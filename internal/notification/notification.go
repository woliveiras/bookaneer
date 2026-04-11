// Package notification provides the notification channel interface,
// service for managing notification configurations, and event dispatch.
package notification

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// Channel is implemented by each notification provider (webhook, etc.).
type Channel interface {
	Name() string
	Type() string
	Test(ctx context.Context) error
	Send(ctx context.Context, event Event) error
}

// EventType identifies the kind of notification event.
type EventType string

const (
	EventGrab        EventType = "grab"
	EventDownload    EventType = "download"
	EventUpgrade     EventType = "upgrade"
	EventHealthIssue EventType = "health_issue"
	EventImport      EventType = "import"
	EventRename      EventType = "rename"
	EventTest        EventType = "test"
)

// BookInfo carries book data for notification payloads.
type BookInfo struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Author string `json:"author"`
}

// Event is dispatched to notification channels.
type Event struct {
	Type       EventType `json:"event"`
	Title      string    `json:"title"`
	Message    string    `json:"message"`
	Book       *BookInfo `json:"book,omitempty"`
	OccurredAt time.Time `json:"occurred_at"`
}

// Config represents a persisted notification channel configuration.
type Config struct {
	ID         int64           `json:"id"`
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Settings   json.RawMessage `json:"settings"`
	OnGrab     bool            `json:"onGrab"`
	OnDownload bool            `json:"onDownload"`
	OnUpgrade  bool            `json:"onUpgrade"`
	Enabled    bool            `json:"enabled"`
}

// CreateInput is the payload for creating a notification channel.
type CreateInput struct {
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Settings   json.RawMessage `json:"settings"`
	OnGrab     bool            `json:"onGrab"`
	OnDownload bool            `json:"onDownload"`
	OnUpgrade  bool            `json:"onUpgrade"`
	Enabled    bool            `json:"enabled"`
}

// UpdateInput is the payload for updating a notification channel.
type UpdateInput struct {
	Name       *string          `json:"name,omitempty"`
	Type       *string          `json:"type,omitempty"`
	Settings   *json.RawMessage `json:"settings,omitempty"`
	OnGrab     *bool            `json:"onGrab,omitempty"`
	OnDownload *bool            `json:"onDownload,omitempty"`
	OnUpgrade  *bool            `json:"onUpgrade,omitempty"`
	Enabled    *bool            `json:"enabled,omitempty"`
}

var (
	ErrNotFound    = errors.New("notification not found")
	ErrUnsupported = errors.New("unsupported notification type")
)

// ChannelFactory creates a Channel from a Config.
type ChannelFactory func(cfg Config) (Channel, error)

// Service manages notification configurations and dispatches events.
type Service struct {
	db        *sql.DB
	mu        sync.RWMutex
	factories map[string]ChannelFactory
}

// New creates a new notification Service.
func New(db *sql.DB) *Service {
	return &Service{
		db:        db,
		factories: make(map[string]ChannelFactory),
	}
}

// RegisterFactory registers a channel factory for a given type.
func (s *Service) RegisterFactory(typeName string, f ChannelFactory) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.factories[typeName] = f
}

