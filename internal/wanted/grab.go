package wanted

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/download"
	"github.com/woliveiras/bookaneer/internal/search"
)

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
	format := detectEbookFormat(r.Title)

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

// GrabIndexerRequest holds the data needed to grab an indexer result.
type GrabIndexerRequest struct {
	GUID         string `json:"guid"`
	ReleaseTitle string `json:"releaseTitle"`
	DownloadURL  string `json:"downloadUrl"`
	Size         int64  `json:"size"`
	Seeders      int    `json:"seeders"`
	IndexerID    int64  `json:"indexerId"`
	IndexerName  string `json:"indexerName"`
}

// GrabIndexerRelease grabs an indexer result, routing to the appropriate torrent or usenet client.
func (s *Service) GrabIndexerRelease(ctx context.Context, bookID int64, req GrabIndexerRequest) (*GrabResult, error) {
	b, err := s.bookService.FindByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("find book: %w", err)
	}

	r := &search.Result{
		GUID:        req.GUID,
		Title:       req.ReleaseTitle,
		Size:        req.Size,
		DownloadURL: req.DownloadURL,
		Seeders:     req.Seeders,
		IndexerID:   req.IndexerID,
		IndexerName: req.IndexerName,
	}

	return s.grabFromIndexer(ctx, b, r)
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

	format := detectEbookFormat(downloadURL + releaseTitle)

	if format != "unknown" && !strings.HasSuffix(strings.ToLower(filename), "."+format) {
		filename = filename + "." + format
	}

	authorDir, expectedSavePath := buildAuthorSavePath(cfg.DownloadDir, b.AuthorName, filename)

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
