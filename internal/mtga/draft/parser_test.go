package draft

import (
	"encoding/json"
	"testing"
	"time"
)

func TestExtractJSON(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		wantJSON bool
	}{
		{
			name:     "simple object",
			line:     `[2024-01-01 12:00:00] {"CardsInPack":[1,2,3]}`,
			wantJSON: true,
		},
		{
			name:     "nested object",
			line:     `[2024-01-01 12:00:00] {"DraftPack":[1,2],"DraftStatus":"PickNext"}`,
			wantJSON: true,
		},
		{
			name:     "object with escaped quotes",
			line:     `[2024-01-01 12:00:00] {"Name":"Card \"Name\"","Id":123}`,
			wantJSON: true,
		},
		{
			name:     "no JSON",
			line:     `[2024-01-01 12:00:00] Some log message without JSON`,
			wantJSON: false,
		},
		{
			name:     "deeply nested",
			line:     `[2024-01-01 12:00:00] {"Outer":{"Inner":{"Deep":[1,2,3]}}}`,
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSON(tt.line)
			if tt.wantJSON && result == "" {
				t.Errorf("extractJSON() returned empty, expected JSON")
			}
			if !tt.wantJSON && result != "" {
				t.Errorf("extractJSON() returned JSON, expected empty")
			}
			// If we got JSON, verify it's valid
			if result != "" {
				var obj map[string]interface{}
				if err := json.Unmarshal([]byte(result), &obj); err != nil {
					t.Errorf("extractJSON() returned invalid JSON: %v", err)
				}
			}
		})
	}
}

func TestDetectEventType(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantType  LogEventType
		wantEmpty bool
	}{
		{
			name:     "CardsInPack event",
			line:     `[2024-01-01 12:00:00] {"CardsInPack":[89019,89020,89021]}`,
			wantType: LogEventCardsInPack,
		},
		{
			name:     "Draft.Notify event",
			line:     `[2024-01-01 12:00:00] {"Draft.Notify":{"DraftPack":[1,2,3],"PackNumber":1,"PickNumber":2}}`,
			wantType: LogEventDraftNotify,
		},
		{
			name:     "DraftNotify alternate",
			line:     `[2024-01-01 12:00:00] {"DraftNotify":{"DraftPack":[1,2,3],"PackNumber":1,"PickNumber":2}}`,
			wantType: LogEventDraftNotify,
		},
		{
			name:     "DraftPack event",
			line:     `[2024-01-01 12:00:00] {"DraftPack":[1,2,3],"DraftStatus":"PickNext"}`,
			wantType: LogEventDraftPack,
		},
		{
			name:     "PlayerDraftMakePick event",
			line:     `[2024-01-01 12:00:00] {"Event_PlayerDraftMakePick":{"GrpId":89019,"Pack":1,"Pick":1}}`,
			wantType: LogEventPlayerDraftPick,
		},
		{
			name:     "MakeHumanDraftPick event",
			line:     `[2024-01-01 12:00:00] {"Draft.MakeHumanDraftPick":{"GrpId":89019,"Pack":1,"Pick":1}}`,
			wantType: LogEventHumanDraftPick,
		},
		{
			name:     "BotDraft_DraftPick event",
			line:     `[2024-01-01 12:00:00] {"BotDraft_DraftPick":{"GrpId":89019,"Pack":1,"Pick":1}}`,
			wantType: LogEventBotDraftPick,
		},
		{
			name:     "GrantCardPool event",
			line:     `[2024-01-01 12:00:00] {"Event_GrantCardPool":{"CardsAdded":[{"GrpId":123,"Quantity":1}]}}`,
			wantType: LogEventGrantCardPool,
		},
		{
			name:     "Courses CardPool event",
			line:     `[2024-01-01 12:00:00] {"Courses":[{"CardPool":[1,2,3]}]}`,
			wantType: LogEventCoursesCardPool,
		},
		{
			name:      "non-draft event",
			line:      `[2024-01-01 12:00:00] {"SomeOtherEvent":"data"}`,
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			eventType, jsonStr := p.detectEventType(tt.line)

			if tt.wantEmpty {
				if eventType != "" {
					t.Errorf("detectEventType() returned %s, want empty", eventType)
				}
				return
			}

			if eventType != tt.wantType {
				t.Errorf("detectEventType() type = %s, want %s", eventType, tt.wantType)
			}
			if jsonStr == "" {
				t.Error("detectEventType() returned empty JSON")
			}
		})
	}
}

