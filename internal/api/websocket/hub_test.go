package websocket

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()

	if hub == nil {
		t.Fatal("NewHub() returned nil")
	}

	if hub.clients == nil {
		t.Error("Hub clients map is nil")
	}

	if hub.broadcast == nil {
		t.Error("Hub broadcast channel is nil")
	}

	if hub.register == nil {
		t.Error("Hub register channel is nil")
	}

	if hub.unregister == nil {
		t.Error("Hub unregister channel is nil")
	}
}

func TestHub_ClientCount(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Initially no clients
	if count := hub.ClientCount(); count != 0 {
		t.Errorf("Expected 0 clients, got %d", count)
	}
}

func TestHub_BroadcastEvent(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create a test event
	event := Event{
		Type: "test:event",
		Data: map[string]interface{}{
			"message": "hello",
		},
	}

	// Broadcasting with no clients should not panic
	hub.BroadcastEvent(event)

	// Give time for the goroutine to process
	time.Sleep(10 * time.Millisecond)
}

func TestEvent_JSON(t *testing.T) {
	event := Event{
		Type: "stats:updated",
		Data: map[string]interface{}{
			"matches": 10,
			"games":   25,
		},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	var decoded Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if decoded.Type != event.Type {
		t.Errorf("Expected type %s, got %s", event.Type, decoded.Type)
	}

	dataMap, ok := decoded.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected Data to be a map")
	}

	if matches, ok := dataMap["matches"].(float64); !ok || int(matches) != 10 {
		t.Errorf("Expected matches=10, got %v", dataMap["matches"])
	}
}

func TestHub_WebSocketConnection(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(hub.ServeWs))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect to WebSocket
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	defer conn.Close()

	// Give time for registration
	time.Sleep(50 * time.Millisecond)

	// Check client count
	if count := hub.ClientCount(); count != 1 {
		t.Errorf("Expected 1 client, got %d", count)
	}

	// Broadcast an event
	testEvent := Event{
		Type: "test:message",
		Data: map[string]string{"content": "hello"},
	}
	hub.BroadcastEvent(testEvent)

	// Read the message
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	// Verify the message
	var received Event
	if err := json.Unmarshal(message, &received); err != nil {
		t.Fatalf("Failed to unmarshal received message: %v", err)
	}

	if received.Type != "test:message" {
		t.Errorf("Expected type test:message, got %s", received.Type)
	}
}

func TestHub_MultipleClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(hub.ServeWs))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect multiple clients
	dialer := websocket.Dialer{}
	var conns []*websocket.Conn

	for i := 0; i < 3; i++ {
		conn, _, err := dialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect client %d: %v", i, err)
		}
		conns = append(conns, conn)
	}

	// Cleanup
	defer func() {
		for _, conn := range conns {
			conn.Close()
		}
	}()

	// Give time for registrations
	time.Sleep(50 * time.Millisecond)

	// Check client count
	if count := hub.ClientCount(); count != 3 {
		t.Errorf("Expected 3 clients, got %d", count)
	}

	// Broadcast an event
	testEvent := Event{
		Type: "broadcast:test",
		Data: map[string]int{"value": 42},
	}
	hub.BroadcastEvent(testEvent)

	// All clients should receive the message
	for i, conn := range conns {
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Errorf("Client %d failed to read message: %v", i, err)
			continue
		}

		var received Event
		if err := json.Unmarshal(message, &received); err != nil {
			t.Errorf("Client %d failed to unmarshal message: %v", i, err)
			continue
		}

		if received.Type != "broadcast:test" {
			t.Errorf("Client %d expected type broadcast:test, got %s", i, received.Type)
		}
	}
}

func TestHub_ClientDisconnect(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(hub.ServeWs))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect a client
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Give time for registration
	time.Sleep(50 * time.Millisecond)

	if count := hub.ClientCount(); count != 1 {
		t.Errorf("Expected 1 client after connect, got %d", count)
	}

	// Disconnect
	conn.Close()

	// Give time for unregistration
	time.Sleep(100 * time.Millisecond)

	if count := hub.ClientCount(); count != 0 {
		t.Errorf("Expected 0 clients after disconnect, got %d", count)
	}
}
