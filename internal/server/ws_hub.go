package server

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"nhooyr.io/websocket"
)

// WSEvent is a message sent over WebSocket to clients.
type WSEvent struct {
	Type string `json:"type"` // "created", "updated", "deleted"
	Path string `json:"path"` // relative path of the changed file
}

// Hub manages WebSocket connections and broadcasts events.
type Hub struct {
	mu      sync.RWMutex
	clients map[*websocket.Conn]context.CancelFunc
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*websocket.Conn]context.CancelFunc),
	}
}

// ServeWS handles WebSocket upgrade and registers the client.
// The handler blocks until the connection is closed.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // Allow all origins for dev
	})
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(r.Context())
	h.mu.Lock()
	h.clients[conn] = cancel
	h.mu.Unlock()

	// Block on reading client messages until the connection closes
	defer func() {
		h.remove(conn)
		conn.Close(websocket.StatusNormalClosure, "")
	}()
	for {
		_, _, err := conn.Read(ctx)
		if err != nil {
			return
		}
	}
}

// Broadcast sends an event to all connected clients.
func (h *Hub) Broadcast(event WSEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	h.mu.RLock()
	clients := make(map[*websocket.Conn]context.CancelFunc, len(h.clients))
	for c, cancel := range h.clients {
		clients[c] = cancel
	}
	h.mu.RUnlock()

	for conn := range clients {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
			h.remove(conn)
			conn.Close(websocket.StatusGoingAway, "write error")
		}
		cancel()
	}
}

// Close disconnects all clients.
func (h *Hub) Close() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for conn, cancel := range h.clients {
		cancel()
		conn.Close(websocket.StatusGoingAway, "server shutting down")
	}
	h.clients = make(map[*websocket.Conn]context.CancelFunc)
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) remove(conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if cancel, ok := h.clients[conn]; ok {
		cancel()
		delete(h.clients, conn)
	}
}