func TestParseCardsInPack(t *testing.T) {
	payload := CardsInPackPayload{
		CardsInPack: []int{89019, 89020, 89021, 89022, 89023},
	}
	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	pack, err := ParseCardsInPack(json.RawMessage(data))
	if err != nil {
		t.Fatalf("ParseCardsInPack() error = %v", err)
	}

	// CardsInPack is always P1P1
	if pack.PackNumber != 1 {
		t.Errorf("PackNumber = %d, want 1", pack.PackNumber)
	}
	if pack.PickNumber != 1 {
		t.Errorf("PickNumber = %d, want 1", pack.PickNumber)
	}
	if len(pack.CardIDs) != 5 {
		t.Errorf("CardIDs length = %d, want 5", len(pack.CardIDs))
	}
}

func TestParseDraftNotify(t *testing.T) {
	tests := []struct {
		name         string
		payload      DraftNotifyPayload
		wantPackNum  int
		wantPickNum  int
		wantCardsCnt int
		wantPick     bool
		wantPickCard int
	}{
		{
			name: "first pick with no SelfPick",
			payload: DraftNotifyPayload{
				DraftPack:  []int{1, 2, 3, 4, 5},
				PackNumber: 1,
				PickNumber: 1,
				SelfPick:   0,
			},
			wantPackNum:  1,
			wantPickNum:  1,
			wantCardsCnt: 5,
			wantPick:     false,
		},
		{
			name: "second pick with SelfPick",
			payload: DraftNotifyPayload{
				DraftPack:  []int{2, 3, 4, 5},
				PackNumber: 1,
				PickNumber: 2,
				SelfPick:   1,
			},
			wantPackNum:  1,
			wantPickNum:  2,
			wantCardsCnt: 4,
			wantPick:     true,
			wantPickCard: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal test data: %v", err)
			}

			pack, pick, err := ParseDraftNotify(json.RawMessage(data))
			if err != nil {
				t.Fatalf("ParseDraftNotify() error = %v", err)
			}

			if pack.PackNumber != tt.wantPackNum {
				t.Errorf("PackNumber = %d, want %d", pack.PackNumber, tt.wantPackNum)
			}
			if pack.PickNumber != tt.wantPickNum {
				t.Errorf("PickNumber = %d, want %d", pack.PickNumber, tt.wantPickNum)
			}
			if len(pack.CardIDs) != tt.wantCardsCnt {
				t.Errorf("CardIDs length = %d, want %d", len(pack.CardIDs), tt.wantCardsCnt)
			}

			if tt.wantPick {
				if pick == nil {
					t.Error("Expected pick, got nil")
				} else if pick.CardID != tt.wantPickCard {
					t.Errorf("Pick CardID = %d, want %d", pick.CardID, tt.wantPickCard)
				}
			} else {
				if pick != nil {
					t.Error("Expected no pick, got pick")
				}
			}
		})
	}
}

func TestParseDraftPack(t *testing.T) {
	tests := []struct {
		name         string
		payload      DraftPackPayload
		wantPack     bool
		wantCardsCnt int
	}{
		{
			name: "PickNext status",
			payload: DraftPackPayload{
				DraftPack:   []int{1, 2, 3, 4, 5},
				DraftStatus: "PickNext",
			},
			wantPack:     true,
			wantCardsCnt: 5,
		},
		{
			name: "other status",
			payload: DraftPackPayload{
				DraftPack:   []int{1, 2, 3, 4, 5},
				DraftStatus: "Waiting",
			},
			wantPack: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal test data: %v", err)
			}

			pack, err := ParseDraftPack(json.RawMessage(data))
			if err != nil {
				t.Fatalf("ParseDraftPack() error = %v", err)
			}

			if tt.wantPack {
				if pack == nil {
					t.Error("Expected pack, got nil")
				} else if len(pack.CardIDs) != tt.wantCardsCnt {
					t.Errorf("CardIDs length = %d, want %d", len(pack.CardIDs), tt.wantCardsCnt)
				}
			} else {
				if pack != nil {
					t.Error("Expected no pack, got pack")
				}
			}
		})
	}
}

