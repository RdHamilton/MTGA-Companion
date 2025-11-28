package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// setupRecommendationFeedbackTestDB creates an in-memory database with recommendation_feedback table.
func setupRecommendationFeedbackTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			is_default INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		);

		CREATE TABLE recommendation_feedback (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			account_id INTEGER NOT NULL,
			recommendation_type TEXT NOT NULL CHECK(recommendation_type IN ('card_pick', 'deck_card', 'archetype', 'sideboard')),
			recommendation_id TEXT NOT NULL,
			recommended_card_id INTEGER,
			recommended_archetype TEXT,
			context_data TEXT NOT NULL,
			action TEXT NOT NULL CHECK(action IN ('accepted', 'rejected', 'ignored', 'alternate')),
			alternate_choice_id INTEGER,
			outcome_match_id TEXT,
			outcome_result TEXT CHECK(outcome_result IN ('win', 'loss')),
			recommendation_score REAL,
			recommendation_rank INTEGER,
			recommended_at DATETIME NOT NULL,
			responded_at DATETIME,
			outcome_recorded_at DATETIME,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (account_id) REFERENCES accounts(id) ON DELETE CASCADE
		);

		INSERT INTO accounts (id, name, is_default, created_at, updated_at)
		VALUES (1, 'Test Account', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestRecommendationFeedbackRepository_Create(t *testing.T) {
	db := setupRecommendationFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewRecommendationFeedbackRepository(db)
	ctx := context.Background()

	cardID := 12345
	score := 0.85
	rank := 1

	feedback := &models.RecommendationFeedback{
		AccountID:           1,
		RecommendationType:  "card_pick",
		RecommendationID:    "rec-123",
		RecommendedCardID:   &cardID,
		ContextData:         `{"pack_number": 1, "pick_number": 1}`,
		Action:              "accepted",
		RecommendationScore: &score,
		RecommendationRank:  &rank,
		RecommendedAt:       time.Now(),
	}

	err := repo.Create(ctx, feedback)
	if err != nil {
		t.Fatalf("failed to create feedback: %v", err)
	}

	if feedback.ID == 0 {
		t.Error("expected feedback ID to be set")
	}
}

func TestRecommendationFeedbackRepository_GetByID(t *testing.T) {
	db := setupRecommendationFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewRecommendationFeedbackRepository(db)
	ctx := context.Background()

	cardID := 12345
	feedback := &models.RecommendationFeedback{
		AccountID:          1,
		RecommendationType: "card_pick",
		RecommendationID:   "rec-123",
		RecommendedCardID:  &cardID,
		ContextData:        `{"pack_number": 1}`,
		Action:             "accepted",
		RecommendedAt:      time.Now(),
	}

	if err := repo.Create(ctx, feedback); err != nil {
		t.Fatalf("failed to create feedback: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, feedback.ID)
	if err != nil {
		t.Fatalf("failed to get feedback: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected feedback, got nil")
	}

	if retrieved.RecommendationID != "rec-123" {
		t.Errorf("expected recommendation ID 'rec-123', got '%s'", retrieved.RecommendationID)
	}

	if *retrieved.RecommendedCardID != 12345 {
		t.Errorf("expected card ID 12345, got %d", *retrieved.RecommendedCardID)
	}
}

func TestRecommendationFeedbackRepository_GetByRecommendationID(t *testing.T) {
	db := setupRecommendationFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewRecommendationFeedbackRepository(db)
	ctx := context.Background()

	feedback := &models.RecommendationFeedback{
		AccountID:          1,
		RecommendationType: "card_pick",
		RecommendationID:   "unique-rec-id",
		ContextData:        `{}`,
		Action:             "rejected",
		RecommendedAt:      time.Now(),
	}

	if err := repo.Create(ctx, feedback); err != nil {
		t.Fatalf("failed to create feedback: %v", err)
	}

	retrieved, err := repo.GetByRecommendationID(ctx, "unique-rec-id")
	if err != nil {
		t.Fatalf("failed to get feedback: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected feedback, got nil")
	}

	if retrieved.ID != feedback.ID {
		t.Errorf("expected ID %d, got %d", feedback.ID, retrieved.ID)
	}
}

func TestRecommendationFeedbackRepository_GetByAccount(t *testing.T) {
	db := setupRecommendationFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewRecommendationFeedbackRepository(db)
	ctx := context.Background()

	// Create multiple feedback entries
	for i := 0; i < 5; i++ {
		feedback := &models.RecommendationFeedback{
			AccountID:          1,
			RecommendationType: "card_pick",
			RecommendationID:   "rec-" + string(rune('0'+i)),
			ContextData:        `{}`,
			Action:             "accepted",
			RecommendedAt:      time.Now().Add(time.Duration(-i) * time.Minute),
		}
		if err := repo.Create(ctx, feedback); err != nil {
			t.Fatalf("failed to create feedback: %v", err)
		}
	}

	feedbacks, err := repo.GetByAccount(ctx, 1, 3)
	if err != nil {
		t.Fatalf("failed to get feedbacks: %v", err)
	}

	if len(feedbacks) != 3 {
		t.Errorf("expected 3 feedbacks, got %d", len(feedbacks))
	}
}

func TestRecommendationFeedbackRepository_GetByType(t *testing.T) {
	db := setupRecommendationFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewRecommendationFeedbackRepository(db)
	ctx := context.Background()

	// Create feedback of different types
	types := []string{"card_pick", "card_pick", "deck_card", "archetype"}
	for i, recType := range types {
		feedback := &models.RecommendationFeedback{
			AccountID:          1,
			RecommendationType: recType,
			RecommendationID:   "rec-" + string(rune('0'+i)),
			ContextData:        `{}`,
			Action:             "accepted",
			RecommendedAt:      time.Now(),
		}
		if err := repo.Create(ctx, feedback); err != nil {
			t.Fatalf("failed to create feedback: %v", err)
		}
	}

	cardPickFeedbacks, err := repo.GetByType(ctx, 1, "card_pick", 10)
	if err != nil {
		t.Fatalf("failed to get feedbacks by type: %v", err)
	}

	if len(cardPickFeedbacks) != 2 {
		t.Errorf("expected 2 card_pick feedbacks, got %d", len(cardPickFeedbacks))
	}
}

func TestRecommendationFeedbackRepository_UpdateAction(t *testing.T) {
	db := setupRecommendationFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewRecommendationFeedbackRepository(db)
	ctx := context.Background()

	feedback := &models.RecommendationFeedback{
		AccountID:          1,
		RecommendationType: "card_pick",
		RecommendationID:   "rec-123",
		ContextData:        `{}`,
		Action:             "ignored", // Start with ignored
		RecommendedAt:      time.Now(),
	}

	if err := repo.Create(ctx, feedback); err != nil {
		t.Fatalf("failed to create feedback: %v", err)
	}

	alternateID := 67890
	err := repo.UpdateAction(ctx, feedback.ID, "alternate", &alternateID)
	if err != nil {
		t.Fatalf("failed to update action: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, feedback.ID)
	if err != nil {
		t.Fatalf("failed to get feedback: %v", err)
	}

	if retrieved.Action != "alternate" {
		t.Errorf("expected action 'alternate', got '%s'", retrieved.Action)
	}

	if retrieved.AlternateChoiceID == nil || *retrieved.AlternateChoiceID != 67890 {
		t.Error("expected alternate choice ID to be set to 67890")
	}

	if retrieved.RespondedAt == nil {
		t.Error("expected responded_at to be set")
	}
}

func TestRecommendationFeedbackRepository_UpdateOutcome(t *testing.T) {
	db := setupRecommendationFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewRecommendationFeedbackRepository(db)
	ctx := context.Background()

	feedback := &models.RecommendationFeedback{
		AccountID:          1,
		RecommendationType: "card_pick",
		RecommendationID:   "rec-123",
		ContextData:        `{}`,
		Action:             "accepted",
		RecommendedAt:      time.Now(),
	}

	if err := repo.Create(ctx, feedback); err != nil {
		t.Fatalf("failed to create feedback: %v", err)
	}

	err := repo.UpdateOutcome(ctx, feedback.ID, "match-456", "win")
	if err != nil {
		t.Fatalf("failed to update outcome: %v", err)
	}

	retrieved, err := repo.GetByID(ctx, feedback.ID)
	if err != nil {
		t.Fatalf("failed to get feedback: %v", err)
	}

	if retrieved.OutcomeMatchID == nil || *retrieved.OutcomeMatchID != "match-456" {
		t.Error("expected outcome match ID to be 'match-456'")
	}

	if retrieved.OutcomeResult == nil || *retrieved.OutcomeResult != "win" {
		t.Error("expected outcome result to be 'win'")
	}

	if retrieved.OutcomeRecordedAt == nil {
		t.Error("expected outcome_recorded_at to be set")
	}
}

func TestRecommendationFeedbackRepository_GetStats(t *testing.T) {
	db := setupRecommendationFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewRecommendationFeedbackRepository(db)
	ctx := context.Background()

	// Create feedback with various actions
	actions := []string{"accepted", "accepted", "accepted", "rejected", "ignored"}
	for i, action := range actions {
		feedback := &models.RecommendationFeedback{
			AccountID:          1,
			RecommendationType: "card_pick",
			RecommendationID:   "rec-" + string(rune('0'+i)),
			ContextData:        `{}`,
			Action:             action,
			RecommendedAt:      time.Now(),
		}
		if err := repo.Create(ctx, feedback); err != nil {
			t.Fatalf("failed to create feedback: %v", err)
		}
	}

	stats, err := repo.GetStats(ctx, 1, nil)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats.TotalRecommendations != 5 {
		t.Errorf("expected 5 total, got %d", stats.TotalRecommendations)
	}

	if stats.AcceptedCount != 3 {
		t.Errorf("expected 3 accepted, got %d", stats.AcceptedCount)
	}

	if stats.RejectedCount != 1 {
		t.Errorf("expected 1 rejected, got %d", stats.RejectedCount)
	}

	if stats.IgnoredCount != 1 {
		t.Errorf("expected 1 ignored, got %d", stats.IgnoredCount)
	}

	expectedAcceptanceRate := 0.6
	if stats.AcceptanceRate != expectedAcceptanceRate {
		t.Errorf("expected acceptance rate %f, got %f", expectedAcceptanceRate, stats.AcceptanceRate)
	}
}

func TestRecommendationFeedbackRepository_GetStatsByType(t *testing.T) {
	db := setupRecommendationFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewRecommendationFeedbackRepository(db)
	ctx := context.Background()

	// Create feedback of different types
	feedbackData := []struct {
		recType string
		action  string
	}{
		{"card_pick", "accepted"},
		{"card_pick", "rejected"},
		{"deck_card", "accepted"},
		{"archetype", "accepted"},
	}

	for i, fd := range feedbackData {
		feedback := &models.RecommendationFeedback{
			AccountID:          1,
			RecommendationType: fd.recType,
			RecommendationID:   "rec-" + string(rune('0'+i)),
			ContextData:        `{}`,
			Action:             fd.action,
			RecommendedAt:      time.Now(),
		}
		if err := repo.Create(ctx, feedback); err != nil {
			t.Fatalf("failed to create feedback: %v", err)
		}
	}

	recType := "card_pick"
	stats, err := repo.GetStats(ctx, 1, &recType)
	if err != nil {
		t.Fatalf("failed to get stats by type: %v", err)
	}

	if stats.TotalRecommendations != 2 {
		t.Errorf("expected 2 card_pick recommendations, got %d", stats.TotalRecommendations)
	}

	if stats.AcceptedCount != 1 {
		t.Errorf("expected 1 accepted, got %d", stats.AcceptedCount)
	}
}

func TestRecommendationFeedbackRepository_GetPendingFeedback(t *testing.T) {
	db := setupRecommendationFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewRecommendationFeedbackRepository(db)
	ctx := context.Background()

	// Create feedback - some responded, some not
	for i := 0; i < 4; i++ {
		feedback := &models.RecommendationFeedback{
			AccountID:          1,
			RecommendationType: "card_pick",
			RecommendationID:   "rec-" + string(rune('0'+i)),
			ContextData:        `{}`,
			Action:             "ignored",
			RecommendedAt:      time.Now(),
		}
		if err := repo.Create(ctx, feedback); err != nil {
			t.Fatalf("failed to create feedback: %v", err)
		}

		// Mark some as responded
		if i < 2 {
			if err := repo.UpdateAction(ctx, feedback.ID, "accepted", nil); err != nil {
				t.Fatalf("failed to update action: %v", err)
			}
		}
	}

	pending, err := repo.GetPendingFeedback(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get pending feedback: %v", err)
	}

	if len(pending) != 2 {
		t.Errorf("expected 2 pending feedbacks, got %d", len(pending))
	}
}

func TestRecommendationFeedbackRepository_GetForMLTraining(t *testing.T) {
	db := setupRecommendationFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewRecommendationFeedbackRepository(db)
	ctx := context.Background()

	// Create feedback with outcomes
	for i := 0; i < 5; i++ {
		feedback := &models.RecommendationFeedback{
			AccountID:          1,
			RecommendationType: "card_pick",
			RecommendationID:   "rec-" + string(rune('0'+i)),
			ContextData:        `{"deck_state": "test"}`,
			Action:             "accepted",
			RecommendedAt:      time.Now(),
		}
		if err := repo.Create(ctx, feedback); err != nil {
			t.Fatalf("failed to create feedback: %v", err)
		}

		// Only some have outcomes
		if i < 3 {
			if err := repo.UpdateAction(ctx, feedback.ID, "accepted", nil); err != nil {
				t.Fatalf("failed to update action: %v", err)
			}
			result := "win"
			if i%2 == 0 {
				result = "loss"
			}
			if err := repo.UpdateOutcome(ctx, feedback.ID, "match-"+string(rune('0'+i)), result); err != nil {
				t.Fatalf("failed to update outcome: %v", err)
			}
		}
	}

	training, err := repo.GetForMLTraining(ctx, 10)
	if err != nil {
		t.Fatalf("failed to get ML training data: %v", err)
	}

	if len(training) != 3 {
		t.Errorf("expected 3 training entries, got %d", len(training))
	}

	for _, f := range training {
		if f.OutcomeResult == nil {
			t.Error("training data should have outcome_result")
		}
		if f.RespondedAt == nil {
			t.Error("training data should have responded_at")
		}
	}
}

func TestRecommendationFeedbackRepository_GetStatsByDateRange(t *testing.T) {
	db := setupRecommendationFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := NewRecommendationFeedbackRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create feedback at different times
	times := []time.Time{
		now.Add(-48 * time.Hour), // 2 days ago
		now.Add(-24 * time.Hour), // 1 day ago
		now.Add(-12 * time.Hour), // 12 hours ago
		now,                      // now
	}

	for i, ts := range times {
		feedback := &models.RecommendationFeedback{
			AccountID:          1,
			RecommendationType: "card_pick",
			RecommendationID:   "rec-" + string(rune('0'+i)),
			ContextData:        `{}`,
			Action:             "accepted",
			RecommendedAt:      ts,
		}
		if err := repo.Create(ctx, feedback); err != nil {
			t.Fatalf("failed to create feedback: %v", err)
		}
	}

	// Query for last 36 hours
	start := now.Add(-36 * time.Hour)
	end := now.Add(time.Hour)

	stats, err := repo.GetStatsByDateRange(ctx, 1, start, end)
	if err != nil {
		t.Fatalf("failed to get stats by date range: %v", err)
	}

	if stats.TotalRecommendations != 3 {
		t.Errorf("expected 3 recommendations in range, got %d", stats.TotalRecommendations)
	}
}
