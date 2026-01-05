package models

import "time"

// GamePlay represents a single play/action made during a game.
// Actions include card plays, attacks, blocks, land drops, and mulligans.
type GamePlay struct {
	ID             int
	GameID         int
	MatchID        string
	TurnNumber     int
	Phase          string  // Main1, Combat, Main2, etc.
	Step           string  // BeginCombat, DeclareAttackers, etc.
	PlayerType     string  // "player" or "opponent"
	ActionType     string  // "play_card", "attack", "block", "land_drop", "mulligan"
	CardID         *int    // Arena card ID (nullable for some actions)
	CardName       *string // Card name for display (nullable)
	ZoneFrom       *string // Source zone (hand, library, graveyard, etc.)
	ZoneTo         *string // Destination zone (battlefield, graveyard, etc.)
	Timestamp      time.Time
	SequenceNumber int // Order within the game
	CreatedAt      time.Time
}

// GameStateSnapshot captures the board state at a specific turn.
type GameStateSnapshot struct {
	ID                  int
	GameID              int
	MatchID             string
	TurnNumber          int
	ActivePlayer        string // "player" or "opponent"
	PlayerLife          *int
	OpponentLife        *int
	PlayerCardsInHand   *int
	OpponentCardsInHand *int
	PlayerLandsInPlay   *int
	OpponentLandsInPlay *int
	BoardStateJSON      *string // JSON snapshot of all permanents on the battlefield
	Timestamp           time.Time
}

// OpponentCardObserved tracks cards revealed by the opponent during a game.
type OpponentCardObserved struct {
	ID            int
	GameID        int
	MatchID       string
	CardID        int     // Arena card ID
	CardName      *string // Card name for display
	ZoneObserved  string  // Where the card was seen (hand, battlefield, graveyard)
	TurnFirstSeen int
	TimesSeen     int
}

// PlayTimelineEntry represents a group of plays during a specific turn/phase.
type PlayTimelineEntry struct {
	Turn     int
	Phase    string
	Plays    []*GamePlay
	Snapshot *GameStateSnapshot
}

// GamePlayFilter provides filtering options for game play queries.
type GamePlayFilter struct {
	MatchID    *string
	GameID     *int
	TurnNumber *int
	PlayerType *string // "player" or "opponent"
	ActionType *string // "play_card", "attack", "block", "land_drop", "mulligan"
}

// GamePlaySummary provides a summary of plays for a game.
type GamePlaySummary struct {
	MatchID           string
	GameID            int
	TotalPlays        int
	PlayerPlays       int
	OpponentPlays     int
	CardPlays         int
	Attacks           int
	Blocks            int
	LandDrops         int
	TotalTurns        int
	OpponentCardsSeen int
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
