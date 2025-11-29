package logreader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseDecks(t *testing.T) {
	tests := []struct {
		name    string
		entry   *LogEntry
		want    *DeckLibrary
		wantNil bool
	}{
		{
			name: "no deck data",
			entry: &LogEntry{
				IsJSON: true,
				JSON: map[string]interface{}{
					"otherEvent": map[string]interface{}{},
				},
			},
			wantNil: true,
		},
		{
			name: "DeckUpsertDeckV2 with deck data",
			entry: &LogEntry{
				IsJSON: true,
				JSON: map[string]interface{}{
					"id":      "123",
					"request": `{"Summary":{"DeckId":"deck-1","Name":"Test Deck","Description":"Decks/Test/TestDeck","Attributes":[{"name":"Format","value":"Standard"},{"name":"LastPlayed","value":"\"2024-06-21T09:35:17.8958228-04:00\""}]},"Deck":{"MainDeck":[{"cardId":12345,"quantity":4}],"Sideboard":[]}}`,
				},
			},
			want: &DeckLibrary{
				Decks: map[string]*PlayerDeck{
					"deck-1": {
						DeckID:      "deck-1",
						Name:        "Test Deck",
						Description: "Decks/Test/TestDeck",
						Format:      "Standard",
						MainDeck:    []DeckCard{{CardID: 12345, Quantity: 4}},
						Sideboard:   []DeckCard{},
					},
				},
				TotalDecks: 1,
			},
		},
		{
			name:  "multiple DeckUpsertDeckV2 entries",
			entry: nil, // Will be handled with multiple entries
			want: &DeckLibrary{
				Decks: map[string]*PlayerDeck{
					"deck-1": {
						DeckID:    "deck-1",
						Name:      "Explorer Deck",
						Format:    "Explorer",
						MainDeck:  []DeckCard{},
						Sideboard: []DeckCard{},
					},
					"deck-2": {
						DeckID:    "deck-2",
						Name:      "Historic Deck",
						Format:    "Historic",
						MainDeck:  []DeckCard{},
						Sideboard: []DeckCard{},
					},
				},
				TotalDecks: 2,
			},
		},
		{
			name: "empty request field",
			entry: &LogEntry{
				IsJSON: true,
				JSON: map[string]interface{}{
					"id":      "123",
					"request": "",
				},
			},
			wantNil: true,
		},
		{
			name: "request without Summary",
			entry: &LogEntry{
				IsJSON: true,
				JSON: map[string]interface{}{
					"id":      "123",
					"request": `{"Deck":{"MainDeck":[]}}`,
				},
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var entries []*LogEntry

			// Handle the multiple entries test case
			if tt.name == "multiple DeckUpsertDeckV2 entries" {
				entries = []*LogEntry{
					{
						IsJSON: true,
						JSON: map[string]interface{}{
							"id":      "123",
							"request": `{"Summary":{"DeckId":"deck-1","Name":"Explorer Deck","Attributes":[{"name":"Format","value":"Explorer"}]},"Deck":{"MainDeck":[],"Sideboard":[]}}`,
						},
					},
					{
						IsJSON: true,
						JSON: map[string]interface{}{
							"id":      "124",
							"request": `{"Summary":{"DeckId":"deck-2","Name":"Historic Deck","Attributes":[{"name":"Format","value":"Historic"}]},"Deck":{"MainDeck":[],"Sideboard":[]}}`,
						},
					},
				}
			} else {
				entries = []*LogEntry{tt.entry}
			}

			got, err := ParseDecks(entries)
			if err != nil {
				t.Errorf("ParseDecks() error = %v", err)
				return
			}

			if tt.wantNil {
				if got != nil {
					t.Errorf("ParseDecks() expected nil, got %v", got)
				}
				return
			}

			if got == nil {
				t.Errorf("ParseDecks() expected library, got nil")
				return
			}

			// Check total decks
			if got.TotalDecks != tt.want.TotalDecks {
				t.Errorf("ParseDecks() TotalDecks = %d, want %d", got.TotalDecks, tt.want.TotalDecks)
			}

			// Check deck count
			if len(got.Decks) != len(tt.want.Decks) {
				t.Errorf("ParseDecks() deck count = %d, want %d", len(got.Decks), len(tt.want.Decks))
			}

			// Check individual decks
			for deckID, wantDeck := range tt.want.Decks {
				gotDeck, ok := got.Decks[deckID]
				if !ok {
					t.Errorf("ParseDecks() missing deck ID %s", deckID)
					continue
				}

				if gotDeck.DeckID != wantDeck.DeckID {
					t.Errorf("ParseDecks() deck %s DeckID = %s, want %s", deckID, gotDeck.DeckID, wantDeck.DeckID)
				}

				if gotDeck.Name != wantDeck.Name {
					t.Errorf("ParseDecks() deck %s Name = %s, want %s", deckID, gotDeck.Name, wantDeck.Name)
				}

				if gotDeck.Format != wantDeck.Format {
					t.Errorf("ParseDecks() deck %s Format = %s, want %s", deckID, gotDeck.Format, wantDeck.Format)
				}
			}
		})
	}
}

