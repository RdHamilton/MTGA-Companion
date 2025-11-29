package meta

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewTop8Client(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		client := NewTop8Client(nil)
		if client == nil {
			t.Fatal("expected non-nil client")
		}
		if client.baseURL != "https://www.mtgtop8.com" {
			t.Errorf("expected default base URL, got %s", client.baseURL)
		}
		if client.cacheTTL != 6*time.Hour {
			t.Errorf("expected 6 hour cache TTL, got %v", client.cacheTTL)
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &Top8Config{
			BaseURL:        "https://custom.url",
			CacheTTL:       3 * time.Hour,
			RequestTimeout: 15 * time.Second,
			RateLimitMs:    2000,
		}
		client := NewTop8Client(config)
		if client.baseURL != "https://custom.url" {
			t.Errorf("expected custom base URL, got %s", client.baseURL)
		}
		if client.cacheTTL != 3*time.Hour {
			t.Errorf("expected 3 hour cache TTL, got %v", client.cacheTTL)
		}
	})
}

func TestDefaultTop8Config(t *testing.T) {
	config := DefaultTop8Config()

	if config.BaseURL != "https://www.mtgtop8.com" {
		t.Errorf("unexpected BaseURL: %s", config.BaseURL)
	}
	if config.CacheTTL != 6*time.Hour {
		t.Errorf("unexpected CacheTTL: %v", config.CacheTTL)
	}
	if config.RequestTimeout != 30*time.Second {
		t.Errorf("unexpected RequestTimeout: %v", config.RequestTimeout)
	}
	if config.RateLimitMs != 1500 {
		t.Errorf("unexpected RateLimitMs: %d", config.RateLimitMs)
	}
}

