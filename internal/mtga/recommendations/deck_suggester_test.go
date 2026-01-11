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
		combo         ColorCombination
		minScore      float64
		maxScore      float64
	}{
		{
			name: "High quality two-color deck",
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
			combo:    ColorCombination{Colors: []string{"W", "U"}, Name: "Azorius"},
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
			combo:    ColorCombination{Colors: []string{"W", "U"}, Name: "Azorius"},
			minScore: 0.0,
			maxScore: 0.5,
		},
		{
			name: "Three-color deck has lower score due to mana",
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
			combo:    ColorCombination{Colors: []string{"W", "U", "B"}, Name: "Esper"},
			minScore: 0.6,
			maxScore: 0.9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := suggester.calculateDeckScore(tt.selectedCards, tt.analysis, tt.combo)

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
	// Verify we have all 25 color combinations (5 mono + 10 two-color + 10 three-color)
	if len(allColorCombinations) != 25 {
		t.Errorf("expected 25 color combinations, got %d", len(allColorCombinations))
	}

	// Check mono-color count
	monoCount := 0
	twoColorCount := 0
	threeColorCount := 0
	for _, combo := range allColorCombinations {
		switch len(combo.Colors) {
		case 1:
			monoCount++
		case 2:
			twoColorCount++
		case 3:
			threeColorCount++
		}
	}

	if monoCount != 5 {
		t.Errorf("expected 5 mono-color combinations, got %d", monoCount)
	}
	if twoColorCount != 10 {
		t.Errorf("expected 10 two-color combinations, got %d", twoColorCount)
	}
	if threeColorCount != 10 {
		t.Errorf("expected 10 three-color combinations, got %d", threeColorCount)
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

// Tests for archetype-based deck building (#180)

func TestArchetypeTargets(t *testing.T) {
	// Verify all three archetypes are defined
	expected := []string{"aggro", "midrange", "control"}
	for _, key := range expected {
		target, ok := archetypeTargets[key]
		if !ok {
			t.Errorf("missing archetype target: %s", key)
			continue
		}
		if target.Name == "" {
			t.Errorf("archetype %s has empty name", key)
		}
		if target.CreatureMin <= 0 || target.CreatureMax <= 0 {
			t.Errorf("archetype %s has invalid creature counts", key)
		}
		if target.MaxAvgCMC <= 0 {
			t.Errorf("archetype %s has invalid max avg CMC", key)
		}
		if target.LandCount <= 0 {
			t.Errorf("archetype %s has invalid land count", key)
		}
		if len(target.PreferredCurve) == 0 {
			t.Errorf("archetype %s has empty preferred curve", key)
		}
	}
}

func TestScoreCMCForArchetype(t *testing.T) {
	suggester := &DeckSuggester{}

	tests := []struct {
		name      string
		card      *cards.Card
		archetype string
		minScore  float64
		maxScore  float64
	}{
		{
			name:      "1-drop for aggro (good)",
			card:      &cards.Card{CMC: 1},
			archetype: "aggro",
			minScore:  0.7,
			maxScore:  1.0,
		},
		{
			name:      "2-drop for aggro (best)",
			card:      &cards.Card{CMC: 2},
			archetype: "aggro",
			minScore:  0.9,
			maxScore:  1.0,
		},
		{
			name:      "5-drop for aggro (poor)",
			card:      &cards.Card{CMC: 5},
			archetype: "aggro",
			minScore:  0.0,
			maxScore:  0.3,
		},
		{
			name:      "5-drop for control (acceptable)",
			card:      &cards.Card{CMC: 5},
			archetype: "control",
			minScore:  0.5,
			maxScore:  0.9,
		},
		{
			name:      "3-drop for midrange (good)",
			card:      &cards.Card{CMC: 3},
			archetype: "midrange",
			minScore:  0.8,
			maxScore:  1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := archetypeTargets[tt.archetype]
			score := suggester.scoreCMCForArchetype(tt.card, target)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("expected score between %.2f and %.2f, got %.2f", tt.minScore, tt.maxScore, score)
			}
		})
	}
}

func TestScoreTypeForDraftArchetype(t *testing.T) {
	suggester := &DeckSuggester{}
	oracleDestroy := "Destroy target creature."
	oracleDraw := "Draw two cards."

	tests := []struct {
		name      string
		card      *cards.Card
		archetype string
		minScore  float64
		maxScore  float64
	}{
		{
			name:      "creature for aggro (best)",
			card:      &cards.Card{TypeLine: "Creature — Human"},
			archetype: "aggro",
			minScore:  0.9,
			maxScore:  1.0,
		},
		{
			name:      "spell for aggro (worse)",
			card:      &cards.Card{TypeLine: "Instant"},
			archetype: "aggro",
			minScore:  0.5,
			maxScore:  0.8,
		},
		{
			name:      "removal spell for control (best)",
			card:      &cards.Card{TypeLine: "Instant", OracleText: &oracleDestroy},
			archetype: "control",
			minScore:  0.9,
			maxScore:  1.0,
		},
		{
			name:      "draw spell for control (best)",
			card:      &cards.Card{TypeLine: "Sorcery", OracleText: &oracleDraw},
			archetype: "control",
			minScore:  0.9,
			maxScore:  1.0,
		},
		{
			name:      "creature for midrange (balanced)",
			card:      &cards.Card{TypeLine: "Creature"},
			archetype: "midrange",
			minScore:  0.6,
			maxScore:  1.0, // Midrange also benefits from creatures
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := archetypeTargets[tt.archetype]
			score := suggester.scoreTypeForArchetype(tt.card, target)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("expected score between %.2f and %.2f, got %.2f", tt.minScore, tt.maxScore, score)
			}
		})
	}
}

