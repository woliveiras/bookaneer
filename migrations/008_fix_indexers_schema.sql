-- +goose Up
-- Fix missing columns in indexers table (from inconsistent migration state)

-- Add missing columns if they don't exist
-- SQLite doesn't support IF NOT EXISTS for ALTER TABLE, so we use a workaround

-- Check and add base_url
CREATE TABLE IF NOT EXISTS _indexers_check (x INT);
DROP TABLE _indexers_check;

-- Using PRAGMA to check columns is complex, so we just try to add and ignore errors
-- This migration assumes the columns might be missing

-- The safest approach: recreate the table with correct schema
ALTER TABLE indexers RENAME TO indexers_old;

CREATE TABLE indexers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('newznab', 'torznab')),
    base_url TEXT NOT NULL DEFAULT '',
    api_path TEXT NOT NULL DEFAULT '/api',
    api_key TEXT NOT NULL DEFAULT '',
    categories TEXT NOT NULL DEFAULT '',
    priority INTEGER NOT NULL DEFAULT 50,
    enabled INTEGER NOT NULL DEFAULT 1,
    enable_rss INTEGER NOT NULL DEFAULT 1,
    enable_automatic_search INTEGER NOT NULL DEFAULT 1,
    enable_interactive_search INTEGER NOT NULL DEFAULT 1,
    additional_parameters TEXT NOT NULL DEFAULT '',
    minimum_seeders INTEGER NOT NULL DEFAULT 1,
    seed_ratio REAL,
    seed_time INTEGER,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(name)
);

-- Migrate data from old table, extracting from JSON settings if needed
INSERT INTO indexers (id, name, type, base_url, api_key, categories, priority, enabled, api_path, enable_rss, enable_automatic_search, enable_interactive_search, additional_parameters, minimum_seeders, seed_ratio, seed_time)
SELECT 
    id, 
    name, 
    type,
    COALESCE(json_extract(settings, '$.baseUrl'), '') as base_url,
    COALESCE(json_extract(settings, '$.apiKey'), '') as api_key,
    COALESCE(json_extract(settings, '$.categories'), '') as categories,
    priority,
    enabled,
    COALESCE(api_path, '/api') as api_path,
    COALESCE(enable_rss, 1) as enable_rss,
    COALESCE(enable_automatic_search, 1) as enable_automatic_search,
    COALESCE(enable_interactive_search, 1) as enable_interactive_search,
    COALESCE(additional_parameters, '') as additional_parameters,
    COALESCE(minimum_seeders, 1) as minimum_seeders,
    seed_ratio,
    seed_time
FROM indexers_old;

DROP TABLE indexers_old;

-- +goose Down
-- This is a one-way migration, going back would lose data
