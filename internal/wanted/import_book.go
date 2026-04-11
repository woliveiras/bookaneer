package wanted

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/core/naming"
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

	// Determine format from source file if not in queue
	if format == "" || format == "unknown" {
		ext := strings.ToLower(filepath.Ext(sourcePath))
		format = strings.TrimPrefix(ext, ".")
	}

	// Build destination path using naming engine
	nc := s.buildNamingContext(ctx, b, format, filepath.Base(sourcePath))
	namingSettings, err := s.namingEngine.LoadSettings(ctx)
	if err != nil {
		slog.Warn("Failed to load naming settings, using defaults", "error", err)
		namingSettings = nil // Format() uses defaults for nil
	}
	result := s.namingEngine.Format(rootPath, nc, namingSettings)

	authorDir := filepath.Dir(result.FullPath)
	if err := os.MkdirAll(authorDir, 0755); err != nil {
		return fmt.Errorf("create author directory: %w", err)
	}

	destPath := result.FullPath
	relativePath := result.RelativePath

	// Check for duplicate: if book already has a file, handle it
	var existingFileID int64
	err = s.db.QueryRowContext(ctx, `SELECT id FROM book_files WHERE book_id = ?`, bookID).Scan(&existingFileID)
	if err == nil {
		slog.Info("Book already has a file, replacing", "bookId", bookID, "existingFileId", existingFileID)
		_, _ = s.db.ExecContext(ctx, `DELETE FROM book_files WHERE id = ?`, existingFileID)
	}

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

	// Add to book_files
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO book_files (book_id, path, relative_path, size, format, quality, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, bookID, destPath, relativePath, info.Size(), format, format, hash)
	if err != nil {
		return fmt.Errorf("insert book_file: %w", err)
	}

	// Update queue status to completed
	if err := s.UpdateQueueItemStatus(ctx, queueID, "completed", 100); err != nil {
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

	// Try to remove parent directory if empty (best effort)
	sourceDir := filepath.Dir(sourcePath)
	_ = os.Remove(sourceDir)

	return nil
}

// buildNamingContext creates a naming.Context from book data and series info.
func (s *Service) buildNamingContext(ctx context.Context, b *book.Book, format, originalName string) naming.Context {
	nc := naming.Context{
		Author:       b.AuthorName,
		Title:        b.Title,
		Format:       format,
		OriginalName: originalName,
	}

	// Get author sort name
	var sortName string
	err := s.db.QueryRowContext(ctx, `SELECT COALESCE(sort_name, '') FROM authors WHERE id = ?`, b.AuthorID).Scan(&sortName)
	if err == nil && sortName != "" {
		nc.SortAuthor = sortName
	}

	// Get year from release date
	if b.ReleaseDate != "" && len(b.ReleaseDate) >= 4 {
		nc.Year = b.ReleaseDate[:4]
	}

	// Get series info
	var seriesTitle, position string
	err = s.db.QueryRowContext(ctx, `
		SELECT s.title, sb.position
		FROM series_books sb
		JOIN series s ON sb.series_id = s.id
		WHERE sb.book_id = ?
		LIMIT 1
	`, b.ID).Scan(&seriesTitle, &position)
	if err == nil {
		nc.Series = seriesTitle
		nc.SeriesPosition = position
	}

	return nc
}
