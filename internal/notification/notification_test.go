package notification

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestService_CRUD(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := New(db)
	ctx := context.Background()

	t.Run("list empty", func(t *testing.T) {
		configs, err := svc.List(ctx)
		require.NoError(t, err)
		assert.Empty(t, configs)
	})

	t.Run("create", func(t *testing.T) {
		settings := json.RawMessage(`{"url":"https://example.com/hook"}`)
		cfg, err := svc.Create(ctx, CreateInput{
			Name:       "Test Webhook",
			Type:       "webhook",
			Settings:   settings,
			OnGrab:     true,
			OnDownload: true,
			OnUpgrade:  false,
			Enabled:    true,
		})
		require.NoError(t, err)
		assert.Equal(t, "Test Webhook", cfg.Name)
		assert.Equal(t, "webhook", cfg.Type)
		assert.True(t, cfg.OnGrab)
		assert.True(t, cfg.OnDownload)
		assert.False(t, cfg.OnUpgrade)
		assert.True(t, cfg.Enabled)
		assert.Greater(t, cfg.ID, int64(0))
	})

	t.Run("find by id", func(t *testing.T) {
		settings := json.RawMessage(`{"url":"https://example.com/hook2"}`)
		created, err := svc.Create(ctx, CreateInput{
			Name: "Find Me", Type: "webhook", Settings: settings, Enabled: true,
		})
		require.NoError(t, err)

		found, err := svc.FindByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		assert.Equal(t, "Find Me", found.Name)
	})

	t.Run("find by id not found", func(t *testing.T) {
		_, err := svc.FindByID(ctx, 99999)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("update", func(t *testing.T) {
		created, err := svc.Create(ctx, CreateInput{
			Name: "Update Me", Type: "webhook", Enabled: true,
		})
		require.NoError(t, err)

		newName := "Updated"
		enabled := false
		updated, err := svc.Update(ctx, created.ID, UpdateInput{
			Name:    &newName,
			Enabled: &enabled,
		})
		require.NoError(t, err)
		assert.Equal(t, "Updated", updated.Name)
		assert.False(t, updated.Enabled)
		assert.Equal(t, "webhook", updated.Type) // unchanged
	})

	t.Run("update not found", func(t *testing.T) {
		newName := "test"
		_, err := svc.Update(ctx, 99999, UpdateInput{Name: &newName})
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("delete", func(t *testing.T) {
		created, err := svc.Create(ctx, CreateInput{
			Name: "Delete Me", Type: "webhook", Enabled: true,
		})
		require.NoError(t, err)

		err = svc.Delete(ctx, created.ID)
		require.NoError(t, err)

		_, err = svc.FindByID(ctx, created.ID)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("delete not found", func(t *testing.T) {
		err := svc.Delete(ctx, 99999)
		assert.ErrorIs(t, err, ErrNotFound)
	})

	t.Run("list returns all", func(t *testing.T) {
		configs, err := svc.List(ctx)
		require.NoError(t, err)
		assert.NotEmpty(t, configs)
	})
}

func TestService_Dispatch(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := New(db)
	ctx := context.Background()

	var sentEvents []Event
	svc.RegisterFactory("mock", func(cfg Config) (Channel, error) {
		return &mockChannel{sendFn: func(e Event) { sentEvents = append(sentEvents, e) }}, nil
	})

	_, err := svc.Create(ctx, CreateInput{
		Name: "Mock", Type: "mock", OnGrab: true, OnDownload: false, Enabled: true,
	})
	require.NoError(t, err)

	t.Run("dispatch matching event", func(t *testing.T) {
		sentEvents = nil
		svc.Dispatch(ctx, Event{Type: EventGrab, Title: "Book grabbed"})
		assert.Len(t, sentEvents, 1)
		assert.Equal(t, EventGrab, sentEvents[0].Type)
	})

	t.Run("dispatch non-matching event", func(t *testing.T) {
		sentEvents = nil
		svc.Dispatch(ctx, Event{Type: EventDownload, Title: "Book downloaded"})
		assert.Empty(t, sentEvents)
	})

	t.Run("dispatch test event always sent", func(t *testing.T) {
		sentEvents = nil
		svc.Dispatch(ctx, Event{Type: EventTest, Title: "Test"})
		assert.Len(t, sentEvents, 1)
	})
}

func TestService_ShouldSend(t *testing.T) {
	svc := &Service{}

	tests := []struct {
		name      string
		cfg       Config
		eventType EventType
		want      bool
	}{
		{"grab enabled", Config{OnGrab: true}, EventGrab, true},
		{"grab disabled", Config{OnGrab: false}, EventGrab, false},
		{"download enabled", Config{OnDownload: true}, EventDownload, true},
		{"download disabled", Config{OnDownload: false}, EventDownload, false},
		{"import uses on_download", Config{OnDownload: true}, EventImport, true},
		{"upgrade enabled", Config{OnUpgrade: true}, EventUpgrade, true},
		{"upgrade disabled", Config{OnUpgrade: false}, EventUpgrade, false},
		{"test always true", Config{}, EventTest, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.shouldSend(tt.cfg, tt.eventType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestService_TestChannel(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := New(db)
	ctx := context.Background()

	var tested bool
	svc.RegisterFactory("mock", func(cfg Config) (Channel, error) {
		return &mockChannel{testFn: func() error { tested = true; return nil }}, nil
	})

	created, err := svc.Create(ctx, CreateInput{Name: "TestMe", Type: "mock", Enabled: true})
	require.NoError(t, err)

	err = svc.TestChannel(ctx, created.ID)
	require.NoError(t, err)
	assert.True(t, tested)
}

func TestService_TestChannel_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := New(db)
	err := svc.TestChannel(context.Background(), 99999)
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestService_TestChannel_UnsupportedType(t *testing.T) {
	db := testutil.OpenTestDB(t)
	svc := New(db)
	ctx := context.Background()

	created, err := svc.Create(ctx, CreateInput{Name: "Bad", Type: "nonexistent", Enabled: true})
	require.NoError(t, err)

	err = svc.TestChannel(ctx, created.ID)
	assert.ErrorIs(t, err, ErrUnsupported)
}

func TestBoolToInt(t *testing.T) {
	assert.Equal(t, 1, boolToInt(true))
	assert.Equal(t, 0, boolToInt(false))
}

// mockChannel is a test double for the Channel interface.
type mockChannel struct {
	testFn func() error
	sendFn func(Event)
}

func (m *mockChannel) Name() string { return "mock" }
func (m *mockChannel) Type() string { return "mock" }

func (m *mockChannel) Test(ctx context.Context) error {
	if m.testFn != nil {
		return m.testFn()
	}
	return nil
}

func (m *mockChannel) Send(ctx context.Context, event Event) error {
	if m.sendFn != nil {
		m.sendFn(event)
	}
	return nil
}
