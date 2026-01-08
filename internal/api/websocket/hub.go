package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for development
		// TODO: Restrict origins in production
		return true
	},
}

// Event represents a WebSocket event to be broadcast.
type Event struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Client represents a WebSocket client connection.
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from clients (for future use).
	broadcast chan []byte

	// Register requests from clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client

	// Signal to stop the hub.
	done chan struct{}

	// Ensures Stop() is idempotent.
	stopOnce sync.Once

	// Indicates hub has been stopped.
	stopped bool

	// Mutex for thread-safe client operations.
	mu sync.RWMutex
}

// NewHub creates a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
	}
}

// Run starts the hub's main loop.
func (h *Hub) Run() {
	for {
		select {
		case <-h.done:
			// Clean up all clients before exiting
			h.mu.Lock()
			h.stopped = true
			for client := range h.clients {
				// Delete from map first to prevent other paths from accessing
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Println("WebSocket hub stopped")
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					h.mu.RUnlock()
					h.mu.Lock()
					close(client.send)
					delete(h.clients, client)
					h.mu.Unlock()
					h.mu.RLock()
				}
			}
			h.mu.RUnlock()
		}
	}
}

// BroadcastEvent broadcasts an event to all connected clients.
// Returns false if the hub has been stopped.
func (h *Hub) BroadcastEvent(event Event) bool {
	h.mu.RLock()
	stopped := h.stopped
	h.mu.RUnlock()

	if stopped {
		return false
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling WebSocket event: %v", err)
		return false
	}

	// Use non-blocking send to avoid blocking if hub is stopping
	select {
	case h.broadcast <- data:
		return true
	case <-h.done:
		return false
	}
}

// ClientCount returns the number of connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// Stop gracefully stops the hub and cleans up all client connections.
// Safe to call multiple times - subsequent calls are no-ops.
func (h *Hub) Stop() {
	h.stopOnce.Do(func() {
		close(h.done)
	})
}

// IsStopped returns true if the hub has been stopped.
func (h *Hub) IsStopped() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.stopped
}

// ServeWs handles WebSocket requests from clients.
// Returns immediately if the hub has been stopped.
func (h *Hub) ServeWs(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	stopped := h.stopped
	h.mu.RUnlock()

	if stopped {
		http.Error(w, "WebSocket hub is not running", http.StatusServiceUnavailable)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	client := &Client{
		hub:  h,
		conn: conn,
		send: make(chan []byte, 256),
	}

	// Use non-blocking send to avoid blocking if hub stops during registration
	select {
	case h.register <- client:
		// Start goroutines for reading and writing
		go client.writePump()
		go client.readPump()
	case <-h.done:
		// Hub stopped, close connection
		if err := conn.Close(); err != nil {
			log.Printf("WebSocket close error: %v", err)
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		// Use non-blocking send to avoid blocking if hub has stopped
		select {
		case c.hub.unregister <- c:
		case <-c.hub.done:
			// Hub already stopped, client cleanup handled there
		}
		if err := c.conn.Close(); err != nil {
			log.Printf("WebSocket close error: %v", err)
		}
	}()

	c.conn.SetReadLimit(maxMessageSize)
	if err := c.conn.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Printf("WebSocket SetReadDeadline error: %v", err)
		return
	}
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}
		// For now, we don't process incoming messages from clients
		// This could be extended for bidirectional communication
	}
}

// writePump pumps messages from the hub to the WebSocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if err := c.conn.Close(); err != nil {
			log.Printf("WebSocket close error: %v", err)
		}
	}()

	for {
		select {
		case message, ok := <-c.send:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("WebSocket SetWriteDeadline error: %v", err)
				return
			}
			if !ok {
				// The hub closed the channel.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

			// Add queued messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				if _, err := w.Write([]byte{'\n'}); err != nil {
					log.Printf("WebSocket write error: %v", err)
					return
				}
				if _, err := w.Write(<-c.send); err != nil {
					log.Printf("WebSocket write error: %v", err)
					return
				}
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			if err := c.conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				log.Printf("WebSocket SetWriteDeadline error: %v", err)
				return
			}
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
