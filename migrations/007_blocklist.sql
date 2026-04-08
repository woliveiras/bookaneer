-- +goose Up
-- Blocklist table for releases that should not be downloaded again

CREATE TABLE IF NOT EXISTS blocklist (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id         INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    source_title    TEXT    NOT NULL,
    quality         TEXT    NOT NULL DEFAULT '',
    reason          TEXT    NOT NULL DEFAULT '',
    date            TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX idx_blocklist_book_id ON blocklist(book_id);
CREATE INDEX idx_blocklist_source_title ON blocklist(source_title);

-- +goose Down
DROP INDEX IF EXISTS idx_blocklist_source_title;
DROP INDEX IF EXISTS idx_blocklist_book_id;
DROP TABLE IF EXISTS blocklist;
