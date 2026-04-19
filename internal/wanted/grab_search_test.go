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
	"github.com/woliveiras/bookaneer/internal/core/naming"
	"github.com/woliveiras/bookaneer/internal/bypass"
	"github.com/woliveiras/bookaneer/internal/download"
	_ "github.com/woliveiras/bookaneer/internal/download/direct" // register embedded downloader factory
	"github.com/woliveiras/bookaneer/internal/testutil"
	"github.com/woliveiras/bookaneer/internal/wanted"
)

// TestGrabRelease_EmptyTitle covers the branch in GrabRelease where
// releaseTitle is "" and the filename is derived from AuthorName + BookTitle.
func TestGrabRelease_EmptyTitle(t *testing.T) {
	db := testutil.OpenTestDBX(t)
	libDir := t.TempDir()
	testutil.SeedRootFolder(t, db.DB, libDir, "Library")

	authorID := testutil.SeedAuthor(t, db.DB, "Tolkien")
	bookID := testutil.SeedBook(t, db.DB, authorID, "The Hobbit")

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db.DB, bypass.Noop{})
	svc := wanted.New(db.DB, bookSvc, nil, nil, downloadSvc, naming.New(db), nil, nil)
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
	db := testutil.OpenTestDBX(t)
	// No root folder → GetDirectClient errors.
	authorID := testutil.SeedAuthor(t, db.DB, "Tolkien")
	bookID := testutil.SeedBook(t, db.DB, authorID, "The Hobbit")

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db.DB, bypass.Noop{})
	svc := wanted.New(db.DB, bookSvc, nil, nil, downloadSvc, naming.New(db), nil, nil)
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

// TestGrabRelease_ConfiguredDirectClient covers the cfg.ID != 0 branch in
// GrabRelease when a real 'direct' download client is configured in the DB.
func TestGrabRelease_ConfiguredDirectClient(t *testing.T) {
	dlDir := t.TempDir()

	db := testutil.OpenTestDBX(t)
	authorID := testutil.SeedAuthor(t, db.DB, "Tolkien")
	bookID := testutil.SeedBook(t, db.DB, authorID, "The Hobbit")

	_, err := db.Exec(`
INSERT INTO download_clients (name, type, host, port, enabled, priority, download_dir)
VALUES ('DirectClient', 'direct', '', 0, 1, 0, ?)
`, dlDir)
	require.NoError(t, err)

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db.DB, bypass.Noop{})
	svc := wanted.New(db.DB, bookSvc, nil, nil, downloadSvc, naming.New(db), nil, nil)
	ctx := context.Background()

	result, err := svc.GrabRelease(ctx, bookID, "http://127.0.0.1:1/test.epub", "Test.epub", 1024)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "manual", result.Source)
	assert.Equal(t, "DirectClient", result.ClientName)

	queue, err := svc.GetDownloadQueue(ctx)
	require.NoError(t, err)
	require.Len(t, queue, 1)
	assert.NotNil(t, queue[0].DownloadClientID, "expected non-nil client ID from configured client")
}

// TestProcessDownloads_ImportFailsNoRootFolder covers the
// slog.Warn("Failed to import download") branch in ProcessDownloads when
// importCompletedDownload fails because the root folder is removed from DB.
func TestProcessDownloads_ImportFailsNoRootFolder(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Disposition", `attachment; filename="hobbit.epub"`)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("epub"))
	}))
	t.Cleanup(srv.Close)

	db := testutil.OpenTestDBX(t)
	libDir := t.TempDir()
	folderID := testutil.SeedRootFolder(t, db.DB, libDir, "Library")

	authorID := testutil.SeedAuthor(t, db.DB, "Tolkien")
	bookID := testutil.SeedBook(t, db.DB, authorID, "The Hobbit")

	bookSvc := book.New(db)
	downloadSvc := download.NewService(db.DB, bypass.Noop{})
	svc := wanted.New(db.DB, bookSvc, nil, nil, downloadSvc, naming.New(db), nil, nil)
	ctx := context.Background()

	_, err := svc.GrabRelease(ctx, bookID, srv.URL+"/hobbit.epub", "The Hobbit.epub", 1024)
	require.NoError(t, err)

	_, err = db.Exec("DELETE FROM root_folders WHERE id = ?", folderID)
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	result, err := svc.ProcessDownloads(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, result.Checked)
	assert.Equal(t, 1, result.Completed)
	assert.Equal(t, 0, result.Imported)
}
