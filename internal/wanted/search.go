package wanted

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"

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
func (s *Service) searchAllSources(ctx context.Context, query string) ([]*unifiedResult, error) {
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
		for i := range results {
			if isEbookFormat(results[i].Title) {
				indexerResults = append(indexerResults, newIndexerResult(&results[i]))
			}
		}
	}()

	wg.Wait()

	if libraryErr != nil {
		slog.Warn("Library search failed", "error", libraryErr)
	}
	if indexerErr != nil {
		slog.Warn("Indexer search failed", "error", indexerErr)
	}

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

func (s *Service) saveSearchResults(ctx context.Context, bookID int64, results []*unifiedResult) error {
	for i, r := range results {
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO search_results (book_id, provider, title, download_url, format, size, score, priority, status)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, bookID, r.sourceName, r.title, r.downloadURL, r.format, r.size, i, i, searchResultPending)
		if err != nil {
			return fmt.Errorf("save result %d: %w", i, err)
		}
	}
	return nil
}
