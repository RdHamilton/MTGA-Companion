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

func TestParseQuestsDetailed_RerollDetection(t *testing.T) {
	// First response has quest-A, second response has quest-B (quest-A was rerolled)
	entries := []*LogEntry{
		{
			IsJSON:    true,
			Timestamp: "[UnityCrossThreadLogger]2024-01-15 10:30:45",
			JSON: map[string]interface{}{
				"quests": []interface{}{
					map[string]interface{}{
						"questId":        "quest-A",
						"locKey":         "Win 2 games",
						"goal":           float64(2),
						"canSwap":        true,
						"endingProgress": float64(0),
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
						"questId":        "quest-B", // Different quest - quest-A was rerolled
						"locKey":         "Cast 10 spells",
						"goal":           float64(10),
						"canSwap":        true,
						"endingProgress": float64(0),
					},
				},
				"canSwap": true,
			},
		},
	}

	result, err := ParseQuestsDetailed(entries)
	if err != nil {
		t.Fatalf("ParseQuestsDetailed returned error: %v", err)
	}

	// Should have 2 quests: one completed (quest-A) and one active (quest-B)
	if len(result.Quests) != 2 {
		t.Fatalf("Expected 2 quests, got %d", len(result.Quests))
	}

	// Find quest-A and quest-B
	var questA, questB *QuestData
	for _, q := range result.Quests {
		switch q.QuestID {
		case "quest-A":
			questA = q
		case "quest-B":
			questB = q
		}
	}

	if questA == nil {
		t.Fatal("quest-A not found")
	}
	if questB == nil {
		t.Fatal("quest-B not found")
	}

	// Quest-A should be marked as completed (disappeared from response)
	if !questA.Completed {
		t.Error("quest-A should be marked as completed when it disappears")
	}
	if questA.CompletedAt == nil {
		t.Error("quest-A CompletedAt should be set when quest disappears")
	}

	// Quest-B should be active
	if questB.Completed {
		t.Error("quest-B should not be marked as completed")
	}

	// CurrentQuestIDs should only contain quest-B (from the latest response)
	if result.CurrentQuestIDs["quest-A"] {
		t.Error("CurrentQuestIDs should NOT contain quest-A")
	}
	if !result.CurrentQuestIDs["quest-B"] {
		t.Error("CurrentQuestIDs should contain quest-B")
	}
}

func TestParseQuests_RerollDetection_SameQuestId_DifferentDetails(t *testing.T) {
	// Simulate a reroll where MTGA reuses the same questId but changes the quest details
	// Entry 1: Quest "Win 4 games" with progress
	// Entry 2: Same questId but now "Cast 20 spells" (rerolled) - different locKey and goal
	entries := []*LogEntry{
		{
			IsJSON:    true,
			Timestamp: "[UnityCrossThreadLogger]2024-01-15 10:30:45",
			JSON: map[string]interface{}{
				"quests": []interface{}{
					map[string]interface{}{
						"questId":          "quest-123",
						"locKey":           "Win 4 games",
						"goal":             float64(4),
						"canSwap":          true,
						"startingProgress": float64(0),
						"endingProgress":   float64(2), // Has progress
					},
				},
				"canSwap": true,
			},
		},
		{
			IsJSON:    true,
			Timestamp: "[UnityCrossThreadLogger]2024-01-15 11:00:00",
			JSON: map[string]interface{}{
				"quests": []interface{}{
					map[string]interface{}{
						"questId":          "quest-123",      // Same questId!
						"locKey":           "Cast 20 spells", // Different quest type
						"goal":             float64(20),      // Different goal
						"canSwap":          true,
						"startingProgress": float64(0), // Reset to 0
						"endingProgress":   float64(0), // No progress yet
					},
				},
				"canSwap": true,
			},
		},
	}

	quests, err := ParseQuests(entries)
	if err != nil {
		t.Fatalf("ParseQuests returned error: %v", err)
	}

	// Should have 2 quests: the old one marked as rerolled, and the new one
	if len(quests) != 2 {
		t.Fatalf("Expected 2 quests (old rerolled + new), got %d", len(quests))
	}

	// Find the original and new quest
	var originalQuest, newQuest *QuestData
	for _, q := range quests {
		if q.QuestType == "Win 4 games" {
			originalQuest = q
		} else if q.QuestType == "Cast 20 spells" {
			newQuest = q
		}
	}

	if originalQuest == nil {
		t.Fatal("Original quest 'Win 4 games' not found")
	}
	if newQuest == nil {
		t.Fatal("New quest 'Cast 20 spells' not found")
	}

	// Original quest should be marked as rerolled
	if !originalQuest.Rerolled {
		t.Error("Original quest should be marked as rerolled")
	}

	// Original quest should NOT be marked as completed (it was rerolled, not completed)
	if originalQuest.Completed {
		t.Error("Original quest should NOT be marked as completed")
	}

	// Original quest should have its original progress preserved
	if originalQuest.EndingProgress != 2 {
		t.Errorf("Original quest should have EndingProgress 2, got %d", originalQuest.EndingProgress)
	}

	// New quest should NOT be marked as rerolled or completed
	if newQuest.Rerolled {
		t.Error("New quest should NOT be marked as rerolled")
	}
	if newQuest.Completed {
		t.Error("New quest should NOT be marked as completed")
	}

	// New quest should have the new details
	if newQuest.Goal != 20 {
		t.Errorf("New quest should have Goal 20, got %d", newQuest.Goal)
	}
	if newQuest.EndingProgress != 0 {
		t.Errorf("New quest should have EndingProgress 0, got %d", newQuest.EndingProgress)
	}
}

