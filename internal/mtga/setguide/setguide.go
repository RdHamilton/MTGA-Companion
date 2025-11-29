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
		primaryType := getPrimaryCardType(cardData.Types)

		// Filter by card type if specified
		if opts.CardType != "" && primaryType != opts.CardType {
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
			CardType: primaryType,
			GIHWR:    rating.GIHWR,
			ALSA:     rating.ALSA,
			ATA:      rating.ATA,
			GIH:      rating.GIH,
			Tier:     calculateTier(rating.GIHWR),
			Category: categorizeCard(cardData.Name, rating.GIHWR, primaryType),
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
		Limit:    5,
	})
	overview.TopArtifacts = artifacts

	enchantments, _ := sg.GetTierList(setCode, TierListOptions{
		CardType: "Enchantment",
		Limit:    5,
	})
	overview.TopEnchants = enchantments

	// Calculate type statistics
	overview.TypeStats = sg.calculateTypeStats(setFile)

	// Get top color pairs
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
	CardType string // Filter by card type ("Creature", "Instant", "Sorcery", etc.)
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

func categorizeCard(name string, gihwr float64, cardType string) string {
	// Categorize based on card type and name patterns
	nameLower := name

	// Check for removal spells (instants and sorceries with removal keywords)
	if cardType == "Instant" || cardType == "Sorcery" {
		removalKeywords := []string{"Murder", "Destroy", "Exile", "Kill", "Strike", "Bolt", "Doom", "Terror", "Slay", "Burn", "Flames", "Lightning", "Fire", "Smite", "Banish"}
		for _, keyword := range removalKeywords {
			if stringContainsIgnoreCase(nameLower, keyword) {
				return "Removal"
			}
		}
	}

	// High win rate cards are bombs
	if gihwr >= 0.60 { // GIHWR is decimal, so 0.60 = 60%
		return "Bomb"
	}

	// Combat tricks (instants that likely buff creatures)
	if cardType == "Instant" {
		trickKeywords := []string{"Giant Growth", "Might", "Pump", "Trick", "Strike", "Bite"}
		for _, keyword := range trickKeywords {
			if stringContainsIgnoreCase(nameLower, keyword) {
				return "Combat Trick"
			}
		}
	}

	// Land fixing
	if cardType == "Land" {
		if stringContainsIgnoreCase(nameLower, "Evolving") ||
			stringContainsIgnoreCase(nameLower, "Wilds") ||
			stringContainsIgnoreCase(nameLower, "Terramorphic") ||
			stringContainsIgnoreCase(nameLower, "Dual") ||
			stringContainsIgnoreCase(nameLower, "Gate") {
			return "Fixing"
		}
	}

	// Categorize by card type
	switch cardType {
	case "Creature":
		return "Creature"
	case "Planeswalker":
		return "Bomb" // Planeswalkers are usually bombs
	case "Enchantment":
		return "Enchantment"
	case "Artifact":
		return "Artifact"
	case "Land":
		return "Land"
	default:
		return "Playable"
	}
}

// stringContainsIgnoreCase checks if s contains substr (case-insensitive).
func stringContainsIgnoreCase(s, substr string) bool {
	sLower := ""
	substrLower := ""
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			sLower += string(c + 32)
		} else {
			sLower += string(c)
		}
	}
	for _, c := range substr {
		if c >= 'A' && c <= 'Z' {
			substrLower += string(c + 32)
		} else {
			substrLower += string(c)
		}
	}
	for i := 0; i <= len(sLower)-len(substrLower); i++ {
		if sLower[i:i+len(substrLower)] == substrLower {
			return true
		}
	}
	return false
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

// Card type functions

// primaryCardTypes defines the order of precedence for card types.
// MTG cards can have multiple types, but we extract the most relevant one for categorization.
var primaryCardTypes = []string{
	"Creature",
	"Planeswalker",
	"Instant",
	"Sorcery",
	"Enchantment",
	"Artifact",
	"Land",
}

// getPrimaryCardType extracts the primary card type from a list of types.
// For example, "Artifact Creature" returns "Creature", "Enchantment Creature" returns "Creature".
func getPrimaryCardType(types []string) string {
	if len(types) == 0 {
		return "Unknown"
	}

	// Check for each primary type in order of precedence
	for _, primaryType := range primaryCardTypes {
		for _, cardType := range types {
			if cardType == primaryType {
				return primaryType
			}
		}
	}

	// Return first type if no primary type found
	return types[0]
}

