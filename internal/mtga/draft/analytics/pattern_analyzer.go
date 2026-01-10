// Package analytics provides services for analyzing draft performance and patterns.
package analytics

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// PatternAnalyzer analyzes drafting patterns from pick history.
type PatternAnalyzer struct {
	draftRepo     repository.DraftRepository
	analyticsRepo repository.DraftAnalyticsRepository
	cardStore     CardStore
}

// CardStore provides access to card metadata.
type CardStore interface {
	GetCard(arenaID int) (*cards.Card, error)
	GetCardByName(name string) (*cards.Card, error)
}

// NewPatternAnalyzer creates a new pattern analyzer.
func NewPatternAnalyzer(draftRepo repository.DraftRepository, analyticsRepo repository.DraftAnalyticsRepository, cardStore CardStore) *PatternAnalyzer {
	return &PatternAnalyzer{
		draftRepo:     draftRepo,
		analyticsRepo: analyticsRepo,
		cardStore:     cardStore,
	}
}

// AnalyzePatterns analyzes drafting patterns for a set (or overall if setCode is nil).
func (p *PatternAnalyzer) AnalyzePatterns(ctx context.Context, setCode *string) (*models.DraftPatternAnalysis, error) {
	// Get all completed draft sessions
	sessions, err := p.draftRepo.GetCompletedSessions(ctx, 1000)
	if err != nil {
		return nil, err
	}

	// Filter by set code if provided
	var filteredSessions []*models.DraftSession
	for _, s := range sessions {
		if setCode == nil || s.SetCode == *setCode {
			filteredSessions = append(filteredSessions, s)
		}
	}

	if len(filteredSessions) == 0 {
		// Return empty analysis
		return &models.DraftPatternAnalysis{
			SetCode:      setCode,
			SampleSize:   0,
			CalculatedAt: time.Now(),
		}, nil
	}

	// Collect all picks from sessions
	var allPicks []*pickWithCard
	for _, session := range filteredSessions {
		picks, err := p.draftRepo.GetPicksBySession(ctx, session.ID)
		if err != nil {
			continue
		}

		for _, pick := range picks {
			// Get card metadata
			cardID, err := strconv.Atoi(pick.CardID)
			if err != nil {
				continue
			}
			card, err := p.cardStore.GetCard(cardID)
			if err != nil || card == nil {
				continue
			}

			allPicks = append(allPicks, &pickWithCard{
				pick: pick,
				card: card,
			})
		}
	}

	// Calculate color preferences
	colorPrefs := p.calculateColorPreferences(allPicks)

	// Calculate type preferences
	typePrefs := p.calculateTypePreferences(allPicks)

	// Calculate pick order patterns
	pickOrderPatterns := p.calculatePickOrderPatterns(allPicks)

	// Calculate archetype affinities
	archetypeAffinities := p.calculateArchetypeAffinities(ctx, filteredSessions)

	// Create analysis record
	analysis := &models.DraftPatternAnalysis{
		SetCode:      setCode,
		SampleSize:   len(filteredSessions),
		CalculatedAt: time.Now(),
	}

	// Set JSON fields
	if err := analysis.SetColorPreferences(colorPrefs); err != nil {
		return nil, err
	}
	if err := analysis.SetTypePreferences(typePrefs); err != nil {
		return nil, err
	}
	if err := analysis.SetPickOrderPatterns(pickOrderPatterns); err != nil {
		return nil, err
	}
	if err := analysis.SetArchetypeAffinities(archetypeAffinities); err != nil {
		return nil, err
	}

	// Save to database
	if err := p.analyticsRepo.SavePatternAnalysis(ctx, analysis); err != nil {
		return nil, err
	}

	return analysis, nil
}

// GetPatternAnalysis retrieves cached pattern analysis.
func (p *PatternAnalyzer) GetPatternAnalysis(ctx context.Context, setCode *string) (*models.DraftPatternAnalysis, error) {
	return p.analyticsRepo.GetPatternAnalysis(ctx, setCode)
}

type pickWithCard struct {
	pick *models.DraftPickSession
	card *cards.Card
}

