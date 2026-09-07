---
slug: indexes
title: Indexes
description: Why some queries are instant
level: beginner
reading_minutes: 15
order: 1
aliases: index, index-basics, btree
tags: performance, btree, explain, tutorial
---

## Why some queries are instant

Imagine this book had no index at the back. To find every mention of
"vacuum" you would flip through all 1,000 pages. With the index you jump
straight to page 412.

A database index is exactly that. Without one, Postgres answers
`WHERE email = '...'` by reading every row in the table, top to bottom.
That is called a sequential scan. With an index, Postgres jumps straight
to the matching rows.

This chapter is hands-on. You will build a table with a million rows,
watch a query crawl, add an index, and watch the same query become
instant. Then you will learn when an index does not help, and what it
costs.

### How to use this chapter

- Keep this tab open for reading.
- Open a second terminal tab for Postgres. Every `sql` block below is
  meant to be pasted there.
- Each step ends with **Your turn**: a short checklist. Do it before
  moving on.

Nothing here touches real data. You will create one throwaway table
and drop it at the end. Plan on 20 to 30 minutes.

## Step 1: Get a Postgres to play with

If you already have a Postgres you can log into with `psql`, use it and
skip to Your turn.

Otherwise, start one in Docker. In your second tab:

```bash
docker run --rm -d --name pgbook-lab \
  -e POSTGRES_PASSWORD=pgbook \
  -p 5432:5432 postgres:17
```

This starts Postgres 17 in the background. `--rm` means the container,
and everything in it, disappears when you stop it. Give it a few
seconds, then open a SQL prompt inside it:

```bash
docker exec -it pgbook-lab psql -U postgres
```

You should see:

```text
psql (17.x)
Type "help" for help.

postgres=#
```

That `postgres=#` prompt is where you type SQL for the rest of this
chapter. Turn on timing, so psql prints how long every query took:

```sql
\timing on
```

### Your turn

- [ ] Start Postgres (Docker, or one you already have).
- [ ] Open `psql` and see the `postgres=#` prompt.
- [ ] Run `\timing on`.
- [ ] Run `SELECT version();` to confirm the connection works.

> Stuck? `docker logs pgbook-lab` shows why the container did not
> start. The usual cause is another Postgres already using port 5432.

## Step 2: Watch a query crawl

Create a `users` table and fill it with one million rows.
`generate_series` is Postgres's built-in row factory: it returns the
numbers 1 to 1,000,000, and each one becomes a user. Paste this in your
psql tab:

```sql
CREATE TABLE users (
  id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  email      text NOT NULL,
  country    text NOT NULL,
  created_at timestamptz NOT NULL
);

INSERT INTO users (email, country, created_at)
SELECT 'user' || n || '@example.com',
       (ARRAY['US', 'DE', 'BR', 'IN', 'JP'])[1 + n % 5],
       now() - n * interval '1 second'
FROM generate_series(1, 1000000) AS n;

ANALYZE users;
```

The insert takes a few seconds. `ANALYZE` makes Postgres look at the
data and record statistics about it, such as how many rows there are
and how many distinct emails. Postgres does this on its own in the
background; running it by hand just means you do not have to wait.

Now find one user by email:

```sql
SELECT * FROM users WHERE email = 'user424242@example.com';
```

One row comes back, and `\timing` reports tens of milliseconds. Run it
again. The first run can be slower while the table is read from disk,
but after that it is the same every time.

Why so slow for one row? Ask Postgres how it ran the query:

```sql
EXPLAIN ANALYZE
SELECT * FROM users WHERE email = 'user424242@example.com';
```

`EXPLAIN ANALYZE` runs the query and prints the plan Postgres used.
Yours will have different numbers, but look for these lines:

```text
Gather  (actual time=9.440..22.478 rows=1 loops=1)
  Workers Planned: 2
  Workers Launched: 2
  ->  Parallel Seq Scan on users  (actual time=14.0..17.5 rows=0 loops=3)
        Filter: (email = 'user424242@example.com'::text)
        Rows Removed by Filter: 333333
Execution Time: 22.494 ms
```

