package wanted_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/download"
	_ "github.com/woliveiras/bookaneer/internal/download/direct" // register embedded downloader factory
	"github.com/woliveiras/bookaneer/internal/library"
	"github.com/woliveiras/bookaneer/internal/search"
	"github.com/woliveiras/bookaneer/internal/testutil"
	"github.com/woliveiras/bookaneer/internal/wanted"
)

// mockLibraryProvider is a test-only implementation of library.Provider.
type mockLibraryProvider struct {
	results []library.SearchResult
	err     error
}

func (m *mockLibraryProvider) Name() string { return "mock" }
func (m *mockLibraryProvider) Search(_ context.Context, _ string) ([]library.SearchResult, error) {
	return m.results, m.err
}
func (m *mockLibraryProvider) GetDownloadLink(_ context.Context, _ string) (string, error) {
	return "", nil
}

// TestSearchAndGrab_LibrarySucceeds exercises the full happy-path for
// searchDigitalLibraries → grabNextSearchResult → grabFromLibrary.
// Port 1 is used as the download URL: the Add call (synchronous) succeeds
// while the background goroutine fails fast without writing any files, so
// the t.TempDir cleanup never races against an in-progress download.
func TestSearchAndGrab_LibrarySucceeds(t *testing.T) {
	db := testutil.OpenTestDB(t)
	libDir := t.TempDir()
	testutil.SeedRootFolder(t, db, libDir, "Library")

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	mock := &mockLibraryProvider{
		results: []library.SearchResult{
			{
				Provider:    "mock",
				Title:       "The Hobbit",
				DownloadURL: "http://127.0.0.1:1/hobbit.epub",
				Format:      "epub",
				Size:        1024,
			},
		},
	}
	agg := library.NewAggregator(mock)

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, agg, nil, downloadSvc)
	ctx := context.Background()

	result, err := svc.SearchAndGrab(ctx, bookID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, bookID, result.BookID)
	assert.Equal(t, "library", result.Source)
	assert.Equal(t, "epub", result.Format)
	assert.Equal(t, "mock", result.ProviderName)

	// recordDownload should have inserted a row in download_queue.
	queue, err := svc.GetDownloadQueue(ctx)
	require.NoError(t, err)
	assert.Len(t, queue, 1)
	assert.Equal(t, "The Hobbit", queue[0].Title)

	// recordHistory should have inserted a 'grabbed' event.
	history, err := svc.GetHistory(ctx, 10, "grabbed")
	require.NoError(t, err)
	assert.Len(t, history, 1)
}

// TestSearchAndGrab_LibraryReturnsInvalidFormats covers the filtering branch
// inside searchDigitalLibraries where all results have non-downloadable formats
// (djvu, txt, etc.) so validResults stays empty and the function returns nil,nil.
func TestSearchAndGrab_LibraryReturnsInvalidFormats(t *testing.T) {
	db := testutil.OpenTestDB(t)
	libDir := t.TempDir()
	testutil.SeedRootFolder(t, db, libDir, "Library")

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	mock := &mockLibraryProvider{
		results: []library.SearchResult{
			{Provider: "mock", Title: "The Hobbit", DownloadURL: "http://example.com/book.djvu", Format: "djvu", Size: 1024},
			{Provider: "mock", Title: "The Hobbit", DownloadURL: "http://example.com/book.txt", Format: "txt", Size: 512},
		},
	}
	agg := library.NewAggregator(mock)

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, agg, nil, downloadSvc)
	ctx := context.Background()

	_, err := svc.SearchAndGrab(ctx, bookID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no suitable download found")
}

// TestSearchAndGrab_LibraryReturnsNoDownloadURL covers the
// `if r.DownloadURL == ""` continue branch inside searchDigitalLibraries.
func TestSearchAndGrab_LibraryReturnsNoDownloadURL(t *testing.T) {
	db := testutil.OpenTestDB(t)
	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	mock := &mockLibraryProvider{
		results: []library.SearchResult{
			{Provider: "mock", Title: "The Hobbit", DownloadURL: "", Format: "epub", Size: 1024},
		},
	}
	agg := library.NewAggregator(mock)

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, agg, nil, downloadSvc)
	ctx := context.Background()

	_, err := svc.SearchAndGrab(ctx, bookID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no suitable download found")
}

