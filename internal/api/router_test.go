package api_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arturovm/pond/internal/api"
)

func TestRouter_GETSubscriptions_Returns405(t *testing.T) {
	mock := &mockSubscriber{}
	router := api.NewRouter(mock, &mockAccountCreator{}, &mockAuthenticator{}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestRouter_POSTSubscriptions_RoutesToSubscribeHandler(t *testing.T) {
	mock := &mockSubscriber{}
	router := api.NewRouter(mock, &mockAccountCreator{}, &mockAuthenticator{}, slog.Default())

	body := strings.NewReader(`{"url":"https://example.com/feed.rss"}`)
	req := httptest.NewRequest(http.MethodPost, "/subscriptions", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
}

func TestRouter_POSTAccounts_RoutesToCreateAccountHandler(t *testing.T) {
	mock := &mockAccountCreator{token: []byte("tokentokentokenx")}
	router := api.NewRouter(&mockSubscriber{}, mock, &mockAuthenticator{}, slog.Default())

	body := strings.NewReader(`{"username":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
}

func TestRouter_GETAccounts_Returns405(t *testing.T) {
	router := api.NewRouter(&mockSubscriber{}, &mockAccountCreator{}, &mockAuthenticator{}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/accounts", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestRouter_GETSessions_Returns405(t *testing.T) {
	router := api.NewRouter(&mockSubscriber{}, &mockAccountCreator{}, &mockAuthenticator{}, slog.Default())

	req := httptest.NewRequest(http.MethodGet, "/sessions", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestRouter_POSTSessions_RoutesToLoginHandler(t *testing.T) {
	mock := &mockAuthenticator{token: []byte("tokentokentokenx")}
	router := api.NewRouter(&mockSubscriber{}, &mockAccountCreator{}, mock, slog.Default())

	body := strings.NewReader(`{"username":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
}
