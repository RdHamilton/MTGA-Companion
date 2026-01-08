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

	if hub.done == nil {
		t.Error("Hub done channel is nil")
	}

	if hub.stopped {
		t.Error("Hub should not be stopped initially")
	}
}

func TestHub_ClientCount(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

	// Initially no clients
	if count := hub.ClientCount(); count != 0 {
		t.Errorf("Expected 0 clients, got %d", count)
	}
}

func TestHub_BroadcastEvent(t *testing.T) {
	hub := NewHub()
	go hub.Run()
	defer hub.Stop()

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
	defer hub.Stop()

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
	defer hub.Stop()

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
	defer hub.Stop()

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

func TestHub_Stop(t *testing.T) {
	hub := NewHub()

	// Start the hub in a goroutine
	done := make(chan struct{})
	go func() {
		hub.Run()
		close(done)
	}()

	// Give time for hub to start
	time.Sleep(10 * time.Millisecond)

	// Stop the hub
	hub.Stop()

	// Wait for hub to stop with timeout
	select {
	case <-done:
		// Hub stopped successfully
	case <-time.After(1 * time.Second):
		t.Fatal("Hub did not stop within timeout")
	}
}

func TestHub_Stop_CleansUpClients(t *testing.T) {
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

	// Give time for registrations
	time.Sleep(50 * time.Millisecond)

	// Verify clients connected
	if count := hub.ClientCount(); count != 3 {
		t.Errorf("Expected 3 clients before stop, got %d", count)
	}

	// Stop the hub
	hub.Stop()

	// Give time for cleanup
	time.Sleep(50 * time.Millisecond)

	// After stop, client count should be 0
	if count := hub.ClientCount(); count != 0 {
		t.Errorf("Expected 0 clients after stop, got %d", count)
	}

	// Clean up connections
	for _, conn := range conns {
		conn.Close()
	}
}

func TestHub_Stop_Idempotent(t *testing.T) {
	hub := NewHub()

	done := make(chan struct{})
	go func() {
		hub.Run()
		close(done)
	}()

	time.Sleep(10 * time.Millisecond)

	// Stop should not panic when called once
	hub.Stop()

	select {
	case <-done:
		// Hub stopped successfully
	case <-time.After(1 * time.Second):
		t.Fatal("Hub did not stop within timeout")
	}

	// Calling Stop() multiple times should not panic (idempotent)
	hub.Stop()
	hub.Stop()
	hub.Stop()
}

func TestHub_IsStopped(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	if hub.IsStopped() {
		t.Error("Expected IsStopped() to be false before Stop()")
	}

	hub.Stop()

	// Give time for stop to propagate
	time.Sleep(50 * time.Millisecond)

	if !hub.IsStopped() {
		t.Error("Expected IsStopped() to be true after Stop()")
	}
}

func TestHub_BroadcastEvent_AfterStop(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	hub.Stop()

	// Give time for stop to propagate
	time.Sleep(50 * time.Millisecond)

	// BroadcastEvent should return false after stop
	event := Event{
		Type: "test:event",
		Data: map[string]interface{}{"key": "value"},
	}

	result := hub.BroadcastEvent(event)
	if result {
		t.Error("Expected BroadcastEvent to return false after Stop()")
	}
}
