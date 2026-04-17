package database

import (
	"database/sql"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

const driverName = "sqlite3"

// Open opens a SQLite database at the given path and returns the connection.
func Open(path string) (*sql.DB, error) {
	dsn := "file:" + path
	return sql.Open(driverName, dsn)
}

// Migrate applies all pending migrations from the given directory.
func Migrate(db *sql.DB, dir string) error {
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	slog.Debug("running migrations", "dir", dir)
	return goose.Up(db, dir)
}
