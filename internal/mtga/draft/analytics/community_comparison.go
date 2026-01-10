package analytics

import (
	"context"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// CommunityComparisonAnalyzer compares user performance to 17Lands community data.
type CommunityComparisonAnalyzer struct {
	draftRepo       repository.DraftRepository
	analyticsRepo   repository.DraftAnalyticsRepository
	ratingsProvider RatingsProvider
}

// RatingsProvider provides access to 17Lands community data.
type RatingsProvider interface {
	GetSetAverageWinRate(setCode, format string) (float64, error)
	GetColorPairWinRates(setCode, format string) (map[string]float64, error)
}

// NewCommunityComparisonAnalyzer creates a new community comparison analyzer.
func NewCommunityComparisonAnalyzer(
	draftRepo repository.DraftRepository,
	analyticsRepo repository.DraftAnalyticsRepository,
	ratingsProvider RatingsProvider,
) *CommunityComparisonAnalyzer {
	return &CommunityComparisonAnalyzer{
		draftRepo:       draftRepo,
		analyticsRepo:   analyticsRepo,
		ratingsProvider: ratingsProvider,
	}
}

// CompareToCommunit compares user performance to 17Lands community averages.
func (c *CommunityComparisonAnalyzer) CompareToCommunity(ctx context.Context, setCode, draftFormat string) (*models.DraftCommunityComparison, error) {
	// Calculate user's win rate for this set/format
	sessions, err := c.draftRepo.GetCompletedSessions(ctx, 1000)
	if err != nil {
		return nil, err
	}

	// Filter to this set
	var userMatches, userWins int
	for _, session := range sessions {
		if session.SetCode != setCode {
			continue
		}

		results, err := c.analyticsRepo.GetDraftMatchResults(ctx, session.ID)
		if err != nil {
			continue
		}

		for _, r := range results {
			userMatches++
			if r.Result == "win" {
				userWins++
			}
		}
	}

	if userMatches == 0 {
		return nil, nil // No data for this set
	}

	userWinRate := float64(userWins) / float64(userMatches)

	// Get community average from 17Lands
	var communityWinRate float64
	if c.ratingsProvider != nil {
		communityWinRate, err = c.ratingsProvider.GetSetAverageWinRate(setCode, draftFormat)
		if err != nil {
			// Default to typical community average if unavailable
			communityWinRate = 0.52
		}
	} else {
		// Default community average
		communityWinRate = 0.52
	}

	// Calculate percentile rank
	percentile := c.calculatePercentile(userWinRate, communityWinRate)

	comparison := &models.DraftCommunityComparison{
		SetCode:             setCode,
		DraftFormat:         draftFormat,
		UserWinRate:         userWinRate,
		CommunityAvgWinRate: communityWinRate,
		PercentileRank:      &percentile,
		SampleSize:          userMatches,
		CalculatedAt:        time.Now(),
	}

	// Save to database
	if err := c.analyticsRepo.SaveCommunityComparison(ctx, comparison); err != nil {
		return nil, err
	}

	return comparison, nil
}

// calculatePercentile estimates user's percentile based on win rate.
// This uses a simplified model based on typical draft win rate distribution.
func (c *CommunityComparisonAnalyzer) calculatePercentile(userWinRate, communityAvg float64) float64 {
	// Win rates in draft typically follow a normal-ish distribution
	// centered around 50% with standard deviation ~10%
	//
	// Rough percentile mapping:
	// - 40% win rate = ~25th percentile
	// - 50% win rate = ~50th percentile
	// - 55% win rate = ~65th percentile
	// - 60% win rate = ~80th percentile
	// - 65% win rate = ~90th percentile
	// - 70% win rate = ~95th percentile

	// Simplified linear approximation
	// Each 1% above 50% is roughly 3 percentile points
	delta := (userWinRate - 0.50) * 300

	percentile := 50 + delta

	// Clamp to valid range
	if percentile < 1 {
		percentile = 1
	}
	if percentile > 99 {
		percentile = 99
	}

	return percentile
}

// GetCommunityComparison retrieves cached community comparison.
func (c *CommunityComparisonAnalyzer) GetCommunityComparison(ctx context.Context, setCode, draftFormat string) (*models.DraftCommunityComparison, error) {
	return c.analyticsRepo.GetCommunityComparison(ctx, setCode, draftFormat)
}

// GetAllComparisons retrieves all cached community comparisons.
func (c *CommunityComparisonAnalyzer) GetAllComparisons(ctx context.Context) ([]*models.DraftCommunityComparison, error) {
	return c.analyticsRepo.GetAllCommunityComparisons(ctx)
}

// CompareArchetypePerformance compares user archetype performance to community.
func (c *CommunityComparisonAnalyzer) CompareArchetypePerformance(ctx context.Context, setCode, draftFormat string) ([]*ArchetypeComparisonEntry, error) {
	// Get user's archetype stats
	userStats, err := c.analyticsRepo.GetArchetypeStats(ctx, setCode)
	if err != nil {
		return nil, err
	}

	// Get community archetype win rates
	var communityRates map[string]float64
	if c.ratingsProvider != nil {
		communityRates, err = c.ratingsProvider.GetColorPairWinRates(setCode, draftFormat)
		if err != nil {
			communityRates = make(map[string]float64)
		}
	} else {
		communityRates = make(map[string]float64)
	}

	// Build comparison entries
	var entries []*ArchetypeComparisonEntry
	for _, stats := range userStats {
		userWinRate := stats.WinRate()

		// Get community rate for this color pair
		communityRate, ok := communityRates[stats.ColorCombination]
		if !ok {
			// Default community rate
			communityRate = 0.52
		}

		entries = append(entries, &ArchetypeComparisonEntry{
			ColorCombination:  stats.ColorCombination,
			ArchetypeName:     stats.ArchetypeName,
			UserWinRate:       userWinRate,
			CommunityWinRate:  communityRate,
			WinRateDelta:      userWinRate - communityRate,
			UserMatchesPlayed: stats.MatchesPlayed,
			PercentileRank:    c.calculatePercentile(userWinRate, communityRate),
			IsAboveCommunity:  userWinRate > communityRate,
		})
	}

	return entries, nil
}

// ArchetypeComparisonEntry represents a comparison of one archetype.
type ArchetypeComparisonEntry struct {
	ColorCombination  string  `json:"colorCombination"`
	ArchetypeName     string  `json:"archetypeName"`
	UserWinRate       float64 `json:"userWinRate"`
	CommunityWinRate  float64 `json:"communityWinRate"`
	WinRateDelta      float64 `json:"winRateDelta"`
	UserMatchesPlayed int     `json:"userMatchesPlayed"`
	PercentileRank    float64 `json:"percentileRank"`
	IsAboveCommunity  bool    `json:"isAboveCommunity"`
}

// CommunityComparisonResponse is the API response for community comparison.
type CommunityComparisonResponse struct {
	SetCode             string                      `json:"setCode"`
	DraftFormat         string                      `json:"draftFormat"`
	UserWinRate         float64                     `json:"userWinRate"`
	CommunityAvgWinRate float64                     `json:"communityAvgWinRate"`
	WinRateDelta        float64                     `json:"winRateDelta"`
	PercentileRank      float64                     `json:"percentileRank"`
	SampleSize          int                         `json:"sampleSize"`
	Rank                string                      `json:"rank"` // e.g., "Top 20%", "Above Average"
	ArchetypeComparison []*ArchetypeComparisonEntry `json:"archetypeComparison,omitempty"`
}

// BuildCommunityComparisonResponse builds the full API response.
func BuildCommunityComparisonResponse(
	comparison *models.DraftCommunityComparison,
	archetypeComparison []*ArchetypeComparisonEntry,
) *CommunityComparisonResponse {
	if comparison == nil {
		return nil
	}

	percentile := 50.0
	if comparison.PercentileRank != nil {
		percentile = *comparison.PercentileRank
	}

	return &CommunityComparisonResponse{
		SetCode:             comparison.SetCode,
		DraftFormat:         comparison.DraftFormat,
		UserWinRate:         comparison.UserWinRate,
		CommunityAvgWinRate: comparison.CommunityAvgWinRate,
		WinRateDelta:        comparison.WinRateDelta(),
		PercentileRank:      percentile,
		SampleSize:          comparison.SampleSize,
		Rank:                getRankLabel(percentile),
		ArchetypeComparison: archetypeComparison,
	}
}

func getRankLabel(percentile float64) string {
	switch {
	case percentile >= 95:
		return "Top 5%"
	case percentile >= 90:
		return "Top 10%"
	case percentile >= 80:
		return "Top 20%"
	case percentile >= 60:
		return "Above Average"
	case percentile >= 40:
		return "Average"
	case percentile >= 20:
		return "Below Average"
	default:
		return "Needs Improvement"
	}
}

// Default17LandsProvider is a basic implementation that uses default values.
type Default17LandsProvider struct {
	// In a real implementation, this would fetch from 17Lands API or cached data
	setAverages   map[string]float64
	colorPairData map[string]map[string]float64
}

// NewDefault17LandsProvider creates a provider with default values.
func NewDefault17LandsProvider() *Default17LandsProvider {
	return &Default17LandsProvider{
		setAverages: map[string]float64{
			// Default win rates for recent sets
			"FDN": 0.52,
			"TLA": 0.51,
			"DSK": 0.52,
			"MH3": 0.51,
			"OTJ": 0.52,
		},
		colorPairData: map[string]map[string]float64{
			"FDN": {
				"WU": 0.53, "WB": 0.51, "WR": 0.52, "WG": 0.54,
				"UB": 0.52, "UR": 0.51, "UG": 0.53,
				"BR": 0.50, "BG": 0.51,
				"RG": 0.52,
			},
		},
	}
}

// GetSetAverageWinRate returns the community average win rate for a set.
func (p *Default17LandsProvider) GetSetAverageWinRate(setCode, format string) (float64, error) {
	if rate, ok := p.setAverages[setCode]; ok {
		return rate, nil
	}
	return 0.52, nil // Default
}

// GetColorPairWinRates returns community win rates by color pair.
func (p *Default17LandsProvider) GetColorPairWinRates(setCode, format string) (map[string]float64, error) {
	if rates, ok := p.colorPairData[setCode]; ok {
		return rates, nil
	}
	// Return default rates
	return map[string]float64{
		"WU": 0.52, "WB": 0.52, "WR": 0.52, "WG": 0.52,
		"UB": 0.52, "UR": 0.52, "UG": 0.52,
		"BR": 0.52, "BG": 0.52,
		"RG": 0.52,
	}, nil
}
