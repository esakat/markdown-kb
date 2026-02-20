package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"nhooyr.io/websocket"
)

func TestHub_BroadcastToClients(t *testing.T) {
	hub := NewHub()
	defer hub.Close()

	// Create a test server with the hub
	srv := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer srv.Close()

	// Connect two clients
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c1, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial c1: %v", err)
	}
	defer c1.Close(websocket.StatusNormalClosure, "")

	c2, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial c2: %v", err)
	}
	defer c2.Close(websocket.StatusNormalClosure, "")

	// Wait for connections to register
	time.Sleep(100 * time.Millisecond)

	if hub.ClientCount() != 2 {
		t.Fatalf("expected 2 clients, got %d", hub.ClientCount())
	}

	// Broadcast an event
	hub.Broadcast(WSEvent{Type: "updated", Path: "guide.md"})

	// Read from both clients
	for i, c := range []*websocket.Conn{c1, c2} {
		_, data, err := c.Read(ctx)
		if err != nil {
			t.Fatalf("read client %d: %v", i+1, err)
		}
		var ev WSEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			t.Fatalf("unmarshal client %d: %v", i+1, err)
		}
		if ev.Type != "updated" || ev.Path != "guide.md" {
			t.Errorf("client %d: expected {updated guide.md}, got %+v", i+1, ev)
		}
	}
}

func TestHub_ClientDisconnect(t *testing.T) {
	hub := NewHub()
	defer hub.Close()

	srv := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if hub.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", hub.ClientCount())
	}

	// Close the client
	c.Close(websocket.StatusNormalClosure, "bye")

	// Wait for server to detect disconnect
	time.Sleep(200 * time.Millisecond)

	// Broadcast should clean up the dead connection
	hub.Broadcast(WSEvent{Type: "updated", Path: "test.md"})
	time.Sleep(100 * time.Millisecond)

	if hub.ClientCount() != 0 {
		t.Errorf("expected 0 clients after disconnect, got %d", hub.ClientCount())
	}
}

func TestHub_Close(t *testing.T) {
	hub := NewHub()

	srv := httptest.NewServer(http.HandlerFunc(hub.ServeWS))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer c.Close(websocket.StatusNormalClosure, "")

	time.Sleep(100 * time.Millisecond)

	hub.Close()

	if hub.ClientCount() != 0 {
		t.Errorf("expected 0 clients after Close, got %d", hub.ClientCount())
	}
}
