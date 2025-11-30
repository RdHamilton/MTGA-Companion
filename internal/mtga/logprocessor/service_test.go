package logprocessor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// setupTestService creates a test storage service with a temporary database
func setupTestService(t *testing.T) (*storage.Service, func()) {
	t.Helper()

	// Create temporary database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open database connection with auto-migrate
	config := storage.DefaultConfig(dbPath)
	config.AutoMigrate = true
	db, err := storage.Open(config)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create service
	service := storage.NewService(db)

	cleanup := func() {
		if err := service.Close(); err != nil {
			t.Errorf("Failed to close service: %v", err)
		}
		os.RemoveAll(tmpDir)
	}

	return service, cleanup
}

func TestProcessLogEntries_ArenaStats(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	processor := NewService(service)

	// Create test log entries with match data
	// Note: This is a simplified test - actual log format may differ
	entries := []*logreader.LogEntry{
		{
			IsJSON:    true,
			Timestamp: "2025-11-15 10:00:00",
			JSON: map[string]interface{}{
				"greToClientEvent": map[string]interface{}{
					"greToClientMessages": []interface{}{
						map[string]interface{}{
							"type": "GREMessageType_GameStateMessage",
							"gameStateMessage": map[string]interface{}{
								"gameInfo": map[string]interface{}{
									"matchID":    "match-123",
									"gameNumber": float64(1),
									"matchState": "MatchState_GameInProgress",
								},
								"turnInfo": map[string]interface{}{
									"turnNumber": float64(1),
								},
							},
						},
					},
				},
			},
		},
	}

	// Process entries - service should handle gracefully even if no data is found
	result, err := processor.ProcessLogEntries(ctx, entries)
	if err != nil {
		t.Fatalf("ProcessLogEntries failed: %v", err)
	}

	// Verify service ran without errors
	// Note: Match/game counts may be 0 if parsers don't recognize this test data format
	if len(result.Errors) > 0 {
		t.Errorf("Unexpected errors: %v", result.Errors)
	}
}

func TestProcessLogEntries_Decks(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	processor := NewService(service)

	// Create test log entries with deck data
	// Note: This is a simplified test - actual log format may differ
	entries := []*logreader.LogEntry{
		{
			IsJSON:    true,
			Timestamp: "2025-11-15 10:00:00",
			JSON: map[string]interface{}{
				"EventGetCoursesV2": map[string]interface{}{
					"Courses": []interface{}{
						map[string]interface{}{
							"InternalEventName": "Ladder",
							"CurrentEventState": "EventState_Active",
						},
					},
				},
			},
		},
	}

	// Process entries - service should handle gracefully even if no data is found
	result, err := processor.ProcessLogEntries(ctx, entries)
	if err != nil {
		t.Fatalf("ProcessLogEntries failed: %v", err)
	}

	// Verify service ran without errors
	// Note: Deck counts may be 0 if parsers don't recognize this test data format
	if len(result.Errors) > 0 {
		t.Errorf("Unexpected errors: %v", result.Errors)
	}
}

func TestProcessLogEntries_RankUpdates(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	processor := NewService(service)

	// Create test log entries with rank update data
	entries := []*logreader.LogEntry{
		{
			IsJSON:    true,
			Timestamp: "2025-11-15 10:00:00",
			JSON: map[string]interface{}{
				"constructedSeasonOrdinal": float64(83),
				"constructedClass":         "Gold",
				"constructedLevel":         float64(4),
				"constructedStep":          float64(2),
			},
		},
	}

	// Process entries
	result, err := processor.ProcessLogEntries(ctx, entries)
	if err != nil {
		t.Fatalf("ProcessLogEntries failed: %v", err)
	}

	// Verify results
	if result.RanksStored == 0 {
		t.Error("Expected rank updates to be stored, got 0")
	}
	if len(result.Errors) > 0 {
		t.Errorf("Unexpected errors: %v", result.Errors)
	}
}

