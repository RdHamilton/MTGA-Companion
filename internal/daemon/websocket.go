package daemon

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Event represents a WebSocket event to be broadcast to clients.
type Event struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// WebSocketServer manages WebSocket connections and event broadcasting.
type WebSocketServer struct {
	port      int
	clients   map[*websocket.Conn]bool
	clientsMu sync.RWMutex
	broadcast chan Event
	upgrader  websocket.Upgrader
	server    *http.Server
}

// NewWebSocketServer creates a new WebSocket server.
func NewWebSocketServer(port int) *WebSocketServer {
	return &WebSocketServer{
		port:      port,
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan Event, 100),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for development
				// TODO: Make this configurable for production
				return true
			},
		},
	}
}

// Start starts the WebSocket server.
func (s *WebSocketServer) Start() error {
	// Start broadcast handler
	go s.handleBroadcasts()

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleWebSocket)
	mux.HandleFunc("/status", s.handleStatus)

	s.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	log.Printf("WebSocket server starting on port %d", s.port)
	return s.server.ListenAndServe()
}

// Stop gracefully stops the WebSocket server.
func (s *WebSocketServer) Stop() error {
	log.Println("Stopping WebSocket server...")

	// Close all client connections
	s.clientsMu.Lock()
	for client := range s.clients {
		if err := client.Close(); err != nil {
			log.Printf("Error closing client connection: %v", err)
		}
	}
	s.clients = make(map[*websocket.Conn]bool)
	s.clientsMu.Unlock()

	// Close broadcast channel
	close(s.broadcast)

	// Shutdown HTTP server
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// Broadcast sends an event to all connected clients.
func (s *WebSocketServer) Broadcast(event Event) {
	event.Timestamp = time.Now()
	select {
	case s.broadcast <- event:
	default:
		log.Println("Warning: Broadcast channel full, dropping event")
	}
}

// ClientCount returns the number of connected clients.
func (s *WebSocketServer) ClientCount() int {
	s.clientsMu.RLock()
	defer s.clientsMu.RUnlock()
	return len(s.clients)
}

// handleWebSocket handles WebSocket upgrade requests.
func (s *WebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Register client
	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	log.Printf("Client connected (total: %d)", s.ClientCount())

	// Send welcome message
	welcome := Event{
		Type: "daemon:connected",
		Data: map[string]interface{}{
			"message": "Connected to MTGA Companion daemon",
		},
		Timestamp: time.Now(),
	}
	if err := conn.WriteJSON(welcome); err != nil {
		log.Printf("Error sending welcome message: %v", err)
	}

	// Handle client messages
	go s.handleClient(conn)
}

// handleClient handles messages from a specific client.
func (s *WebSocketServer) handleClient(conn *websocket.Conn) {
	defer func() {
		// Unregister client
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()

		if err := conn.Close(); err != nil {
			log.Printf("Error closing client connection: %v", err)
		}
		log.Printf("Client disconnected (total: %d)", s.ClientCount())
	}()

	for {
		var msg map[string]interface{}
		if err := conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle client messages
		msgType, ok := msg["type"].(string)
		if !ok {
			continue
		}

		switch msgType {
		case "ping":
			// Respond with pong
			pong := Event{
				Type:      "pong",
				Data:      map[string]interface{}{},
				Timestamp: time.Now(),
			}
			if err := conn.WriteJSON(pong); err != nil {
				log.Printf("Error sending pong: %v", err)
				return
			}
		case "subscribe":
			// TODO: Implement selective event subscription
			log.Printf("Client subscribed to events: %v", msg["events"])
		}
	}
}

// handleBroadcasts handles broadcasting events to all clients.
func (s *WebSocketServer) handleBroadcasts() {
	for event := range s.broadcast {
		s.clientsMu.RLock()
		for client := range s.clients {
			if err := client.WriteJSON(event); err != nil {
				log.Printf("Error broadcasting to client: %v", err)
				if err := client.Close(); err != nil {
					log.Printf("Error closing client after broadcast error: %v", err)
				}
				s.clientsMu.Lock()
				delete(s.clients, client)
				s.clientsMu.Unlock()
			}
		}
		s.clientsMu.RUnlock()
	}
}

// handleStatus handles HTTP status requests.
func (s *WebSocketServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":  "running",
		"clients": s.ClientCount(),
		"time":    time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Printf("Error encoding status response: %v", err)
	}
}
