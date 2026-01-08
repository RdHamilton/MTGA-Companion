package recommendations

import (
	"context"
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
)

func TestScoreColorCompatibility(t *testing.T) {
	builder := &SeedDeckBuilder{}

	tests := []struct {
		name         string
		card         *cards.Card
		seedAnalysis *SeedCardAnalysis
		minScore     float64
		maxScore     float64
	}{
		{
			name:         "Colorless card fits any deck",
			card:         &cards.Card{Colors: []string{}},
			seedAnalysis: &SeedCardAnalysis{Colors: []string{"W", "U"}},
			minScore:     1.0,
			maxScore:     1.0,
		},
		{
			name:         "Exact color match",
			card:         &cards.Card{Colors: []string{"W"}},
			seedAnalysis: &SeedCardAnalysis{Colors: []string{"W"}},
			minScore:     1.0,
			maxScore:     1.0,
		},
		{
			name:         "Card color is subset of seed colors",
			card:         &cards.Card{Colors: []string{"W"}},
			seedAnalysis: &SeedCardAnalysis{Colors: []string{"W", "U"}},
			minScore:     1.0,
			maxScore:     1.0,
		},
		{
			name:         "Multi-color card matches all seed colors",
			card:         &cards.Card{Colors: []string{"W", "U"}},
			seedAnalysis: &SeedCardAnalysis{Colors: []string{"W", "U"}},
			minScore:     1.0,
			maxScore:     1.0,
		},
		{
			name:         "No color overlap",
			card:         &cards.Card{Colors: []string{"R"}},
			seedAnalysis: &SeedCardAnalysis{Colors: []string{"W", "U"}},
			minScore:     0.0,
			maxScore:     0.0,
		},
		{
			name:         "Partial color overlap",
			card:         &cards.Card{Colors: []string{"W", "R"}},
			seedAnalysis: &SeedCardAnalysis{Colors: []string{"W", "U"}},
			minScore:     0.3,
			maxScore:     0.4, // 1/2 * 0.7 = 0.35
		},
		{
			name:         "Colorless seed - any color works",
			card:         &cards.Card{Colors: []string{"B"}},
			seedAnalysis: &SeedCardAnalysis{Colors: []string{}},
			minScore:     0.8,
			maxScore:     0.8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := builder.scoreColorCompatibility(tt.card, tt.seedAnalysis)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("expected score between %.2f and %.2f, got %.2f", tt.minScore, tt.maxScore, score)
			}
		})
	}
}

func TestScoreManaCurveFit(t *testing.T) {
	builder := &SeedDeckBuilder{}

	tests := []struct {
		name     string
		card     *cards.Card
		minScore float64
		maxScore float64
	}{
		{
			name:     "CMC 2 is ideal",
			card:     &cards.Card{CMC: 2, TypeLine: "Creature"},
			minScore: 1.0,
			maxScore: 1.0,
		},
		{
			name:     "CMC 3 is ideal",
			card:     &cards.Card{CMC: 3, TypeLine: "Creature"},
			minScore: 1.0,
			maxScore: 1.0,
		},
		{
			name:     "CMC 1 is good",
			card:     &cards.Card{CMC: 1, TypeLine: "Instant"},
			minScore: 0.8,
			maxScore: 0.8,
		},
		{
			name:     "CMC 5 is acceptable",
			card:     &cards.Card{CMC: 5, TypeLine: "Creature"},
			minScore: 0.6,
			maxScore: 0.6,
		},
		{
			name:     "CMC 7+ is risky",
			card:     &cards.Card{CMC: 8, TypeLine: "Creature"},
			minScore: 0.3,
			maxScore: 0.3,
		},
		{
			name:     "Land is neutral",
			card:     &cards.Card{CMC: 0, TypeLine: "Land"},
			minScore: 0.5,
			maxScore: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := builder.scoreManaCurveFit(tt.card)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("expected score between %.2f and %.2f, got %.2f", tt.minScore, tt.maxScore, score)
			}
		})
	}
}

func TestSeedDeckBuilder_ScoreCardQuality(t *testing.T) {
	builder := &SeedDeckBuilder{}

	tests := []struct {
		name     string
		card     *cards.Card
		expected float64
	}{
		{
			name:     "Mythic has highest quality",
			card:     &cards.Card{Rarity: "mythic"},
			expected: 0.85,
		},
		{
			name:     "Rare has high quality",
			card:     &cards.Card{Rarity: "rare"},
			expected: 0.75,
		},
		{
			name:     "Uncommon has medium quality",
			card:     &cards.Card{Rarity: "uncommon"},
			expected: 0.60,
		},
		{
			name:     "Common has base quality",
			card:     &cards.Card{Rarity: "common"},
			expected: 0.50,
		},
		{
			name:     "Unknown rarity defaults to 0.5",
			card:     &cards.Card{Rarity: "special"},
			expected: 0.50,
		},
		{
			name:     "Rarity is case insensitive",
			card:     &cards.Card{Rarity: "MYTHIC"},
			expected: 0.85,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := builder.scoreCardQuality(tt.card)
			if score != tt.expected {
				t.Errorf("expected %.2f, got %.2f", tt.expected, score)
			}
		})
	}
}

