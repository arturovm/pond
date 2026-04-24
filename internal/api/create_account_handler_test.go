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

type mockAccountCreator struct {
	calledWithUsername string
	calledWithPassword string
	calledWithIP       netip.Addr
	token              []byte
}

func (m *mockAccountCreator) CreateAccount(username, password string, ip netip.Addr) ([]byte, error) {
	m.calledWithUsername = username
	m.calledWithPassword = password
	m.calledWithIP = ip
	return m.token, nil
}

var _ pond.AccountCreator = (*mockAccountCreator)(nil)

type errorAccountCreator struct {
	err error
}

func (e *errorAccountCreator) CreateAccount(_, _ string, _ netip.Addr) ([]byte, error) {
	return nil, e.err
}

var _ pond.AccountCreator = (*errorAccountCreator)(nil)

func TestCreateAccountHandler_UsernameTaken_Returns409(t *testing.T) {
	ac := &errorAccountCreator{err: pond.ErrUsernameTaken}
	handler := api.NewCreateAccountHandler(ac, slog.Default())

	rec := doPost(t, handler, `{"username":"alice","password":"secret"}`)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected status %d, got %d", http.StatusConflict, rec.Code)
	}
}

func TestCreateAccountHandler_PortError_LogsAndReturns500(t *testing.T) {
	ac := &errorAccountCreator{err: errors.New("db failed")}
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	handler := api.NewCreateAccountHandler(ac, logger)

	rec := doPost(t, handler, `{"username":"alice","password":"secret"}`)

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

	rec := doPost(t, handler, `{"username":"","password":"secret"}`)

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

	rec := doPost(t, handler, `{"username":"alice","password":""}`)

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

	rec := doPost(t, handler, `{"username":"alice","password":"secret"}`)

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

	rec := doPost(t, handler, `not json`)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestCreateAccountHandler_ValidCredentials_ReturnsSessionTokenInBody(t *testing.T) {
	tokenBytes := []byte("abc123token")
	mock := &mockAccountCreator{token: tokenBytes}
	handler := api.NewCreateAccountHandler(mock, slog.Default())

	rec := doPost(t, handler, `{"username":"alice","password":"secret"}`)

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

func TestCreateAccountHandler_ForwardsClientIP(t *testing.T) {
	mock := &mockAccountCreator{}
	handler := api.NewCreateAccountHandler(mock, slog.Default())

	req := httptest.NewRequest(http.MethodPost, "/accounts", strings.NewReader(`{"username":"alice","password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "192.0.2.1:4321"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	want := netip.MustParseAddr("192.0.2.1")
	if mock.calledWithIP != want {
		t.Errorf("expected CreateAccount called with IP %v, got %v", want, mock.calledWithIP)
	}
}
