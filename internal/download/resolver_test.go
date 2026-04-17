package download_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/bypass"
	"github.com/woliveiras/bookaneer/internal/download"
	_ "github.com/woliveiras/bookaneer/internal/download/direct"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

// TestGetDirectClient_ConfiguredInDB verifies that when a direct client exists in the database,
// GetDirectClient returns it (exercising getClientByType → getClientByTypes → getOrCreateClient).
func TestGetDirectClient_ConfiguredInDB(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db, bypass.Noop{})
	ctx := context.Background()

	dir := t.TempDir()
	cfg := &download.ClientConfig{
		Name:        "Configured Direct",
		Type:        download.ClientTypeDirect,
		DownloadDir: dir,
		Enabled:     true,
	}
	require.NoError(t, svc.CreateClient(ctx, cfg))

	client, got, err := svc.GetDirectClient(ctx)
	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "Configured Direct", got.Name)
	assert.Equal(t, download.ClientTypeDirect, got.Type)
	assert.Equal(t, dir, got.DownloadDir)
}

// TestGetOrCreateClient_CachesClient verifies that the same client instance is returned on
// repeated calls (exercises getOrCreateClient double-check locking).
func TestGetOrCreateClient_CachesClient(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db, bypass.Noop{})
	ctx := context.Background()

	dir := t.TempDir()
	cfg := &download.ClientConfig{
		Name:        "Cached Direct",
		Type:        download.ClientTypeDirect,
		DownloadDir: dir,
		Enabled:     true,
	}
	require.NoError(t, svc.CreateClient(ctx, cfg))

	c1, cfg1, err := svc.GetDirectClient(ctx)
	require.NoError(t, err)

	c2, cfg2, err := svc.GetDirectClient(ctx)
	require.NoError(t, err)

	assert.Equal(t, c1, c2, "should return the cached client instance")
	assert.Equal(t, cfg1.Name, cfg2.Name)
}

// TestGetClientQueue_WithDirectClient exercises getOrCreateClient via the public GetClientQueue
// path (GetClient → getOrCreateClient → GetQueue).
func TestGetClientQueue_WithDirectClient(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db, bypass.Noop{})
	ctx := context.Background()

	dir := t.TempDir()
	cfg := &download.ClientConfig{
		Name:        "Direct Queue",
		Type:        download.ClientTypeDirect,
		DownloadDir: dir,
		Enabled:     true,
	}
	require.NoError(t, svc.CreateClient(ctx, cfg))

	queue, err := svc.GetClientQueue(ctx, cfg.ID)
	require.NoError(t, err)
	assert.Empty(t, queue)
}

// TestGetQueue_WithEnabledDirectClient verifies GetQueue aggregates results from all
// enabled clients (exercises getOrCreateClient via the GetQueue loop).
func TestGetQueue_WithEnabledDirectClient(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db, bypass.Noop{})
	ctx := context.Background()

	dir := t.TempDir()
	require.NoError(t, svc.CreateClient(ctx, &download.ClientConfig{
		Name:        "Active Direct",
		Type:        download.ClientTypeDirect,
		DownloadDir: dir,
		Enabled:     true,
	}))

	queue, err := svc.GetQueue(ctx)
	require.NoError(t, err)
	assert.Empty(t, queue) // direct client has no queued downloads initially
}

// TestGetClientByTypes_DisabledClientSkipped verifies that GetQueue skips disabled clients.
func TestGetClientByTypes_DisabledClientSkipped(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db, bypass.Noop{})
	ctx := context.Background()

	dir := t.TempDir()
	require.NoError(t, svc.CreateClient(ctx, &download.ClientConfig{
		Name:        "Disabled Direct",
		Type:        download.ClientTypeDirect,
		DownloadDir: dir,
		Enabled:     false,
	}))

	// GetQueue only processes enabled clients
	queue, err := svc.GetQueue(ctx)
	require.NoError(t, err)
	assert.Empty(t, queue)
}

// TestGetClientByTypes_WithUnknownFactoryType verifies that when a client exists in DB with
// an unregistered type, getOrCreateClient returns an error (NewClient fails), and
// GetClientQueue propagates that error.
func TestGetClientByTypes_WithUnknownFactoryType(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := download.NewService(db, bypass.Noop{})
	ctx := context.Background()

	cfg := &download.ClientConfig{
		Name:    "Unknown Factory",
		Type:    "not-a-real-client-type",
		Host:    "localhost",
		Enabled: true,
	}
	require.NoError(t, svc.CreateClient(ctx, cfg))

	_, err := svc.GetClientQueue(ctx, cfg.ID)
	require.Error(t, err)
}