func TestParseMakePick(t *testing.T) {
	tests := []struct {
		name       string
		payload    MakePickPayload
		wantCardID int
		wantPack   int
		wantPick   int
		wantError  bool
	}{
		{
			name: "pick with GrpId",
			payload: MakePickPayload{
				GrpId: 89019,
				Pack:  1,
				Pick:  1,
			},
			wantCardID: 89019,
			wantPack:   1,
			wantPick:   1,
		},
		{
			name: "pick with CardId",
			payload: MakePickPayload{
				CardId: 89020,
				Pack:   2,
				Pick:   3,
			},
			wantCardID: 89020,
			wantPack:   2,
			wantPick:   3,
		},
		{
			name:      "pick with no card ID",
			payload:   MakePickPayload{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal test data: %v", err)
			}

			pick, err := ParseMakePick(json.RawMessage(data))

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseMakePick() error = %v", err)
			}

			if pick.CardID != tt.wantCardID {
				t.Errorf("CardID = %d, want %d", pick.CardID, tt.wantCardID)
			}
			if pick.PackNumber != tt.wantPack {
				t.Errorf("PackNumber = %d, want %d", pick.PackNumber, tt.wantPack)
			}
			if pick.PickNumber != tt.wantPick {
				t.Errorf("PickNumber = %d, want %d", pick.PickNumber, tt.wantPick)
			}
		})
	}
}

func TestParseGrantCardPool(t *testing.T) {
	payload := GrantCardPoolPayload{
		CardsAdded: []struct {
			GrpId    int `json:"GrpId"`
			Quantity int `json:"Quantity"`
		}{
			{GrpId: 100, Quantity: 1},
			{GrpId: 200, Quantity: 2},
			{GrpId: 300, Quantity: 1},
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	cardIDs, err := ParseGrantCardPool(json.RawMessage(data))
	if err != nil {
		t.Fatalf("ParseGrantCardPool() error = %v", err)
	}

	// Should have 4 cards total (1 + 2 + 1)
	if len(cardIDs) != 4 {
		t.Errorf("CardIDs length = %d, want 4", len(cardIDs))
	}

	// Check that card 200 appears twice
	count200 := 0
	for _, id := range cardIDs {
		if id == 200 {
			count200++
		}
	}
	if count200 != 2 {
		t.Errorf("Card 200 appears %d times, want 2", count200)
	}
}

func TestParseCoursesCardPool(t *testing.T) {
	tests := []struct {
		name      string
		payload   CoursesPayload
		wantCount int
		wantError bool
	}{
		{
			name: "valid courses",
			payload: CoursesPayload{
				Courses: []struct {
					CardPool []int `json:"CardPool"`
				}{
					{CardPool: []int{1, 2, 3, 4, 5}},
				},
			},
			wantCount: 5,
		},
		{
			name: "empty courses",
			payload: CoursesPayload{
				Courses: []struct {
					CardPool []int `json:"CardPool"`
				}{},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal test data: %v", err)
			}

			cardIDs, err := ParseCoursesCardPool(json.RawMessage(data))

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseCoursesCardPool() error = %v", err)
			}

			if len(cardIDs) != tt.wantCount {
				t.Errorf("CardIDs length = %d, want %d", len(cardIDs), tt.wantCount)
			}
		})
	}
}

