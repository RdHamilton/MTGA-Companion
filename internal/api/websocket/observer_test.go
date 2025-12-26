package websocket

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"

	"github.com/ramonehamilton/MTGA-Companion/internal/events"
)

func TestNewWebSocketObserver(t *testing.T) {
	hub := NewHub()
	observer := NewWebSocketObserver(hub)

	if observer == nil {
		t.Fatal("NewWebSocketObserver() returned nil")
	}

	if observer.hub != hub {
		t.Error("Observer hub reference is incorrect")
	}

	if observer.name != "WebSocketObserver" {
		t.Errorf("Expected name 'WebSocketObserver', got '%s'", observer.name)
	}
}

func TestWebSocketObserver_GetName(t *testing.T) {
	hub := NewHub()
	observer := NewWebSocketObserver(hub)

	name := observer.GetName()
	if name != "WebSocketObserver" {
		t.Errorf("Expected 'WebSocketObserver', got '%s'", name)
	}
}

func TestWebSocketObserver_ShouldHandle(t *testing.T) {
	hub := NewHub()
	observer := NewWebSocketObserver(hub)

	// Should handle all event types
	eventTypes := []string{
		"stats:updated",
		"rank:updated",
		"quest:updated",
		"draft:updated",
		"deck:updated",
		"collection:updated",
		"daemon:status",
		"daemon:connected",
		"daemon:error",
		"replay:started",
		"replay:progress",
		"replay:completed",
		"custom:event",
	}

	for _, eventType := range eventTypes {
		if !observer.ShouldHandle(eventType) {
			t.Errorf("Expected ShouldHandle(%s) to return true", eventType)
		}
	}
}

func TestWebSocketObserver_OnEvent_NilHub(t *testing.T) {
	observer := &WebSocketObserver{
		name: "TestObserver",
		hub:  nil,
	}

	event := events.Event{
		Type: "test:event",
		Data: map[string]interface{}{"key": "value"},
	}

	// Should not panic and return nil error
	err := observer.OnEvent(event)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

func TestWebSocketObserver_OnEvent_WithData(t *testing.T) {
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
	defer conn.Close()

	// Give time for registration
	time.Sleep(50 * time.Millisecond)

	// Create observer
	observer := NewWebSocketObserver(hub)

	// Send an event
	event := events.Event{
		Type: "stats:updated",
		Data: map[string]interface{}{
			"matches": 15,
			"games":   30,
		},
		Context: context.Background(),
	}

	err = observer.OnEvent(event)
	if err != nil {
		t.Errorf("OnEvent returned error: %v", err)
	}

	// Read the message from WebSocket
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	// Verify the message
	var received Event
	if err := json.Unmarshal(message, &received); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if received.Type != "stats:updated" {
		t.Errorf("Expected type 'stats:updated', got '%s'", received.Type)
	}

	dataMap, ok := received.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected Data to be a map")
	}

	if matches, ok := dataMap["matches"].(float64); !ok || int(matches) != 15 {
		t.Errorf("Expected matches=15, got %v", dataMap["matches"])
	}
}

func TestWebSocketObserver_OnEvent_WithTypedData(t *testing.T) {
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
	defer conn.Close()

	// Give time for registration
	time.Sleep(50 * time.Millisecond)

	// Create observer
	observer := NewWebSocketObserver(hub)

	// Define a typed event payload
	type StatsPayload struct {
		Matches int `json:"matches"`
		Games   int `json:"games"`
	}

	// Send an event with TypedData
	event := events.Event{
		Type: "stats:updated",
		Data: map[string]interface{}{
			"matches": 10, // This should be ignored
		},
		TypedData: StatsPayload{
			Matches: 25,
			Games:   50,
		},
		Context: context.Background(),
	}

	err = observer.OnEvent(event)
	if err != nil {
		t.Errorf("OnEvent returned error: %v", err)
	}

	// Read the message from WebSocket
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	// Verify the message uses TypedData
	var received struct {
		Type string       `json:"type"`
		Data StatsPayload `json:"data"`
	}
	if err := json.Unmarshal(message, &received); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if received.Type != "stats:updated" {
		t.Errorf("Expected type 'stats:updated', got '%s'", received.Type)
	}

	// Should use TypedData values
	if received.Data.Matches != 25 {
		t.Errorf("Expected matches=25 (from TypedData), got %d", received.Data.Matches)
	}

	if received.Data.Games != 50 {
		t.Errorf("Expected games=50 (from TypedData), got %d", received.Data.Games)
	}
}

func TestWebSocketObserver_ImplementsInterface(t *testing.T) {
	hub := NewHub()
	observer := NewWebSocketObserver(hub)

	// Verify it implements events.Observer
	var _ events.Observer = observer
}