func TestProcessLogEntries_MultipleTypes(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	processor := NewService(service)

	// Create test log entries with multiple data types
	entries := []*logreader.LogEntry{
		// Rank update
		{
			IsJSON:    true,
			Timestamp: "2025-11-15 09:00:00",
			JSON: map[string]interface{}{
				"constructedSeasonOrdinal": float64(83),
				"constructedClass":         "Gold",
				"constructedLevel":         float64(4),
			},
		},
		// Match data
		{
			IsJSON:    true,
			Timestamp: "2025-11-15 10:00:00",
			JSON: map[string]interface{}{
				"matchGameRoomStateChangedEvent": map[string]interface{}{
					"gameRoomInfo": map[string]interface{}{
						"gameRoomConfig": map[string]interface{}{
							"matchId": "match-456",
							"eventId": "Ladder",
						},
						"finalMatchResult": map[string]interface{}{
							"resultList": []interface{}{
								map[string]interface{}{
									"scope":  "MatchScope_Match",
									"result": "ResultType_Lost",
								},
							},
						},
					},
				},
			},
		},
	}

	// Process entries
	result, err := processor.ProcessLogEntries(ctx, entries)
	if err != nil {
		t.Fatalf("ProcessLogEntries failed: %v", err)
	}

	// Verify that multiple types were processed
	// Note: Actual counts may vary based on parser implementation
	if result.RanksStored == 0 && result.MatchesStored == 0 {
		t.Error("Expected either ranks or matches to be stored, got 0 for both")
	}
}

func TestProcessLogEntries_EmptyEntries(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	processor := NewService(service)

	// Process empty entries
	result, err := processor.ProcessLogEntries(ctx, []*logreader.LogEntry{})
	if err != nil {
		t.Fatalf("ProcessLogEntries failed on empty entries: %v", err)
	}

	// Verify nothing was stored
	if result.MatchesStored != 0 {
		t.Errorf("Expected 0 matches, got %d", result.MatchesStored)
	}
	if result.DecksStored != 0 {
		t.Errorf("Expected 0 decks, got %d", result.DecksStored)
	}
	if result.RanksStored != 0 {
		t.Errorf("Expected 0 ranks, got %d", result.RanksStored)
	}
}

func TestProcessLogEntries_InvalidData(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()
	processor := NewService(service)

	// Create test log entries with invalid data
	entries := []*logreader.LogEntry{
		{
			IsJSON:    true,
			Timestamp: "invalid-timestamp",
			JSON: map[string]interface{}{
				"someInvalidKey": "someInvalidValue",
			},
		},
	}

	// Process entries - should not fail even with invalid data
	result, err := processor.ProcessLogEntries(ctx, entries)
	if err != nil {
		t.Fatalf("ProcessLogEntries should handle invalid data gracefully: %v", err)
	}

	// Verify nothing was stored (invalid data should be skipped)
	if result.MatchesStored != 0 {
		t.Errorf("Expected 0 matches from invalid data, got %d", result.MatchesStored)
	}
}

func TestProcessLogEntries_ContextCancellation(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	processor := NewService(service)

	entries := []*logreader.LogEntry{
		{
			IsJSON:    true,
			Timestamp: "2025-11-15 10:00:00",
			JSON:      map[string]interface{}{},
		},
	}

	// Process with cancelled context
	// Note: Current implementation may not check context cancellation in all paths
	// This test ensures it doesn't panic or hang
	_, _ = processor.ProcessLogEntries(ctx, entries)
}

func TestNewService(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	processor := NewService(service)
	if processor == nil {
		t.Fatal("NewService returned nil")
	}
	if processor.storage != service {
		t.Error("NewService did not set storage correctly")
	}
}

func TestProcessResult_Structure(t *testing.T) {
	result := &ProcessResult{
		MatchesStored: 5,
		GamesStored:   10,
		DecksStored:   3,
		RanksStored:   2,
		Errors:        []error{},
	}

	if result.MatchesStored != 5 {
		t.Errorf("Expected MatchesStored=5, got %d", result.MatchesStored)
	}
	if result.GamesStored != 10 {
		t.Errorf("Expected GamesStored=10, got %d", result.GamesStored)
	}
	if result.DecksStored != 3 {
		t.Errorf("Expected DecksStored=3, got %d", result.DecksStored)
	}
	if result.RanksStored != 2 {
		t.Errorf("Expected RanksStored=2, got %d", result.RanksStored)
	}
	if len(result.Errors) != 0 {
		t.Errorf("Expected 0 errors, got %d", len(result.Errors))
	}
}

