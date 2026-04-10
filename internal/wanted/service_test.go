package wanted_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/download"
	"github.com/woliveiras/bookaneer/internal/testutil"
	"github.com/woliveiras/bookaneer/internal/core/naming"
	"github.com/woliveiras/bookaneer/internal/wanted"
)

func TestGetWantedBooks_Empty(t *testing.T) {
	svc, ctx := newTestService(t)

	books, err := svc.GetWantedBooks(ctx)
	require.NoError(t, err)
	assert.Empty(t, books)
}

func TestGetWantedBooks_WithMonitoredBooks(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	testutil.SeedBook(t, db, authorID, "The Hobbit")
	testutil.SeedBook(t, db, authorID, "The Lord of the Rings")

	books, err := svc.GetWantedBooks(ctx)
	require.NoError(t, err)
	assert.Len(t, books, 2)
}

func TestGetWantedBooks_ExcludesBooksWithFiles(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	book1 := testutil.SeedBook(t, db, authorID, "The Hobbit")
	testutil.SeedBook(t, db, authorID, "Silmarillion")

	_, err := db.Exec(`INSERT INTO book_files (book_id, path, relative_path, size, format, quality) VALUES (?, '/lib/hobbit.epub', 'hobbit.epub', 1024, 'epub', 'epub')`, book1)
	require.NoError(t, err)

	books, err := svc.GetWantedBooks(ctx)
	require.NoError(t, err)
	assert.Len(t, books, 1)
	assert.Equal(t, "Silmarillion", books[0].Title)
}

func TestSearchAndGrab_BookNotFound(t *testing.T) {
	svc, ctx := newTestService(t)

	_, err := svc.SearchAndGrab(ctx, 9999)
	require.Error(t, err)
}

func TestSearchAndGrab_NotMonitored(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	result, err := db.Exec(`INSERT INTO books (author_id, title, sort_title, monitored) VALUES (?, 'Unmonitored', 'Unmonitored', 0)`, authorID)
	require.NoError(t, err)
	bookID, _ := result.LastInsertId()

	_, err = svc.SearchAndGrab(ctx, bookID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not monitored")
}

func TestSearchAndGrab_ActiveDownloadExists(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")
	testutil.SeedQueueItem(t, db, bookID, "The Hobbit EPUB", "downloading")

	_, err := svc.SearchAndGrab(ctx, bookID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "active download")
}

func TestSearchAndGrab_NoProviders(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	_, err := svc.SearchAndGrab(ctx, bookID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no suitable download found")
}

func TestGetBookInfo(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	title, authorName, err := svc.GetBookInfo(ctx, bookID)
	require.NoError(t, err)
	assert.Equal(t, "The Hobbit", title)
	assert.Equal(t, "Tolkien", authorName)
}

func TestGetBookInfo_NotFound(t *testing.T) {
	svc, ctx := newTestService(t)

	_, _, err := svc.GetBookInfo(ctx, 9999)
	require.Error(t, err)
}

func TestGetPendingSourcesCount_NoPending(t *testing.T) {
	svc, ctx := newTestService(t)

	count := svc.GetPendingSourcesCount(ctx, 9999)
	assert.Equal(t, 0, count)
}

func TestGetPendingSourcesCount_WithPending(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	_, err := db.Exec(`INSERT INTO search_results (book_id, provider, title, download_url, format, size, score, priority, status) VALUES (?, 'archive', 'Result 1', 'http://a.com', 'epub', 1024, 100, 0, 'pending')`, bookID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO search_results (book_id, provider, title, download_url, format, size, score, priority, status) VALUES (?, 'archive', 'Result 2', 'http://b.com', 'epub', 2048, 90, 1, 'pending')`, bookID)
	require.NoError(t, err)
	_, err = db.Exec(`INSERT INTO search_results (book_id, provider, title, download_url, format, size, score, priority, status) VALUES (?, 'archive', 'Result 3', 'http://c.com', 'epub', 3072, 80, 2, 'tried')`, bookID)
	require.NoError(t, err)

	count := svc.GetPendingSourcesCount(ctx, bookID)
	assert.Equal(t, 2, count)
}

func TestProcessDownloads_EmptyQueue(t *testing.T) {
	svc, ctx := newTestService(t)

	result, err := svc.ProcessDownloads(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Checked)
	assert.Equal(t, 0, result.Completed)
	assert.Equal(t, 0, result.Failed)
}

func TestSearchAllWanted_NoBooks(t *testing.T) {
	svc, ctx := newTestService(t)

	results, err := svc.SearchAllWanted(ctx)
	require.NoError(t, err)
	assert.Empty(t, results)
}

// TestSearchAllWanted_WithBooksAndNoProviders covers the SearchAllWanted loop
// body where SearchAndGrab fails for every book (no providers configured) and
// the function returns an empty result list without an error.
func TestSearchAllWanted_WithBooksAndNoProviders(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	testutil.SeedBook(t, db, authorID, "The Hobbit")
	testutil.SeedBook(t, db, authorID, "LOTR")

	results, err := svc.SearchAllWanted(ctx)
	require.NoError(t, err)
	assert.Empty(t, results)
}
