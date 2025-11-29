package prediction

import (
	"testing"
)

func TestCalculateSynergy_EmptyDeck(t *testing.T) {
	result := CalculateSynergy([]CardData{})
	if result.OverallScore != 0.5 {
		t.Errorf("Expected neutral score 0.5 for empty deck, got %f", result.OverallScore)
	}
}

func TestCalculateSynergy_SingleCard(t *testing.T) {
	cards := []CardData{
		{Name: "Llanowar Elves", Types: []string{"Elf", "Druid"}},
	}
	result := CalculateSynergy(cards)
	if result.OverallScore != 0.5 {
		t.Errorf("Expected neutral score 0.5 for single card, got %f", result.OverallScore)
	}
}

func TestCalculateSynergy_TribalSynergy(t *testing.T) {
	cards := []CardData{
		{Name: "Llanowar Elves", Types: []string{"Elf", "Druid"}},
		{Name: "Elvish Mystic", Types: []string{"Elf", "Druid"}},
		{Name: "Elvish Archdruid", Types: []string{"Elf", "Druid"}},
	}

	result := CalculateSynergy(cards)

	if result.TribalSynergies == 0 {
		t.Error("Expected tribal synergies for Elf deck")
	}

	if result.OverallScore <= 0.5 {
		t.Errorf("Expected above-neutral score for synergistic deck, got %f", result.OverallScore)
	}
}

func TestCalculateSynergy_MechanicalSynergy_Sacrifice(t *testing.T) {
	cards := []CardData{
		{
			Name:       "Viscera Seer",
			Keywords:   []string{"sacrifice"},
			OracleText: "Sacrifice a creature: Scry 1.",
		},
		{
			Name:       "Blood Artist",
			Keywords:   []string{"death trigger"},
			OracleText: "Whenever a creature dies, target opponent loses 1 life.",
		},
	}

	result := CalculateSynergy(cards)

	if result.MechSynergies == 0 {
		t.Error("Expected mechanical synergies for sacrifice deck")
	}
}

func TestCalculateSynergy_MechanicalSynergy_Tokens(t *testing.T) {
	cards := []CardData{
		{
			Name:       "Raise the Alarm",
			Keywords:   []string{"token"},
			OracleText: "Create two 1/1 white Soldier creature tokens.",
		},
		{
			Name:       "Glorious Anthem",
			Keywords:   []string{"anthem"},
			OracleText: "Creatures you control get +1/+1.",
		},
	}

	result := CalculateSynergy(cards)

	if result.MechSynergies == 0 {
		t.Error("Expected mechanical synergies for token + anthem")
	}
}

func TestCalculateSynergy_ColorSynergy(t *testing.T) {
	cards := []CardData{
		{Name: "Lightning Bolt", Color: "R"},
		{Name: "Shock", Color: "R"},
		{Name: "Searing Spear", Color: "R"},
	}

	result := CalculateSynergy(cards)

	if result.ColorSynergies == 0 {
		t.Error("Expected color synergies for mono-red deck")
	}
}

func TestCalculateSynergy_NoSynergy(t *testing.T) {
	cards := []CardData{
		{Name: "Giant Growth", Color: "G", Types: []string{}},
		{Name: "Lightning Bolt", Color: "R", Types: []string{}},
		{Name: "Doom Blade", Color: "B", Types: []string{}},
	}

	result := CalculateSynergy(cards)

	// Should have minimal synergies (no same type, no mechanical synergy)
	if result.TribalSynergies > 0 {
		t.Errorf("Expected no tribal synergies, got %d", result.TribalSynergies)
	}

	if result.ColorSynergies > 0 {
		t.Errorf("Expected no color synergies, got %d", result.ColorSynergies)
	}
}

func TestCalculateSynergy_SpellSynergy(t *testing.T) {
	cards := []CardData{
		{
			Name:    "Lightning Bolt",
			IsSpell: true,
		},
		{
			Name:       "Monastery Swiftspear",
			Keywords:   []string{"magecraft"},
			OracleText: "Whenever you cast or copy an instant or sorcery spell, this creature gets +1/+0.",
		},
	}

	result := CalculateSynergy(cards)

	if result.MechSynergies == 0 {
		t.Error("Expected mechanical synergy between spell and magecraft creature")
	}
}

func TestCheckTribalSynergy_SharedType(t *testing.T) {
	cardA := CardData{Name: "Elf Warrior", Types: []string{"Elf", "Warrior"}}
	cardB := CardData{Name: "Elvish Visionary", Types: []string{"Elf", "Wizard"}}

	synergy := checkTribalSynergy(cardA, cardB)

	if synergy == nil {
		t.Error("Expected tribal synergy for shared Elf type")
	}

	if synergy != nil && synergy.SynergyType != SynergyTribal {
		t.Errorf("Expected tribal synergy type, got %s", synergy.SynergyType)
	}
}

func TestCheckTribalSynergy_NoSharedType(t *testing.T) {
	cardA := CardData{Name: "Goblin Guide", Types: []string{"Goblin", "Scout"}}
	cardB := CardData{Name: "Elvish Visionary", Types: []string{"Elf", "Wizard"}}

	synergy := checkTribalSynergy(cardA, cardB)

	if synergy != nil {
		t.Error("Expected no tribal synergy for different types")
	}
}

func TestCheckColorSynergy_SameColor(t *testing.T) {
	cardA := CardData{Name: "Lightning Bolt", Color: "R"}
	cardB := CardData{Name: "Shock", Color: "R"}

	synergy := checkColorSynergy(cardA, cardB)

	if synergy == nil {
		t.Error("Expected color synergy for same-color cards")
	}

	if synergy != nil && synergy.SynergyType != SynergyColor {
		t.Errorf("Expected color synergy type, got %s", synergy.SynergyType)
	}
}

