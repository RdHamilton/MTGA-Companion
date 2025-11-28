package events

import (
	"context"
	"testing"
)

func TestNewTypedEvent(t *testing.T) {
	ctx := context.Background()

	event := NewTypedEvent("stats:updated", StatsUpdatedEvent{
		Matches: 5,
		Games:   10,
	}, ctx)

	if event.Type != "stats:updated" {
		t.Errorf("Expected type 'stats:updated', got '%s'", event.Type)
	}

	if event.TypedData == nil {
		t.Error("Expected TypedData to be set")
	}

	typed, ok := event.TypedData.(StatsUpdatedEvent)
	if !ok {
		t.Error("Expected TypedData to be StatsUpdatedEvent")
	}

	if typed.Matches != 5 || typed.Games != 10 {
		t.Errorf("Expected Matches=5, Games=10, got Matches=%d, Games=%d", typed.Matches, typed.Games)
	}
}

func TestGetTypedData(t *testing.T) {
	ctx := context.Background()

	event := NewTypedEvent("rank:updated", RankUpdatedEvent{
		Format: "Constructed",
		Tier:   "Gold",
		Step:   "2",
	}, ctx)

	// Test successful extraction
	data, ok := GetTypedData[RankUpdatedEvent](event)
	if !ok {
		t.Error("Expected GetTypedData to succeed")
	}

	if data.Format != "Constructed" {
		t.Errorf("Expected Format 'Constructed', got '%s'", data.Format)
	}

	// Test wrong type extraction
	_, ok = GetTypedData[StatsUpdatedEvent](event)
	if ok {
		t.Error("Expected GetTypedData to fail for wrong type")
	}
}

func TestGetTypedData_NilTypedData(t *testing.T) {
	event := Event{
		Type:      "test",
		Data:      map[string]interface{}{"key": "value"},
		TypedData: nil,
	}

	_, ok := GetTypedData[StatsUpdatedEvent](event)
	if ok {
		t.Error("Expected GetTypedData to fail for nil TypedData")
	}
}

func TestStructToMap_NilInput(t *testing.T) {
	result := structToMap(nil)

	if result == nil {
		t.Error("Expected non-nil map")
	}

	if len(result) != 0 {
		t.Errorf("Expected empty map, got %d elements", len(result))
	}
}

func TestStructToMap_MapInput(t *testing.T) {
	input := map[string]interface{}{
		"key": "value",
	}

	result := structToMap(input)

	if result["key"] != "value" {
		t.Errorf("Expected key='value', got '%v'", result["key"])
	}
}

func TestEventMessageTypes(t *testing.T) {
	// Verify all event types can be instantiated
	tests := []struct {
		name  string
		event any
	}{
		{"StatsUpdatedEvent", StatsUpdatedEvent{Matches: 1, Games: 2}},
		{"RankUpdatedEvent", RankUpdatedEvent{Format: "Constructed", Tier: "Gold", Step: "1"}},
		{"QuestUpdatedEvent", QuestUpdatedEvent{Completed: 1, Count: 3}},
		{"DraftUpdatedEvent", DraftUpdatedEvent{Count: 1, Picks: 15}},
		{"DeckUpdatedEvent", DeckUpdatedEvent{Count: 5}},
		{"CollectionUpdatedEvent", CollectionUpdatedEvent{NewCards: 10, CardsAdded: 40}},
		{"DaemonStatusEvent", DaemonStatusEvent{Status: "connected", Connected: true}},
		{"DaemonConnectedEvent", DaemonConnectedEvent{Version: "1.0.0"}},
		{"DaemonErrorEvent", DaemonErrorEvent{Error: "test error", Code: "ERR_001"}},
		{"ReplayStartedEvent", ReplayStartedEvent{TotalFiles: 10}},
		{"ReplayProgressEvent", ReplayProgressEvent{Current: 5, Total: 10, Percentage: 50.0}},
		{"ReplayPausedEvent", ReplayPausedEvent{Current: 5, Total: 10}},
		{"ReplayResumedEvent", ReplayResumedEvent{Current: 5, Total: 10}},
		{"ReplayCompletedEvent", ReplayCompletedEvent{FilesProcessed: 10, Duration: 5.5}},
		{"ReplayErrorEvent", ReplayErrorEvent{Error: "replay error"}},
		{"ReplayDraftDetectedEvent", ReplayDraftDetectedEvent{DraftID: "123", SetCode: "DSK"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			event := NewTypedEvent("test", tt.event, ctx)

			if event.TypedData == nil {
				t.Error("Expected TypedData to be set")
			}
		})
	}
}

func TestOutgoingMessageTypes(t *testing.T) {
	// Verify all outgoing message types can be instantiated
	tests := []struct {
		name    string
		message any
	}{
		{"ReplayLogsMessage", ReplayLogsMessage{Type: "replay_logs", ClearData: true}},
		{"StartReplayMessage", StartReplayMessage{Type: "start_replay", Files: []string{"file1.log"}}},
		{"PauseReplayMessage", PauseReplayMessage{Type: "pause_replay"}},
		{"ResumeReplayMessage", ResumeReplayMessage{Type: "resume_replay"}},
		{"StopReplayMessage", StopReplayMessage{Type: "stop_replay"}},
		{"PingMessage", PingMessage{Type: "ping"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.message == nil {
				t.Error("Expected message to be set")
			}
		})
	}
}
