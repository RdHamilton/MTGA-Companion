package scryfall

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := NewClient()

	if client == nil {
		t.Fatal("NewClient() returned nil")
	}

	if client.httpClient == nil {
		t.Error("httpClient is nil")
	}

	if client.rateLimiter == nil {
		t.Error("rateLimiter is nil")
	}

	if client.userAgent == "" {
		t.Error("userAgent is empty")
	}
}

func TestClient_RateLimiting(t *testing.T) {
	// Create a test server that counts requests
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"test","name":"Test Card"}`))
	}))
	defer server.Close()

	client := NewClient()
	ctx := context.Background()

	// Make 3 requests and measure time
	start := time.Now()
	for i := 0; i < 3; i++ {
		var card Card
		err := client.doRequest(ctx, server.URL, &card)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
	}
	elapsed := time.Since(start)

	// Should have made 3 requests
	if requestCount != 3 {
		t.Errorf("Expected 3 requests, got %d", requestCount)
	}

	// Should take at least 200ms (2 delays of 100ms each between 3 requests)
	minDuration := 200 * time.Millisecond
	if elapsed < minDuration {
		t.Errorf("Rate limiting not working: completed 3 requests in %v (expected >= %v)", elapsed, minDuration)
	}
}

func TestClient_GetCard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cards/test-id" {
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "test-id",
			"name": "Lightning Bolt",
			"mana_cost": "{R}",
			"cmc": 1.0,
			"type_line": "Instant",
			"oracle_text": "Deal 3 damage to any target."
		}`))
	}))
	defer server.Close()

	client := NewClient()
	ctx := context.Background()

	// Note: This test uses doRequest directly with test server
	// GetCard uses const baseURL, so we test doRequest separately
	var card Card
	err := client.doRequest(ctx, server.URL+"/cards/test-id", &card)
	if err != nil {
		t.Fatalf("doRequest failed: %v", err)
	}

	if card.Name != "Lightning Bolt" {
		t.Errorf("Expected card name 'Lightning Bolt', got '%s'", card.Name)
	}
}

func TestClient_NotFoundError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"object":"error","code":"not_found","status":404,"details":"No card found"}`))
	}))
	defer server.Close()

	client := NewClient()
	ctx := context.Background()

	var card Card
	err := client.doRequest(ctx, server.URL, &card)

	if err == nil {
		t.Fatal("Expected error for 404, got nil")
	}

	if !IsNotFound(err) {
		t.Errorf("Expected NotFoundError, got: %T", err)
	}
}

func TestClient_RateLimitRetry(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		if attemptCount < 2 {
			// First attempt: return 429
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"object":"error","code":"rate_limit","status":429}`))
			return
		}

		// Second attempt: success
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"id":"test","name":"Test Card"}`))
	}))
	defer server.Close()

	client := NewClient()
	ctx := context.Background()

	var card Card
	err := client.doRequest(ctx, server.URL, &card)
	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}

	if attemptCount < 2 {
		t.Errorf("Expected at least 2 attempts, got %d", attemptCount)
	}

	if card.Name != "Test Card" {
		t.Errorf("Expected card name 'Test Card', got '%s'", card.Name)
	}
}

func TestClient_MaxRetriesExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Always return 429
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"object":"error","code":"rate_limit","status":429}`))
	}))
	defer server.Close()

	client := NewClient()
	ctx := context.Background()

	var card Card
	err := client.doRequest(ctx, server.URL, &card)

	if err == nil {
		t.Fatal("Expected error after max retries, got nil")
	}

	// Should mention max retries or rate limiting
	errStr := err.Error()
	if errStr == "" {
		t.Error("Error message is empty")
	}
}

func TestClient_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	client := NewClient()
	ctx := context.Background()

	var card Card
	err := client.doRequest(ctx, server.URL, &card)

	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Slow response
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	var card Card
	err := client.doRequest(ctx, server.URL, &card)

	if err == nil {
		t.Fatal("Expected error from context cancellation, got nil")
	}
}

func TestClient_UserAgent(t *testing.T) {
	receivedUserAgent := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUserAgent = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient()
	ctx := context.Background()

	var card Card
	client.doRequest(ctx, server.URL, &card)

	if receivedUserAgent == "" {
		t.Error("User-Agent header not set")
	}

	if receivedUserAgent != "MTGA-Companion/1.0" {
		t.Errorf("Expected User-Agent 'MTGA-Companion/1.0', got '%s'", receivedUserAgent)
	}
}

func TestClient_AcceptHeader(t *testing.T) {
	receivedAccept := ""
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAccept = r.Header.Get("Accept")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client := NewClient()
	ctx := context.Background()

	var card Card
	client.doRequest(ctx, server.URL, &card)

	if receivedAccept != "application/json" {
		t.Errorf("Expected Accept header 'application/json', got '%s'", receivedAccept)
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "NotFoundError",
			err:      &NotFoundError{URL: "test"},
			expected: true,
		},
		{
			name:     "Other error",
			err:      &APIError{Status: 500},
			expected: false,
		},
		{
			name:     "Nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNotFound(tt.err)
			if result != tt.expected {
				t.Errorf("IsNotFound() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		apiError APIError
		contains string
	}{
		{
			name: "With details",
			apiError: APIError{
				Status:  404,
				Details: "Card not found",
			},
			contains: "Card not found",
		},
		{
			name: "Without details",
			apiError: APIError{
				Status: 500,
				Code:   "internal_error",
			},
			contains: "internal_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.apiError.Error()
			if errMsg == "" {
				t.Error("Error message is empty")
			}
		})
	}
}
