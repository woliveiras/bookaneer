-- +goose Up
-- Bookmarks for the web reader

CREATE TABLE IF NOT EXISTS bookmarks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    book_file_id INTEGER NOT NULL REFERENCES book_files(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    position TEXT NOT NULL,          -- EPUB CFI or page number
    title TEXT NOT NULL DEFAULT '',  -- User-provided title or auto-generated
    note TEXT NOT NULL DEFAULT '',   -- Optional user note
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    
    UNIQUE(book_file_id, user_id, position)
);

CREATE INDEX idx_bookmarks_book_file_user ON bookmarks(book_file_id, user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_bookmarks_book_file_user;
DROP TABLE IF EXISTS bookmarks;
