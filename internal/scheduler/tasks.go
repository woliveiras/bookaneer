package scheduler

import (
	"context"
	"database/sql"
	"log/slog"
	"time"
)

// GetScheduledTasks returns all scheduled tasks.
func (s *Scheduler) GetScheduledTasks(ctx context.Context) ([]ScheduledTask, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT name, interval_seconds, last_run_at, next_run_at
		FROM scheduled_tasks
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanTasks(rows)
}

// TriggerTask immediately queues a scheduled task.
func (s *Scheduler) TriggerTask(ctx context.Context, name string) (string, error) {
	return s.QueueCommand(ctx, CommandName(name), TriggerManual, nil)
}

// getDueScheduledTasks returns tasks that are due to run.
func (s *Scheduler) getDueScheduledTasks(ctx context.Context) ([]ScheduledTask, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	rows, err := s.db.QueryContext(ctx, `
		SELECT name, interval_seconds, last_run_at, next_run_at
		FROM scheduled_tasks
		WHERE next_run_at <= ?
	`, now)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanTasks(rows)
}

// updateTaskNextRun updates a task's next run time.
func (s *Scheduler) updateTaskNextRun(ctx context.Context, name string, intervalSeconds int) error {
	now := time.Now().UTC()
	nextRun := now.Add(time.Duration(intervalSeconds) * time.Second)

	_, err := s.db.ExecContext(ctx, `
		UPDATE scheduled_tasks
		SET last_run_at = ?, next_run_at = ?
		WHERE name = ?
	`, now.Format(time.RFC3339), nextRun.Format(time.RFC3339), name)

	return err
}

// checkScheduledTasks checks and queues due scheduled tasks.
func (s *Scheduler) checkScheduledTasks(ctx context.Context) {
	tasks, err := s.getDueScheduledTasks(ctx)
	if err != nil {
		slog.Error("Failed to get scheduled tasks", "error", err)
		return
	}

	for _, task := range tasks {
		_, err := s.QueueCommand(ctx, CommandName(task.Name), TriggerScheduled, nil)
		if err != nil {
			slog.Error("Failed to queue scheduled task", "task", task.Name, "error", err)
			continue
		}

		if err := s.updateTaskNextRun(ctx, task.Name, task.IntervalSeconds); err != nil {
			slog.Error("Failed to update task next run", "task", task.Name, "error", err)
		}
	}
}

// scanTasks scans multiple task rows.
func scanTasks(rows *sql.Rows) ([]ScheduledTask, error) {
	var tasks []ScheduledTask
	for rows.Next() {
		var task ScheduledTask
		var lastRunAt, nextRunAt sql.NullString

		if err := rows.Scan(&task.Name, &task.IntervalSeconds, &lastRunAt, &nextRunAt); err != nil {
			return nil, err
		}

		if lastRunAt.Valid {
			t, _ := time.Parse(time.RFC3339, lastRunAt.String)
			task.LastRunAt = &t
		}
		if nextRunAt.Valid {
			task.NextRunAt, _ = time.Parse(time.RFC3339, nextRunAt.String)
		}

		tasks = append(tasks, task)
	}

	return tasks, rows.Err()
}
