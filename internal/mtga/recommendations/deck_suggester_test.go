package recommendations

import (
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
)

func TestFilterByColorFit(t *testing.T) {
	suggester := &DeckSuggester{}

	// Create test cards
	poolCards := []*cards.Card{
		{ArenaID: 1, Name: "White Knight", Colors: []string{"W"}, TypeLine: "Creature"},
		{ArenaID: 2, Name: "Blue Mage", Colors: []string{"U"}, TypeLine: "Creature"},
		{ArenaID: 3, Name: "Azorius Charm", Colors: []string{"W", "U"}, TypeLine: "Instant"},
		{ArenaID: 4, Name: "Firebolt", Colors: []string{"R"}, TypeLine: "Sorcery"},
		{ArenaID: 5, Name: "Sol Ring", Colors: []string{}, TypeLine: "Artifact"},
		{ArenaID: 6, Name: "Plains", Colors: []string{}, TypeLine: "Basic Land"},
	}

	tests := []struct {
		name          string
		combo         ColorCombination
		expectedCount int
		expectedNames []string
	}{
		{
			name:          "Two-color WU filters correctly",
			combo:         ColorCombination{Colors: []string{"W", "U"}, Name: "Azorius"},
			expectedCount: 4, // White Knight, Blue Mage, Azorius Charm, Sol Ring (lands filtered)
			expectedNames: []string{"White Knight", "Blue Mage", "Azorius Charm", "Sol Ring"},
		},
		{
			name:          "Mono-W filters correctly",
			combo:         ColorCombination{Colors: []string{"W"}, Name: "Mono-White"},
			expectedCount: 2, // White Knight, Sol Ring
			expectedNames: []string{"White Knight", "Sol Ring"},
		},
		{
			name:          "RG pair excludes WU cards",
			combo:         ColorCombination{Colors: []string{"R", "G"}, Name: "Gruul"},
			expectedCount: 2, // Firebolt, Sol Ring
			expectedNames: []string{"Firebolt", "Sol Ring"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := suggester.filterByColorFit(poolCards, tt.combo)
			if len(result) != tt.expectedCount {
				t.Errorf("expected %d cards, got %d", tt.expectedCount, len(result))
			}

			// Check that expected names are present
			for _, expectedName := range tt.expectedNames {
				found := false
				for _, card := range result {
					if card.Name == expectedName {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected card %s not found in result", expectedName)
				}
			}
		})
	}
}

func TestIsViable(t *testing.T) {
	suggester := &DeckSuggester{}

	tests := []struct {
		name       string
		candidates []*cards.Card
		expected   bool
	}{
		{
			name:       "Not viable - too few cards",
			candidates: make([]*cards.Card, 10),
			expected:   false,
		},
		{
			name: "Not viable - not enough creatures",
			candidates: func() []*cards.Card {
				result := make([]*cards.Card, 20)
				for i := range result {
					result[i] = &cards.Card{TypeLine: "Instant"}
				}
				return result
			}(),
			expected: false,
		},
		{
			name: "Viable - enough cards and creatures",
			candidates: func() []*cards.Card {
				result := make([]*cards.Card, 20)
				for i := range result {
					if i < 12 {
						result[i] = &cards.Card{TypeLine: "Creature"}
					} else {
						result[i] = &cards.Card{TypeLine: "Instant"}
					}
				}
				return result
			}(),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := suggester.isViable(tt.candidates)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSelectBestCards(t *testing.T) {
	suggester := &DeckSuggester{}

	// Create scored cards with varying CMCs and scores
	scoredCards := []*scoredCard{
		{card: &cards.Card{ArenaID: 1, Name: "Card1", CMC: 2}, score: 0.9, reasoning: "best"},
		{card: &cards.Card{ArenaID: 2, Name: "Card2", CMC: 2}, score: 0.85, reasoning: "good"},
		{card: &cards.Card{ArenaID: 3, Name: "Card3", CMC: 3}, score: 0.8, reasoning: "good"},
		{card: &cards.Card{ArenaID: 4, Name: "Card4", CMC: 3}, score: 0.75, reasoning: "ok"},
		{card: &cards.Card{ArenaID: 5, Name: "Card5", CMC: 4}, score: 0.7, reasoning: "ok"},
		{card: &cards.Card{ArenaID: 6, Name: "Card6", CMC: 5}, score: 0.65, reasoning: "ok"},
		{card: &cards.Card{ArenaID: 7, Name: "Card7", CMC: 1}, score: 0.6, reasoning: "filler"},
	}

	tests := []struct {
		name        string
		targetCount int
	}{
		{
			name:        "Select 5 cards",
			targetCount: 5,
		},
		{
			name:        "Select all available",
			targetCount: 7,
		},
		{
			name:        "Select more than available",
			targetCount: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := suggester.selectBestCards(scoredCards, tt.targetCount)

			expectedLen := tt.targetCount
			if expectedLen > len(scoredCards) {
				expectedLen = len(scoredCards)
			}

			if len(result) != expectedLen {
				t.Errorf("expected %d cards, got %d", expectedLen, len(result))
			}

			// Verify no duplicates
			seen := make(map[int]bool)
			for _, sc := range result {
				if seen[sc.card.ArenaID] {
					t.Errorf("duplicate card found: %s", sc.card.Name)
				}
				seen[sc.card.ArenaID] = true
			}
		})
	}
}

func TestDistributeLands(t *testing.T) {
	suggester := &DeckSuggester{}

	tests := []struct {
		name          string
		selectedCards []*scoredCard
		combo         ColorCombination
		expectedTotal int
	}{
		{
			name: "Two-color distribution with mana costs",
			selectedCards: []*scoredCard{
				{card: &cards.Card{ManaCost: ptr("{W}{W}"), Colors: []string{"W"}}},
				{card: &cards.Card{ManaCost: ptr("{1}{U}"), Colors: []string{"U"}}},
				{card: &cards.Card{ManaCost: ptr("{W}{U}"), Colors: []string{"W", "U"}}},
			},
			combo:         ColorCombination{Colors: []string{"W", "U"}, Name: "Azorius"},
			expectedTotal: 17,
		},
		{
			name:          "Mono-color all one type",
			selectedCards: []*scoredCard{},
			combo:         ColorCombination{Colors: []string{"R"}, Name: "Mono-Red"},
			expectedTotal: 17,
		},
		{
			name: "Two-color even split (no mana costs)",
			selectedCards: []*scoredCard{
				{card: &cards.Card{Colors: []string{"G"}}},
				{card: &cards.Card{Colors: []string{"W"}}},
			},
			combo:         ColorCombination{Colors: []string{"G", "W"}, Name: "Selesnya"},
			expectedTotal: 17,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lands := suggester.distributeLands(tt.selectedCards, tt.combo)

			// Count total lands
			total := 0
			for _, land := range lands {
				total += land.Quantity
			}

			if total != tt.expectedTotal {
				t.Errorf("expected %d total lands, got %d", tt.expectedTotal, total)
			}

			// Verify all lands are in the color combination
			for _, land := range lands {
				found := false
				for _, color := range tt.combo.Colors {
					if land.Color == color {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("land color %s not in combo %v", land.Color, tt.combo.Colors)
				}
			}
		})
	}
}

func TestCalculateDeckScore(t *testing.T) {
	suggester := &DeckSuggester{}

	tests := []struct {
		name          string
		selectedCards []*scoredCard
		analysis      *DeckSuggestionAnalysis
		minScore      float64
		maxScore      float64
	}{
		{
			name: "High quality deck",
			selectedCards: func() []*scoredCard {
				cards := make([]*scoredCard, 23)
				for i := range cards {
					cards[i] = &scoredCard{score: 0.8}
				}
				return cards
			}(),
			analysis: &DeckSuggestionAnalysis{
				CreatureCount: 15,
				ManaCurve:     map[int]int{2: 5, 3: 5, 4: 4},
			},
			minScore: 0.7,
			maxScore: 1.0,
		},
		{
			name: "Low quality deck",
			selectedCards: func() []*scoredCard {
				cards := make([]*scoredCard, 23)
				for i := range cards {
					cards[i] = &scoredCard{score: 0.4}
				}
				return cards
			}(),
			analysis: &DeckSuggestionAnalysis{
				CreatureCount: 8,
				ManaCurve:     map[int]int{5: 10, 6: 8},
			},
			minScore: 0.0,
			maxScore: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := suggester.calculateDeckScore(tt.selectedCards, tt.analysis)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("expected score between %.2f and %.2f, got %.2f", tt.minScore, tt.maxScore, score)
			}
		})
	}
}

func TestDetermineViability(t *testing.T) {
	suggester := &DeckSuggester{}

	tests := []struct {
		name     string
		score    float64
		analysis *DeckSuggestionAnalysis
		expected string
	}{
		{
			name:  "Strong deck",
			score: 0.75,
			analysis: &DeckSuggestionAnalysis{
				CreatureCount: 15,
				PlayableCount: 30,
			},
			expected: "strong",
		},
		{
			name:  "Viable deck",
			score: 0.55,
			analysis: &DeckSuggestionAnalysis{
				CreatureCount: 12,
				PlayableCount: 20,
			},
			expected: "viable",
		},
		{
			name:  "Weak deck",
			score: 0.4,
			analysis: &DeckSuggestionAnalysis{
				CreatureCount: 8,
				PlayableCount: 15,
			},
			expected: "weak",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := suggester.determineViability(tt.score, tt.analysis)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestColorCombinations(t *testing.T) {
	// Verify we have all 15 color combinations
	if len(allColorCombinations) != 15 {
		t.Errorf("expected 15 color combinations, got %d", len(allColorCombinations))
	}

	// Check mono-color count
	monoCount := 0
	twoColorCount := 0
	for _, combo := range allColorCombinations {
		if len(combo.Colors) == 1 {
			monoCount++
		} else if len(combo.Colors) == 2 {
			twoColorCount++
		}
	}

	if monoCount != 5 {
		t.Errorf("expected 5 mono-color combinations, got %d", monoCount)
	}
	if twoColorCount != 10 {
		t.Errorf("expected 10 two-color combinations, got %d", twoColorCount)
	}
}

func TestBasicLandsByColor(t *testing.T) {
	expectedLands := map[string]string{
		"W": "Plains",
		"U": "Island",
		"B": "Swamp",
		"R": "Mountain",
		"G": "Forest",
	}

	for color, expectedName := range expectedLands {
		land, ok := basicLandsByColor[color]
		if !ok {
			t.Errorf("missing land for color %s", color)
			continue
		}
		if land.Name != expectedName {
			t.Errorf("expected %s for color %s, got %s", expectedName, color, land.Name)
		}
		if land.ArenaID == 0 {
			t.Errorf("missing ArenaID for %s", land.Name)
		}
	}
}

// Helper function to create string pointer
func ptr(s string) *string {
	return &s
}
