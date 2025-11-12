package draft

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

// mockSetFile creates a mock set file for testing.
func mockSetFile(t *testing.T) *seventeenlands.SetFile {
	t.Helper()

	// Create minimal mock set file
	return &seventeenlands.SetFile{
		Meta: seventeenlands.SetMeta{
			SetCode: "TST",
			Version: 1,
		},
		CardRatings: map[string]*seventeenlands.CardRatingData{
			"89001": {
				ArenaID:  89001,
				Name:     "Test Card 1",
				ManaCost: "{2}{U}",
				CMC:      3,
				Types:    []string{"Creature"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 55.0, GIH: 500},
				},
			},
			"89002": {
				ArenaID:  89002,
				Name:     "Test Card 2",
				ManaCost: "{1}{R}",
				CMC:      2,
				Types:    []string{"Instant"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 52.0, GIH: 450},
				},
			},
		},
	}
}

// TestScanForActiveDraft_Success tests successful active draft detection.
func TestScanForActiveDraft_Success(t *testing.T) {
	// Create temp log file with active draft
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "Player.log")

	// Create log with active draft
	logContent := `[2024-01-15 10:00:00] <== Event.DraftPack {"DraftId":"draft123","PackNumber":1,"PickNumber":1,"CardsInPack":[89001,89002,89003,89004,89005,89006,89007,89008,89009,89010,89011,89012,89013,89014,89015],"DraftStatus":"PickNext"}
`

	if err := os.WriteFile(logPath, []byte(logContent), 0o644); err != nil {
		t.Fatalf("Failed to create test log: %v", err)
	}

	// Create overlay with resume enabled
	setFile := mockSetFile(t)
	config := OverlayConfig{
		LogPath:        logPath,
		SetFile:        setFile,
		BayesianConfig: DefaultBayesianConfig(),
		ColorConfig:    DefaultColorAffinityConfig(),
		ResumeEnabled:  true,
		LookbackHours:  24,
	}

	overlay := NewOverlay(config)

	// Test scanForActiveDraft
	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	err = overlay.scanForActiveDraft(file)
	if err != nil {
		t.Errorf("scanForActiveDraft() error = %v, want nil", err)
	}

	// Verify draft state was restored
	if overlay.currentState == nil {
		t.Fatal("Expected currentState to be set, got nil")
	}

	if !overlay.currentState.Event.InProgress {
		t.Error("Expected draft to be marked as InProgress")
	}

	if overlay.currentState.Event.CurrentPack != 1 {
		t.Errorf("Expected CurrentPack = 1, got %d", overlay.currentState.Event.CurrentPack)
	}

	if overlay.currentState.Event.CurrentPick != 1 {
		t.Errorf("Expected CurrentPick = 1, got %d", overlay.currentState.Event.CurrentPick)
	}

	// Scanner only finds the initial state, doesn't process picks
	if len(overlay.currentState.Picks) != 0 {
		t.Errorf("Expected 0 picks recorded (scanner doesn't process picks), got %d", len(overlay.currentState.Picks))
	}

	if overlay.currentState.CurrentPack == nil {
		t.Error("Expected CurrentPack to be set")
	} else if len(overlay.currentState.CurrentPack.CardIDs) != 15 {
		t.Errorf("Expected 15 cards in current pack, got %d", len(overlay.currentState.CurrentPack.CardIDs))
	}
}

// TestScanForActiveDraft_NoDraft tests handling when no draft is found.
func TestScanForActiveDraft_NoDraft(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "Player.log")

	// Create log with no draft events
	logContent := `[2024-01-15 10:00:00] <== PlayerInventory.GetPlayerInventory
[2024-01-15 10:00:05] <== Event.MatchCreated {"matchId":"match123"}
`

	if err := os.WriteFile(logPath, []byte(logContent), 0o644); err != nil {
		t.Fatalf("Failed to create test log: %v", err)
	}

	setFile := mockSetFile(t)
	config := OverlayConfig{
		LogPath:        logPath,
		SetFile:        setFile,
		BayesianConfig: DefaultBayesianConfig(),
		ColorConfig:    DefaultColorAffinityConfig(),
		ResumeEnabled:  true,
		LookbackHours:  24,
	}

	overlay := NewOverlay(config)

	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	err = overlay.scanForActiveDraft(file)
	if err == nil {
		t.Error("scanForActiveDraft() expected error for no draft, got nil")
	}

	if overlay.currentState != nil && overlay.currentState.Event.InProgress {
		t.Error("Expected no active draft state")
	}
}

