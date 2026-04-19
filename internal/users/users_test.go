package users_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/arturovm/pond/internal/users"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE users (
		username TEXT NOT NULL UNIQUE
	)`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSQLiteUsers_Exists_DBFailure_ReturnsError(t *testing.T) {
	db := openTestDB(t)
	repo := users.NewSQLite(db)
	db.Close()

	_, err := repo.Exists("alice")

	if err == nil {
		t.Error("expected an error when DB is closed, got nil")
	}
}

func TestSQLiteUsers_Exists_KnownUsername_ReturnsTrue(t *testing.T) {
	db := openTestDB(t)
	repo := users.NewSQLite(db)
	_, err := db.Exec(`INSERT INTO users (username) VALUES (?)`, "alice")
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	exists, err := repo.Exists("alice")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("expected true for known username, got false")
	}
}

func TestSQLiteUsers_Exists_UnknownUsername_ReturnsFalse(t *testing.T) {
	db := openTestDB(t)
	repo := users.NewSQLite(db)

	exists, err := repo.Exists("alice")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("expected false for unknown username, got true")
	}
}
