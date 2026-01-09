package models

import "time"

// OpponentDeckProfile represents a reconstructed opponent deck from observed cards.
type OpponentDeckProfile struct {
	ID                  int       `json:"id" db:"id"`
	MatchID             string    `json:"matchId" db:"match_id"`
	DetectedArchetype   *string   `json:"detectedArchetype" db:"detected_archetype"`
	ArchetypeConfidence float64   `json:"archetypeConfidence" db:"archetype_confidence"`
	ColorIdentity       string    `json:"colorIdentity" db:"color_identity"`
	DeckStyle           *string   `json:"deckStyle" db:"deck_style"` // aggro, control, midrange, combo
	CardsObserved       int       `json:"cardsObserved" db:"cards_observed"`
	EstimatedDeckSize   int       `json:"estimatedDeckSize" db:"estimated_deck_size"`
	ObservedCardIDs     *string   `json:"observedCardIds" db:"observed_card_ids"` // JSON array
	InferredCardIDs     *string   `json:"inferredCardIds" db:"inferred_card_ids"` // JSON array
	SignatureCards      *string   `json:"signatureCards" db:"signature_cards"`    // JSON array
	Format              *string   `json:"format" db:"format"`
	MetaArchetypeID     *int      `json:"metaArchetypeId" db:"meta_archetype_id"`
	CreatedAt           time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt           time.Time `json:"updatedAt" db:"updated_at"`
}

// OpponentAnalysis provides detailed analysis of an opponent's deck.
type OpponentAnalysis struct {
	Profile           *OpponentDeckProfile `json:"profile"`
	ObservedCards     []ObservedCard       `json:"observedCards"`
	ExpectedCards     []ExpectedCard       `json:"expectedCards"`
	StrategicInsights []StrategicInsight   `json:"strategicInsights"`
	MatchupStats      *MatchupStatistic    `json:"matchupStats,omitempty"`
	MetaArchetype     *MetaArchetypeMatch  `json:"metaArchetype,omitempty"`
}

// ObservedCard represents a card that was observed during the match.
type ObservedCard struct {
	CardID        int     `json:"cardId"`
	CardName      string  `json:"cardName"`
	Zone          string  `json:"zone"`
	TurnFirstSeen int     `json:"turnFirstSeen"`
	TimesSeen     int     `json:"timesSeen"`
	IsSignature   bool    `json:"isSignature"`
	Category      *string `json:"category,omitempty"` // removal, threat, etc.
}

// ExpectedCard represents a card likely in opponent's deck based on archetype.
type ExpectedCard struct {
	CardID        int     `json:"cardId"`
	CardName      string  `json:"cardName"`
	InclusionRate float64 `json:"inclusionRate"` // 0.0-1.0
	AvgCopies     float64 `json:"avgCopies"`
	WasSeen       bool    `json:"wasSeen"`
	Category      string  `json:"category"`             // removal, threat, wincon, interaction
	PlayAround    string  `json:"playAround,omitempty"` // How to play around this card
}

// StrategicInsight provides actionable advice based on detected archetype.
type StrategicInsight struct {
	Type        string `json:"type"`        // removal, counter, threat, wincon
	Description string `json:"description"` // What to expect/play around
	Priority    string `json:"priority"`    // high, medium, low
	Cards       []int  `json:"cards"`       // Related card IDs
}

// MetaArchetypeMatch represents a match to a known meta archetype.
type MetaArchetypeMatch struct {
	ArchetypeID   int     `json:"archetypeId"`
	ArchetypeName string  `json:"archetypeName"`
	MetaShare     float64 `json:"metaShare"`  // Percentage in meta
	Tier          int     `json:"tier"`       // 1-4
	Confidence    float64 `json:"confidence"` // Match confidence
	Source        string  `json:"source"`     // mtggoldfish, mtgtop8, etc.
}

// MatchupStatistic tracks win/loss stats against an archetype.
type MatchupStatistic struct {
	ID                int        `json:"id" db:"id"`
	AccountID         int        `json:"accountId" db:"account_id"`
	PlayerArchetype   string     `json:"playerArchetype" db:"player_archetype"`
	OpponentArchetype string     `json:"opponentArchetype" db:"opponent_archetype"`
	Format            string     `json:"format" db:"format"`
	TotalMatches      int        `json:"totalMatches" db:"total_matches"`
	Wins              int        `json:"wins" db:"wins"`
	Losses            int        `json:"losses" db:"losses"`
	WinRate           float64    `json:"winRate"` // Calculated field
	AvgGameDuration   *int       `json:"avgGameDuration" db:"avg_game_duration"`
	LastMatchAt       *time.Time `json:"lastMatchAt" db:"last_match_at"`
	CreatedAt         time.Time  `json:"createdAt" db:"created_at"`
	UpdatedAt         time.Time  `json:"updatedAt" db:"updated_at"`
}

// ArchetypeExpectedCard represents expected cards for an archetype.
type ArchetypeExpectedCard struct {
	ID            int       `json:"id" db:"id"`
	ArchetypeName string    `json:"archetypeName" db:"archetype_name"`
	Format        string    `json:"format" db:"format"`
	CardID        int       `json:"cardId" db:"card_id"`
	CardName      string    `json:"cardName" db:"card_name"`
	InclusionRate float64   `json:"inclusionRate" db:"inclusion_rate"`
	AvgCopies     float64   `json:"avgCopies" db:"avg_copies"`
	IsSignature   bool      `json:"isSignature" db:"is_signature"`
	Category      *string   `json:"category" db:"category"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
}

// OpponentHistorySummary provides aggregated stats about opponents faced.
type OpponentHistorySummary struct {
	TotalOpponents      int                       `json:"totalOpponents"`
	UniqueArchetypes    int                       `json:"uniqueArchetypes"`
	MostCommonArchetype string                    `json:"mostCommonArchetype"`
	MostCommonCount     int                       `json:"mostCommonCount"`
	ArchetypeBreakdown  []ArchetypeBreakdownEntry `json:"archetypeBreakdown"`
	ColorIdentityStats  []ColorIdentityStatsEntry `json:"colorIdentityStats"`
}

// ArchetypeBreakdownEntry represents stats for a single archetype.
type ArchetypeBreakdownEntry struct {
	Archetype  string  `json:"archetype"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
	WinRate    float64 `json:"winRate"`
}

// ColorIdentityStatsEntry represents stats for a color identity.
type ColorIdentityStatsEntry struct {
	ColorIdentity string  `json:"colorIdentity"`
	Count         int     `json:"count"`
	Percentage    float64 `json:"percentage"`
	WinRate       float64 `json:"winRate"`
}

// Constants for deck styles.
const (
	DeckStyleAggro    = "aggro"
	DeckStyleMidrange = "midrange"
	DeckStyleControl  = "control"
	DeckStyleCombo    = "combo"
	DeckStyleTempo    = "tempo"
)

// Constants for card categories.
const (
	CardCategoryRemoval     = "removal"
	CardCategoryThreat      = "threat"
	CardCategoryInteraction = "interaction"
	CardCategoryWincon      = "wincon"
	CardCategoryUtility     = "utility"
	CardCategoryRamp        = "ramp"
	CardCategoryCardDraw    = "card_draw"
)

// Constants for insight priority.
const (
	InsightPriorityHigh   = "high"
	InsightPriorityMedium = "medium"
	InsightPriorityLow    = "low"
)
