package models

import "time"

// Note categories for deck notes.
const (
	NoteCategoryGeneral   = "general"
	NoteCategoryMatchup   = "matchup"
	NoteCategorySideboard = "sideboard"
	NoteCategoryMulligan  = "mulligan"
)

// Suggestion types for improvement suggestions.
const (
	SuggestionTypeCurve      = "curve"
	SuggestionTypeRemoval    = "removal"
	SuggestionTypeMana       = "mana"
	SuggestionTypeSequencing = "sequencing"
	SuggestionTypeSideboard  = "sideboard"
)

// Suggestion priority levels.
const (
	SuggestionPriorityLow    = "low"
	SuggestionPriorityMedium = "medium"
	SuggestionPriorityHigh   = "high"
)

// DeckNote represents a timestamped note on a deck.
type DeckNote struct {
	ID        int64
	DeckID    string
	Content   string
	Category  string // general, matchup, sideboard, mulligan
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ImprovementSuggestion represents an auto-generated improvement suggestion for a deck.
type ImprovementSuggestion struct {
	ID             int64
	DeckID         string
	SuggestionType string // curve, removal, mana, sequencing, sideboard
	Priority       string // low, medium, high
	Title          string
	Description    string
	Evidence       *string // JSON array of supporting play data (nullable)
	CardReferences *string // JSON array of card IDs involved (nullable)
	IsDismissed    bool
	CreatedAt      time.Time
}

// MatchNotes represents notes and rating for a match.
// These fields are stored in the matches table, not a separate table.
type MatchNotes struct {
	MatchID string
	Notes   string
	Rating  int // 1-5 stars, 0 = not rated
}
