package recommendations

import (
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
)

func TestClassifyDeck_Aggro(t *testing.T) {
	hasteText := "Haste"
	burnText := "Lightning Bolt deals 3 damage to any target."

	// Build an aggro deck: low curve creatures + burn
	var deckCards []*cards.Card
	for i := 0; i < 12; i++ {
		deckCards = append(deckCards, &cards.Card{
			ArenaID:    i,
			Name:       "Cheap Creature",
			TypeLine:   "Creature — Goblin",
			CMC:        1,
			OracleText: &hasteText,
		})
	}
	for i := 0; i < 8; i++ {
		deckCards = append(deckCards, &cards.Card{
			ArenaID:    100 + i,
			Name:       "Burn Spell",
			TypeLine:   "Instant",
			CMC:        1,
			OracleText: &burnText,
		})
	}

	scores := ClassifyDeck(deckCards)

	if len(scores) == 0 {
		t.Fatal("Expected archetype scores, got none")
	}

	// Aggro should be top or near top
	foundAggro := false
	for i, score := range scores {
		if score.Archetype == ArchetypeAggro {
			foundAggro = true
			if i > 1 {
				t.Errorf("Expected Aggro to be top archetype, but it was #%d", i+1)
			}
			if score.Score < 0.5 {
				t.Errorf("Expected Aggro score >= 0.5, got %.2f", score.Score)
			}
			break
		}
	}

	if !foundAggro {
		t.Error("Expected to find Aggro archetype in scores")
	}
}

func TestClassifyDeck_Control(t *testing.T) {
	counterText := "Counter target spell."
	removeText := "Destroy target creature."
	drawText := "Draw two cards."
	wipeText := "Destroy all creatures."

	// Build a control deck: counters + removal + card draw + few creatures
	var deckCards []*cards.Card
	for i := 0; i < 6; i++ {
		deckCards = append(deckCards, &cards.Card{
			ArenaID:    i,
			Name:       "Counterspell",
			TypeLine:   "Instant",
			CMC:        2,
			OracleText: &counterText,
		})
	}
	for i := 0; i < 6; i++ {
		deckCards = append(deckCards, &cards.Card{
			ArenaID:    100 + i,
			Name:       "Removal",
			TypeLine:   "Instant",
			CMC:        3,
			OracleText: &removeText,
		})
	}
	for i := 0; i < 4; i++ {
		deckCards = append(deckCards, &cards.Card{
			ArenaID:    200 + i,
			Name:       "Draw Spell",
			TypeLine:   "Sorcery",
			CMC:        4,
			OracleText: &drawText,
		})
	}
	for i := 0; i < 3; i++ {
		deckCards = append(deckCards, &cards.Card{
			ArenaID:    300 + i,
			Name:       "Board Wipe",
			TypeLine:   "Sorcery",
			CMC:        5,
			OracleText: &wipeText,
		})
	}
	// Only a few creatures
	for i := 0; i < 4; i++ {
		creatureText := "Flying"
		deckCards = append(deckCards, &cards.Card{
			ArenaID:    400 + i,
			Name:       "Win Condition",
			TypeLine:   "Creature — Dragon",
			CMC:        6,
			OracleText: &creatureText,
		})
	}

	scores := ClassifyDeck(deckCards)

	if len(scores) == 0 {
		t.Fatal("Expected archetype scores, got none")
	}

	// Control should be top or near top
	foundControl := false
	for i, score := range scores {
		if score.Archetype == ArchetypeControl {
			foundControl = true
			if i > 1 {
				t.Errorf("Expected Control to be top archetype, but it was #%d", i+1)
			}
			break
		}
	}

	if !foundControl {
		t.Error("Expected to find Control archetype in scores")
	}
}

func TestClassifyDeck_Tokens(t *testing.T) {
	tokenText := "Create two 1/1 white Soldier creature tokens."
	anthemText := "Other creatures you control get +1/+1."
	payoffText := "Whenever a creature enters the battlefield, draw a card."

	// Build a tokens deck
	var deckCards []*cards.Card
	for i := 0; i < 10; i++ {
		deckCards = append(deckCards, &cards.Card{
			ArenaID:    i,
			Name:       "Token Maker",
			TypeLine:   "Creature — Human Soldier",
			CMC:        2,
			OracleText: &tokenText,
		})
	}
	for i := 0; i < 4; i++ {
		deckCards = append(deckCards, &cards.Card{
			ArenaID:    100 + i,
			Name:       "Anthem",
			TypeLine:   "Enchantment",
			CMC:        3,
			OracleText: &anthemText,
		})
	}
	for i := 0; i < 3; i++ {
		deckCards = append(deckCards, &cards.Card{
			ArenaID:    200 + i,
			Name:       "Payoff",
			TypeLine:   "Creature — Human",
			CMC:        4,
			OracleText: &payoffText,
		})
	}

	scores := ClassifyDeck(deckCards)

	// Tokens should appear in scores
	foundTokens := false
	for _, score := range scores {
		if score.Archetype == ArchetypeTokens {
			foundTokens = true
			if score.Score < 0.3 {
				t.Errorf("Expected Tokens score >= 0.3, got %.2f", score.Score)
			}
			break
		}
	}

	if !foundTokens {
		t.Error("Expected to find Tokens archetype in scores")
	}
}

