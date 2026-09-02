// gensite builds the static pgbook.dev API files from topics/*.md.
//
// Usage: go run ./tools/gensite [-topics topics] [-out site]
//
// It writes:
//
//	site/api/topics          — the versioned topic index (no content)
//	site/api/topics/<slug>   — one JSON document per topic, with content
//	site/api/book            — PDF edition metadata, only when
//	                           site/downloads/postgres-book.pdf exists
//
// The same source files feed the CLI, the website, and the PDF, so the
// three can never drift apart.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pgrundev/pgbook/internal/topic"
)

const (
	bookVersion = "0.1"
	baseURL     = "https://pgbook.dev"
	pdfName     = "postgres-book.pdf"
)

type index struct {
	Version string        `json:"version"`
	Topics  []topic.Topic `json:"topics"`
}

type bookMeta struct {
	Version     string `json:"version"`
	Topics      int    `json:"topics"`
	Pages       int    `json:"pages"`
	Filename    string `json:"filename"`
	DownloadURL string `json:"download_url"`
	SHA256      string `json:"sha256"`
}

func generate(topicsDir, outDir string) error {
	paths, err := filepath.Glob(filepath.Join(topicsDir, "*.md"))
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return fmt.Errorf("no topic files in %s", topicsDir)
	}

	var topics []topic.Topic
	seen := map[string]string{}
	for _, p := range paths {
		src, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		t, err := topic.Parse(src)
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		}
		if prev, dup := seen[t.Slug]; dup {
			return fmt.Errorf("duplicate slug %q in %s and %s", t.Slug, prev, p)
		}
		seen[t.Slug] = p
		topics = append(topics, t)
	}
	sort.SliceStable(topics, func(i, j int) bool { return topics[i].Order < topics[j].Order })

	apiDir := filepath.Join(outDir, "api", "topics")
	if err := os.MkdirAll(apiDir, 0o755); err != nil {
		return err
	}

	// Per-topic documents, with content.
	for _, t := range topics {
		if err := writeJSON(filepath.Join(apiDir, t.Slug), t); err != nil {
			return err
		}
	}

	// The index, without content. Named topics.json because the
	// "topics" path is the directory of per-slug documents; the site's
	// _redirects file rewrites GET /api/topics to it.
	slim := make([]topic.Topic, len(topics))
	copy(slim, topics)
	for i := range slim {
		slim[i].Content = ""
	}
	if err := writeJSON(filepath.Join(outDir, "api", "topics.json"), index{Version: bookVersion, Topics: slim}); err != nil {
		return err
	}

	// PDF edition metadata, only when the PDF has been built.
	pdfPath := filepath.Join(outDir, "downloads", pdfName)
	if data, err := os.ReadFile(pdfPath); err == nil {
		sum := sha256.Sum256(data)
		meta := bookMeta{
			Version:     bookVersion,
			Topics:      len(topics),
			Pages:       countPDFPages(data),
			Filename:    pdfName,
			DownloadURL: baseURL + "/downloads/" + pdfName,
			SHA256:      hex.EncodeToString(sum[:]),
		}
		if err := writeJSON(filepath.Join(outDir, "api", "book"), meta); err != nil {
			return err
		}
	}
	return nil
}

// writeJSON writes v as pretty JSON; extensionless API files get their
// content type from the site's _headers file.
func writeJSON(path string, v any) error {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

// countPDFPages counts page objects in a simple PDF. Good enough for
// metadata; a wrong count harms nothing.
func countPDFPages(data []byte) int {
	return strings.Count(string(data), "/Type /Page") - strings.Count(string(data), "/Type /Pages")
}

func main() {
	topicsDir := flag.String("topics", "topics", "directory of topic source files")
	outDir := flag.String("out", "site", "site output directory")
	flag.Parse()
	if err := generate(*topicsDir, *outDir); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("generated %s/api from %s\n", *outDir, *topicsDir)
}
