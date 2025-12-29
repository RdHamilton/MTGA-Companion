package logreader

import (
	"testing"
	"time"
)

func TestParseQuests_LastSeenAtUsesCurrentTime(t *testing.T) {
	// Create a log entry - the timestamp doesn't matter because LastSeenAt
	// should always use time.Now() when processing quests
	entries := []*LogEntry{
		{
			IsJSON:    true,
			Timestamp: "[UnityCrossThreadLogger]2024-01-15 10:30:45",
			JSON: map[string]interface{}{
				"quests": []interface{}{
					map[string]interface{}{
						"questId":        "quest-123",
						"locKey":         "Win 4 games",
						"goal":           float64(4),
						"canSwap":        true,
						"endingProgress": float64(2),
					},
				},
				"canSwap": true, // Indicates QuestGetQuests response
			},
		},
	}

	beforeParse := time.Now()
	quests, err := ParseQuests(entries)
	afterParse := time.Now()

	if err != nil {
		t.Fatalf("ParseQuests returned error: %v", err)
	}

	if len(quests) != 1 {
		t.Fatalf("Expected 1 quest, got %d", len(quests))
	}

	quest := quests[0]

	// The main fix: LastSeenAt should be set to current time (within test execution window)
	// This ensures quests appear as "active" even when reading old log entries
	if quest.LastSeenAt == nil {
		t.Fatal("LastSeenAt should not be nil")
	}

	if quest.LastSeenAt.Before(beforeParse) || quest.LastSeenAt.After(afterParse) {
		t.Errorf("LastSeenAt should be between %v and %v, got %v",
			beforeParse, afterParse, *quest.LastSeenAt)
	}

	// Verify quest data was parsed correctly
	if quest.QuestID != "quest-123" {
		t.Errorf("Expected QuestID 'quest-123', got %s", quest.QuestID)
	}
	if quest.Goal != 4 {
		t.Errorf("Expected Goal 4, got %d", quest.Goal)
	}
	if quest.EndingProgress != 2 {
		t.Errorf("Expected EndingProgress 2, got %d", quest.EndingProgress)
	}
}

func TestParseQuestsDetailed_LastSeenAtUsesCurrentTime(t *testing.T) {
	entries := []*LogEntry{
		{
			IsJSON:    true,
			Timestamp: "[UnityCrossThreadLogger]2024-01-15 10:30:45",
			JSON: map[string]interface{}{
				"quests": []interface{}{
					map[string]interface{}{
						"questId":        "quest-456",
						"locKey":         "Cast 20 spells",
						"goal":           float64(20),
						"canSwap":        true,
						"endingProgress": float64(10),
					},
				},
				"canSwap": true,
			},
		},
	}

	beforeParse := time.Now()
	result, err := ParseQuestsDetailed(entries)
	afterParse := time.Now()

	if err != nil {
		t.Fatalf("ParseQuestsDetailed returned error: %v", err)
	}

	if len(result.Quests) != 1 {
		t.Fatalf("Expected 1 quest, got %d", len(result.Quests))
	}

	quest := result.Quests[0]

	// The main fix: LastSeenAt should be set to current time
	if quest.LastSeenAt == nil {
		t.Fatal("LastSeenAt should not be nil")
	}

	if quest.LastSeenAt.Before(beforeParse) || quest.LastSeenAt.After(afterParse) {
		t.Errorf("LastSeenAt should be between %v and %v, got %v",
			beforeParse, afterParse, *quest.LastSeenAt)
	}

	// Verify HasQuestResponse flag is set
	if !result.HasQuestResponse {
		t.Error("HasQuestResponse should be true")
	}

	// Verify CurrentQuestIDs is populated
	if !result.CurrentQuestIDs["quest-456"] {
		t.Error("CurrentQuestIDs should contain 'quest-456'")
	}
}

func TestParseQuests_NewQuestsEvent_LastSeenAtUsesCurrentTime(t *testing.T) {
	// Test newQuests event - when a new quest is assigned
	entries := []*LogEntry{
		{
			IsJSON:    true,
			Timestamp: "[UnityCrossThreadLogger]2024-01-15 10:30:45",
			JSON: map[string]interface{}{
				"newQuests": []interface{}{
					map[string]interface{}{
						"questId": "quest-789",
						"locKey":  "Play 3 lands",
						"goal":    float64(3),
						"canSwap": true,
					},
				},
			},
		},
	}

	beforeParse := time.Now()
	quests, err := ParseQuests(entries)
	afterParse := time.Now()

	if err != nil {
		t.Fatalf("ParseQuests returned error: %v", err)
	}

	if len(quests) != 1 {
		t.Fatalf("Expected 1 quest, got %d", len(quests))
	}

	quest := quests[0]

	// The main fix: LastSeenAt should be set to current time even for newQuests events
	if quest.LastSeenAt == nil {
		t.Fatal("LastSeenAt should not be nil")
	}

	if quest.LastSeenAt.Before(beforeParse) || quest.LastSeenAt.After(afterParse) {
		t.Errorf("LastSeenAt should be between %v and %v, got %v",
			beforeParse, afterParse, *quest.LastSeenAt)
	}
}

