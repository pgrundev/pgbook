package client

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const indexJSON = `{"version":"0.1","topics":[
  {"slug":"indexes","title":"Indexes","description":"Why some queries are instant","level":"beginner","reading_minutes":8,"order":1},
  {"slug":"locks","title":"Locks","description":"Why a query is stuck, not slow","level":"intermediate","reading_minutes":10,"order":7,"aliases":["locking"]}
]}`

const locksJSON = `{"slug":"locks","title":"Locks","description":"Why a query is stuck, not slow","level":"intermediate","reading_minutes":10,"order":7,"content":"## What Postgres locks\n\nBody.\n"}`

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/topics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(indexJSON))
	})
	mux.HandleFunc("/api/topics/locks", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(locksJSON))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func setup(t *testing.T) (cacheDir string) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	return dir
}

func TestTopicsFetchesAndCaches(t *testing.T) {
	dir := setup(t)
	srv := testServer(t)
	c := New(srv.URL)

	ts, err := c.Topics()
	if err != nil {
		t.Fatalf("Topics: %v", err)
	}
	if len(ts) != 2 || ts[0].Slug != "indexes" || ts[1].Slug != "locks" {
		t.Fatalf("Topics = %+v", ts)
	}
	if _, err := os.Stat(filepath.Join(dir, "pgbook", "topics.json")); err != nil {
		t.Errorf("index not cached under $XDG_CACHE_HOME/pgbook: %v", err)
	}
}

func TestTopicsFallsBackToCacheWhenOffline(t *testing.T) {
	setup(t)
	srv := testServer(t)
	c := New(srv.URL)
	if _, err := c.Topics(); err != nil {
		t.Fatalf("warm-up fetch: %v", err)
	}
	srv.Close()

	ts, err := c.Topics()
	if err != nil {
		t.Fatalf("Topics offline with cache: %v", err)
	}
	if len(ts) != 2 {
		t.Errorf("cached Topics = %d entries, want 2", len(ts))
	}
}

func TestTopicsOfflineWithoutCacheExplains(t *testing.T) {
	setup(t)
	srv := testServer(t)
	srv.Close()
	c := New(srv.URL)
	_, err := c.Topics()
	if err == nil {
		t.Fatal("Topics offline without cache succeeded, want error")
	}
	if !strings.Contains(err.Error(), "pgbook.dev") && !strings.Contains(err.Error(), srv.URL) {
		t.Errorf("error should name the API it could not reach: %v", err)
	}
}

func TestTopicFetchesCachesAndFallsBack(t *testing.T) {
	dir := setup(t)
	srv := testServer(t)
	c := New(srv.URL)

	tp, err := c.Topic("locks")
	if err != nil {
		t.Fatalf("Topic: %v", err)
	}
	if tp.Slug != "locks" || !strings.Contains(tp.Content, "What Postgres locks") {
		t.Fatalf("Topic = %+v", tp)
	}
	if _, err := os.Stat(filepath.Join(dir, "pgbook", "topics", "locks.json")); err != nil {
		t.Errorf("topic not cached: %v", err)
	}

	srv.Close()
	tp, err = c.Topic("locks")
	if err != nil {
		t.Fatalf("Topic offline with cache: %v", err)
	}
	if !strings.Contains(tp.Content, "What Postgres locks") {
		t.Errorf("cached Topic content = %q", tp.Content)
	}
}

func TestTopicNotFound(t *testing.T) {
	setup(t)
	srv := testServer(t)
	c := New(srv.URL)
	_, err := c.Topic("nope")
	if err == nil {
		t.Fatal("Topic(nope) succeeded, want error")
	}
	if !strings.Contains(err.Error(), "nope") {
		t.Errorf("error should name the missing topic: %v", err)
	}
}
