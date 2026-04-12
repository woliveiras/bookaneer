package ws

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHub_BroadcastAndClientCount(t *testing.T) {
	hub := NewHub()
	assert.Equal(t, 0, hub.ClientCount())

	// Create Echo + httptest server with WS endpoint
	e := echo.New()
	e.GET("/ws", hub.HandleWS)
	server := httptest.NewServer(e)
	defer server.Close()

	// Connect WS client
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() { _ = conn.Close() }()

	// Give the hub time to register the client
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, hub.ClientCount())

	// Broadcast a message
	hub.Broadcast("test_event", map[string]string{"key": "value"})

	// Read the message from the client
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, msg, err := conn.ReadMessage()
	require.NoError(t, err)
	assert.Contains(t, string(msg), `"event":"test_event"`)
	assert.Contains(t, string(msg), `"key":"value"`)
}

func TestHub_MultipleClients(t *testing.T) {
	hub := NewHub()

	e := echo.New()
	e.GET("/ws", hub.HandleWS)
	server := httptest.NewServer(e)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"

	// Connect two clients
	conn1, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() { _ = conn1.Close() }()

	conn2, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer func() { _ = conn2.Close() }()

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 2, hub.ClientCount())

	hub.Broadcast("hello", "world")

	// Both clients should receive the message
	for _, conn := range []*websocket.Conn{conn1, conn2} {
		_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, msg, err := conn.ReadMessage()
		require.NoError(t, err)
		assert.Contains(t, string(msg), `"event":"hello"`)
	}
}

func TestHub_ClientDisconnect(t *testing.T) {
	hub := NewHub()

	e := echo.New()
	e.GET("/ws", hub.HandleWS)
	server := httptest.NewServer(e)
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)

	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, 1, hub.ClientCount())

	_ = conn.Close()
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 0, hub.ClientCount())
}

func TestHub_HandleWS_UpgradeFails(t *testing.T) {
	hub := NewHub()

	e := echo.New()
	e.GET("/ws", hub.HandleWS)

	// Normal HTTP request (not WS upgrade) should fail
	req := httptest.NewRequest(http.MethodGet, "/ws", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
