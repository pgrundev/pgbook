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
| 04 | **Index basics** — why some queries are instant | _in progress_ |
| 05 | **Transactions** — grouping statements safely | _in progress_ |
| 06 | Reading EXPLAIN | _in progress_ |

### Intermediate

| # | Topic | |
|---|-------|---|
| 07 | **Locks** — why a query is stuck, not slow | _in progress_ |
| 08 | **Transactions and isolation** — what one query can see of another | _in progress_ |
| 09 | **JSONB** — semi-structured data, indexed | _in progress_ |
| 10 | **Window functions** — running totals without collapsing rows | _in progress_ |
| 11 | **Row-level security** — access control inside the database | _in progress_ |
| 12 | **Vacuum and autovacuum** — why deleted rows still take space | _in progress_ |
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
| 20 | Replication — read replicas and failover | _in progress_ |
| 21 | Partitioning | _in progress_ |
| 22 | Query-plan optimization | _in progress_ |

All topics are in progress. Every example runs against a real Postgres — copy it straight into your own database.

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

---

MIT licensed · Contributions welcome