- `Seq Scan`: Postgres read the whole table, top to bottom. `Parallel`
  means it split the work across several processes, which is why
  `loops=3`: the main process plus two helpers each read a third.
- `Rows Removed by Filter: 333333`: each of the three looked at 333,333
  rows and threw away all but, at most, one. A million rows examined to
  find one.

That is the "flipping through every page" problem. On a small machine
you may see a plain `Seq Scan` with `Rows Removed by Filter: 999999`
instead: the same work, done by one process.

### Your turn

- [ ] Create and fill the `users` table, then `ANALYZE` it.
- [ ] Run the `SELECT` twice and note the time.
- [ ] Run `EXPLAIN ANALYZE`. Find `Seq Scan` and `Rows Removed by Filter`.
- [ ] Write the `Execution Time` down. You will compare it in Step 3.

## Step 3: Add an index and run it again

Create an index on the `email` column:

```sql
CREATE INDEX users_email_idx ON users (email);
```

This takes a second or two: Postgres reads every email once and stores
them, sorted, in a separate structure on disk. Now run the exact same
query as before:

```sql
SELECT * FROM users WHERE email = 'user424242@example.com';
```

Well under a millisecond. Dozens of times faster, and the query text
did not change at all. Check the plan:

```sql
EXPLAIN ANALYZE
SELECT * FROM users WHERE email = 'user424242@example.com';
```

```text
Index Scan using users_email_idx on users  (actual rows=1 loops=1)
  Index Cond: (email = 'user424242@example.com'::text)
Execution Time: 0.014 ms
```

`Index Scan` replaced `Seq Scan`, the helpers are gone, and so is
`Rows Removed by Filter`. Postgres jumped straight to the row.

### What the index actually is

The default index type is a B-tree: every value, kept in sorted order,
with a small tree on top so any value can be found in a handful of
hops. Finding one email among a million sorted entries takes about 20
comparisons instead of a million.

Because it is sorted, a B-tree helps with more than exact matches:

- `WHERE email = 'x'`: equality
- `WHERE created_at > now() - interval '1 hour'`: ranges
- `ORDER BY created_at DESC LIMIT 10`: ordering, without a sort
- `WHERE email LIKE 'user4242%'`: prefixes (with a caveat, see Step 4)

### Your turn

- [ ] Create `users_email_idx`.
- [ ] Re-run the `SELECT`. Compare the time with Step 2.
- [ ] Re-run `EXPLAIN ANALYZE` and confirm `Index Scan using users_email_idx`.
- [ ] Try a range query:

```sql
EXPLAIN ANALYZE
SELECT count(*) FROM users WHERE created_at > now() - interval '1 hour';
```

It is a `Seq Scan`. Why? There is no index on `created_at` yet. Create
one, `users_created_at_idx`, and run it again. The plan should now name
your new index.

## Step 4: When the index does not help

Having an index does not guarantee it will make a query fast. Three
cases catch people every week. Before you run each query, guess what
the plan will be.

**1. A function wrapped around the column.**

```sql
EXPLAIN ANALYZE
SELECT * FROM users WHERE lower(email) = 'user424242@example.com';
```

`Seq Scan`, and slower than Step 2. The index stores `email`, not
`lower(email)`. To know whether a row matches, Postgres has to compute
`lower()` on every row, which is the same work as having no index,
plus a million function calls. The fix is either to stop wrapping the
column, or to index the expression itself:

```sql
CREATE INDEX users_lower_email_idx ON users (lower(email));
```

Run the `EXPLAIN ANALYZE` again. The plan now names
`users_lower_email_idx`, and the time is back under a millisecond.
(You may see `Bitmap Index Scan` rather than `Index Scan`. It is a
cousin: both jump through the index instead of reading the table.)

**2. A wildcard at the start.**

```sql
EXPLAIN ANALYZE
SELECT count(*) FROM users WHERE email LIKE '%42@example.com';
```

`Seq Scan`. A sorted list is useless when you do not know how the value
starts, like finding every surname ending in "-son" in a phone book.
(The prefix form `LIKE 'user42%'` is only served by a B-tree when the
index uses `text_pattern_ops` or the database collation is `C`. When in
doubt, `EXPLAIN` it.)

