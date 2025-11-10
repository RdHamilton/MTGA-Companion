package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// RankTimelineEntry represents a single point in the rank progression timeline.
type RankTimelineEntry struct {
	Timestamp     time.Time `json:"timestamp"`
	Date          string    `json:"date"` // Formatted date for display
	Rank          string    `json:"rank"` // Formatted rank string (e.g., "Gold 2")
	RankClass     *string   `json:"rank_class"`
	RankLevel     *int      `json:"rank_level"`
	RankStep      *int      `json:"rank_step"`
	Percentile    *float64  `json:"percentile"`
	Format        string    `json:"format"`
	SeasonOrdinal int       `json:"season_ordinal"`
	IsChange      bool      `json:"is_change"`    // True if rank changed from previous entry
	IsMilestone   bool      `json:"is_milestone"` // True if achieved new rank class or level
}

// RankTimeline represents a collection of rank progression timeline entries.
type RankTimeline struct {
	Format         string               `json:"format"`
	StartDate      time.Time            `json:"start_date"`
	EndDate        time.Time            `json:"end_date"`
	Entries        []*RankTimelineEntry `json:"entries"`
	TotalChanges   int                  `json:"total_changes"`   // Number of rank changes
	Milestones     int                  `json:"milestones"`      // Number of milestones achieved
	StartRank      string               `json:"start_rank"`      // First rank in timeline
	EndRank        string               `json:"end_rank"`        // Last rank in timeline
	HighestRank    string               `json:"highest_rank"`    // Best rank in timeline
	LowestRank     string               `json:"lowest_rank"`     // Worst rank in timeline
	SeasonsCovered []int                `json:"seasons_covered"` // List of seasons in timeline
}

// TimelinePeriod defines how to group timeline entries.
type TimelinePeriod string

const (
	PeriodAll     TimelinePeriod = "all"     // All rank snapshots
	PeriodDaily   TimelinePeriod = "daily"   // One entry per day (latest)
	PeriodWeekly  TimelinePeriod = "weekly"  // One entry per week (latest)
	PeriodMonthly TimelinePeriod = "monthly" // One entry per month (latest)
)

// GetRankProgressionTimeline generates a timeline of rank progression.
// period determines the granularity: "all", "daily", "weekly", or "monthly"
func (s *Service) GetRankProgressionTimeline(ctx context.Context, format string, startDate, endDate *time.Time, period TimelinePeriod) (*RankTimeline, error) {
	// Get rank history for format
	var history []*models.RankHistory
	var err error

	if startDate != nil && endDate != nil {
		history, err = s.GetRankHistoryByDateRange(ctx, *startDate, *endDate)
		if err != nil {
			return nil, fmt.Errorf("failed to get rank history: %w", err)
		}
		// Filter by format
		filtered := make([]*models.RankHistory, 0)
		for _, h := range history {
			if h.Format == format {
				filtered = append(filtered, h)
			}
		}
		history = filtered
	} else {
		history, err = s.GetRankHistoryByFormat(ctx, format)
		if err != nil {
			return nil, fmt.Errorf("failed to get rank history: %w", err)
		}
	}

	if len(history) == 0 {
		return nil, fmt.Errorf("no rank history found for format: %s", format)
	}

	// Sort by timestamp (oldest first)
	sortRankHistoryByTimestamp(history)

	// Group by period if needed
	grouped := groupByPeriod(history, period)

	// Create timeline entries
	entries := make([]*RankTimelineEntry, 0, len(grouped))
	var prevRank *models.RankHistory
	totalChanges := 0
	milestones := 0
	seasonsMap := make(map[int]bool)

	// Track highest and lowest ranks
	var highestRank, lowestRank *models.RankHistory

	for _, rank := range grouped {
		rankStr := formatRankHistoryString(rank)
		isChange := false
		isMilestone := false

		// Detect changes
		if prevRank != nil {
			// Check if rank changed
			if !ranksEqual(prevRank, rank) {
				isChange = true
				totalChanges++

				// Check if it's a milestone (new rank class or level)
				if isMilestoneChange(prevRank, rank) {
					isMilestone = true
					milestones++
				}
			}
		}

		entry := &RankTimelineEntry{
			Timestamp:     rank.Timestamp,
			Date:          rank.Timestamp.Format("2006-01-02"),
			Rank:          rankStr,
			RankClass:     rank.RankClass,
			RankLevel:     rank.RankLevel,
			RankStep:      rank.RankStep,
			Percentile:    rank.Percentile,
			Format:        rank.Format,
			SeasonOrdinal: rank.SeasonOrdinal,
			IsChange:      isChange,
			IsMilestone:   isMilestone,
		}
		entries = append(entries, entry)

		// Track seasons
		seasonsMap[rank.SeasonOrdinal] = true

		// Track highest and lowest
		if highestRank == nil || compareRanks(rank, highestRank) > 0 {
			highestRank = rank
		}
		if lowestRank == nil || compareRanks(rank, lowestRank) < 0 {
			lowestRank = rank
		}

		prevRank = rank
	}

	// Build seasons list
	seasons := make([]int, 0, len(seasonsMap))
	for season := range seasonsMap {
		seasons = append(seasons, season)
	}

	// Sort seasons
	for i := 0; i < len(seasons); i++ {
		for j := i + 1; j < len(seasons); j++ {
			if seasons[i] > seasons[j] {
				seasons[i], seasons[j] = seasons[j], seasons[i]
			}
		}
	}

	timeline := &RankTimeline{
		Format:         format,
		StartDate:      grouped[0].Timestamp,
		EndDate:        grouped[len(grouped)-1].Timestamp,
		Entries:        entries,
		TotalChanges:   totalChanges,
		Milestones:     milestones,
		StartRank:      formatRankHistoryString(grouped[0]),
		EndRank:        formatRankHistoryString(grouped[len(grouped)-1]),
		HighestRank:    formatRankHistoryString(highestRank),
		LowestRank:     formatRankHistoryString(lowestRank),
		SeasonsCovered: seasons,
	}

	return timeline, nil
}

