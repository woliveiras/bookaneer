-- +goose Up
-- Add save_path column to download_queue to persist the downloaded file path
-- This is needed to properly import files after server restarts

ALTER TABLE download_queue ADD COLUMN save_path TEXT NOT NULL DEFAULT '';

-- +goose Down
-- SQLite doesn't support DROP COLUMN easily, so we recreate the table
CREATE TABLE download_queue_new (
    id                 INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id            INTEGER NOT NULL,
    download_client_id INTEGER,
    indexer_id         INTEGER,
    external_id        TEXT    NOT NULL DEFAULT '',
    title              TEXT    NOT NULL,
    size               INTEGER NOT NULL DEFAULT 0,
    format             TEXT    NOT NULL DEFAULT 'unknown',
    status             TEXT    NOT NULL DEFAULT 'queued',
    progress           REAL    NOT NULL DEFAULT 0.0,
    download_url       TEXT    NOT NULL DEFAULT '',
    added_at           TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
    FOREIGN KEY (download_client_id) REFERENCES download_clients(id) ON DELETE SET NULL,
    FOREIGN KEY (indexer_id) REFERENCES indexers(id) ON DELETE SET NULL
);

INSERT INTO download_queue_new 
SELECT id, book_id, download_client_id, indexer_id, external_id, title, size, format, status, progress, download_url, added_at
FROM download_queue;

DROP TABLE download_queue;
ALTER TABLE download_queue_new RENAME TO download_queue;