// TestSearchAndGrab_LibraryReturnsMultipleFormats covers the full filter loop
// in searchDigitalLibraries, including pdf and mobi results alongside epub.
// Port 1 URLs keep goroutines from writing to the test's TempDir.
func TestSearchAndGrab_LibraryReturnsMultipleFormats(t *testing.T) {
	db := testutil.OpenTestDB(t)
	libDir := t.TempDir()
	testutil.SeedRootFolder(t, db, libDir, "Library")

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	mock := &mockLibraryProvider{
		results: []library.SearchResult{
			{Provider: "mock", Title: "The Hobbit PDF", DownloadURL: "http://127.0.0.1:1/a.pdf", Format: "pdf", Size: 512},
			{Provider: "mock", Title: "The Hobbit MOBI", DownloadURL: "http://127.0.0.1:1/b.mobi", Format: "mobi", Size: 256},
			{Provider: "mock", Title: "The Hobbit EPUB", DownloadURL: "http://127.0.0.1:1/c.epub", Format: "epub", Size: 1024},
		},
	}
	agg := library.NewAggregator(mock)

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, agg, nil, downloadSvc)
	ctx := context.Background()

	result, err := svc.SearchAndGrab(ctx, bookID)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "library", result.Source)
}

// TestSearchAndGrab_LibraryNoDirectClient covers the error path in grabFromLibrary
// where GetDirectClient fails because no root folder is configured.
func TestSearchAndGrab_LibraryNoDirectClient(t *testing.T) {
	db := testutil.OpenTestDB(t)
	// No root folder → GetDirectClient returns error.
	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	mock := &mockLibraryProvider{
		results: []library.SearchResult{
			{Provider: "mock", Title: "The Hobbit", DownloadURL: "http://example.com/book.epub", Format: "epub", Size: 1024},
		},
	}
	agg := library.NewAggregator(mock)

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, agg, nil, downloadSvc)
	ctx := context.Background()

	_, err := svc.SearchAndGrab(ctx, bookID)
	require.Error(t, err)
}

// TestSearchAndGrab_WithNonNilSearchService covers searchIndexers when the
// service is non-nil but has no clients loaded (returns nil results).
func TestSearchAndGrab_WithNonNilSearchService(t *testing.T) {
	db := testutil.OpenTestDB(t)
	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	bookSvc := book.New(db)
	searchSvc := search.NewService(db) // no clients loaded → Search returns nil
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, searchSvc, downloadSvc)
	ctx := context.Background()

	_, err := svc.SearchAndGrab(ctx, bookID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no suitable download found")
}

// TestGrabRelease_EmptyTitle covers the branch in GrabRelease where
// releaseTitle is "" and the filename is derived from AuthorName + BookTitle.
func TestGrabRelease_EmptyTitle(t *testing.T) {
	db := testutil.OpenTestDB(t)
	libDir := t.TempDir()
	testutil.SeedRootFolder(t, db, libDir, "Library")

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	// Empty releaseTitle → filename built from "Tolkien - The Hobbit"
	result, err := svc.GrabRelease(ctx, bookID, "http://127.0.0.1:1/test.epub", "", 0)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "manual", result.Source)
	assert.Equal(t, "epub", result.Format)
}

// TestGrabRelease_NoRootFolder covers the GetDirectClient error path in
// GrabRelease when no root folder is configured.
func TestGrabRelease_NoRootFolder(t *testing.T) {
	db := testutil.OpenTestDB(t)
	// No root folder → GetDirectClient errors.
	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	_, err := svc.GrabRelease(ctx, bookID, "http://127.0.0.1:1/test.epub", "Test.epub", 1024)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "get download client")
}

// TestGrabRelease_BookNotFound covers the FindByID error path in GrabRelease.
func TestGrabRelease_BookNotFound(t *testing.T) {
	svc, ctx := newTestService(t)

	_, err := svc.GrabRelease(ctx, 9999, "http://127.0.0.1:1/test.epub", "Test.epub", 1024)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "find book")
}

