-- +goose Up
-- Global indexer options (like Radarr/Sonarr)

CREATE TABLE IF NOT EXISTS indexer_options (
    id INTEGER PRIMARY KEY CHECK (id = 1),  -- Only one row allowed
    minimum_age INTEGER NOT NULL DEFAULT 0,           -- Minutes (Usenet: min age before grab)
    retention INTEGER NOT NULL DEFAULT 0,             -- Days (Usenet: 0 = unlimited)
    maximum_size INTEGER NOT NULL DEFAULT 0,          -- MB (0 = unlimited)
    rss_sync_interval INTEGER NOT NULL DEFAULT 30,    -- Minutes (0 = disabled)
    prefer_indexer_flags INTEGER NOT NULL DEFAULT 0,  -- Boolean
    availability_delay INTEGER NOT NULL DEFAULT 0,    -- Days
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Insert default row
INSERT INTO indexer_options (id) VALUES (1);

-- +goose Down
DROP TABLE IF EXISTS indexer_options;