// calculateTypeStats calculates statistics for each card type in the set.
func (sg *SetGuide) calculateTypeStats(setFile *seventeenlands.SetFile) []TypeStats {
	// Track counts and totals by type
	typeCounts := make(map[string]int)
	typeGIHWRSum := make(map[string]float64)
	typeGIHWRCount := make(map[string]int)
	totalCards := 0

	for _, cardData := range setFile.CardRatings {
		primaryType := getPrimaryCardType(cardData.Types)
		typeCounts[primaryType]++
		totalCards++

		// Get GIHWR from ALL colors rating
		if rating, ok := cardData.DeckColors["ALL"]; ok && rating.GIH >= 50 {
			typeGIHWRSum[primaryType] += rating.GIHWR
			typeGIHWRCount[primaryType]++
		}
	}

	// Build type stats list
	var stats []TypeStats
	for _, typeName := range primaryCardTypes {
		count := typeCounts[typeName]
		if count == 0 {
			continue
		}

		percentage := 0.0
		if totalCards > 0 {
			percentage = float64(count) / float64(totalCards) * 100
		}

		avgGIHWR := 0.0
		if typeGIHWRCount[typeName] > 0 {
			avgGIHWR = typeGIHWRSum[typeName] / float64(typeGIHWRCount[typeName])
		}

		stats = append(stats, TypeStats{
			Type:       typeName,
			Count:      count,
			Percentage: percentage,
			AvgGIHWR:   avgGIHWR,
		})
	}

	// Sort by count descending
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	return stats
}

// Archetype generation functions

// guildNames maps color pairs to their guild names.
var guildNames = map[string]string{
	"WU": "Azorius", "UW": "Azorius",
	"UB": "Dimir", "BU": "Dimir",
	"BR": "Rakdos", "RB": "Rakdos",
	"RG": "Gruul", "GR": "Gruul",
	"GW": "Selesnya", "WG": "Selesnya",
	"WB": "Orzhov", "BW": "Orzhov",
	"UR": "Izzet", "RU": "Izzet",
	"BG": "Golgari", "GB": "Golgari",
	"RW": "Boros", "WR": "Boros",
	"GU": "Simic", "UG": "Simic",
}

// generateArchetypesFromColorPairs creates archetypes from the top 5 color pair win rates.
func (sg *SetGuide) generateArchetypesFromColorPairs(setCode string, colorPairs []struct {
	Colors  string
	WinRate float64
}, setFile *seventeenlands.SetFile,
) []Archetype {
	var archetypes []Archetype

	// Take top 5 color pairs
	limit := 5
	if len(colorPairs) < limit {
		limit = len(colorPairs)
	}

	for i := 0; i < limit; i++ {
		cp := colorPairs[i]
		archetype := sg.createArchetypeFromColorPair(cp.Colors, cp.WinRate, setFile)
		archetypes = append(archetypes, archetype)
	}

	// Store in cache
	sg.archetypes[setCode] = archetypes

	return archetypes
}

// createArchetypeFromColorPair creates a single archetype from a color pair.
func (sg *SetGuide) createArchetypeFromColorPair(colorPair string, winRate float64, setFile *seventeenlands.SetFile) Archetype {
	colors := []string{string(colorPair[0]), string(colorPair[1])}
	guildName := guildNames[colorPair]
	if guildName == "" {
		guildName = colorPair
	}

	style := sg.determineArchetypeStyle(colors, setFile)
	keyCards := sg.findKeyCardsForColors(colors, setFile)

	return Archetype{
		Name:     guildName + " " + style,
		Colors:   colors,
		Strategy: generateStrategyDescription(guildName, style, winRate),
		KeyCards: keyCards,
		Curve:    determineCurveFromStyle(style),
		WinRate:  winRate,
	}
}

// findKeyCardsForColors finds the top performing cards for a color pair.
func (sg *SetGuide) findKeyCardsForColors(colors []string, setFile *seventeenlands.SetFile) []string {
	type cardScore struct {
		name  string
		score float64
	}

	var cards []cardScore

	for _, data := range setFile.CardRatings {
		// Check if card matches colors (is in at least one of the colors or is colorless)
		matchesColor := false
		for _, cardColor := range data.Colors {
			if containsColor(colors, cardColor) {
				matchesColor = true
				break
			}
		}
		// Also include colorless cards
		if len(data.Colors) == 0 {
			matchesColor = true
		}
		// Include gold cards that are both colors
		if len(data.Colors) == 2 && containsColor(colors, data.Colors[0]) && containsColor(colors, data.Colors[1]) {
			matchesColor = true
		}

		if !matchesColor {
			continue
		}

		// Get rating for ALL
		rating, ok := data.DeckColors["ALL"]
		if !ok || rating.GIH < 50 { // Minimum sample size
			continue
		}

		cards = append(cards, cardScore{
			name:  data.Name,
			score: rating.GIHWR,
		})
	}

	// Sort by GIHWR descending
	sort.Slice(cards, func(i, j int) bool {
		return cards[i].score > cards[j].score
	})

	// Return top 5 card names
	var result []string
	limit := 5
	if len(cards) < limit {
		limit = len(cards)
	}
	for i := 0; i < limit; i++ {
		result = append(result, cards[i].name)
	}

	return result
}

