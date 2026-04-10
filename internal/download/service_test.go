package download_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/download"
	_ "github.com/woliveiras/bookaneer/internal/download/direct"
	"github.com/woliveiras/bookaneer/internal/testutil"
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
