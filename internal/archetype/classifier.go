package archetype

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// ColorPair represents a two-color combination.
type ColorPair struct {
	Colors string // e.g., "WU", "BR", "GW"
	Name   string // e.g., "Azorius", "Rakdos", "Selesnya"
}

// DefaultColorPairs returns the standard MTG color pairs with guild names.
var DefaultColorPairs = []ColorPair{
	{Colors: "WU", Name: "Azorius"},
	{Colors: "UB", Name: "Dimir"},
	{Colors: "BR", Name: "Rakdos"},
	{Colors: "RG", Name: "Gruul"},
	{Colors: "GW", Name: "Selesnya"},
	{Colors: "WB", Name: "Orzhov"},
	{Colors: "UR", Name: "Izzet"},
	{Colors: "BG", Name: "Golgari"},
	{Colors: "RW", Name: "Boros"},
	{Colors: "GU", Name: "Simic"},
}

// ClassificationResult represents the result of classifying a deck.
type ClassificationResult struct {
	PrimaryArchetype    string               // Primary archetype name
	SecondaryArchetype  *string              // Secondary archetype if applicable
	Confidence          float64              // 0.0-1.0
	ColorIdentity       string               // Detected color identity
	DominantColors      []string             // Primary colors in the deck
	ColorPair           *ColorPair           // Detected color pair if applicable
	SignatureCards      []int                // Arena IDs of signature cards
	ArchetypeIndicators []ArchetypeIndicator // Cards that indicate the archetype
	TotalCards          int
	Analysis            *DeckAnalysis
}

// ArchetypeIndicator represents a card that indicates a specific archetype.
type ArchetypeIndicator struct {
	CardID   int
	CardName string
	Weight   float64 // How strongly this card indicates the archetype
	Reason   string  // Why this card indicates the archetype
}

// DeckAnalysis provides detailed breakdown of deck composition.
type DeckAnalysis struct {
	// Color distribution
	ColorCounts    map[string]int // W, U, B, R, G -> count
	ColorlessCount int
	GoldCount      int // Multi-colored cards

	// Type distribution
	CreatureCount     int
	InstantCount      int
	SorceryCount      int
	ArtifactCount     int
	EnchantmentCount  int
	PlaneswalkerCount int
	LandCount         int

	// Mana curve
	ManaCurve map[int]int // CMC -> count
	AvgCMC    float64

	// Card quality
	RareCounts map[string]int // rarity -> count
}

// Classifier provides archetype classification for decks.
type Classifier struct {
	cardService *cards.Service
	deckRepo    repository.DeckRepository
	perfRepo    repository.DeckPerformanceRepository
}

// NewClassifier creates a new archetype classifier.
func NewClassifier(cardService *cards.Service, deckRepo repository.DeckRepository, perfRepo repository.DeckPerformanceRepository) *Classifier {
	return &Classifier{
		cardService: cardService,
		deckRepo:    deckRepo,
		perfRepo:    perfRepo,
	}
}

// ClassifyDeck classifies a deck based on its cards.
func (c *Classifier) ClassifyDeck(ctx context.Context, deckID string) (*ClassificationResult, error) {
	// Get deck cards
	deckCards, err := c.deckRepo.GetCards(ctx, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deck cards: %w", err)
	}

	if len(deckCards) == 0 {
		return nil, fmt.Errorf("deck has no cards")
	}

	// Extract card IDs (only main deck, excluding sideboard for classification)
	cardIDs := make([]int, 0, len(deckCards))
	cardQuantities := make(map[int]int)
	for _, card := range deckCards {
		if card.Board == "main" {
			cardIDs = append(cardIDs, card.CardID)
			cardQuantities[card.CardID] = card.Quantity
		}
	}

	return c.ClassifyCards(cardIDs, cardQuantities)
}

// ClassifyCards classifies a set of cards with their quantities.
func (c *Classifier) ClassifyCards(cardIDs []int, quantities map[int]int) (*ClassificationResult, error) {
	if len(cardIDs) == 0 {
		return nil, fmt.Errorf("no cards to classify")
	}

	// Get card metadata
	cardMetadata, err := c.cardService.GetCards(cardIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to get card metadata: %w", err)
	}

	// Perform deck analysis
	analysis := c.analyzeDeck(cardMetadata, quantities)

	// Detect color identity
	colorIdentity := c.detectColorIdentity(analysis.ColorCounts)
	dominantColors := c.getDominantColors(analysis.ColorCounts, len(cardIDs))
	colorPair := c.detectColorPair(dominantColors)

	// Build classification result
	result := &ClassificationResult{
		ColorIdentity:  colorIdentity,
		DominantColors: dominantColors,
		ColorPair:      colorPair,
		TotalCards:     c.countTotalCards(quantities),
		Analysis:       analysis,
	}

	// Classify into archetype based on analysis
	c.classifyArchetype(result, cardMetadata, quantities)

	return result, nil
}

