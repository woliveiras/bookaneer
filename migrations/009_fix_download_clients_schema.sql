-- +goose Up
-- Fix download_clients table schema: migrate from JSON settings to individual columns
-- This is needed because migration 001 created the table with JSON settings,
-- and migration 006 tried to create it with individual columns using IF NOT EXISTS

-- Step 1: Rename the old table
ALTER TABLE download_clients RENAME TO download_clients_old;

-- Step 2: Create the table with proper schema
CREATE TABLE download_clients (
    id                     INTEGER PRIMARY KEY AUTOINCREMENT,
    name                   TEXT    NOT NULL UNIQUE,
    type                   TEXT    NOT NULL,  -- sabnzbd, qbittorrent, transmission, blackhole, direct
    host                   TEXT    NOT NULL DEFAULT '',
    port                   INTEGER NOT NULL DEFAULT 0,
    use_tls                INTEGER NOT NULL DEFAULT 0,
    username               TEXT    NOT NULL DEFAULT '',
    password               TEXT    NOT NULL DEFAULT '',
    api_key                TEXT    NOT NULL DEFAULT '',
    category               TEXT    NOT NULL DEFAULT '',
    recent_priority        INTEGER NOT NULL DEFAULT 0,
    older_priority         INTEGER NOT NULL DEFAULT 0,
    remove_completed_after INTEGER NOT NULL DEFAULT 0, -- minutes, 0 = never
    enabled                INTEGER NOT NULL DEFAULT 1,
    priority               INTEGER NOT NULL DEFAULT 1,  -- client priority, lower = preferred
    nzb_folder             TEXT    NOT NULL DEFAULT '', -- blackhole
    torrent_folder         TEXT    NOT NULL DEFAULT '', -- blackhole
    watch_folder           TEXT    NOT NULL DEFAULT '', -- blackhole
    download_dir           TEXT    NOT NULL DEFAULT '', -- direct downloader output path
    created_at             TEXT    NOT NULL DEFAULT (datetime('now')),
    updated_at             TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- Step 3: Migrate data from old table if any exists
-- Extract settings from JSON and insert into new schema
INSERT INTO download_clients (id, name, type, host, port, use_tls, username, password, api_key, category, enabled, priority)
SELECT 
    id,
    name,
    type,
    COALESCE(json_extract(settings, '$.host'), ''),
    COALESCE(json_extract(settings, '$.port'), 0),
    COALESCE(json_extract(settings, '$.useTls'), 0),
    COALESCE(json_extract(settings, '$.username'), ''),
    COALESCE(json_extract(settings, '$.password'), ''),
    COALESCE(json_extract(settings, '$.apiKey'), ''),
    COALESCE(json_extract(settings, '$.category'), ''),
    enabled,
    priority
FROM download_clients_old;

-- Step 4: Drop the old table
DROP TABLE download_clients_old;

-- +goose Down
-- Reverting to the old schema would lose data, so we just leave it
-- This is intentionally empty to prevent accidental rollback
