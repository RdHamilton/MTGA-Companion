package logreader

import (
	"testing"
	"time"
)

func TestGetPlayerSeatID_FromConnectResp(t *testing.T) {
	entries := []*LogEntry{
		{
			IsJSON: true,
			JSON: map[string]interface{}{
				"connectResp": map[string]interface{}{
					"systemSeatIds": []interface{}{float64(1)},
					"teamId":        float64(2),
				},
			},
		},
	}

	conn := GetPlayerSeatID(entries)

	if conn == nil {
		t.Fatal("Expected connection info to be non-nil")
	}
	if conn.SystemSeatID != 1 {
		t.Errorf("Expected SystemSeatID 1, got %d", conn.SystemSeatID)
	}
	if conn.TeamID != 2 {
		t.Errorf("Expected TeamID 2, got %d", conn.TeamID)
	}
}

func TestGetPlayerSeatID_FromMatchEvent(t *testing.T) {
	entries := []*LogEntry{
		{
			IsJSON: true,
			JSON: map[string]interface{}{
				"matchGameRoomStateChangedEvent": map[string]interface{}{
					"gameRoomInfo": map[string]interface{}{
						"gameRoomConfig": map[string]interface{}{
							"reservedPlayers": []interface{}{
								map[string]interface{}{
									"systemSeatId": float64(2),
									"teamId":       float64(1),
									"playerName":   "TestPlayer",
								},
							},
						},
					},
				},
			},
		},
	}

	conn := GetPlayerSeatID(entries)

	if conn == nil {
		t.Fatal("Expected connection info to be non-nil")
	}
	if conn.SystemSeatID != 2 {
		t.Errorf("Expected SystemSeatID 2, got %d", conn.SystemSeatID)
	}
	if conn.TeamID != 1 {
		t.Errorf("Expected TeamID 1, got %d", conn.TeamID)
	}
}

func TestGetPlayerSeatID_NoMatch(t *testing.T) {
	entries := []*LogEntry{
		{
			IsJSON: true,
			JSON: map[string]interface{}{
				"someOtherEvent": map[string]interface{}{},
			},
		},
	}

	conn := GetPlayerSeatID(entries)

	if conn != nil {
		t.Error("Expected connection info to be nil for non-matching entries")
	}
}

