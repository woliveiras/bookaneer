package wanted

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/woliveiras/bookaneer/internal/download"
)

// ProcessDownloadsResult contains the results of processing downloads.
type ProcessDownloadsResult struct {
	Checked   int `json:"checked"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
	Imported  int `json:"imported"`
}

// ProcessDownloads checks active downloads and updates their status.
func (s *Service) ProcessDownloads(ctx context.Context) (*ProcessDownloadsResult, error) {
	result := &ProcessDownloadsResult{}

	// First, process any completed downloads that have a save_path but weren't imported
	// This handles server restarts where the in-memory download state was lost
	if imported, err := s.importPendingCompletedDownloads(ctx); err != nil {
		slog.Warn("Failed to import pending downloads", "error", err)
	} else {
		result.Imported = imported
	}

	// Get active downloads (queued, downloading, paused)
	rows, err := s.db.QueryContext(ctx, `
		SELECT q.id, q.download_client_id, q.external_id, q.status
		FROM download_queue q
		WHERE q.status IN ('queued', 'downloading', 'paused', 'sent')
	`)
	if err != nil {
		return nil, fmt.Errorf("query active downloads: %w", err)
	}
	defer rows.Close()

	type activeDownload struct {
		ID       int64
		ClientID sql.NullInt64
		ExtID    string
		Status   string
	}

	var downloads []activeDownload
	for rows.Next() {
		var d activeDownload
		if err := rows.Scan(&d.ID, &d.ClientID, &d.ExtID, &d.Status); err != nil {
			continue
		}
		downloads = append(downloads, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate downloads: %w", err)
	}

	result.Checked = len(downloads)

	// Check status of each download
	for _, d := range downloads {
		// Get appropriate client - use embedded client for NULL clientID
		client, _, err := s.downloadService.GetDirectClient(ctx)
		if err != nil || client == nil {
			slog.Warn("Could not get download client", "queueId", d.ID, "error", err)
			continue
		}

		status, err := client.GetStatus(ctx, d.ExtID)
		if err != nil {
			// Download not found in client - probably lost after restart
			// Try to restart the download
			slog.Info("Restarting lost download", "queueId", d.ID, "externalId", d.ExtID)
			if err := s.restartDownload(ctx, d.ID, client); err != nil {
				slog.Warn("Failed to restart download", "queueId", d.ID, "error", err)
			}
			continue
		}

		// Update status based on download client response (including save_path)
		newStatus := string(status.Status)
		if status.SavePath != "" {
			if err := s.UpdateQueueItemStatusWithPath(ctx, d.ID, newStatus, status.Progress, status.SavePath); err != nil {
				slog.Warn("Failed to update queue status", "id", d.ID, "error", err)
				continue
			}
		} else {
			if err := s.UpdateQueueItemStatus(ctx, d.ID, newStatus, status.Progress); err != nil {
				slog.Warn("Failed to update queue status", "id", d.ID, "error", err)
				continue
			}
		}

		switch status.Status {
		case download.StatusCompleted:
			result.Completed++
			// Import file to library
			if status.SavePath != "" {
				if err := s.importCompletedDownload(ctx, d.ID, status.SavePath); err != nil {
					slog.Warn("Failed to import download",
						"queueId", d.ID,
						"path", status.SavePath,
						"error", err,
					)
				} else {
					slog.Info("Download imported to library",
						"queueId", d.ID,
						"path", status.SavePath,
					)
					result.Imported++

					// Clean up search results after successful import
					s.cleanupSearchResults(ctx, d.ID)
				}
			}
		case download.StatusFailed:
			result.Failed++
			slog.Warn("Download failed",
				"queueId", d.ID,
				"error", status.ErrorMessage,
			)

			// Try next available source automatically
			if retried := s.tryNextSource(ctx, d.ID, status.ErrorMessage); retried {
				slog.Info("Automatically trying next download source", "queueId", d.ID)
			}
		}
	}

	return result, nil
}

// importPendingCompletedDownloads imports downloads that completed but weren't imported
// (e.g., because the server restarted before import could happen).
func (s *Service) importPendingCompletedDownloads(ctx context.Context) (int, error) {
	// Find completed downloads with save_path that haven't been imported yet
	// (not imported = no entry in book_files for that book_id)
	rows, err := s.db.QueryContext(ctx, `
		SELECT q.id, q.book_id, q.save_path
		FROM download_queue q
		WHERE q.status = 'completed'
		  AND q.save_path != ''
		  AND NOT EXISTS (SELECT 1 FROM book_files bf WHERE bf.book_id = q.book_id)
	`)
	if err != nil {
		return 0, fmt.Errorf("query pending imports: %w", err)
	}

	// Collect all pending imports first, then close rows before processing
	// This avoids SQLite lock issues when doing writes during iteration
	type pendingImport struct {
		queueID  int64
		bookID   int64
		savePath string
	}
	var pending []pendingImport
	for rows.Next() {
		var p pendingImport
		if err := rows.Scan(&p.queueID, &p.bookID, &p.savePath); err != nil {
			slog.Warn("Failed to scan pending import", "error", err)
			continue
		}
		pending = append(pending, p)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return 0, fmt.Errorf("iterate pending imports: %w", err)
	}
	rows.Close() // Close before processing to avoid SQLite locks

	var imported int
	for _, p := range pending {
		// Check if file still exists
		if _, err := os.Stat(p.savePath); os.IsNotExist(err) {
			slog.Warn("Download file no longer exists, marking as failed",
				"queueId", p.queueID,
				"path", p.savePath,
			)
			_ = s.UpdateQueueItemStatus(ctx, p.queueID, "failed", 0)
			continue
		}

		// Import the download
		if err := s.importCompletedDownload(ctx, p.queueID, p.savePath); err != nil {
			slog.Warn("Failed to import pending download",
				"queueId", p.queueID,
				"path", p.savePath,
				"error", err,
			)
		} else {
			slog.Info("Successfully imported pending download",
				"queueId", p.queueID,
				"path", p.savePath,
			)
			imported++
		}
	}

	return imported, nil
}

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

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// hashFile calculates SHA256 hash of a file.
func hashFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}

// sanitizeFilename removes unsafe characters from filename.
func sanitizeFilename(name string) string {
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_",
		"?", "_", "\"", "_", "<", "_", ">", "_", "|", "_",
	)
	return replacer.Replace(name)
}
