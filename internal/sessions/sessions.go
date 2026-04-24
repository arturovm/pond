package sessions

import (
	"database/sql"

	"github.com/arturovm/pond/internal/pond"
)

// SQLite is a SQLite-backed adapter for the pond.Sessions port.
type SQLite struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite Sessions adapter.
func NewSQLite(db *sql.DB) *SQLite {
	return &SQLite{db: db}
}

var _ pond.Sessions = (*SQLite)(nil)

// Save persists a session to the database.
func (s *SQLite) Save(sess pond.Session) error {
	_, err := s.db.Exec(
		`INSERT INTO sessions (token, user_id, ip, created_at, expires_at) VALUES (?, ?, ?, ?, ?)`,
		sess.Token,
		sess.UserID.String(),
		sess.IP.String(),
		sess.CreatedAt.Unix(),
		sess.ExpiresAt.Unix(),
	)
	return err
}
