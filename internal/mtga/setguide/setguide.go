package setguide

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

const (
	// CacheValidity is how long cached set files are considered valid
	CacheValidity = 24 * time.Hour
)

// SetGuide provides pre-draft preparation and set analysis.
type SetGuide struct {
	client     *seventeenlands.Client
	cacheDir   string
	setFiles   map[string]*seventeenlands.SetFile // Cached set files by set code
	archetypes map[string][]Archetype             // Archetypes by set code
}

// Archetype represents a draft archetype (e.g., "WU Fliers", "BR Aggro").
type Archetype struct {
	Name       string   // e.g., "WU Fliers"
	Colors     []string // e.g., ["W", "U"]
	Strategy   string   // Strategy description
	KeyCards   []string // Key card names
	Curve      string   // "Low (2-3)", "Medium (3-4)", "High (4+)"
	WinRate    float64  // Expected win rate
	Priorities []string // P1P1 priorities
	Avoid      []string // Cards to avoid
}

// CardTier represents a card with its tier and rating.
type CardTier struct {
	Name     string
	Color    string
	Rarity   string
	CardType string // Primary card type: "Creature", "Instant", "Sorcery", etc.
	GIHWR    float64
	ALSA     float64
	ATA      float64
	GIH      int
	Tier     string // "S", "A", "B", "C", "D", "F"
	Category string // "Bomb", "Removal", "Fixing", etc.
}

// TypeStats represents statistics for a card type within a set.
type TypeStats struct {
	Type       string  // Card type (e.g., "Creature", "Instant")
	Count      int     // Number of cards of this type
	Percentage float64 // Percentage of set (0-100)
	AvgGIHWR   float64 // Average GIHWR for this type
	TopCards   []CardTier
}

// SetOverview provides high-level set information.
type SetOverview struct {
	SetCode      string
	SetName      string
	Mechanics    []string
	TopArchetype []Archetype
	TopCommons   []CardTier
	TopUncommons []CardTier
	KeyRemoval   []CardTier
	TypeStats    []TypeStats // Card type distribution and statistics
	TopCreatures []CardTier  // Best creatures in set
	TopInstants  []CardTier  // Best instants (often removal/tricks)
	TopSorceries []CardTier  // Best sorceries
	TopArtifacts []CardTier  // Best artifacts
	TopEnchants  []CardTier  // Best enchantments
}

// NewSetGuide creates a new set guide.
func NewSetGuide(client *seventeenlands.Client, cacheDir string) *SetGuide {
	return &SetGuide{
		client:     client,
		cacheDir:   cacheDir,
		setFiles:   make(map[string]*seventeenlands.SetFile),
		archetypes: make(map[string][]Archetype),
	}
}

