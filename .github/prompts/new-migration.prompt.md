---
description: "Create a new goose SQL migration file. Use when adding or modifying database tables, indexes, or constraints."
agent: "agent"
argument-hint: "Migration purpose (e.g., 'add reading_stats table')"
---

Create a new goose SQL migration for Bookaneer.

## Steps

1. **Determine the next migration number**: Check `migrations/` for the highest numbered file and increment by 1
2. **Create the file** at `migrations/{NNN}_{snake_case_name}.sql`

## Template

```sql
-- +goose Up
-- SQL statements here

-- +goose Down
-- Reverse the Up statements
```

## Rules

Follow [SQLite instructions](../instructions/sqlite.instructions.md):

- `INTEGER PRIMARY KEY AUTOINCREMENT` for domain entity IDs
- `TEXT` for timestamps, default `strftime('%Y-%m-%dT%H:%M:%SZ', 'now')`
- `INTEGER` (0/1) for booleans with `NOT NULL DEFAULT 0`
- `TEXT DEFAULT '{}'` for JSON fields
- Always define foreign keys with `ON DELETE CASCADE` or `ON DELETE SET NULL`
- Create indexes for columns used in WHERE, JOIN, ORDER BY
- Down migration must drop exactly what Up created
- Never modify an existing migration — always create a new one
