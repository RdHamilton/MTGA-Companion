package models

import "time"

// CardCooccurrence represents a pair of cards that appear together in decks.
type CardCooccurrence struct {
	ID           int64     `json:"id" db:"id"`
	CardAArenaID int       `json:"cardAArenaId" db:"card_a_arena_id"`
	CardBArenaID int       `json:"cardBArenaId" db:"card_b_arena_id"`
	Format       string    `json:"format" db:"format"`
	Count        int       `json:"count" db:"count"`
	PMIScore     float64   `json:"pmiScore" db:"pmi_score"`
	LastUpdated  time.Time `json:"lastUpdated" db:"last_updated"`
}

// CooccurrenceSource tracks the sources of co-occurrence data.
type CooccurrenceSource struct {
	ID         int64     `json:"id" db:"id"`
	SourceType string    `json:"sourceType" db:"source_type"`
	SourceID   string    `json:"sourceId" db:"source_id"`
	Format     string    `json:"format" db:"format"`
	DeckCount  int       `json:"deckCount" db:"deck_count"`
	CardCount  int       `json:"cardCount" db:"card_count"`
	LastSynced time.Time `json:"lastSynced" db:"last_synced"`
}

// CardFrequency tracks how often a card appears across analyzed decks.
type CardFrequency struct {
	ID          int64     `json:"id" db:"id"`
	CardArenaID int       `json:"cardArenaId" db:"card_arena_id"`
	Format      string    `json:"format" db:"format"`
	DeckCount   int       `json:"deckCount" db:"deck_count"`
	TotalDecks  int       `json:"totalDecks" db:"total_decks"`
	Frequency   float64   `json:"frequency" db:"frequency"`
	LastUpdated time.Time `json:"lastUpdated" db:"last_updated"`
}
