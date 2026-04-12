-- +goose Up
ALTER TABLE book_files ADD COLUMN content_mismatch INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE book_files DROP COLUMN content_mismatch;