func TestProcessCollection_FromDecks(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// First, store some decks with cards to the database
	// Create a deck directly in the database
	now := time.Now()
	_, err := service.GetDB().ExecContext(ctx, `
		INSERT INTO decks (id, account_id, name, format, source, created_at, modified_at)
		VALUES ('test-deck-1', 1, 'Test Deck', 'Standard', 'arena', ?, ?)
	`, now, now)
	if err != nil {
		t.Fatalf("Failed to insert test deck: %v", err)
	}

	// Add cards to the deck
	_, err = service.GetDB().ExecContext(ctx, `
		INSERT INTO deck_cards (deck_id, card_id, quantity, board)
		VALUES
			('test-deck-1', 12345, 4, 'main'),
			('test-deck-1', 67890, 2, 'main'),
			('test-deck-1', 11111, 1, 'sideboard')
	`)
	if err != nil {
		t.Fatalf("Failed to insert deck cards: %v", err)
	}

	// Process collection
	processor := NewService(service)
	result := &ProcessResult{}
	err = processor.processCollection(ctx, result)
	if err != nil {
		t.Fatalf("processCollection failed: %v", err)
	}

	// Verify collection was updated
	if result.CollectionNewCards != 3 {
		t.Errorf("Expected 3 new cards, got %d", result.CollectionNewCards)
	}

	// Verify quantities
	collection, err := service.CollectionRepo().GetAll(ctx)
	if err != nil {
		t.Fatalf("Failed to get collection: %v", err)
	}

	if collection[12345] != 4 {
		t.Errorf("Expected card 12345 to have quantity 4, got %d", collection[12345])
	}
	if collection[67890] != 2 {
		t.Errorf("Expected card 67890 to have quantity 2, got %d", collection[67890])
	}
	if collection[11111] != 1 {
		t.Errorf("Expected card 11111 to have quantity 1, got %d", collection[11111])
	}
}

func TestProcessCollection_CapAt4Copies(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create a deck with more than 4 copies across multiple decks
	now := time.Now()
	_, err := service.GetDB().ExecContext(ctx, `
		INSERT INTO decks (id, account_id, name, format, source, created_at, modified_at)
		VALUES
			('test-deck-1', 1, 'Test Deck 1', 'Standard', 'arena', ?, ?),
			('test-deck-2', 1, 'Test Deck 2', 'Standard', 'arena', ?, ?)
	`, now, now, now, now)
	if err != nil {
		t.Fatalf("Failed to insert test decks: %v", err)
	}

	// Add cards - total should exceed 4
	_, err = service.GetDB().ExecContext(ctx, `
		INSERT INTO deck_cards (deck_id, card_id, quantity, board)
		VALUES
			('test-deck-1', 12345, 4, 'main'),
			('test-deck-2', 12345, 4, 'main')
	`)
	if err != nil {
		t.Fatalf("Failed to insert deck cards: %v", err)
	}

	// Process collection
	processor := NewService(service)
	result := &ProcessResult{}
	err = processor.processCollection(ctx, result)
	if err != nil {
		t.Fatalf("processCollection failed: %v", err)
	}

	// Verify collection was capped at 4
	collection, err := service.CollectionRepo().GetAll(ctx)
	if err != nil {
		t.Fatalf("Failed to get collection: %v", err)
	}

	if collection[12345] != 4 {
		t.Errorf("Expected card 12345 to be capped at 4, got %d", collection[12345])
	}
}

func TestProcessCollection_NoChanges(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Process empty collection - should not error
	processor := NewService(service)
	result := &ProcessResult{}
	err := processor.processCollection(ctx, result)
	if err != nil {
		t.Fatalf("processCollection failed on empty data: %v", err)
	}

	// Verify no changes reported
	if result.CollectionNewCards != 0 {
		t.Errorf("Expected 0 new cards, got %d", result.CollectionNewCards)
	}
	if result.CollectionCardsAdded != 0 {
		t.Errorf("Expected 0 cards added, got %d", result.CollectionCardsAdded)
	}
}

