package logreader

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// DraftSessionEvent represents a parsed draft session event from MTGA logs.
type DraftSessionEvent struct {
	Type         string   // "started", "status_updated", "pick_made", "ended"
	SessionID    string   // Unique session identifier
	EventName    string   // e.g., "QuickDraft_TDM_20251111"
	SetCode      string   // e.g., "TDM"
	Context      string   // e.g., "BotDraft"
	PackNumber   int      // 0-indexed (0 = Pack 1)
	PickNumber   int      // 1-indexed
	DraftPack    []string // Card IDs available in current pack
	PickedCards  []string // Card IDs already picked
	SelectedCard []string // Card IDs selected in this pick
	Timestamp    time.Time
}

// ParseDraftSessionEvent parses a log entry for draft session events.
// Returns nil if the entry is not a draft-related event.
func ParseDraftSessionEvent(entry *LogEntry) (*DraftSessionEvent, error) {
	if !entry.IsJSON {
		return nil, nil
	}

	// Check for draft start: Client.SceneChange with toSceneName="Draft"
	if sceneChange, ok := entry.JSON["toSceneName"]; ok && sceneChange == "Draft" {
		context, _ := entry.JSON["context"].(string)
		// BotDraft = Quick Draft, HumanDraft = Premier/Traditional Draft
		if context != "BotDraft" && context != "HumanDraft" {
			return nil, nil // Not a draft we're tracking
		}

		return &DraftSessionEvent{
			Type:      "started",
			Context:   context,
			Timestamp: time.Now(), // Use current time as log may not have timestamp
		}, nil
	}

	// Check for draft end: Client.SceneChange with fromSceneName="Draft"
	if fromScene, ok := entry.JSON["fromSceneName"]; ok && fromScene == "Draft" {
		if toScene, ok := entry.JSON["toSceneName"]; ok && toScene == "DeckBuilder" {
			return &DraftSessionEvent{
				Type:      "ended",
				Timestamp: time.Now(),
			}, nil
		}
	}

	// Check for BotDraftDraftStatus (Quick Draft initial state)
	if strings.Contains(entry.Raw, "BotDraftDraftStatus") && strings.Contains(entry.Raw, "<==") {
		return parseDraftStatus(entry)
	}

	// Check for BotDraftDraftPick (Quick Draft pick)
	if strings.Contains(entry.Raw, "BotDraftDraftPick") && strings.Contains(entry.Raw, "==>") {
		return parseDraftPick(entry)
	}

	// Check for EventPlayerDraftMakePick (Premier Draft pick)
	if strings.Contains(entry.Raw, "EventPlayerDraftMakePick") && strings.Contains(entry.Raw, "==>") {
		return parsePremierDraftPick(entry)
	}

	// Check for Draft.Notify (Premier Draft pack notification)
	if strings.Contains(entry.Raw, "Draft.Notify") {
		return parsePremierDraftNotify(entry)
	}

	// Check for EventJoin (captures event name for Premier Draft)
	if strings.Contains(entry.Raw, "EventJoin") && strings.Contains(entry.Raw, "<==") {
		return parseEventJoin(entry)
	}

	return nil, nil
}

// parseDraftStatus parses a BotDraftDraftStatus response.
func parseDraftStatus(entry *LogEntry) (*DraftSessionEvent, error) {
	// Look for the Payload field which contains the actual draft data
	var payload map[string]interface{}

	if currentModule, ok := entry.JSON["CurrentModule"]; ok && currentModule == "BotDraft" {
		if payloadStr, ok := entry.JSON["Payload"].(string); ok {
			// The Payload is a JSON string nested inside the JSON
			if err := json.Unmarshal([]byte(payloadStr), &payload); err != nil {
				return nil, fmt.Errorf("parse draft status payload: %w", err)
			}
		} else {
			return nil, nil
		}
	} else {
		return nil, nil
	}

	// Extract event name to get set code
	eventName, _ := payload["EventName"].(string)
	setCode := extractSetCode(eventName)

	// Extract pack number and pick number
	packNumber := 0
	if pn, ok := payload["PackNumber"].(float64); ok {
		packNumber = int(pn)
	}

	pickNumber := 0
	if pn, ok := payload["PickNumber"].(float64); ok {
		pickNumber = int(pn)
	}

	// Extract draft pack (available cards)
	draftPack := []string{}
	if pack, ok := payload["DraftPack"].([]interface{}); ok {
		for _, cardID := range pack {
			if id, ok := cardID.(string); ok {
				draftPack = append(draftPack, id)
			}
		}
	}

	// Extract picked cards
	pickedCards := []string{}
	if picked, ok := payload["PickedCards"].([]interface{}); ok {
		for _, cardID := range picked {
			if id, ok := cardID.(string); ok {
				pickedCards = append(pickedCards, id)
			}
		}
	}

	return &DraftSessionEvent{
		Type:        "status_updated",
		EventName:   eventName,
		SetCode:     setCode,
		PackNumber:  packNumber,
		PickNumber:  pickNumber,
		DraftPack:   draftPack,
		PickedCards: pickedCards,
		Timestamp:   time.Now(),
	}, nil
}