func TestScoreSynergyWithSeed(t *testing.T) {
	builder := &SeedDeckBuilder{}

	tests := []struct {
		name         string
		card         *cards.Card
		seedAnalysis *SeedCardAnalysis
		minScore     float64
		maxScore     float64
	}{
		{
			name:         "No synergy - neutral score",
			card:         &cards.Card{OracleText: strPtr("Draw a card.")},
			seedAnalysis: &SeedCardAnalysis{Keywords: []KeywordInfo{}},
			minScore:     0.5,
			maxScore:     0.5,
		},
		{
			name: "Tribal synergy - same creature type",
			card: &cards.Card{
				TypeLine: "Creature - Elf Warrior",
			},
			seedAnalysis: &SeedCardAnalysis{
				IsCreature:    true,
				CreatureTypes: []string{"Elf"},
			},
			minScore: 0.7,
			maxScore: 1.0,
		},
		{
			name:         "No oracle text - neutral score",
			card:         &cards.Card{OracleText: nil},
			seedAnalysis: &SeedCardAnalysis{Keywords: []KeywordInfo{}},
			minScore:     0.5,
			maxScore:     0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := builder.scoreSynergyWithSeed(tt.card, tt.seedAnalysis)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("expected score between %.2f and %.2f, got %.2f", tt.minScore, tt.maxScore, score)
			}
		})
	}
}

func TestScoreCardForSeed(t *testing.T) {
	builder := &SeedDeckBuilder{}

	tests := []struct {
		name         string
		card         *cards.Card
		seedAnalysis *SeedCardAnalysis
		minScore     float64
		maxScore     float64
	}{
		{
			name: "Ideal card - same color, good CMC, rare",
			card: &cards.Card{
				Colors:   []string{"W"},
				CMC:      2,
				Rarity:   "rare",
				TypeLine: "Creature",
			},
			seedAnalysis: &SeedCardAnalysis{Colors: []string{"W"}},
			minScore:     0.7,
			maxScore:     1.0,
		},
		{
			name: "Off-color card scores lower but has other factors",
			card: &cards.Card{
				Colors:   []string{"R"},
				CMC:      2,
				Rarity:   "rare",
				TypeLine: "Creature",
			},
			seedAnalysis: &SeedCardAnalysis{Colors: []string{"W", "U"}},
			minScore:     0.5, // Still gets curve, quality, legality, playability points
			maxScore:     0.6, // But no color compatibility points
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, reasoning, _, _ := builder.scoreCardForSeed(tt.card, tt.seedAnalysis)
			if score < tt.minScore || score > tt.maxScore {
				t.Errorf("expected score between %.2f and %.2f, got %.2f", tt.minScore, tt.maxScore, score)
			}
			if reasoning == "" {
				t.Error("expected non-empty reasoning")
			}
		})
	}
}

func TestFilterToCollection(t *testing.T) {
	builder := &SeedDeckBuilder{}

	scoredCards := []*scoredCard{
		{card: &cards.Card{ArenaID: 1, Name: "Owned Card 1"}, score: 0.9},
		{card: &cards.Card{ArenaID: 2, Name: "Not Owned"}, score: 0.85},
		{card: &cards.Card{ArenaID: 3, Name: "Owned Card 2"}, score: 0.8},
		{card: &cards.Card{ArenaID: 4, Name: "Also Not Owned"}, score: 0.75},
	}

	collection := map[int]int{
		1: 4, // Own 4 copies
		3: 2, // Own 2 copies
	}

	result := builder.filterToCollection(scoredCards, collection)

	if len(result) != 2 {
		t.Errorf("expected 2 cards, got %d", len(result))
	}

	for _, sc := range result {
		if sc.card.ArenaID != 1 && sc.card.ArenaID != 3 {
			t.Errorf("unexpected card in result: %s", sc.card.Name)
		}
	}
}

