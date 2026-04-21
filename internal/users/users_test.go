package users_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/arturovm/pond/internal/pond"
	"github.com/arturovm/pond/internal/users"
	"github.com/google/uuid"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE users (
		id       TEXT PRIMARY KEY,
		username TEXT NOT NULL UNIQUE
	)`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSQLiteUsers_Save_PersistsUserInDB(t *testing.T) {
	db := openTestDB(t)
	repo := users.NewSQLite(db)
	u := pond.User{ID: uuid.MustParse("01960000-0000-7000-8000-000000000001"), Username: "alice"}

	err := repo.Save(u)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var gotID, gotUsername string
	err = db.QueryRow(`SELECT id, username FROM users WHERE id = ?`, u.ID.String()).Scan(&gotID, &gotUsername)
	if err != nil {
		t.Fatalf("row not found: %v", err)
	}
	if gotID != u.ID.String() {
		t.Errorf("expected id %q, got %q", u.ID.String(), gotID)
	}
	if gotUsername != u.Username {
		t.Errorf("expected username %q, got %q", u.Username, gotUsername)
	}
}

func TestSQLiteUsers_Save_DBFailure_ReturnsError(t *testing.T) {
	db := openTestDB(t)
	repo := users.NewSQLite(db)
	db.Close()

	err := repo.Save(pond.User{ID: uuid.MustParse("01960000-0000-7000-8000-000000000001"), Username: "alice"})

	if err == nil {
		t.Error("expected an error when DB is closed, got nil")
	}
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
	_, err := db.Exec(`INSERT INTO users (id, username) VALUES (?, ?)`, "test-id", "alice")
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