// parseDraftPick parses a BotDraftDraftPick request.
func parseDraftPick(entry *LogEntry) (*DraftSessionEvent, error) {
	// The request is nested in a JSON string
	requestStr, ok := entry.JSON["request"].(string)
	if !ok {
		return nil, nil
	}

	var request map[string]interface{}
	if err := json.Unmarshal([]byte(requestStr), &request); err != nil {
		return nil, fmt.Errorf("parse draft pick request: %w", err)
	}

	// Extract event name
	eventName, _ := request["EventName"].(string)
	setCode := extractSetCode(eventName)

	// Extract pick info
	pickInfo, ok := request["PickInfo"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	packNumber := 0
	if pn, ok := pickInfo["PackNumber"].(float64); ok {
		packNumber = int(pn)
	}

	pickNumber := 0
	if pn, ok := pickInfo["PickNumber"].(float64); ok {
		pickNumber = int(pn)
	}

	// Extract selected cards
	selectedCard := []string{}
	if cards, ok := pickInfo["CardIds"].([]interface{}); ok {
		for _, cardID := range cards {
			if id, ok := cardID.(string); ok {
				selectedCard = append(selectedCard, id)
			}
		}
	}

	return &DraftSessionEvent{
		Type:         "pick_made",
		EventName:    eventName,
		SetCode:      setCode,
		PackNumber:   packNumber,
		PickNumber:   pickNumber,
		SelectedCard: selectedCard,
		Timestamp:    time.Now(),
	}, nil
}

// parsePremierDraftPick parses an EventPlayerDraftMakePick request (Premier Draft).
// Example: ==> EventPlayerDraftMakePick {"request":"{\"DraftId\":\"...\",\"GrpIds\":[97380],\"Pack\":1,\"Pick\":1}"}
func parsePremierDraftPick(entry *LogEntry) (*DraftSessionEvent, error) {
	// The request is nested in a JSON string
	requestStr, ok := entry.JSON["request"].(string)
	if !ok {
		return nil, nil
	}

	var request map[string]interface{}
	if err := json.Unmarshal([]byte(requestStr), &request); err != nil {
		return nil, fmt.Errorf("parse premier draft pick request: %w", err)
	}

	// Extract draft ID
	draftID, _ := request["DraftId"].(string)

	// Extract pack and pick numbers (1-indexed in Premier Draft)
	packNumber := 0
	if pn, ok := request["Pack"].(float64); ok {
		packNumber = int(pn) - 1 // Convert to 0-indexed
	}

	pickNumber := 0
	if pn, ok := request["Pick"].(float64); ok {
		pickNumber = int(pn)
	}

	// Extract selected card IDs
	selectedCard := []string{}
	if grpIds, ok := request["GrpIds"].([]interface{}); ok {
		for _, cardID := range grpIds {
			if idFloat, ok := cardID.(float64); ok {
				selectedCard = append(selectedCard, fmt.Sprintf("%d", int(idFloat)))
			}
		}
	}

	return &DraftSessionEvent{
		Type:         "pick_made",
		SessionID:    draftID,
		PackNumber:   packNumber,
		PickNumber:   pickNumber,
		SelectedCard: selectedCard,
		Timestamp:    time.Now(),
	}, nil
}

// parsePremierDraftNotify parses a Draft.Notify message (Premier Draft pack update).
// Example: Draft.Notify {"draftId":"...","SelfPick":2,"SelfPack":1,"PackCards":"97530,97468,..."}
func parsePremierDraftNotify(entry *LogEntry) (*DraftSessionEvent, error) {
	// Extract draft ID
	draftID, _ := entry.JSON["draftId"].(string)

	// Extract pack and pick numbers (1-indexed)
	selfPack := 0
	packNumber := 0
	if pn, ok := entry.JSON["SelfPack"].(float64); ok {
		selfPack = int(pn)
		packNumber = int(pn) - 1 // Convert to 0-indexed
	}

	selfPick := 0
	pickNumber := 0
	if pn, ok := entry.JSON["SelfPick"].(float64); ok {
		selfPick = int(pn)
		pickNumber = int(pn)
	}

	// Extract pack cards (comma-separated string)
	draftPack := []string{}
	if packCardsStr, ok := entry.JSON["PackCards"].(string); ok {
		if packCardsStr != "" {
			draftPack = strings.Split(packCardsStr, ",")
		}
	}

	fmt.Printf("[parsePremierDraftNotify] Draft.Notify: SelfPack=%d, SelfPick=%d -> pack=%d, pick=%d, cards=%d\n",
		selfPack, selfPick, packNumber, pickNumber, len(draftPack))

	return &DraftSessionEvent{
		Type:       "status_updated",
		SessionID:  draftID,
		PackNumber: packNumber,
		PickNumber: pickNumber,
		DraftPack:  draftPack,
		Timestamp:  time.Now(),
	}, nil
}

// parseEventJoin parses an EventJoin response to capture event name and session ID.
// Example: <== EventJoin(...) {"Course":{"CourseId":"...","InternalEventName":"PremierDraft_TLA_20251118",...}}
func parseEventJoin(entry *LogEntry) (*DraftSessionEvent, error) {
	// Check if this is a draft event (Premier or Quick Draft)
	course, ok := entry.JSON["Course"].(map[string]interface{})
	if !ok {
		return nil, nil
	}

	eventName, _ := course["InternalEventName"].(string)
	if !strings.Contains(eventName, "Draft") {
		return nil, nil // Not a draft event
	}

	courseID, _ := course["CourseId"].(string)
	setCode := extractSetCode(eventName)

	return &DraftSessionEvent{
		Type:      "session_info",
		SessionID: courseID,
		EventName: eventName,
		SetCode:   setCode,
		Timestamp: time.Now(),
	}, nil
}

// extractSetCode extracts the set code from an event name.
// Example: "QuickDraft_TDM_20251111" -> "TDM"
func extractSetCode(eventName string) string {
	// Pattern: QuickDraft_XXX_YYYYMMDD or PremierDraft_XXX_YYYYMMDD
	re := regexp.MustCompile(`(?:QuickDraft|PremierDraft)_([A-Z0-9]+)_\d+`)
	matches := re.FindStringSubmatch(eventName)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