func (p *PatternAnalyzer) calculateColorPreferences(picks []*pickWithCard) []models.ColorPreference {
	colorCounts := make(map[string]int)
	colorPickOrder := make(map[string][]int)
	totalPicks := 0

	for _, pwc := range picks {
		if pwc.card == nil {
			continue
		}
		totalPicks++

		// Count colors
		if len(pwc.card.Colors) == 0 {
			colorCounts["C"]++ // Colorless
			colorPickOrder["C"] = append(colorPickOrder["C"], pwc.pick.PickNumber)
		} else {
			for _, color := range pwc.card.Colors {
				colorCounts[color]++
				colorPickOrder[color] = append(colorPickOrder[color], pwc.pick.PickNumber)
			}
		}
	}

	// Calculate preferences
	var prefs []models.ColorPreference
	for color, count := range colorCounts {
		avgOrder := 0.0
		if len(colorPickOrder[color]) > 0 {
			sum := 0
			for _, o := range colorPickOrder[color] {
				sum += o
			}
			avgOrder = float64(sum) / float64(len(colorPickOrder[color]))
		}

		percentOfPool := 0.0
		if totalPicks > 0 {
			percentOfPool = float64(count) / float64(totalPicks) * 100
		}

		prefs = append(prefs, models.ColorPreference{
			Color:         color,
			TotalPicks:    count,
			PercentOfPool: percentOfPool,
			AvgPickOrder:  avgOrder,
		})
	}

	// Sort by total picks (descending)
	sort.Slice(prefs, func(i, j int) bool {
		return prefs[i].TotalPicks > prefs[j].TotalPicks
	})

	return prefs
}

func (p *PatternAnalyzer) calculateTypePreferences(picks []*pickWithCard) []models.TypePreference {
	typeCounts := make(map[string]int)
	typePickOrder := make(map[string][]int)
	totalPicks := 0

	for _, pwc := range picks {
		if pwc.card == nil {
			continue
		}
		totalPicks++

		// Extract primary type from type line
		primaryType := extractPrimaryType(pwc.card.TypeLine)
		typeCounts[primaryType]++
		typePickOrder[primaryType] = append(typePickOrder[primaryType], pwc.pick.PickNumber)
	}

	// Calculate preferences
	var prefs []models.TypePreference
	for cardType, count := range typeCounts {
		avgOrder := 0.0
		if len(typePickOrder[cardType]) > 0 {
			sum := 0
			for _, o := range typePickOrder[cardType] {
				sum += o
			}
			avgOrder = float64(sum) / float64(len(typePickOrder[cardType]))
		}

		percentOfPool := 0.0
		if totalPicks > 0 {
			percentOfPool = float64(count) / float64(totalPicks) * 100
		}

		prefs = append(prefs, models.TypePreference{
			Type:          cardType,
			TotalPicks:    count,
			PercentOfPool: percentOfPool,
			AvgPickOrder:  avgOrder,
		})
	}

	// Sort by total picks (descending)
	sort.Slice(prefs, func(i, j int) bool {
		return prefs[i].TotalPicks > prefs[j].TotalPicks
	})

	return prefs
}

func extractPrimaryType(typeLine string) string {
	// Handle split type lines (e.g., "Creature — Human Wizard")
	parts := strings.Split(typeLine, " — ")
	if len(parts) == 0 {
		return "Unknown"
	}

	// Get the types before the dash
	typesPart := strings.TrimSpace(parts[0])

	// Check for common types in order of precedence
	if strings.Contains(typesPart, "Creature") {
		return "Creature"
	}
	if strings.Contains(typesPart, "Planeswalker") {
		return "Planeswalker"
	}
	if strings.Contains(typesPart, "Instant") {
		return "Instant"
	}
	if strings.Contains(typesPart, "Sorcery") {
		return "Sorcery"
	}
	if strings.Contains(typesPart, "Enchantment") {
		return "Enchantment"
	}
	if strings.Contains(typesPart, "Artifact") {
		return "Artifact"
	}
	if strings.Contains(typesPart, "Land") {
		return "Land"
	}

	return typesPart
}

