package meta

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewGoldfishClient(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		client := NewGoldfishClient(nil)
		if client == nil {
			t.Fatal("expected non-nil client")
		}
		if client.baseURL != "https://www.mtggoldfish.com" {
			t.Errorf("expected default base URL, got %s", client.baseURL)
		}
		if client.cacheTTL != 4*time.Hour {
			t.Errorf("expected 4 hour cache TTL, got %v", client.cacheTTL)
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &GoldfishConfig{
			BaseURL:        "https://custom.url",
			CacheTTL:       2 * time.Hour,
			RequestTimeout: 10 * time.Second,
			RateLimitMs:    500,
		}
		client := NewGoldfishClient(config)
		if client.baseURL != "https://custom.url" {
			t.Errorf("expected custom base URL, got %s", client.baseURL)
		}
		if client.cacheTTL != 2*time.Hour {
			t.Errorf("expected 2 hour cache TTL, got %v", client.cacheTTL)
		}
	})
}

func TestDefaultGoldfishConfig(t *testing.T) {
	config := DefaultGoldfishConfig()

	if config.BaseURL != "https://www.mtggoldfish.com" {
		t.Errorf("unexpected BaseURL: %s", config.BaseURL)
	}
	if config.CacheTTL != 4*time.Hour {
		t.Errorf("unexpected CacheTTL: %v", config.CacheTTL)
	}
	if config.RequestTimeout != 30*time.Second {
		t.Errorf("unexpected RequestTimeout: %v", config.RequestTimeout)
	}
	if config.RateLimitMs != 1000 {
		t.Errorf("unexpected RateLimitMs: %d", config.RateLimitMs)
	}
}

