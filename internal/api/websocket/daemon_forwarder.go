package websocket

import (
	"log"
	"reflect"
)

// DaemonEventForwarder forwards daemon events to the API server's WebSocket hub.
// This bridges the gap between the daemon (which runs on a separate port)
// and the frontend (which connects to the API server's WebSocket).
type DaemonEventForwarder struct {
	hub *Hub
}

// NewDaemonEventForwarder creates a new forwarder that sends events to the given hub.
func NewDaemonEventForwarder(hub *Hub) *DaemonEventForwarder {
	return &DaemonEventForwarder{hub: hub}
}

// ForwardEvent forwards a daemon event to all connected WebSocket clients.
// The event parameter should be a daemon.Event but we accept interface{}
// to implement the daemon.EventForwarder interface without import cycles.
func (f *DaemonEventForwarder) ForwardEvent(event interface{}) {
	var eventType string
	var eventData interface{}

	// Extract Type and Data fields using reflection
	// This avoids import cycles with the daemon package
	if v, ok := getFieldByName(event, "Type"); ok {
		eventType, _ = v.(string)
	}
	if v, ok := getFieldByName(event, "Data"); ok {
		eventData = v
	}

	if eventType == "" {
		log.Printf("[DaemonForwarder] Warning: Could not extract event type from %T", event)
		return
	}

	wsEvent := Event{
		Type: eventType,
		Data: eventData,
	}
	f.hub.BroadcastEvent(wsEvent)
	log.Printf("[DaemonForwarder] Forwarded event %s to %d clients", wsEvent.Type, f.hub.ClientCount())
}

// getFieldByName extracts a field value from a struct by name using reflection.
func getFieldByName(obj interface{}, fieldName string) (interface{}, bool) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil, false
	}
	f := v.FieldByName(fieldName)
	if !f.IsValid() {
		return nil, false
	}
	return f.Interface(), true
}
