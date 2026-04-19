package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arturovm/pond/internal/api"
)

func TestRouter_GETSubscriptions_Returns405(t *testing.T) {
	mock := &mockSubscriber{}
	router := api.NewRouter(mock)

	req := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestRouter_POSTSubscriptions_RoutesToSubscribeHandler(t *testing.T) {
	mock := &mockSubscriber{}
	router := api.NewRouter(mock)

	body := strings.NewReader(`{"url":"https://example.com/feed.rss"}`)
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
}