func TestIsQuestRerolled(t *testing.T) {
	tests := []struct {
		name     string
		existing *QuestData
		new      *QuestData
		expected bool
	}{
		{
			name: "same quest, progress updated - not rerolled",
			existing: &QuestData{
				QuestID:        "q1",
				QuestType:      "Win 4 games",
				Goal:           4,
				EndingProgress: 2,
			},
			new: &QuestData{
				QuestID:        "q1",
				QuestType:      "Win 4 games",
				Goal:           4,
				EndingProgress: 3,
			},
			expected: false,
		},
		{
			name: "different quest type - rerolled",
			existing: &QuestData{
				QuestID:        "q1",
				QuestType:      "Win 4 games",
				Goal:           4,
				EndingProgress: 2,
			},
			new: &QuestData{
				QuestID:          "q1",
				QuestType:        "Cast 20 spells",
				Goal:             20,
				StartingProgress: 0,
				EndingProgress:   0,
			},
			expected: true,
		},
		{
			name: "same quest type but different goal - rerolled",
			existing: &QuestData{
				QuestID:        "q1",
				QuestType:      "Win X games",
				Goal:           4,
				EndingProgress: 2,
			},
			new: &QuestData{
				QuestID:          "q1",
				QuestType:        "Win X games",
				Goal:             10, // Different goal
				StartingProgress: 0,
				EndingProgress:   0,
			},
			expected: true,
		},
		{
			name: "same quest type, progress reset - rerolled",
			existing: &QuestData{
				QuestID:        "q1",
				QuestType:      "Win 4 games",
				Goal:           4,
				EndingProgress: 3, // Had progress
			},
			new: &QuestData{
				QuestID:          "q1",
				QuestType:        "Win 4 games",
				Goal:             4,
				StartingProgress: 0,
				EndingProgress:   0, // Progress reset
			},
			expected: true,
		},
		{
			name: "new quest with no existing progress - not rerolled",
			existing: &QuestData{
				QuestID:        "q1",
				QuestType:      "Win 4 games",
				Goal:           4,
				EndingProgress: 0, // No progress yet
			},
			new: &QuestData{
				QuestID:          "q1",
				QuestType:        "Win 4 games",
				Goal:             4,
				StartingProgress: 0,
				EndingProgress:   0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isQuestRerolled(tt.existing, tt.new)
			if result != tt.expected {
				t.Errorf("isQuestRerolled() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
