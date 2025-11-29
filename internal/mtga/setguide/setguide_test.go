package setguide

import (
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

func TestGetPrimaryCardType(t *testing.T) {
	tests := []struct {
		name     string
		types    []string
		expected string
	}{
		{"Empty types", []string{}, "Unknown"},
		{"Single creature", []string{"Creature"}, "Creature"},
		{"Artifact creature", []string{"Artifact", "Creature"}, "Creature"},
		{"Enchantment creature", []string{"Enchantment", "Creature"}, "Creature"},
		{"Instant", []string{"Instant"}, "Instant"},
		{"Sorcery", []string{"Sorcery"}, "Sorcery"},
		{"Planeswalker", []string{"Planeswalker"}, "Planeswalker"},
		{"Basic land", []string{"Land", "Basic"}, "Land"},
		{"Enchantment", []string{"Enchantment"}, "Enchantment"},
		{"Artifact", []string{"Artifact"}, "Artifact"},
		{"Unknown type", []string{"Tribal"}, "Tribal"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPrimaryCardType(tt.types)
			if got != tt.expected {
				t.Errorf("getPrimaryCardType(%v) = %s, want %s", tt.types, got, tt.expected)
			}
		})
	}
}

func TestCategorizeCard(t *testing.T) {
	tests := []struct {
		name     string
		cardName string
		gihwr    float64
		cardType string
		expected string
	}{
		// Removal spells
		{"Murder instant", "Murder", 0.55, "Instant", "Removal"},
		{"Destroy sorcery", "Destroy Evil", 0.53, "Sorcery", "Removal"},
		{"Lightning bolt", "Lightning Bolt", 0.58, "Instant", "Removal"},
		{"Exile spell", "Exile to the Beyond", 0.54, "Sorcery", "Removal"},

		// Bombs
		{"High win rate creature", "Amazing Dragon", 0.65, "Creature", "Bomb"},
		{"High win rate instant", "Power Spell", 0.62, "Instant", "Bomb"},
		{"Planeswalker", "Jace the Wise", 0.55, "Planeswalker", "Bomb"},

		// Combat tricks
		{"Giant Growth", "Giant Growth", 0.52, "Instant", "Combat Trick"},
		{"Mighty Leap", "Mighty Leap", 0.50, "Instant", "Combat Trick"},

		// Land fixing
		{"Evolving Wilds", "Evolving Wilds", 0.48, "Land", "Fixing"},
		{"Terramorphic Expanse", "Terramorphic Expanse", 0.47, "Land", "Fixing"},

		// Type-based categorization
		{"Regular creature", "Grizzly Bears", 0.50, "Creature", "Creature"},
		{"Regular enchantment", "Pacifism", 0.52, "Enchantment", "Enchantment"},
		{"Regular artifact", "Mana Rock", 0.51, "Artifact", "Artifact"},
		{"Regular land", "Mountain", 0.45, "Land", "Land"},

		// Unknown defaults to playable
		{"Unknown type", "Mystery Card", 0.50, "Unknown", "Playable"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := categorizeCard(tt.cardName, tt.gihwr, tt.cardType)
			if got != tt.expected {
				t.Errorf("categorizeCard(%s, %v, %s) = %s, want %s", tt.cardName, tt.gihwr, tt.cardType, got, tt.expected)
			}
		})
	}
}

func TestStringContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"Lightning Bolt", "lightning", true},
		{"Lightning Bolt", "BOLT", true},
		{"Murder", "murder", true},
		{"Hello World", "xyz", false},
		{"Short", "LongSubstring", false},
		{"", "test", false},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			got := stringContainsIgnoreCase(tt.s, tt.substr)
			if got != tt.expected {
				t.Errorf("stringContainsIgnoreCase(%s, %s) = %v, want %v", tt.s, tt.substr, got, tt.expected)
			}
		})
	}
}

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
			if !stringContainsIgnoreCase(got, "55.0%") && !stringContainsIgnoreCase(got, "52.0%") && !stringContainsIgnoreCase(got, "50.0%") {
				t.Errorf("strategy description should contain win rate, got: %s", got)
			}
		})
	}
}

