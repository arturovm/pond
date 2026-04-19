package pond

// Subscriber is the incoming port for subscribing to a feed.
type Subscriber interface {
	Subscribe(feedURL string) error
}

// Pond is the application hexagon.
type Pond struct{}