func TestProcessCollection_DryRunMode(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create a deck with cards
	now := time.Now()
	_, err := service.GetDB().ExecContext(ctx, `
		INSERT INTO decks (id, account_id, name, format, source, created_at, modified_at)
		VALUES ('test-deck-1', 1, 'Test Deck', 'Standard', 'arena', ?, ?)
	`, now, now)
	if err != nil {
		t.Fatalf("Failed to insert test deck: %v", err)
	}

	_, err = service.GetDB().ExecContext(ctx, `
		INSERT INTO deck_cards (deck_id, card_id, quantity, board)
		VALUES ('test-deck-1', 12345, 4, 'main')
	`)
	if err != nil {
		t.Fatalf("Failed to insert deck cards: %v", err)
	}

	// Enable dry run mode
	processor := NewService(service)
	processor.SetDryRun(true)

	result := &ProcessResult{}
	err = processor.processCollection(ctx, result)
	if err != nil {
		t.Fatalf("processCollection failed in dry run mode: %v", err)
	}

	// Verify collection was NOT updated
	collection, err := service.CollectionRepo().GetAll(ctx)
	if err != nil {
		t.Fatalf("Failed to get collection: %v", err)
	}

	if len(collection) != 0 {
		t.Errorf("Expected empty collection in dry run mode, got %d cards", len(collection))
	}
}

func TestSplitCompletedDraftSessions_NewDraftAfterCompleted(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create a completed draft session in the database
	now := time.Now()
	_, err := service.GetDB().ExecContext(ctx, `
		INSERT INTO draft_sessions (id, event_name, set_code, draft_type, start_time, status, total_picks, created_at, updated_at)
		VALUES ('QuickDraft_TLA_20251127', 'QuickDraft_TLA_20251127', 'TLA', 'QuickDraft', ?, 'completed', 42, ?, ?)
	`, now, now, now)
	if err != nil {
		t.Fatalf("Failed to insert completed draft session: %v", err)
	}

	// Add 42 picks to make it complete
	for i := 0; i < 42; i++ {
		packNum := i / 14
		pickNum := (i % 14) + 1
		_, err = service.GetDB().ExecContext(ctx, `
			INSERT INTO draft_picks (session_id, pack_number, pick_number, card_id, timestamp)
			VALUES ('QuickDraft_TLA_20251127', ?, ?, '12345', ?)
		`, packNum, pickNum, now)
		if err != nil {
			t.Fatalf("Failed to insert pick %d: %v", i, err)
		}
	}

	// Create events for a NEW draft with the same event name but P0P0 pack data
	events := map[string][]*logreader.DraftSessionEvent{
		"QuickDraft_TLA_20251127": {
			{
				Type:       "status_updated",
				EventName:  "QuickDraft_TLA_20251127",
				PackNumber: 0,
				PickNumber: 0,
				DraftPack:  []string{"card1", "card2", "card3"},
				Timestamp:  now,
			},
		},
	}

	// Process with splitCompletedDraftSessions
	processor := NewService(service)
	result := processor.splitCompletedDraftSessions(ctx, events)

	// Verify a new session ID was generated
	if len(result) != 1 {
		t.Fatalf("Expected 1 result group, got %d", len(result))
	}

	// The new session ID should NOT be the original
	for newSessionID := range result {
		if newSessionID == "QuickDraft_TLA_20251127" {
			t.Error("Expected new session ID to be different from completed session")
		}
		// Should start with the original prefix
		if len(newSessionID) < len("QuickDraft_TLA_20251127_") {
			t.Errorf("Expected new session ID to start with original prefix, got %s", newSessionID)
		}
	}
}