func TestCalculateTypeStats(t *testing.T) {
	sg := NewSetGuide(nil, "")

	setFile := &seventeenlands.SetFile{
		CardRatings: map[string]*seventeenlands.CardRatingData{
			"1": {
				Name:   "Creature 1",
				Types:  []string{"Creature"},
				Colors: []string{"W"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 55.0, GIH: 100},
				},
			},
			"2": {
				Name:   "Creature 2",
				Types:  []string{"Creature"},
				Colors: []string{"U"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 50.0, GIH: 100},
				},
			},
			"3": {
				Name:   "Instant 1",
				Types:  []string{"Instant"},
				Colors: []string{"R"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 53.0, GIH: 100},
				},
			},
			"4": {
				Name:   "Land 1",
				Types:  []string{"Land"},
				Colors: []string{},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 45.0, GIH: 100},
				},
			},
		},
	}

	stats := sg.calculateTypeStats(setFile)

	if len(stats) < 3 {
		t.Errorf("Expected at least 3 type stats, got %d", len(stats))
	}

	// Find creature stats
	var creatureStats *TypeStats
	for i := range stats {
		if stats[i].Type == "Creature" {
			creatureStats = &stats[i]
			break
		}
	}

	if creatureStats == nil {
		t.Fatal("Expected to find Creature stats")
	}

	if creatureStats.Count != 2 {
		t.Errorf("Expected 2 creatures, got %d", creatureStats.Count)
	}

	// Average GIHWR should be (55 + 50) / 2 = 52.5
	expectedAvg := 52.5
	if creatureStats.AvgGIHWR != expectedAvg {
		t.Errorf("Expected average GIHWR %v, got %v", expectedAvg, creatureStats.AvgGIHWR)
	}
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
			"1": {
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
			"1": {
				Name:   "White Card",
				Colors: []string{"W"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 60.0, GIH: 100},
				},
			},
			"2": {
				Name:   "Blue Card",
				Colors: []string{"U"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 55.0, GIH: 100},
				},
			},
			"3": {
				Name:   "Red Card",
				Colors: []string{"R"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 58.0, GIH: 100},
				},
			},
			"4": {
				Name:   "Gold Card",
				Colors: []string{"W", "U"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 65.0, GIH: 100},
				},
			},
			"5": {
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
			"1": {Colors: []string{"R"}, CMC: 1.0},
			"2": {Colors: []string{"R"}, CMC: 2.0},
			"3": {Colors: []string{"W"}, CMC: 2.0},
			"4": {Colors: []string{"W"}, CMC: 3.0},
		},
	}

	style := sg.determineArchetypeStyle([]string{"R", "W"}, setFile)

	// With low average CMC and R/W colors, should be Aggro
	if style != "Aggro" {
		t.Errorf("Expected Aggro style for low CMC RW deck, got %s", style)
	}
}

func TestGetTierListWithCardType(t *testing.T) {
	sg := NewSetGuide(nil, "")

	setFile := &seventeenlands.SetFile{
		CardRatings: map[string]*seventeenlands.CardRatingData{
			"1": {
				Name:   "Creature 1",
				Types:  []string{"Creature"},
				Colors: []string{"W"},
				Rarity: "common",
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 55.0, GIH: 100},
				},
			},
			"2": {
				Name:   "Instant 1",
				Types:  []string{"Instant"},
				Colors: []string{"R"},
				Rarity: "common",
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 53.0, GIH: 100},
				},
			},
		},
	}

	sg.setFiles = map[string]*seventeenlands.SetFile{
		"TEST": setFile,
	}

	// Filter by creature type
	creatures, err := sg.GetTierList("TEST", TierListOptions{
		CardType: "Creature",
	})
	if err != nil {
		t.Errorf("GetTierList returned error: %v", err)
	}

	if len(creatures) != 1 {
		t.Errorf("Expected 1 creature, got %d", len(creatures))
	}

	if creatures[0].CardType != "Creature" {
		t.Errorf("Expected CardType to be Creature, got %s", creatures[0].CardType)
	}

	// Filter by instant type
	instants, err := sg.GetTierList("TEST", TierListOptions{
		CardType: "Instant",
	})
	if err != nil {
		t.Errorf("GetTierList returned error: %v", err)
	}

	if len(instants) != 1 {
		t.Errorf("Expected 1 instant, got %d", len(instants))
	}

	if instants[0].CardType != "Instant" {
		t.Errorf("Expected CardType to be Instant, got %s", instants[0].CardType)
	}
}
