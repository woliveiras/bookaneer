// Package wanted provides services for searching and grabbing wanted books.
// It coordinates between metadata providers, digital libraries, indexers, and download clients.
package wanted

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/download"
	"github.com/woliveiras/bookaneer/internal/library"
	"github.com/woliveiras/bookaneer/internal/search"
)

// GrabResult represents the result of a grab attempt.
type GrabResult struct {
	BookID       int64  `json:"bookId"`
	Title        string `json:"title"`
	Source       string `json:"source"` // "library" or "indexer"
	ProviderName string `json:"providerName"`
	Format       string `json:"format"`
	Size         int64  `json:"size"`
	DownloadID   string `json:"downloadId"` // ID from download client
	ClientName   string `json:"clientName"`
}

// DownloadQueueItem represents an item in the download queue.
type DownloadQueueItem struct {
	ID               int64   `json:"id"`
	BookID           int64   `json:"bookId"`
	DownloadClientID int64   `json:"downloadClientId"`
	IndexerID        *int64  `json:"indexerId,omitempty"`
	ExternalID       string  `json:"externalId"`
	Title            string  `json:"title"`
	Size             int64   `json:"size"`
	Format           string  `json:"format"`
	Status           string  `json:"status"`
	Progress         float64 `json:"progress"`
	DownloadURL      string  `json:"downloadUrl"`
	AddedAt          string  `json:"addedAt"`
	BookTitle        string  `json:"bookTitle"`
	ClientName       string  `json:"clientName"`
}

// Service handles searching and grabbing wanted books.
type Service struct {
	db              *sql.DB
	bookService     *book.Service
	libraryService  *library.Aggregator
	searchService   *search.Service
	downloadService *download.Service
}

// New creates a new Wanted service.
func New(
	db *sql.DB,
	bookService *book.Service,
	libraryService *library.Aggregator,
	searchService *search.Service,
	downloadService *download.Service,
) *Service {
	return &Service{
		db:              db,
		bookService:     bookService,
		libraryService:  libraryService,
		searchService:   searchService,
		downloadService: downloadService,
	}
}

// SearchAndGrab searches for a book and grabs the best result.
func (s *Service) SearchAndGrab(ctx context.Context, bookID int64) (*GrabResult, error) {
	// Get book details
	b, err := s.bookService.FindByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("find book: %w", err)
	}

	if !b.Monitored {
		return nil, fmt.Errorf("book %d is not monitored", bookID)
	}

	slog.Info("Searching for book", "id", bookID, "title", b.Title)

	// Build search query
	query := b.Title
	if b.AuthorName != "" {
		query = fmt.Sprintf("%s %s", b.Title, b.AuthorName)
	}

	// Try digital libraries first (free, direct download)
	result, err := s.searchDigitalLibraries(ctx, b, query)
	if err == nil && result != nil {
		return result, nil
	}
	if err != nil {
		slog.Warn("Digital library search failed", "error", err)
	}

	// Fall back to indexers (torrent/usenet)
	result, err = s.searchIndexers(ctx, b, query)
	if err == nil && result != nil {
		return result, nil
	}
	if err != nil {
		slog.Warn("Indexer search failed", "error", err)
	}

	return nil, fmt.Errorf("no suitable download found for %q", b.Title)
}

// searchDigitalLibraries searches digital libraries and grabs the best result.
func (s *Service) searchDigitalLibraries(ctx context.Context, b *book.Book, query string) (*GrabResult, error) {
	if s.libraryService == nil {
		return nil, nil // Not configured
	}

	results, err := s.libraryService.Search(ctx, b.Title) // Use just title for library search
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	// Find best match (already sorted by score in aggregator)
	for _, r := range results {
		// Verify it's a downloadable format
		format := strings.ToLower(r.Format)
		if format != "epub" && format != "pdf" && format != "mobi" {
			continue
		}

		// Need a download URL
		if r.DownloadURL == "" {
			continue
		}

		// Grab it using direct downloader
		grabResult, err := s.grabFromLibrary(ctx, b, &r)
		if err != nil {
			slog.Warn("Failed to grab from library", "provider", r.Provider, "error", err)
			continue
		}

		return grabResult, nil
	}

	return nil, nil
}

