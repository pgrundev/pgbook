// Package search finds topics by title, description, headings, tags,
// and lesson text.
package search

import (
	"sort"
	"strings"

	"github.com/pgrundev/pgbook/internal/topic"
)

// Result is one search hit.
type Result struct {
	Topic   topic.Topic
	Snippet string
}

// Rank weights: earlier fields rank higher.
const (
	rankTitle = iota
	rankDescription
	rankHeading
	rankTag
	rankBody
)

// Search returns matching topics, best match first.
func Search(ts []topic.Topic, query string) []Result {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	type hit struct {
		r    Result
		rank int
	}
	var hits []hit
	for _, t := range ts {
		h, ok := match(t, q)
		if ok {
			hits = append(hits, h)
		}
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].rank != hits[j].rank {
			return hits[i].rank < hits[j].rank
		}
		return hits[i].r.Topic.Order < hits[j].r.Topic.Order
	})
	out := make([]Result, len(hits))
	for i, h := range hits {
		out[i] = h.r
	}
	return out
}

func match(t topic.Topic, q string) (h struct {
	r    Result
	rank int
}, ok bool) {
	h.r.Topic = t
	if strings.Contains(strings.ToLower(t.Title), q) {
		h.r.Snippet = t.Description
		h.rank = rankTitle
		return h, true
	}
	if strings.Contains(strings.ToLower(t.Description), q) {
		h.r.Snippet = t.Description
		h.rank = rankDescription
		return h, true
	}
	var bodyLine string
	for _, line := range strings.Split(t.Content, "\n") {
		if !strings.Contains(strings.ToLower(line), q) {
			continue
		}
		clean := strings.TrimSpace(strings.TrimLeft(line, "# "))
		if strings.HasPrefix(line, "#") {
			h.r.Snippet = clean
			h.rank = rankHeading
			return h, true
		}
		if bodyLine == "" {
			bodyLine = clean
		}
	}
	for _, tag := range t.Tags {
		if strings.Contains(strings.ToLower(tag), q) {
			h.r.Snippet = t.Description
			h.rank = rankTag
			return h, true
		}
	}
	if bodyLine != "" {
		h.r.Snippet = bodyLine
		h.rank = rankBody
		return h, true
	}
	return h, false
}

// Highlight wraps occurrences of query in s with bold when color is on.
func Highlight(s, query string, color bool) string {
	if !color || query == "" {
		return s
	}
	lower := strings.ToLower(s)
	q := strings.ToLower(query)
	var b strings.Builder
	for {
		i := strings.Index(lower, q)
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		b.WriteString(s[:i])
		b.WriteString("\x1b[1m" + s[i:i+len(q)] + "\x1b[0m")
		s = s[i+len(q):]
		lower = lower[i+len(q):]
	}
}