func TestParseGREMessages_GameStateMessage(t *testing.T) {
	entries := []*LogEntry{
		{
			IsJSON:    true,
			Timestamp: "2024-01-15 10:30:45",
			JSON: map[string]interface{}{
				"greToClientEvent": map[string]interface{}{
					"greToClientMessages": []interface{}{
						map[string]interface{}{
							"type": "GREMessageType_GameStateMessage",
							"gameStateMessage": map[string]interface{}{
								"turnInfo": map[string]interface{}{
									"turnNumber":     float64(3),
									"phase":          "Phase_Main1",
									"step":           "",
									"activePlayer":   float64(1),
									"priorityPlayer": float64(1),
								},
								"players": []interface{}{
									map[string]interface{}{
										"seatId":    float64(1),
										"lifeTotal": float64(20),
										"teamId":    float64(1),
									},
									map[string]interface{}{
										"seatId":    float64(2),
										"lifeTotal": float64(18),
										"teamId":    float64(2),
									},
								},
								"gameObjects": []interface{}{
									map[string]interface{}{
										"instanceId":       float64(100),
										"grpId":            float64(12345),
										"ownerSeatId":      float64(1),
										"controllerSeatId": float64(1),
										"zoneId":           float64(3), // battlefield
										"cardTypes":        []interface{}{"CardType_Creature"},
									},
								},
								"gameInfo": map[string]interface{}{
									"matchID":    "match-123",
									"gameNumber": float64(1),
								},
							},
						},
					},
				},
			},
		},
	}

	messages, err := ParseGREMessages(entries)
	if err != nil {
		t.Fatalf("ParseGREMessages failed: %v", err)
	}

	if len(messages) != 1 {
		t.Fatalf("Expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.MatchID != "match-123" {
		t.Errorf("Expected MatchID 'match-123', got '%s'", msg.MatchID)
	}
	if msg.GameNumber != 1 {
		t.Errorf("Expected GameNumber 1, got %d", msg.GameNumber)
	}

	if msg.TurnInfo == nil {
		t.Fatal("Expected TurnInfo to be non-nil")
	}
	if msg.TurnInfo.TurnNumber != 3 {
		t.Errorf("Expected TurnNumber 3, got %d", msg.TurnInfo.TurnNumber)
	}
	if msg.TurnInfo.Phase != "Phase_Main1" {
		t.Errorf("Expected Phase 'Phase_Main1', got '%s'", msg.TurnInfo.Phase)
	}

	if len(msg.Players) != 2 {
		t.Fatalf("Expected 2 players, got %d", len(msg.Players))
	}
	if msg.Players[0].LifeTotal != 20 {
		t.Errorf("Expected player 1 life 20, got %d", msg.Players[0].LifeTotal)
	}

	if len(msg.GameObjects) != 1 {
		t.Fatalf("Expected 1 game object, got %d", len(msg.GameObjects))
	}
	if msg.GameObjects[0].GRPId != 12345 {
		t.Errorf("Expected GRPId 12345, got %d", msg.GameObjects[0].GRPId)
	}
}

func TestParseGREMessages_SkipsNonGameStateMessages(t *testing.T) {
	entries := []*LogEntry{
		{
			IsJSON: true,
			JSON: map[string]interface{}{
				"greToClientEvent": map[string]interface{}{
					"greToClientMessages": []interface{}{
						map[string]interface{}{
							"type": "GREMessageType_QueuedGameStateMessage",
						},
						map[string]interface{}{
							"type": "GREMessageType_UIMessage",
						},
					},
				},
			},
		},
	}

	messages, err := ParseGREMessages(entries)
	if err != nil {
		t.Fatalf("ParseGREMessages failed: %v", err)
	}

	if len(messages) != 0 {
		t.Errorf("Expected 0 messages for non-GameStateMessage types, got %d", len(messages))
	}
}

func TestNormalizePhase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Phase_Beginning", "Beginning"},
		{"Phase_Main1", "Main1"},
		{"Phase_Combat", "Combat"},
		{"Phase_Main2", "Main2"},
		{"Phase_Ending", "Ending"},
		{"UnknownPhase", "UnknownPhase"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizePhase(tt.input)
			if result != tt.expected {
				t.Errorf("normalizePhase(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeStep(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Step_Upkeep", "Upkeep"},
		{"Step_Draw", "Draw"},
		{"Step_BeginCombat", "BeginCombat"},
		{"Step_DeclareAttack", "DeclareAttackers"},
		{"Step_DeclareBlock", "DeclareBlockers"},
		{"Step_CombatDamage", "CombatDamage"},
		{"Step_End", "EndStep"},
		{"Step_Cleanup", "Cleanup"},
		{"UnknownStep", "UnknownStep"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeStep(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeStep(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestZoneIDToName(t *testing.T) {
	tests := []struct {
		zoneID   int
		expected string
	}{
		{1, "hand"},
		{11, "hand"},
		{2, "library"},
		{3, "battlefield"},
		{4, "graveyard"},
		{5, "exile"},
		{6, "stack"},
		{7, "command"},
		{8, "zone_8"}, // Unknown zone
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := zoneIDToName(tt.zoneID)
			if result != tt.expected {
				t.Errorf("zoneIDToName(%d) = %q, expected %q", tt.zoneID, result, tt.expected)
			}
		})
	}
}

func TestParseGamePlays_DetectsZoneChanges(t *testing.T) {
	// Create two game states where a card moves from hand to battlefield
	entries := []*LogEntry{
		// First state: card in hand
		{
			IsJSON:    true,
			Timestamp: "2024-01-15 10:30:45",
			JSON: map[string]interface{}{
				"greToClientEvent": map[string]interface{}{
					"greToClientMessages": []interface{}{
						map[string]interface{}{
							"type": "GREMessageType_GameStateMessage",
							"gameStateMessage": map[string]interface{}{
								"turnInfo": map[string]interface{}{
									"turnNumber":   float64(1),
									"phase":        "Phase_Main1",
									"activePlayer": float64(1),
								},
								"gameObjects": []interface{}{
									map[string]interface{}{
										"instanceId":       float64(100),
										"grpId":            float64(12345),
										"controllerSeatId": float64(1),
										"zoneId":           float64(1), // hand
										"cardTypes":        []interface{}{"CardType_Creature"},
									},
								},
								"gameInfo": map[string]interface{}{
									"matchID":    "match-123",
									"gameNumber": float64(1),
								},
							},
						},
					},
				},
			},
		},
		// Second state: card on battlefield
		{
			IsJSON:    true,
			Timestamp: "2024-01-15 10:30:46",
			JSON: map[string]interface{}{
				"greToClientEvent": map[string]interface{}{
					"greToClientMessages": []interface{}{
						map[string]interface{}{
							"type": "GREMessageType_GameStateMessage",
							"gameStateMessage": map[string]interface{}{
								"turnInfo": map[string]interface{}{
									"turnNumber":   float64(1),
									"phase":        "Phase_Main1",
									"activePlayer": float64(1),
								},
								"gameObjects": []interface{}{
									map[string]interface{}{
										"instanceId":       float64(100),
										"grpId":            float64(12345),
										"controllerSeatId": float64(1),
										"zoneId":           float64(3), // battlefield
										"cardTypes":        []interface{}{"CardType_Creature"},
									},
								},
								"gameInfo": map[string]interface{}{
									"matchID":    "match-123",
									"gameNumber": float64(1),
								},
							},
						},
					},
				},
			},
		},
	}

	playerConn := &GREConnection{SeatID: 1, SystemSeatID: 1}
	plays, err := ParseGamePlays(entries, playerConn)
	if err != nil {
		t.Fatalf("ParseGamePlays failed: %v", err)
	}

	if len(plays) != 1 {
		t.Fatalf("Expected 1 play, got %d", len(plays))
	}

	play := plays[0]
	if play.ActionType != "play_card" {
		t.Errorf("Expected ActionType 'play_card', got '%s'", play.ActionType)
	}
	if play.PlayerType != "player" {
		t.Errorf("Expected PlayerType 'player', got '%s'", play.PlayerType)
	}
	if play.ZoneFrom != "hand" {
		t.Errorf("Expected ZoneFrom 'hand', got '%s'", play.ZoneFrom)
	}
	if play.ZoneTo != "battlefield" {
		t.Errorf("Expected ZoneTo 'battlefield', got '%s'", play.ZoneTo)
	}
	if play.CardID != 12345 {
		t.Errorf("Expected CardID 12345, got %d", play.CardID)
	}
}

func TestParseGamePlays_DetectsLandDrop(t *testing.T) {
	entries := []*LogEntry{
		// First state: land in hand
		{
			IsJSON:    true,
			Timestamp: "2024-01-15 10:30:45",
			JSON: map[string]interface{}{
				"greToClientEvent": map[string]interface{}{
					"greToClientMessages": []interface{}{
						map[string]interface{}{
							"type": "GREMessageType_GameStateMessage",
							"gameStateMessage": map[string]interface{}{
								"turnInfo": map[string]interface{}{
									"turnNumber":   float64(1),
									"phase":        "Phase_Main1",
									"activePlayer": float64(1),
								},
								"gameObjects": []interface{}{
									map[string]interface{}{
										"instanceId":       float64(200),
										"grpId":            float64(67890),
										"controllerSeatId": float64(1),
										"zoneId":           float64(1), // hand
										"cardTypes":        []interface{}{"CardType_Land"},
									},
								},
							},
						},
					},
				},
			},
		},
		// Second state: land on battlefield
		{
			IsJSON:    true,
			Timestamp: "2024-01-15 10:30:46",
			JSON: map[string]interface{}{
				"greToClientEvent": map[string]interface{}{
					"greToClientMessages": []interface{}{
						map[string]interface{}{
							"type": "GREMessageType_GameStateMessage",
							"gameStateMessage": map[string]interface{}{
								"turnInfo": map[string]interface{}{
									"turnNumber":   float64(1),
									"phase":        "Phase_Main1",
									"activePlayer": float64(1),
								},
								"gameObjects": []interface{}{
									map[string]interface{}{
										"instanceId":       float64(200),
										"grpId":            float64(67890),
										"controllerSeatId": float64(1),
										"zoneId":           float64(3), // battlefield
										"cardTypes":        []interface{}{"CardType_Land"},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	playerConn := &GREConnection{SeatID: 1, SystemSeatID: 1}
	plays, err := ParseGamePlays(entries, playerConn)
	if err != nil {
		t.Fatalf("ParseGamePlays failed: %v", err)
	}

	if len(plays) != 1 {
		t.Fatalf("Expected 1 play, got %d", len(plays))
	}

	play := plays[0]
	if play.ActionType != "land_drop" {
		t.Errorf("Expected ActionType 'land_drop', got '%s'", play.ActionType)
	}
}

func TestParseGamePlays_DetectsAttack(t *testing.T) {
	entries := []*LogEntry{
		// First state: creature not attacking
		{
			IsJSON:    true,
			Timestamp: "2024-01-15 10:30:45",
			JSON: map[string]interface{}{
				"greToClientEvent": map[string]interface{}{
					"greToClientMessages": []interface{}{
						map[string]interface{}{
							"type": "GREMessageType_GameStateMessage",
							"gameStateMessage": map[string]interface{}{
								"turnInfo": map[string]interface{}{
									"turnNumber":   float64(2),
									"phase":        "Phase_Combat",
									"step":         "Step_BeginCombat",
									"activePlayer": float64(1),
								},
								"gameObjects": []interface{}{
									map[string]interface{}{
										"instanceId":       float64(100),
										"grpId":            float64(12345),
										"controllerSeatId": float64(1),
										"zoneId":           float64(3),
									},
								},
							},
						},
					},
				},
			},
		},
		// Second state: creature attacking
		{
			IsJSON:    true,
			Timestamp: "2024-01-15 10:30:46",
			JSON: map[string]interface{}{
				"greToClientEvent": map[string]interface{}{
					"greToClientMessages": []interface{}{
						map[string]interface{}{
							"type": "GREMessageType_GameStateMessage",
							"gameStateMessage": map[string]interface{}{
								"turnInfo": map[string]interface{}{
									"turnNumber":   float64(2),
									"phase":        "Phase_Combat",
									"step":         "Step_DeclareAttack",
									"activePlayer": float64(1),
								},
								"gameObjects": []interface{}{
									map[string]interface{}{
										"instanceId":       float64(100),
										"grpId":            float64(12345),
										"controllerSeatId": float64(1),
										"zoneId":           float64(3),
										"attackState":      "AttackState_Attacking",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	playerConn := &GREConnection{SeatID: 1, SystemSeatID: 1}
	plays, err := ParseGamePlays(entries, playerConn)
	if err != nil {
		t.Fatalf("ParseGamePlays failed: %v", err)
	}

	// Should have at least one attack play
	hasAttack := false
	for _, play := range plays {
		if play.ActionType == "attack" {
			hasAttack = true
			if play.PlayerType != "player" {
				t.Errorf("Expected PlayerType 'player' for attack, got '%s'", play.PlayerType)
			}
			break
		}
	}

	if !hasAttack {
		t.Error("Expected to detect an attack action")
	}
}

func TestExtractOpponentCards(t *testing.T) {
	entries := []*LogEntry{
		{
			IsJSON: true,
			JSON: map[string]interface{}{
				"greToClientEvent": map[string]interface{}{
					"greToClientMessages": []interface{}{
						map[string]interface{}{
							"type": "GREMessageType_GameStateMessage",
							"gameStateMessage": map[string]interface{}{
								"turnInfo": map[string]interface{}{
									"turnNumber": float64(2),
								},
								"gameObjects": []interface{}{
									// Player's card
									map[string]interface{}{
										"instanceId":       float64(100),
										"grpId":            float64(11111),
										"controllerSeatId": float64(1),
										"zoneId":           float64(3),
									},
									// Opponent's card
									map[string]interface{}{
										"instanceId":       float64(200),
										"grpId":            float64(22222),
										"controllerSeatId": float64(2),
										"zoneId":           float64(3),
									},
									// Another opponent card
									map[string]interface{}{
										"instanceId":       float64(201),
										"grpId":            float64(33333),
										"controllerSeatId": float64(2),
										"zoneId":           float64(4), // graveyard
									},
								},
							},
						},
					},
				},
			},
		},
	}

	playerConn := &GREConnection{SeatID: 1, SystemSeatID: 1}
	cards, err := ExtractOpponentCards(entries, playerConn)
	if err != nil {
		t.Fatalf("ExtractOpponentCards failed: %v", err)
	}

	if len(cards) != 2 {
		t.Fatalf("Expected 2 opponent cards, got %d", len(cards))
	}

	// Check that both opponent cards are captured
	foundCard1 := false
	foundCard2 := false
	for _, card := range cards {
		if card.CardID == 22222 {
			foundCard1 = true
			if card.ZoneObserved != "battlefield" {
				t.Errorf("Expected zone 'battlefield' for card 22222, got '%s'", card.ZoneObserved)
			}
		}
		if card.CardID == 33333 {
			foundCard2 = true
			if card.ZoneObserved != "graveyard" {
				t.Errorf("Expected zone 'graveyard' for card 33333, got '%s'", card.ZoneObserved)
			}
		}
	}

	if !foundCard1 {
		t.Error("Expected to find opponent card 22222")
	}
	if !foundCard2 {
		t.Error("Expected to find opponent card 33333")
	}
}

func TestExtractGameSnapshots(t *testing.T) {
	entries := []*LogEntry{
		{
			IsJSON:    true,
			Timestamp: "2024-01-15 10:30:45",
			JSON: map[string]interface{}{
				"greToClientEvent": map[string]interface{}{
					"greToClientMessages": []interface{}{
						map[string]interface{}{
							"type": "GREMessageType_GameStateMessage",
							"gameStateMessage": map[string]interface{}{
								"turnInfo": map[string]interface{}{
									"turnNumber":   float64(1),
									"activePlayer": float64(1),
								},
								"players": []interface{}{
									map[string]interface{}{
										"seatId":    float64(1),
										"lifeTotal": float64(20),
									},
									map[string]interface{}{
										"seatId":    float64(2),
										"lifeTotal": float64(18),
									},
								},
								"gameObjects": []interface{}{
									// Player's hand
									map[string]interface{}{
										"instanceId":       float64(100),
										"grpId":            float64(11111),
										"controllerSeatId": float64(1),
										"zoneId":           float64(1), // hand
									},
									// Player's land
									map[string]interface{}{
										"instanceId":       float64(101),
										"grpId":            float64(22222),
										"controllerSeatId": float64(1),
										"zoneId":           float64(3), // battlefield
										"cardTypes":        []interface{}{"CardType_Land"},
									},
								},
								"gameInfo": map[string]interface{}{
									"matchID":    "match-123",
									"gameNumber": float64(1),
								},
							},
						},
					},
				},
			},
		},
	}

	playerConn := &GREConnection{SeatID: 1, SystemSeatID: 1}
	snapshots, err := ExtractGameSnapshots(entries, playerConn)
	if err != nil {
		t.Fatalf("ExtractGameSnapshots failed: %v", err)
	}

	if len(snapshots) != 1 {
		t.Fatalf("Expected 1 snapshot, got %d", len(snapshots))
	}

	snapshot := snapshots[0]
	if snapshot.MatchID != "match-123" {
		t.Errorf("Expected MatchID 'match-123', got '%s'", snapshot.MatchID)
	}
	if snapshot.TurnNumber != 1 {
		t.Errorf("Expected TurnNumber 1, got %d", snapshot.TurnNumber)
	}
	if snapshot.ActivePlayer != "player" {
		t.Errorf("Expected ActivePlayer 'player', got '%s'", snapshot.ActivePlayer)
	}
	if snapshot.PlayerLife != 20 {
		t.Errorf("Expected PlayerLife 20, got %d", snapshot.PlayerLife)
	}
	if snapshot.OpponentLife != 18 {
		t.Errorf("Expected OpponentLife 18, got %d", snapshot.OpponentLife)
	}
	if snapshot.PlayerCardsInHand != 1 {
		t.Errorf("Expected PlayerCardsInHand 1, got %d", snapshot.PlayerCardsInHand)
	}
	if snapshot.PlayerLandsInPlay != 1 {
		t.Errorf("Expected PlayerLandsInPlay 1, got %d", snapshot.PlayerLandsInPlay)
	}
}

func TestParseTurnInfo(t *testing.T) {
	turnInfo := map[string]interface{}{
		"turnNumber":     float64(5),
		"phase":          "Phase_Combat",
		"step":           "Step_DeclareAttack",
		"activePlayer":   float64(2),
		"priorityPlayer": float64(1),
		"decisionPlayer": float64(1),
		"nextPhase":      "Phase_Main2",
		"nextStep":       "",
	}

	ti := parseTurnInfo(turnInfo)

	if ti.TurnNumber != 5 {
		t.Errorf("Expected TurnNumber 5, got %d", ti.TurnNumber)
	}
	if ti.Phase != "Phase_Combat" {
		t.Errorf("Expected Phase 'Phase_Combat', got '%s'", ti.Phase)
	}
	if ti.Step != "Step_DeclareAttack" {
		t.Errorf("Expected Step 'Step_DeclareAttack', got '%s'", ti.Step)
	}
	if ti.ActivePlayer != 2 {
		t.Errorf("Expected ActivePlayer 2, got %d", ti.ActivePlayer)
	}
	if ti.PriorityPlayer != 1 {
		t.Errorf("Expected PriorityPlayer 1, got %d", ti.PriorityPlayer)
	}
}

func TestParsePlayerState(t *testing.T) {
	playerMap := map[string]interface{}{
		"seatId":        float64(1),
		"lifeTotal":     float64(17),
		"teamId":        float64(1),
		"maxHandSize":   float64(7),
		"systemSeatId":  float64(1),
		"timerState":    "Running",
		"timeRemaining": float64(300),
	}

	ps := parsePlayerState(playerMap)

	if ps.SeatID != 1 {
		t.Errorf("Expected SeatID 1, got %d", ps.SeatID)
	}
	if ps.LifeTotal != 17 {
		t.Errorf("Expected LifeTotal 17, got %d", ps.LifeTotal)
	}
	if ps.TeamID != 1 {
		t.Errorf("Expected TeamID 1, got %d", ps.TeamID)
	}
	if ps.MaxHandSize != 7 {
		t.Errorf("Expected MaxHandSize 7, got %d", ps.MaxHandSize)
	}
}

func TestParseGameObject(t *testing.T) {
	objMap := map[string]interface{}{
		"instanceId":       float64(42),
		"grpId":            float64(12345),
		"ownerSeatId":      float64(1),
		"controllerSeatId": float64(1),
		"zoneId":           float64(3),
		"cardTypes":        []interface{}{"CardType_Creature", "CardType_Legendary"},
		"power": map[string]interface{}{
			"value": float64(4),
		},
		"toughness": map[string]interface{}{
			"value": float64(5),
		},
		"isTapped":    true,
		"attackState": "AttackState_Attacking",
		"counters": []interface{}{
			map[string]interface{}{
				"type":  "+1/+1",
				"count": float64(2),
			},
		},
	}

	obj := parseGameObject(objMap)

	if obj.InstanceID != 42 {
		t.Errorf("Expected InstanceID 42, got %d", obj.InstanceID)
	}
	if obj.GRPId != 12345 {
		t.Errorf("Expected GRPId 12345, got %d", obj.GRPId)
	}
	if obj.ControllerSeatID != 1 {
		t.Errorf("Expected ControllerSeatID 1, got %d", obj.ControllerSeatID)
	}
	if obj.ZoneName != "battlefield" {
		t.Errorf("Expected ZoneName 'battlefield', got '%s'", obj.ZoneName)
	}
	if len(obj.CardTypes) != 2 {
		t.Errorf("Expected 2 card types, got %d", len(obj.CardTypes))
	}
	if obj.Power != 4 {
		t.Errorf("Expected Power 4, got %d", obj.Power)
	}
	if obj.Toughness != 5 {
		t.Errorf("Expected Toughness 5, got %d", obj.Toughness)
	}
	if !obj.IsTapped {
		t.Error("Expected IsTapped to be true")
	}
	if !obj.IsAttacking {
		t.Error("Expected IsAttacking to be true")
	}
	if obj.Counters["+1/+1"] != 2 {
		t.Errorf("Expected 2 +1/+1 counters, got %d", obj.Counters["+1/+1"])
	}
}

func TestParseGamePlays_EmptyEntries(t *testing.T) {
	entries := []*LogEntry{}

	playerConn := &GREConnection{SeatID: 1, SystemSeatID: 1}
	plays, err := ParseGamePlays(entries, playerConn)
	if err != nil {
		t.Fatalf("ParseGamePlays failed: %v", err)
	}

	if plays != nil && len(plays) != 0 {
		t.Errorf("Expected nil or empty plays for empty entries, got %d", len(plays))
	}
}

func TestParseGamePlays_SingleEntry(t *testing.T) {
	entries := []*LogEntry{
		{
			IsJSON: true,
			JSON: map[string]interface{}{
				"greToClientEvent": map[string]interface{}{
					"greToClientMessages": []interface{}{
						map[string]interface{}{
							"type": "GREMessageType_GameStateMessage",
							"gameStateMessage": map[string]interface{}{
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

	playerConn := &GREConnection{SeatID: 1, SystemSeatID: 1}
	plays, err := ParseGamePlays(entries, playerConn)
	if err != nil {
		t.Fatalf("ParseGamePlays failed: %v", err)
	}

	// With only one state, no zone changes can be detected
	if plays != nil && len(plays) != 0 {
		t.Errorf("Expected nil or empty plays for single entry, got %d", len(plays))
	}
}

// Benchmarks

func BenchmarkParseGREMessages(b *testing.B) {
	// Create a realistic set of entries
	entries := make([]*LogEntry, 100)
	for i := range entries {
		entries[i] = &LogEntry{
			IsJSON:    true,
			Timestamp: time.Now().Format("2006-01-02 15:04:05"),
			JSON: map[string]interface{}{
				"greToClientEvent": map[string]interface{}{
					"greToClientMessages": []interface{}{
						map[string]interface{}{
							"type": "GREMessageType_GameStateMessage",
							"gameStateMessage": map[string]interface{}{
								"turnInfo": map[string]interface{}{
									"turnNumber":   float64(i),
									"phase":        "Phase_Main1",
									"activePlayer": float64(1),
								},
								"players": []interface{}{
									map[string]interface{}{
										"seatId":    float64(1),
										"lifeTotal": float64(20),
									},
									map[string]interface{}{
										"seatId":    float64(2),
										"lifeTotal": float64(20),
									},
								},
								"gameObjects": []interface{}{
									map[string]interface{}{
										"instanceId":       float64(100 + i),
										"grpId":            float64(12345),
										"controllerSeatId": float64(1),
										"zoneId":           float64(3),
									},
								},
							},
						},
					},
				},
			},
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseGREMessages(entries)
	}
}
