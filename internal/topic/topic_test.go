package topic

import (
	"reflect"
	"testing"
)

const sample = `---
slug: locks
title: Locks
description: Why a query is stuck, not slow
level: intermediate
reading_minutes: 10
order: 7
aliases: locking, lock
tags: transactions, concurrency
---

## What Postgres locks

Body text here.
`

func TestParseFrontMatter(t *testing.T) {
	tp, err := Parse([]byte(sample))
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if tp.Slug != "locks" {
		t.Errorf("Slug = %q, want %q", tp.Slug, "locks")
	}
	if tp.Title != "Locks" {
		t.Errorf("Title = %q, want %q", tp.Title, "Locks")
	}
	if tp.Description != "Why a query is stuck, not slow" {
		t.Errorf("Description = %q", tp.Description)
	}
	if tp.Level != "intermediate" {
		t.Errorf("Level = %q", tp.Level)
	}
	if tp.ReadingMinutes != 10 {
		t.Errorf("ReadingMinutes = %d, want 10", tp.ReadingMinutes)
	}
	if tp.Order != 7 {
		t.Errorf("Order = %d, want 7", tp.Order)
	}
	if want := []string{"locking", "lock"}; !reflect.DeepEqual(tp.Aliases, want) {
		t.Errorf("Aliases = %v, want %v", tp.Aliases, want)
	}
	if want := []string{"transactions", "concurrency"}; !reflect.DeepEqual(tp.Tags, want) {
		t.Errorf("Tags = %v, want %v", tp.Tags, want)
	}
	if tp.Content == "" || tp.Content[:2] != "##" {
		t.Errorf("Content should start at the markdown body, got %q", tp.Content)
	}
}

func TestParseErrors(t *testing.T) {
	cases := map[string]string{
		"no front matter":     "just markdown",
		"unterminated":        "---\nslug: x\n",
		"missing slug":        "---\ntitle: X\norder: 1\n---\nbody",
		"bad reading_minutes": "---\nslug: x\ntitle: X\norder: 1\nreading_minutes: ten\n---\nbody",
	}
	for name, in := range cases {
		if _, err := Parse([]byte(in)); err == nil {
			t.Errorf("%s: Parse succeeded, want error", name)
		}
	}
}

func TestNormalizeSlug(t *testing.T) {
	cases := map[string]string{
		"  Locks ":          "locks",
		"JSONB":             "jsonb",
		"Window Functions":  "window-functions",
		"vacuum_autovacuum": "vacuum-autovacuum",
	}
	for in, want := range cases {
		if got := NormalizeSlug(in); got != want {
			t.Errorf("NormalizeSlug(%q) = %q, want %q", in, got, want)
		}
	}
}

func testTopics() []Topic {
	return []Topic{
		{Slug: "indexes", Title: "Indexes", Order: 1, Aliases: []string{"index", "index-basics"}},
		{Slug: "jsonb", Title: "JSON & JSONB", Order: 3, Aliases: []string{"json"}},
		{Slug: "locks", Title: "Locks", Order: 7, Aliases: []string{"locking"}},
	}
}

func TestResolveBySlugAliasAndNormalization(t *testing.T) {
	ts := testTopics()
	for in, want := range map[string]string{
		"locks":     "locks",
		"locking":   "locks", // alias
		"JSON":      "jsonb", // alias, case-insensitive
		" Indexes ": "indexes",
	} {
		got, ok := Resolve(ts, in)
		if !ok || got.Slug != want {
			t.Errorf("Resolve(%q) = %v/%v, want %s", in, got.Slug, ok, want)
		}
	}
	if _, ok := Resolve(ts, "nope"); ok {
		t.Error("Resolve(nope) succeeded, want miss")
	}
}

func TestClosestSuggestions(t *testing.T) {
	ts := testTopics()
	got := Closest(ts, "lokcs", 2)
	if len(got) == 0 || got[0].Slug != "locks" {
		t.Errorf("Closest(lokcs) = %v, want locks first", got)
	}
}