func (p *PatternAnalyzer) calculatePickOrderPatterns(picks []*pickWithCard) []models.PickOrderPattern {
	// Group picks by phase: early (1-5), mid (6-10), late (11-14)
	phases := map[string]*phaseStats{
		"early": {phase: "early", minPick: 1, maxPick: 5},
		"mid":   {phase: "mid", minPick: 6, maxPick: 10},
		"late":  {phase: "late", minPick: 11, maxPick: 14},
	}

	for _, pwc := range picks {
		if pwc.card == nil {
			continue
		}

		var phase *phaseStats
		pickNum := pwc.pick.PickNumber
		switch {
		case pickNum >= 1 && pickNum <= 5:
			phase = phases["early"]
		case pickNum >= 6 && pickNum <= 10:
			phase = phases["mid"]
		default:
			phase = phases["late"]
		}

		phase.totalPicks++
		if pwc.pick.PickedCardGIHWR != nil {
			phase.ratingSum += *pwc.pick.PickedCardGIHWR
			phase.ratingCount++
		}

		// Count rarities
		switch pwc.card.Rarity {
		case "rare", "mythic":
			phase.rarePicks++
		case "common":
			phase.commonPicks++
		}
	}

	var patterns []models.PickOrderPattern
	for _, phaseName := range []string{"early", "mid", "late"} {
		phase := phases[phaseName]
		avgRating := 0.0
		if phase.ratingCount > 0 {
			avgRating = phase.ratingSum / float64(phase.ratingCount)
		}

		patterns = append(patterns, models.PickOrderPattern{
			Phase:       phaseName,
			AvgRating:   avgRating,
			TotalPicks:  phase.totalPicks,
			RarePicks:   phase.rarePicks,
			CommonPicks: phase.commonPicks,
		})
	}

	return patterns
}

type phaseStats struct {
	phase       string
	minPick     int
	maxPick     int
	totalPicks  int
	ratingSum   float64
	ratingCount int
	rarePicks   int
	commonPicks int
}

func (p *PatternAnalyzer) calculateArchetypeAffinities(ctx context.Context, sessions []*models.DraftSession) []models.ArchetypeAffinity {
	// Track color pairs and their performance
	colorPairStats := make(map[string]*archetypeStats)

	for _, session := range sessions {
		// Get picks for this session
		picks, err := p.draftRepo.GetPicksBySession(ctx, session.ID)
		if err != nil || len(picks) == 0 {
			continue
		}

		// Count colors in picks
		colorCounts := make(map[string]int)
		for _, pick := range picks {
			cardID, err := strconv.Atoi(pick.CardID)
			if err != nil {
				continue
			}
			card, err := p.cardStore.GetCard(cardID)
			if err != nil || card == nil {
				continue
			}
			for _, color := range card.Colors {
				colorCounts[color]++
			}
		}

		// Determine primary color pair
		colorPair := determinePrimaryColorPair(colorCounts)
		if colorPair == "" {
			continue
		}

		if _, ok := colorPairStats[colorPair]; !ok {
			colorPairStats[colorPair] = &archetypeStats{
				colorPair: colorPair,
			}
		}

		stats := colorPairStats[colorPair]
		stats.timesBuilt++

		// Get match results for this draft if available
		results, err := p.analyticsRepo.GetDraftMatchResults(ctx, session.ID)
		if err == nil && len(results) > 0 {
			wins := 0
			total := 0
			for _, r := range results {
				total++
				if r.Result == "win" {
					wins++
				}
			}
			if total > 0 {
				stats.totalMatches += total
				stats.totalWins += wins
			}
		}
	}

	// Convert to affinities
	totalDrafts := len(sessions)
	var affinities []models.ArchetypeAffinity
	for _, stats := range colorPairStats {
		avgWinRate := 0.0
		if stats.totalMatches > 0 {
			avgWinRate = float64(stats.totalWins) / float64(stats.totalMatches)
		}

		affinityScore := 0.0
		if totalDrafts > 0 {
			affinityScore = float64(stats.timesBuilt) / float64(totalDrafts)
		}

		affinities = append(affinities, models.ArchetypeAffinity{
			ColorPair:     stats.colorPair,
			ArchetypeName: getArchetypeName(stats.colorPair),
			TimesBuilt:    stats.timesBuilt,
			AvgWinRate:    avgWinRate,
			AffinityScore: affinityScore,
		})
	}

	// Sort by times built (descending)
	sort.Slice(affinities, func(i, j int) bool {
		return affinities[i].TimesBuilt > affinities[j].TimesBuilt
	})

	return affinities
}

