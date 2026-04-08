// Package scheduler provides a job scheduler and command queue processor.
// It reads commands from the database and executes them as goroutines,
// and runs scheduled tasks at configured intervals.
package scheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CommandName represents the type of command.
type CommandName string

const (
	// Book-related commands.
	CommandBookSearch        CommandName = "BookSearch"
	CommandMissingBookSearch CommandName = "MissingBookSearch"
	CommandRefreshBook       CommandName = "RefreshBook"
	CommandAutomaticSearch   CommandName = "AutomaticSearch"

	// Library commands.
	CommandLibraryScan  CommandName = "LibraryScan"
	CommandRenameFiles  CommandName = "RenameFiles"
	CommandManualImport CommandName = "ManualImport"
	CommandDeleteFiles  CommandName = "DeleteFiles"

	// Download commands.
	CommandDownloadGrab    CommandName = "DownloadGrab"
	CommandDownloadMonitor CommandName = "DownloadMonitor"

	// System commands.
	CommandRssSync      CommandName = "RssSync"
	CommandBackup       CommandName = "Backup"
	CommandHousekeeping CommandName = "Housekeeping"
)

// CommandStatus represents the status of a command.
type CommandStatus string

const (
	StatusQueued    CommandStatus = "queued"
	StatusRunning   CommandStatus = "running"
	StatusCompleted CommandStatus = "completed"
	StatusFailed    CommandStatus = "failed"
	StatusCancelled CommandStatus = "cancelled"
)

// CommandTrigger represents what triggered the command.
type CommandTrigger string

const (
	TriggerManual    CommandTrigger = "manual"
	TriggerScheduled CommandTrigger = "scheduled"
	TriggerAutomatic CommandTrigger = "automatic"
)

// Command represents a queued command.
type Command struct {
	ID        string         `json:"id"`
	Name      CommandName    `json:"name"`
	Status    CommandStatus  `json:"status"`
	Priority  int            `json:"priority"`
	Payload   map[string]any `json:"payload"`
	Result    map[string]any `json:"result"`
	Trigger   CommandTrigger `json:"trigger"`
	QueuedAt  time.Time      `json:"queuedAt"`
	StartedAt *time.Time     `json:"startedAt,omitempty"`
	EndedAt   *time.Time     `json:"endedAt,omitempty"`
}

// ScheduledTask represents a recurring task.
type ScheduledTask struct {
	Name            string     `json:"name"`
	IntervalSeconds int        `json:"intervalSeconds"`
	LastRunAt       *time.Time `json:"lastRunAt,omitempty"`
	NextRunAt       time.Time  `json:"nextRunAt"`
}

// CommandHandler is a function that executes a command.
type CommandHandler func(ctx context.Context, cmd *Command) error

// Scheduler manages the command queue and scheduled tasks.
type Scheduler struct {
	db       *sql.DB
	handlers map[CommandName]CommandHandler

	mu        sync.RWMutex
	running   map[string]context.CancelFunc // command ID -> cancel func
	workerSem chan struct{}                 // limit concurrent workers

	stopCh chan struct{}
	wg     sync.WaitGroup
}

// New creates a new Scheduler.
func New(db *sql.DB, maxConcurrent int) *Scheduler {
	if maxConcurrent <= 0 {
		maxConcurrent = 3
	}
	return &Scheduler{
		db:        db,
		handlers:  make(map[CommandName]CommandHandler),
		running:   make(map[string]context.CancelFunc),
		workerSem: make(chan struct{}, maxConcurrent),
		stopCh:    make(chan struct{}),
	}
}

// RegisterHandler registers a handler for a command type.
func (s *Scheduler) RegisterHandler(name CommandName, handler CommandHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[name] = handler
}

// Start starts the scheduler.
func (s *Scheduler) Start(ctx context.Context) {
	slog.Info("Starting scheduler")

	// Recover any orphaned running commands from crash
	s.recoverOrphanedCommands(ctx)

	// Start the main loop
	s.wg.Add(1)
	go s.run(ctx)
}

// Stop stops the scheduler gracefully.
func (s *Scheduler) Stop() {
	slog.Info("Stopping scheduler")
	close(s.stopCh)

	// Cancel all running commands
	s.mu.Lock()
	for id, cancel := range s.running {
		slog.Info("Cancelling command", "id", id)
		cancel()
	}
	s.mu.Unlock()

	// Wait for all workers to finish
	s.wg.Wait()
	slog.Info("Scheduler stopped")
}

// run is the main scheduler loop.
func (s *Scheduler) run(ctx context.Context) {
	defer s.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.dispatchPendingCommands(ctx)
			s.checkScheduledTasks(ctx)
		}
	}
}

// dispatchPendingCommands fetches and dispatches queued commands.
func (s *Scheduler) dispatchPendingCommands(ctx context.Context) {
	// Get next queued command
	cmd, err := s.getNextQueuedCommand(ctx)
	if err != nil {
		if err != sql.ErrNoRows {
			slog.Error("Failed to get next command", "error", err)
		}
		return
	}

	// Try to acquire worker slot
	select {
	case s.workerSem <- struct{}{}:
		// Got a slot, dispatch the command
		s.wg.Add(1)
		go s.executeCommand(ctx, cmd)
	default:
		// No worker slot available, try again later
	}
}

