-- Bookaneer — Initial Schema
-- Migration: 001_initial_schema.sql
-- SQLite (WAL mode) — applied via goose
--
-- +goose Up

-- =============================================================================
-- Core: Library
-- =============================================================================

CREATE TABLE authors (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT    NOT NULL,
    sort_name     TEXT    NOT NULL,          -- "Tolkien, J.R.R."
    foreign_id    TEXT    UNIQUE,            -- OpenLibrary author key
    overview      TEXT    DEFAULT '',
    image_url     TEXT    DEFAULT '',
    status        TEXT    NOT NULL DEFAULT 'active',  -- active, paused, ended
    monitored     INTEGER NOT NULL DEFAULT 1,
    path          TEXT    NOT NULL,          -- /library/J.R.R. Tolkien
    added_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX idx_authors_foreign_id ON authors(foreign_id);
CREATE INDEX idx_authors_monitored ON authors(monitored);

CREATE TABLE books (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    author_id       INTEGER NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
    title           TEXT    NOT NULL,
    sort_title      TEXT    NOT NULL,
    foreign_id      TEXT    UNIQUE,          -- OpenLibrary work key
    isbn            TEXT    DEFAULT '',
    isbn13          TEXT    DEFAULT '',
    release_date    TEXT    DEFAULT '',       -- YYYY-MM-DD
    overview        TEXT    DEFAULT '',
    image_url       TEXT    DEFAULT '',
    page_count      INTEGER DEFAULT 0,
    monitored       INTEGER NOT NULL DEFAULT 1,
    added_at        TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX idx_books_author_id ON books(author_id);
CREATE INDEX idx_books_foreign_id ON books(foreign_id);
CREATE INDEX idx_books_monitored ON books(monitored);

CREATE TABLE editions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id         INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    foreign_id      TEXT    UNIQUE,          -- OpenLibrary edition key
    title           TEXT    NOT NULL,
    isbn            TEXT    DEFAULT '',
    isbn13          TEXT    DEFAULT '',
    format          TEXT    DEFAULT '',       -- epub, mobi, pdf, hardcover, paperback
    publisher       TEXT    DEFAULT '',
    release_date    TEXT    DEFAULT '',
    page_count      INTEGER DEFAULT 0,
    language        TEXT    DEFAULT '',       -- ISO 639-1
    monitored       INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_editions_book_id ON editions(book_id);

CREATE TABLE series (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    foreign_id    TEXT    UNIQUE,
    title         TEXT    NOT NULL,
    description   TEXT    DEFAULT '',
    monitored     INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE series_books (
    series_id     INTEGER NOT NULL REFERENCES series(id) ON DELETE CASCADE,
    book_id       INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    position      TEXT    NOT NULL DEFAULT '',  -- "1", "2.5", "1-3"
    PRIMARY KEY (series_id, book_id)
);

CREATE INDEX idx_series_books_book_id ON series_books(book_id);

CREATE TABLE book_files (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id         INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    edition_id      INTEGER REFERENCES editions(id) ON DELETE SET NULL,
    path            TEXT    NOT NULL UNIQUE,
    relative_path   TEXT    NOT NULL,          -- relative to root folder
    size            INTEGER NOT NULL DEFAULT 0,
    format          TEXT    NOT NULL,          -- epub, mobi, azw3, pdf, cbz
    quality         TEXT    NOT NULL DEFAULT 'unknown',
    hash            TEXT    DEFAULT '',        -- SHA-256 of the file
    added_at        TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX idx_book_files_book_id ON book_files(book_id);

-- =============================================================================
-- Core: Reading Progress
-- =============================================================================

CREATE TABLE reading_progress (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    book_file_id    INTEGER NOT NULL REFERENCES book_files(id) ON DELETE CASCADE,
    user_id         INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    position        TEXT    NOT NULL DEFAULT '',   -- epubcfi string
    percentage      REAL    NOT NULL DEFAULT 0.0,
    updated_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    UNIQUE(book_file_id, user_id)
);

-- =============================================================================
-- Infrastructure: Folders & Profiles
-- =============================================================================

CREATE TABLE root_folders (
    id                        INTEGER PRIMARY KEY AUTOINCREMENT,
    path                      TEXT    NOT NULL UNIQUE,
    name                      TEXT    NOT NULL,
    default_quality_profile_id INTEGER REFERENCES quality_profiles(id) ON DELETE SET NULL
);

CREATE TABLE quality_profiles (
    id        INTEGER PRIMARY KEY AUTOINCREMENT,
    name      TEXT    NOT NULL,
    cutoff    TEXT    NOT NULL DEFAULT 'epub',
    items     TEXT    NOT NULL DEFAULT '[]'     -- JSON array of {quality, allowed}
);

-- =============================================================================
-- Infrastructure: Indexers & Download Clients
-- =============================================================================

CREATE TABLE indexers (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT    NOT NULL,
    type          TEXT    NOT NULL,               -- newznab, torznab
    settings      TEXT    NOT NULL DEFAULT '{}',  -- JSON: {baseUrl, apiKey, categories}
    enabled       INTEGER NOT NULL DEFAULT 1,
    priority      INTEGER NOT NULL DEFAULT 25     -- lower = higher priority
);

CREATE TABLE download_clients (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT    NOT NULL,
    type          TEXT    NOT NULL,               -- qbittorrent, transmission, sabnzbd, nzbget, blackhole
    settings      TEXT    NOT NULL DEFAULT '{}',  -- JSON: connection settings
    enabled       INTEGER NOT NULL DEFAULT 1,
    priority      INTEGER NOT NULL DEFAULT 1
);

-- =============================================================================
-- Infrastructure: Notifications
-- =============================================================================

CREATE TABLE notifications (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    name          TEXT    NOT NULL,
    type          TEXT    NOT NULL,               -- webhook, discord, email, telegram, gotify, apprise
    settings      TEXT    NOT NULL DEFAULT '{}',  -- JSON
    on_grab       INTEGER NOT NULL DEFAULT 0,
    on_download   INTEGER NOT NULL DEFAULT 0,
    on_upgrade    INTEGER NOT NULL DEFAULT 0,
    enabled       INTEGER NOT NULL DEFAULT 1
);

-- =============================================================================
-- Infrastructure: History
-- =============================================================================

CREATE TABLE history (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id         INTEGER REFERENCES books(id) ON DELETE SET NULL,
    author_id       INTEGER REFERENCES authors(id) ON DELETE SET NULL,
    event_type      TEXT    NOT NULL,   -- grabbed, downloadCompleted, downloadFailed, bookFileDeleted, bookFileRenamed, bookImported
    source_title    TEXT    NOT NULL DEFAULT '',
    quality         TEXT    NOT NULL DEFAULT '',
    data            TEXT    NOT NULL DEFAULT '{}',  -- JSON: extra context (indexer, client, etc)
    date            TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX idx_history_book_id ON history(book_id);
CREATE INDEX idx_history_author_id ON history(author_id);
CREATE INDEX idx_history_event_type ON history(event_type);
CREATE INDEX idx_history_date ON history(date);

-- =============================================================================
-- Infrastructure: Config
-- =============================================================================

CREATE TABLE config (
    key     TEXT PRIMARY KEY,
    value   TEXT NOT NULL DEFAULT ''
);

-- Defaults
INSERT INTO config (key, value) VALUES
    ('general.port', '8787'),
    ('general.bindAddress', '0.0.0.0'),
    ('general.urlBase', ''),
    ('general.apiKey', ''),           -- generated on first run
    ('general.authMethod', 'forms'),  -- none, forms, basic
    ('naming.enabled', '1'),
    ('naming.authorFolderFormat', '$Author'),
    ('naming.bookFileFormat', '$Author - $Title{ ($SeriesName #$SeriesPosition)}'),
    ('naming.replaceSpaces', '0'),
    ('naming.colonReplacement', 'dash');

-- =============================================================================
-- Infrastructure: Users
-- =============================================================================

CREATE TABLE users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    username        TEXT    NOT NULL UNIQUE,
    password_hash   TEXT    NOT NULL,         -- bcrypt
    api_key         TEXT    NOT NULL UNIQUE,
    role            TEXT    NOT NULL DEFAULT 'user',  -- admin, user
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

-- =============================================================================
-- Infrastructure: Tags
-- =============================================================================

CREATE TABLE tags (
    id    INTEGER PRIMARY KEY AUTOINCREMENT,
    label TEXT    NOT NULL UNIQUE
);

CREATE TABLE tag_mappings (
    tag_id        INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    entity_type   TEXT    NOT NULL,   -- author, book, indexer, download_client, notification
    entity_id     INTEGER NOT NULL,
    PRIMARY KEY (tag_id, entity_type, entity_id)
);

CREATE INDEX idx_tag_mappings_entity ON tag_mappings(entity_type, entity_id);

-- =============================================================================
-- Jobs: Command Queue & Scheduler
-- =============================================================================

CREATE TABLE commands (
    id            TEXT    PRIMARY KEY,  -- ULID
    name          TEXT    NOT NULL,     -- BookSearch, MissingBookSearch, LibraryScan, RssSync, RenameFiles, Backup
    status        TEXT    NOT NULL DEFAULT 'queued',  -- queued, running, completed, failed, cancelled
    priority      INTEGER NOT NULL DEFAULT 0,
    payload       TEXT    NOT NULL DEFAULT '{}',  -- JSON: command-specific args
    result        TEXT    NOT NULL DEFAULT '{}',  -- JSON: output/errors
    trigger       TEXT    NOT NULL DEFAULT 'manual',  -- manual, scheduled, automatic
    queued_at     TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    started_at    TEXT,
    ended_at      TEXT
);

CREATE INDEX idx_commands_status ON commands(status);
CREATE INDEX idx_commands_queued ON commands(status, priority DESC, queued_at ASC);

CREATE TABLE scheduled_tasks (
    name              TEXT    PRIMARY KEY,
    interval_seconds  INTEGER NOT NULL,
    last_run_at       TEXT,
    next_run_at       TEXT    NOT NULL
);

-- Default scheduled tasks
INSERT INTO scheduled_tasks (name, interval_seconds, next_run_at) VALUES
    ('RssSync',            900,  strftime('%Y-%m-%dT%H:%M:%SZ', 'now', '+15 minutes')),
    ('LibraryScan',       86400, strftime('%Y-%m-%dT%H:%M:%SZ', 'now', '+24 hours')),
    ('MissingBookSearch', 86400, strftime('%Y-%m-%dT%H:%M:%SZ', 'now', '+24 hours')),
    ('Backup',           604800, strftime('%Y-%m-%dT%H:%M:%SZ', 'now', '+7 days')),
    ('Housekeeping',      86400, strftime('%Y-%m-%dT%H:%M:%SZ', 'now', '+24 hours'));

-- =============================================================================
-- Jobs: Download Queue (tracking active downloads)
-- =============================================================================

CREATE TABLE download_queue (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id             INTEGER NOT NULL REFERENCES books(id) ON DELETE CASCADE,
    download_client_id  INTEGER REFERENCES download_clients(id) ON DELETE SET NULL,
    indexer_id          INTEGER REFERENCES indexers(id) ON DELETE SET NULL,
    external_id         TEXT    NOT NULL DEFAULT '',   -- hash/nzo_id from client
    title               TEXT    NOT NULL,
    size                INTEGER NOT NULL DEFAULT 0,
    format              TEXT    NOT NULL DEFAULT 'unknown',
    status              TEXT    NOT NULL DEFAULT 'queued',  -- queued, downloading, paused, completed, failed, importing
    progress            REAL    NOT NULL DEFAULT 0.0,
    download_url        TEXT    NOT NULL DEFAULT '',
    added_at            TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX idx_download_queue_book_id ON download_queue(book_id);
CREATE INDEX idx_download_queue_status ON download_queue(status);

-- =============================================================================
-- +goose Down
-- =============================================================================

DROP TABLE IF EXISTS download_queue;
DROP TABLE IF EXISTS scheduled_tasks;
DROP TABLE IF EXISTS commands;
DROP TABLE IF EXISTS tag_mappings;
DROP TABLE IF EXISTS tags;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS config;
DROP TABLE IF EXISTS history;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS download_clients;
DROP TABLE IF EXISTS indexers;
DROP TABLE IF EXISTS quality_profiles;
DROP TABLE IF EXISTS root_folders;
DROP TABLE IF EXISTS reading_progress;
DROP TABLE IF EXISTS book_files;
DROP TABLE IF EXISTS series_books;
DROP TABLE IF EXISTS series;
DROP TABLE IF EXISTS editions;
DROP TABLE IF EXISTS books;
DROP TABLE IF EXISTS authors;
