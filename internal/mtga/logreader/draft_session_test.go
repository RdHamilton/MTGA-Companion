package logreader

import (
	"testing"
)

func TestParseDraftSessionEvent_DraftStart(t *testing.T) {
	entry := &LogEntry{
		Raw:    `[UnityCrossThreadLogger]Client.SceneChange {"fromSceneName":"EventLanding","toSceneName":"Draft","initiator":"System","context":"BotDraft"}`,
		IsJSON: true,
		JSON: map[string]interface{}{
			"fromSceneName": "EventLanding",
			"toSceneName":   "Draft",
			"initiator":     "System",
			"context":       "BotDraft",
		},
	}

	event, err := ParseDraftSessionEvent(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event == nil {
		t.Fatal("expected event, got nil")
	}

	if event.Type != "started" {
		t.Errorf("expected Type 'started', got '%s'", event.Type)
	}

	if event.Context != "BotDraft" {
		t.Errorf("expected Context 'BotDraft', got '%s'", event.Context)
	}
}

func TestParseDraftSessionEvent_DraftEnd(t *testing.T) {
	entry := &LogEntry{
		Raw:    `[UnityCrossThreadLogger]Client.SceneChange {"fromSceneName":"Draft","toSceneName":"DeckBuilder","initiator":"System","context":"deck builder"}`,
		IsJSON: true,
		JSON: map[string]interface{}{
			"fromSceneName": "Draft",
			"toSceneName":   "DeckBuilder",
			"initiator":     "System",
			"context":       "deck builder",
		},
	}

	event, err := ParseDraftSessionEvent(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event == nil {
		t.Fatal("expected event, got nil")
	}

	if event.Type != "ended" {
		t.Errorf("expected Type 'ended', got '%s'", event.Type)
	}
}

func TestParseDraftSessionEvent_NotDraftRelated(t *testing.T) {
	entry := &LogEntry{
		Raw:    `[UnityCrossThreadLogger]Some random log line`,
		IsJSON: false,
		JSON:   nil,
	}

	event, err := ParseDraftSessionEvent(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event != nil {
		t.Errorf("expected nil event for non-draft log, got %+v", event)
	}
}

func TestExtractSetCode(t *testing.T) {
	tests := []struct {
		eventName    string
		expectedCode string
	}{
		{"QuickDraft_TDM_20251111", "TDM"},
		{"QuickDraft_BLB_20240801", "BLB"},
		{"PremierDraft_OTJ_20240515", "OTJ"},
		{"QuickDraft_MKM_20240201", "MKM"},
		{"invalid_format", ""},
	}

	for _, tt := range tests {
		t.Run(tt.eventName, func(t *testing.T) {
			code := extractSetCode(tt.eventName)
			if code != tt.expectedCode {
				t.Errorf("expected '%s', got '%s'", tt.expectedCode, code)
			}
		})
	}
}
