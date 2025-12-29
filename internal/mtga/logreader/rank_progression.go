package logreader

import (
	"strings"
	"time"
)

// RankUpdate represents a rank change event from MTGA logs.
type RankUpdate struct {
	PlayerID         string
	SeasonOrdinal    int
	NewClass         string
	OldClass         string
	NewLevel         int
	OldLevel         int
	NewStep          int
	OldStep          int
	WasLossProtected bool
	RankUpdateType   string // "Constructed" or "Limited"
	Timestamp        time.Time
}

// ParseRankUpdates extracts rank progression data from log entries.
// It looks for rank data in the current MTGA log format (constructedClass, limitedClass, etc.).
// Note: Older logs used RankUpdated events, but current logs use individual rank fields.
func ParseRankUpdates(entries []*LogEntry) ([]*RankUpdate, error) {
	var rankUpdates []*RankUpdate

	for _, entry := range entries {
		if !entry.IsJSON {
			continue
		}

		// Parse timestamp
		timestamp := time.Now()
		if entry.Timestamp != "" {
			if parsedTime, err := parseLogTimestamp(entry.Timestamp); err == nil {
				timestamp = parsedTime
			}
		}

		// Try current MTGA log format first (individual rank fields)
		// Check for constructed rank data
		if hasConstructedRank(entry.JSON) {
			update := parseConstructedRank(entry.JSON, timestamp)
			if update != nil {
				rankUpdates = append(rankUpdates, update)
			}
		}

		// Check for limited rank data
		if hasLimitedRank(entry.JSON) {
			update := parseLimitedRank(entry.JSON, timestamp)
			if update != nil {
				rankUpdates = append(rankUpdates, update)
			}
		}

		// Legacy support: Check for old RankUpdated event format
		if rankData, ok := entry.JSON["RankUpdated"]; ok {
			rankMap, ok := rankData.(map[string]interface{})
			if !ok {
				continue
			}

			update := parseLegacyRankUpdate(rankMap, timestamp)
			if update != nil {
				rankUpdates = append(rankUpdates, update)
			}
		}
	}

	return rankUpdates, nil
}

// hasConstructedRank checks if the entry contains constructed rank data
func hasConstructedRank(json map[string]interface{}) bool {
	_, hasClass := json["constructedClass"]
	_, hasLevel := json["constructedLevel"]
	_, hasSeason := json["constructedSeasonOrdinal"]
	return (hasClass || hasLevel) && hasSeason
}

// hasLimitedRank checks if the entry contains limited rank data
func hasLimitedRank(json map[string]interface{}) bool {
	_, hasClass := json["limitedClass"]
	_, hasLevel := json["limitedLevel"]
	_, hasSeason := json["limitedSeasonOrdinal"]
	return (hasClass || hasLevel) && hasSeason
}

// parseConstructedRank extracts constructed rank from current log format
func parseConstructedRank(json map[string]interface{}, timestamp time.Time) *RankUpdate {
	update := &RankUpdate{
		RankUpdateType: "Constructed",
		Timestamp:      timestamp,
	}

	// Extract season ordinal
	if seasonOrdinal, ok := json["constructedSeasonOrdinal"].(float64); ok {
		update.SeasonOrdinal = int(seasonOrdinal)
	}

	// Extract rank class
	if rankClass, ok := json["constructedClass"].(string); ok {
		update.NewClass = rankClass
		update.OldClass = rankClass // No old value in current format
	}

	// Extract rank level
	if rankLevel, ok := json["constructedLevel"].(float64); ok {
		update.NewLevel = int(rankLevel)
		update.OldLevel = int(rankLevel) // No old value in current format
	}

	// Extract rank step (if present)
	if rankStep, ok := json["constructedStep"].(float64); ok {
		update.NewStep = int(rankStep)
		update.OldStep = int(rankStep) // No old value in current format
	}

	// Only add if we have essential data - require rank class to prevent "Unranked" entries
	if update.SeasonOrdinal > 0 && update.NewClass != "" {
		return update
	}

	return nil
}

// parseLimitedRank extracts limited rank from current log format
func parseLimitedRank(json map[string]interface{}, timestamp time.Time) *RankUpdate {
	update := &RankUpdate{
		RankUpdateType: "Limited",
		Timestamp:      timestamp,
	}

	// Extract season ordinal
	if seasonOrdinal, ok := json["limitedSeasonOrdinal"].(float64); ok {
		update.SeasonOrdinal = int(seasonOrdinal)
	}

	// Extract rank class
	if rankClass, ok := json["limitedClass"].(string); ok {
		update.NewClass = rankClass
		update.OldClass = rankClass // No old value in current format
	}

	// Extract rank level
	if rankLevel, ok := json["limitedLevel"].(float64); ok {
		update.NewLevel = int(rankLevel)
		update.OldLevel = int(rankLevel) // No old value in current format
	}

	// Extract rank step (if present)
	if rankStep, ok := json["limitedStep"].(float64); ok {
		update.NewStep = int(rankStep)
		update.OldStep = int(rankStep) // No old value in current format
	}

	// Only add if we have essential data - require rank class to prevent "Unranked" entries
	if update.SeasonOrdinal > 0 && update.NewClass != "" {
		return update
	}

	return nil
}

// parseLegacyRankUpdate parses old RankUpdated event format for backwards compatibility
func parseLegacyRankUpdate(rankMap map[string]interface{}, timestamp time.Time) *RankUpdate {
	update := &RankUpdate{Timestamp: timestamp}

	// Extract player ID
	if playerID, ok := rankMap["playerId"].(string); ok {
		update.PlayerID = playerID
	}

	// Extract season ordinal
	if seasonOrdinal, ok := rankMap["seasonOrdinal"].(float64); ok {
		update.SeasonOrdinal = int(seasonOrdinal)
	}

	// Extract rank classes
	if newClass, ok := rankMap["newClass"].(string); ok {
		update.NewClass = newClass
	}
	if oldClass, ok := rankMap["oldClass"].(string); ok {
		update.OldClass = oldClass
	}

	// Extract rank levels
	if newLevel, ok := rankMap["newLevel"].(float64); ok {
		update.NewLevel = int(newLevel)
	}
	if oldLevel, ok := rankMap["oldLevel"].(float64); ok {
		update.OldLevel = int(oldLevel)
	}

	// Extract rank steps
	if newStep, ok := rankMap["newStep"].(float64); ok {
		update.NewStep = int(newStep)
	}
	if oldStep, ok := rankMap["oldStep"].(float64); ok {
		update.OldStep = int(oldStep)
	}

	// Extract loss protection status
	if lossProtected, ok := rankMap["wasLossProtected"].(bool); ok {
		update.WasLossProtected = lossProtected
	}

	// Extract rank update type (Constructed or Limited)
	if rankType, ok := rankMap["rankUpdateType"].(string); ok {
		update.RankUpdateType = rankType
	}

	// Only add if we have the essential data
	if update.NewClass != "" && update.RankUpdateType != "" {
		return update
	}

	return nil
}

// FormatToDBFormat converts the MTGA rank update type to database format.
// "Constructed" -> "constructed", "Limited" -> "limited"
func (r *RankUpdate) FormatToDBFormat() string {
	return strings.ToLower(r.RankUpdateType)
}