func TestParseQuests_UpdateExistingQuest_LastSeenAtUpdated(t *testing.T) {
	// Test that when a quest is seen again (progress updated), LastSeenAt is updated
	entries := []*LogEntry{
		{
			IsJSON:    true,
			Timestamp: "[UnityCrossThreadLogger]2024-01-15 10:30:45",
			JSON: map[string]interface{}{
				"quests": []interface{}{
					map[string]interface{}{
						"questId":        "quest-progress",
						"locKey":         "Win 5 games",
						"goal":           float64(5),
						"canSwap":        true,
						"endingProgress": float64(1),
					},
				},
				"canSwap": true,
			},
		},
		{
			IsJSON:    true,
			Timestamp: "[UnityCrossThreadLogger]2024-01-15 11:30:45",
			JSON: map[string]interface{}{
				"quests": []interface{}{
					map[string]interface{}{
						"questId":        "quest-progress",
						"locKey":         "Win 5 games",
						"goal":           float64(5),
						"canSwap":        true,
						"endingProgress": float64(3), // Progress updated
					},
				},
				"canSwap": true,
			},
		},
	}

	beforeParse := time.Now()
	quests, err := ParseQuests(entries)
	afterParse := time.Now()

	if err != nil {
		t.Fatalf("ParseQuests returned error: %v", err)
	}

	if len(quests) != 1 {
		t.Fatalf("Expected 1 quest, got %d", len(quests))
	}

	quest := quests[0]

	// Progress should be updated to latest value
	if quest.EndingProgress != 3 {
		t.Errorf("Expected EndingProgress 3, got %d", quest.EndingProgress)
	}

	// LastSeenAt should be current time (updated on second entry)
	if quest.LastSeenAt == nil {
		t.Fatal("LastSeenAt should not be nil")
	}

	if quest.LastSeenAt.Before(beforeParse) || quest.LastSeenAt.After(afterParse) {
		t.Errorf("LastSeenAt should be between %v and %v, got %v",
			beforeParse, afterParse, *quest.LastSeenAt)
	}
}

func TestParseQuests_MultipleQuests(t *testing.T) {
	entries := []*LogEntry{
		{
			IsJSON:    true,
			Timestamp: "[UnityCrossThreadLogger]2024-01-15 10:30:45",
			JSON: map[string]interface{}{
				"quests": []interface{}{
					map[string]interface{}{
						"questId":        "quest-1",
						"locKey":         "Win 2 games",
						"goal":           float64(2),
						"canSwap":        true,
						"endingProgress": float64(0),
					},
					map[string]interface{}{
						"questId":        "quest-2",
						"locKey":         "Cast 10 spells",
						"goal":           float64(10),
						"canSwap":        false,
						"endingProgress": float64(5),
					},
					map[string]interface{}{
						"questId":        "quest-3",
						"locKey":         "Play 5 lands",
						"goal":           float64(5),
						"canSwap":        true,
						"endingProgress": float64(3),
					},
				},
				"canSwap": true,
			},
		},
	}

	beforeParse := time.Now()
	quests, err := ParseQuests(entries)
	afterParse := time.Now()

	if err != nil {
		t.Fatalf("ParseQuests returned error: %v", err)
	}

	if len(quests) != 3 {
		t.Fatalf("Expected 3 quests, got %d", len(quests))
	}

	// All quests should have LastSeenAt set to current time
	for _, quest := range quests {
		if quest.LastSeenAt == nil {
			t.Errorf("Quest %s: LastSeenAt should not be nil", quest.QuestID)
			continue
		}

		if quest.LastSeenAt.Before(beforeParse) || quest.LastSeenAt.After(afterParse) {
			t.Errorf("Quest %s: LastSeenAt should be between %v and %v, got %v",
				quest.QuestID, beforeParse, afterParse, *quest.LastSeenAt)
		}
	}
}

func TestParseQuests_QuestCompletion(t *testing.T) {
	// First response has the quest, second response doesn't - quest was completed
	entries := []*LogEntry{
		{
			IsJSON:    true,
			Timestamp: "[UnityCrossThreadLogger]2024-01-15 10:30:45",
			JSON: map[string]interface{}{
				"quests": []interface{}{
					map[string]interface{}{
						"questId":        "quest-complete",
						"locKey":         "Win 2 games",
						"goal":           float64(2),
						"canSwap":        true,
						"endingProgress": float64(1),
					},
				},
				"canSwap": true,
			},
		},
		{
			IsJSON:    true,
			Timestamp: "[UnityCrossThreadLogger]2024-01-15 11:30:45",
			JSON: map[string]interface{}{
				"quests":  []interface{}{}, // Quest disappeared - completed
				"canSwap": true,
			},
		},
	}

	quests, err := ParseQuests(entries)
	if err != nil {
		t.Fatalf("ParseQuests returned error: %v", err)
	}

	if len(quests) != 1 {
		t.Fatalf("Expected 1 quest, got %d", len(quests))
	}

	quest := quests[0]

	// Verify the quest is marked as completed
	if !quest.Completed {
		t.Error("Quest should be marked as completed")
	}

	// CompletedAt should be set
	if quest.CompletedAt == nil {
		t.Error("CompletedAt should not be nil for completed quest")
	}

	// EndingProgress should be set to goal
	if quest.EndingProgress != quest.Goal {
		t.Errorf("EndingProgress should equal Goal (%d), got %d", quest.Goal, quest.EndingProgress)
	}
}
