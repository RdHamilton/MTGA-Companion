package prediction

import (
	"strings"
)

// SynergyType represents the type of synergy between cards.
type SynergyType string

const (
	SynergyTribal     SynergyType = "tribal"
	SynergyMechanical SynergyType = "mechanical"
	SynergyColor      SynergyType = "color"
	SynergyArchetype  SynergyType = "archetype"
)

// SynergyScore represents a synergy connection between two cards.
type SynergyScore struct {
	CardA       string      `json:"card_a"`
	CardB       string      `json:"card_b"`
	SynergyType SynergyType `json:"synergy_type"`
	Score       float64     `json:"score"`  // 0.0 to 1.0
	Reason      string      `json:"reason"` // Human-readable explanation
	Weight      float64     `json:"weight"` // Contribution to overall synergy
}

// SynergyResult contains the complete synergy analysis for a deck.
type SynergyResult struct {
	OverallScore    float64        `json:"overall_score"`    // 0.0 to 1.0
	SynergyPairs    []SynergyScore `json:"synergy_pairs"`    // Individual synergy pairs
	TribalSynergies int            `json:"tribal_synergies"` // Count of tribal synergies
	MechSynergies   int            `json:"mech_synergies"`   // Count of mechanical synergies
	ColorSynergies  int            `json:"color_synergies"`  // Count of color synergies
	TopSynergies    []string       `json:"top_synergies"`    // Top synergy explanations
}

// CardData contains expanded card information for synergy analysis.
type CardData struct {
	Name          string
	CMC           int
	Color         string
	GIHWR         float64
	Rarity        string
	Types         []string // Creature types: "Elf", "Goblin", "Vampire", etc.
	Keywords      []string // Card keywords: "flying", "sacrifice", "token", etc.
	OracleText    string   // Full card text for pattern matching
	IsCreature    bool
	IsSpell       bool
	IsArtifact    bool
	IsEnchantment bool
}

// tribalLords maps creature types to cards that benefit that type.
var tribalLords = map[string][]string{
	"elf":       {"elvish archdruid", "elvish warmaster", "dwynen", "imperious perfect"},
	"goblin":    {"goblin chieftain", "goblin warchief", "krenko", "goblin trashmaster"},
	"vampire":   {"bloodline keeper", "vampire nocturnus", "captivating vampire"},
	"zombie":    {"lord of the accursed", "death baron", "undead warchief"},
	"human":     {"thalia's lieutenant", "champion of the parish", "mayor of avabruck"},
	"warrior":   {"blood-chin fanatic", "chief of the edge", "blood-chin rager"},
	"merfolk":   {"master of the pearl trident", "lord of atlantis", "merfolk mistbinder"},
	"wizard":    {"naban", "patron wizard", "sage of fables"},
	"knight":    {"knight exemplar", "acclaimed contender", "inspiring veteran"},
	"spirit":    {"drogskol captain", "supreme phantom", "rattlechains"},
	"elemental": {"risen reef", "omnath", "creeping trailblazer"},
	"cleric":    {"righteous valkyrie", "orah", "speaker of the heavens"},
	"rogue":     {"soaring thought-thief", "thieves' guild enforcer", "robber of the rich"},
	"angel":     {"resplendent angel", "lyra dawnbringer", "angel of vitality"},
	"dragon":    {"utvara hellkite", "lathliss", "dragon tempest"},
}

// mechanicalSynergies maps mechanics to synergistic keywords/patterns.
var mechanicalSynergies = map[string][]string{
	"sacrifice":    {"death trigger", "blood artist", "mayhem devil", "cruel celebrant", "creates token", "dies"},
	"tokens":       {"populate", "anthem", "goes wide", "+1/+1 counters", "convoke", "sacrifice"},
	"counters":     {"proliferate", "modular", "evolve", "mentor", "adapt"},
	"graveyard":    {"escape", "flashback", "surveil", "delve", "self-mill", "reanimate"},
	"spells":       {"magecraft", "prowess", "storm", "cantrip", "instant", "sorcery"},
	"lifegain":     {"soul sister", "ajani's pridemate", "heliod", "resplendent angel"},
	"artifacts":    {"affinity", "metalcraft", "improvise", "modular"},
	"enchantments": {"constellation", "aura", "enchantress"},
	"flying":       {"favorable winds", "empyrean eagle", "thunderclap wyvern"},
	"deathtouch":   {"fight", "first strike", "ping", "pinger"},
}

