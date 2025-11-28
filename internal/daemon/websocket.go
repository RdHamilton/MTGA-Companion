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
	port       int
	clients    map[*websocket.Conn]bool
	clientsMu  sync.RWMutex
	broadcast  chan Event
	upgrader   websocket.Upgrader
	server     *http.Server
	service    *Service // Reference to parent service for health checks
	corsConfig CORSConfig
}

// NewWebSocketServer creates a new WebSocket server.
func NewWebSocketServer(port int) *WebSocketServer {
	return NewWebSocketServerWithCORS(port, DefaultCORSConfig())
}

// NewWebSocketServerWithCORS creates a new WebSocket server with custom CORS configuration.
func NewWebSocketServerWithCORS(port int, corsConfig CORSConfig) *WebSocketServer {
	s := &WebSocketServer{
		port:       port,
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan Event, 100),
		corsConfig: corsConfig,
	}

	s.upgrader = websocket.Upgrader{
		CheckOrigin: s.checkOrigin,
	}

	return s
}

// checkOrigin validates the request origin against the configured CORS policy.
func (s *WebSocketServer) checkOrigin(r *http.Request) bool {
	// If AllowAllOrigins is true, allow everything
	if s.corsConfig.AllowAllOrigins {
		return true
	}

	origin := r.Header.Get("Origin")
	if origin == "" {
		// No origin header means same-origin request, allow it
		return true
	}

	// If no specific origins are configured and AllowAllOrigins is false,
	// deny all cross-origin requests (same-origin only)
	if len(s.corsConfig.AllowedOrigins) == 0 {
		log.Printf("CORS: Rejected origin %s (no allowed origins configured)", origin)
		return false
	}

	// Check if origin is in the allowed list
	for _, allowed := range s.corsConfig.AllowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
	}

	log.Printf("CORS: Rejected origin %s (allowed: %v)", origin, s.corsConfig.AllowedOrigins)
	return false
}

// SetService sets the parent service reference for health checks.
func (s *WebSocketServer) SetService(service *Service) {
	s.service = service
}

