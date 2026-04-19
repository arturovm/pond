package database

import (
	"database/sql"
	"embed"
	"log/slog"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

//go:embed migrations
var migrations embed.FS

const driverName = "sqlite3"

// Open opens a SQLite database at the given path and returns the connection.
func Open(path string) (*sql.DB, error) {
	dsn := "file:" + path
	return sql.Open(driverName, dsn)
}

// Migrate applies all pending migrations.
func Migrate(db *sql.DB) error {
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("sqlite3"); err != nil {
		return err
	}
	goose.SetBaseFS(migrations)
	slog.Debug("running migrations")
	return goose.Up(db, "migrations")
}
