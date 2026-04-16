-- +goose Up
-- Remove automatic search infrastructure.
-- Search is now always manual: user triggers a search, receives a list of results
-- sorted by file size (largest first), and chooses which release to download.
-- Books not found during search can be saved to the wanted list by the user.

-- Drop the automatic search flag from indexers.
-- All enabled indexers are now always available for interactive (manual) search.
ALTER TABLE indexers DROP COLUMN enable_automatic_search;

-- Remove scheduled tasks that triggered automatic searches.
-- DownloadMonitor and other system tasks are kept.
DELETE FROM scheduled_tasks WHERE name IN ('RssSync', 'MissingBookSearch');

-- +goose Down
-- Restore automatic search column with default disabled.
ALTER TABLE indexers ADD COLUMN enable_automatic_search INTEGER NOT NULL DEFAULT 0;

-- Restore scheduled tasks (intervals in seconds).
INSERT OR IGNORE INTO scheduled_tasks (name, interval_seconds, next_run_at)
VALUES
    ('RssSync', 900, datetime('now')),
    ('MissingBookSearch', 86400, datetime('now'));
