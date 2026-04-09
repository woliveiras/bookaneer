package wanted

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"

	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/download"
	"github.com/woliveiras/bookaneer/internal/library"
	"github.com/woliveiras/bookaneer/internal/search"
)

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
		SavePath:    authorDir,
	})
	if err != nil {
		return nil, fmt.Errorf("add to download client: %w", err)
	}

	// Record in download queue with expected save path
	var clientIDPtr *int64
	if cfg.ID != 0 {
		clientIDPtr = &cfg.ID
	}
	if err := s.recordDownload(ctx, b.ID, clientIDPtr, nil, r.Title, r.Size, r.Format, r.DownloadURL, downloadID, expectedSavePath); err != nil {
		return nil, fmt.Errorf("record download in queue: %w", err)
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

	indexerID := &r.IndexerID

	var clientIDPtr *int64
	if cfg.ID != 0 {
		clientIDPtr = &cfg.ID
	}
	if err := s.recordDownload(ctx, b.ID, clientIDPtr, indexerID, r.Title, r.Size, format, r.DownloadURL, downloadID, ""); err != nil {
		return nil, fmt.Errorf("record download in queue: %w", err)
	}

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
	b, err := s.bookService.FindByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("find book: %w", err)
	}

	client, cfg, err := s.downloadService.GetDirectClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("get download client: %w", err)
	}
	if client == nil {
		return nil, fmt.Errorf("no download client configured")
	}

	filename := releaseTitle
	if filename == "" {
		filename = fmt.Sprintf("%s - %s", b.AuthorName, b.Title)
	}
	filename = sanitizeFilename(filename)

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

	if format != "unknown" && !strings.HasSuffix(strings.ToLower(filename), "."+format) {
		filename = filename + "." + format
	}

	authorFolder := sanitizeFilename(b.AuthorName)
	authorDir := filepath.Join(cfg.DownloadDir, authorFolder)
	expectedSavePath := filepath.Join(authorDir, filename)

	downloadID, err := client.Add(ctx, download.AddItem{
		Name:        filename,
		DownloadURL: downloadURL,
		Category:    "books",
		SavePath:    authorDir,
	})
	if err != nil {
		return nil, fmt.Errorf("add to download client: %w", err)
	}

	var clientIDPtr *int64
	if cfg.ID != 0 {
		clientIDPtr = &cfg.ID
	}
	if err := s.recordDownload(ctx, b.ID, clientIDPtr, nil, releaseTitle, size, format, downloadURL, downloadID, expectedSavePath); err != nil {
		return nil, fmt.Errorf("record download in queue: %w", err)
	}

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
