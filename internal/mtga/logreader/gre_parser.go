package logreader

import (
	"fmt"
	"time"
)

// GREGameStateMessage represents a parsed game state message from the GRE.
type GREGameStateMessage struct {
	MatchID       string
	GameNumber    int
	TurnInfo      *GRETurnInfo
	Players       []GREPlayerState
	GameObjects   []GREGameObject
	PrevGameState *GREGameStateMessage // For comparing state changes
	Timestamp     time.Time
}

// GRETurnInfo contains information about the current turn.
type GRETurnInfo struct {
	TurnNumber          int
	Phase               string // "Phase_Main1", "Phase_Combat", etc.
	Step                string // "Step_BeginCombat", "Step_DeclareAttackers", etc.
	ActivePlayer        int    // Seat ID of the active player
	PriorityPlayer      int    // Seat ID of player with priority
	DecisionPlayer      int    // Seat ID of player making a decision
	NextPhase           string
	NextStep            string
	StormCount          int
	ManaSpent           int
	PhasePaymentOptions []interface{}
}

// GREPlayerState represents a player's state in the game.
type GREPlayerState struct {
	SeatID          int
	LifeTotal       int
	TeamID          int
	MaxHandSize     int
	PendingMessages int
	TimerState      string
	TimeRemaining   int
	SystemSeatID    int // For identifying player vs opponent
}

// GREGameObject represents an object in the game (card, token, etc.).
type GREGameObject struct {
	InstanceID           int
	GRPId                int    // Arena card ID
	OwnerSeatID          int    // Who owns this object
	ControllerSeatID     int    // Who controls this object
	ZoneID               int    // Current zone
	ZoneName             string // Derived from ZoneID
	CardTypes            []string
	Subtypes             []string
	SuperTypes           []string
	Power                int
	Toughness            int
	IsTapped             bool
	IsAttacking          bool
	IsBlocking           bool
	HasSummoningSickness bool
	Counters             map[string]int
	Abilities            []int
}

// GamePlayEvent represents a detected game play/action.
type GamePlayEvent struct {
	MatchID        string
	GameNumber     int
	TurnNumber     int
	Phase          string
	Step           string
	PlayerType     string // "player" or "opponent"
	ActionType     string // "play_card", "attack", "block", "land_drop", "mulligan"
	CardID         int    // Arena card ID (GRPId)
	CardName       string // Will be populated later from card database
	ZoneFrom       string
	ZoneTo         string
	Timestamp      time.Time
	SequenceNumber int
}

// GREConnection stores the player's connection info for seat identification.
type GREConnection struct {
	SeatID       int
	SystemSeatID int
	TeamID       int
}

// GetPlayerSeatID extracts the player's seat ID from connectResp messages.
// This is used to identify which objects belong to the player vs opponent.
func GetPlayerSeatID(entries []*LogEntry) *GREConnection {
	for _, entry := range entries {
		if !entry.IsJSON {
			continue
		}

		// Look for connectResp
		if connectResp, ok := entry.JSON["connectResp"]; ok {
			connMap, ok := connectResp.(map[string]interface{})
			if !ok {
				continue
			}

			conn := &GREConnection{}

			// Get system seat IDs
			if seatIDs, ok := connMap["systemSeatIds"].([]interface{}); ok && len(seatIDs) > 0 {
				if seatID, ok := seatIDs[0].(float64); ok {
					conn.SystemSeatID = int(seatID)
					conn.SeatID = int(seatID) // Usually the same
				}
			}

			// Try to get teamId
			if teamID, ok := connMap["teamId"].(float64); ok {
				conn.TeamID = int(teamID)
			}

			if conn.SeatID != 0 || conn.SystemSeatID != 0 {
				return conn
			}
		}

		// Also look for matchCreated/matchGameRoomStateChangedEvent for seat info
		if matchEvent, ok := entry.JSON["matchGameRoomStateChangedEvent"]; ok {
			eventMap, ok := matchEvent.(map[string]interface{})
			if !ok {
				continue
			}

			gameRoomInfo, ok := eventMap["gameRoomInfo"].(map[string]interface{})
			if !ok {
				continue
			}

			gameRoomConfig, ok := gameRoomInfo["gameRoomConfig"].(map[string]interface{})
			if !ok {
				continue
			}

			reservedPlayers, ok := gameRoomConfig["reservedPlayers"].([]interface{})
			if !ok {
				continue
			}

			// The first player in the list is usually the logged-in player
			// But we need to match by screen name to be sure
			if len(reservedPlayers) > 0 {
				player, ok := reservedPlayers[0].(map[string]interface{})
				if !ok {
					continue
				}

				conn := &GREConnection{}
				if seatID, ok := player["systemSeatId"].(float64); ok {
					conn.SystemSeatID = int(seatID)
					conn.SeatID = int(seatID)
				}
				if teamID, ok := player["teamId"].(float64); ok {
					conn.TeamID = int(teamID)
				}

				if conn.SeatID != 0 || conn.SystemSeatID != 0 {
					return conn
				}
			}
		}
	}

	return nil
}

