package events

// ============================================================================
// Event Message Types
// These types define the structure of data sent with events.
// Using typed structs provides compile-time safety and IDE support.
// ============================================================================

// StatsUpdatedEvent is the payload for stats:updated events.
// Sent when match/game statistics are updated.
type StatsUpdatedEvent struct {
	Matches int `json:"matches"` // Number of matches updated
	Games   int `json:"games"`   // Number of games updated
}

// RankUpdatedEvent is the payload for rank:updated events.
// Sent when player rank changes.
type RankUpdatedEvent struct {
	Format string `json:"format"` // Ranked format (e.g., "Constructed", "Limited")
	Tier   string `json:"tier"`   // Rank tier (e.g., "Gold", "Platinum")
	Step   string `json:"step"`   // Step within tier (e.g., "1", "2", "3", "4")
}

// QuestUpdatedEvent is the payload for quest:updated events.
// Sent when quest progress changes.
type QuestUpdatedEvent struct {
	Completed int `json:"completed"` // Number of quests completed
	Count     int `json:"count"`     // Total number of quests
}

// DraftUpdatedEvent is the payload for draft:updated events.
// Sent when draft session data changes.
type DraftUpdatedEvent struct {
	Count int `json:"count"` // Number of draft sessions updated
	Picks int `json:"picks"` // Number of picks made
}

// DeckUpdatedEvent is the payload for deck:updated events.
// Sent when deck data changes.
type DeckUpdatedEvent struct {
	Count int `json:"count"` // Number of decks updated
}

// CollectionUpdatedEvent is the payload for collection:updated events.
// Sent when collection data changes (cards added from decks/drafts).
type CollectionUpdatedEvent struct {
	NewCards   int `json:"newCards"`   // Number of new unique cards added
	CardsAdded int `json:"cardsAdded"` // Total cards added to collection
}

// AchievementUpdatedEvent is the payload for achievement:updated events.
// Sent when achievement progress changes.
type AchievementUpdatedEvent struct {
	Count int `json:"count"` // Number of achievements updated
}

// DaemonStatusEvent is the payload for daemon:status events.
// Sent when daemon connection status changes.
type DaemonStatusEvent struct {
	Status    string `json:"status"`    // Connection status ("connected", "standalone", "disconnected")
	Connected bool   `json:"connected"` // Whether daemon is connected
}

// DaemonConnectedEvent is the payload for daemon:connected events.
// Sent when daemon connection is established.
type DaemonConnectedEvent struct {
	Version string `json:"version,omitempty"` // Daemon version (optional)
}

// DaemonErrorEvent is the payload for daemon:error events.
// Sent when daemon encounters an error.
type DaemonErrorEvent struct {
	Error   string `json:"error"`             // Error message
	Code    string `json:"code,omitempty"`    // Error code (optional)
	Details string `json:"details,omitempty"` // Additional details (optional)
}

// ReplayStartedEvent is the payload for replay:started events.
// Sent when log replay begins.
type ReplayStartedEvent struct {
	TotalFiles int `json:"totalFiles"` // Total number of files to replay
}

// ReplayProgressEvent is the payload for replay:progress events.
// Sent during log replay to indicate progress.
type ReplayProgressEvent struct {
	Current     int     `json:"current"`     // Current file being processed
	Total       int     `json:"total"`       // Total files to process
	Percentage  float64 `json:"percentage"`  // Progress percentage (0-100)
	CurrentFile string  `json:"currentFile"` // Name of current file
}

// ReplayPausedEvent is the payload for replay:paused events.
// Sent when replay is paused.
type ReplayPausedEvent struct {
	Current int `json:"current"` // Current position when paused
	Total   int `json:"total"`   // Total files
}

// ReplayResumedEvent is the payload for replay:resumed events.
// Sent when replay is resumed.
type ReplayResumedEvent struct {
	Current int `json:"current"` // Current position when resumed
	Total   int `json:"total"`   // Total files
}

// ReplayCompletedEvent is the payload for replay:completed events.
// Sent when replay finishes successfully.
type ReplayCompletedEvent struct {
	FilesProcessed int     `json:"filesProcessed"` // Number of files processed
	Duration       float64 `json:"duration"`       // Duration in seconds
	MatchesFound   int     `json:"matchesFound"`   // Number of matches found
	DraftsFound    int     `json:"draftsFound"`    // Number of drafts found
}

// ReplayErrorEvent is the payload for replay:error events.
// Sent when replay encounters an error.
type ReplayErrorEvent struct {
	Error   string `json:"error"`             // Error message
	Code    string `json:"code,omitempty"`    // Error code (optional)
	Details string `json:"details,omitempty"` // Additional details (optional)
}

// ReplayDraftDetectedEvent is the payload for replay:draft_detected events.
// Sent when a draft is detected during replay.
type ReplayDraftDetectedEvent struct {
	DraftID   string `json:"draftId"`   // ID of the detected draft
	SetCode   string `json:"setCode"`   // Set code (e.g., "DSK", "BLB")
	EventType string `json:"eventType"` // Draft event type (e.g., "PremierDraft")
}

// ============================================================================
// Outgoing Message Types (sent to daemon)
// ============================================================================

// ReplayLogsMessage is sent to trigger log replay.
type ReplayLogsMessage struct {
	Type      string `json:"type"`      // Always "replay_logs"
	ClearData bool   `json:"clearData"` // Whether to clear existing data first
}

// StartReplayMessage is sent to start replay with specific files.
type StartReplayMessage struct {
	Type  string   `json:"type"`  // Always "start_replay"
	Files []string `json:"files"` // File paths to replay
}

// PauseReplayMessage is sent to pause an active replay.
type PauseReplayMessage struct {
	Type string `json:"type"` // Always "pause_replay"
}

// ResumeReplayMessage is sent to resume a paused replay.
type ResumeReplayMessage struct {
	Type string `json:"type"` // Always "resume_replay"
}

// StopReplayMessage is sent to stop an active replay.
type StopReplayMessage struct {
	Type string `json:"type"` // Always "stop_replay"
}

// PingMessage is sent as a keep-alive.
type PingMessage struct {
	Type string `json:"type"` // Always "ping"
}
