package rootfolder

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// MoveRootFolder moves all files from the old root folder path to a new path.
// This is a synchronous operation that:
// 1. Creates the new directory if needed
// 2. Moves all author folders and their contents
// 3. Updates all paths in the database (authors, book_files)
// 4. Updates the root_folder path
func (s *Service) MoveRootFolder(ctx context.Context, id int64, newPath string) (*RootFolder, error) {
	existing, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldPath := existing.Path

	// Same path, nothing to do
	if oldPath == newPath {
		return existing, nil
	}

	// Ensure old path has no trailing slash for consistent comparisons
	oldPath = strings.TrimSuffix(oldPath, "/")
	newPath = strings.TrimSuffix(newPath, "/")

	slog.Info("Starting root folder migration", "id", id, "oldPath", oldPath, "newPath", newPath)

	// Step 1: Create new directory if it doesn't exist
	if err := os.MkdirAll(newPath, 0755); err != nil {
		return nil, fmt.Errorf("create new directory: %w", err)
	}

	// Step 2: Get all authors in this root folder
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, path FROM authors WHERE path LIKE ?
	`, oldPath+"%")
	if err != nil {
		return nil, fmt.Errorf("query authors: %w", err)
	}

	type authorPath struct {
		id      int64
		oldPath string
		newPath string
	}
	var authors []authorPath

	for rows.Next() {
		var id int64
		var path string
		if err := rows.Scan(&id, &path); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan author: %w", err)
		}
		// Calculate new path by replacing the root folder prefix
		newAuthorPath := strings.Replace(path, oldPath, newPath, 1)
		authors = append(authors, authorPath{id: id, oldPath: path, newPath: newAuthorPath})
	}
	_ = rows.Close()

	slog.Info("Found authors to migrate", "count", len(authors))

	// Step 3: Move each author folder
	for _, author := range authors {
		// Check if source exists
		if _, err := os.Stat(author.oldPath); os.IsNotExist(err) {
			slog.Warn("Author path does not exist, skipping", "path", author.oldPath)
			continue
		}

		// Create parent directory for new path
		if err := os.MkdirAll(filepath.Dir(author.newPath), 0755); err != nil {
			return nil, fmt.Errorf("create author directory: %w", err)
		}

		// Try rename first (fastest if same filesystem)
		err := os.Rename(author.oldPath, author.newPath)
		if err != nil {
			// If rename fails (cross-device), do a copy+delete
			slog.Debug("Rename failed, using copy", "error", err)
			if err := copyDir(author.oldPath, author.newPath); err != nil {
				return nil, fmt.Errorf("copy author folder %s: %w", author.oldPath, err)
			}
			// Remove old directory after successful copy
			if err := os.RemoveAll(author.oldPath); err != nil {
				slog.Warn("Failed to remove old author folder", "path", author.oldPath, "error", err)
			}
		}

		slog.Debug("Moved author folder", "from", author.oldPath, "to", author.newPath)
	}

	// Step 4: Update database paths in a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Update authors paths
	_, err = tx.ExecContext(ctx, `
		UPDATE authors 
		SET path = ? || SUBSTR(path, ?)
		WHERE path LIKE ?
	`, newPath, len(oldPath)+1, oldPath+"%")
	if err != nil {
		return nil, fmt.Errorf("update authors paths: %w", err)
	}

	// Update book_files paths
	_, err = tx.ExecContext(ctx, `
		UPDATE book_files 
		SET path = ? || SUBSTR(path, ?)
		WHERE path LIKE ?
	`, newPath, len(oldPath)+1, oldPath+"%")
	if err != nil {
		return nil, fmt.Errorf("update book_files paths: %w", err)
	}

	// Update root_folders path
	_, err = tx.ExecContext(ctx, `
		UPDATE root_folders SET path = ? WHERE id = ?
	`, newPath, id)
	if err != nil {
		return nil, fmt.Errorf("update root folder path: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	slog.Info("Root folder migration completed", "id", id, "newPath", newPath, "authorsMoved", len(authors))

	return s.FindByID(ctx, id)
}

// copyDir recursively copies a directory tree.
func copyDir(src string, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
