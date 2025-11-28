package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// setupDraftTestDB creates an in-memory database with draft tables.
func setupDraftTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE draft_sessions (
			id TEXT PRIMARY KEY,
			event_name TEXT NOT NULL,
			set_code TEXT NOT NULL,
			draft_type TEXT NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			status TEXT NOT NULL DEFAULT 'in_progress',
			total_picks INTEGER DEFAULT 0,
			overall_grade TEXT,
			overall_score INTEGER,
			pick_quality_score REAL,
			color_discipline_score REAL,
			deck_composition_score REAL,
			strategic_score REAL,
			predicted_win_rate REAL,
			predicted_win_rate_min REAL,
			predicted_win_rate_max REAL,
			prediction_factors TEXT,
			predicted_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE draft_picks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			pack_number INTEGER NOT NULL,
			pick_number INTEGER NOT NULL,
			card_id TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			pick_quality_grade TEXT,
			pick_quality_rank INTEGER,
			pack_best_gihwr REAL,
			picked_card_gihwr REAL,
			alternatives_json TEXT,
			FOREIGN KEY (session_id) REFERENCES draft_sessions(id),
			UNIQUE(session_id, pack_number, pick_number)
		);

		CREATE TABLE draft_packs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			pack_number INTEGER NOT NULL,
			pick_number INTEGER NOT NULL,
			card_ids TEXT NOT NULL,
			timestamp DATETIME NOT NULL,
			FOREIGN KEY (session_id) REFERENCES draft_sessions(id),
			UNIQUE(session_id, pack_number, pick_number)
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestDraftRepository_CreateSession(t *testing.T) {
	db := setupDraftTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewDraftRepository(db)
	ctx := context.Background()

	session := &models.DraftSession{
		ID:         "test-session-1",
		EventName:  "QuickDraft_FDN",
		SetCode:    "FDN",
		DraftType:  "QuickDraft",
		StartTime:  time.Now(),
		Status:     "in_progress",
		TotalPicks: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err := repo.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Verify it was created
	retrieved, err := repo.GetSession(ctx, "test-session-1")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if retrieved == nil {
		t.Fatal("GetSession returned nil")
	}
	if retrieved.ID != "test-session-1" {
		t.Errorf("expected ID 'test-session-1', got '%s'", retrieved.ID)
	}
	if retrieved.SetCode != "FDN" {
		t.Errorf("expected SetCode 'FDN', got '%s'", retrieved.SetCode)
	}
}

func TestDraftRepository_UpdateSessionPrediction(t *testing.T) {
	db := setupDraftTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewDraftRepository(db)
	ctx := context.Background()

	// Create a session first
	session := &models.DraftSession{
		ID:         "test-session-pred",
		EventName:  "QuickDraft_FDN",
		SetCode:    "FDN",
		DraftType:  "QuickDraft",
		StartTime:  time.Now(),
		Status:     "completed",
		TotalPicks: 45,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := repo.CreateSession(ctx, session); err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Update prediction
	predictedAt := time.Now()
	factorsJSON := `{"base_win_rate":0.55,"card_quality":0.05,"synergy":-0.02}`

	err := repo.UpdateSessionPrediction(ctx, "test-session-pred", 0.58, 0.52, 0.64, factorsJSON, predictedAt)
	if err != nil {
		t.Fatalf("UpdateSessionPrediction failed: %v", err)
	}

	// Verify prediction was saved
	retrieved, err := repo.GetSession(ctx, "test-session-pred")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if retrieved.PredictedWinRate == nil {
		t.Fatal("PredictedWinRate is nil after update")
	}
	if *retrieved.PredictedWinRate != 0.58 {
		t.Errorf("expected PredictedWinRate 0.58, got %f", *retrieved.PredictedWinRate)
	}
	if *retrieved.PredictedWinRateMin != 0.52 {
		t.Errorf("expected PredictedWinRateMin 0.52, got %f", *retrieved.PredictedWinRateMin)
	}
	if *retrieved.PredictedWinRateMax != 0.64 {
		t.Errorf("expected PredictedWinRateMax 0.64, got %f", *retrieved.PredictedWinRateMax)
	}
	if retrieved.PredictionFactors == nil || *retrieved.PredictionFactors != factorsJSON {
		t.Errorf("PredictionFactors mismatch")
	}
}

func TestDraftRepository_UpdateSessionTotalPicks(t *testing.T) {
	db := setupDraftTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewDraftRepository(db)
	ctx := context.Background()

	// Create a session first
	session := &models.DraftSession{
		ID:         "test-session-picks",
		EventName:  "QuickDraft_FDN",
		SetCode:    "FDN",
		DraftType:  "QuickDraft",
		StartTime:  time.Now(),
		Status:     "in_progress",
		TotalPicks: 0,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := repo.CreateSession(ctx, session); err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Update total picks
	err := repo.UpdateSessionTotalPicks(ctx, "test-session-picks", 45)
	if err != nil {
		t.Fatalf("UpdateSessionTotalPicks failed: %v", err)
	}

	// Verify
	retrieved, err := repo.GetSession(ctx, "test-session-picks")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if retrieved.TotalPicks != 45 {
		t.Errorf("expected TotalPicks 45, got %d", retrieved.TotalPicks)
	}
}

func TestDraftRepository_UpdateSessionGrade(t *testing.T) {
	db := setupDraftTestDB(t)
	defer func() { _ = db.Close() }()

	repo := NewDraftRepository(db)
	ctx := context.Background()

	// Create a session first
	session := &models.DraftSession{
		ID:         "test-session-grade",
		EventName:  "QuickDraft_FDN",
		SetCode:    "FDN",
		DraftType:  "QuickDraft",
		StartTime:  time.Now(),
		Status:     "completed",
		TotalPicks: 45,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := repo.CreateSession(ctx, session); err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Update grade
	err := repo.UpdateSessionGrade(ctx, "test-session-grade", "A-", 85, 35.0, 18.0, 20.0, 12.0)
	if err != nil {
		t.Fatalf("UpdateSessionGrade failed: %v", err)
	}

	// Verify
	retrieved, err := repo.GetSession(ctx, "test-session-grade")
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}
	if retrieved.OverallGrade == nil || *retrieved.OverallGrade != "A-" {
		t.Errorf("expected OverallGrade 'A-', got %v", retrieved.OverallGrade)
	}
	if retrieved.OverallScore == nil || *retrieved.OverallScore != 85 {
		t.Errorf("expected OverallScore 85, got %v", retrieved.OverallScore)
	}
}
