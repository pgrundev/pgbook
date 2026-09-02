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
| 04 | **Index basics** — why some queries are instant | ✅ |
| 05 | **Transactions** — grouping statements safely | ✅ |
| 06 | Reading EXPLAIN | _in progress_ |

### Intermediate

| # | Topic | |
|---|-------|---|
| 07 | **Locks** — why a query is stuck, not slow | ✅ |
| 08 | **Transactions and isolation** — what one query can see of another | ✅ |
| 09 | **JSONB** — semi-structured data, indexed | ✅ |
| 10 | **Window functions** — running totals without collapsing rows | ✅ |
| 11 | **Row-level security** — access control inside the database | ✅ |
| 12 | **Vacuum and autovacuum** — why deleted rows still take space | ✅ |
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

Eight topics written so far, more in progress. Every example runs against a real Postgres — copy it straight into your own database.

## Built to be read, not searched

Postgres docs are exhaustive but hard to start in. pgbook picks the topics that actually trip people up, and explains each one in a page, not a chapter.

**Topic-first** — No table of contents to hunt through. Each topic is a single, self-contained page.

**Runnable examples** — Every example is real SQL you can paste into your own Postgres and run immediately.

**Free and open source** — MIT licensed. No account, no paywall, no tracking.

## Read it from the terminal

```console
$ pgbook list
$ pgbook read transactions
$ pgbook run transactions --example 1
```

Or just read it at [pgbook.dev](https://pgbook.dev) — no install required.

---

MIT licensed · Contributions welcome