// CalculateSynergy calculates the synergy score for a deck.
func CalculateSynergy(cards []CardData) *SynergyResult {
	result := &SynergyResult{
		SynergyPairs: []SynergyScore{},
		TopSynergies: []string{},
	}

	if len(cards) < 2 {
		result.OverallScore = 0.5 // Neutral for very small decks
		return result
	}

	// Analyze all card pairs for synergies
	for i := 0; i < len(cards); i++ {
		for j := i + 1; j < len(cards); j++ {
			synergies := findSynergies(cards[i], cards[j])
			result.SynergyPairs = append(result.SynergyPairs, synergies...)
		}
	}

	// Count synergy types and calculate overall score
	totalSynergyWeight := 0.0
	for _, synergy := range result.SynergyPairs {
		switch synergy.SynergyType {
		case SynergyTribal:
			result.TribalSynergies++
		case SynergyMechanical:
			result.MechSynergies++
		case SynergyColor:
			result.ColorSynergies++
		}
		totalSynergyWeight += synergy.Weight
	}

	// Calculate overall score (0.0 to 1.0)
	// Base score of 0.5, modified by synergy count relative to deck size
	maxPairs := len(cards) * (len(cards) - 1) / 2
	synergyDensity := float64(len(result.SynergyPairs)) / float64(maxPairs)

	// Score formula: base 0.5, plus up to 0.3 for synergy density, plus up to 0.2 for synergy weight
	result.OverallScore = 0.5 + (synergyDensity * 0.3) + (min(totalSynergyWeight/float64(len(cards)), 1.0) * 0.2)
	result.OverallScore = min(1.0, max(0.0, result.OverallScore))

	// Get top synergy explanations
	result.TopSynergies = getTopSynergyReasons(result.SynergyPairs, 5)

	return result
}

// findSynergies finds synergies between two cards.
func findSynergies(cardA, cardB CardData) []SynergyScore {
	var synergies []SynergyScore

	// 1. Check tribal synergies
	tribalSynergy := checkTribalSynergy(cardA, cardB)
	if tribalSynergy != nil {
		synergies = append(synergies, *tribalSynergy)
	}

	// 2. Check mechanical synergies
	mechSynergies := checkMechanicalSynergies(cardA, cardB)
	synergies = append(synergies, mechSynergies...)

	// 3. Check color synergies (same color bonus)
	colorSynergy := checkColorSynergy(cardA, cardB)
	if colorSynergy != nil {
		synergies = append(synergies, *colorSynergy)
	}

	return synergies
}

// checkTribalSynergy checks for tribal synergies between two cards.
func checkTribalSynergy(cardA, cardB CardData) *SynergyScore {
	// Check if one card is a lord that buffs the other's type
	for _, typeA := range cardA.Types {
		typeLower := strings.ToLower(typeA)
		if lords, ok := tribalLords[typeLower]; ok {
			for _, lord := range lords {
				if containsIgnoreCase(cardB.Name, lord) {
					return &SynergyScore{
						CardA:       cardA.Name,
						CardB:       cardB.Name,
						SynergyType: SynergyTribal,
						Score:       0.8,
						Reason:      cardB.Name + " buffs " + typeA + " creatures like " + cardA.Name,
						Weight:      0.15,
					}
				}
			}
		}
	}

	// Check reverse direction
	for _, typeB := range cardB.Types {
		typeLower := strings.ToLower(typeB)
		if lords, ok := tribalLords[typeLower]; ok {
			for _, lord := range lords {
				if containsIgnoreCase(cardA.Name, lord) {
					return &SynergyScore{
						CardA:       cardA.Name,
						CardB:       cardB.Name,
						SynergyType: SynergyTribal,
						Score:       0.8,
						Reason:      cardA.Name + " buffs " + typeB + " creatures like " + cardB.Name,
						Weight:      0.15,
					}
				}
			}
		}
	}

	// Check if both creatures share a tribal type
	for _, typeA := range cardA.Types {
		for _, typeB := range cardB.Types {
			if strings.EqualFold(typeA, typeB) && isRelevantCreatureType(typeA) {
				return &SynergyScore{
					CardA:       cardA.Name,
					CardB:       cardB.Name,
					SynergyType: SynergyTribal,
					Score:       0.5,
					Reason:      "Both cards are " + typeA + "s",
					Weight:      0.05,
				}
			}
		}
	}

	return nil
}

// checkMechanicalSynergies checks for mechanical synergies between two cards.
func checkMechanicalSynergies(cardA, cardB CardData) []SynergyScore {
	var synergies []SynergyScore

	// Check each mechanical category
	for mechanic, keywords := range mechanicalSynergies {
		// Check if cardA has the mechanic and cardB has synergistic keywords
		if hasMechanic(cardA, mechanic) {
			for _, keyword := range keywords {
				if hasKeyword(cardB, keyword) {
					synergies = append(synergies, SynergyScore{
						CardA:       cardA.Name,
						CardB:       cardB.Name,
						SynergyType: SynergyMechanical,
						Score:       0.7,
						Reason:      cardA.Name + " (" + mechanic + ") synergizes with " + cardB.Name,
						Weight:      0.12,
					})
					break // Only count one synergy per mechanic pair
				}
			}
		}

		// Check reverse direction
		if hasMechanic(cardB, mechanic) {
			for _, keyword := range keywords {
				if hasKeyword(cardA, keyword) {
					// Avoid duplicates
					alreadyFound := false
					for _, existing := range synergies {
						if existing.CardA == cardA.Name && existing.CardB == cardB.Name {
							alreadyFound = true
							break
						}
					}
					if !alreadyFound {
						synergies = append(synergies, SynergyScore{
							CardA:       cardA.Name,
							CardB:       cardB.Name,
							SynergyType: SynergyMechanical,
							Score:       0.7,
							Reason:      cardB.Name + " (" + mechanic + ") synergizes with " + cardA.Name,
							Weight:      0.12,
						})
					}
					break
				}
			}
		}
	}

	// Check spell synergies (instants/sorceries with spell payoffs)
	if cardA.IsSpell && hasKeyword(cardB, "magecraft") {
		synergies = append(synergies, SynergyScore{
			CardA:       cardA.Name,
			CardB:       cardB.Name,
			SynergyType: SynergyMechanical,
			Score:       0.75,
			Reason:      cardA.Name + " triggers " + cardB.Name + "'s spell payoff",
			Weight:      0.13,
		})
	}
	if cardB.IsSpell && hasKeyword(cardA, "magecraft") {
		synergies = append(synergies, SynergyScore{
			CardA:       cardA.Name,
			CardB:       cardB.Name,
			SynergyType: SynergyMechanical,
			Score:       0.75,
			Reason:      cardB.Name + " triggers " + cardA.Name + "'s spell payoff",
			Weight:      0.13,
		})
	}

	return synergies
}

