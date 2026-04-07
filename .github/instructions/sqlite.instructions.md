---
description: "Use when writing SQL queries, designing schema, working with SQLite, creating migrations, optimizing queries, or configuring the database. Covers SQLite WAL mode, parameterized queries, pure Go driver (modernc), goose migrations, and performance patterns."
applyTo: "**/*.sql, internal/database/**/*.go, migrations/**"
---

# SQLite Best Practices

## Driver: modernc.org/sqlite

- Pure Go, zero CGo — cross-compiles to any platform without C toolchain
- Import as `_ "modernc.org/sqlite"` and use via `database/sql`
- Connection string: `file:/path/to/db.db?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON`

## Connection Setup

```go
func OpenDB(path string) (*sql.DB, error) {
    dsn := fmt.Sprintf("file:%s?_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=ON&_synchronous=NORMAL", path)
    db, err := sql.Open("sqlite", dsn)
    if err != nil {
        return nil, fmt.Errorf("open database: %w", err)
    }
    // SQLite performs best with a single writer connection
    db.SetMaxOpenConns(1)
    // Enable WAL checkpoint on close
    db.SetConnMaxLifetime(0)
    return db, nil
}
```

Critical pragmas (set via DSN or after open):
- `_journal_mode=WAL` — Write-Ahead Logging for concurrent reads
- `_busy_timeout=5000` — Wait up to 5s on lock instead of immediate SQLITE_BUSY
- `_foreign_keys=ON` — Enforce foreign key constraints (off by default!)
- `_synchronous=NORMAL` — Good balance of safety and performance with WAL

## WAL Mode

- WAL allows concurrent reads while writing — essential for a web application
- Only ONE writer at a time — `SetMaxOpenConns(1)` prevents write contention
- Readers never block writers, writers never block readers
- WAL file (`*.db-wal`) and shared memory file (`*.db-shm`) are created alongside the DB
- Back up all three files together: `*.db`, `*.db-wal`, `*.db-shm`

## Parameterized Queries — ALWAYS

```go
// CORRECT — parameterized
row := db.QueryRowContext(ctx, "SELECT title FROM books WHERE id = ?", id)

// CORRECT — multiple parameters
rows, err := db.QueryContext(ctx,
    "SELECT id, title FROM books WHERE author_id = ? AND monitored = ? ORDER BY title LIMIT ? OFFSET ?",
    authorID, 1, limit, offset,
)

// NEVER — string concatenation (SQL injection)
db.Query("SELECT * FROM books WHERE id = " + id)
db.Query(fmt.Sprintf("SELECT * FROM books WHERE title = '%s'", title))
```

## Schema Conventions

- Primary keys: `INTEGER PRIMARY KEY AUTOINCREMENT` for domain entities
- Text primary keys: ULID for commands (sortable, unique, no sequences)
- Timestamps: `TEXT` in ISO 8601 format (`strftime('%Y-%m-%dT%H:%M:%SZ', 'now')`)
- Booleans: `INTEGER` (0/1) with `NOT NULL DEFAULT 0`
- JSON fields: `TEXT` with `DEFAULT '{}'` for flexible settings
- Indexes: create for columns used in WHERE, JOIN, ORDER BY
- Foreign keys: always define with `ON DELETE CASCADE` or `ON DELETE SET NULL`
- Use `IF NOT EXISTS` in production schema, but not in migrations (migrations should be explicit)

## Migrations (goose)

- SQL-based migrations in `migrations/` directory
- File naming: `001_initial_schema.sql`, `002_add_reading_progress.sql`
- Each file has `-- +goose Up` and `-- +goose Down` sections
- Down migrations must be reversible — drop what Up created
- Never modify an existing migration after it's applied — create a new one
- Run migrations on app startup, not as a separate step

```go
// Run migrations programmatically on startup
func RunMigrations(db *sql.DB, dir string) error {
    goose.SetDialect("sqlite3")
    return goose.Up(db, dir)
}
```

## Query Patterns

### Pagination
```sql
SELECT id, title, author_id
FROM books
WHERE monitored = 1
ORDER BY title ASC
LIMIT ? OFFSET ?
```
Always ORDER BY before LIMIT/OFFSET for consistent results.

### Count with pagination
```sql
SELECT COUNT(*) FROM books WHERE monitored = 1
```
Run as a separate query before the paginated query for total count.

### Upsert
```sql
INSERT INTO config (key, value) VALUES (?, ?)
ON CONFLICT(key) DO UPDATE SET value = excluded.value
```

### JSON fields
```sql
-- Read JSON field
SELECT json_extract(settings, '$.baseUrl') FROM indexers WHERE id = ?

-- Query inside JSON
SELECT * FROM indexers WHERE json_extract(settings, '$.enabled') = 1
```

## Performance

- SQLite is fast enough for <10k books — don't over-optimize
- Use `EXPLAIN QUERY PLAN` to verify index usage on slow queries
- Batch inserts in a transaction for bulk operations (library scan):
  ```go
  tx, _ := db.BeginTx(ctx, nil)
  defer tx.Rollback()
  stmt, _ := tx.PrepareContext(ctx, "INSERT INTO books (...) VALUES (?, ?, ?)")
  for _, book := range books {
      stmt.ExecContext(ctx, book.Title, book.AuthorID, book.ISBN)
  }
  tx.Commit()
  ```
- Use prepared statements for repeated queries in hot paths
- `VACUUM` periodically (during backup or maintenance task) to reclaim space

## Backup

SQLite's online backup API via the `VACUUM INTO` command:

```sql
VACUUM INTO '/data/backups/bookaneer-2026-04-07.db'
```

This creates a consistent backup without stopping the application. The backup file is a complete, standalone database.

## Testing

- Use in-memory database for unit tests: `file::memory:?cache=shared`
- Apply migrations before each test suite
- Use `t.Cleanup()` to close DB after tests
- Each test should set up its own data — no shared state between tests

```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite", "file::memory:?cache=shared&_foreign_keys=ON")
    require.NoError(t, err)
    t.Cleanup(func() { db.Close() })
    // Run migrations
    require.NoError(t, goose.Up(db, "../../migrations"))
    return db
}
```

## Anti-patterns

- Multiple writer connections (`SetMaxOpenConns > 1`) — causes SQLITE_BUSY
- Missing `_foreign_keys=ON` — FK constraints silently not enforced
- Missing `_busy_timeout` — immediate SQLITE_BUSY errors under concurrent access
- String concatenation in queries — SQL injection
- Storing binary data (BLOBs) in SQLite — store file paths instead
- Using SQLite as a cache with high write throughput — use in-memory cache
- Not backing up WAL and SHM files alongside the main DB file
- Auto-incrementing IDs for distributed systems — use ULID/UUID if needed
