package pond

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/netip"
	"time"

	"github.com/google/uuid"
)

// Feed represents raw feed content fetched from a URL.
type Feed struct {
	Body []byte
}

// Metadata holds feed channel-level information extracted from a Feed.
type Metadata struct {
	Title       string
	Link        string
	Description string
}

// Source represents a feed source extracted from Metadata.
type Source struct {
	Title       string
	Link        string
	Description string
}

// Subscription represents a user's subscription to a Source.
type Subscription struct {
	UserID string
	Source Source
}

// ExtractSource builds a Source from feed Metadata.
func ExtractSource(meta Metadata) Source {
	return Source{
		Title:       meta.Title,
		Link:        meta.Link,
		Description: meta.Description,
	}
}

// FeedFetcher is the outgoing port for fetching feeds.
type FeedFetcher interface {
	Fetch(url string) (Feed, error)
}

// Sources is the outgoing port for persisting feed sources.
type Sources interface {
	Save(Source) error
}

// Subscriptions is the outgoing port for persisting subscriptions.
type Subscriptions interface {
	Save(Subscription) error
}

// Subscriber is the incoming port for subscribing to a feed.
type Subscriber interface {
	Subscribe(userID, feedURL string) error
}

// User represents a registered user.
type User struct {
	ID       string
	Username string
}

// ErrUsernameTaken is returned when a username already exists.
var ErrUsernameTaken = errors.New("username taken")

// Users is the outgoing port for user persistence.
type Users interface {
	Exists(username string) (bool, error)
	Save(User) error
}

// Credential holds the hashed password material for a user.
type Credential struct {
	UserID string
	Hash   []byte
	Salt   []byte
}

// Credentials is the outgoing port for persisting credentials.
type Credentials interface {
	Save(Credential) error
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

// Sessions is the outgoing port for persisting sessions.
type Sessions interface {
	Save(Session) error
}

// AccountCreator is the incoming port for creating an account.
type AccountCreator interface {
	CreateAccount(username, password string, ip netip.Addr) (string, error)
}

// Pond is the application hexagon.
type Pond struct {
	fetcher       FeedFetcher
	sources       Sources
	subscriptions Subscriptions
	users         Users
	credentials   Credentials
	sessions      Sessions
}

func New(fetcher FeedFetcher, sources Sources, subscriptions Subscriptions, users Users, credentials Credentials, sessions Sessions) *Pond {
	return &Pond{fetcher: fetcher, sources: sources, subscriptions: subscriptions, users: users, credentials: credentials, sessions: sessions}
}

func (p *Pond) CreateAccount(username, password string, ip netip.Addr) (string, error) {
	exists, err := p.users.Exists(username)
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
	if err := p.users.Save(User{ID: id.String(), Username: username}); err != nil {
		return "", err
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	hash := HashPassword(password, salt)
	if err := p.credentials.Save(Credential{UserID: id.String(), Hash: hash, Salt: salt}); err != nil {
		return "", err
	}
	tokenBytes := make([]byte, 16)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(tokenBytes)
	now := time.Now()
	if err := p.sessions.Save(Session{Token: token, UserID: id.String(), IP: ip, CreatedAt: now, ExpiresAt: now.Add(sessionDuration)}); err != nil {
		return "", err
	}
	return token, nil
}

func (p *Pond) Subscribe(userID, feedURL string) error {
	feed, err := p.fetcher.Fetch(feedURL)
	if err != nil {
		return err
	}
	meta, err := ParseMetadata(feed)
	if err != nil {
		return err
	}
	source := ExtractSource(meta)
	if err := p.sources.Save(source); err != nil {
		return err
	}
	return p.subscriptions.Save(Subscription{UserID: userID, Source: source})
}
