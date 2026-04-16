package wanted

import (
	"fmt"
	"strings"

	"github.com/woliveiras/bookaneer/internal/library"
	"github.com/woliveiras/bookaneer/internal/search"
)

// unifiedResult wraps either a library or indexer result for unified ranking.
type unifiedResult struct {
	// Source identification
	isLibrary   bool
	isIndexer   bool
	sourceName  string
	downloadURL string
	title       string
	format      string
	size        int64

	// Type-specific data (one will be nil)
	libraryResult *library.SearchResult
	indexerResult *search.Result
}

// newLibraryResult creates a unified result from a library search result.
func newLibraryResult(r *library.SearchResult) *unifiedResult {
	return &unifiedResult{
		isLibrary:     true,
		sourceName:    r.Provider,
		downloadURL:   r.DownloadURL,
		title:         r.Title,
		format:        r.Format,
		size:          r.Size,
		libraryResult: r,
	}
}

// newIndexerResult creates a unified result from an indexer search result.
func newIndexerResult(r *search.Result) *unifiedResult {
	format := extractFormat(r.Title)
	return &unifiedResult{
		isIndexer:     true,
		sourceName:    r.IndexerName,
		downloadURL:   r.DownloadURL,
		title:         r.Title,
		format:        format,
		size:          r.Size,
		indexerResult: r,
	}
}

// sortBySize sorts results by file size descending (largest first).
// Larger files indicate better quality (richer formatting, images, fonts preserved).
func sortBySize(results []*unifiedResult) {
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].size > results[i].size {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

// extractFormat attempts to extract the format from a title or filename.
func extractFormat(title string) string {
	lower := strings.ToLower(title)
	formats := []string{"epub", "mobi", "azw3", "azw", "pdf", "cbz", "cbr"}
	for _, f := range formats {
		if strings.Contains(lower, "."+f) || strings.Contains(lower, " "+f+" ") || strings.HasSuffix(lower, " "+f) {
			return f
		}
	}
	return "unknown"
}

// String returns a human-readable representation of the result for logging.
func (r *unifiedResult) String() string {
	sizeMB := float64(r.size) / (1024 * 1024)
	source := "library"
	if r.isIndexer {
		source = "indexer"
	}
	return fmt.Sprintf("[%s:%s] %s %.1fMB", source, r.sourceName, r.format, sizeMB)
}