// LoadSet loads or fetches a set file.
func (sg *SetGuide) LoadSet(ctx context.Context, setCode, format string) error {
	// Check if already loaded
	if _, exists := sg.setFiles[setCode]; exists {
		return nil
	}

	// Try to load from cache first
	cacheKey := fmt.Sprintf("%s_%s", setCode, format)
	cachedSetFile, err := sg.loadFromCache(cacheKey)
	if err == nil && cachedSetFile != nil {
		sg.setFiles[setCode] = cachedSetFile
		return nil
	}

	// Fetch from 17Lands
	// Use last 365 days as default date range for color ratings
	// This ensures we get data even for older sets
	endDate := time.Now().Format("2006-01-02")
	startDate := time.Now().AddDate(-1, 0, 0).Format("2006-01-02")

	params := seventeenlands.QueryParams{
		Expansion: setCode,
		Format:    format,
		EventType: format, // EventType is required for color ratings
		StartDate: startDate,
		EndDate:   endDate,
	}

	ratings, err := sg.client.GetCardRatings(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to fetch card ratings: %w", err)
	}

	colorRatings, err := sg.client.GetColorRatings(ctx, params)
	if err != nil {
		// Color ratings are optional - silently continue without them
		colorRatings = []seventeenlands.ColorRating{}
	}

	// Build set file
	setFile := &seventeenlands.SetFile{
		Meta: seventeenlands.SetMeta{
			SetCode:               setCode,
			DraftFormat:           format,
			CollectionDate:        time.Now(),
			SeventeenLandsEnabled: true,
		},
		CardRatings:  make(map[string]*seventeenlands.CardRatingData),
		ColorRatings: make(map[string]float64),
	}

	// Convert ratings to CardRatingData format
	for _, rating := range ratings {
		cardData := &seventeenlands.CardRatingData{
			Name:     rating.Name,
			ArenaID:  rating.MTGAID,
			ManaCost: "", // Will need to fetch from Scryfall
			Colors:   []string{rating.Color},
			Rarity:   rating.Rarity,
			Types:    []string{}, // Will need to fetch
			DeckColors: map[string]*seventeenlands.DeckColorRatings{
				"ALL": {
					GIHWR: rating.GIHWR,
					ALSA:  rating.ALSA,
					ATA:   rating.ATA,
					IWD:   rating.GIHWRDelta,
					GIH:   rating.GIH,
				},
			},
		}
		setFile.CardRatings[fmt.Sprintf("%d", rating.MTGAID)] = cardData
	}

	// Store color ratings
	for _, cr := range colorRatings {
		setFile.ColorRatings[cr.ColorName] = cr.WinRate
	}

	sg.setFiles[setCode] = setFile

	// Save to cache (silently ignore errors - caching is optional)
	_ = sg.saveToCache(cacheKey, setFile)

	return nil
}

// GetTierList returns a sorted tier list for a set.
func (sg *SetGuide) GetTierList(setCode string, opts TierListOptions) ([]CardTier, error) {
	setFile, exists := sg.setFiles[setCode]
	if !exists {
		return nil, fmt.Errorf("set %s not loaded", setCode)
	}

	var tiers []CardTier

	for _, cardData := range setFile.CardRatings {
		// Filter by options
		if opts.Rarity != "" && cardData.Rarity != opts.Rarity {
			continue
		}
		if opts.Color != "" && !contains(cardData.Colors, opts.Color) {
			continue
		}

		// Extract primary card type
		cardType := getPrimaryCardType(cardData.Types)

		// Filter by card type if specified
		if opts.CardType != "" && !stringContainsIgnoreCase(cardType, opts.CardType) {
			continue
		}

		// Get rating for "ALL" colors
		rating, ok := cardData.DeckColors["ALL"]
		if !ok {
			continue
		}

		tier := CardTier{
			Name:     cardData.Name,
			Color:    colorString(cardData.Colors),
			Rarity:   cardData.Rarity,
			CardType: cardType,
			GIHWR:    rating.GIHWR,
			ALSA:     rating.ALSA,
			ATA:      rating.ATA,
			GIH:      rating.GIH,
			Tier:     calculateTier(rating.GIHWR),
			Category: categorizeCard(cardData.Name, cardType, rating.GIHWR),
		}

		tiers = append(tiers, tier)
	}

	// Sort by GIHWR descending
	sort.Slice(tiers, func(i, j int) bool {
		return tiers[i].GIHWR > tiers[j].GIHWR
	})

	// Apply limit
	if opts.Limit > 0 && len(tiers) > opts.Limit {
		tiers = tiers[:opts.Limit]
	}

	return tiers, nil
}

