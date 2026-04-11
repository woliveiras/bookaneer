package scheduler

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/woliveiras/bookaneer/internal/database"
)

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
	row := s.db.QueryRowContext(ctx, `
		SELECT id, name, status, priority, payload, result, trigger, queued_at, started_at, ended_at
		FROM commands WHERE id = ?
	`, id)

	cmd, err := scanCommand(row)
	if err != nil {
		return nil, err
	}
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
	defer func() { _ = rows.Close() }()

	return scanCommands(rows)
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
	defer func() { _ = rows.Close() }()

	return scanCommands(rows)
}

// CancelCommand cancels a running or queued command.
func (s *Scheduler) CancelCommand(ctx context.Context, id string) error {
	s.mu.Lock()
	if cancel, ok := s.running[id]; ok {
		cancel()
	}
	s.mu.Unlock()

	return s.updateCommandStatus(ctx, id, StatusCancelled, nil)
}

// getNextQueuedCommand gets the next command to execute (by priority then age).
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
	if err := json.Unmarshal([]byte(payloadJSON), &cmd.Payload); err != nil {
		slog.Warn("failed to unmarshal command payload", "id", cmd.ID, "error", err)
	}

	return &cmd, nil
}

// updateCommandStatus updates a command's status and timestamps.
func (s *Scheduler) updateCommandStatus(ctx context.Context, id string, status CommandStatus, result map[string]any) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		slog.Warn("failed to marshal command result", "id", id, "error", err)
		resultJSON = []byte("null")
	}

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

	_, err = s.db.ExecContext(ctx, query, args...)
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

// scanner abstracts sql.Row and sql.Rows for shared scanning.
type scanner = database.Scanner

// scanCommandRow scans a single command row.
func scanCommandRow(s scanner) (Command, error) {
	var cmd Command
	var payloadJSON, resultJSON string
	var startedAt, endedAt sql.NullString
	var queuedAt string

	if err := s.Scan(&cmd.ID, &cmd.Name, &cmd.Status, &cmd.Priority, &payloadJSON, &resultJSON, &cmd.Trigger, &queuedAt, &startedAt, &endedAt); err != nil {
		return Command{}, err
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

	if err := json.Unmarshal([]byte(payloadJSON), &cmd.Payload); err != nil {
		slog.Warn("failed to unmarshal command payload", "id", cmd.ID, "error", err)
	}
	if resultJSON != "" && resultJSON != "null" {
		if err := json.Unmarshal([]byte(resultJSON), &cmd.Result); err != nil {
			slog.Warn("failed to unmarshal command result", "id", cmd.ID, "error", err)
		}
	}

	return cmd, nil
}

// scanCommand scans a single row into a Command.
func scanCommand(row *sql.Row) (Command, error) {
	return scanCommandRow(row)
}

// scanCommands scans multiple rows into a Command slice.
func scanCommands(rows *sql.Rows) ([]Command, error) {
	var commands []Command
	for rows.Next() {
		cmd, err := scanCommandRow(rows)
		if err != nil {
			return nil, err
		}
		commands = append(commands, cmd)
	}
	return commands, rows.Err()
}
