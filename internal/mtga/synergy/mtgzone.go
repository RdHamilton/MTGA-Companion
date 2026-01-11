package synergy

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// MTGZoneClient fetches archetype and deck data from MTGZone.
type MTGZoneClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewMTGZoneClient creates a new MTGZone client.
func NewMTGZoneClient() *MTGZoneClient {
	return &MTGZoneClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://mtgazone.com",
	}
}

// ScrapedDeck represents a deck scraped from MTGZone.
type ScrapedDeck struct {
	Name        string
	Format      string
	Archetype   string
	Tier        string
	Description string
	PlayStyle   string
	SourceURL   string
	MainDeck    []DeckEntry
	Sideboard   []DeckEntry
	Author      string
	Date        string
}

// DeckEntry represents a card entry in a deck.
type DeckEntry struct {
	Quantity int
	CardName string
	Role     string // "core", "flex", "sideboard"
}

// KnownArchetypes provides a mapping of archetype names to their play styles.
var KnownArchetypes = map[string]string{
	// Aggro archetypes
	"mono-red aggro":   "aggro",
	"red deck wins":    "aggro",
	"boros convoke":    "aggro",
	"gruul aggro":      "aggro",
	"mono-white aggro": "aggro",
	"azorius soldiers": "aggro",
	"rakdos aggro":     "aggro",
	"mono-black aggro": "aggro",
	"selesnya tokens":  "aggro",

	// Midrange archetypes
	"domain":             "midrange",
	"esper midrange":     "midrange",
	"golgari midrange":   "midrange",
	"rakdos midrange":    "midrange",
	"jund midrange":      "midrange",
	"orzhov midrange":    "midrange",
	"abzan midrange":     "midrange",
	"sultai midrange":    "midrange",
	"bant midrange":      "midrange",
	"four-color legends": "midrange",

	// Control archetypes
	"azorius control":   "control",
	"esper control":     "control",
	"dimir control":     "control",
	"jeskai control":    "control",
	"mono-blue control": "control",
	"grixis control":    "control",

	// Combo archetypes
	"lotus field":  "combo",
	"reanimator":   "combo",
	"creativity":   "combo",
	"greasefang":   "combo",
	"jeskai combo": "combo",

	// Tempo archetypes
	"azorius tempo":   "tempo",
	"izzet tempo":     "tempo",
	"mono-blue tempo": "tempo",
	"dimir rogues":    "tempo",

	// Ramp archetypes
	"mono-green ramp": "ramp",
	"simic ramp":      "ramp",
	"temur ramp":      "ramp",
}

// ParseDeckFromText parses a deck list from plain text format.
// Supports formats like:
// - "4 Lightning Bolt"
// - "4x Lightning Bolt"
// - "Lightning Bolt x4"
func ParseDeckFromText(text string) ([]DeckEntry, []DeckEntry, error) {
	var mainDeck, sideboard []DeckEntry
	inSideboard := false

	// Pattern: "N CardName" or "Nx CardName" or "N x CardName"
	deckPattern := regexp.MustCompile(`^(\d+)x?\s+(.+)$`)
	altPattern := regexp.MustCompile(`^(.+)\s+x?(\d+)$`)

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for sideboard marker
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "sideboard") {
			inSideboard = true
			continue
		}

		// Skip comments or headers
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}

		var quantity int
		var cardName string

		// Try primary pattern
		if matches := deckPattern.FindStringSubmatch(line); len(matches) == 3 {
			_, _ = fmt.Sscanf(matches[1], "%d", &quantity)
			cardName = strings.TrimSpace(matches[2])
		} else if matches := altPattern.FindStringSubmatch(line); len(matches) == 3 {
			// Try alternative pattern
			cardName = strings.TrimSpace(matches[1])
			_, _ = fmt.Sscanf(matches[2], "%d", &quantity)
		} else {
			// Try to parse as just a card name (assume 1 copy)
			cardName = line
			quantity = 1
		}

		if cardName == "" || quantity == 0 {
			continue
		}

		// Clean up card name (remove set codes, collector numbers, etc.)
		cardName = cleanCardName(cardName)

		entry := DeckEntry{
			Quantity: quantity,
			CardName: cardName,
		}

		if inSideboard {
			entry.Role = "sideboard"
			sideboard = append(sideboard, entry)
		} else {
			mainDeck = append(mainDeck, entry)
		}
	}

	return mainDeck, sideboard, nil
}

// cleanCardName removes extra information from card names.
func cleanCardName(name string) string {
	// Remove set codes in parentheses: "Lightning Bolt (2XM)"
	parenPattern := regexp.MustCompile(`\s*\([^)]+\)\s*$`)
	name = parenPattern.ReplaceAllString(name, "")

	// Remove collector numbers: "Lightning Bolt #123"
	hashPattern := regexp.MustCompile(`\s*#\d+\s*$`)
	name = hashPattern.ReplaceAllString(name, "")

	// Remove foil/showcase markers
	name = strings.TrimSuffix(name, " (Foil)")
	name = strings.TrimSuffix(name, " (Showcase)")
	name = strings.TrimSuffix(name, " (Extended Art)")

	return strings.TrimSpace(name)
}

