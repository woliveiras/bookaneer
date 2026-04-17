-- Add error_message column to download_queue to persist failure reasons
-- Migration: 017_queue_error_message.sql
--
-- +goose Up

ALTER TABLE download_queue ADD COLUMN error_message TEXT NOT NULL DEFAULT '';

-- +goose Down

-- SQLite does not support DROP COLUMN on older versions; recreate table to roll back
CREATE TABLE download_queue_old AS SELECT * FROM download_queue;
DROP TABLE download_queue;
CREATE TABLE download_queue AS
    SELECT id, book_id, download_client_id, indexer_id, external_id, title, size, format,
           status, progress, download_url, added_at, save_path
    FROM download_queue_old;
DROP TABLE download_queue_old;
