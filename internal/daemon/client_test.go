package daemon

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	config := DefaultClientConfig(9999)
	client := NewClient(config)

	if client == nil {
		t.Fatal("expected client to be created")
	}
	if client.config.BaseURL != "http://localhost:9999" {
		t.Errorf("expected base URL http://localhost:9999, got %s", client.config.BaseURL)
	}
}

func TestDefaultClientConfig(t *testing.T) {
	config := DefaultClientConfig(8080)

	if config.BaseURL != "http://localhost:8080" {
		t.Errorf("expected base URL http://localhost:8080, got %s", config.BaseURL)
	}
	if config.Timeout != 10*time.Second {
		t.Errorf("expected timeout 10s, got %v", config.Timeout)
	}
	if config.MaxRetries != 3 {
		t.Errorf("expected max retries 3, got %d", config.MaxRetries)
	}
	if config.RetryBaseDelay != 500*time.Millisecond {
		t.Errorf("expected retry base delay 500ms, got %v", config.RetryBaseDelay)
	}
}

func TestClient_GetStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/status" {
			t.Errorf("expected path /status, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected method GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Status{
			Status:        "connected",
			Connected:     true,
			Version:       "1.0.0",
			MTGAConnected: true,
			PlayerID:      "player123",
			LastUpdate:    "2025-11-25T12:00:00Z",
		})
	}))
	defer server.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	client := NewClient(config)

	ctx := context.Background()
	status, err := client.GetStatus(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.Status != "connected" {
		t.Errorf("expected status connected, got %s", status.Status)
	}
	if !status.Connected {
		t.Error("expected connected to be true")
	}
	if status.Version != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", status.Version)
	}
	if status.PlayerID != "player123" {
		t.Errorf("expected player ID player123, got %s", status.PlayerID)
	}
}

func TestClient_GetCards(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cards" {
			t.Errorf("expected path /cards, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CardCollection{
			Cards: map[int]int{
				12345: 4,
				67890: 2,
				11111: 1,
			},
		})
	}))
	defer server.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	client := NewClient(config)

	ctx := context.Background()
	collection, err := client.GetCards(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(collection.Cards) != 3 {
		t.Errorf("expected 3 cards, got %d", len(collection.Cards))
	}
	if collection.Cards[12345] != 4 {
		t.Errorf("expected card 12345 to have 4 copies, got %d", collection.Cards[12345])
	}
	if collection.Cards[67890] != 2 {
		t.Errorf("expected card 67890 to have 2 copies, got %d", collection.Cards[67890])
	}
}

func TestClient_GetInventory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/inventory" {
			t.Errorf("expected path /inventory, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Inventory{
			Gold:          25000,
			Gems:          3500,
			CommonWC:      45,
			UncommonWC:    32,
			RareWC:        15,
			MythicWC:      8,
			VaultProgress: 75.5,
			DraftTokens:   2,
			SealedTokens:  1,
		})
	}))
	defer server.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	client := NewClient(config)

	ctx := context.Background()
	inventory, err := client.GetInventory(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if inventory.Gold != 25000 {
		t.Errorf("expected gold 25000, got %d", inventory.Gold)
	}
	if inventory.Gems != 3500 {
		t.Errorf("expected gems 3500, got %d", inventory.Gems)
	}
	if inventory.RareWC != 15 {
		t.Errorf("expected rare wildcards 15, got %d", inventory.RareWC)
	}
	if inventory.MythicWC != 8 {
		t.Errorf("expected mythic wildcards 8, got %d", inventory.MythicWC)
	}
	if inventory.VaultProgress != 75.5 {
		t.Errorf("expected vault progress 75.5, got %f", inventory.VaultProgress)
	}
}

func TestClient_GetPlayerID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/playerId" {
			t.Errorf("expected path /playerId, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(PlayerInfo{
			PlayerID:   "ABCD1234",
			PlayerName: "TestPlayer",
		})
	}))
	defer server.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	client := NewClient(config)

	ctx := context.Background()
	playerID, err := client.GetPlayerID(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if playerID != "ABCD1234" {
		t.Errorf("expected player ID ABCD1234, got %s", playerID)
	}
}

func TestClient_IsHealthy(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{"connected", "connected", true},
		{"healthy", "healthy", true},
		{"disconnected", "disconnected", false},
		{"error", "error", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(Status{Status: tt.status})
			}))
			defer server.Close()

			config := DefaultClientConfig(9999)
			config.BaseURL = server.URL
			client := NewClient(config)

			ctx := context.Background()
			healthy := client.IsHealthy(ctx)

			if healthy != tt.expected {
				t.Errorf("expected healthy=%v, got %v", tt.expected, healthy)
			}
		})
	}
}

func TestClient_RetryOnServerError(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempt := atomic.AddInt32(&attempts, 1)
		if attempt < 3 {
			// First two attempts fail with 500
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("server error"))
			return
		}
		// Third attempt succeeds
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Status{Status: "connected"})
	}))
	defer server.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	config.RetryBaseDelay = 10 * time.Millisecond // Speed up test
	client := NewClient(config)

	ctx := context.Background()
	status, err := client.GetStatus(ctx)
	if err != nil {
		t.Fatalf("unexpected error after retries: %v", err)
	}

	if status.Status != "connected" {
		t.Errorf("expected status connected, got %s", status.Status)
	}

	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClient_NoRetryOn4xxError(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("not found"))
	}))
	defer server.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	config.RetryBaseDelay = 10 * time.Millisecond
	client := NewClient(config)

	ctx := context.Background()
	_, err := client.GetStatus(ctx)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}

	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("expected 1 attempt (no retry for 4xx), got %d", attempts)
	}
}

func TestClient_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Simulate slow response
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	config.Timeout = 30 * time.Second // Long timeout, we'll cancel via context
	client := NewClient(config)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.GetStatus(ctx)
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
}

func TestClient_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	client := NewClient(config)

	ctx := context.Background()
	_, err := client.GetStatus(ctx)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestClient_SetBaseURL(t *testing.T) {
	config := DefaultClientConfig(9999)
	client := NewClient(config)

	client.SetBaseURL("http://newhost:8080")

	if client.GetBaseURL() != "http://newhost:8080" {
		t.Errorf("expected base URL http://newhost:8080, got %s", client.GetBaseURL())
	}
}

func TestClient_MaxRetriesExhausted(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer server.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	config.MaxRetries = 2
	config.RetryBaseDelay = 10 * time.Millisecond
	client := NewClient(config)

	ctx := context.Background()
	_, err := client.GetStatus(ctx)
	if err == nil {
		t.Fatal("expected error after max retries exhausted")
	}

	// Should be 1 initial + 2 retries = 3 total attempts
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestClient_EmptyCards(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CardCollection{
			Cards: map[int]int{},
		})
	}))
	defer server.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	client := NewClient(config)

	ctx := context.Background()
	collection, err := client.GetCards(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(collection.Cards) != 0 {
		t.Errorf("expected 0 cards, got %d", len(collection.Cards))
	}
}

func TestClient_AcceptHeaderSet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		if accept != "application/json" {
			t.Errorf("expected Accept header application/json, got %s", accept)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Status{Status: "connected"})
	}))
	defer server.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	client := NewClient(config)

	ctx := context.Background()
	_, _ = client.GetStatus(ctx)
}