type archetypeStats struct {
	colorPair    string
	timesBuilt   int
	totalMatches int
	totalWins    int
}

func determinePrimaryColorPair(colorCounts map[string]int) string {
	// Sort colors by count
	type colorCount struct {
		color string
		count int
	}
	var counts []colorCount
	for c, n := range colorCounts {
		counts = append(counts, colorCount{c, n})
	}
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})

	if len(counts) == 0 {
		return ""
	}
	if len(counts) == 1 {
		return counts[0].color
	}

	// Return top 2 colors in WUBRG order
	topTwo := []string{counts[0].color, counts[1].color}
	sort.Slice(topTwo, func(i, j int) bool {
		return colorOrder(topTwo[i]) < colorOrder(topTwo[j])
	})
	return strings.Join(topTwo, "")
}

func colorOrder(color string) int {
	order := map[string]int{"W": 0, "U": 1, "B": 2, "R": 3, "G": 4}
	if o, ok := order[color]; ok {
		return o
	}
	return 5
}

func getArchetypeName(colorPair string) string {
	// Common MTG archetype names (in WUBRG order)
	archetypes := map[string]string{
		"WU": "Azorius",
		"WB": "Orzhov",
		"WR": "Boros",
		"WG": "Selesnya",
		"UB": "Dimir",
		"UR": "Izzet",
		"UG": "Simic",
		"BR": "Rakdos",
		"BG": "Golgari",
		"RG": "Gruul",
		"W":  "Mono-White",
		"U":  "Mono-Blue",
		"B":  "Mono-Black",
		"R":  "Mono-Red",
		"G":  "Mono-Green",
	}
	if name, ok := archetypes[colorPair]; ok {
		return name
	}
	// Try reverse order if not found (e.g., "GW" -> "WG")
	if len(colorPair) == 2 {
		reversed := string(colorPair[1]) + string(colorPair[0])
		if name, ok := archetypes[reversed]; ok {
			return name
		}
	}
	return colorPair
}

// PatternAnalysisResponse is the API response for pattern analysis.
type PatternAnalysisResponse struct {
	SetCode             *string                    `json:"setCode,omitempty"`
	ColorPreferences    []models.ColorPreference   `json:"colorPreferences"`
	TypePreferences     []models.TypePreference    `json:"typePreferences"`
	PickOrderPatterns   []models.PickOrderPattern  `json:"pickOrderPatterns"`
	ArchetypeAffinities []models.ArchetypeAffinity `json:"archetypeAffinities"`
	SampleSize          int                        `json:"sampleSize"`
	CalculatedAt        time.Time                  `json:"calculatedAt"`
}

// ToPatternAnalysisResponse converts a DraftPatternAnalysis to an API response.
func ToPatternAnalysisResponse(a *models.DraftPatternAnalysis) (*PatternAnalysisResponse, error) {
	if a == nil {
		return nil, nil
	}

	colorPrefs, err := a.GetColorPreferences()
	if err != nil {
		return nil, err
	}
	typePrefs, err := a.GetTypePreferences()
	if err != nil {
		return nil, err
	}
	pickPatterns, err := a.GetPickOrderPatterns()
	if err != nil {
		return nil, err
	}
	archetypes, err := a.GetArchetypeAffinities()
	if err != nil {
		return nil, err
	}

	return &PatternAnalysisResponse{
		SetCode:             a.SetCode,
		ColorPreferences:    colorPrefs,
		TypePreferences:     typePrefs,
		PickOrderPatterns:   pickPatterns,
		ArchetypeAffinities: archetypes,
		SampleSize:          a.SampleSize,
		CalculatedAt:        a.CalculatedAt,
	}, nil
}
