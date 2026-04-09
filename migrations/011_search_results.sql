-- +goose Up
-- Table to store search results for automatic fallback when downloads fail

CREATE TABLE search_results (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id         INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    provider        TEXT NOT NULL,          -- internet-archive, libgen, anna
    title           TEXT NOT NULL,
    download_url    TEXT NOT NULL,
    format          TEXT NOT NULL,          -- epub, pdf, mobi
    size            INTEGER DEFAULT 0,
    score           REAL DEFAULT 0,         -- match score from search
    priority        INTEGER NOT NULL,       -- order to try (lower = try first)
    status          TEXT NOT NULL DEFAULT 'pending', -- pending, tried, failed, success
    error_message   TEXT DEFAULT '',
    created_at      TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    tried_at        TEXT
);

CREATE INDEX idx_search_results_book_id ON search_results(book_id);
CREATE INDEX idx_search_results_status ON search_results(status);

-- +goose Down
DROP TABLE IF EXISTS search_results;
