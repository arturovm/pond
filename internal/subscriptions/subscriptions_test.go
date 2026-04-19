package subscriptions_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"github.com/arturovm/pond/internal/pond"
	"github.com/arturovm/pond/internal/subscriptions"
)

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE subscriptions (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id     TEXT NOT NULL,
		source_link TEXT NOT NULL,
		UNIQUE(user_id, source_link)
	)`)
	if err != nil {
		t.Fatalf("failed to create subscriptions table: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSQLiteSubscriptions_Save_ReturnsErrorOnDBFailure(t *testing.T) {
	db := openTestDB(t)
	repo := subscriptions.NewSQLite(db)
	db.Close()

	err := repo.Save(pond.Subscription{UserID: "user-1", Source: pond.Source{Link: "https://example.com/feed.rss"}})

	if err == nil {
		t.Error("expected an error when DB is closed, got nil")
	}
}

func TestSQLiteSubscriptions_Save_PersistsSubscription(t *testing.T) {
	db := openTestDB(t)
	repo := subscriptions.NewSQLite(db)

	sub := pond.Subscription{
		UserID: "user-1",
		Source: pond.Source{Link: "https://example.com/feed.rss"},
	}
	err := repo.Save(sub)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var gotUserID, gotSourceLink string
	row := db.QueryRow(`SELECT user_id, source_link FROM subscriptions WHERE user_id = ? AND source_link = ?`, sub.UserID, sub.Source.Link)
	if err := row.Scan(&gotUserID, &gotSourceLink); err != nil {
		t.Fatalf("subscription not found in database: %v", err)
	}
	if gotUserID != sub.UserID || gotSourceLink != sub.Source.Link {
		t.Errorf("expected user_id=%q source_link=%q, got user_id=%q source_link=%q",
			sub.UserID, sub.Source.Link, gotUserID, gotSourceLink)
	}
}
