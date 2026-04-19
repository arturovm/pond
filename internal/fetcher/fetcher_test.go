package fetcher_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/arturovm/pond/internal/fetcher"
)

func TestHTTPFetcher_ReturnsErrorForNon2xxResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	f := fetcher.NewHTTPFetcher()
	_, err := f.Fetch(srv.URL)

	if err == nil {
		t.Fatal("expected error for non-2xx response, got nil")
	}
}

func TestHTTPFetcher_FetchesContentFromURL(t *testing.T) {
	const body = `<?xml version="1.0"?><rss></rss>`
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	f := fetcher.NewHTTPFetcher()
	feed, err := f.Fetch(srv.URL)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(feed.Body) != body {
		t.Errorf("expected body %q, got %q", body, string(feed.Body))
	}
}
