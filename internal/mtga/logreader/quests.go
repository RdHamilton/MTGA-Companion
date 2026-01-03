package logreader

import (
	"fmt"
	"log"
	"time"
)

// QuestData represents a quest parsed from MTGA logs.
type QuestData struct {
	QuestID          string
	QuestType        string
	Goal             int
	StartingProgress int
	EndingProgress   int
	CanSwap          bool
	Rewards          string // ChestDescription as JSON string
	AssignedAt       time.Time
	CompletedAt      *time.Time
	LastSeenAt       *time.Time // Tracks when quest was last seen in QuestGetQuests response
	Completed        bool
	Rerolled         bool
}

// ParseQuestsResult contains the results of parsing quests from log entries.
type ParseQuestsResult struct {
	Quests             []*QuestData    // All parsed quests
	HasQuestResponse   bool            // Whether we found any QuestGetQuests responses
	CurrentQuestIDs    map[string]bool // Quest IDs present in the most recent QuestGetQuests response
	LatestResponseTime time.Time       // Timestamp of the most recent QuestGetQuests response
}

// ParseQuests extracts quest data from log entries.
// It looks for QuestGetQuests responses to track quest state and detect completion via disappearance.
func ParseQuests(entries []*LogEntry) ([]*QuestData, error) {
	var quests []*QuestData
	questMap := make(map[string]*QuestData) // Track by questId to detect updates

	questsFound := 0
	responsesFound := 0

	for _, entry := range entries {
		if !entry.IsJSON {
			continue
		}

		// Parse timestamp from log entry for AssignedAt/CompletedAt
		// Many log entries don't have timestamps, so we silently fall back to time.Now()
		logTimestamp := time.Now()
		if entry.Timestamp != "" {
			if parsedTime, err := parseLogTimestamp(entry.Timestamp); err == nil {
				logTimestamp = parsedTime
			}
			// Silently use current time if parsing fails - this is expected for many entry types
		}

		// Check for QuestGetQuests response (contains current active quests)
		if questsData, ok := entry.JSON["quests"]; ok {
			// Accept as QuestGetQuests response if:
			// - canSwap field exists (normal response with quests)
			// - quests array is empty (all quests completed - MTGA returns {"quests":[]} without canSwap)
			isEmptyQuests := false
			if questArray, ok := questsData.([]interface{}); ok && len(questArray) == 0 {
				isEmptyQuests = true
			}
			_, hasCanSwap := entry.JSON["canSwap"]
			if hasCanSwap || isEmptyQuests {
				// This is a QuestGetQuests response
				responsesFound++

				// Track which quest IDs are present in this response
				currentQuestIDs := make(map[string]bool)

				// Use current time for LastSeenAt - this indicates when we processed the quest,
				// not when the log entry was written. This ensures quests appear as "active"
				// when the daemon processes them, even if reading old log entries.
				now := time.Now()

				if questArray, ok := questsData.([]interface{}); ok {
					for _, q := range questArray {
						if questJSON, ok := q.(map[string]interface{}); ok {
							quest := parseQuestFromMap(questJSON, logTimestamp)
							if quest != nil {
								currentQuestIDs[quest.QuestID] = true

								// Update or add quest
								if existing, exists := questMap[quest.QuestID]; exists {
									// Check if this is a reroll - same questId but different quest details
									// MTGA reuses questIds when rerolling, but the QuestType or Goal changes
									if isQuestRerolled(existing, quest) {
										// Mark old quest as rerolled
										existing.Rerolled = true
										log.Printf("Quest parser: Quest %s rerolled - type changed from '%s' to '%s'",
											quest.QuestID, existing.QuestType, quest.QuestType)

										// Create new quest entry with modified key to track both
										quest.LastSeenAt = &now
										newQuestKey := quest.QuestID + "_rerolled_" + logTimestamp.Format("20060102150405")
										questMap[newQuestKey] = quest
										questsFound++
									} else {
										// Same quest, just update progress and last seen timestamp
										existing.EndingProgress = quest.EndingProgress
										existing.CanSwap = quest.CanSwap
										existing.LastSeenAt = &now
									}
								} else {
									// New quest - set last seen to current time
									quest.LastSeenAt = &now
									questMap[quest.QuestID] = quest
									questsFound++
								}
							}
						}
					}
				}

				// Check for quest disappearance (completion detection)
				// If we previously saw a quest but it's not in this response, it was completed
				// Use log timestamp for CompletedAt since it reflects when the action happened
				// Note: Use quest.QuestID (not map key) to check currentQuestIDs because
				// rerolled quests have a different map key but same QuestID
				for _, quest := range questMap {
					if !quest.Completed && !quest.Rerolled && !currentQuestIDs[quest.QuestID] {
						// Quest disappeared - mark as completed
						quest.Completed = true
						completedAt := logTimestamp
						quest.CompletedAt = &completedAt
						// Set progress to goal when completed
						quest.EndingProgress = quest.Goal
						log.Printf("Quest parser: Quest %s completed (disappeared from response)", quest.QuestID)
					}
				}
			}
		}

		// Check for "newQuests" event (newly assigned quests)
		if newQuestsData, ok := entry.JSON["newQuests"]; ok {
			if questArray, ok := newQuestsData.([]interface{}); ok {
				now := time.Now()
				for _, q := range questArray {
					if questJSON, ok := q.(map[string]interface{}); ok {
						quest := parseQuestFromMap(questJSON, logTimestamp)
						if quest != nil {
							quest.LastSeenAt = &now

							// Add or update quest
							if _, exists := questMap[quest.QuestID]; !exists {
								questMap[quest.QuestID] = quest
								questsFound++
							}
						}
					}
				}
			}
		}
	}

	// Convert map to slice
	for _, quest := range questMap {
		quests = append(quests, quest)
	}

	if responsesFound > 0 || questsFound > 0 {
		log.Printf("Quest parser: Found %d QuestGetQuests responses, parsed %d unique quests", responsesFound, questsFound)
	}

	return quests, nil
}

