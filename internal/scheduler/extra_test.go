package scheduler

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/core/book"
	"github.com/woliveiras/bookaneer/internal/core/naming"
	"github.com/woliveiras/bookaneer/internal/download"
	"github.com/woliveiras/bookaneer/internal/library"
	"github.com/woliveiras/bookaneer/internal/search"
	"github.com/woliveiras/bookaneer/internal/testutil"
	"github.com/woliveiras/bookaneer/internal/wanted"
	_ "modernc.org/sqlite"
)

// closedDB returns an already-closed database so that all operations on it fail,
// enabling tests of error-handling paths without modifying testutil.OpenTestDB.
func closedDB() *sql.DB {
	db, _ := sql.Open("sqlite", ":memory:")
	_ = db.Close()
	return db
}

// newTestWantedSvc creates a minimal *wanted.Service backed by a fresh in-memory
// database. No digital-library providers or indexers are registered, so
// SearchAllWanted / ProcessDownloads succeed but find nothing.
func newTestWantedSvc(t *testing.T) (*wanted.Service, *sql.DB) {
	t.Helper()
	db := testutil.OpenTestDB(t)
	return wanted.New(
		db,
		book.New(db),
		library.NewAggregator(), // zero providers → searchDigitalLibraries returns (nil, nil)
		search.NewService(db),
		download.NewService(db),
		naming.New(db),
		nil, // scanner not needed for these tests
		nil, // pathMapper not needed for these tests
	), db
}

// ── QueueCommand ─────────────────────────────────────────────────────────────

// TestQueueCommand_DBError covers the db.ExecContext error path in QueueCommand.
func TestQueueCommand_DBError(t *testing.T) {
	s := New(closedDB(), 3)
	_, err := s.QueueCommand(context.Background(), CommandDownloadGrab, TriggerManual, nil)
	require.Error(t, err)
}

// ── ListCommands ──────────────────────────────────────────────────────────────

// TestListCommands_DBError covers the db.QueryContext error path in ListCommands.
func TestListCommands_DBError(t *testing.T) {
	s := New(closedDB(), 3)
	_, err := s.ListCommands(context.Background(), 10)
	require.Error(t, err)
}

// ── GetActiveCommands ─────────────────────────────────────────────────────────

// TestGetActiveCommands_DBError covers the error path in GetActiveCommands.
func TestGetActiveCommands_DBError(t *testing.T) {
	s := New(closedDB(), 3)
	_, err := s.GetActiveCommands(context.Background())
	require.Error(t, err)
}

// ── CancelCommand ─────────────────────────────────────────────────────────────

// TestCancelCommand_WithRunningCommand covers the cancel() call inside CancelCommand
// when the command ID is present in s.running (the running-command branch).
func TestCancelCommand_WithRunningCommand(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	id, err := s.QueueCommand(ctx, CommandDownloadGrab, TriggerManual, nil)
	require.NoError(t, err)

	cancelCalled := false
	s.mu.Lock()
	s.running[id] = func() { cancelCalled = true }
	s.mu.Unlock()

	require.NoError(t, s.CancelCommand(ctx, id))
	assert.True(t, cancelCalled, "cancel should have been called for the running command")
}

// ── Stop ──────────────────────────────────────────────────────────────────────

// TestStop_CancelsRunningCommands covers the loop body inside Stop that iterates
// s.running and calls each cancel function.
func TestStop_CancelsRunningCommands(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	cancelCalled := false
	s.mu.Lock()
	s.running["fake-cmd-id"] = func() { cancelCalled = true }
	s.mu.Unlock()

	s.Start(ctx)
	s.Stop()

	assert.True(t, cancelCalled, "Stop should cancel all running commands")
}

// ── run ───────────────────────────────────────────────────────────────────────

// TestRun_TickerFires covers the ticker.C branch inside run(), which calls
// dispatchPendingCommands and checkScheduledTasks.
func TestRun_TickerFires(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	s.Start(ctx)
	time.Sleep(1100 * time.Millisecond) // wait for at least one 1-second tick
	s.Stop()
	// No panic / deadlock reaching here means the ticker path executed successfully.
}

// ── recoverOrphanedCommands ───────────────────────────────────────────────────

// TestRecoverOrphanedCommands_DBError covers the db.ExecContext error path inside
// recoverOrphanedCommands (logged but does not propagate).
func TestRecoverOrphanedCommands_DBError(t *testing.T) {
	s := New(closedDB(), 3)
	// Must not panic; the error is logged and swallowed.
	s.recoverOrphanedCommands(context.Background())
}

// ── checkScheduledTasks ───────────────────────────────────────────────────────

// TestCheckScheduledTasks_DBError covers the getDueScheduledTasks error path inside
// checkScheduledTasks (logged but does not propagate).
func TestCheckScheduledTasks_DBError(t *testing.T) {
	s := New(closedDB(), 3)
	// Must not panic.
	s.checkScheduledTasks(context.Background())
}

// ── GetScheduledTasks ─────────────────────────────────────────────────────────

// TestGetScheduledTasks_DBError covers the db.QueryContext error path in
// GetScheduledTasks.
func TestGetScheduledTasks_DBError(t *testing.T) {
	s := New(closedDB(), 3)
	_, err := s.GetScheduledTasks(context.Background())
	require.Error(t, err)
}

// ── RegisterWantedHandlers – inner handler bodies ─────────────────────────────

