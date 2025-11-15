package ipc

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewClient(t *testing.T) {
	client := NewClient("ws://localhost:9999")

	if client == nil {
		t.Fatal("NewClient returned nil")
	}

	if client.url != "ws://localhost:9999" {
		t.Errorf("Expected URL ws://localhost:9999, got %s", client.url)
	}

	if client.handlers == nil {
		t.Error("handlers map not initialized")
	}

	if !client.reconnect {
		t.Error("reconnect should be true by default")
	}
}

func TestClient_Connect(t *testing.T) {
	// Create test WebSocket server
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade: %v", err)
		}
		defer conn.Close()

		// Keep connection open
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create client and connect
	client := NewClient(wsURL)
	if err := client.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Stop()

	if !client.IsConnected() {
		t.Error("Expected client to be connected")
	}

	if client.conn == nil {
		t.Error("Expected conn to be set")
	}
}

func TestClient_On(t *testing.T) {
	client := NewClient("ws://localhost:9999")

	called := false
	handler := func(data map[string]interface{}) {
		called = true
	}

	client.On("test:event", handler)

	// Check handler was registered
	client.handlersMu.RLock()
	handlers, ok := client.handlers["test:event"]
	client.handlersMu.RUnlock()

	if !ok {
		t.Error("Handler not registered")
	}

	if len(handlers) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(handlers))
	}

	// Test handler execution
	handlers[0](nil)
	time.Sleep(10 * time.Millisecond) // Give goroutine time to execute

	if !called {
		t.Error("Handler was not called")
	}
}

func TestClient_DispatchEvent(t *testing.T) {
	client := NewClient("ws://localhost:9999")

	receivedData := make(map[string]interface{})
	done := make(chan bool)

	client.On("stats:updated", func(data map[string]interface{}) {
		receivedData = data
		done <- true
	})

	event := Event{
		Type: "stats:updated",
		Data: map[string]interface{}{
			"matches": float64(5),
			"games":   float64(10),
		},
		Timestamp: time.Now(),
	}

	// Dispatch event
	client.dispatchEvent(event)

	// Wait for handler
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Handler was not called within timeout")
	}

	if receivedData["matches"] != float64(5) {
		t.Errorf("Expected matches 5, got %v", receivedData["matches"])
	}
}

func TestClient_SendPing(t *testing.T) {
	// Create test WebSocket server
	upgrader := websocket.Upgrader{}
	received := make(chan map[string]interface{})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("Failed to upgrade: %v", err)
		}
		defer conn.Close()

		// Read ping message
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			t.Fatalf("Failed to read message: %v", err)
		}
		received <- msg
	}))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create client and connect
	client := NewClient(wsURL)
	if err := client.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}
	defer client.Stop()

	// Send ping
	if err := client.SendPing(); err != nil {
		t.Fatalf("SendPing failed: %v", err)
	}

	// Wait for server to receive ping
	select {
	case msg := <-received:
		if msg["type"] != "ping" {
			t.Errorf("Expected type ping, got %v", msg["type"])
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Server did not receive ping within timeout")
	}
}

func TestClient_Stop(t *testing.T) {
	// Create test WebSocket server
	upgrader := websocket.Upgrader{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}))
	defer server.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Create client and connect
	client := NewClient(wsURL)
	if err := client.Connect(); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if !client.IsConnected() {
		t.Error("Expected client to be connected")
	}

	// Stop client
	client.Stop()

	time.Sleep(50 * time.Millisecond) // Give time for cleanup

	if client.IsConnected() {
		t.Error("Expected client to be disconnected after Stop()")
	}

	if client.reconnect {
		t.Error("Expected reconnect to be disabled after Stop()")
	}
}

func TestClient_IsConnected(t *testing.T) {
	client := NewClient("ws://localhost:9999")

	if client.IsConnected() {
		t.Error("Expected client to not be connected initially")
	}

	client.setConnected(true)

	if !client.IsConnected() {
		t.Error("Expected client to be connected after setConnected(true)")
	}

	client.setConnected(false)

	if client.IsConnected() {
		t.Error("Expected client to not be connected after setConnected(false)")
	}
}

func TestClient_GetURL(t *testing.T) {
	url := "ws://localhost:9999"
	client := NewClient(url)

	if client.GetURL() != url {
		t.Errorf("Expected URL %s, got %s", url, client.GetURL())
	}
}

func TestMarshalEvent(t *testing.T) {
	event := Event{
		Type: "test:event",
		Data: map[string]interface{}{
			"key": "value",
		},
		Timestamp: time.Date(2025, 11, 15, 10, 30, 0, 0, time.UTC),
	}

	jsonStr, err := MarshalEvent(event)
	if err != nil {
		t.Fatalf("MarshalEvent failed: %v", err)
	}

	// Unmarshal to verify
	var decoded Event
	if err := json.Unmarshal([]byte(jsonStr), &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.Type != "test:event" {
		t.Errorf("Expected type test:event, got %s", decoded.Type)
	}

	if decoded.Data["key"] != "value" {
		t.Errorf("Expected key=value, got %v", decoded.Data["key"])
	}
}

func TestEvent_Structure(t *testing.T) {
	event := Event{
		Type: "stats:updated",
		Data: map[string]interface{}{
			"matches": 5,
			"games":   10,
		},
		Timestamp: time.Now(),
	}

	// Test JSON marshaling
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	// Test JSON unmarshaling
	var decoded Event
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if decoded.Type != "stats:updated" {
		t.Errorf("Expected type stats:updated, got %s", decoded.Type)
	}
}
