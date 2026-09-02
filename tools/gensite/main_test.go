package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeTopic(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestGenerate(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()
	writeTopic(t, src, "02-locks.md",
		"---\nslug: locks\ntitle: Locks\ndescription: Why a query is stuck\nlevel: intermediate\nreading_minutes: 10\norder: 2\naliases: locking\n---\n\n## Body\n\nLock text.\n")
	writeTopic(t, src, "01-indexes.md",
		"---\nslug: indexes\ntitle: Indexes\ndescription: Why some queries are instant\nlevel: beginner\nreading_minutes: 8\norder: 1\n---\n\n## Body\n\nIndex text.\n")

	if err := generate(src, out); err != nil {
		t.Fatalf("generate: %v", err)
	}

	// Index: ordered, no content, versioned. Written as topics.json —
	// a static host rewrites /api/topics to it (see site/_redirects) —
	// because /api/topics/<slug> files occupy the "topics" directory.
	data, err := os.ReadFile(filepath.Join(out, "api", "topics.json"))
	if err != nil {
		t.Fatalf("index not written: %v", err)
	}
	var idx struct {
		Version string `json:"version"`
		Topics  []struct {
			Slug    string `json:"slug"`
			Order   int    `json:"order"`
			Content string `json:"content"`
		} `json:"topics"`
	}
	if err := json.Unmarshal(data, &idx); err != nil {
		t.Fatalf("index is not JSON: %v", err)
	}
	if idx.Version == "" {
		t.Error("index has no version")
	}
	if len(idx.Topics) != 2 || idx.Topics[0].Slug != "indexes" || idx.Topics[1].Slug != "locks" {
		t.Errorf("index topics wrong: %+v", idx.Topics)
	}
	for _, tp := range idx.Topics {
		if tp.Content != "" {
			t.Errorf("index should not embed content, %s has it", tp.Slug)
		}
	}

	// Per-topic file: full content.
	data, err = os.ReadFile(filepath.Join(out, "api", "topics", "locks"))
	if err != nil {
		t.Fatalf("topic file not written: %v", err)
	}
	var tp struct {
		Slug    string `json:"slug"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(data, &tp); err != nil {
		t.Fatalf("topic file is not JSON: %v", err)
	}
	if tp.Slug != "locks" || tp.Content == "" {
		t.Errorf("topic file wrong: %+v", tp)
	}
}

func TestGenerateRejectsDuplicateSlugs(t *testing.T) {
	src := t.TempDir()
	out := t.TempDir()
	writeTopic(t, src, "a.md", "---\nslug: locks\ntitle: A\norder: 1\n---\nbody")
	writeTopic(t, src, "b.md", "---\nslug: locks\ntitle: B\norder: 2\n---\nbody")
	if err := generate(src, out); err == nil {
		t.Error("generate accepted duplicate slugs")
	}
}
