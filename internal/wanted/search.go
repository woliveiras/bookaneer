package wanted

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/core/qualityprofile"
	"github.com/woliveiras/bookaneer/internal/library"
	"github.com/woliveiras/bookaneer/internal/search"
)

// Search result status values stored in the search_results table.
const (
	searchResultPending = "pending"
	searchResultTried   = "tried"
	searchResultFailed  = "failed"
	searchResultSuccess = "success"
)

// searchAllSources searches both libraries and indexers in parallel and returns unified results.
func (s *Service) searchAllSources(ctx context.Context, b *book.Book, query string) ([]*unifiedResult, error) {
	var (
		libraryResults []*unifiedResult
		indexerResults []*unifiedResult
		libraryErr     error
		indexerErr     error
		wg             sync.WaitGroup
	)

	// Search libraries in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		if s.libraryService == nil {
			return
		}
		results, err := s.libraryService.Search(ctx, query)
		if err != nil {
			libraryErr = err
			return
		}
		// Convert to unified results
		for i := range results {
			if isValidLibraryResult(&results[i]) {
				libraryResults = append(libraryResults, newLibraryResult(&results[i]))
			}
		}
	}()

	// Search indexers in parallel
	wg.Add(1)
	go func() {
		defer wg.Done()
		if s.searchService == nil {
			return
		}
		results, err := s.searchService.Search(ctx, search.SearchQuery{Query: query})
		if err != nil {
			indexerErr = err
			return
		}
		// Convert to unified results (filter for ebook formats)
		for i := range results {
			if isEbookFormat(results[i].Title) {
				indexerResults = append(indexerResults, newIndexerResult(&results[i]))
			}
		}
	}()

	wg.Wait()

	// Log any errors but don't fail if at least one source worked
	if libraryErr != nil {
		slog.Warn("Library search failed", "error", libraryErr)
	}
	if indexerErr != nil {
		slog.Warn("Indexer search failed", "error", indexerErr)
	}

	// Combine results
	allResults := append(libraryResults, indexerResults...)

	slog.Info("Search completed",
		"query", query,
		"libraryResults", len(libraryResults),
		"indexerResults", len(indexerResults),
		"total", len(allResults))

	return allResults, nil
}

// isValidLibraryResult checks if a library result is usable.
func isValidLibraryResult(r *library.SearchResult) bool {
	if r.DownloadURL == "" {
		return false
	}
	format := strings.ToLower(r.Format)
	return format == "epub" || format == "pdf" || format == "mobi" || format == "azw" || format == "azw3"
}

// getQualityProfile retrieves the quality profile for a book (uses default if not configured).
func (s *Service) getQualityProfile(ctx context.Context, b *book.Book) *qualityprofile.QualityProfile {
	// For now, always return the default profile
	// TODO: Wire up author → root_folder → quality_profile_id relationship
	return qualityprofile.DefaultProfile()
}

// saveSearchResults saves all unified results to the search_results table for fallback.
func (s *Service) saveSearchResults(ctx context.Context, bookID int64, results []*unifiedResult) error {
	for i, r := range results {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO search_results (book_id, provider, title, download_url, format, size, score, priority, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, bookID, r.sourceName, r.title, r.downloadURL, r.format, r.size, r.score, i, searchResultPending)
		if err != nil {
			return fmt.Errorf("save result %d: %w", i, err)
		}
	}
	return nil
}

// tryGrabBestResult attempts to grab the best result, falling back to alternatives on failure.
func (s *Service) tryGrabBestResult(ctx context.Context, b *book.Book, results []*unifiedResult) (*GrabResult, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("no results available")
	}

	// Try the best result first
	best := results[0]
	slog.Info("Attempting to grab best result", "book", b.Title, "source", best.sourceName, "format", best.format)

	var grabResult *GrabResult
	var err error

	if best.isLibrary {
		grabResult, err = s.grabFromLibrary(ctx, b, best.libraryResult)
	} else if best.isIndexer {
		grabResult, err = s.grabFromIndexer(ctx, b, best.indexerResult)
	}

	if err == nil {
		// Mark as success in search_results
		_, _ = s.db.ExecContext(ctx, `
			UPDATE search_results 
			SET status = ?, tried_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
			WHERE book_id = ? AND priority = 0
		`, searchResultSuccess, b.ID)
		return grabResult, nil
	}

	// First grab failed, mark it and try fallbacks
	slog.Warn("Failed to grab best result, trying fallbacks", "error", err)
	_, _ = s.db.ExecContext(ctx, `
		UPDATE search_results 
		SET status = ?, error_message = ?, tried_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		WHERE book_id = ? AND priority = 0
	`, searchResultFailed, err.Error(), b.ID)

	// Try remaining results
	for i := 1; i < len(results) && i < 5; i++ { // Try up to top 5 results
		r := results[i]
		slog.Info("Trying fallback option", "rank", i+1, "source", r.sourceName, "format", r.format)

		if r.isLibrary {
			grabResult, err = s.grabFromLibrary(ctx, b, r.libraryResult)
		} else if r.isIndexer {
			grabResult, err = s.grabFromIndexer(ctx, b, r.indexerResult)
		}

		if err == nil {
			_, _ = s.db.ExecContext(ctx, `
				UPDATE search_results 
				SET status = ?, tried_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
				WHERE book_id = ? AND priority = ?
			`, searchResultSuccess, b.ID, i)
			return grabResult, nil
		}

		_, _ = s.db.ExecContext(ctx, `
			UPDATE search_results 
			SET status = ?, error_message = ?, tried_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
			WHERE book_id = ? AND priority = ?
		`, searchResultFailed, err.Error(), b.ID, i)
	}

	return nil, fmt.Errorf("all grab attempts failed for %q", b.Title)
}