**3. The condition matches a big slice of the table.**

```sql
CREATE INDEX users_country_idx ON users (country);

EXPLAIN ANALYZE
SELECT count(*), min(created_at) FROM users WHERE country = 'DE';
```

Every fifth user is in `DE`, so this matches 200,000 rows. Depending on
your machine you will see either a `Seq Scan`, or a `Bitmap Heap Scan`
that mentions `users_country_idx`. If it is the latter, look at
`Heap Blocks`: Postgres used the index to make a list of matching rows,
then still had to visit nearly every page of the table, because `DE`
users are on every page.

Either way, compare `Execution Time` with the seq scan in Step 2. About
the same. The index exists, may even be used, and buys nothing.

Indexes shine when the condition is selective: it picks out a small
fraction of the rows, so that most of the table can be skipped.

### Your turn

- [ ] Run all three queries. Confirm none of them is faster than Step 2.
- [ ] Create the `lower(email)` index and re-run query 1.
- [ ] For every index you are tempted to add in production, ask: is the
  condition selective, and does the query use the column bare?

## Step 5: Indexes are not free

An index is a second copy of a column, kept sorted on disk. Look at what
you have built so far:

```sql
SELECT relname, pg_size_pretty(pg_relation_size(oid)) AS size
FROM pg_class
WHERE relname LIKE 'users%' AND relkind IN ('r', 'i')
ORDER BY pg_relation_size(oid) DESC;
```

```text
        relname        |  size
-----------------------+---------
 users                 | 73 MB
 users_email_idx       | 39 MB
 users_lower_email_idx | 39 MB
 users_created_at_idx  | 21 MB
 users_pkey            | 21 MB
 users_country_idx     | 6800 kB
```

The two email indexes together are bigger than the table. Each index
takes disk, and each one must be updated on every `INSERT`, `UPDATE`,
and `DELETE`. Time a write with all of them in place:

```sql
INSERT INTO users (email, country, created_at)
SELECT 'late' || n || '@example.com', 'US', now()
FROM generate_series(1, 200000) AS n;
```

Now drop the two indexes you do not need and run the same insert again:

```sql
DROP INDEX users_country_idx;
DROP INDEX users_lower_email_idx;

INSERT INTO users (email, country, created_at)
SELECT 'later' || n || '@example.com', 'US', now()
FROM generate_series(1, 200000) AS n;
```

Noticeably faster. Fewer indexes, cheaper writes. On a table that takes
thousands of writes a second, every extra index is a tax on every one
of them.

On a real system, find the indexes nobody uses:

```sql
SELECT indexrelname, idx_scan,
       pg_size_pretty(pg_relation_size(indexrelid)) AS size
FROM pg_stat_user_indexes
WHERE relname = 'users'
ORDER BY idx_scan;
```

`idx_scan` counts how many times each index has been used since the
statistics were last reset. An index with `0` scans on a busy table is
usually safe to drop, with two exceptions: check replicas and rare
monthly reports first, and never drop a primary key or unique index
because of this number. Those enforce rules about your data, not just
speed.

### Clean up

```sql
DROP TABLE users;
\q
```

If you used Docker, `docker stop pgbook-lab` removes the container too.

### Your turn

- [ ] Compare the size of the table with the size of each index.
- [ ] Time the insert with all your indexes, then with two fewer.
- [ ] Run the unused-index query and read the `idx_scan` column.
- [ ] Clean up.

## What you learned

- Without an index, a `WHERE` clause means reading every row: a
  `Seq Scan`.
- `EXPLAIN ANALYZE` shows which plan Postgres chose and how long it
  took. Read it before and after every index you add.
- A B-tree index serves equality, ranges, ordering, and prefixes.
- An index does not help when the column is wrapped in a function, the
  pattern starts with `%`, or the condition matches a big slice of the
  table.
- Every index costs disk and slows every write. Add the ones your
  queries need, and check `pg_stat_user_indexes` for the ones nobody
  uses.

Run `pgbook next` to continue with transactions.
