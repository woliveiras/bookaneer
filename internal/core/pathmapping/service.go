package pathmapping

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jmoiron/sqlx"
)

var (
	ErrNotFound     = errors.New("remote path mapping not found")
	ErrInvalidInput = errors.New("invalid remote path mapping input")
)

// Service provides remote path mapping operations.
type Service struct {
	db *sqlx.DB
}

// New creates a new path mapping service.
func New(db *sqlx.DB) *Service {
	return &Service{db: db}
}

// List returns all remote path mappings.
func (s *Service) List(ctx context.Context) ([]RemotePathMapping, error) {
	var mappings []RemotePathMapping
	if err := s.db.SelectContext(ctx, &mappings, `
		SELECT id, host, remote_path, local_path, created_at
		FROM remote_path_mappings
		ORDER BY id
	`); err != nil {
		return nil, fmt.Errorf("list remote path mappings: %w", err)
	}
	return mappings, nil
}

// FindByID returns a mapping by ID.
func (s *Service) FindByID(ctx context.Context, id int64) (*RemotePathMapping, error) {
	var m RemotePathMapping
	err := s.db.GetContext(ctx, &m, `
		SELECT id, host, remote_path, local_path, created_at
		FROM remote_path_mappings WHERE id = ?
	`, id)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find remote path mapping %d: %w", id, err)
	}
	return &m, nil
}

// Create creates a new remote path mapping.
func (s *Service) Create(ctx context.Context, input CreateInput) (*RemotePathMapping, error) {
	if input.RemotePath == "" || input.LocalPath == "" {
		return nil, ErrInvalidInput
	}

	// Normalize paths to always end with separator
	input.RemotePath = normalizePath(input.RemotePath)
	input.LocalPath = normalizePath(input.LocalPath)

	result, err := s.db.NamedExecContext(ctx, `
		INSERT INTO remote_path_mappings (host, remote_path, local_path)
		VALUES (:host, :remote_path, :local_path)
	`, input)
	if err != nil {
		return nil, fmt.Errorf("create remote path mapping: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get mapping id: %w", err)
	}

	return s.FindByID(ctx, id)
}

// Update updates an existing mapping.
func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*RemotePathMapping, error) {
	existing, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	host := existing.Host
	remotePath := existing.RemotePath
	localPath := existing.LocalPath

	if input.Host != nil {
		host = *input.Host
	}
	if input.RemotePath != nil {
		if *input.RemotePath == "" {
			return nil, ErrInvalidInput
		}
		remotePath = normalizePath(*input.RemotePath)
	}
	if input.LocalPath != nil {
		if *input.LocalPath == "" {
			return nil, ErrInvalidInput
		}
		localPath = normalizePath(*input.LocalPath)
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE remote_path_mappings SET host = ?, remote_path = ?, local_path = ?
		WHERE id = ?
	`, host, remotePath, localPath, id)
	if err != nil {
		return nil, fmt.Errorf("update remote path mapping: %w", err)
	}

	return s.FindByID(ctx, id)
}

// Delete removes a mapping.
func (s *Service) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM remote_path_mappings WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete remote path mapping: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

// MapPath translates a remote path to a local path using the configured mappings.
// If no mapping matches, the original path is returned unchanged.
func (s *Service) MapPath(ctx context.Context, remotePath string) string {
	mappings, err := s.List(ctx)
	if err != nil || len(mappings) == 0 {
		return remotePath
	}

	for _, m := range mappings {
		if strings.HasPrefix(remotePath, m.RemotePath) {
			return filepath.Join(m.LocalPath, remotePath[len(m.RemotePath):])
		}
	}

	return remotePath
}

// normalizePath ensures a directory path ends with a path separator.
func normalizePath(p string) string {
	p = filepath.Clean(p)
	if !strings.HasSuffix(p, string(filepath.Separator)) {
		p += string(filepath.Separator)
	}
	return p
}
