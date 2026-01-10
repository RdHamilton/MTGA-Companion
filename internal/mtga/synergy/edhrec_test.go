package synergy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSanitizeCardName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple card name",
			input:    "Lightning Bolt",
			expected: "lightning-bolt",
		},
		{
			name:     "card name with apostrophe",
			input:    "Sol'kanar the Swamp King",
			expected: "solkanar-the-swamp-king",
		},
		{
			name:     "card name with comma",
			input:    "Venser, Shaper Savant",
			expected: "venser-shaper-savant",
		},
		{
			name:     "card name with colon",
			input:    "Oko, Thief of Crowns: The Card",
			expected: "oko-thief-of-crowns-the-card",
		},
		{
			name:     "already lowercase",
			input:    "murder",
			expected: "murder",
		},
		{
			name:     "multiple spaces",
			input:    "Sword of Fire and Ice",
			expected: "sword-of-fire-and-ice",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeCardName(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeCardName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNewEDHRECClient(t *testing.T) {
	client := NewEDHRECClient()
	if client == nil {
		t.Error("NewEDHRECClient() returned nil")
	}
	if client.httpClient == nil {
		t.Error("NewEDHRECClient() returned client with nil httpClient")
	}
	if client.baseURL == "" {
		t.Error("NewEDHRECClient() returned client with empty baseURL")
	}
}

func TestGetCardSynergy_NotFound(t *testing.T) {
	// Create a test server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewEDHRECClient()
	client.baseURL = server.URL

	_, err := client.GetCardSynergy(context.Background(), "nonexistent-card")
	if err == nil {
		t.Error("Expected error for nonexistent card, got nil")
	}
}

func TestGetCardSynergy_Success(t *testing.T) {
	// Create a test server that returns valid JSON
	responseJSON := `{
		"container": {
			"json_dict": {
				"card": {
					"id": "12345",
					"name": "Sol Ring",
					"sanitized": "sol-ring",
					"cmc": 1,
					"color_identity": [],
					"salt": 0.5,
					"num_decks": 100000
				},
				"cardlists": [
					{
						"tag": "highsynergycards",
						"header": "High Synergy Cards",
						"cardviews": [
							{
								"id": "67890",
								"name": "Arcane Signet",
								"sanitized": "arcane-signet",
								"synergy": 0.8,
								"lift": 1.5,
								"inclusion": 90,
								"num_decks": 50000
							}
						]
					},
					{
						"tag": "topcards",
						"header": "Top Cards",
						"cardviews": [
							{
								"id": "11111",
								"name": "Command Tower",
								"sanitized": "command-tower",
								"synergy": 0.2,
								"inclusion": 95,
								"num_decks": 80000
							}
						]
					}
				]
			}
		},
		"similar": [
			{
				"id": "22222",
				"name": "Mana Crypt",
				"sanitized": "mana-crypt",
				"cmc": 0,
				"color_identity": [],
				"primary_type": "Artifact",
				"rarity": "Mythic"
			}
		]
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	}))
	defer server.Close()

	client := NewEDHRECClient()
	client.baseURL = server.URL

	data, err := client.GetCardSynergy(context.Background(), "sol-ring")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if data == nil {
		t.Fatal("Expected data, got nil")
	}

	if data.CardName != "sol-ring" {
		t.Errorf("CardName = %q, want %q", data.CardName, "sol-ring")
	}

	if data.NumDecks != 100000 {
		t.Errorf("NumDecks = %d, want %d", data.NumDecks, 100000)
	}

	if data.Salt != 0.5 {
		t.Errorf("Salt = %f, want %f", data.Salt, 0.5)
	}

	if len(data.HighSynergy) != 1 {
		t.Errorf("HighSynergy length = %d, want %d", len(data.HighSynergy), 1)
	} else if data.HighSynergy[0].Name != "Arcane Signet" {
		t.Errorf("HighSynergy[0].Name = %q, want %q", data.HighSynergy[0].Name, "Arcane Signet")
	}

	if len(data.TopCards) != 1 {
		t.Errorf("TopCards length = %d, want %d", len(data.TopCards), 1)
	}

	if len(data.SimilarCards) != 1 {
		t.Errorf("SimilarCards length = %d, want %d", len(data.SimilarCards), 1)
	}
}

func TestGetThemeSynergy_Success(t *testing.T) {
	responseJSON := `{
		"container": {
			"json_dict": {
				"cardlists": [
					{
						"tag": "highsynergycards",
						"header": "High Synergy Cards",
						"cardviews": [
							{
								"id": "12345",
								"name": "Smothering Tithe",
								"synergy": 0.9
							}
						]
					},
					{
						"tag": "topcards",
						"header": "Top Cards",
						"cardviews": [
							{
								"id": "67890",
								"name": "Anointed Procession",
								"synergy": 0.7
							}
						]
					}
				]
			}
		},
		"header": "Tokens",
		"description": "Decks that focus on creating token creatures"
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	}))
	defer server.Close()

	client := NewEDHRECClient()
	client.baseURL = server.URL

	data, err := client.GetThemeSynergy(context.Background(), "tokens")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if data == nil {
		t.Fatal("Expected data, got nil")
	}

	if data.ThemeName != "tokens" {
		t.Errorf("ThemeName = %q, want %q", data.ThemeName, "tokens")
	}

	if len(data.HighSynergy) != 1 {
		t.Errorf("HighSynergy length = %d, want %d", len(data.HighSynergy), 1)
	}

	if len(data.TopCards) != 1 {
		t.Errorf("TopCards length = %d, want %d", len(data.TopCards), 1)
	}
}