// ParseGREMessages extracts game state messages from log entries.
func ParseGREMessages(entries []*LogEntry) ([]*GREGameStateMessage, error) {
	var messages []*GREGameStateMessage

	for _, entry := range entries {
		if !entry.IsJSON {
			continue
		}

		// Parse entry timestamp
		entryTime := time.Now()
		if entry.Timestamp != "" {
			if t, err := parseLogTimestamp(entry.Timestamp); err == nil {
				entryTime = t
			}
		}

		// Look for greToClientEvent messages
		if greEvent, ok := entry.JSON["greToClientEvent"]; ok {
			eventMap, ok := greEvent.(map[string]interface{})
			if !ok {
				continue
			}

			greToClientMsgs, ok := eventMap["greToClientMessages"].([]interface{})
			if !ok {
				continue
			}

			for _, msgData := range greToClientMsgs {
				msgMap, ok := msgData.(map[string]interface{})
				if !ok {
					continue
				}

				// Check message type
				msgType, _ := msgMap["type"].(string)
				if msgType != "GREMessageType_GameStateMessage" {
					continue
				}

				msg := parseGameStateMessage(msgMap, entryTime)
				if msg != nil {
					messages = append(messages, msg)
				}
			}
		}
	}

	return messages, nil
}

// parseGameStateMessage parses a single game state message.
func parseGameStateMessage(msgMap map[string]interface{}, timestamp time.Time) *GREGameStateMessage {
	msg := &GREGameStateMessage{
		Timestamp: timestamp,
	}

	// Get game state info
	gameStateMsg, ok := msgMap["gameStateMessage"].(map[string]interface{})
	if !ok {
		return nil
	}

	// Parse turn info
	if turnInfo, ok := gameStateMsg["turnInfo"].(map[string]interface{}); ok {
		msg.TurnInfo = parseTurnInfo(turnInfo)
	}

	// Parse players
	if players, ok := gameStateMsg["players"].([]interface{}); ok {
		for _, playerData := range players {
			playerMap, ok := playerData.(map[string]interface{})
			if !ok {
				continue
			}
			player := parsePlayerState(playerMap)
			msg.Players = append(msg.Players, player)
		}
	}

	// Parse game objects
	if gameObjects, ok := gameStateMsg["gameObjects"].([]interface{}); ok {
		for _, objData := range gameObjects {
			objMap, ok := objData.(map[string]interface{})
			if !ok {
				continue
			}
			obj := parseGameObject(objMap)
			msg.GameObjects = append(msg.GameObjects, obj)
		}
	}

	// Get game info if available
	if gameInfo, ok := gameStateMsg["gameInfo"].(map[string]interface{}); ok {
		if matchID, ok := gameInfo["matchID"].(string); ok {
			msg.MatchID = matchID
		}
		if gameNumber, ok := gameInfo["gameNumber"].(float64); ok {
			msg.GameNumber = int(gameNumber)
		}
	}

	return msg
}

// parseTurnInfo parses turn information from the game state.
func parseTurnInfo(turnInfo map[string]interface{}) *GRETurnInfo {
	ti := &GRETurnInfo{}

	if turnNumber, ok := turnInfo["turnNumber"].(float64); ok {
		ti.TurnNumber = int(turnNumber)
	}
	if phase, ok := turnInfo["phase"].(string); ok {
		ti.Phase = phase
	}
	if step, ok := turnInfo["step"].(string); ok {
		ti.Step = step
	}
	if activePlayer, ok := turnInfo["activePlayer"].(float64); ok {
		ti.ActivePlayer = int(activePlayer)
	}
	if priorityPlayer, ok := turnInfo["priorityPlayer"].(float64); ok {
		ti.PriorityPlayer = int(priorityPlayer)
	}
	if decisionPlayer, ok := turnInfo["decisionPlayer"].(float64); ok {
		ti.DecisionPlayer = int(decisionPlayer)
	}
	if nextPhase, ok := turnInfo["nextPhase"].(string); ok {
		ti.NextPhase = nextPhase
	}
	if nextStep, ok := turnInfo["nextStep"].(string); ok {
		ti.NextStep = nextStep
	}

	return ti
}