// ParseQuestsDetailed extracts quest data from log entries with detailed information
// about the current quest state. Use this when you need to detect rerolled quests.
func ParseQuestsDetailed(entries []*LogEntry) (*ParseQuestsResult, error) {
	result := &ParseQuestsResult{
		Quests:          []*QuestData{},
		CurrentQuestIDs: make(map[string]bool),
	}

	questMap := make(map[string]*QuestData) // Track by questId to detect updates
	var latestResponseTime time.Time
	var latestResponseQuestIDs map[string]bool

	questsFound := 0
	responsesFound := 0

	for _, entry := range entries {
		if !entry.IsJSON {
			continue
		}

		// Parse timestamp from log entry for AssignedAt/CompletedAt
		// Many log entries don't have timestamps, so we silently fall back to time.Now()
		logTimestamp := time.Now()
		if entry.Timestamp != "" {
			if parsedTime, err := parseLogTimestamp(entry.Timestamp); err == nil {
				logTimestamp = parsedTime
			}
			// Silently use current time if parsing fails - this is expected for many entry types
		}

		// Check for QuestGetQuests response (contains current active quests)
		if questsData, ok := entry.JSON["quests"]; ok {
			// Accept as QuestGetQuests response if:
			// - canSwap field exists (normal response with quests)
			// - quests array is empty (all quests completed - MTGA returns {"quests":[]} without canSwap)
			isEmptyQuests := false
			if questArray, ok := questsData.([]interface{}); ok && len(questArray) == 0 {
				isEmptyQuests = true
			}
			_, hasCanSwap := entry.JSON["canSwap"]
			if hasCanSwap || isEmptyQuests {
				// This is a QuestGetQuests response
				responsesFound++
				result.HasQuestResponse = true

				// Track which quest IDs are present in this response
				currentQuestIDs := make(map[string]bool)

				// Use current time for LastSeenAt - this indicates when we processed the quest,
				// not when the log entry was written. This ensures quests appear as "active"
				// when the daemon processes them, even if reading old log entries.
				now := time.Now()

				if questArray, ok := questsData.([]interface{}); ok {
					for _, q := range questArray {
						if questJSON, ok := q.(map[string]interface{}); ok {
							quest := parseQuestFromMap(questJSON, logTimestamp)
							if quest != nil {
								currentQuestIDs[quest.QuestID] = true

								// Update or add quest
								if existing, exists := questMap[quest.QuestID]; exists {
									// Check if this is a reroll - same questId but different quest details
									// MTGA reuses questIds when rerolling, but the QuestType or Goal changes
									if isQuestRerolled(existing, quest) {
										// Mark old quest as rerolled
										existing.Rerolled = true
										log.Printf("Quest parser: Quest %s rerolled - type changed from '%s' to '%s'",
											quest.QuestID, existing.QuestType, quest.QuestType)

										// Create new quest entry with modified key to track both
										quest.LastSeenAt = &now
										newQuestKey := quest.QuestID + "_rerolled_" + logTimestamp.Format("20060102150405")
										questMap[newQuestKey] = quest
										questsFound++
									} else {
										// Same quest, just update progress and last seen timestamp
										existing.EndingProgress = quest.EndingProgress
										existing.CanSwap = quest.CanSwap
										existing.LastSeenAt = &now
									}
								} else {
									// New quest - set last seen to current time
									quest.LastSeenAt = &now
									questMap[quest.QuestID] = quest
									questsFound++
								}
							}
						}
					}
				}

				// Track the latest response's quest IDs (use log timestamp for historical tracking)
				if logTimestamp.After(latestResponseTime) {
					latestResponseTime = logTimestamp
					latestResponseQuestIDs = currentQuestIDs
				}

				// Check for quest disappearance (completion detection)
				// Use log timestamp for CompletedAt since it reflects when the action happened
				// Note: Use quest.QuestID (not map key) to check currentQuestIDs because
				// rerolled quests have a different map key but same QuestID
				for _, quest := range questMap {
					if !quest.Completed && !quest.Rerolled && !currentQuestIDs[quest.QuestID] {
						quest.Completed = true
						completedAt := logTimestamp
						quest.CompletedAt = &completedAt
						quest.EndingProgress = quest.Goal
						log.Printf("Quest parser: Quest %s completed (disappeared from response)", quest.QuestID)
					}
				}
			}
		}

		// Check for "newQuests" event (newly assigned quests)
		if newQuestsData, ok := entry.JSON["newQuests"]; ok {
			if questArray, ok := newQuestsData.([]interface{}); ok {
				now := time.Now()
				for _, q := range questArray {
					if questJSON, ok := q.(map[string]interface{}); ok {
						quest := parseQuestFromMap(questJSON, logTimestamp)
						if quest != nil {
							quest.LastSeenAt = &now

							if _, exists := questMap[quest.QuestID]; !exists {
								questMap[quest.QuestID] = quest
								questsFound++
							}
						}
					}
				}
			}
		}
	}

	// Convert map to slice
	for _, quest := range questMap {
		result.Quests = append(result.Quests, quest)
	}

	// Set the current quest IDs from the latest response
	if latestResponseQuestIDs != nil {
		result.CurrentQuestIDs = latestResponseQuestIDs
		result.LatestResponseTime = latestResponseTime
	}

	if responsesFound > 0 || questsFound > 0 {
		log.Printf("Quest parser (detailed): Found %d QuestGetQuests responses, parsed %d unique quests, %d current quest IDs",
			responsesFound, questsFound, len(result.CurrentQuestIDs))
		// Log the latest response timestamp and current quest IDs for debugging
		if len(result.CurrentQuestIDs) > 0 {
			log.Printf("Quest parser (detailed): Latest response timestamp: %s", latestResponseTime.Format("2006-01-02 15:04:05 MST"))
			for questID := range result.CurrentQuestIDs {
				// Find the quest to log its type
				for _, q := range result.Quests {
					if q.QuestID == questID && !q.Completed && !q.Rerolled {
						// Truncate quest ID for logging (max 8 chars)
						displayID := q.QuestID
						if len(displayID) > 8 {
							displayID = displayID[:8]
						}
						log.Printf("Quest parser (detailed): Current quest: %s (%s) %d/%d",
							displayID, q.QuestType, q.EndingProgress, q.Goal)
						break
					}
				}
			}
		}
	}

	return result, nil
}

