// Package scheduler provides a job scheduler and command queue processor.
// It reads commands from the database and executes them as goroutines,
// and runs scheduled tasks at configured intervals.
package scheduler

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"
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
		if uerr := s.updateCommandStatus(ctx, cmd.ID, StatusFailed, map[string]any{"error": err.Error()}); uerr != nil {
			slog.Error("Failed to update command status", "id", cmd.ID, "error", uerr)
		}
		slog.Error("No handler for command", "name", cmd.Name)
		return
	}

	// Execute handler
	err := handler(cmdCtx, cmd)

	// Update status based on result
	if err != nil {
		slog.Error("Command failed", "id", cmd.ID, "name", cmd.Name, "error", err)
		if uerr := s.updateCommandStatus(ctx, cmd.ID, StatusFailed, map[string]any{"error": err.Error()}); uerr != nil {
			slog.Error("Failed to update command status", "id", cmd.ID, "error", uerr)
		}
	} else {
		slog.Info("Command completed", "id", cmd.ID, "name", cmd.Name)
		if uerr := s.updateCommandStatus(ctx, cmd.ID, StatusCompleted, cmd.Result); uerr != nil {
			slog.Error("Failed to update command status", "id", cmd.ID, "error", uerr)
		}
	}
}