// parsePlayerState parses a player's state from the game state.
func parsePlayerState(playerMap map[string]interface{}) GREPlayerState {
	ps := GREPlayerState{}

	if seatID, ok := playerMap["seatId"].(float64); ok {
		ps.SeatID = int(seatID)
	}
	if lifeTotal, ok := playerMap["lifeTotal"].(float64); ok {
		ps.LifeTotal = int(lifeTotal)
	}
	if teamID, ok := playerMap["teamId"].(float64); ok {
		ps.TeamID = int(teamID)
	}
	if maxHandSize, ok := playerMap["maxHandSize"].(float64); ok {
		ps.MaxHandSize = int(maxHandSize)
	}
	if systemSeatID, ok := playerMap["systemSeatId"].(float64); ok {
		ps.SystemSeatID = int(systemSeatID)
	}
	if timerState, ok := playerMap["timerState"].(string); ok {
		ps.TimerState = timerState
	}
	if timeRemaining, ok := playerMap["timeRemaining"].(float64); ok {
		ps.TimeRemaining = int(timeRemaining)
	}

	return ps
}

// parseGameObject parses a game object from the game state.
func parseGameObject(objMap map[string]interface{}) GREGameObject {
	obj := GREGameObject{
		Counters: make(map[string]int),
	}

	if instanceID, ok := objMap["instanceId"].(float64); ok {
		obj.InstanceID = int(instanceID)
	}
	if grpID, ok := objMap["grpId"].(float64); ok {
		obj.GRPId = int(grpID)
	}
	if ownerSeatID, ok := objMap["ownerSeatId"].(float64); ok {
		obj.OwnerSeatID = int(ownerSeatID)
	}
	if controllerSeatID, ok := objMap["controllerSeatId"].(float64); ok {
		obj.ControllerSeatID = int(controllerSeatID)
	}
	if zoneID, ok := objMap["zoneId"].(float64); ok {
		obj.ZoneID = int(zoneID)
		obj.ZoneName = zoneIDToName(int(zoneID))
	}

	// Parse card types
	if cardTypes, ok := objMap["cardTypes"].([]interface{}); ok {
		for _, ct := range cardTypes {
			if ctStr, ok := ct.(string); ok {
				obj.CardTypes = append(obj.CardTypes, ctStr)
			}
		}
	}

	// Parse other attributes
	if power, ok := objMap["power"].(map[string]interface{}); ok {
		if val, ok := power["value"].(float64); ok {
			obj.Power = int(val)
		}
	}
	if toughness, ok := objMap["toughness"].(map[string]interface{}); ok {
		if val, ok := toughness["value"].(float64); ok {
			obj.Toughness = int(val)
		}
	}
	if isTapped, ok := objMap["isTapped"].(bool); ok {
		obj.IsTapped = isTapped
	}
	if attacking, ok := objMap["attackState"].(string); ok {
		obj.IsAttacking = attacking == "AttackState_Attacking"
	}
	if blocking, ok := objMap["blockState"].(string); ok {
		obj.IsBlocking = blocking != "" && blocking != "BlockState_None"
	}

	// Parse counters
	if counters, ok := objMap["counters"].([]interface{}); ok {
		for _, counterData := range counters {
			counterMap, ok := counterData.(map[string]interface{})
			if !ok {
				continue
			}
			if counterType, ok := counterMap["type"].(string); ok {
				if count, ok := counterMap["count"].(float64); ok {
					obj.Counters[counterType] = int(count)
				}
			}
		}
	}

	return obj
}

