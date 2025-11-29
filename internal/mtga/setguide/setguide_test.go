package setguide

import (
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

func TestGuildNames(t *testing.T) {
	tests := []struct {
		colors   string
		expected string
	}{
		{"WU", "Azorius"},
		{"UW", "Azorius"},
		{"UB", "Dimir"},
		{"BR", "Rakdos"},
		{"RG", "Gruul"},
		{"GW", "Selesnya"},
		{"WB", "Orzhov"},
		{"UR", "Izzet"},
		{"BG", "Golgari"},
		{"RW", "Boros"},
		{"GU", "Simic"},
	}

	for _, tt := range tests {
		t.Run(tt.colors, func(t *testing.T) {
			got := guildNames[tt.colors]
			if got != tt.expected {
				t.Errorf("guildNames[%s] = %s, want %s", tt.colors, got, tt.expected)
			}
		})
	}
}

func TestContainsColor(t *testing.T) {
	tests := []struct {
		colors   []string
		target   string
		expected bool
	}{
		{[]string{"W", "U"}, "W", true},
		{[]string{"W", "U"}, "B", false},
		{[]string{}, "W", false},
		{[]string{"R", "G", "B"}, "G", true},
	}

	for _, tt := range tests {
		got := containsColor(tt.colors, tt.target)
		if got != tt.expected {
			t.Errorf("containsColor(%v, %s) = %v, want %v", tt.colors, tt.target, got, tt.expected)
		}
	}
}

func TestEstimateCMCForColor(t *testing.T) {
	tests := []struct {
		color    string
		expected float64
	}{
		{"W", 2.5},
		{"U", 3.0},
		{"B", 3.0},
		{"R", 2.5},
		{"G", 3.5},
		{"X", 3.0}, // Unknown color
	}

	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			got := estimateCMCForColor(tt.color)
			if got != tt.expected {
				t.Errorf("estimateCMCForColor(%s) = %v, want %v", tt.color, got, tt.expected)
			}
		})
	}
}

func TestDetermineCurveFromStyle(t *testing.T) {
	tests := []struct {
		style    string
		expected string
	}{
		{"Aggro", "Low (1-3)"},
		{"Tempo", "Low-Medium (2-3)"},
		{"Spells", "Low-Medium (2-3)"},
		{"Midrange", "Medium (3-4)"},
		{"Control", "High (4+)"},
		{"Unknown", "Medium (3-4)"},
	}

	for _, tt := range tests {
		t.Run(tt.style, func(t *testing.T) {
			got := determineCurveFromStyle(tt.style)
			if got != tt.expected {
				t.Errorf("determineCurveFromStyle(%s) = %s, want %s", tt.style, got, tt.expected)
			}
		})
	}
}

func TestClassifyStyleFromColors(t *testing.T) {
	tests := []struct {
		colors   []string
		avgCMC   float64
		expected string
	}{
		// Aggressive combos
		{[]string{"R", "W"}, 2.5, "Aggro"},
		{[]string{"R", "B"}, 2.5, "Aggro"},
		// Control combos
		{[]string{"U", "W"}, 3.5, "Control"},
		{[]string{"U", "B"}, 3.5, "Control"},
		// Tempo
		{[]string{"U", "W"}, 3.0, "Tempo"},
		// Green combos
		{[]string{"G", "B"}, 3.0, "Midrange"},
		{[]string{"G", "W"}, 3.0, "Go-Wide"},
		{[]string{"G", "R"}, 3.0, "Stompy"},
		// Spells
		{[]string{"U", "R"}, 3.0, "Spells"},
	}

	for _, tt := range tests {
		name := tt.colors[0] + tt.colors[1]
		t.Run(name, func(t *testing.T) {
			got := classifyStyleFromColors(tt.colors, tt.avgCMC)
			if got != tt.expected {
				t.Errorf("classifyStyleFromColors(%v, %v) = %s, want %s", tt.colors, tt.avgCMC, got, tt.expected)
			}
		})
	}
}

