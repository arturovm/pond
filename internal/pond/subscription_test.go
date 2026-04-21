package pond_test

import (
	"errors"
	"testing"

	"github.com/arturovm/pond/internal/pond"
	"github.com/google/uuid"
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

func TestSubscriptionService_Subscribe_SavesSubscriptionInSubscriptions(t *testing.T) {
	fetcher := &mockFeedFetcher{feed: validFeed}
	sources := &mockSources{}
	subscriptions := &mockSubscriptions{}
	s := pond.NewSubscriptionService(fetcher, sources, subscriptions)

	err := s.Subscribe(uuid.UUID{1}, "https://example.com/feed.rss")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := pond.Subscription{
		UserID: uuid.UUID{1},
		Source: pond.Source{Title: "T", Link: "https://example.com", Description: "D"},
	}
	if subscriptions.saved != want {
		t.Errorf("expected subscription %+v, got %+v", want, subscriptions.saved)
	}
}

func TestSubscriptionService_Subscribe_SavesSourceInSources(t *testing.T) {
	fetcher := &mockFeedFetcher{feed: validFeed}
	sources := &mockSources{}
	s := pond.NewSubscriptionService(fetcher, sources, &mockSubscriptions{})

	err := s.Subscribe(uuid.UUID{1}, "https://example.com/feed.rss")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := pond.Source{Title: "T", Link: "https://example.com", Description: "D"}
	if sources.saved != want {
		t.Errorf("expected source %+v, got %+v", want, sources.saved)
	}
}

func TestSubscriptionService_Subscribe_CallsFeedFetcherWithURL(t *testing.T) {
	fetcher := &mockFeedFetcher{feed: validFeed}
	s := pond.NewSubscriptionService(fetcher, &mockSources{}, &mockSubscriptions{})

	err := s.Subscribe(uuid.UUID{1}, "https://example.com/feed.rss")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetcher.calledWith != "https://example.com/feed.rss" {
		t.Errorf("expected Fetch to be called with %q, got %q", "https://example.com/feed.rss", fetcher.calledWith)
	}
}

func TestSubscriptionService_Subscribe_ReturnsSourcesError(t *testing.T) {
	saveErr := errors.New("save failed")
	fetcher := &mockFeedFetcher{feed: validFeed}
	sources := &mockSources{err: saveErr}
	s := pond.NewSubscriptionService(fetcher, sources, &mockSubscriptions{})

	err := s.Subscribe(uuid.UUID{1}, "https://example.com/feed.rss")

	if !errors.Is(err, saveErr) {
		t.Errorf("expected error %v, got %v", saveErr, err)
	}
}

func TestSubscriptionService_Subscribe_ReturnsFeedFetcherError(t *testing.T) {
	fetchErr := errors.New("fetch failed")
	fetcher := &mockFeedFetcher{err: fetchErr}
	s := pond.NewSubscriptionService(fetcher, &mockSources{}, &mockSubscriptions{})

	err := s.Subscribe(uuid.UUID{1}, "https://example.com/feed.rss")

	if !errors.Is(err, fetchErr) {
		t.Errorf("expected error %v, got %v", fetchErr, err)
	}
}

func TestSubscriptionService_Subscribe_ReturnsSubscriptionsError(t *testing.T) {
	saveErr := errors.New("subscriptions save failed")
	fetcher := &mockFeedFetcher{feed: validFeed}
	subscriptions := &mockSubscriptions{err: saveErr}
	s := pond.NewSubscriptionService(fetcher, &mockSources{}, subscriptions)

	err := s.Subscribe(uuid.UUID{1}, "https://example.com/feed.rss")

	if !errors.Is(err, saveErr) {
		t.Errorf("expected error %v, got %v", saveErr, err)
	}
}

func TestSubscriptionService_Subscribe_ReturnsParseError(t *testing.T) {
	fetcher := &mockFeedFetcher{feed: pond.Feed{Body: []byte("not xml")}}
	s := pond.NewSubscriptionService(fetcher, &mockSources{}, &mockSubscriptions{})

	err := s.Subscribe(uuid.UUID{1}, "https://example.com/feed.rss")

	if err == nil {
		t.Error("expected a parse error, got nil")
	}
}
