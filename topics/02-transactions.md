---
slug: transactions
title: Transactions & isolation
description: What one query can see of another
level: intermediate
reading_minutes: 10
order: 2
aliases: isolation, transactions-and-isolation, transaction
tags: concurrency, mvcc, isolation
---

## Grouping statements safely

A transaction makes a group of statements atomic: either all of them
commit, or none do.

```sql
BEGIN;
UPDATE accounts SET balance = balance - 100 WHERE id = 1;
UPDATE accounts SET balance = balance + 100 WHERE id = 2;
COMMIT;
```

If the connection drops between the two updates, Postgres rolls the
first one back. Money never disappears.

## Every statement is a transaction

Without `BEGIN`, each statement runs in its own tiny transaction that
commits immediately. `BEGIN` just stretches that boundary across
several statements.

## What one transaction sees of another

Open two `psql` sessions and try it:

```sql
-- session A
BEGIN;
UPDATE accounts SET balance = 0 WHERE id = 1;

-- session B
SELECT balance FROM accounts WHERE id = 1;  -- still the old value
```

Session B does not see A's uncommitted change. Postgres never shows
dirty (uncommitted) data, at any isolation level.

## Read Committed — the default

Under `READ COMMITTED`, every *statement* sees a fresh snapshot: data
committed before the statement began. Two identical `SELECT`s inside
one transaction can return different rows if another transaction
commits in between. That surprise is called a non-repeatable read.

## Repeatable Read

```sql
BEGIN ISOLATION LEVEL REPEATABLE READ;
```

Now the *transaction* takes one snapshot at its first query and keeps
it. Reports that read many tables consistently belong here. The cost:
if you try to update a row that someone else changed after your
snapshot, Postgres aborts you with a serialization error — retry the
transaction.

## Serializable

```sql
BEGIN ISOLATION LEVEL SERIALIZABLE;
```

The strictest level: Postgres guarantees the outcome equals *some*
serial order of the transactions, detecting write patterns the lower
levels miss. Any serializable transaction can fail with SQLSTATE
`40001`; your application must be ready to retry.

> Note: retries are not an edge case at higher isolation levels — they
> are the contract. Wrap the transaction in a retry loop.

## Choosing a level

- `READ COMMITTED` — the default; fine for most OLTP writes
- `REPEATABLE READ` — multi-statement reads that must be consistent
- `SERIALIZABLE` — invariants spanning multiple rows or tables

Whatever the level, keep transactions short. A transaction held open
for minutes blocks vacuum and can bloat every table it touched.
