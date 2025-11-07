package logreader

import (
	"testing"
	"time"
)

func TestNewNotifier(t *testing.T) {
	t.Run("DefaultConfig", func(t *testing.T) {
		config := DefaultNotificationConfig()
		notifier := NewNotifier(config)
		if notifier == nil {
			t.Fatal("NewNotifier() returned nil")
		}
	})

	t.Run("NilConfig", func(t *testing.T) {
		notifier := NewNotifier(nil)
		if notifier == nil {
			t.Fatal("NewNotifier(nil) returned nil")
		}
	})
}

func TestNotifier_Notify(t *testing.T) {
	config := DefaultNotificationConfig()
	config.EnableConsole = false // Disable console for testing
	notifier := NewNotifier(config)

	var receivedEvent *Event
	notifier.AddHandler(func(event *Event) {
		receivedEvent = event
	})

	// Send notification
	event := &Event{
		Type:       EventMatchComplete,
		Importance: ImportanceMedium,
		Message:    "Test match completed",
	}

	notifier.Notify(event)

	if receivedEvent == nil {
		t.Fatal("Event not received by handler")
	}

	if receivedEvent.Message != event.Message {
		t.Errorf("Expected message %q, got %q", event.Message, receivedEvent.Message)
	}
}

func TestNotifier_EventFiltering(t *testing.T) {
	config := DefaultNotificationConfig()
	config.EnableConsole = false
	config.EnabledEvents = []EventType{EventMatchComplete}
	notifier := NewNotifier(config)

	var receivedCount int
	notifier.AddHandler(func(event *Event) {
		receivedCount++
	})

	// Send enabled event
	notifier.Notify(&Event{
		Type:       EventMatchComplete,
		Importance: ImportanceMedium,
		Message:    "Match completed",
	})

	// Send disabled event
	notifier.Notify(&Event{
		Type:       EventRankChange,
		Importance: ImportanceMedium,
		Message:    "Rank changed",
	})

	// Should only receive one event
	if receivedCount != 1 {
		t.Errorf("Expected 1 event, got %d", receivedCount)
	}
}

func TestNotifier_ImportanceFiltering(t *testing.T) {
	config := DefaultNotificationConfig()
	config.EnableConsole = false
	config.MinImportance = ImportanceHigh
	notifier := NewNotifier(config)

	var receivedCount int
	notifier.AddHandler(func(event *Event) {
		receivedCount++
	})

	// Send low importance event (should be filtered)
	notifier.Notify(&Event{
		Type:       EventMatchComplete,
		Importance: ImportanceLow,
		Message:    "Low importance",
	})

	// Send high importance event (should pass)
	notifier.Notify(&Event{
		Type:       EventMatchComplete,
		Importance: ImportanceHigh,
		Message:    "High importance",
	})

	// Should only receive high importance event
	if receivedCount != 1 {
		t.Errorf("Expected 1 event, got %d", receivedCount)
	}
}

func TestNotifier_RateLimit(t *testing.T) {
	config := DefaultNotificationConfig()
	config.EnableConsole = false
	config.RateLimit = 200 * time.Millisecond
	notifier := NewNotifier(config)

	var receivedCount int
	notifier.AddHandler(func(event *Event) {
		receivedCount++
	})

	event := &Event{
		Type:       EventMatchComplete,
		Importance: ImportanceMedium,
		Message:    "Test",
	}

	// Send first event
	notifier.Notify(event)

	// Send second event immediately (should be rate limited)
	notifier.Notify(event)

	if receivedCount != 1 {
		t.Errorf("Expected 1 event due to rate limit, got %d", receivedCount)
	}

	// Wait for rate limit to pass
	time.Sleep(250 * time.Millisecond)

	// Send third event (should pass)
	notifier.Notify(event)

	if receivedCount != 2 {
		t.Errorf("Expected 2 events after rate limit, got %d", receivedCount)
	}
}

