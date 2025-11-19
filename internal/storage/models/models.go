package models

import "time"

// Account represents a player account.
type Account struct {
	ID           int
	Name         string
	ScreenName   *string // Nullable
	ClientID     *string // Nullable
	DailyWins    int     // Current daily win count (0-15)
	WeeklyWins   int     // Current weekly win count (0-15)
	MasteryLevel int     // Current mastery pass level
	MasteryPass  string  // "Basic" (free) or "Advanced" (paid)
	MasteryMax   int     // Maximum mastery level for current season
	IsDefault    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
	OpponentName    *string // Nullable: opponent's display name
	OpponentID      *string // Nullable: opponent's unique identifier
	CreatedAt       time.Time
}

// Game represents a single game within a match.
type Game struct {
	ID              int
	MatchID         string
	GameNumber      int
	Result          string  // "win" or "loss"
	DurationSeconds *int    // Nullable
	ResultReason    *string // Nullable: "concede", "timeout", "normal", etc.
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

// DraftPick represents a single pick made during a draft event.
type DraftPick struct {
	ID             int
	DraftEventID   string // Foreign key to draft_events
	PackNumber     int
	PickNumber     int
	AvailableCards []int     // Card IDs available in the pack
	SelectedCard   int       // Card ID that was picked
	Timestamp      time.Time // When the pick was made
	CreatedAt      time.Time
}

// StatsFilter provides filtering options for statistics queries.
type StatsFilter struct {
	AccountID    *int // Filter by account ID, nil means all accounts
	StartDate    *time.Time
	EndDate      *time.Time
	Format       *string  // Single format filter (for backward compatibility) - filters matches.format (Ladder/Play)
	Formats      []string // Multiple format filter (e.g., ["Ladder", "Play"]) - filters matches.format
	DeckFormat   *string  // Filter by deck format (Standard, Historic, etc.) - filters decks.format via JOIN
	DeckID       *string
	EventName    *string  // Filter by event name (exact match)
	EventNames   []string // Multiple event names (OR logic)
	OpponentName *string  // Filter by opponent name (exact match)
	OpponentID   *string  // Filter by opponent ID
	Result       *string  // Filter by result ("win" or "loss")
	RankClass    *string  // Filter by rank class (e.g., "Mythic", "Diamond")
	RankMinClass *string  // Minimum rank class
	RankMaxClass *string  // Maximum rank class
	ResultReason *string  // Filter by result reason (e.g., "concede", "timeout")
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

// StreakStats represents win/loss streak information.
type StreakStats struct {
	CurrentStreak     int // Positive = wins, negative = losses, 0 = no streak
	LongestWinStreak  int
	LongestLossStreak int
}

// PerformanceMetrics represents duration-based performance metrics.
type PerformanceMetrics struct {
	AvgMatchDuration *float64 // Average match duration in seconds
	AvgGameDuration  *float64 // Average game duration in seconds
	FastestMatch     *int     // Fastest match duration in seconds
	SlowestMatch     *int     // Slowest match duration in seconds
	FastestGame      *int     // Fastest game duration in seconds
	SlowestGame      *int     // Slowest game duration in seconds
}

// CurrencyHistory tracks changes to gems and gold over time.
type CurrencyHistory struct {
	ID        int
	AccountID int       // Foreign key to accounts
	Timestamp time.Time // When the currency snapshot was taken
	Gems      int       // Current gems amount
	Gold      int       // Current gold amount
	GemsDelta int       // Change in gems since last snapshot
	GoldDelta int       // Change in gold since last snapshot
	Source    *string   // Nullable: where the currency came from
	CreatedAt time.Time
}

// CurrencySnapshot represents a summary of currency at a point in time.
type CurrencySnapshot struct {
	Gems      int
	Gold      int
	Timestamp time.Time
}

// SetCompletion represents completion statistics for a MTG set.
type SetCompletion struct {
	SetCode         string
	SetName         string
	TotalCards      int
	OwnedCards      int
	Percentage      float64
	RarityBreakdown map[string]*RarityCompletion
}

// RarityCompletion represents completion for a specific rarity.
type RarityCompletion struct {
	Rarity     string
	Total      int
	Owned      int
	Percentage float64
}

// SeasonalRankSummary represents rank information for a single season.
type SeasonalRankSummary struct {
	SeasonOrdinal  int
	Format         string
	StartRank      *string // First recorded rank in the season
	EndRank        *string // Last recorded rank in the season
	HighestRank    *string // Best rank achieved in the season
	LowestRank     *string // Worst rank during the season
	TotalSnapshots int     // Number of rank snapshots in the season
	FirstSeen      time.Time
	LastSeen       time.Time
}

// RankAchievement represents a milestone or achievement in rank progression.
type RankAchievement struct {
	Format        string
	RankClass     string    // The rank class achieved (Bronze, Silver, Gold, etc.)
	RankLevel     *int      // Tier within class (optional)
	FirstAchieved time.Time // When this rank was first reached
	SeasonOrdinal int       // Season when first achieved
	IsHighest     bool      // Whether this is the highest rank ever achieved
}

// RankProgression represents progress toward next rank tier.
type RankProgression struct {
	CurrentRank      string    // Current rank (e.g., "Gold 2")
	NextRank         string    // Next rank target (e.g., "Gold 1")
	CurrentStep      int       // Current step within tier
	StepsToNext      int       // Steps needed to reach next tier
	IsAtFloor        bool      // Whether current rank is at a floor
	EstimatedMatches *int      // Estimated matches needed (based on win rate)
	WinRateUsed      *float64  // Win rate used for estimation
	Format           string    // "constructed" or "limited"
	LastUpdated      time.Time // When this was calculated
}

// RankFloor represents a rank floor (ranks below which you cannot drop).
type RankFloor struct {
	RankClass string // "Bronze", "Silver", "Gold", etc.
	RankLevel int    // Tier level (e.g., 4 for Bronze 4)
	Format    string // "constructed" or "limited"
}

// DoubleRankUp represents a detected double rank up event.
type DoubleRankUp struct {
	PreviousRank  string    // Rank before the jump
	NewRank       string    // Rank after the jump
	SkippedRank   string    // The rank that was skipped
	MatchID       string    // Match that triggered the double rank up
	Timestamp     time.Time // When it occurred
	Format        string    // "constructed" or "limited"
	SeasonOrdinal int       // Season when it occurred
}

// Achievement represents a player achievement/milestone in MTGA.
type Achievement struct {
	ID              int
	AccountID       int // Foreign key to accounts
	GraphID         string
	NodeID          string
	Status          string // "Available", "Completed", "InProgress"
	CurrentProgress int    // Current progress toward achievement
	MaxProgress     *int   // Nullable: max progress required (if applicable)
	CompletedAt     *time.Time
	FirstSeen       time.Time // When first detected
	LastUpdated     time.Time // Last progress update
	CreatedAt       time.Time
}

// AchievementStats represents achievement statistics.
type AchievementStats struct {
	TotalAchievements     int
	CompletedAchievements int
	InProgressCount       int
	CompletionRate        float64
	RecentlyCompleted     int // Completed in last 7 days
	CloseToComplete       int // Within 90% of completion
}

// DraftSession represents a Quick Draft session parsed from MTGA logs.
type DraftSession struct {
	ID                   string
	EventName            string
	SetCode              string
	DraftType            string
	StartTime            time.Time
	EndTime              *time.Time
	Status               string // "in_progress", "completed", "abandoned"
	TotalPicks           int
	OverallGrade         *string  // A+, A, A-, B+, etc.
	OverallScore         *int     // 0-100
	PickQualityScore     *float64 // Component score (0-40)
	ColorDisciplineScore *float64 // Component score (0-20)
	DeckCompositionScore *float64 // Component score (0-25)
	StrategicScore       *float64 // Component score (0-15)
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// DraftPickSession represents a single pick made during a draft session.
type DraftPickSession struct {
	ID                 int
	SessionID          string
	PackNumber         int
	PickNumber         int
	CardID             string
	Timestamp          time.Time
	PickQualityGrade   *string  // A+, A, B, C, D, F
	PickQualityRank    *int     // Rank in pack (1 = best)
	PackBestGIHWR      *float64 // Best GIHWR in pack
	PickedCardGIHWR    *float64 // GIHWR of picked card
	AlternativesJSON   *string  // JSON array of alternative picks
}

// DraftPackSession represents the cards available in a pack during a draft.
type DraftPackSession struct {
	ID         int
	SessionID  string
	PackNumber int
	PickNumber int
	CardIDs    []string
	Timestamp  time.Time
}

// SetCard represents a card from a specific MTG set, cached from Scryfall.
type SetCard struct {
	ID            int
	SetCode       string
	ArenaID       string
	ScryfallID    string
	Name          string
	ManaCost      string
	CMC           int
	Types         []string
	Colors        []string
	Rarity        string
	Text          string
	Power         string
	Toughness     string
	ImageURL      string
	ImageURLSmall string
	ImageURLArt   string
	FetchedAt     time.Time
}
