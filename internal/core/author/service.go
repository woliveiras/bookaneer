package author

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Service provides author-related operations.
type Service struct {
	db *sql.DB
}

// New creates a new author service.
func New(db *sql.DB) *Service {
	return &Service{db: db}
}

// FindByID returns an author by ID.
func (s *Service) FindByID(ctx context.Context, id int64) (*Author, error) {
	var a Author
	var monitored int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, sort_name, COALESCE(foreign_id, ''), overview, image_url, status, monitored, path, added_at, updated_at
		FROM authors WHERE id = ?
	`, id).Scan(
		&a.ID, &a.Name, &a.SortName, &a.ForeignID, &a.Overview, &a.ImageURL,
		&a.Status, &monitored, &a.Path, &a.AddedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find author %d: %w", id, err)
	}
	a.Monitored = monitored == 1
	return &a, nil
}

// FindByForeignID returns an author by foreign ID (e.g., OpenLibrary key).
func (s *Service) FindByForeignID(ctx context.Context, foreignID string) (*Author, error) {
	var a Author
	var monitored int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, sort_name, COALESCE(foreign_id, ''), overview, image_url, status, monitored, path, added_at, updated_at
		FROM authors WHERE foreign_id = ?
	`, foreignID).Scan(
		&a.ID, &a.Name, &a.SortName, &a.ForeignID, &a.Overview, &a.ImageURL,
		&a.Status, &monitored, &a.Path, &a.AddedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find author by foreign id %s: %w", foreignID, err)
	}
	a.Monitored = monitored == 1
	return &a, nil
}

// FindByName returns an author by exact name match (case-insensitive).
func (s *Service) FindByName(ctx context.Context, name string) (*Author, error) {
	var a Author
	var monitored int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, sort_name, COALESCE(foreign_id, ''), overview, image_url, status, monitored, path, added_at, updated_at
		FROM authors WHERE LOWER(name) = LOWER(?)
	`, name).Scan(
		&a.ID, &a.Name, &a.SortName, &a.ForeignID, &a.Overview, &a.ImageURL,
		&a.Status, &monitored, &a.Path, &a.AddedAt, &a.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find author by name %s: %w", name, err)
	}
	a.Monitored = monitored == 1
	return &a, nil
}

// List returns authors matching the filter.
func (s *Service) List(ctx context.Context, filter ListAuthorsFilter) ([]Author, int, error) {
	var conditions []string
	var args []any

	if filter.Monitored != nil {
		if *filter.Monitored {
			conditions = append(conditions, "monitored = 1")
		} else {
			conditions = append(conditions, "monitored = 0")
		}
	}
	if filter.Status != "" {
		conditions = append(conditions, "status = ?")
		args = append(args, filter.Status)
	}
	if filter.Search != "" {
		conditions = append(conditions, "(name LIKE ? OR sort_name LIKE ?)")
		search := "%" + filter.Search + "%"
		args = append(args, search, search)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) FROM authors " + where
	if err := s.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count authors: %w", err)
	}

	// Build ORDER BY
	sortBy := "name"
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "sortName":
			sortBy = "sort_name"
		case "addedAt":
			sortBy = "added_at"
		default:
			sortBy = "name"
		}
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
		SELECT id, name, sort_name, COALESCE(foreign_id, ''), overview, image_url, status, monitored, path, added_at, updated_at
		FROM authors %s ORDER BY %s %s LIMIT ? OFFSET ?
	`, where, sortBy, sortDir)
	args = append(args, limit, offset)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list authors: %w", err)
	}
	defer rows.Close()

	var authors []Author
	for rows.Next() {
		var a Author
		var monitored int
		if err := rows.Scan(
			&a.ID, &a.Name, &a.SortName, &a.ForeignID, &a.Overview, &a.ImageURL,
			&a.Status, &monitored, &a.Path, &a.AddedAt, &a.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan author: %w", err)
		}
		a.Monitored = monitored == 1
		authors = append(authors, a)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate authors: %w", err)
	}

	return authors, total, nil
}

