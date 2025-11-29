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

func TestParseDraftSessionEvent_BotDraftDraftStatus(t *testing.T) {
	// This tests the BotDraftDraftStatus parsing which has CurrentModule: BotDraft
	// and a Payload field containing the draft state
	entry := &LogEntry{
		Raw:    `{"CurrentModule":"BotDraft","Payload":"{\"Result\":\"Success\",\"EventName\":\"QuickDraft_TLA_20251127\",\"DraftStatus\":\"Draft\",\"PackNumber\":0,\"PickNumber\":0,\"DraftPack\":[\"97380\",\"97468\"],\"PickedCards\":[]}"}`,
		IsJSON: true,
		JSON: map[string]interface{}{
			"CurrentModule": "BotDraft",
			"Payload":       `{"Result":"Success","EventName":"QuickDraft_TLA_20251127","DraftStatus":"Draft","PackNumber":0,"PickNumber":0,"DraftPack":["97380","97468"],"PickedCards":[]}`,
		},
	}

	event, err := ParseDraftSessionEvent(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event == nil {
		t.Fatal("expected event, got nil")
	}

	if event.Type != "status_updated" {
		t.Errorf("expected Type 'status_updated', got '%s'", event.Type)
	}

	if event.EventName != "QuickDraft_TLA_20251127" {
		t.Errorf("expected EventName 'QuickDraft_TLA_20251127', got '%s'", event.EventName)
	}

	if event.SetCode != "TLA" {
		t.Errorf("expected SetCode 'TLA', got '%s'", event.SetCode)
	}

	if event.PackNumber != 0 {
		t.Errorf("expected PackNumber 0, got %d", event.PackNumber)
	}

	if event.PickNumber != 0 {
		t.Errorf("expected PickNumber 0, got %d", event.PickNumber)
	}

	if len(event.DraftPack) != 2 {
		t.Errorf("expected 2 cards in DraftPack, got %d", len(event.DraftPack))
	}
}

func TestParseDraftSessionEvent_BotDraftDraftStatus_WithPickedCards(t *testing.T) {
	// Test mid-draft state with some cards already picked
	entry := &LogEntry{
		Raw:    `{"CurrentModule":"BotDraft","Payload":"{\"Result\":\"Success\",\"EventName\":\"QuickDraft_TLA_20251127\",\"DraftStatus\":\"Draft\",\"PackNumber\":1,\"PickNumber\":5,\"DraftPack\":[\"97400\",\"97401\",\"97402\"],\"PickedCards\":[\"97380\",\"97381\",\"97382\",\"97383\"]}"}`,
		IsJSON: true,
		JSON: map[string]interface{}{
			"CurrentModule": "BotDraft",
			"Payload":       `{"Result":"Success","EventName":"QuickDraft_TLA_20251127","DraftStatus":"Draft","PackNumber":1,"PickNumber":5,"DraftPack":["97400","97401","97402"],"PickedCards":["97380","97381","97382","97383"]}`,
		},
	}

	event, err := ParseDraftSessionEvent(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event == nil {
		t.Fatal("expected event, got nil")
	}

	if event.PackNumber != 1 {
		t.Errorf("expected PackNumber 1, got %d", event.PackNumber)
	}

	if event.PickNumber != 5 {
		t.Errorf("expected PickNumber 5, got %d", event.PickNumber)
	}

	if len(event.DraftPack) != 3 {
		t.Errorf("expected 3 cards in DraftPack, got %d", len(event.DraftPack))
	}

	if len(event.PickedCards) != 4 {
		t.Errorf("expected 4 cards in PickedCards, got %d", len(event.PickedCards))
	}
}

func TestParseDraftSessionEvent_NotBotDraftModule(t *testing.T) {
	// Test that non-BotDraft modules with Payload are ignored
	entry := &LogEntry{
		Raw:    `{"CurrentModule":"SomeOtherModule","Payload":"{}"}`,
		IsJSON: true,
		JSON: map[string]interface{}{
			"CurrentModule": "SomeOtherModule",
			"Payload":       "{}",
		},
	}

	event, err := ParseDraftSessionEvent(entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if event != nil {
		t.Errorf("expected nil event for non-BotDraft module, got %+v", event)
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
		{"QuickDraft_TLA_20251127", "TLA"},
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
