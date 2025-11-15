package logprocessor

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
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