// TestProcessDownloads_TryNextSourceOnFailed verifies that tryNextSource is
// called when a download's in-memory status becomes StatusFailed and pending
// search results exist for the book.
//
// GrabRelease plants the item in the embedded direct client's map with a
// refused-connection URL. After a short sleep the background goroutine will
// have set the status to Failed; ProcessDownloads then sees that status and
// calls tryNextSource.
func TestProcessDownloads_TryNextSourceOnFailed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	libDir := t.TempDir()
	testutil.SeedRootFolder(t, db, libDir, "Library")

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	// Pending fallback so tryNextSource has a source to retry with.
	_, err := db.Exec(`
		INSERT INTO search_results (book_id, provider, title, download_url, format, size, score, priority, status)
		VALUES (?, 'mock', 'Fallback', 'http://127.0.0.1:1/fallback.epub', 'epub', 1024, 50, 0, 'pending')
	`, bookID)
	require.NoError(t, err)

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	// GrabRelease registers the download in the embedded direct client.
	// Port 1 is always refused so the goroutine fails almost instantly.
	_, err = svc.GrabRelease(ctx, bookID, "http://127.0.0.1:1/hobbit.epub", "The Hobbit.epub", 1024)
	require.NoError(t, err)

	// Give the goroutine time to hit the refused connection.
	time.Sleep(150 * time.Millisecond)

	pdResult, err := svc.ProcessDownloads(ctx)
	require.NoError(t, err)
	// The item was checked and reported as failed.
	assert.Equal(t, 1, pdResult.Checked)
	assert.Equal(t, 1, pdResult.Failed)
}

// TestProcessDownloads_TryNextSourceNoPending covers the branch in tryNextSource
// where pendingCount == 0, so it logs and returns false without retrying.
func TestProcessDownloads_TryNextSourceNoPending(t *testing.T) {
	db := testutil.OpenTestDB(t)
	libDir := t.TempDir()
	testutil.SeedRootFolder(t, db, libDir, "Library")

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	// No pending search results → tryNextSource returns false.
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	_, err := svc.GrabRelease(ctx, bookID, "http://127.0.0.1:1/hobbit.epub", "The Hobbit.epub", 1024)
	require.NoError(t, err)

	time.Sleep(150 * time.Millisecond)

	pdResult, err := svc.ProcessDownloads(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, pdResult.Checked)
	assert.Equal(t, 1, pdResult.Failed)
}

