---
slug: row-level-security
title: Row-level security
description: Access control inside the database
level: intermediate
reading_minutes: 10
order: 5
aliases: rls, row-security, policies
tags: security, multi-tenant, policies
---

## The database enforces who sees what

Row-level security (RLS) attaches a filter to a table that every query
must pass through. Instead of trusting each application query to add
`WHERE tenant_id = ...`, the database adds it for you — and there is no
way to forget.

```sql
CREATE TABLE documents (
  id        bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  tenant_id int  NOT NULL,
  title     text NOT NULL
);

INSERT INTO documents (tenant_id, title) VALUES
  (1, 'tenant one doc'), (2, 'tenant two doc');
```

## Turning it on

```sql
ALTER TABLE documents ENABLE ROW LEVEL SECURITY;

CREATE POLICY tenant_isolation ON documents
  USING (tenant_id = current_setting('app.tenant_id')::int);
```

With RLS enabled and no matching policy, a table is empty for
non-owners — RLS fails closed.

## Trying it

```sql
CREATE ROLE app_user LOGIN;
GRANT SELECT, INSERT, UPDATE, DELETE ON documents TO app_user;

SET ROLE app_user;
SET app.tenant_id = '1';
SELECT * FROM documents;   -- only tenant one's rows

SET app.tenant_id = '2';
SELECT * FROM documents;   -- only tenant two's rows
```

The application sets `app.tenant_id` once per connection (or per
transaction with `SET LOCAL`), and every query is scoped automatically.

## USING vs WITH CHECK

- `USING (...)` — which existing rows are visible (SELECT, UPDATE,
  DELETE)
- `WITH CHECK (...)` — which new row versions are allowed (INSERT,
  UPDATE)

```sql
CREATE POLICY tenant_writes ON documents FOR INSERT
  WITH CHECK (tenant_id = current_setting('app.tenant_id')::int);
```

Without `WITH CHECK`, a tenant could insert rows *into another
tenant's* view of the table while never being able to read them back.

## Who bypasses RLS

Superusers and roles with `BYPASSRLS` skip policies entirely. Table
owners do too, unless you force it:

```sql
ALTER TABLE documents FORCE ROW LEVEL SECURITY;
```

> Note: your migration/admin role bypassing RLS is a feature; your
> application role must never own the tables it queries.

## The cost

Policies are just predicates — they are planned like any `WHERE`
clause, and indexes on the policy columns (`tenant_id` here) apply.
Keep policies simple and indexed and the overhead is negligible.