// checkScheduledTasks checks and queues due scheduled tasks.
func (s *Scheduler) checkScheduledTasks(ctx context.Context) {
	tasks, err := s.getDueScheduledTasks(ctx)
	if err != nil {
		slog.Error("Failed to get scheduled tasks", "error", err)
		return
	}

	for _, task := range tasks {
		// Queue the task as a command
		_, err := s.QueueCommand(ctx, CommandName(task.Name), TriggerScheduled, nil)
		if err != nil {
			slog.Error("Failed to queue scheduled task", "task", task.Name, "error", err)
			continue
		}

		// Update next run time
		if err := s.updateTaskNextRun(ctx, task.Name, task.IntervalSeconds); err != nil {
			slog.Error("Failed to update task next run", "task", task.Name, "error", err)
		}
	}
}

// executeCommand executes a command in a goroutine.
func (s *Scheduler) executeCommand(ctx context.Context, cmd *Command) {
	defer s.wg.Done()
	defer func() { <-s.workerSem }() // Release worker slot

	// Create cancellable context
	cmdCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Track running command
	s.mu.Lock()
	s.running[cmd.ID] = cancel
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.running, cmd.ID)
		s.mu.Unlock()
	}()

	// Mark as running
	if err := s.updateCommandStatus(ctx, cmd.ID, StatusRunning, nil); err != nil {
		slog.Error("Failed to mark command as running", "id", cmd.ID, "error", err)
		return
	}

	slog.Info("Executing command", "id", cmd.ID, "name", cmd.Name)

	// Get handler
	s.mu.RLock()
	handler, ok := s.handlers[cmd.Name]
	s.mu.RUnlock()

	if !ok {
		err := fmt.Errorf("no handler registered for command: %s", cmd.Name)
		s.updateCommandStatus(ctx, cmd.ID, StatusFailed, map[string]any{"error": err.Error()})
		slog.Error("No handler for command", "name", cmd.Name)
		return
	}

	// Execute handler
	err := handler(cmdCtx, cmd)

	// Update status based on result
	if err != nil {
		slog.Error("Command failed", "id", cmd.ID, "name", cmd.Name, "error", err)
		s.updateCommandStatus(ctx, cmd.ID, StatusFailed, map[string]any{"error": err.Error()})
	} else {
		slog.Info("Command completed", "id", cmd.ID, "name", cmd.Name)
		s.updateCommandStatus(ctx, cmd.ID, StatusCompleted, cmd.Result)
	}
}

// QueueCommand adds a new command to the queue.
func (s *Scheduler) QueueCommand(ctx context.Context, name CommandName, trigger CommandTrigger, payload map[string]any) (string, error) {
	id := uuid.New().String()

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal payload: %w", err)
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO commands (id, name, status, priority, payload, trigger, queued_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, id, name, StatusQueued, 0, string(payloadJSON), trigger, time.Now().UTC().Format(time.RFC3339))

	if err != nil {
		return "", fmt.Errorf("insert command: %w", err)
	}

	slog.Info("Command queued", "id", id, "name", name, "trigger", trigger)
	return id, nil
}

