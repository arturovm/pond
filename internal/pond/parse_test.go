package pond_test

import (
	"testing"

	"github.com/arturovm/pond/internal/pond"
)

const minimalRSSFeed = `<rss version="2.0">
  <channel>
    <title>My Feed</title>
    <link>https://example.com</link>
    <description>A feed about things.</description>
  </channel>
</rss>`

func TestParseMetadata_ReturnsErrorWhenChannelIsMissing(t *testing.T) {
	feed := pond.Feed{Body: []byte(`<rss version="2.0"></rss>`)}

	_, err := pond.ParseMetadata(feed)

	if err == nil {
		t.Error("expected an error when channel is missing, got nil")
	}
}

func TestParseMetadata_ReturnsErrorForMalformedXML(t *testing.T) {
	feed := pond.Feed{Body: []byte("not xml")}

	_, err := pond.ParseMetadata(feed)

	if err == nil {
		t.Error("expected an error for malformed XML, got nil")
	}
}

func TestParseMetadata_ExtractsDescription(t *testing.T) {
	feed := pond.Feed{Body: []byte(minimalRSSFeed)}

	meta, err := pond.ParseMetadata(feed)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Description != "A feed about things." {
		t.Errorf("expected description %q, got %q", "A feed about things.", meta.Description)
	}
}

func TestParseMetadata_ExtractsLink(t *testing.T) {
	feed := pond.Feed{Body: []byte(minimalRSSFeed)}

	meta, err := pond.ParseMetadata(feed)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Link != "https://example.com" {
		t.Errorf("expected link %q, got %q", "https://example.com", meta.Link)
	}
}

func TestParseMetadata_ExtractsTitle(t *testing.T) {
	feed := pond.Feed{Body: []byte(minimalRSSFeed)}

	meta, err := pond.ParseMetadata(feed)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.Title != "My Feed" {
		t.Errorf("expected title %q, got %q", "My Feed", meta.Title)
	}
}

func TestExtractSource_MapsAllMetadataFields(t *testing.T) {
	meta := pond.Metadata{Title: "My Feed", Link: "https://example.com", Description: "A feed about things."}

	source := pond.ExtractSource(meta)

	if source.Title != meta.Title {
		t.Errorf("expected title %q, got %q", meta.Title, source.Title)
	}
	if source.Link != meta.Link {
		t.Errorf("expected link %q, got %q", meta.Link, source.Link)
	}
	if source.Description != meta.Description {
		t.Errorf("expected description %q, got %q", meta.Description, source.Description)
	}
}
