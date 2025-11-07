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
			name: "decks as array",
			entry: &LogEntry{
				IsJSON: true,
				JSON: map[string]interface{}{
					"Deck.GetDeckLists": []interface{}{
						map[string]interface{}{
							"id":     "deck-1",
							"name":   "Test Deck",
							"format": "Standard",
							"mainDeck": []interface{}{
								map[string]interface{}{
									"cardId":   float64(12345),
									"quantity": float64(4),
								},
							},
						},
					},
				},
			},
			want: &DeckLibrary{
				Decks: map[string]*PlayerDeck{
					"deck-1": {
						DeckID:   "deck-1",
						Name:     "Test Deck",
						Format:   "Standard",
						MainDeck: []DeckCard{{CardID: 12345, Quantity: 4}},
					},
				},
				TotalDecks: 1,
			},
		},
		{
			name: "decks in object",
			entry: &LogEntry{
				IsJSON: true,
				JSON: map[string]interface{}{
					"getDeckLists": map[string]interface{}{
						"decks": []interface{}{
							map[string]interface{}{
								"Id":     "deck-2",
								"Name":   "Another Deck",
								"Format": "Historic",
							},
						},
					},
				},
			},
			want: &DeckLibrary{
				Decks: map[string]*PlayerDeck{
					"deck-2": {
						DeckID: "deck-2",
						Name:   "Another Deck",
						Format: "Historic",
					},
				},
				TotalDecks: 1,
			},
		},
		{
			name: "empty deck list",
			entry: &LogEntry{
				IsJSON: true,
				JSON: map[string]interface{}{
					"Deck.GetDeckLists": []interface{}{},
				},
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := []*LogEntry{tt.entry}
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

	// Create test log data with decks
	testData := `[UnityCrossThreadLogger]{"Deck.GetDeckLists":[{"id":"deck-1","name":"Test Deck","format":"Standard","mainDeck":[{"cardId":12345,"quantity":4}]}]}
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