func TestGetHighSynergyCards_WithLimit(t *testing.T) {
	responseJSON := `{
		"container": {
			"json_dict": {
				"card": {
					"id": "12345",
					"name": "Test Card"
				},
				"cardlists": [
					{
						"tag": "highsynergycards",
						"cardviews": [
							{"id": "1", "name": "Card 1", "synergy": 0.9},
							{"id": "2", "name": "Card 2", "synergy": 0.8},
							{"id": "3", "name": "Card 3", "synergy": 0.7},
							{"id": "4", "name": "Card 4", "synergy": 0.6},
							{"id": "5", "name": "Card 5", "synergy": 0.5}
						]
					}
				]
			}
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(responseJSON))
	}))
	defer server.Close()

	client := NewEDHRECClient()
	client.baseURL = server.URL

	// Test with limit
	cards, err := client.GetHighSynergyCards(context.Background(), "test-card", 3)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(cards) != 3 {
		t.Errorf("GetHighSynergyCards with limit 3 returned %d cards, want 3", len(cards))
	}

	// Test without limit (0 means all)
	cards, err = client.GetHighSynergyCards(context.Background(), "test-card", 0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(cards) != 5 {
		t.Errorf("GetHighSynergyCards with limit 0 returned %d cards, want 5", len(cards))
	}
}

func TestKnownThemes(t *testing.T) {
	if len(KnownThemes) == 0 {
		t.Error("KnownThemes should not be empty")
	}

	// Check for some expected themes
	expectedThemes := []string{"tokens", "aristocrats", "counters", "lifegain", "mill"}
	for _, expected := range expectedThemes {
		found := false
		for _, theme := range KnownThemes {
			if theme == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected theme %q not found in KnownThemes", expected)
		}
	}
}

func TestExtractSynergyData_EmptyContainer(t *testing.T) {
	client := NewEDHRECClient()

	// Test with nil container
	page := &EDHRECCardPage{
		Container: nil,
	}
	data := client.extractSynergyData("test", page)

	if data.CardName != "test" {
		t.Errorf("CardName = %q, want %q", data.CardName, "test")
	}
	if len(data.HighSynergy) != 0 {
		t.Errorf("HighSynergy should be empty, got %d items", len(data.HighSynergy))
	}

	// Test with nil json_dict
	page = &EDHRECCardPage{
		Container: &EDHRECContainer{
			JSONDict: nil,
		},
	}
	data = client.extractSynergyData("test", page)
	if len(data.HighSynergy) != 0 {
		t.Errorf("HighSynergy should be empty, got %d items", len(data.HighSynergy))
	}
}

func TestExtractThemeData_EmptyContainer(t *testing.T) {
	client := NewEDHRECClient()

	// Test with nil container
	page := &EDHRECThemePage{
		Container:   nil,
		Description: "Test description",
	}
	data := client.extractThemeData("test-theme", page)

	if data.ThemeName != "test-theme" {
		t.Errorf("ThemeName = %q, want %q", data.ThemeName, "test-theme")
	}
	if data.Description != "Test description" {
		t.Errorf("Description = %q, want %q", data.Description, "Test description")
	}
	if len(data.HighSynergy) != 0 {
		t.Errorf("HighSynergy should be empty, got %d items", len(data.HighSynergy))
	}
}
