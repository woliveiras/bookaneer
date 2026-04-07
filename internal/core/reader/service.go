package reader

import (
	"context"
	"database/sql"
	"errors"
)

// Service provides book file and reading progress operations.
type Service struct {
	db *sql.DB
}

// New creates a new reader service.
func New(db *sql.DB) *Service {
	return &Service{db: db}
}

// GetBookFile retrieves a book file with metadata.
func (s *Service) GetBookFile(ctx context.Context, id int64) (*BookFile, error) {
	row := s.db.QueryRow(`
		SELECT bf.id, bf.book_id, bf.edition_id, bf.path, bf.relative_path,
		       bf.size, bf.format, bf.quality, bf.hash, bf.added_at,
		       b.title, a.name, b.cover_url
		FROM book_files bf
		JOIN books b ON b.id = bf.book_id
		LEFT JOIN authors a ON a.id = b.author_id
		WHERE bf.id = ?
	`, id)

	var bf BookFile
	var bookTitle, authorName, coverURL sql.NullString

	err := row.Scan(
		&bf.ID, &bf.BookID, &bf.EditionID, &bf.Path, &bf.RelativePath,
		&bf.Size, &bf.Format, &bf.Quality, &bf.Hash, &bf.AddedAt,
		&bookTitle, &authorName, &coverURL,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrBookFileNotFound
		}
		return nil, err
	}

	bf.BookTitle = bookTitle.String
	bf.AuthorName = authorName.String
	bf.CoverURL = coverURL.String

	return &bf, nil
}

// ListBookFiles returns all book files for a given book.
func (s *Service) ListBookFiles(ctx context.Context, bookID int64) ([]BookFile, error) {
	rows, err := s.db.Query(`
		SELECT bf.id, bf.book_id, bf.edition_id, bf.path, bf.relative_path,
		       bf.size, bf.format, bf.quality, bf.hash, bf.added_at
		FROM book_files bf
		WHERE bf.book_id = ?
		ORDER BY bf.format, bf.quality
	`, bookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []BookFile
	for rows.Next() {
		var bf BookFile
		if err := rows.Scan(
			&bf.ID, &bf.BookID, &bf.EditionID, &bf.Path, &bf.RelativePath,
			&bf.Size, &bf.Format, &bf.Quality, &bf.Hash, &bf.AddedAt,
		); err != nil {
			return nil, err
		}
		files = append(files, bf)
	}

	return files, rows.Err()
}

// GetProgress retrieves reading progress for a user and book file.
func (s *Service) GetProgress(ctx context.Context, bookFileID, userID int64) (*ReadingProgress, error) {
	row := s.db.QueryRow(`
		SELECT id, book_file_id, user_id, position, percentage, updated_at
		FROM reading_progress
		WHERE book_file_id = ? AND user_id = ?
	`, bookFileID, userID)

	var rp ReadingProgress
	err := row.Scan(&rp.ID, &rp.BookFileID, &rp.UserID, &rp.Position, &rp.Percentage, &rp.UpdatedAt)
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
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO reading_progress (book_file_id, user_id, position, percentage, updated_at)
		VALUES (?, ?, ?, ?, datetime('now'))
		ON CONFLICT(book_file_id, user_id) 
		DO UPDATE SET position = excluded.position, 
		              percentage = excluded.percentage,
		              updated_at = excluded.updated_at
	`, bookFileID, userID, position, percentage)
	if err != nil {
		return nil, err
	}

	return s.GetProgress(ctx, bookFileID, userID)
}
