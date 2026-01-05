package daemon

import (
	"sync"
	"testing"
	"time"
)

// mockEventForwarder is a test implementation of EventForwarder.
type mockEventForwarder struct {
	mu     sync.Mutex
	events []interface{}
}

func newMockEventForwarder() *mockEventForwarder {
	return &mockEventForwarder{
		events: make([]interface{}, 0),
	}
}

func (m *mockEventForwarder) ForwardEvent(event interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
}

func (m *mockEventForwarder) GetEvents() []interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]interface{}{}, m.events...)
}

func (m *mockEventForwarder) EventCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.events)
}

func TestService_RegisterEventForwarder(t *testing.T) {
	// Create a minimal service for testing
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	// Register a forwarder
	forwarder := newMockEventForwarder()
	service.RegisterEventForwarder(forwarder)

	// Verify forwarder was registered
	service.forwardersMu.RLock()
	count := len(service.forwarders)
	service.forwardersMu.RUnlock()

	if count != 1 {
		t.Errorf("Expected 1 forwarder, got %d", count)
	}
}

func TestService_RegisterEventForwarder_Multiple(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	// Register multiple forwarders
	forwarder1 := newMockEventForwarder()
	forwarder2 := newMockEventForwarder()
	forwarder3 := newMockEventForwarder()

	service.RegisterEventForwarder(forwarder1)
	service.RegisterEventForwarder(forwarder2)
	service.RegisterEventForwarder(forwarder3)

	service.forwardersMu.RLock()
	count := len(service.forwarders)
	service.forwardersMu.RUnlock()

	if count != 3 {
		t.Errorf("Expected 3 forwarders, got %d", count)
	}
}

func TestService_broadcastEvent_NoForwarders(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	event := Event{
		Type: "test:event",
		Data: map[string]interface{}{"key": "value"},
	}

	// Should not panic with no forwarders
	service.broadcastEvent(event)
}

func TestService_broadcastEvent_WithForwarder(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	forwarder := newMockEventForwarder()
	service.RegisterEventForwarder(forwarder)

	event := Event{
		Type: "stats:updated",
		Data: map[string]interface{}{
			"matches": 10,
			"games":   25,
		},
	}

	service.broadcastEvent(event)

	// Give time for processing
	time.Sleep(10 * time.Millisecond)

	if forwarder.EventCount() != 1 {
		t.Errorf("Expected 1 event forwarded, got %d", forwarder.EventCount())
	}

	events := forwarder.GetEvents()
	forwarded, ok := events[0].(Event)
	if !ok {
		t.Fatal("Expected Event type")
	}

	if forwarded.Type != "stats:updated" {
		t.Errorf("Expected type stats:updated, got %s", forwarded.Type)
	}
}

func TestService_broadcastEvent_MultipleForwarders(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	forwarder1 := newMockEventForwarder()
	forwarder2 := newMockEventForwarder()
	forwarder3 := newMockEventForwarder()

	service.RegisterEventForwarder(forwarder1)
	service.RegisterEventForwarder(forwarder2)
	service.RegisterEventForwarder(forwarder3)

	event := Event{
		Type: "quest:updated",
		Data: map[string]interface{}{"quests": 5},
	}

	service.broadcastEvent(event)

	// Give time for processing
	time.Sleep(10 * time.Millisecond)

	// All forwarders should receive the event
	if forwarder1.EventCount() != 1 {
		t.Errorf("Forwarder 1: expected 1 event, got %d", forwarder1.EventCount())
	}
	if forwarder2.EventCount() != 1 {
		t.Errorf("Forwarder 2: expected 1 event, got %d", forwarder2.EventCount())
	}
	if forwarder3.EventCount() != 1 {
		t.Errorf("Forwarder 3: expected 1 event, got %d", forwarder3.EventCount())
	}
}

func TestService_broadcastEvent_AllDaemonEventTypes(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	forwarder := newMockEventForwarder()
	service.RegisterEventForwarder(forwarder)

	// Test all daemon event types
	eventTypes := []string{
		"daemon:status",
		"daemon:error",
		"stats:updated",
		"deck:updated",
		"rank:updated",
		"quest:updated",
		"draft:updated",
		"replay:started",
		"replay:paused",
		"replay:resumed",
		"replay:completed",
		"replay:progress",
		"replay:error",
		"replay:draft_detected",
	}

	for _, eventType := range eventTypes {
		event := Event{
			Type: eventType,
			Data: map[string]interface{}{"test": true},
		}
		service.broadcastEvent(event)
	}

	// Give time for processing
	time.Sleep(10 * time.Millisecond)

	if forwarder.EventCount() != len(eventTypes) {
		t.Errorf("Expected %d events forwarded, got %d", len(eventTypes), forwarder.EventCount())
	}
}

func TestService_broadcastEvent_ConcurrentSafety(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	forwarder := newMockEventForwarder()
	service.RegisterEventForwarder(forwarder)

	// Broadcast events concurrently
	var wg sync.WaitGroup
	numGoroutines := 10
	eventsPerGoroutine := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < eventsPerGoroutine; j++ {
				event := Event{
					Type: "concurrent:test",
					Data: map[string]interface{}{
						"goroutine": id,
						"iteration": j,
					},
				}
				service.broadcastEvent(event)
			}
		}(i)
	}

	wg.Wait()

	// Give time for all events to be processed
	time.Sleep(50 * time.Millisecond)

	expectedEvents := numGoroutines * eventsPerGoroutine
	if forwarder.EventCount() != expectedEvents {
		t.Errorf("Expected %d events, got %d", expectedEvents, forwarder.EventCount())
	}
}

func TestService_RegisterEventForwarder_ConcurrentSafety(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	// Register forwarders concurrently
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			forwarder := newMockEventForwarder()
			service.RegisterEventForwarder(forwarder)
		}()
	}

	wg.Wait()

	service.forwardersMu.RLock()
	count := len(service.forwarders)
	service.forwardersMu.RUnlock()

	if count != numGoroutines {
		t.Errorf("Expected %d forwarders, got %d", numGoroutines, count)
	}
}

func TestEventForwarder_Interface(t *testing.T) {
	// Verify mockEventForwarder implements EventForwarder
	var _ EventForwarder = (*mockEventForwarder)(nil)
}
