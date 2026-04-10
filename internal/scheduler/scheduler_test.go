package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestQueueCommand(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	id, err := s.QueueCommand(ctx, CommandBookSearch, TriggerManual, map[string]any{"bookId": 42})
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	cmd, err := s.GetCommand(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, id, cmd.ID)
	assert.Equal(t, CommandBookSearch, cmd.Name)
	assert.Equal(t, StatusQueued, cmd.Status)
	assert.Equal(t, TriggerManual, cmd.Trigger)
	assert.Equal(t, float64(42), cmd.Payload["bookId"])
}

func TestQueueCommand_NilPayload(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	id, err := s.QueueCommand(ctx, CommandRssSync, TriggerScheduled, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	cmd, err := s.GetCommand(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, CommandRssSync, cmd.Name)
	assert.Equal(t, TriggerScheduled, cmd.Trigger)
}

func TestGetCommand_NotFound(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	cmd, err := s.GetCommand(ctx, "nonexistent-id")
	assert.Error(t, err)
	assert.Nil(t, cmd)
}

func TestListCommands(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		_, err := s.QueueCommand(ctx, CommandBookSearch, TriggerManual, nil)
		require.NoError(t, err)
	}

	cmds, err := s.ListCommands(ctx, 3)
	require.NoError(t, err)
	assert.Len(t, cmds, 3)
}

func TestListCommands_DefaultLimit(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	_, err := s.QueueCommand(ctx, CommandBookSearch, TriggerManual, nil)
	require.NoError(t, err)

	cmds, err := s.ListCommands(ctx, 0)
	require.NoError(t, err)
	assert.Len(t, cmds, 1)
}

func TestGetActiveCommands(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	id, err := s.QueueCommand(ctx, CommandBookSearch, TriggerManual, nil)
	require.NoError(t, err)

	active, err := s.GetActiveCommands(ctx)
	require.NoError(t, err)
	assert.Len(t, active, 1)
	assert.Equal(t, id, active[0].ID)
	assert.Equal(t, StatusQueued, active[0].Status)
}

func TestGetActiveCommands_ExcludesCompleted(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	id, err := s.QueueCommand(ctx, CommandBookSearch, TriggerManual, nil)
	require.NoError(t, err)

	err = s.updateCommandStatus(ctx, id, StatusCompleted, nil)
	require.NoError(t, err)

	active, err := s.GetActiveCommands(ctx)
	require.NoError(t, err)
	assert.Empty(t, active)
}

func TestCancelCommand(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	id, err := s.QueueCommand(ctx, CommandBookSearch, TriggerManual, nil)
	require.NoError(t, err)

	err = s.CancelCommand(ctx, id)
	require.NoError(t, err)

	cmd, err := s.GetCommand(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, StatusCancelled, cmd.Status)
	assert.NotNil(t, cmd.EndedAt)
}

func TestUpdateCommandStatus_Running(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	id, err := s.QueueCommand(ctx, CommandBookSearch, TriggerManual, nil)
	require.NoError(t, err)

	err = s.updateCommandStatus(ctx, id, StatusRunning, nil)
	require.NoError(t, err)

	cmd, err := s.GetCommand(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, StatusRunning, cmd.Status)
	assert.NotNil(t, cmd.StartedAt)
	assert.Nil(t, cmd.EndedAt)
}

func TestUpdateCommandStatus_Failed(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	id, err := s.QueueCommand(ctx, CommandBookSearch, TriggerManual, nil)
	require.NoError(t, err)

	err = s.updateCommandStatus(ctx, id, StatusFailed, map[string]any{"error": "timeout"})
	require.NoError(t, err)

	cmd, err := s.GetCommand(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, cmd.Status)
	assert.NotNil(t, cmd.EndedAt)
	assert.Equal(t, "timeout", cmd.Result["error"])
}