// grabFromLibrary sends a library result to the direct downloader.
func (s *Service) grabFromLibrary(ctx context.Context, b *book.Book, r *library.SearchResult) (*GrabResult, error) {
	// Find a direct download client
	client, cfg, err := s.downloadService.GetDirectClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("no direct download client configured: %w", err)
	}

	// Build filename
	filename := fmt.Sprintf("%s - %s.%s", b.AuthorName, b.Title, r.Format)
	filename = sanitizeFilename(filename)

	// Add to download client
	downloadID, err := client.Add(ctx, download.AddItem{
		Name:        filename,
		DownloadURL: r.DownloadURL,
		Category:    "books",
	})
	if err != nil {
		return nil, fmt.Errorf("add to download client: %w", err)
	}

	// Record in download queue (use nil for embedded client)
	var clientIDPtr *int64
	if cfg.ID != 0 {
		clientIDPtr = &cfg.ID
	}
	if err := s.recordDownload(ctx, b.ID, clientIDPtr, nil, r.Title, r.Size, r.Format, r.DownloadURL, downloadID); err != nil {
		slog.Error("Failed to record download", "error", err)
	}

	// Record in history
	s.recordHistory(ctx, b.ID, b.AuthorID, "grabbed", r.Title, r.Format, map[string]any{
		"provider":   r.Provider,
		"downloadId": downloadID,
		"client":     cfg.Name,
	})

	slog.Info("Grabbed book from library",
		"book", b.Title,
		"provider", r.Provider,
		"format", r.Format,
		"client", cfg.Name,
	)

	return &GrabResult{
		BookID:       b.ID,
		Title:        r.Title,
		Source:       "library",
		ProviderName: r.Provider,
		Format:       r.Format,
		Size:         r.Size,
		DownloadID:   downloadID,
		ClientName:   cfg.Name,
	}, nil
}

// searchIndexers searches torrent/usenet indexers and grabs the best result.
func (s *Service) searchIndexers(ctx context.Context, b *book.Book, query string) (*GrabResult, error) {
	if s.searchService == nil {
		return nil, nil
	}

	results, err := s.searchService.Search(ctx, search.SearchQuery{Query: query})
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	// Filter for ebook formats
	var filtered []search.Result
	for _, r := range results {
		title := strings.ToLower(r.Title)
		if strings.Contains(title, "epub") ||
			strings.Contains(title, "pdf") ||
			strings.Contains(title, "mobi") ||
			strings.Contains(title, "ebook") {
			filtered = append(filtered, r)
		}
	}

	if len(filtered) == 0 {
		return nil, nil
	}

	// Try to grab the first suitable result
	for _, r := range filtered {
		grabResult, err := s.grabFromIndexer(ctx, b, &r)
		if err != nil {
			slog.Warn("Failed to grab from indexer", "indexer", r.IndexerName, "error", err)
			continue
		}
		return grabResult, nil
	}

	return nil, nil
}

// grabFromIndexer sends an indexer result to a download client.
func (s *Service) grabFromIndexer(ctx context.Context, b *book.Book, r *search.Result) (*GrabResult, error) {
	// Get appropriate download client based on result characteristics.
	// If seeders > 0, it's a torrent; otherwise assume usenet.
	var client download.Client
	var cfg *download.ClientConfig
	var err error

	if r.Seeders > 0 {
		client, cfg, err = s.downloadService.GetTorrentClient(ctx)
	} else {
		client, cfg, err = s.downloadService.GetUsenetClient(ctx)
	}
	if err != nil {
		return nil, fmt.Errorf("no suitable download client: %w", err)
	}
	if client == nil {
		return nil, fmt.Errorf("no download client configured for this result type")
	}

	// Add to download client
	downloadID, err := client.Add(ctx, download.AddItem{
		Name:        r.Title,
		DownloadURL: r.DownloadURL,
		Category:    "books",
	})
	if err != nil {
		return nil, fmt.Errorf("add to download client: %w", err)
	}

	// Determine format from title
	format := "unknown"
	titleLower := strings.ToLower(r.Title)
	switch {
	case strings.Contains(titleLower, "epub"):
		format = "epub"
	case strings.Contains(titleLower, "pdf"):
		format = "pdf"
	case strings.Contains(titleLower, "mobi"):
		format = "mobi"
	}

	// Get indexer ID
	indexerID := &r.IndexerID

	// Record in download queue (use nil for embedded client)
	var clientIDPtr *int64
	if cfg.ID != 0 {
		clientIDPtr = &cfg.ID
	}
	if err := s.recordDownload(ctx, b.ID, clientIDPtr, indexerID, r.Title, r.Size, format, r.DownloadURL, downloadID); err != nil {
		slog.Error("Failed to record download", "error", err)
	}

	// Record in history
	protocol := "torrent"
	if r.Seeders == 0 {
		protocol = "usenet"
	}
	s.recordHistory(ctx, b.ID, b.AuthorID, "grabbed", r.Title, format, map[string]any{
		"indexer":    r.IndexerName,
		"downloadId": downloadID,
		"client":     cfg.Name,
		"protocol":   protocol,
	})

	slog.Info("Grabbed book from indexer",
		"book", b.Title,
		"indexer", r.IndexerName,
		"client", cfg.Name,
	)

	return &GrabResult{
		BookID:       b.ID,
		Title:        r.Title,
		Source:       "indexer",
		ProviderName: r.IndexerName,
		Format:       format,
		Size:         r.Size,
		DownloadID:   downloadID,
		ClientName:   cfg.Name,
	}, nil
}