// DetectArchetype attempts to detect the archetype from a deck's contents.
func DetectArchetype(mainDeck []DeckEntry) string {
	// Count colors from card names (basic heuristic)
	colorWords := map[string]string{
		"mountain": "R",
		"forest":   "G",
		"island":   "U",
		"plains":   "W",
		"swamp":    "B",
	}

	colors := make(map[string]bool)
	creatureCount := 0
	spellCount := 0

	for _, entry := range mainDeck {
		lowerName := strings.ToLower(entry.CardName)

		// Detect colors from basic lands
		for word, color := range colorWords {
			if strings.Contains(lowerName, word) {
				colors[color] = true
			}
		}
	}

	// Simple play style detection based on card patterns
	playStyle := "midrange" // default

	// More creatures = more aggro
	if creatureCount > 20 {
		playStyle = "aggro"
	} else if spellCount > 25 {
		playStyle = "control"
	}

	// Build color prefix
	colorOrder := []string{"W", "U", "B", "R", "G"}
	var colorPrefix string
	for _, c := range colorOrder {
		if colors[c] {
			colorPrefix += c
		}
	}

	if colorPrefix == "" {
		colorPrefix = "colorless"
	}

	return fmt.Sprintf("%s %s", colorPrefix, playStyle)
}

// ClassifyCardRole determines if a card is core, flex, or sideboard.
func ClassifyCardRole(cardName string, quantity int, isInSideboard bool) models.CardRole {
	if isInSideboard {
		return models.CardRoleSideboard
	}

	// 4-ofs are usually core cards
	if quantity >= 4 {
		return models.CardRoleCore
	}

	// 3-ofs can be core or flex depending on the card
	if quantity == 3 {
		// Check for typically core card patterns
		lowerName := strings.ToLower(cardName)
		corePatterns := []string{
			"bolt", "counterspell", "thoughtseize", "fatal push",
			"opt", "consider", "play with fire", "cut down",
		}
		for _, pattern := range corePatterns {
			if strings.Contains(lowerName, pattern) {
				return models.CardRoleCore
			}
		}
	}

	return models.CardRoleFlex
}

// ExtractSynergiesFromText extracts card synergy mentions from article text.
// Returns synergies found with reasons.
func ExtractSynergiesFromText(text string, knownCardNames []string) []models.MTGZoneSynergy {
	var synergies []models.MTGZoneSynergy

	// Synergy indicator patterns
	synergyPatterns := []struct {
		pattern string
		reason  string
	}{
		{`works well with`, "synergy"},
		{`combos with`, "combo"},
		{`pairs nicely with`, "synergy"},
		{`synergizes with`, "synergy"},
		{`enables`, "enabler"},
		{`is enabled by`, "payoff"},
		{`goes well with`, "synergy"},
		{`alongside`, "synergy"},
	}

	lowerText := strings.ToLower(text)

	// For each known card, look for synergy mentions
	for i, cardA := range knownCardNames {
		lowerCardA := strings.ToLower(cardA)

		// Check if this card is mentioned
		if !strings.Contains(lowerText, lowerCardA) {
			continue
		}

		// Look for synergy patterns near this card mention
		for _, pattern := range synergyPatterns {
			if !strings.Contains(lowerText, pattern.pattern) {
				continue
			}

			// Check for mentions of other cards near the synergy pattern
			for j, cardB := range knownCardNames {
				if i == j {
					continue
				}

				lowerCardB := strings.ToLower(cardB)
				if !strings.Contains(lowerText, lowerCardB) {
					continue
				}

				// Check if both cards and the pattern appear in proximity
				// (This is a simplified check - a more sophisticated version
				// would use sentence boundaries and NLP)
				synergies = append(synergies, models.MTGZoneSynergy{
					CardA:      cardA,
					CardB:      cardB,
					Reason:     fmt.Sprintf("%s %s %s", cardA, pattern.reason, cardB),
					Confidence: 0.6,
				})
			}
		}
	}

	return synergies
}

// ArchetypeData represents extracted archetype information.
type ArchetypeData struct {
	Archetype models.MTGZoneArchetype
	CoreCards []models.MTGZoneArchetypeCard
	FlexCards []models.MTGZoneArchetypeCard
	Sideboard []models.MTGZoneArchetypeCard
	Synergies []models.MTGZoneSynergy
}

