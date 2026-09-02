---
slug: replication
title: Replication
description: Read replicas and failover
level: advanced
reading_minutes: 11
order: 8
aliases: replicas, read-replicas, failover, standby
tags: wal, high-availability, streaming
---

## One log, many copies

Every change in Postgres is first written to the write-ahead log (WAL).
Replication is simply shipping that log to another server that replays
it. The replica is a byte-for-byte copy, seconds (usually milliseconds)
behind.

## Setting up a streaming replica

On the primary, create a replication role and let it connect:

```sql
CREATE ROLE replicator WITH REPLICATION LOGIN PASSWORD 'secret';
```

```text
# pg_hba.conf on the primary
host replication replicator 10.0.0.2/32 scram-sha-256
```

On the replica machine, clone the primary — this is the whole setup:

```text
pg_basebackup -h primary.internal -U replicator \
  -D /var/lib/postgresql/data -R -P
```

The `-R` flag writes the connection settings and a `standby.signal`
file; start Postgres and it begins streaming.

## Watching replication

On the primary:

```sql
SELECT client_addr, state, sent_lsn, replay_lsn,
       pg_wal_lsn_diff(sent_lsn, replay_lsn) AS replay_lag_bytes
FROM pg_stat_replication;
```

On the replica:

```sql
SELECT pg_is_in_recovery();                      -- true
SELECT now() - pg_last_xact_replay_timestamp()   AS replication_lag;
```

## Using a replica

A streaming replica is read-only. Point reports, analytics, and
read-heavy endpoints at it. Two caveats:

- **Lag is real.** Read-your-own-writes can fail: a user updates their
  profile on the primary and reloads the page from a replica that
  hasn't replayed it yet. Route such reads to the primary.
- **Long queries vs replay.** A long replica query can conflict with
  incoming WAL; by default Postgres cancels the query after
  `max_standby_streaming_delay`. `hot_standby_feedback = on` avoids the
  cancellations at the price of bloat on the primary.

## Replication slots

A slot makes the primary retain WAL until the replica has consumed it,
so a replica that falls behind (or reboots) never loses its place:

```sql
SELECT * FROM pg_create_physical_replication_slot('replica_1');
```

> Note: a slot for a replica that is *gone* forces the primary to keep
> WAL forever and will eventually fill the disk. Monitor
> `pg_replication_slots` for inactive slots and drop them.

## Failover

When the primary dies, promote the replica:

```text
pg_ctl promote -D /var/lib/postgresql/data
```

It stops recovery and starts accepting writes. The hard parts are
around the edges: making sure the old primary never comes back as a
second writer (fencing), and repointing clients. Tools like Patroni or
pg_auto_failover automate the dance; managed services hide it entirely.

## Sync vs async

By default replication is asynchronous — a crash can lose the last few
transactions. `synchronous_commit = on` with `synchronous_standby_names`
makes the primary wait for the replica's acknowledgment: zero data loss,
higher commit latency, and writes stall if the sync replica dies. Most
setups run async and accept seconds of exposure.
