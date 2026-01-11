package models

import "time"

// EDHRECSynergy represents a synergy relationship between two cards from EDHREC.
type EDHRECSynergy struct {
	ID              int64     `json:"id" db:"id"`
	CardName        string    `json:"cardName" db:"card_name"`
	SynergyCardName string    `json:"synergyCardName" db:"synergy_card_name"`
	SynergyScore    float64   `json:"synergyScore" db:"synergy_score"`
	InclusionCount  int       `json:"inclusionCount" db:"inclusion_count"`
	NumDecks        int       `json:"numDecks" db:"num_decks"`
	Lift            float64   `json:"lift" db:"lift"`
	LastUpdated     time.Time `json:"lastUpdated" db:"last_updated"`
}

// EDHRECCardMetadata represents card-level metadata from EDHREC.
type EDHRECCardMetadata struct {
	ID            int64     `json:"id" db:"id"`
	CardName      string    `json:"cardName" db:"card_name"`
	SanitizedName string    `json:"sanitizedName" db:"sanitized_name"`
	NumDecks      int       `json:"numDecks" db:"num_decks"`
	SaltScore     float64   `json:"saltScore" db:"salt_score"`
	ColorIdentity string    `json:"colorIdentity" db:"color_identity"`
	LastUpdated   time.Time `json:"lastUpdated" db:"last_updated"`
}

// EDHRECThemeCard represents a card associated with a theme from EDHREC.
type EDHRECThemeCard struct {
	ID            int64     `json:"id" db:"id"`
	ThemeName     string    `json:"themeName" db:"theme_name"`
	CardName      string    `json:"cardName" db:"card_name"`
	SynergyScore  float64   `json:"synergyScore" db:"synergy_score"`
	IsTopCard     bool      `json:"isTopCard" db:"is_top_card"`
	IsHighSynergy bool      `json:"isHighSynergy" db:"is_high_synergy"`
	LastUpdated   time.Time `json:"lastUpdated" db:"last_updated"`
}

// NormalizeSynergyScore normalizes EDHREC synergy scores to 0.0-1.0 range.
// EDHREC synergy scores typically range from -1.0 to +1.0.
func NormalizeSynergyScore(score float64) float64 {
	// EDHREC synergy is already roughly in -1.0 to 1.0 range
	// Shift and scale to 0.0-1.0
	normalized := (score + 1.0) / 2.0
	if normalized < 0.0 {
		normalized = 0.0
	}
	if normalized > 1.0 {
		normalized = 1.0
	}
	return normalized
}
