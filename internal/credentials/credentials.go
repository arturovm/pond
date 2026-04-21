package credentials

import (
	"database/sql"

	"github.com/arturovm/pond/internal/pond"
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

// Save persists a credential to the database.
func (s *SQLite) Save(c pond.Credential) error {
	_, err := s.db.Exec(
		`INSERT INTO credentials (user_id, hash, salt) VALUES (?, ?, ?)`,
		c.UserID.String(), c.Hash, c.Salt,
	)
	return err
}
