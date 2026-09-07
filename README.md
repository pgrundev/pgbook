# pgbook

**Postgres Book, one topic at a time.**

No 1,000-page manual. Pick a topic, understand how it works, try the examples yourself.

```console
$ pgbook read indexes
```

Free and open source · No signup required

## Install

```bash
curl -fsSL https://pgbook.dev/install.sh | sh
```

Or with Homebrew:

```bash
brew install pgrundev/tap/pgbook
```

Or grab a binary from the [latest release](https://github.com/pgrundev/pgbook/releases/latest) (macOS and Linux, arm64 and amd64), or build from source:

```bash
go install github.com/pgrundev/pgbook@latest
```

To update, run the same command again: the installer always fetches the latest release, and Homebrew users can `brew update && brew upgrade pgbook`. The installer puts the binary in `/usr/local/bin` (asking for sudo if needed); set `PGBOOK_INSTALL_DIR` to install somewhere else:

```bash
curl -fsSL https://pgbook.dev/install.sh | PGBOOK_INSTALL_DIR=~/.local/bin sh
```

Chapter text is fetched from pgbook.dev when you open it, so new and revised chapters arrive without reinstalling anything.

## Table of contents

### Beginner

| # | Topic | |
|---|-------|---|
| 01 | Tables and data types | _in progress_ |
| 02 | SELECT, INSERT, UPDATE, DELETE | _in progress_ |
| 03 | Joins | _in progress_ |
| 04 | **Index basics** — why some queries are instant (hands-on tutorial, five steps) | ✅ `pgbook read indexes` |
| 05 | **Transactions** — grouping statements safely | _in progress_ |
| 06 | Reading EXPLAIN | _in progress_ |

### Intermediate

| # | Topic | |
|---|-------|---|
| 07 | **Locks** — why a query is stuck, not slow | 🔍 in review · `pgbook read locks` |
| 08 | **Transactions and isolation** — what one query can see of another | 🔍 in review · `pgbook read transactions` |
| 09 | **JSONB** — semi-structured data, indexed | 🔍 in review · `pgbook read jsonb` |
| 10 | **Window functions** — running totals without collapsing rows | 🔍 in review · `pgbook read window-functions` |
| 11 | **Row-level security** — access control inside the database | 🔍 in review · `pgbook read row-level-security` |
| 12 | **Vacuum and autovacuum** — why deleted rows still take space | 🔍 in review · `pgbook read vacuum` |
| 13 | Connection pooling | _in progress_ |
| 14 | Finding slow queries | _in progress_ |

### Advanced

| # | Topic | |
|---|-------|---|
| 15 | MVCC | _in progress_ |
| 16 | Query planner | _in progress_ |
| 17 | Index internals | _in progress_ |
| 18 | Deadlocks | _in progress_ |
| 19 | WAL and checkpoints | _in progress_ |
| 20 | Replication — read replicas and failover | 🔍 in review · `pgbook read replication` |
| 21 | Partitioning | _in progress_ |
| 22 | Query-plan optimization | _in progress_ |

✅ finished · 🔍 in review: readable now as a draft, being reworked into the hands-on format below · _in progress_: not yet written.

Every example runs against a real Postgres — copy it straight into your own database.

## Built to be read, not searched

Postgres docs are exhaustive but hard to start in. pgbook picks the topics that actually trip people up, and explains each one in a page, not a chapter.

**Topic-first** — No table of contents to hunt through. Each topic is a single, self-contained page.

**Runnable examples** — Every example is real SQL you can paste into your own Postgres and run immediately.

**Free and open source** — MIT licensed. No account, no paywall, no tracking.

## Read it from the terminal

```console
$ pgbook                  # greeting and where to start
$ pgbook list
$ pgbook read indexes
$ pgbook search locks
$ pgbook next             # continue where you left off
$ pgbook pdf
```

Or just read it at [pgbook.dev](https://pgbook.dev) — no install required.

### Hands-on chapters

Finished chapters are tutorials, not reference pages. Keep `pgbook read` open in one terminal tab and `psql` in another, paste each SQL block as you reach it, and tick off the **Your turn** checklist at the end of every step. The terminal shows where you are:

```text
────────────────────────────────────────────────────────
 1 ─ [2] ─ 3 ─ 4 ─ 5   STEP 2 OF 5
 WATCH A QUERY CRAWL
────────────────────────────────────────────────────────
```

SQL blocks are syntax-colored, and every chapter starts with a one-line Docker command for a throwaway Postgres, so you need nothing installed but Docker (or a database you already have). Nothing in a chapter touches real data.

Topics are fetched from pgbook.dev, so the book updates without a new CLI release, and every topic you open is cached for offline reading. The CLI only displays lessons — it never connects to a database and never executes SQL.

A single static binary — no signup, no Node, no Postgres required to read. See [Install](#install).

### `pgbook pdf`

Download the latest complete edition of Postgres Book as a PDF:

```bash
pgbook pdf
```

Expected output:

```text
Downloading Postgres Book…

✓ Saved to ./postgres-book.pdf
  8 topics · 64 pages · version 0.1
```

Supports a custom destination:

```bash
pgbook pdf --output ~/Downloads/postgres-book.pdf
pgbook pdf -o postgres.pdf
```

Behavior:

- Downloads the latest PDF from pgbook.dev.
- Saves it as `postgres-book.pdf` in the current directory by default.
- Shows download progress, edition version, topic count, page count, and final path.
- Never silently overwrites an existing file — asks for confirmation, or requires `--force`.
- Downloads to a temporary file and renames it only after the download succeeds.
- Validates the HTTP response, content type, file size, and published checksum.
- Removes partial temporary files after failures.
- Returns a non-zero exit code with a useful error message when the download fails.

The PDF is generated from the same source files used by the website, so the website, CLI lessons, and downloadable book always contain the same content.

## API

Public, read-only, versioned JSON. Generated from the same `topics/*.md` source files that feed the CLI and the PDF, so the website, CLI lessons, and downloadable book always contain the same content.

### `GET /api/topics`

The topic index — slug, title, description, level, reading time, order, aliases, and tags for every topic (no lesson content).

### `GET /api/topics/:slug`

One topic with its full markdown lesson content:

```json
{
  "slug": "locks",
  "title": "Locks",
  "description": "Why a query is stuck, not slow",
  "level": "intermediate",
  "reading_minutes": 10,
  "order": 7,
  "aliases": ["locking", "lock", "blocking"],
  "tags": ["concurrency", "transactions", "blocking"],
  "content": "Markdown lesson content"
}
```

### `GET /api/book`

Public endpoint returning metadata about the current edition and its download URL:

```json
{
  "version": "0.1",
  "topics": 8,
  "pages": 64,
  "filename": "postgres-book.pdf",
  "download_url": "https://pgbook.dev/downloads/postgres-book.pdf",
  "sha256": "..."
}
```

## Development

```bash
make test    # run all tests (Go, no cgo, no external deps)
make build   # build the pgbook binary
make site    # regenerate site/api from topics/*.md
make serve   # preview pgbook.dev locally on :8391
```

Lessons live in `topics/*.md` — front matter plus markdown. Edit one, run `make site`, and the CLI, website, and API all pick it up. `PGBOOK_BASE_URL=http://127.0.0.1:8391 pgbook read indexes` points the CLI at your local preview.

### Writing a chapter

- `## Step N: Title` headings become the step tracker; number them from 1 with no gaps.
- End each step with `### Your turn` and a `- [ ]` checklist. The terminal renders the boxes as ☐.
- Fence SQL as ```` ```sql ```` to get syntax colors; use ```` ```bash ```` and ```` ```text ```` for everything else.
- Keep source lines at 76 characters or fewer so nothing wraps in an 80-column terminal (code blocks are indented four spaces).
- Run every SQL block, in order, against a throwaway `postgres:17` container and paste the real plans into the text. Add `ANALYZE <table>;` after bulk loads, or the planner's guesses will not match what you wrote.

### Shipping

- **Chapter text**: `make site`, commit, then `cd deploy/cloudflare && npx wrangler deploy`. Readers get it on their next `pgbook read`; no CLI release needed.
- **The CLI**: push a `v*` tag. CI tests, cross-builds for macOS/Linux (arm64 + amd64), publishes the binaries and `checksums.txt` as a GitHub release. Then copy the version and the four sha256 values into `packaging/homebrew/pgbook.rb` and push the same file to `pgrundev/homebrew-tap` as `Formula/pgbook.rb`.

---

MIT licensed · Contributions welcome
