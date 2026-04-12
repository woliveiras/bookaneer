-- Remote Path Mappings — translate download client paths to local paths
-- +goose Up

CREATE TABLE remote_path_mappings (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    host        TEXT    NOT NULL DEFAULT '',       -- download client host (for display/organization)
    remote_path TEXT    NOT NULL,                  -- path as seen by the download client
    local_path  TEXT    NOT NULL,                  -- path as seen by Bookaneer
    created_at  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

-- +goose Down

DROP TABLE IF EXISTS remote_path_mappings;