func TestRecoverOrphanedCommands(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	id, err := s.QueueCommand(ctx, CommandBookSearch, TriggerManual, nil)
	require.NoError(t, err)
	err = s.updateCommandStatus(ctx, id, StatusRunning, nil)
	require.NoError(t, err)

	s.recoverOrphanedCommands(ctx)

	cmd, err := s.GetCommand(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, StatusQueued, cmd.Status, "orphaned running command should be re-queued")
}

func TestGetNextQueuedCommand_PriorityOrder(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	_, err := s.QueueCommand(ctx, CommandRssSync, TriggerScheduled, nil)
	require.NoError(t, err)
	id2, err := s.QueueCommand(ctx, CommandBookSearch, TriggerManual, nil)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx, "UPDATE commands SET priority = 10 WHERE id = ?", id2)
	require.NoError(t, err)

	cmd, err := s.getNextQueuedCommand(ctx)
	require.NoError(t, err)
	assert.Equal(t, id2, cmd.ID, "higher priority command should be dispatched first")
}

func TestRegisterHandler(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)

	called := false
	s.RegisterHandler(CommandBookSearch, func(_ context.Context, _ *Command) error {
		called = true
		return nil
	})

	s.mu.RLock()
	_, ok := s.handlers[CommandBookSearch]
	s.mu.RUnlock()

	assert.True(t, ok, "handler should be registered")
	assert.False(t, called, "handler should not be called yet")
}

func TestTriggerTask(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	id, err := s.TriggerTask(ctx, "RssSync")
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	cmd, err := s.GetCommand(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, CommandName("RssSync"), cmd.Name)
	assert.Equal(t, TriggerManual, cmd.Trigger)
}

func TestStartStop(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	s.Start(ctx)
	s.Stop()
}

func TestStartStop_Graceful(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx, cancel := context.WithCancel(context.Background())

	s.Start(ctx)
	// Cancel context to trigger the ctx.Done() path in run()
	cancel()
	s.Stop()
}

func TestExecuteCommand_NoHandler(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	id, err := s.QueueCommand(ctx, "UnknownCommand", TriggerManual, nil)
	require.NoError(t, err)

	cmd, err := s.GetCommand(ctx, id)
	require.NoError(t, err)

	s.workerSem <- struct{}{}
	s.wg.Add(1)
	s.executeCommand(ctx, cmd)

	updated, err := s.GetCommand(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, updated.Status)
}

func TestExecuteCommand_HandlerFails(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	s.RegisterHandler(CommandBookSearch, func(_ context.Context, _ *Command) error {
		return assert.AnError
	})

	id, err := s.QueueCommand(ctx, CommandBookSearch, TriggerManual, nil)
	require.NoError(t, err)

	cmd, err := s.GetCommand(ctx, id)
	require.NoError(t, err)

	s.workerSem <- struct{}{}
	s.wg.Add(1)
	s.executeCommand(ctx, cmd)

	updated, err := s.GetCommand(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, StatusFailed, updated.Status)
}

func TestExecuteCommand_HandlerSucceeds(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	s.RegisterHandler(CommandBookSearch, func(_ context.Context, cmd *Command) error {
		cmd.Result = map[string]any{"found": 5}
		return nil
	})

	id, err := s.QueueCommand(ctx, CommandBookSearch, TriggerManual, nil)
	require.NoError(t, err)

	cmd, err := s.GetCommand(ctx, id)
	require.NoError(t, err)

	s.workerSem <- struct{}{}
	s.wg.Add(1)
	s.executeCommand(ctx, cmd)

	updated, err := s.GetCommand(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, updated.Status)
}

func TestDispatchPendingCommands_NoQueued(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	// Should not panic with empty queue
	s.dispatchPendingCommands(context.Background())
}

func TestCheckScheduledTasks_RunsDueTasks(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	// Set DownloadMonitor to be due (its default next_run is "now")
	s.checkScheduledTasks(ctx)

	// Should have queued at least the DownloadMonitor task
	cmds, err := s.ListCommands(ctx, 50)
	require.NoError(t, err)
	assert.NotEmpty(t, cmds)
}