func TestParserUpdateState(t *testing.T) {
	p := NewParser()

	// Test CardsInPack event
	cardsInPackData, _ := json.Marshal(CardsInPackPayload{
		CardsInPack: []int{1, 2, 3, 4, 5},
	})
	event := &LogEvent{
		Timestamp: time.Now(),
		Type:      LogEventCardsInPack,
		Data:      json.RawMessage(cardsInPackData),
	}

	err := p.UpdateState(event)
	if err != nil {
		t.Fatalf("UpdateState() error = %v", err)
	}

	state := p.GetCurrentState()
	if state == nil {
		t.Fatal("GetCurrentState() returned nil")
	}

	if !state.Event.InProgress {
		t.Error("Draft should be in progress")
	}

	if state.Event.CurrentPack != 1 {
		t.Errorf("CurrentPack = %d, want 1", state.Event.CurrentPack)
	}

	if state.Event.CurrentPick != 1 {
		t.Errorf("CurrentPick = %d, want 1", state.Event.CurrentPick)
	}

	if len(state.AllPacks) != 1 {
		t.Errorf("AllPacks length = %d, want 1", len(state.AllPacks))
	}
}

func TestParserUpdateStateWithPicks(t *testing.T) {
	p := NewParser()

	// Initialize with CardsInPack
	cardsInPackData, _ := json.Marshal(CardsInPackPayload{
		CardsInPack: []int{1, 2, 3, 4, 5},
	})
	p.UpdateState(&LogEvent{
		Timestamp: time.Now(),
		Type:      LogEventCardsInPack,
		Data:      json.RawMessage(cardsInPackData),
	})

	// Add a pick
	pickData, _ := json.Marshal(MakePickPayload{
		GrpId: 1,
		Pack:  1,
		Pick:  1,
	})
	p.UpdateState(&LogEvent{
		Timestamp: time.Now(),
		Type:      LogEventPlayerDraftPick,
		Data:      json.RawMessage(pickData),
	})

	state := p.GetCurrentState()
	if len(state.Picks) != 1 {
		t.Errorf("Picks length = %d, want 1", len(state.Picks))
	}

	if state.Picks[0].CardID != 1 {
		t.Errorf("Pick CardID = %d, want 1", state.Picks[0].CardID)
	}
}

func TestParserUpdateStateQuickDraft(t *testing.T) {
	p := NewParser()

	// Quick Draft uses DraftPack events
	for i := 1; i <= 3; i++ {
		draftPackData, _ := json.Marshal(DraftPackPayload{
			DraftPack:   []int{1, 2, 3, 4, 5},
			DraftStatus: "PickNext",
		})

		err := p.UpdateState(&LogEvent{
			Timestamp: time.Now(),
			Type:      LogEventDraftPack,
			Data:      json.RawMessage(draftPackData),
		})
		if err != nil {
			t.Fatalf("UpdateState() error = %v", err)
		}
	}

	state := p.GetCurrentState()
	if state.Event.CurrentPick != 3 {
		t.Errorf("CurrentPick = %d, want 3", state.Event.CurrentPick)
	}

	if len(state.AllPacks) != 3 {
		t.Errorf("AllPacks length = %d, want 3", len(state.AllPacks))
	}
}

func TestParserUpdateStateSealed(t *testing.T) {
	p := NewParser()

	// Sealed uses GrantCardPool
	grantData, _ := json.Marshal(GrantCardPoolPayload{
		CardsAdded: []struct {
			GrpId    int `json:"GrpId"`
			Quantity int `json:"Quantity"`
		}{
			{GrpId: 1, Quantity: 1},
			{GrpId: 2, Quantity: 2},
			{GrpId: 3, Quantity: 1},
		},
	})

	err := p.UpdateState(&LogEvent{
		Timestamp: time.Now(),
		Type:      LogEventGrantCardPool,
		Data:      json.RawMessage(grantData),
	})
	if err != nil {
		t.Fatalf("UpdateState() error = %v", err)
	}

	state := p.GetCurrentState()
	if state.Event.Type != DraftTypeSealed {
		t.Errorf("Event.Type = %s, want %s", state.Event.Type, DraftTypeSealed)
	}

	if len(state.AllPacks) != 1 {
		t.Errorf("AllPacks length = %d, want 1 (sealed pool)", len(state.AllPacks))
	}

	// Sealed pool should have 4 cards (1 + 2 + 1)
	if len(state.AllPacks[0].CardIDs) != 4 {
		t.Errorf("Sealed pool CardIDs = %d, want 4", len(state.AllPacks[0].CardIDs))
	}
}

