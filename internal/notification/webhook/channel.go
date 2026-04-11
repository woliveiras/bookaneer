// Package webhook implements a webhook notification channel.
package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/woliveiras/bookaneer/internal/notification"
)

// Settings holds webhook-specific configuration.
type Settings struct {
	URL     string            `json:"url"`
	Method  string            `json:"method,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Channel sends notifications via HTTP webhooks.
type Channel struct {
	name     string
	settings Settings
	client   *http.Client
}

// New creates a Channel from a notification config.
func New(cfg notification.Config) (notification.Channel, error) {
	var s Settings
	if err := json.Unmarshal(cfg.Settings, &s); err != nil {
		return nil, fmt.Errorf("parse webhook settings: %w", err)
	}
	if s.URL == "" {
		return nil, fmt.Errorf("webhook URL is required")
	}
	if _, err := url.ParseRequestURI(s.URL); err != nil {
		return nil, fmt.Errorf("invalid webhook URL: %w", err)
	}
	if s.Method == "" {
		s.Method = http.MethodPost
	}
	return &Channel{
		name:     cfg.Name,
		settings: s,
		client:   &http.Client{Timeout: 15 * time.Second},
	}, nil
}

func (ch *Channel) Name() string { return ch.name }
func (ch *Channel) Type() string { return "webhook" }

// Test sends a test notification to verify the webhook is reachable.
func (ch *Channel) Test(ctx context.Context) error {
	event := notification.Event{
		Type:       notification.EventTest,
		Title:      "Bookaneer Test",
		Message:    "This is a test notification from Bookaneer.",
		OccurredAt: time.Now().UTC(),
	}
	return ch.Send(ctx, event)
}

// Send dispatches the event to the webhook URL with retries.
func (ch *Channel) Send(ctx context.Context, event notification.Event) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt) * 2 * time.Second):
			}
		}

		req, err := http.NewRequestWithContext(ctx, ch.settings.Method, ch.settings.URL, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Bookaneer-Event", string(event.Type))
		req.Header.Set("User-Agent", "Bookaneer/1.0")

		for k, v := range ch.settings.Headers {
			req.Header.Set(k, v)
		}

		resp, err := ch.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("webhook request failed: %w", err)
			continue
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}
		lastErr = fmt.Errorf("webhook returned status %d", resp.StatusCode)
		if resp.StatusCode < 500 {
			return lastErr // Don't retry client errors
		}
	}
	return lastErr
}