// GetCommand retrieves a command by ID.
func (s *Scheduler) GetCommand(ctx context.Context, id string) (*Command, error) {
	var cmd Command
	var payloadJSON, resultJSON string
	var startedAt, endedAt sql.NullString
	var queuedAt string

	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, status, priority, payload, result, trigger, queued_at, started_at, ended_at
		FROM commands WHERE id = ?
	`, id).Scan(&cmd.ID, &cmd.Name, &cmd.Status, &cmd.Priority, &payloadJSON, &resultJSON, &cmd.Trigger, &queuedAt, &startedAt, &endedAt)

	if err != nil {
		return nil, err
	}

	cmd.QueuedAt, _ = time.Parse(time.RFC3339, queuedAt)
	if startedAt.Valid {
		t, _ := time.Parse(time.RFC3339, startedAt.String)
		cmd.StartedAt = &t
	}
	if endedAt.Valid {
		t, _ := time.Parse(time.RFC3339, endedAt.String)
		cmd.EndedAt = &t
	}

	json.Unmarshal([]byte(payloadJSON), &cmd.Payload)
	json.Unmarshal([]byte(resultJSON), &cmd.Result)

	return &cmd, nil
}

// ListCommands lists recent commands.
func (s *Scheduler) ListCommands(ctx context.Context, limit int) ([]Command, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, status, priority, payload, result, trigger, queued_at, started_at, ended_at
		FROM commands
		ORDER BY queued_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []Command
	for rows.Next() {
		var cmd Command
		var payloadJSON, resultJSON string
		var startedAt, endedAt sql.NullString
		var queuedAt string

		if err := rows.Scan(&cmd.ID, &cmd.Name, &cmd.Status, &cmd.Priority, &payloadJSON, &resultJSON, &cmd.Trigger, &queuedAt, &startedAt, &endedAt); err != nil {
			return nil, err
		}

		cmd.QueuedAt, _ = time.Parse(time.RFC3339, queuedAt)
		if startedAt.Valid {
			t, _ := time.Parse(time.RFC3339, startedAt.String)
			cmd.StartedAt = &t
		}
		if endedAt.Valid {
			t, _ := time.Parse(time.RFC3339, endedAt.String)
			cmd.EndedAt = &t
		}

		json.Unmarshal([]byte(payloadJSON), &cmd.Payload)
		json.Unmarshal([]byte(resultJSON), &cmd.Result)

		commands = append(commands, cmd)
	}

	return commands, rows.Err()
}

// GetActiveCommands returns commands that are queued or running.
func (s *Scheduler) GetActiveCommands(ctx context.Context) ([]Command, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, status, priority, payload, result, trigger, queued_at, started_at, ended_at
		FROM commands
		WHERE status IN (?, ?)
		ORDER BY queued_at DESC
	`, StatusQueued, StatusRunning)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commands []Command
	for rows.Next() {
		var cmd Command
		var payloadJSON, resultJSON string
		var startedAt, endedAt sql.NullString
		var queuedAt string

		if err := rows.Scan(&cmd.ID, &cmd.Name, &cmd.Status, &cmd.Priority, &payloadJSON, &resultJSON, &cmd.Trigger, &queuedAt, &startedAt, &endedAt); err != nil {
			return nil, err
		}

		cmd.QueuedAt, _ = time.Parse(time.RFC3339, queuedAt)
		if startedAt.Valid {
			t, _ := time.Parse(time.RFC3339, startedAt.String)
			cmd.StartedAt = &t
		}
		if endedAt.Valid {
			t, _ := time.Parse(time.RFC3339, endedAt.String)
			cmd.EndedAt = &t
		}

		json.Unmarshal([]byte(payloadJSON), &cmd.Payload)
		json.Unmarshal([]byte(resultJSON), &cmd.Result)

		commands = append(commands, cmd)
	}

	return commands, rows.Err()
}

// CancelCommand cancels a running or queued command.
func (s *Scheduler) CancelCommand(ctx context.Context, id string) error {
	// If running, cancel it
	s.mu.Lock()
	if cancel, ok := s.running[id]; ok {
		cancel()
	}
	s.mu.Unlock()

	// Update status
	return s.updateCommandStatus(ctx, id, StatusCancelled, nil)
}

// getNextQueuedCommand gets the next command to execute.
func (s *Scheduler) getNextQueuedCommand(ctx context.Context) (*Command, error) {
	var cmd Command
	var payloadJSON string
	var queuedAt string

	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, status, priority, payload, trigger, queued_at
		FROM commands
		WHERE status = ?
		ORDER BY priority DESC, queued_at ASC
		LIMIT 1
	`, StatusQueued).Scan(&cmd.ID, &cmd.Name, &cmd.Status, &cmd.Priority, &payloadJSON, &cmd.Trigger, &queuedAt)

	if err != nil {
		return nil, err
	}

	cmd.QueuedAt, _ = time.Parse(time.RFC3339, queuedAt)
	json.Unmarshal([]byte(payloadJSON), &cmd.Payload)

	return &cmd, nil
}

// updateCommandStatus updates a command's status.
func (s *Scheduler) updateCommandStatus(ctx context.Context, id string, status CommandStatus, result map[string]any) error {
	resultJSON, _ := json.Marshal(result)

	var query string
	var args []any

	switch status {
	case StatusRunning:
		query = `UPDATE commands SET status = ?, started_at = ? WHERE id = ?`
		args = []any{status, time.Now().UTC().Format(time.RFC3339), id}
	case StatusCompleted, StatusFailed, StatusCancelled:
		query = `UPDATE commands SET status = ?, result = ?, ended_at = ? WHERE id = ?`
		args = []any{status, string(resultJSON), time.Now().UTC().Format(time.RFC3339), id}
	default:
		query = `UPDATE commands SET status = ? WHERE id = ?`
		args = []any{status, id}
	}

	_, err := s.db.ExecContext(ctx, query, args...)
	return err
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
	defer rows.Close()

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

// recoverOrphanedCommands marks running commands as queued after crash.
func (s *Scheduler) recoverOrphanedCommands(ctx context.Context) {
	result, err := s.db.ExecContext(ctx, `
		UPDATE commands SET status = ? WHERE status = ?
	`, StatusQueued, StatusRunning)

	if err != nil {
		slog.Error("Failed to recover orphaned commands", "error", err)
		return
	}

	if affected, _ := result.RowsAffected(); affected > 0 {
		slog.Info("Recovered orphaned commands", "count", affected)
	}
}

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
	defer rows.Close()

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

// TriggerTask immediately queues a scheduled task.
func (s *Scheduler) TriggerTask(ctx context.Context, name string) (string, error) {
	return s.QueueCommand(ctx, CommandName(name), TriggerManual, nil)
}