func TestIsViableForArchetype(t *testing.T) {
	suggester := &DeckSuggester{}

	tests := []struct {
		name       string
		candidates []*cards.Card
		archetype  string
		expected   bool
	}{
		{
			name: "viable aggro pool",
			candidates: func() []*cards.Card {
				result := make([]*cards.Card, 25)
				for i := range result {
					if i < 15 {
						result[i] = &cards.Card{TypeLine: "Creature"}
					} else {
						result[i] = &cards.Card{TypeLine: "Instant"}
					}
				}
				return result
			}(),
			archetype: "aggro",
			expected:  true,
		},
		{
			name: "not viable aggro - too few creatures",
			candidates: func() []*cards.Card {
				result := make([]*cards.Card, 20)
				for i := range result {
					if i < 8 {
						result[i] = &cards.Card{TypeLine: "Creature"}
					} else {
						result[i] = &cards.Card{TypeLine: "Instant"}
					}
				}
				return result
			}(),
			archetype: "aggro",
			expected:  false,
		},
		{
			name: "viable control pool",
			candidates: func() []*cards.Card {
				result := make([]*cards.Card, 20)
				for i := range result {
					if i < 10 {
						result[i] = &cards.Card{TypeLine: "Creature"}
					} else {
						result[i] = &cards.Card{TypeLine: "Instant"}
					}
				}
				return result
			}(),
			archetype: "control",
			expected:  true, // Control needs fewer creatures
		},
		{
			name:       "not viable - too few total cards",
			candidates: make([]*cards.Card, 10),
			archetype:  "midrange",
			expected:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := archetypeTargets[tt.archetype]
			result := suggester.isViableForArchetype(tt.candidates, target)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestSelectCardsForArchetype(t *testing.T) {
	suggester := &DeckSuggester{}

	// Create a pool of cards at various CMCs
	scoredCards := []*scoredCard{
		{card: &cards.Card{ArenaID: 1, CMC: 1}, score: 0.9},
		{card: &cards.Card{ArenaID: 2, CMC: 1}, score: 0.85},
		{card: &cards.Card{ArenaID: 3, CMC: 2}, score: 0.8},
		{card: &cards.Card{ArenaID: 4, CMC: 2}, score: 0.75},
		{card: &cards.Card{ArenaID: 5, CMC: 2}, score: 0.7},
		{card: &cards.Card{ArenaID: 6, CMC: 3}, score: 0.65},
		{card: &cards.Card{ArenaID: 7, CMC: 3}, score: 0.6},
		{card: &cards.Card{ArenaID: 8, CMC: 4}, score: 0.55},
		{card: &cards.Card{ArenaID: 9, CMC: 5}, score: 0.5},
	}

	tests := []struct {
		name        string
		archetype   string
		targetCount int
	}{
		{
			name:        "aggro selection",
			archetype:   "aggro",
			targetCount: 5,
		},
		{
			name:        "control selection",
			archetype:   "control",
			targetCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := archetypeTargets[tt.archetype]
			result := suggester.selectCardsForArchetype(scoredCards, tt.targetCount, target)

			if len(result) > tt.targetCount {
				t.Errorf("selected too many cards: got %d, max %d", len(result), tt.targetCount)
			}

			// Verify no duplicates
			seen := make(map[int]bool)
			for _, sc := range result {
				if seen[sc.card.ArenaID] {
					t.Errorf("duplicate card selected: ArenaID %d", sc.card.ArenaID)
				}
				seen[sc.card.ArenaID] = true
			}
		})
	}
}

func TestDistributeArchetypeLands(t *testing.T) {
	suggester := &DeckSuggester{}

	tests := []struct {
		name          string
		selectedCards []*scoredCard
		combo         ColorCombination
		landCount     int
	}{
		{
			name: "aggro land distribution (16 lands)",
			selectedCards: []*scoredCard{
				{card: &cards.Card{ManaCost: ptr("{R}{R}")}},
				{card: &cards.Card{ManaCost: ptr("{1}{G}")}},
			},
			combo:     ColorCombination{Colors: []string{"R", "G"}, Name: "Gruul"},
			landCount: 16,
		},
		{
			name: "control land distribution (18 lands)",
			selectedCards: []*scoredCard{
				{card: &cards.Card{ManaCost: ptr("{U}{U}")}},
				{card: &cards.Card{ManaCost: ptr("{1}{B}")}},
				{card: &cards.Card{ManaCost: ptr("{2}{U}{B}")}},
			},
			combo:     ColorCombination{Colors: []string{"U", "B"}, Name: "Dimir"},
			landCount: 18,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lands := suggester.distributeArchetypeLands(tt.selectedCards, tt.combo, tt.landCount)

			total := 0
			for _, land := range lands {
				total += land.Quantity
			}

			if total != tt.landCount {
				t.Errorf("expected %d total lands, got %d", tt.landCount, total)
			}
		})
	}
}

func TestCalculateArchetypeScore(t *testing.T) {
	suggester := &DeckSuggester{}

	tests := []struct {
		name          string
		selectedCards []*scoredCard
		analysis      *DeckSuggestionAnalysis
		combo         ColorCombination
		archetype     string
		minScore      float64
		maxScore      float64
	}{
		{
			name: "good aggro deck",
			selectedCards: func() []*scoredCard {
				cards := make([]*scoredCard, 24)
				for i := range cards {
					cards[i] = &scoredCard{score: 0.8}
				}
				return cards
			}(),
			analysis: &DeckSuggestionAnalysis{
				CreatureCount: 17,
				AverageCMC:    2.3,
			},
			combo:     ColorCombination{Colors: []string{"R", "G"}},
			archetype: "aggro",
			minScore:  0.7,
			maxScore:  1.0,
		},
		{
			name: "poor aggro deck (too few creatures)",
			selectedCards: func() []*scoredCard {
				cards := make([]*scoredCard, 24)
				for i := range cards {
					cards[i] = &scoredCard{score: 0.7}
				}
				return cards
			}(),
			analysis: &DeckSuggestionAnalysis{
				CreatureCount: 10,
				AverageCMC:    2.5,
			},
			combo:     ColorCombination{Colors: []string{"R", "G"}},
			archetype: "aggro",
			minScore:  0.4,
			maxScore:  0.8, // Slightly higher due to good CMC and card quality
		},
		{
			name: "good control deck",
			selectedCards: func() []*scoredCard {
				cards := make([]*scoredCard, 22)
				for i := range cards {
					cards[i] = &scoredCard{score: 0.8}
				}
				return cards
			}(),
			analysis: &DeckSuggestionAnalysis{
				CreatureCount: 11,
				AverageCMC:    3.2,
			},
			combo:     ColorCombination{Colors: []string{"U", "B"}},
			archetype: "control",
			minScore:  0.7,
			maxScore:  1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			target := archetypeTargets[tt.archetype]
			score := suggester.calculateArchetypeScore(tt.selectedCards, tt.analysis, tt.combo, target)

			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("expected score between %.2f and %.2f, got %.2f", tt.minScore, tt.maxScore, score)
			}
		})
	}
}

