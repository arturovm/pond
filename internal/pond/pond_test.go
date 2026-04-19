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

var validFeed = pond.Feed{Body: []byte(`<rss version="2.0"><channel><title>T</title><link>https://example.com</link><description>D</description></channel></rss>`)}

func TestPond_Subscribe_SavesSourceInSources(t *testing.T) {
	fetcher := &mockFeedFetcher{feed: validFeed}
	sources := &mockSources{}
	p := pond.New(fetcher, sources)

	err := p.Subscribe("https://example.com/feed.rss")

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
	p := pond.New(fetcher, sources)

	err := p.Subscribe("https://example.com/feed.rss")

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
	p := pond.New(fetcher, sources)

	err := p.Subscribe("https://example.com/feed.rss")

	if !errors.Is(err, saveErr) {
		t.Errorf("expected error %v, got %v", saveErr, err)
	}
}

func TestPond_Subscribe_ReturnsFeedFetcherError(t *testing.T) {
	fetchErr := errors.New("fetch failed")
	fetcher := &mockFeedFetcher{err: fetchErr}
	p := pond.New(fetcher, &mockSources{})

	err := p.Subscribe("https://example.com/feed.rss")

	if !errors.Is(err, fetchErr) {
		t.Errorf("expected error %v, got %v", fetchErr, err)
	}
}

func TestPond_Subscribe_ReturnsParseError(t *testing.T) {
	fetcher := &mockFeedFetcher{feed: pond.Feed{Body: []byte("not xml")}}
	p := pond.New(fetcher, &mockSources{})

	err := p.Subscribe("https://example.com/feed.rss")

	if err == nil {
		t.Error("expected a parse error, got nil")
	}
}
