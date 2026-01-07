package models

import "time"

// StandardConfig holds configuration for Standard format rotation.
type StandardConfig struct {
	ID               int       `json:"id"`
	NextRotationDate time.Time `json:"nextRotationDate"`
	RotationEnabled  bool      `json:"rotationEnabled"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// StandardSet represents a set with Standard-specific metadata.
type StandardSet struct {
	Code              string  `json:"code"`
	Name              string  `json:"name"`
	ReleasedAt        *string `json:"releasedAt,omitempty"`
	RotationDate      *string `json:"rotationDate,omitempty"`
	IsStandardLegal   bool    `json:"isStandardLegal"`
	IconSVGURI        *string `json:"iconSvgUri,omitempty"`
	CardCount         *int    `json:"cardCount,omitempty"`
	DaysUntilRotation *int    `json:"daysUntilRotation,omitempty"`
	IsRotatingSoon    bool    `json:"isRotatingSoon"` // < 90 days until rotation
}

// CardLegality represents format legalities for a card.
// Values are: "legal", "not_legal", "banned", "restricted"
type CardLegality struct {
	Standard  string `json:"standard"`
	Historic  string `json:"historic"`
	Explorer  string `json:"explorer"`
	Pioneer   string `json:"pioneer"`
	Modern    string `json:"modern"`
	Alchemy   string `json:"alchemy"`
	Brawl     string `json:"brawl"`
	Commander string `json:"commander"`
}

// DeckValidationResult contains results of Standard validation.
type DeckValidationResult struct {
	IsLegal       bool                `json:"isLegal"`
	Errors        []ValidationError   `json:"errors,omitempty"`
	Warnings      []ValidationWarning `json:"warnings,omitempty"`
	RotatingCards []RotatingCard      `json:"rotatingCards,omitempty"`
	SetBreakdown  []DeckSetInfo       `json:"setBreakdown"`
}

// ValidationError represents a hard legality error that makes a deck illegal.
type ValidationError struct {
	CardID   int    `json:"cardId"`
	CardName string `json:"cardName"`
	Reason   string `json:"reason"`  // "not_legal", "banned", "too_many_copies", "deck_size"
	Details  string `json:"details"` // Human-readable explanation
}

// ValidationWarning represents a soft warning (rotating cards, etc.) that doesn't affect legality.
type ValidationWarning struct {
	CardID   int    `json:"cardId"`
	CardName string `json:"cardName"`
	Type     string `json:"type"`    // "rotating_soon", "alchemy_only", "restricted"
	Details  string `json:"details"` // Human-readable explanation
}

// RotatingCard represents a card that will rotate out of Standard.
type RotatingCard struct {
	CardID            int    `json:"cardId"`
	CardName          string `json:"cardName"`
	SetCode           string `json:"setCode"`
	SetName           string `json:"setName"`
	RotationDate      string `json:"rotationDate"`
	DaysUntilRotation int    `json:"daysUntilRotation"`
}

// DeckSetInfo shows how many cards in a deck are from each set.
type DeckSetInfo struct {
	SetCode    string `json:"setCode"`
	SetName    string `json:"setName"`
	CardCount  int    `json:"cardCount"`
	IconSVGURI string `json:"iconSvgUri"`
	IsRotating bool   `json:"isRotating"` // Will rotate at next rotation
}

// RotationAffectedDeck represents a deck that will lose cards at the next rotation.
type RotationAffectedDeck struct {
	DeckID            string         `json:"deckId"`
	DeckName          string         `json:"deckName"`
	Format            string         `json:"format"`
	RotatingCardCount int            `json:"rotatingCardCount"`
	TotalCards        int            `json:"totalCards"`
	PercentAffected   float64        `json:"percentAffected"`
	RotatingCards     []RotatingCard `json:"rotatingCards"`
}

// UpcomingRotation contains information about the next Standard rotation.
type UpcomingRotation struct {
	NextRotationDate  string        `json:"nextRotationDate"`
	DaysUntilRotation int           `json:"daysUntilRotation"`
	RotatingSets      []StandardSet `json:"rotatingSets"`
	RotatingCardCount int           `json:"rotatingCardCount"`
	AffectedDecks     int           `json:"affectedDecks"`
}