func TestClassifyDeck_Ramp(t *testing.T) {
	dorkText := "Tap: Add {G}."
	searchText := "Search your library for a basic land card and put it onto the battlefield."
	bigText := "Flying, trample"

	// Build a ramp deck
	var deckCards []*cards.Card
	for i := 0; i < 8; i++ {
		deckCards = append(deckCards, &cards.Card{
			ArenaID:    i,
			Name:       "Mana Dork",
			TypeLine:   "Creature — Elf Druid",
			CMC:        1,
			OracleText: &dorkText,
		})
	}
	for i := 0; i < 4; i++ {
		deckCards = append(deckCards, &cards.Card{
			ArenaID:    100 + i,
			Name:       "Rampant Growth",
			TypeLine:   "Sorcery",
			CMC:        2,
			OracleText: &searchText,
		})
	}
	for i := 0; i < 6; i++ {
		deckCards = append(deckCards, &cards.Card{
			ArenaID:    200 + i,
			Name:       "Big Finisher",
			TypeLine:   "Creature — Dragon",
			CMC:        7,
			OracleText: &bigText,
		})
	}

	scores := ClassifyDeck(deckCards)

	// Ramp should appear in scores
	foundRamp := false
	for _, score := range scores {
		if score.Archetype == ArchetypeRamp {
			foundRamp = true
			break
		}
	}

	if !foundRamp {
		t.Error("Expected to find Ramp archetype in scores")
	}
}

func TestClassifyDeck_EmptyDeck(t *testing.T) {
	scores := ClassifyDeck([]*cards.Card{})

	if scores != nil && len(scores) > 0 {
		t.Error("Expected no scores for empty deck")
	}
}

func TestGetPrimaryArchetype(t *testing.T) {
	burnText := "Lightning Bolt deals 3 damage to any target."

	var deckCards []*cards.Card
	for i := 0; i < 20; i++ {
		deckCards = append(deckCards, &cards.Card{
			ArenaID:    i,
			Name:       "Burn Spell",
			TypeLine:   "Instant",
			CMC:        1,
			OracleText: &burnText,
		})
	}

	primary := GetPrimaryArchetype(deckCards)

	if primary == nil {
		t.Fatal("Expected primary archetype, got nil")
	}

	if primary.Score < 0.2 {
		t.Errorf("Expected primary archetype score >= 0.2, got %.2f", primary.Score)
	}
}

func TestGetPrimaryArchetype_EmptyDeck(t *testing.T) {
	primary := GetPrimaryArchetype([]*cards.Card{})

	if primary != nil {
		t.Error("Expected nil for empty deck")
	}
}

func TestArchetypeSignals_AllHavePatterns(t *testing.T) {
	for archetype, signals := range archetypeSignals {
		if len(signals) == 0 {
			t.Errorf("Archetype %q has no signals", archetype)
		}

		for _, signal := range signals {
			if signal.Name == "" {
				t.Errorf("Archetype %q has signal with empty name", archetype)
			}
			if signal.Weight <= 0 {
				t.Errorf("Archetype %q signal %q has invalid weight: %.2f", archetype, signal.Name, signal.Weight)
			}
		}
	}
}

func TestArchetypeDescriptions_AllArchetypesHaveDescriptions(t *testing.T) {
	archetypes := []Archetype{
		ArchetypeAggro,
		ArchetypeControl,
		ArchetypeMidrange,
		ArchetypeCombo,
		ArchetypeTempo,
		ArchetypeRamp,
		ArchetypeTribal,
		ArchetypeTokens,
		ArchetypeArtifacts,
	}

	for _, arch := range archetypes {
		desc := GetArchetypeDescription(arch)
		if desc == "" || desc == "Unknown archetype" {
			t.Errorf("Archetype %q has no description", arch)
		}
	}
}

func TestIsAggroArchetype(t *testing.T) {
	tests := []struct {
		archetype Archetype
		want      bool
	}{
		{ArchetypeAggro, true},
		{ArchetypeTempo, true},
		{ArchetypeControl, false},
		{ArchetypeMidrange, false},
	}

	for _, tt := range tests {
		got := IsAggroArchetype(tt.archetype)
		if got != tt.want {
			t.Errorf("IsAggroArchetype(%q) = %v, want %v", tt.archetype, got, tt.want)
		}
	}
}

func TestIsControlArchetype(t *testing.T) {
	tests := []struct {
		archetype Archetype
		want      bool
	}{
		{ArchetypeControl, true},
		{ArchetypeAggro, false},
		{ArchetypeMidrange, false},
	}

	for _, tt := range tests {
		got := IsControlArchetype(tt.archetype)
		if got != tt.want {
			t.Errorf("IsControlArchetype(%q) = %v, want %v", tt.archetype, got, tt.want)
		}
	}
}

func TestClassifyDeck_ReturnsSignals(t *testing.T) {
	counterText := "Counter target spell."

	var deckCards []*cards.Card
	for i := 0; i < 8; i++ {
		deckCards = append(deckCards, &cards.Card{
			ArenaID:    i,
			Name:       "Counterspell",
			TypeLine:   "Instant",
			CMC:        2,
			OracleText: &counterText,
		})
	}

	scores := ClassifyDeck(deckCards)

	for _, score := range scores {
		if score.Archetype == ArchetypeControl || score.Archetype == ArchetypeTempo {
			if len(score.Signals) == 0 {
				t.Errorf("Expected signals for %q archetype", score.Archetype)
			}
		}
	}
}
