package pond

import "errors"

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

// ErrUsernameTaken is returned when a username already exists.
var ErrUsernameTaken = errors.New("username taken")

// Users is the outgoing port for checking user existence.
type Users interface {
	Exists(username string) (bool, error)
}

// AccountCreator is the incoming port for creating an account.
type AccountCreator interface {
	CreateAccount(username, password string) error
}

// Pond is the application hexagon.
type Pond struct {
	fetcher       FeedFetcher
	sources       Sources
	subscriptions Subscriptions
	users         Users
}

func New(fetcher FeedFetcher, sources Sources, subscriptions Subscriptions, users Users) *Pond {
	return &Pond{fetcher: fetcher, sources: sources, subscriptions: subscriptions, users: users}
}

func (p *Pond) CreateAccount(username, password string) error {
	exists, err := p.users.Exists(username)
	if err != nil {
		return err
	}
	if exists {
		return ErrUsernameTaken
	}
	return nil
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