// isQuestRerolled detects if a quest was rerolled by comparing the existing quest
// with the new quest data. MTGA reuses questIds when rerolling, but changes the
// quest type (locKey), goal, or resets starting progress.
func isQuestRerolled(existing, new *QuestData) bool {
	// If the quest type (locKey) changed, it's definitely a reroll
	if existing.QuestType != new.QuestType && existing.QuestType != "" && new.QuestType != "" {
		return true
	}

	// If the goal changed significantly, it's a reroll
	// (Small changes might be MTGA adjustments, but a completely different goal is a reroll)
	if existing.Goal != new.Goal && existing.Goal > 0 && new.Goal > 0 {
		return true
	}

	// If the starting progress was reset while we had progress, it's a reroll
	// This catches cases where the same quest type is rerolled to itself
	if existing.EndingProgress > 0 && new.StartingProgress == 0 && new.EndingProgress == 0 {
		return true
	}

	return false
}

// parseQuestFromMap extracts quest data from a JSON map.
func parseQuestFromMap(json map[string]interface{}, timestamp time.Time) *QuestData {
	quest := &QuestData{
		AssignedAt: timestamp,
	}

	// Extract quest ID
	if questID, ok := json["questId"].(string); ok {
		quest.QuestID = questID
	} else {
		return nil // Quest ID is required
	}

	// Extract quest type (prefer locKey for descriptive name, fallback to questTrack)
	if locKey, ok := json["locKey"].(string); ok {
		quest.QuestType = locKey
	} else if questTrack, ok := json["questTrack"].(string); ok {
		quest.QuestType = questTrack
	}

	// Extract goal
	if goal, ok := json["goal"].(float64); ok {
		quest.Goal = int(goal)
	}

	// Extract starting progress
	if startingProgress, ok := json["startingProgress"].(float64); ok {
		quest.StartingProgress = int(startingProgress)
	}

	// Extract ending progress (current progress)
	if endingProgress, ok := json["endingProgress"].(float64); ok {
		quest.EndingProgress = int(endingProgress)
	}

	// Check if quest can be swapped/rerolled
	if canSwap, ok := json["canSwap"].(bool); ok {
		quest.CanSwap = canSwap
	} else {
		quest.CanSwap = true // Default to true
	}

	// Extract reward description
	if chestDesc, ok := json["chestDescription"].(map[string]interface{}); ok {
		// Extract gold/reward quantity from chestDescription object
		if quantity, ok := chestDesc["quantity"].(string); ok {
			quest.Rewards = quantity
		} else if quantityNum, ok := chestDesc["quantity"].(float64); ok {
			quest.Rewards = fmt.Sprintf("%.0f", quantityNum)
		}
	} else if chestDescStr, ok := json["chestDescription"].(string); ok {
		// Fallback: if it's already a string
		quest.Rewards = chestDescStr
	}

	// Primary completion detection is by quest disappearance from QuestGetQuests responses.
	// As a fallback, if progress >= goal, mark the quest as completed.
	// This helps when we only have one response (e.g., reading from old logs).
	if quest.Goal > 0 && quest.EndingProgress >= quest.Goal {
		quest.Completed = true
		log.Printf("Quest parser: Quest %s marked completed by progress (%d/%d)",
			quest.QuestID, quest.EndingProgress, quest.Goal)
	}

	return quest
}