func TestParseDecks_FromLogFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test_player.log")

	// Create test log data with DeckUpsertDeckV2 format
	// Note: The request field is a JSON string, so we need to escape the inner quotes
	testData := `[UnityCrossThreadLogger]{"id":"123","request":"{\"Summary\":{\"DeckId\":\"deck-1\",\"Name\":\"Test Deck\",\"Attributes\":[{\"name\":\"Format\",\"value\":\"Standard\"}]},\"Deck\":{\"MainDeck\":[{\"cardId\":12345,\"quantity\":4}],\"Sideboard\":[]}}"}
`
	if err := os.WriteFile(logPath, []byte(testData), 0o644); err != nil {
		t.Fatalf("Failed to create test log file: %v", err)
	}

	// Read entries
	reader, err := NewReader(logPath)
	if err != nil {
		t.Fatalf("Failed to create reader: %v", err)
	}
	defer func() {
		if err := reader.Close(); err != nil {
			t.Errorf("Error closing reader: %v", err)
		}
	}()

	entries, err := reader.ReadAllJSON()
	if err != nil {
		t.Fatalf("Failed to read entries: %v", err)
	}

	// Parse decks
	library, err := ParseDecks(entries)
	if err != nil {
		t.Fatalf("ParseDecks() error = %v", err)
	}

	if library == nil {
		t.Fatal("ParseDecks() expected library, got nil")
	}

	if library.TotalDecks != 1 {
		t.Errorf("ParseDecks() TotalDecks = %d, want 1", library.TotalDecks)
	}

	// Check specific deck
	if deck, ok := library.Decks["deck-1"]; !ok {
		t.Error("ParseDecks() deck-1 not found")
	} else if deck.Name != "Test Deck" {
		t.Errorf("ParseDecks() deck name = %s, want Test Deck", deck.Name)
	}
}

func TestCleanDeckName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "regular name unchanged",
			input:    "My Cool Deck",
			expected: "My Cool Deck",
		},
		{
			name:     "precon localization key",
			input:    "?=?Loc/Decks/Precon/Precon_EPP2024_UW",
			expected: "Precon EPP2024 UW",
		},
		{
			name:     "precon with player name",
			input:    "?=?Loc/Decks/Precon/2022_WC/Player_JanM",
			expected: "Player JanM",
		},
		{
			name:     "precon NPE historic brawl",
			input:    "?=?Loc/Decks/Precon/Precon_NPE_HistoricBrawl_GW",
			expected: "Precon NPE HistoricBrawl GW",
		},
		{
			name:     "color codes deck",
			input:    "?=?Loc/Decks/Precon/CC_ANB_W",
			expected: "CC ANB W",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "non-deck localization key",
			input:    "?=?Loc/Other/Something",
			expected: "Something",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanDeckName(tt.input)
			if result != tt.expected {
				t.Errorf("cleanDeckName(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
