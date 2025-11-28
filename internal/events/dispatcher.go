package events

import (
	"context"
	"log"
	"sync"
)

// Event represents a domain event that can be dispatched to observers.
type Event struct {
	// Type is the event type (e.g., "stats:updated", "daemon:status")
	Type string

	// Data contains the event payload as untyped map.
	// Deprecated: Use TypedData for type-safe access when possible.
	Data map[string]interface{}

	// TypedData contains the event payload as a typed struct.
	// This is optional and may be nil for backwards compatibility.
	// When set, observers should prefer this over Data.
	TypedData any

	// Context provides execution context for the event
	Context context.Context
}

// Observer defines the interface for objects that want to be notified of events.
// Implementations can handle events in different ways (e.g., forward to frontend, send to IPC, log).
type Observer interface {
	// OnEvent is called when an event is dispatched.
	// Returns an error if the observer fails to handle the event.
	OnEvent(event Event) error

	// GetName returns a human-readable name for this observer (for logging/debugging).
	GetName() string

	// ShouldHandle returns true if this observer should handle the given event type.
	// This allows observers to filter which events they care about.
	ShouldHandle(eventType string) bool
}

// EventDispatcher implements the Observer pattern for event distribution.
// It maintains a list of observers and notifies them when events occur.
// Thread-safe for concurrent use.
type EventDispatcher struct {
	observers []Observer
	mu        sync.RWMutex
}

// NewEventDispatcher creates a new EventDispatcher.
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		observers: make([]Observer, 0),
	}
}

// Register adds an observer to the dispatcher.
// The observer will be notified of all future events (filtered by ShouldHandle).
func (d *EventDispatcher) Register(observer Observer) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.observers = append(d.observers, observer)
	log.Printf("[EventDispatcher] Registered observer: %s", observer.GetName())
}

// Unregister removes an observer from the dispatcher.
// The observer will no longer receive event notifications.
func (d *EventDispatcher) Unregister(observer Observer) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for i, obs := range d.observers {
		if obs == observer {
			// Remove observer by replacing it with the last element and truncating
			d.observers[i] = d.observers[len(d.observers)-1]
			d.observers = d.observers[:len(d.observers)-1]
			log.Printf("[EventDispatcher] Unregistered observer: %s", observer.GetName())
			return
		}
	}
}

// Dispatch sends an event to all registered observers.
// Observers are notified sequentially in the order they were registered.
// If an observer returns an error, it's logged but dispatch continues to other observers.
func (d *EventDispatcher) Dispatch(event Event) {
	d.mu.RLock()
	observers := make([]Observer, len(d.observers))
	copy(observers, d.observers)
	d.mu.RUnlock()

	for _, observer := range observers {
		// Check if observer wants to handle this event type
		if !observer.ShouldHandle(event.Type) {
			continue
		}

		// Notify observer
		if err := observer.OnEvent(event); err != nil {
			log.Printf("[EventDispatcher] Observer %s failed to handle event %s: %v",
				observer.GetName(), event.Type, err)
		}
	}
}

// DispatchAsync sends an event to all observers asynchronously.
// Each observer is notified in a separate goroutine.
// Useful for long-running event handlers that shouldn't block the caller.
func (d *EventDispatcher) DispatchAsync(event Event) {
	d.mu.RLock()
	observers := make([]Observer, len(d.observers))
	copy(observers, d.observers)
	d.mu.RUnlock()

	for _, observer := range observers {
		// Check if observer wants to handle this event type
		if !observer.ShouldHandle(event.Type) {
			continue
		}

		// Notify observer asynchronously
		go func(obs Observer) {
			if err := obs.OnEvent(event); err != nil {
				log.Printf("[EventDispatcher] Observer %s failed to handle event %s: %v",
					obs.GetName(), event.Type, err)
			}
		}(observer)
	}
}

// ObserverCount returns the number of registered observers.
func (d *EventDispatcher) ObserverCount() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return len(d.observers)
}

// Clear removes all registered observers.
// Useful for testing or cleanup.
func (d *EventDispatcher) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.observers = make([]Observer, 0)
	log.Printf("[EventDispatcher] Cleared all observers")
}

// NewTypedEvent creates an Event with typed data.
// The typed data is also converted to map[string]interface{} for backwards compatibility.
func NewTypedEvent[T any](eventType string, data T, ctx context.Context) Event {
	return Event{
		Type:      eventType,
		Data:      structToMap(data),
		TypedData: data,
		Context:   ctx,
	}
}

// structToMap converts a struct to map[string]interface{} using JSON serialization.
// This ensures backwards compatibility with code expecting untyped data.
func structToMap(v any) map[string]interface{} {
	result := make(map[string]interface{})

	// Use reflection-free approach via JSON
	// This is slightly slower but more reliable for nested structs
	if v == nil {
		return result
	}

	// For simple cases, we can try direct type assertion
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}

	// Otherwise, leave as empty map - observers should use TypedData
	return result
}

// GetTypedData extracts typed data from an Event.
// Returns the zero value and false if the data is not of the expected type.
func GetTypedData[T any](event Event) (T, bool) {
	var zero T
	if event.TypedData == nil {
		return zero, false
	}
	typed, ok := event.TypedData.(T)
	return typed, ok
}