// List returns all notification configs.
func (s *Service) List(ctx context.Context) ([]Config, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, name, type, settings, on_grab, on_download, on_upgrade, enabled
		FROM notifications ORDER BY id
	`)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	var configs []Config
	for rows.Next() {
		var c Config
		var settings string
		var onGrab, onDownload, onUpgrade, enabled int
		if err := rows.Scan(&c.ID, &c.Name, &c.Type, &settings, &onGrab, &onDownload, &onUpgrade, &enabled); err != nil {
			return nil, fmt.Errorf("scan notification: %w", err)
		}
		c.Settings = json.RawMessage(settings)
		c.OnGrab = onGrab == 1
		c.OnDownload = onDownload == 1
		c.OnUpgrade = onUpgrade == 1
		c.Enabled = enabled == 1
		configs = append(configs, c)
	}
	return configs, rows.Err()
}

// FindByID returns a notification config by ID.
func (s *Service) FindByID(ctx context.Context, id int64) (*Config, error) {
	var c Config
	var settings string
	var onGrab, onDownload, onUpgrade, enabled int
	err := s.db.QueryRowContext(ctx, `
		SELECT id, name, type, settings, on_grab, on_download, on_upgrade, enabled
		FROM notifications WHERE id = ?
	`, id).Scan(&c.ID, &c.Name, &c.Type, &settings, &onGrab, &onDownload, &onUpgrade, &enabled)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find notification %d: %w", id, err)
	}
	c.Settings = json.RawMessage(settings)
	c.OnGrab = onGrab == 1
	c.OnDownload = onDownload == 1
	c.OnUpgrade = onUpgrade == 1
	c.Enabled = enabled == 1
	return &c, nil
}

// Create inserts a new notification config.
func (s *Service) Create(ctx context.Context, input CreateInput) (*Config, error) {
	settings := input.Settings
	if settings == nil {
		settings = json.RawMessage("{}")
	}
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO notifications (name, type, settings, on_grab, on_download, on_upgrade, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, input.Name, input.Type, string(settings), boolToInt(input.OnGrab), boolToInt(input.OnDownload), boolToInt(input.OnUpgrade), boolToInt(input.Enabled))
	if err != nil {
		return nil, fmt.Errorf("create notification: %w", err)
	}
	id, _ := result.LastInsertId()
	return s.FindByID(ctx, id)
}

// Update modifies an existing notification config.
func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*Config, error) {
	existing, err := s.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if input.Name != nil {
		existing.Name = *input.Name
	}
	if input.Type != nil {
		existing.Type = *input.Type
	}
	if input.Settings != nil {
		existing.Settings = *input.Settings
	}
	if input.OnGrab != nil {
		existing.OnGrab = *input.OnGrab
	}
	if input.OnDownload != nil {
		existing.OnDownload = *input.OnDownload
	}
	if input.OnUpgrade != nil {
		existing.OnUpgrade = *input.OnUpgrade
	}
	if input.Enabled != nil {
		existing.Enabled = *input.Enabled
	}

	_, err = s.db.ExecContext(ctx, `
		UPDATE notifications
		SET name = ?, type = ?, settings = ?, on_grab = ?, on_download = ?, on_upgrade = ?, enabled = ?
		WHERE id = ?
	`, existing.Name, existing.Type, string(existing.Settings),
		boolToInt(existing.OnGrab), boolToInt(existing.OnDownload), boolToInt(existing.OnUpgrade), boolToInt(existing.Enabled), id)
	if err != nil {
		return nil, fmt.Errorf("update notification %d: %w", id, err)
	}
	return existing, nil
}

// Delete removes a notification config.
func (s *Service) Delete(ctx context.Context, id int64) error {
	result, err := s.db.ExecContext(ctx, "DELETE FROM notifications WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete notification %d: %w", id, err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

// TestChannel creates a channel from config and sends a test event.
func (s *Service) TestChannel(ctx context.Context, id int64) error {
	cfg, err := s.FindByID(ctx, id)
	if err != nil {
		return err
	}
	ch, err := s.buildChannel(*cfg)
	if err != nil {
		return err
	}
	return ch.Test(ctx)
}

// Dispatch sends an event to all enabled channels that subscribe to the event type.
func (s *Service) Dispatch(ctx context.Context, event Event) {
	configs, err := s.List(ctx)
	if err != nil {
		slog.Warn("Failed to list notification channels", "error", err)
		return
	}
	for _, cfg := range configs {
		if !cfg.Enabled {
			continue
		}
		if !s.shouldSend(cfg, event.Type) {
			continue
		}
		ch, err := s.buildChannel(cfg)
		if err != nil {
			slog.Warn("Failed to build notification channel", "name", cfg.Name, "error", err)
			continue
		}
		if err := ch.Send(ctx, event); err != nil {
			slog.Warn("Failed to send notification", "name", cfg.Name, "type", cfg.Type, "error", err)
		}
	}
}

func (s *Service) shouldSend(cfg Config, eventType EventType) bool {
	switch eventType {
	case EventGrab:
		return cfg.OnGrab
	case EventDownload, EventImport:
		return cfg.OnDownload
	case EventUpgrade:
		return cfg.OnUpgrade
	case EventTest:
		return true
	default:
		return cfg.OnDownload // default to on_download for unknown events
	}
}

func (s *Service) buildChannel(cfg Config) (Channel, error) {
	s.mu.RLock()
	factory, ok := s.factories[cfg.Type]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupported, cfg.Type)
	}
	return factory(cfg)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
