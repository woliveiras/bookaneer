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
	DownloadClientID *int64  `json:"downloadClientId,omitempty"`
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

// HistoryItem represents a history event.
type HistoryItem struct {
	ID          int64          `json:"id"`
	BookID      *int64         `json:"bookId,omitempty"`
	AuthorID    *int64         `json:"authorId,omitempty"`
	EventType   string         `json:"eventType"`
	SourceTitle string         `json:"sourceTitle"`
	Quality     string         `json:"quality"`
	Data        map[string]any `json:"data"`
	Date        string         `json:"date"`
	BookTitle   string         `json:"bookTitle,omitempty"`
	AuthorName  string         `json:"authorName,omitempty"`
}

// BlocklistItem represents a blocked release.
type BlocklistItem struct {
	ID          int64  `json:"id"`
	BookID      int64  `json:"bookId"`
	SourceTitle string `json:"sourceTitle"`
	Quality     string `json:"quality"`
	Reason      string `json:"reason"`
	Date        string `json:"date"`
	BookTitle   string `json:"bookTitle"`
	AuthorName  string `json:"authorName"`
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

// GetBookInfo returns book title and author for display purposes.
func (s *Service) GetBookInfo(ctx context.Context, bookID int64) (title, authorName string, err error) {
	b, err := s.bookService.FindByID(ctx, bookID)
	if err != nil {
		return "", "", err
	}
	return b.Title, b.AuthorName, nil
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

	// Check if there's already an active download for this book
	var activeCount int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM download_queue 
		WHERE book_id = ? AND status IN ('queued', 'downloading', 'paused', 'importing')
	`, bookID).Scan(&activeCount)
	if err != nil {
		slog.Warn("Failed to check for active downloads", "error", err)
	} else if activeCount > 0 {
		slog.Info("Skipping search - book already has active download", "id", bookID, "title", b.Title)
		return nil, fmt.Errorf("book already has an active download in queue")
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

// searchDigitalLibraries searches digital libraries and saves all results for fallback.
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

	// Clear any previous search results for this book
	_, _ = s.db.ExecContext(ctx, `DELETE FROM search_results WHERE book_id = ?`, b.ID)

	// Filter and save all valid results for fallback
	var validResults []library.SearchResult
	priority := 0
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

		validResults = append(validResults, r)

		// Save to search_results table for fallback
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO search_results (book_id, provider, title, download_url, format, size, score, priority, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'pending')
		`, b.ID, r.Provider, r.Title, r.DownloadURL, r.Format, r.Size, r.Score, priority)
		if err != nil {
			slog.Warn("Failed to save search result", "error", err)
		}
		priority++
	}

	if len(validResults) == 0 {
		return nil, nil
	}

	slog.Info("Found download sources", "book", b.Title, "count", len(validResults))

	// Try to grab the first (best) result
	return s.grabNextSearchResult(ctx, b)
}

// grabNextSearchResult tries to grab the next pending search result for a book.
func (s *Service) grabNextSearchResult(ctx context.Context, b *book.Book) (*GrabResult, error) {
	// Get next pending result
	var resultID int64
	var provider, title, downloadURL, format string
	var size int64

	err := s.db.QueryRowContext(ctx, `
		SELECT id, provider, title, download_url, format, size 
		FROM search_results 
		WHERE book_id = ? AND status = 'pending'
		ORDER BY priority ASC
		LIMIT 1
	`, b.ID).Scan(&resultID, &provider, &title, &downloadURL, &format, &size)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no more download sources available for %q", b.Title)
		}
		return nil, fmt.Errorf("get next search result: %w", err)
	}

	// Mark as tried
	_, _ = s.db.ExecContext(ctx, `
		UPDATE search_results SET status = 'tried', tried_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		WHERE id = ?
	`, resultID)

	// Build library result struct
	r := &library.SearchResult{
		Provider:    provider,
		Title:       title,
		DownloadURL: downloadURL,
		Format:      format,
		Size:        size,
	}

	// Try to grab it
	grabResult, err := s.grabFromLibrary(ctx, b, r)
	if err != nil {
		// Mark as failed and record error
		_, _ = s.db.ExecContext(ctx, `
			UPDATE search_results SET status = 'failed', error_message = ? WHERE id = ?
		`, err.Error(), resultID)
		return nil, err
	}

	// Mark as success
	_, _ = s.db.ExecContext(ctx, `UPDATE search_results SET status = 'success' WHERE id = ?`, resultID)

	return grabResult, nil
}

