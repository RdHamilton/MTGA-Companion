package models

import "github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"

// CollectionCardView represents a collection card with full metadata.
type CollectionCardView struct {
	// Card ID and quantity from collection
	CardID   int
	Quantity int

	// Card metadata (may be nil if not available)
	Metadata *cards.Card
}

// DeckCardView represents a deck card with full metadata.
type DeckCardView struct {
	// DeckCard data
	ID       int
	DeckID   string
	CardID   int
	Quantity int
	Board    string // "main" or "sideboard"

	// Card metadata (may be nil if not available)
	Metadata *cards.Card
}

// DraftCardView represents a card picked in a draft with full metadata.
type DraftCardView struct {
	// Card ID
	CardID int

	// Draft context
	Pack  int // Which pack (1, 2, or 3)
	Pick  int // Which pick in the pack
	Round int // Which pick overall in the draft

	// Card metadata (may be nil if not available)
	Metadata *cards.Card
}

// DeckView represents a deck with all its cards and metadata.
type DeckView struct {
	// Deck information
	Deck *Deck

	// Cards in the deck with metadata
	MainboardCards  []*DeckCardView
	SideboardCards  []*DeckCardView
	TotalMainboard  int // Total number of cards in mainboard
	TotalSideboard  int // Total number of cards in sideboard
	ColorIdentity   []string
	ManaCurve       map[int]int // CMC -> count
}
