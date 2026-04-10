package wanted_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/download"
	"github.com/woliveiras/bookaneer/internal/testutil"
	"github.com/woliveiras/bookaneer/internal/wanted"
)

func newTestService(t *testing.T) (*wanted.Service, context.Context) {
	t.Helper()
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	return svc, context.Background()
}

func seedTestBook(t *testing.T, db interface {
	Exec(string, ...any) (interface{ LastInsertId() (int64, error) }, error)
}) (int64, int64) {
	t.Helper()
	// We need raw DB access for seeding, so this helper won't work with the service alone.
	// Use testutil.SeedAuthor and testutil.SeedBook instead.
	return 0, 0
}

func TestRemoveFromQueue_ItemExists(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "J.R.R. Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")
	queueID := testutil.SeedQueueItem(t, db, bookID, "The Hobbit EPUB", "queued")

	err := svc.RemoveFromQueue(ctx, queueID)
	require.NoError(t, err)

	// Verify it's gone
	queue, err := svc.GetDownloadQueue(ctx)
	require.NoError(t, err)
	assert.Empty(t, queue)
}

func TestRemoveFromQueue_ItemNotFound(t *testing.T) {
	svc, ctx := newTestService(t)

	err := svc.RemoveFromQueue(ctx, 9999)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestGetDownloadQueue_Empty(t *testing.T) {
	svc, ctx := newTestService(t)

	queue, err := svc.GetDownloadQueue(ctx)
	require.NoError(t, err)
	assert.Empty(t, queue)
}

func TestGetDownloadQueue_WithItems(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "J.R.R. Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")
	testutil.SeedQueueItem(t, db, bookID, "The Hobbit EPUB", "queued")
	testutil.SeedQueueItem(t, db, bookID, "The Hobbit PDF", "downloading")

	queue, err := svc.GetDownloadQueue(ctx)
	require.NoError(t, err)
	assert.Len(t, queue, 2)

	// Verify fields are populated
	for _, item := range queue {
		assert.NotZero(t, item.ID)
		assert.Equal(t, bookID, item.BookID)
		assert.Equal(t, "The Hobbit", item.BookTitle)
		assert.Equal(t, "Embedded Downloader", item.ClientName) // No client configured
	}
}

func TestGetDownloadQueue_NullJoins(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")

	// Insert queue item with NULL download_client_id and NULL indexer_id
	_, err := db.Exec(`
		INSERT INTO download_queue (book_id, title, size, format, status, download_url)
		VALUES (?, 'Test Release', 2048, 'pdf', 'queued', 'https://example.com/test.pdf')`,
		bookID)
	require.NoError(t, err)

	queue, err := svc.GetDownloadQueue(ctx)
	require.NoError(t, err)
	require.Len(t, queue, 1)
	assert.Nil(t, queue[0].DownloadClientID)
	assert.Nil(t, queue[0].IndexerID)
	assert.Equal(t, "Embedded Downloader", queue[0].ClientName)
}

func TestUpdateQueueItemStatus(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	queueID := testutil.SeedQueueItem(t, db, bookID, "Release", "queued")

	err := svc.UpdateQueueItemStatus(ctx, queueID, "downloading", 42.5)
	require.NoError(t, err)

	// Verify
	queue, err := svc.GetDownloadQueue(ctx)
	require.NoError(t, err)
	require.Len(t, queue, 1)
	assert.Equal(t, "downloading", queue[0].Status)
	assert.InDelta(t, 42.5, queue[0].Progress, 0.01)
}

func TestUpdateQueueItemStatusWithPath(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	queueID := testutil.SeedQueueItem(t, db, bookID, "Release", "downloading")

	err := svc.UpdateQueueItemStatusWithPath(ctx, queueID, "completed", 100.0, "/library/Author/Book.epub")
	require.NoError(t, err)

	// Verify via raw query (save_path is not in GetDownloadQueue response)
	var status string
	var savePath string
	err = db.QueryRow(`SELECT status, save_path FROM download_queue WHERE id = ?`, queueID).Scan(&status, &savePath)
	require.NoError(t, err)
	assert.Equal(t, "completed", status)
	assert.Equal(t, "/library/Author/Book.epub", savePath)
}