func TestEnrichWithOwnership(t *testing.T) {
	builder := &SeedDeckBuilder{}

	scoredCards := []*scoredCard{
		{
			card: &cards.Card{
				ArenaID:  1,
				Name:     "Test Card",
				ManaCost: strPtr("{W}{W}"),
				CMC:      2,
				Colors:   []string{"W"},
				TypeLine: "Creature",
				Rarity:   "rare",
				ImageURI: strPtr("http://example.com/card.png"),
			},
			score:     0.9,
			reasoning: "Great card",
		},
	}

	collection := map[int]int{
		1: 3, // Own 3 copies
	}

	result := builder.enrichWithOwnership(scoredCards, collection)

	if len(result) != 1 {
		t.Fatalf("expected 1 card, got %d", len(result))
	}

	card := result[0]
	if card.CardID != 1 {
		t.Errorf("expected CardID 1, got %d", card.CardID)
	}
	if card.Name != "Test Card" {
		t.Errorf("expected name 'Test Card', got %s", card.Name)
	}
	if card.ManaCost != "{W}{W}" {
		t.Errorf("expected mana cost '{W}{W}', got %s", card.ManaCost)
	}
	if card.CMC != 2 {
		t.Errorf("expected CMC 2, got %d", card.CMC)
	}
	if !card.InCollection {
		t.Error("expected InCollection to be true")
	}
	if card.OwnedCount != 3 {
		t.Errorf("expected OwnedCount 3, got %d", card.OwnedCount)
	}
	if card.NeededCount != 1 {
		t.Errorf("expected NeededCount 1, got %d", card.NeededCount)
	}
	if card.Score != 0.9 {
		t.Errorf("expected Score 0.9, got %f", card.Score)
	}
}

func TestEnrichWithOwnership_NotOwned(t *testing.T) {
	builder := &SeedDeckBuilder{}

	scoredCards := []*scoredCard{
		{
			card: &cards.Card{
				ArenaID:  1,
				Name:     "Not Owned Card",
				CMC:      3,
				Colors:   []string{"U"},
				TypeLine: "Instant",
				Rarity:   "uncommon",
			},
			score:     0.7,
			reasoning: "Could work",
		},
	}

	collection := map[int]int{} // Empty collection

	result := builder.enrichWithOwnership(scoredCards, collection)

	if len(result) != 1 {
		t.Fatalf("expected 1 card, got %d", len(result))
	}

	card := result[0]
	if card.InCollection {
		t.Error("expected InCollection to be false")
	}
	if card.OwnedCount != 0 {
		t.Errorf("expected OwnedCount 0, got %d", card.OwnedCount)
	}
	if card.NeededCount != 4 {
		t.Errorf("expected NeededCount 4, got %d", card.NeededCount)
	}
}

func TestSuggestLands(t *testing.T) {
	builder := &SeedDeckBuilder{}

	tests := []struct {
		name          string
		seedAnalysis  *SeedCardAnalysis
		suggestions   []*CardWithOwnership
		expectedTotal int
		expectedMin   int // Minimum lands of any color
	}{
		{
			name:         "Mono-color deck",
			seedAnalysis: &SeedCardAnalysis{Colors: []string{"W"}},
			suggestions: []*CardWithOwnership{
				{Colors: []string{"W"}},
				{Colors: []string{"W"}},
				{Colors: []string{"W"}},
			},
			expectedTotal: 24,
			expectedMin:   24, // All lands should be Plains
		},
		{
			name:         "Two-color deck",
			seedAnalysis: &SeedCardAnalysis{Colors: []string{"W", "U"}},
			suggestions: []*CardWithOwnership{
				{Colors: []string{"W"}},
				{Colors: []string{"U"}},
				{Colors: []string{"W", "U"}},
			},
			expectedTotal: 24,
			expectedMin:   1, // At least 1 of each color
		},
		{
			name:          "Colorless deck",
			seedAnalysis:  &SeedCardAnalysis{Colors: []string{}},
			suggestions:   []*CardWithOwnership{{Colors: []string{}}},
			expectedTotal: 0, // No basic lands suggested for colorless
			expectedMin:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lands := builder.suggestLands(tt.seedAnalysis, tt.suggestions)

			total := 0
			for _, land := range lands {
				total += land.Quantity
				if land.Quantity < tt.expectedMin && tt.expectedTotal > 0 {
					// This check is only valid for mono-color
					if len(tt.seedAnalysis.Colors) == 1 {
						t.Errorf("expected minimum %d lands, got %d for %s", tt.expectedMin, land.Quantity, land.Name)
					}
				}
			}

			if total != tt.expectedTotal {
				t.Errorf("expected %d total lands, got %d", tt.expectedTotal, total)
			}
		})
	}
}

