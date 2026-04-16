package book

import (
	"context"
	"fmt"
	"strings"
)

// GetWithEditions returns a book with its editions and files.
func (s *Service) GetWithEditions(ctx context.Context, id int64) (*BookWithEditions, error) {
	book, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &BookWithEditions{
		Book:     *book,
		Editions: []Edition{},
		Files:    []BookFile{},
	}

	// Get editions
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, book_id, foreign_id, title, isbn, isbn13, format, publisher, release_date, page_count, language
		FROM editions WHERE book_id = ?
	`, id)
	if err != nil {
		return nil, fmt.Errorf("get editions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var e Edition
		if err := rows.Scan(
			&e.ID, &e.BookID, &e.ForeignID, &e.Title, &e.ISBN, &e.ISBN13,
			&e.Format, &e.Publisher, &e.ReleaseDate, &e.PageCount, &e.Language,
		); err != nil {
			return nil, fmt.Errorf("scan edition: %w", err)
		}
		result.Editions = append(result.Editions, e)
	}

	// Get files
	fileRows, err := s.db.QueryContext(ctx, `
		SELECT id, book_id, edition_id, path, relative_path, size, format, quality, hash, added_at, content_mismatch
		FROM book_files WHERE book_id = ?
	`, id)
	if err != nil {
		return nil, fmt.Errorf("get book files: %w", err)
	}
	defer func() { _ = fileRows.Close() }()

	for fileRows.Next() {
		var f BookFile
		if err := fileRows.Scan(
			&f.ID, &f.BookID, &f.EditionID, &f.Path, &f.RelativePath, &f.Size, &f.Format, &f.Quality, &f.Hash, &f.AddedAt, &f.ContentMismatch,
		); err != nil {
			return nil, fmt.Errorf("scan book file: %w", err)
		}
		result.Files = append(result.Files, f)
	}

	return result, nil
}

// CreateEdition creates a new edition for a book.
func (s *Service) CreateEdition(ctx context.Context, input CreateEditionInput) (*Edition, error) {
	if input.BookID == 0 || input.Title == "" {
		return nil, ErrInvalidInput
	}

	// Check book exists
	_, err := s.FindByID(ctx, input.BookID)
	if err != nil {
		return nil, err
	}

	result, err := s.db.ExecContext(ctx, `
		INSERT INTO editions (book_id, foreign_id, title, isbn, isbn13, format, publisher, release_date, page_count, language)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.BookID, input.ForeignID, input.Title, input.ISBN, input.ISBN13, input.Format, input.Publisher, input.ReleaseDate, input.PageCount, input.Language)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("create edition: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get edition id: %w", err)
	}

	var e Edition
	err = s.db.QueryRowContext(ctx, `
		SELECT id, book_id, foreign_id, title, isbn, isbn13, format, publisher, release_date, page_count, language
		FROM editions WHERE id = ?
	`, id).Scan(
		&e.ID, &e.BookID, &e.ForeignID, &e.Title, &e.ISBN, &e.ISBN13,
		&e.Format, &e.Publisher, &e.ReleaseDate, &e.PageCount, &e.Language,
	)
	if err != nil {
		return nil, fmt.Errorf("get created edition: %w", err)
	}

	return &e, nil
}

// DeleteEdition deletes an edition by ID.
func (s *Service) DeleteEdition(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM editions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete edition %d: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted rows: %w", err)
	}
	if rows == 0 {
		return ErrEditionNotFound
	}

	return nil
}
