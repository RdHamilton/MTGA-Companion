package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// SeasonStats represents statistics for a specific season/time period.
type SeasonStats struct {
	SeasonName   string    `json:"season_name"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	TotalMatches int       `json:"total_matches"`
	MatchesWon   int       `json:"matches_won"`
	MatchesLost  int       `json:"matches_lost"`
	TotalGames   int       `json:"total_games"`
	GamesWon     int       `json:"games_won"`
	GamesLost    int       `json:"games_lost"`
	WinRate      float64   `json:"win_rate"`      // Match win rate percentage
	GameWinRate  float64   `json:"game_win_rate"` // Game win rate percentage
}

// SeasonComparison compares two seasons/time periods.
type SeasonComparison struct {
	Season1              *SeasonStats `json:"season1"`
	Season2              *SeasonStats `json:"season2"`
	WinRateChange        float64      `json:"win_rate_change"`         // Percentage point change
	GameWinRateChange    float64      `json:"game_win_rate_change"`    // Percentage point change
	MatchCountChange     int          `json:"match_count_change"`      // Absolute change in matches played
	MatchCountChangePerc float64      `json:"match_count_change_perc"` // Percentage change in matches played
	Trend                string       `json:"trend"`                   // "improving", "declining", "stable"
}

// MultiSeasonComparison compares multiple seasons/time periods.
type MultiSeasonComparison struct {
	Seasons      []SeasonStats `json:"seasons"`
	BestSeason   string        `json:"best_season"`   // Season name with highest win rate
	WorstSeason  string        `json:"worst_season"`  // Season name with lowest win rate
	MostActive   string        `json:"most_active"`   // Season with most matches
	OverallTrend string        `json:"overall_trend"` // "improving", "declining", "stable", "variable"
}

// GetSeasonStats calculates statistics for a specific season/time period.
func (s *Service) GetSeasonStats(ctx context.Context, seasonName string, startDate, endDate time.Time, formatFilter *string) (*SeasonStats, error) {
	filter := models.StatsFilter{
		StartDate: &startDate,
		EndDate:   &endDate,
	}
	if formatFilter != nil {
		filter.Format = formatFilter
	}

	matches, err := s.GetMatches(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches for season %s: %w", seasonName, err)
	}

	stats := &SeasonStats{
		SeasonName: seasonName,
		StartDate:  startDate,
		EndDate:    endDate,
	}

	// Calculate statistics
	for _, match := range matches {
		stats.TotalMatches++
		stats.TotalGames += match.PlayerWins + match.OpponentWins
		stats.GamesWon += match.PlayerWins
		stats.GamesLost += match.OpponentWins

		if match.Result == "win" {
			stats.MatchesWon++
		} else {
			stats.MatchesLost++
		}
	}

	// Calculate win rates
	if stats.TotalMatches > 0 {
		stats.WinRate = float64(stats.MatchesWon) / float64(stats.TotalMatches) * 100
	}
	if stats.TotalGames > 0 {
		stats.GameWinRate = float64(stats.GamesWon) / float64(stats.TotalGames) * 100
	}

	return stats, nil
}

// CompareSeasons compares two seasons/time periods.
func (s *Service) CompareSeasons(ctx context.Context, season1Name string, season1Start, season1End time.Time, season2Name string, season2Start, season2End time.Time, formatFilter *string) (*SeasonComparison, error) {
	season1, err := s.GetSeasonStats(ctx, season1Name, season1Start, season1End, formatFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get season1 stats: %w", err)
	}

	season2, err := s.GetSeasonStats(ctx, season2Name, season2Start, season2End, formatFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get season2 stats: %w", err)
	}

	comparison := &SeasonComparison{
		Season1:           season1,
		Season2:           season2,
		WinRateChange:     season2.WinRate - season1.WinRate,
		GameWinRateChange: season2.GameWinRate - season1.GameWinRate,
		MatchCountChange:  season2.TotalMatches - season1.TotalMatches,
	}

	// Calculate percentage change in match count
	if season1.TotalMatches > 0 {
		comparison.MatchCountChangePerc = float64(comparison.MatchCountChange) / float64(season1.TotalMatches) * 100
	}

	// Determine trend (±3% threshold for "stable")
	switch {
	case comparison.WinRateChange > 3.0:
		comparison.Trend = "improving"
	case comparison.WinRateChange < -3.0:
		comparison.Trend = "declining"
	default:
		comparison.Trend = "stable"
	}

	return comparison, nil
}

// CompareMultipleSeasons compares multiple seasons/time periods.
func (s *Service) CompareMultipleSeasons(ctx context.Context, seasons []struct {
	Name      string
	StartDate time.Time
	EndDate   time.Time
}, formatFilter *string,
) (*MultiSeasonComparison, error) {
	if len(seasons) < 2 {
		return nil, fmt.Errorf("need at least 2 seasons for comparison")
	}

	// Get stats for all seasons
	seasonStats := make([]SeasonStats, len(seasons))
	for i, season := range seasons {
		stats, err := s.GetSeasonStats(ctx, season.Name, season.StartDate, season.EndDate, formatFilter)
		if err != nil {
			return nil, fmt.Errorf("failed to get stats for season %s: %w", season.Name, err)
		}
		seasonStats[i] = *stats
	}

	comparison := &MultiSeasonComparison{
		Seasons: seasonStats,
	}

	// Find best/worst/most active seasons
	var bestWinRate, worstWinRate float64 = -1, 101
	var maxMatches int

	for _, stats := range seasonStats {
		if stats.TotalMatches == 0 {
			continue // Skip seasons with no matches
		}

		if stats.WinRate > bestWinRate {
			bestWinRate = stats.WinRate
			comparison.BestSeason = stats.SeasonName
		}
		if stats.WinRate < worstWinRate {
			worstWinRate = stats.WinRate
			comparison.WorstSeason = stats.SeasonName
		}
		if stats.TotalMatches > maxMatches {
			maxMatches = stats.TotalMatches
			comparison.MostActive = stats.SeasonName
		}
	}

	// Determine overall trend by comparing first and last seasons
	if len(seasonStats) >= 2 {
		first := seasonStats[0]
		last := seasonStats[len(seasonStats)-1]

		if first.TotalMatches > 0 && last.TotalMatches > 0 {
			winRateChange := last.WinRate - first.WinRate

			switch {
			case winRateChange > 5.0:
				comparison.OverallTrend = "improving"
			case winRateChange < -5.0:
				comparison.OverallTrend = "declining"
			case len(seasonStats) > 2:
				// Check for variability
				var changes []float64
				for i := 1; i < len(seasonStats); i++ {
					if seasonStats[i].TotalMatches > 0 && seasonStats[i-1].TotalMatches > 0 {
						changes = append(changes, seasonStats[i].WinRate-seasonStats[i-1].WinRate)
					}
				}

				if len(changes) >= 2 {
					var positiveChanges, negativeChanges int
					for _, change := range changes {
						if change > 3 {
							positiveChanges++
						} else if change < -3 {
							negativeChanges++
						}
					}

					if positiveChanges > 0 && negativeChanges > 0 {
						comparison.OverallTrend = "variable"
					} else {
						comparison.OverallTrend = "stable"
					}
				} else {
					comparison.OverallTrend = "stable"
				}
			default:
				comparison.OverallTrend = "stable"
			}
		}
	}

	return comparison, nil
}