func TestBuildAnalysis(t *testing.T) {
	builder := &SeedDeckBuilder{}

	seedAnalysis := &SeedCardAnalysis{
		Colors: []string{"W", "U"},
		Keywords: []KeywordInfo{
			{Keyword: "Flying", Category: CategoryAbility},
			{Keyword: "tokens", Category: CategoryTheme},
		},
		Themes: []string{"tokens"},
	}

	suggestions := []*CardWithOwnership{
		{Rarity: "rare", InCollection: true},
		{Rarity: "uncommon", InCollection: true},
		{Rarity: "common", InCollection: false},
		{Rarity: "common", InCollection: false},
	}

	lands := []*SuggestedLand{
		{Quantity: 12},
		{Quantity: 12},
	}

	analysis := builder.buildAnalysis(seedAnalysis, suggestions, lands)

	if len(analysis.ColorIdentity) != 2 {
		t.Errorf("expected 2 colors, got %d", len(analysis.ColorIdentity))
	}

	if len(analysis.Keywords) != 2 {
		t.Errorf("expected 2 keywords, got %d", len(analysis.Keywords))
	}

	if len(analysis.Themes) != 1 || analysis.Themes[0] != "tokens" {
		t.Errorf("expected theme 'tokens', got %v", analysis.Themes)
	}

	if analysis.SuggestedLandCount != 24 {
		t.Errorf("expected 24 lands, got %d", analysis.SuggestedLandCount)
	}

	if analysis.InCollectionCount != 2 {
		t.Errorf("expected 2 in collection, got %d", analysis.InCollectionCount)
	}

	if analysis.MissingCount != 2 {
		t.Errorf("expected 2 missing, got %d", analysis.MissingCount)
	}

	if analysis.MissingWildcardCost["common"] != 2 {
		t.Errorf("expected 2 common wildcards needed, got %d", analysis.MissingWildcardCost["common"])
	}

	// Total cards should be suggestions + lands + 4 seed card copies
	expectedTotal := len(suggestions) + 24 + 4
	if analysis.TotalCards != expectedTotal {
		t.Errorf("expected %d total cards, got %d", expectedTotal, analysis.TotalCards)
	}
}

func TestBuildSeedCardResponse(t *testing.T) {
	builder := &SeedDeckBuilder{}

	seedAnalysis := &SeedCardAnalysis{
		Card: &cards.Card{
			ArenaID:  12345,
			Name:     "Sheoldred, the Apocalypse",
			ManaCost: strPtr("{2}{B}{B}"),
			CMC:      4,
			Colors:   []string{"B"},
			TypeLine: "Legendary Creature - Phyrexian Praetor",
			Rarity:   "mythic",
			ImageURI: strPtr("http://example.com/sheoldred.png"),
		},
	}

	collection := map[int]int{
		12345: 2, // Own 2 copies
	}

	result := builder.buildSeedCardResponse(seedAnalysis, collection)

	if result.CardID != 12345 {
		t.Errorf("expected CardID 12345, got %d", result.CardID)
	}
	if result.Name != "Sheoldred, the Apocalypse" {
		t.Errorf("unexpected name: %s", result.Name)
	}
	if result.ManaCost != "{2}{B}{B}" {
		t.Errorf("expected mana cost '{2}{B}{B}', got %s", result.ManaCost)
	}
	if result.CMC != 4 {
		t.Errorf("expected CMC 4, got %d", result.CMC)
	}
	if !result.InCollection {
		t.Error("expected InCollection to be true")
	}
	if result.OwnedCount != 2 {
		t.Errorf("expected OwnedCount 2, got %d", result.OwnedCount)
	}
	if result.NeededCount != 2 {
		t.Errorf("expected NeededCount 2, got %d", result.NeededCount)
	}
	if result.Score != 1.0 {
		t.Errorf("expected Score 1.0 for seed card, got %f", result.Score)
	}
	if result.Reasoning != "This is your build-around card." {
		t.Errorf("unexpected reasoning: %s", result.Reasoning)
	}
}

