package fetcher

import (
	"fmt"
	"io"
	"net/http"

	"github.com/arturovm/pond/internal/pond"
)

type HTTPFetcher struct {
	client *http.Client
}

func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{client: &http.Client{}}
}

func (f *HTTPFetcher) Fetch(url string) (pond.Feed, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return pond.Feed{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return pond.Feed{}, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return pond.Feed{}, err
	}

	return pond.Feed{Body: body}, nil
}
