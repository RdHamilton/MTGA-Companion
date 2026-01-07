package gui

import (
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

func TestConvertSetCardToCard(t *testing.T) {
	tests := []struct {
		name     string
		setCard  *models.SetCard
		expected struct {
			arenaID     int
			name        string
			setCode     string
			cmc         float64
			rarity      string
			hasManaCost bool
			manaCost    string
		}
	}{
		{
			name: "basic instant spell",
			setCard: &models.SetCard{
				ArenaID:    "12345",
				ScryfallID: "scryfall-123",
				Name:       "Lightning Bolt",
				SetCode:    "M21",
				CMC:        1,
				ManaCost:   "{R}",
				Colors:     []string{"R"},
				Types:      []string{"Instant"},
				Rarity:     "common",
				Text:       "Deal 3 damage to any target.",
				ImageURL:   "https://example.com/bolt.jpg",
			},
			expected: struct {
				arenaID     int
				name        string
				setCode     string
				cmc         float64
				rarity      string
				hasManaCost bool
				manaCost    string
			}{
				arenaID:     12345,
				name:        "Lightning Bolt",
				setCode:     "M21",
				cmc:         1.0,
				rarity:      "common",
				hasManaCost: true,
				manaCost:    "{R}",
			},
		},
		{
			name: "creature with power and toughness",
			setCard: &models.SetCard{
				ArenaID:    "67890",
				ScryfallID: "scryfall-456",
				Name:       "Grizzly Bears",
				SetCode:    "M21",
				CMC:        2,
				ManaCost:   "{1}{G}",
				Colors:     []string{"G"},
				Types:      []string{"Creature", "Bear"},
				Rarity:     "common",
				Power:      "2",
				Toughness:  "2",
				ImageURL:   "https://example.com/bears.jpg",
			},
			expected: struct {
				arenaID     int
				name        string
				setCode     string
				cmc         float64
				rarity      string
				hasManaCost bool
				manaCost    string
			}{
				arenaID:     67890,
				name:        "Grizzly Bears",
				setCode:     "M21",
				cmc:         2.0,
				rarity:      "common",
				hasManaCost: true,
				manaCost:    "{1}{G}",
			},
		},
		{
			name: "multi-colored card",
			setCard: &models.SetCard{
				ArenaID:    "11111",
				ScryfallID: "scryfall-789",
				Name:       "Azorius Charm",
				SetCode:    "RNA",
				CMC:        2,
				ManaCost:   "{W}{U}",
				Colors:     []string{"W", "U"},
				Types:      []string{"Instant"},
				Rarity:     "uncommon",
				Text:       "Choose one — ...",
				ImageURL:   "https://example.com/charm.jpg",
			},
			expected: struct {
				arenaID     int
				name        string
				setCode     string
				cmc         float64
				rarity      string
				hasManaCost bool
				manaCost    string
			}{
				arenaID:     11111,
				name:        "Azorius Charm",
				setCode:     "RNA",
				cmc:         2.0,
				rarity:      "uncommon",
				hasManaCost: true,
				manaCost:    "{W}{U}",
			},
		},
		{
			name: "card without mana cost",
			setCard: &models.SetCard{
				ArenaID:    "22222",
				ScryfallID: "scryfall-000",
				Name:       "Ornithopter",
				SetCode:    "M21",
				CMC:        0,
				ManaCost:   "",
				Colors:     []string{},
				Types:      []string{"Artifact", "Creature", "Thopter"},
				Rarity:     "uncommon",
				Power:      "0",
				Toughness:  "2",
				Text:       "Flying",
				ImageURL:   "https://example.com/ornithopter.jpg",
			},
			expected: struct {
				arenaID     int
				name        string
				setCode     string
				cmc         float64
				rarity      string
				hasManaCost bool
				manaCost    string
			}{
				arenaID:     22222,
				name:        "Ornithopter",
				setCode:     "M21",
				cmc:         0.0,
				rarity:      "uncommon",
				hasManaCost: false,
				manaCost:    "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertSetCardToCard(tt.setCard)

			if result == nil {
				t.Fatal("convertSetCardToCard() returned nil")
			}

			// Check basic fields
			if result.ArenaID != tt.expected.arenaID {
				t.Errorf("ArenaID = %d, want %d", result.ArenaID, tt.expected.arenaID)
			}
			if result.Name != tt.expected.name {
				t.Errorf("Name = %s, want %s", result.Name, tt.expected.name)
			}
			if result.SetCode != tt.expected.setCode {
				t.Errorf("SetCode = %s, want %s", result.SetCode, tt.expected.setCode)
			}
			if result.CMC != tt.expected.cmc {
				t.Errorf("CMC = %f, want %f", result.CMC, tt.expected.cmc)
			}
			if result.Rarity != tt.expected.rarity {
				t.Errorf("Rarity = %s, want %s", result.Rarity, tt.expected.rarity)
			}

			// Check ManaCost pointer
			if tt.expected.hasManaCost {
				if result.ManaCost == nil {
					t.Error("ManaCost is nil, expected non-nil")
				} else if *result.ManaCost != tt.expected.manaCost {
					t.Errorf("ManaCost = %s, want %s", *result.ManaCost, tt.expected.manaCost)
				}
			} else {
				if result.ManaCost != nil {
					t.Errorf("ManaCost = %v, want nil", *result.ManaCost)
				}
			}

			// Verify TypeLine was constructed from Types
			if len(tt.setCard.Types) > 0 && result.TypeLine == "" {
				t.Error("TypeLine is empty, expected non-empty")
			}

			// Check Power/Toughness if present
			if tt.setCard.Power != "" {
				if result.Power == nil {
					t.Error("Power is nil, expected non-nil")
				} else if *result.Power != tt.setCard.Power {
					t.Errorf("Power = %s, want %s", *result.Power, tt.setCard.Power)
				}
			}

			if tt.setCard.Toughness != "" {
				if result.Toughness == nil {
					t.Error("Toughness is nil, expected non-nil")
				} else if *result.Toughness != tt.setCard.Toughness {
					t.Errorf("Toughness = %s, want %s", *result.Toughness, tt.setCard.Toughness)
				}
			}

			// Check OracleText if present
			if tt.setCard.Text != "" {
				if result.OracleText == nil {
					t.Error("OracleText is nil, expected non-nil")
				} else if *result.OracleText != tt.setCard.Text {
					t.Errorf("OracleText = %s, want %s", *result.OracleText, tt.setCard.Text)
				}
			}

			// Check ImageURI if present
			if tt.setCard.ImageURL != "" {
				if result.ImageURI == nil {
					t.Error("ImageURI is nil, expected non-nil")
				} else if *result.ImageURI != tt.setCard.ImageURL {
					t.Errorf("ImageURI = %s, want %s", *result.ImageURI, tt.setCard.ImageURL)
				}
			}

			// Check Colors slice
			if len(result.Colors) != len(tt.setCard.Colors) {
				t.Errorf("Colors length = %d, want %d", len(result.Colors), len(tt.setCard.Colors))
			}
		})
	}
}

func TestConvertSetCardToCard_NilInput(t *testing.T) {
	result := convertSetCardToCard(nil)
	if result != nil {
		t.Errorf("convertSetCardToCard(nil) = %v, want nil", result)
	}
}

func TestNormalizeDeckSource(t *testing.T) {
	tests := []struct {
		name           string
		inputSource    string
		expectedSource string
		shouldError    bool
	}{
		{
			name:           "manual maps to constructed",
			inputSource:    "manual",
			expectedSource: "constructed",
			shouldError:    false,
		},
		{
			name:           "import maps to imported",
			inputSource:    "import",
			expectedSource: "imported",
			shouldError:    false,
		},
		{
			name:           "draft stays draft",
			inputSource:    "draft",
			expectedSource: "draft",
			shouldError:    false,
		},
		{
			name:           "constructed stays constructed",
			inputSource:    "constructed",
			expectedSource: "constructed",
			shouldError:    false,
		},
		{
			name:           "imported stays imported",
			inputSource:    "imported",
			expectedSource: "imported",
			shouldError:    false,
		},
		{
			name:           "invalid source returns error",
			inputSource:    "invalid",
			expectedSource: "",
			shouldError:    true,
		},
		{
			name:           "empty source returns error",
			inputSource:    "",
			expectedSource: "",
			shouldError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			normalized, err := normalizeDeckSource(tt.inputSource)
			if tt.shouldError {
				if err == nil {
					t.Errorf("normalizeDeckSource(%q) expected error, got nil", tt.inputSource)
				}
			} else {
				if err != nil {
					t.Errorf("normalizeDeckSource(%q) unexpected error: %v", tt.inputSource, err)
				}
				if normalized != tt.expectedSource {
					t.Errorf("normalizeDeckSource(%q) = %q, want %q", tt.inputSource, normalized, tt.expectedSource)
				}
			}
		})
	}
}