// zoneIDToName maps zone IDs to readable zone names.
// Zone IDs in MTGA are player-specific (different IDs for each player's zones).
func zoneIDToName(zoneID int) string {
	// These mappings are based on typical MTGA zone IDs
	// Zone IDs 1-10 are typically player 1's zones, 11-20 are player 2's
	// This is a rough mapping - actual zone detection requires tracking
	// the zone structure from earlier messages

	// Common pattern:
	// zoneID % 10 gives the zone type
	zoneType := zoneID % 10
	switch zoneType {
	case 1:
		return "hand"
	case 2:
		return "library"
	case 3:
		return "battlefield"
	case 4:
		return "graveyard"
	case 5:
		return "exile"
	case 6:
		return "stack"
	case 7:
		return "command"
	default:
		return fmt.Sprintf("zone_%d", zoneID)
	}
}

// ParseGamePlays extracts game plays by comparing consecutive game states.
func ParseGamePlays(entries []*LogEntry, playerConn *GREConnection) ([]*GamePlayEvent, error) {
	messages, err := ParseGREMessages(entries)
	if err != nil {
		return nil, err
	}

	if len(messages) < 2 {
		return nil, nil
	}

	var plays []*GamePlayEvent
	sequenceNum := 0

	// Compare consecutive states to detect changes
	for i := 1; i < len(messages); i++ {
		prev := messages[i-1]
		curr := messages[i]

		// Detect zone changes
		zoneChanges := detectZoneChanges(prev, curr, playerConn)
		for _, change := range zoneChanges {
			change.SequenceNumber = sequenceNum
			sequenceNum++
			plays = append(plays, change)
		}

		// Detect attacks
		attacks := detectAttacks(prev, curr, playerConn)
		for _, attack := range attacks {
			attack.SequenceNumber = sequenceNum
			sequenceNum++
			plays = append(plays, attack)
		}

		// Detect blocks
		blocks := detectBlocks(prev, curr, playerConn)
		for _, block := range blocks {
			block.SequenceNumber = sequenceNum
			sequenceNum++
			plays = append(plays, block)
		}
	}

	return plays, nil
}

// detectZoneChanges finds objects that moved between zones.
func detectZoneChanges(prev, curr *GREGameStateMessage, playerConn *GREConnection) []*GamePlayEvent {
	var events []*GamePlayEvent

	if curr.TurnInfo == nil {
		return events
	}

	// Build maps of objects by instance ID
	prevObjects := make(map[int]GREGameObject)
	for _, obj := range prev.GameObjects {
		prevObjects[obj.InstanceID] = obj
	}

	currObjects := make(map[int]GREGameObject)
	for _, obj := range curr.GameObjects {
		currObjects[obj.InstanceID] = obj
	}

	// Check for objects that changed zones
	for instanceID, currObj := range currObjects {
		prevObj, existed := prevObjects[instanceID]

		// New object or zone change
		if !existed || prevObj.ZoneID != currObj.ZoneID {
			// Determine player type
			playerType := "opponent"
			if playerConn != nil && currObj.ControllerSeatID == playerConn.SeatID {
				playerType = "player"
			}

			// Determine action type based on zone transition
			actionType := "play_card"
			fromZone := "unknown"
			if existed {
				fromZone = prevObj.ZoneName
			}
			toZone := currObj.ZoneName

			// Hand to battlefield = play or land drop
			if fromZone == "hand" && toZone == "battlefield" {
				// Check if it's a land
				for _, cardType := range currObj.CardTypes {
					if cardType == "CardType_Land" {
						actionType = "land_drop"
						break
					}
				}
			} else if fromZone == "library" && toZone == "hand" {
				// Draw - skip, not a "play"
				continue
			} else if toZone == "graveyard" {
				// Something died or was discarded - track for opponent cards
				actionType = "play_card" // Could be more specific
			} else if fromZone == "unknown" && toZone == "hand" {
				// Mulligan or game start
				continue
			}

			event := &GamePlayEvent{
				MatchID:    curr.MatchID,
				GameNumber: curr.GameNumber,
				TurnNumber: curr.TurnInfo.TurnNumber,
				Phase:      normalizePhase(curr.TurnInfo.Phase),
				Step:       normalizeStep(curr.TurnInfo.Step),
				PlayerType: playerType,
				ActionType: actionType,
				CardID:     currObj.GRPId,
				ZoneFrom:   fromZone,
				ZoneTo:     toZone,
				Timestamp:  curr.Timestamp,
			}
			events = append(events, event)
		}
	}

	return events
}

