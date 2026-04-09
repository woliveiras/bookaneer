package search

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func newTestIndexerConfig(name string) IndexerConfig {
	return IndexerConfig{
		Name:                    name,
		Type:                    "newznab",
		BaseURL:                 "http://localhost:9999",
		APIPath:                 "/api",
		APIKey:                  "test-key",
		Categories:              "7020,8000",
		Priority:                1,
		Enabled:                 true,
		EnableRSS:               true,
		EnableAutomaticSearch:   true,
		EnableInteractiveSearch: true,
	}
}

func TestCreateIndexer(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	cfg := newTestIndexerConfig("NZBgeek")
	id, err := s.CreateIndexer(ctx, cfg)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))
}

func TestGetIndexer(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	cfg := newTestIndexerConfig("NZBgeek")
	id, err := s.CreateIndexer(ctx, cfg)
	require.NoError(t, err)

	got, err := s.GetIndexer(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "NZBgeek", got.Name)
	assert.Equal(t, "newznab", got.Type)
	assert.True(t, got.Enabled)
}

func TestGetIndexer_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	_, err := s.GetIndexer(ctx, 999)
	assert.ErrorIs(t, err, ErrIndexerNotFound)
}

func TestListIndexers(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	_, err := s.CreateIndexer(ctx, newTestIndexerConfig("Alpha"))
	require.NoError(t, err)
	cfg := newTestIndexerConfig("Beta")
	cfg.Priority = 2
	_, err = s.CreateIndexer(ctx, cfg)
	require.NoError(t, err)

	indexers, err := s.ListIndexers(ctx)
	require.NoError(t, err)
	assert.Len(t, indexers, 2)
	assert.Equal(t, "Alpha", indexers[0].Name, "should be sorted by priority")
}

func TestUpdateIndexer(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	id, err := s.CreateIndexer(ctx, newTestIndexerConfig("Old Name"))
	require.NoError(t, err)

	cfg := newTestIndexerConfig("New Name")
	cfg.ID = id
	err = s.UpdateIndexer(ctx, cfg)
	require.NoError(t, err)

	got, err := s.GetIndexer(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "New Name", got.Name)
}

func TestUpdateIndexer_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	cfg := newTestIndexerConfig("Ghost")
	cfg.ID = 999
	err := s.UpdateIndexer(ctx, cfg)
	assert.ErrorIs(t, err, ErrIndexerNotFound)
}

func TestDeleteIndexer(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	id, err := s.CreateIndexer(ctx, newTestIndexerConfig("ToDelete"))
	require.NoError(t, err)

	err = s.DeleteIndexer(ctx, id)
	require.NoError(t, err)

	_, err = s.GetIndexer(ctx, id)
	assert.ErrorIs(t, err, ErrIndexerNotFound)
}

func TestDeleteIndexer_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	err := s.DeleteIndexer(ctx, 999)
	assert.ErrorIs(t, err, ErrIndexerNotFound)
}

func TestGetOptions(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	opts, err := s.GetOptions(ctx)
	require.NoError(t, err)
	assert.NotNil(t, opts)
	assert.GreaterOrEqual(t, opts.RSSSyncInterval, 0)
}

func TestUpdateOptions(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	err := s.UpdateOptions(ctx, IndexerOptions{
		MinimumAge:         30,
		Retention:          365,
		MaximumSize:        500,
		RSSSyncInterval:    15,
		PreferIndexerFlags: true,
		AvailabilityDelay:  1,
	})
	require.NoError(t, err)

	opts, err := s.GetOptions(ctx)
	require.NoError(t, err)
	assert.Equal(t, 30, opts.MinimumAge)
	assert.Equal(t, 365, opts.Retention)
	assert.True(t, opts.PreferIndexerFlags)
}