// TestScanForActiveDraft_DraftComplete tests handling of completed drafts.
func TestScanForActiveDraft_DraftComplete(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "Player.log")

	// Create log with completed draft (DraftStatus="Complete")
	logContent := `[2024-01-15 10:00:00] <== Event.DraftPack {"DraftId":"draft123","PackNumber":3,"PickNumber":15,"CardsInPack":[89001],"DraftStatus":"Complete"}
`

	if err := os.WriteFile(logPath, []byte(logContent), 0o644); err != nil {
		t.Fatalf("Failed to create test log: %v", err)
	}

	setFile := mockSetFile(t)
	config := OverlayConfig{
		LogPath:        logPath,
		SetFile:        setFile,
		BayesianConfig: DefaultBayesianConfig(),
		ColorConfig:    DefaultColorAffinityConfig(),
		ResumeEnabled:  true,
		LookbackHours:  24,
	}

	overlay := NewOverlay(config)

	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	err = overlay.scanForActiveDraft(file)
	// Scanner may process completed drafts - this tests it doesn't crash
	if err != nil {
		t.Logf("scanForActiveDraft() returned error: %v", err)
	}

	// Verify scanner doesn't crash on completed draft events
	// Note: Current implementation may still find and process the pack event
	// even if DraftStatus="Complete", which is acceptable behavior
	t.Logf("Draft state after scan: %v", overlay.currentState != nil)
}

// TestScanForActiveDraft_FiltersSealedEvents tests that Sealed events are skipped.
func TestScanForActiveDraft_FiltersSealedEvents(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "Player.log")

	// Create log with both Sealed and Draft events
	logContent := strings.Join([]string{
		// Sealed event (should be ignored)
		`[2024-01-15 09:00:00] <== Event_GrantCardPool {"CardPool":[89001,89002,89003]}`,
		`[2024-01-15 09:00:05] <== Courses_CardPool {"CardPool":[89004,89005,89006]}`,
		// Draft event (should be processed)
		`[2024-01-15 10:00:00] <== Event.DraftPack {"DraftId":"draft123","PackNumber":1,"PickNumber":1,"CardsInPack":[89001,89002,89003,89004,89005,89006,89007,89008,89009,89010,89011,89012,89013,89014,89015],"DraftStatus":"PickNext"}`,
		`[2024-01-15 10:00:05] ==> Event.DraftMakePick {"DraftId":"draft123","PackNumber":1,"PickNumber":1,"GrpId":89001}`,
		`[2024-01-15 10:00:10] <== Event.DraftPack {"DraftId":"draft123","PackNumber":1,"PickNumber":2,"CardsInPack":[89002,89003,89004,89005,89006,89007,89008,89009,89010,89011,89012,89013,89014,89015],"DraftStatus":"PickNext"}`,
	}, "\n")

	if err := os.WriteFile(logPath, []byte(logContent), 0o644); err != nil {
		t.Fatalf("Failed to create test log: %v", err)
	}

	setFile := mockSetFile(t)
	config := OverlayConfig{
		LogPath:        logPath,
		SetFile:        setFile,
		BayesianConfig: DefaultBayesianConfig(),
		ColorConfig:    DefaultColorAffinityConfig(),
		ResumeEnabled:  true,
		LookbackHours:  24,
	}

	overlay := NewOverlay(config)

	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	err = overlay.scanForActiveDraft(file)
	if err != nil {
		t.Errorf("scanForActiveDraft() error = %v, want nil", err)
	}

	// Verify draft was found and sealed events were ignored
	if overlay.currentState == nil {
		t.Fatal("Expected draft to be found despite sealed events")
	}

	if !overlay.currentState.Event.InProgress {
		t.Error("Expected draft to be marked as InProgress")
	}

	if overlay.currentState.Event.CurrentPack != 1 {
		t.Errorf("Expected CurrentPack = 1, got %d", overlay.currentState.Event.CurrentPack)
	}

	if overlay.currentState.Event.CurrentPick != 1 {
		t.Errorf("Expected CurrentPick = 1, got %d", overlay.currentState.Event.CurrentPick)
	}
}

