package wanted

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// importCompletedDownload imports a completed download to the library.
func (s *Service) importCompletedDownload(ctx context.Context, queueID int64, sourcePath string) error {
	// Get queue item to find book_id
	var bookID int64
	var format string
	err := s.db.QueryRowContext(ctx, `
		SELECT book_id, format FROM download_queue WHERE id = ?
	`, queueID).Scan(&bookID, &format)
	if err != nil {
		return fmt.Errorf("get queue item: %w", err)
	}

	// Get book info
	b, err := s.bookService.FindByID(ctx, bookID)
	if err != nil {
		return fmt.Errorf("find book: %w", err)
	}

	// Get first root folder
	var rootPath string
	err = s.db.QueryRowContext(ctx, `SELECT path FROM root_folders ORDER BY id LIMIT 1`).Scan(&rootPath)
	if err != nil {
		return fmt.Errorf("get root folder: %w", err)
	}

	// Build destination path: rootPath/AuthorName/BookTitle.format
	authorDir := filepath.Join(rootPath, sanitizeFilename(b.AuthorName))
	if err := os.MkdirAll(authorDir, 0755); err != nil {
		return fmt.Errorf("create author directory: %w", err)
	}

	// Determine format from source file if not in queue
	if format == "" || format == "unknown" {
		ext := strings.ToLower(filepath.Ext(sourcePath))
		format = strings.TrimPrefix(ext, ".")
	}

	// Build filename: AuthorName - BookTitle.format
	filename := fmt.Sprintf("%s - %s.%s", sanitizeFilename(b.AuthorName), sanitizeFilename(b.Title), format)
	destPath := filepath.Join(authorDir, filename)

	// Check if source and destination are the same file
	// This happens when the download was saved directly to the library location
	srcAbs, _ := filepath.Abs(sourcePath)
	dstAbs, _ := filepath.Abs(destPath)
	if srcAbs != dstAbs {
		// Copy file to library (copy instead of move for safety)
		if err := copyFile(sourcePath, destPath); err != nil {
			return fmt.Errorf("copy file: %w", err)
		}
	} else {
		slog.Debug("Source and destination are the same, skipping copy", "path", destPath)
	}

	// Get file info
	info, err := os.Stat(destPath)
	if err != nil {
		return fmt.Errorf("stat destination: %w", err)
	}

	// Calculate hash for smaller files
	hash := ""
	if info.Size() < 50*1024*1024 {
		hash = hashFile(destPath)
	}

	// Calculate relative path from root
	relativePath := filepath.Join(sanitizeFilename(b.AuthorName), filename)

	// Add to book_files
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO book_files (book_id, path, relative_path, size, format, quality, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, bookID, destPath, relativePath, info.Size(), format, format, hash)
	if err != nil {
		return fmt.Errorf("insert book_file: %w", err)
	}

	// Update queue status to imported
	if err := s.UpdateQueueItemStatus(ctx, queueID, "imported", 100); err != nil {
		return fmt.Errorf("update queue status: %w", err)
	}

	// Record history
	s.recordHistory(ctx, bookID, b.AuthorID, "bookImported", b.Title, format, map[string]any{
		"path":       destPath,
		"size":       info.Size(),
		"sourcePath": sourcePath,
	})

	// Try to remove source file (best effort)
	_ = os.Remove(sourcePath)

	return nil
}