func TestGenerateStrategyDescription(t *testing.T) {
	tests := []struct {
		guildName string
		style     string
		winRate   float64
	}{
		{"Boros", "Aggro", 0.55},
		{"Azorius", "Control", 0.52},
		{"Golgari", "Midrange", 0.50},
	}

	for _, tt := range tests {
		t.Run(tt.guildName+" "+tt.style, func(t *testing.T) {
			got := generateStrategyDescription(tt.guildName, tt.style, tt.winRate)
			if got == "" {
				t.Error("generateStrategyDescription returned empty string")
			}
			// Check that it contains the win rate
			if !stringContains(got, "55.0%") && !stringContains(got, "52.0%") && !stringContains(got, "50.0%") {
				t.Errorf("strategy description should contain win rate, got: %s", got)
			}
		})
	}
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestGenerateArchetypesFromColorPairs(t *testing.T) {
	sg := NewSetGuide(nil, "")

	colorPairs := []struct {
		Colors  string
		WinRate float64
	}{
		{"WU", 0.55},
		{"BR", 0.53},
		{"GW", 0.52},
		{"UB", 0.51},
		{"RG", 0.50},
		{"WB", 0.49}, // This should be excluded (only top 5)
	}

	setFile := &seventeenlands.SetFile{
		CardRatings: map[string]*seventeenlands.CardRatingData{},
	}

	archetypes := sg.generateArchetypesFromColorPairs("TEST", colorPairs, setFile)

	if len(archetypes) != 5 {
		t.Errorf("Expected 5 archetypes, got %d", len(archetypes))
	}

	// Check first archetype
	if archetypes[0].WinRate != 0.55 {
		t.Errorf("First archetype should have highest win rate 0.55, got %f", archetypes[0].WinRate)
	}

	// Check that archetypes are stored in cache
	cached, err := sg.GetArchetypes("TEST")
	if err != nil {
		t.Errorf("GetArchetypes returned error: %v", err)
	}
	if len(cached) != 5 {
		t.Errorf("Cached archetypes should have 5 items, got %d", len(cached))
	}
}

func TestCreateArchetypeFromColorPair(t *testing.T) {
	sg := NewSetGuide(nil, "")

	setFile := &seventeenlands.SetFile{
		CardRatings: map[string]*seventeenlands.CardRatingData{
			"Test Card": {
				Name:   "Test Card",
				Colors: []string{"W"},
				CMC:    2.0,
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 55.0, GIH: 100},
				},
			},
		},
	}

	archetype := sg.createArchetypeFromColorPair("WU", 0.55, setFile)

	if archetype.Name == "" {
		t.Error("Archetype name should not be empty")
	}

	if len(archetype.Colors) != 2 {
		t.Errorf("Expected 2 colors, got %d", len(archetype.Colors))
	}

	if archetype.WinRate != 0.55 {
		t.Errorf("Expected win rate 0.55, got %f", archetype.WinRate)
	}

	if archetype.Strategy == "" {
		t.Error("Strategy should not be empty")
	}

	if archetype.Curve == "" {
		t.Error("Curve should not be empty")
	}
}

func TestFindKeyCardsForColors(t *testing.T) {
	sg := NewSetGuide(nil, "")

	setFile := &seventeenlands.SetFile{
		CardRatings: map[string]*seventeenlands.CardRatingData{
			"White Card": {
				Name:   "White Card",
				Colors: []string{"W"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 60.0, GIH: 100},
				},
			},
			"Blue Card": {
				Name:   "Blue Card",
				Colors: []string{"U"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 55.0, GIH: 100},
				},
			},
			"Red Card": {
				Name:   "Red Card",
				Colors: []string{"R"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 58.0, GIH: 100},
				},
			},
			"Gold Card": {
				Name:   "Gold Card",
				Colors: []string{"W", "U"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 65.0, GIH: 100},
				},
			},
			"Low Sample Card": {
				Name:   "Low Sample Card",
				Colors: []string{"W"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 70.0, GIH: 10}, // Too few games
				},
			},
		},
	}

	keyCards := sg.findKeyCardsForColors([]string{"W", "U"}, setFile)

	// Should find Gold Card (65%), White Card (60%), Blue Card (55%)
	// Should not include Red Card (wrong color) or Low Sample Card (too few games)
	if len(keyCards) != 3 {
		t.Errorf("Expected 3 key cards, got %d: %v", len(keyCards), keyCards)
	}

	// Gold card should be first (highest GIHWR)
	if len(keyCards) > 0 && keyCards[0] != "Gold Card" {
		t.Errorf("Expected Gold Card to be first (highest GIHWR), got %s", keyCards[0])
	}
}

func TestDetermineArchetypeStyle(t *testing.T) {
	sg := NewSetGuide(nil, "")

	// Create a set file with cards that have actual CMC values
	setFile := &seventeenlands.SetFile{
		CardRatings: map[string]*seventeenlands.CardRatingData{
			"Cheap Card 1": {Colors: []string{"R"}, CMC: 1.0},
			"Cheap Card 2": {Colors: []string{"R"}, CMC: 2.0},
			"Cheap Card 3": {Colors: []string{"W"}, CMC: 2.0},
			"Mid Card":     {Colors: []string{"W"}, CMC: 3.0},
		},
	}

	style := sg.determineArchetypeStyle([]string{"R", "W"}, setFile)

	// With low average CMC and R/W colors, should be Aggro
	if style != "Aggro" {
		t.Errorf("Expected Aggro style for low CMC RW deck, got %s", style)
	}
}
