package draft

import (
	"encoding/json"
	"time"
)

// DraftType represents the type of draft event.
type DraftType string

const (
	DraftTypePremier     DraftType = "PremierDraft"
	DraftTypeQuick       DraftType = "QuickDraft"
	DraftTypeTraditional DraftType = "TraditionalDraft"
	DraftTypeSealed      DraftType = "Sealed"
)

// DraftEvent represents a draft session.
type DraftEvent struct {
	ID          string     // Event ID
	Type        DraftType  // Draft type
	SetCode     string     // Set being drafted (e.g., "BLB", "DSK")
	StartTime   time.Time  // When draft started
	EndTime     *time.Time // When draft ended (nil if in progress)
	CurrentPack int        // Current pack number (1-3)
	CurrentPick int        // Current pick number (1-15)
	InProgress  bool       // Whether draft is currently active
}

// Pack represents a single pack of cards in the draft.
type Pack struct {
	PackNumber int       // Pack number (1-3)
	PickNumber int       // Pick number (1-15)
	CardIDs    []int     // Arena IDs of cards in the pack
	Timestamp  time.Time // When this pack was seen
}

// Pick represents a card pick made during the draft.
type Pick struct {
	PackNumber int       // Pack number (1-3)
	PickNumber int       // Pick number (1-15)
	CardID     int       // Arena ID of picked card
	Timestamp  time.Time // When pick was made
}

// DraftState tracks the current state of an active draft.
type DraftState struct {
	Event       DraftEvent // Draft event info
	CurrentPack *Pack      // Current pack being viewed
	Picks       []Pick     // All picks made so far
	AllPacks    []Pack     // All packs seen (for history)
}

// LogEvent represents a parsed log entry relevant to drafting.
type LogEvent struct {
	Timestamp time.Time
	Type      LogEventType
	Data      json.RawMessage
}

// LogEventType represents the type of log event.
type LogEventType string

const (
	// Draft start/end events
	LogEventDraftStart LogEventType = "draft_start"
	LogEventDraftEnd   LogEventType = "draft_end"

	// Pack events
	LogEventNewPack         LogEventType = "new_pack"          // New pack received
	LogEventCardsInPack     LogEventType = "cards_in_pack"     // Pack contents
	LogEventDraftNotify     LogEventType = "draft_notify"      // Draft notification (Premier)
	LogEventDraftPack       LogEventType = "draft_pack"        // Draft pack (Quick)
	LogEventGrantCardPool   LogEventType = "grant_card_pool"   // Sealed pool
	LogEventCoursesCardPool LogEventType = "courses_card_pool" // Sealed pool (alternate format)

	// Pick events
	LogEventMakePick          LogEventType = "make_pick"            // Player made pick
	LogEventPlayerDraftPick   LogEventType = "player_draft_pick"    // Event_PlayerDraftMakePick
	LogEventHumanDraftPick    LogEventType = "human_draft_pick"     // Draft.MakeHumanDraftPick
	LogEventBotDraftPick      LogEventType = "bot_draft_pick"       // BotDraft_DraftPick (Quick)
	LogEventDraftMakePickResp LogEventType = "draft_make_pick_resp" // Draft.MakePick response
)

// CardsInPackPayload represents the "CardsInPack" log entry.
type CardsInPackPayload struct {
	CardsInPack []int `json:"CardsInPack"`
}

// DraftNotifyPayload represents the "Draft.Notify" log entry.
type DraftNotifyPayload struct {
	DraftPack  []int `json:"DraftPack"`
	PackNumber int   `json:"PackNumber"`
	PickNumber int   `json:"PickNumber"`
	SelfPick   int   `json:"SelfPick"`
}

// DraftPackPayload represents the "DraftPack" log entry (Quick Draft).
type DraftPackPayload struct {
	DraftPack   []int  `json:"DraftPack"`
	DraftStatus string `json:"DraftStatus"` // "PickNext" indicates player's turn
}

// MakePickPayload represents pick events.
type MakePickPayload struct {
	GrpId    int  `json:"GrpId"`    // Card ID picked
	Pack     int  `json:"Pack"`     // Pack number
	Pick     int  `json:"Pick"`     // Pick number
	CardId   int  `json:"CardId"`   // Alternate field name
	AutoPick bool `json:"AutoPick"` // Whether it was auto-picked
}

// GrantCardPoolPayload represents sealed pool events.
type GrantCardPoolPayload struct {
	CardsAdded []struct {
		GrpId    int `json:"GrpId"`
		Quantity int `json:"Quantity"`
	} `json:"CardsAdded"`
}

// CoursesPayload represents the alternate sealed pool format.
type CoursesPayload struct {
	Courses []struct {
		CardPool []int `json:"CardPool"`
	} `json:"Courses"`
}

// IsDraftEvent returns true if the event type is a draft-related event.
func IsDraftEvent(eventType LogEventType) bool {
	switch eventType {
	case LogEventDraftStart, LogEventDraftEnd,
		LogEventNewPack, LogEventCardsInPack, LogEventDraftNotify,
		LogEventDraftPack, LogEventGrantCardPool, LogEventCoursesCardPool,
		LogEventMakePick, LogEventPlayerDraftPick, LogEventHumanDraftPick,
		LogEventBotDraftPick, LogEventDraftMakePickResp:
		return true
	default:
		return false
	}
}

// IsPackEvent returns true if the event represents a new pack.
func IsPackEvent(eventType LogEventType) bool {
	switch eventType {
	case LogEventNewPack, LogEventCardsInPack, LogEventDraftNotify,
		LogEventDraftPack, LogEventGrantCardPool, LogEventCoursesCardPool:
		return true
	default:
		return false
	}
}

// IsPickEvent returns true if the event represents a card pick.
func IsPickEvent(eventType LogEventType) bool {
	switch eventType {
	case LogEventMakePick, LogEventPlayerDraftPick, LogEventHumanDraftPick,
		LogEventBotDraftPick, LogEventDraftMakePickResp:
		return true
	default:
		return false
	}
}
