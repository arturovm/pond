package credentials_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/arturovm/pond/internal/credentials"
	"github.com/arturovm/pond/internal/pond"
	"github.com/google/uuid"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE credentials (
		user_id TEXT PRIMARY KEY,
		hash    BLOB NOT NULL,
		salt    BLOB NOT NULL
	)`)
	if err != nil {
		t.Fatalf("failed to create credentials table: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSQLiteCredentials_Save_DBFailure_ReturnsError(t *testing.T) {
	db := openTestDB(t)
	repo := credentials.NewSQLite(db)
	db.Close()

	err := repo.Save(pond.Credential{UserID: uuid.MustParse("01960000-0000-7000-8000-000000000001"), Hash: []byte("h"), Salt: []byte("s")})

	if err == nil {
		t.Error("expected an error when DB is closed, got nil")
	}
}

func TestSQLiteCredentials_Save_PersistsCredentialInDB(t *testing.T) {
	db := openTestDB(t)
	repo := credentials.NewSQLite(db)
	c := pond.Credential{
		UserID: uuid.MustParse("01960000-0000-7000-8000-000000000001"),
		Hash:   []byte("hashvalue"),
		Salt:   []byte("saltvalue"),
	}

	err := repo.Save(c)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var gotUserID string
	var gotHash, gotSalt []byte
	err = db.QueryRow(`SELECT user_id, hash, salt FROM credentials WHERE user_id = ?`, c.UserID.String()).
		Scan(&gotUserID, &gotHash, &gotSalt)
	if err != nil {
		t.Fatalf("row not found: %v", err)
	}
	if gotUserID != c.UserID.String() {
		t.Errorf("expected user_id %q, got %q", c.UserID.String(), gotUserID)
	}
	if string(gotHash) != string(c.Hash) {
		t.Errorf("expected hash %q, got %q", c.Hash, gotHash)
	}
	if string(gotSalt) != string(c.Salt) {
		t.Errorf("expected salt %q, got %q", c.Salt, gotSalt)
	}
}
