// Package commands implements the pgbook subcommands.
package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pgrundev/pgbook/internal/client"
	"github.com/pgrundev/pgbook/internal/progress"
	"github.com/pgrundev/pgbook/internal/render"
	"github.com/pgrundev/pgbook/internal/search"
	"github.com/pgrundev/pgbook/internal/topic"
)

// App wires the subcommands to their output and content source.
type App struct {
	Client *client.Client
	Out    io.Writer
	Color  bool
}

func (a *App) bold(s string) string {
	if !a.Color {
		return s
	}
	return "\x1b[1m" + s + "\x1b[0m"
}

func (a *App) dim(s string) string {
	if !a.Color {
		return s
	}
	return "\x1b[2m" + s + "\x1b[0m"
}

func (a *App) topics() ([]topic.Topic, error) {
	ts, err := a.Client.Topics()
	if err != nil {
		return nil, err
	}
	sort.SliceStable(ts, func(i, j int) bool { return ts[i].Order < ts[j].Order })
	return ts, nil
}

// List prints the table of topics.
func (a *App) List() error {
	ts, err := a.topics()
	if err != nil {
		return err
	}
	fmt.Fprintf(a.Out, "%s\n\n", a.bold("POSTGRES BOOK"))
	width := 0
	for _, t := range ts {
		if len(t.Title) > width {
			width = len(t.Title)
		}
	}
	for _, t := range ts {
		fmt.Fprintf(a.Out, "%02d  %-*s  %s\n", t.Order, width, t.Title, a.dim(t.Level))
	}
	example := "locks"
	if _, ok := topic.Resolve(ts, example); !ok && len(ts) > 0 {
		example = ts[0].Slug
	}
	fmt.Fprintf(a.Out, "\nRead a topic: pgbook read %s\n", example)
	return nil
}

// Read renders one topic and records progress.
func (a *App) Read(name string) error {
	ts, err := a.topics()
	if err != nil {
		return err
	}
	meta, ok := topic.Resolve(ts, name)
	if !ok {
		msg := fmt.Sprintf("unknown topic %q", name)
		if close := topic.Closest(ts, name, 1); len(close) > 0 {
			msg += fmt.Sprintf(" — did you mean %q?", close[0].Slug)
		}
		return fmt.Errorf("%s (run `pgbook list` to see all topics)", msg)
	}
	t, err := a.Client.Topic(meta.Slug)
	if err != nil {
		return err
	}
	fmt.Fprintf(a.Out, "%s\n", a.bold(strings.ToUpper(t.Title)))
	if t.Description != "" {
		fmt.Fprintf(a.Out, "%s\n", t.Description)
	}
	if t.Level != "" || t.ReadingMinutes > 0 {
		fmt.Fprintf(a.Out, "\n%s\n", a.dim(fmt.Sprintf("%s · %d min", titleCase(t.Level), t.ReadingMinutes)))
	}
	fmt.Fprintf(a.Out, "\n%s", render.Render(t.Content, render.Options{Color: a.Color}))
	return progress.SaveLastRead(t.Slug)
}

// Search prints topics matching the query.
func (a *App) Search(query string) error {
	ts, err := a.topics()
	if err != nil {
		return err
	}
	// Pull lesson text (cached after first fetch) so body matches work.
	for i, t := range ts {
		if t.Content != "" {
			continue
		}
		if full, err := a.Client.Topic(t.Slug); err == nil {
			ts[i].Content = full.Content
		}
	}
	rs := search.Search(ts, query)
	if len(rs) == 0 {
		fmt.Fprintf(a.Out, "No results for %q.\n", query)
		if close := topic.Closest(ts, query, 3); len(close) > 0 {
			fmt.Fprintf(a.Out, "\nClosest topics:\n")
			for _, t := range close {
				fmt.Fprintf(a.Out, "  %s — pgbook read %s\n", t.Title, t.Slug)
			}
		}
		return nil
	}
	plural := "results"
	if len(rs) == 1 {
		plural = "result"
	}
	fmt.Fprintf(a.Out, "%d %s\n\n", len(rs), plural)
	for _, r := range rs {
		fmt.Fprintf(a.Out, "%s\n", a.bold(search.Highlight(r.Topic.Title, query, a.Color)))
		if r.Snippet != "" {
			fmt.Fprintf(a.Out, "%s\n", search.Highlight(r.Snippet, query, a.Color))
		}
		fmt.Fprintf(a.Out, "%s\n\n", a.dim("pgbook read "+r.Topic.Slug))
	}
	return nil
}

// Next opens the topic after the last one read.
func (a *App) Next() error {
	ts, err := a.topics()
	if err != nil {
		return err
	}
	last, err := progress.LastRead()
	if err != nil {
		return err
	}
	next, done := progress.Next(ts, last)
	if done {
		fmt.Fprintf(a.Out, "You've reached the end of the book — nice work.\nRun `pgbook list` to browse all topics again.\n")
		return nil
	}
	return a.Read(next.Slug)
}

// PDF downloads the book PDF to dest ("" means the published filename in
// the current directory). It never silently overwrites: pass force.
func (a *App) PDF(dest string, force bool) error {
	meta, err := a.Client.Book()
	if err != nil {
		return fmt.Errorf("cannot fetch book metadata: %w", err)
	}
	if dest == "" {
		dest = meta.Filename
		if dest == "" {
			dest = "postgres-book.pdf"
		}
	}
	if _, err := os.Stat(dest); err == nil && !force {
		return fmt.Errorf("%s already exists — pass --force to overwrite it", dest)
	}

	fmt.Fprintf(a.Out, "Downloading Postgres Book…\n\n")

	resp, err := a.Client.Download(meta.DownloadURL)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: %s returned HTTP %s", meta.DownloadURL, resp.Status)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "application/pdf") {
		return fmt.Errorf("download failed: expected a PDF, got Content-Type %q", ct)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dest), ".pgbook-*.tmp")
	if err != nil {
		return err
	}
	defer func() {
		tmp.Close()
		os.Remove(tmp.Name()) // no-op after a successful rename
	}()

	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(tmp, h), resp.Body)
	if err != nil {
		return fmt.Errorf("download interrupted: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("download failed: empty response")
	}
	if resp.ContentLength > 0 && n != resp.ContentLength {
		return fmt.Errorf("download truncated: got %d of %d bytes", n, resp.ContentLength)
	}
	if meta.SHA256 != "" {
		if got := hex.EncodeToString(h.Sum(nil)); got != meta.SHA256 {
			return fmt.Errorf("checksum mismatch: got %s, published %s", got, meta.SHA256)
		}
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), dest); err != nil {
		return err
	}

	fmt.Fprintf(a.Out, "✓ Saved to %s\n  %d topics · %d pages · version %s\n",
		dest, meta.Topics, meta.Pages, meta.Version)
	return nil
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