// GetSetOverview returns a high-level set overview.
func (sg *SetGuide) GetSetOverview(setCode string) (*SetOverview, error) {
	setFile, exists := sg.setFiles[setCode]
	if !exists {
		return nil, fmt.Errorf("set %s not loaded", setCode)
	}

	overview := &SetOverview{
		SetCode: setCode,
		SetName: setCodeToName(setCode),
	}

	// Get top commons
	commons, _ := sg.GetTierList(setCode, TierListOptions{
		Rarity: "common",
		Limit:  10,
	})
	overview.TopCommons = commons

	// Get top uncommons
	uncommons, _ := sg.GetTierList(setCode, TierListOptions{
		Rarity: "uncommon",
		Limit:  10,
	})
	overview.TopUncommons = uncommons

	// Get top cards by type
	creatures, _ := sg.GetTierList(setCode, TierListOptions{
		CardType: "Creature",
		Limit:    10,
	})
	overview.TopCreatures = creatures

	instants, _ := sg.GetTierList(setCode, TierListOptions{
		CardType: "Instant",
		Limit:    10,
	})
	overview.TopInstants = instants

	sorceries, _ := sg.GetTierList(setCode, TierListOptions{
		CardType: "Sorcery",
		Limit:    10,
	})
	overview.TopSorceries = sorceries

	artifacts, _ := sg.GetTierList(setCode, TierListOptions{
		CardType: "Artifact",
		Limit:    10,
	})
	overview.TopArtifacts = artifacts

	enchantments, _ := sg.GetTierList(setCode, TierListOptions{
		CardType: "Enchantment",
		Limit:    10,
	})
	overview.TopEnchants = enchantments

	// Calculate type statistics
	overview.TypeStats = sg.calculateTypeStats(setCode, setFile)

	// Get top color pairs and create archetypes
	var colorPairs []struct {
		Colors  string
		WinRate float64
	}
	for colors, winRate := range setFile.ColorRatings {
		if len(colors) == 2 { // Two-color pairs
			colorPairs = append(colorPairs, struct {
				Colors  string
				WinRate float64
			}{colors, winRate})
		}
	}
	sort.Slice(colorPairs, func(i, j int) bool {
		return colorPairs[i].WinRate > colorPairs[j].WinRate
	})

	// Generate archetypes from top color pairs
	topArchetypes := sg.generateArchetypesFromColorPairs(setCode, colorPairs, setFile)
	overview.TopArchetype = topArchetypes

	return overview, nil
}

// GetArchetypes returns draft archetypes for a set.
func (sg *SetGuide) GetArchetypes(setCode string) ([]Archetype, error) {
	archetypes, exists := sg.archetypes[setCode]
	if !exists {
		return nil, fmt.Errorf("no archetypes defined for set %s", setCode)
	}
	return archetypes, nil
}

// GetColorRatings returns color pair win rates for a set.
func (sg *SetGuide) GetColorRatings(setCode string) (map[string]float64, error) {
	setFile, exists := sg.setFiles[setCode]
	if !exists {
		return nil, fmt.Errorf("set %s not loaded", setCode)
	}
	return setFile.ColorRatings, nil
}

// TierListOptions configures tier list generation.
type TierListOptions struct {
	Rarity   string // Filter by rarity ("common", "uncommon", "rare", "mythic")
	Color    string // Filter by color ("W", "U", "B", "R", "G")
	CardType string // Filter by card type ("Creature", "Instant", "Sorcery", "Artifact", "Enchantment", "Land", "Planeswalker")
	Limit    int    // Limit number of results
}

// Cache management

// getCachePath returns the cache file path for a given cache key.
func (sg *SetGuide) getCachePath(cacheKey string) string {
	return filepath.Join(sg.cacheDir, fmt.Sprintf("%s_guide.json", cacheKey))
}

// loadFromCache loads a set file from cache if it exists and is valid.
func (sg *SetGuide) loadFromCache(cacheKey string) (*seventeenlands.SetFile, error) {
	cachePath := sg.getCachePath(cacheKey)

	// Check if cache file exists
	info, err := os.Stat(cachePath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("cache file not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to stat cache file: %w", err)
	}

	// Check if cache is still valid
	if time.Since(info.ModTime()) > CacheValidity {
		return nil, fmt.Errorf("cache expired")
	}

	// Read cache file
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read cache file: %w", err)
	}

	// Parse cache file
	var setFile seventeenlands.SetFile
	if err := json.Unmarshal(data, &setFile); err != nil {
		return nil, fmt.Errorf("failed to parse cache file: %w", err)
	}

	return &setFile, nil
}

