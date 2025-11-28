package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// RankSnapshot represents a rank at a specific point in time.
type RankSnapshot struct {
	Timestamp     time.Time
	Format        string // "constructed" or "limited"
	RankClass     *string
	RankLevel     *int
	RankStep      *int
	SeasonOrdinal int
}

// formatRankString formats a rank snapshot as a string (e.g., "Gold 2 Step 3").
func formatRankString(snapshot RankSnapshot) string {
	if snapshot.RankClass == nil {
		return "Unranked"
	}

	parts := []string{*snapshot.RankClass}
	if snapshot.RankLevel != nil {
		parts = append(parts, fmt.Sprintf("%d", *snapshot.RankLevel))
	}
	if snapshot.RankStep != nil {
		parts = append(parts, fmt.Sprintf("Step %d", *snapshot.RankStep))
	}

	return strings.Join(parts, " ")
}

// extractRankSnapshots extracts rank information from log entries.
func extractRankSnapshots(entries []*logreader.LogEntry) []RankSnapshot {
	var snapshots []RankSnapshot

	for _, entry := range entries {
		if !entry.IsJSON {
			continue
		}

		// Check for rank information
		_, hasConstructed := entry.JSON["constructedSeasonOrdinal"]
		_, hasLimited := entry.JSON["limitedSeasonOrdinal"]

		if !hasConstructed && !hasLimited {
			continue
		}

		// Parse timestamp
		timestamp := time.Now()
		if entry.Timestamp != "" {
			if parsedTime, err := parseLogTimestamp(entry.Timestamp); err == nil {
				timestamp = parsedTime
			}
		}

		// Extract constructed rank
		if hasConstructed {
			snapshot := RankSnapshot{
				Timestamp: timestamp,
				Format:    "constructed",
			}

			if season, ok := entry.JSON["constructedSeasonOrdinal"].(float64); ok {
				snapshot.SeasonOrdinal = int(season)
			}
			if class, ok := entry.JSON["constructedClass"].(string); ok && class != "" {
				snapshot.RankClass = &class
			}
			if level, ok := entry.JSON["constructedLevel"].(float64); ok {
				levelInt := int(level)
				snapshot.RankLevel = &levelInt
			}
			if step, ok := entry.JSON["constructedStep"].(float64); ok {
				stepInt := int(step)
				snapshot.RankStep = &stepInt
			}

			snapshots = append(snapshots, snapshot)
		}

		// Extract limited rank
		if hasLimited {
			snapshot := RankSnapshot{
				Timestamp: timestamp,
				Format:    "limited",
			}

			if season, ok := entry.JSON["limitedSeasonOrdinal"].(float64); ok {
				snapshot.SeasonOrdinal = int(season)
			}
			if class, ok := entry.JSON["limitedClass"].(string); ok && class != "" {
				snapshot.RankClass = &class
			}
			if level, ok := entry.JSON["limitedLevel"].(float64); ok {
				levelInt := int(level)
				snapshot.RankLevel = &levelInt
			}
			if step, ok := entry.JSON["limitedStep"].(float64); ok {
				stepInt := int(step)
				snapshot.RankStep = &stepInt
			}

			snapshots = append(snapshots, snapshot)
		}
	}

	return snapshots
}

