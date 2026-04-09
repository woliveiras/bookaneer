package scheduler

import (
	"context"
	"testing"

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