func TestSplitCompletedDraftSessions_InProgressSession(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create an in-progress draft session
	now := time.Now()
	_, err := service.GetDB().ExecContext(ctx, `
		INSERT INTO draft_sessions (id, event_name, set_code, draft_type, start_time, status, total_picks, created_at, updated_at)
		VALUES ('QuickDraft_TLA_20251127', 'QuickDraft_TLA_20251127', 'TLA', 'QuickDraft', ?, 'in_progress', 42, ?, ?)
	`, now, now, now)
	if err != nil {
		t.Fatalf("Failed to insert in-progress draft session: %v", err)
	}

	// Add only 10 picks (not complete)
	for i := 0; i < 10; i++ {
		_, err = service.GetDB().ExecContext(ctx, `
			INSERT INTO draft_picks (session_id, pack_number, pick_number, card_id, timestamp)
			VALUES ('QuickDraft_TLA_20251127', 0, ?, '12345', ?)
		`, i+1, now)
		if err != nil {
			t.Fatalf("Failed to insert pick %d: %v", i, err)
		}
	}

	// Create events with P0P11 pack data (continuing the draft)
	events := map[string][]*logreader.DraftSessionEvent{
		"QuickDraft_TLA_20251127": {
			{
				Type:       "status_updated",
				EventName:  "QuickDraft_TLA_20251127",
				PackNumber: 0,
				PickNumber: 11,
				DraftPack:  []string{"card1", "card2", "card3"},
				Timestamp:  now,
			},
		},
	}

	// Process with splitCompletedDraftSessions
	processor := NewService(service)
	result := processor.splitCompletedDraftSessions(ctx, events)

	// Verify the original session ID is preserved (no split)
	if len(result) != 1 {
		t.Fatalf("Expected 1 result group, got %d", len(result))
	}

	for sessionID := range result {
		if sessionID != "QuickDraft_TLA_20251127" {
			t.Errorf("Expected session ID to remain 'QuickDraft_TLA_20251127', got %s", sessionID)
		}
	}
}

func TestSplitCompletedDraftSessions_UUIDPassthrough(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// UUID-based sessions (Premier Draft) should pass through unchanged
	events := map[string][]*logreader.DraftSessionEvent{
		"73e1c7a3-75ee-4b38-b32b-d6854e5c6c9c": {
			{
				Type:       "status_updated",
				SessionID:  "73e1c7a3-75ee-4b38-b32b-d6854e5c6c9c",
				PackNumber: 0,
				PickNumber: 0,
				DraftPack:  []string{"card1", "card2", "card3"},
				Timestamp:  time.Now(),
			},
		},
	}

	processor := NewService(service)
	result := processor.splitCompletedDraftSessions(ctx, events)

	// UUID sessions should pass through unchanged
	if len(result) != 1 {
		t.Fatalf("Expected 1 result group, got %d", len(result))
	}

	for sessionID := range result {
		if sessionID != "73e1c7a3-75ee-4b38-b32b-d6854e5c6c9c" {
			t.Errorf("Expected UUID session ID to remain unchanged, got %s", sessionID)
		}
	}
}

func TestSplitCompletedDraftSessions_ReuseExistingInProgressSession(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create a completed session for the base event name
	completedSession := &models.DraftSession{
		ID:         "QuickDraft_TLA_20251127",
		EventName:  "QuickDraft_TLA_20251127",
		SetCode:    "TLA",
		DraftType:  "QuickDraft",
		StartTime:  time.Now().Add(-2 * time.Hour),
		Status:     "completed",
		TotalPicks: 42,
		CreatedAt:  time.Now().Add(-2 * time.Hour),
		UpdatedAt:  time.Now().Add(-time.Hour),
	}
	if err := service.DraftRepo().CreateSession(ctx, completedSession); err != nil {
		t.Fatalf("Failed to create completed session: %v", err)
	}

	// Create an existing in_progress session with timestamp suffix
	existingInProgressSession := &models.DraftSession{
		ID:         "QuickDraft_TLA_20251127_1234567890",
		EventName:  "QuickDraft_TLA_20251127",
		SetCode:    "TLA",
		DraftType:  "QuickDraft",
		StartTime:  time.Now().Add(-30 * time.Minute),
		Status:     "in_progress",
		TotalPicks: 0,
		CreatedAt:  time.Now().Add(-30 * time.Minute),
		UpdatedAt:  time.Now(),
	}
	if err := service.DraftRepo().CreateSession(ctx, existingInProgressSession); err != nil {
		t.Fatalf("Failed to create in_progress session: %v", err)
	}

	// Simulate new events coming for the base event name with P0P1 pack data
	events := map[string][]*logreader.DraftSessionEvent{
		"QuickDraft_TLA_20251127": {
			{
				Type:       "status_updated",
				EventName:  "QuickDraft_TLA_20251127",
				PackNumber: 0,
				PickNumber: 1,
				DraftPack:  []string{"card1", "card2", "card3"},
				Timestamp:  time.Now(),
			},
		},
	}

	processor := NewService(service)
	result := processor.splitCompletedDraftSessions(ctx, events)

	// Should reuse the existing in_progress session instead of creating a new one
	if len(result) != 1 {
		t.Fatalf("Expected 1 result group, got %d", len(result))
	}

	for sessionID := range result {
		if sessionID != "QuickDraft_TLA_20251127_1234567890" {
			t.Errorf("Expected to reuse existing session 'QuickDraft_TLA_20251127_1234567890', got %s", sessionID)
		}
	}
}

