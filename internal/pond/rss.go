package pond

import (
	"encoding/xml"
	"fmt"
)

type rssChannel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
}

type rssFeed struct {
	Channel *rssChannel `xml:"channel"`
}

// ParseMetadata extracts channel-level Metadata from a fetched Feed.
func ParseMetadata(feed Feed) (Metadata, error) {
	var rss rssFeed
	if err := xml.Unmarshal(feed.Body, &rss); err != nil {
		return Metadata{}, err
	}
	if rss.Channel == nil {
		return Metadata{}, fmt.Errorf("feed has no channel element")
	}
	return Metadata{
		Title:       rss.Channel.Title,
		Link:        rss.Channel.Link,
		Description: rss.Channel.Description,
	}, nil
}
