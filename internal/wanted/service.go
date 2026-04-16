// Package wanted provides services for searching and grabbing wanted books.
// It coordinates between metadata providers, digital libraries, indexers, and download clients.
package wanted

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"

	"github.com/woliveiras/bookaneer/internal/core/book"
	corelibrary "github.com/woliveiras/bookaneer/internal/core/library"
	"github.com/woliveiras/bookaneer/internal/core/naming"
	"github.com/woliveiras/bookaneer/internal/core/pathmapping"
	"github.com/woliveiras/bookaneer/internal/download"
	"github.com/woliveiras/bookaneer/internal/library"
	"github.com/woliveiras/bookaneer/internal/search"
)

// SearchBookResult represents a single search result shown to the user.
type SearchBookResult struct {
	Title       string `json:"title"`
	Source      string `json:"source"`   // "library" or "indexer"
	Provider    string `json:"provider"` // provider/indexer name
	Format      string `json:"format"`
	Size        int64  `json:"size"`
	DownloadURL string `json:"downloadUrl"`
	Seeders     int    `json:"seeders,omitempty"`
}

// Service handles searching and grabbing wanted books.
type Service struct {
	db              *sql.DB
	bookService     *book.Service
	libraryService  *library.Aggregator
	searchService   *search.Service
	downloadService *download.Service
	namingEngine    *naming.Engine
	scanner         *corelibrary.Scanner
	pathMapper      *pathmapping.Service
}

// New creates a new Wanted service.
func New(
	db *sql.DB,
	bookService *book.Service,
	libraryService *library.Aggregator,
	searchService *search.Service,
	downloadService *download.Service,
	namingEngine *naming.Engine,
	scanner *corelibrary.Scanner,
	pathMapper *pathmapping.Service,
) *Service {
	return &Service{
		db:              db,
		bookService:     bookService,
		libraryService:  libraryService,
		searchService:   searchService,
		downloadService: downloadService,
		namingEngine:    namingEngine,
		scanner:         scanner,
		pathMapper:      pathMapper,
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

// Search searches for a book across all sources and returns results sorted by file size (largest first).
// An empty slice is returned when no results are found — the caller decides what to do next.
func (s *Service) Search(ctx context.Context, bookID int64) ([]SearchBookResult, error) {
	b, err := s.bookService.FindByID(ctx, bookID)
	if err != nil {
		return nil, fmt.Errorf("find book: %w", err)
	}

	isbn := b.ISBN13
	if isbn == "" {
		isbn = b.ISBN
	}

	query := isbn
	if query == "" {
		query = b.Title
		if b.AuthorName != "" {
			query = fmt.Sprintf("%s %s", b.Title, b.AuthorName)
		}
	} else {
		slog.Info("Using ISBN for search precision", "isbn", query, "book", b.Title)
	}

	allResults, err := s.searchAllSources(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// If ISBN search returned nothing, retry with title+author
	if len(allResults) == 0 && isbn != "" {
		slog.Info("ISBN search returned no results, retrying with title+author", "book", b.Title)
		titleQuery := b.Title
		if b.AuthorName != "" {
			titleQuery = fmt.Sprintf("%s %s", b.Title, b.AuthorName)
		}
		allResults, err = s.searchAllSources(ctx, titleQuery)
		if err != nil {
			return nil, fmt.Errorf("fallback search failed: %w", err)
		}
	}

	if len(allResults) == 0 {
		slog.Info("No results found for book", "id", bookID, "title", b.Title)
		return []SearchBookResult{}, nil
	}

	// Sort by file size descending (largest = heaviest = shown first)
	sortBySize(allResults)

	slog.Info("Found download sources", "book", b.Title, "count", len(allResults))
	for i, r := range allResults {
		if i >= 5 {
			break
		}
		slog.Debug("Candidate", "rank", i+1, "result", r.String())
	}

	// Persist results for audit trail and future grab reference
	if _, err := s.db.ExecContext(ctx, `DELETE FROM search_results WHERE book_id = ?`, b.ID); err != nil {
		slog.Warn("failed to clear search results", "bookId", b.ID, "error", err)
	}
	if err := s.saveSearchResults(ctx, b.ID, allResults); err != nil {
		slog.Warn("failed to save search results", "error", err)
	}

	out := make([]SearchBookResult, 0, len(allResults))
	for _, r := range allResults {
		res := SearchBookResult{
			Title:       r.title,
			Source:      "indexer",
			Provider:    r.sourceName,
			Format:      r.format,
			Size:        r.size,
			DownloadURL: r.downloadURL,
		}
		if r.isLibrary {
			res.Source = "library"
		}
		if r.isIndexer && r.indexerResult != nil {
			res.Seeders = r.indexerResult.Seeders
		}
		out = append(out, res)
	}
	return out, nil
}

// GetWantedBooks returns all monitored books without files.
func (s *Service) GetWantedBooks(ctx context.Context) ([]book.Book, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT b.id, b.author_id, b.title, COALESCE(b.sort_title,''), COALESCE(b.foreign_id,''), COALESCE(b.isbn,''), COALESCE(b.isbn13,''),
		       COALESCE(b.release_date,''), COALESCE(b.overview,''), COALESCE(b.image_url,''), b.page_count, b.monitored,
		       b.added_at, b.updated_at,
		       COALESCE(a.name,'') as author_name,
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
	defer func() { _ = rows.Close() }()

	var books []book.Book
	for rows.Next() {
		var b book.Book
		var monitored, hasFile int
		if err := rows.Scan(
			&b.ID, &b.AuthorID, &b.Title, &b.SortTitle, &b.ForeignID, &b.ISBN, &b.ISBN13,
			&b.ReleaseDate, &b.Overview, &b.ImageURL, &b.PageCount, &monitored,
			&b.AddedAt, &b.UpdatedAt, &b.AuthorName, &hasFile,
		); err != nil {
			return nil, err
		}
		b.Monitored = monitored == 1
		b.HasFile = hasFile == 1
		books = append(books, b)
	}

	return books, rows.Err()
}
