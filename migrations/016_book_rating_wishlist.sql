-- +goose Up
ALTER TABLE books ADD COLUMN user_rating INTEGER DEFAULT NULL CHECK (user_rating IS NULL OR (user_rating >= 1 AND user_rating <= 5));
ALTER TABLE books ADD COLUMN in_wishlist INTEGER NOT NULL DEFAULT 0;

-- +goose Down
-- SQLite does not support DROP COLUMN in older versions; migration is intentionally left without a down step.