func TestCheckColorSynergy_DifferentColor(t *testing.T) {
	cardA := CardData{Name: "Lightning Bolt", Color: "R"}
	cardB := CardData{Name: "Giant Growth", Color: "G"}

	synergy := checkColorSynergy(cardA, cardB)

	if synergy != nil {
		t.Error("Expected no color synergy for different-color cards")
	}
}

func TestCheckColorSynergy_Colorless(t *testing.T) {
	cardA := CardData{Name: "Sol Ring", Color: "C"}
	cardB := CardData{Name: "Mana Crypt", Color: "C"}

	synergy := checkColorSynergy(cardA, cardB)

	// Colorless cards don't get color synergy
	if synergy != nil {
		t.Error("Expected no color synergy for colorless cards")
	}
}

func TestHasMechanic(t *testing.T) {
	card := CardData{
		Name:       "Viscera Seer",
		Keywords:   []string{"sacrifice"},
		OracleText: "Sacrifice a creature: Scry 1.",
	}

	if !hasMechanic(card, "sacrifice") {
		t.Error("Expected card to have sacrifice mechanic via keyword")
	}
}

func TestHasKeyword(t *testing.T) {
	card := CardData{
		Name:       "Blood Artist",
		OracleText: "Whenever a creature dies, target opponent loses 1 life.",
	}

	if !hasKeyword(card, "dies") {
		t.Error("Expected card to have 'dies' keyword in oracle text")
	}
}

func TestIsRelevantCreatureType(t *testing.T) {
	relevantTypes := []string{"Elf", "Goblin", "Vampire", "Zombie", "Human", "Dragon"}
	irrelevantTypes := []string{"Coward", "Survivor", "Citizen"}

	for _, ct := range relevantTypes {
		if !isRelevantCreatureType(ct) {
			t.Errorf("Expected %s to be a relevant creature type", ct)
		}
	}

	for _, ct := range irrelevantTypes {
		if isRelevantCreatureType(ct) {
			t.Errorf("Expected %s to NOT be a relevant creature type", ct)
		}
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		haystack string
		needle   string
		expected bool
	}{
		{"Hello World", "world", true},
		{"Hello World", "WORLD", true},
		{"Hello World", "xyz", false},
		{"sacrifice", "Sacrifice", true},
		{"", "test", false},
	}

	for _, tt := range tests {
		got := containsIgnoreCase(tt.haystack, tt.needle)
		if got != tt.expected {
			t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v",
				tt.haystack, tt.needle, got, tt.expected)
		}
	}
}

func TestColorName(t *testing.T) {
	tests := []struct {
		abbrev   string
		expected string
	}{
		{"W", "White"},
		{"U", "Blue"},
		{"B", "Black"},
		{"R", "Red"},
		{"G", "Green"},
		{"C", "Colorless"},
		{"X", "X"}, // Unknown color
	}

	for _, tt := range tests {
		got := colorName(tt.abbrev)
		if got != tt.expected {
			t.Errorf("colorName(%q) = %q, want %q", tt.abbrev, got, tt.expected)
		}
	}
}

func TestGetTopSynergyReasons(t *testing.T) {
	synergies := []SynergyScore{
		{Reason: "Low weight", Weight: 0.05},
		{Reason: "High weight", Weight: 0.15},
		{Reason: "Medium weight", Weight: 0.10},
	}

	reasons := getTopSynergyReasons(synergies, 2)

	if len(reasons) != 2 {
		t.Errorf("Expected 2 reasons, got %d", len(reasons))
	}

	if reasons[0] != "High weight" {
		t.Errorf("Expected 'High weight' first, got %s", reasons[0])
	}
}

func TestConvertCardsToCardData(t *testing.T) {
	cards := []Card{
		{Name: "Test Card", CMC: 3, Color: "R", GIHWR: 0.55, Rarity: "common"},
	}

	result := ConvertCardsToCardData(cards)

	if len(result) != 1 {
		t.Fatalf("Expected 1 card, got %d", len(result))
	}

	if result[0].Name != "Test Card" {
		t.Errorf("Expected name 'Test Card', got %s", result[0].Name)
	}

	if result[0].CMC != 3 {
		t.Errorf("Expected CMC 3, got %d", result[0].CMC)
	}
}

func TestMinMax(t *testing.T) {
	if min(3.0, 5.0) != 3.0 {
		t.Error("min(3.0, 5.0) should be 3.0")
	}

	if min(5.0, 3.0) != 3.0 {
		t.Error("min(5.0, 3.0) should be 3.0")
	}

	if max(3.0, 5.0) != 5.0 {
		t.Error("max(3.0, 5.0) should be 5.0")
	}

	if max(5.0, 3.0) != 5.0 {
		t.Error("max(5.0, 3.0) should be 5.0")
	}
}

func TestCalculateSynergy_SynergyScoreRange(t *testing.T) {
	// Test with highly synergistic deck
	cards := []CardData{
		{Name: "Elf 1", Types: []string{"Elf"}, Color: "G"},
		{Name: "Elf 2", Types: []string{"Elf"}, Color: "G"},
		{Name: "Elf 3", Types: []string{"Elf"}, Color: "G"},
		{Name: "Elf 4", Types: []string{"Elf"}, Color: "G"},
		{Name: "Elf 5", Types: []string{"Elf"}, Color: "G"},
	}

	result := CalculateSynergy(cards)

	if result.OverallScore < 0.0 || result.OverallScore > 1.0 {
		t.Errorf("Synergy score out of range: %f", result.OverallScore)
	}

	if result.OverallScore <= 0.5 {
		t.Errorf("Expected above-neutral score for highly synergistic deck, got %f", result.OverallScore)
	}
}
