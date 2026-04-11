package book

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/woliveiras/bookaneer/internal/database"
)

// Service provides book-related operations.
type Service struct {
	db *sql.DB
}

// New creates a new book service.
func New(db *sql.DB) *Service {
	return &Service{db: db}
}

// FindByID returns a book by ID.
func (s *Service) FindByID(ctx context.Context, id int64) (*Book, error) {
	var b Book
	var monitored int
	var hasFile int
	err := s.db.QueryRowContext(ctx, `
		SELECT b.id, b.author_id, b.title, COALESCE(b.sort_title,''), COALESCE(b.foreign_id,''), COALESCE(b.isbn,''), COALESCE(b.isbn13,''),
		       COALESCE(b.release_date,''), COALESCE(b.overview,''), COALESCE(b.image_url,''), b.page_count, b.monitored, b.added_at, b.updated_at,
		       a.name,
		       (SELECT COUNT(*) > 0 FROM book_files bf WHERE bf.book_id = b.id) as has_file,
		       COALESCE((SELECT bf.format FROM book_files bf WHERE bf.book_id = b.id ORDER BY bf.added_at DESC LIMIT 1), '') as file_format
		FROM books b
		JOIN authors a ON b.author_id = a.id
		WHERE b.id = ?
	`, id).Scan(
		&b.ID, &b.AuthorID, &b.Title, &b.SortTitle, &b.ForeignID, &b.ISBN, &b.ISBN13,
		&b.ReleaseDate, &b.Overview, &b.ImageURL, &b.PageCount, &monitored, &b.AddedAt, &b.UpdatedAt,
		&b.AuthorName, &hasFile, &b.FileFormat,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find book %d: %w", id, err)
	}
	b.Monitored = monitored == 1
	b.HasFile = hasFile == 1

	return &b, nil
}

// FindByForeignID returns a book by foreign ID.
func (s *Service) FindByForeignID(ctx context.Context, foreignID string) (*Book, error) {
	var b Book
	var monitored int
	err := s.db.QueryRowContext(ctx, `
		SELECT b.id, b.author_id, b.title, COALESCE(b.sort_title,''), COALESCE(b.foreign_id,''), COALESCE(b.isbn,''), COALESCE(b.isbn13,''),
		       COALESCE(b.release_date,''), COALESCE(b.overview,''), COALESCE(b.image_url,''), b.page_count, b.monitored, b.added_at, b.updated_at,
		       a.name
		FROM books b
		JOIN authors a ON b.author_id = a.id
		WHERE b.foreign_id = ?
	`, foreignID).Scan(
		&b.ID, &b.AuthorID, &b.Title, &b.SortTitle, &b.ForeignID, &b.ISBN, &b.ISBN13,
		&b.ReleaseDate, &b.Overview, &b.ImageURL, &b.PageCount, &monitored, &b.AddedAt, &b.UpdatedAt,
		&b.AuthorName,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find book by foreign id %s: %w", foreignID, err)
	}
	b.Monitored = monitored == 1
	return &b, nil
}

// List returns books matching the filter.
func (s *Service) List(ctx context.Context, filter ListBooksFilter) ([]Book, int, error) {
	var conditions []string
	var args []any

	if filter.AuthorID != nil {
		conditions = append(conditions, "b.author_id = ?")
		args = append(args, *filter.AuthorID)
	}
	if filter.Monitored != nil {
		if *filter.Monitored {
			conditions = append(conditions, "b.monitored = 1")
		} else {
			conditions = append(conditions, "b.monitored = 0")
		}
	}
	if filter.Missing {
		conditions = append(conditions, "NOT EXISTS (SELECT 1 FROM book_files bf WHERE bf.book_id = b.id)")
	}
	if filter.Search != "" {
		conditions = append(conditions, "(b.title LIKE ? OR b.sort_title LIKE ? OR b.isbn LIKE ? OR b.isbn13 LIKE ?)")
		search := "%" + filter.Search + "%"
		args = append(args, search, search, search, search)
	}

	where := database.BuildWhereClause(conditions)

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) FROM books b " + where
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count books: %w", err)
	}

	// Build ORDER BY
	sortBy := "b.title"
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "sortTitle":
			sortBy = "b.sort_title"
		case "releaseDate":
			sortBy = "b.release_date"
		case "addedAt":
			sortBy = "b.added_at"
		default:
			sortBy = "b.title"
		}
	}
	sortDir := database.NormaliseSortDir(filter.SortDir)
	limit := database.ClampLimit(filter.Limit, 50, 500)
	offset := filter.Offset

	query := fmt.Sprintf(`
		SELECT b.id, b.author_id, b.title, COALESCE(b.sort_title,''), COALESCE(b.foreign_id,''), COALESCE(b.isbn,''), COALESCE(b.isbn13,''),
		       COALESCE(b.release_date,''), COALESCE(b.overview,''), COALESCE(b.image_url,''), b.page_count, b.monitored, b.added_at, b.updated_at,
		       a.name,
		       EXISTS (SELECT 1 FROM book_files bf WHERE bf.book_id = b.id) as has_file,
		       COALESCE((SELECT bf.format FROM book_files bf WHERE bf.book_id = b.id ORDER BY bf.added_at DESC LIMIT 1), '') as file_format
		FROM books b
		JOIN authors a ON b.author_id = a.id
		%s ORDER BY %s %s LIMIT ? OFFSET ?
	`, where, sortBy, sortDir)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list books: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var books []Book
	for rows.Next() {
		var b Book
		var monitored, hasFile int
		if err := rows.Scan(
			&b.ID, &b.AuthorID, &b.Title, &b.SortTitle, &b.ForeignID, &b.ISBN, &b.ISBN13,
			&b.ReleaseDate, &b.Overview, &b.ImageURL, &b.PageCount, &monitored, &b.AddedAt, &b.UpdatedAt,
			&b.AuthorName, &hasFile, &b.FileFormat,
		); err != nil {
			return nil, 0, fmt.Errorf("scan book: %w", err)
		}
		b.Monitored = monitored == 1
		b.HasFile = hasFile == 1
		books = append(books, b)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate books: %w", err)
	}

	return books, total, nil
}

