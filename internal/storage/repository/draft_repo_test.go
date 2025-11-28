package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// setupDraftTestDB creates an in-memory database with all draft-related tables.
func setupDraftTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE draft_sessions (
			id TEXT PRIMARY KEY,
			event_name TEXT NOT NULL,
			set_code TEXT NOT NULL,
			draft_type TEXT DEFAULT 'quick_draft',
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP,
			status TEXT DEFAULT 'in_progress',
			total_picks INTEGER DEFAULT 0,
			overall_grade TEXT,
			overall_score INTEGER,
			pick_quality_score REAL,
			color_discipline_score REAL,
			deck_composition_score REAL,
			strategic_score REAL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE draft_picks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			pack_number INTEGER NOT NULL,
			pick_number INTEGER NOT NULL,
			card_id TEXT NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			pick_quality_grade TEXT,
			pick_quality_rank INTEGER,
			pack_best_gihwr REAL,
			picked_card_gihwr REAL,
			alternatives_json TEXT,
			FOREIGN KEY (session_id) REFERENCES draft_sessions(id) ON DELETE CASCADE,
			UNIQUE(session_id, pack_number, pick_number)
		);

		CREATE TABLE draft_packs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			pack_number INTEGER NOT NULL,
			pick_number INTEGER NOT NULL,
			card_ids TEXT NOT NULL,
			timestamp TIMESTAMP NOT NULL,
			FOREIGN KEY (session_id) REFERENCES draft_sessions(id) ON DELETE CASCADE,
			UNIQUE(session_id, pack_number, pick_number)
		);

		CREATE INDEX idx_draft_picks_session ON draft_picks(session_id);
		CREATE INDEX idx_draft_packs_session ON draft_packs(session_id);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestDraftRepository_CreateSession(t *testing.T) {
	db := setupDraftTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRepository(db)
	ctx := context.Background()

	now := time.Now()

	session := &models.DraftSession{
		ID:         "session-1",
		EventName:  "Quick Draft FDN",
		SetCode:    "FDN",
		DraftType:  "quick_draft",
		StartTime:  now,
		Status:     "in_progress",
		TotalPicks: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err := repo.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Verify it was created
	retrieved, err := repo.GetSession(ctx, "session-1")
	if err != nil {
		t.Fatalf("failed to retrieve session: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected session to be found")
	}

	if retrieved.EventName != "Quick Draft FDN" {
		t.Errorf("expected event name 'Quick Draft FDN', got '%s'", retrieved.EventName)
	}

	if retrieved.SetCode != "FDN" {
		t.Errorf("expected set code 'FDN', got '%s'", retrieved.SetCode)
	}

	if retrieved.Status != "in_progress" {
		t.Errorf("expected status 'in_progress', got '%s'", retrieved.Status)
	}
}

func TestDraftRepository_GetSessionCount(t *testing.T) {
	db := setupDraftTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRepository(db)
	ctx := context.Background()

	// Initially should be 0
	count, err := repo.GetSessionCount(ctx)
	if err != nil {
		t.Fatalf("failed to get session count: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 sessions, got %d", count)
	}

	// Create sessions
	now := time.Now()
	for i := 1; i <= 3; i++ {
		session := &models.DraftSession{
			ID:         "session-" + string(rune('0'+i)),
			EventName:  "Quick Draft",
			SetCode:    "FDN",
			DraftType:  "quick_draft",
			StartTime:  now,
			Status:     "in_progress",
			TotalPicks: 0,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := repo.CreateSession(ctx, session); err != nil {
			t.Fatalf("failed to create session: %v", err)
		}
	}

	count, err = repo.GetSessionCount(ctx)
	if err != nil {
		t.Fatalf("failed to get session count: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 sessions, got %d", count)
	}
}

func TestDraftRepository_GetPickCount(t *testing.T) {
	db := setupDraftTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRepository(db)
	ctx := context.Background()

	// Initially should be 0
	count, err := repo.GetPickCount(ctx)
	if err != nil {
		t.Fatalf("failed to get pick count: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 picks, got %d", count)
	}

	// Create a session and picks
	now := time.Now()
	session := &models.DraftSession{
		ID:         "session-1",
		EventName:  "Quick Draft",
		SetCode:    "FDN",
		DraftType:  "quick_draft",
		StartTime:  now,
		Status:     "in_progress",
		TotalPicks: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := repo.CreateSession(ctx, session); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add picks
	for i := 1; i <= 5; i++ {
		pick := &models.DraftPickSession{
			SessionID:  "session-1",
			PackNumber: 1,
			PickNumber: i,
			CardID:     "12345",
			Timestamp:  now,
		}
		if err := repo.SavePick(ctx, pick); err != nil {
			t.Fatalf("failed to save pick: %v", err)
		}
	}

	count, err = repo.GetPickCount(ctx)
	if err != nil {
		t.Fatalf("failed to get pick count: %v", err)
	}

	if count != 5 {
		t.Errorf("expected 5 picks, got %d", count)
	}
}

func TestDraftRepository_GetPackCount(t *testing.T) {
	db := setupDraftTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRepository(db)
	ctx := context.Background()

	// Initially should be 0
	count, err := repo.GetPackCount(ctx)
	if err != nil {
		t.Fatalf("failed to get pack count: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 packs, got %d", count)
	}

	// Create a session and packs
	now := time.Now()
	session := &models.DraftSession{
		ID:         "session-1",
		EventName:  "Quick Draft",
		SetCode:    "FDN",
		DraftType:  "quick_draft",
		StartTime:  now,
		Status:     "in_progress",
		TotalPicks: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := repo.CreateSession(ctx, session); err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Add packs
	for i := 1; i <= 3; i++ {
		pack := &models.DraftPackSession{
			SessionID:  "session-1",
			PackNumber: i,
			PickNumber: 1,
			CardIDs:    []string{"12345", "67890", "11111"},
			Timestamp:  now,
		}
		if err := repo.SavePack(ctx, pack); err != nil {
			t.Fatalf("failed to save pack: %v", err)
		}
	}

	count, err = repo.GetPackCount(ctx)
	if err != nil {
		t.Fatalf("failed to get pack count: %v", err)
	}

	if count != 3 {
		t.Errorf("expected 3 packs, got %d", count)
	}
}

func TestDraftRepository_ClearAllSessions(t *testing.T) {
	db := setupDraftTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create multiple sessions with picks and packs
	for i := 1; i <= 2; i++ {
		sessionID := "session-" + string(rune('0'+i))
		session := &models.DraftSession{
			ID:         sessionID,
			EventName:  "Quick Draft",
			SetCode:    "FDN",
			DraftType:  "quick_draft",
			StartTime:  now,
			Status:     "in_progress",
			TotalPicks: 0,
			CreatedAt:  now,
			UpdatedAt:  now,
		}
		if err := repo.CreateSession(ctx, session); err != nil {
			t.Fatalf("failed to create session: %v", err)
		}

		// Add picks for each session
		for j := 1; j <= 3; j++ {
			pick := &models.DraftPickSession{
				SessionID:  sessionID,
				PackNumber: 1,
				PickNumber: j,
				CardID:     "12345",
				Timestamp:  now,
			}
			if err := repo.SavePick(ctx, pick); err != nil {
				t.Fatalf("failed to save pick: %v", err)
			}
		}

		// Add packs for each session
		for j := 1; j <= 2; j++ {
			pack := &models.DraftPackSession{
				SessionID:  sessionID,
				PackNumber: j,
				PickNumber: 1,
				CardIDs:    []string{"12345", "67890"},
				Timestamp:  now,
			}
			if err := repo.SavePack(ctx, pack); err != nil {
				t.Fatalf("failed to save pack: %v", err)
			}
		}
	}

	// Verify data exists before clearing
	sessionCount, _ := repo.GetSessionCount(ctx)
	pickCount, _ := repo.GetPickCount(ctx)
	packCount, _ := repo.GetPackCount(ctx)

	if sessionCount != 2 {
		t.Errorf("expected 2 sessions before clear, got %d", sessionCount)
	}
	if pickCount != 6 {
		t.Errorf("expected 6 picks before clear, got %d", pickCount)
	}
	if packCount != 4 {
		t.Errorf("expected 4 packs before clear, got %d", packCount)
	}

	// Clear all sessions
	sessionsDeleted, picksDeleted, packsDeleted, err := repo.ClearAllSessions(ctx)
	if err != nil {
		t.Fatalf("failed to clear all sessions: %v", err)
	}

	// Verify returned counts
	if sessionsDeleted != 2 {
		t.Errorf("expected 2 sessions deleted, got %d", sessionsDeleted)
	}
	if picksDeleted != 6 {
		t.Errorf("expected 6 picks deleted, got %d", picksDeleted)
	}
	if packsDeleted != 4 {
		t.Errorf("expected 4 packs deleted, got %d", packsDeleted)
	}

	// Verify data is gone
	sessionCount, _ = repo.GetSessionCount(ctx)
	pickCount, _ = repo.GetPickCount(ctx)
	packCount, _ = repo.GetPackCount(ctx)

	if sessionCount != 0 {
		t.Errorf("expected 0 sessions after clear, got %d", sessionCount)
	}
	if pickCount != 0 {
		t.Errorf("expected 0 picks after clear, got %d", pickCount)
	}
	if packCount != 0 {
		t.Errorf("expected 0 packs after clear, got %d", packCount)
	}
}

func TestDraftRepository_ClearAllSessions_Empty(t *testing.T) {
	db := setupDraftTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRepository(db)
	ctx := context.Background()

	// Clear on empty database should not error
	sessionsDeleted, picksDeleted, packsDeleted, err := repo.ClearAllSessions(ctx)
	if err != nil {
		t.Fatalf("failed to clear all sessions on empty db: %v", err)
	}

	if sessionsDeleted != 0 || picksDeleted != 0 || packsDeleted != 0 {
		t.Errorf("expected all 0 deletions on empty db, got sessions=%d, picks=%d, packs=%d",
			sessionsDeleted, picksDeleted, packsDeleted)
	}
}

func TestDraftRepository_SavePick(t *testing.T) {
	db := setupDraftTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create a session first
	session := &models.DraftSession{
		ID:         "session-1",
		EventName:  "Quick Draft FDN",
		SetCode:    "FDN",
		DraftType:  "quick_draft",
		StartTime:  now,
		Status:     "in_progress",
		TotalPicks: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err := repo.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Save a pick
	pick := &models.DraftPickSession{
		SessionID:  "session-1",
		PackNumber: 1,
		PickNumber: 1,
		CardID:     "12345",
		Timestamp:  now,
	}

	err = repo.SavePick(ctx, pick)
	if err != nil {
		t.Fatalf("failed to save pick: %v", err)
	}

	if pick.ID == 0 {
		t.Error("expected pick ID to be set")
	}

	// Retrieve picks
	picks, err := repo.GetPicksBySession(ctx, "session-1")
	if err != nil {
		t.Fatalf("failed to get picks: %v", err)
	}

	if len(picks) != 1 {
		t.Fatalf("expected 1 pick, got %d", len(picks))
	}

	if picks[0].CardID != "12345" {
		t.Errorf("expected card ID '12345', got '%s'", picks[0].CardID)
	}
}

func TestDraftRepository_SavePack(t *testing.T) {
	db := setupDraftTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create a session first
	session := &models.DraftSession{
		ID:         "session-1",
		EventName:  "Quick Draft FDN",
		SetCode:    "FDN",
		DraftType:  "quick_draft",
		StartTime:  now,
		Status:     "in_progress",
		TotalPicks: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err := repo.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Save a pack
	pack := &models.DraftPackSession{
		SessionID:  "session-1",
		PackNumber: 1,
		PickNumber: 1,
		CardIDs:    []string{"12345", "67890", "11111"},
		Timestamp:  now,
	}

	err = repo.SavePack(ctx, pack)
	if err != nil {
		t.Fatalf("failed to save pack: %v", err)
	}

	if pack.ID == 0 {
		t.Error("expected pack ID to be set")
	}

	// Retrieve pack
	retrieved, err := repo.GetPack(ctx, "session-1", 1, 1)
	if err != nil {
		t.Fatalf("failed to get pack: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected pack to be found")
	}

	if len(retrieved.CardIDs) != 3 {
		t.Errorf("expected 3 cards, got %d", len(retrieved.CardIDs))
	}
}

func TestDraftRepository_UpdateSessionStatus(t *testing.T) {
	db := setupDraftTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create a session
	session := &models.DraftSession{
		ID:         "session-1",
		EventName:  "Quick Draft FDN",
		SetCode:    "FDN",
		DraftType:  "quick_draft",
		StartTime:  now,
		Status:     "in_progress",
		TotalPicks: 0,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	err := repo.CreateSession(ctx, session)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Update status to completed
	endTime := time.Now()
	err = repo.UpdateSessionStatus(ctx, "session-1", "completed", &endTime)
	if err != nil {
		t.Fatalf("failed to update session status: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetSession(ctx, "session-1")
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}

	if retrieved.Status != "completed" {
		t.Errorf("expected status 'completed', got '%s'", retrieved.Status)
	}

	if retrieved.EndTime == nil {
		t.Error("expected end time to be set")
	}
}

func TestDraftRepository_GetActiveSessions(t *testing.T) {
	db := setupDraftTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create active and completed sessions
	sessions := []*models.DraftSession{
		{
			ID:         "session-1",
			EventName:  "Quick Draft",
			SetCode:    "FDN",
			DraftType:  "quick_draft",
			StartTime:  now,
			Status:     "in_progress",
			TotalPicks: 0,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "session-2",
			EventName:  "Quick Draft",
			SetCode:    "FDN",
			DraftType:  "quick_draft",
			StartTime:  now.Add(-time.Hour),
			Status:     "completed",
			TotalPicks: 45,
			CreatedAt:  now.Add(-time.Hour),
			UpdatedAt:  now,
		},
		{
			ID:         "session-3",
			EventName:  "Quick Draft",
			SetCode:    "FDN",
			DraftType:  "quick_draft",
			StartTime:  now.Add(-30 * time.Minute),
			Status:     "in_progress",
			TotalPicks: 15,
			CreatedAt:  now.Add(-30 * time.Minute),
			UpdatedAt:  now,
		},
	}

	for _, s := range sessions {
		if err := repo.CreateSession(ctx, s); err != nil {
			t.Fatalf("failed to create session: %v", err)
		}
	}

	// Get active sessions
	active, err := repo.GetActiveSessions(ctx)
	if err != nil {
		t.Fatalf("failed to get active sessions: %v", err)
	}

	if len(active) != 2 {
		t.Errorf("expected 2 active sessions, got %d", len(active))
	}

	// Should be ordered by start_time DESC
	if len(active) >= 2 && active[0].ID != "session-1" {
		t.Errorf("expected session-1 first (most recent start), got %s", active[0].ID)
	}
}

func TestDraftRepository_GetCompletedSessions(t *testing.T) {
	db := setupDraftTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftRepository(db)
	ctx := context.Background()

	now := time.Now()

	// Create sessions with different statuses
	sessions := []*models.DraftSession{
		{
			ID:         "session-1",
			EventName:  "Quick Draft",
			SetCode:    "FDN",
			DraftType:  "quick_draft",
			StartTime:  now,
			Status:     "in_progress",
			TotalPicks: 0,
			CreatedAt:  now,
			UpdatedAt:  now,
		},
		{
			ID:         "session-2",
			EventName:  "Quick Draft",
			SetCode:    "FDN",
			DraftType:  "quick_draft",
			StartTime:  now.Add(-time.Hour),
			Status:     "completed",
			TotalPicks: 45,
			CreatedAt:  now.Add(-time.Hour),
			UpdatedAt:  now,
		},
		{
			ID:         "session-3",
			EventName:  "Quick Draft",
			SetCode:    "DSK",
			DraftType:  "quick_draft",
			StartTime:  now.Add(-2 * time.Hour),
			Status:     "completed",
			TotalPicks: 45,
			CreatedAt:  now.Add(-2 * time.Hour),
			UpdatedAt:  now,
		},
	}

	for _, s := range sessions {
		if err := repo.CreateSession(ctx, s); err != nil {
			t.Fatalf("failed to create session: %v", err)
		}
	}

	// Get completed sessions
	completed, err := repo.GetCompletedSessions(ctx, 10)
	if err != nil {
		t.Fatalf("failed to get completed sessions: %v", err)
	}

	if len(completed) != 2 {
		t.Errorf("expected 2 completed sessions, got %d", len(completed))
	}

	// Test limit
	limited, err := repo.GetCompletedSessions(ctx, 1)
	if err != nil {
		t.Fatalf("failed to get limited completed sessions: %v", err)
	}

	if len(limited) != 1 {
		t.Errorf("expected 1 completed session with limit, got %d", len(limited))
	}
}