// ClassifyDraftPool classifies a draft pool based on picked cards.
func (c *Classifier) ClassifyDraftPool(cardIDs []int) (*ClassificationResult, error) {
	// For draft pools, each card has quantity 1
	quantities := make(map[int]int)
	for _, id := range cardIDs {
		quantities[id] = 1
	}
	return c.ClassifyCards(cardIDs, quantities)
}

// analyzeDeck performs detailed analysis of deck composition.
func (c *Classifier) analyzeDeck(cardMetadata map[int]*cards.Card, quantities map[int]int) *DeckAnalysis {
	analysis := &DeckAnalysis{
		ColorCounts: make(map[string]int),
		ManaCurve:   make(map[int]int),
		RareCounts:  make(map[string]int),
	}

	totalCMC := 0.0
	totalNonlandCards := 0

	for cardID, card := range cardMetadata {
		if card == nil {
			continue
		}
		qty := quantities[cardID]
		if qty == 0 {
			qty = 1
		}

		// Color analysis
		colors := card.Colors
		if len(colors) == 0 {
			analysis.ColorlessCount += qty
		} else if len(colors) > 1 {
			analysis.GoldCount += qty
			// Also count individual colors
			for _, color := range colors {
				analysis.ColorCounts[color] += qty
			}
		} else {
			for _, color := range colors {
				analysis.ColorCounts[color] += qty
			}
		}

		// Type analysis
		typeLine := ""
		if card.TypeLine != "" {
			typeLine = strings.ToLower(card.TypeLine)
		}

		if strings.Contains(typeLine, "creature") {
			analysis.CreatureCount += qty
		}
		if strings.Contains(typeLine, "instant") {
			analysis.InstantCount += qty
		}
		if strings.Contains(typeLine, "sorcery") {
			analysis.SorceryCount += qty
		}
		if strings.Contains(typeLine, "artifact") {
			analysis.ArtifactCount += qty
		}
		if strings.Contains(typeLine, "enchantment") {
			analysis.EnchantmentCount += qty
		}
		if strings.Contains(typeLine, "planeswalker") {
			analysis.PlaneswalkerCount += qty
		}
		if strings.Contains(typeLine, "land") {
			analysis.LandCount += qty
		} else {
			// Track mana curve for non-lands
			cmc := int(card.CMC)
			if cmc > 7 {
				cmc = 7 // Group 7+ together
			}
			analysis.ManaCurve[cmc] += qty
			totalCMC += card.CMC * float64(qty)
			totalNonlandCards += qty
		}

		// Rarity analysis
		rarity := strings.ToLower(card.Rarity)
		analysis.RareCounts[rarity] += qty
	}

	// Calculate average CMC
	if totalNonlandCards > 0 {
		analysis.AvgCMC = totalCMC / float64(totalNonlandCards)
	}

	return analysis
}

// detectColorIdentity returns the color identity string (e.g., "WU", "BRG").
func (c *Classifier) detectColorIdentity(colorCounts map[string]int) string {
	colorOrder := []string{"W", "U", "B", "R", "G"}
	var identity strings.Builder

	for _, color := range colorOrder {
		if colorCounts[color] > 0 {
			identity.WriteString(color)
		}
	}

	if identity.Len() == 0 {
		return "C" // Colorless
	}
	return identity.String()
}

// getDominantColors returns colors that represent significant portion of the deck.
func (c *Classifier) getDominantColors(colorCounts map[string]int, totalCards int) []string {
	if totalCards == 0 {
		return nil
	}

	// Calculate total colored cards
	totalColored := 0
	for _, count := range colorCounts {
		totalColored += count
	}

	if totalColored == 0 {
		return nil
	}

	// Sort colors by count
	type colorCount struct {
		color string
		count int
	}
	var counts []colorCount
	for color, count := range colorCounts {
		counts = append(counts, colorCount{color, count})
	}
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})

	// A color is dominant if it represents at least 15% of colored cards
	threshold := float64(totalColored) * 0.15
	var dominant []string
	for _, cc := range counts {
		if float64(cc.count) >= threshold {
			dominant = append(dominant, cc.color)
		}
	}

	return dominant
}

