package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// ComparisonGroup represents a labeled group of matches for comparison.
type ComparisonGroup struct {
	Label      string              // Human-readable label for this group
	Filter     models.StatsFilter  // Filter defining this group
	Statistics *models.Statistics  // Aggregated statistics for this group
	MatchCount int                 // Number of matches in this group
}

// ComparisonResult represents the result of comparing two or more groups.
type ComparisonResult struct {
	Groups          []*ComparisonGroup
	BestGroup       *ComparisonGroup // Group with highest win rate
	WorstGroup      *ComparisonGroup // Group with lowest win rate
	WinRateDiff     float64          // Difference between best and worst
	TotalMatches    int              // Total matches across all groups
	ComparisonDate  time.Time        // When comparison was performed
}

// ComparisonDiff represents the difference between two specific groups.
type ComparisonDiff struct {
	Group1Label     string
	Group2Label     string
	WinRateDiff     float64 // Group1 - Group2
	GameWinRateDiff float64
	MatchCountDiff  int
	GamesPlayedDiff int
	Trend           string // "improving", "declining", "stable"
}

// CompareMatches compares multiple groups of matches based on different filters.
// Each group is defined by a filter and a label.
func (s *Service) CompareMatches(ctx context.Context, groups []ComparisonGroup) (*ComparisonResult, error) {
	if len(groups) < 2 {
		return nil, fmt.Errorf("need at least 2 groups to compare")
	}

	result := &ComparisonResult{
		Groups:         make([]*ComparisonGroup, 0, len(groups)),
		ComparisonDate: time.Now(),
	}

	// Fetch statistics for each group
	for i := range groups {
		group := &groups[i]

		// Get statistics for this group
		stats, err := s.GetStats(ctx, group.Filter)
		if err != nil {
			return nil, fmt.Errorf("failed to get stats for group %s: %w", group.Label, err)
		}

		group.Statistics = stats
		group.MatchCount = stats.TotalMatches
		result.Groups = append(result.Groups, group)
		result.TotalMatches += stats.TotalMatches
	}

	// Find best and worst performing groups
	if len(result.Groups) > 0 {
		result.BestGroup = result.Groups[0]
		result.WorstGroup = result.Groups[0]

		for _, group := range result.Groups[1:] {
			if group.Statistics.WinRate > result.BestGroup.Statistics.WinRate {
				result.BestGroup = group
			}
			if group.Statistics.WinRate < result.WorstGroup.Statistics.WinRate {
				result.WorstGroup = group
			}
		}

		result.WinRateDiff = result.BestGroup.Statistics.WinRate - result.WorstGroup.Statistics.WinRate
	}

	return result, nil
}

// CompareTwoGroups compares exactly two groups and returns detailed differences.
func (s *Service) CompareTwoGroups(ctx context.Context, group1, group2 ComparisonGroup) (*ComparisonDiff, error) {
	// Get statistics for both groups
	stats1, err := s.GetStats(ctx, group1.Filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats for group 1: %w", err)
	}

	stats2, err := s.GetStats(ctx, group2.Filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats for group 2: %w", err)
	}

	diff := &ComparisonDiff{
		Group1Label:     group1.Label,
		Group2Label:     group2.Label,
		WinRateDiff:     stats1.WinRate - stats2.WinRate,
		GameWinRateDiff: stats1.GameWinRate - stats2.GameWinRate,
		MatchCountDiff:  stats1.TotalMatches - stats2.TotalMatches,
		GamesPlayedDiff: stats1.TotalGames - stats2.TotalGames,
	}

	// Determine trend
	if diff.WinRateDiff > 0.05 {
		diff.Trend = "improving"
	} else if diff.WinRateDiff < -0.05 {
		diff.Trend = "declining"
	} else {
		diff.Trend = "stable"
	}

	return diff, nil
}

// CompareFormats compares performance across different formats.
func (s *Service) CompareFormats(ctx context.Context, formats []string, baseFilter models.StatsFilter) (*ComparisonResult, error) {
	groups := make([]ComparisonGroup, 0, len(formats))

	for _, format := range formats {
		filter := baseFilter
		filter.Format = &format

		groups = append(groups, ComparisonGroup{
			Label:  format,
			Filter: filter,
		})
	}

	return s.CompareMatches(ctx, groups)
}

// CompareTimePeriods compares performance across different time periods.
func (s *Service) CompareTimePeriods(ctx context.Context, periods []struct {
	Label string
	Start time.Time
	End   time.Time
}, baseFilter models.StatsFilter) (*ComparisonResult, error) {
	groups := make([]ComparisonGroup, 0, len(periods))

	for _, period := range periods {
		filter := baseFilter
		filter.StartDate = &period.Start
		filter.EndDate = &period.End

		groups = append(groups, ComparisonGroup{
			Label:  period.Label,
			Filter: filter,
		})
	}

	return s.CompareMatches(ctx, groups)
}

// CompareDecks compares performance across different decks.
func (s *Service) CompareDecks(ctx context.Context, deckIDs []string, baseFilter models.StatsFilter) (*ComparisonResult, error) {
	groups := make([]ComparisonGroup, 0, len(deckIDs))

	for _, deckID := range deckIDs {
		filter := baseFilter
		filter.DeckID = &deckID

		// Use deck ID as label (could be enhanced to fetch deck name in future)
		groups = append(groups, ComparisonGroup{
			Label:  deckID,
			Filter: filter,
		})
	}

	return s.CompareMatches(ctx, groups)
}
