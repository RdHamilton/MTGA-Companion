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
// It looks for RankUpdated events in the log.
func ParseRankUpdates(entries []*LogEntry) ([]*RankUpdate, error) {
	var rankUpdates []*RankUpdate

	for _, entry := range entries {
		if !entry.IsJSON {
			continue
		}

		// Check for RankUpdated event
		if rankData, ok := entry.JSON["RankUpdated"]; ok {
			rankMap, ok := rankData.(map[string]interface{})
			if !ok {
				continue
			}

			update := &RankUpdate{}

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

			// Parse timestamp
			timestamp := time.Now()
			if entry.Timestamp != "" {
				if parsedTime, err := parseLogTimestamp(entry.Timestamp); err == nil {
					timestamp = parsedTime
				}
			}
			update.Timestamp = timestamp

			// Only add if we have the essential data
			if update.NewClass != "" && update.RankUpdateType != "" {
				rankUpdates = append(rankUpdates, update)
			}
		}
	}

	return rankUpdates, nil
}

// FormatToDBFormat converts the MTGA rank update type to database format.
// "Constructed" -> "constructed", "Limited" -> "limited"
func (r *RankUpdate) FormatToDBFormat() string {
	return strings.ToLower(r.RankUpdateType)
}
