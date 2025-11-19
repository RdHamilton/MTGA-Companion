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

// ParseQuests extracts quest data from log entries.
// It looks for QuestGetQuests responses to track quest state and detect completion via disappearance.
func ParseQuests(entries []*LogEntry) ([]*QuestData, error) {
	var quests []*QuestData
	questMap := make(map[string]*QuestData)        // Track by questId to detect updates
	lastSeenQuestIDs := make(map[string]time.Time) // Track when each quest was last seen

	questsFound := 0
	responsesFound := 0

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

		// Check for QuestGetQuests response (contains current active quests)
		if questsData, ok := entry.JSON["quests"]; ok {
			if _, hasCanSwap := entry.JSON["canSwap"]; hasCanSwap {
				// This is a QuestGetQuests response
				responsesFound++

				// Track which quest IDs are present in this response
				currentQuestIDs := make(map[string]bool)

				if questArray, ok := questsData.([]interface{}); ok {
					for _, q := range questArray {
						if questJSON, ok := q.(map[string]interface{}); ok {
							quest := parseQuestFromMap(questJSON, timestamp)
							if quest != nil {
								currentQuestIDs[quest.QuestID] = true
								lastSeenQuestIDs[quest.QuestID] = timestamp

								// Update or add quest
								if existing, exists := questMap[quest.QuestID]; exists {
									// Update existing quest progress and last seen timestamp
									existing.EndingProgress = quest.EndingProgress
									existing.CanSwap = quest.CanSwap
									existing.LastSeenAt = &timestamp
								} else {
									// New quest - set last seen to current timestamp
									quest.LastSeenAt = &timestamp
									questMap[quest.QuestID] = quest
									questsFound++
								}
							}
						}
					}
				}

				// Check for quest disappearance (completion detection)
				// If we previously saw a quest but it's not in this response, it was completed
				for questID, quest := range questMap {
					if !quest.Completed && !currentQuestIDs[questID] {
						// Quest disappeared - mark as completed
						quest.Completed = true
						quest.CompletedAt = &timestamp
						// Set progress to goal when completed
						quest.EndingProgress = quest.Goal
						log.Printf("Quest parser: Quest %s completed (disappeared from response)", questID)
					}
				}
			}
		}

		// Check for "newQuests" event (newly assigned quests)
		if newQuestsData, ok := entry.JSON["newQuests"]; ok {
			if questArray, ok := newQuestsData.([]interface{}); ok {
				for _, q := range questArray {
					if questJSON, ok := q.(map[string]interface{}); ok {
						quest := parseQuestFromMap(questJSON, timestamp)
						if quest != nil {
							quest.AssignedAt = timestamp
							lastSeenQuestIDs[quest.QuestID] = timestamp

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

	// Note: Completion is detected by quest disappearance from QuestGetQuests responses,
	// not by checking progress >= goal. MTGA removes quests from the response when completed.

	return quest
}
