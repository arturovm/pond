package users

import (
	"database/sql"

	"github.com/arturovm/pond/internal/pond"
)

// SQLite is a SQLite-backed adapter for the pond.Users port.
type SQLite struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite Users adapter.
func NewSQLite(db *sql.DB) *SQLite {
	return &SQLite{db: db}
}

var _ pond.Users = (*SQLite)(nil)

// Exists reports whether a username is present in the database.
func (s *SQLite) Exists(username string) (bool, error) {
	var exists bool
	err := s.db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)`,
		username,
	).Scan(&exists)
	return exists, err
}

// Save persists a user to the database.
func (s *SQLite) Save(u pond.User) error {
	_, err := s.db.Exec(
		`INSERT INTO users (id, username) VALUES (?, ?)`,
		u.ID, u.Username,
	)
	return err
}