// detectAttacks finds creatures that started attacking.
func detectAttacks(prev, curr *GREGameStateMessage, playerConn *GREConnection) []*GamePlayEvent {
	var events []*GamePlayEvent

	if curr.TurnInfo == nil {
		return events
	}

	// Build map of previous attacking state
	prevAttacking := make(map[int]bool)
	for _, obj := range prev.GameObjects {
		prevAttacking[obj.InstanceID] = obj.IsAttacking
	}

	// Check for new attackers
	for _, obj := range curr.GameObjects {
		wasAttacking := prevAttacking[obj.InstanceID]
		if obj.IsAttacking && !wasAttacking {
			playerType := "opponent"
			if playerConn != nil && obj.ControllerSeatID == playerConn.SeatID {
				playerType = "player"
			}

			event := &GamePlayEvent{
				MatchID:    curr.MatchID,
				GameNumber: curr.GameNumber,
				TurnNumber: curr.TurnInfo.TurnNumber,
				Phase:      normalizePhase(curr.TurnInfo.Phase),
				Step:       normalizeStep(curr.TurnInfo.Step),
				PlayerType: playerType,
				ActionType: "attack",
				CardID:     obj.GRPId,
				ZoneFrom:   "battlefield",
				ZoneTo:     "battlefield",
				Timestamp:  curr.Timestamp,
			}
			events = append(events, event)
		}
	}

	return events
}

// detectBlocks finds creatures that started blocking.
func detectBlocks(prev, curr *GREGameStateMessage, playerConn *GREConnection) []*GamePlayEvent {
	var events []*GamePlayEvent

	if curr.TurnInfo == nil {
		return events
	}

	// Build map of previous blocking state
	prevBlocking := make(map[int]bool)
	for _, obj := range prev.GameObjects {
		prevBlocking[obj.InstanceID] = obj.IsBlocking
	}

	// Check for new blockers
	for _, obj := range curr.GameObjects {
		wasBlocking := prevBlocking[obj.InstanceID]
		if obj.IsBlocking && !wasBlocking {
			playerType := "opponent"
			if playerConn != nil && obj.ControllerSeatID == playerConn.SeatID {
				playerType = "player"
			}

			event := &GamePlayEvent{
				MatchID:    curr.MatchID,
				GameNumber: curr.GameNumber,
				TurnNumber: curr.TurnInfo.TurnNumber,
				Phase:      normalizePhase(curr.TurnInfo.Phase),
				Step:       normalizeStep(curr.TurnInfo.Step),
				PlayerType: playerType,
				ActionType: "block",
				CardID:     obj.GRPId,
				ZoneFrom:   "battlefield",
				ZoneTo:     "battlefield",
				Timestamp:  curr.Timestamp,
			}
			events = append(events, event)
		}
	}

	return events
}

// normalizePhase converts MTGA phase names to readable names.
func normalizePhase(phase string) string {
	switch phase {
	case "Phase_Beginning":
		return "Beginning"
	case "Phase_Main1":
		return "Main1"
	case "Phase_Combat":
		return "Combat"
	case "Phase_Main2":
		return "Main2"
	case "Phase_Ending":
		return "Ending"
	default:
		return phase
	}
}

// normalizeStep converts MTGA step names to readable names.
func normalizeStep(step string) string {
	switch step {
	case "Step_Upkeep":
		return "Upkeep"
	case "Step_Draw":
		return "Draw"
	case "Step_BeginCombat":
		return "BeginCombat"
	case "Step_DeclareAttack":
		return "DeclareAttackers"
	case "Step_DeclareBlock":
		return "DeclareBlockers"
	case "Step_CombatDamage":
		return "CombatDamage"
	case "Step_FirstStrikeDamage":
		return "FirstStrikeDamage"
	case "Step_EndCombat":
		return "EndCombat"
	case "Step_End":
		return "EndStep"
	case "Step_Cleanup":
		return "Cleanup"
	default:
		return step
	}
}

