package websocket

import (
	"log"

	"github.com/ramonehamilton/MTGA-Companion/internal/events"
)

// WebSocketObserver forwards events to WebSocket clients.
// Implements the events.Observer interface to listen for domain events
// and broadcast them to all connected WebSocket clients.
type WebSocketObserver struct {
	name string
	hub  *Hub
}

// NewWebSocketObserver creates a new observer that forwards events to WebSocket clients.
func NewWebSocketObserver(hub *Hub) *WebSocketObserver {
	return &WebSocketObserver{
		name: "WebSocketObserver",
		hub:  hub,
	}
}

// OnEvent forwards the event to all connected WebSocket clients.
func (o *WebSocketObserver) OnEvent(event events.Event) error {
	if o.hub == nil {
		log.Printf("[%s] Cannot emit event %s: hub is nil", o.name, event.Type)
		return nil
	}

	// Convert the event to WebSocket format
	wsEvent := Event{
		Type: event.Type,
		Data: event.Data,
	}

	// If TypedData is available, prefer it over the untyped map
	if event.TypedData != nil {
		wsEvent.Data = event.TypedData
	}

	// Broadcast to all connected clients
	o.hub.BroadcastEvent(wsEvent)
	log.Printf("[%s] Broadcast event to %d clients: %s", o.name, o.hub.ClientCount(), event.Type)

	return nil
}

// GetName returns the observer's name.
func (o *WebSocketObserver) GetName() string {
	return o.name
}

// ShouldHandle returns true for all events (forwards everything to WebSocket clients).
func (o *WebSocketObserver) ShouldHandle(eventType string) bool {
	// Forward all events to WebSocket clients
	return true
}

// Ensure WebSocketObserver implements the Observer interface.
var _ events.Observer = (*WebSocketObserver)(nil)
