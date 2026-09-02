package progress

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pgrundev/pgbook/internal/topic"
)

func withStateDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", dir)
	return dir
}

func TestSaveAndLoadLastRead(t *testing.T) {
	dir := withStateDir(t)
	if err := SaveLastRead("locks"); err != nil {
		t.Fatalf("SaveLastRead: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "pgbook", "progress.json")); err != nil {
		t.Fatalf("progress file not under $XDG_STATE_HOME/pgbook: %v", err)
	}
	got, err := LastRead()
	if err != nil {
		t.Fatalf("LastRead: %v", err)
	}
	if got != "locks" {
		t.Errorf("LastRead = %q, want %q", got, "locks")
	}
}

func TestLastReadEmptyWhenNoState(t *testing.T) {
	withStateDir(t)
	got, err := LastRead()
	if err != nil {
		t.Fatalf("LastRead: %v", err)
	}
	if got != "" {
		t.Errorf("LastRead = %q, want empty", got)
	}
}

func topics() []topic.Topic {
	return []topic.Topic{
		{Slug: "indexes", Order: 1},
		{Slug: "transactions", Order: 2},
		{Slug: "locks", Order: 3},
	}
}

func TestNextTopic(t *testing.T) {
	ts := topics()

	// No progress: first topic.
	got, done := Next(ts, "")
	if done || got.Slug != "indexes" {
		t.Errorf("Next(\"\") = %v/%v, want indexes", got.Slug, done)
	}

	// Mid-book: the following topic.
	got, done = Next(ts, "indexes")
	if done || got.Slug != "transactions" {
		t.Errorf("Next(indexes) = %v/%v, want transactions", got.Slug, done)
	}

	// Last topic read: done.
	if _, done = Next(ts, "locks"); !done {
		t.Error("Next(locks) done = false, want true")
	}

	// Unknown saved slug (topic removed upstream): start over.
	got, done = Next(ts, "gone")
	if done || got.Slug != "indexes" {
		t.Errorf("Next(gone) = %v/%v, want indexes", got.Slug, done)
	}
}
