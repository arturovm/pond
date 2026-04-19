package sources_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/arturovm/pond/internal/pond"
	"github.com/arturovm/pond/internal/sources"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE sources (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		title       TEXT NOT NULL,
		link        TEXT NOT NULL UNIQUE,
		description TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatalf("failed to create sources table: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSQLiteSources_Save_ReturnsErrorOnDBFailure(t *testing.T) {
	db := openTestDB(t)
	repo := sources.NewSQLite(db)
	db.Close() // force all subsequent ops to fail

	err := repo.Save(pond.Source{Title: "T", Link: "https://example.com", Description: "D"})

	if err == nil {
		t.Error("expected an error when DB is closed, got nil")
	}
}

func TestSQLiteSources_Save_PersistsSource(t *testing.T) {
	db := openTestDB(t)
	repo := sources.NewSQLite(db)

	src := pond.Source{Title: "My Feed", Link: "https://example.com", Description: "A feed about things."}
	err := repo.Save(src)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var got pond.Source
	row := db.QueryRow(`SELECT title, link, description FROM sources WHERE link = ?`, src.Link)
	if err := row.Scan(&got.Title, &got.Link, &got.Description); err != nil {
		t.Fatalf("source not found in database: %v", err)
	}
	if got != src {
		t.Errorf("expected %+v, got %+v", src, got)
	}
}