func TestGoldfishClient_GetMeta(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return mock HTML with archetype data
		html := `
		<html>
		<body>
		<div class="archetype-tile">
			<span class="archetype-tile-title">Mono Red Aggro</span>
			<span class="archetype-tile-statistic">15.5%</span>
		</div>
		<div class="archetype-tile">
			<span class="archetype-tile-title">Azorius Control</span>
			<span class="archetype-tile-statistic">12.3%</span>
		</div>
		<div class="archetype-tile">
			<span class="archetype-tile-title">Golgari Midrange</span>
			<span class="archetype-tile-statistic">8.7%</span>
		</div>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &GoldfishConfig{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10, // Fast rate limit for tests
	}
	client := NewGoldfishClient(config)

	ctx := context.Background()
	meta, err := client.GetMeta(ctx, "standard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta == nil {
		t.Fatal("expected non-nil meta")
	}
	if meta.Format != "standard" {
		t.Errorf("expected format 'standard', got %s", meta.Format)
	}
	if meta.Source != "mtggoldfish" {
		t.Errorf("expected source 'mtggoldfish', got %s", meta.Source)
	}
	if len(meta.Decks) != 3 {
		t.Errorf("expected 3 decks, got %d", len(meta.Decks))
	}
}

func TestGoldfishClient_GetMeta_UnsupportedFormat(t *testing.T) {
	config := &GoldfishConfig{
		BaseURL:        "https://test.url",
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewGoldfishClient(config)

	ctx := context.Background()
	_, err := client.GetMeta(ctx, "unsupported")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestGoldfishClient_GetTopDecks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
		<body>
		<div class="archetype-tile">
			<span class="archetype-tile-title">Deck One</span>
			<span class="archetype-tile-statistic">20.0%</span>
		</div>
		<div class="archetype-tile">
			<span class="archetype-tile-title">Deck Two</span>
			<span class="archetype-tile-statistic">15.0%</span>
		</div>
		<div class="archetype-tile">
			<span class="archetype-tile-title">Deck Three</span>
			<span class="archetype-tile-statistic">10.0%</span>
		</div>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &GoldfishConfig{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewGoldfishClient(config)

	ctx := context.Background()
	decks, err := client.GetTopDecks(ctx, "standard", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(decks) != 2 {
		t.Errorf("expected 2 decks, got %d", len(decks))
	}
}

func TestGoldfishClient_GetDeckByArchetype(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
		<body>
		<div class="archetype-tile">
			<span class="archetype-tile-title">Mono Red Aggro</span>
			<span class="archetype-tile-statistic">15.0%</span>
		</div>
		<div class="archetype-tile">
			<span class="archetype-tile-title">Azorius Control</span>
			<span class="archetype-tile-statistic">10.0%</span>
		</div>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &GoldfishConfig{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewGoldfishClient(config)

	ctx := context.Background()

	t.Run("found archetype", func(t *testing.T) {
		deck, err := client.GetDeckByArchetype(ctx, "standard", "mono red aggro")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if deck.Name != "Mono Red Aggro" {
			t.Errorf("expected 'Mono Red Aggro', got %s", deck.Name)
		}
	})

	t.Run("not found archetype", func(t *testing.T) {
		_, err := client.GetDeckByArchetype(ctx, "standard", "nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent archetype")
		}
	})
}

func TestGoldfishClient_GetMetaShare(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
		<body>
		<div class="archetype-tile">
			<span class="archetype-tile-title">Test Deck</span>
			<span class="archetype-tile-statistic">25.5%</span>
		</div>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &GoldfishConfig{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewGoldfishClient(config)

	ctx := context.Background()
	share, err := client.GetMetaShare(ctx, "standard", "test deck")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if share != 25.5 {
		t.Errorf("expected meta share 25.5, got %f", share)
	}
}

func TestGoldfishClient_Cache(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		html := `
		<html>
		<body>
		<div class="archetype-tile">
			<span class="archetype-tile-title">Cached Deck</span>
			<span class="archetype-tile-statistic">10.0%</span>
		</div>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &GoldfishConfig{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewGoldfishClient(config)

	ctx := context.Background()

	// First request should hit server
	_, err := client.GetMeta(ctx, "standard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 1 {
		t.Errorf("expected 1 request, got %d", requestCount)
	}

	// Second request should use cache
	_, err = client.GetMeta(ctx, "standard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 1 {
		t.Errorf("expected still 1 request (cached), got %d", requestCount)
	}
}

func TestGoldfishClient_ClearCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
		<body>
		<div class="archetype-tile">
			<span class="archetype-tile-title">Test</span>
			<span class="archetype-tile-statistic">10.0%</span>
		</div>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &GoldfishConfig{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewGoldfishClient(config)

	ctx := context.Background()

	// Populate cache
	_, _ = client.GetMeta(ctx, "standard")

	// Check cache status
	cached, _ := client.GetCacheStatus("standard")
	if !cached {
		t.Error("expected cache to be populated")
	}

	// Clear cache
	client.ClearCache()

	// Check cache status after clear
	cached, _ = client.GetCacheStatus("standard")
	if cached {
		t.Error("expected cache to be cleared")
	}
}

func TestGoldfishClient_RefreshMeta(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		html := `
		<html>
		<body>
		<div class="archetype-tile">
			<span class="archetype-tile-title">Refreshed Deck</span>
			<span class="archetype-tile-statistic">10.0%</span>
		</div>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &GoldfishConfig{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewGoldfishClient(config)

	ctx := context.Background()

	// First request
	_, _ = client.GetMeta(ctx, "standard")

	// Refresh should bypass cache
	_, err := client.RefreshMeta(ctx, "standard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requestCount != 2 {
		t.Errorf("expected 2 requests after refresh, got %d", requestCount)
	}
}

func TestGoldfishClient_ExtractColorsFromName(t *testing.T) {
	client := NewGoldfishClient(nil)

	tests := []struct {
		name     string
		expected []string
	}{
		{"Mono Red Aggro", []string{"R"}},
		{"Azorius Control", []string{"W", "U"}},
		{"Golgari Midrange", []string{"B", "G"}},
		{"Esper Control", []string{"W", "U", "B"}},
		{"Jund Sacrifice", []string{"B", "R", "G"}},
		{"Unknown Deck", []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			colors := client.extractColorsFromName(tt.name)
			if len(colors) != len(tt.expected) {
				t.Errorf("expected %d colors, got %d", len(tt.expected), len(colors))
				return
			}
			for i, c := range tt.expected {
				if colors[i] != c {
					t.Errorf("expected color %s at position %d, got %s", c, i, colors[i])
				}
			}
		})
	}
}

func TestGoldfishClient_NormalizeArchetypeName(t *testing.T) {
	client := NewGoldfishClient(nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"Standard Mono Red", "mono red"},
		{"Historic Control", "control"},
		{"  Trimmed Name  ", "trimmed name"},
		{"Pioneer Aggro", "aggro"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := client.normalizeArchetypeName(tt.input)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestFormatMeta_Serialize(t *testing.T) {
	meta := &FormatMeta{
		Format: "standard",
		Decks: []*MetaDeck{
			{
				Name:      "Test Deck",
				MetaShare: 15.5,
			},
		},
		TotalDecks:  1,
		LastUpdated: time.Now(),
		Source:      "mtggoldfish",
	}

	data, err := meta.Serialize()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty serialized data")
	}

	// Deserialize and verify
	restored, err := DeserializeFormatMeta(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if restored.Format != meta.Format {
		t.Errorf("expected format %s, got %s", meta.Format, restored.Format)
	}
	if len(restored.Decks) != 1 {
		t.Errorf("expected 1 deck, got %d", len(restored.Decks))
	}
}

func TestDeserializeFormatMeta_InvalidJSON(t *testing.T) {
	_, err := DeserializeFormatMeta([]byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestGoldfishClient_ParseMetaPageTableFormat(t *testing.T) {
	client := NewGoldfishClient(nil)

	html := `
	<html>
	<body>
	<table>
	<tr>
		<td><a href="/archetype/mono-red#paper">Mono Red</a></td>
		<td>18.5%</td>
	</tr>
	<tr>
		<td><a href="/archetype/control#paper">Blue Control</a></td>
		<td>12.3%</td>
	</tr>
	</table>
	</body>
	</html>
	`

	meta := client.parseMetaPage(html, "standard")

	if len(meta.Decks) < 2 {
		t.Errorf("expected at least 2 decks from table format, got %d", len(meta.Decks))
	}
}

func TestGoldfishClient_GetCacheStatus_NotCached(t *testing.T) {
	client := NewGoldfishClient(nil)

	cached, expiresAt := client.GetCacheStatus("nonexistent")
	if cached {
		t.Error("expected not cached")
	}
	if !expiresAt.IsZero() {
		t.Error("expected zero time for nonexistent cache entry")
	}
}