func TestNotifier_History(t *testing.T) {
	config := DefaultNotificationConfig()
	config.EnableConsole = false
	config.MaxHistory = 5
	config.RateLimit = 10 * time.Millisecond // Short rate limit for testing
	notifier := NewNotifier(config)

	// Send multiple events of different types to avoid rate limiting
	eventTypes := []EventType{
		EventMatchComplete,
		EventRankChange,
		EventDraftComplete,
		EventMilestone,
		EventCollection,
	}

	for i := 0; i < 10; i++ {
		notifier.Notify(&Event{
			Type:       eventTypes[i%len(eventTypes)],
			Importance: ImportanceMedium,
			Message:    "Test",
		})
		time.Sleep(15 * time.Millisecond) // Wait for rate limit
	}

	history := notifier.GetHistory()

	// Should only keep max history
	if len(history) != config.MaxHistory {
		t.Errorf("Expected history length %d, got %d", config.MaxHistory, len(history))
	}
}

func TestNotifier_ProcessEntry(t *testing.T) {
	config := DefaultNotificationConfig()
	config.EnableConsole = false
	notifier := NewNotifier(config)

	var receivedEvents []*Event
	notifier.AddHandler(func(event *Event) {
		receivedEvents = append(receivedEvents, event)
	})

	// Test match completion
	entry := &LogEntry{
		IsJSON: true,
		JSON: map[string]interface{}{
			"CurrentEventState": "MatchCompleted",
			"outcome":           "win",
		},
	}

	notifier.ProcessEntry(entry)

	// Wait a bit for processing
	time.Sleep(50 * time.Millisecond)

	if len(receivedEvents) < 1 {
		t.Errorf("Expected at least 1 event, got %d", len(receivedEvents))
	}

	if len(receivedEvents) > 0 {
		if receivedEvents[0].Type != EventMatchComplete {
			t.Errorf("Expected EventMatchComplete, got %v", receivedEvents[0].Type)
		}
	}
}

func TestNotifier_ProcessEntry_RankChange(t *testing.T) {
	config := DefaultNotificationConfig()
	config.EnableConsole = false
	notifier := NewNotifier(config)

	var receivedEvent *Event
	notifier.AddHandler(func(event *Event) {
		receivedEvent = event
	})

	// Test rank change
	entry := &LogEntry{
		IsJSON: true,
		JSON: map[string]interface{}{
			"rankClass": "Bronze",
			"rankTier":  float64(3),
		},
	}

	notifier.ProcessEntry(entry)

	time.Sleep(50 * time.Millisecond)

	if receivedEvent == nil {
		t.Fatal("Expected rank change event, got nil")
	}

	if receivedEvent.Type != EventRankChange {
		t.Errorf("Expected EventRankChange, got %v", receivedEvent.Type)
	}

	if receivedEvent.Importance != ImportanceHigh {
		t.Errorf("Expected ImportanceHigh for rank change, got %v", receivedEvent.Importance)
	}
}

func TestNotifier_ProcessEntry_DraftComplete(t *testing.T) {
	config := DefaultNotificationConfig()
	config.EnableConsole = false
	notifier := NewNotifier(config)

	var receivedEvent *Event
	notifier.AddHandler(func(event *Event) {
		receivedEvent = event
	})

	// Test draft completion
	entry := &LogEntry{
		IsJSON: true,
		JSON: map[string]interface{}{
			"draftStatus": "Complete",
		},
	}

	notifier.ProcessEntry(entry)

	time.Sleep(50 * time.Millisecond)

	if receivedEvent == nil {
		t.Fatal("Expected draft complete event, got nil")
	}

	if receivedEvent.Type != EventDraftComplete {
		t.Errorf("Expected EventDraftComplete, got %v", receivedEvent.Type)
	}
}

func TestNotifier_MultipleHandlers(t *testing.T) {
	config := DefaultNotificationConfig()
	config.EnableConsole = false
	notifier := NewNotifier(config)

	var handler1Called, handler2Called bool

	notifier.AddHandler(func(event *Event) {
		handler1Called = true
	})

	notifier.AddHandler(func(event *Event) {
		handler2Called = true
	})

	notifier.Notify(&Event{
		Type:       EventMatchComplete,
		Importance: ImportanceMedium,
		Message:    "Test",
	})

	if !handler1Called {
		t.Error("Handler 1 was not called")
	}

	if !handler2Called {
		t.Error("Handler 2 was not called")
	}
}