// recordDownload adds an entry to the download_queue table.
// clientID can be nil for embedded client (no database entry).
func (s *Service) recordDownload(ctx context.Context, bookID int64, clientID *int64, indexerID *int64, title string, size int64, format, downloadURL, externalID string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO download_queue (book_id, download_client_id, indexer_id, external_id, title, size, format, status, download_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'queued', ?)
	`, bookID, clientID, indexerID, externalID, title, size, format, downloadURL)
	return err
}

// recordHistory adds an entry to the history table.
func (s *Service) recordHistory(ctx context.Context, bookID, authorID int64, eventType, sourceTitle, quality string, data map[string]any) {
	dataJSON, _ := json.Marshal(data)
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO history (book_id, author_id, event_type, source_title, quality, data)
		VALUES (?, ?, ?, ?, ?, ?)
	`, bookID, authorID, eventType, sourceTitle, quality, string(dataJSON))
	if err != nil {
		slog.Error("Failed to record history", "error", err)
	}
}

// GetWantedBooks returns all monitored books without files.
func (s *Service) GetWantedBooks(ctx context.Context) ([]book.Book, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT b.id, b.author_id, b.title, b.sort_title, b.foreign_id, b.isbn, b.isbn13,
		       b.release_date, b.overview, b.image_url, b.page_count, b.monitored,
		       b.added_at, b.updated_at,
		       a.name as author_name,
		       EXISTS(SELECT 1 FROM book_files bf WHERE bf.book_id = b.id) as has_file
		FROM books b
		LEFT JOIN authors a ON a.id = b.author_id
		WHERE b.monitored = 1
		  AND NOT EXISTS (SELECT 1 FROM book_files bf WHERE bf.book_id = b.id)
		ORDER BY b.added_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var books []book.Book
	for rows.Next() {
		var b book.Book
		if err := rows.Scan(
			&b.ID, &b.AuthorID, &b.Title, &b.SortTitle, &b.ForeignID, &b.ISBN, &b.ISBN13,
			&b.ReleaseDate, &b.Overview, &b.ImageURL, &b.PageCount, &b.Monitored,
			&b.AddedAt, &b.UpdatedAt, &b.AuthorName, &b.HasFile,
		); err != nil {
			return nil, err
		}
		books = append(books, b)
	}

	return books, rows.Err()
}

// SearchAllWanted searches for all wanted books.
func (s *Service) SearchAllWanted(ctx context.Context) ([]GrabResult, error) {
	books, err := s.GetWantedBooks(ctx)
	if err != nil {
		return nil, fmt.Errorf("get wanted books: %w", err)
	}

	slog.Info("Searching for wanted books", "count", len(books))

	var results []GrabResult
	for _, b := range books {
		result, err := s.SearchAndGrab(ctx, b.ID)
		if err != nil {
			slog.Warn("Failed to grab book", "id", b.ID, "title", b.Title, "error", err)
			continue
		}
		if result != nil {
			results = append(results, *result)
		}

		// Small delay between searches to be nice to providers
		time.Sleep(500 * time.Millisecond)
	}

	return results, nil
}