// grabFromLibrary sends a library result to the direct downloader.
func (s *Service) grabFromLibrary(ctx context.Context, b *book.Book, r *library.SearchResult) (*GrabResult, error) {
	// Find a direct download client
	client, cfg, err := s.downloadService.GetDirectClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("no direct download client configured: %w", err)
	}

	// Build author folder and filename
	// Structure: RootFolder/AuthorName/AuthorName - BookTitle.format
	authorFolder := sanitizeFilename(b.AuthorName)
	filename := fmt.Sprintf("%s - %s.%s", sanitizeFilename(b.AuthorName), sanitizeFilename(b.Title), r.Format)

	// Build expected save path (including author folder)
	authorDir := filepath.Join(cfg.DownloadDir, authorFolder)
	expectedSavePath := filepath.Join(authorDir, filename)

	// Add to download client with the full path including author folder
	downloadID, err := client.Add(ctx, download.AddItem{
		Name:        filename,
		DownloadURL: r.DownloadURL,
		Category:    "books",
		SavePath:    authorDir, // Tell client to save in author folder
	})
	if err != nil {
		return nil, fmt.Errorf("add to download client: %w", err)
	}

	// Record in download queue with expected save path (use nil for embedded client)
	var clientIDPtr *int64
	if cfg.ID != 0 {
		clientIDPtr = &cfg.ID
	}
	if err := s.recordDownload(ctx, b.ID, clientIDPtr, nil, r.Title, r.Size, r.Format, r.DownloadURL, downloadID, expectedSavePath); err != nil {
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

	// For indexer downloads, we don't know the exact save path yet (depends on client)
	// Record in download queue (use nil for embedded client)
	var clientIDPtr *int64
	if cfg.ID != 0 {
		clientIDPtr = &cfg.ID
	}
	if err := s.recordDownload(ctx, b.ID, clientIDPtr, indexerID, r.Title, r.Size, format, r.DownloadURL, downloadID, ""); err != nil {
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
func (s *Service) recordDownload(ctx context.Context, bookID int64, clientID *int64, indexerID *int64, title string, size int64, format, downloadURL, externalID, savePath string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO download_queue (book_id, download_client_id, indexer_id, external_id, title, size, format, status, download_url, save_path)
		VALUES (?, ?, ?, ?, ?, ?, ?, 'queued', ?, ?)
	`, bookID, clientID, indexerID, externalID, title, size, format, downloadURL, savePath)
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

// GetHistory returns recent history events.
func (s *Service) GetHistory(ctx context.Context, limit int, eventType string) ([]HistoryItem, error) {
	if limit <= 0 {
		limit = 50
	}

	query := `
		SELECT h.id, h.book_id, h.author_id, h.event_type, h.source_title, h.quality, h.data, h.date,
		       COALESCE(b.title, '') as book_title,
		       COALESCE(a.name, '') as author_name
		FROM history h
		LEFT JOIN books b ON b.id = h.book_id
		LEFT JOIN authors a ON a.id = h.author_id
	`
	var args []any
	if eventType != "" {
		query += " WHERE h.event_type = ?"
		args = append(args, eventType)
	}
	query += " ORDER BY h.date DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []HistoryItem
	for rows.Next() {
		var item HistoryItem
		var bookID, authorID sql.NullInt64
		var dataJSON string
		if err := rows.Scan(&item.ID, &bookID, &authorID, &item.EventType, &item.SourceTitle, &item.Quality, &dataJSON, &item.Date, &item.BookTitle, &item.AuthorName); err != nil {
			return nil, err
		}
		if bookID.Valid {
			item.BookID = &bookID.Int64
		}
		if authorID.Valid {
			item.AuthorID = &authorID.Int64
		}
		json.Unmarshal([]byte(dataJSON), &item.Data)
		items = append(items, item)
	}

	return items, rows.Err()
}

// GetBlocklist returns all blocklisted releases.
func (s *Service) GetBlocklist(ctx context.Context) ([]BlocklistItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT bl.id, bl.book_id, bl.source_title, bl.quality, bl.reason, bl.date,
		       COALESCE(b.title, '') as book_title,
		       COALESCE(a.name, '') as author_name
		FROM blocklist bl
		LEFT JOIN books b ON b.id = bl.book_id
		LEFT JOIN authors a ON a.id = b.author_id
		ORDER BY bl.date DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []BlocklistItem
	for rows.Next() {
		var item BlocklistItem
		if err := rows.Scan(&item.ID, &item.BookID, &item.SourceTitle, &item.Quality, &item.Reason, &item.Date, &item.BookTitle, &item.AuthorName); err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// AddToBlocklist adds a release to the blocklist.
func (s *Service) AddToBlocklist(ctx context.Context, bookID int64, sourceTitle, quality, reason string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO blocklist (book_id, source_title, quality, reason)
		VALUES (?, ?, ?, ?)
	`, bookID, sourceTitle, quality, reason)
	return err
}

// RemoveFromBlocklist removes an item from the blocklist.
func (s *Service) RemoveFromBlocklist(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM blocklist WHERE id = ?`, id)
	return err
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
		var clientID sql.NullInt64
		var indexerID sql.NullInt64
		var bookTitle sql.NullString
		var clientName sql.NullString
		if err := rows.Scan(
			&item.ID, &item.BookID, &clientID, &indexerID, &item.ExternalID,
			&item.Title, &item.Size, &item.Format, &item.Status, &item.Progress, &item.DownloadURL, &item.AddedAt,
			&bookTitle, &clientName,
		); err != nil {
			return nil, err
		}
		if clientID.Valid {
			item.DownloadClientID = &clientID.Int64
		}
		if indexerID.Valid {
			item.IndexerID = &indexerID.Int64
		}
		if bookTitle.Valid {
			item.BookTitle = bookTitle.String
		} else {
			item.BookTitle = item.Title // Fallback to release title
		}
		if clientName.Valid {
			item.ClientName = clientName.String
		} else {
			item.ClientName = "Embedded Downloader"
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// For embedded client items, get real-time status from the direct client
	client, _, err := s.downloadService.GetDirectClient(ctx)
	if err == nil && client != nil {
		for i := range items {
			// Check items with no client ID (embedded) that have active statuses
			if items[i].DownloadClientID == nil && items[i].ExternalID != "" {
				status, err := client.GetStatus(ctx, items[i].ExternalID)
				if err == nil {
					// Update with real-time status from embedded client
					items[i].Status = string(status.Status)
					items[i].Progress = status.Progress
					// Also update DB to persist the status
					_ = s.UpdateQueueItemStatus(ctx, items[i].ID, items[i].Status, items[i].Progress)
				}
			}
		}
	}

	return items, nil
}

// UpdateQueueItemStatus updates the status of a queue item.
func (s *Service) UpdateQueueItemStatus(ctx context.Context, id int64, status string, progress float64) error {
	_, err := s.db.ExecContext(ctx, `UPDATE download_queue SET status = ?, progress = ? WHERE id = ?`, status, progress, id)
	return err
}

// UpdateQueueItemStatusWithPath updates the status and save_path of a queue item.
func (s *Service) UpdateQueueItemStatusWithPath(ctx context.Context, id int64, status string, progress float64, savePath string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE download_queue SET status = ?, progress = ?, save_path = ? WHERE id = ?`, status, progress, savePath, id)
	return err
}

