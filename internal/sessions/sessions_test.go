package sessions_test

import (
	"database/sql"
	"net/netip"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/arturovm/pond/internal/pond"
	"github.com/arturovm/pond/internal/sessions"
	"github.com/google/uuid"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE sessions (
		token      BLOB PRIMARY KEY,
		user_id    TEXT NOT NULL,
		ip         TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		expires_at INTEGER NOT NULL
	)`)
	if err != nil {
		t.Fatalf("failed to create sessions table: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSQLiteSessions_Save_PersistsSessionInDB(t *testing.T) {
	db := openTestDB(t)
	repo := sessions.NewSQLite(db)
	sess := pond.Session{
		Token:     []byte("tokentokentokenx"),
		UserID:    uuid.MustParse("01960000-0000-7000-8000-000000000001"),
		IP:        netip.MustParseAddr("192.0.2.1"),
		CreatedAt: time.Now().UTC().Truncate(time.Second),
		ExpiresAt: time.Now().UTC().Truncate(time.Second).Add(30 * 24 * time.Hour),
	}

	err := repo.Save(sess)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var gotToken []byte
	err = db.QueryRow(`SELECT token FROM sessions WHERE token = ?`, sess.Token).Scan(&gotToken)
	if err != nil {
		t.Fatalf("row not found: %v", err)
	}
	if string(gotToken) != string(sess.Token) {
		t.Errorf("expected token %x, got %x", sess.Token, gotToken)
	}
}

func TestSQLiteSessions_Save_DBFailure_ReturnsError(t *testing.T) {
	db := openTestDB(t)
	repo := sessions.NewSQLite(db)
	db.Close()

	err := repo.Save(pond.Session{
		Token:     []byte("tokentokentokenx"),
		UserID:    uuid.MustParse("01960000-0000-7000-8000-000000000001"),
		IP:        netip.MustParseAddr("192.0.2.1"),
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(30 * 24 * time.Hour),
	})

	if err == nil {
		t.Error("expected an error when DB is closed, got nil")
	}
}
