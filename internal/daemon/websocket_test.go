package daemon

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestNewWebSocketServer(t *testing.T) {
	server := NewWebSocketServer(9999)

	if server == nil {
		t.Fatal("NewWebSocketServer returned nil")
	}

	if server.port != 9999 {
		t.Errorf("Expected port 9999, got %d", server.port)
	}

	if server.clients == nil {
		t.Error("clients map not initialized")
	}

	if server.broadcast == nil {
		t.Error("broadcast channel not initialized")
	}
}

func TestWebSocketServer_ClientCount(t *testing.T) {
	server := NewWebSocketServer(9999)

	count := server.ClientCount()
	if count != 0 {
		t.Errorf("Expected 0 clients initially, got %d", count)
	}
}

func TestWebSocketServer_Broadcast(t *testing.T) {
	server := NewWebSocketServer(9999)

	event := Event{
		Type: "test:event",
		Data: map[string]interface{}{
			"message": "test",
		},
	}

	// Should not panic even with no clients
	server.Broadcast(event)
}

func TestEvent_Structure(t *testing.T) {
	event := Event{
		Type: "stats:updated",
		Data: map[string]interface{}{
			"matches": 5,
			"games":   10,
		},
		Timestamp: time.Now(),
	}

	// Test JSON marshaling
	jsonBytes, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("Failed to marshal event: %v", err)
	}

	// Test JSON unmarshaling
	var decoded Event
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal event: %v", err)
	}

	if decoded.Type != "stats:updated" {
		t.Errorf("Expected type stats:updated, got %s", decoded.Type)
	}

	if decoded.Data["matches"].(float64) != 5 {
		t.Errorf("Expected matches 5, got %v", decoded.Data["matches"])
	}
}

func TestWebSocketServer_StatusHandler(t *testing.T) {
	server := NewWebSocketServer(9999)

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()

	server.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var status map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if status["status"] != "running" {
		t.Errorf("Expected status running, got %v", status["status"])
	}

	if status["clients"] != float64(0) {
		t.Errorf("Expected 0 clients, got %v", status["clients"])
	}
}

func TestWebSocketServer_UpgradeConnection(t *testing.T) {
	server := NewWebSocketServer(9999)

	// Create test HTTP server
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWebSocket))
	defer httpServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")

	// Try to connect
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Read welcome message
	var welcome Event
	if err := conn.ReadJSON(&welcome); err != nil {
		t.Fatalf("Failed to read welcome message: %v", err)
	}

	if welcome.Type != "daemon:connected" {
		t.Errorf("Expected welcome type daemon:connected, got %s", welcome.Type)
	}

	// Give server time to register client
	time.Sleep(50 * time.Millisecond)

	// Check client count
	if server.ClientCount() != 1 {
		t.Errorf("Expected 1 client, got %d", server.ClientCount())
	}
}

func TestWebSocketServer_PingPong(t *testing.T) {
	server := NewWebSocketServer(9999)

	// Create test HTTP server
	httpServer := httptest.NewServer(http.HandlerFunc(server.handleWebSocket))
	defer httpServer.Close()

	// Convert http:// to ws://
	wsURL := "ws" + strings.TrimPrefix(httpServer.URL, "http")

	// Connect
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Read welcome message
	var welcome Event
	if err := conn.ReadJSON(&welcome); err != nil {
		t.Fatalf("Failed to read welcome: %v", err)
	}

	// Send ping
	ping := map[string]interface{}{
		"type": "ping",
	}
	if err := conn.WriteJSON(ping); err != nil {
		t.Fatalf("Failed to send ping: %v", err)
	}

	// Read pong
	var pong Event
	if err := conn.ReadJSON(&pong); err != nil {
		t.Fatalf("Failed to read pong: %v", err)
	}

	if pong.Type != "pong" {
		t.Errorf("Expected pong type, got %s", pong.Type)
	}
}

func TestWebSocketServer_Stop(t *testing.T) {
	server := NewWebSocketServer(9999)

	// Should not panic when stopping without starting
	if err := server.Stop(); err != nil {
		t.Errorf("Stop returned error: %v", err)
	}

	// Check clients cleared
	if len(server.clients) != 0 {
		t.Errorf("Expected clients to be cleared, got %d", len(server.clients))
	}
}

