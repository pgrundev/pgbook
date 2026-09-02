---
slug: vacuum
title: Vacuum & autovacuum
description: Why deleted rows still take space
level: intermediate
reading_minutes: 10
order: 6
aliases: autovacuum, vacuum-autovacuum, bloat
tags: mvcc, maintenance, bloat
---

## DELETE does not delete

When you `DELETE` or `UPDATE` a row, Postgres does not reclaim its
space. It marks the old version dead and leaves it in place, so
transactions that started earlier can still see it (that's MVCC —
readers never block writers). An `UPDATE` is really an insert of a new
version plus a soft-delete of the old one.

Dead versions pile up. `VACUUM` is the process that reclaims them.

## Watching it happen

```sql
CREATE TABLE counters (id int PRIMARY KEY, n int);
INSERT INTO counters SELECT g, 0 FROM generate_series(1, 100000) g;

UPDATE counters SET n = n + 1;   -- rewrites every row

SELECT n_dead_tup, n_live_tup, last_autovacuum
FROM pg_stat_user_tables WHERE relname = 'counters';
```

After the update, `n_dead_tup` is about 100,000 — the whole table
exists twice on disk until vacuum runs.

## What VACUUM does (and does not)

```sql
VACUUM counters;
```

- marks dead space reusable by future inserts and updates
- updates the visibility map (which speeds up index-only scans)
- prevents transaction ID wraparound (the reason vacuum can never be
  disabled entirely)

It does **not** shrink the file on disk except in the tail-end case.
`VACUUM FULL` does shrink, but takes an exclusive lock and rewrites the
whole table — an outage, not maintenance.

## Autovacuum

Postgres runs vacuum for you when dead tuples pass a threshold —
roughly 20% of the table by default. For a big table that is a lot of
bloat before anything happens. High-churn tables deserve tighter
settings:

```sql
ALTER TABLE counters SET (
  autovacuum_vacuum_scale_factor = 0.02,
  autovacuum_vacuum_cost_delay   = 1
);
```

## What blocks vacuum

Vacuum can only remove versions no transaction can still see. The
classic causes of "vacuum runs but nothing shrinks":

- a transaction open for hours (`idle in transaction`)
- an abandoned replication slot
- a long-running query on a hot-standby with feedback on

```sql
SELECT pid, state, xact_start, query
FROM pg_stat_activity
WHERE state <> 'idle'
ORDER BY xact_start;
```

> Note: if `n_dead_tup` keeps climbing while autovacuum is running,
> don't tune vacuum first — find the old transaction or slot pinning
> the horizon and kill it.