// RemoveFromQueue removes an item from the download queue.
func (s *Service) RemoveFromQueue(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM download_queue WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete query failed: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("queue item %d not found", id)
	}
	return nil
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

	// Determine format from URL or title first (needed for filename)
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

	// Add format extension to filename if missing
	if format != "unknown" && !strings.HasSuffix(strings.ToLower(filename), "."+format) {
		filename = filename + "." + format
	}

	// Build author folder and expected save path
	// Structure: RootFolder/AuthorName/AuthorName - BookTitle.format
	authorFolder := sanitizeFilename(b.AuthorName)
	authorDir := filepath.Join(cfg.DownloadDir, authorFolder)
	expectedSavePath := filepath.Join(authorDir, filename)

	// Add to download client
	downloadID, err := client.Add(ctx, download.AddItem{
		Name:        filename,
		DownloadURL: downloadURL,
		Category:    "books",
		SavePath:    authorDir, // Tell client to save in author folder
	})
	if err != nil {
		return nil, fmt.Errorf("add to download client: %w", err)
	}

	// Record in download queue with expected save path (use nil for embedded client)
	var clientIDPtr *int64
	if cfg.ID != 0 {
		clientIDPtr = &cfg.ID
	}
	if err := s.recordDownload(ctx, b.ID, clientIDPtr, nil, releaseTitle, size, format, downloadURL, downloadID, expectedSavePath); err != nil {
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

// restartDownload restarts a download that was lost (e.g., after server restart).
func (s *Service) restartDownload(ctx context.Context, queueID int64, client download.Client) error {
	// Get download info from queue
	var title, downloadURL string
	err := s.db.QueryRowContext(ctx, `
		SELECT title, download_url FROM download_queue WHERE id = ?
	`, queueID).Scan(&title, &downloadURL)
	if err != nil {
		return fmt.Errorf("get queue item: %w", err)
	}

	if downloadURL == "" {
		return fmt.Errorf("no download URL for queue item %d", queueID)
	}

	// Add to client again
	newID, err := client.Add(ctx, download.AddItem{
		Name:        title,
		DownloadURL: downloadURL,
		Category:    "books",
	})
	if err != nil {
		return fmt.Errorf("add to client: %w", err)
	}

	// Update external_id in queue
	_, err = s.db.ExecContext(ctx, `UPDATE download_queue SET external_id = ?, status = 'queued' WHERE id = ?`, newID, queueID)
	if err != nil {
		return fmt.Errorf("update queue: %w", err)
	}

	slog.Info("Download restarted", "queueId", queueID, "newExternalId", newID)
	return nil
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

// tryNextSource attempts to download from the next available source when a download fails.
// Returns true if a new download was started, false if no more sources are available.
func (s *Service) tryNextSource(ctx context.Context, queueID int64, errorMessage string) bool {
	// Get book_id from queue
	var bookID int64
	err := s.db.QueryRowContext(ctx, `SELECT book_id FROM download_queue WHERE id = ?`, queueID).Scan(&bookID)
	if err != nil {
		slog.Warn("Failed to get book_id for retry", "queueId", queueID, "error", err)
		return false
	}

	// Get book info
	b, err := s.bookService.FindByID(ctx, bookID)
	if err != nil {
		slog.Warn("Failed to get book for retry", "bookId", bookID, "error", err)
		return false
	}

	// Check if there are more sources to try
	var pendingCount int
	err = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM search_results WHERE book_id = ? AND status = 'pending'
	`, bookID).Scan(&pendingCount)
	if err != nil || pendingCount == 0 {
		slog.Info("No more download sources available", "book", b.Title)
		return false
	}

	// Remove the failed queue entry
	_, _ = s.db.ExecContext(ctx, `DELETE FROM download_queue WHERE id = ?`, queueID)

	// Try next source
	grabResult, err := s.grabNextSearchResult(ctx, b)
	if err != nil {
		slog.Warn("Failed to grab next source", "book", b.Title, "error", err)
		return false
	}

	slog.Info("Retrying with alternative source",
		"book", b.Title,
		"source", grabResult.ProviderName,
		"remaining", pendingCount-1,
	)

	return true
}

// cleanupSearchResults removes search results after successful download.
func (s *Service) cleanupSearchResults(ctx context.Context, queueID int64) {
	// Get book_id from queue
	var bookID int64
	err := s.db.QueryRowContext(ctx, `SELECT book_id FROM download_queue WHERE id = ?`, queueID).Scan(&bookID)
	if err != nil {
		return
	}

	// Delete all search results for this book
	_, _ = s.db.ExecContext(ctx, `DELETE FROM search_results WHERE book_id = ?`, bookID)
}

// GetPendingSourcesCount returns the number of pending download sources for a book.
func (s *Service) GetPendingSourcesCount(ctx context.Context, bookID int64) int {
	var count int
	_ = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM search_results WHERE book_id = ? AND status = 'pending'
	`, bookID).Scan(&count)
	return count
}
