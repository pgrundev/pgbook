package commands

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/pgrundev/pgbook/internal/client"
)

var pdfBody = []byte("%PDF-1.4 fake book body")

func api(t *testing.T) *httptest.Server {
	t.Helper()
	sum := sha256.Sum256(pdfBody)
	mux := http.NewServeMux()
	mux.HandleFunc("/api/topics", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"version":"0.1","topics":[
		  {"slug":"indexes","title":"Indexes","description":"Why some queries are instant","level":"beginner","reading_minutes":8,"order":1},
		  {"slug":"jsonb","title":"JSON & JSONB","description":"Semi-structured data, indexed","level":"intermediate","reading_minutes":9,"order":3,"aliases":["json"]},
		  {"slug":"locks","title":"Locks","description":"Why a query is stuck, not slow","level":"intermediate","reading_minutes":10,"order":7,"aliases":["locking"],"tags":["concurrency"]}
		]}`)
	})
	mux.HandleFunc("/api/topics/locks", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"slug":"locks","title":"Locks","description":"Why a query is stuck, not slow","level":"intermediate","reading_minutes":10,"order":7,"content":"## What Postgres locks\n\nEvery statement takes locks.\n"}`)
	})
	mux.HandleFunc("/api/topics/indexes", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"slug":"indexes","title":"Indexes","description":"Why some queries are instant","level":"beginner","reading_minutes":8,"order":1,"content":"## B-tree basics\n\nMost indexes are B-trees.\n"}`)
	})
	mux.HandleFunc("/api/book", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"version":"0.1","topics":8,"pages":64,"filename":"postgres-book.pdf","download_url":"http://%s/downloads/postgres-book.pdf","sha256":"%s"}`,
			r.Host, hex.EncodeToString(sum[:]))
	})
	mux.HandleFunc("/downloads/postgres-book.pdf", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(pdfBody)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func app(t *testing.T, srv *httptest.Server) (*App, *bytes.Buffer) {
	t.Helper()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	t.Setenv("XDG_STATE_HOME", t.TempDir())
	var out bytes.Buffer
	return &App{
		Client: client.New(srv.URL),
		Out:    &out,
		Color:  false,
	}, &out
}

func TestList(t *testing.T) {
	a, out := app(t, api(t))
	if err := a.List(); err != nil {
		t.Fatalf("List: %v", err)
	}
	s := out.String()
	for _, want := range []string{
		"POSTGRES BOOK",
		"01", "Indexes", "beginner",
		"07", "Locks", "intermediate",
		"pgbook read",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("list output missing %q:\n%s", want, s)
		}
	}
	// Ordered by book order.
	if strings.Index(s, "Indexes") > strings.Index(s, "Locks") {
		t.Errorf("topics out of order:\n%s", s)
	}
}

func TestReadRendersAndSavesProgress(t *testing.T) {
	a, out := app(t, api(t))
	if err := a.Read("locks"); err != nil {
		t.Fatalf("Read: %v", err)
	}
	s := out.String()
	for _, want := range []string{
		"LOCKS",
		"Why a query is stuck, not slow",
		"Intermediate · 10 min",
		"WHAT POSTGRES LOCKS",
		"Every statement takes locks.",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("read output missing %q:\n%s", want, s)
		}
	}
	// Progress saved so `next` continues after locks.
	data, err := os.ReadFile(filepath.Join(os.Getenv("XDG_STATE_HOME"), "pgbook", "progress.json"))
	if err != nil || !strings.Contains(string(data), "locks") {
		t.Errorf("progress not saved: %v %s", err, data)
	}
}

func TestReadResolvesAliases(t *testing.T) {
	a, out := app(t, api(t))
	if err := a.Read("locking"); err != nil {
		t.Fatalf("Read(locking): %v", err)
	}
	if !strings.Contains(out.String(), "LOCKS") {
		t.Errorf("alias did not resolve:\n%s", out.String())
	}
}

func TestReadUnknownTopicSuggests(t *testing.T) {
	a, _ := app(t, api(t))
	err := a.Read("lokcs")
	if err == nil {
		t.Fatal("Read(lokcs) succeeded, want error")
	}
	if !strings.Contains(err.Error(), "locks") {
		t.Errorf("error should suggest the closest topic: %v", err)
	}
}

