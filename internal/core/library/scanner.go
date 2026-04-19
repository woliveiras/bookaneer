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

	"github.com/jmoiron/sqlx"

	"github.com/woliveiras/bookaneer/internal/core/mediafile"
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
	db *sqlx.DB
}

// NewScanner creates a new library scanner.
func NewScanner(db *sqlx.DB) *Scanner {
	return &Scanner{db: db}
}

// ScanRootFolder scans a root folder for ebook files.
func (s *Scanner) ScanRootFolder(ctx context.Context, rootFolderID int64) (*ScanResult, error) {
	var rootPath string
	err := s.db.GetContext(ctx, &rootPath, "SELECT path FROM root_folders WHERE id = ?", rootFolderID)
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
	var paths []string
	if err := s.db.SelectContext(ctx, &paths, "SELECT path FROM book_files WHERE path LIKE ?", rootPath+"%"); err != nil {
		return nil, fmt.Errorf("get existing files: %w", err)
	}
	for _, path := range paths {
		existingFiles[path] = struct{}{}
	}

	foundFiles := make(map[string]struct{})
	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
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
	err := s.db.GetContext(ctx, &authorPath, "SELECT path FROM authors WHERE id = ?", authorID)
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

		discoveredID, discoverErr := s.discoverBookFromFile(ctx, rootPath, filePath)
		if discoverErr != nil {
			return fmt.Errorf("discover book from file: %w", discoverErr)
		}
		bookID = &discoveredID
	}

	_, err = s.db.NamedExecContext(ctx, `
		INSERT INTO book_files (book_id, path, relative_path, size, format, quality, hash)
		VALUES (:book_id, :path, :relative_path, :size, :format, :quality, :hash)
	`, map[string]any{
		"book_id":       bookID,
		"path":          filePath,
		"relative_path": relativePath,
		"size":          info.Size(),
		"format":        format,
		"quality":       format,
		"hash":          hash,
	})
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
	if err := s.db.GetContext(ctx, &storedSize, "SELECT size FROM book_files WHERE path = ?", filePath); err != nil {
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
	err := s.db.GetContext(ctx, &bookID, `
		SELECT b.id FROM books b
		JOIN authors a ON b.author_id = a.id
		WHERE (b.title = ? OR b.sort_title = ?)
		AND a.path = ?
		LIMIT 1
	`, baseName, baseName, dir)
	if err == sql.ErrNoRows {
		bookDir := filepath.Base(dir)
		authorDir := filepath.Dir(dir)
		if err = s.db.GetContext(ctx, &bookID, `
			SELECT b.id FROM books b
			JOIN authors a ON b.author_id = a.id
			WHERE (b.title = ? OR b.sort_title = ?)
			AND a.path = ?
			LIMIT 1
		`, bookDir, bookDir, authorDir); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &bookID, nil
}

// discoverBookFromFile parses directory structure and EPUB metadata to
// auto-create author and book records for files not yet known to the library.
func (s *Scanner) discoverBookFromFile(ctx context.Context, rootPath, filePath string) (int64, error) {
	authorName, bookTitle := parsePathForBookInfo(rootPath, filePath)

	// Enrich with EPUB metadata when available.
	if strings.ToLower(filepath.Ext(filePath)) == ".epub" {
		meta, err := mediafile.ExtractMetadata(filePath)
		if err == nil && meta != nil {
			if meta.Title != "" {
				bookTitle = meta.Title
			}
			if len(meta.Authors) > 0 && meta.Authors[0] != "" {
				authorName = meta.Authors[0]
			}
		}
	}

	if authorName == "" {
		authorName = "Unknown Author"
	}
	if bookTitle == "" {
		bookTitle = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	}

	authorID, err := s.findOrCreateAuthor(ctx, authorName, rootPath)
	if err != nil {
		return 0, fmt.Errorf("find or create author %q: %w", authorName, err)
	}

	bookID, err := s.findOrCreateBook(ctx, authorID, bookTitle, filePath)
	if err != nil {
		return 0, fmt.Errorf("find or create book %q: %w", bookTitle, err)
	}

	return bookID, nil
}

// parsePathForBookInfo extracts author name and book title from a file's path
// relative to the root folder. Supports these layouts:
//
//	Root/Author/Book.epub          → author from directory, title from filename
//	Root/Author/BookTitle/Book.epub → author from first dir, title from second dir
//	Root/Author - Title.epub       → parsed from "Author - Title" filename
//	Root/Book.epub                 → no author, title from filename
func parsePathForBookInfo(rootPath, filePath string) (authorName, bookTitle string) {
	rel, err := filepath.Rel(rootPath, filePath)
	if err != nil {
		return "", ""
	}

	parts := strings.Split(filepath.ToSlash(rel), "/")
	fileName := parts[len(parts)-1]
	ext := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, ext)

	switch len(parts) {
	case 1:
		// Root/BookFile.epub — try "Author - Title" pattern in filename.
		if idx := strings.Index(baseName, " - "); idx > 0 {
			return baseName[:idx], baseName[idx+3:]
		}
		return "", baseName
	case 2:
		// Root/AuthorName/BookFile.epub
		return parts[0], baseName
	default:
		// Root/AuthorName/BookTitle/BookFile.epub (or deeper)
		return parts[0], parts[1]
	}
}

