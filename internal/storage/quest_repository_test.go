package storage

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// setupTestDB creates an in-memory SQLite database with the quests table
func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create quests table
	_, err = db.Exec(`
		CREATE TABLE quests (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			quest_id TEXT NOT NULL,
			quest_type TEXT,
			goal INTEGER DEFAULT 0,
			starting_progress INTEGER DEFAULT 0,
			ending_progress INTEGER DEFAULT 0,
			completed INTEGER DEFAULT 0,
			can_swap INTEGER DEFAULT 1,
			rewards TEXT,
			assigned_at TEXT NOT NULL,
			completed_at TEXT,
			last_seen_at TEXT,
			rerolled INTEGER DEFAULT 0,
			created_at TEXT DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create quests table: %v", err)
	}

	return db
}

func TestParseGoldFromRewards(t *testing.T) {
	tests := []struct {
		name     string
		rewards  string
		expected int
	}{
		{
			name:     "standard 500 gold",
			rewards:  "500",
			expected: 500,
		},
		{
			name:     "750 gold quest",
			rewards:  "750",
			expected: 750,
		},
		{
			name:     "empty string defaults to 500",
			rewards:  "",
			expected: 500,
		},
		{
			name:     "whitespace only defaults to 500",
			rewards:  "   ",
			expected: 500,
		},
		{
			name:     "invalid string defaults to 500",
			rewards:  "invalid",
			expected: 500,
		},
		{
			name:     "negative number defaults to 500",
			rewards:  "-100",
			expected: 500,
		},
		{
			name:     "zero defaults to 500",
			rewards:  "0",
			expected: 500,
		},
		{
			name:     "number with whitespace",
			rewards:  " 750 ",
			expected: 750,
		},
		{
			name:     "1000 gold",
			rewards:  "1000",
			expected: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseGoldFromRewards(tt.rewards)
			if result != tt.expected {
				t.Errorf("parseGoldFromRewards(%q) = %d, want %d", tt.rewards, result, tt.expected)
			}
		})
	}
}

func TestCalculateTotalGoldEarned(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewQuestRepository(db)

	// Insert test quests
	now := time.Now().UTC()
	assignedAt := now.Add(-24 * time.Hour).Format("2006-01-02 15:04:05")
	completedAt := now.Format("2006-01-02 15:04:05")
	createdAt := now.Format("2006-01-02 15:04:05")

	// Quest 1: 500 gold, completed
	_, err := db.Exec(`
		INSERT INTO quests (quest_id, quest_type, goal, ending_progress, completed, rewards, assigned_at, completed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "quest-1", "Daily", 5, 5, 1, "500", assignedAt, completedAt, createdAt)
	if err != nil {
		t.Fatalf("Failed to insert quest 1: %v", err)
	}

	// Quest 2: 750 gold, completed
	_, err = db.Exec(`
		INSERT INTO quests (quest_id, quest_type, goal, ending_progress, completed, rewards, assigned_at, completed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "quest-2", "Daily", 10, 10, 1, "750", assignedAt, completedAt, createdAt)
	if err != nil {
		t.Fatalf("Failed to insert quest 2: %v", err)
	}

	// Quest 3: Not completed (should not be counted)
	_, err = db.Exec(`
		INSERT INTO quests (quest_id, quest_type, goal, ending_progress, completed, rewards, assigned_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, "quest-3", "Daily", 5, 2, 0, "500", assignedAt, createdAt)
	if err != nil {
		t.Fatalf("Failed to insert quest 3: %v", err)
	}

	// Quest 4: Empty rewards, completed (should default to 500)
	_, err = db.Exec(`
		INSERT INTO quests (quest_id, quest_type, goal, ending_progress, completed, rewards, assigned_at, completed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "quest-4", "Daily", 3, 3, 1, "", assignedAt, completedAt, createdAt)
	if err != nil {
		t.Fatalf("Failed to insert quest 4: %v", err)
	}

	// Calculate total: 500 + 750 + 500 (default) = 1750
	total := repo.calculateTotalGoldEarned(nil, nil)
	expected := 1750

	if total != expected {
		t.Errorf("calculateTotalGoldEarned() = %d, want %d", total, expected)
	}
}

func TestCalculateTotalGoldEarnedWithDateFilter(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewQuestRepository(db)

	// Insert quests at different dates
	oldDate := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	recentDate := time.Date(2024, 11, 15, 12, 0, 0, 0, time.UTC)

	// Old quest: 500 gold
	_, err := db.Exec(`
		INSERT INTO quests (quest_id, quest_type, goal, ending_progress, completed, rewards, assigned_at, completed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "old-quest", "Daily", 5, 5, 1, "500",
		oldDate.Add(-24*time.Hour).Format("2006-01-02 15:04:05"),
		oldDate.Format("2006-01-02 15:04:05"),
		oldDate.Format("2006-01-02 15:04:05"))
	if err != nil {
		t.Fatalf("Failed to insert old quest: %v", err)
	}

	// Recent quest: 750 gold
	_, err = db.Exec(`
		INSERT INTO quests (quest_id, quest_type, goal, ending_progress, completed, rewards, assigned_at, completed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "recent-quest", "Daily", 5, 5, 1, "750",
		recentDate.Add(-24*time.Hour).Format("2006-01-02 15:04:05"),
		recentDate.Format("2006-01-02 15:04:05"),
		recentDate.Format("2006-01-02 15:04:05"))
	if err != nil {
		t.Fatalf("Failed to insert recent quest: %v", err)
	}

	// Test with date filter that only includes recent quest
	startDate := time.Date(2024, 11, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 11, 30, 0, 0, 0, 0, time.UTC)

	total := repo.calculateTotalGoldEarned(&startDate, &endDate)
	expected := 750 // Only the recent quest

	if total != expected {
		t.Errorf("calculateTotalGoldEarned() with date filter = %d, want %d", total, expected)
	}

	// Test without date filter (should include both)
	totalAll := repo.calculateTotalGoldEarned(nil, nil)
	expectedAll := 1250 // 500 + 750

	if totalAll != expectedAll {
		t.Errorf("calculateTotalGoldEarned() without filter = %d, want %d", totalAll, expectedAll)
	}
}

func TestGetQuestStats(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewQuestRepository(db)

	// Insert test quests with small time difference to avoid float precision issues
	now := time.Now().UTC()
	// Use a very small time difference (1 second) to get an integer result from AVG
	assignedAt := now.Add(-1 * time.Second).Format("2006-01-02 15:04:05")
	completedAt := now.Format("2006-01-02 15:04:05")
	createdAt := now.Format("2006-01-02 15:04:05")

	// Completed quest with 750 gold
	_, err := db.Exec(`
		INSERT INTO quests (quest_id, quest_type, goal, ending_progress, completed, rewards, assigned_at, completed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "quest-1", "Daily", 5, 5, 1, "750", assignedAt, completedAt, createdAt)
	if err != nil {
		t.Fatalf("Failed to insert quest: %v", err)
	}

	// Active quest (no completed_at)
	_, err = db.Exec(`
		INSERT INTO quests (quest_id, quest_type, goal, ending_progress, completed, rewards, assigned_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, "quest-2", "Daily", 10, 5, 0, "500", assignedAt, createdAt)
	if err != nil {
		t.Fatalf("Failed to insert quest: %v", err)
	}

	stats, err := repo.GetQuestStats(nil, nil)
	if err != nil {
		t.Fatalf("GetQuestStats failed: %v", err)
	}

	if stats.TotalQuests != 2 {
		t.Errorf("TotalQuests = %d, want 2", stats.TotalQuests)
	}

	if stats.CompletedQuests != 1 {
		t.Errorf("CompletedQuests = %d, want 1", stats.CompletedQuests)
	}

	if stats.ActiveQuests != 1 {
		t.Errorf("ActiveQuests = %d, want 1", stats.ActiveQuests)
	}

	if stats.TotalGoldEarned != 750 {
		t.Errorf("TotalGoldEarned = %d, want 750", stats.TotalGoldEarned)
	}
}

func TestQuestRepositorySave(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewQuestRepository(db)

	quest := &models.Quest{
		QuestID:          "test-quest-123",
		QuestType:        "Daily_Win",
		Goal:             5,
		StartingProgress: 0,
		EndingProgress:   2,
		Completed:        false,
		CanSwap:          true,
		Rewards:          "750",
		AssignedAt:       time.Now().UTC(),
	}

	err := repo.Save(quest)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if quest.ID == 0 {
		t.Error("Quest ID should be set after save")
	}

	// Retrieve and verify
	retrieved, err := repo.GetQuestByID(quest.ID)
	if err != nil {
		t.Fatalf("GetQuestByID failed: %v", err)
	}

	if retrieved.QuestID != quest.QuestID {
		t.Errorf("QuestID = %s, want %s", retrieved.QuestID, quest.QuestID)
	}

	if retrieved.Rewards != "750" {
		t.Errorf("Rewards = %s, want 750", retrieved.Rewards)
	}
}

func TestDeduplicateQuestsByQuestID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewQuestRepository(db)

	// Insert multiple entries for the same quest_id with different created_at
	now := time.Now().UTC()

	// Older entry
	_, err := db.Exec(`
		INSERT INTO quests (quest_id, quest_type, goal, ending_progress, completed, rewards, assigned_at, completed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "dup-quest", "Daily", 5, 3, 0, "500",
		now.Add(-48*time.Hour).Format("2006-01-02 15:04:05"),
		nil,
		now.Add(-24*time.Hour).Format("2006-01-02 15:04:05"))
	if err != nil {
		t.Fatalf("Failed to insert older quest: %v", err)
	}

	// Newer entry (completed with 750 gold)
	_, err = db.Exec(`
		INSERT INTO quests (quest_id, quest_type, goal, ending_progress, completed, rewards, assigned_at, completed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, "dup-quest", "Daily", 5, 5, 1, "750",
		now.Add(-48*time.Hour).Format("2006-01-02 15:04:05"),
		now.Format("2006-01-02 15:04:05"),
		now.Format("2006-01-02 15:04:05"))
	if err != nil {
		t.Fatalf("Failed to insert newer quest: %v", err)
	}

	// Calculate gold - should only count the newer entry (750)
	total := repo.calculateTotalGoldEarned(nil, nil)
	if total != 750 {
		t.Errorf("calculateTotalGoldEarned with duplicates = %d, want 750", total)
	}
}

func TestQuestReassignment(t *testing.T) {
	// Test that when MTGA reuses a quest_id for a new quest after the old one was completed,
	// we create a new record instead of updating the old completed one.
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewQuestRepository(db)
	now := time.Now().UTC()

	// First, save a quest and complete it
	quest1 := &models.Quest{
		QuestID:        "reused-quest-id",
		QuestType:      "First Quest",
		Goal:           5,
		EndingProgress: 5,
		Completed:      true,
		AssignedAt:     now.Add(-48 * time.Hour),
		CompletedAt:    &now,
		LastSeenAt:     &now,
	}
	completedAt := now.Add(-24 * time.Hour)
	quest1.CompletedAt = &completedAt

	err := repo.Save(quest1)
	if err != nil {
		t.Fatalf("Failed to save first quest: %v", err)
	}
	firstQuestID := quest1.ID

	// Now MTGA reuses the same quest_id for a NEW quest
	newLastSeen := now
	quest2 := &models.Quest{
		QuestID:        "reused-quest-id", // Same ID!
		QuestType:      "Second Quest (Reused ID)",
		Goal:           10,
		EndingProgress: 0,
		Completed:      false, // Not completed - this is a NEW quest
		AssignedAt:     now,
		LastSeenAt:     &newLastSeen,
	}

	err = repo.Save(quest2)
	if err != nil {
		t.Fatalf("Failed to save second quest: %v", err)
	}

	// The second quest should get a NEW ID (not update the first)
	if quest2.ID == firstQuestID {
		t.Errorf("Quest reassignment should create new record, but got same ID: %d", quest2.ID)
	}

	// Verify we now have 2 records with the same quest_id
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM quests WHERE quest_id = ?", "reused-quest-id").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count quests: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 quests with same quest_id, got %d", count)
	}

	// Verify the first quest is still completed
	var firstCompleted bool
	err = db.QueryRow("SELECT completed FROM quests WHERE id = ?", firstQuestID).Scan(&firstCompleted)
	if err != nil {
		t.Fatalf("Failed to query first quest: %v", err)
	}
	if !firstCompleted {
		t.Error("First quest should still be completed")
	}

	// Verify the second quest is not completed
	var secondCompleted bool
	err = db.QueryRow("SELECT completed FROM quests WHERE id = ?", quest2.ID).Scan(&secondCompleted)
	if err != nil {
		t.Fatalf("Failed to query second quest: %v", err)
	}
	if secondCompleted {
		t.Error("Second quest should not be completed")
	}
}