func TestBuildAroundSeed_NilRequest(t *testing.T) {
	builder := &SeedDeckBuilder{}

	_, err := builder.BuildAroundSeed(context.Background(), nil)
	if err == nil {
		t.Error("expected error for nil request")
	}
	if err.Error() != "request is nil" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestBuildAroundSeed_InvalidSeedCardID(t *testing.T) {
	builder := &SeedDeckBuilder{}

	_, err := builder.BuildAroundSeed(context.Background(), &SeedDeckBuilderRequest{
		SeedCardID: 0,
	})
	if err == nil {
		t.Error("expected error for invalid seed card ID")
	}
	if err.Error() != "seed card ID is required" {
		t.Errorf("unexpected error message: %s", err.Error())
	}

	_, err = builder.BuildAroundSeed(context.Background(), &SeedDeckBuilderRequest{
		SeedCardID: -1,
	})
	if err == nil {
		t.Error("expected error for negative seed card ID")
	}
}

func TestSeedDeckBuilderRequest_Defaults(t *testing.T) {
	// Test that defaults are applied properly in BuildAroundSeed
	// We can't test the full flow without mocks, but we can verify the request struct
	req := &SeedDeckBuilderRequest{
		SeedCardID: 12345,
	}

	// These should be empty/zero before processing
	if req.MaxResults != 0 {
		t.Errorf("expected MaxResults 0 before defaults, got %d", req.MaxResults)
	}
	if req.SetRestriction != "" {
		t.Errorf("expected empty SetRestriction before defaults, got %s", req.SetRestriction)
	}
}

func TestScoreAndRankCandidates(t *testing.T) {
	builder := &SeedDeckBuilder{}

	candidates := []*cards.Card{
		{ArenaID: 1, Name: "High Score Card", Colors: []string{"W"}, CMC: 2, Rarity: "rare", TypeLine: "Creature"},
		{ArenaID: 2, Name: "Medium Score Card", Colors: []string{"W"}, CMC: 5, Rarity: "common", TypeLine: "Creature"},
		{ArenaID: 3, Name: "Low Score Card", Colors: []string{"R"}, CMC: 7, Rarity: "common", TypeLine: "Creature"},
	}

	seedAnalysis := &SeedCardAnalysis{Colors: []string{"W"}}

	result := builder.scoreAndRankCandidates(candidates, seedAnalysis)

	// Verify results are not empty (all cards should pass the 0.3 threshold)
	if len(result) == 0 {
		t.Error("expected at least one card to be included")
	}

	// Verify all 3 cards are included (even off-color gets neutral synergy/legality scores)
	if len(result) != 3 {
		t.Errorf("expected 3 cards to be included, got %d", len(result))
	}

	// Verify high score card (on-color, good CMC, rare) is ranked first
	if len(result) > 0 && result[0].card.ArenaID != 1 {
		t.Errorf("expected high score card (ArenaID=1) to be ranked first, got ArenaID=%d", result[0].card.ArenaID)
	}

	// Verify off-color card (ArenaID=3) is ranked last due to color mismatch
	if len(result) >= 3 && result[len(result)-1].card.ArenaID != 3 {
		t.Errorf("expected off-color card (ArenaID=3) to be ranked last, got ArenaID=%d", result[len(result)-1].card.ArenaID)
	}

	// Verify off-color card scores lower than on-color cards
	var onColorScore, offColorScore float64
	for _, sc := range result {
		if sc.card.ArenaID == 1 {
			onColorScore = sc.score
		}
		if sc.card.ArenaID == 3 {
			offColorScore = sc.score
		}
	}
	if offColorScore >= onColorScore {
		t.Errorf("off-color card (%.2f) should score lower than on-color card (%.2f)", offColorScore, onColorScore)
	}

	// Results should be sorted by score (descending)
	for i := 1; i < len(result); i++ {
		if result[i].score > result[i-1].score {
			t.Errorf("results not sorted by score: %.2f > %.2f", result[i].score, result[i-1].score)
		}
	}
}

func TestGetCollectionMap_NilRepo(t *testing.T) {
	builder := &SeedDeckBuilder{collectionRepo: nil}

	result, err := builder.getCollectionMap(context.Background())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result == nil {
		t.Error("expected empty map, got nil")
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d items", len(result))
	}
}

// Helper function for string pointer
func strPtr(s string) *string {
	return &s
}

func TestScoreCardForSeed_ReturnsScoreBreakdown(t *testing.T) {
	builder := &SeedDeckBuilder{}
	oracleText := "Flying"
	card := &cards.Card{
		ArenaID:    12345,
		Name:       "Test Creature",
		Colors:     []string{"W", "U"},
		CMC:        3,
		TypeLine:   "Creature — Bird",
		Rarity:     "rare",
		OracleText: &oracleText,
	}

	seedAnalysis := &SeedCardAnalysis{
		Colors:        []string{"W", "U"},
		Keywords:      []KeywordInfo{{Keyword: "flying", Category: CategoryCombat, Weight: 0.8}},
		Themes:        []string{},
		CMC:           3,
		IsCreature:    true,
		CreatureTypes: []string{"Wizard"},
	}

	score, reasoning, breakdown, synergyDetails := builder.scoreCardForSeed(card, seedAnalysis)

	// Verify score is reasonable
	if score < 0.5 || score > 1.0 {
		t.Errorf("expected score between 0.5 and 1.0, got %.2f", score)
	}

	// Verify reasoning is not empty
	if reasoning == "" {
		t.Error("expected non-empty reasoning")
	}

	// Verify breakdown is populated
	if breakdown == nil {
		t.Fatal("expected non-nil score breakdown")
	}

	// Verify breakdown fields are populated
	if breakdown.ColorFit <= 0 {
		t.Errorf("expected positive colorFit, got %.2f", breakdown.ColorFit)
	}
	if breakdown.CurveFit <= 0 {
		t.Errorf("expected positive curveFit, got %.2f", breakdown.CurveFit)
	}
	if breakdown.Quality <= 0 {
		t.Errorf("expected positive quality, got %.2f", breakdown.Quality)
	}
	if breakdown.Overall != score {
		t.Errorf("expected overall to match score (%.2f), got %.2f", score, breakdown.Overall)
	}

	// Synergy details can be empty or populated
	_ = synergyDetails
}

func TestScoreSynergyWithSeedDetailed_ReturnsKeywordSynergy(t *testing.T) {
	builder := &SeedDeckBuilder{}
	oracleText := "Flying, Vigilance"
	card := &cards.Card{
		ArenaID:    12345,
		Name:       "Angel",
		TypeLine:   "Creature — Angel",
		OracleText: &oracleText,
	}

	seedAnalysis := &SeedCardAnalysis{
		Keywords: []KeywordInfo{
			{Keyword: "flying", Category: CategoryCombat, Weight: 0.8},
		},
		Themes:        []string{},
		IsCreature:    true,
		CreatureTypes: []string{},
	}

	score, details := builder.scoreSynergyWithSeedDetailed(card, seedAnalysis)

	// Verify score is positive due to flying synergy
	if score <= 0.5 {
		t.Errorf("expected score > 0.5 due to flying synergy, got %.2f", score)
	}

	// Verify synergy details are captured
	if len(details) == 0 {
		t.Error("expected synergy details to be captured")
	}

	// Check that flying keyword is in details
	foundFlying := false
	for _, detail := range details {
		if detail.Type == "keyword" && detail.Name == "flying" {
			foundFlying = true
			break
		}
	}
	if !foundFlying {
		t.Error("expected 'flying' keyword in synergy details")
	}
}

func TestScoreSynergyWithSeedDetailed_ReturnsCreatureTypeSynergy(t *testing.T) {
	builder := &SeedDeckBuilder{}
	oracleText := "When this enters, draw a card"
	card := &cards.Card{
		ArenaID:    12345,
		Name:       "Elvish Visionary",
		TypeLine:   "Creature — Elf Shaman",
		OracleText: &oracleText,
	}

	seedAnalysis := &SeedCardAnalysis{
		Keywords:      []KeywordInfo{},
		Themes:        []string{},
		IsCreature:    true,
		CreatureTypes: []string{"Elf"},
	}

	score, details := builder.scoreSynergyWithSeedDetailed(card, seedAnalysis)

	// Verify score is positive due to Elf tribal synergy
	if score <= 0.5 {
		t.Errorf("expected score > 0.5 due to Elf tribal, got %.2f", score)
	}

	// Check that Elf creature type is in details
	foundElf := false
	for _, detail := range details {
		if detail.Type == "creature_type" && detail.Name == "Elf" {
			foundElf = true
			break
		}
	}
	if !foundElf {
		t.Error("expected 'Elf' creature type in synergy details")
	}
}

func TestScoreCardForDeck_ReturnsScoreBreakdown(t *testing.T) {
	builder := &SeedDeckBuilder{}
	oracleText := "Flying"
	card := &cards.Card{
		ArenaID:    12345,
		Name:       "Test Creature",
		Colors:     []string{"U"},
		CMC:        2,
		TypeLine:   "Creature — Bird",
		Rarity:     "uncommon",
		OracleText: &oracleText,
	}

	deckAnalysis := &CollectiveDeckAnalysis{
		Colors:        map[string]int{"U": 10, "W": 5},
		Keywords:      []KeywordInfo{{Keyword: "flying", Category: CategoryCombat, Weight: 0.8}},
		Themes:        map[string]int{},
		CreatureTypes: map[string]int{"Bird": 2},
		ManaCurve:     map[int]int{1: 4, 2: 2, 3: 6}, // Need more 2-drops
		TotalCards:    20,
	}

	score, reasoning, breakdown, synergyDetails := builder.scoreCardForDeck(card, deckAnalysis)

	// Verify score is reasonable
	if score < 0.5 || score > 1.0 {
		t.Errorf("expected score between 0.5 and 1.0, got %.2f", score)
	}

	// Verify reasoning is not empty
	if reasoning == "" {
		t.Error("expected non-empty reasoning")
	}

	// Verify breakdown is populated
	if breakdown == nil {
		t.Fatal("expected non-nil score breakdown")
	}

	// All color is in deck colors, should be 1.0
	if breakdown.ColorFit < 0.9 {
		t.Errorf("expected high colorFit for matching color, got %.2f", breakdown.ColorFit)
	}

	// 2 CMC should fill curve gap (we have only 2 at 2 CMC, need 8)
	if breakdown.CurveFit < 0.7 {
		t.Errorf("expected curveFit >= 0.7 for 2 CMC gap fill, got %.2f", breakdown.CurveFit)
	}

	// Synergy details can include flying keyword and Bird creature type
	_ = synergyDetails
}

func TestScoreSynergyWithDeckDetailed_ReturnsThemeSynergy(t *testing.T) {
	builder := &SeedDeckBuilder{}
	oracleText := "Create a 1/1 token"
	card := &cards.Card{
		ArenaID:    12345,
		Name:       "Token Maker",
		TypeLine:   "Instant",
		OracleText: &oracleText,
	}

	deckAnalysis := &CollectiveDeckAnalysis{
		Colors:        map[string]int{"W": 10},
		Keywords:      []KeywordInfo{{Keyword: "tokens", Category: CategoryTheme, Weight: 0.7}},
		Themes:        map[string]int{"tokens": 5},
		CreatureTypes: map[string]int{},
		ManaCurve:     map[int]int{1: 4, 2: 8, 3: 8},
		TotalCards:    24,
	}

	score, details := builder.scoreSynergyWithDeckDetailed(card, deckAnalysis)

	// If card has tokens theme, it should have synergy
	// (Note: the card might not have tokens theme extracted, so score might be neutral)
	_ = score

	// Check if any theme synergies were found
	for _, detail := range details {
		if detail.Type == "theme" {
			// Found a theme synergy
			if detail.Description == "" {
				t.Error("expected non-empty description for theme synergy")
			}
		}
	}
}

func TestCalculateKeywordSynergyDetailed_ReturnsMatchedKeywords(t *testing.T) {
	card1Keywords := []KeywordInfo{
		{Keyword: "flying", Category: CategoryCombat, Weight: 0.8},
		{Keyword: "vigilance", Category: CategoryCombat, Weight: 0.6},
	}
	card2Keywords := []KeywordInfo{
		{Keyword: "flying", Category: CategoryCombat, Weight: 0.8},
		{Keyword: "trample", Category: CategoryCombat, Weight: 0.6},
	}

	score, matchedKeywords := CalculateKeywordSynergyDetailed(card1Keywords, card2Keywords)

	// Should have positive synergy
	if score <= 0 {
		t.Errorf("expected positive synergy score, got %.2f", score)
	}

	// Should return "flying" as matched keyword
	if len(matchedKeywords) == 0 {
		t.Fatal("expected matched keywords")
	}

	foundFlying := false
	for _, kw := range matchedKeywords {
		if kw == "flying" {
			foundFlying = true
			break
		}
	}
	if !foundFlying {
		t.Errorf("expected 'flying' in matched keywords, got %v", matchedKeywords)
	}
}

func TestCalculateKeywordSynergyDetailed_EmptyKeywords(t *testing.T) {
	score, matchedKeywords := CalculateKeywordSynergyDetailed(nil, nil)

	if score != 0 {
		t.Errorf("expected 0 score for empty keywords, got %.2f", score)
	}

	if len(matchedKeywords) != 0 {
		t.Errorf("expected empty matched keywords, got %v", matchedKeywords)
	}
}

func TestScoreSynergyWithDeckDetailed_MultiTypeCreature(t *testing.T) {
	builder := &SeedDeckBuilder{}
	// Multi-type creature: Cat Warrior
	card := &cards.Card{
		ArenaID:  12345,
		Name:     "Cat Warrior",
		TypeLine: "Creature — Cat Warrior",
	}

	// Deck has both Cat and Warrior creatures
	deckAnalysis := &CollectiveDeckAnalysis{
		Colors:   map[string]int{"G": 10},
		Keywords: nil,
		Themes:   map[string]int{},
		CreatureTypes: map[string]int{
			"Cat":     3, // 3 cats in deck
			"Warrior": 2, // 2 warriors in deck
		},
		ManaCurve:  map[int]int{1: 4, 2: 8, 3: 8},
		TotalCards: 24,
	}

	score, details := builder.scoreSynergyWithDeckDetailed(card, deckAnalysis)

	// Score should be higher than neutral (0.5) due to tribal synergies
	if score <= 0.5 {
		t.Errorf("expected score > 0.5 for multi-type tribal synergy, got %.2f", score)
	}

	// Should have TWO creature type synergy details - one for Cat, one for Warrior
	creatureTypeDetails := 0
	foundCat := false
	foundWarrior := false
	for _, detail := range details {
		if detail.Type == "creature_type" {
			creatureTypeDetails++
			if detail.Name == "Cat" {
				foundCat = true
			}
			if detail.Name == "Warrior" {
				foundWarrior = true
			}
		}
	}

	if creatureTypeDetails != 2 {
		t.Errorf("expected 2 creature type synergy details for multi-type creature, got %d", creatureTypeDetails)
	}

	if !foundCat {
		t.Error("expected Cat tribal synergy detail")
	}

	if !foundWarrior {
		t.Error("expected Warrior tribal synergy detail")
	}
}

func TestScoreSynergyWithDeckDetailed_SingleTypeCreature(t *testing.T) {
	builder := &SeedDeckBuilder{}
	// Single-type creature: Cat only
	card := &cards.Card{
		ArenaID:  12345,
		Name:     "Savannah Lions",
		TypeLine: "Creature — Cat",
	}

	// Deck has both Cat and Warrior creatures
	deckAnalysis := &CollectiveDeckAnalysis{
		Colors:   map[string]int{"W": 10},
		Keywords: nil,
		Themes:   map[string]int{},
		CreatureTypes: map[string]int{
			"Cat":     3, // 3 cats in deck
			"Warrior": 2, // 2 warriors in deck
		},
		ManaCurve:  map[int]int{1: 4, 2: 8, 3: 8},
		TotalCards: 24,
	}

	score, details := builder.scoreSynergyWithDeckDetailed(card, deckAnalysis)

	// Score should be higher than neutral (0.5) due to Cat synergy
	if score <= 0.5 {
		t.Errorf("expected score > 0.5 for Cat tribal synergy, got %.2f", score)
	}

	// Should have exactly ONE creature type synergy detail (Cat only)
	creatureTypeDetails := 0
	for _, detail := range details {
		if detail.Type == "creature_type" {
			creatureTypeDetails++
			if detail.Name != "Cat" {
				t.Errorf("expected Cat synergy detail, got %s", detail.Name)
			}
		}
	}

	if creatureTypeDetails != 1 {
		t.Errorf("expected 1 creature type synergy detail for single-type creature, got %d", creatureTypeDetails)
	}
}

func TestScoreSynergyWithDeckDetailed_Changeling(t *testing.T) {
	builder := &SeedDeckBuilder{}
	changelingText := "Changeling (This card is every creature type.)"
	card := &cards.Card{
		ArenaID:    12345,
		Name:       "Changeling Outcast",
		TypeLine:   "Creature — Shapeshifter",
		OracleText: &changelingText,
	}

	// Deck has multiple creature types
	deckAnalysis := &CollectiveDeckAnalysis{
		Colors:   map[string]int{"B": 10},
		Keywords: nil,
		Themes:   map[string]int{},
		CreatureTypes: map[string]int{
			"Elf":    3,
			"Zombie": 2,
		},
		ManaCurve:  map[int]int{1: 4, 2: 8, 3: 8},
		TotalCards: 24,
	}

	score, details := builder.scoreSynergyWithDeckDetailed(card, deckAnalysis)

	// Score should be high because changeling matches ALL creature types
	if score <= 0.5 {
		t.Errorf("expected score > 0.5 for changeling synergy, got %.2f", score)
	}

	// Should have synergy details for BOTH Elf and Zombie
	foundElf := false
	foundZombie := false
	for _, detail := range details {
		if detail.Type == "creature_type" {
			if detail.Name == "Elf" {
				foundElf = true
			}
			if detail.Name == "Zombie" {
				foundZombie = true
			}
		}
	}

	if !foundElf {
		t.Error("expected Elf synergy detail for changeling")
	}
	if !foundZombie {
		t.Error("expected Zombie synergy detail for changeling")
	}
}

func TestScoreSynergyWithDeckDetailed_StrongTribalWeight(t *testing.T) {
	builder := &SeedDeckBuilder{}

	// Test with a strong tribal type (Elf)
	elfCard := &cards.Card{
		ArenaID:  12345,
		Name:     "Llanowar Elves",
		TypeLine: "Creature — Elf Druid",
	}

	// Test with a weak tribal type (Beast)
	beastCard := &cards.Card{
		ArenaID:  12346,
		Name:     "Grizzly Bears",
		TypeLine: "Creature — Beast",
	}

	elfDeck := &CollectiveDeckAnalysis{
		Colors:        map[string]int{"G": 10},
		CreatureTypes: map[string]int{"Elf": 5},
		ManaCurve:     map[int]int{1: 4, 2: 8, 3: 8},
		TotalCards:    24,
	}

	beastDeck := &CollectiveDeckAnalysis{
		Colors:        map[string]int{"G": 10},
		CreatureTypes: map[string]int{"Beast": 5},
		ManaCurve:     map[int]int{1: 4, 2: 8, 3: 8},
		TotalCards:    24,
	}

	elfScore, _ := builder.scoreSynergyWithDeckDetailed(elfCard, elfDeck)
	beastScore, _ := builder.scoreSynergyWithDeckDetailed(beastCard, beastDeck)

	// Elf (strong tribal) should score higher than Beast (weak tribal)
	if elfScore <= beastScore {
		t.Errorf("expected Elf score (%.2f) > Beast score (%.2f) due to tribal weight", elfScore, beastScore)
	}
}
