package synergy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestArchidektClient_GetDeck(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/decks/12345/" {
			http.NotFound(w, r)
			return
		}

		deck := ArchidektDeck{
			ID:         12345,
			Name:       "Test Deck",
			DeckFormat: 3, // Commander
			ViewCount:  100,
			Cards: []*ArchidektDeckCard{
				{
					ID:       1,
					Quantity: 1,
					Card: &ArchidektCard{
						ID: 101,
						OracleCard: &ArchidektOracleCard{
							ID:   201,
							Name: "Sol Ring",
							CMC:  1,
						},
					},
				},
				{
					ID:       2,
					Quantity: 1,
					Card: &ArchidektCard{
						ID: 102,
						OracleCard: &ArchidektOracleCard{
							ID:   202,
							Name: "Command Tower",
							CMC:  0,
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(deck)
	}))
	defer server.Close()

	client := NewArchidektClient()
	client.baseURL = server.URL

	deck, err := client.GetDeck(context.Background(), 12345)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if deck.ID != 12345 {
		t.Errorf("expected deck ID 12345, got %d", deck.ID)
	}

	if deck.Name != "Test Deck" {
		t.Errorf("expected deck name 'Test Deck', got '%s'", deck.Name)
	}

	if len(deck.Cards) != 2 {
		t.Errorf("expected 2 cards, got %d", len(deck.Cards))
	}
}

func TestArchidektClient_GetDeck_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := NewArchidektClient()
	client.baseURL = server.URL

	_, err := client.GetDeck(context.Background(), 99999)
	if err == nil {
		t.Fatal("expected error for non-existent deck")
	}
}

func TestArchidektDeck_ExtractCardNames(t *testing.T) {
	deck := &ArchidektDeck{
		Cards: []*ArchidektDeckCard{
			{
				ID:       1,
				Quantity: 4,
				Card: &ArchidektCard{
					OracleCard: &ArchidektOracleCard{Name: "Lightning Bolt"},
				},
			},
			{
				ID:       2,
				Quantity: 4,
				Card: &ArchidektCard{
					OracleCard: &ArchidektOracleCard{Name: "Mountain"},
				},
			},
			{
				ID:       3,
				Quantity: 4,
				Card: &ArchidektCard{
					OracleCard: &ArchidektOracleCard{Name: "Lightning Bolt"}, // Duplicate
				},
			},
		},
	}

	names := deck.ExtractCardNames()

	if len(names) != 2 {
		t.Errorf("expected 2 unique card names, got %d", len(names))
	}

	// Check that both cards are present
	hasLightningBolt := false
	hasMountain := false
	for _, name := range names {
		if name == "Lightning Bolt" {
			hasLightningBolt = true
		}
		if name == "Mountain" {
			hasMountain = true
		}
	}

	if !hasLightningBolt {
		t.Error("expected 'Lightning Bolt' in card names")
	}
	if !hasMountain {
		t.Error("expected 'Mountain' in card names")
	}
}

func TestArchidektDeck_ExtractCardPairs(t *testing.T) {
	deck := &ArchidektDeck{
		Cards: []*ArchidektDeckCard{
			{Card: &ArchidektCard{OracleCard: &ArchidektOracleCard{Name: "A"}}},
			{Card: &ArchidektCard{OracleCard: &ArchidektOracleCard{Name: "B"}}},
			{Card: &ArchidektCard{OracleCard: &ArchidektOracleCard{Name: "C"}}},
		},
	}

	pairs := deck.ExtractCardPairs()

	// 3 cards = 3 pairs: (A,B), (A,C), (B,C)
	if len(pairs) != 3 {
		t.Errorf("expected 3 pairs, got %d", len(pairs))
	}

	// Check that pairs are ordered alphabetically
	for _, pair := range pairs {
		if pair[0] > pair[1] {
			t.Errorf("pair not in alphabetical order: (%s, %s)", pair[0], pair[1])
		}
	}
}

func TestArchidektDeck_GetFormatName(t *testing.T) {
	tests := []struct {
		format   int
		expected string
	}{
		{1, "Standard"},
		{2, "Modern"},
		{3, "Commander"},
		{8, "Pioneer"},
		{9, "Historic"},
		{99, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			deck := &ArchidektDeck{DeckFormat: tt.format}
			if got := deck.GetFormatName(); got != tt.expected {
				t.Errorf("GetFormatName() = %s, want %s", got, tt.expected)
			}
		})
	}
}

func TestFormatNameToID(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{"Standard", 1},
		{"Modern", 2},
		{"Commander", 3},
		{"Historic", 9},
		{"Unknown Format", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatNameToID(tt.name); got != tt.expected {
				t.Errorf("FormatNameToID(%s) = %d, want %d", tt.name, got, tt.expected)
			}
		})
	}
}

func TestNewArchidektClient(t *testing.T) {
	client := NewArchidektClient()

	if client == nil {
		t.Fatal("expected non-nil client")
	}

	if client.httpClient == nil {
		t.Error("expected non-nil httpClient")
	}

	if client.baseURL != "https://archidekt.com/api" {
		t.Errorf("expected base URL 'https://archidekt.com/api', got '%s'", client.baseURL)
	}
}

func TestLocalDeckSource_FetchDecks(t *testing.T) {
	decks := []*SimpleDeck{
		{ID: "1", Name: "Standard Deck 1", Format: "Standard", CardNames: []string{"A", "B"}},
		{ID: "2", Name: "Standard Deck 2", Format: "Standard", CardNames: []string{"C", "D"}},
		{ID: "3", Name: "Historic Deck", Format: "Historic", CardNames: []string{"E", "F"}},
	}

	source := NewLocalDeckSource(decks)

	// Test fetching all formats
	all, err := source.FetchDecks(context.Background(), "all", 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 decks, got %d", len(all))
	}

	// Test fetching by format
	standard, err := source.FetchDecks(context.Background(), "Standard", 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(standard) != 2 {
		t.Errorf("expected 2 Standard decks, got %d", len(standard))
	}

	// Test limit
	limited, err := source.FetchDecks(context.Background(), "all", 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(limited) != 1 {
		t.Errorf("expected 1 deck with limit, got %d", len(limited))
	}
}

func TestLocalDeckSource_SourceName(t *testing.T) {
	source := NewLocalDeckSource(nil)
	if source.SourceName() != "local" {
		t.Errorf("expected source name 'local', got '%s'", source.SourceName())
	}
}
