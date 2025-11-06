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
