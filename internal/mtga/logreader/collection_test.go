package logreader

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseCollection(t *testing.T) {
	tests := []struct {
		name    string
		entry   *LogEntry
		want    *CardCollection
		wantErr bool
		wantNil bool
	}{
		{
			name: "no collection data",
			entry: &LogEntry{
				IsJSON: true,
				JSON: map[string]interface{}{
					"InventoryInfo": map[string]interface{}{
						"Gems": float64(1000),
						"Gold": float64(5000),
					},
				},
			},
			wantNil: true,
		},
		{
			name: "collection with cards as map",
			entry: &LogEntry{
				IsJSON: true,
				JSON: map[string]interface{}{
					"InventoryInfo": map[string]interface{}{
						"Cards": map[string]interface{}{
							"12345": float64(4),
							"67890": float64(2),
							"11111": float64(1),
						},
					},
				},
			},
			want: &CardCollection{
				Cards: map[int]*Card{
					12345: {CardID: 12345, Quantity: 4, MaxQuantity: 4},
					67890: {CardID: 67890, Quantity: 2, MaxQuantity: 4},
					11111: {CardID: 11111, Quantity: 1, MaxQuantity: 4},
				},
				TotalCards:  7,
				UniqueCards: 3,
			},
		},
		{
			name: "collection with cards as array",
			entry: &LogEntry{
				IsJSON: true,
				JSON: map[string]interface{}{
					"InventoryInfo": map[string]interface{}{
						"CardInventory": []interface{}{
							map[string]interface{}{
								"cardId":   float64(12345),
								"quantity": float64(4),
							},
							map[string]interface{}{
								"CardId":   float64(67890),
								"Quantity": float64(2),
							},
						},
					},
				},
			},
			want: &CardCollection{
				Cards: map[int]*Card{
					12345: {CardID: 12345, Quantity: 4, MaxQuantity: 4},
					67890: {CardID: 67890, Quantity: 2, MaxQuantity: 4},
				},
				TotalCards:  6,
				UniqueCards: 2,
			},
		},
		{
			name: "empty collection",
			entry: &LogEntry{
				IsJSON: true,
				JSON: map[string]interface{}{
					"InventoryInfo": map[string]interface{}{
						"Cards": map[string]interface{}{},
					},
				},
			},
			wantNil: true,
		},
		{
			name: "non-JSON entry",
			entry: &LogEntry{
				IsJSON: false,
				JSON:   nil,
			},
			wantNil: true,
		},
		{
			name: "no InventoryInfo",
			entry: &LogEntry{
				IsJSON: true,
				JSON: map[string]interface{}{
					"otherEvent": map[string]interface{}{},
				},
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := []*LogEntry{tt.entry}
			got, err := ParseCollection(entries)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseCollection() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseCollection() error = %v", err)
				return
			}

			if tt.wantNil {
				if got != nil {
					t.Errorf("ParseCollection() expected nil, got %v", got)
				}
				return
			}

			if got == nil {
				t.Errorf("ParseCollection() expected collection, got nil")
				return
			}

			// Check total cards
			if got.TotalCards != tt.want.TotalCards {
				t.Errorf("ParseCollection() TotalCards = %d, want %d", got.TotalCards, tt.want.TotalCards)
			}

			// Check unique cards
			if got.UniqueCards != tt.want.UniqueCards {
				t.Errorf("ParseCollection() UniqueCards = %d, want %d", got.UniqueCards, tt.want.UniqueCards)
			}

			// Check card count
			if len(got.Cards) != len(tt.want.Cards) {
				t.Errorf("ParseCollection() card count = %d, want %d", len(got.Cards), len(tt.want.Cards))
			}

			// Check individual cards
			for cardID, wantCard := range tt.want.Cards {
				gotCard, ok := got.Cards[cardID]
				if !ok {
					t.Errorf("ParseCollection() missing card ID %d", cardID)
					continue
				}

				if gotCard.CardID != wantCard.CardID {
					t.Errorf("ParseCollection() card %d CardID = %d, want %d", cardID, gotCard.CardID, wantCard.CardID)
				}

				if gotCard.Quantity != wantCard.Quantity {
					t.Errorf("ParseCollection() card %d Quantity = %d, want %d", cardID, gotCard.Quantity, wantCard.Quantity)
				}
			}
		})
	}
}

func TestParseCollection_FromLogFile(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test_player.log")

	// Create test log data with collection
	testData := `[UnityCrossThreadLogger]{"InventoryInfo":{"Cards":{"12345":4,"67890":2,"11111":1}}}
[UnityCrossThreadLogger]{"InventoryInfo":{"Gems":1000,"Gold":5000}}
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

	// Parse collection
	collection, err := ParseCollection(entries)
	if err != nil {
		t.Fatalf("ParseCollection() error = %v", err)
	}

	if collection == nil {
		t.Fatal("ParseCollection() expected collection, got nil")
	}

	if collection.TotalCards != 7 {
		t.Errorf("ParseCollection() TotalCards = %d, want 7", collection.TotalCards)
	}

	if collection.UniqueCards != 3 {
		t.Errorf("ParseCollection() UniqueCards = %d, want 3", collection.UniqueCards)
	}

	// Check specific cards
	if card, ok := collection.Cards[12345]; !ok || card.Quantity != 4 {
		t.Errorf("ParseCollection() card 12345 not found or wrong quantity")
	}

	if card, ok := collection.Cards[67890]; !ok || card.Quantity != 2 {
		t.Errorf("ParseCollection() card 67890 not found or wrong quantity")
	}

	if card, ok := collection.Cards[11111]; !ok || card.Quantity != 1 {
		t.Errorf("ParseCollection() card 11111 not found or wrong quantity")
	}
}
