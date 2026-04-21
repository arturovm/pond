package subscriptions

import (
	"database/sql"

	"github.com/arturovm/pond/internal/pond"
)

// SQLite is a SQLite-backed adapter for the pond.Subscriptions port.
type SQLite struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite Subscriptions adapter.
func NewSQLite(db *sql.DB) *SQLite {
	return &SQLite{db: db}
}

var _ pond.Subscriptions = (*SQLite)(nil)

// Save persists a Subscription to the database.
func (s *SQLite) Save(sub pond.Subscription) error {
	_, err := s.db.Exec(
		`INSERT INTO subscriptions (user_id, source_link) VALUES (?, ?)`,
		sub.UserID.String(), sub.Source.Link,
	)
	return err
}
