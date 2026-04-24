package pond

import (
	"crypto/rand"
	"errors"
	"net/netip"
	"time"

	"github.com/google/uuid"
)

// User represents a registered user.
type User struct {
	ID       uuid.UUID
	Username string
}

// ErrUsernameTaken is returned when a username already exists.
var ErrUsernameTaken = errors.New("username taken")

// ErrInvalidCredentials is returned when login credentials are incorrect.
var ErrInvalidCredentials = errors.New("invalid credentials")

// ErrUserNotFound is returned when a user cannot be found by username.
var ErrUserNotFound = errors.New("user not found")

// Credential holds the hashed password material for a user.
type Credential struct {
	UserID uuid.UUID
	Hash   []byte
	Salt   []byte
}

const sessionDuration = 30 * 24 * time.Hour

// Session represents an authenticated session.
type Session struct {
	Token     []byte
	UserID    uuid.UUID
	IP        netip.Addr
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Users is the outgoing port for user persistence.
type Users interface {
	Exists(username string) (bool, error)
	Save(User) error
	FindByUsername(username string) (User, error)
}

// Credentials is the outgoing port for persisting credentials.
type Credentials interface {
	Save(Credential) error
	FindByUserID(userID uuid.UUID) (Credential, error)
}

// Sessions is the outgoing port for persisting sessions.
type Sessions interface {
	Save(Session) error
}

// AccountCreator is the incoming port for creating an account.
type AccountCreator interface {
	CreateAccount(username, password string, ip netip.Addr) ([]byte, error)
}

// Authenticator is the incoming port for logging in.
type Authenticator interface {
	Authenticate(username, password string, ip netip.Addr) ([]byte, error)
}

// AccountService implements account management use cases.
type AccountService struct {
	users       Users
	credentials Credentials
	sessions    Sessions
}

func NewAccountService(users Users, credentials Credentials, sessions Sessions) *AccountService {
	return &AccountService{users: users, credentials: credentials, sessions: sessions}
}

func (s *AccountService) Authenticate(username, password string, ip netip.Addr) ([]byte, error) {
	user, err := s.users.FindByUsername(username)
	if errors.Is(err, ErrUserNotFound) {
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}
	if _, err := s.credentials.FindByUserID(user.ID); err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *AccountService) CreateAccount(username, password string, ip netip.Addr) ([]byte, error) {
	exists, err := s.users.Exists(username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUsernameTaken
	}
	id, err := uuid.NewV7()
	if err != nil {
		return nil, err
	}
	if err := s.users.Save(User{ID: id, Username: username}); err != nil {
		return nil, err
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	hash := HashPassword(password, salt)
	if err := s.credentials.Save(Credential{UserID: id, Hash: hash, Salt: salt}); err != nil {
		return nil, err
	}
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		return nil, err
	}
	now := time.Now()
	if err := s.sessions.Save(Session{Token: token, UserID: id, IP: ip, CreatedAt: now, ExpiresAt: now.Add(sessionDuration)}); err != nil {
		return nil, err
	}
	return token, nil
}