// detectColorPair detects if the deck matches a two-color pair.
func (c *Classifier) detectColorPair(dominantColors []string) *ColorPair {
	if len(dominantColors) != 2 {
		return nil
	}

	// Normalize color order
	colors := strings.Join(dominantColors, "")

	for _, pair := range DefaultColorPairs {
		// Check both orderings
		if colors == pair.Colors || c.reverseString(colors) == pair.Colors {
			return &pair
		}
	}

	return nil
}

// reverseString reverses a string.
func (c *Classifier) reverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// countTotalCards counts total cards including quantities.
func (c *Classifier) countTotalCards(quantities map[int]int) int {
	total := 0
	for _, qty := range quantities {
		total += qty
	}
	return total
}

// classifyArchetype determines the archetype based on deck analysis.
func (c *Classifier) classifyArchetype(result *ClassificationResult, cardMetadata map[int]*cards.Card, quantities map[int]int) {
	analysis := result.Analysis

	// Start with color-based archetype
	if result.ColorPair != nil {
		result.PrimaryArchetype = result.ColorPair.Name
		result.Confidence = 0.5 // Base confidence for color match
	} else if len(result.DominantColors) == 1 {
		result.PrimaryArchetype = "Mono-" + c.colorName(result.DominantColors[0])
		result.Confidence = 0.5
	} else if len(result.DominantColors) > 2 {
		result.PrimaryArchetype = "Multi-color"
		result.Confidence = 0.4
	} else {
		result.PrimaryArchetype = "Unknown"
		result.Confidence = 0.2
	}

	// Analyze deck style to refine archetype
	style := c.detectDeckStyle(analysis)
	if style != "" && result.ColorPair != nil {
		result.PrimaryArchetype = result.ColorPair.Name + " " + style
		result.Confidence += 0.2
	} else if style != "" {
		if result.PrimaryArchetype != "Unknown" {
			result.PrimaryArchetype = result.PrimaryArchetype + " " + style
		} else {
			result.PrimaryArchetype = style
		}
		result.Confidence += 0.15
	}

	// Look for signature cards and archetype indicators
	indicators := c.findArchetypeIndicators(cardMetadata, quantities)
	result.ArchetypeIndicators = indicators

	// Collect signature card IDs
	for _, indicator := range indicators {
		if indicator.Weight >= 2.0 {
			result.SignatureCards = append(result.SignatureCards, indicator.CardID)
		}
	}

	// Boost confidence based on indicators
	if len(indicators) > 0 {
		indicatorBoost := float64(len(indicators)) * 0.05
		if indicatorBoost > 0.2 {
			indicatorBoost = 0.2
		}
		result.Confidence += indicatorBoost
	}

	// Cap confidence at 1.0
	if result.Confidence > 1.0 {
		result.Confidence = 1.0
	}
}

// colorName returns the full name for a color code.
func (c *Classifier) colorName(color string) string {
	names := map[string]string{
		"W": "White",
		"U": "Blue",
		"B": "Black",
		"R": "Red",
		"G": "Green",
	}
	if name, ok := names[color]; ok {
		return name
	}
	return color
}

// detectDeckStyle analyzes the deck composition to determine play style.
func (c *Classifier) detectDeckStyle(analysis *DeckAnalysis) string {
	totalNonLand := analysis.CreatureCount + analysis.InstantCount + analysis.SorceryCount +
		analysis.ArtifactCount + analysis.EnchantmentCount + analysis.PlaneswalkerCount

	if totalNonLand == 0 {
		return ""
	}

	creatureRatio := float64(analysis.CreatureCount) / float64(totalNonLand)
	spellRatio := float64(analysis.InstantCount+analysis.SorceryCount) / float64(totalNonLand)

	// Check mana curve for aggro vs control tendencies
	lowCurve := analysis.ManaCurve[1] + analysis.ManaCurve[2]
	midCurve := analysis.ManaCurve[3] + analysis.ManaCurve[4]
	highCurve := analysis.ManaCurve[5] + analysis.ManaCurve[6] + analysis.ManaCurve[7]

	totalCurve := lowCurve + midCurve + highCurve
	if totalCurve == 0 {
		return ""
	}

	lowRatio := float64(lowCurve) / float64(totalCurve)
	highRatio := float64(highCurve) / float64(totalCurve)

	// Aggro: Low curve, high creature count
	if lowRatio > 0.5 && creatureRatio > 0.6 && analysis.AvgCMC < 2.5 {
		return "Aggro"
	}

	// Control: High curve, spell-heavy
	if highRatio > 0.3 && spellRatio > 0.4 && analysis.AvgCMC > 3.5 {
		return "Control"
	}

	// Midrange: Balanced curve
	if creatureRatio > 0.4 && analysis.AvgCMC >= 2.5 && analysis.AvgCMC <= 3.5 {
		return "Midrange"
	}

	// Tempo: Mixed creatures and spells with lower curve
	if creatureRatio > 0.3 && spellRatio > 0.3 && analysis.AvgCMC < 3.0 {
		return "Tempo"
	}

	return ""
}