func TestWebSocketServer_HealthHandler_NoService(t *testing.T) {
	server := NewWebSocketServer(9999)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["status"] != "unavailable" {
		t.Errorf("Expected status unavailable, got %v", response["status"])
	}

	if response["message"] != "Service not fully initialized" {
		t.Errorf("Expected initialization message, got %v", response["message"])
	}
}

func TestWebSocketServer_HealthHandler_Healthy(t *testing.T) {
	// Create a minimal service for testing
	// Note: This is a simplified test - in production, a full service would be initialized
	server := NewWebSocketServer(9999)

	// Create a mock service with healthy state
	mockService := &Service{
		startTime: time.Now().Add(-1 * time.Hour), // Running for 1 hour
		wsServer:  server,                         // Wire up the server
	}
	mockService.healthMu.Lock()
	mockService.lastLogRead = time.Now().Add(-1 * time.Minute)
	mockService.lastDBWrite = time.Now().Add(-2 * time.Minute)
	mockService.totalProcessed = 100
	mockService.totalErrors = 0
	mockService.healthMu.Unlock()

	server.SetService(mockService)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	var health HealthStatus
	if err := json.NewDecoder(w.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if health.Status != "healthy" {
		t.Errorf("Expected status healthy, got %s", health.Status)
	}

	if health.Version != Version {
		t.Errorf("Expected version %s, got %s", Version, health.Version)
	}

	if health.Database.Status != "ok" {
		t.Errorf("Expected database status ok, got %s", health.Database.Status)
	}

	if health.LogMonitor.Status != "ok" {
		t.Errorf("Expected log monitor status ok, got %s", health.LogMonitor.Status)
	}

	if health.WebSocket.Status != "ok" {
		t.Errorf("Expected websocket status ok, got %s", health.WebSocket.Status)
	}

	if health.Metrics.TotalProcessed != 100 {
		t.Errorf("Expected 100 processed, got %d", health.Metrics.TotalProcessed)
	}

	if health.Metrics.TotalErrors != 0 {
		t.Errorf("Expected 0 errors, got %d", health.Metrics.TotalErrors)
	}
}

func TestWebSocketServer_HealthHandler_Degraded_StaleLog(t *testing.T) {
	server := NewWebSocketServer(9999)

	// Create a mock service with stale log reads
	mockService := &Service{
		startTime: time.Now().Add(-1 * time.Hour),
		wsServer:  server, // Wire up the server
	}
	mockService.healthMu.Lock()
	mockService.lastLogRead = time.Now().Add(-10 * time.Minute) // 10 minutes ago (stale)
	mockService.lastDBWrite = time.Now().Add(-2 * time.Minute)
	mockService.totalProcessed = 100
	mockService.totalErrors = 0
	mockService.healthMu.Unlock()

	server.SetService(mockService)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for degraded, got %d", w.Code)
	}

	var health HealthStatus
	if err := json.NewDecoder(w.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if health.Status != "degraded" {
		t.Errorf("Expected status degraded, got %s", health.Status)
	}

	if health.LogMonitor.Status != "warning" {
		t.Errorf("Expected log monitor warning, got %s", health.LogMonitor.Status)
	}
}

func TestWebSocketServer_HealthHandler_Degraded_HighErrorRate(t *testing.T) {
	server := NewWebSocketServer(9999)

	// Create a mock service with high error rate
	mockService := &Service{
		startTime: time.Now().Add(-1 * time.Hour),
		wsServer:  server, // Wire up the server
	}
	mockService.healthMu.Lock()
	mockService.lastLogRead = time.Now().Add(-1 * time.Minute)
	mockService.lastDBWrite = time.Now().Add(-2 * time.Minute)
	mockService.totalProcessed = 100
	mockService.totalErrors = 15 // 15% error rate
	mockService.healthMu.Unlock()

	server.SetService(mockService)

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.handleHealth(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for degraded, got %d", w.Code)
	}

	var health HealthStatus
	if err := json.NewDecoder(w.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if health.Status != "degraded" {
		t.Errorf("Expected status degraded due to high error rate, got %s", health.Status)
	}

	if health.Metrics.TotalErrors != 15 {
		t.Errorf("Expected 15 errors, got %d", health.Metrics.TotalErrors)
	}
}

func TestWebSocketServer_SetService(t *testing.T) {
	server := NewWebSocketServer(9999)

	if server.service != nil {
		t.Error("Expected nil service initially")
	}

	mockService := &Service{
		startTime: time.Now(),
	}

	server.SetService(mockService)

	if server.service != mockService {
		t.Error("Expected service to be set")
	}
}

func TestNewWebSocketServerWithCORS(t *testing.T) {
	corsConfig := CORSConfig{
		AllowAllOrigins: false,
		AllowedOrigins:  []string{"https://example.com"},
	}

	server := NewWebSocketServerWithCORS(9999, corsConfig)

	if server == nil {
		t.Fatal("NewWebSocketServerWithCORS returned nil")
	}

	if server.corsConfig.AllowAllOrigins {
		t.Error("Expected AllowAllOrigins to be false")
	}

	if len(server.corsConfig.AllowedOrigins) != 1 {
		t.Errorf("Expected 1 allowed origin, got %d", len(server.corsConfig.AllowedOrigins))
	}
}

func TestWebSocketServer_CheckOrigin_AllowAll(t *testing.T) {
	server := NewWebSocketServerWithCORS(9999, CORSConfig{
		AllowAllOrigins: true,
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://random-origin.com")

	if !server.checkOrigin(req) {
		t.Error("Expected origin to be allowed when AllowAllOrigins is true")
	}
}

func TestWebSocketServer_CheckOrigin_SpecificOrigins(t *testing.T) {
	server := NewWebSocketServerWithCORS(9999, CORSConfig{
		AllowAllOrigins: false,
		AllowedOrigins:  []string{"https://example.com", "https://app.example.com"},
	})

	// Test allowed origin
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")

	if !server.checkOrigin(req) {
		t.Error("Expected https://example.com to be allowed")
	}

	// Test second allowed origin
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("Origin", "https://app.example.com")

	if !server.checkOrigin(req2) {
		t.Error("Expected https://app.example.com to be allowed")
	}

	// Test disallowed origin
	req3 := httptest.NewRequest("GET", "/", nil)
	req3.Header.Set("Origin", "https://evil.com")

	if server.checkOrigin(req3) {
		t.Error("Expected https://evil.com to be rejected")
	}
}

func TestWebSocketServer_CheckOrigin_NoOriginHeader(t *testing.T) {
	server := NewWebSocketServerWithCORS(9999, CORSConfig{
		AllowAllOrigins: false,
		AllowedOrigins:  []string{"https://example.com"},
	})

	// Request without Origin header (same-origin request)
	req := httptest.NewRequest("GET", "/", nil)

	if !server.checkOrigin(req) {
		t.Error("Expected same-origin request (no Origin header) to be allowed")
	}
}

func TestWebSocketServer_CheckOrigin_EmptyOriginsAndNotAllowAll(t *testing.T) {
	server := NewWebSocketServerWithCORS(9999, CORSConfig{
		AllowAllOrigins: false,
		AllowedOrigins:  nil, // No origins allowed
	})

	// Request with Origin header should be rejected
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://example.com")

	if server.checkOrigin(req) {
		t.Error("Expected cross-origin request to be rejected when no origins allowed")
	}

	// Request without Origin header should still be allowed (same-origin)
	req2 := httptest.NewRequest("GET", "/", nil)

	if !server.checkOrigin(req2) {
		t.Error("Expected same-origin request to be allowed")
	}
}

func TestWebSocketServer_CheckOrigin_Wildcard(t *testing.T) {
	server := NewWebSocketServerWithCORS(9999, CORSConfig{
		AllowAllOrigins: false,
		AllowedOrigins:  []string{"*"},
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://any-origin.com")

	if !server.checkOrigin(req) {
		t.Error("Expected wildcard to allow any origin")
	}
}
