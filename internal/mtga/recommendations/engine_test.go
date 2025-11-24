package recommendations

import (
	"context"
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

func TestNewRuleBasedEngine(t *testing.T) {
	engine := NewRuleBasedEngine(nil, nil)
	if engine == nil {
		t.Fatal("NewRuleBasedEngine returned nil")
	}
	if engine.cardService != nil {
		t.Error("Expected cardService to be nil")
	}
}

func TestScoreColorFit(t *testing.T) {
	tests := []struct {
		name          string
		cardColors    []string
		deckColors    map[string]int
		primaryColors []string
		expectedMin   float64
		expectedMax   float64
	}{
		{
			name:          "colorless card",
			cardColors:    []string{},
			deckColors:    map[string]int{"Red": 10, "Blue": 5},
			primaryColors: []string{"Red", "Blue"},
			expectedMin:   1.0,
			expectedMax:   1.0,
		},
		{
			name:          "perfect match with primary colors",
			cardColors:    []string{"Red"},
			deckColors:    map[string]int{"Red": 10, "Blue": 5},
			primaryColors: []string{"Red", "Blue"},
			expectedMin:   1.0,
			expectedMax:   1.0,
		},
		{
			name:          "match but not primary",
			cardColors:    []string{"Blue"},
			deckColors:    map[string]int{"Red": 10, "Blue": 2},
			primaryColors: []string{"Red"},
			expectedMin:   0.8,
			expectedMax:   0.9,
		},
		{
			name:          "no color match",
			cardColors:    []string{"Green"},
			deckColors:    map[string]int{"Red": 10, "Blue": 5},
			primaryColors: []string{"Red", "Blue"},
			expectedMin:   0.0,
			expectedMax:   0.0,
		},
		{
			name:          "partial match",
			cardColors:    []string{"Red", "Green"},
			deckColors:    map[string]int{"Red": 10, "Blue": 5},
			primaryColors: []string{"Red", "Blue"},
			expectedMin:   0.2,
			expectedMax:   0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &cards.Card{
				Colors: tt.cardColors,
			}
			analysis := &DeckAnalysis{
				Colors:        tt.deckColors,
				ColorIdentity: getColorIdentity(tt.deckColors),
				PrimaryColors: tt.primaryColors,
			}

			score := scoreColorFit(card, analysis)

			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("scoreColorFit() = %v, want between %v and %v",
					score, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestScoreManaCurve(t *testing.T) {
	tests := []struct {
		name        string
		cardCMC     float64
		typeLine    string
		manaCurve   map[int]int
		expectedMin float64
		expectedMax float64
	}{
		{
			name:        "fills gap at CMC 2",
			cardCMC:     2.0,
			typeLine:    "Creature — Human",
			manaCurve:   map[int]int{1: 2, 3: 5, 4: 4},
			expectedMin: 0.8,
			expectedMax: 1.0,
		},
		{
			name:        "at ideal count",
			cardCMC:     3.0,
			typeLine:    "Creature — Elf",
			manaCurve:   map[int]int{1: 2, 2: 5, 3: 5, 4: 4},
			expectedMin: 0.6,
			expectedMax: 0.6,
		},
		{
			name:        "over ideal count",
			cardCMC:     2.0,
			typeLine:    "Instant",
			manaCurve:   map[int]int{1: 2, 2: 8, 3: 3},
			expectedMin: 0.1,
			expectedMax: 0.4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &cards.Card{
				CMC:      tt.cardCMC,
				TypeLine: tt.typeLine,
			}
			analysis := &DeckAnalysis{
				ManaCurve: tt.manaCurve,
			}

			score := scoreManaCurve(card, analysis)

			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("scoreManaCurve() = %v, want between %v and %v",
					score, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestScoreCardQuality(t *testing.T) {
	tests := []struct {
		name        string
		rarity      string
		expectedMin float64
		expectedMax float64
	}{
		{
			name:        "mythic rarity",
			rarity:      "mythic",
			expectedMin: 0.85,
			expectedMax: 0.85,
		},
		{
			name:        "rare rarity",
			rarity:      "rare",
			expectedMin: 0.75,
			expectedMax: 0.75,
		},
		{
			name:        "uncommon rarity",
			rarity:      "uncommon",
			expectedMin: 0.60,
			expectedMax: 0.60,
		},
		{
			name:        "common rarity",
			rarity:      "common",
			expectedMin: 0.50,
			expectedMax: 0.50,
		},
		{
			name:        "unknown rarity",
			rarity:      "unknown",
			expectedMin: 0.50,
			expectedMax: 0.50,
		},
	}

	engine := NewRuleBasedEngine(nil, nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &cards.Card{
				Rarity: tt.rarity,
			}

			score := engine.fallbackQualityScore(card)

			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("fallbackQualityScore() = %v, want between %v and %v",
					score, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestScoreSynergy(t *testing.T) {
	flyingText := "Flying, Vigilance"
	humanTypeLine := "Creature — Human Warrior"

	tests := []struct {
		name        string
		oracleText  *string
		typeLine    string
		keywords    map[string]int
		creatures   map[string]int
		expectedMin float64
		expectedMax float64
	}{
		{
			name:        "keyword synergy",
			oracleText:  &flyingText,
			typeLine:    "Creature — Bird",
			keywords:    map[string]int{"Flying": 5, "Vigilance": 3},
			creatures:   map[string]int{},
			expectedMin: 0.2,
			expectedMax: 0.3,
		},
		{
			name:        "tribal synergy",
			oracleText:  nil,
			typeLine:    humanTypeLine,
			keywords:    map[string]int{},
			creatures:   map[string]int{"Human": 5, "Warrior": 4},
			expectedMin: 0.45, // Improved tribal scoring: 5 Humans=0.6, 4 Warriors=0.4, averaged=0.5
			expectedMax: 0.55,
		},
		{
			name:        "strong tribal synergy",
			oracleText:  nil,
			typeLine:    "Creature — Ally Warrior",
			keywords:    map[string]int{},
			creatures:   map[string]int{"Ally": 10}, // 10+ Allies = strong tribal deck
			expectedMin: 0.75, // 10 Allies gives 0.8 bonus
			expectedMax: 0.85,
		},
		{
			name:        "no synergy",
			oracleText:  nil,
			typeLine:    "Instant",
			keywords:    map[string]int{},
			creatures:   map[string]int{},
			expectedMin: 0.5,
			expectedMax: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &cards.Card{
				OracleText: tt.oracleText,
				TypeLine:   tt.typeLine,
			}
			deck := &DeckContext{
				Format: "Limited",
			}
			analysis := &DeckAnalysis{
				Keywords:      tt.keywords,
				CreatureTypes: tt.creatures,
			}

			score := scoreSynergy(card, deck, analysis)

			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("scoreSynergy() = %v, want between %v and %v",
					score, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestScorePlayability(t *testing.T) {
	tests := []struct {
		name         string
		format       string
		draftCardIDs []int
		cardID       int
		expectedMin  float64
		expectedMax  float64
	}{
		{
			name:         "Limited format in draft pool",
			format:       "Limited",
			draftCardIDs: []int{123, 456, 789},
			cardID:       456,
			expectedMin:  0.9,
			expectedMax:  0.9,
		},
		{
			name:         "Limited format not in draft pool",
			format:       "Limited",
			draftCardIDs: []int{123, 456},
			cardID:       789,
			expectedMin:  0.1,
			expectedMax:  0.1,
		},
		{
			name:         "Constructed format",
			format:       "Standard",
			draftCardIDs: nil,
			cardID:       123,
			expectedMin:  0.8,
			expectedMax:  0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &cards.Card{
				ArenaID: tt.cardID,
			}
			deck := &DeckContext{
				Format:       tt.format,
				DraftCardIDs: tt.draftCardIDs,
			}

			score := scorePlayability(card, deck)

			if score < tt.expectedMin || score > tt.expectedMax {
				t.Errorf("scorePlayability() = %v, want between %v and %v",
					score, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestGenerateExplanation(t *testing.T) {
	card := &cards.Card{
		CMC: 3.0,
	}

	tests := []struct {
		name     string
		factors  *ScoreFactors
		contains string
	}{
		{
			name: "high color fit",
			factors: &ScoreFactors{
				ColorFit:  0.9,
				ManaCurve: 0.5,
				Quality:   0.6,
				Synergy:   0.5,
				Playable:  0.8,
			},
			contains: "matches your deck's colors perfectly",
		},
		{
			name: "fills mana curve gap",
			factors: &ScoreFactors{
				ColorFit:  0.7,
				ManaCurve: 0.8,
				Quality:   0.6,
				Synergy:   0.5,
				Playable:  0.8,
			},
			contains: "fills a gap in your mana curve",
		},
		{
			name: "high quality",
			factors: &ScoreFactors{
				ColorFit:  0.7,
				ManaCurve: 0.5,
				Quality:   0.85,
				Synergy:   0.5,
				Playable:  0.8,
			},
			contains: "high-quality card",
		},
		{
			name: "strong synergy",
			factors: &ScoreFactors{
				ColorFit:  0.7,
				ManaCurve: 0.5,
				Quality:   0.6,
				Synergy:   0.75,
				Playable:  0.8,
			},
			contains: "strong synergy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := &DeckAnalysis{}
			explanation := generateExplanation(card, tt.factors, analysis)

			if explanation == "" {
				t.Error("generateExplanation() returned empty string")
			}

			if tt.contains != "" && !contains(explanation, tt.contains) {
				t.Errorf("generateExplanation() = %q, want to contain %q",
					explanation, tt.contains)
			}
		})
	}
}

func TestDeterminePrimarySource(t *testing.T) {
	tests := []struct {
		name     string
		factors  *ScoreFactors
		expected string
	}{
		{
			name: "color fit primary",
			factors: &ScoreFactors{
				ColorFit:  0.9,
				ManaCurve: 0.6,
				Quality:   0.5,
				Synergy:   0.5,
				Playable:  0.7,
			},
			expected: "color-fit",
		},
		{
			name: "mana curve primary",
			factors: &ScoreFactors{
				ColorFit:  0.6,
				ManaCurve: 0.9,
				Quality:   0.5,
				Synergy:   0.5,
				Playable:  0.7,
			},
			expected: "mana-curve",
		},
		{
			name: "quality primary",
			factors: &ScoreFactors{
				ColorFit:  0.6,
				ManaCurve: 0.5,
				Quality:   0.95,
				Synergy:   0.5,
				Playable:  0.7,
			},
			expected: "quality",
		},
		{
			name: "synergy primary",
			factors: &ScoreFactors{
				ColorFit:  0.6,
				ManaCurve: 0.5,
				Quality:   0.5,
				Synergy:   0.9,
				Playable:  0.7,
			},
			expected: "synergy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source := determinePrimarySource(tt.factors)
			if source != tt.expected {
				t.Errorf("determinePrimarySource() = %v, want %v", source, tt.expected)
			}
		})
	}
}

func TestCalculateConfidence(t *testing.T) {
	tests := []struct {
		name        string
		factors     *ScoreFactors
		expectedMin float64
		expectedMax float64
	}{
		{
			name: "all factors positive",
			factors: &ScoreFactors{
				ColorFit:  0.8,
				ManaCurve: 0.7,
				Quality:   0.9,
				Synergy:   0.65,
				Playable:  0.8,
			},
			expectedMin: 0.9,
			expectedMax: 1.0,
		},
		{
			name: "mixed factors",
			factors: &ScoreFactors{
				ColorFit:  0.7,
				ManaCurve: 0.4,
				Quality:   0.8,
				Synergy:   0.5,
				Playable:  0.7,
			},
			expectedMin: 0.5,
			expectedMax: 0.7,
		},
		{
			name: "low confidence",
			factors: &ScoreFactors{
				ColorFit:  0.3,
				ManaCurve: 0.4,
				Quality:   0.5,
				Synergy:   0.4,
				Playable:  0.3,
			},
			expectedMin: 0.0,
			expectedMax: 0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := calculateConfidence(tt.factors)

			if confidence < tt.expectedMin || confidence > tt.expectedMax {
				t.Errorf("calculateConfidence() = %v, want between %v and %v",
					confidence, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestIsCardInDeck(t *testing.T) {
	cards := []*models.DeckCard{
		{CardID: 123, Quantity: 4},
		{CardID: 456, Quantity: 2},
		{CardID: 789, Quantity: 1},
	}

	tests := []struct {
		name     string
		cardID   int
		expected bool
	}{
		{
			name:     "card in deck",
			cardID:   456,
			expected: true,
		},
		{
			name:     "card not in deck",
			cardID:   999,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCardInDeck(tt.cardID, cards)
			if result != tt.expected {
				t.Errorf("isCardInDeck() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMatchesFilters(t *testing.T) {
	tests := []struct {
		name     string
		card     *cards.Card
		filters  *Filters
		expected bool
	}{
		{
			name: "matches color filter",
			card: &cards.Card{
				Colors: []string{"Red"},
			},
			filters: &Filters{
				Colors: []string{"Red", "Blue"},
			},
			expected: true,
		},
		{
			name: "doesn't match color filter",
			card: &cards.Card{
				Colors: []string{"Green"},
			},
			filters: &Filters{
				Colors: []string{"Red", "Blue"},
			},
			expected: false,
		},
		{
			name: "matches type filter",
			card: &cards.Card{
				TypeLine: "Creature — Human",
			},
			filters: &Filters{
				CardTypes: []string{"Creature"},
			},
			expected: true,
		},
		{
			name: "doesn't match type filter",
			card: &cards.Card{
				TypeLine: "Instant",
			},
			filters: &Filters{
				CardTypes: []string{"Creature"},
			},
			expected: false,
		},
		{
			name: "matches CMC range",
			card: &cards.Card{
				CMC: 3.0,
			},
			filters: &Filters{
				CMCRange: &CMCRange{Min: 2, Max: 4},
			},
			expected: true,
		},
		{
			name: "doesn't match CMC range",
			card: &cards.Card{
				CMC: 6.0,
			},
			filters: &Filters{
				CMCRange: &CMCRange{Min: 2, Max: 4},
			},
			expected: false,
		},
		{
			name: "land excluded when IncludeLands false",
			card: &cards.Card{
				TypeLine: "Basic Land — Mountain",
			},
			filters: &Filters{
				IncludeLands: false,
			},
			expected: false,
		},
		{
			name: "land included when IncludeLands true",
			card: &cards.Card{
				TypeLine: "Basic Land — Forest",
			},
			filters: &Filters{
				IncludeLands: true,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := matchesFilters(tt.card, tt.filters)
			if result != tt.expected {
				t.Errorf("matchesFilters() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSortRecommendations(t *testing.T) {
	recommendations := []*CardRecommendation{
		{Score: 0.5},
		{Score: 0.9},
		{Score: 0.3},
		{Score: 0.7},
	}

	sortRecommendations(recommendations)

	expectedOrder := []float64{0.9, 0.7, 0.5, 0.3}
	for i, rec := range recommendations {
		if rec.Score != expectedOrder[i] {
			t.Errorf("sortRecommendations() position %d = %v, want %v",
				i, rec.Score, expectedOrder[i])
		}
	}
}

func TestGetColorIdentity(t *testing.T) {
	tests := []struct {
		name     string
		colors   map[string]int
		expected int // Number of colors in identity
	}{
		{
			name: "two colors",
			colors: map[string]int{
				"Red":  10,
				"Blue": 5,
			},
			expected: 2,
		},
		{
			name: "mono color",
			colors: map[string]int{
				"Green": 15,
			},
			expected: 1,
		},
		{
			name: "ignores zero counts",
			colors: map[string]int{
				"Red":   10,
				"Blue":  0,
				"Green": 5,
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			identity := getColorIdentity(tt.colors)
			if len(identity) != tt.expected {
				t.Errorf("getColorIdentity() returned %d colors, want %d",
					len(identity), tt.expected)
			}
		})
	}
}

func TestGetPrimaryColors(t *testing.T) {
	colors := map[string]int{
		"Red":   15,
		"Blue":  8,
		"Green": 3,
		"White": 1,
	}

	primary := getPrimaryColors(colors, 2)

	if len(primary) != 2 {
		t.Errorf("getPrimaryColors() returned %d colors, want 2", len(primary))
	}

	// Should return Red and Blue (highest counts)
	hasRed := false
	hasBlue := false
	for _, color := range primary {
		if color == "Red" {
			hasRed = true
		}
		if color == "Blue" {
			hasBlue = true
		}
	}

	if !hasRed || !hasBlue {
		t.Errorf("getPrimaryColors() = %v, want Red and Blue", primary)
	}
}

func TestExtractKeywordsFromText(t *testing.T) {
	tests := []struct {
		name         string
		text         string
		expectedKeys []string
	}{
		{
			name:         "single keyword",
			text:         "Creature with Flying",
			expectedKeys: []string{"Flying"},
		},
		{
			name:         "multiple keywords",
			text:         "First strike, Deathtouch, and Lifelink",
			expectedKeys: []string{"First strike", "Deathtouch", "Lifelink"},
		},
		{
			name:         "no keywords",
			text:         "Draw a card",
			expectedKeys: []string{},
		},
		{
			name:         "case insensitive",
			text:         "FLYING and TRAMPLE",
			expectedKeys: []string{"Flying", "Trample"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := extractKeywordsFromText(tt.text)

			if len(keywords) != len(tt.expectedKeys) {
				t.Errorf("extractKeywordsFromText() found %d keywords, want %d",
					len(keywords), len(tt.expectedKeys))
			}

			for _, expected := range tt.expectedKeys {
				if !keywords[expected] {
					t.Errorf("extractKeywordsFromText() missing keyword %q", expected)
				}
			}
		})
	}
}

func TestExtractCreatureTypesFromLine(t *testing.T) {
	tests := []struct {
		name          string
		typeLine      string
		expectedTypes []string
	}{
		{
			name:          "standard format",
			typeLine:      "Creature — Human Warrior",
			expectedTypes: []string{"Human", "Warrior"},
		},
		{
			name:          "legendary creature",
			typeLine:      "Legendary Creature — Elf Wizard",
			expectedTypes: []string{"Elf", "Wizard"},
		},
		{
			name:          "single dash",
			typeLine:      "Creature - Dragon",
			expectedTypes: []string{"Dragon"},
		},
		{
			name:          "no creature types",
			typeLine:      "Creature",
			expectedTypes: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			types := extractCreatureTypesFromLine(tt.typeLine)

			if len(types) != len(tt.expectedTypes) {
				t.Errorf("extractCreatureTypesFromLine() found %d types, want %d",
					len(types), len(tt.expectedTypes))
			}

			for _, expected := range tt.expectedTypes {
				if !types[expected] {
					t.Errorf("extractCreatureTypesFromLine() missing type %q", expected)
				}
			}
		})
	}
}

func TestContainsTypeInTypeLine(t *testing.T) {
	tests := []struct {
		name       string
		typeLine   string
		targetType string
		expected   bool
	}{
		{
			name:       "contains creature type",
			typeLine:   "Legendary Creature — Human Warrior",
			targetType: "Creature",
			expected:   true,
		},
		{
			name:       "contains legendary",
			typeLine:   "Legendary Creature — Elf",
			targetType: "Legendary",
			expected:   true,
		},
		{
			name:       "doesn't contain type",
			typeLine:   "Instant",
			targetType: "Creature",
			expected:   false,
		},
		{
			name:       "case insensitive",
			typeLine:   "Creature — Dragon",
			targetType: "creature",
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsTypeInTypeLine(tt.typeLine, tt.targetType)
			if result != tt.expected {
				t.Errorf("containsTypeInTypeLine() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAnalyzeDeck(t *testing.T) {
	text := "Flying, Vigilance"
	deck := &DeckContext{
		Cards: []*models.DeckCard{
			{CardID: 1, Quantity: 4, Board: "main"},
			{CardID: 2, Quantity: 3, Board: "main"},
			{CardID: 3, Quantity: 10, Board: "main"},
			{CardID: 4, Quantity: 2, Board: "sideboard"}, // Should be skipped
		},
		CardMetadata: map[int]*cards.Card{
			1: {
				ArenaID:    1,
				Colors:     []string{"Red"},
				TypeLine:   "Creature — Human Warrior",
				CMC:        2.0,
				OracleText: &text,
			},
			2: {
				ArenaID:  2,
				Colors:   []string{"Red"},
				TypeLine: "Instant",
				CMC:      1.0,
			},
			3: {
				ArenaID:  3,
				Colors:   []string{},
				TypeLine: "Land — Mountain",
			},
			4: {
				ArenaID:  4,
				Colors:   []string{"Blue"},
				TypeLine: "Creature — Merfolk",
				CMC:      3.0,
			},
		},
	}

	analysis := analyzeDeck(deck)

	// Check total cards (mainboard only)
	if analysis.TotalCards != 17 {
		t.Errorf("TotalCards = %d, want 17", analysis.TotalCards)
	}

	// Check non-land count
	if analysis.TotalNonLands != 7 {
		t.Errorf("TotalNonLands = %d, want 7", analysis.TotalNonLands)
	}

	// Check color distribution (should only count non-lands)
	if analysis.Colors["Red"] != 7 {
		t.Errorf("Red count = %d, want 7", analysis.Colors["Red"])
	}

	// Check mana curve
	if analysis.ManaCurve[2] != 4 {
		t.Errorf("CMC 2 count = %d, want 4", analysis.ManaCurve[2])
	}

	// Check keywords
	if analysis.Keywords["Flying"] == 0 {
		t.Error("Expected Flying keyword to be detected")
	}

	// Check creature types
	if analysis.CreatureTypes["Human"] != 4 {
		t.Errorf("Human count = %d, want 4", analysis.CreatureTypes["Human"])
	}
}

func TestGetRecommendations_EmptyDeck(t *testing.T) {
	engine := NewRuleBasedEngine(nil, nil)

	result, err := engine.GetRecommendations(context.Background(), nil, nil)
	if err == nil {
		t.Error("Expected error for nil deck context")
	}
	if result != nil {
		t.Error("Expected nil result for nil deck context")
	}
}

func TestRecordAcceptance(t *testing.T) {
	engine := NewRuleBasedEngine(nil, nil)

	// Phase 1A: This should be a no-op
	err := engine.RecordAcceptance(context.Background(), "deck-123", 456, true)
	if err != nil {
		t.Errorf("RecordAcceptance() error = %v, want nil", err)
	}
}

func TestConvertSetCardToCardsCard(t *testing.T) {
	tests := []struct {
		name     string
		setCard  *models.SetCard
		expected *cards.Card
	}{
		{
			name: "basic card conversion",
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
				Power:      "",
				Toughness:  "",
				Text:       "Deal 3 damage to any target.",
				ImageURL:   "https://example.com/bolt.jpg",
			},
			expected: &cards.Card{
				ArenaID:    12345,
				ScryfallID: "scryfall-123",
				Name:       "Lightning Bolt",
				TypeLine:   "Instant",
				SetCode:    "M21",
				CMC:        1.0,
				Colors:     []string{"R"},
				Rarity:     "common",
			},
		},
		{
			name: "creature with power/toughness",
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
				Text:       "",
				ImageURL:   "https://example.com/bears.jpg",
			},
			expected: &cards.Card{
				ArenaID:    67890,
				ScryfallID: "scryfall-456",
				Name:       "Grizzly Bears",
				TypeLine:   "Creature Bear",
				SetCode:    "M21",
				CMC:        2.0,
				Colors:     []string{"G"},
				Rarity:     "common",
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
			expected: &cards.Card{
				ArenaID:    11111,
				ScryfallID: "scryfall-789",
				Name:       "Azorius Charm",
				TypeLine:   "Instant",
				SetCode:    "RNA",
				CMC:        2.0,
				Colors:     []string{"W", "U"},
				Rarity:     "uncommon",
			},
		},
		{
			name:     "nil setCard",
			setCard:  nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertSetCardToCardsCard(tt.setCard)

			if tt.expected == nil {
				if result != nil {
					t.Errorf("convertSetCardToCardsCard() = %v, want nil", result)
				}
				return
			}

			if result == nil {
				t.Fatal("convertSetCardToCardsCard() returned nil, want non-nil")
			}

			// Check basic fields
			if result.ArenaID != tt.expected.ArenaID {
				t.Errorf("ArenaID = %d, want %d", result.ArenaID, tt.expected.ArenaID)
			}
			if result.Name != tt.expected.Name {
				t.Errorf("Name = %s, want %s", result.Name, tt.expected.Name)
			}
			if result.SetCode != tt.expected.SetCode {
				t.Errorf("SetCode = %s, want %s", result.SetCode, tt.expected.SetCode)
			}
			if result.CMC != tt.expected.CMC {
				t.Errorf("CMC = %f, want %f", result.CMC, tt.expected.CMC)
			}
			if result.Rarity != tt.expected.Rarity {
				t.Errorf("Rarity = %s, want %s", result.Rarity, tt.expected.Rarity)
			}

			// Check ManaCost pointer
			if tt.setCard.ManaCost != "" {
				if result.ManaCost == nil {
					t.Error("ManaCost is nil, expected non-nil")
				} else if *result.ManaCost != tt.setCard.ManaCost {
					t.Errorf("ManaCost = %s, want %s", *result.ManaCost, tt.setCard.ManaCost)
				}
			}

			// Check Power/Toughness pointers
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

			// Check OracleText pointer
			if tt.setCard.Text != "" {
				if result.OracleText == nil {
					t.Error("OracleText is nil, expected non-nil")
				} else if *result.OracleText != tt.setCard.Text {
					t.Errorf("OracleText = %s, want %s", *result.OracleText, tt.setCard.Text)
				}
			}

			// Check ImageURI pointer
			if tt.setCard.ImageURL != "" {
				if result.ImageURI == nil {
					t.Error("ImageURI is nil, expected non-nil")
				} else if *result.ImageURI != tt.setCard.ImageURL {
					t.Errorf("ImageURI = %s, want %s", *result.ImageURI, tt.setCard.ImageURL)
				}
			}

			// Check Colors slice
			if len(result.Colors) != len(tt.expected.Colors) {
				t.Errorf("Colors length = %d, want %d", len(result.Colors), len(tt.expected.Colors))
			}
		})
	}
}

func TestNewRuleBasedEngineWithSetRepo(t *testing.T) {
	engine := NewRuleBasedEngineWithSetRepo(nil, nil, nil)
	if engine == nil {
		t.Fatal("NewRuleBasedEngineWithSetRepo returned nil")
	}
	if engine.cardService != nil {
		t.Error("Expected cardService to be nil")
	}
	if engine.setCardRepo != nil {
		t.Error("Expected setCardRepo to be nil")
	}
	if engine.ratingsRepo != nil {
		t.Error("Expected ratingsRepo to be nil")
	}
}

// Helper functions

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > 0 && len(substr) > 0 && containsHelper(s, substr)))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
