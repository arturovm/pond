package sources

import (
	"database/sql"

	"github.com/arturovm/pond/internal/pond"
)

// SQLite is a SQLite-backed adapter for the pond.Sources port.
type SQLite struct {
	db *sql.DB
}

// NewSQLite creates a new SQLite Sources adapter.
func NewSQLite(db *sql.DB) *SQLite {
	return &SQLite{db: db}
}

var _ pond.Sources = (*SQLite)(nil)

// Save persists a Source to the database.
func (s *SQLite) Save(src pond.Source) error {
	_, err := s.db.Exec(
		`INSERT INTO sources (title, link, description) VALUES (?, ?, ?)`,
		src.Title, src.Link, src.Description,
	)
	return err
}
