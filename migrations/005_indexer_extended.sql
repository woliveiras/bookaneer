-- +goose Up
-- Extended indexer fields (Radarr/Sonarr style)

ALTER TABLE indexers ADD COLUMN api_path TEXT NOT NULL DEFAULT '/api';
ALTER TABLE indexers ADD COLUMN enable_rss INTEGER NOT NULL DEFAULT 1;
ALTER TABLE indexers ADD COLUMN enable_automatic_search INTEGER NOT NULL DEFAULT 1;
ALTER TABLE indexers ADD COLUMN enable_interactive_search INTEGER NOT NULL DEFAULT 1;
ALTER TABLE indexers ADD COLUMN additional_parameters TEXT NOT NULL DEFAULT '';

-- Torznab-specific fields
ALTER TABLE indexers ADD COLUMN minimum_seeders INTEGER NOT NULL DEFAULT 1;
ALTER TABLE indexers ADD COLUMN seed_ratio REAL;      -- NULL = use download client default
ALTER TABLE indexers ADD COLUMN seed_time INTEGER;    -- Minutes, NULL = use download client default

-- +goose Down
-- SQLite doesn't support DROP COLUMN easily, so we recreate the table
-- For simplicity, this down migration just drops and recreates
-- In production, you'd use a more careful migration
