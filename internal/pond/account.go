package pond

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/netip"
	"time"

	"github.com/google/uuid"
)

// User represents a registered user.
type User struct {
	ID       string
	Username string
}

// ErrUsernameTaken is returned when a username already exists.
var ErrUsernameTaken = errors.New("username taken")

// Credential holds the hashed password material for a user.
type Credential struct {
	UserID string
	Hash   []byte
	Salt   []byte
}

const sessionDuration = 30 * 24 * time.Hour

// Session represents an authenticated session.
type Session struct {
	Token     string
	UserID    string
	IP        netip.Addr
	CreatedAt time.Time
	ExpiresAt time.Time
}

// Users is the outgoing port for user persistence.
type Users interface {
	Exists(username string) (bool, error)
	Save(User) error
}

// Credentials is the outgoing port for persisting credentials.
type Credentials interface {
	Save(Credential) error
}

// Sessions is the outgoing port for persisting sessions.
type Sessions interface {
	Save(Session) error
}

// AccountCreator is the incoming port for creating an account.
type AccountCreator interface {
	CreateAccount(username, password string, ip netip.Addr) (string, error)
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

func (s *AccountService) CreateAccount(username, password string, ip netip.Addr) (string, error) {
	exists, err := s.users.Exists(username)
	if err != nil {
		return "", err
	}
	if exists {
		return "", ErrUsernameTaken
	}
	id, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	if err := s.users.Save(User{ID: id.String(), Username: username}); err != nil {
		return "", err
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := HashPassword(password, salt)
	if err := s.credentials.Save(Credential{UserID: id.String(), Hash: hash, Salt: salt}); err != nil {
		return "", err
	}
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)
	now := time.Now()
	if err := s.sessions.Save(Session{Token: token, UserID: id.String(), IP: ip, CreatedAt: now, ExpiresAt: now.Add(sessionDuration)}); err != nil {
		return "", err
	}
	return token, nil
}
