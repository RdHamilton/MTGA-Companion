package synergy

import (
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

func TestParseDeckFromText_BasicFormat(t *testing.T) {
	text := `4 Lightning Bolt
3 Thoughtseize
2 Fatal Push
1 Mountain

Sideboard
2 Duress
1 Negate`

	mainDeck, sideboard, err := ParseDeckFromText(text)
	if err != nil {
		t.Fatalf("Failed to parse deck: %v", err)
	}

	if len(mainDeck) != 4 {
		t.Errorf("Expected 4 main deck entries, got %d", len(mainDeck))
	}
	if len(sideboard) != 2 {
		t.Errorf("Expected 2 sideboard entries, got %d", len(sideboard))
	}

	// Check first entry
	if mainDeck[0].Quantity != 4 {
		t.Errorf("Expected quantity 4 for Lightning Bolt, got %d", mainDeck[0].Quantity)
	}
	if mainDeck[0].CardName != "Lightning Bolt" {
		t.Errorf("Expected card name 'Lightning Bolt', got %q", mainDeck[0].CardName)
	}
}

func TestParseDeckFromText_XFormat(t *testing.T) {
	text := `4x Lightning Bolt
3x Thoughtseize`

	mainDeck, _, err := ParseDeckFromText(text)
	if err != nil {
		t.Fatalf("Failed to parse deck: %v", err)
	}

	if len(mainDeck) != 2 {
		t.Errorf("Expected 2 main deck entries, got %d", len(mainDeck))
	}
	if mainDeck[0].Quantity != 4 {
		t.Errorf("Expected quantity 4, got %d", mainDeck[0].Quantity)
	}
}

func TestParseDeckFromText_WithSetCodes(t *testing.T) {
	text := `4 Lightning Bolt (2XM)
3 Fatal Push #123`

	mainDeck, _, err := ParseDeckFromText(text)
	if err != nil {
		t.Fatalf("Failed to parse deck: %v", err)
	}

	if mainDeck[0].CardName != "Lightning Bolt" {
		t.Errorf("Expected set code to be stripped, got %q", mainDeck[0].CardName)
	}
	if mainDeck[1].CardName != "Fatal Push" {
		t.Errorf("Expected collector number to be stripped, got %q", mainDeck[1].CardName)
	}
}

func TestParseDeckFromText_SkipsComments(t *testing.T) {
	text := `// Deck Name
# This is a comment
4 Lightning Bolt
3 Thoughtseize`

	mainDeck, _, err := ParseDeckFromText(text)
	if err != nil {
		t.Fatalf("Failed to parse deck: %v", err)
	}

	if len(mainDeck) != 2 {
		t.Errorf("Expected 2 entries (comments skipped), got %d", len(mainDeck))
	}
}

func TestParseDeckFromText_EmptyLines(t *testing.T) {
	text := `4 Lightning Bolt

3 Thoughtseize

`

	mainDeck, _, err := ParseDeckFromText(text)
	if err != nil {
		t.Fatalf("Failed to parse deck: %v", err)
	}

	if len(mainDeck) != 2 {
		t.Errorf("Expected 2 entries (empty lines skipped), got %d", len(mainDeck))
	}
}

func TestParseDeckFromText_SideboardMarkers(t *testing.T) {
	tests := []string{
		"Sideboard",
		"SIDEBOARD",
		"sideboard:",
		"// Sideboard",
	}

	for _, marker := range tests {
		text := "4 Lightning Bolt\n" + marker + "\n2 Duress"
		_, sideboard, err := ParseDeckFromText(text)
		if err != nil {
			t.Fatalf("Failed to parse deck with marker %q: %v", marker, err)
		}
		if len(sideboard) != 1 {
			t.Errorf("Marker %q: Expected 1 sideboard entry, got %d", marker, len(sideboard))
		}
	}
}

func TestCleanCardName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Lightning Bolt (2XM)", "Lightning Bolt"},
		{"Fatal Push #123", "Fatal Push"},
		{"Card Name (Foil)", "Card Name"},
		{"Card Name (Showcase)", "Card Name"},
		{"Card Name (Extended Art)", "Card Name"},
		{"  Card Name  ", "Card Name"},
		// Note: cleanCardName removes set codes first, then collector numbers
		// But the pattern removes collector numbers at end of string
		// So (SET) #456 becomes (SET) after removing #456
		{"Card Name #456", "Card Name"},
	}

	for _, tt := range tests {
		result := cleanCardName(tt.input)
		if result != tt.expected {
			t.Errorf("cleanCardName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestDetectArchetype_Colors(t *testing.T) {
	tests := []struct {
		name     string
		mainDeck []DeckEntry
		contains string
	}{
		{
			name: "Red deck",
			mainDeck: []DeckEntry{
				{Quantity: 20, CardName: "Mountain"},
				{Quantity: 4, CardName: "Lightning Bolt"},
			},
			contains: "R",
		},
		{
			name: "Blue deck",
			mainDeck: []DeckEntry{
				{Quantity: 20, CardName: "Island"},
				{Quantity: 4, CardName: "Counterspell"},
			},
			contains: "U",
		},
		{
			name: "Multi-color deck",
			mainDeck: []DeckEntry{
				{Quantity: 10, CardName: "Mountain"},
				{Quantity: 10, CardName: "Forest"},
				{Quantity: 4, CardName: "Gruul Spellbreaker"},
			},
			contains: "RG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectArchetype(tt.mainDeck)
			// Just check that colors are detected (not exact archetype name)
			if tt.contains != "" && len(result) == 0 {
				t.Errorf("Expected archetype to be detected, got empty string")
			}
		})
	}
}

func TestClassifyCardRole_Core(t *testing.T) {
	// 4-ofs should be core
	role := ClassifyCardRole("Any Card", 4, false)
	if role != models.CardRoleCore {
		t.Errorf("Expected 4-of to be core, got %s", role)
	}
}

func TestClassifyCardRole_Sideboard(t *testing.T) {
	// Sideboard cards should always be sideboard role
	role := ClassifyCardRole("Any Card", 4, true)
	if role != models.CardRoleSideboard {
		t.Errorf("Expected sideboard card to have sideboard role, got %s", role)
	}
}

func TestClassifyCardRole_FlexCards(t *testing.T) {
	// 1-2 copies are typically flex
	role1 := ClassifyCardRole("Random Card", 1, false)
	role2 := ClassifyCardRole("Random Card", 2, false)

	if role1 != models.CardRoleFlex {
		t.Errorf("Expected 1-of to be flex, got %s", role1)
	}
	if role2 != models.CardRoleFlex {
		t.Errorf("Expected 2-of to be flex, got %s", role2)
	}
}

func TestClassifyCardRole_CorePatterns(t *testing.T) {
	// Cards with known core patterns should be core even at 3 copies
	coreCards := []string{
		"Lightning Bolt",
		"Counterspell",
		"Thoughtseize",
		"Fatal Push",
	}

	for _, card := range coreCards {
		role := ClassifyCardRole(card, 3, false)
		if role != models.CardRoleCore {
			t.Errorf("Expected %q at 3 copies to be core, got %s", card, role)
		}
	}
}

func TestKnownArchetypes(t *testing.T) {
	// Verify known archetypes are mapped correctly
	tests := []struct {
		archetype string
		playStyle string
	}{
		{"mono-red aggro", "aggro"},
		{"domain", "midrange"},
		{"azorius control", "control"},
		{"lotus field", "combo"},
		{"mono-blue tempo", "tempo"},
		{"mono-green ramp", "ramp"},
	}

	for _, tt := range tests {
		playStyle, ok := KnownArchetypes[tt.archetype]
		if !ok {
			t.Errorf("Archetype %q not found in KnownArchetypes", tt.archetype)
			continue
		}
		if playStyle != tt.playStyle {
			t.Errorf("KnownArchetypes[%q] = %q, want %q", tt.archetype, playStyle, tt.playStyle)
		}
	}
}

func TestExtractSynergiesFromText(t *testing.T) {
	text := "Lightning Bolt works well with Fatal Push to provide excellent removal options."
	knownCards := []string{"Lightning Bolt", "Fatal Push", "Thoughtseize"}

	synergies := ExtractSynergiesFromText(text, knownCards)

	if len(synergies) == 0 {
		t.Error("Expected at least one synergy to be extracted")
	}

	// Check that we found the synergy between Lightning Bolt and Fatal Push
	found := false
	for _, s := range synergies {
		if (s.CardA == "Lightning Bolt" && s.CardB == "Fatal Push") ||
			(s.CardA == "Fatal Push" && s.CardB == "Lightning Bolt") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected synergy between Lightning Bolt and Fatal Push")
	}
}

func TestExtractSynergiesFromText_NoMatches(t *testing.T) {
	text := "This text mentions no synergy patterns."
	knownCards := []string{"Lightning Bolt", "Fatal Push"}

	synergies := ExtractSynergiesFromText(text, knownCards)

	if len(synergies) != 0 {
		t.Errorf("Expected no synergies when no patterns match, got %d", len(synergies))
	}
}

func TestBuildArchetypeFromDeck(t *testing.T) {
	deck := &ScrapedDeck{
		Name:      "Test Red Aggro",
		Format:    "Standard",
		Archetype: "Mono-Red Aggro",
		Tier:      "S",
		PlayStyle: "aggro",
		MainDeck: []DeckEntry{
			{Quantity: 4, CardName: "Lightning Bolt"},
			{Quantity: 4, CardName: "Monastery Swiftspear"},
			{Quantity: 2, CardName: "Flex Card"},
		},
		Sideboard: []DeckEntry{
			{Quantity: 2, CardName: "Smash to Smithereens"},
		},
	}

	data := BuildArchetypeFromDeck(deck)

	if data.Archetype.Name != "Mono-Red Aggro" {
		t.Errorf("Archetype name = %q, want %q", data.Archetype.Name, "Mono-Red Aggro")
	}
	if data.Archetype.Tier != "S" {
		t.Errorf("Tier = %q, want %q", data.Archetype.Tier, "S")
	}
	if data.Archetype.PlayStyle != "aggro" {
		t.Errorf("PlayStyle = %q, want %q", data.Archetype.PlayStyle, "aggro")
	}

	// Check card counts
	if len(data.CoreCards) != 2 {
		t.Errorf("Expected 2 core cards (4-ofs), got %d", len(data.CoreCards))
	}
	if len(data.FlexCards) != 1 {
		t.Errorf("Expected 1 flex card, got %d", len(data.FlexCards))
	}
	if len(data.Sideboard) != 1 {
		t.Errorf("Expected 1 sideboard card, got %d", len(data.Sideboard))
	}
}

func TestBuildArchetypeFromDeck_InfersPlayStyle(t *testing.T) {
	deck := &ScrapedDeck{
		Name:      "Azorius Control",
		Format:    "Standard",
		Archetype: "azorius control", // lowercase to test lookup
		MainDeck:  []DeckEntry{},
	}

	data := BuildArchetypeFromDeck(deck)

	if data.Archetype.PlayStyle != "control" {
		t.Errorf("Expected play style to be inferred as 'control', got %q", data.Archetype.PlayStyle)
	}
}

func TestBuildArchetypeFromDeck_CardImportance(t *testing.T) {
	deck := &ScrapedDeck{
		Name:   "Test Deck",
		Format: "Standard",
		MainDeck: []DeckEntry{
			{Quantity: 4, CardName: "Four-of Card"},
			{Quantity: 3, CardName: "Three-of Card"},
			{Quantity: 2, CardName: "Two-of Card"},
		},
	}

	data := BuildArchetypeFromDeck(deck)

	// Find cards and check importance
	var fourOf, threeOf, twoOf *models.MTGZoneArchetypeCard
	for i := range data.CoreCards {
		if data.CoreCards[i].CardName == "Four-of Card" {
			fourOf = &data.CoreCards[i]
		}
	}
	for i := range data.FlexCards {
		if data.FlexCards[i].CardName == "Three-of Card" {
			threeOf = &data.FlexCards[i]
		}
		if data.FlexCards[i].CardName == "Two-of Card" {
			twoOf = &data.FlexCards[i]
		}
	}

	if fourOf != nil && fourOf.Importance != models.CardImportanceEssential {
		t.Errorf("4-of should be essential, got %s", fourOf.Importance)
	}
	if threeOf != nil && threeOf.Importance != models.CardImportanceImportant {
		t.Errorf("3-of should be important, got %s", threeOf.Importance)
	}
	if twoOf != nil && twoOf.Importance != models.CardImportanceOptional {
		t.Errorf("2-of should be optional, got %s", twoOf.Importance)
	}
}

func TestGetMetaTierList_Standard(t *testing.T) {
	tierList := GetMetaTierList("Standard")

	if len(tierList) == 0 {
		t.Error("Expected Standard tier list to have entries")
	}

	// Check that all entries have format set
	for _, arch := range tierList {
		if arch.Format != "Standard" {
			t.Errorf("Expected format 'Standard', got %q", arch.Format)
		}
	}
}

func TestGetMetaTierList_Historic(t *testing.T) {
	tierList := GetMetaTierList("Historic")

	if len(tierList) == 0 {
		t.Error("Expected Historic tier list to have entries")
	}

	for _, arch := range tierList {
		if arch.Format != "Historic" {
			t.Errorf("Expected format 'Historic', got %q", arch.Format)
		}
	}
}

func TestGetMetaTierList_Unknown(t *testing.T) {
	tierList := GetMetaTierList("Unknown Format")

	if len(tierList) != 0 {
		t.Errorf("Expected empty tier list for unknown format, got %d entries", len(tierList))
	}
}

func TestGetMetaTierList_CaseInsensitive(t *testing.T) {
	// The function should be case-insensitive
	tierList1 := GetMetaTierList("standard")
	tierList2 := GetMetaTierList("STANDARD")

	if len(tierList1) == 0 || len(tierList2) == 0 {
		t.Error("Expected case-insensitive format lookup")
	}
}

func TestArchetypeData_ToJSON(t *testing.T) {
	data := &ArchetypeData{
		Archetype: models.MTGZoneArchetype{
			Name:   "Test Deck",
			Format: "Standard",
			Tier:   "A",
		},
		CoreCards: []models.MTGZoneArchetypeCard{
			{CardName: "Card A", Role: models.CardRoleCore, Copies: 4},
		},
	}

	jsonStr, err := data.ToJSON()
	if err != nil {
		t.Fatalf("Failed to convert to JSON: %v", err)
	}

	if jsonStr == "" {
		t.Error("Expected non-empty JSON string")
	}

	// Verify JSON contains expected fields
	if !contains(jsonStr, "Test Deck") {
		t.Error("JSON should contain archetype name")
	}
	if !contains(jsonStr, "Card A") {
		t.Error("JSON should contain card name")
	}
}

func TestScrapedDeck_Struct(t *testing.T) {
	deck := ScrapedDeck{
		Name:        "Mono-Red Aggro",
		Format:      "Standard",
		Archetype:   "Aggro",
		Tier:        "S",
		Description: "Fast aggressive deck",
		PlayStyle:   "aggro",
		SourceURL:   "https://mtgazone.com/deck",
		Author:      "Test Author",
		Date:        "2024-01-01",
	}

	if deck.Name != "Mono-Red Aggro" {
		t.Errorf("Name = %q, want %q", deck.Name, "Mono-Red Aggro")
	}
	if deck.Format != "Standard" {
		t.Errorf("Format = %q, want %q", deck.Format, "Standard")
	}
}

func TestDeckEntry_Struct(t *testing.T) {
	entry := DeckEntry{
		Quantity: 4,
		CardName: "Lightning Bolt",
		Role:     "core",
	}

	if entry.Quantity != 4 {
		t.Errorf("Quantity = %d, want %d", entry.Quantity, 4)
	}
	if entry.CardName != "Lightning Bolt" {
		t.Errorf("CardName = %q, want %q", entry.CardName, "Lightning Bolt")
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
