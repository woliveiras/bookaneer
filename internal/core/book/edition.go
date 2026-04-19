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
	var editions []Edition
	if err := s.db.SelectContext(ctx, &editions, `
		SELECT id, book_id, foreign_id, title, isbn, isbn13, format, publisher, release_date, page_count, language
		FROM editions WHERE book_id = ?
	`, id); err != nil {
		return nil, fmt.Errorf("get editions: %w", err)
	}
	if editions != nil {
		result.Editions = editions
	}

	// Get files
	var files []BookFile
	if err := s.db.SelectContext(ctx, &files, `
		SELECT id, book_id, edition_id, path, relative_path, size, format, quality, hash, added_at, content_mismatch
		FROM book_files WHERE book_id = ?
	`, id); err != nil {
		return nil, fmt.Errorf("get book files: %w", err)
	}
	if files != nil {
		result.Files = files
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

	res, err := s.db.NamedExecContext(ctx, `
		INSERT INTO editions (book_id, foreign_id, title, isbn, isbn13, format, publisher, release_date, page_count, language)
		VALUES (:book_id, :foreign_id, :title, :isbn, :isbn13, :format, :publisher, :release_date, :page_count, :language)
	`, input)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("create edition: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get edition id: %w", err)
	}

	var e Edition
	if err := s.db.GetContext(ctx, &e, `
		SELECT id, book_id, foreign_id, title, isbn, isbn13, format, publisher, release_date, page_count, language
		FROM editions WHERE id = ?
	`, id); err != nil {
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