// TestProcessDownloads_CleanupSearchResultsOnSuccess verifies that
// cleanupSearchResults deletes search_results rows for the book after a
// successful import triggered through ProcessDownloads' active-download loop.
func TestProcessDownloads_CleanupSearchResultsOnSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="hobbit.epub"`)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("EPUB content"))
	}))
	t.Cleanup(srv.Close)

	db := testutil.OpenTestDB(t)
	libDir := t.TempDir()
	testutil.SeedRootFolder(t, db, libDir, "Library")

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	// Seed a pending search result that cleanupSearchResults should delete.
	_, err := db.Exec(`
		INSERT INTO search_results (book_id, provider, title, download_url, format, size, score, priority, status)
		VALUES (?, 'mock', 'Hobbit', ?, 'epub', 1024, 100, 0, 'pending')
	`, bookID, srv.URL+"/hobbit.epub")
	require.NoError(t, err)

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	// Start a download via GrabRelease so the embedded client tracks it.
	_, err = svc.GrabRelease(ctx, bookID, srv.URL+"/hobbit.epub", "The Hobbit.epub", 1024)
	require.NoError(t, err)

	// Poll until ProcessDownloads reports an imported file.
	assert.Eventually(t, func() bool {
		r, err := svc.ProcessDownloads(ctx)
		return err == nil && r.Imported > 0
	}, 5*time.Second, 100*time.Millisecond, "import should succeed within 5s")

	// After successful import cleanupSearchResults should have removed the row.
	var count int
	err = db.QueryRow(`SELECT COUNT(*) FROM search_results WHERE book_id = ?`, bookID).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "search_results should be cleaned up after import")
}

// TestGetPendingSourcesCount_AfterSearch verifies GetPendingSourcesCount after
// searchDigitalLibraries populates the search_results table.
// Port 1 URLs keep background goroutines from writing to the TempDir.
func TestGetPendingSourcesCount_AfterSearch(t *testing.T) {
	db := testutil.OpenTestDB(t)
	libDir := t.TempDir()
	testutil.SeedRootFolder(t, db, libDir, "Library")

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	// Two valid epub results; after the grab the first becomes 'tried'/'success',
	// the second stays 'pending'.
	mock := &mockLibraryProvider{
		results: []library.SearchResult{
			{Provider: "mock", Title: "Hobbit 1", DownloadURL: "http://127.0.0.1:1/a.epub", Format: "epub", Size: 1024},
			{Provider: "mock", Title: "Hobbit 2", DownloadURL: "http://127.0.0.1:1/b.epub", Format: "epub", Size: 2048},
		},
	}
	agg := library.NewAggregator(mock)

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, agg, nil, downloadSvc)
	ctx := context.Background()

	_, err := svc.SearchAndGrab(ctx, bookID)
	require.NoError(t, err)

	// One pending source should remain after the first was grabbed.
	count := svc.GetPendingSourcesCount(ctx, bookID)
	assert.Equal(t, 1, count)
}

// mockIndexer is a test-only implementation of search.Indexer that returns
// pre-configured results.
type mockIndexer struct {
	results []search.Result
}

func (m *mockIndexer) Name() string { return "mock-indexer" }
func (m *mockIndexer) Type() string { return "mock-indexer-type" }
func (m *mockIndexer) Search(_ context.Context, _ search.SearchQuery) ([]search.Result, error) {
	return m.results, nil
}
func (m *mockIndexer) Caps(_ context.Context) (*search.Capabilities, error) {
	return &search.Capabilities{}, nil
}
func (m *mockIndexer) Test(_ context.Context) error { return nil }

// TestSearchAndGrab_IndexerReturnsEbookResults covers searchIndexers when the
// search service has a loaded indexer that returns ebook-format results.
// grabFromIndexer will error (no download client configured) but the loop
// and format-filtering code paths are fully exercised.
func TestSearchAndGrab_IndexerReturnsEbookResults(t *testing.T) {
	// Use 'newznab' — a type accepted by the DB CHECK constraint.
	// Neither torznab nor newznab is imported in the wanted tests, so
	// registering here does not override any real factory.
	const mockType = "newznab"

	results := []search.Result{
		// With seeders → torrent path in grabFromIndexer.
		{
			Title:       "The Hobbit EPUB",
			DownloadURL: "http://127.0.0.1:1/hobbit.epub",
			IndexerID:   1,
			IndexerName: "mock",
			Size:        1024,
			Seeders:     10,
		},
		// Without seeders → usenet path in grabFromIndexer.
		{
			Title:       "The Hobbit EPUB NZB",
			DownloadURL: "http://127.0.0.1:1/hobbit.nzb",
			IndexerID:   1,
			IndexerName: "mock",
			Size:        1024,
			Seeders:     0,
		},
	}

	// Register the mock factory; idempotent across multiple test runs.
	search.RegisterFactory(mockType, func(cfg search.IndexerConfig) (search.Indexer, error) {
		return &mockIndexer{results: results}, nil
	})

	db := testutil.OpenTestDB(t)
	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	// Insert a mock indexer row that LoadIndexers will pick up.
	_, err := db.Exec(`
		INSERT INTO indexers
		  (name, type, base_url, api_path, api_key, categories, priority,
		   enabled, enable_rss, enable_automatic_search, enable_interactive_search,
		   additional_parameters, minimum_seeders)
		VALUES ('Mock', ?, 'http://localhost', '/api', 'k', '7030', 10,
		        1, 0, 1, 1, '', 0)
	`, mockType)
	require.NoError(t, err)

	searchSvc := search.NewService(db)
	require.NoError(t, searchSvc.LoadIndexers(context.Background()))

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	// No download client configured → grabFromIndexer will return "no client" errors.
	svc := wanted.New(db, bookSvc, nil, searchSvc, downloadSvc)
	ctx := context.Background()

	_, err = svc.SearchAndGrab(ctx, bookID)
	// All indexer grabs fail (no client), so SearchAndGrab ultimately returns
	// "no suitable download found".
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no suitable download found")
}

// TestSearchAndGrab_IndexerReturnsNonEbookResults covers the filter in
// searchIndexers that skips results whose titles contain no ebook keyword.
func TestSearchAndGrab_IndexerReturnsNonEbookResults(t *testing.T) {
	const mockType = "torznab"

	search.RegisterFactory(mockType, func(cfg search.IndexerConfig) (search.Indexer, error) {
		return &mockIndexer{results: []search.Result{
			{Title: "Some.Video.File.mp4", DownloadURL: "http://example.com/video.mp4", IndexerID: 1, IndexerName: "mock"},
		}}, nil
	})

	db := testutil.OpenTestDB(t)
	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	_, err := db.Exec(`
		INSERT INTO indexers
		  (name, type, base_url, api_path, api_key, categories, priority,
		   enabled, enable_rss, enable_automatic_search, enable_interactive_search,
		   additional_parameters, minimum_seeders)
		VALUES ('MockNE', ?, 'http://localhost', '/api', 'k', '7030', 10,
		        1, 0, 1, 1, '', 0)
	`, mockType)
	require.NoError(t, err)

	searchSvc := search.NewService(db)
	require.NoError(t, searchSvc.LoadIndexers(context.Background()))

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, searchSvc, downloadSvc)
	ctx := context.Background()

	_, err = svc.SearchAndGrab(ctx, bookID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no suitable download found")
}

// mockDownloadClient is a test-only download.Client that always succeeds.
type mockDownloadClient struct{}

func (m *mockDownloadClient) Name() string { return "mock" }
func (m *mockDownloadClient) Type() string { return "qbittorrent" }
func (m *mockDownloadClient) Test(_ context.Context) error { return nil }
func (m *mockDownloadClient) Add(_ context.Context, _ download.AddItem) (string, error) {
return "mock-dl-id", nil
}
func (m *mockDownloadClient) Remove(_ context.Context, _ string, _ bool) error { return nil }
func (m *mockDownloadClient) GetStatus(_ context.Context, _ string) (*download.ItemStatus, error) {
return nil, nil
}
func (m *mockDownloadClient) GetQueue(_ context.Context) ([]download.ItemStatus, error) {
return nil, nil
}

// TestSearchAndGrab_IndexerGrabSuccess exercises the full grabFromIndexer happy
// path: a mock 'newznab' indexer returns an ebook result with seeders, a mock
// 'qbittorrent' client accepts the download, and the function records the grab
// in both download_queue and history.
func TestSearchAndGrab_IndexerGrabSuccess(t *testing.T) {
const indexerType = "newznab"
const clientType = "qbittorrent"

// Register mock indexer factory (returns seeded ebook result).
search.RegisterFactory(indexerType, func(cfg search.IndexerConfig) (search.Indexer, error) {
return &mockIndexer{results: []search.Result{
{
Title:       "The Hobbit EPUB",
DownloadURL: "http://127.0.0.1:1/hobbit.epub",
IndexerID:   1,
IndexerName: "mock",
Size:        2048,
Seeders:     5,
},
}}, nil
})

// Register mock download client factory that always succeeds.
download.RegisterFactory(clientType, func(cfg download.ClientConfig) (download.Client, error) {
return &mockDownloadClient{}, nil
})

db := testutil.OpenTestDB(t)
authorID := testutil.SeedAuthor(t, db, "Tolkien")
bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

// Seed the indexer row.
_, err := db.Exec(`
INSERT INTO indexers
  (name, type, base_url, api_path, api_key, categories, priority,
   enabled, enable_rss, enable_automatic_search, enable_interactive_search,
   additional_parameters, minimum_seeders)