func TestSplitCompletedDraftSessions_MixedOldAndNewEvents(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create a completed draft session in the database
	now := time.Now()
	_, err := service.GetDB().ExecContext(ctx, `
		INSERT INTO draft_sessions (id, event_name, set_code, draft_type, start_time, status, total_picks, created_at, updated_at)
		VALUES ('QuickDraft_TLA_20251127', 'QuickDraft_TLA_20251127', 'TLA', 'QuickDraft', ?, 'completed', 42, ?, ?)
	`, now.Add(-2*time.Hour), now.Add(-2*time.Hour), now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("Failed to insert completed draft session: %v", err)
	}

	// Add 42 picks to make it complete
	for i := 0; i < 42; i++ {
		packNum := i / 14
		pickNum := (i % 14) + 1
		_, err = service.GetDB().ExecContext(ctx, `
			INSERT INTO draft_picks (session_id, pack_number, pick_number, card_id, timestamp)
			VALUES ('QuickDraft_TLA_20251127', ?, ?, '12345', ?)
		`, packNum, pickNum, now.Add(-2*time.Hour))
		if err != nil {
			t.Fatalf("Failed to insert pick %d: %v", i, err)
		}
	}

	// Simulate reprocessing the entire log file: events from BOTH old and new drafts mixed together
	// This is what happens when the daemon restarts with ReadFromStart=true
	events := map[string][]*logreader.DraftSessionEvent{
		"QuickDraft_TLA_20251127": {
			// OLD draft events (42 picks worth)
			{
				Type:        "status_updated",
				EventName:   "QuickDraft_TLA_20251127",
				PackNumber:  0,
				PickNumber:  1,
				DraftPack:   []string{"old1", "old2", "old3", "old4", "old5", "old6", "old7", "old8", "old9", "old10", "old11", "old12", "old13", "old14"},
				PickedCards: []string{}, // First pick of old draft
				Timestamp:   now.Add(-2 * time.Hour),
			},
			{
				Type:         "pick_made",
				EventName:    "QuickDraft_TLA_20251127",
				PackNumber:   0,
				PickNumber:   1,
				SelectedCard: []string{"old1"},
				Timestamp:    now.Add(-2 * time.Hour),
			},
			{
				Type:        "status_updated",
				EventName:   "QuickDraft_TLA_20251127",
				PackNumber:  0,
				PickNumber:  2,
				DraftPack:   []string{"old2", "old3", "old4", "old5", "old6", "old7", "old8", "old9", "old10", "old11", "old12", "old13"},
				PickedCards: []string{"old1"}, // Has picked cards - NOT a new draft
				Timestamp:   now.Add(-2 * time.Hour),
			},
			// ... more old draft events would be here in real scenario ...
			{
				Type:      "ended",
				EventName: "QuickDraft_TLA_20251127",
				Timestamp: now.Add(-time.Hour),
			},
			// NEW draft events (starting fresh)
			{
				Type:        "status_updated",
				EventName:   "QuickDraft_TLA_20251127",
				PackNumber:  0,
				PickNumber:  1,
				DraftPack:   []string{"new1", "new2", "new3", "new4", "new5", "new6", "new7", "new8", "new9", "new10", "new11", "new12", "new13", "new14"},
				PickedCards: []string{}, // Empty! This is a NEW draft
				Timestamp:   now,
			},
			{
				Type:         "pick_made",
				EventName:    "QuickDraft_TLA_20251127",
				PackNumber:   0,
				PickNumber:   1,
				SelectedCard: []string{"new1"},
				Timestamp:    now,
			},
			{
				Type:        "status_updated",
				EventName:   "QuickDraft_TLA_20251127",
				PackNumber:  0,
				PickNumber:  2,
				DraftPack:   []string{"new2", "new3", "new4", "new5", "new6", "new7", "new8", "new9", "new10", "new11", "new12", "new13"},
				PickedCards: []string{"new1"},
				Timestamp:   now,
			},
		},
	}

	processor := NewService(service)
	result := processor.splitCompletedDraftSessions(ctx, events)

	// Should create a new session ID for the new draft
	if len(result) != 1 {
		t.Fatalf("Expected 1 result group, got %d", len(result))
	}

	var newSessionID string
	var newEvents []*logreader.DraftSessionEvent
	for id, evts := range result {
		newSessionID = id
		newEvents = evts
	}

	// The new session ID should NOT be the original (should have timestamp suffix)
	if newSessionID == "QuickDraft_TLA_20251127" {
		t.Error("Expected new session ID to be different from completed session")
	}

	// Should have filtered out old draft events
	// We should only have the new draft events (the one with empty PickedCards and events after it)
	// Expected: status_updated (new P0P1), pick_made (new), status_updated (new P0P2)
	// But we may also include 'started' and 'session_info' events if any
	hasOldEvents := false
	for _, evt := range newEvents {
		// Check if any old draft events are included
		if evt.Type == "ended" {
			hasOldEvents = true
			t.Error("Old draft 'ended' event should have been filtered out")
		}
		// Check for old draft status_updated with PickedCards
		if evt.Type == "status_updated" && len(evt.PickedCards) > 0 && evt.PickedCards[0] == "old1" {
			hasOldEvents = true
			t.Error("Old draft status_updated with PickedCards should have been filtered out")
		}
		// Check for old pick_made
		if evt.Type == "pick_made" && len(evt.SelectedCard) > 0 && evt.SelectedCard[0] == "old1" {
			hasOldEvents = true
			t.Error("Old draft pick_made event should have been filtered out")
		}
	}

	if hasOldEvents {
		t.Errorf("Expected no old draft events in result, but some were found. Total events: %d", len(newEvents))
		for i, evt := range newEvents {
			t.Logf("Event %d: Type=%s, Pack=%d, Pick=%d", i, evt.Type, evt.PackNumber, evt.PickNumber)
		}
	}

	// Verify new draft events are present
	hasNewStatusUpdate := false
	hasNewPickMade := false
	for _, evt := range newEvents {
		if evt.Type == "status_updated" && len(evt.DraftPack) > 0 && evt.DraftPack[0] == "new1" {
			hasNewStatusUpdate = true
		}
		if evt.Type == "pick_made" && len(evt.SelectedCard) > 0 && evt.SelectedCard[0] == "new1" {
			hasNewPickMade = true
		}
	}

	if !hasNewStatusUpdate {
		t.Error("Expected new draft status_updated event to be present")
	}
	if !hasNewPickMade {
		t.Error("Expected new draft pick_made event to be present")
	}
}

