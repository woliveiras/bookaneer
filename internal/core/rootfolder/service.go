package rootfolder

import (
	"context"
	"database/sql"
	"fmt"
	"os"
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
	defer func() { _ = rows.Close() }()

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