// findArchetypeIndicators looks for cards that indicate specific archetypes.
func (c *Classifier) findArchetypeIndicators(cardMetadata map[int]*cards.Card, quantities map[int]int) []ArchetypeIndicator {
	var indicators []ArchetypeIndicator

	for cardID, card := range cardMetadata {
		if card == nil {
			continue
		}

		qty := quantities[cardID]
		if qty == 0 {
			qty = 1
		}

		// Check for archetype-indicating cards based on text
		oracleText := ""
		if card.OracleText != nil {
			oracleText = strings.ToLower(*card.OracleText)
		}
		typeLine := strings.ToLower(card.TypeLine)

		// Flying theme
		if strings.Contains(oracleText, "flying") || strings.Contains(typeLine, "flying") {
			if strings.Contains(oracleText, "creatures you control with flying") ||
				strings.Contains(oracleText, "whenever a creature with flying") {
				indicators = append(indicators, ArchetypeIndicator{
					CardID:   cardID,
					CardName: card.Name,
					Weight:   2.5,
					Reason:   "Flying synergy payoff",
				})
			}
		}

		// Sacrifice theme
		if strings.Contains(oracleText, "sacrifice") {
			if strings.Contains(oracleText, "whenever you sacrifice") ||
				strings.Contains(oracleText, "when this creature dies") {
				indicators = append(indicators, ArchetypeIndicator{
					CardID:   cardID,
					CardName: card.Name,
					Weight:   2.0,
					Reason:   "Sacrifice synergy",
				})
			}
		}

		// Graveyard theme
		if strings.Contains(oracleText, "from your graveyard") ||
			strings.Contains(oracleText, "in your graveyard") {
			indicators = append(indicators, ArchetypeIndicator{
				CardID:   cardID,
				CardName: card.Name,
				Weight:   1.5,
				Reason:   "Graveyard synergy",
			})
		}

		// Token theme
		if strings.Contains(oracleText, "create") && strings.Contains(oracleText, "token") {
			if strings.Contains(oracleText, "whenever") || qty >= 2 {
				indicators = append(indicators, ArchetypeIndicator{
					CardID:   cardID,
					CardName: card.Name,
					Weight:   1.5,
					Reason:   "Token generation",
				})
			}
		}

		// Counters/+1/+1 theme
		if strings.Contains(oracleText, "+1/+1 counter") {
			if strings.Contains(oracleText, "whenever") ||
				strings.Contains(oracleText, "each creature") {
				indicators = append(indicators, ArchetypeIndicator{
					CardID:   cardID,
					CardName: card.Name,
					Weight:   2.0,
					Reason:   "+1/+1 counter synergy",
				})
			}
		}

		// Spells matter theme
		if strings.Contains(oracleText, "whenever you cast") &&
			(strings.Contains(oracleText, "instant") || strings.Contains(oracleText, "sorcery")) {
			indicators = append(indicators, ArchetypeIndicator{
				CardID:   cardID,
				CardName: card.Name,
				Weight:   2.0,
				Reason:   "Spells matter",
			})
		}

		// Ramp theme
		if strings.Contains(oracleText, "search your library for a") &&
			strings.Contains(oracleText, "land") {
			indicators = append(indicators, ArchetypeIndicator{
				CardID:   cardID,
				CardName: card.Name,
				Weight:   1.5,
				Reason:   "Ramp/land search",
			})
		}
	}

	// Sort by weight descending
	sort.Slice(indicators, func(i, j int) bool {
		return indicators[i].Weight > indicators[j].Weight
	})

	return indicators
}

// GetColorPairArchetypes returns known archetypes for a color pair in a specific set.
func (c *Classifier) GetColorPairArchetypes(ctx context.Context, colorPair string, setCode string) ([]*models.DeckArchetype, error) {
	format := "draft"
	return c.perfRepo.ListArchetypes(ctx, &setCode, &format)
}

// GetArchetypeByName retrieves an archetype definition.
func (c *Classifier) GetArchetypeByName(ctx context.Context, name string, setCode *string, format string) (*models.DeckArchetype, error) {
	return c.perfRepo.GetArchetypeByName(ctx, name, setCode, format)
}