// correlateRanksWithMatches correlates rank snapshots with matches to determine rank before/after.
func correlateRanksWithMatches(matches []matchData, rankSnapshots []RankSnapshot) {
	// Sort snapshots by timestamp
	// Group snapshots by format
	snapshotsByFormat := make(map[string][]RankSnapshot)
	for _, snapshot := range rankSnapshots {
		snapshotsByFormat[snapshot.Format] = append(snapshotsByFormat[snapshot.Format], snapshot)
	}

	// For each match, find the rank before and after
	for i := range matches {
		match := matches[i].Match
		matchTime := match.Timestamp

		// Determine format (constructed or limited) from match format
		// This is a simplification - in reality, we'd need to map event IDs to formats
		format := "constructed" // Default
		if strings.Contains(strings.ToLower(match.Format), "draft") ||
			strings.Contains(strings.ToLower(match.Format), "sealed") ||
			strings.Contains(strings.ToLower(match.Format), "limited") {
			format = "limited"
		}

		// Find rank before match (most recent snapshot before match time)
		var rankBefore *string
		var rankAfter *string

		snapshots := snapshotsByFormat[format]
		for j, snapshot := range snapshots {
			// Find the most recent snapshot before the match
			if snapshot.Timestamp.Before(matchTime) || snapshot.Timestamp.Equal(matchTime) {
				rankStr := formatRankString(snapshot)
				rankBefore = &rankStr

				// Look for the next snapshot after the match to determine rank after
				for k := j + 1; k < len(snapshots); k++ {
					nextSnapshot := snapshots[k]
					if nextSnapshot.Timestamp.After(matchTime) {
						rankAfterStr := formatRankString(nextSnapshot)
						rankAfter = &rankAfterStr
						break
					}
				}
				break
			}
		}

		// If no rank before found, try to find the first snapshot after
		if rankBefore == nil && len(snapshots) > 0 {
			// Use the first snapshot as rank before if it's close to match time
			firstSnapshot := snapshots[0]
			if firstSnapshot.Timestamp.After(matchTime) && firstSnapshot.Timestamp.Sub(matchTime) < 5*time.Minute {
				rankAfterStr := formatRankString(firstSnapshot)
				rankAfter = &rankAfterStr
			}
		}

		match.RankBefore = rankBefore
		match.RankAfter = rankAfter
	}
}

// StoreRankHistory stores rank progression in the database.
// It extracts rank snapshots from log entries and stores them using the rank history repository.
func (s *Service) StoreRankHistory(ctx context.Context, entries []*logreader.LogEntry) error {
	snapshots := extractRankSnapshots(entries)

	for _, snapshot := range snapshots {
		// Convert RankSnapshot to models.RankHistory
		rank := &models.RankHistory{
			AccountID:     s.currentAccountID,
			Timestamp:     snapshot.Timestamp,
			Format:        snapshot.Format,
			SeasonOrdinal: snapshot.SeasonOrdinal,
			RankClass:     snapshot.RankClass,
			RankLevel:     snapshot.RankLevel,
			RankStep:      snapshot.RankStep,
			CreatedAt:     snapshot.Timestamp,
		}

		if err := s.rankHistory.Create(ctx, rank); err != nil {
			return err
		}
	}

	return nil
}

// GetRankTierStats calculates statistics grouped by rank tier.
func (s *Service) GetRankTierStats(ctx context.Context, filter models.StatsFilter) (map[string]*models.Statistics, error) {
	// Get all matches
	matches, err := s.GetMatches(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}

	// Group matches by rank tier
	statsByTier := make(map[string]*models.Statistics)

	for _, match := range matches {
		// Determine rank tier from rank_before
		tier := "Unknown"
		if match.RankBefore != nil {
			tier = extractRankTier(*match.RankBefore)
		}

		if statsByTier[tier] == nil {
			statsByTier[tier] = &models.Statistics{}
		}

		stats := statsByTier[tier]
		stats.TotalMatches++

		if match.Result == "win" {
			stats.MatchesWon++
		} else {
			stats.MatchesLost++
		}

		// Calculate win rate
		if stats.TotalMatches > 0 {
			stats.WinRate = float64(stats.MatchesWon) / float64(stats.TotalMatches)
		}
	}

	return statsByTier, nil
}

// extractRankTier extracts the rank tier (class) from a rank string.
func extractRankTier(rankStr string) string {
	// Rank string format: "Gold 2 Step 3" or "Silver 1" or "Bronze"
	parts := strings.Fields(rankStr)
	if len(parts) > 0 {
		return parts[0] // Return the class (Bronze, Silver, Gold, etc.)
	}
	return "Unknown"
}
