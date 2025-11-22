package deckimport

import (
	"testing"
)

func TestParseArenaFormat(t *testing.T) {
	parser := NewParser(nil)

	tests := []struct {
		name          string
		input         string
		wantMainboard int
		wantSideboard int
		wantOK        bool
		wantErrors    int
	}{
		{
			name: "valid arena format with sideboard",
			input: `Deck
4 Lightning Bolt (M21) 123
3 Shock (M21) 124
2 Mountain (M21) 275

2 Duress (M21) 95
1 Negate (M21) 56`,
			wantMainboard: 3,
			wantSideboard: 2,
			wantOK:        true,
			wantErrors:    0,
		},
		{
			name: "arena format without set codes",
			input: `Deck
4 Lightning Bolt
3 Shock
2 Mountain`,
			wantMainboard: 3,
			wantSideboard: 0,
			wantOK:        true,
			wantErrors:    0,
		},
		{
			name: "arena format mainboard only",
			input: `4 Lightning Bolt (M21) 123
3 Shock (M21) 124`,
			wantMainboard: 2,
			wantSideboard: 0,
			wantOK:        true,
			wantErrors:    0,
		},
		{
			name:          "empty input",
			input:         "",
			wantMainboard: 0,
			wantSideboard: 0,
			wantOK:        false,
			wantErrors:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParseArenaFormat(tt.input)
			if err != nil {
				t.Fatalf("ParseArenaFormat() error = %v", err)
			}

			if len(result.Deck.Mainboard) != tt.wantMainboard {
				t.Errorf("mainboard cards = %d, want %d", len(result.Deck.Mainboard), tt.wantMainboard)
			}

			if len(result.Deck.Sideboard) != tt.wantSideboard {
				t.Errorf("sideboard cards = %d, want %d", len(result.Deck.Sideboard), tt.wantSideboard)
			}

			if result.Deck.ParsedOK != tt.wantOK {
				t.Errorf("ParsedOK = %v, want %v", result.Deck.ParsedOK, tt.wantOK)
			}

			if len(result.Deck.Errors) != tt.wantErrors {
				t.Errorf("errors = %d, want %d: %v", len(result.Deck.Errors), tt.wantErrors, result.Deck.Errors)
			}
		})
	}
}

func TestParseArenaFormat_Quantities(t *testing.T) {
	parser := NewParser(nil)

	input := `Deck
4 Lightning Bolt (M21) 123
3 Shock (M21) 124
1 Mountain (M21) 275`

	result, err := parser.ParseArenaFormat(input)
	if err != nil {
		t.Fatalf("ParseArenaFormat() error = %v", err)
	}

	expectedQuantities := []int{4, 3, 1}
	for i, card := range result.Deck.Mainboard {
		if card.Quantity != expectedQuantities[i] {
			t.Errorf("card %d quantity = %d, want %d", i, card.Quantity, expectedQuantities[i])
		}
	}
}

func TestParseArenaFormat_CardNames(t *testing.T) {
	parser := NewParser(nil)

	input := `Deck
4 Lightning Bolt (M21) 123
3 Shock (M21) 124`

	result, err := parser.ParseArenaFormat(input)
	if err != nil {
		t.Fatalf("ParseArenaFormat() error = %v", err)
	}

	expectedNames := []string{"Lightning Bolt", "Shock"}
	for i, card := range result.Deck.Mainboard {
		if card.Name != expectedNames[i] {
			t.Errorf("card %d name = %q, want %q", i, card.Name, expectedNames[i])
		}
	}
}

func TestParseArenaFormat_SetCodes(t *testing.T) {
	parser := NewParser(nil)

	input := `Deck
4 Lightning Bolt (M21) 123
3 Shock`

	result, err := parser.ParseArenaFormat(input)
	if err != nil {
		t.Fatalf("ParseArenaFormat() error = %v", err)
	}

	if result.Deck.Mainboard[0].SetCode != "M21" {
		t.Errorf("first card set code = %q, want 'M21'", result.Deck.Mainboard[0].SetCode)
	}

	if result.Deck.Mainboard[1].SetCode != "" {
		t.Errorf("second card set code = %q, want empty", result.Deck.Mainboard[1].SetCode)
	}
}

