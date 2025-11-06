package logreader

// PlayerProfile contains player identification information.
type PlayerProfile struct {
	ScreenName string
	ClientID   string
}

// PlayerInventory contains player resources and collection information.
type PlayerInventory struct {
	Gems               int
	Gold               int
	TotalVaultProgress int
	WildCardCommons    int
	WildCardUncommons  int
	WildCardRares      int
	WildCardMythics    int
	Boosters           []Booster
	CustomTokens       map[string]int
}

// Booster represents a booster pack in the player's inventory.
type Booster struct {
	SetCode     string
	Count       int
	CollationID int
}

// PlayerRank contains player ranking information for both constructed and limited formats.
type PlayerRank struct {
	// Constructed format
	ConstructedSeasonOrdinal int
	ConstructedClass         string
	ConstructedLevel         int
	ConstructedPercentile    float64
	ConstructedStep          int

	// Limited format
	LimitedSeasonOrdinal int
	LimitedClass         string
	LimitedLevel         int
	LimitedPercentile    float64
	LimitedStep          int

	// Match statistics
	LimitedMatchesWon  int
	LimitedMatchesLost int
}

// DraftHistory contains a history of draft/limited events.
type DraftHistory struct {
	Drafts []DraftEvent
}

// DraftEvent represents a single draft or limited event.
type DraftEvent struct {
	EventID   string // CourseId from MTGA
	EventName string // InternalEventName (e.g., "PremierDraft_BLB")
	Status    string // Current module/status (e.g., "Draft", "DeckBuild", "CreateMatch")
	Wins      int    // CurrentWins
	Losses    int    // CurrentLosses if available
	Deck      DraftDeck
}

// DraftDeck represents the deck built during a draft.
type DraftDeck struct {
	Name     string
	MainDeck []DeckCard
}

// DeckCard represents a card in a deck with its quantity.
type DeckCard struct {
	CardID   int
	Quantity int
}