func TestSearchOutput(t *testing.T) {
	a, out := app(t, api(t))
	if err := a.Search("instant"); err != nil {
		t.Fatalf("Search: %v", err)
	}
	s := out.String()
	for _, want := range []string{"1 result", "Indexes", "pgbook read indexes"} {
		if !strings.Contains(s, want) {
			t.Errorf("search output missing %q:\n%s", want, s)
		}
	}
}

func TestSearchNoResultsSuggestsClosest(t *testing.T) {
	a, out := app(t, api(t))
	if err := a.Search("lockz"); err != nil {
		t.Fatalf("Search: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "No results") || !strings.Contains(s, "locks") {
		t.Errorf("no-result output should suggest closest topics:\n%s", s)
	}
}

func TestNextFromNothingOpensFirstTopic(t *testing.T) {
	a, out := app(t, api(t))
	if err := a.Next(); err != nil {
		t.Fatalf("Next: %v", err)
	}
	if !strings.Contains(out.String(), "INDEXES") {
		t.Errorf("next without progress should open the first topic:\n%s", out.String())
	}
}

func TestNextAfterFinalTopicExplains(t *testing.T) {
	a, out := app(t, api(t))
	if err := a.Read("locks"); err != nil { // locks is the last topic (order 7)
		t.Fatalf("Read: %v", err)
	}
	out.Reset()
	if err := a.Next(); err != nil {
		t.Fatalf("Next: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "end") || !strings.Contains(s, "pgbook list") {
		t.Errorf("end-of-book message should mention the end and pgbook list:\n%s", s)
	}
}

func TestPDFDownloads(t *testing.T) {
	a, out := app(t, api(t))
	dest := filepath.Join(t.TempDir(), "postgres-book.pdf")
	if err := a.PDF(dest, false); err != nil {
		t.Fatalf("PDF: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil || !bytes.Equal(got, pdfBody) {
		t.Fatalf("downloaded file wrong: %v", err)
	}
	s := out.String()
	for _, want := range []string{"Saved to", "8 topics", "64 pages", "version 0.1"} {
		if !strings.Contains(s, want) {
			t.Errorf("pdf output missing %q:\n%s", want, s)
		}
	}
}

func TestPDFRefusesOverwriteWithoutForce(t *testing.T) {
	a, _ := app(t, api(t))
	dest := filepath.Join(t.TempDir(), "postgres-book.pdf")
	if err := os.WriteFile(dest, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := a.PDF(dest, false)
	if err == nil {
		t.Fatal("PDF overwrote existing file without --force")
	}
	if !strings.Contains(err.Error(), "--force") {
		t.Errorf("error should mention --force: %v", err)
	}
	if got, _ := os.ReadFile(dest); string(got) != "old" {
		t.Errorf("existing file was modified")
	}
	// --force overwrites.
	if err := a.PDF(dest, true); err != nil {
		t.Fatalf("PDF --force: %v", err)
	}
	if got, _ := os.ReadFile(dest); !bytes.Equal(got, pdfBody) {
		t.Errorf("--force did not replace the file")
	}
}

func TestPDFChecksumMismatchFailsAndCleansUp(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/book", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"version":"0.1","topics":8,"pages":64,"filename":"postgres-book.pdf","download_url":"http://%s/downloads/postgres-book.pdf","sha256":"deadbeef"}`, r.Host)
	})
	mux.HandleFunc("/downloads/postgres-book.pdf", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		w.Write(pdfBody)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	a, _ := app(t, srv)
	dir := t.TempDir()
	dest := filepath.Join(dir, "postgres-book.pdf")
	err := a.PDF(dest, false)
	if err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("PDF with bad checksum: err = %v, want checksum error", err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("partial files left behind: %v", entries)
	}
}

func TestPDFWrongContentTypeFails(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/book", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"version":"0.1","topics":8,"pages":64,"filename":"postgres-book.pdf","download_url":"http://%s/downloads/postgres-book.pdf","sha256":""}`, r.Host)
	})
	mux.HandleFunc("/downloads/postgres-book.pdf", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, "<html>not a pdf</html>")
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	a, _ := app(t, srv)
	dest := filepath.Join(t.TempDir(), "postgres-book.pdf")
	if err := a.PDF(dest, false); err == nil {
		t.Fatal("PDF accepted a text/html response")
	}
}
