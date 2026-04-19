package pond_test

import (
	"errors"
	"testing"

	"github.com/arturovm/pond/internal/pond"
)

type mockFeedFetcher struct {
	calledWith string
	err        error
}

func (m *mockFeedFetcher) Fetch(url string) (pond.Feed, error) {
	m.calledWith = url
	return pond.Feed{}, m.err
}

var _ pond.FeedFetcher = (*mockFeedFetcher)(nil)

func TestPond_Subscribe_CallsFeedFetcherWithURL(t *testing.T) {
	fetcher := &mockFeedFetcher{}
	p := pond.New(fetcher)

	err := p.Subscribe("https://example.com/feed.rss")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fetcher.calledWith != "https://example.com/feed.rss" {
		t.Errorf("expected Fetch to be called with %q, got %q", "https://example.com/feed.rss", fetcher.calledWith)
	}
}

func TestPond_Subscribe_ReturnsFeedFetcherError(t *testing.T) {
	fetchErr := errors.New("fetch failed")
	fetcher := &mockFeedFetcher{err: fetchErr}
	p := pond.New(fetcher)

	err := p.Subscribe("https://example.com/feed.rss")

	if !errors.Is(err, fetchErr) {
		t.Errorf("expected error %v, got %v", fetchErr, err)
	}
}
