// Package topic defines the lesson model shared by the CLI, the search
// index, and the site generator.
package topic

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// Topic is one lesson of the book. The same shape is served as JSON by
// pgbook.dev under /api/topics.
type Topic struct {
	Slug           string   `json:"slug"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Level          string   `json:"level"`
	ReadingMinutes int      `json:"reading_minutes"`
	Order          int      `json:"order"`
	Aliases        []string `json:"aliases,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	Content        string   `json:"content,omitempty"`
}

// Parse reads a topic source file: `---` front matter followed by markdown.
func Parse(src []byte) (Topic, error) {
	var t Topic
	text := strings.ReplaceAll(string(src), "\r\n", "\n")
	if !strings.HasPrefix(text, "---\n") {
		return t, fmt.Errorf("missing front matter (expected leading ---)")
	}
	rest := text[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return t, fmt.Errorf("unterminated front matter")
	}
	head := rest[:end]
	body := rest[end+len("\n---"):]
	body = strings.TrimPrefix(body, "\n")

	for _, line := range strings.Split(head, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, ":")
		if !ok {
			return t, fmt.Errorf("bad front matter line %q", line)
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch key {
		case "slug":
			t.Slug = val
		case "title":
			t.Title = val
		case "description":
			t.Description = val
		case "level":
			t.Level = val
		case "reading_minutes":
			n, err := strconv.Atoi(val)
			if err != nil {
				return t, fmt.Errorf("reading_minutes: %q is not a number", val)
			}
			t.ReadingMinutes = n
		case "order":
			n, err := strconv.Atoi(val)
			if err != nil {
				return t, fmt.Errorf("order: %q is not a number", val)
			}
			t.Order = n
		case "aliases":
			t.Aliases = splitList(val)
		case "tags":
			t.Tags = splitList(val)
		default:
			return t, fmt.Errorf("unknown front matter key %q", key)
		}
	}
	if t.Slug == "" {
		return t, fmt.Errorf("front matter is missing slug")
	}
	t.Content = strings.TrimSpace(body) + "\n"
	return t, nil
}

func splitList(v string) []string {
	var out []string
	for _, p := range strings.Split(v, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// NormalizeSlug lowercases and hyphenates user input for matching.
func NormalizeSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.Map(func(r rune) rune {
		switch r {
		case ' ', '_':
			return '-'
		}
		return r
	}, s)
	return s
}

// Resolve finds the topic named by user input via slug or alias.
func Resolve(ts []Topic, input string) (Topic, bool) {
	want := NormalizeSlug(input)
	for _, t := range ts {
		if NormalizeSlug(t.Slug) == want {
			return t, true
		}
	}
	for _, t := range ts {
		for _, a := range t.Aliases {
			if NormalizeSlug(a) == want {
				return t, true
			}
		}
	}
	return Topic{}, false
}

// Closest returns up to n topics nearest to the input, for suggestions.
func Closest(ts []Topic, input string, n int) []Topic {
	want := NormalizeSlug(input)
	type scored struct {
		t Topic
		d int
	}
	var all []scored
	for _, t := range ts {
		d := editDistance(want, NormalizeSlug(t.Slug))
		for _, a := range t.Aliases {
			if ad := editDistance(want, NormalizeSlug(a)); ad < d {
				d = ad
			}
		}
		all = append(all, scored{t, d})
	}
	sort.SliceStable(all, func(i, j int) bool { return all[i].d < all[j].d })
	var out []Topic
	for _, s := range all {
		if len(out) == n {
			break
		}
		out = append(out, s.t)
	}
	return out
}

func editDistance(a, b string) int {
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(min(cur[j-1]+1, prev[j]+1), prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