func TestParserReset(t *testing.T) {
	p := NewParser()

	// Initialize state
	cardsInPackData, _ := json.Marshal(CardsInPackPayload{
		CardsInPack: []int{1, 2, 3, 4, 5},
	})
	p.UpdateState(&LogEvent{
		Timestamp: time.Now(),
		Type:      LogEventCardsInPack,
		Data:      json.RawMessage(cardsInPackData),
	})

	// Verify state exists
	if p.GetCurrentState() == nil {
		t.Fatal("State should exist before reset")
	}

	// Reset
	p.Reset()

	// Verify state is cleared
	if p.GetCurrentState() != nil {
		t.Error("State should be nil after reset")
	}
}

func TestParseLogEntry(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantEvent bool
		wantType  LogEventType
	}{
		{
			name:      "CardsInPack event",
			line:      `[2024-01-01 12:00:00] {"CardsInPack":[1,2,3]}`,
			wantEvent: true,
			wantType:  LogEventCardsInPack,
		},
		{
			name:      "non-draft event",
			line:      `[2024-01-01 12:00:00] {"SomeOther":"event"}`,
			wantEvent: false,
		},
		{
			name:      "empty line",
			line:      "",
			wantEvent: false,
		},
		{
			name:      "whitespace only",
			line:      "   \t  ",
			wantEvent: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewParser()
			event, err := p.ParseLogEntry(tt.line, time.Now())
			if err != nil {
				t.Fatalf("ParseLogEntry() error = %v", err)
			}

			if tt.wantEvent {
				if event == nil {
					t.Error("Expected event, got nil")
				} else if event.Type != tt.wantType {
					t.Errorf("Event type = %s, want %s", event.Type, tt.wantType)
				}
			} else {
				if event != nil {
					t.Error("Expected no event, got event")
				}
			}
		})
	}
}

func TestIsDraftEvent(t *testing.T) {
	draftEvents := []LogEventType{
		LogEventDraftStart,
		LogEventDraftEnd,
		LogEventCardsInPack,
		LogEventDraftNotify,
		LogEventDraftPack,
		LogEventPlayerDraftPick,
		LogEventHumanDraftPick,
		LogEventBotDraftPick,
		LogEventGrantCardPool,
		LogEventCoursesCardPool,
	}

	for _, eventType := range draftEvents {
		if !IsDraftEvent(eventType) {
			t.Errorf("IsDraftEvent(%s) = false, want true", eventType)
		}
	}

	// Test non-draft event
	nonDraftEvent := LogEventType("non_draft_event")
	if IsDraftEvent(nonDraftEvent) {
		t.Errorf("IsDraftEvent(%s) = true, want false", nonDraftEvent)
	}
}

func TestIsPackEvent(t *testing.T) {
	packEvents := []LogEventType{
		LogEventNewPack,
		LogEventCardsInPack,
		LogEventDraftNotify,
		LogEventDraftPack,
		LogEventGrantCardPool,
		LogEventCoursesCardPool,
	}

	for _, eventType := range packEvents {
		if !IsPackEvent(eventType) {
			t.Errorf("IsPackEvent(%s) = false, want true", eventType)
		}
	}

	// Test non-pack event
	if IsPackEvent(LogEventPlayerDraftPick) {
		t.Error("IsPackEvent(LogEventPlayerDraftPick) = true, want false")
	}
}

func TestIsPickEvent(t *testing.T) {
	pickEvents := []LogEventType{
		LogEventMakePick,
		LogEventPlayerDraftPick,
		LogEventHumanDraftPick,
		LogEventBotDraftPick,
		LogEventDraftMakePickResp,
	}

	for _, eventType := range pickEvents {
		if !IsPickEvent(eventType) {
			t.Errorf("IsPickEvent(%s) = false, want true", eventType)
		}
	}

	// Test non-pick event
	if IsPickEvent(LogEventCardsInPack) {
		t.Error("IsPickEvent(LogEventCardsInPack) = true, want false")
	}
}