// Create creates a new author.
func (s *Service) Create(ctx context.Context, input CreateAuthorInput) (*Author, error) {
	if input.Name == "" {
		return nil, ErrInvalidInput
	}
	if input.SortName == "" {
		input.SortName = input.Name
	}
	if input.Status == "" {
		input.Status = "active"
	}

	// Check if author with same foreignId already exists
	if input.ForeignID != "" {
		existing, err := s.FindByForeignID(ctx, input.ForeignID)
		if err == nil {
			// Author exists - update monitored status and return
			monitoredTrue := true
			return s.Update(ctx, existing.ID, UpdateAuthorInput{Monitored: &monitoredTrue})
		}
		if err != ErrNotFound {
			return nil, fmt.Errorf("check existing author by foreign_id: %w", err)
		}
	}

	// Check if author with same name already exists (to avoid duplicates)
	existing, err := s.FindByName(ctx, input.Name)
	if err == nil {
		// Author with same name exists - update monitored status and return existing
		monitoredTrue := true
		return s.Update(ctx, existing.ID, UpdateAuthorInput{Monitored: &monitoredTrue})
	}
	if err != ErrNotFound {
		return nil, fmt.Errorf("check existing author by name: %w", err)
	}

	monitored := 0
	if input.Monitored {
		monitored = 1
	}

	// Convert empty foreign_id to NULL to avoid UNIQUE constraint issues
	var foreignID any = input.ForeignID
	if input.ForeignID == "" {
		foreignID = nil
	}

	result, err := s.db.ExecContext(ctx, `
		INSERT INTO authors (name, sort_name, foreign_id, overview, image_url, status, monitored, path)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, input.Name, input.SortName, foreignID, input.Overview, input.ImageURL, input.Status, monitored, input.Path)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			// Race condition - author was created between our check and insert
			// Try to find and return the existing author
			if input.ForeignID != "" {
				if existing, findErr := s.FindByForeignID(ctx, input.ForeignID); findErr == nil {
					monitoredTrue := true
					return s.Update(ctx, existing.ID, UpdateAuthorInput{Monitored: &monitoredTrue})
				}
			}
			if existing, findErr := s.FindByName(ctx, input.Name); findErr == nil {
				monitoredTrue := true
				return s.Update(ctx, existing.ID, UpdateAuthorInput{Monitored: &monitoredTrue})
			}
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("create author: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get author id: %w", err)
	}

	return s.FindByID(ctx, id)
}

// Update updates an existing author.
func (s *Service) Update(ctx context.Context, id int64, input UpdateAuthorInput) (*Author, error) {
	// First check if author exists
	existing, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Build update query dynamically
	var sets []string
	var args []any

	if input.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *input.Name)
	}
	if input.SortName != nil {
		sets = append(sets, "sort_name = ?")
		args = append(args, *input.SortName)
	}
	if input.ForeignID != nil {
		sets = append(sets, "foreign_id = ?")
		args = append(args, *input.ForeignID)
	}
	if input.Overview != nil {
		sets = append(sets, "overview = ?")
		args = append(args, *input.Overview)
	}
	if input.ImageURL != nil {
		sets = append(sets, "image_url = ?")
		args = append(args, *input.ImageURL)
	}
	if input.Status != nil {
		sets = append(sets, "status = ?")
		args = append(args, *input.Status)
	}
	if input.Monitored != nil {
		m := 0
		if *input.Monitored {
			m = 1
		}
		sets = append(sets, "monitored = ?")
		args = append(args, m)
	}
	if input.Path != nil {
		sets = append(sets, "path = ?")
		args = append(args, *input.Path)
	}

	if len(sets) == 0 {
		return existing, nil
	}

	sets = append(sets, "updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')")
	args = append(args, id)

	query := fmt.Sprintf("UPDATE authors SET %s WHERE id = ?", strings.Join(sets, ", "))
	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("update author %d: %w", id, err)
	}

	return s.FindByID(ctx, id)
}

// Delete deletes an author by ID.
func (s *Service) Delete(ctx context.Context, id int64, deleteFiles bool) error {
	// Get author info first (needed for deleting files)
	author, err := s.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// If deleteFiles is true, delete the author's folder and all files
	if deleteFiles {
		if err := s.deleteAuthorFiles(ctx, author); err != nil {
			return fmt.Errorf("delete author files: %w", err)
		}
	}

	// Delete author from database (CASCADE will delete books and book_files)
	result, err := s.db.ExecContext(ctx, "DELETE FROM authors WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete author %d: %w", id, err)
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

// deleteAuthorFiles deletes all files in the author's folder.
func (s *Service) deleteAuthorFiles(ctx context.Context, author *Author) error {
	// Get first root folder
	var rootPath string
	err := s.db.QueryRowContext(ctx, `SELECT path FROM root_folders ORDER BY id LIMIT 1`).Scan(&rootPath)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // No root folder, nothing to delete
		}
		return fmt.Errorf("get root folder: %w", err)
	}

	// Build author folder path
	authorFolder := sanitizeFolderName(author.Name)
	authorPath := filepath.Join(rootPath, authorFolder)

	// Check if folder exists
	if _, err := os.Stat(authorPath); os.IsNotExist(err) {
		return nil // Folder doesn't exist, nothing to delete
	}

	// Delete the entire author folder
	if err := os.RemoveAll(authorPath); err != nil {
		return fmt.Errorf("remove author folder %s: %w", authorPath, err)
	}

	return nil
}

// sanitizeFolderName sanitizes a name for use as a folder name.
func sanitizeFolderName(name string) string {
	// Replace characters that are invalid in filenames
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "",
		"?", "",
		"\"", "'",
		"<", "",
		">", "",
		"|", "-",
	)
	return strings.TrimSpace(replacer.Replace(name))
}

// GetStats returns statistics for an author.
func (s *Service) GetStats(ctx context.Context, id int64) (*AuthorStats, error) {
	var stats AuthorStats

	err := s.db.QueryRowContext(ctx, `
		SELECT 
			(SELECT COUNT(*) FROM books WHERE author_id = ?) as book_count,
			(SELECT COUNT(*) FROM book_files bf JOIN books b ON bf.book_id = b.id WHERE b.author_id = ?) as file_count,
			(SELECT COUNT(*) FROM books b WHERE b.author_id = ? AND b.monitored = 1 AND NOT EXISTS (SELECT 1 FROM book_files bf WHERE bf.book_id = b.id)) as missing,
			(SELECT COALESCE(SUM(bf.size), 0) FROM book_files bf JOIN books b ON bf.book_id = b.id WHERE b.author_id = ?) as total_size
	`, id, id, id, id).Scan(&stats.BookCount, &stats.BookFileCount, &stats.MissingBooks, &stats.TotalSizeBytes)
	if err != nil {
		return nil, fmt.Errorf("get author stats: %w", err)
	}

	return &stats, nil
}
