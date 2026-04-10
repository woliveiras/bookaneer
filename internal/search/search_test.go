package search_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/search"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestNewService(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := search.NewService(db)
	assert.NotNil(t, svc)
}

func TestLoadIndexers_NoIndexers(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := search.NewService(db)
	ctx := context.Background()

	err := svc.LoadIndexers(ctx)
	require.NoError(t, err)
}

func TestSearch_NoLoadedIndexers(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := search.NewService(db)
	ctx := context.Background()

	err := svc.LoadIndexers(ctx)
	require.NoError(t, err)

	results, err := svc.Search(ctx, search.SearchQuery{Query: "lord of the rings"})
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestListIndexers_Multiple(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := search.NewService(db)
	ctx := context.Background()

	_, err := svc.CreateIndexer(ctx, search.IndexerConfig{Name: "A", Type: "newznab", BaseURL: "http://a.com", Enabled: true})
	require.NoError(t, err)
	_, err = svc.CreateIndexer(ctx, search.IndexerConfig{Name: "B", Type: "torznab", BaseURL: "http://b.com", Enabled: true})
	require.NoError(t, err)

	indexers, err := svc.ListIndexers(ctx)
	require.NoError(t, err)
	assert.Len(t, indexers, 2)
}

func TestUpdateIndexer_ChangeName(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := search.NewService(db)
	ctx := context.Background()

	id, err := svc.CreateIndexer(ctx, search.IndexerConfig{Name: "Old", Type: "newznab", BaseURL: "http://old.com", Enabled: true})
	require.NoError(t, err)

	err = svc.UpdateIndexer(ctx, search.IndexerConfig{ID: id, Name: "New", Type: "newznab", BaseURL: "http://new.com", Enabled: true})
	require.NoError(t, err)

	got, err := svc.GetIndexer(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "New", got.Name)
}

func TestRegisterFactory(t *testing.T) {
	search.RegisterFactory("test-type", nil)
	_, ok := search.GetFactory("test-type")
	assert.True(t, ok)
}

func TestGetFactory_Unknown(t *testing.T) {
	_, ok := search.GetFactory("nonexistent-xyz")
	assert.False(t, ok)
}

// TestTestIndexer_UnknownType exercises the error path of Service.TestIndexer when no
// factory is registered for the given type (covers 0% TestIndexer function lines).
func TestTestIndexer_UnknownType(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := search.NewService(db)
	ctx := context.Background()

	err := svc.TestIndexer(ctx, search.IndexerConfig{
		Type:    "no-such-indexer-type-for-external-test",
		BaseURL: "http://test.com",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown indexer type")
}

// TestListIndexers_Empty verifies that ListIndexers returns an empty slice when no
// indexers have been created.
func TestListIndexers_Empty(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := search.NewService(db)
	ctx := context.Background()

	indexers, err := svc.ListIndexers(ctx)
	require.NoError(t, err)
	assert.Empty(t, indexers)
}

// TestLoadIndexers_WithEnabledAndDisabled verifies that LoadIndexers succeeds when the
// DB contains both enabled and disabled indexers (only enabled ones are loaded into
// the in-memory client map).
func TestLoadIndexers_WithEnabledAndDisabled(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := search.NewService(db)
	ctx := context.Background()

	_, err := svc.CreateIndexer(ctx, search.IndexerConfig{
		Name: "EnabledIdx", Type: "newznab", BaseURL: "http://enabled.com", Enabled: true,
	})
	require.NoError(t, err)
	_, err = svc.CreateIndexer(ctx, search.IndexerConfig{
		Name: "DisabledIdx", Type: "newznab", BaseURL: "http://disabled.com", Enabled: false,
	})
	require.NoError(t, err)

	require.NoError(t, svc.LoadIndexers(ctx))

	// Both records exist in the DB.
	all, err := svc.ListIndexers(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

// TestUpdateOptions_FlagsOff verifies UpdateOptions round-trips with
// PreferIndexerFlags = false (the preferFlags == 0 branch).
func TestUpdateOptions_FlagsOff(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := search.NewService(db)
	ctx := context.Background()

	err := svc.UpdateOptions(ctx, search.IndexerOptions{
		RSSSyncInterval:    60,
		PreferIndexerFlags: false,
	})
	require.NoError(t, err)

	opts, err := svc.GetOptions(ctx)
	require.NoError(t, err)
	assert.False(t, opts.PreferIndexerFlags)
	assert.Equal(t, 60, opts.RSSSyncInterval)
}

// TestSearch_NoIndexersLoaded verifies Search returns empty results (not an error)
// when LoadIndexers has not been called or no indexers are registered.
func TestSearch_NoIndexersLoaded(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := search.NewService(db)
	ctx := context.Background()

	results, err := svc.Search(ctx, search.SearchQuery{Query: "anything"})
	require.NoError(t, err)
	assert.Empty(t, results)
}

