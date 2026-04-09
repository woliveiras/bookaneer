// Package wanted provides services for searching and grabbing wanted books.
// It coordinates between metadata providers, digital libraries, indexers, and download clients.
package wanted

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
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
