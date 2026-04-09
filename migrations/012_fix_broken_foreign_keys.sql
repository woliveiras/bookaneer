-- +goose Up
-- Fix broken foreign key references in download_queue and grabs tables.
-- Migrations 008/009 renamed indexers/download_clients to _old variants,
-- which caused SQLite to update FK references in child tables automatically.
-- After the _old tables were dropped, the FK references became invalid,
-- breaking all INSERTs when PRAGMA foreign_keys = ON.

-- Must disable FK checks during table recreation to avoid constraint errors
PRAGMA foreign_keys = OFF;

-- === Fix download_queue ===

CREATE TABLE download_queue_new (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id             INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    download_client_id  INTEGER REFERENCES download_clients(id) ON DELETE SET NULL,
    indexer_id          INTEGER REFERENCES indexers(id) ON DELETE SET NULL,
    external_id         TEXT    NOT NULL DEFAULT '',
    title               TEXT    NOT NULL,
    size                INTEGER NOT NULL DEFAULT 0,
    format              TEXT    NOT NULL DEFAULT 'unknown',
    status              TEXT    NOT NULL DEFAULT 'queued',
    progress            REAL    NOT NULL DEFAULT 0.0,
    download_url        TEXT    NOT NULL DEFAULT '',
    added_at            TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    save_path           TEXT    NOT NULL DEFAULT ''
);

INSERT INTO download_queue_new
    SELECT id, book_id, download_client_id, indexer_id, external_id,
           title, size, format, status, progress, download_url, added_at, save_path
    FROM download_queue;

DROP TABLE download_queue;
ALTER TABLE download_queue_new RENAME TO download_queue;

CREATE INDEX idx_download_queue_book_id ON download_queue(book_id);
CREATE INDEX idx_download_queue_status ON download_queue(status);

-- === Fix grabs ===

CREATE TABLE grabs_new (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id       INTEGER NOT NULL,
    indexer_id    INTEGER NOT NULL,
    release_title TEXT    NOT NULL,
    download_url  TEXT    NOT NULL,
    size          INTEGER NOT NULL DEFAULT 0,
    quality       TEXT    NOT NULL DEFAULT '',
    client_id     INTEGER NOT NULL,
    download_id   TEXT    NOT NULL DEFAULT '',
    status        TEXT    NOT NULL DEFAULT 'pending',
    error_message TEXT    NOT NULL DEFAULT '',
    grabbed_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    completed_at  TEXT,
    FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
    FOREIGN KEY (indexer_id) REFERENCES indexers(id) ON DELETE SET NULL,
    FOREIGN KEY (client_id) REFERENCES download_clients(id) ON DELETE SET NULL
);

INSERT INTO grabs_new
    SELECT id, book_id, indexer_id, release_title, download_url,
           size, quality, client_id, download_id, status, error_message,
           grabbed_at, completed_at
    FROM grabs;

DROP TABLE grabs;
ALTER TABLE grabs_new RENAME TO grabs;

-- Re-enable FK checks
PRAGMA foreign_keys = ON;

-- +goose Down
-- One-way migration: the fixed FK references are strictly correct.
