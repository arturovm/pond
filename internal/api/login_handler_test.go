package api_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"

	"github.com/arturovm/pond/internal/api"
	"github.com/arturovm/pond/internal/pond"
)

type mockAuthenticator struct {
	calledWithUsername string
	calledWithPassword string
	calledWithIP       netip.Addr
	token              []byte
}

func (m *mockAuthenticator) Authenticate(username, password string, ip netip.Addr) ([]byte, error) {
	m.calledWithUsername = username
	m.calledWithPassword = password
	m.calledWithIP = ip
	return m.token, nil
}

var _ pond.Authenticator = (*mockAuthenticator)(nil)

type errorAuthenticator struct {
	err error
}

func (e *errorAuthenticator) Authenticate(_, _ string, _ netip.Addr) ([]byte, error) {
	return nil, e.err
}

var _ pond.Authenticator = (*errorAuthenticator)(nil)

func TestLoginHandler_EmptyUsername_ReturnsBadRequest(t *testing.T) {
	mock := &mockAuthenticator{}
	handler := api.NewLoginHandler(mock, slog.Default())

	body := strings.NewReader(`{"username":"","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if mock.calledWithUsername != "" {
		t.Errorf("expected Authenticate not to be called, but it was called with username %q", mock.calledWithUsername)
	}
}

func TestLoginHandler_EmptyPassword_ReturnsBadRequest(t *testing.T) {
	mock := &mockAuthenticator{}
	handler := api.NewLoginHandler(mock, slog.Default())

	body := strings.NewReader(`{"username":"alice","password":""}`)
	req := httptest.NewRequest(http.MethodPost, "/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if mock.calledWithUsername != "" {
		t.Errorf("expected Authenticate not to be called, but it was called with username %q", mock.calledWithUsername)
	}
}

func TestLoginHandler_PortError_LogsAndReturns500(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	handler := api.NewLoginHandler(&errorAuthenticator{err: errors.New("db failed")}, logger)

	body := strings.NewReader(`{"username":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
	if !strings.Contains(buf.String(), "db failed") {
		t.Errorf("expected log to contain error message, got: %s", buf.String())
	}
}

func TestLoginHandler_ForwardsClientIP(t *testing.T) {
	mock := &mockAuthenticator{}
	handler := api.NewLoginHandler(mock, slog.Default())

	body := strings.NewReader(`{"username":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.0.2.1:4321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	want := netip.MustParseAddr("192.0.2.1")
	if mock.calledWithIP != want {
		t.Errorf("expected Authenticate called with IP %v, got %v", want, mock.calledWithIP)
	}
}

func TestLoginHandler_InvalidCredentials_Returns401(t *testing.T) {
	handler := api.NewLoginHandler(&errorAuthenticator{err: pond.ErrInvalidCredentials}, slog.Default())

	body := strings.NewReader(`{"username":"alice","password":"wrong"}`)
	req := httptest.NewRequest(http.MethodPost, "/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestLoginHandler_MalformedJSON_ReturnsBadRequest(t *testing.T) {
	mock := &mockAuthenticator{}
	handler := api.NewLoginHandler(mock, slog.Default())

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if mock.calledWithUsername != "" {
		t.Errorf("expected Authenticate not to be called, but it was called with username %q", mock.calledWithUsername)
	}
}

func TestLoginHandler_ValidCredentials_ForwardsToPortAndReturns202WithToken(t *testing.T) {
	tokenBytes := []byte("abc123token")
	mock := &mockAuthenticator{token: tokenBytes}
	handler := api.NewLoginHandler(mock, slog.Default())

	body := strings.NewReader(`{"username":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/sessions", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
	if mock.calledWithUsername != "alice" {
		t.Errorf("expected Authenticate called with username %q, got %q", "alice", mock.calledWithUsername)
	}
	if mock.calledWithPassword != "secret" {
		t.Errorf("expected Authenticate called with password %q, got %q", "secret", mock.calledWithPassword)
	}
	var resp struct {
		SessionToken string `json:"session_token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	want := hex.EncodeToString(tokenBytes)
	if resp.SessionToken != want {
		t.Errorf("expected session_token %q, got %q", want, resp.SessionToken)
	}
}
