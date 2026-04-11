// Package ws provides a WebSocket hub for real-time event broadcasting.
package ws

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMsgSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Rely on auth middleware / API-key check
	},
}

// Message is a typed event sent over WebSocket.
type Message struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// Hub maintains a set of active WebSocket clients and broadcasts messages.
type Hub struct {
	mu      sync.RWMutex
	clients map[*client]struct{}
}

type client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// NewHub creates a new WebSocket Hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*client]struct{}),
	}
}

// Broadcast sends a typed event to all connected clients.
func (h *Hub) Broadcast(event string, data interface{}) {
	msg := Message{Event: event, Data: data}
	raw, err := json.Marshal(msg)
	if err != nil {
		slog.Warn("ws: failed to marshal broadcast", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients {
		select {
		case c.send <- raw:
		default:
			go h.removeClient(c)
		}
	}
}

// ClientCount returns the number of active connections.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) addClient(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = struct{}{}
	slog.Debug("ws: client connected", "total", len(h.clients))
}

func (h *Hub) removeClient(c *client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
		slog.Debug("ws: client disconnected", "total", len(h.clients))
	}
}

// HandleWS is the Echo handler for WebSocket upgrades.
func (h *Hub) HandleWS(c echo.Context) error {
	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	cl := &client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}
	h.addClient(cl)

	go cl.writePump()
	go cl.readPump()

	return nil
}

func (c *client) readPump() {
	defer func() {
		c.hub.removeClient(c)
		_ = c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMsgSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case msg, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
