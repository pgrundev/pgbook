---
slug: jsonb
title: JSON & JSONB
description: Semi-structured data, indexed
level: intermediate
reading_minutes: 9
order: 3
aliases: json, json-jsonb
tags: jsonb, gin, schema-design
---

## Two JSON types, one right answer

Postgres has `json` (stores the exact text) and `jsonb` (stores a
parsed binary form). Use `jsonb` — it supports indexing, containment
queries, and ignores insignificant whitespace and duplicate keys.

```sql
CREATE TABLE events (
  id      bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  payload jsonb NOT NULL
);

INSERT INTO events (payload) VALUES
  ('{"type": "signup", "user": {"id": 7, "plan": "pro"}}'),
  ('{"type": "login",  "user": {"id": 7}}'),
  ('{"type": "signup", "user": {"id": 9, "plan": "free"}}');
```

## Reading values

```sql
SELECT payload->>'type'                 AS type,     -- text
       payload->'user'->>'plan'         AS plan,     -- nested
       payload#>>'{user,id}'            AS user_id   -- path form
FROM events;
```

`->` returns `jsonb`, `->>` returns `text`. Chain `->` while
navigating, use `->>` for the final value.

## Containment: the jsonb superpower

`@>` asks "does this document contain this shape?":

```sql
SELECT * FROM events
WHERE payload @> '{"type": "signup"}';

SELECT * FROM events
WHERE payload @> '{"user": {"plan": "pro"}}';
```

## Indexing JSONB with GIN

Containment queries stay sequential scans until you add a GIN index:

```sql
CREATE INDEX events_payload_idx ON events USING gin (payload);

EXPLAIN ANALYZE SELECT * FROM events
WHERE payload @> '{"type": "signup"}';
```

With enough rows the plan becomes a `Bitmap Index Scan`. If you only
ever query containment, the smaller `jsonb_path_ops` variant is faster:

```sql
CREATE INDEX events_payload_path_idx ON events
USING gin (payload jsonb_path_ops);
```

## Indexing one field with a B-tree

When one key is queried with equality or ranges, an expression index
beats GIN:

```sql
CREATE INDEX events_type_idx ON events ((payload->>'type'));

SELECT count(*) FROM events WHERE payload->>'type' = 'login';
```

## When not to use JSONB

If every row has the same keys and you filter or join on them, they are
columns. JSONB shines for genuinely variable data: webhook payloads,
user-defined attributes, API responses.

> Note: `jsonb` values are stored whole. Updating one key rewrites the
> entire document (and every index on it) — high-churn counters do not
> belong inside a big JSONB blob.
