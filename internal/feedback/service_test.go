package feedback

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

func setupFeedbackTestDB(t *testing.T) *sql.DB {
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

func TestService_RecordRecommendation(t *testing.T) {
	db := setupFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := repository.NewRecommendationFeedbackRepository(db)
	svc := NewService(repo, 1)
	ctx := context.Background()

	cardID := 12345
	score := 0.85
	rank := 1

	recContext := &RecommendationContext{
		DeckID:            "deck-123",
		DeckCardCount:     40,
		DeckColorIdentity: "WU",
		PackNumber:        1,
		PickNumber:        1,
		Timestamp:         time.Now(),
	}

	req := &RecordRecommendationRequest{
		RecommendationType: "card_pick",
		RecommendedCardID:  &cardID,
		Context:            recContext,
		Score:              &score,
		Rank:               &rank,
	}

	recommendationID, err := svc.RecordRecommendation(ctx, req)
	if err != nil {
		t.Fatalf("failed to record recommendation: %v", err)
	}

	if recommendationID == "" {
		t.Error("expected non-empty recommendation ID")
	}

	// Verify it was saved
	feedback, err := repo.GetByRecommendationID(ctx, recommendationID)
	if err != nil {
		t.Fatalf("failed to get feedback: %v", err)
	}

	if feedback == nil {
		t.Fatal("expected feedback, got nil")
	}

	if feedback.Action != "ignored" {
		t.Errorf("expected initial action 'ignored', got '%s'", feedback.Action)
	}

	if *feedback.RecommendedCardID != 12345 {
		t.Errorf("expected card ID 12345, got %d", *feedback.RecommendedCardID)
	}
}

func TestService_RecordAction(t *testing.T) {
	db := setupFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := repository.NewRecommendationFeedbackRepository(db)
	svc := NewService(repo, 1)
	ctx := context.Background()

	// First create a recommendation
	cardID := 12345
	req := &RecordRecommendationRequest{
		RecommendationType: "card_pick",
		RecommendedCardID:  &cardID,
		Context:            &RecommendationContext{Timestamp: time.Now()},
	}

	recommendationID, err := svc.RecordRecommendation(ctx, req)
	if err != nil {
		t.Fatalf("failed to record recommendation: %v", err)
	}

	// Record acceptance
	err = svc.RecordAction(ctx, recommendationID, "accepted", nil)
	if err != nil {
		t.Fatalf("failed to record action: %v", err)
	}

	// Verify
	feedback, err := repo.GetByRecommendationID(ctx, recommendationID)
	if err != nil {
		t.Fatalf("failed to get feedback: %v", err)
	}

	if feedback.Action != "accepted" {
		t.Errorf("expected action 'accepted', got '%s'", feedback.Action)
	}

	if feedback.RespondedAt == nil {
		t.Error("expected responded_at to be set")
	}
}

func TestService_RecordAction_Alternate(t *testing.T) {
	db := setupFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := repository.NewRecommendationFeedbackRepository(db)
	svc := NewService(repo, 1)
	ctx := context.Background()

	// Create a recommendation
	cardID := 12345
	req := &RecordRecommendationRequest{
		RecommendationType: "card_pick",
		RecommendedCardID:  &cardID,
		Context:            &RecommendationContext{Timestamp: time.Now()},
	}

	recommendationID, err := svc.RecordRecommendation(ctx, req)
	if err != nil {
		t.Fatalf("failed to record recommendation: %v", err)
	}

	// Record alternate choice
	alternateID := 67890
	err = svc.RecordAction(ctx, recommendationID, "alternate", &alternateID)
	if err != nil {
		t.Fatalf("failed to record alternate action: %v", err)
	}

	// Verify
	feedback, err := repo.GetByRecommendationID(ctx, recommendationID)
	if err != nil {
		t.Fatalf("failed to get feedback: %v", err)
	}

	if feedback.Action != "alternate" {
		t.Errorf("expected action 'alternate', got '%s'", feedback.Action)
	}

	if feedback.AlternateChoiceID == nil || *feedback.AlternateChoiceID != 67890 {
		t.Error("expected alternate choice ID 67890")
	}
}

func TestService_RecordOutcome(t *testing.T) {
	db := setupFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := repository.NewRecommendationFeedbackRepository(db)
	svc := NewService(repo, 1)
	ctx := context.Background()

	// Create and accept a recommendation
	cardID := 12345
	req := &RecordRecommendationRequest{
		RecommendationType: "card_pick",
		RecommendedCardID:  &cardID,
		Context:            &RecommendationContext{Timestamp: time.Now()},
	}

	recommendationID, err := svc.RecordRecommendation(ctx, req)
	if err != nil {
		t.Fatalf("failed to record recommendation: %v", err)
	}

	err = svc.RecordAction(ctx, recommendationID, "accepted", nil)
	if err != nil {
		t.Fatalf("failed to record action: %v", err)
	}

	// Record win outcome
	err = svc.RecordOutcome(ctx, recommendationID, "match-123", "win")
	if err != nil {
		t.Fatalf("failed to record outcome: %v", err)
	}

	// Verify
	feedback, err := repo.GetByRecommendationID(ctx, recommendationID)
	if err != nil {
		t.Fatalf("failed to get feedback: %v", err)
	}

	if feedback.OutcomeMatchID == nil || *feedback.OutcomeMatchID != "match-123" {
		t.Error("expected outcome match ID 'match-123'")
	}

	if feedback.OutcomeResult == nil || *feedback.OutcomeResult != "win" {
		t.Error("expected outcome result 'win'")
	}

	if feedback.OutcomeRecordedAt == nil {
		t.Error("expected outcome_recorded_at to be set")
	}
}

func TestService_GetStats(t *testing.T) {
	db := setupFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := repository.NewRecommendationFeedbackRepository(db)
	svc := NewService(repo, 1)
	ctx := context.Background()

	// Create multiple recommendations with different outcomes
	for i := 0; i < 5; i++ {
		cardID := 10000 + i
		req := &RecordRecommendationRequest{
			RecommendationType: "card_pick",
			RecommendedCardID:  &cardID,
			Context:            &RecommendationContext{Timestamp: time.Now()},
		}

		id, err := svc.RecordRecommendation(ctx, req)
		if err != nil {
			t.Fatalf("failed to record recommendation: %v", err)
		}

		// Accept first 3, reject 2
		action := "accepted"
		if i >= 3 {
			action = "rejected"
		}

		err = svc.RecordAction(ctx, id, action, nil)
		if err != nil {
			t.Fatalf("failed to record action: %v", err)
		}
	}

	stats, err := svc.GetStats(ctx, nil)
	if err != nil {
		t.Fatalf("failed to get stats: %v", err)
	}

	if stats.TotalRecommendations != 5 {
		t.Errorf("expected 5 total, got %d", stats.TotalRecommendations)
	}

	if stats.AcceptedCount != 3 {
		t.Errorf("expected 3 accepted, got %d", stats.AcceptedCount)
	}

	if stats.RejectedCount != 2 {
		t.Errorf("expected 2 rejected, got %d", stats.RejectedCount)
	}

	expectedAcceptanceRate := 0.6
	if stats.AcceptanceRate != expectedAcceptanceRate {
		t.Errorf("expected acceptance rate %f, got %f", expectedAcceptanceRate, stats.AcceptanceRate)
	}
}

func TestService_GetDashboardMetrics(t *testing.T) {
	db := setupFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := repository.NewRecommendationFeedbackRepository(db)
	svc := NewService(repo, 1)
	ctx := context.Background()

	// Create recommendations of different types
	types := []string{"card_pick", "card_pick", "deck_card", "archetype"}
	for i, recType := range types {
		cardID := 10000 + i
		req := &RecordRecommendationRequest{
			RecommendationType: recType,
			RecommendedCardID:  &cardID,
			Context:            &RecommendationContext{Timestamp: time.Now()},
		}

		id, err := svc.RecordRecommendation(ctx, req)
		if err != nil {
			t.Fatalf("failed to record recommendation: %v", err)
		}

		err = svc.RecordAction(ctx, id, "accepted", nil)
		if err != nil {
			t.Fatalf("failed to record action: %v", err)
		}
	}

	metrics, err := svc.GetDashboardMetrics(ctx)
	if err != nil {
		t.Fatalf("failed to get dashboard metrics: %v", err)
	}

	if metrics.TotalRecommendations != 4 {
		t.Errorf("expected 4 total recommendations, got %d", metrics.TotalRecommendations)
	}

	if metrics.AcceptanceRate != 1.0 {
		t.Errorf("expected 100%% acceptance rate, got %f", metrics.AcceptanceRate)
	}

	// Check type breakdown
	if _, ok := metrics.ByType["card_pick"]; !ok {
		t.Error("expected card_pick in type breakdown")
	}

	if metrics.ByType["card_pick"].Total != 2 {
		t.Errorf("expected 2 card_pick recommendations, got %d", metrics.ByType["card_pick"].Total)
	}
}

func TestService_ExportForMLTraining(t *testing.T) {
	db := setupFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := repository.NewRecommendationFeedbackRepository(db)
	svc := NewService(repo, 1)
	ctx := context.Background()

	// Create recommendations with full lifecycle
	for i := 0; i < 3; i++ {
		cardID := 10000 + i
		recContext := &RecommendationContext{
			DeckID:            "deck-test",
			DeckCardCount:     40,
			DeckColorIdentity: "WU",
			Timestamp:         time.Now(),
		}

		req := &RecordRecommendationRequest{
			RecommendationType: "card_pick",
			RecommendedCardID:  &cardID,
			Context:            recContext,
		}

		id, err := svc.RecordRecommendation(ctx, req)
		if err != nil {
			t.Fatalf("failed to record recommendation: %v", err)
		}

		err = svc.RecordAction(ctx, id, "accepted", nil)
		if err != nil {
			t.Fatalf("failed to record action: %v", err)
		}

		result := "win"
		if i == 2 {
			result = "loss"
		}

		err = svc.RecordOutcome(ctx, id, "match-"+string(rune('0'+i)), result)
		if err != nil {
			t.Fatalf("failed to record outcome: %v", err)
		}
	}

	// Export for ML training
	trainingData, err := svc.ExportForMLTraining(ctx, 100)
	if err != nil {
		t.Fatalf("failed to export for ML training: %v", err)
	}

	if len(trainingData) != 3 {
		t.Errorf("expected 3 training entries, got %d", len(trainingData))
	}

	// Verify data structure
	for _, td := range trainingData {
		if td.Context == nil {
			t.Error("expected context to be parsed")
		}
		if td.Action != "accepted" {
			t.Errorf("expected action 'accepted', got '%s'", td.Action)
		}
		if td.OutcomeResult == nil {
			t.Error("expected outcome result to be set")
		}
	}
}

func TestService_RecordBatchRecommendations(t *testing.T) {
	db := setupFeedbackTestDB(t)
	defer func() {
		_ = db.Close()
	}()

	repo := repository.NewRecommendationFeedbackRepository(db)
	svc := NewService(repo, 1)
	ctx := context.Background()

	recContext := &RecommendationContext{
		DeckID:    "deck-test",
		Timestamp: time.Now(),
	}

	cardIDs := []int{10001, 10002, 10003, 10004, 10005}

	ids, err := svc.RecordBatchRecommendations(ctx, "deck_card", recContext, cardIDs)
	if err != nil {
		t.Fatalf("failed to record batch: %v", err)
	}

	if len(ids) != 5 {
		t.Errorf("expected 5 IDs, got %d", len(ids))
	}

	// Verify each was created with correct rank
	for i, id := range ids {
		feedback, err := repo.GetByRecommendationID(ctx, id)
		if err != nil {
			t.Fatalf("failed to get feedback: %v", err)
		}

		expectedRank := i + 1
		if feedback.RecommendationRank == nil || *feedback.RecommendationRank != expectedRank {
			t.Errorf("expected rank %d, got %v", expectedRank, feedback.RecommendationRank)
		}
	}
}