// Create creates a new book.
func (s *Service) Create(ctx context.Context, input CreateBookInput) (*Book, error) {
	if input.Title == "" {
		return nil, ErrInvalidInput
	}
	if input.AuthorID == 0 {
		return nil, ErrInvalidInput
	}
	if input.SortTitle == "" {
		input.SortTitle = input.Title
	}

	// Check if book with same foreignId already exists (e.g., was previously removed from wanted)
	if input.ForeignID != "" {
		existing, err := s.FindByForeignID(ctx, input.ForeignID)
		if err == nil {
			// Book exists - update it to set monitored=true
			monitoredTrue := true
			return s.Update(ctx, existing.ID, UpdateBookInput{Monitored: &monitoredTrue})
		}
		// If not found, continue to create new book
		if err != ErrNotFound {
			return nil, fmt.Errorf("check existing book: %w", err)
		}
	}

	// Check author exists
	var authorExists int
	err := s.db.QueryRowContext(ctx, "SELECT 1 FROM authors WHERE id = ?", input.AuthorID).Scan(&authorExists)
	if err == sql.ErrNoRows {
		return nil, ErrAuthorNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("check author: %w", err)
	}

	monitored := 0
	if input.Monitored {
		monitored = 1
	}

	result, err := s.db.ExecContext(ctx, `
		INSERT INTO books (author_id, title, sort_title, foreign_id, isbn, isbn13, release_date, overview, image_url, page_count, monitored)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.AuthorID, input.Title, input.SortTitle, input.ForeignID, input.ISBN, input.ISBN13, input.ReleaseDate, input.Overview, input.ImageURL, input.PageCount, monitored)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("create book: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get book id: %w", err)
	}

	return s.FindByID(ctx, id)
}

// Update updates an existing book.
func (s *Service) Update(ctx context.Context, id int64, input UpdateBookInput) (*Book, error) {
	existing, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var sets []string
	var args []any

	if input.AuthorID != nil {
		// Check author exists
		var authorExists int
		err := s.db.QueryRowContext(ctx, "SELECT 1 FROM authors WHERE id = ?", *input.AuthorID).Scan(&authorExists)
		if err == sql.ErrNoRows {
			return nil, ErrAuthorNotFound
		}
		sets = append(sets, "author_id = ?")
		args = append(args, *input.AuthorID)
	}
	if input.Title != nil {
		sets = append(sets, "title = ?")
		args = append(args, *input.Title)
	}
	if input.SortTitle != nil {
		sets = append(sets, "sort_title = ?")
		args = append(args, *input.SortTitle)
	}
	if input.ForeignID != nil {
		sets = append(sets, "foreign_id = ?")
		args = append(args, *input.ForeignID)
	}
	if input.ISBN != nil {
		sets = append(sets, "isbn = ?")
		args = append(args, *input.ISBN)
	}
	if input.ISBN13 != nil {
		sets = append(sets, "isbn13 = ?")
		args = append(args, *input.ISBN13)
	}
	if input.ReleaseDate != nil {
		sets = append(sets, "release_date = ?")
		args = append(args, *input.ReleaseDate)
	}
	if input.Overview != nil {
		sets = append(sets, "overview = ?")
		args = append(args, *input.Overview)
	}
	if input.ImageURL != nil {
		sets = append(sets, "image_url = ?")
		args = append(args, *input.ImageURL)
	}
	if input.PageCount != nil {
		sets = append(sets, "page_count = ?")
		args = append(args, *input.PageCount)
	}
	if input.Monitored != nil {
		m := 0
		if *input.Monitored {
			m = 1
		}
		sets = append(sets, "monitored = ?")
		args = append(args, m)

		// If unmonitoring, remove pending downloads from queue
		if !*input.Monitored {
			_, _ = s.db.ExecContext(ctx, `
				DELETE FROM download_queue 
				WHERE book_id = ? AND status IN ('queued', 'downloading', 'paused')
			`, id)
		}
	}

	if len(sets) == 0 {
		return existing, nil
	}

	sets = append(sets, "updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE books SET %s WHERE id = ?", strings.Join(sets, ", "))
	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("update book %d: %w", id, err)
	}

	return s.FindByID(ctx, id)
}

// Delete deletes a book by ID.
func (s *Service) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM books WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete book %d: %w", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted rows: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
