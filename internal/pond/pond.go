package pond

// Feed represents raw feed content fetched from a URL.
type Feed struct {
	Body []byte
}

// FeedFetcher is the outgoing port for fetching feeds.
type FeedFetcher interface {
	Fetch(url string) (Feed, error)
}

// Subscriber is the incoming port for subscribing to a feed.
type Subscriber interface {
	Subscribe(feedURL string) error
}

// Pond is the application hexagon.
type Pond struct {
	fetcher FeedFetcher
}

func New(fetcher FeedFetcher) *Pond {
	return &Pond{fetcher: fetcher}
}

func (p *Pond) Subscribe(feedURL string) error {
	_, err := p.fetcher.Fetch(feedURL)
	return err
}
