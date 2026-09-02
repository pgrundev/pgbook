package search

import (
	"strings"
	"testing"

	"github.com/pgrundev/pgbook/internal/topic"
)

func corpus() []topic.Topic {
	return []topic.Topic{
		{Slug: "indexes", Title: "Indexes", Description: "Why some queries are instant", Order: 1,
			Content: "## B-tree basics\n\nMost indexes are B-trees.\n"},
		{Slug: "jsonb", Title: "JSON & JSONB", Description: "Semi-structured data, indexed", Order: 3,
			Tags:    []string{"json"},
			Content: "## Indexing JSONB with GIN\n\nGIN indexes make containment fast.\n"},
		{Slug: "locks", Title: "Locks", Description: "Why a query is stuck, not slow", Order: 7,
			Tags:    []string{"concurrency"},
			Content: "## Deadlocks\n\nTwo transactions waiting on each other.\n"},
	}
}

func TestSearchMatchesTitleDescriptionHeadingsTagsAndText(t *testing.T) {
	ts := corpus()

	rs := Search(ts, "indexes")
	if len(rs) != 2 {
		t.Fatalf("Search(indexes) = %d results, want 2 (indexes, jsonb): %+v", len(rs), rs)
	}
	if rs[0].Topic.Slug != "indexes" {
		t.Errorf("title match should rank first, got %s", rs[0].Topic.Slug)
	}

	// Heading match surfaces the heading as the snippet.
	rs = Search(ts, "gin")
	if len(rs) != 1 || rs[0].Topic.Slug != "jsonb" {
		t.Fatalf("Search(gin) = %+v", rs)
	}
	if !strings.Contains(rs[0].Snippet, "Indexing JSONB with GIN") {
		t.Errorf("snippet = %q, want the matching heading", rs[0].Snippet)
	}

	// Tag match.
	if rs = Search(ts, "concurrency"); len(rs) != 1 || rs[0].Topic.Slug != "locks" {
		t.Errorf("Search(concurrency) = %+v, want locks", rs)
	}

	// Body text match.
	if rs = Search(ts, "waiting on each other"); len(rs) != 1 || rs[0].Topic.Slug != "locks" {
		t.Errorf("Search(body text) = %+v, want locks", rs)
	}

	// Case-insensitive.
	if rs = Search(ts, "DEADLOCKS"); len(rs) != 1 {
		t.Errorf("Search should be case-insensitive, got %+v", rs)
	}
}

func TestSearchNoResults(t *testing.T) {
	if rs := Search(corpus(), "zzzz"); len(rs) != 0 {
		t.Errorf("Search(zzzz) = %+v, want none", rs)
	}
}

func TestHighlight(t *testing.T) {
	got := Highlight("Indexing JSONB with GIN", "gin", true)
	if !strings.Contains(got, "\x1b[1mGIN\x1b[0m") {
		t.Errorf("Highlight color = %q, want GIN wrapped in bold", got)
	}
	if got := Highlight("Indexing JSONB with GIN", "gin", false); got != "Indexing JSONB with GIN" {
		t.Errorf("Highlight plain = %q, want unchanged", got)
	}
}
