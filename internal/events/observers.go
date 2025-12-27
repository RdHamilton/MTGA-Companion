package events

import (
	"log"
)

// IPCObserver forwards events to the IPC daemon.
// Implements the Observer interface to listen for domain events
// and send them to the daemon via IPC.
type IPCObserver struct {
	name      string
	ipcClient IPCClient
}

// IPCClient defines the interface for sending events to the daemon.
// This allows the observer to be decoupled from the specific IPC implementation.
type IPCClient interface {
	// Emit sends an event to the daemon
	Emit(eventType string, data map[string]interface{})

	// IsConnected returns true if connected to the daemon
	IsConnected() bool
}

// NewIPCObserver creates a new observer that forwards events to the IPC daemon.
func NewIPCObserver(ipcClient IPCClient) *IPCObserver {
	return &IPCObserver{
		name:      "IPCObserver",
		ipcClient: ipcClient,
	}
}

// OnEvent forwards the event to the IPC daemon.
func (o *IPCObserver) OnEvent(event Event) error {
	if o.ipcClient == nil || !o.ipcClient.IsConnected() {
		// Don't emit if not connected - this is expected behavior
		return nil
	}

	// Emit event to daemon via IPC
	o.ipcClient.Emit(event.Type, event.Data)
	log.Printf("[%s] Emitted event to daemon: %s", o.name, event.Type)
	return nil
}

// GetName returns the observer's name.
func (o *IPCObserver) GetName() string {
	return o.name
}

// ShouldHandle returns true for events that should be sent to the daemon.
// Filters out daemon-originated events to avoid loops.
func (o *IPCObserver) ShouldHandle(eventType string) bool {
	// Don't send daemon events back to daemon (avoid loops)
	switch eventType {
	case "daemon:status", "daemon:error", "daemon:connected":
		return false
	default:
		return true
	}
}

// LoggingObserver logs all events for debugging purposes.
// Useful for development and troubleshooting.
type LoggingObserver struct {
	name    string
	verbose bool
}

// NewLoggingObserver creates a new observer that logs events.
func NewLoggingObserver(verbose bool) *LoggingObserver {
	return &LoggingObserver{
		name:    "LoggingObserver",
		verbose: verbose,
	}
}

// OnEvent logs the event details.
func (o *LoggingObserver) OnEvent(event Event) error {
	if o.verbose {
		log.Printf("[%s] Event: %s, Data: %v", o.name, event.Type, event.Data)
	} else {
		log.Printf("[%s] Event: %s", o.name, event.Type)
	}
	return nil
}

// GetName returns the observer's name.
func (o *LoggingObserver) GetName() string {
	return o.name
}

// ShouldHandle returns true for all events (logs everything).
func (o *LoggingObserver) ShouldHandle(eventType string) bool {
	return true
}
