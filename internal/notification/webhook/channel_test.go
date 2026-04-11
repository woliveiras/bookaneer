package webhook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/woliveiras/bookaneer/internal/notification"
)

func TestNew(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := notification.Config{
			Name:     "test",
			Settings: json.RawMessage(`{"url":"https://example.com/hook"}`),
		}
		ch, err := New(cfg)
		require.NoError(t, err)
		assert.Equal(t, "test", ch.Name())
		assert.Equal(t, "webhook", ch.Type())
	})

	t.Run("missing URL", func(t *testing.T) {
		cfg := notification.Config{
			Settings: json.RawMessage(`{}`),
		}
		_, err := New(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "URL is required")
	})

	t.Run("invalid URL", func(t *testing.T) {
		cfg := notification.Config{
			Settings: json.RawMessage(`{"url":"not a url"}`),
		}
		_, err := New(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid webhook URL")
	})

	t.Run("invalid JSON", func(t *testing.T) {
		cfg := notification.Config{
			Settings: json.RawMessage(`{bad json}`),
		}
		_, err := New(cfg)
		assert.Error(t, err)
	})

	t.Run("default method is POST", func(t *testing.T) {
		cfg := notification.Config{
			Settings: json.RawMessage(`{"url":"https://example.com/hook"}`),
		}
		ch, err := New(cfg)
		require.NoError(t, err)
		wh := ch.(*Channel)
		assert.Equal(t, http.MethodPost, wh.settings.Method)
	})

	t.Run("custom method", func(t *testing.T) {
		cfg := notification.Config{
			Settings: json.RawMessage(`{"url":"https://example.com/hook","method":"PUT"}`),
		}
		ch, err := New(cfg)
		require.NoError(t, err)
		wh := ch.(*Channel)
		assert.Equal(t, "PUT", wh.settings.Method)
	})
}

func TestChannel_Send(t *testing.T) {
	t.Run("successful send", func(t *testing.T) {
		var receivedBody []byte
		var receivedHeaders http.Header
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedBody, _ = io.ReadAll(r.Body)
			receivedHeaders = r.Header
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := notification.Config{
			Name:     "test",
			Settings: json.RawMessage(`{"url":"` + server.URL + `"}`),
		}
		ch, err := New(cfg)
		require.NoError(t, err)

		event := notification.Event{
			Type:       notification.EventGrab,
			Title:      "Test Book",
			Message:    "Book grabbed",
			OccurredAt: time.Now().UTC(),
		}
		err = ch.Send(context.Background(), event)
		require.NoError(t, err)

		assert.Equal(t, "application/json", receivedHeaders.Get("Content-Type"))
		assert.Equal(t, "grab", receivedHeaders.Get("X-Bookaneer-Event"))
		assert.Equal(t, "Bookaneer/1.0", receivedHeaders.Get("User-Agent"))

		var received notification.Event
		err = json.Unmarshal(receivedBody, &received)
		require.NoError(t, err)
		assert.Equal(t, notification.EventGrab, received.Type)
		assert.Equal(t, "Test Book", received.Title)
	})

	t.Run("custom headers", func(t *testing.T) {
		var receivedHeaders http.Header
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedHeaders = r.Header
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := notification.Config{
			Settings: json.RawMessage(`{"url":"` + server.URL + `","headers":{"X-Custom":"value123"}}`),
		}
		ch, err := New(cfg)
		require.NoError(t, err)

		err = ch.Send(context.Background(), notification.Event{Type: notification.EventTest})
		require.NoError(t, err)
		assert.Equal(t, "value123", receivedHeaders.Get("X-Custom"))
	})

	t.Run("client error no retry", func(t *testing.T) {
		var callCount int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			w.WriteHeader(http.StatusBadRequest)
		}))
		defer server.Close()

		cfg := notification.Config{
			Settings: json.RawMessage(`{"url":"` + server.URL + `"}`),
		}
		ch, err := New(cfg)
		require.NoError(t, err)

		err = ch.Send(context.Background(), notification.Event{Type: notification.EventTest})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "400")
		assert.Equal(t, 1, callCount) // No retry for 4xx
	})

	t.Run("server error retries", func(t *testing.T) {
		var callCount int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount < 3 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		cfg := notification.Config{
			Settings: json.RawMessage(`{"url":"` + server.URL + `"}`),
		}
		ch, err := New(cfg)
		require.NoError(t, err)

		err = ch.Send(context.Background(), notification.Event{Type: notification.EventTest})
		require.NoError(t, err)
		assert.Equal(t, 3, callCount) // Retried twice, succeeded third time
	})

	t.Run("context cancellation stops retries", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		cfg := notification.Config{
			Settings: json.RawMessage(`{"url":"` + server.URL + `"}`),
		}
		ch, err := New(cfg)
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		err = ch.Send(ctx, notification.Event{Type: notification.EventTest})
		assert.Error(t, err)
	})
}

func TestChannel_Test(t *testing.T) {
	var receivedEvent string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedEvent = r.Header.Get("X-Bookaneer-Event")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	cfg := notification.Config{
		Name:     "test-hook",
		Settings: json.RawMessage(`{"url":"` + server.URL + `"}`),
	}
	ch, err := New(cfg)
	require.NoError(t, err)

	err = ch.Test(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "test", receivedEvent)
}
