package download_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/download"
	_ "github.com/woliveiras/bookaneer/internal/download/direct"
	"github.com/woliveiras/bookaneer/internal/testutil"
	_ "modernc.org/sqlite"
)

func TestNewService(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	assert.NotNil(t, svc)
}

func TestListGrabs_Empty(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	grabs, err := svc.ListGrabs(context.Background())
	require.NoError(t, err)
	assert.Empty(t, grabs)
}

func TestCreateGrab(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	_, err := db.Exec(`INSERT INTO indexers (name, type, base_url, enabled) VALUES ('test', 'newznab', 'http://test.com', 1)`)
	require.NoError(t, err)
	var indexerID int64
	require.NoError(t, db.QueryRow("SELECT id FROM indexers LIMIT 1").Scan(&indexerID))
	_, err = db.Exec(`INSERT INTO download_clients (name, type, host, port, enabled) VALUES ('dc', 'direct', 'localhost', 0, 1)`)
	require.NoError(t, err)
	var clientID int64
	require.NoError(t, db.QueryRow("SELECT id FROM download_clients LIMIT 1").Scan(&clientID))

	grab := &download.GrabItem{
		BookID:       bookID,
		IndexerID:    indexerID,
		ClientID:     clientID,
		ReleaseTitle: "Test Book EPUB",
		DownloadURL:  "https://example.com/book.epub",
		Size:         1024,
		Quality:      "epub",
	}
	err = svc.CreateGrab(ctx, grab)
	require.NoError(t, err)
	assert.NotZero(t, grab.ID)
	assert.Equal(t, download.GrabStatusPending, grab.Status)
}

func TestListGrabs_WithData(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	_, _ = db.Exec(`INSERT INTO indexers (name, type, base_url, enabled) VALUES ('idx', 'newznab', 'http://x.com', 1)`)
	var indexerID int64
	_ = db.QueryRow("SELECT id FROM indexers LIMIT 1").Scan(&indexerID)
	_, _ = db.Exec(`INSERT INTO download_clients (name, type, host, port, enabled) VALUES ('dc', 'direct', 'localhost', 0, 1)`)
	var clientID int64
	_ = db.QueryRow("SELECT id FROM download_clients LIMIT 1").Scan(&clientID)

	require.NoError(t, svc.CreateGrab(ctx, &download.GrabItem{
		BookID: bookID, IndexerID: indexerID, ClientID: clientID, ReleaseTitle: "Book", DownloadURL: "http://a.com", Size: 100, Quality: "epub",
	}))

	grabs, err := svc.ListGrabs(ctx)
	require.NoError(t, err)
	assert.Len(t, grabs, 1)
	assert.Equal(t, "Book", grabs[0].ReleaseTitle)
}

func TestGetQueue_NoClients(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	queue, err := svc.GetQueue(context.Background())
	require.NoError(t, err)
	assert.Empty(t, queue)
}

func TestGetClientQueue_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	_, err := svc.GetClientQueue(context.Background(), 9999)
	require.Error(t, err)
}

func TestRegisterFactory(t *testing.T) {
	download.RegisterFactory("test-type-dl", func(cfg download.ClientConfig) (download.Client, error) {
		return nil, nil
	})
}

func TestNewClient_UnknownType(t *testing.T) {
	_, err := download.NewClient(download.ClientConfig{Type: "nonexistent"})
	require.Error(t, err)
}

func TestGetDirectClient_NoRootFolder(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	_, _, err := svc.GetDirectClient(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "root folder")
}

func TestGetDirectClient_WithRootFolder(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	dir := t.TempDir()
	testutil.SeedRootFolder(t, db, dir, "Library")

	client, cfg, err := svc.GetDirectClient(ctx)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "Embedded Downloader", cfg.Name)
	assert.Equal(t, dir, cfg.DownloadDir)
}

func TestGetUsenetClient_NoneConfigured(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	client, cfg, err := svc.GetUsenetClient(context.Background())
	require.NoError(t, err)
	assert.Nil(t, client)
	assert.Nil(t, cfg)
}

func TestGetTorrentClient_NoneConfigured(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	client, cfg, err := svc.GetTorrentClient(context.Background())
	require.NoError(t, err)
	assert.Nil(t, client)
	assert.Nil(t, cfg)
}