// grabNextSearchResult tries to grab the next pending search result for a book.
// Used for fallback retry when a download fails.
func (s *Service) grabNextSearchResult(ctx context.Context, b *book.Book) (*GrabResult, error) {
	// Get next pending result
	var resultID int64
	var provider, title, downloadURL, format string
	var size int64

	err := s.db.QueryRowContext(ctx, `
		SELECT id, provider, title, download_url, format, size 
		FROM search_results 
		WHERE book_id = ? AND status = ?
		ORDER BY priority ASC
		LIMIT 1
	`, b.ID, searchResultPending).Scan(&resultID, &provider, &title, &downloadURL, &format, &size)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no more download sources available for %q", b.Title)
		}
		return nil, fmt.Errorf("get next search result: %w", err)
	}

	// Mark as tried
	if _, err := s.db.ExecContext(ctx, `
		UPDATE search_results SET status = ?, tried_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		WHERE id = ?
	`, searchResultTried, resultID); err != nil {
		slog.Warn("failed to mark search result as tried", "resultId", resultID, "error", err)
	}

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
		if _, ferr := s.db.ExecContext(ctx, `
			UPDATE search_results SET status = ?, error_message = ? WHERE id = ?
		`, searchResultFailed, err.Error(), resultID); ferr != nil {
			slog.Warn("failed to mark search result as failed", "resultId", resultID, "error", ferr)
		}
		return nil, err
	}

	// Mark as success
	if _, err := s.db.ExecContext(ctx, `UPDATE search_results SET status = ? WHERE id = ?`, searchResultSuccess, resultID); err != nil {
		slog.Warn("failed to mark search result as success", "resultId", resultID, "error", err)
	}

	return grabResult, nil
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
		SELECT COUNT(*) FROM search_results WHERE book_id = ? AND status = ?
	`, bookID, searchResultPending).Scan(&pendingCount)
	if err != nil || pendingCount == 0 {
		slog.Info("No more download sources available", "book", b.Title)
		return false
	}

	// Remove the failed queue entry
	if _, err := s.db.ExecContext(ctx, `DELETE FROM download_queue WHERE id = ?`, queueID); err != nil {
		slog.Warn("failed to remove failed queue entry", "queueId", queueID, "error", err)
	}

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
	if _, err := s.db.ExecContext(ctx, `DELETE FROM search_results WHERE book_id = ?`, bookID); err != nil {
		slog.Warn("failed to cleanup search results", "bookId", bookID, "error", err)
	}
}

// GetPendingSourcesCount returns the number of pending download sources for a book.
func (s *Service) GetPendingSourcesCount(ctx context.Context, bookID int64) int {
	var count int
	_ = s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM search_results WHERE book_id = ? AND status = ?
	`, bookID, searchResultPending).Scan(&count)
	return count
}

// tryNextSourceForMismatch handles content mismatch by blocklisting the bad source
// and attempting the next available one. The mismatched file is kept (flagged)
// so the user can inspect it.
func (s *Service) tryNextSourceForMismatch(ctx context.Context, queueID int64) {
	var bookID int64
	var title, downloadURL, format string
	err := s.db.QueryRowContext(ctx, `
		SELECT book_id, title, download_url, format FROM download_queue WHERE id = ?
	`, queueID).Scan(&bookID, &title, &downloadURL, &format)
	if err != nil {
		slog.Warn("failed to get queue item for mismatch retry", "queueId", queueID, "error", err)
		return
	}

	// Blocklist the bad source so we don't try it again
	_ = s.AddToBlocklist(ctx, bookID, title, format, "content mismatch detected automatically")

	// Check if there are more sources to try
	pending := s.GetPendingSourcesCount(ctx, bookID)
	if pending == 0 {
		slog.Info("No alternative sources for content mismatch", "book", title)
		return
	}

	b, err := s.bookService.FindByID(ctx, bookID)
	if err != nil {
		return
	}

	grabResult, err := s.grabNextSearchResult(ctx, b)
	if err != nil {
		slog.Warn("Failed to grab next source after mismatch", "book", b.Title, "error", err)
		return
	}

	slog.Info("Retrying with alternative source after content mismatch",
		"book", b.Title,
		"source", grabResult.ProviderName,
		"remaining", pending-1,
	)
}
