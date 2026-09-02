---
slug: window-functions
title: Window functions
description: Running totals without collapsing rows
level: intermediate
reading_minutes: 9
order: 4
aliases: windows, window, over, partition-by
tags: sql, analytics, aggregates
---

## Aggregates that keep the rows

`GROUP BY` collapses rows into one row per group. A window function
computes the same kind of aggregate but leaves every row in place:

```sql
CREATE TABLE payments (
  id       bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id  int  NOT NULL,
  amount   numeric NOT NULL,
  paid_at  date NOT NULL
);

INSERT INTO payments (user_id, amount, paid_at) VALUES
  (1, 50, '2026-01-01'), (1, 30, '2026-01-05'), (1, 20, '2026-02-01'),
  (2, 90, '2026-01-02'), (2, 10, '2026-01-20');
```

```sql
SELECT user_id, paid_at, amount,
       sum(amount) OVER (PARTITION BY user_id ORDER BY paid_at)
         AS running_total
FROM payments;
```

Each row shows its own amount *and* the running total for that user so
far. No self-join, no subquery.

## The three parts of OVER

- `PARTITION BY user_id` — restart the calculation per user
- `ORDER BY paid_at` — the order rows accumulate in
- frame (optional) — which neighboring rows are visible

Leave `PARTITION BY` out to compute over the whole result; leave
`ORDER BY` out to aggregate the entire partition on every row.

## Ranking

```sql
SELECT user_id, amount,
       row_number() OVER w,
       rank()       OVER w,
       dense_rank() OVER w
FROM payments
WINDOW w AS (PARTITION BY user_id ORDER BY amount DESC);
```

`row_number` always counts 1, 2, 3…; `rank` leaves gaps after ties;
`dense_rank` does not. The `WINDOW` clause names a definition so you
write it once.

## Top-N per group

The classic use — latest payment per user:

```sql
SELECT * FROM (
  SELECT p.*,
         row_number() OVER (PARTITION BY user_id ORDER BY paid_at DESC) AS rn
  FROM payments p
) ranked
WHERE rn = 1;
```

## Looking at neighbors

`lag` and `lead` read other rows of the partition:

```sql
SELECT user_id, paid_at, amount,
       amount - lag(amount) OVER (PARTITION BY user_id ORDER BY paid_at)
         AS change_from_previous
FROM payments;
```

> Note: window functions run after `WHERE` and `GROUP BY`. To filter on
> a window result (like `rn = 1` above), wrap it in a subquery or CTE —
> `WHERE rn = 1` directly is an error.