// BuildArchetypeFromDeck creates archetype data from a scraped deck.
func BuildArchetypeFromDeck(deck *ScrapedDeck) *ArchetypeData {
	archetype := models.MTGZoneArchetype{
		Name:        deck.Archetype,
		Format:      deck.Format,
		Tier:        deck.Tier,
		Description: deck.Description,
		PlayStyle:   deck.PlayStyle,
		SourceURL:   deck.SourceURL,
	}

	// If no archetype name, try to detect it
	if archetype.Name == "" {
		archetype.Name = DetectArchetype(deck.MainDeck)
	}

	// If no play style, infer from known archetypes
	if archetype.PlayStyle == "" {
		lowerName := strings.ToLower(archetype.Name)
		if style, ok := KnownArchetypes[lowerName]; ok {
			archetype.PlayStyle = style
		}
	}

	data := &ArchetypeData{
		Archetype: archetype,
	}

	// Process main deck cards
	for _, entry := range deck.MainDeck {
		role := ClassifyCardRole(entry.CardName, entry.Quantity, false)

		card := models.MTGZoneArchetypeCard{
			CardName: entry.CardName,
			Role:     role,
			Copies:   entry.Quantity,
		}

		// Set importance based on quantity
		if entry.Quantity >= 4 {
			card.Importance = models.CardImportanceEssential
		} else if entry.Quantity >= 3 {
			card.Importance = models.CardImportanceImportant
		} else {
			card.Importance = models.CardImportanceOptional
		}

		if role == models.CardRoleCore {
			data.CoreCards = append(data.CoreCards, card)
		} else {
			data.FlexCards = append(data.FlexCards, card)
		}
	}

	// Process sideboard
	for _, entry := range deck.Sideboard {
		card := models.MTGZoneArchetypeCard{
			CardName:   entry.CardName,
			Role:       models.CardRoleSideboard,
			Copies:     entry.Quantity,
			Importance: models.CardImportanceOptional,
		}
		data.Sideboard = append(data.Sideboard, card)
	}

	return data
}

// FetchPage fetches a page from MTGZone.
func (c *MTGZoneClient) FetchPage(ctx context.Context, path string) (string, error) {
	url := c.baseURL + path
	if strings.HasPrefix(path, "http") {
		url = path
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	req.Header.Set("User-Agent", "MTGA-Companion/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch page: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	return string(body), nil
}

// StandardMetaTierList represents known Standard meta archetypes with estimated tiers.
// This serves as a fallback when scraping isn't available.
var StandardMetaTierList = []ArchetypeDefinition{
	{Name: "Domain Ramp", Tier: "S", PlayStyle: "midrange", Format: "Standard"},
	{Name: "Mono-Red Aggro", Tier: "S", PlayStyle: "aggro", Format: "Standard"},
	{Name: "Esper Midrange", Tier: "A", PlayStyle: "midrange", Format: "Standard"},
	{Name: "Azorius Control", Tier: "A", PlayStyle: "control", Format: "Standard"},
	{Name: "Boros Convoke", Tier: "A", PlayStyle: "aggro", Format: "Standard"},
	{Name: "Golgari Midrange", Tier: "A", PlayStyle: "midrange", Format: "Standard"},
	{Name: "Dimir Control", Tier: "B", PlayStyle: "control", Format: "Standard"},
	{Name: "Gruul Aggro", Tier: "B", PlayStyle: "aggro", Format: "Standard"},
	{Name: "Rakdos Midrange", Tier: "B", PlayStyle: "midrange", Format: "Standard"},
	{Name: "Selesnya Tokens", Tier: "B", PlayStyle: "aggro", Format: "Standard"},
}

// HistoricMetaTierList represents known Historic meta archetypes.
var HistoricMetaTierList = []ArchetypeDefinition{
	{Name: "Rakdos Arcanist", Tier: "S", PlayStyle: "midrange", Format: "Historic"},
	{Name: "Mono-Red Goblins", Tier: "A", PlayStyle: "aggro", Format: "Historic"},
	{Name: "Azorius Control", Tier: "A", PlayStyle: "control", Format: "Historic"},
	{Name: "Gruul Aggro", Tier: "A", PlayStyle: "aggro", Format: "Historic"},
	{Name: "Jeskai Control", Tier: "B", PlayStyle: "control", Format: "Historic"},
	{Name: "Mono-Blue Tempo", Tier: "B", PlayStyle: "tempo", Format: "Historic"},
}

// ArchetypeDefinition is a simple archetype definition.
type ArchetypeDefinition struct {
	Name      string
	Tier      string
	PlayStyle string
	Format    string
	CoreCards []string
}

// GetMetaTierList returns the meta tier list for a format.
func GetMetaTierList(format string) []ArchetypeDefinition {
	switch strings.ToLower(format) {
	case "standard":
		return StandardMetaTierList
	case "historic":
		return HistoricMetaTierList
	default:
		return []ArchetypeDefinition{}
	}
}

// ToJSON serializes archetype data to JSON.
func (a *ArchetypeData) ToJSON() (string, error) {
	data, err := json.Marshal(a)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