VALUES ('MockIndexer', ?, 'http://localhost', '/api', 'k', '7030', 5,
        1, 0, 1, 1, '', 0)
`, indexerType)
require.NoError(t, err)

// Seed the download client row with a non-zero ID (autoincrement ensures it).
_, err = db.Exec(`
INSERT INTO download_clients (name, type, host, port, enabled, priority)
VALUES ('MockQBit', ?, '127.0.0.1', 8080, 1, 0)
`, clientType)
require.NoError(t, err)

searchSvc := search.NewService(db)
require.NoError(t, searchSvc.LoadIndexers(context.Background()))

bookSvc := book.New(db)
downloadSvc := download.NewService(db)
// libraryService is nil → falls through to searchIndexers.
svc := wanted.New(db, bookSvc, nil, searchSvc, downloadSvc)
ctx := context.Background()

result, err := svc.SearchAndGrab(ctx, bookID)
require.NoError(t, err)
require.NotNil(t, result)
assert.Equal(t, bookID, result.BookID)
assert.Equal(t, "indexer", result.Source)
assert.Equal(t, "mock-dl-id", result.DownloadID)
assert.Equal(t, "MockQBit", result.ClientName)

// recordDownload must have written to download_queue.
queue, err := svc.GetDownloadQueue(ctx)
require.NoError(t, err)
assert.Len(t, queue, 1)
assert.Equal(t, "The Hobbit EPUB", queue[0].Title)

// recordHistory must have written a 'grabbed' event.
history, err := svc.GetHistory(ctx, 10, "grabbed")
require.NoError(t, err)
assert.Len(t, history, 1)
}

// TestGrabRelease_ConfiguredDirectClient covers the `cfg.ID != 0` branch in
// GrabRelease when a real 'direct' download client is configured in the DB
// (rather than using the embedded client with ID==0).
func TestGrabRelease_ConfiguredDirectClient(t *testing.T) {
dlDir := t.TempDir()

db := testutil.OpenTestDB(t)
// No root_folder needed: GetDirectClient finds the DB-configured client first.
authorID := testutil.SeedAuthor(t, db, "Tolkien")
bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

// Insert a 'direct' client with a non-zero ID (autoincrement).
_, err := db.Exec(`
INSERT INTO download_clients (name, type, host, port, enabled, priority, download_dir)
VALUES ('DirectClient', 'direct', '', 0, 1, 0, ?)
`, dlDir)
require.NoError(t, err)

bookSvc := book.New(db)
downloadSvc := download.NewService(db)
svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
ctx := context.Background()

// Port 1 → the goroutine fails fast without writing to dlDir.
result, err := svc.GrabRelease(ctx, bookID, "http://127.0.0.1:1/test.epub", "Test.epub", 1024)
require.NoError(t, err)
require.NotNil(t, result)
assert.Equal(t, "manual", result.Source)
assert.Equal(t, "DirectClient", result.ClientName)

// The download_queue entry should reference the configured client (ID != 0).
queue, err := svc.GetDownloadQueue(ctx)
require.NoError(t, err)
require.Len(t, queue, 1)
assert.NotNil(t, queue[0].DownloadClientID, "expected non-nil client ID from configured client")
}

// TestSearchAndGrab_LibraryConfiguredDirectClient covers the `cfg.ID != 0`
// branch in grabFromLibrary when a direct download client is configured in DB.
func TestSearchAndGrab_LibraryConfiguredDirectClient(t *testing.T) {
dlDir := t.TempDir()

db := testutil.OpenTestDB(t)
authorID := testutil.SeedAuthor(t, db, "Tolkien")
bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

_, err := db.Exec(`
INSERT INTO download_clients (name, type, host, port, enabled, priority, download_dir)
VALUES ('DirectDB', 'direct', '', 0, 1, 0, ?)
`, dlDir)
require.NoError(t, err)

mock := &mockLibraryProvider{
results: []library.SearchResult{
{
Provider:    "mock",
Title:       "The Hobbit",
DownloadURL: "http://127.0.0.1:1/hobbit.epub",
Format:      "epub",
Size:        1024,
},
},
}
agg := library.NewAggregator(mock)

bookSvc := book.New(db)
downloadSvc := download.NewService(db)
svc := wanted.New(db, bookSvc, agg, nil, downloadSvc)
ctx := context.Background()

result, err := svc.SearchAndGrab(ctx, bookID)
require.NoError(t, err)
require.NotNil(t, result)
assert.Equal(t, "library", result.Source)
assert.Equal(t, "DirectDB", result.ClientName)

queue, err := svc.GetDownloadQueue(ctx)
require.NoError(t, err)
require.Len(t, queue, 1)
assert.NotNil(t, queue[0].DownloadClientID, "grabFromLibrary should set client ID when cfg.ID != 0")
}

// TestProcessDownloads_ImportFailsNoRootFolder covers the
// `slog.Warn("Failed to import download")` branch in ProcessDownloads'
// active-downloads loop when importCompletedDownload fails because the root
// folder is removed from the DB after the download completes.
func TestProcessDownloads_ImportFailsNoRootFolder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="hobbit.epub"`)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("epub"))
	}))
	t.Cleanup(srv.Close)

	db := testutil.OpenTestDB(t)
	libDir := t.TempDir()
	folderID := testutil.SeedRootFolder(t, db, libDir, "Library")

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	// GrabRelease caches the embedded client with DownloadDir=libDir.
	_, err := svc.GrabRelease(ctx, bookID, srv.URL+"/hobbit.epub", "The Hobbit.epub", 1024)
	require.NoError(t, err)

	// Delete root folder from DB; the embedded client's cache is unaffected
	// so the download still proceeds.  importCompletedDownload will fail when
	// it tries to re-query root_folders during the import step.
	_, err = db.Exec("DELETE FROM root_folders WHERE id = ?", folderID)
	require.NoError(t, err)

	// Wait long enough for the small file to download.
	time.Sleep(500 * time.Millisecond)

	result, err := svc.ProcessDownloads(ctx)
	require.NoError(t, err)
	// The item was completed but import failed (no root folder).
	assert.Equal(t, 1, result.Checked)
	assert.Equal(t, 1, result.Completed)
	assert.Equal(t, 0, result.Imported)
}
