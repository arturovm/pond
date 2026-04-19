package api_test

import (
	"bytes"
	"errors"
	"log/slog"
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

func (m *mockSubscriber) Subscribe(userID, feedURL string) error {
	m.calledWith = feedURL
	return nil
}

var _ pond.Subscriber = (*mockSubscriber)(nil)

type errorSubscriber struct {
	err error
}

func (e *errorSubscriber) Subscribe(userID, feedURL string) error {
	return e.err
}

var _ pond.Subscriber = (*errorSubscriber)(nil)

func TestSubscribeHandler_MalformedJSON_ReturnsBadRequest(t *testing.T) {
	mock := &mockSubscriber{}
	handler := api.NewSubscribeHandler(mock, slog.Default())

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSubscribeHandler_EmptyURL_ReturnsBadRequest(t *testing.T) {
	mock := &mockSubscriber{}
	handler := api.NewSubscribeHandler(mock, slog.Default())

	body := strings.NewReader(`{"url":""}`)
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if mock.calledWith != "" {
		t.Errorf("expected Subscribe not to be called, but it was called with %q", mock.calledWith)
	}
}

func TestSubscribeHandler_MalformedURL_ReturnsBadRequest(t *testing.T) {
	mock := &mockSubscriber{}
	handler := api.NewSubscribeHandler(mock, slog.Default())

	body := strings.NewReader(`{"url":"not a url"}`)
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if mock.calledWith != "" {
		t.Errorf("expected Subscribe not to be called, but it was called with %q", mock.calledWith)
	}
}

func TestSubscribeHandler_SubscribeError_LogsError(t *testing.T) {
	sub := &errorSubscriber{err: errors.New("fetch failed")}
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	handler := api.NewSubscribeHandler(sub, logger)

	body := strings.NewReader(`{"url":"https://example.com/feed.rss"}`)
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !strings.Contains(buf.String(), "fetch failed") {
		t.Errorf("expected log to contain error message, got: %s", buf.String())
	}
}

func TestSubscribeHandler_ValidURL_ForwardsToPort(t *testing.T) {
	mock := &mockSubscriber{}
	handler := api.NewSubscribeHandler(mock, slog.Default())

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