// saveToCache saves a set file to cache.
func (sg *SetGuide) saveToCache(cacheKey string, setFile *seventeenlands.SetFile) error {
	// Ensure cache directory exists
	if err := os.MkdirAll(sg.cacheDir, 0o755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(setFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal set file: %w", err)
	}

	// Write to cache file
	cachePath := sg.getCachePath(cacheKey)
	if err := os.WriteFile(cachePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Helper functions

func calculateTier(gihwr float64) string {
	// GIHWR is a decimal (e.g., 0.60 = 60%), so compare against decimal values
	switch {
	case gihwr >= 0.60:
		return "S"
	case gihwr >= 0.57:
		return "A"
	case gihwr >= 0.54:
		return "B"
	case gihwr >= 0.50:
		return "C"
	case gihwr >= 0.45:
		return "D"
	default:
		return "F"
	}
}

func categorizeCard(name, cardType string, gihwr float64) string {
	// Categorization using card type and name patterns
	nameLower := name

	// Check for removal patterns
	if stringContainsIgnoreCase(name, "Murder") ||
		stringContainsIgnoreCase(name, "Destroy") ||
		stringContainsIgnoreCase(name, "Exile") ||
		stringContainsIgnoreCase(name, "Kill") ||
		stringContainsIgnoreCase(name, "Strike") ||
		stringContainsIgnoreCase(name, "Bolt") {
		return "Removal"
	}

	// Instants and sorceries are often removal or tricks
	if cardType == "Instant" {
		if gihwr >= 0.55 {
			return "Removal/Trick"
		}
		return "Trick"
	}

	if cardType == "Sorcery" {
		if gihwr >= 0.55 {
			return "Removal/Sweeper"
		}
		return "Utility"
	}

	// High win rate cards are bombs
	if gihwr >= 0.60 { // GIHWR is decimal, so 0.60 = 60%
		return "Bomb"
	}

	// Lands with name patterns are fixing
	if cardType == "Land" ||
		contains([]string{"Evolving", "Wilds", "Terramorphic"}, nameLower) {
		return "Fixing"
	}

	// Artifacts can be ramp or utility
	if cardType == "Artifact" {
		return "Artifact"
	}

	// Enchantments are usually utility
	if cardType == "Enchantment" {
		return "Enchantment"
	}

	// Creatures are the default
	if cardType == "Creature" {
		if gihwr >= 0.57 {
			return "Premium Creature"
		}
		return "Creature"
	}

	return "Playable"
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func colorString(colors []string) string {
	if len(colors) == 0 {
		return "C" // Colorless
	}
	return colors[0] // Simplified for now
}

func setCodeToName(code string) string {
	// Map of set codes to names
	names := map[string]string{
		"BLB": "Bloomburrow",
		"DSK": "Duskmourn: House of Horror",
		"MKM": "Murders at Karlov Manor",
		"LCI": "The Lost Caverns of Ixalan",
		"WOE": "Wilds of Eldraine",
		"ONE": "Phyrexia: All Will Be One",
	}
	if name, ok := names[code]; ok {
		return name
	}
	return code
}

// guildNames maps two-color combinations to their guild names.
var guildNames = map[string]string{
	"WU": "Azorius",
	"UW": "Azorius",
	"UB": "Dimir",
	"BU": "Dimir",
	"BR": "Rakdos",
	"RB": "Rakdos",
	"RG": "Gruul",
	"GR": "Gruul",
	"GW": "Selesnya",
	"WG": "Selesnya",
	"WB": "Orzhov",
	"BW": "Orzhov",
	"UR": "Izzet",
	"RU": "Izzet",
	"BG": "Golgari",
	"GB": "Golgari",
	"RW": "Boros",
	"WR": "Boros",
	"GU": "Simic",
	"UG": "Simic",
}

// generateArchetypesFromColorPairs creates Archetype objects from the top color pairs.
func (sg *SetGuide) generateArchetypesFromColorPairs(
	setCode string,
	colorPairs []struct {
		Colors  string
		WinRate float64
	},
	setFile *seventeenlands.SetFile,
) []Archetype {
	// Take top 5 color pairs (or fewer if not available)
	limit := 5
	if len(colorPairs) < limit {
		limit = len(colorPairs)
	}

	archetypes := make([]Archetype, 0, limit)
	for i := 0; i < limit; i++ {
		cp := colorPairs[i]
		archetype := sg.createArchetypeFromColorPair(cp.Colors, cp.WinRate, setFile)
		archetypes = append(archetypes, archetype)
	}

	// Store in cache for GetArchetypes method
	sg.archetypes[setCode] = archetypes

	return archetypes
}

// createArchetypeFromColorPair creates an Archetype from a color pair and its win rate.
func (sg *SetGuide) createArchetypeFromColorPair(
	colors string,
	winRate float64,
	setFile *seventeenlands.SetFile,
) Archetype {
	// Get guild name
	guildName := guildNames[colors]
	if guildName == "" {
		guildName = colors
	}

	// Split colors into individual color letters
	colorList := make([]string, 0, len(colors))
	for _, c := range colors {
		colorList = append(colorList, string(c))
	}

	// Find key cards for this color pair
	keyCards := sg.findKeyCardsForColors(colorList, setFile)

	// Determine archetype style based on card analysis
	style := sg.determineArchetypeStyle(colorList, setFile)

	// Create archetype name (e.g., "Azorius Tempo" or just "Azorius" if no style)
	name := guildName
	if style != "" {
		name = guildName + " " + style
	}

	// Determine curve preference based on style
	curve := determineCurveFromStyle(style)

	// Generate strategy description
	strategy := generateStrategyDescription(guildName, style, winRate)

	return Archetype{
		Name:     name,
		Colors:   colorList,
		Strategy: strategy,
		KeyCards: keyCards,
		Curve:    curve,
		WinRate:  winRate,
	}
}

// findKeyCardsForColors finds the top-performing cards in the given colors.
func (sg *SetGuide) findKeyCardsForColors(colors []string, setFile *seventeenlands.SetFile) []string {
	type cardScore struct {
		name  string
		gihwr float64
	}

	var candidates []cardScore

	// Build color pair key for deck-specific ratings (e.g., "WU")
	colorPairKey := ""
	if len(colors) >= 2 {
		colorPairKey = colors[0] + colors[1]
	}

	for name, data := range setFile.CardRatings {
		// Check if card is in one of the archetype colors
		cardColors := data.Colors
		isInColors := false

		// Colorless cards fit any archetype
		if len(cardColors) == 0 {
			isInColors = true
		} else {
			// Check if any of the card's colors match archetype colors
			for _, cardColor := range cardColors {
				for _, c := range colors {
					if cardColor == c {
						isInColors = true
						break
					}
				}
				if isInColors {
					break
				}
			}

			// Also include gold cards that match both archetype colors
			if !isInColors && len(cardColors) >= 2 && len(colors) >= 2 {
				hasFirst := false
				hasSecond := false
				for _, cc := range cardColors {
					if cc == colors[0] {
						hasFirst = true
					}
					if cc == colors[1] {
						hasSecond = true
					}
				}
				if hasFirst && hasSecond {
					isInColors = true
				}
			}
		}

		if !isInColors {
			continue
		}

		// Get ratings - prefer color pair specific, fall back to "ALL"
		var ratings *seventeenlands.DeckColorRatings
		if colorPairKey != "" && data.DeckColors != nil {
			ratings = data.DeckColors[colorPairKey]
		}
		if ratings == nil && data.DeckColors != nil {
			ratings = data.DeckColors["ALL"]
		}
		if ratings == nil {
			continue
		}

		// Only consider cards with meaningful data
		if ratings.GIH < 50 || ratings.GIHWR == 0 {
			continue
		}

		candidates = append(candidates, cardScore{
			name:  name,
			gihwr: ratings.GIHWR,
		})
	}

	// Sort by win rate
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].gihwr > candidates[j].gihwr
	})

	// Take top 5 cards
	limit := 5
	if len(candidates) < limit {
		limit = len(candidates)
	}

	keyCards := make([]string, 0, limit)
	for i := 0; i < limit; i++ {
		keyCards = append(keyCards, candidates[i].name)
	}

	return keyCards
}

// determineArchetypeStyle analyzes cards to determine the archetype's play style.
func (sg *SetGuide) determineArchetypeStyle(colors []string, setFile *seventeenlands.SetFile) string {
	var totalCMC float64
	var cardCount int

	for _, data := range setFile.CardRatings {
		// Check if card is in archetype colors
		cardColors := data.Colors
		isInColors := false

		// Colorless cards fit any archetype
		if len(cardColors) == 0 {
			isInColors = true
		} else {
			for _, cardColor := range cardColors {
				for _, c := range colors {
					if cardColor == c {
						isInColors = true
						break
					}
				}
				if isInColors {
					break
				}
			}
		}

		if !isInColors {
			continue
		}

		cardCount++

		// Use actual CMC if available, otherwise estimate from colors
		if data.CMC > 0 {
			totalCMC += data.CMC
		} else if len(data.Colors) > 0 {
			totalCMC += estimateCMCForColor(data.Colors[0])
		} else {
			totalCMC += 3.0 // Default for colorless
		}
	}

	if cardCount == 0 {
		return ""
	}

	avgCMC := totalCMC / float64(cardCount)

	// Determine style based on color tendencies and estimated CMC
	return classifyStyleFromColors(colors, avgCMC)
}

// estimateCMCForColor provides a rough CMC estimate based on color tendencies.
func estimateCMCForColor(color string) float64 {
	// Color tendencies for average CMC
	switch color {
	case "W":
		return 2.5 // White tends toward weenies
	case "U":
		return 3.0 // Blue is medium
	case "B":
		return 3.0 // Black is medium
	case "R":
		return 2.5 // Red is aggressive
	case "G":
		return 3.5 // Green tends toward bigger creatures
	default:
		return 3.0 // Colorless/multicolor average
	}
}

// classifyStyleFromColors determines archetype style based on colors and CMC.
func classifyStyleFromColors(colors []string, avgCMC float64) string {
	hasW := containsColor(colors, "W")
	hasU := containsColor(colors, "U")
	hasB := containsColor(colors, "B")
	hasR := containsColor(colors, "R")
	hasG := containsColor(colors, "G")

	// Aggressive combinations
	if (hasR && hasW) || (hasR && hasB) {
		if avgCMC < 2.8 {
			return "Aggro"
		}
	}

	// Control combinations
	if (hasU && hasW) || (hasU && hasB) {
		if avgCMC > 3.2 {
			return "Control"
		}
		return "Tempo"
	}

	// Midrange combinations
	if hasG {
		if hasB {
			return "Midrange"
		}
		if hasW {
			return "Go-Wide"
		}
		if hasR {
			return "Stompy"
		}
	}

	// Spells-matter
	if hasU && hasR {
		return "Spells"
	}

	// Default based on CMC
	if avgCMC < 2.8 {
		return "Aggro"
	} else if avgCMC > 3.2 {
		return "Control"
	}
	return "Midrange"
}

// containsColor checks if a color list contains a specific color.
func containsColor(colors []string, target string) bool {
	for _, c := range colors {
		if c == target {
			return true
		}
	}
	return false
}

// determineCurveFromStyle returns the recommended curve based on archetype style.
func determineCurveFromStyle(style string) string {
	switch style {
	case "Aggro":
		return "Low (1-3)"
	case "Tempo", "Spells":
		return "Low-Medium (2-3)"
	case "Midrange", "Go-Wide", "Stompy":
		return "Medium (3-4)"
	case "Control":
		return "High (4+)"
	default:
		return "Medium (3-4)"
	}
}

// generateStrategyDescription creates a strategy description for the archetype.
func generateStrategyDescription(guildName, style string, winRate float64) string {
	winRateStr := fmt.Sprintf("%.1f%%", winRate*100)

	switch style {
	case "Aggro":
		return fmt.Sprintf("%s focuses on aggressive early plays. Win rate: %s. Prioritize low-cost creatures and removal.", guildName, winRateStr)
	case "Tempo":
		return fmt.Sprintf("%s balances threats with interaction. Win rate: %s. Look for efficient creatures and cheap counterspells/removal.", guildName, winRateStr)
	case "Control":
		return fmt.Sprintf("%s aims to control the board and win late. Win rate: %s. Prioritize removal, card draw, and finishers.", guildName, winRateStr)
	case "Midrange":
		return fmt.Sprintf("%s plays efficient threats on curve. Win rate: %s. Value 2-for-1s and resilient creatures.", guildName, winRateStr)
	case "Go-Wide":
		return fmt.Sprintf("%s builds a wide board of creatures. Win rate: %s. Prioritize token makers and anthems.", guildName, winRateStr)
	case "Stompy":
		return fmt.Sprintf("%s plays the biggest creatures. Win rate: %s. Prioritize ramp and large threats.", guildName, winRateStr)
	case "Spells":
		return fmt.Sprintf("%s leverages instants and sorceries. Win rate: %s. Look for spell payoffs and cantrips.", guildName, winRateStr)
	default:
		return fmt.Sprintf("%s archetype. Win rate: %s.", guildName, winRateStr)
	}
}

// getPrimaryCardType extracts the primary card type from the type line.
// MTG cards can have multiple types (e.g., "Artifact Creature"), this returns the primary one.
func getPrimaryCardType(types []string) string {
	if len(types) == 0 {
		return "Unknown"
	}

	// Priority order for primary type determination
	// Check for specific types first
	for _, t := range types {
		switch t {
		case "Creature":
			return "Creature"
		case "Planeswalker":
			return "Planeswalker"
		}
	}

	for _, t := range types {
		switch t {
		case "Instant":
			return "Instant"
		case "Sorcery":
			return "Sorcery"
		}
	}

	for _, t := range types {
		switch t {
		case "Artifact":
			return "Artifact"
		case "Enchantment":
			return "Enchantment"
		case "Land":
			return "Land"
		}
	}

	// Return first type if no match
	return types[0]
}

// stringContainsIgnoreCase checks if haystack contains needle (case-insensitive).
func stringContainsIgnoreCase(haystack, needle string) bool {
	if len(needle) == 0 {
		return true
	}
	if len(haystack) < len(needle) {
		return false
	}

	// Convert both to lowercase for comparison
	haystackLower := toLowerSimple(haystack)
	needleLower := toLowerSimple(needle)

	for i := 0; i <= len(haystackLower)-len(needleLower); i++ {
		if haystackLower[i:i+len(needleLower)] == needleLower {
			return true
		}
	}
	return false
}

// toLowerSimple converts a string to lowercase (ASCII only for performance).
func toLowerSimple(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

// calculateTypeStats computes statistics for each card type in the set.
func (sg *SetGuide) calculateTypeStats(setCode string, setFile *seventeenlands.SetFile) []TypeStats {
	// Map to accumulate stats per type
	typeData := make(map[string]struct {
		count     int
		totalGIHR float64
	})

	totalCards := 0

	for _, cardData := range setFile.CardRatings {
		cardType := getPrimaryCardType(cardData.Types)

		// Get rating
		rating, ok := cardData.DeckColors["ALL"]
		if !ok {
			continue
		}

		totalCards++

		data := typeData[cardType]
		data.count++
		data.totalGIHR += rating.GIHWR
		typeData[cardType] = data
	}

	// Convert to slice and calculate averages
	var stats []TypeStats
	for typeName, data := range typeData {
		avgGIHWR := 0.0
		if data.count > 0 {
			avgGIHWR = data.totalGIHR / float64(data.count)
		}

		percentage := 0.0
		if totalCards > 0 {
			percentage = (float64(data.count) / float64(totalCards)) * 100
		}

		// Get top cards for this type
		topCards, _ := sg.GetTierList(setCode, TierListOptions{
			CardType: typeName,
			Limit:    5,
		})

		stats = append(stats, TypeStats{
			Type:       typeName,
			Count:      data.count,
			Percentage: percentage,
			AvgGIHWR:   avgGIHWR,
			TopCards:   topCards,
		})
	}

	// Sort by count (most common types first)
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	return stats
}
