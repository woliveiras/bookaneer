package scheduler_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/scheduler"
	"github.com/woliveiras/bookaneer/internal/testutil"
)

func TestGetScheduledTasks_DefaultSeeds(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := scheduler.New(db, 3)

	// The migration seeds 4 default scheduled tasks (RssSync and MissingBookSearch removed).
	tasks, err := s.GetScheduledTasks(context.Background())
	require.NoError(t, err)
	assert.Len(t, tasks, 4)

	names := make([]string, len(tasks))
	for i, task := range tasks {
		names[i] = task.Name
	}
	assert.Contains(t, names, "LibraryScan")
	assert.Contains(t, names, "DownloadMonitor")
}

func TestGetScheduledTasks_AfterDelete(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := scheduler.New(db, 3)

	_, err := db.Exec(`DELETE FROM scheduled_tasks`)
	require.NoError(t, err)

	tasks, err := s.GetScheduledTasks(context.Background())
	require.NoError(t, err)
	assert.Empty(t, tasks)
}

func TestGetScheduledTasks_CustomTask(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := scheduler.New(db, 3)

	_, err := db.Exec(`DELETE FROM scheduled_tasks`)
	require.NoError(t, err)

	now := time.Now().UTC().Format(time.RFC3339)
	_, err = db.Exec(`INSERT INTO scheduled_tasks (name, interval_seconds, next_run_at) VALUES ('CustomTask', 3600, ?)`, now)
	require.NoError(t, err)

	tasks, err := s.GetScheduledTasks(context.Background())
	require.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, "CustomTask", tasks[0].Name)
	assert.Equal(t, 3600, tasks[0].IntervalSeconds)
}

func TestTriggerTask_QueuesCommand(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := scheduler.New(db, 3)
	ctx := context.Background()

	id, err := s.TriggerTask(ctx, "BookSearch")
	require.NoError(t, err)
	assert.NotEmpty(t, id)

	cmds, err := s.ListCommands(ctx, 10)
	require.NoError(t, err)
	assert.NotEmpty(t, cmds)
	assert.Equal(t, scheduler.CommandName("BookSearch"), cmds[0].Name)
	assert.Equal(t, scheduler.TriggerManual, cmds[0].Trigger)
}

func TestNew_DefaultConcurrency(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := scheduler.New(db, 0)
	assert.NotNil(t, s)
}

func TestNew_NegativeConcurrency(t *testing.T) {
	db := testutil.OpenTestDB(t)
	s := scheduler.New(db, -1)
	assert.NotNil(t, s)
}