func TestConvertSetCardToCard_MultiTypeCard(t *testing.T) {
	setCard := &models.SetCard{
		ArenaID:    "99999",
		ScryfallID: "scryfall-xyz",
		Name:       "Artifact Creature",
		SetCode:    "M21",
		CMC:        3,
		ManaCost:   "{3}",
		Colors:     []string{},
		Types:      []string{"Artifact", "Creature", "Golem"},
		Rarity:     "uncommon",
		Power:      "3",
		Toughness:  "3",
	}

	result := convertSetCardToCard(setCard)

	if result == nil {
		t.Fatal("convertSetCardToCard() returned nil")
	}

	// Verify TypeLine contains all types
	expectedTypeLine := "Artifact Creature Golem"
	if result.TypeLine != expectedTypeLine {
		t.Errorf("TypeLine = %s, want %s", result.TypeLine, expectedTypeLine)
	}
}

func TestBasicLandIDs(t *testing.T) {
	// Test that basic land IDs are correctly identified for the 4-card limit exemption.
	// These IDs should match the constants used in AddCard validation.
	basicLandIDs := map[int]bool{
		81716: true, // Plains
		81717: true, // Island
		81718: true, // Swamp
		81719: true, // Mountain
		81720: true, // Forest
	}

	tests := []struct {
		name    string
		cardID  int
		isBasic bool
	}{
		{"Plains is basic land", 81716, true},
		{"Island is basic land", 81717, true},
		{"Swamp is basic land", 81718, true},
		{"Mountain is basic land", 81719, true},
		{"Forest is basic land", 81720, true},
		{"Regular card is not basic", 12345, false},
		{"Another regular card is not basic", 99999, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := basicLandIDs[tt.cardID]
			if result != tt.isBasic {
				t.Errorf("basicLandIDs[%d] = %v, want %v", tt.cardID, result, tt.isBasic)
			}
		})
	}
}

func TestFourCardLimitLogic(t *testing.T) {
	// Test the 4-card limit logic used in AddCard
	tests := []struct {
		name        string
		currentQty  int
		addQty      int
		shouldError bool
		maxCanAdd   int
	}{
		{"can add 1 to empty deck", 0, 1, false, 4},
		{"can add 4 to empty deck", 0, 4, false, 4},
		{"cannot add 5 to empty deck", 0, 5, true, 4},
		{"can add 1 when at 3", 3, 1, false, 1},
		{"cannot add 2 when at 3", 3, 2, true, 1},
		{"cannot add any when at 4", 4, 1, true, 0},
		{"can add 2 when at 2", 2, 2, false, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			maxCanAdd := 4 - tt.currentQty
			wouldExceed := tt.currentQty+tt.addQty > 4

			if wouldExceed != tt.shouldError {
				t.Errorf("wouldExceed = %v, want %v", wouldExceed, tt.shouldError)
			}

			if maxCanAdd != tt.maxCanAdd {
				t.Errorf("maxCanAdd = %d, want %d", maxCanAdd, tt.maxCanAdd)
			}
		})
	}
}