// checkColorSynergy checks for color synergies between two cards.
func checkColorSynergy(cardA, cardB CardData) *SynergyScore {
	// Same color cards have minor synergy (mana consistency)
	if cardA.Color == cardB.Color && cardA.Color != "" && cardA.Color != "C" {
		return &SynergyScore{
			CardA:       cardA.Name,
			CardB:       cardB.Name,
			SynergyType: SynergyColor,
			Score:       0.4,
			Reason:      "Both cards are " + colorName(cardA.Color) + " (color consistency)",
			Weight:      0.03,
		}
	}
	return nil
}

// hasMechanic checks if a card has a particular mechanic.
func hasMechanic(card CardData, mechanic string) bool {
	// Check keywords
	for _, kw := range card.Keywords {
		if containsIgnoreCase(kw, mechanic) {
			return true
		}
	}

	// Check oracle text
	if containsIgnoreCase(card.OracleText, mechanic) {
		return true
	}

	// Check card type for spells mechanic
	if mechanic == "spells" && card.IsSpell {
		return true
	}

	return false
}

// hasKeyword checks if a card has a particular keyword or text pattern.
func hasKeyword(card CardData, keyword string) bool {
	// Check keywords
	for _, kw := range card.Keywords {
		if containsIgnoreCase(kw, keyword) {
			return true
		}
	}

	// Check oracle text
	if containsIgnoreCase(card.OracleText, keyword) {
		return true
	}

	// Check card types for specific keywords
	if keyword == "instant" && card.IsSpell && !card.IsCreature {
		return true
	}
	if keyword == "sorcery" && card.IsSpell && !card.IsCreature {
		return true
	}

	return false
}

// isRelevantCreatureType checks if a creature type is relevant for tribal synergies.
func isRelevantCreatureType(creatureType string) bool {
	relevantTypes := []string{
		"elf", "goblin", "vampire", "zombie", "human", "warrior", "merfolk",
		"wizard", "knight", "spirit", "elemental", "cleric", "rogue", "angel",
		"dragon", "soldier", "pirate", "dinosaur", "faerie", "beast", "cat",
		"bird", "dog", "rat", "sliver",
	}

	typeLower := strings.ToLower(creatureType)
	for _, rt := range relevantTypes {
		if rt == typeLower {
			return true
		}
	}
	return false
}

// containsIgnoreCase checks if haystack contains needle (case-insensitive).
func containsIgnoreCase(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), strings.ToLower(needle))
}

// colorName returns the full color name from abbreviation.
func colorName(abbrev string) string {
	colors := map[string]string{
		"W": "White",
		"U": "Blue",
		"B": "Black",
		"R": "Red",
		"G": "Green",
		"C": "Colorless",
	}
	if name, ok := colors[abbrev]; ok {
		return name
	}
	return abbrev
}

// getTopSynergyReasons returns the top N synergy explanations.
func getTopSynergyReasons(synergies []SynergyScore, n int) []string {
	// Sort by weight (highest first)
	sorted := make([]SynergyScore, len(synergies))
	copy(sorted, synergies)

	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Weight > sorted[i].Weight {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Extract top N reasons
	reasons := make([]string, 0, n)
	for i := 0; i < len(sorted) && i < n; i++ {
		reasons = append(reasons, sorted[i].Reason)
	}

	return reasons
}

// min returns the minimum of two float64 values.
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// max returns the maximum of two float64 values.
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// ConvertCardsToCardData converts basic Card slice to CardData slice.
// This is a helper for integration with existing prediction code.
func ConvertCardsToCardData(cards []Card) []CardData {
	result := make([]CardData, len(cards))
	for i, card := range cards {
		result[i] = CardData{
			Name:       card.Name,
			CMC:        card.CMC,
			Color:      card.Color,
			GIHWR:      card.GIHWR,
			Rarity:     card.Rarity,
			Types:      []string{},
			Keywords:   []string{},
			IsCreature: true, // Default assumption
		}
	}
	return result
}
