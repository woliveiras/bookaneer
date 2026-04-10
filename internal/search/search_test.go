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
