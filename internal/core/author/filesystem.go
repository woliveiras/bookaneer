package author

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// deleteAuthorFiles deletes all files in the author's folder.
func (s *Service) deleteAuthorFiles(ctx context.Context, author *Author) error {
	// Get first root folder
	var rootPath string
	err := s.db.GetContext(ctx, &rootPath, `SELECT path FROM root_folders ORDER BY id LIMIT 1`)
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

	err := s.db.GetContext(ctx, &stats, `
		SELECT
			(SELECT COUNT(*) FROM books WHERE author_id = ?) as book_count,
			(SELECT COUNT(*) FROM book_files bf JOIN books b ON bf.book_id = b.id WHERE b.author_id = ?) as file_count,
			(SELECT COUNT(*) FROM books b WHERE b.author_id = ? AND b.monitored = 1 AND NOT EXISTS (SELECT 1 FROM book_files bf WHERE bf.book_id = b.id)) as missing,
			(SELECT COALESCE(SUM(bf.size), 0) FROM book_files bf JOIN books b ON bf.book_id = b.id WHERE b.author_id = ?) as total_size
	`, id, id, id, id)
	if err != nil {
		return nil, fmt.Errorf("get author stats: %w", err)
	}

	return &stats, nil
}
