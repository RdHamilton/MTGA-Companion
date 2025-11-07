package logreader

import "strings"

// DraftStatistics represents aggregated draft performance statistics.
type DraftStatistics struct {
	TotalDrafts    int
	TotalWins      int
	TotalLosses    int
	AverageWins    float64
	BestRecord     DraftRecord
	WorstRecord    DraftRecord
	DraftsByFormat map[string]*FormatDraftStats
	DraftsBySet    map[string]*SetDraftStats
	TrophyCount    int // 7+ wins
}

// DraftRecord represents a single draft record.
type DraftRecord struct {
	EventName string
	Wins      int
	Losses    int
	Format    string
	Set       string
}

// FormatDraftStats represents statistics for a specific format.
type FormatDraftStats struct {
	Format      string
	DraftCount  int
	Wins        int
	Losses      int
	WinRate     float64
	AverageWins float64
}

// SetDraftStats represents statistics for a specific set.
type SetDraftStats struct {
	SetCode     string
	DraftCount  int
	Wins        int
	Losses      int
	WinRate     float64
	AverageWins float64
}

// CalculateDraftStatistics calculates aggregated statistics from draft history.
func CalculateDraftStatistics(history *DraftHistory) *DraftStatistics {
	if history == nil || len(history.Drafts) == 0 {
		return nil
	}

	stats := &DraftStatistics{
		DraftsByFormat: make(map[string]*FormatDraftStats),
		DraftsBySet:    make(map[string]*SetDraftStats),
	}

	var bestRecord, worstRecord *DraftRecord
	bestWins := -1
	worstWins := 999

	for _, draft := range history.Drafts {
		stats.TotalDrafts++
		stats.TotalWins += draft.Wins
		stats.TotalLosses += draft.Losses

		// Track best/worst records
		if draft.Wins > bestWins || (draft.Wins == bestWins && draft.Losses < worstRecord.Losses) {
			bestWins = draft.Wins
			bestRecord = &DraftRecord{
				EventName: draft.EventName,
				Wins:      draft.Wins,
				Losses:    draft.Losses,
				Format:    extractFormat(draft.EventName),
				Set:       extractSet(draft.EventName),
			}
		}

		if draft.Wins < worstWins || (draft.Wins == worstWins && draft.Losses > worstRecord.Losses) {
			worstWins = draft.Wins
			worstRecord = &DraftRecord{
				EventName: draft.EventName,
				Wins:      draft.Wins,
				Losses:    draft.Losses,
				Format:    extractFormat(draft.EventName),
				Set:       extractSet(draft.EventName),
			}
		}

		// Count trophies (7+ wins)
		if draft.Wins >= 7 {
			stats.TrophyCount++
		}

		// Aggregate by format
		format := extractFormat(draft.EventName)
		if stats.DraftsByFormat[format] == nil {
			stats.DraftsByFormat[format] = &FormatDraftStats{
				Format: format,
			}
		}
		formatStats := stats.DraftsByFormat[format]
		formatStats.DraftCount++
		formatStats.Wins += draft.Wins
		formatStats.Losses += draft.Losses

		// Aggregate by set
		set := extractSet(draft.EventName)
		if stats.DraftsBySet[set] == nil {
			stats.DraftsBySet[set] = &SetDraftStats{
				SetCode: set,
			}
		}
		setStats := stats.DraftsBySet[set]
		setStats.DraftCount++
		setStats.Wins += draft.Wins
		setStats.Losses += draft.Losses
	}

	// Calculate averages and win rates
	if stats.TotalDrafts > 0 {
		stats.AverageWins = float64(stats.TotalWins) / float64(stats.TotalDrafts)
	}

	if bestRecord != nil {
		stats.BestRecord = *bestRecord
	}
	if worstRecord != nil {
		stats.WorstRecord = *worstRecord
	}

	// Calculate format statistics
	for _, formatStats := range stats.DraftsByFormat {
		if formatStats.DraftCount > 0 {
			formatStats.AverageWins = float64(formatStats.Wins) / float64(formatStats.DraftCount)
			totalMatches := formatStats.Wins + formatStats.Losses
			if totalMatches > 0 {
				formatStats.WinRate = float64(formatStats.Wins) / float64(totalMatches)
			}
		}
	}

	// Calculate set statistics
	for _, setStats := range stats.DraftsBySet {
		if setStats.DraftCount > 0 {
			setStats.AverageWins = float64(setStats.Wins) / float64(setStats.DraftCount)
			totalMatches := setStats.Wins + setStats.Losses
			if totalMatches > 0 {
				setStats.WinRate = float64(setStats.Wins) / float64(totalMatches)
			}
		}
	}

	return stats
}

// extractFormat extracts the format from an event name.
func extractFormat(eventName string) string {
	// Simple heuristic: look for format keywords
	if contains(eventName, "Premier") {
		return "Premier"
	}
	if contains(eventName, "Quick") {
		return "Quick"
	}
	if contains(eventName, "Traditional") {
		return "Traditional"
	}
	if contains(eventName, "Sealed") {
		return "Sealed"
	}
	return "Unknown"
}

// extractSet extracts the set code from an event name.
func extractSet(eventName string) string {
	// Simple heuristic: look for set codes (usually 3 letters)
	// Event names often have format like "PremierDraft_BLB" where BLB is the set
	parts := strings.Split(eventName, "_")
	if len(parts) > 1 {
		setCode := parts[len(parts)-1]
		if len(setCode) == 3 {
			return setCode
		}
	}
	return "Unknown"
}
