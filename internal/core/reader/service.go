package reader

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

// Service provides book file and reading progress operations.
type Service struct {
	db *sqlx.DB
}

// New creates a new reader service.
func New(db *sqlx.DB) *Service {
	return &Service{db: db}
}

// GetBookFile retrieves a book file with metadata.
func (s *Service) GetBookFile(ctx context.Context, id int64) (*BookFile, error) {
	var bf BookFile
	err := s.db.GetContext(ctx, &bf, `
		SELECT bf.id, bf.book_id, bf.edition_id, bf.path, bf.relative_path,
		       bf.size, bf.format, bf.quality, bf.hash, bf.added_at,
		       b.title AS book_title, COALESCE(a.name, '') AS author_name, COALESCE(b.image_url, '') AS cover_url
		FROM book_files bf
		JOIN books b ON b.id = bf.book_id
		LEFT JOIN authors a ON a.id = b.author_id
		WHERE bf.id = ?
	`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBookFileNotFound
		}
		return nil, err
	}
	return &bf, nil
}

// ListBookFiles returns all book files for a given book.
func (s *Service) ListBookFiles(ctx context.Context, bookID int64) ([]BookFile, error) {
	var files []BookFile
	err := s.db.SelectContext(ctx, &files, `
		SELECT bf.id, bf.book_id, bf.edition_id, bf.path, bf.relative_path,
		       bf.size, bf.format, bf.quality, bf.hash, bf.added_at
		FROM book_files bf
		WHERE bf.book_id = ?
		ORDER BY bf.format, bf.quality
	`, bookID)
	return files, err
}

// GetProgress retrieves reading progress for a user and book file.
func (s *Service) GetProgress(ctx context.Context, bookFileID, userID int64) (*ReadingProgress, error) {
	var rp ReadingProgress
	err := s.db.GetContext(ctx, &rp, `
		SELECT id, book_file_id, user_id, position, percentage, updated_at
		FROM reading_progress
		WHERE book_file_id = ? AND user_id = ?
	`, bookFileID, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProgressNotFound
		}
		return nil, err
	}
	return &rp, nil
}

// SaveProgress saves or updates reading progress and returns the saved progress.
func (s *Service) SaveProgress(ctx context.Context, bookFileID, userID int64, position string, percentage float64) (*ReadingProgress, error) {
	_, err := s.db.NamedExecContext(ctx, `
		INSERT INTO reading_progress (book_file_id, user_id, position, percentage, updated_at)
		VALUES (:book_file_id, :user_id, :position, :percentage, datetime('now'))
		ON CONFLICT(book_file_id, user_id)
		DO UPDATE SET position = excluded.position,
		              percentage = excluded.percentage,
		              updated_at = excluded.updated_at
	`, map[string]any{
		"book_file_id": bookFileID,
		"user_id":      userID,
		"position":     position,
		"percentage":   percentage,
	})
	if err != nil {
		return nil, err
	}

	return s.GetProgress(ctx, bookFileID, userID)
}

// Bookmark operations

// ListBookmarks returns all bookmarks for a user and book file.
func (s *Service) ListBookmarks(ctx context.Context, bookFileID, userID int64) ([]Bookmark, error) {
	var bookmarks []Bookmark
	err := s.db.SelectContext(ctx, &bookmarks, `
		SELECT id, book_file_id, user_id, position, title, note, created_at
		FROM bookmarks
		WHERE book_file_id = ? AND user_id = ?
		ORDER BY created_at DESC
	`, bookFileID, userID)
	return bookmarks, err
}

// CreateBookmark creates a new bookmark.
func (s *Service) CreateBookmark(ctx context.Context, bookFileID, userID int64, position, title, note string) (*Bookmark, error) {
	result, err := s.db.NamedExecContext(ctx, `
		INSERT INTO bookmarks (book_file_id, user_id, position, title, note)
		VALUES (:book_file_id, :user_id, :position, :title, :note)
	`, map[string]any{
		"book_file_id": bookFileID,
		"user_id":      userID,
		"position":     position,
		"title":        title,
		"note":         note,
	})
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return s.GetBookmark(ctx, id, userID)
}

// GetBookmark retrieves a bookmark by ID.
func (s *Service) GetBookmark(ctx context.Context, id, userID int64) (*Bookmark, error) {
	var bm Bookmark
	err := s.db.GetContext(ctx, &bm, `
		SELECT id, book_file_id, user_id, position, title, note, created_at
		FROM bookmarks
		WHERE id = ? AND user_id = ?
	`, id, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBookmarkNotFound
		}
		return nil, err
	}
	return &bm, nil
}

// DeleteBookmark deletes a bookmark.
func (s *Service) DeleteBookmark(ctx context.Context, id, userID int64) error {
	result, err := s.db.ExecContext(ctx, `
		DELETE FROM bookmarks WHERE id = ? AND user_id = ?
	`, id, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrBookmarkNotFound
	}

	return nil
}