// determineArchetypeStyle analyzes cards in colors to determine the archetype style.
func (sg *SetGuide) determineArchetypeStyle(colors []string, setFile *seventeenlands.SetFile) string {
	// Calculate average CMC of cards in these colors
	var totalCMC float64
	var count int

	for _, data := range setFile.CardRatings {
		matchesColor := false
		for _, cardColor := range data.Colors {
			if containsColor(colors, cardColor) {
				matchesColor = true
				break
			}
		}
		if !matchesColor {
			continue
		}

		if data.CMC > 0 {
			totalCMC += data.CMC
			count++
		}
	}

	avgCMC := 3.0 // Default
	if count > 0 {
		avgCMC = totalCMC / float64(count)
	} else {
		// Estimate based on colors if no card data
		for _, color := range colors {
			avgCMC = (avgCMC + estimateCMCForColor(color)) / 2
		}
	}

	return classifyStyleFromColors(colors, avgCMC)
}

// estimateCMCForColor returns an estimated average CMC for a color.
func estimateCMCForColor(color string) float64 {
	switch color {
	case "W":
		return 2.5 // White tends to be weenie-focused
	case "U":
		return 3.0 // Blue is medium
	case "B":
		return 3.0 // Black is medium
	case "R":
		return 2.5 // Red tends to be aggressive
	case "G":
		return 3.5 // Green tends to have bigger creatures
	default:
		return 3.0
	}
}

// classifyStyleFromColors determines deck style based on colors and average CMC.
func classifyStyleFromColors(colors []string, avgCMC float64) string {
	hasRed := containsColor(colors, "R")
	hasWhite := containsColor(colors, "W")
	hasBlue := containsColor(colors, "U")
	hasBlack := containsColor(colors, "B")
	hasGreen := containsColor(colors, "G")

	// Aggressive color combinations
	if (hasRed && hasWhite) || (hasRed && hasBlack) {
		if avgCMC < 3.0 {
			return "Aggro"
		}
	}

	// Control color combinations
	if hasBlue && (hasWhite || hasBlack) {
		if avgCMC >= 3.5 {
			return "Control"
		}
		return "Tempo"
	}

	// Spells matter (Izzet)
	if hasBlue && hasRed {
		return "Spells"
	}

	// Green combinations
	if hasGreen {
		if hasBlack {
			return "Midrange"
		}
		if hasWhite {
			return "Go-Wide"
		}
		if hasRed {
			return "Stompy"
		}
	}

	// Default based on CMC
	if avgCMC < 2.5 {
		return "Aggro"
	} else if avgCMC > 3.5 {
		return "Control"
	}
	return "Midrange"
}

// containsColor checks if a color is in the colors slice.
func containsColor(colors []string, target string) bool {
	for _, c := range colors {
		if c == target {
			return true
		}
	}
	return false
}

// determineCurveFromStyle returns a curve description based on archetype style.
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

// generateStrategyDescription creates a human-readable strategy description.
func generateStrategyDescription(guildName, style string, winRate float64) string {
	winPercent := winRate * 100

	switch style {
	case "Aggro":
		return fmt.Sprintf("%s aggro aims to win quickly with efficient creatures and burn spells. Win rate: %.1f%%", guildName, winPercent)
	case "Control":
		return fmt.Sprintf("%s control focuses on card advantage and removal, winning in the late game. Win rate: %.1f%%", guildName, winPercent)
	case "Tempo":
		return fmt.Sprintf("%s tempo combines efficient threats with countermagic and bounce spells. Win rate: %.1f%%", guildName, winPercent)
	case "Spells":
		return fmt.Sprintf("%s spells rewards you for casting instants and sorceries with prowess effects. Win rate: %.1f%%", guildName, winPercent)
	case "Midrange":
		return fmt.Sprintf("%s midrange plays powerful threats at every point on the curve. Win rate: %.1f%%", guildName, winPercent)
	case "Go-Wide":
		return fmt.Sprintf("%s go-wide creates many creatures and uses anthems and buffs to overwhelm. Win rate: %.1f%%", guildName, winPercent)
	case "Stompy":
		return fmt.Sprintf("%s stompy plays the biggest creatures and attacks through blockers. Win rate: %.1f%%", guildName, winPercent)
	default:
		return fmt.Sprintf("%s %s has a %.1f%% win rate in this format.", guildName, style, winPercent)
	}
}
