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

func newTestService(t *testing.T) (*wanted.Service, context.Context) {
	t.Helper()
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
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
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
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
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
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
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
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
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
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

// TestGetDownloadQueue_WithCompletedAndFailedItems verifies that GetDownloadQueue
// returns items regardless of their status (completed, failed, imported).
func TestGetDownloadQueue_WithCompletedAndFailedItems(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	testutil.SeedQueueItem(t, db, bookID, "Completed Release", "completed")
	testutil.SeedQueueItem(t, db, bookID, "Failed Release", "failed")
	testutil.SeedQueueItem(t, db, bookID, "Imported Release", "imported")

	queue, err := svc.GetDownloadQueue(ctx)
	require.NoError(t, err)
	assert.Len(t, queue, 3)

	statuses := make(map[string]bool)
	for _, item := range queue {
		statuses[item.Status] = true
	}
	assert.True(t, statuses["completed"])
	assert.True(t, statuses["failed"])
	assert.True(t, statuses["imported"])
}

// TestGrabRelease_RecordsDownloadAndHistory calls GrabRelease end-to-end to
// exercise recordDownload (inserts into download_queue) and recordHistory.
// Requires the embedded direct client factory (imported via import_test.go).
func TestGrabRelease_RecordsDownloadAndHistory(t *testing.T) {
	db := testutil.OpenTestDB(t)
	dir := t.TempDir()
	testutil.SeedRootFolder(t, db, dir, "Library")

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
	ctx := context.Background()

	// Port 1 is always refused — the background download fails fast, but
	// recordDownload and recordHistory run synchronously before any download.
	result, err := svc.GrabRelease(ctx, bookID, "http://127.0.0.1:1/hobbit.epub", "The Hobbit.epub", 1024)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, bookID, result.BookID)
	assert.Equal(t, "manual", result.Source)
	assert.Equal(t, "epub", result.Format)

	// recordDownload must have inserted a row in download_queue.
	queue, err := svc.GetDownloadQueue(ctx)
	require.NoError(t, err)
	require.Len(t, queue, 1)
	assert.Equal(t, "The Hobbit.epub", queue[0].Title)
	assert.Equal(t, int64(1024), queue[0].Size)

	// recordHistory must have inserted a 'grabbed' event.
	history, err := svc.GetHistory(ctx, 10, "grabbed")
	require.NoError(t, err)
	assert.Len(t, history, 1)
}

// TestProcessDownloads_RestartsLostDownload exercises restartDownload via
// ProcessDownloads: the queue item exists in the DB but not in the embedded
// client's in-memory map, so ProcessDownloads restarts it.
// Requires the embedded direct client factory (imported via import_test.go).
func TestProcessDownloads_RestartsLostDownload(t *testing.T) {
	db := testutil.OpenTestDB(t)
	dir := t.TempDir()
	testutil.SeedRootFolder(t, db, dir, "Library")

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	// Use a URL that fails fast (port 1 always refused) to avoid slow goroutines.
	_, err := db.Exec(`
		INSERT INTO download_queue (book_id, title, size, format, status, download_url, external_id)
		VALUES (?, 'The Hobbit EPUB', 1024, 'epub', 'queued', 'http://127.0.0.1:1/book.epub', 'ext-lost')
	`, bookID)
	require.NoError(t, err)

	var queueID int64
	require.NoError(t, db.QueryRow(`SELECT id FROM download_queue WHERE external_id = 'ext-lost'`).Scan(&queueID))

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
	ctx := context.Background()

	result, err := svc.ProcessDownloads(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Checked)

	// restartDownload should have updated external_id to a new value.
	var newExtID string
	require.NoError(t, db.QueryRow(`SELECT external_id FROM download_queue WHERE id = ?`, queueID).Scan(&newExtID))
	assert.NotEqual(t, "ext-lost", newExtID)
	assert.NotEmpty(t, newExtID)
}

func TestUpdateQueueItemStatusWithPath(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
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

func TestRemoveFromQueue_DBError(t *testing.T) {
	db := testutil.OpenTestDB(t)
	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db))
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	queueID := testutil.SeedQueueItem(t, db, bookID, "Release", "queued")

	db.Close()

	err := svc.RemoveFromQueue(ctx, queueID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete query failed")
}
