package archetype

import (
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
)

func TestClassifier_DetectColorIdentity(t *testing.T) {
	c := &Classifier{}

	tests := []struct {
		name        string
		colorCounts map[string]int
		want        string
	}{
		{
			name:        "mono white",
			colorCounts: map[string]int{"W": 10},
			want:        "W",
		},
		{
			name:        "azorius (WU)",
			colorCounts: map[string]int{"W": 10, "U": 10},
			want:        "WU",
		},
		{
			name:        "rakdos (BR)",
			colorCounts: map[string]int{"B": 8, "R": 12},
			want:        "BR",
		},
		{
			name:        "jund (BRG)",
			colorCounts: map[string]int{"B": 5, "R": 8, "G": 7},
			want:        "BRG",
		},
		{
			name:        "five color",
			colorCounts: map[string]int{"W": 2, "U": 3, "B": 4, "R": 5, "G": 6},
			want:        "WUBRG",
		},
		{
			name:        "colorless",
			colorCounts: map[string]int{},
			want:        "C",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.detectColorIdentity(tt.colorCounts)
			if got != tt.want {
				t.Errorf("detectColorIdentity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifier_GetDominantColors(t *testing.T) {
	c := &Classifier{}

	tests := []struct {
		name        string
		colorCounts map[string]int
		totalCards  int
		wantLen     int
		wantFirst   string
	}{
		{
			name:        "clear two-color deck",
			colorCounts: map[string]int{"W": 15, "U": 15, "B": 2},
			totalCards:  40,
			wantLen:     2,
			wantFirst:   "W", // Same count, but W comes first
		},
		{
			name:        "mono-color splash",
			colorCounts: map[string]int{"R": 25, "B": 3},
			totalCards:  40,
			wantLen:     1,
			wantFirst:   "R",
		},
		{
			name:        "three even colors",
			colorCounts: map[string]int{"B": 10, "R": 10, "G": 10},
			totalCards:  40,
			wantLen:     3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.getDominantColors(tt.colorCounts, tt.totalCards)
			if len(got) != tt.wantLen {
				t.Errorf("getDominantColors() len = %v, want %v", len(got), tt.wantLen)
			}
			if tt.wantFirst != "" && len(got) > 0 && got[0] != tt.wantFirst {
				t.Errorf("getDominantColors() first = %v, want %v", got[0], tt.wantFirst)
			}
		})
	}
}

func TestClassifier_DetectColorPair(t *testing.T) {
	c := &Classifier{}

	tests := []struct {
		name           string
		dominantColors []string
		wantPairName   string
		wantNil        bool
	}{
		{
			name:           "azorius",
			dominantColors: []string{"W", "U"},
			wantPairName:   "Azorius",
		},
		{
			name:           "dimir",
			dominantColors: []string{"U", "B"},
			wantPairName:   "Dimir",
		},
		{
			name:           "rakdos",
			dominantColors: []string{"B", "R"},
			wantPairName:   "Rakdos",
		},
		{
			name:           "gruul",
			dominantColors: []string{"R", "G"},
			wantPairName:   "Gruul",
		},
		{
			name:           "selesnya",
			dominantColors: []string{"G", "W"},
			wantPairName:   "Selesnya",
		},
		{
			name:           "orzhov",
			dominantColors: []string{"W", "B"},
			wantPairName:   "Orzhov",
		},
		{
			name:           "izzet",
			dominantColors: []string{"U", "R"},
			wantPairName:   "Izzet",
		},
		{
			name:           "golgari",
			dominantColors: []string{"B", "G"},
			wantPairName:   "Golgari",
		},
		{
			name:           "boros",
			dominantColors: []string{"R", "W"},
			wantPairName:   "Boros",
		},
		{
			name:           "simic",
			dominantColors: []string{"G", "U"},
			wantPairName:   "Simic",
		},
		{
			name:           "mono color - no pair",
			dominantColors: []string{"R"},
			wantNil:        true,
		},
		{
			name:           "three color - no pair",
			dominantColors: []string{"W", "U", "B"},
			wantNil:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.detectColorPair(tt.dominantColors)
			if tt.wantNil {
				if got != nil {
					t.Errorf("detectColorPair() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Errorf("detectColorPair() = nil, want %v", tt.wantPairName)
				return
			}
			if got.Name != tt.wantPairName {
				t.Errorf("detectColorPair() = %v, want %v", got.Name, tt.wantPairName)
			}
		})
	}
}

func TestClassifier_AnalyzeDeck(t *testing.T) {
	c := &Classifier{}

	// Create mock card metadata
	cardMetadata := map[int]*cards.Card{
		1: {
			ArenaID:  1,
			Name:     "Elite Vanguard",
			TypeLine: "Creature — Human Soldier",
			Colors:   []string{"W"},
			CMC:      1,
			Rarity:   "common",
		},
		2: {
			ArenaID:  2,
			Name:     "Serra Angel",
			TypeLine: "Creature — Angel",
			Colors:   []string{"W"},
			CMC:      5,
			Rarity:   "uncommon",
		},
		3: {
			ArenaID:  3,
			Name:     "Counterspell",
			TypeLine: "Instant",
			Colors:   []string{"U"},
			CMC:      2,
			Rarity:   "common",
		},
		4: {
			ArenaID:  4,
			Name:     "Sol Ring",
			TypeLine: "Artifact",
			Colors:   []string{},
			CMC:      1,
			Rarity:   "rare",
		},
		5: {
			ArenaID:  5,
			Name:     "Plains",
			TypeLine: "Basic Land — Plains",
			Colors:   []string{},
			CMC:      0,
			Rarity:   "common",
		},
	}

	quantities := map[int]int{
		1: 4,  // Elite Vanguard x4
		2: 2,  // Serra Angel x2
		3: 4,  // Counterspell x4
		4: 1,  // Sol Ring x1
		5: 10, // Plains x10
	}

	analysis := c.analyzeDeck(cardMetadata, quantities)

	// Test color counts
	if analysis.ColorCounts["W"] != 6 {
		t.Errorf("ColorCounts[W] = %d, want 6", analysis.ColorCounts["W"])
	}
	if analysis.ColorCounts["U"] != 4 {
		t.Errorf("ColorCounts[U] = %d, want 4", analysis.ColorCounts["U"])
	}

	// Test colorless count
	if analysis.ColorlessCount != 11 { // Sol Ring (1) + Plains (10)
		t.Errorf("ColorlessCount = %d, want 11", analysis.ColorlessCount)
	}

	// Test type counts
	if analysis.CreatureCount != 6 {
		t.Errorf("CreatureCount = %d, want 6", analysis.CreatureCount)
	}
	if analysis.InstantCount != 4 {
		t.Errorf("InstantCount = %d, want 4", analysis.InstantCount)
	}
	if analysis.ArtifactCount != 1 {
		t.Errorf("ArtifactCount = %d, want 1", analysis.ArtifactCount)
	}
	if analysis.LandCount != 10 {
		t.Errorf("LandCount = %d, want 10", analysis.LandCount)
	}

	// Test mana curve
	if analysis.ManaCurve[1] != 5 { // Elite Vanguard x4 + Sol Ring x1
		t.Errorf("ManaCurve[1] = %d, want 5", analysis.ManaCurve[1])
	}
	if analysis.ManaCurve[2] != 4 { // Counterspell x4
		t.Errorf("ManaCurve[2] = %d, want 4", analysis.ManaCurve[2])
	}
	if analysis.ManaCurve[5] != 2 { // Serra Angel x2
		t.Errorf("ManaCurve[5] = %d, want 2", analysis.ManaCurve[5])
	}
}

func TestClassifier_DetectDeckStyle(t *testing.T) {
	c := &Classifier{}

	tests := []struct {
		name     string
		analysis *DeckAnalysis
		want     string
	}{
		{
			name: "aggro deck",
			analysis: &DeckAnalysis{
				CreatureCount: 25,
				InstantCount:  4,
				SorceryCount:  2,
				ManaCurve:     map[int]int{1: 8, 2: 10, 3: 5, 4: 2},
				AvgCMC:        2.0,
			},
			want: "Aggro",
		},
		{
			name: "control deck",
			analysis: &DeckAnalysis{
				CreatureCount: 6,
				InstantCount:  12,
				SorceryCount:  8,
				ManaCurve:     map[int]int{2: 4, 3: 6, 4: 5, 5: 6, 6: 3, 7: 2},
				AvgCMC:        4.0,
			},
			want: "Control",
		},
		{
			name: "midrange deck",
			analysis: &DeckAnalysis{
				CreatureCount: 18,
				InstantCount:  6,
				SorceryCount:  4,
				ManaCurve:     map[int]int{1: 2, 2: 6, 3: 8, 4: 6, 5: 4, 6: 2},
				AvgCMC:        3.0,
			},
			want: "Midrange",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.detectDeckStyle(tt.analysis)
			if got != tt.want {
				t.Errorf("detectDeckStyle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClassifier_FindArchetypeIndicators(t *testing.T) {
	c := &Classifier{}

	cardMetadata := map[int]*cards.Card{
		1: {
			ArenaID:    1,
			Name:       "Empyrean Eagle",
			TypeLine:   "Creature — Bird Spirit",
			OracleText: stringPtr("Flying\nOther creatures you control with flying get +1/+1."),
			Colors:     []string{"W", "U"},
		},
		2: {
			ArenaID:    2,
			Name:       "Woe Strider",
			TypeLine:   "Creature — Horror",
			OracleText: stringPtr("Sacrifice another creature: Scry 1."),
			Colors:     []string{"B"},
		},
		3: {
			ArenaID:    3,
			Name:       "Young Pyromancer",
			TypeLine:   "Creature — Human Shaman",
			OracleText: stringPtr("Whenever you cast an instant or sorcery spell, create a 1/1 red Elemental creature token."),
			Colors:     []string{"R"},
		},
	}

	quantities := map[int]int{1: 1, 2: 1, 3: 1}

	indicators := c.findArchetypeIndicators(cardMetadata, quantities)

	if len(indicators) == 0 {
		t.Error("expected at least one indicator")
		return
	}

	// Should find Empyrean Eagle as a flying payoff
	foundFlying := false
	foundSpells := false
	for _, ind := range indicators {
		if ind.CardName == "Empyrean Eagle" && ind.Reason == "Flying synergy payoff" {
			foundFlying = true
		}
		if ind.CardName == "Young Pyromancer" && ind.Reason == "Spells matter" {
			foundSpells = true
		}
	}

	if !foundFlying {
		t.Error("expected Empyrean Eagle to be identified as flying payoff")
	}
	if !foundSpells {
		t.Error("expected Young Pyromancer to be identified as spells matter")
	}
}

func TestClassifier_ClassifyCards(t *testing.T) {
	// Skip this test in unit tests since it requires the cards service
	// This test is for integration testing
	t.Skip("Integration test - requires cards service")
}

func stringPtr(s string) *string {
	return &s
}
