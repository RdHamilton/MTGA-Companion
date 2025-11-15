package ipc

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Event represents an event received from the daemon.
type Event struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventHandler is a function that handles events.
type EventHandler func(data map[string]interface{})

// Client represents a WebSocket client for IPC with the daemon.
type Client struct {
	url         string
	conn        *websocket.Conn
	handlers    map[string][]EventHandler
	handlersMu  sync.RWMutex
	connected   bool
	connectedMu sync.RWMutex
	stopChan    chan struct{}
	reconnect   bool
}

// NewClient creates a new IPC client.
func NewClient(url string) *Client {
	return &Client{
		url:       url,
		handlers:  make(map[string][]EventHandler),
		stopChan:  make(chan struct{}),
		reconnect: true,
	}
}

// Connect establishes a connection to the daemon WebSocket server.
func (c *Client) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return err
	}

	c.conn = conn
	c.setConnected(true)

	log.Printf("Connected to daemon at %s", c.url)
	return nil
}

// Start begins listening for events from the daemon.
func (c *Client) Start() {
	go c.listen()
}

// Stop stops the client and closes the connection.
func (c *Client) Stop() {
	close(c.stopChan)
	c.reconnect = false

	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("Error closing WebSocket connection: %v", err)
		}
	}

	c.setConnected(false)
}

// On registers an event handler for a specific event type.
func (c *Client) On(eventType string, handler EventHandler) {
	c.handlersMu.Lock()
	defer c.handlersMu.Unlock()

	c.handlers[eventType] = append(c.handlers[eventType], handler)
}

// IsConnected returns whether the client is currently connected.
func (c *Client) IsConnected() bool {
	c.connectedMu.RLock()
	defer c.connectedMu.RUnlock()
	return c.connected
}

// setConnected sets the connected status.
func (c *Client) setConnected(connected bool) {
	c.connectedMu.Lock()
	defer c.connectedMu.Unlock()
	c.connected = connected
}

// listen listens for events from the daemon.
func (c *Client) listen() {
	for {
		select {
		case <-c.stopChan:
			return
		default:
			if c.conn == nil {
				if c.reconnect {
					c.attemptReconnect()
				} else {
					return
				}
				continue
			}

			var event Event
			if err := c.conn.ReadJSON(&event); err != nil {
				log.Printf("Error reading from daemon: %v", err)
				c.setConnected(false)

				if c.reconnect {
					c.attemptReconnect()
				} else {
					return
				}
				continue
			}

			// Dispatch event to handlers
			c.dispatchEvent(event)
		}
	}
}

// attemptReconnect attempts to reconnect to the daemon.
func (c *Client) attemptReconnect() {
	log.Println("Attempting to reconnect to daemon...")
	c.setConnected(false)

	// Close old connection if it exists
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("Error closing old connection: %v", err)
		}
		c.conn = nil
	}

	// Wait before reconnecting
	time.Sleep(5 * time.Second)

	// Try to connect
	if err := c.Connect(); err != nil {
		log.Printf("Reconnection failed: %v", err)
		return
	}

	log.Println("Reconnected to daemon successfully")
}

// dispatchEvent dispatches an event to registered handlers.
func (c *Client) dispatchEvent(event Event) {
	c.handlersMu.RLock()
	handlers, ok := c.handlers[event.Type]
	c.handlersMu.RUnlock()

	if !ok {
		// No handlers for this event type
		return
	}

	// Call all handlers for this event type
	for _, handler := range handlers {
		go handler(event.Data)
	}
}

// SendPing sends a ping message to the daemon.
func (c *Client) SendPing() error {
	if c.conn == nil {
		return websocket.ErrCloseSent
	}

	ping := map[string]interface{}{
		"type": "ping",
	}

	return c.conn.WriteJSON(ping)
}

// GetURL returns the WebSocket URL.
func (c *Client) GetURL() string {
	return c.url
}

// MarshalEvent marshals an event to JSON for logging/debugging.
func MarshalEvent(event Event) (string, error) {
	bytes, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
