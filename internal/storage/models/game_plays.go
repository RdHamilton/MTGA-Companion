package models

import "time"

// GamePlay represents a single play/action made during a game.
// Actions include card plays, attacks, blocks, land drops, and mulligans.
type GamePlay struct {
	ID             int       `json:"id" db:"id"`
	GameID         int       `json:"game_id" db:"game_id"`
	MatchID        string    `json:"match_id" db:"match_id"`
	TurnNumber     int       `json:"turn_number" db:"turn_number"`
	Phase          string    `json:"phase" db:"phase"`                   // Main1, Combat, Main2, etc.
	Step           string    `json:"step,omitempty" db:"step"`           // BeginCombat, DeclareAttackers, etc.
	PlayerType     string    `json:"player_type" db:"player_type"`       // "player" or "opponent"
	ActionType     string    `json:"action_type" db:"action_type"`       // "play_card", "attack", "block", "land_drop", "mulligan"
	CardID         *int      `json:"card_id,omitempty" db:"card_id"`     // Arena card ID (nullable for some actions)
	CardName       *string   `json:"card_name,omitempty" db:"card_name"` // Card name for display (nullable)
	ZoneFrom       *string   `json:"zone_from,omitempty" db:"zone_from"` // Source zone (hand, library, graveyard, etc.)
	ZoneTo         *string   `json:"zone_to,omitempty" db:"zone_to"`     // Destination zone (battlefield, graveyard, etc.)
	Timestamp      time.Time `json:"timestamp" db:"timestamp"`
	SequenceNumber int       `json:"sequence_number" db:"sequence_number"` // Order within the game
	CreatedAt      time.Time `json:"created_at,omitempty" db:"created_at"`
}

// GameStateSnapshot captures the board state at a specific turn.
type GameStateSnapshot struct {
	ID                  int       `json:"id" db:"id"`
	GameID              int       `json:"game_id" db:"game_id"`
	MatchID             string    `json:"match_id" db:"match_id"`
	TurnNumber          int       `json:"turn_number" db:"turn_number"`
	ActivePlayer        string    `json:"active_player" db:"active_player"` // "player" or "opponent"
	PlayerLife          *int      `json:"player_life,omitempty" db:"player_life"`
	OpponentLife        *int      `json:"opponent_life,omitempty" db:"opponent_life"`
	PlayerCardsInHand   *int      `json:"player_cards_in_hand,omitempty" db:"player_cards_in_hand"`
	OpponentCardsInHand *int      `json:"opponent_cards_in_hand,omitempty" db:"opponent_cards_in_hand"`
	PlayerLandsInPlay   *int      `json:"player_lands_in_play,omitempty" db:"player_lands_in_play"`
	OpponentLandsInPlay *int      `json:"opponent_lands_in_play,omitempty" db:"opponent_lands_in_play"`
	BoardStateJSON      *string   `json:"board_state_json,omitempty" db:"board_state_json"` // JSON snapshot of all permanents on the battlefield
	Timestamp           time.Time `json:"timestamp" db:"timestamp"`
}

// OpponentCardObserved tracks cards revealed by the opponent during a game.
type OpponentCardObserved struct {
	ID            int     `json:"id" db:"id"`
	GameID        int     `json:"game_id" db:"game_id"`
	MatchID       string  `json:"match_id" db:"match_id"`
	CardID        int     `json:"card_id" db:"card_id"`               // Arena card ID
	CardName      *string `json:"card_name,omitempty" db:"card_name"` // Card name for display
	ZoneObserved  string  `json:"zone_observed" db:"zone_observed"`   // Where the card was seen (hand, battlefield, graveyard)
	TurnFirstSeen int     `json:"turn_first_seen" db:"turn_first_seen"`
	TimesSeen     int     `json:"times_seen" db:"times_seen"`
}

// PlayTimelineEntry represents a group of plays during a specific turn/phase.
type PlayTimelineEntry struct {
	Turn     int                `json:"turn"`
	Phase    string             `json:"phase"`
	Plays    []*GamePlay        `json:"plays"`
	Snapshot *GameStateSnapshot `json:"snapshot,omitempty"`
}

// GamePlayFilter provides filtering options for game play queries.
type GamePlayFilter struct {
	MatchID    *string `json:"match_id,omitempty"`
	GameID     *int    `json:"game_id,omitempty"`
	TurnNumber *int    `json:"turn_number,omitempty"`
	PlayerType *string `json:"player_type,omitempty"` // "player" or "opponent"
	ActionType *string `json:"action_type,omitempty"` // "play_card", "attack", "block", "land_drop", "mulligan"
}

// GamePlaySummary provides aggregated play statistics at the match level (across all games).
type GamePlaySummary struct {
	MatchID           string `json:"match_id"`
	TotalPlays        int    `json:"total_plays"`
	PlayerPlays       int    `json:"player_plays"`
	OpponentPlays     int    `json:"opponent_plays"`
	CardPlays         int    `json:"card_plays"`
	Attacks           int    `json:"attacks"`
	Blocks            int    `json:"blocks"`
	LandDrops         int    `json:"land_drops"`
	TotalTurns        int    `json:"total_turns"`
	OpponentCardsSeen int    `json:"opponent_cards_seen"`
}

// Constants for player types.
const (
	PlayerTypePlayer   = "player"
	PlayerTypeOpponent = "opponent"
)

// Constants for action types.
const (
	ActionTypePlayCard = "play_card"
	ActionTypeAttack   = "attack"
	ActionTypeBlock    = "block"
	ActionTypeLandDrop = "land_drop"
	ActionTypeMulligan = "mulligan"
)

// Constants for game phases.
const (
	PhaseBeginning = "Beginning"
	PhaseMain1     = "Main1"
	PhaseCombat    = "Combat"
	PhaseMain2     = "Main2"
	PhaseEnding    = "Ending"
)

// Constants for combat steps.
const (
	StepBeginCombat      = "BeginCombat"
	StepDeclareAttackers = "DeclareAttackers"
	StepDeclareBlockers  = "DeclareBlockers"
	StepCombatDamage     = "CombatDamage"
	StepEndCombat        = "EndCombat"
)

// Constants for zones.
const (
	ZoneHand        = "hand"
	ZoneLibrary     = "library"
	ZoneBattlefield = "battlefield"
	ZoneGraveyard   = "graveyard"
	ZoneExile       = "exile"
	ZoneStack       = "stack"
	ZoneCommand     = "command"
)
