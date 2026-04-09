package series

import (
	"context"
	"fmt"
	"strings"
)

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
	_, err := s.FindByID(ctx, seriesID)
	if err != nil {
		return err
	}

	var bookExists int
	err = s.db.QueryRowContext(ctx, "SELECT 1 FROM books WHERE id = ?", input.BookID).Scan(&bookExists)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return ErrBookNotFound
		}
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
