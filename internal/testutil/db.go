package testutil

import (
	"database/sql"
	"testing"

	bookaneer "github.com/woliveiras/bookaneer"

	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

// OpenTestDB creates an in-memory SQLite database with all migrations applied.
// The database is closed automatically when the test finishes.
func OpenTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Use legacy alter table behavior during migrations so ALTER TABLE RENAME
	// doesn't rewrite FK references in other tables' schemas (SQLite 3.25+).
	if _, err := db.Exec("PRAGMA legacy_alter_table = ON"); err != nil {
		t.Fatalf("set legacy_alter_table: %v", err)
	}

	// Run migrations with FK disabled so ALTER TABLE RENAME doesn't break FK refs.
	goose.SetBaseFS(bookaneer.MigrationsFS)
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("set dialect: %v", err)
	}
	if err := goose.Up(db, "migrations"); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	// Restore normal alter table behavior and enable foreign keys.
	if _, err := db.Exec("PRAGMA legacy_alter_table = OFF"); err != nil {
		t.Fatalf("reset legacy_alter_table: %v", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

	return db
}

// SeedAuthor inserts a test author and returns its ID.
func SeedAuthor(t *testing.T, db *sql.DB, name string) int64 {
	t.Helper()
	result, err := db.Exec(`
		INSERT INTO authors (name, sort_name, monitored, path)
		VALUES (?, ?, 1, ?)`,
		name, name, "/library/"+name)
	if err != nil {
		t.Fatalf("seed author: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}

// SeedBook inserts a test book and returns its ID.
func SeedBook(t *testing.T, db *sql.DB, authorID int64, title string) int64 {
	t.Helper()
	result, err := db.Exec(`
		INSERT INTO books (author_id, title, sort_title, monitored)
		VALUES (?, ?, ?, 1)`,
		authorID, title, title)
	if err != nil {
		t.Fatalf("seed book: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}

// SeedRootFolder inserts a test root folder and returns its ID.
func SeedRootFolder(t *testing.T, db *sql.DB, path, name string) int64 {
	t.Helper()
	result, err := db.Exec(`
		INSERT INTO root_folders (path, name)
		VALUES (?, ?)`,
		path, name)
	if err != nil {
		t.Fatalf("seed root folder: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}

// SeedQueueItem inserts a download queue item and returns its ID.
func SeedQueueItem(t *testing.T, db *sql.DB, bookID int64, title, status string) int64 {
	t.Helper()
	result, err := db.Exec(`
		INSERT INTO download_queue (book_id, title, size, format, status, download_url, external_id)
		VALUES (?, ?, 1024, 'epub', ?, 'https://example.com/book.epub', 'ext-123')`,
		bookID, title, status)
	if err != nil {
		t.Fatalf("seed queue item: %v", err)
	}
	id, _ := result.LastInsertId()
	return id
}
