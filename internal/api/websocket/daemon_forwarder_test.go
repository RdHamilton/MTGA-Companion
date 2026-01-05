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

// MockDaemonEvent simulates a daemon.Event struct for testing.
// This mimics the daemon.Event without importing the daemon package.
type MockDaemonEvent struct {
	Type string
	Data interface{}
}

func TestNewDaemonEventForwarder(t *testing.T) {
	hub := NewHub()
	forwarder := NewDaemonEventForwarder(hub)

	if forwarder == nil {
		t.Fatal("NewDaemonEventForwarder() returned nil")
	}

	if forwarder.hub != hub {
		t.Error("Forwarder hub reference is incorrect")
	}
}

func TestDaemonEventForwarder_ForwardEvent_NoClients(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	forwarder := NewDaemonEventForwarder(hub)

	// Create a daemon event
	event := MockDaemonEvent{
		Type: "quest:updated",
		Data: map[string]interface{}{
			"quests": 5,
		},
	}

	// Should not panic when no clients are connected
	forwarder.ForwardEvent(event)

	// Give time for the goroutine to process
	time.Sleep(10 * time.Millisecond)
}

func TestDaemonEventForwarder_ForwardEvent_AllDaemonEventTypes(t *testing.T) {
	// List all daemon event types that should be forwarded
	daemonEventTypes := []struct {
		eventType string
		data      interface{}
	}{
		{"daemon:status", map[string]interface{}{"status": "connected", "watching": true}},
		{"daemon:error", map[string]interface{}{"error": "connection failed"}},
		{"stats:updated", map[string]interface{}{"matches": 10, "wins": 5}},
		{"deck:updated", map[string]interface{}{"deckId": 123}},
		{"rank:updated", map[string]interface{}{"rank": "Gold", "tier": 2}},
		{"quest:updated", map[string]interface{}{"quests": []string{"quest1", "quest2"}}},
		{"draft:updated", map[string]interface{}{"sessionId": "draft123", "picks": 5}},
		{"replay:started", map[string]interface{}{"totalEntries": 1000, "speed": 2.0, "filter": "draft"}},
		{"replay:error", map[string]interface{}{"error": "file not found"}},
		{"replay:completed", map[string]interface{}{"totalEntries": 1000, "elapsed": 60.5}},
		{"replay:progress", map[string]interface{}{"currentEntry": 500, "totalEntries": 1000, "percentComplete": 50.0}},
		{"replay:paused", map[string]interface{}{"currentEntry": 250, "totalEntries": 1000}},
		{"replay:resumed", map[string]interface{}{"currentEntry": 250, "totalEntries": 1000}},
		{"replay:draft_detected", map[string]interface{}{"currentEntry": 100, "message": "Draft event detected"}},
	}

	for _, tc := range daemonEventTypes {
		t.Run(tc.eventType, func(t *testing.T) {
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

			// Create forwarder and forward event
			forwarder := NewDaemonEventForwarder(hub)
			event := MockDaemonEvent{
				Type: tc.eventType,
				Data: tc.data,
			}
			forwarder.ForwardEvent(event)

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

			if received.Type != tc.eventType {
				t.Errorf("Expected type '%s', got '%s'", tc.eventType, received.Type)
			}

			if received.Data == nil {
				t.Error("Expected Data to not be nil")
			}
		})
	}
}

func TestDaemonEventForwarder_ForwardEvent_MultipleClients(t *testing.T) {
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

	defer func() {
		for _, conn := range conns {
			conn.Close()
		}
	}()

	// Give time for registrations
	time.Sleep(50 * time.Millisecond)

	// Verify all clients connected
	if count := hub.ClientCount(); count != 3 {
		t.Errorf("Expected 3 clients, got %d", count)
	}

	// Create forwarder and forward event
	forwarder := NewDaemonEventForwarder(hub)
	event := MockDaemonEvent{
		Type: "quest:updated",
		Data: map[string]interface{}{
			"quests": 5,
		},
	}
	forwarder.ForwardEvent(event)

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

		if received.Type != "quest:updated" {
			t.Errorf("Client %d expected type 'quest:updated', got '%s'", i, received.Type)
		}
	}
}

func TestDaemonEventForwarder_ForwardEvent_EmptyType(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	forwarder := NewDaemonEventForwarder(hub)

	// Event with empty type should be logged as warning but not panic
	event := MockDaemonEvent{
		Type: "",
		Data: map[string]interface{}{"key": "value"},
	}

	// Should not panic
	forwarder.ForwardEvent(event)
}

func TestDaemonEventForwarder_ForwardEvent_PointerEvent(t *testing.T) {
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

	forwarder := NewDaemonEventForwarder(hub)

	// Forward a pointer to event (testing reflection handling)
	event := &MockDaemonEvent{
		Type: "stats:updated",
		Data: map[string]interface{}{"matches": 10},
	}
	forwarder.ForwardEvent(event)

	// Read the message
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read message: %v", err)
	}

	var received Event
	if err := json.Unmarshal(message, &received); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if received.Type != "stats:updated" {
		t.Errorf("Expected type 'stats:updated', got '%s'", received.Type)
	}
}

func TestDaemonEventForwarder_ForwardEvent_NonStructType(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	forwarder := NewDaemonEventForwarder(hub)

	// Forward a non-struct type (should log warning and return)
	forwarder.ForwardEvent("not a struct")
	forwarder.ForwardEvent(123)
	forwarder.ForwardEvent(nil)

	// Should not panic
}

