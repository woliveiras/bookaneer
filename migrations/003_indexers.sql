-- +goose Up
-- Indexer configuration (Newznab/Torznab)

CREATE TABLE IF NOT EXISTS indexers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('newznab', 'torznab')),
    base_url TEXT NOT NULL,
    api_key TEXT NOT NULL DEFAULT '',
    categories TEXT NOT NULL DEFAULT '',  -- Comma-separated category IDs
    priority INTEGER NOT NULL DEFAULT 50, -- Lower = higher priority (1-100)
    enabled INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    
    UNIQUE(name)
);

-- Search history for caching/analytics
CREATE TABLE IF NOT EXISTS search_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    query TEXT NOT NULL,
    author TEXT NOT NULL DEFAULT '',
    title TEXT NOT NULL DEFAULT '',
    results_count INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_search_history_created ON search_history(created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_search_history_created;
DROP TABLE IF EXISTS search_history;
DROP TABLE IF EXISTS indexers;