func TestSendGrab_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	err := svc.SendGrab(context.Background(), 9999, 1)
	require.Error(t, err)
}

func TestGrabItem_UnmarshalJSON(t *testing.T) {
	data := []byte(`{"bookId":1,"releaseTitle":"Test","downloadUrl":"http://a.com","size":100}`)
	var g download.GrabItem
	err := json.Unmarshal(data, &g)
	require.NoError(t, err)
	assert.Equal(t, int64(1), g.BookID)
	assert.Equal(t, "Test", g.ReleaseTitle)
}

func TestTestClient_Direct(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	dir := t.TempDir()
	err := svc.TestClient(context.Background(), &download.ClientConfig{
		Name:        "DirectTest",
		Type:        download.ClientTypeDirect,
		DownloadDir: dir,
		Enabled:     true,
	})
	require.NoError(t, err)
}

func TestTestClient_UnknownType(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	err := svc.TestClient(context.Background(), &download.ClientConfig{
		Type: "totally-nonexistent-type",
	})
	require.Error(t, err)
}

func TestSendGrab_ClientNotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	_, err := db.Exec(`INSERT INTO indexers (name, type, base_url, enabled) VALUES ('idx', 'newznab', 'http://x.com', 1)`)
	require.NoError(t, err)
	var indexerID int64
	require.NoError(t, db.QueryRow("SELECT id FROM indexers LIMIT 1").Scan(&indexerID))

	// A placeholder client is needed to satisfy the FK constraint on grabs.client_id.
	placeholderCfg := &download.ClientConfig{
		Name: "Placeholder", Type: download.ClientTypeDirect, Host: "localhost", Enabled: false,
	}
	require.NoError(t, svc.CreateClient(ctx, placeholderCfg))

	grab := &download.GrabItem{
		BookID: bookID, IndexerID: indexerID, ClientID: placeholderCfg.ID,
		ReleaseTitle: "Book", DownloadURL: "http://a.com", Size: 100, Quality: "epub",
	}
	require.NoError(t, svc.CreateGrab(ctx, grab))

	err = svc.SendGrab(ctx, grab.ID, 9999)
	require.ErrorIs(t, err, download.ErrNotFound)
}

func TestSendGrab_DisabledClient(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	_, err := db.Exec(`INSERT INTO indexers (name, type, base_url, enabled) VALUES ('idx', 'newznab', 'http://x.com', 1)`)
	require.NoError(t, err)
	var indexerID int64
	require.NoError(t, db.QueryRow("SELECT id FROM indexers LIMIT 1").Scan(&indexerID))

	dir := t.TempDir()
	clientCfg := &download.ClientConfig{
		Name:        "Disabled",
		Type:        download.ClientTypeDirect,
		DownloadDir: dir,
		Enabled:     false,
	}
	require.NoError(t, svc.CreateClient(ctx, clientCfg))

	grab := &download.GrabItem{
		BookID: bookID, IndexerID: indexerID, ClientID: clientCfg.ID,
		ReleaseTitle: "Book", DownloadURL: "http://a.com", Size: 100, Quality: "epub",
	}
	require.NoError(t, svc.CreateGrab(ctx, grab))

	err = svc.SendGrab(ctx, grab.ID, clientCfg.ID)
	require.ErrorIs(t, err, download.ErrClientDisabled)
}

func TestSendGrab_WithDirectClient(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	_, err := db.Exec(`INSERT INTO indexers (name, type, base_url, enabled) VALUES ('idx', 'newznab', 'http://x.com', 1)`)
	require.NoError(t, err)
	var indexerID int64
	require.NoError(t, db.QueryRow("SELECT id FROM indexers LIMIT 1").Scan(&indexerID))

	dir := t.TempDir()
	clientCfg := &download.ClientConfig{
		Name:        "Direct",
		Type:        download.ClientTypeDirect,
		DownloadDir: dir,
		Enabled:     true,
	}
	require.NoError(t, svc.CreateClient(ctx, clientCfg))

	grab := &download.GrabItem{
		BookID: bookID, IndexerID: indexerID, ClientID: clientCfg.ID,
		ReleaseTitle: "Test Book EPUB",
		DownloadURL:  "http://example.com/book.epub", Size: 1024, Quality: "epub",
	}
	require.NoError(t, svc.CreateGrab(ctx, grab))

	err = svc.SendGrab(ctx, grab.ID, clientCfg.ID)
	require.NoError(t, err)

	grabs, err := svc.ListGrabs(ctx)
	require.NoError(t, err)
	require.Len(t, grabs, 1)
	assert.Equal(t, download.GrabStatusSent, grabs[0].Status)
	assert.NotEmpty(t, grabs[0].DownloadID)
}

