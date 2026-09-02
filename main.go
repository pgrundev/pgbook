// pgbook is the terminal reader for the Postgres Book (pgbook.dev).
// It only displays lessons — it never connects to a database or runs SQL.
package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/pgrundev/pgbook/commands"
	"github.com/pgrundev/pgbook/internal/client"
	"github.com/pgrundev/pgbook/internal/render"
)

// version is overridden at release time via -ldflags "-X main.version=...".
var version = "0.1.0"

const usage = `pgbook — the Postgres Book in your terminal (pgbook.dev)

Usage:
  pgbook list                 Show all topics
  pgbook read <topic>         Read a topic (e.g. pgbook read locks)
  pgbook search <query>       Search titles, headings, and lesson text
  pgbook next                 Continue where you left off
  pgbook pdf [-o file]        Download the book as a PDF

Flags:
  -o, --output <file>         pdf: destination path
      --force                 pdf: overwrite an existing file
  -h, --help                  Show this help
      --version               Show version

Topics are fetched from pgbook.dev and cached for offline reading.
pgbook never connects to a database and never executes SQL.
`

func run(args []string, stdout, stderr io.Writer) int {
	baseURL := os.Getenv("PGBOOK_BASE_URL")
	if baseURL == "" {
		baseURL = "https://pgbook.dev"
	}
	app := &commands.App{
		Client: client.New(baseURL),
		Out:    stdout,
		Color:  render.ShouldColor(isTTY(stdout), os.Getenv("NO_COLOR")),
	}

	if len(args) == 0 {
		fmt.Fprint(stdout, usage)
		return 0
	}
	cmd, rest := args[0], args[1:]

	fail := func(err error) int {
		fmt.Fprintf(stderr, "pgbook: %v\n", err)
		return 1
	}

	switch cmd {
	case "help", "--help", "-h":
		fmt.Fprint(stdout, usage)
		return 0
	case "--version", "version", "-v":
		fmt.Fprintf(stdout, "pgbook %s\n", version)
		return 0
	case "list":
		if err := app.List(); err != nil {
			return fail(err)
		}
	case "read":
		if len(rest) != 1 || strings.HasPrefix(rest[0], "-") {
			fmt.Fprintln(stderr, "usage: pgbook read <topic>   (see: pgbook list)")
			return 2
		}
		if err := paged(app, stdout, func() error { return app.Read(rest[0]) }); err != nil {
			return fail(err)
		}
	case "search":
		if len(rest) == 0 {
			fmt.Fprintln(stderr, "usage: pgbook search <query>")
			return 2
		}
		if err := app.Search(strings.Join(rest, " ")); err != nil {
			return fail(err)
		}
	case "next":
		if err := paged(app, stdout, app.Next); err != nil {
			return fail(err)
		}
	case "pdf":
		dest := ""
		force := false
		for i := 0; i < len(rest); i++ {
			switch rest[i] {
			case "-o", "--output":
				if i+1 == len(rest) {
					fmt.Fprintln(stderr, "usage: pgbook pdf [-o <file>] [--force]")
					return 2
				}
				i++
				dest = rest[i]
			case "--force":
				force = true
			default:
				fmt.Fprintf(stderr, "pgbook pdf: unknown flag %q\n", rest[i])
				return 2
			}
		}
		if err := app.PDF(dest, force); err != nil {
			return fail(err)
		}
	default:
		fmt.Fprintf(stderr, "pgbook: unknown command %q (try: pgbook --help)\n", cmd)
		return 2
	}
	return 0
}

// paged pipes a lesson through the system pager when stdout is a
// terminal; redirected output is written straight through.
func paged(app *commands.App, stdout io.Writer, f func() error) error {
	if !isTTY(stdout) {
		return f()
	}
	pager := os.Getenv("PAGER")
	if pager == "" {
		pager = "less -R"
	}
	parts := strings.Fields(pager)
	path, err := exec.LookPath(parts[0])
	if err != nil {
		return f() // no pager available: print directly
	}
	cmd := exec.Command(path, parts[1:]...)
	pipe, err := cmd.StdinPipe()
	if err != nil {
		return f()
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return f()
	}
	prev := app.Out
	app.Out = pipe
	runErr := f()
	app.Out = prev
	pipe.Close()
	if err := cmd.Wait(); err != nil && runErr == nil {
		runErr = err
	}
	return runErr
}

func isTTY(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}
