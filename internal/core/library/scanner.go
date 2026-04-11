package library

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// SupportedFormats lists the ebook formats recognized by the scanner.
var SupportedFormats = map[string]bool{
	".epub": true,
	".mobi": true,
	".azw3": true,
	".pdf":  true,
	".cbz":  true,
}

// ScanResult holds the results of a library scan.
type ScanResult struct {
	TotalFiles   int      `json:"totalFiles"`
	NewFiles     int      `json:"newFiles"`
	UpdatedFiles int      `json:"updatedFiles"`
	RemovedFiles int      `json:"removedFiles"`
	Errors       []string `json:"errors,omitempty"`
}

// Scanner scans the library for ebook files.
type Scanner struct {
	db *sql.DB
}

// NewScanner creates a new library scanner.
func NewScanner(db *sql.DB) *Scanner {
	return &Scanner{db: db}
}

// ScanRootFolder scans a root folder for ebook files.
func (s *Scanner) ScanRootFolder(ctx context.Context, rootFolderID int64) (*ScanResult, error) {
	var rootPath string
	err := s.db.QueryRowContext(ctx, "SELECT path FROM root_folders WHERE id = ?", rootFolderID).Scan(&rootPath)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("root folder not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get root folder: %w", err)
	}
	return s.ScanPath(ctx, rootPath)
}

// ScanPath scans a path for ebook files.
func (s *Scanner) ScanPath(ctx context.Context, rootPath string) (*ScanResult, error) {
	result := &ScanResult{}

	existingFiles := make(map[string]struct{})
	rows, err := s.db.QueryContext(ctx, "SELECT path FROM book_files WHERE path LIKE ?", rootPath+"%")
	if err != nil {
		return nil, fmt.Errorf("get existing files: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, fmt.Errorf("scan file path: %w", err)
		}
		existingFiles[path] = struct{}{}
	}

	foundFiles := make(map[string]struct{})
	err = filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("walk error: %s: %v", path, err))
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") && d.Name() != "." {
				return fs.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !SupportedFormats[ext] {
			return nil
		}

		result.TotalFiles++
		foundFiles[path] = struct{}{}

		if _, exists := existingFiles[path]; exists {
			if err := s.updateFileIfChanged(ctx, path); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("update %s: %v", path, err))
			} else {
				result.UpdatedFiles++
			}
		} else {
			if err := s.addNewFile(ctx, rootPath, path); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("add %s: %v", path, err))
			} else {
				result.NewFiles++
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk directory: %w", err)
	}

	for existingPath := range existingFiles {
		if _, found := foundFiles[existingPath]; !found {
			if err := s.removeFile(ctx, existingPath); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("remove %s: %v", existingPath, err))
			} else {
				result.RemovedFiles++
			}
		}
	}

	return result, nil
}

// ScanAuthorFolder scans a specific author folder.
func (s *Scanner) ScanAuthorFolder(ctx context.Context, authorID int64) (*ScanResult, error) {
	var authorPath string
	err := s.db.QueryRowContext(ctx, "SELECT path FROM authors WHERE id = ?", authorID).Scan(&authorPath)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("author not found")
	}
	if err != nil {
		return nil, fmt.Errorf("get author: %w", err)
	}
	return s.ScanPath(ctx, authorPath)
}

func (s *Scanner) addNewFile(ctx context.Context, rootPath, filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	relativePath, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		relativePath = filePath
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	format := strings.TrimPrefix(ext, ".")

	hash := ""
	if info.Size() < 50*1024*1024 {
		hash, _ = hashFile(filePath)
	}

	bookID, err := s.matchFileToBook(ctx, filePath)
	if err != nil {
		slog.Debug("could not match file to book", "path", filePath, "error", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO book_files (book_id, path, relative_path, size, format, quality, hash)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, bookID, filePath, relativePath, info.Size(), format, format, hash)
	if err != nil {
		return fmt.Errorf("insert file: %w", err)
	}

	return nil
}

func (s *Scanner) updateFileIfChanged(ctx context.Context, filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	var storedSize int64
	err = s.db.QueryRowContext(ctx, "SELECT size FROM book_files WHERE path = ?", filePath).Scan(&storedSize)
	if err != nil {
		return fmt.Errorf("get stored size: %w", err)
	}

	if storedSize == info.Size() {
		return nil
	}

	hash := ""
	if info.Size() < 50*1024*1024 {
		hash, _ = hashFile(filePath)
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE book_files SET size = ?, hash = ? WHERE path = ?
	`, info.Size(), hash, filePath)
	if err != nil {
		return fmt.Errorf("update file: %w", err)
	}

	return nil
}

func (s *Scanner) removeFile(ctx context.Context, filePath string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM book_files WHERE path = ?", filePath)
	if err != nil {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

func (s *Scanner) matchFileToBook(ctx context.Context, filePath string) (*int64, error) {
	dir := filepath.Dir(filePath)
	fileName := filepath.Base(filePath)
	ext := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, ext)

	var bookID int64
	err := s.db.QueryRowContext(ctx, `
		SELECT b.id FROM books b
		JOIN authors a ON b.author_id = a.id
		WHERE (b.title = ? OR b.sort_title = ?)
		AND a.path = ?
		LIMIT 1
	`, baseName, baseName, dir).Scan(&bookID)
	if err == sql.ErrNoRows {
		bookDir := filepath.Base(dir)
		authorDir := filepath.Dir(dir)
		err = s.db.QueryRowContext(ctx, `
			SELECT b.id FROM books b
			JOIN authors a ON b.author_id = a.id
			WHERE (b.title = ? OR b.sort_title = ?)
			AND a.path = ?
			LIMIT 1
		`, bookDir, bookDir, authorDir).Scan(&bookID)
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &bookID, nil
}

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
