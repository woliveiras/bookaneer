-- +goose Up
-- Download clients table
CREATE TABLE IF NOT EXISTS download_clients (
    id                    INTEGER PRIMARY KEY AUTOINCREMENT,
    name                  TEXT    NOT NULL UNIQUE,
    type                  TEXT    NOT NULL,  -- sabnzbd, qbittorrent, transmission, blackhole
    host                  TEXT    NOT NULL DEFAULT '',
    port                  INTEGER NOT NULL DEFAULT 0,
    use_tls               INTEGER NOT NULL DEFAULT 0,
    username              TEXT    NOT NULL DEFAULT '',
    password              TEXT    NOT NULL DEFAULT '',
    api_key               TEXT    NOT NULL DEFAULT '',
    category              TEXT    NOT NULL DEFAULT '',
    recent_priority       INTEGER NOT NULL DEFAULT 0,
    older_priority        INTEGER NOT NULL DEFAULT 0,
    remove_completed_after INTEGER NOT NULL DEFAULT 0, -- minutes, 0 = never
    enabled               INTEGER NOT NULL DEFAULT 1,
    priority              INTEGER NOT NULL DEFAULT 1,  -- client priority, lower = preferred
    nzb_folder            TEXT    NOT NULL DEFAULT '', -- blackhole
    torrent_folder        TEXT    NOT NULL DEFAULT '', -- blackhole
    watch_folder          TEXT    NOT NULL DEFAULT '', -- blackhole
    created_at            TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at            TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- Grabs table - tracks downloads sent to clients
CREATE TABLE IF NOT EXISTS grabs (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id       INTEGER NOT NULL,
    indexer_id    INTEGER NOT NULL,
    release_title TEXT    NOT NULL,
    download_url  TEXT    NOT NULL,
    size          INTEGER NOT NULL DEFAULT 0,
    quality       TEXT    NOT NULL DEFAULT '',
    client_id     INTEGER NOT NULL,
    download_id   TEXT    NOT NULL DEFAULT '', -- ID from download client
    status        TEXT    NOT NULL DEFAULT 'pending', -- pending, sent, downloading, completed, failed, imported
    error_message TEXT    NOT NULL DEFAULT '',
    grabbed_at    TEXT    NOT NULL DEFAULT (datetime('now')),
    completed_at  TEXT,
    FOREIGN KEY (book_id) REFERENCES books(id) ON DELETE CASCADE,
    FOREIGN KEY (indexer_id) REFERENCES indexers(id) ON DELETE SET NULL,
    FOREIGN KEY (client_id) REFERENCES download_clients(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_grabs_book_id ON grabs(book_id);
CREATE INDEX IF NOT EXISTS idx_grabs_status ON grabs(status);

-- +goose Down
DROP TABLE IF EXISTS grabs;
DROP TABLE IF EXISTS download_clients;
