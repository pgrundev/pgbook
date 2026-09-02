// Package client fetches book content from pgbook.dev and keeps an
// offline cache under $XDG_CACHE_HOME/pgbook.
package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/pgrundev/pgbook/internal/topic"
)

// Client talks to the pgbook.dev content API.
type Client struct {
	baseURL string
	http    *http.Client
}

// New returns a client for the given base URL (e.g. https://pgbook.dev).
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type index struct {
	Version string        `json:"version"`
	Topics  []topic.Topic `json:"topics"`
}

// Topics returns the topic index, from the API or the local cache.
func (c *Client) Topics() ([]topic.Topic, error) {
	data, fetchErr := c.get("/api/topics")
	if fetchErr == nil {
		writeCache("topics.json", data)
	} else {
		var cacheErr error
		data, cacheErr = readCache("topics.json")
		if cacheErr != nil {
			return nil, fmt.Errorf("cannot reach %s and no cached copy exists: %w", c.baseURL, fetchErr)
		}
	}
	var idx index
	if err := json.Unmarshal(data, &idx); err != nil {
		return nil, fmt.Errorf("bad topic index from %s: %w", c.baseURL, err)
	}
	return idx.Topics, nil
}

// Topic returns one topic with its full content, from the API or cache.
func (c *Client) Topic(slug string) (topic.Topic, error) {
	var t topic.Topic
	cacheName := filepath.Join("topics", slug+".json")
	data, fetchErr := c.get("/api/topics/" + slug)
	if fetchErr == nil {
		writeCache(cacheName, data)
	} else {
		var cacheErr error
		data, cacheErr = readCache(cacheName)
		if cacheErr != nil {
			return t, fmt.Errorf("cannot load topic %q: %w", slug, fetchErr)
		}
	}
	if err := json.Unmarshal(data, &t); err != nil {
		return t, fmt.Errorf("bad topic %q from %s: %w", slug, c.baseURL, err)
	}
	return t, nil
}

func (c *Client) get(path string) ([]byte, error) {
	resp, err := c.http.Get(c.baseURL + path)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%s%s: not found", c.baseURL, path)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s%s: HTTP %s", c.baseURL, path, resp.Status)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 10<<20))
}

func cacheDir() (string, error) {
	dir := os.Getenv("XDG_CACHE_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".cache")
	}
	return filepath.Join(dir, "pgbook"), nil
}

func writeCache(name string, data []byte) {
	dir, err := cacheDir()
	if err != nil {
		return // caching is best-effort
	}
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	_ = os.WriteFile(path, data, 0o644)
}

func readCache(name string) ([]byte, error) {
	dir, err := cacheDir()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(dir, name))
}