// GetDownloadQueue returns the current download queue.
func (s *Service) GetDownloadQueue(ctx context.Context) ([]DownloadQueueItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT dq.id, dq.book_id, dq.download_client_id, dq.indexer_id, dq.external_id,
		       dq.title, dq.size, dq.format, dq.status, dq.progress, dq.download_url, dq.added_at,
		       b.title as book_title,
		       dc.name as client_name
		FROM download_queue dq
		LEFT JOIN books b ON b.id = dq.book_id
		LEFT JOIN download_clients dc ON dc.id = dq.download_client_id
		ORDER BY dq.added_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []DownloadQueueItem
	for rows.Next() {
		var item DownloadQueueItem
		var clientName sql.NullString
		if err := rows.Scan(
			&item.ID, &item.BookID, &item.DownloadClientID, &item.IndexerID, &item.ExternalID,
			&item.Title, &item.Size, &item.Format, &item.Status, &item.Progress, &item.DownloadURL, &item.AddedAt,
			&item.BookTitle, &clientName,
		); err != nil {
			return nil, err
		}
		if clientName.Valid {
			item.ClientName = clientName.String
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// UpdateQueueItemStatus updates the status of a queue item.
func (s *Service) UpdateQueueItemStatus(ctx context.Context, id int64, status string, progress float64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE download_queue SET status = ?, progress = ? WHERE id = ?`, status, progress, id)
	return err
}

// RemoveFromQueue removes an item from the download queue.
func (s *Service) RemoveFromQueue(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM download_queue WHERE id = ?`, id)
	return err
}

// GrabRelease manually grabs a release by URL and sends it to a download client.
func (s *Service) GrabRelease(ctx context.Context, bookID int64, downloadURL, releaseTitle string, size int64) (*GrabResult, error) {
	// Get book details
	b, err := s.bookService.FindByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("find book: %w", err)
	}

	// Find a direct download client (for HTTP URLs from digital libraries)
	client, cfg, err := s.downloadService.GetDirectClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get download client: %w", err)
	}
	if client == nil {
		return nil, fmt.Errorf("no download client configured")
	}

	// Build filename from release title or book info
	filename := releaseTitle
	if filename == "" {
		filename = fmt.Sprintf("%s - %s", b.AuthorName, b.Title)
	}
	filename = sanitizeFilename(filename)

	// Add to download client
	downloadID, err := client.Add(ctx, download.AddItem{
		Name:        filename,
		DownloadURL: downloadURL,
		Category:    "books",
	})
	if err != nil {
		return nil, fmt.Errorf("add to download client: %w", err)
	}

	// Determine format from URL or title
	format := "unknown"
	urlLower := strings.ToLower(downloadURL + releaseTitle)
	switch {
	case strings.Contains(urlLower, "epub"):
		format = "epub"
	case strings.Contains(urlLower, "pdf"):
		format = "pdf"
	case strings.Contains(urlLower, "mobi"):
		format = "mobi"
	}

	// Record in download queue (use nil for embedded client)
	var clientIDPtr *int64
	if cfg.ID != 0 {
		clientIDPtr = &cfg.ID
	}
	if err := s.recordDownload(ctx, b.ID, clientIDPtr, nil, releaseTitle, size, format, downloadURL, downloadID); err != nil {
		slog.Error("Failed to record download", "error", err)
	}

	// Record in history
	s.recordHistory(ctx, b.ID, b.AuthorID, "grabbed", releaseTitle, format, map[string]any{
		"downloadId": downloadID,
		"client":     cfg.Name,
		"manual":     true,
	})

	slog.Info("Manually grabbed release",
		"book", b.Title,
		"url", downloadURL,
		"client", cfg.Name,
	)

	return &GrabResult{
		BookID:       b.ID,
		Title:        releaseTitle,
		Source:       "manual",
		ProviderName: "manual",
		Format:       format,
		Size:         size,
		DownloadID:   downloadID,
		ClientName:   cfg.Name,
	}, nil
}

// ProcessDownloadsResult contains the results of processing downloads.
type ProcessDownloadsResult struct {
	Checked   int `json:"checked"`
	Completed int `json:"completed"`
	Failed    int `json:"failed"`
}

// ProcessDownloads checks active downloads and updates their status.
func (s *Service) ProcessDownloads(ctx context.Context) (*ProcessDownloadsResult, error) {
	result := &ProcessDownloadsResult{}

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
		ClientID int64
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
		client, _, err := s.downloadService.GetDirectClient(ctx)
		if err != nil || client == nil {
			slog.Warn("Could not get download client", "clientId", d.ClientID, "error", err)
			continue
		}

		status, err := client.GetStatus(ctx, d.ExtID)
		if err != nil {
			slog.Warn("Could not get download status", "externalId", d.ExtID, "error", err)
			continue
		}

		// Update status based on download client response
		newStatus := string(status.Status)
		if err := s.UpdateQueueItemStatus(ctx, d.ID, newStatus, status.Progress); err != nil {
			slog.Warn("Failed to update queue status", "id", d.ID, "error", err)
			continue
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
				}
			}
		case download.StatusFailed:
			result.Failed++
			slog.Warn("Download failed",
				"queueId", d.ID,
				"error", status.ErrorMessage,
			)
		}
	}

	return result, nil
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

	// Copy file to library (copy instead of move for safety)
	if err := copyFile(sourcePath, destPath); err != nil {
		return fmt.Errorf("copy file: %w", err)
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
