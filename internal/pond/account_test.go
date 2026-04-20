package pond_test

import (
	"errors"
	"net/netip"
	"testing"

	"github.com/arturovm/pond/internal/pond"
)

type mockUsers struct {
	exists  bool
	err     error
	saved   pond.User
	saveErr error
}

func (m *mockUsers) Exists(username string) (bool, error) {
	return m.exists, m.err
}

func (m *mockUsers) Save(u pond.User) error {
	m.saved = u
	return m.saveErr
}

var _ pond.Users = (*mockUsers)(nil)

type mockCredentials struct {
	saved   pond.Credential
	saveErr error
}

func (m *mockCredentials) Save(c pond.Credential) error {
	m.saved = c
	return m.saveErr
}

var _ pond.Credentials = (*mockCredentials)(nil)

type mockSessions struct {
	saved   pond.Session
	saveErr error
}

func (m *mockSessions) Save(s pond.Session) error {
	m.saved = s
	return m.saveErr
}

var _ pond.Sessions = (*mockSessions)(nil)

func TestAccountService_CreateAccount_ExistingUsername_ReturnsErrUsernameTaken(t *testing.T) {
	users := &mockUsers{exists: true}
	s := pond.NewAccountService(users, nil, nil)

	_, err := s.CreateAccount("alice", "secret", netip.Addr{})

	if !errors.Is(err, pond.ErrUsernameTaken) {
		t.Errorf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestAccountService_CreateAccount_UsersPortError_ReturnsError(t *testing.T) {
	portErr := errors.New("db failed")
	users := &mockUsers{err: portErr}
	s := pond.NewAccountService(users, nil, nil)

	_, err := s.CreateAccount("alice", "secret", netip.Addr{})

	if !errors.Is(err, portErr) {
		t.Errorf("expected %v, got %v", portErr, err)
	}
}

func TestAccountService_CreateAccount_NewUsername_SavesUserInUsers(t *testing.T) {
	users := &mockUsers{exists: false}
	s := pond.NewAccountService(users, &mockCredentials{}, &mockSessions{})

	_, err := s.CreateAccount("alice", "secret", netip.Addr{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if users.saved.Username != "alice" {
		t.Errorf("expected saved username %q, got %q", "alice", users.saved.Username)
	}
	if users.saved.ID == "" {
		t.Error("expected non-empty user ID")
	}
}

func TestAccountService_CreateAccount_SaveError_ReturnsError(t *testing.T) {
	saveErr := errors.New("save failed")
	users := &mockUsers{exists: false, saveErr: saveErr}
	s := pond.NewAccountService(users, nil, nil)

	_, err := s.CreateAccount("alice", "secret", netip.Addr{})

	if !errors.Is(err, saveErr) {
		t.Errorf("expected %v, got %v", saveErr, err)
	}
}

func TestAccountService_CreateAccount_NewUsername_DoesNotReturnError(t *testing.T) {
	users := &mockUsers{exists: false}
	credentials := &mockCredentials{}
	s := pond.NewAccountService(users, credentials, &mockSessions{})

	_, err := s.CreateAccount("alice", "secret", netip.Addr{})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestAccountService_CreateAccount_CredentialsPortError_ReturnsError(t *testing.T) {
	credErr := errors.New("credentials save failed")
	users := &mockUsers{exists: false}
	credentials := &mockCredentials{saveErr: credErr}
	s := pond.NewAccountService(users, credentials, &mockSessions{})

	_, err := s.CreateAccount("alice", "secret", netip.Addr{})

	if !errors.Is(err, credErr) {
		t.Errorf("expected %v, got %v", credErr, err)
	}
}

func TestAccountService_CreateAccount_SavesCredentialWithNonEmptyHashAndSalt(t *testing.T) {
	users := &mockUsers{exists: false}
	credentials := &mockCredentials{}
	s := pond.NewAccountService(users, credentials, &mockSessions{})

	_, err := s.CreateAccount("alice", "secret", netip.Addr{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(credentials.saved.Hash) == 0 {
		t.Error("expected non-empty Hash in saved credential")
	}
	if len(credentials.saved.Salt) == 0 {
		t.Error("expected non-empty Salt in saved credential")
	}
}

func TestAccountService_CreateAccount_SavesCredentialWithMatchingUserID(t *testing.T) {
	users := &mockUsers{exists: false}
	credentials := &mockCredentials{}
	s := pond.NewAccountService(users, credentials, &mockSessions{})

	_, err := s.CreateAccount("alice", "secret", netip.Addr{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if credentials.saved.UserID != users.saved.ID {
		t.Errorf("expected credential UserID %q to match saved user ID %q", credentials.saved.UserID, users.saved.ID)
	}
}

func TestAccountService_CreateAccount_TwoCallsProduceDifferentSessionTokens(t *testing.T) {
	s := pond.NewAccountService(&mockUsers{exists: false}, &mockCredentials{}, &mockSessions{})

	token1, err := s.CreateAccount("alice", "secret1", netip.Addr{})
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}
	token2, err := s.CreateAccount("bob", "secret2", netip.Addr{})
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	if token1 == token2 {
		t.Error("expected two calls to produce different session tokens")
	}
}

func TestAccountService_CreateAccount_SessionTokenIsAtLeast32Characters(t *testing.T) {
	users := &mockUsers{exists: false}
	credentials := &mockCredentials{}
	s := pond.NewAccountService(users, credentials, &mockSessions{})

	token, err := s.CreateAccount("alice", "secret", netip.Addr{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token) < 32 {
		t.Errorf("expected token length >= 32, got %d", len(token))
	}
}

func TestAccountService_CreateAccount_ReturnsNonEmptySessionToken(t *testing.T) {
	users := &mockUsers{exists: false}
	credentials := &mockCredentials{}
	s := pond.NewAccountService(users, credentials, &mockSessions{})

	token, err := s.CreateAccount("alice", "secret", netip.Addr{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty session token")
	}
}

func TestAccountService_CreateAccount_SavesSessionWithIP(t *testing.T) {
	users := &mockUsers{exists: false}
	sessions := &mockSessions{}
	s := pond.NewAccountService(users, &mockCredentials{}, sessions)

	ip := netip.MustParseAddr("192.0.2.1")
	_, err := s.CreateAccount("alice", "secret", ip)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessions.saved.IP != ip {
		t.Errorf("expected saved session IP %v, got %v", ip, sessions.saved.IP)
	}
}

func TestAccountService_CreateAccount_SessionsPortError_ReturnsError(t *testing.T) {
	sessErr := errors.New("sessions save failed")
	sessions := &mockSessions{saveErr: sessErr}
	s := pond.NewAccountService(&mockUsers{exists: false}, &mockCredentials{}, sessions)

	_, err := s.CreateAccount("alice", "secret", netip.Addr{})

	if !errors.Is(err, sessErr) {
		t.Errorf("expected %v, got %v", sessErr, err)
	}
}

func TestAccountService_CreateAccount_SavesSessionWithExpiresAtAfterCreatedAt(t *testing.T) {
	users := &mockUsers{exists: false}
	sessions := &mockSessions{}
	s := pond.NewAccountService(users, &mockCredentials{}, sessions)

	_, err := s.CreateAccount("alice", "secret", netip.Addr{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sessions.saved.ExpiresAt.After(sessions.saved.CreatedAt) {
		t.Errorf("expected ExpiresAt %v to be after CreatedAt %v", sessions.saved.ExpiresAt, sessions.saved.CreatedAt)
	}
}

func TestAccountService_CreateAccount_SavesSessionWithNonZeroCreatedAt(t *testing.T) {
	users := &mockUsers{exists: false}
	sessions := &mockSessions{}
	s := pond.NewAccountService(users, &mockCredentials{}, sessions)

	_, err := s.CreateAccount("alice", "secret", netip.Addr{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessions.saved.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt on saved session")
	}
}

func TestAccountService_CreateAccount_SavesSessionWithMatchingUserID(t *testing.T) {
	users := &mockUsers{exists: false}
	credentials := &mockCredentials{}
	sessions := &mockSessions{}
	s := pond.NewAccountService(users, credentials, sessions)

	_, err := s.CreateAccount("alice", "secret", netip.Addr{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessions.saved.UserID != users.saved.ID {
		t.Errorf("expected saved session UserID %q to match user ID %q", sessions.saved.UserID, users.saved.ID)
	}
}

func TestAccountService_CreateAccount_SavesSessionWithGeneratedToken(t *testing.T) {
	users := &mockUsers{exists: false}
	credentials := &mockCredentials{}
	sessions := &mockSessions{}
	s := pond.NewAccountService(users, credentials, sessions)

	token, err := s.CreateAccount("alice", "secret", netip.Addr{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessions.saved.Token != token {
		t.Errorf("expected saved session token %q, got %q", token, sessions.saved.Token)
	}
}
