package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// BookMeta describes the downloadable edition, served at /api/book.
type BookMeta struct {
	Version     string `json:"version"`
	Topics      int    `json:"topics"`
	Pages       int    `json:"pages"`
	Filename    string `json:"filename"`
	DownloadURL string `json:"download_url"`
	SHA256      string `json:"sha256"`
}

// Book fetches metadata about the current PDF edition.
func (c *Client) Book() (BookMeta, error) {
	var m BookMeta
	data, err := c.get("/api/book")
	if err != nil {
		return m, err
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return m, fmt.Errorf("bad book metadata from %s: %w", c.baseURL, err)
	}
	return m, nil
}

// Download starts an HTTP GET for a large file, with a generous timeout.
// The caller must close the response body.
func (c *Client) Download(url string) (*http.Response, error) {
	dl := &http.Client{Timeout: 5 * time.Minute}
	return dl.Get(url)
}