func TestExportToArenaFormat(t *testing.T) {
	tests := []struct {
		name     string
		deck     *SuggestedDeck
		contains []string
		isEmpty  bool
	}{
		{
			name:    "nil deck",
			deck:    nil,
			isEmpty: true,
		},
		{
			name: "basic deck export",
			deck: &SuggestedDeck{
				ColorCombo: ColorCombination{Name: "Gruul"},
				Spells: []*SuggestedCard{
					{Name: "Goblin Guide"},
					{Name: "Goblin Guide"},
					{Name: "Llanowar Elves"},
				},
				Lands: []*SuggestedLand{
					{Name: "Mountain", Quantity: 8},
					{Name: "Forest", Quantity: 8},
				},
			},
			contains: []string{
				"Deck: Gruul Draft",
				"2 Goblin Guide",
				"1 Llanowar Elves",
				"8 Mountain",
				"8 Forest",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExportToArenaFormat(tt.deck)

			if tt.isEmpty {
				if result != "" {
					t.Errorf("expected empty string, got: %s", result)
				}
				return
			}

			for _, substr := range tt.contains {
				if !containsString(result, substr) {
					t.Errorf("expected result to contain %q, got:\n%s", substr, result)
				}
			}
		})
	}
}

func TestGetAvailableDraftArchetypes(t *testing.T) {
	archetypes := GetAvailableDraftArchetypes()

	if len(archetypes) != 3 {
		t.Errorf("expected 3 archetypes, got %d", len(archetypes))
	}

	expected := map[string]bool{"aggro": true, "midrange": true, "control": true}
	for _, arch := range archetypes {
		if !expected[arch] {
			t.Errorf("unexpected archetype: %s", arch)
		}
	}
}

func TestGetDraftArchetypeDescription(t *testing.T) {
	tests := []struct {
		key      string
		contains string
	}{
		{"aggro", "Fast"},
		{"midrange", "Balanced"},
		{"control", "Slower"},
		{"invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			desc := GetDraftArchetypeDescription(tt.key)
			if tt.contains == "" {
				if desc != "" {
					t.Errorf("expected empty description for invalid key, got: %s", desc)
				}
			} else {
				if !containsString(desc, tt.contains) {
					t.Errorf("expected description to contain %q, got: %s", tt.contains, desc)
				}
			}
		})
	}
}

// containsString is a helper to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
