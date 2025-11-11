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
	GIHWR    float64
	ALSA     float64
	ATA      float64
	GIH      int
	Tier     string // "S", "A", "B", "C", "D", "F"
	Category string // "Bomb", "Removal", "Fixing", etc.
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
		// Color ratings are optional - log warning but don't fail
		fmt.Printf("Warning: failed to fetch color ratings: %v\n", err)
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

	// Save to cache
	if err := sg.saveToCache(cacheKey, setFile); err != nil {
		// Log error but don't fail - caching is optional
		fmt.Printf("Warning: failed to save set file to cache: %v\n", err)
	}

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

		// Get rating for "ALL" colors
		rating, ok := cardData.DeckColors["ALL"]
		if !ok {
			continue
		}

		tier := CardTier{
			Name:     cardData.Name,
			Color:    colorString(cardData.Colors),
			Rarity:   cardData.Rarity,
			GIHWR:    rating.GIHWR,
			ALSA:     rating.ALSA,
			ATA:      rating.ATA,
			GIH:      rating.GIH,
			Tier:     calculateTier(rating.GIHWR),
			Category: categorizeCard(cardData.Name, rating.GIHWR),
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

	// TODO: Add archetypes based on top color pairs

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
	Rarity string // Filter by rarity ("common", "uncommon", "rare", "mythic")
	Color  string // Filter by color ("W", "U", "B", "R", "G")
	Limit  int    // Limit number of results
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

func categorizeCard(name string, gihwr float64) string {
	// Simple categorization based on name patterns
	// TODO: Use card types for better categorization
	nameLower := name
	if contains([]string{"Murder", "Destroy", "Exile", "Kill", "Strike", "Bolt"}, nameLower) {
		return "Removal"
	}
	if gihwr >= 0.60 { // GIHWR is decimal, so 0.60 = 60%
		return "Bomb"
	}
	if contains([]string{"Evolving", "Wilds", "Terramorphic"}, nameLower) {
		return "Fixing"
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
