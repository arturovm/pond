package api_test

import (
	"bytes"
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

type mockAccountCreator struct {
	calledWithUsername string
	calledWithPassword string
	calledWithIP       netip.Addr
	token              string
}

func (m *mockAccountCreator) CreateAccount(username, password string, ip netip.Addr) (string, error) {
	m.calledWithUsername = username
	m.calledWithPassword = password
	m.calledWithIP = ip
	return m.token, nil
}

var _ pond.AccountCreator = (*mockAccountCreator)(nil)

type errorAccountCreator struct {
	err error
}

func (e *errorAccountCreator) CreateAccount(_, _ string, _ netip.Addr) (string, error) {
	return "", e.err
}

var _ pond.AccountCreator = (*errorAccountCreator)(nil)

func TestCreateAccountHandler_UsernameTaken_Returns409(t *testing.T) {
	ac := &errorAccountCreator{err: pond.ErrUsernameTaken}
	handler := api.NewCreateAccountHandler(ac, slog.Default())

	body := strings.NewReader(`{"username":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestCreateAccountHandler_PortError_LogsAndReturns500(t *testing.T) {
	ac := &errorAccountCreator{err: errors.New("db failed")}
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	handler := api.NewCreateAccountHandler(ac, logger)

	body := strings.NewReader(`{"username":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", body)
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

func TestCreateAccountHandler_EmptyUsername_ReturnsBadRequest(t *testing.T) {
	mock := &mockAccountCreator{}
	handler := api.NewCreateAccountHandler(mock, slog.Default())

	body := strings.NewReader(`{"username":"","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if mock.calledWithUsername != "" {
		t.Errorf("expected CreateAccount not to be called, but it was called with username %q", mock.calledWithUsername)
	}
}

func TestCreateAccountHandler_EmptyPassword_ReturnsBadRequest(t *testing.T) {
	mock := &mockAccountCreator{}
	handler := api.NewCreateAccountHandler(mock, slog.Default())

	body := strings.NewReader(`{"username":"alice","password":""}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
	if mock.calledWithUsername != "" {
		t.Errorf("expected CreateAccount not to be called, but it was called with username %q", mock.calledWithUsername)
	}
}

func TestCreateAccountHandler_ValidCredentials_ForwardsToPort(t *testing.T) {
	mock := &mockAccountCreator{}
	handler := api.NewCreateAccountHandler(mock, slog.Default())

	body := strings.NewReader(`{"username":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}
	if mock.calledWithUsername != "alice" {
		t.Errorf("expected CreateAccount called with username %q, got %q", "alice", mock.calledWithUsername)
	}
	if mock.calledWithPassword != "secret" {
		t.Errorf("expected CreateAccount called with password %q, got %q", "secret", mock.calledWithPassword)
	}
}

func TestCreateAccountHandler_MalformedJSON_ReturnsBadRequest(t *testing.T) {
	mock := &mockAccountCreator{}
	handler := api.NewCreateAccountHandler(mock, slog.Default())

	body := strings.NewReader(`not json`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateAccountHandler_ValidCredentials_ReturnsSessionTokenInBody(t *testing.T) {
	mock := &mockAccountCreator{token: "abc123token"}
	handler := api.NewCreateAccountHandler(mock, slog.Default())

	body := strings.NewReader(`{"username":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	var resp struct {
		SessionToken string `json:"session_token"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if resp.SessionToken != "abc123token" {
		t.Errorf("expected session_token %q, got %q", "abc123token", resp.SessionToken)
	}
}

func TestCreateAccountHandler_ForwardsClientIP(t *testing.T) {
	mock := &mockAccountCreator{}
	handler := api.NewCreateAccountHandler(mock, slog.Default())

	body := strings.NewReader(`{"username":"alice","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/accounts", body)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.0.2.1:4321"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	want := netip.MustParseAddr("192.0.2.1")
	if mock.calledWithIP != want {
		t.Errorf("expected CreateAccount called with IP %v, got %v", want, mock.calledWithIP)
	}
}
