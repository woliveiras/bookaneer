package wanted

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"

	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/library"
	"github.com/woliveiras/bookaneer/internal/search"
)

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
		if isEbookFormat(r.Title) {
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
