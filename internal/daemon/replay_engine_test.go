package daemon

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
)

func TestNewReplayEngine(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	if engine == nil {
		t.Fatal("NewReplayEngine returned nil")
	}

	if engine.service != service {
		t.Error("Engine service not set correctly")
	}

	if engine.speed != 1.0 {
		t.Errorf("Expected default speed 1.0, got %.2f", engine.speed)
	}

	if engine.filterType != "all" {
		t.Errorf("Expected default filter 'all', got %s", engine.filterType)
	}

	if engine.pauseChan == nil {
		t.Error("pauseChan is nil")
	}

	if engine.resumeChan == nil {
		t.Error("resumeChan is nil")
	}

	if engine.stopChan == nil {
		t.Error("stopChan is nil")
	}

	if engine.ctx == nil {
		t.Error("ctx is nil")
	}

	if engine.cancel == nil {
		t.Error("cancel is nil")
	}
}

func TestReplayEngine_GetStatus_NotActive(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	status := engine.GetStatus()

	isActive, ok := status["isActive"].(bool)
	if !ok {
		t.Fatal("isActive not found in status")
	}

	if isActive {
		t.Error("Expected isActive to be false")
	}
}

func TestReplayEngine_Pause_NotActive(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	err := engine.Pause()
	if err == nil {
		t.Error("Expected error when pausing inactive replay")
	}

	if err.Error() != "replay not active" {
		t.Errorf("Expected 'replay not active' error, got: %v", err)
	}
}

func TestReplayEngine_Resume_NotActive(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	err := engine.Resume()
	if err == nil {
		t.Error("Expected error when resuming inactive replay")
	}

	if err.Error() != "replay not active" {
		t.Errorf("Expected 'replay not active' error, got: %v", err)
	}
}

func TestReplayEngine_Stop_NotActive(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	err := engine.Stop()
	if err == nil {
		t.Error("Expected error when stopping inactive replay")
	}

	if err.Error() != "replay not active" {
		t.Errorf("Expected 'replay not active' error, got: %v", err)
	}
}

