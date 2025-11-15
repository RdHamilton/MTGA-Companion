package logreader

import (
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
			if questArray, ok := questsData.([]interface{}); ok {
				for _, q := range questArray {
					if questMap, ok := q.(map[string]interface{}); ok {
						quest := parseQuestFromMap(questMap, timestamp)
						if quest != nil {
							questMap[quest.QuestID] = quest
						}
					}
				}
			}
		}

		// Check for "newQuests" event (newly assigned quests)
		if newQuestsData, ok := entry.JSON["newQuests"]; ok {
			if questArray, ok := newQuestsData.([]interface{}); ok {
				for _, q := range questArray {
					if questMapData, ok := q.(map[string]interface{}); ok {
						quest := parseQuestFromMap(questMapData, timestamp)
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

	// Extract quest type (from questTrack or locKey)
	if questTrack, ok := json["questTrack"].(string); ok {
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
	if chestDesc, ok := json["chestDescription"].(string); ok {
		quest.Rewards = chestDesc
	}

	// Check if completed (when ending progress >= goal)
	if quest.Goal > 0 && quest.EndingProgress >= quest.Goal {
		quest.Completed = true
		completedAt := timestamp
		quest.CompletedAt = &completedAt
	}

	return quest
}
