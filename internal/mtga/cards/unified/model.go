package unified

import (
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
)

// UnifiedCard combines Scryfall metadata with 17Lands draft statistics.
type UnifiedCard struct {
	// Core identity (Scryfall)
	ID              string
	ArenaID         int
	Name            string

	// Card metadata (Scryfall)
	ManaCost        string
	CMC             float64
	TypeLine        string
	OracleText      string
	Colors          []string
	ColorIdentity   []string
	Rarity          string
	SetCode         string
	CollectorNumber string
	Power           string
	Toughness       string
	Loyalty         string
	ImageURIs       *scryfall.ImageURIs
	Layout          string
	CardFaces       []scryfall.CardFace
	Legalities      map[string]string
	ReleasedAt      string

	// Draft statistics (17Lands) - nil if unavailable
	DraftStats      *DraftStatistics

	// Data freshness tracking
	MetadataAge     time.Duration // Age of Scryfall data
	StatsAge        time.Duration // Age of 17Lands data
	MetadataSource  DataSource    // Where metadata came from
	StatsSource     DataSource    // Where stats came from
}

// DraftStatistics contains 17Lands draft performance data for a card.
type DraftStatistics struct {
	// Win rate metrics
	GIHWR       float64   // Games in Hand Win Rate
	OHWR        float64   // Opening Hand Win Rate
	GPWR        float64   // Game Present Win Rate
	GDWR        float64   // Game Drawn Win Rate
	IHDWR       float64   // In Hand Drawn Win Rate

	// Improvement metrics
	GIHWRDelta  float64   // GIH Win Rate Delta (improvement)
	OHWRDelta   float64   // OH Win Rate Delta
	GDWRDelta   float64   // GD Win Rate Delta
	IHDWRDelta  float64   // IHD Win Rate Delta

	// Draft metrics
	ALSA        float64   // Average Last Seen At (pick position)
	ATA         float64   // Average Taken At (pick position)

	// Sample sizes
	GIH         int       // Games in Hand count
	OH          int       // Opening Hand count
	GP          int       // Game Present count
	GD          int       // Game Drawn count
	IHD         int       // In Hand Drawn count

	// Deck metrics
	GamesPlayed int       // Total games with this card
	NumDecks    int       // Number of decks containing this card

	// Metadata
	Format      string    // PremierDraft, QuickDraft, TradDraft, etc.
	Colors      string    // Color filter used (empty = all colors)
	StartDate   string    // YYYY-MM-DD format
	EndDate     string    // YYYY-MM-DD format
	LastUpdated time.Time // When this data was cached
}

// DataSource indicates where data came from.
type DataSource int

const (
	SourceUnknown DataSource = iota
	SourceAPI                // Fresh from API
	SourceCache              // From local cache
	SourceFallback           // Fallback/partial data
)

func (ds DataSource) String() string {
	switch ds {
	case SourceAPI:
		return "API"
	case SourceCache:
		return "Cache"
	case SourceFallback:
		return "Fallback"
	default:
		return "Unknown"
	}
}

// HasDraftStats returns true if the card has draft statistics.
func (uc *UnifiedCard) HasDraftStats() bool {
	return uc.DraftStats != nil
}

// HasFreshMetadata returns true if metadata is less than the given age.
func (uc *UnifiedCard) HasFreshMetadata(maxAge time.Duration) bool {
	return uc.MetadataAge < maxAge
}

// HasFreshStats returns true if draft stats are less than the given age.
func (uc *UnifiedCard) HasFreshStats(maxAge time.Duration) bool {
	return uc.HasDraftStats() && uc.StatsAge < maxAge
}

// IsComplete returns true if the card has both metadata and stats.
func (uc *UnifiedCard) IsComplete() bool {
	return uc.ID != "" && uc.HasDraftStats()
}

// GetSampleSize returns the primary sample size (GIH count).
func (ds *DraftStatistics) GetSampleSize() int {
	return ds.GIH
}

// IsSignificant returns true if the sample size is large enough for statistical significance.
// Uses threshold of 100 games as default.
func (ds *DraftStatistics) IsSignificant() bool {
	return ds.GetSampleSize() >= 100
}

// GetPrimaryWinRate returns the most reliable win rate metric (GIHWR).
func (ds *DraftStatistics) GetPrimaryWinRate() float64 {
	return ds.GIHWR
}