func TestGetFieldByName_ValidStruct(t *testing.T) {
	event := MockDaemonEvent{
		Type: "test:event",
		Data: map[string]interface{}{"key": "value"},
	}

	// Test getting Type field
	typeVal, ok := getFieldByName(event, "Type")
	if !ok {
		t.Error("Expected to find Type field")
	}
	if typeVal != "test:event" {
		t.Errorf("Expected 'test:event', got '%v'", typeVal)
	}

	// Test getting Data field
	dataVal, ok := getFieldByName(event, "Data")
	if !ok {
		t.Error("Expected to find Data field")
	}
	if dataVal == nil {
		t.Error("Expected Data to not be nil")
	}
}

func TestGetFieldByName_PointerStruct(t *testing.T) {
	event := &MockDaemonEvent{
		Type: "test:event",
		Data: map[string]interface{}{"key": "value"},
	}

	// Test getting Type field from pointer
	typeVal, ok := getFieldByName(event, "Type")
	if !ok {
		t.Error("Expected to find Type field from pointer")
	}
	if typeVal != "test:event" {
		t.Errorf("Expected 'test:event', got '%v'", typeVal)
	}
}

func TestGetFieldByName_NonExistentField(t *testing.T) {
	event := MockDaemonEvent{
		Type: "test:event",
		Data: nil,
	}

	_, ok := getFieldByName(event, "NonExistent")
	if ok {
		t.Error("Expected not to find NonExistent field")
	}
}

func TestGetFieldByName_NonStruct(t *testing.T) {
	// Test with string
	_, ok := getFieldByName("not a struct", "Type")
	if ok {
		t.Error("Expected not to find field in string")
	}

	// Test with int
	_, ok = getFieldByName(123, "Type")
	if ok {
		t.Error("Expected not to find field in int")
	}

	// Test with nil
	_, ok = getFieldByName(nil, "Type")
	if ok {
		t.Error("Expected not to find field in nil")
	}
}

func TestGetFieldByName_NestedData(t *testing.T) {
	type NestedStruct struct {
		ID   int
		Name string
	}

	type EventWithNested struct {
		Type string
		Data NestedStruct
	}

	event := EventWithNested{
		Type: "nested:event",
		Data: NestedStruct{
			ID:   42,
			Name: "test",
		},
	}

	dataVal, ok := getFieldByName(event, "Data")
	if !ok {
		t.Error("Expected to find Data field")
	}

	nested, ok := dataVal.(NestedStruct)
	if !ok {
		t.Fatal("Expected Data to be NestedStruct")
	}

	if nested.ID != 42 {
		t.Errorf("Expected ID=42, got %d", nested.ID)
	}
}

func TestDaemonEventForwarder_IntegrationWithRealDaemonEventTypes(t *testing.T) {
	// This test simulates the full flow from daemon event to WebSocket client
	hub := NewHub()
	go hub.Run()

	// Create test server
	server := httptest.NewServer(http.HandlerFunc(hub.ServeWs))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	// Connect a client (simulating the frontend)
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Give time for registration
	time.Sleep(50 * time.Millisecond)

	forwarder := NewDaemonEventForwarder(hub)

	// Test quest:updated event (the original failing case)
	questEvent := MockDaemonEvent{
		Type: "quest:updated",
		Data: map[string]interface{}{
			"active_quests": []map[string]interface{}{
				{"id": "quest1", "progress": 4, "goal": 30},
				{"id": "quest2", "progress": 7, "goal": 15},
			},
			"daily_wins":  6,
			"weekly_wins": 15,
		},
	}
	forwarder.ForwardEvent(questEvent)

	// Verify client receives the event
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, message, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("Failed to read quest:updated message: %v", err)
	}

	var received Event
	if err := json.Unmarshal(message, &received); err != nil {
		t.Fatalf("Failed to unmarshal message: %v", err)
	}

	if received.Type != "quest:updated" {
		t.Errorf("Expected type 'quest:updated', got '%s'", received.Type)
	}

	// Verify the data structure
	dataMap, ok := received.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected Data to be a map")
	}

	if dataMap["daily_wins"] == nil {
		t.Error("Expected daily_wins in data")
	}

	if dataMap["weekly_wins"] == nil {
		t.Error("Expected weekly_wins in data")
	}
}

func TestDaemonEventForwarder_ForwardEvent_SequentialEvents(t *testing.T) {
	// Test that multiple events are forwarded in order
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

	forwarder := NewDaemonEventForwarder(hub)

	// Send multiple events
	events := []MockDaemonEvent{
		{Type: "stats:updated", Data: map[string]interface{}{"order": 1}},
		{Type: "quest:updated", Data: map[string]interface{}{"order": 2}},
		{Type: "deck:updated", Data: map[string]interface{}{"order": 3}},
	}

	for _, event := range events {
		forwarder.ForwardEvent(event)
		// Small delay to prevent message buffering issues in WebSocket
		time.Sleep(10 * time.Millisecond)
	}

	// Verify all events are received
	for i := 0; i < len(events); i++ {
		conn.SetReadDeadline(time.Now().Add(1 * time.Second))
		_, message, err := conn.ReadMessage()
		if err != nil {
			t.Fatalf("Failed to read message %d: %v", i, err)
		}

		var received Event
		if err := json.Unmarshal(message, &received); err != nil {
			t.Fatalf("Failed to unmarshal message %d: %v", i, err)
		}

		if received.Type != events[i].Type {
			t.Errorf("Event %d: expected type '%s', got '%s'", i, events[i].Type, received.Type)
		}
	}
}
