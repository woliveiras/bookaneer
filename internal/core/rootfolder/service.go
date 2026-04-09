package rootfolder

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// Service provides root folder-related operations.
type Service struct {
	db *sql.DB
}

// New creates a new root folder service.
func New(db *sql.DB) *Service {
	return &Service{db: db}
}

// FindByID returns a root folder by ID.
func (s *Service) FindByID(ctx context.Context, id int64) (*RootFolder, error) {
	var rf RootFolder
	err := s.db.QueryRowContext(ctx, `
		SELECT id, path, name, default_quality_profile_id
		FROM root_folders WHERE id = ?
	`, id).Scan(&rf.ID, &rf.Path, &rf.Name, &rf.DefaultQualityProfileID)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find root folder %d: %w", id, err)
	}

	s.enrichWithDiskInfo(&rf)
	s.enrichWithAuthorCount(ctx, &rf)

	return &rf, nil
}

// List returns all root folders.
func (s *Service) List(ctx context.Context) ([]RootFolder, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, path, name, default_quality_profile_id
		FROM root_folders ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("list root folders: %w", err)
	}
	defer rows.Close()

	var folders []RootFolder
	for rows.Next() {
		var rf RootFolder
		if err := rows.Scan(&rf.ID, &rf.Path, &rf.Name, &rf.DefaultQualityProfileID); err != nil {
			return nil, fmt.Errorf("scan root folder: %w", err)
		}
		folders = append(folders, rf)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate root folders: %w", err)
	}

	// Enrich after closing rows to avoid SQLite connection contention
	for i := range folders {
		s.enrichWithDiskInfo(&folders[i])
		s.enrichWithAuthorCount(ctx, &folders[i])
	}

	return folders, nil
}

// Create creates a new root folder.
func (s *Service) Create(ctx context.Context, input CreateRootFolderInput) (*RootFolder, error) {
	if input.Path == "" || input.Name == "" {
		return nil, ErrInvalidInput
	}

	// Create directory if it doesn't exist
	info, err := os.Stat(input.Path)
	if os.IsNotExist(err) {
		// Create the directory with appropriate permissions
		if err := os.MkdirAll(input.Path, 0755); err != nil {
			return nil, fmt.Errorf("create directory: %w", err)
		}
	} else if err != nil {
		return nil, ErrPathNotAccessible
	} else if !info.IsDir() {
		return nil, ErrPathNotAccessible
	}

	result, err := s.db.ExecContext(ctx, `
		INSERT INTO root_folders (path, name, default_quality_profile_id)
		VALUES (?, ?, ?)
	`, input.Path, input.Name, input.DefaultQualityProfileID)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("create root folder: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get root folder id: %w", err)
	}

	return s.FindByID(ctx, id)
}

// Update updates an existing root folder.
// If MoveFiles is true and Path is being changed, it will move all files to the new location.
func (s *Service) Update(ctx context.Context, id int64, input UpdateRootFolderInput) (*RootFolder, error) {
	existing, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// If MoveFiles is true and path is changing, use the migration flow
	if input.MoveFiles && input.Path != nil && *input.Path != existing.Path {
		return s.MoveRootFolder(ctx, id, *input.Path)
	}

	var sets []string
	var args []any

	if input.Path != nil {
		// Verify path exists and is accessible
		info, err := os.Stat(*input.Path)
		if err != nil && !os.IsNotExist(err) {
			return nil, ErrPathNotAccessible
		}
		// Create if doesn't exist
		if os.IsNotExist(err) {
			if err := os.MkdirAll(*input.Path, 0755); err != nil {
				return nil, fmt.Errorf("create directory: %w", err)
			}
		} else if !info.IsDir() {
			return nil, ErrPathNotAccessible
		}
		sets = append(sets, "path = ?")
		args = append(args, *input.Path)
	}
	if input.Name != nil {
		sets = append(sets, "name = ?")
		args = append(args, *input.Name)
	}
	if input.DefaultQualityProfileID != nil {
		sets = append(sets, "default_quality_profile_id = ?")
		args = append(args, *input.DefaultQualityProfileID)
	}

	if len(sets) == 0 {
		return existing, nil
	}

	args = append(args, id)
	query := fmt.Sprintf("UPDATE root_folders SET %s WHERE id = ?", strings.Join(sets, ", "))
	_, err = s.db.ExecContext(ctx, query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("update root folder %d: %w", id, err)
	}

	return s.FindByID(ctx, id)
}

// Delete deletes a root folder by ID.
func (s *Service) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM root_folders WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete root folder %d: %w", id, err)
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

// enrichWithDiskInfo adds disk usage info to a root folder.
func (s *Service) enrichWithDiskInfo(rf *RootFolder) {
	var stat unix.Statfs_t
	if err := unix.Statfs(rf.Path, &stat); err != nil {
		rf.Accessible = false
		return
	}

	rf.Accessible = true
	rf.FreeSpace = int64(stat.Bavail) * int64(stat.Bsize)
	rf.TotalSpace = int64(stat.Blocks) * int64(stat.Bsize)
}

// enrichWithAuthorCount adds the count of authors in this root folder.
func (s *Service) enrichWithAuthorCount(ctx context.Context, rf *RootFolder) {
	var count int
	// Count authors whose path starts with this root folder path
	_ = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM authors WHERE path LIKE ?
	`, rf.Path+"%").Scan(&count)
	rf.AuthorCount = count
}

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
			rows.Close()
			return nil, fmt.Errorf("scan author: %w", err)
		}
		// Calculate new path by replacing the root folder prefix
		newAuthorPath := strings.Replace(path, oldPath, newPath, 1)
		authors = append(authors, authorPath{id: id, oldPath: path, newPath: newAuthorPath})
	}
	rows.Close()

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
	defer tx.Rollback()

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
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