// TestRegisterWantedHandlers_DownloadGrab_MissingBookId covers the bookId validation
// inside the CommandDownloadGrab handler (wantedService is never reached).
func TestRegisterWantedHandlers_DownloadGrab_MissingBookId(t *testing.T) {
	wantedSvc, db := newTestWantedSvc(t)
	s := New(db, 3)
	s.RegisterWantedHandlers(wantedSvc)

	s.mu.RLock()
	handler := s.handlers[CommandDownloadGrab]
	s.mu.RUnlock()

	cmd := &Command{ID: "test-grab", Payload: map[string]any{}}
	err := handler(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bookId")
}

// TestRegisterWantedHandlers_DownloadGrab_MissingURL covers the downloadUrl validation.
func TestRegisterWantedHandlers_DownloadGrab_MissingURL(t *testing.T) {
	wantedSvc, db := newTestWantedSvc(t)
	s := New(db, 3)
	s.RegisterWantedHandlers(wantedSvc)

	s.mu.RLock()
	handler := s.handlers[CommandDownloadGrab]
	s.mu.RUnlock()

	cmd := &Command{ID: "test-grab", Payload: map[string]any{"bookId": float64(1)}}
	err := handler(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "downloadUrl")
}

// TestRegisterWantedHandlers_DownloadGrab_EmptyURL covers the empty-string downloadUrl check.
func TestRegisterWantedHandlers_DownloadGrab_EmptyURL(t *testing.T) {
	wantedSvc, db := newTestWantedSvc(t)
	s := New(db, 3)
	s.RegisterWantedHandlers(wantedSvc)

	s.mu.RLock()
	handler := s.handlers[CommandDownloadGrab]
	s.mu.RUnlock()

	cmd := &Command{
		ID:      "test-grab",
		Payload: map[string]any{"bookId": float64(1), "downloadUrl": ""},
	}
	err := handler(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "downloadUrl")
}

// TestRegisterWantedHandlers_DownloadMonitor exercises the CommandDownloadMonitor
// handler body: ProcessDownloads with empty DB returns a zero result struct.
func TestRegisterWantedHandlers_DownloadMonitor(t *testing.T) {
	wantedSvc, db := newTestWantedSvc(t)
	s := New(db, 3)
	s.RegisterWantedHandlers(wantedSvc)

	s.mu.RLock()
	handler := s.handlers[CommandDownloadMonitor]
	s.mu.RUnlock()
	require.NotNil(t, handler)

	cmd := &Command{ID: "test-monitor", Payload: map[string]any{}}
	require.NoError(t, handler(context.Background(), cmd))
	assert.NotNil(t, cmd.Result)
	assert.Equal(t, 0, cmd.Result["checked"])
}

// TestRegisterWantedHandlers_DownloadGrab_GrabError covers lines 77-83 of handlers.go:
// when bookId and downloadUrl are valid, GrabRelease is called and returns an error
// because the book doesn't exist in the database.
func TestRegisterWantedHandlers_DownloadGrab_GrabError(t *testing.T) {
	wantedSvc, db := newTestWantedSvc(t)
	s := New(db, 3)
	s.RegisterWantedHandlers(wantedSvc)

	s.mu.RLock()
	handler := s.handlers[CommandDownloadGrab]
	s.mu.RUnlock()

	cmd := &Command{
		ID: "test-grab-err",
		Payload: map[string]any{
			"bookId":       float64(99999), // book doesn't exist → GrabRelease returns error
			"downloadUrl":  "http://test.com/book.epub",
			"releaseTitle": "Test Book",
			"size":         float64(1024),
		},
	}
	err := handler(context.Background(), cmd)
	require.Error(t, err, "GrabRelease should fail for a non-existent book")
}

// TestRegisterWantedHandlers_DownloadMonitor_Error covers the error-return branch in
// the CommandDownloadMonitor handler when ProcessDownloads returns an error.
func TestRegisterWantedHandlers_DownloadMonitor_Error(t *testing.T) {
	wantedSvc, db := newTestWantedSvc(t)
	s := New(db, 3)
	s.RegisterWantedHandlers(wantedSvc)

	s.mu.RLock()
	handler := s.handlers[CommandDownloadMonitor]
	s.mu.RUnlock()

	_ = db.Close() // close DB so ProcessDownloads fails
	cmd := &Command{ID: "test-monitor-err", Payload: map[string]any{}}
	err := handler(context.Background(), cmd)
	require.Error(t, err)
}

// TestCheckScheduledTasks_QueueCommandError covers the "Failed to queue scheduled task"
// error branch in checkScheduledTasks: a task is due but the commands table is gone so
// QueueCommand fails, triggering the slog.Error + continue path.
func TestCheckScheduledTasks_QueueCommandError(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	// Make one task immediately due.
	_, err := db.ExecContext(ctx, `UPDATE scheduled_tasks SET next_run_at = datetime('now', '-1 minute') WHERE name = 'DownloadMonitor'`)
	require.NoError(t, err)
	// Set all other tasks to the future to keep the test focused.
	_, err = db.ExecContext(ctx, `UPDATE scheduled_tasks SET next_run_at = datetime('now', '+1 year') WHERE name != 'DownloadMonitor'`)
	require.NoError(t, err)

	// Drop the commands table so QueueCommand fails while getDueScheduledTasks still works.
	_, err = db.ExecContext(ctx, `DROP TABLE commands`)
	require.NoError(t, err)

	// Must not panic; the error is logged and the loop continues.
	s.checkScheduledTasks(ctx)
}
