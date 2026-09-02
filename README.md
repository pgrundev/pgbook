# pgbook

**Postgres Book, one topic at a time.**

No 1,000-page manual. Pick a topic, understand how it works, try the examples yourself.

```console
$ pgbook read locks
```

Free and open source · No signup required

## Table of contents

### Beginner

| # | Topic | |
|---|-------|---|
| 01 | Tables and data types | _in progress_ |
| 02 | SELECT, INSERT, UPDATE, DELETE | _in progress_ |
| 03 | Joins | _in progress_ |
| 04 | **Index basics** — why some queries are instant | ✅ `pgbook read indexes` |
| 05 | **Transactions** — grouping statements safely | _in progress_ |
| 06 | Reading EXPLAIN | _in progress_ |

### Intermediate

| # | Topic | |
|---|-------|---|
| 07 | **Locks** — why a query is stuck, not slow | ✅ `pgbook read locks` |
| 08 | **Transactions and isolation** — what one query can see of another | ✅ `pgbook read transactions` |
| 09 | **JSONB** — semi-structured data, indexed | ✅ `pgbook read jsonb` |
| 10 | **Window functions** — running totals without collapsing rows | ✅ `pgbook read window-functions` |
| 11 | **Row-level security** — access control inside the database | ✅ `pgbook read row-level-security` |
| 12 | **Vacuum and autovacuum** — why deleted rows still take space | ✅ `pgbook read vacuum` |
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
| 20 | Replication — read replicas and failover | ✅ `pgbook read replication` |
| 21 | Partitioning | _in progress_ |
| 22 | Query-plan optimization | _in progress_ |

Eight topics are written so far, more in progress. Every example runs against a real Postgres — copy it straight into your own database.

## Built to be read, not searched

Postgres docs are exhaustive but hard to start in. pgbook picks the topics that actually trip people up, and explains each one in a page, not a chapter.

**Topic-first** — No table of contents to hunt through. Each topic is a single, self-contained page.

**Runnable examples** — Every example is real SQL you can paste into your own Postgres and run immediately.

**Free and open source** — MIT licensed. No account, no paywall, no tracking.

## Read it from the terminal

```console
$ pgbook list
$ pgbook read locks
$ pgbook search indexes
$ pgbook next
$ pgbook pdf
```

Or just read it at [pgbook.dev](https://pgbook.dev) — no install required.

Topics are fetched from pgbook.dev, so the book updates without a new CLI release, and every topic you open is cached for offline reading. The CLI only displays lessons — it never connects to a database and never executes SQL.

### Install

```bash
curl -fsSL https://pgbook.dev/install.sh | sh
```

Or with Homebrew:

```bash
brew install pgrundev/tap/pgbook
```

macOS and Linux, arm64 and amd64. A single static binary — no signup, no Node, no Postgres required to read.

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

Lessons live in `topics/*.md` — front matter plus markdown. Edit one, run `make site`, and the CLI, website, and API all pick it up. `PGBOOK_BASE_URL=http://127.0.0.1:8391 pgbook list` points the CLI at your local preview.

Releases are cut by pushing a `v*` tag: CI tests, cross-builds for macOS/Linux (arm64 + amd64), checksums, and publishes the binaries as a GitHub release.

---

MIT licensed · Contributions welcome
