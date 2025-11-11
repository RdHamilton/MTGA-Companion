package draft

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Parser extracts draft events from MTGA log entries.
type Parser struct {
	currentState *DraftState
}

// NewParser creates a new draft event parser.
func NewParser() *Parser {
	return &Parser{}
}

// ParseLogEntry parses a single log line and extracts draft-related events.
// Returns nil if the line doesn't contain draft information.
func (p *Parser) ParseLogEntry(line string, timestamp time.Time) (*LogEvent, error) {
	// Trim whitespace
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, nil
	}

	// Try to detect event type and extract JSON payload
	eventType, jsonStr := p.detectEventType(line)
	if eventType == "" {
		return nil, nil // Not a draft event
	}

	return &LogEvent{
		Timestamp: timestamp,
		Type:      eventType,
		Data:      json.RawMessage(jsonStr),
	}, nil
}

// detectEventType detects the type of draft event from a log line.
// Returns the event type and the JSON payload string.
func (p *Parser) detectEventType(line string) (LogEventType, string) {
	// Check for BotDraft module (Quick Draft with escaped Payload)
	if strings.Contains(line, `"CurrentModule":"BotDraft"`) {
		return p.extractBotDraftPayload(line)
	}

	// CardsInPack event (Premier Draft P1P1)
	if strings.Contains(line, `"CardsInPack"`) {
		if jsonStr := extractJSON(line); jsonStr != "" {
			return LogEventCardsInPack, jsonStr
		}
	}

	// Draft.Notify event (Premier Draft)
	if strings.Contains(line, `"Draft.Notify"`) || strings.Contains(line, `"DraftNotify"`) {
		if jsonStr := extractJSON(line); jsonStr != "" {
			return LogEventDraftNotify, jsonStr
		}
	}

	// DraftPack event (Quick Draft - direct format, not BotDraft module)
	if strings.Contains(line, `"DraftPack"`) && strings.Contains(line, `"DraftStatus"`) {
		if jsonStr := extractJSON(line); jsonStr != "" {
			return LogEventDraftPack, jsonStr
		}
	}

	// Event_PlayerDraftMakePick
	if strings.Contains(line, `"Event_PlayerDraftMakePick"`) || strings.Contains(line, `"PlayerDraftMakePick"`) {
		if jsonStr := extractJSON(line); jsonStr != "" {
			return LogEventPlayerDraftPick, jsonStr
		}
	}

	// Draft.MakeHumanDraftPick
	if strings.Contains(line, `"Draft.MakeHumanDraftPick"`) || strings.Contains(line, `"MakeHumanDraftPick"`) {
		if jsonStr := extractJSON(line); jsonStr != "" {
			return LogEventHumanDraftPick, jsonStr
		}
	}

	// BotDraft_DraftPick (Quick Draft)
	if strings.Contains(line, `"BotDraft_DraftPick"`) {
		if jsonStr := extractJSON(line); jsonStr != "" {
			return LogEventBotDraftPick, jsonStr
		}
	}

	// EventGrantCardPool (Sealed)
	if strings.Contains(line, `"Event_GrantCardPool"`) || strings.Contains(line, `"EventGrantCardPool"`) {
		if jsonStr := extractJSON(line); jsonStr != "" {
			return LogEventGrantCardPool, jsonStr
		}
	}

	// Courses CardPool (Sealed alternate format)
	if strings.Contains(line, `"Courses"`) && strings.Contains(line, `"CardPool"`) {
		if jsonStr := extractJSON(line); jsonStr != "" {
			return LogEventCoursesCardPool, jsonStr
		}
	}

	return "", ""
}

