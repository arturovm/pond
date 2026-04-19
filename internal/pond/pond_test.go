package pond_test

import (
	"errors"
	"testing"

	"github.com/arturovm/pond/internal/pond"
)

type mockFeedFetcher struct {
	calledWith string
	feed       pond.Feed
	err        error
}

func (m *mockFeedFetcher) Fetch(url string) (pond.Feed, error) {
	m.calledWith = url
	return m.feed, m.err
}

var _ pond.FeedFetcher = (*mockFeedFetcher)(nil)

type mockSources struct {
	saved pond.Source
	err   error
}

func (m *mockSources) Save(s pond.Source) error {
	m.saved = s
	return m.err
}

var _ pond.Sources = (*mockSources)(nil)

type mockSubscriptions struct {
	saved pond.Subscription
	err   error
}

func (m *mockSubscriptions) Save(s pond.Subscription) error {
	m.saved = s
	return m.err
}

var _ pond.Subscriptions = (*mockSubscriptions)(nil)

var validFeed = pond.Feed{Body: []byte(`<rss version="2.0"><channel><title>T</title><link>https://example.com</link><description>D</description></channel></rss>`)}

func TestPond_Subscribe_SavesSubscriptionInSubscriptions(t *testing.T) {
	fetcher := &mockFeedFetcher{feed: validFeed}
	sources := &mockSources{}
	subscriptions := &mockSubscriptions{}
	p := pond.New(fetcher, sources, subscriptions, nil)

	err := p.Subscribe("user1", "https://example.com/feed.rss")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := pond.Subscription{
		UserID: "user1",
		Source: pond.Source{Title: "T", Link: "https://example.com", Description: "D"},
	}
	if subscriptions.saved != want {
		t.Errorf("expected subscription %+v, got %+v", want, subscriptions.saved)
	}
}

func TestPond_Subscribe_SavesSourceInSources(t *testing.T) {
	fetcher := &mockFeedFetcher{feed: validFeed}
	sources := &mockSources{}
	p := pond.New(fetcher, sources, &mockSubscriptions{}, nil)

	err := p.Subscribe("user1", "https://example.com/feed.rss")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := pond.Source{Title: "T", Link: "https://example.com", Description: "D"}
	if sources.saved != want {
		t.Errorf("expected source %+v, got %+v", want, sources.saved)
	}
}

func TestPond_Subscribe_CallsFeedFetcherWithURL(t *testing.T) {
	fetcher := &mockFeedFetcher{feed: validFeed}
	sources := &mockSources{}
	p := pond.New(fetcher, sources, &mockSubscriptions{}, nil)

	err := p.Subscribe("user1", "https://example.com/feed.rss")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetcher.calledWith != "https://example.com/feed.rss" {
		t.Errorf("expected Fetch to be called with %q, got %q", "https://example.com/feed.rss", fetcher.calledWith)
	}
}

func TestPond_Subscribe_ReturnsSourcesError(t *testing.T) {
	saveErr := errors.New("save failed")
	fetcher := &mockFeedFetcher{feed: validFeed}
	sources := &mockSources{err: saveErr}
	p := pond.New(fetcher, sources, &mockSubscriptions{}, nil)

	err := p.Subscribe("user1", "https://example.com/feed.rss")

	if !errors.Is(err, saveErr) {
		t.Errorf("expected error %v, got %v", saveErr, err)
	}
}

func TestPond_Subscribe_ReturnsFeedFetcherError(t *testing.T) {
	fetchErr := errors.New("fetch failed")
	fetcher := &mockFeedFetcher{err: fetchErr}
	p := pond.New(fetcher, &mockSources{}, &mockSubscriptions{}, nil)

	err := p.Subscribe("user1", "https://example.com/feed.rss")

	if !errors.Is(err, fetchErr) {
		t.Errorf("expected error %v, got %v", fetchErr, err)
	}
}

func TestPond_Subscribe_ReturnsSubscriptionsError(t *testing.T) {
	saveErr := errors.New("subscriptions save failed")
	fetcher := &mockFeedFetcher{feed: validFeed}
	subscriptions := &mockSubscriptions{err: saveErr}
	p := pond.New(fetcher, &mockSources{}, subscriptions, nil)

	err := p.Subscribe("user1", "https://example.com/feed.rss")

	if !errors.Is(err, saveErr) {
		t.Errorf("expected error %v, got %v", saveErr, err)
	}
}

func TestPond_Subscribe_ReturnsParseError(t *testing.T) {
	fetcher := &mockFeedFetcher{feed: pond.Feed{Body: []byte("not xml")}}
	p := pond.New(fetcher, &mockSources{}, &mockSubscriptions{}, nil)

	err := p.Subscribe("user1", "https://example.com/feed.rss")

	if err == nil {
		t.Error("expected a parse error, got nil")
	}
}

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

func TestPond_CreateAccount_ExistingUsername_ReturnsErrUsernameTaken(t *testing.T) {
	users := &mockUsers{exists: true}
	p := pond.New(nil, nil, nil, users)

	err := p.CreateAccount("alice", "secret")

	if !errors.Is(err, pond.ErrUsernameTaken) {
		t.Errorf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestPond_CreateAccount_UsersPortError_ReturnsError(t *testing.T) {
	portErr := errors.New("db failed")
	users := &mockUsers{err: portErr}
	p := pond.New(nil, nil, nil, users)

	err := p.CreateAccount("alice", "secret")

	if !errors.Is(err, portErr) {
		t.Errorf("expected %v, got %v", portErr, err)
	}
}

func TestPond_CreateAccount_NewUsername_SavesUserInUsers(t *testing.T) {
	users := &mockUsers{exists: false}
	p := pond.New(nil, nil, nil, users)

	err := p.CreateAccount("alice", "secret")

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

func TestPond_CreateAccount_SaveError_ReturnsError(t *testing.T) {
	saveErr := errors.New("save failed")
	users := &mockUsers{exists: false, saveErr: saveErr}
	p := pond.New(nil, nil, nil, users)

	err := p.CreateAccount("alice", "secret")

	if !errors.Is(err, saveErr) {
		t.Errorf("expected %v, got %v", saveErr, err)
	}
}

func TestPond_CreateAccount_NewUsername_DoesNotReturnError(t *testing.T) {
	users := &mockUsers{exists: false}
	p := pond.New(nil, nil, nil, users)

	err := p.CreateAccount("alice", "secret")

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}
