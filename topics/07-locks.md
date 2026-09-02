---
slug: locks
title: Locks
description: Why a query is stuck, not slow
level: intermediate
reading_minutes: 10
order: 7
aliases: locking, lock, blocking
tags: concurrency, transactions, blocking
---

## What Postgres locks

A query that is "slow" every time has a plan problem. A query that is
usually instant but sometimes hangs has a lock problem. Locks are how
Postgres stops two transactions from changing the same thing at once —
reads never block reads, but writes serialize.

## Row locks vs table locks

Ordinary DML takes light table locks plus row locks on the rows it
changes:

```sql
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
-- row 1 is now locked until COMMIT or ROLLBACK
```

DDL is the dangerous one: `ALTER TABLE` takes an `ACCESS EXCLUSIVE`
lock that conflicts with *everything*, including `SELECT`.

## SELECT ... FOR UPDATE

Lock rows you are about to update, so the read and the write are one
atomic step:

```sql
BEGIN;
SELECT * FROM accounts WHERE id = 1 FOR UPDATE;
-- no other transaction can change or lock this row now
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
COMMIT;
```

## How blocking happens

Open two sessions:

```sql
-- session A
BEGIN;
UPDATE accounts SET balance = 0 WHERE id = 1;

-- session B (hangs until A commits or rolls back)
UPDATE accounts SET balance = 1 WHERE id = 1;
```

B is not slow — it is waiting. The most common real-world shape: a
transaction updates a row, then does something long (an API call, user
input) before committing, and every other writer of that row queues up
behind it.

## Finding the blocking query

```sql
SELECT blocked.pid            AS blocked_pid,
       blocked.query          AS blocked_query,
       blocking.pid           AS blocking_pid,
       blocking.query         AS blocking_query,
       blocking.state         AS blocking_state
FROM pg_stat_activity blocked
JOIN pg_stat_activity blocking
  ON blocking.pid = ANY(pg_blocking_pids(blocked.pid));
```

A blocker whose state is `idle in transaction` is the smoking gun: it
holds locks but is running nothing. End it with
`SELECT pg_terminate_backend(pid)`.

## Deadlocks

If A waits for B while B waits for A, nobody can ever proceed.
Postgres detects the cycle after `deadlock_timeout` (1s by default) and
kills one transaction with error `40P01`. The fix is ordering: make
every transaction lock rows in the same order (for example, always by
ascending id).

## Lock timeouts

Never queue forever. Set a ceiling per session, per transaction, or per
role:

```sql
SET lock_timeout = '2s';
```

Migrations deserve one always — an `ALTER TABLE` waiting for a lock
also *blocks everyone behind it* while it waits:

```sql
SET lock_timeout = '5s';
ALTER TABLE accounts ADD COLUMN note text;  -- fails fast instead of stalling the app
```

## Safe locking patterns

- Keep transactions short; never hold one across network calls
- `SET lock_timeout` in migrations, and retry on failure
- Lock rows in a consistent order to prevent deadlocks
- Use `SELECT ... FOR UPDATE SKIP LOCKED` for job queues:

```sql
SELECT id FROM jobs
WHERE status = 'pending'
ORDER BY id
FOR UPDATE SKIP LOCKED
LIMIT 1;
```

Workers grab different jobs without ever waiting on each other.