func TestListGrabs_WithCompletedAt(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	authorID := testutil.SeedAuthor(t, db, "Author")
	bookID := testutil.SeedBook(t, db, authorID, "Book")
	_, err := db.Exec(`INSERT INTO indexers (name, type, base_url, enabled) VALUES ('idx', 'newznab', 'http://x.com', 1)`)
	require.NoError(t, err)
	var indexerID int64
	require.NoError(t, db.QueryRow("SELECT id FROM indexers LIMIT 1").Scan(&indexerID))

	// A download client is required to satisfy the grabs.client_id FK constraint.
	clientCfg := &download.ClientConfig{
		Name: "Direct", Type: download.ClientTypeDirect, DownloadDir: t.TempDir(), Enabled: true,
	}
	require.NoError(t, svc.CreateClient(ctx, clientCfg))

	grab := &download.GrabItem{
		BookID: bookID, IndexerID: indexerID, ClientID: clientCfg.ID,
		ReleaseTitle: "Completed Book",
		DownloadURL:  "http://a.com", Size: 500, Quality: "epub",
	}
	require.NoError(t, svc.CreateGrab(ctx, grab))

	_, err = db.ExecContext(ctx,
		"UPDATE grabs SET completed_at = datetime('now'), status = ? WHERE id = ?",
		download.GrabStatusCompleted, grab.ID)
	require.NoError(t, err)

	grabs, err := svc.ListGrabs(ctx)
	require.NoError(t, err)
	require.Len(t, grabs, 1)
	assert.Equal(t, download.GrabStatusCompleted, grabs[0].Status)
	assert.NotNil(t, grabs[0].CompletedAt)
}

func TestGetDirectClient_Cached(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db)
	ctx := context.Background()

	dir := t.TempDir()
	testutil.SeedRootFolder(t, db, dir, "Library")

	c1, cfg1, err := svc.GetDirectClient(ctx)
	require.NoError(t, err)

	c2, cfg2, err := svc.GetDirectClient(ctx)
	require.NoError(t, err)
	assert.Equal(t, c1, c2)
	assert.Equal(t, cfg1.Name, cfg2.Name)
}

// ── DB-closed error paths ──────────────────────────────────────────────────────

// TestGetQueue_DBError covers the error path in GetQueue when the underlying
// ListClients call fails because the DB is closed.
func TestGetQueue_DBError(t *testing.T) {
	// Use the openClosedDB helper defined in repository_test.go (same package).
	db, err := openClosedDBForService()
	require.NoError(t, err)
	svc := download.NewService(db)
	_, queueErr := svc.GetQueue(context.Background())
	require.Error(t, queueErr)
}

// TestListGrabs_DBError covers the db.QueryContext error path in ListGrabs.
func TestListGrabs_DBError(t *testing.T) {
	db, err := openClosedDBForService()
	require.NoError(t, err)
	svc := download.NewService(db)
	_, listErr := svc.ListGrabs(context.Background())
	require.Error(t, listErr)
}

// TestCreateGrab_DBError covers the db.ExecContext error path in CreateGrab.
func TestCreateGrab_DBError(t *testing.T) {
	db, err := openClosedDBForService()
	require.NoError(t, err)
	svc := download.NewService(db)
	grabErr := svc.CreateGrab(context.Background(), &download.GrabItem{
		BookID: 1, IndexerID: 1, ReleaseTitle: "Book", DownloadURL: "http://a.com",
	})
	require.Error(t, grabErr)
}

// openClosedDBForService is a package-level helper (mirrors openClosedDB in repository_test.go)
// that works without *testing.T so it can be used in inline variable declarations.
// It returns (nil, nil) on unexpected SQL open errors.
func openClosedDBForService() (*sql.DB, error) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}
	if err := db.Close(); err != nil {
		return nil, err
	}
	return db, nil
}