func (s *Scanner) findOrCreateAuthor(ctx context.Context, name, rootPath string) (int64, error) {
	var authorID int64
	err := s.db.GetContext(ctx, &authorID,
		`SELECT id FROM authors WHERE LOWER(name) = LOWER(?) LIMIT 1`, name)
	if err == nil {
		return authorID, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("find author: %w", err)
	}

	authorPath := filepath.Join(rootPath, name)
	result, err := s.db.NamedExecContext(ctx, `
		INSERT INTO authors (name, sort_name, status, monitored, path)
		VALUES (:name, :sort_name, 'active', 1, :path)
	`, map[string]any{
		"name":      name,
		"sort_name": makeSortName(name),
		"path":      authorPath,
	})
	if err != nil {
		// UNIQUE race: author may have been created concurrently.
		if strings.Contains(err.Error(), "UNIQUE") {
			if err2 := s.db.GetContext(ctx, &authorID,
				`SELECT id FROM authors WHERE LOWER(name) = LOWER(?) LIMIT 1`, name); err2 == nil {
				return authorID, nil
			}
		}
		return 0, fmt.Errorf("create author: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get author id: %w", err)
	}

	slog.Info("discovered author from library scan", "name", name, "id", id)
	return id, nil
}

func (s *Scanner) findOrCreateBook(ctx context.Context, authorID int64, title, filePath string) (int64, error) {
	var bookID int64
	err := s.db.GetContext(ctx, &bookID, `
		SELECT id FROM books
		WHERE author_id = ? AND (LOWER(title) = LOWER(?) OR LOWER(sort_title) = LOWER(?))
		LIMIT 1
	`, authorID, title, title)
	if err == nil {
		return bookID, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("find book: %w", err)
	}

	// Optionally grab ISBN from EPUB metadata for the new record.
	var isbn string
	if strings.ToLower(filepath.Ext(filePath)) == ".epub" {
		meta, metaErr := mediafile.ExtractMetadata(filePath)
		if metaErr == nil && meta != nil {
			isbn = meta.ISBN
		}
	}

	result, err := s.db.NamedExecContext(ctx, `
		INSERT INTO books (author_id, title, sort_title, isbn, monitored)
		VALUES (:author_id, :title, :sort_title, :isbn, 1)
	`, map[string]any{
		"author_id":  authorID,
		"title":      title,
		"sort_title": title,
		"isbn":       isbn,
	})
	if err != nil {
		return 0, fmt.Errorf("create book: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("get book id: %w", err)
	}

	slog.Info("discovered book from library scan", "title", title, "authorId", authorID, "id", id)
	return id, nil
}

// makeSortName converts "J.R.R. Tolkien" → "Tolkien, J.R.R.".
func makeSortName(name string) string {
	parts := strings.Fields(name)
	if len(parts) < 2 {
		return name
	}
	last := parts[len(parts)-1]
	rest := strings.Join(parts[:len(parts)-1], " ")
	return last + ", " + rest
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
