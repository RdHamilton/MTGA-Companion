package seventeenlands

import (
	"context"
	"time"
)

// StorageAdapter adapts the storage layer to implement CacheStorage interface.
// This allows the 17Lands client to use the database for caching.
type StorageAdapter struct {
	storage StorageService
}

// StorageService defines the storage interface needed by the adapter.
type StorageService interface {
	SaveCardRatings(ctx context.Context, ratings []CardRating, expansion, format, colors, startDate, endDate string) error
	GetCardRatingsForSet(ctx context.Context, expansion, format, colors string) ([]*DraftCardRating, error)
	SaveColorRatings(ctx context.Context, ratings []ColorRating, expansion, eventType, startDate, endDate string) error
	GetColorRatings(ctx context.Context, expansion, eventType string) ([]*DraftColorRating, error)
}

// DraftCardRating represents a cached card rating from storage.
type DraftCardRating struct {
	ID        int
	ArenaID   int
	Expansion string
	Format    string
	Colors    string

	// Win rate metrics
	GIHWR      float64
	OHWR       float64
	GPWR       float64
	GDWR       float64
	IHDWR      float64
	GIHWRDelta float64
	OHWRDelta  float64
	GDWRDelta  float64
	IHDWRDelta float64

	// Draft metrics
	ALSA float64
	ATA  float64

	// Sample sizes
	GIH int
	OH  int
	GP  int
	GD  int
	IHD int

	// Deck metrics
	GamesPlayed int
	NumDecks    int

	// Metadata
	StartDate   string
	EndDate     string
	CachedAt    time.Time
	LastUpdated time.Time
}

// DraftColorRating represents a cached color combination rating from storage.
type DraftColorRating struct {
	ID               int
	Expansion        string
	EventType        string
	ColorCombination string

	// Metrics
	WinRate     float64
	GamesPlayed int
	NumDecks    int

	// Metadata
	StartDate   string
	EndDate     string
	CachedAt    time.Time
	LastUpdated time.Time
}

// NewStorageAdapter creates a new storage adapter.
func NewStorageAdapter(storage StorageService) *StorageAdapter {
	return &StorageAdapter{
		storage: storage,
	}
}

// SaveCardRatings implements CacheStorage.SaveCardRatings.
func (sa *StorageAdapter) SaveCardRatings(ctx context.Context, ratings []CardRating, expansion, format, colors, startDate, endDate string) error {
	return sa.storage.SaveCardRatings(ctx, ratings, expansion, format, colors, startDate, endDate)
}

// GetCardRatingsForSet implements CacheStorage.GetCardRatingsForSet.
func (sa *StorageAdapter) GetCardRatingsForSet(ctx context.Context, expansion, format, colors string) ([]CardRating, time.Time, error) {
	cached, err := sa.storage.GetCardRatingsForSet(ctx, expansion, format, colors)
	if err != nil {
		return nil, time.Time{}, err
	}

	if len(cached) == 0 {
		return nil, time.Time{}, nil
	}

	// Convert storage types to API types
	ratings := make([]CardRating, len(cached))
	var mostRecent time.Time

	for i, c := range cached {
		ratings[i] = CardRating{
			MTGAID:      c.ArenaID,
			GIHWR:       c.GIHWR,
			OHWR:        c.OHWR,
			GPWR:        c.GPWR,
			GDWR:        c.GDWR,
			IHDWR:       c.IHDWR,
			GIHWRDelta:  c.GIHWRDelta,
			OHWRDelta:   c.OHWRDelta,
			GDWRDelta:   c.GDWRDelta,
			IHDWRDelta:  c.IHDWRDelta,
			ALSA:        c.ALSA,
			ATA:         c.ATA,
			GIH:         c.GIH,
			OH:          c.OH,
			GP:          c.GP,
			GD:          c.GD,
			IHD:         c.IHD,
			GamesPlayed: c.GamesPlayed,
			NumberDecks: c.NumDecks,
		}

		// Track most recent cache time
		if c.LastUpdated.After(mostRecent) {
			mostRecent = c.LastUpdated
		}
	}

	return ratings, mostRecent, nil
}

// SaveColorRatings implements CacheStorage.SaveColorRatings.
func (sa *StorageAdapter) SaveColorRatings(ctx context.Context, ratings []ColorRating, expansion, eventType, startDate, endDate string) error {
	return sa.storage.SaveColorRatings(ctx, ratings, expansion, eventType, startDate, endDate)
}

// GetColorRatings implements CacheStorage.GetColorRatings.
func (sa *StorageAdapter) GetColorRatings(ctx context.Context, expansion, eventType string) ([]ColorRating, time.Time, error) {
	cached, err := sa.storage.GetColorRatings(ctx, expansion, eventType)
	if err != nil {
		return nil, time.Time{}, err
	}

	if len(cached) == 0 {
		return nil, time.Time{}, nil
	}

	// Convert storage types to API types
	ratings := make([]ColorRating, len(cached))
	var mostRecent time.Time

	for i, c := range cached {
		ratings[i] = ColorRating{
			ColorName:   c.ColorCombination,
			WinRate:     c.WinRate,
			GamesPlayed: c.GamesPlayed,
			NumberDecks: c.NumDecks,
		}

		// Track most recent cache time
		if c.LastUpdated.After(mostRecent) {
			mostRecent = c.LastUpdated
		}
	}

	return ratings, mostRecent, nil
}
