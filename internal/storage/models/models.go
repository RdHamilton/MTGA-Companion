package models

import "time"

// Account represents a player account.
type Account struct {
	ID         int
	Name       string
	ScreenName *string // Nullable
	ClientID   *string // Nullable
	IsDefault  bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Match represents a single match in MTGA.
// A match may consist of multiple games (best-of-3).
type Match struct {
	ID              string
	AccountID       int // Foreign key to accounts
	EventID         string
	EventName       string
	Timestamp       time.Time
	DurationSeconds *int // Nullable
	PlayerWins      int
	OpponentWins    int
	PlayerTeamID    int
	DeckID          *string // Nullable, foreign key to decks
	RankBefore      *string // Nullable
	RankAfter       *string // Nullable
	Format          string
	Result          string  // "win" or "loss"
	ResultReason    *string // Nullable: "concede", "timeout", "normal", etc.
	CreatedAt       time.Time
}

// Game represents a single game within a match.
type Game struct {
	ID              int
	MatchID         string
	GameNumber      int
	Result          string // "win" or "loss"
	DurationSeconds *int   // Nullable
	CreatedAt       time.Time
}

// PlayerStats represents aggregated player statistics for a time period.
type PlayerStats struct {
	ID            int
	AccountID     int // Foreign key to accounts
	Date          time.Time
	Format        string
	MatchesPlayed int
	MatchesWon    int
	GamesPlayed   int
	GamesWon      int
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Deck represents a deck list.
type Deck struct {
	ID            string
	AccountID     int // Foreign key to accounts
	Name          string
	Format        string
	Description   *string // Nullable
	ColorIdentity *string // Nullable
	CreatedAt     time.Time
	ModifiedAt    time.Time
	LastPlayed    *time.Time // Nullable
}

// DeckCard represents a card in a deck.
type DeckCard struct {
	ID       int
	DeckID   string
	CardID   int
	Quantity int
	Board    string // "main" or "sideboard"
}

// CollectionCard represents a card in the player's collection.
type CollectionCard struct {
	AccountID int // Foreign key to accounts
	CardID    int
	Quantity  int
	UpdatedAt time.Time
}

// CollectionHistory tracks changes to the collection over time.
type CollectionHistory struct {
	ID            int
	AccountID     int // Foreign key to accounts
	CardID        int
	QuantityDelta int // Positive or negative change
	QuantityAfter int // Quantity after this change
	Timestamp     time.Time
	Source        *string // Nullable: "pack", "draft", "craft", etc.
	CreatedAt     time.Time
}

// RankHistory tracks rank progression over time.
type RankHistory struct {
	ID            int
	AccountID     int // Foreign key to accounts
	Timestamp     time.Time
	Format        string // "constructed" or "limited"
	SeasonOrdinal int
	RankClass     *string  // Nullable: "Bronze", "Silver", "Gold", etc.
	RankLevel     *int     // Nullable: tier within class
	RankStep      *int     // Nullable: step within tier
	Percentile    *float64 // Nullable: percentile ranking
	CreatedAt     time.Time
}

// DraftEvent represents a draft or sealed event.
type DraftEvent struct {
	ID        string
	AccountID int // Foreign key to accounts
	EventName string
	SetCode   string
	StartTime time.Time
	EndTime   *time.Time // Nullable if event is ongoing
	Wins      int
	Losses    int
	Status    string  // "active", "completed", "abandoned"
	DeckID    *string // Nullable, foreign key to decks
	EntryFee  *string // Nullable: description of entry fee
	Rewards   *string // Nullable: description of rewards
	CreatedAt time.Time
}

// StatsFilter provides filtering options for statistics queries.
type StatsFilter struct {
	AccountID *int // Filter by account ID, nil means all accounts
	StartDate *time.Time
	EndDate   *time.Time
	Format    *string
	DeckID    *string
}

// Statistics represents aggregated statistics.
type Statistics struct {
	TotalMatches int
	MatchesWon   int
	MatchesLost  int
	TotalGames   int
	GamesWon     int
	GamesLost    int
	WinRate      float64
	GameWinRate  float64
}