// sortRankHistoryByTimestamp sorts rank history by timestamp (oldest first).
func sortRankHistoryByTimestamp(history []*models.RankHistory) {
	for i := 0; i < len(history); i++ {
		for j := i + 1; j < len(history); j++ {
			if history[i].Timestamp.After(history[j].Timestamp) {
				history[i], history[j] = history[j], history[i]
			}
		}
	}
}

// groupByPeriod groups rank history by time period.
func groupByPeriod(history []*models.RankHistory, period TimelinePeriod) []*models.RankHistory {
	if period == PeriodAll || len(history) == 0 {
		return history
	}

	grouped := make(map[string]*models.RankHistory)

	for _, rank := range history {
		var key string
		switch period {
		case PeriodDaily:
			key = rank.Timestamp.Format("2006-01-02")
		case PeriodWeekly:
			year, week := rank.Timestamp.ISOWeek()
			key = fmt.Sprintf("%d-W%02d", year, week)
		case PeriodMonthly:
			key = rank.Timestamp.Format("2006-01")
		default:
			key = rank.Timestamp.Format("2006-01-02")
		}

		// Keep the latest entry for each period
		if existing, exists := grouped[key]; !exists || rank.Timestamp.After(existing.Timestamp) {
			grouped[key] = rank
		}
	}

	// Convert map back to slice
	result := make([]*models.RankHistory, 0, len(grouped))
	for _, rank := range grouped {
		result = append(result, rank)
	}

	// Sort by timestamp
	sortRankHistoryByTimestamp(result)

	return result
}

// ranksEqual checks if two ranks are equal.
func ranksEqual(a, b *models.RankHistory) bool {
	// Compare rank class
	if (a.RankClass == nil) != (b.RankClass == nil) {
		return false
	}
	if a.RankClass != nil && b.RankClass != nil && *a.RankClass != *b.RankClass {
		return false
	}

	// Compare rank level
	if (a.RankLevel == nil) != (b.RankLevel == nil) {
		return false
	}
	if a.RankLevel != nil && b.RankLevel != nil && *a.RankLevel != *b.RankLevel {
		return false
	}

	// Compare rank step
	if (a.RankStep == nil) != (b.RankStep == nil) {
		return false
	}
	if a.RankStep != nil && b.RankStep != nil && *a.RankStep != *b.RankStep {
		return false
	}

	return true
}

// isMilestoneChange checks if rank change represents a milestone (new class or level).
func isMilestoneChange(prev, current *models.RankHistory) bool {
	// Check rank class change
	if prev.RankClass != nil && current.RankClass != nil {
		if *prev.RankClass != *current.RankClass {
			return true
		}
	}

	// Check rank level change
	if prev.RankLevel != nil && current.RankLevel != nil {
		if *prev.RankLevel != *current.RankLevel {
			return true
		}
	}

	return false
}
