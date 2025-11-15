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
	Completed        bool
	Rerolled         bool
}

// ParseQuests extracts quest data from log entries.
// It looks for "quests" and "newQuests" events in the MTGA logs.
func ParseQuests(entries []*LogEntry) ([]*QuestData, error) {
	var quests []*QuestData
	questMap := make(map[string]*QuestData) // Track by questId to detect updates

	questsFound := 0
	newQuestsFound := 0

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

		// Check for "quests" event (quest state/progress updates)
		if questsData, ok := entry.JSON["quests"]; ok {
			questsFound++
			if questArray, ok := questsData.([]interface{}); ok {
				for _, q := range questArray {
					if questJSON, ok := q.(map[string]interface{}); ok {
						quest := parseQuestFromMap(questJSON, timestamp)
						if quest != nil {
							questMap[quest.QuestID] = quest
						}
					}
				}
			}
		}

		// Check for "newQuests" event (newly assigned quests)
		if newQuestsData, ok := entry.JSON["newQuests"]; ok {
			newQuestsFound++
			if questArray, ok := newQuestsData.([]interface{}); ok {
				for _, q := range questArray {
					if questJSON, ok := q.(map[string]interface{}); ok {
						quest := parseQuestFromMap(questJSON, timestamp)
						if quest != nil {
							quest.AssignedAt = timestamp
							questMap[quest.QuestID] = quest
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

	if questsFound > 0 || newQuestsFound > 0 {
		log.Printf("Quest parser: Found %d 'quests' events and %d 'newQuests' events, parsed %d unique quests", questsFound, newQuestsFound, len(quests))
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

	// Check if completed (when ending progress >= goal)
	if quest.Goal > 0 && quest.EndingProgress >= quest.Goal {
		quest.Completed = true
		completedAt := timestamp
		quest.CompletedAt = &completedAt
	}

	return quest
}