// extractBotDraftPayload extracts and unescapes the Payload from BotDraft module logs.
// BotDraft logs have format: {"CurrentModule":"BotDraft","Payload":"{escaped JSON}"}
func (p *Parser) extractBotDraftPayload(line string) (LogEventType, string) {
	// Extract outer JSON
	outerJSON := extractJSON(line)
	if outerJSON == "" {
		return "", ""
	}

	// Parse outer JSON to get Payload field
	var envelope struct {
		CurrentModule string `json:"CurrentModule"`
		Payload       string `json:"Payload"`
	}

	if err := json.Unmarshal([]byte(outerJSON), &envelope); err != nil {
		return "", ""
	}

	// Payload contains escaped JSON - it's already a string, so we can use it directly
	payloadJSON := envelope.Payload

	// Determine event type based on content
	if strings.Contains(payloadJSON, `"DraftPack"`) && strings.Contains(payloadJSON, `"DraftStatus"`) {
		return LogEventDraftPack, payloadJSON
	}

	// Check for pick events (the response after making a pick)
	if strings.Contains(payloadJSON, `"PickInfo"`) {
		return LogEventBotDraftPick, payloadJSON
	}

	return "", ""
}

// extractJSON extracts JSON payload from a log line.
// Looks for content between curly braces.
func extractJSON(line string) string {
	// Find first opening brace
	start := strings.Index(line, "{")
	if start == -1 {
		return ""
	}

	// Find matching closing brace
	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(line); i++ {
		c := line[i]

		if escaped {
			escaped = false
			continue
		}

		if c == '\\' {
			escaped = true
			continue
		}

		if c == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		switch c {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return line[start : i+1]
			}
		}
	}

	return ""
}

// ParseCardsInPack parses a CardsInPack event.
func ParseCardsInPack(data json.RawMessage) (*Pack, error) {
	var payload CardsInPackPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse CardsInPack: %w", err)
	}

	return &Pack{
		PackNumber: 1, // CardsInPack is always P1P1
		PickNumber: 1,
		CardIDs:    []int(payload.CardsInPack), // Convert FlexibleIntArray to []int
		Timestamp:  time.Now(),
	}, nil
}

// ParseDraftNotify parses a Draft.Notify event.
func ParseDraftNotify(data json.RawMessage) (*Pack, *Pick, error) {
	var payload DraftNotifyPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, nil, fmt.Errorf("failed to parse DraftNotify: %w", err)
	}

	pack := &Pack{
		PackNumber: payload.PackNumber,
		PickNumber: payload.PickNumber,
		CardIDs:    []int(payload.DraftPack), // Convert FlexibleIntArray to []int
		Timestamp:  time.Now(),
	}

	var pick *Pick
	if payload.SelfPick > 0 {
		pick = &Pick{
			PackNumber: payload.PackNumber,
			PickNumber: payload.PickNumber - 1, // SelfPick is for previous pick
			CardID:     payload.SelfPick,
			Timestamp:  time.Now(),
		}
	}

	return pack, pick, nil
}

// ParseDraftPack parses a DraftPack event (Quick Draft).
func ParseDraftPack(data json.RawMessage) (*Pack, error) {
	var payload DraftPackPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse DraftPack: %w", err)
	}

	// Only process if it's the player's turn to pick
	if payload.DraftStatus != "PickNext" {
		return nil, nil
	}

	// BotDraft responses include actual pack/pick numbers
	return &Pack{
		PackNumber: payload.PackNumber,
		PickNumber: payload.PickNumber,
		CardIDs:    []int(payload.DraftPack), // Convert FlexibleIntArray to []int
		Timestamp:  time.Now(),
	}, nil
}

// ParseMakePick parses a pick event.
func ParseMakePick(data json.RawMessage) (*Pick, error) {
	var payload MakePickPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse MakePick: %w", err)
	}

	// Determine card ID (different events use different field names)
	cardID := payload.GrpId
	if cardID == 0 {
		cardID = payload.CardId
	}

	if cardID == 0 {
		return nil, fmt.Errorf("no card ID in pick event")
	}

	return &Pick{
		PackNumber: payload.Pack,
		PickNumber: payload.Pick,
		CardID:     cardID,
		Timestamp:  time.Now(),
	}, nil
}