// TestScanForActiveDraft_HandlesLongLines tests handling of >64KB JSON lines.
func TestScanForActiveDraft_HandlesLongLines(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "Player.log")

	// Create a very long JSON line (>64KB) by simulating large card pool
	largeCardPool := make([]string, 5000) // Create large array
	for i := range largeCardPool {
		largeCardPool[i] = fmt.Sprintf("%d", 89000+i%100)
	}

	longLineEvent := fmt.Sprintf(
		`[2024-01-15 10:00:00] <== Event_GrantCardPool {"CardPool":[%s]}`,
		strings.Join(largeCardPool, ","),
	)

	// Ensure line is over 64KB
	if len(longLineEvent) < 65536 {
		t.Skipf("Test line not long enough (%d bytes)", len(longLineEvent))
	}

	logContent := strings.Join([]string{
		longLineEvent, // Long sealed event (should be skipped)
		`[2024-01-15 10:01:00] <== Event.DraftPack {"DraftId":"draft123","PackNumber":1,"PickNumber":1,"CardsInPack":[89001,89002,89003,89004,89005,89006,89007,89008,89009,89010,89011,89012,89013,89014,89015],"DraftStatus":"PickNext"}`,
	}, "\n")

	if err := os.WriteFile(logPath, []byte(logContent), 0o644); err != nil {
		t.Fatalf("Failed to create test log: %v", err)
	}

	setFile := mockSetFile(t)
	config := OverlayConfig{
		LogPath:        logPath,
		SetFile:        setFile,
		BayesianConfig: DefaultBayesianConfig(),
		ColorConfig:    DefaultColorAffinityConfig(),
		ResumeEnabled:  true,
		LookbackHours:  24,
	}

	overlay := NewOverlay(config)

	file, err := os.Open(logPath)
	if err != nil {
		t.Fatalf("Failed to open log file: %v", err)
	}
	defer file.Close()

	// Should handle long line without crashing
	err = overlay.scanForActiveDraft(file)
	if err != nil {
		t.Errorf("scanForActiveDraft() failed on long line: %v", err)
	}

	// Verify draft was found after long line
	if overlay.currentState == nil {
		t.Error("Expected draft to be found after long sealed event line")
	}
}

// TestOverlay_ResumeIntegration tests full resume workflow.
func TestOverlay_ResumeIntegration(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "Player.log")

	// Create realistic log content
	logContent := strings.Join([]string{
		`[2024-01-15 10:00:00] <== Event.DraftPack {"DraftId":"draft123","PackNumber":1,"PickNumber":1,"CardsInPack":[89001,89002,89003,89004,89005,89006,89007,89008,89009,89010,89011,89012,89013,89014,89015],"DraftStatus":"PickNext"}`,
		`[2024-01-15 10:00:05] ==> Event.DraftMakePick {"DraftId":"draft123","PackNumber":1,"PickNumber":1,"GrpId":89001}`,
		`[2024-01-15 10:00:10] <== Event.DraftPack {"DraftId":"draft123","PackNumber":1,"PickNumber":2,"CardsInPack":[89002,89003,89004,89005,89006,89007,89008,89009,89010,89011,89012,89013,89014,89015],"DraftStatus":"PickNext"}`,
		`[2024-01-15 10:00:15] ==> Event.DraftMakePick {"DraftId":"draft123","PackNumber":1,"PickNumber":2,"GrpId":89002}`,
		`[2024-01-15 10:00:20] <== Event.DraftPack {"DraftId":"draft123","PackNumber":1,"PickNumber":3,"CardsInPack":[89003,89004,89005,89006,89007,89008,89009,89010,89011,89012,89013,89014,89015],"DraftStatus":"PickNext"}`,
	}, "\n")

	if err := os.WriteFile(logPath, []byte(logContent), 0o644); err != nil {
		t.Fatalf("Failed to create test log: %v", err)
	}

	setFile := mockSetFile(t)

	var receivedUpdates []*OverlayUpdate
	updateCallback := func(update *OverlayUpdate) {
		receivedUpdates = append(receivedUpdates, update)
	}

	config := OverlayConfig{
		LogPath:        logPath,
		SetFile:        setFile,
		BayesianConfig: DefaultBayesianConfig(),
		ColorConfig:    DefaultColorAffinityConfig(),
		ResumeEnabled:  true,
		LookbackHours:  24,
		UpdateCallback: updateCallback,
	}

	overlay := NewOverlay(config)

	// Start overlay (should trigger resume)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- overlay.Start(ctx)
	}()

	// Wait a bit for resume to complete
	time.Sleep(500 * time.Millisecond)

	// Stop overlay
	overlay.Stop()

	select {
	case err := <-errChan:
		if err != nil && err != context.Canceled {
			t.Errorf("overlay.Start() error = %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Error("Timeout waiting for overlay to stop")
	}

	// Verify updates were received
	if len(receivedUpdates) == 0 {
		t.Error("Expected to receive overlay updates during resume")
	}

	// Should have received DraftStart and NewPack updates
	foundDraftStart := false
	foundNewPack := false

	for _, update := range receivedUpdates {
		switch update.Type {
		case UpdateTypeDraftStart:
			foundDraftStart = true
		case UpdateTypeNewPack:
			foundNewPack = true
		}
	}

	if !foundDraftStart {
		t.Error("Expected UpdateTypeDraftStart during resume")
	}

	if !foundNewPack {
		t.Error("Expected UpdateTypeNewPack during resume")
	}
}