func TestTop8Client_GetTournamentMeta(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
		<body>
		<table>
		<tr>
			<td><a href="/archetype?a=123">Mono Red</a></td>
			<td>25</td>
		</tr>
		<tr>
			<td><a href="/archetype?a=456">Azorius Control</a></td>
			<td>18</td>
		</tr>
		</table>
		<div class="event_title">Pro Tour Test</div>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &Top8Config{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewTop8Client(config)

	ctx := context.Background()
	meta, err := client.GetTournamentMeta(ctx, "standard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if meta == nil {
		t.Fatal("expected non-nil meta")
	}
	if meta.Format != "standard" {
		t.Errorf("expected format 'standard', got %s", meta.Format)
	}
	if meta.Source != "mtgtop8" {
		t.Errorf("expected source 'mtgtop8', got %s", meta.Source)
	}
}

func TestTop8Client_GetTournamentMeta_UnsupportedFormat(t *testing.T) {
	config := &Top8Config{
		BaseURL:        "https://test.url",
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewTop8Client(config)

	ctx := context.Background()
	_, err := client.GetTournamentMeta(ctx, "commander")
	if err == nil {
		t.Error("expected error for unsupported format")
	}
}

func TestTop8Client_GetRecentTournaments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
		<body>
		<div class="event_title">Tournament One</div>
		<div class="event_title">Tournament Two</div>
		<div class="event_title">Tournament Three</div>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &Top8Config{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewTop8Client(config)

	ctx := context.Background()
	tournaments, err := client.GetRecentTournaments(ctx, "standard", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tournaments) > 2 {
		t.Errorf("expected at most 2 tournaments, got %d", len(tournaments))
	}
}

func TestTop8Client_GetArchetypeStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
		<body>
		<table>
		<tr>
			<td><a href="/archetype?a=123">Mono Red</a></td>
			<td>25</td>
		</tr>
		<tr>
			<td><a href="/archetype?a=456">Azorius Control</a></td>
			<td>18</td>
		</tr>
		</table>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &Top8Config{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewTop8Client(config)

	ctx := context.Background()

	t.Run("found archetype", func(t *testing.T) {
		stats, err := client.GetArchetypeStats(ctx, "standard", "mono red")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if stats == nil {
			t.Fatal("expected non-nil stats")
		}
		if stats.Top8Count != 25 {
			t.Errorf("expected Top8Count 25, got %d", stats.Top8Count)
		}
	})

	t.Run("not found archetype", func(t *testing.T) {
		_, err := client.GetArchetypeStats(ctx, "standard", "nonexistent")
		if err == nil {
			t.Error("expected error for nonexistent archetype")
		}
	})
}

func TestTop8Client_GetTopArchetypes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
		<body>
		<table>
		<tr>
			<td><a href="/archetype?a=1">First Archetype</a></td>
			<td>50</td>
		</tr>
		<tr>
			<td><a href="/archetype?a=2">Second Archetype</a></td>
			<td>30</td>
		</tr>
		<tr>
			<td><a href="/archetype?a=3">Third Archetype</a></td>
			<td>20</td>
		</tr>
		</table>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &Top8Config{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewTop8Client(config)

	ctx := context.Background()
	archetypes, err := client.GetTopArchetypes(ctx, "standard", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(archetypes) != 2 {
		t.Errorf("expected 2 archetypes, got %d", len(archetypes))
	}

	// Verify sorting (highest first)
	if len(archetypes) >= 2 && archetypes[0].Top8Count < archetypes[1].Top8Count {
		t.Error("expected archetypes to be sorted by Top8Count descending")
	}
}

func TestTop8Client_Cache(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		html := `
		<html>
		<body>
		<div class="event_title">Cached Tournament</div>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &Top8Config{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewTop8Client(config)

	ctx := context.Background()

	// First request should hit server
	_, err := client.GetTournamentMeta(ctx, "standard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 1 {
		t.Errorf("expected 1 request, got %d", requestCount)
	}

	// Second request should use cache
	_, err = client.GetTournamentMeta(ctx, "standard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount != 1 {
		t.Errorf("expected still 1 request (cached), got %d", requestCount)
	}
}

func TestTop8Client_ClearCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `<html><body><div class="event_title">Test</div></body></html>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &Top8Config{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewTop8Client(config)

	ctx := context.Background()

	// Populate cache
	_, _ = client.GetTournamentMeta(ctx, "standard")

	// Verify cache has data
	cached := client.getFromCache("standard")
	if cached == nil {
		t.Error("expected cache to be populated")
	}

	// Clear cache
	client.ClearCache()

	// Verify cache is empty
	cached = client.getFromCache("standard")
	if cached != nil {
		t.Error("expected cache to be cleared")
	}
}

func TestTop8Client_RefreshMeta(t *testing.T) {
	requestCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		html := `<html><body><div class="event_title">Refreshed</div></body></html>`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &Top8Config{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewTop8Client(config)

	ctx := context.Background()

	// First request
	_, _ = client.GetTournamentMeta(ctx, "standard")

	// Refresh should bypass cache
	_, err := client.RefreshMeta(ctx, "standard")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requestCount != 2 {
		t.Errorf("expected 2 requests after refresh, got %d", requestCount)
	}
}

func TestTop8Client_ExtractColorsFromName(t *testing.T) {
	client := NewTop8Client(nil)

	tests := []struct {
		name     string
		expected []string
	}{
		{"Mono Red", []string{"R"}},
		{"Azorius Control", []string{"W", "U"}},
		{"Golgari Midrange", []string{"B", "G"}},
		{"Esper Control", []string{"W", "U", "B"}},
		{"Jund", []string{"B", "R", "G"}},
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

func TestTop8Client_NormalizeArchetypeName(t *testing.T) {
	client := NewTop8Client(nil)

	tests := []struct {
		input    string
		expected string
	}{
		{"Mono Red", "mono red"},
		{"  Trimmed Name  ", "trimmed name"},
		{"UPPERCASE", "uppercase"},
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

func TestTop8Client_ParseTournamentPage(t *testing.T) {
	client := NewTop8Client(nil)

	html := `
	<html>
	<body>
	<table>
	<tr>
		<td><a href="/archetype?a=123">Control</a></td>
		<td>30</td>
	</tr>
	<tr>
		<td><a href="/archetype?a=456">Aggro</a></td>
		<td>25</td>
	</tr>
	</table>
	<div class="event_title">Grand Prix Test</div>
	<div class="event_title">Pro Tour Test</div>
	</body>
	</html>
	`

	meta := client.parseTournamentPage(html, "modern")

	if meta.Format != "modern" {
		t.Errorf("expected format 'modern', got %s", meta.Format)
	}

	// Should have parsed archetype stats
	if len(meta.ArchetypeStats) < 2 {
		t.Errorf("expected at least 2 archetype stats, got %d", len(meta.ArchetypeStats))
	}

	// Should have parsed tournaments
	if len(meta.Tournaments) < 2 {
		t.Errorf("expected at least 2 tournaments, got %d", len(meta.Tournaments))
	}
}

func TestArchetypeStats_Fields(t *testing.T) {
	stats := &ArchetypeStats{
		ArchetypeName:   "Test Archetype",
		TotalPlacements: 100,
		Top8Count:       50,
		WinCount:        10,
		AvgPlacement:    3.5,
		Colors:          []string{"W", "U"},
		PopularCards:    []string{"Card A", "Card B"},
		TrendDirection:  "up",
		LastSeen:        time.Now(),
	}

	if stats.ArchetypeName != "Test Archetype" {
		t.Errorf("unexpected ArchetypeName: %s", stats.ArchetypeName)
	}
	if stats.TotalPlacements != 100 {
		t.Errorf("unexpected TotalPlacements: %d", stats.TotalPlacements)
	}
	if stats.Top8Count != 50 {
		t.Errorf("unexpected Top8Count: %d", stats.Top8Count)
	}
	if stats.WinCount != 10 {
		t.Errorf("unexpected WinCount: %d", stats.WinCount)
	}
}

func TestTournament_Fields(t *testing.T) {
	tournament := &Tournament{
		Name:    "Test Tournament",
		Format:  "standard",
		Date:    time.Now(),
		Players: 256,
		URL:     "https://example.com/tournament",
		TopDecks: []*TopDeck{
			{
				Placement:     1,
				ArchetypeName: "Winner Deck",
				PlayerName:    "Player One",
				Colors:        []string{"R"},
			},
		},
		LastUpdated: time.Now(),
	}

	if tournament.Name != "Test Tournament" {
		t.Errorf("unexpected Name: %s", tournament.Name)
	}
	if tournament.Players != 256 {
		t.Errorf("unexpected Players: %d", tournament.Players)
	}
	if len(tournament.TopDecks) != 1 {
		t.Errorf("unexpected TopDecks length: %d", len(tournament.TopDecks))
	}
	if tournament.TopDecks[0].Placement != 1 {
		t.Errorf("unexpected Placement: %d", tournament.TopDecks[0].Placement)
	}
}

func TestTournamentMeta_DateRange(t *testing.T) {
	now := time.Now()
	meta := &TournamentMeta{
		Format: "standard",
		DateRange: DateRange{
			Start: now.AddDate(0, 0, -30),
			End:   now,
		},
	}

	if meta.DateRange.End.Before(meta.DateRange.Start) {
		t.Error("expected End to be after Start")
	}
}

func TestTop8Client_GetRecentTournaments_NoLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
		<body>
		<div class="event_title">Tournament One</div>
		<div class="event_title">Tournament Two</div>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &Top8Config{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewTop8Client(config)

	ctx := context.Background()

	// Test with 0 limit (should return all)
	tournaments, err := client.GetRecentTournaments(ctx, "standard", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return all tournaments
	if len(tournaments) != 2 {
		t.Errorf("expected 2 tournaments with no limit, got %d", len(tournaments))
	}
}

func TestTop8Client_GetTopArchetypes_NoLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
		<body>
		<table>
		<tr>
			<td><a href="/archetype?a=1">Archetype One</a></td>
			<td>50</td>
		</tr>
		<tr>
			<td><a href="/archetype?a=2">Archetype Two</a></td>
			<td>30</td>
		</tr>
		</table>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	config := &Top8Config{
		BaseURL:        server.URL,
		CacheTTL:       1 * time.Hour,
		RequestTimeout: 5 * time.Second,
		RateLimitMs:    10,
	}
	client := NewTop8Client(config)

	ctx := context.Background()

	// Test with 0 limit (should return all)
	archetypes, err := client.GetTopArchetypes(ctx, "standard", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should return all archetypes
	if len(archetypes) != 2 {
		t.Errorf("expected 2 archetypes with no limit, got %d", len(archetypes))
	}
}
