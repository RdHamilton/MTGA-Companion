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

// ClientSubscription tracks event subscriptions for a WebSocket client.
type ClientSubscription struct {
	subscriptions map[string]bool
	subscribeAll  bool // If true, client receives all events (default behavior)
	mu            sync.RWMutex
}

// NewClientSubscription creates a new client subscription with default "subscribe to all" behavior.
func NewClientSubscription() *ClientSubscription {
	return &ClientSubscription{
		subscriptions: make(map[string]bool),
		subscribeAll:  true, // Default: receive all events for backwards compatibility
	}
}

// Subscribe adds event types to the subscription list.
// If this is the first explicit subscription, disables "subscribe all" mode.
func (cs *ClientSubscription) Subscribe(eventTypes []string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// First explicit subscription disables "subscribe all" mode
	if cs.subscribeAll && len(eventTypes) > 0 {
		cs.subscribeAll = false
	}

	for _, eventType := range eventTypes {
		cs.subscriptions[eventType] = true
	}
}

// Unsubscribe removes event types from the subscription list.
func (cs *ClientSubscription) Unsubscribe(eventTypes []string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for _, eventType := range eventTypes {
		delete(cs.subscriptions, eventType)
	}
}

// SubscribeAll enables receiving all events.
func (cs *ClientSubscription) SubscribeAll() {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.subscribeAll = true
}

// IsSubscribed checks if the client should receive the given event type.
func (cs *ClientSubscription) IsSubscribed(eventType string) bool {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	// If subscribed to all, always return true
	if cs.subscribeAll {
		return true
	}

	return cs.subscriptions[eventType]
}

// GetSubscriptions returns a copy of current subscriptions.
func (cs *ClientSubscription) GetSubscriptions() []string {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if cs.subscribeAll {
		return []string{"*"}
	}

	result := make([]string, 0, len(cs.subscriptions))
	for eventType := range cs.subscriptions {
		result = append(result, eventType)
	}
	return result
}

// WebSocketServer manages WebSocket connections and event broadcasting.
type WebSocketServer struct {
	port       int
	clients    map[*websocket.Conn]*ClientSubscription
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
		clients:    make(map[*websocket.Conn]*ClientSubscription),
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
	s.clients = make(map[*websocket.Conn]*ClientSubscription)
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

	// Register client with default subscription (all events)
	subscription := NewClientSubscription()
	s.clientsMu.Lock()
	s.clients[conn] = subscription
	s.clientsMu.Unlock()

	log.Printf("Client connected (total: %d)", s.ClientCount())

	// Send welcome message with subscription info
	welcome := Event{
		Type: "daemon:connected",
		Data: map[string]interface{}{
			"message":       "Connected to MTGA Companion daemon",
			"subscriptions": subscription.GetSubscriptions(),
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
			// Get the client's subscription
			s.clientsMu.RLock()
			subscription := s.clients[conn]
			s.clientsMu.RUnlock()

			if subscription == nil {
				log.Printf("Error: subscription not found for client")
				continue
			}

			// Extract event types from message
			eventTypes := extractEventTypes(msg["events"])
			if len(eventTypes) == 0 {
				// If no events specified, subscribe to all
				subscription.SubscribeAll()
				log.Printf("Client subscribed to all events")
			} else {
				subscription.Subscribe(eventTypes)
				log.Printf("Client subscribed to events: %v", eventTypes)
			}

			// Send confirmation
			ack := Event{
				Type: "subscription:updated",
				Data: map[string]interface{}{
					"action":        "subscribe",
					"subscriptions": subscription.GetSubscriptions(),
				},
				Timestamp: time.Now(),
			}
			if err := conn.WriteJSON(ack); err != nil {
				log.Printf("Error sending subscription acknowledgment: %v", err)
				return
			}
		case "unsubscribe":
			// Get the client's subscription
			s.clientsMu.RLock()
			subscription := s.clients[conn]
			s.clientsMu.RUnlock()

			if subscription == nil {
				log.Printf("Error: subscription not found for client")
				continue
			}

			// Extract event types from message
			eventTypes := extractEventTypes(msg["events"])
			if len(eventTypes) > 0 {
				subscription.Unsubscribe(eventTypes)
				log.Printf("Client unsubscribed from events: %v", eventTypes)
			}

			// Send confirmation
			ack := Event{
				Type: "subscription:updated",
				Data: map[string]interface{}{
					"action":        "unsubscribe",
					"subscriptions": subscription.GetSubscriptions(),
				},
				Timestamp: time.Now(),
			}
			if err := conn.WriteJSON(ack); err != nil {
				log.Printf("Error sending unsubscription acknowledgment: %v", err)
				return
			}
		case "get_subscriptions":
			// Get the client's subscription
			s.clientsMu.RLock()
			subscription := s.clients[conn]
			s.clientsMu.RUnlock()

			if subscription == nil {
				log.Printf("Error: subscription not found for client")
				continue
			}

			// Send current subscriptions
			response := Event{
				Type: "subscription:list",
				Data: map[string]interface{}{
					"subscriptions": subscription.GetSubscriptions(),
				},
				Timestamp: time.Now(),
			}
			if err := conn.WriteJSON(response); err != nil {
				log.Printf("Error sending subscription list: %v", err)
				return
			}
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

// handleBroadcasts handles broadcasting events to subscribed clients.
func (s *WebSocketServer) handleBroadcasts() {
	for event := range s.broadcast {
		log.Printf("handleBroadcasts: Processing event %s for %d client(s)", event.Type, len(s.clients))
		s.clientsMu.RLock()
		sentCount := 0
		skippedCount := 0
		clientsToRemove := make([]*websocket.Conn, 0)

		for client, subscription := range s.clients {
			// Check if client is subscribed to this event type
			if !subscription.IsSubscribed(event.Type) {
				skippedCount++
				continue
			}

			if err := client.WriteJSON(event); err != nil {
				log.Printf("Error broadcasting %s to client: %v", event.Type, err)
				if err := client.Close(); err != nil {
					log.Printf("Error closing client after broadcast error: %v", err)
				}
				clientsToRemove = append(clientsToRemove, client)
			} else {
				sentCount++
			}
		}
		s.clientsMu.RUnlock()

		// Remove failed clients
		if len(clientsToRemove) > 0 {
			s.clientsMu.Lock()
			for _, client := range clientsToRemove {
				delete(s.clients, client)
			}
			s.clientsMu.Unlock()
		}

		if skippedCount > 0 {
			log.Printf("handleBroadcasts: Sent %s to %d client(s), skipped %d (not subscribed)", event.Type, sentCount, skippedCount)
		} else {
			log.Printf("handleBroadcasts: Successfully sent %s to %d client(s)", event.Type, sentCount)
		}
	}
}

// extractEventTypes extracts a slice of event type strings from a message field.
func extractEventTypes(events interface{}) []string {
	if events == nil {
		return nil
	}

	// Handle array of strings
	if arr, ok := events.([]interface{}); ok {
		result := make([]string, 0, len(arr))
		for _, item := range arr {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	}

	// Handle single string
	if str, ok := events.(string); ok {
		return []string{str}
	}

	return nil
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
