package wanted

import (
	"context"
	"database/sql"
	"encoding/json"
	"log/slog"
)

// GetHistory returns recent history events.
func (s *Service) GetHistory(ctx context.Context, limit int, eventType string) ([]HistoryItem, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT h.id, h.book_id, h.author_id, h.event_type, h.source_title, h.quality, h.data, h.date,
		       COALESCE(b.title, '') as book_title,
		       COALESCE(a.name, '') as author_name
		FROM history h
		LEFT JOIN books b ON b.id = h.book_id
		LEFT JOIN authors a ON a.id = h.author_id
	`
	var args []any
	if eventType != "" {
		query += " WHERE h.event_type = ?"
		args = append(args, eventType)
	}
	query += " ORDER BY h.date DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []HistoryItem
	for rows.Next() {
		var item HistoryItem
		var bookID, authorID sql.NullInt64
		var dataJSON string
		if err := rows.Scan(&item.ID, &bookID, &authorID, &item.EventType, &item.SourceTitle, &item.Quality, &dataJSON, &item.Date, &item.BookTitle, &item.AuthorName); err != nil {
			return nil, err
		}
		if bookID.Valid {
			item.BookID = &bookID.Int64
		}
		if authorID.Valid {
			item.AuthorID = &authorID.Int64
		}
		_ = json.Unmarshal([]byte(dataJSON), &item.Data)
		items = append(items, item)
	}

	return items, rows.Err()
}

// recordHistory adds an entry to the history table.
func (s *Service) recordHistory(ctx context.Context, bookID, authorID int64, eventType, sourceTitle, quality string, data map[string]any) {
	dataJSON, _ := json.Marshal(data)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO history (book_id, author_id, event_type, source_title, quality, data)
		VALUES (?, ?, ?, ?, ?, ?)
	`, bookID, authorID, eventType, sourceTitle, quality, string(dataJSON))
	if err != nil {
		slog.Error("Failed to record history", "error", err)
	}
}

// GetBlocklist returns all blocklisted releases.
func (s *Service) GetBlocklist(ctx context.Context) ([]BlocklistItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT bl.id, bl.book_id, bl.source_title, bl.quality, bl.reason, bl.date,
		       COALESCE(b.title, '') as book_title,
		       COALESCE(a.name, '') as author_name
		FROM blocklist bl
		LEFT JOIN books b ON b.id = bl.book_id
		LEFT JOIN authors a ON a.id = b.author_id
		ORDER BY bl.date DESC
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []BlocklistItem
	for rows.Next() {
		var item BlocklistItem
		if err := rows.Scan(&item.ID, &item.BookID, &item.SourceTitle, &item.Quality, &item.Reason, &item.Date, &item.BookTitle, &item.AuthorName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// AddToBlocklist adds a release to the blocklist.
func (s *Service) AddToBlocklist(ctx context.Context, bookID int64, sourceTitle, quality, reason string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO blocklist (book_id, source_title, quality, reason)
		VALUES (?, ?, ?, ?)
	`, bookID, sourceTitle, quality, reason)
	return err
}

// RemoveFromBlocklist removes an item from the blocklist.
func (s *Service) RemoveFromBlocklist(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM blocklist WHERE id = ?`, id)
	return err
}
