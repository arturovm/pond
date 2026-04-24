package credentials

import (
	"database/sql"

	"github.com/arturovm/pond/internal/pond"
	"github.com/google/uuid"
)

// SQLite is a SQLite-backed adapter for the pond.Credentials port.
type SQLite struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite Credentials adapter.
func NewSQLite(db *sql.DB) *SQLite {
	return &SQLite{db: db}
}

var _ pond.Credentials = (*SQLite)(nil)

// FindByUserID retrieves a credential by user ID.
func (s *SQLite) FindByUserID(userID uuid.UUID) (pond.Credential, error) {
	var c pond.Credential
	var userIDStr string
	err := s.db.QueryRow(
		`SELECT user_id, hash, salt FROM credentials WHERE user_id = ?`,
		userID.String(),
	).Scan(&userIDStr, &c.Hash, &c.Salt)
	if err != nil {
		return pond.Credential{}, err
	}
	c.UserID, err = uuid.Parse(userIDStr)
	if err != nil {
		return pond.Credential{}, err
	}
	return c, nil
}

// Save persists a credential to the database.
func (s *SQLite) Save(c pond.Credential) error {
	_, err := s.db.Exec(
		`INSERT INTO credentials (user_id, hash, salt) VALUES (?, ?, ?)`,
		c.UserID.String(), c.Hash, c.Salt,
	)
	return err
}