// ParseGrantCardPool parses a sealed pool event.
func ParseGrantCardPool(data json.RawMessage) ([]int, error) {
	var payload GrantCardPoolPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse GrantCardPool: %w", err)
	}

	var cardIDs []int
	for _, card := range payload.CardsAdded {
		// Add card multiple times based on quantity
		for i := 0; i < card.Quantity; i++ {
			cardIDs = append(cardIDs, card.GrpId)
		}
	}

	return cardIDs, nil
}

// ParseCoursesCardPool parses a Courses sealed pool event.
func ParseCoursesCardPool(data json.RawMessage) ([]int, error) {
	var payload CoursesPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse Courses: %w", err)
	}

	if len(payload.Courses) == 0 {
		return nil, fmt.Errorf("no courses in payload")
	}

	// Return the first course's card pool
	return payload.Courses[0].CardPool, nil
}

// UpdateState updates the parser's internal state based on a log event.
func (p *Parser) UpdateState(event *LogEvent) error {
	if event == nil {
		return nil
	}

	// Initialize state if needed
	if p.currentState == nil {
		p.currentState = &DraftState{
			Event: DraftEvent{
				InProgress: true,
				StartTime:  event.Timestamp,
			},
			Picks:    []Pick{},
			AllPacks: []Pack{},
		}
	}

	switch event.Type {
	case LogEventCardsInPack:
		pack, err := ParseCardsInPack(event.Data)
		if err != nil {
			return err
		}
		p.currentState.CurrentPack = pack
		p.currentState.AllPacks = append(p.currentState.AllPacks, *pack)
		p.currentState.Event.CurrentPack = 1
		p.currentState.Event.CurrentPick = 1

	case LogEventDraftNotify:
		pack, pick, err := ParseDraftNotify(event.Data)
		if err != nil {
			return err
		}
		if pack != nil {
			p.currentState.CurrentPack = pack
			p.currentState.AllPacks = append(p.currentState.AllPacks, *pack)
			p.currentState.Event.CurrentPack = pack.PackNumber
			p.currentState.Event.CurrentPick = pack.PickNumber
		}
		if pick != nil {
			p.currentState.Picks = append(p.currentState.Picks, *pick)
		}

	case LogEventDraftPack:
		pack, err := ParseDraftPack(event.Data)
		if err != nil {
			return err
		}
		if pack != nil {
			// Use pack/pick numbers from MTGA response (they're accurate)
			p.currentState.Event.CurrentPack = pack.PackNumber
			p.currentState.Event.CurrentPick = pack.PickNumber

			p.currentState.CurrentPack = pack
			p.currentState.AllPacks = append(p.currentState.AllPacks, *pack)
		}

	case LogEventPlayerDraftPick, LogEventHumanDraftPick, LogEventBotDraftPick:
		pick, err := ParseMakePick(event.Data)
		if err != nil {
			return err
		}
		if pick != nil {
			p.currentState.Picks = append(p.currentState.Picks, *pick)
		}

	case LogEventGrantCardPool, LogEventCoursesCardPool:
		// For sealed, we treat the entire pool as one "pack"
		var cardIDs []int
		var err error

		if event.Type == LogEventGrantCardPool {
			cardIDs, err = ParseGrantCardPool(event.Data)
		} else {
			cardIDs, err = ParseCoursesCardPool(event.Data)
		}

		if err != nil {
			return err
		}

		pack := &Pack{
			PackNumber: 1,
			PickNumber: 1,
			CardIDs:    cardIDs,
			Timestamp:  event.Timestamp,
		}

		p.currentState.CurrentPack = pack
		p.currentState.AllPacks = append(p.currentState.AllPacks, *pack)
		p.currentState.Event.Type = DraftTypeSealed
	}

	return nil
}

// GetCurrentState returns the current draft state.
func (p *Parser) GetCurrentState() *DraftState {
	return p.currentState
}

// Reset resets the parser state (when a draft ends).
func (p *Parser) Reset() {
	if p.currentState != nil {
		p.currentState.Event.InProgress = false
		endTime := time.Now()
		p.currentState.Event.EndTime = &endTime
	}
	p.currentState = nil
}
