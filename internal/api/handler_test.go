package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arturovm/pond/internal/api"
	"github.com/arturovm/pond/internal/pond"
)

type mockSubscriber struct {
	calledWith string
}

func (m *mockSubscriber) Subscribe(feedURL string) error {
	m.calledWith = feedURL
	return nil
}

var _ pond.Subscriber = (*mockSubscriber)(nil)

func TestSubscribeHandler_ValidURL_ForwardsToPort(t *testing.T) {
	mock := &mockSubscriber{}
	handler := api.NewSubscribeHandler(mock)

	body := strings.NewReader(`{"url":"https://example.com/feed.rss"}`)
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
	if mock.calledWith != "https://example.com/feed.rss" {
		t.Errorf("expected Subscribe to be called with %q, got %q", "https://example.com/feed.rss", mock.calledWith)
	}
}
