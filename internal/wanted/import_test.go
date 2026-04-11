package wanted_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/core/naming"
	"github.com/woliveiras/bookaneer/internal/download"
	_ "github.com/woliveiras/bookaneer/internal/download/direct" // register embedded downloader factory
	"github.com/woliveiras/bookaneer/internal/testutil"
	"github.com/woliveiras/bookaneer/internal/wanted"
)

// TestProcessDownloads_ImportsCompletedFile exercises importPendingCompletedDownloads,
// importCompletedDownload, and recordHistory through the ProcessDownloads public API.
func TestProcessDownloads_ImportsCompletedFile(t *testing.T) {
	db := testutil.OpenTestDB(t)
	libDir := t.TempDir()
	testutil.SeedRootFolder(t, db, libDir, "Library")

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	// Create a fake epub file in a download staging directory.
	srcFile := filepath.Join(t.TempDir(), "hobbit.epub")
	require.NoError(t, os.WriteFile(srcFile, []byte("epub content"), 0644))

	// Seed a completed queue item whose save_path points to the file above.
	_, err := db.Exec(`
		INSERT INTO download_queue (book_id, title, size, format, status, download_url, external_id, save_path)
		VALUES (?, 'The Hobbit EPUB', 1024, 'epub', 'completed', 'http://127.0.0.1:1/book.epub', 'ext-done', ?)
	`, bookID, srcFile)
	require.NoError(t, err)

	var queueID int64
	require.NoError(t, db.QueryRow(`SELECT id FROM download_queue WHERE external_id = 'ext-done'`).Scan(&queueID))

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db), nil)
	ctx := context.Background()

	result, err := svc.ProcessDownloads(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Imported)
	assert.Equal(t, 0, result.Checked) // no active (queued/downloading) items

	// Verify book_files was populated by importCompletedDownload.
	var fileCount int
	require.NoError(t, db.QueryRow(`SELECT COUNT(*) FROM book_files WHERE book_id = ?`, bookID).Scan(&fileCount))
	assert.Equal(t, 1, fileCount)

	// Verify queue item was updated to 'completed'.
	var status string
	require.NoError(t, db.QueryRow(`SELECT status FROM download_queue WHERE id = ?`, queueID).Scan(&status))
	assert.Equal(t, "completed", status)

	// Verify recordHistory was called inside importCompletedDownload.
	history, err := svc.GetHistory(ctx, 10, "bookImported")
	require.NoError(t, err)
	require.Len(t, history, 1)
	assert.Equal(t, "The Hobbit", history[0].BookTitle)
}

// TestProcessDownloads_MarksFailedWhenFileGone exercises the branch inside
// importPendingCompletedDownloads that marks a queue item 'failed' when the
// saved file no longer exists on disk.
func TestProcessDownloads_MarksFailedWhenFileGone(t *testing.T) {
	db := testutil.OpenTestDB(t)

	authorID := testutil.SeedAuthor(t, db, "Tolkien")
	bookID := testutil.SeedBook(t, db, authorID, "The Hobbit")

	// Seed a completed item pointing to a path that does not exist.
	_, err := db.Exec(`
		INSERT INTO download_queue (book_id, title, size, format, status, download_url, external_id, save_path)
		VALUES (?, 'The Hobbit EPUB', 1024, 'epub', 'completed', 'http://127.0.0.1:1/book.epub', 'ext-gone', '/nonexistent/path/hobbit.epub')
	`, bookID)
	require.NoError(t, err)

	var queueID int64
	require.NoError(t, db.QueryRow(`SELECT id FROM download_queue WHERE external_id = 'ext-gone'`).Scan(&queueID))

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db)
	svc := wanted.New(db, bookSvc, nil, nil, downloadSvc, naming.New(db), nil)
	ctx := context.Background()

	result, err := svc.ProcessDownloads(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, result.Imported)

	// Queue item must be marked 'failed' because the file is missing.
	var status string
	require.NoError(t, db.QueryRow(`SELECT status FROM download_queue WHERE id = ?`, queueID).Scan(&status))
	assert.Equal(t, "failed", status)
}