func TestSplitCompletedDraftSessions_OngoingPicksRouteToTimestampedSession(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create a completed session for the base event name
	completedSession := &models.DraftSession{
		ID:         "QuickDraft_TLA_20251127",
		EventName:  "QuickDraft_TLA_20251127",
		SetCode:    "TLA",
		DraftType:  "QuickDraft",
		StartTime:  time.Now().Add(-2 * time.Hour),
		Status:     "completed",
		TotalPicks: 42,
		CreatedAt:  time.Now().Add(-2 * time.Hour),
		UpdatedAt:  time.Now().Add(-time.Hour),
	}
	if err := service.DraftRepo().CreateSession(ctx, completedSession); err != nil {
		t.Fatalf("Failed to create completed session: %v", err)
	}

	// Create an existing in_progress session with timestamp suffix (the new draft)
	existingInProgressSession := &models.DraftSession{
		ID:         "QuickDraft_TLA_20251127_1234567890",
		EventName:  "QuickDraft_TLA_20251127",
		SetCode:    "TLA",
		DraftType:  "QuickDraft",
		StartTime:  time.Now().Add(-30 * time.Minute),
		Status:     "in_progress",
		TotalPicks: 42,
		CreatedAt:  time.Now().Add(-30 * time.Minute),
		UpdatedAt:  time.Now(),
	}
	if err := service.DraftRepo().CreateSession(ctx, existingInProgressSession); err != nil {
		t.Fatalf("Failed to create in_progress session: %v", err)
	}

	// Simulate ongoing picks coming in - these do NOT have P0P0/P0P1 pack data
	// This is what happens when the user makes a pick mid-draft
	events := map[string][]*logreader.DraftSessionEvent{
		"QuickDraft_TLA_20251127": {
			{
				Type:        "status_updated",
				EventName:   "QuickDraft_TLA_20251127",
				PackNumber:  0,
				PickNumber:  5, // Pick 5, not pick 0 or 1
				DraftPack:   []string{"card1", "card2", "card3"},
				PickedCards: []string{"prev1", "prev2", "prev3", "prev4"}, // Has previous picks
				Timestamp:   time.Now(),
			},
			{
				Type:         "pick_made",
				EventName:    "QuickDraft_TLA_20251127",
				PackNumber:   0,
				PickNumber:   5,
				SelectedCard: []string{"card1"},
				Timestamp:    time.Now(),
			},
		},
	}

	processor := NewService(service)
	result := processor.splitCompletedDraftSessions(ctx, events)

	// Should route to the existing in_progress session with timestamp suffix
	if len(result) != 1 {
		t.Fatalf("Expected 1 result group, got %d", len(result))
	}

	for sessionID, evts := range result {
		if sessionID != "QuickDraft_TLA_20251127_1234567890" {
			t.Errorf("Expected events to be routed to timestamped session 'QuickDraft_TLA_20251127_1234567890', got %s", sessionID)
		}
		if len(evts) != 2 {
			t.Errorf("Expected 2 events to be routed, got %d", len(evts))
		}
	}
}