// ExtractOpponentCards extracts all cards observed from the opponent.
func ExtractOpponentCards(entries []*LogEntry, playerConn *GREConnection) ([]OpponentCard, error) {
	messages, err := ParseGREMessages(entries)
	if err != nil {
		return nil, err
	}

	// Track seen cards with the turn they were first seen
	seenCards := make(map[int]*OpponentCard)

	for _, msg := range messages {
		turnNumber := 0
		if msg.TurnInfo != nil {
			turnNumber = msg.TurnInfo.TurnNumber
		}

		for _, obj := range msg.GameObjects {
			// Skip player's own cards
			if playerConn != nil && obj.ControllerSeatID == playerConn.SeatID {
				continue
			}

			// Skip cards without a valid GRPId
			if obj.GRPId == 0 {
				continue
			}

			// Track the card
			if existing, exists := seenCards[obj.GRPId]; exists {
				existing.TimesSeen++
				// Update zone if moved to a more revealing zone
				if obj.ZoneName == "battlefield" || obj.ZoneName == "hand" || obj.ZoneName == "graveyard" {
					existing.ZoneObserved = obj.ZoneName
				}
			} else {
				seenCards[obj.GRPId] = &OpponentCard{
					CardID:        obj.GRPId,
					ZoneObserved:  obj.ZoneName,
					TurnFirstSeen: turnNumber,
					TimesSeen:     1,
				}
			}
		}
	}

	// Convert map to slice
	var cards []OpponentCard
	for _, card := range seenCards {
		cards = append(cards, *card)
	}

	return cards, nil
}

// OpponentCard represents an opponent's card that was observed.
type OpponentCard struct {
	CardID        int
	CardName      string // Will be populated later from card database
	ZoneObserved  string
	TurnFirstSeen int
	TimesSeen     int
}

// ExtractGameSnapshots extracts turn-by-turn snapshots from game state messages.
func ExtractGameSnapshots(entries []*LogEntry, playerConn *GREConnection) ([]*GameSnapshot, error) {
	messages, err := ParseGREMessages(entries)
	if err != nil {
		return nil, err
	}

	// Group by turn number and take the last state for each turn
	turnSnapshots := make(map[int]*GREGameStateMessage)
	for _, msg := range messages {
		if msg.TurnInfo == nil {
			continue
		}
		turnSnapshots[msg.TurnInfo.TurnNumber] = msg
	}

	var snapshots []*GameSnapshot
	for turnNumber, msg := range turnSnapshots {
		snapshot := &GameSnapshot{
			MatchID:    msg.MatchID,
			GameNumber: msg.GameNumber,
			TurnNumber: turnNumber,
			Timestamp:  msg.Timestamp,
		}

		if msg.TurnInfo != nil {
			activePlayer := "opponent"
			if playerConn != nil && msg.TurnInfo.ActivePlayer == playerConn.SeatID {
				activePlayer = "player"
			}
			snapshot.ActivePlayer = activePlayer
		}

		// Extract player and opponent states
		for _, player := range msg.Players {
			if playerConn != nil && player.SeatID == playerConn.SeatID {
				snapshot.PlayerLife = player.LifeTotal
			} else {
				snapshot.OpponentLife = player.LifeTotal
			}
		}

		// Count cards and lands by controller
		playerCardsInHand := 0
		opponentCardsInHand := 0
		playerLands := 0
		opponentLands := 0

		for _, obj := range msg.GameObjects {
			isPlayer := playerConn != nil && obj.ControllerSeatID == playerConn.SeatID

			if obj.ZoneName == "hand" {
				if isPlayer {
					playerCardsInHand++
				} else {
					opponentCardsInHand++
				}
			}

			if obj.ZoneName == "battlefield" {
				for _, cardType := range obj.CardTypes {
					if cardType == "CardType_Land" {
						if isPlayer {
							playerLands++
						} else {
							opponentLands++
						}
						break
					}
				}
			}
		}

		snapshot.PlayerCardsInHand = playerCardsInHand
		snapshot.OpponentCardsInHand = opponentCardsInHand
		snapshot.PlayerLandsInPlay = playerLands
		snapshot.OpponentLandsInPlay = opponentLands

		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

// GameSnapshot represents the game state at a specific turn.
type GameSnapshot struct {
	MatchID             string
	GameNumber          int
	TurnNumber          int
	ActivePlayer        string
	PlayerLife          int
	OpponentLife        int
	PlayerCardsInHand   int
	OpponentCardsInHand int
	PlayerLandsInPlay   int
	OpponentLandsInPlay int
	BoardStateJSON      string
	Timestamp           time.Time
}

// parseLogTimestamp is defined in draft_picks.go