// Start starts the WebSocket server.
func (s *WebSocketServer) Start() error {
	// Start broadcast handler
	go s.handleBroadcasts()

	// Setup HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleWebSocket)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/health", s.handleHealth)

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
		log.Printf("Broadcast: Queued event %s for %d client(s)", event.Type, s.ClientCount())
	default:
		log.Printf("Warning: Broadcast channel full, dropping event %s", event.Type)
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
		case "replay_logs":
			// Extract clear_data parameter (default to false)
			clearData := false
			if data, ok := msg["clear_data"].(bool); ok {
				clearData = data
			}

			log.Printf("Received replay_logs command (clear_data: %v)", clearData)

			// Send acknowledgment
			ack := Event{
				Type: "replay:acknowledged",
				Data: map[string]interface{}{
					"clear_data": clearData,
				},
				Timestamp: time.Now(),
			}
			if err := conn.WriteJSON(ack); err != nil {
				log.Printf("Error sending replay acknowledgment: %v", err)
				return
			}

			// Run replay in background goroutine
			// This allows the WebSocket to continue handling other messages
			go func() {
				if err := s.service.ReplayHistoricalLogs(clearData); err != nil {
					log.Printf("Replay error: %v", err)
					// Broadcast error event
					s.Broadcast(Event{
						Type: "replay:error",
						Data: map[string]interface{}{
							"error": err.Error(),
						},
					})
				}
			}()
		case "start_replay":
			// Extract parameters - support both single file_path and multiple file_paths
			var filePaths []string

			// Check for new format (array of paths)
			if paths, ok := msg["file_paths"].([]interface{}); ok {
				for _, p := range paths {
					if pathStr, ok := p.(string); ok {
						filePaths = append(filePaths, pathStr)
					}
				}
			}

			// Fallback to old format (single path) for backward compatibility
			if len(filePaths) == 0 {
				if filePath, ok := msg["file_path"].(string); ok && filePath != "" {
					filePaths = []string{filePath}
				}
			}

			if len(filePaths) == 0 {
				log.Println("No file paths provided in start_replay command")
				errEvent := Event{
					Type: "replay:error",
					Data: map[string]interface{}{
						"error": "No file paths provided",
					},
					Timestamp: time.Now(),
				}
				if err := conn.WriteJSON(errEvent); err != nil {
					log.Printf("Error sending replay error: %v", err)
				}
				continue
			}

			speed := 1.0
			if s, ok := msg["speed"].(float64); ok {
				speed = s
			}
			filterType := "all"
			if f, ok := msg["filter"].(string); ok {
				filterType = f
			}
			pauseOnDraft := false
			if p, ok := msg["pause_on_draft"].(bool); ok {
				pauseOnDraft = p
			}

			log.Printf("Received start_replay command (%d file(s), speed: %.1fx, filter: %s, pauseOnDraft: %v)", len(filePaths), speed, filterType, pauseOnDraft)

			// Start replay with multiple files
			if err := s.service.StartReplay(filePaths, speed, filterType, pauseOnDraft); err != nil {
				log.Printf("Failed to start replay: %v", err)
				errEvent := Event{
					Type: "replay:error",
					Data: map[string]interface{}{
						"error": err.Error(),
					},
					Timestamp: time.Now(),
				}
				if err := conn.WriteJSON(errEvent); err != nil {
					log.Printf("Error sending replay error: %v", err)
				}
			}
		case "pause_replay":
			log.Println("Received pause_replay command")
			if err := s.service.PauseReplay(); err != nil {
				log.Printf("Failed to pause replay: %v", err)
				errEvent := Event{
					Type: "replay:error",
					Data: map[string]interface{}{
						"error": err.Error(),
					},
					Timestamp: time.Now(),
				}
				if err := conn.WriteJSON(errEvent); err != nil {
					log.Printf("Error sending replay error: %v", err)
				}
			}
		case "resume_replay":
			log.Println("Received resume_replay command")
			if err := s.service.ResumeReplay(); err != nil {
				log.Printf("Failed to resume replay: %v", err)
				errEvent := Event{
					Type: "replay:error",
					Data: map[string]interface{}{
						"error": err.Error(),
					},
					Timestamp: time.Now(),
				}
				if err := conn.WriteJSON(errEvent); err != nil {
					log.Printf("Error sending replay error: %v", err)
				}
			}
		case "stop_replay":
			log.Println("Received stop_replay command")
			if err := s.service.StopReplay(); err != nil {
				log.Printf("Failed to stop replay: %v", err)
				errEvent := Event{
					Type: "replay:error",
					Data: map[string]interface{}{
						"error": err.Error(),
					},
					Timestamp: time.Now(),
				}
				if err := conn.WriteJSON(errEvent); err != nil {
					log.Printf("Error sending replay error: %v", err)
				}
			}
		case "get_replay_status":
			log.Println("Received get_replay_status command")
			status := s.service.GetReplayStatus()
			statusEvent := Event{
				Type:      "replay:status",
				Data:      status,
				Timestamp: time.Now(),
			}
			if err := conn.WriteJSON(statusEvent); err != nil {
				log.Printf("Error sending replay status: %v", err)
			}
		}
	}
}

// handleBroadcasts handles broadcasting events to all clients.
func (s *WebSocketServer) handleBroadcasts() {
	for event := range s.broadcast {
		log.Printf("handleBroadcasts: Processing event %s for %d client(s)", event.Type, len(s.clients))
		s.clientsMu.RLock()
		clientCount := 0
		for client := range s.clients {
			if err := client.WriteJSON(event); err != nil {
				log.Printf("Error broadcasting %s to client: %v", event.Type, err)
				if err := client.Close(); err != nil {
					log.Printf("Error closing client after broadcast error: %v", err)
				}
				s.clientsMu.Lock()
				delete(s.clients, client)
				s.clientsMu.Unlock()
			} else {
				clientCount++
			}
		}
		s.clientsMu.RUnlock()
		log.Printf("handleBroadcasts: Successfully sent %s to %d client(s)", event.Type, clientCount)
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

// handleHealth handles HTTP health check requests.
func (s *WebSocketServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	if s.service == nil {
		// Service not initialized yet
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "unavailable",
			"message": "Service not fully initialized",
		})
		return
	}

	health := s.service.GetHealth()

	// Set appropriate HTTP status code based on health
	w.Header().Set("Content-Type", "application/json")
	switch health.Status {
	case "healthy":
		w.WriteHeader(http.StatusOK)
	case "degraded":
		w.WriteHeader(http.StatusOK) // Still return 200 for degraded
	default:
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	if err := json.NewEncoder(w).Encode(health); err != nil {
		log.Printf("Error encoding health response: %v", err)
	}
}