func TestFilterNewDraftEvents(t *testing.T) {
	service, cleanup := setupTestService(t)
	defer cleanup()

	processor := NewService(service)

	events := []*logreader.DraftSessionEvent{
		// Old draft events
		{Type: "started", PackNumber: 0, PickNumber: 0},
		{Type: "status_updated", PackNumber: 0, PickNumber: 1, DraftPack: []string{"a"}, PickedCards: []string{}},
		{Type: "pick_made", PackNumber: 0, PickNumber: 1, SelectedCard: []string{"a"}},
		{Type: "status_updated", PackNumber: 0, PickNumber: 2, DraftPack: []string{"b"}, PickedCards: []string{"a"}},
		{Type: "ended"},
		// New draft events - index 5 is the new draft start (empty PickedCards at P0P1)
		{Type: "status_updated", PackNumber: 0, PickNumber: 1, DraftPack: []string{"x", "y", "z"}, PickedCards: []string{}},
		{Type: "pick_made", PackNumber: 0, PickNumber: 1, SelectedCard: []string{"x"}},
	}

	filtered := processor.filterNewDraftEvents(events, 5)

	// Should have:
	// - started (kept as control event)
	// - new status_updated at index 5
	// - new pick_made at index 6
	// Should NOT have:
	// - old status_updated at index 1, 3
	// - old pick_made at index 2
	// - old ended at index 4

	expectedTypes := map[string]int{
		"started":        1,
		"status_updated": 1,
		"pick_made":      1,
	}

	actualTypes := make(map[string]int)
	for _, evt := range filtered {
		actualTypes[evt.Type]++
	}

	for evtType, expectedCount := range expectedTypes {
		if actualTypes[evtType] != expectedCount {
			t.Errorf("Expected %d %s events, got %d", expectedCount, evtType, actualTypes[evtType])
		}
	}

	// Should NOT have ended event
	if actualTypes["ended"] > 0 {
		t.Error("Should not include 'ended' event from old draft")
	}
}

// Benchmark tests
func BenchmarkProcessLogEntries(b *testing.B) {
	service, cleanup := setupTestService(&testing.T{})
	defer cleanup()

	ctx := context.Background()
	processor := NewService(service)

	// Create sample entries
	entries := []*logreader.LogEntry{
		{
			IsJSON:    true,
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			JSON: map[string]interface{}{
				"constructedSeasonOrdinal": float64(83),
				"constructedClass":         "Gold",
				"constructedLevel":         float64(4),
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = processor.ProcessLogEntries(ctx, entries)
	}
}