func TestReplayEngine_Start_InvalidPath(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	err := engine.Start([]string{"/nonexistent/path/to/log.log"}, 1.0, "all", false)
	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestReplayEngine_Start_AlreadyActive(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	// Set isActive manually to simulate running replay
	engine.mu.Lock()
	engine.isActive = true
	engine.mu.Unlock()

	err := engine.Start([]string{"/some/path"}, 1.0, "all", false)
	if err == nil {
		t.Error("Expected error when starting already active replay")
	}

	if err.Error() != "replay already active" {
		t.Errorf("Expected 'replay already active' error, got: %v", err)
	}
}

func TestReplayEngine_isDraftEntry(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	tests := []struct {
		name     string
		raw      string
		expected bool
	}{
		{
			name:     "Draft status entry",
			raw:      `{"DraftStatus": "InProgress"}`,
			expected: true,
		},
		{
			name:     "Draft pick entry",
			raw:      `[UnityCrossThreadLogger]==> DraftPick {"id": 1234}`,
			expected: true,
		},
		{
			name:     "Draft make pick entry",
			raw:      `[UnityCrossThreadLogger]==> DraftMakePick {"cardId": 5678}`,
			expected: true,
		},
		{
			name:     "Draft dot entry",
			raw:      `Draft.Notify {"event": "start"}`,
			expected: true,
		},
		{
			name:     "Non-draft entry",
			raw:      `[UnityCrossThreadLogger]MatchGameRoomStateChanged`,
			expected: false,
		},
		{
			name:     "Empty entry",
			raw:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &logreader.LogEntry{Raw: tt.raw}
			result := engine.isDraftEntry(entry)
			if result != tt.expected {
				t.Errorf("isDraftEntry() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestReplayEngine_isDraftPickEntry(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	tests := []struct {
		name     string
		raw      string
		expected bool
	}{
		{
			name:     "DraftPick with arrow",
			raw:      `[UnityCrossThreadLogger]==> DraftPick {"id": 1234}`,
			expected: true,
		},
		{
			name:     "DraftMakePick with arrow",
			raw:      `[UnityCrossThreadLogger]==> DraftMakePick {"cardId": 5678}`,
			expected: true,
		},
		{
			name:     "DraftPick without arrow",
			raw:      `DraftPick {"id": 1234}`,
			expected: false,
		},
		{
			name:     "DraftStatus (not a pick)",
			raw:      `DraftStatus {"status": "InProgress"}`,
			expected: false,
		},
		{
			name:     "Non-draft entry",
			raw:      `MatchGameRoomStateChanged`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &logreader.LogEntry{Raw: tt.raw}
			result := engine.isDraftPickEntry(entry)
			if result != tt.expected {
				t.Errorf("isDraftPickEntry() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestReplayEngine_isMatchEntry(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	tests := []struct {
		name     string
		raw      string
		expected bool
	}{
		{
			name:     "MatchGameRoomStateChanged",
			raw:      `[UnityCrossThreadLogger]MatchGameRoomStateChanged`,
			expected: true,
		},
		{
			name:     "GameRoomStateChanged",
			raw:      `GameRoomStateChanged {"state": "active"}`,
			expected: true,
		},
		{
			name:     "EventMatchCreated",
			raw:      `EventMatchCreated {"matchId": "123"}`,
			expected: true,
		},
		{
			name:     "Draft entry",
			raw:      `DraftPick {"id": 1234}`,
			expected: false,
		},
		{
			name:     "Empty entry",
			raw:      "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &logreader.LogEntry{Raw: tt.raw}
			result := engine.isMatchEntry(entry)
			if result != tt.expected {
				t.Errorf("isMatchEntry() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestReplayEngine_isEventEntry(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	tests := []struct {
		name     string
		raw      string
		expected bool
	}{
		{
			name:     "Event underscore",
			raw:      `Event_Join {"eventId": "123"}`,
			expected: true,
		},
		{
			name:     "EventJoin",
			raw:      `EventJoin {"eventId": "456"}`,
			expected: true,
		},
		{
			name:     "EventGetCourses",
			raw:      `EventGetCourses {"courses": []}`,
			expected: true,
		},
		{
			name:     "Draft entry",
			raw:      `DraftPick {"id": 1234}`,
			expected: false,
		},
		{
			name:     "Match entry",
			raw:      `MatchGameRoomStateChanged`,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &logreader.LogEntry{Raw: tt.raw}
			result := engine.isEventEntry(entry)
			if result != tt.expected {
				t.Errorf("isEventEntry() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestReplayEngine_filterEntries_All(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	entries := []*logreader.LogEntry{
		{Raw: "entry1", IsJSON: true},
		{Raw: "entry2", IsJSON: true},
		{Raw: "entry3", IsJSON: true},
	}

	result := engine.filterEntries(entries, "all")

	if len(result) != len(entries) {
		t.Errorf("Expected %d entries, got %d", len(entries), len(result))
	}
}

func TestReplayEngine_filterEntries_Draft(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	entries := []*logreader.LogEntry{
		{Raw: "DraftPick {}", IsJSON: true},
		{Raw: "MatchGameRoomStateChanged", IsJSON: true},
		{Raw: "DraftStatus {}", IsJSON: true},
		{Raw: "EventJoin {}", IsJSON: true},
	}

	result := engine.filterEntries(entries, "draft")

	if len(result) != 2 {
		t.Errorf("Expected 2 draft entries, got %d", len(result))
	}
}

func TestReplayEngine_filterEntries_Match(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	entries := []*logreader.LogEntry{
		{Raw: "DraftPick {}", IsJSON: true},
		{Raw: "MatchGameRoomStateChanged", IsJSON: true},
		{Raw: "GameRoomStateChanged {}", IsJSON: true},
		{Raw: "EventJoin {}", IsJSON: true},
	}

	result := engine.filterEntries(entries, "match")

	if len(result) != 2 {
		t.Errorf("Expected 2 match entries, got %d", len(result))
	}
}

func TestReplayEngine_filterEntries_Event(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	entries := []*logreader.LogEntry{
		{Raw: "DraftPick {}", IsJSON: true},
		{Raw: "EventJoin {}", IsJSON: true},
		{Raw: "EventGetCourses {}", IsJSON: true},
		{Raw: "MatchGameRoomStateChanged", IsJSON: true},
	}

	result := engine.filterEntries(entries, "event")

	if len(result) != 2 {
		t.Errorf("Expected 2 event entries, got %d", len(result))
	}
}

func TestReplayEngine_filterEntries_SkipsNonJSON(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	entries := []*logreader.LogEntry{
		{Raw: "DraftPick {}", IsJSON: true},
		{Raw: "DraftStatus {}", IsJSON: false}, // Non-JSON should be skipped
		{Raw: "DraftMakePick {}", IsJSON: true},
	}

	result := engine.filterEntries(entries, "draft")

	if len(result) != 2 {
		t.Errorf("Expected 2 entries (non-JSON skipped), got %d", len(result))
	}
}

func TestReplayEngine_extractTimestamp(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	tests := []struct {
		name      string
		timestamp string
		expectNil bool
	}{
		{
			name:      "Full timestamp with milliseconds",
			timestamp: "2024-11-16 15:30:45.123",
			expectNil: false,
		},
		{
			name:      "Full timestamp without milliseconds",
			timestamp: "2024-11-16 15:30:45",
			expectNil: false,
		},
		{
			name:      "Time only with milliseconds",
			timestamp: "15:30:45.123",
			expectNil: false,
		},
		{
			name:      "Time only without milliseconds",
			timestamp: "15:30:45",
			expectNil: false,
		},
		{
			name:      "Empty timestamp",
			timestamp: "",
			expectNil: true,
		},
		{
			name:      "Invalid timestamp",
			timestamp: "not-a-timestamp",
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &logreader.LogEntry{Timestamp: tt.timestamp}
			result := engine.extractTimestamp(entry)

			if tt.expectNil && !result.IsZero() {
				t.Error("Expected zero time")
			}
			if !tt.expectNil && result.IsZero() {
				t.Error("Expected non-zero time")
			}
		})
	}
}

func TestReplayEngine_calculateDelay(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	tests := []struct {
		name        string
		prevTime    string
		currTime    string
		expectedMin time.Duration
		expectedMax time.Duration
	}{
		{
			name:        "Normal delay",
			prevTime:    "15:30:45.000",
			currTime:    "15:30:46.000",
			expectedMin: 10 * time.Millisecond,
			expectedMax: 5 * time.Second,
		},
		{
			name:        "Small delay - uses minimum",
			prevTime:    "15:30:45.000",
			currTime:    "15:30:45.005",
			expectedMin: 10 * time.Millisecond,
			expectedMax: 10 * time.Millisecond,
		},
		{
			name:        "Large delay - capped at 5s",
			prevTime:    "15:30:45.000",
			currTime:    "15:40:45.000",
			expectedMin: 5 * time.Second,
			expectedMax: 5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prev := &logreader.LogEntry{Timestamp: tt.prevTime}
			curr := &logreader.LogEntry{Timestamp: tt.currTime}

			delay := engine.calculateDelay(prev, curr)

			if delay < tt.expectedMin {
				t.Errorf("Delay %v is less than minimum %v", delay, tt.expectedMin)
			}
			if delay > tt.expectedMax {
				t.Errorf("Delay %v is greater than maximum %v", delay, tt.expectedMax)
			}
		})
	}
}

func TestReplayEngine_calculateDelay_NilPrev(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	curr := &logreader.LogEntry{Timestamp: "15:30:45.000"}
	delay := engine.calculateDelay(nil, curr)

	if delay != 0 {
		t.Errorf("Expected 0 delay for nil prev, got %v", delay)
	}
}

func TestReplayEngine_calculateDelay_EmptyTimestamps(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	prev := &logreader.LogEntry{Timestamp: ""}
	curr := &logreader.LogEntry{Timestamp: "15:30:45.000"}

	delay := engine.calculateDelay(prev, curr)

	if delay != 0 {
		t.Errorf("Expected 0 delay for empty prev timestamp, got %v", delay)
	}
}

func TestReplayEngine_Start_EmptyLogFile(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	// Create a temporary empty log file
	tmpDir := t.TempDir()
	emptyFile := filepath.Join(tmpDir, "empty.log")
	if err := os.WriteFile(emptyFile, []byte{}, 0o644); err != nil {
		t.Fatalf("Failed to create empty file: %v", err)
	}

	err := engine.Start([]string{emptyFile}, 1.0, "all", false)
	if err == nil {
		t.Error("Expected error for empty log file")
	}

	if err.Error() != "log files contain no entries" {
		t.Errorf("Expected 'log files contain no entries' error, got: %v", err)
	}
}

func TestReplayEngine_Pause_AlreadyPaused(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	// Set state to active and paused
	engine.mu.Lock()
	engine.isActive = true
	engine.isPaused = true
	engine.mu.Unlock()

	err := engine.Pause()
	if err == nil {
		t.Error("Expected error when pausing already paused replay")
	}

	if err.Error() != "replay already paused" {
		t.Errorf("Expected 'replay already paused' error, got: %v", err)
	}
}

func TestReplayEngine_Resume_NotPaused(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	// Set state to active but not paused
	engine.mu.Lock()
	engine.isActive = true
	engine.isPaused = false
	engine.mu.Unlock()

	err := engine.Resume()
	if err == nil {
		t.Error("Expected error when resuming non-paused replay")
	}

	if err.Error() != "replay not paused" {
		t.Errorf("Expected 'replay not paused' error, got: %v", err)
	}
}

func TestReplayEngine_GetStatus_Active(t *testing.T) {
	service := &Service{
		wsServer: NewWebSocketServer(0),
	}

	engine := NewReplayEngine(service)

	// Set state to active with some entries
	engine.mu.Lock()
	engine.isActive = true
	engine.isPaused = false
	engine.entries = make([]*logreader.LogEntry, 100)
	engine.currentIdx = 25
	engine.speed = 2.0
	engine.filterType = "draft"
	engine.startTime = time.Now().Add(-10 * time.Second)
	engine.mu.Unlock()

	status := engine.GetStatus()

	if isActive, ok := status["isActive"].(bool); !ok || !isActive {
		t.Error("Expected isActive to be true")
	}

	if isPaused, ok := status["isPaused"].(bool); !ok || isPaused {
		t.Error("Expected isPaused to be false")
	}

	if currentEntry, ok := status["currentEntry"].(int); !ok || currentEntry != 25 {
		t.Errorf("Expected currentEntry 25, got %v", currentEntry)
	}

	if totalEntries, ok := status["totalEntries"].(int); !ok || totalEntries != 100 {
		t.Errorf("Expected totalEntries 100, got %v", totalEntries)
	}

	if speed, ok := status["speed"].(float64); !ok || speed != 2.0 {
		t.Errorf("Expected speed 2.0, got %v", speed)
	}

	if filter, ok := status["filter"].(string); !ok || filter != "draft" {
		t.Errorf("Expected filter 'draft', got %v", filter)
	}

	if percentComplete, ok := status["percentComplete"].(float64); !ok || percentComplete != 25.0 {
		t.Errorf("Expected percentComplete 25.0, got %v", percentComplete)
	}
}