func TestRecoverOrphanedCommands_NoOrphans(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	// Should not panic with no orphaned commands
	s.recoverOrphanedCommands(ctx)
}

func TestRecoverOrphanedCommands_WithOrphans(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	id, err := s.QueueCommand(ctx, "TestCmd", "manual", nil)
	require.NoError(t, err)

	// Set it to "running" manually
	err = s.updateCommandStatus(ctx, id, StatusRunning, nil)
	require.NoError(t, err)

	s.recoverOrphanedCommands(ctx)

	cmd, err := s.GetCommand(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, StatusQueued, cmd.Status)
}

func TestUpdateCommandStatus_Default(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	id, err := s.QueueCommand(ctx, "TestCmd", "manual", nil)
	require.NoError(t, err)

	// Use a status not in the switch cases to hit default
	err = s.updateCommandStatus(ctx, id, StatusQueued, nil)
	require.NoError(t, err)

	cmd, err := s.GetCommand(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, StatusQueued, cmd.Status)
}

func TestListCommands_WithPayloadAndResult(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	payload := map[string]any{"bookId": float64(42)}
	id, err := s.QueueCommand(ctx, "BookSearch", "manual", payload)
	require.NoError(t, err)

	// Complete with result
	err = s.updateCommandStatus(ctx, id, StatusCompleted, map[string]any{"found": float64(3)})
	require.NoError(t, err)

	cmds, err := s.ListCommands(ctx, 10)
	require.NoError(t, err)
	require.NotEmpty(t, cmds)

	var found bool
	for _, cmd := range cmds {
		if cmd.ID == id {
			found = true
			assert.NotNil(t, cmd.EndedAt)
			break
		}
	}
	assert.True(t, found)
}

func TestCheckScheduledTasks_NoDueTasks(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	// Set all tasks to far future
	_, err := db.ExecContext(ctx, `UPDATE scheduled_tasks SET next_run_at = datetime('now', '+1 year')`)
	require.NoError(t, err)

	s.checkScheduledTasks(ctx)

	// Should NOT have queued any commands (except from earlier)
	cmds, err := s.ListCommands(ctx, 50)
	require.NoError(t, err)
	assert.Empty(t, cmds)
}

// TestMakeBookSearchHandler_MissingBookId tests the handler's validation path when the
// bookId key is absent from the payload (covers 0% makeBookSearchHandler lines).
func TestMakeBookSearchHandler_MissingBookId(t *testing.T) {
	handler := makeBookSearchHandler(nil) // nil is safe: SearchAndGrab is never reached
	cmd := &Command{
		ID:      "test-missing",
		Payload: map[string]any{},
	}
	err := handler(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bookId")
}

// TestMakeBookSearchHandler_InvalidBookIdType covers the type-assertion failure path.
func TestMakeBookSearchHandler_InvalidBookIdType(t *testing.T) {
	handler := makeBookSearchHandler(nil)
	cmd := &Command{
		ID:      "test-invalid-type",
		Payload: map[string]any{"bookId": "not-a-number"},
	}
	err := handler(context.Background(), cmd)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bookId")
}

// TestRegisterWantedHandlers_RegistersAllHandlers verifies that all six command handlers
// are registered after calling RegisterWantedHandlers (covers 0% RegisterWantedHandlers).
// A nil *wanted.Service is safe here because the handlers are only stored in closures
// during registration – they are never invoked.
func TestRegisterWantedHandlers_RegistersAllHandlers(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)

	s.RegisterWantedHandlers(nil)

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, name := range []CommandName{
		CommandBookSearch,
		CommandAutomaticSearch,
		CommandMissingBookSearch,
		CommandDownloadGrab,
		CommandRssSync,
		CommandDownloadMonitor,
	} {
		assert.NotNil(t, s.handlers[name], "handler for %s should be registered", name)
	}
}

