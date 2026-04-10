package search

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/testutil"
	_ "modernc.org/sqlite"
)

// mockIndexer is a test double for the Indexer interface.
type mockIndexer struct {
	indexerName string
	results     []Result
	searchErr   error
	testErr     error
}

func (m *mockIndexer) Name() string                                           { return m.indexerName }
func (m *mockIndexer) Type() string                                           { return "mock" }
func (m *mockIndexer) Search(_ context.Context, _ SearchQuery) ([]Result, error) {
	return m.results, m.searchErr
}
func (m *mockIndexer) Caps(_ context.Context) (*Capabilities, error) { return &Capabilities{}, nil }
func (m *mockIndexer) Test(_ context.Context) error                  { return m.testErr }

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

// TestTestIndexer_UnknownType covers the "unknown indexer type" error path (0% coverage).
func TestTestIndexer_UnknownType(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	err := s.TestIndexer(ctx, IndexerConfig{
		Type:    "totally-nonexistent-type-xyz",
		BaseURL: "http://test.com",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown indexer type")
}

// TestLoadIndexers_SkipsUnknownFactory verifies that LoadIndexers silently skips indexers
// whose type has no registered factory (covers the "!ok" branch inside the loop).
// Since no "newznab" factory is registered in this file by default, an enabled newznab
// indexer exercises that skip path.
func TestLoadIndexers_SkipsUnknownFactory(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	_, err := s.CreateIndexer(ctx, IndexerConfig{
		Name: "NoFactory", Type: "newznab", BaseURL: "http://nofactory.com", Enabled: true,
	})
	require.NoError(t, err)

	require.NoError(t, s.LoadIndexers(ctx))

	// No factory registered → no client loaded → Search returns nil (not error).
	results, err := s.Search(ctx, SearchQuery{Query: "test"})
	require.NoError(t, err)
	assert.Empty(t, results)
}

// withTestFactory registers a temporary factory for typeName and removes it via
// t.Cleanup so it does not pollute the global registry for subsequent tests.
// Accessing factoryReg directly is valid in internal tests (package search).
func withTestFactory(t *testing.T, typeName string, f IndexerFactory) {
	t.Helper()
	RegisterFactory(typeName, f)
	t.Cleanup(func() {
		factoryMu.Lock()
		delete(factoryReg, typeName)
		factoryMu.Unlock()
	})
}

// TestLoadIndexers_WithKnownType covers the success branch of loadClient (factory found,
// indexer created, stored in clients map).
func TestLoadIndexers_WithKnownType(t *testing.T) {
	withTestFactory(t, "newznab", func(cfg IndexerConfig) (Indexer, error) {
		return &mockIndexer{indexerName: cfg.Name}, nil
	})

	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	_, err := s.CreateIndexer(ctx, IndexerConfig{
		Name: "KnownMock", Type: "newznab", BaseURL: "http://mock.com", Enabled: true,
	})
	require.NoError(t, err)

	require.NoError(t, s.LoadIndexers(ctx))

	// Client was loaded; Search should reach it (returns empty, not error).
	results, err := s.Search(ctx, SearchQuery{Query: "test"})
	require.NoError(t, err)
	assert.Empty(t, results)
}

// TestLoadClient_FactoryError covers the "factory returns error" branch of loadClient.
// CreateIndexer must succeed even when the factory fails to build the in-memory client.
func TestLoadClient_FactoryError(t *testing.T) {
	withTestFactory(t, "newznab", func(_ IndexerConfig) (Indexer, error) {
		return nil, errors.New("simulated factory error")
	})

	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	id, err := s.CreateIndexer(ctx, IndexerConfig{
		Name: "FailFactory", Type: "newznab", BaseURL: "http://fail.com", Enabled: true,
	})
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// The record exists in DB even though the in-memory client wasn't created.
	got, err := s.GetIndexer(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, "FailFactory", got.Name)

	// After LoadIndexers the factory still fails → no client loaded → Search returns nil.
	require.NoError(t, s.LoadIndexers(ctx))
	results, err := s.Search(ctx, SearchQuery{Query: "test"})
	require.NoError(t, err)
	assert.Empty(t, results)
}

// TestSearch_WithMockIndexer covers the concurrent result-aggregation path in Search
// (currently at 25% coverage) when at least one indexer returns results.
func TestSearch_WithMockIndexer(t *testing.T) {
	expected := []Result{
		{Title: "Book One", DownloadURL: "http://dl.com/1"},
		{Title: "Book Two", DownloadURL: "http://dl.com/2"},
	}
	withTestFactory(t, "newznab", func(cfg IndexerConfig) (Indexer, error) {
		return &mockIndexer{indexerName: cfg.Name, results: expected}, nil
	})

	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	_, err := s.CreateIndexer(ctx, IndexerConfig{
		Name: "ResultsMock", Type: "newznab", BaseURL: "http://mock.com", Enabled: true,
	})
	require.NoError(t, err)
	require.NoError(t, s.LoadIndexers(ctx))

	results, err := s.Search(ctx, SearchQuery{Query: "book"})
	require.NoError(t, err)
	assert.Len(t, results, 2)
}

// TestSearch_AllIndexersFail covers the "all indexers failed" error branch in Search.
func TestSearch_AllIndexersFail(t *testing.T) {
	withTestFactory(t, "newznab", func(cfg IndexerConfig) (Indexer, error) {
		return &mockIndexer{
			indexerName: cfg.Name,
			searchErr:   errors.New("indexer unreachable"),
		}, nil
	})

	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	_, err := s.CreateIndexer(ctx, IndexerConfig{
		Name: "FailSearch", Type: "newznab", BaseURL: "http://mock.com", Enabled: true,
	})
	require.NoError(t, err)
	require.NoError(t, s.LoadIndexers(ctx))

	results, searchErr := s.Search(ctx, SearchQuery{Query: "book"})
	require.Error(t, searchErr)
	assert.Nil(t, results)
	assert.Contains(t, searchErr.Error(), "all indexers failed")
}

// ── TestIndexer additional paths ──────────────────────────────────────────────

// TestTestIndexer_FactoryError covers the "factory returns error" path in TestIndexer.
func TestTestIndexer_FactoryError(t *testing.T) {
	withTestFactory(t, "newznab", func(_ IndexerConfig) (Indexer, error) {
		return nil, errors.New("factory cannot create indexer")
	})

	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	err := s.TestIndexer(ctx, IndexerConfig{Type: "newznab", BaseURL: "http://test.com"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "create indexer")
}

// TestTestIndexer_TestError covers the path where the factory succeeds but
// indexer.Test() returns an error.
func TestTestIndexer_TestError(t *testing.T) {
	withTestFactory(t, "newznab", func(cfg IndexerConfig) (Indexer, error) {
		return &mockIndexer{testErr: errors.New("connection refused")}, nil
	})

	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	err := s.TestIndexer(ctx, IndexerConfig{Type: "newznab", BaseURL: "http://test.com"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

// TestTestIndexer_Success covers the happy path where factory succeeds and
// indexer.Test() returns nil.
func TestTestIndexer_Success(t *testing.T) {
	withTestFactory(t, "newznab", func(cfg IndexerConfig) (Indexer, error) {
		return &mockIndexer{}, nil // testErr is nil → Test() returns nil
	})

	db := testutil.OpenTestDB(t)
	s := NewService(db)
	ctx := context.Background()

	err := s.TestIndexer(ctx, IndexerConfig{Type: "newznab", BaseURL: "http://test.com"})
	require.NoError(t, err)
}

// ── DB error paths ─────────────────────────────────────────────────────────────

// openAndCloseSQLiteDB opens an in-memory SQLite database and immediately closes
// it so that any subsequent operation on it returns an error.
func openAndCloseSQLiteDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	require.NoError(t, db.Close())
	return db
}

// TestListIndexers_DBError covers the db.QueryContext error path in ListIndexers.
func TestListIndexers_DBError(t *testing.T) {
	db := openAndCloseSQLiteDB(t)
	s := NewService(db)
	_, err := s.ListIndexers(context.Background())
	require.Error(t, err)
}

// TestGetIndexer_DBError covers the db.QueryContext error path in GetIndexer
// (a closed DB produces an error that is not sql.ErrNoRows).
func TestGetIndexer_DBError(t *testing.T) {
	db := openAndCloseSQLiteDB(t)
	s := NewService(db)
	_, err := s.GetIndexer(context.Background(), 1)
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrIndexerNotFound)
}

// TestCreateIndexer_DBError covers the db.ExecContext error path in CreateIndexer.
func TestCreateIndexer_DBError(t *testing.T) {
	db := openAndCloseSQLiteDB(t)
	s := NewService(db)
	_, err := s.CreateIndexer(context.Background(), IndexerConfig{
		Name: "X", Type: "newznab", BaseURL: "http://x.com",
	})
	require.Error(t, err)
}

// TestDeleteIndexer_DBError covers the db.ExecContext error path in DeleteIndexer.
func TestDeleteIndexer_DBError(t *testing.T) {
	db := openAndCloseSQLiteDB(t)
	s := NewService(db)
	err := s.DeleteIndexer(context.Background(), 1)
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrIndexerNotFound)
}

