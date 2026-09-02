---
slug: indexes
title: Indexes
description: Why some queries are instant
level: beginner
reading_minutes: 8
order: 1
aliases: index, index-basics, btree
tags: performance, btree, explain
---

## Why some queries are instant

Without an index, Postgres answers a `WHERE` clause by reading every row
in the table — a sequential scan. An index is a separate structure that
maps column values to row locations, so Postgres can jump straight to
the matching rows.

```sql
CREATE TABLE users (
  id    bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  email text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

INSERT INTO users (email)
SELECT 'user' || n || '@example.com' FROM generate_series(1, 100000) n;
```

## Seeing the difference

`EXPLAIN ANALYZE` shows how Postgres actually ran a query:

```sql
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'user4242@example.com';
```

You will see `Seq Scan on users` — every row was checked. Now add an
index and run it again:

```sql
CREATE INDEX users_email_idx ON users (email);

EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'user4242@example.com';
```

The plan switches to `Index Scan using users_email_idx`, and the
execution time drops from milliseconds to microseconds.

## What a B-tree can do

The default index type is a B-tree. It keeps values sorted, so it
serves more than equality:

- `WHERE email = '...'` — exact matches
- `WHERE created_at > now() - interval '1 day'` — ranges
- `ORDER BY created_at DESC LIMIT 10` — ordering, without a sort
- `WHERE email LIKE 'user42%'` — left-anchored prefixes (with the right
  operator class or collation)

It cannot help with `LIKE '%@gmail.com'` — the sorted order is useless
when the prefix is unknown.

## Multi-column indexes

Column order matters. An index on `(a, b)` is sorted by `a` first, then
`b` — like a phone book sorted by last name, then first name.

```sql
CREATE INDEX orders_customer_created_idx ON orders (customer_id, created_at);
```

This serves `WHERE customer_id = 7` and
`WHERE customer_id = 7 AND created_at > '2026-01-01'`,
but not `WHERE created_at > '2026-01-01'` alone.

## Indexes are not free

Every index slows down writes: each `INSERT`, `UPDATE`, and `DELETE`
must maintain it, and it takes disk space. Find indexes that are never
read:

```sql
SELECT indexrelname, idx_scan
FROM pg_stat_user_indexes
ORDER BY idx_scan ASC;
```

> Note: an `idx_scan` of 0 on a busy table usually means the index can
> be dropped — but check replicas and rare reports first.