// TestCancelCommand_NonExistent verifies that cancelling a non-existent command
// does not return an error (CancelCommand doesn't check rows affected).
func TestCancelCommand_NonExistent(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	err := s.CancelCommand(ctx, "non-existent-id-xyz")
	require.NoError(t, err)
}

// TestCancelCommand_AlreadyCompleted verifies that cancelling a completed command
// simply overwrites its status without error.
func TestCancelCommand_AlreadyCompleted(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	id, err := s.QueueCommand(ctx, CommandBookSearch, TriggerManual, nil)
	require.NoError(t, err)
	require.NoError(t, s.updateCommandStatus(ctx, id, StatusCompleted, nil))

	err = s.CancelCommand(ctx, id)
	require.NoError(t, err)
}

// TestDispatchPendingCommands_WithQueuedCommand verifies that dispatchPendingCommands
// picks up a queued command, executes it via a goroutine, and marks it completed
// (covers the "worker slot available" branch of dispatchPendingCommands).
func TestDispatchPendingCommands_WithQueuedCommand(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	done := make(chan struct{})
	s.RegisterHandler(CommandRssSync, func(_ context.Context, _ *Command) error {
		close(done)
		return nil
	})

	id, err := s.QueueCommand(ctx, CommandRssSync, TriggerManual, nil)
	require.NoError(t, err)

	s.dispatchPendingCommands(ctx)

	select {
	case <-done:
	case <-context.Background().Done():
		t.Fatal("handler was not invoked")
	}
	s.wg.Wait()

	cmd, err := s.GetCommand(ctx, id)
	require.NoError(t, err)
	assert.Equal(t, StatusCompleted, cmd.Status)
}

// TestDispatchPendingCommands_WorkersFull verifies that when all worker slots are occupied,
// dispatchPendingCommands leaves the command queued (covers the "default" select branch).
func TestDispatchPendingCommands_WorkersFull(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 1) // single worker slot
	ctx := context.Background()

	// Occupy the only worker slot so no new work can be dispatched.
	s.workerSem <- struct{}{}

	_, err := s.QueueCommand(ctx, CommandBookSearch, TriggerManual, nil)
	require.NoError(t, err)

	s.dispatchPendingCommands(ctx)

	// Release slot before asserting so the test DB is still usable.
	<-s.workerSem

	cmds, err := s.GetActiveCommands(ctx)
	require.NoError(t, err)
	require.Len(t, cmds, 1)
	assert.Equal(t, StatusQueued, cmds[0].Status)
}

// TestCheckScheduledTasks_UpdatesNextRunTime verifies that after dispatching a due task,
// its next_run_at is pushed into the future (covers updateTaskNextRun call path).
func TestCheckScheduledTasks_UpdatesNextRunTime(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := New(db, 3)
	ctx := context.Background()

	// Force one task to be due right now.
	_, err := db.ExecContext(ctx,
		`UPDATE scheduled_tasks SET next_run_at = datetime('now', '-1 minute') WHERE name = 'RssSync'`)
	require.NoError(t, err)

	// Set all other tasks far in the future to keep the test focused.
	_, err = db.ExecContext(ctx,
		`UPDATE scheduled_tasks SET next_run_at = datetime('now', '+1 year') WHERE name != 'RssSync'`)
	require.NoError(t, err)

	s.checkScheduledTasks(ctx)

	// next_run_at should now be in the future.
	var nextRunAt string
	require.NoError(t, db.QueryRowContext(ctx,
		`SELECT next_run_at FROM scheduled_tasks WHERE name = 'RssSync'`).Scan(&nextRunAt))

	assert.NotEmpty(t, nextRunAt)
	assert.Greater(t, nextRunAt, time.Now().UTC().Add(-5*time.Second).Format(time.RFC3339))
}
