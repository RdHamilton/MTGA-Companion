package daemon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewWebSocketServer(t *testing.T) {
	server := NewWebSocketServer(9999)

	if server == nil {
		t.Fatal("NewWebSocketServer returned nil")
	}

	if server.port != 9999 {
		t.Errorf("Expected port 9999, got %d", server.port)
	}

	if server.clients == nil {
		t.Error("clients map not initialized")
	}

	if server.broadcast == nil {
		t.Error("broadcast channel not initialized")
	}
}

func TestWebSocketServer_ClientCount(t *testing.T) {
	server := NewWebSocketServer(9999)

	count := server.ClientCount()
	if count != 0 {
		t.Errorf("Expected 0 clients initially, got %d", count)
	}
}

func TestWebSocketServer_Broadcast(t *testing.T) {
	server := NewWebSocketServer(9999)

	event := Event{
		Type: "test:event",
		Data: map[string]interface{}{
			"message": "test",
		},
	}

	// Should not panic even with no clients
	server.Broadcast(event)
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

	if decoded.Data["matches"].(float64) != 5 {
		t.Errorf("Expected matches 5, got %v", decoded.Data["matches"])
	}
}

func TestWebSocketServer_StatusHandler(t *testing.T) {
	server := NewWebSocketServer(9999)

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	server.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var status map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if status["status"] != "running" {
		t.Errorf("Expected status running, got %v", status["status"])
	}

	if status["clients"] != float64(0) {
		t.Errorf("Expected 0 clients, got %v", status["clients"])
	}
}

func TestWebSocketServer_UpgradeConnection(t *testing.T) {
	server := NewWebSocketServer(9999)

	// Create test HTTP server
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWebSocket))
	defer httpServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")

	// Try to connect
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Read welcome message
	var welcome Event
	if err := conn.ReadJSON(&welcome); err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}

	if welcome.Type != "daemon:connected" {
		t.Errorf("Expected welcome type daemon:connected, got %s", welcome.Type)
	}

	// Give server time to register client
	time.Sleep(50 * time.Millisecond)

	// Check client count
	if server.ClientCount() != 1 {
		t.Errorf("Expected 1 client, got %d", server.ClientCount())
	}
}

func TestWebSocketServer_PingPong(t *testing.T) {
	server := NewWebSocketServer(9999)

	// Create test HTTP server
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWebSocket))
	defer httpServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")

	// Connect
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Read welcome message
	var welcome Event
	if err := conn.ReadJSON(&welcome); err != nil {
		t.Fatalf("Failed to read welcome: %v", err)
	}

	// Send ping
	ping := map[string]interface{}{
		"type": "ping",
	}
	if err := conn.WriteJSON(ping); err != nil {
		t.Fatalf("Failed to send ping: %v", err)
	}

	// Read pong
	var pong Event
	if err := conn.ReadJSON(&pong); err != nil {
		t.Fatalf("Failed to read pong: %v", err)
	}

	if pong.Type != "pong" {
		t.Errorf("Expected pong type, got %s", pong.Type)
	}
}

func TestWebSocketServer_Stop(t *testing.T) {
	server := NewWebSocketServer(9999)

	// Should not panic when stopping without starting
	if err := server.Stop(); err != nil {
		t.Errorf("Stop returned error: %v", err)
	}

	// Check clients cleared
	if len(server.clients) != 0 {
		t.Errorf("Expected clients to be cleared, got %d", len(server.clients))
	}
}