func TestParsePlainText(t *testing.T) {
	parser := NewParser(nil)

	tests := []struct {
		name          string
		input         string
		wantMainboard int
		wantSideboard int
		wantOK        bool
		wantErrors    int
	}{
		{
			name: "plain text format with x",
			input: `4x Lightning Bolt
3x Shock
2x Mountain`,
			wantMainboard: 3,
			wantSideboard: 0,
			wantOK:        true,
			wantErrors:    0,
		},
		{
			name: "plain text format without x",
			input: `4 Lightning Bolt
3 Shock
2 Mountain`,
			wantMainboard: 3,
			wantSideboard: 0,
			wantOK:        true,
			wantErrors:    0,
		},
		{
			name: "plain text with sideboard",
			input: `4 Lightning Bolt
3 Shock

Sideboard
2 Duress
1 Negate`,
			wantMainboard: 2,
			wantSideboard: 2,
			wantOK:        true,
			wantErrors:    0,
		},
		{
			name: "reverse format card name x4",
			input: `Lightning Bolt x4
Shock x3
Mountain x2`,
			wantMainboard: 3,
			wantSideboard: 0,
			wantOK:        true,
			wantErrors:    0,
		},
		{
			name:          "empty input",
			input:         "",
			wantMainboard: 0,
			wantSideboard: 0,
			wantOK:        false,
			wantErrors:    1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.ParsePlainText(tt.input)
			if err != nil {
				t.Fatalf("ParsePlainText() error = %v", err)
			}

			if len(result.Deck.Mainboard) != tt.wantMainboard {
				t.Errorf("mainboard cards = %d, want %d", len(result.Deck.Mainboard), tt.wantMainboard)
			}

			if len(result.Deck.Sideboard) != tt.wantSideboard {
				t.Errorf("sideboard cards = %d, want %d", len(result.Deck.Sideboard), tt.wantSideboard)
			}

			if result.Deck.ParsedOK != tt.wantOK {
				t.Errorf("ParsedOK = %v, want %v", result.Deck.ParsedOK, tt.wantOK)
			}

			if len(result.Deck.Errors) != tt.wantErrors {
				t.Errorf("errors = %d, want %d", len(result.Deck.Errors), tt.wantErrors)
			}
		})
	}
}

func TestParse_AutoDetect(t *testing.T) {
	parser := NewParser(nil)

	tests := []struct {
		name       string
		input      string
		wantFormat string // "arena" or "plain"
	}{
		{
			name: "detect arena format",
			input: `Deck
4 Lightning Bolt (M21) 123
3 Shock (M21) 124`,
			wantFormat: "arena",
		},
		{
			name: "detect plain text format",
			input: `4x Lightning Bolt
3x Shock`,
			wantFormat: "plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parser.Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if !result.Deck.ParsedOK {
				t.Errorf("Parse() failed to parse %s format", tt.wantFormat)
			}

			if len(result.Deck.Mainboard) == 0 {
				t.Errorf("Parse() found no mainboard cards")
			}
		})
	}
}

func TestValidateDraftImport(t *testing.T) {
	parser := NewParser(nil)

	// Create a mock parsed deck
	result := &ParseResult{
		Deck: &ParsedDeck{
			Mainboard: []*ParsedCard{
				{Name: "Lightning Bolt", Quantity: 4},
				{Name: "Shock", Quantity: 3},
				{Name: "Mountain", Quantity: 10},
			},
			Sideboard: []*ParsedCard{
				{Name: "Duress", Quantity: 2},
			},
		},
		CardIDs: map[string]int{
			"Lightning Bolt": 12345,
			"Shock":          67890,
			"Mountain":       11111,
			"Duress":         22222,
		},
	}

	tests := []struct {
		name          string
		draftCardIDs  []int
		wantErrorsNum int
	}{
		{
			name:          "all cards in draft pool",
			draftCardIDs:  []int{12345, 67890, 11111, 22222, 33333},
			wantErrorsNum: 0,
		},
		{
			name:          "missing one card",
			draftCardIDs:  []int{12345, 67890, 11111}, // Missing Duress (22222)
			wantErrorsNum: 1,
		},
		{
			name:          "missing multiple cards",
			draftCardIDs:  []int{12345}, // Only Lightning Bolt
			wantErrorsNum: 3,            // Shock, Mountain, Duress
		},
		{
			name:          "empty draft pool",
			draftCardIDs:  []int{},
			wantErrorsNum: 4, // All cards missing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := parser.ValidateDraftImport(result, tt.draftCardIDs)
			if len(errors) != tt.wantErrorsNum {
				t.Errorf("ValidateDraftImport() errors = %d, want %d", len(errors), tt.wantErrorsNum)
				for _, err := range errors {
					t.Logf("  Error: %v", err)
				}
			}
		})
	}
}

func TestParsePlainText_Quantities(t *testing.T) {
	parser := NewParser(nil)

	input := `4 Lightning Bolt
3x Shock
Mountain x2`

	result, err := parser.ParsePlainText(input)
	if err != nil {
		t.Fatalf("ParsePlainText() error = %v", err)
	}

	expectedQuantities := []int{4, 3, 2}
	expectedNames := []string{"Lightning Bolt", "Shock", "Mountain"}

	for i, card := range result.Deck.Mainboard {
		if card.Quantity != expectedQuantities[i] {
			t.Errorf("card %d quantity = %d, want %d", i, card.Quantity, expectedQuantities[i])
		}
		if card.Name != expectedNames[i] {
			t.Errorf("card %d name = %q, want %q", i, card.Name, expectedNames[i])
		}
	}
}
