package logreader

import (
	"fmt"
	"strings"
	"time"
)

// DraftPick represents a single pick made during a draft.
type DraftPick struct {
	CourseID       string    // CourseId from the draft event
	PackNumber     int       // Pack number (1, 2, 3)
	PickNumber     int       // Pick number within the pack (1-14 or 1-15)
	AvailableCards []int     // Card IDs available in the pack
	SelectedCard   int       // Card ID selected by the player
	Timestamp      time.Time // Timestamp of the pick
}

// DraftPicks represents all picks for a draft event.
type DraftPicks struct {
	CourseID string
	Picks    []DraftPick
}

// ParseDraftPicks extracts individual draft picks from log entries.
// It looks for humanDraftEvent entries containing pack contents and picks.
func ParseDraftPicks(entries []*LogEntry) ([]*DraftPicks, error) {
	var allDraftPicks []*DraftPicks
	picksByCourse := make(map[string]*DraftPicks)

	for _, entry := range entries {
		if !entry.IsJSON {
			continue
		}

		// Check for humanDraftEvent
		if eventData, ok := entry.JSON["humanDraftEvent"]; ok {
			eventMap, ok := eventData.(map[string]interface{})
			if !ok {
				continue
			}

			// Extract course ID
			courseID, _ := eventMap["CourseId"].(string)
			if courseID == "" {
				continue
			}

			// Initialize picks for this course if not exists
			if picksByCourse[courseID] == nil {
				picksByCourse[courseID] = &DraftPicks{
					CourseID: courseID,
					Picks:    []DraftPick{},
				}
			}

			// Parse timestamp
			timestamp := time.Now()
			if entry.Timestamp != "" {
				if parsedTime, err := parseLogTimestamp(entry.Timestamp); err == nil {
					timestamp = parsedTime
				}
			}

			// Extract pack information
			packNumber := 0
			if pack, ok := eventMap["SelfPack"].(float64); ok {
				packNumber = int(pack)
			} else if pack, ok := eventMap["selfPack"].(float64); ok {
				packNumber = int(pack)
			}

			pickNumber := 0
			if pick, ok := eventMap["SelfPick"].(float64); ok {
				pickNumber = int(pick)
			} else if pick, ok := eventMap["selfPick"].(float64); ok {
				pickNumber = int(pick)
			}

			// Extract pack cards
			var availableCards []int
			if packCards, ok := eventMap["PackCards"].([]interface{}); ok {
				for _, cardData := range packCards {
					if cardID, ok := cardData.(float64); ok {
						availableCards = append(availableCards, int(cardID))
					} else if cardID, ok := cardData.(int); ok {
						availableCards = append(availableCards, cardID)
					}
				}
			} else if packCards, ok := eventMap["packCards"].([]interface{}); ok {
				for _, cardData := range packCards {
					if cardID, ok := cardData.(float64); ok {
						availableCards = append(availableCards, int(cardID))
					} else if cardID, ok := cardData.(int); ok {
						availableCards = append(availableCards, cardID)
					}
				}
			}

			// Extract selected card
			selectedCard := 0
			if selected, ok := eventMap["SelectedCard"].(float64); ok {
				selectedCard = int(selected)
			} else if selected, ok := eventMap["selectedCard"].(float64); ok {
				selectedCard = int(selected)
			} else if selected, ok := eventMap["SelectedCardId"].(float64); ok {
				selectedCard = int(selected)
			} else if selected, ok := eventMap["selectedCardId"].(float64); ok {
				selectedCard = int(selected)
			}

			// Create draft pick if we have valid data
			if packNumber > 0 && pickNumber > 0 && selectedCard > 0 {
				pick := DraftPick{
					CourseID:       courseID,
					PackNumber:     packNumber,
					PickNumber:     pickNumber,
					AvailableCards: availableCards,
					SelectedCard:   selectedCard,
					Timestamp:      timestamp,
				}
				picksByCourse[courseID].Picks = append(picksByCourse[courseID].Picks, pick)
			}
		}
	}

	// Convert map to slice
	for _, picks := range picksByCourse {
		if len(picks.Picks) > 0 {
			allDraftPicks = append(allDraftPicks, picks)
		}
	}

	if len(allDraftPicks) == 0 {
		return nil, nil
	}

	return allDraftPicks, nil
}

// parseLogTimestamp attempts to parse a timestamp from the log entry format.
// MTGA log timestamps are in local time (the user's machine timezone).
// We convert to UTC for consistent storage and comparison with query boundaries.
func parseLogTimestamp(timestampStr string) (time.Time, error) {
	// Format: [UnityCrossThreadLogger]2024-01-15 10:30:45
	// Try to extract the date/time portion
	parts := strings.Fields(timestampStr)
	if len(parts) < 2 {
		return time.Time{}, fmt.Errorf("invalid timestamp format")
	}

	// Try common formats
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.000",
	}

	dateTimeStr := parts[len(parts)-2] + " " + parts[len(parts)-1]
	for _, format := range formats {
		if t, err := time.ParseInLocation(format, dateTimeStr, time.Local); err == nil {
			// Convert local time to UTC for consistent storage and comparison
			// This ensures GetDailyWins/GetWeeklyWins queries (which use UTC boundaries)
			// correctly compare against stored timestamps
			return t.UTC(), nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", timestampStr)
}
