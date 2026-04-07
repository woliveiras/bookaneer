package series

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// Service provides series-related operations.
type Service struct {
	db *sql.DB
}

// New creates a new series service.
func New(db *sql.DB) *Service {
	return &Service{db: db}
}

// FindByID returns a series by ID.
func (s *Service) FindByID(ctx context.Context, id int64) (*Series, error) {
	var ser Series
	var monitored int
	err := s.db.QueryRowContext(ctx, `
		SELECT s.id, s.foreign_id, s.title, s.description, s.monitored,
		       (SELECT COUNT(*) FROM series_books sb WHERE sb.series_id = s.id) as book_count
		FROM series s WHERE s.id = ?
	`, id).Scan(&ser.ID, &ser.ForeignID, &ser.Title, &ser.Description, &monitored, &ser.BookCount)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find series %d: %w", id, err)
	}
	ser.Monitored = monitored == 1
	return &ser, nil
}

// List returns series matching the filter.
func (s *Service) List(ctx context.Context, filter ListSeriesFilter) ([]Series, int, error) {
	var conditions []string
	var args []any

	if filter.Monitored != nil {
		if *filter.Monitored {
			conditions = append(conditions, "monitored = 1")
		} else {
			conditions = append(conditions, "monitored = 0")
		}
	}
	if filter.Search != "" {
		conditions = append(conditions, "title LIKE ?")
		args = append(args, "%"+filter.Search+"%")
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) FROM series " + where
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count series: %w", err)
	}

	// Build ORDER BY
	sortBy := "title"
	if filter.SortBy == "bookCount" {
		sortBy = "book_count"
	}
	sortDir := "ASC"
	if filter.SortDir == "desc" {
		sortDir = "DESC"
	}

	// Apply limit/offset
	limit := 50
	if filter.Limit > 0 && filter.Limit <= 500 {
		limit = filter.Limit
	}
	offset := 0
	if filter.Offset > 0 {
		offset = filter.Offset
	}

	query := fmt.Sprintf(`
		SELECT s.id, s.foreign_id, s.title, s.description, s.monitored,
		       (SELECT COUNT(*) FROM series_books sb WHERE sb.series_id = s.id) as book_count
		FROM series s %s ORDER BY %s %s LIMIT ? OFFSET ?
	`, where, sortBy, sortDir)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list series: %w", err)
	}
	defer rows.Close()

	var seriesList []Series
	for rows.Next() {
		var ser Series
		var monitored int
		if err := rows.Scan(&ser.ID, &ser.ForeignID, &ser.Title, &ser.Description, &monitored, &ser.BookCount); err != nil {
			return nil, 0, fmt.Errorf("scan series: %w", err)
		}
		ser.Monitored = monitored == 1
		seriesList = append(seriesList, ser)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate series: %w", err)
	}

	return seriesList, total, nil
}

// Create creates a new series.
func (s *Service) Create(ctx context.Context, input CreateSeriesInput) (*Series, error) {
	if input.Title == "" {
		return nil, ErrInvalidInput
	}

	monitored := 0
	if input.Monitored {
		monitored = 1
	}

	result, err := s.db.ExecContext(ctx, `
		INSERT INTO series (foreign_id, title, description, monitored)
		VALUES (?, ?, ?, ?)
	`, input.ForeignID, input.Title, input.Description, monitored)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("create series: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get series id: %w", err)
	}

	return s.FindByID(ctx, id)
}

// Update updates an existing series.
func (s *Service) Update(ctx context.Context, id int64, input UpdateSeriesInput) (*Series, error) {
	existing, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var sets []string
	var args []any

	if input.ForeignID != nil {
		sets = append(sets, "foreign_id = ?")
		args = append(args, *input.ForeignID)
	}
	if input.Title != nil {
		sets = append(sets, "title = ?")
		args = append(args, *input.Title)
	}
	if input.Description != nil {
		sets = append(sets, "description = ?")
		args = append(args, *input.Description)
	}
	if input.Monitored != nil {
		m := 0
		if *input.Monitored {
			m = 1
		}
		sets = append(sets, "monitored = ?")
		args = append(args, m)
	}

	if len(sets) == 0 {
		return existing, nil
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE series SET %s WHERE id = ?", strings.Join(sets, ", "))
	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("update series %d: %w", id, err)
	}

	return s.FindByID(ctx, id)
}

// Delete deletes a series by ID.
func (s *Service) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM series WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete series %d: %w", id, err)
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

// GetWithBooks returns a series with its books.
func (s *Service) GetWithBooks(ctx context.Context, id int64) (*SeriesWithBooks, error) {
	ser, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	result := &SeriesWithBooks{
		Series: *ser,
		Books:  []SeriesBook{},
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT sb.series_id, sb.book_id, sb.position, b.title, a.name
		FROM series_books sb
		JOIN books b ON sb.book_id = b.id
		JOIN authors a ON b.author_id = a.id
		WHERE sb.series_id = ?
		ORDER BY sb.position
	`, id)
	if err != nil {
		return nil, fmt.Errorf("get series books: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var sb SeriesBook
		if err := rows.Scan(&sb.SeriesID, &sb.BookID, &sb.Position, &sb.BookTitle, &sb.AuthorName); err != nil {
			return nil, fmt.Errorf("scan series book: %w", err)
		}
		result.Books = append(result.Books, sb)
	}

	return result, nil
}

// AddBook adds a book to a series.
func (s *Service) AddBook(ctx context.Context, seriesID int64, input AddBookInput) error {
	// Check series exists
	_, err := s.FindByID(ctx, seriesID)
	if err != nil {
		return err
	}

	// Check book exists
	var bookExists int
	err = s.db.QueryRowContext(ctx, "SELECT 1 FROM books WHERE id = ?", input.BookID).Scan(&bookExists)
	if err == sql.ErrNoRows {
		return ErrBookNotFound
	}
	if err != nil {
		return fmt.Errorf("check book: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO series_books (series_id, book_id, position)
		VALUES (?, ?, ?)
	`, seriesID, input.BookID, input.Position)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") || strings.Contains(err.Error(), "PRIMARY KEY constraint failed") {
			return ErrBookAlreadyInSeries
		}
		return fmt.Errorf("add book to series: %w", err)
	}

	return nil
}

// RemoveBook removes a book from a series.
func (s *Service) RemoveBook(ctx context.Context, seriesID, bookID int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM series_books WHERE series_id = ? AND book_id = ?", seriesID, bookID)
	if err != nil {
		return fmt.Errorf("remove book from series: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check deleted rows: %w", err)
	}
	if rows == 0 {
		return ErrBookNotFound
	}

	return nil
}

// UpdateBookPosition updates the position of a book in a series.
func (s *Service) UpdateBookPosition(ctx context.Context, seriesID, bookID int64, position string) error {
	result, err := s.db.ExecContext(ctx, `
		UPDATE series_books SET position = ? WHERE series_id = ? AND book_id = ?
	`, position, seriesID, bookID)
	if err != nil {
		return fmt.Errorf("update book position: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check updated rows: %w", err)
	}
	if rows == 0 {
		return ErrBookNotFound
	}

	return nil
}
