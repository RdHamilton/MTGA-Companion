package query

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/unified"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// FallbackMode defines how to handle unavailable data.
type FallbackMode int

const (
	RequireAll   FallbackMode = iota // Error if data missing
	AllowPartial                     // Return partial data
	CacheOnly                        // Never query APIs
)

func (fm FallbackMode) String() string {
	switch fm {
	case RequireAll:
		return "RequireAll"
	case AllowPartial:
		return "AllowPartial"
	case CacheOnly:
		return "CacheOnly"
	default:
		return "Unknown"
	}
}

// QueryOptions configures how card queries are performed.
type QueryOptions struct {
	Format       string        // PremierDraft, QuickDraft, etc.
	IncludeStats bool          // Include draft stats if available
	MaxStaleAge  time.Duration // Max age for cached data (0 = require fresh)
	FallbackMode FallbackMode  // How to handle unavailable data
}

// DefaultQueryOptions returns sensible defaults for queries.
func DefaultQueryOptions() QueryOptions {
	return QueryOptions{
		Format:       "PremierDraft",
		IncludeStats: true,
		MaxStaleAge:  24 * time.Hour, // Accept 24h old cache
		FallbackMode: AllowPartial,
	}
}

// CardQuery provides intelligent query interface with priority system.
type CardQuery interface {
	// Get retrieves a single card with automatic priority and caching.
	Get(ctx context.Context, arenaID int, opts QueryOptions) (*unified.UnifiedCard, error)

	// GetMany retrieves multiple cards with batch optimization.
	GetMany(ctx context.Context, arenaIDs []int, opts QueryOptions) ([]*unified.UnifiedCard, error)

	// Search finds cards by name.
	Search(ctx context.Context, name string, opts QueryOptions) ([]*unified.UnifiedCard, error)

	// GetSet retrieves all cards in a set.
	GetSet(ctx context.Context, setCode string, opts QueryOptions) ([]*unified.UnifiedCard, error)

	// Close shuts down background workers.
	Close() error
}

// refreshJob represents a background refresh task.
type refreshJob struct {
	arenaID int
	opts    QueryOptions
}

// UnifiedCardService provides unified card data operations.
type UnifiedCardService interface {
	GetCard(ctx context.Context, arenaID int, setCode, format string) (*unified.UnifiedCard, error)
	GetCards(ctx context.Context, arenaIDs []int, format string) ([]*unified.UnifiedCard, error)
	GetSetCards(ctx context.Context, setCode, format string) ([]*unified.UnifiedCard, error)
}

// CardStorage provides card storage operations.
type CardStorage interface {
	GetCardByArenaID(ctx context.Context, arenaID int) (*storage.Card, error)
	GetCardRating(ctx context.Context, arenaID int, expansion, format, colors string) (*storage.DraftCardRating, error)
	SearchCards(ctx context.Context, name string) ([]*storage.Card, error)
}

// cardQuery implements CardQuery with smart caching and priority system.
type cardQuery struct {
	unifiedService UnifiedCardService
	storage        CardStorage
	logger         *slog.Logger

	// Background refresh
	refreshQueue chan refreshJob
	refreshWg    sync.WaitGroup
	closed       chan struct{}
}

// QueryConfig configures the card query service.
type QueryConfig struct {
	UnifiedService UnifiedCardService
	Storage        CardStorage
	Logger         *slog.Logger
	RefreshWorkers int // Number of background refresh workers
}

// NewCardQuery creates a new card query service.
func NewCardQuery(config QueryConfig) (CardQuery, error) {
	if config.UnifiedService == nil {
		return nil, fmt.Errorf("unifiedService is required")
	}
	if config.Storage == nil {
		return nil, fmt.Errorf("storage is required")
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}
	if config.RefreshWorkers == 0 {
		config.RefreshWorkers = 2 // Default to 2 workers
	}

	q := &cardQuery{
		unifiedService: config.UnifiedService,
		storage:        config.Storage,
		logger:         config.Logger,
		refreshQueue:   make(chan refreshJob, 100), // Buffer up to 100 refresh jobs
		closed:         make(chan struct{}),
	}

	// Start background refresh workers
	for i := 0; i < config.RefreshWorkers; i++ {
		q.refreshWg.Add(1)
		go q.refreshWorker()
	}

	return q, nil
}

// Get retrieves a single card with priority system.
func (q *cardQuery) Get(ctx context.Context, arenaID int, opts QueryOptions) (*unified.UnifiedCard, error) {
	// 1. Check cache
	cached, cacheErr := q.getCached(ctx, arenaID, opts.Format)
	if cacheErr == nil && q.isFresh(cached, opts.MaxStaleAge) {
		q.logger.Debug("Cache hit (fresh)", "arenaID", arenaID)
		return cached, nil
	}

	// 2. If cache-only mode, return cached or error
	if opts.FallbackMode == CacheOnly {
		if cached != nil {
			q.logger.Debug("Cache-only mode: returning stale cache", "arenaID", arenaID)
			return cached, nil
		}
		return nil, fmt.Errorf("card not in cache (cache-only mode)")
	}

	// 3. Fetch fresh data
	fresh, err := q.fetchFresh(ctx, arenaID, opts)
	if err != nil {
		// 4. Fallback to stale cache
		if cached != nil && opts.FallbackMode == AllowPartial {
			q.logger.Warn("Using stale cache due to fetch failure",
				"arenaID", arenaID,
				"error", err,
				"age", cached.MetadataAge)

			// 5. Queue async refresh
			q.queueRefresh(arenaID, opts)

			return cached, nil
		}

		return nil, fmt.Errorf("failed to fetch card: %w", err)
	}

	// 6. Cache fresh data
	if err := q.saveCached(ctx, fresh, opts.Format); err != nil {
		q.logger.Warn("Failed to cache fresh data", "error", err)
	}

	return fresh, nil
}

// GetMany retrieves multiple cards with batch optimization.
func (q *cardQuery) GetMany(ctx context.Context, arenaIDs []int, opts QueryOptions) ([]*unified.UnifiedCard, error) {
	if len(arenaIDs) == 0 {
		return []*unified.UnifiedCard{}, nil
	}

	// 1. Batch cache lookup
	cached, missing := q.getManyCache(ctx, arenaIDs, opts)

	// 2. If all cached and fresh, return
	if len(missing) == 0 {
		q.logger.Debug("Cache hit (all cards)", "count", len(cached))
		return cached, nil
	}

	// 3. If cache-only mode, return what we have
	if opts.FallbackMode == CacheOnly {
		if len(cached) == 0 {
			return nil, fmt.Errorf("no cards in cache (cache-only mode)")
		}
		q.logger.Debug("Cache-only mode: returning partial results",
			"cached", len(cached),
			"missing", len(missing))
		return cached, nil
	}

	// 4. Batch fetch missing
	fresh, err := q.unifiedService.GetCards(ctx, missing, opts.Format)
	if err != nil && opts.FallbackMode == RequireAll {
		return nil, fmt.Errorf("failed to fetch cards: %w", err)
	}

	// 5. Cache fresh data
	for _, card := range fresh {
		if err := q.saveCached(ctx, card, opts.Format); err != nil {
			q.logger.Warn("Failed to cache card", "arenaID", card.ArenaID, "error", err)
		}
	}

	// 6. Combine cached + fresh
	result := append(cached, fresh...)

	q.logger.Debug("Batch query complete",
		"total", len(result),
		"cached", len(cached),
		"fresh", len(fresh))

	return result, nil
}

// Search finds cards by name (searches cached cards only).
func (q *cardQuery) Search(ctx context.Context, name string, opts QueryOptions) ([]*unified.UnifiedCard, error) {
	// Search in cached cards by name
	cards, err := q.storage.SearchCards(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to search cards: %w", err)
	}

	// Convert to unified cards (without stats for now)
	var result []*unified.UnifiedCard
	for _, card := range cards {
		// TODO: Load draft stats if opts.IncludeStats is true
		uc := &unified.UnifiedCard{
			ID:             card.ID,
			ArenaID:        *card.ArenaID,
			Name:           card.Name,
			ManaCost:       card.ManaCost,
			CMC:            card.CMC,
			TypeLine:       card.TypeLine,
			OracleText:     card.OracleText,
			Colors:         card.Colors,
			ColorIdentity:  card.ColorIdentity,
			Rarity:         card.Rarity,
			SetCode:        card.SetCode,
			MetadataSource: unified.SourceCache,
		}
		result = append(result, uc)
	}

	return result, nil
}

// GetSet retrieves all cards in a set.
func (q *cardQuery) GetSet(ctx context.Context, setCode string, opts QueryOptions) ([]*unified.UnifiedCard, error) {
	// Check if we have cached set data
	// For now, delegate to unified service
	cards, err := q.unifiedService.GetSetCards(ctx, setCode, opts.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to get set cards: %w", err)
	}

	// Cache all cards
	for _, card := range cards {
		if err := q.saveCached(ctx, card, opts.Format); err != nil {
			q.logger.Warn("Failed to cache card", "arenaID", card.ArenaID, "error", err)
		}
	}

	return cards, nil
}

// Close shuts down background workers.
func (q *cardQuery) Close() error {
	close(q.closed)
	close(q.refreshQueue)
	q.refreshWg.Wait()
	return nil
}

// getCached retrieves a card from cache.
func (q *cardQuery) getCached(ctx context.Context, arenaID int, format string) (*unified.UnifiedCard, error) {
	// Get metadata from storage
	card, err := q.storage.GetCardByArenaID(ctx, arenaID)
	if err != nil {
		return nil, err
	}

	// Get draft stats if available
	rating, err := q.storage.GetCardRating(ctx, arenaID, card.SetCode, format, "")
	if err != nil {
		// Stats not available, return card without stats
		uc := &unified.UnifiedCard{
			ID:             card.ID,
			ArenaID:        *card.ArenaID,
			Name:           card.Name,
			ManaCost:       card.ManaCost,
			CMC:            card.CMC,
			TypeLine:       card.TypeLine,
			OracleText:     card.OracleText,
			Colors:         card.Colors,
			ColorIdentity:  card.ColorIdentity,
			Rarity:         card.Rarity,
			SetCode:        card.SetCode,
			MetadataSource: unified.SourceCache,
		}
		return uc, nil
	}

	// Convert to unified card with stats
	uc := &unified.UnifiedCard{
		ID:             card.ID,
		ArenaID:        *card.ArenaID,
		Name:           card.Name,
		ManaCost:       card.ManaCost,
		CMC:            card.CMC,
		TypeLine:       card.TypeLine,
		OracleText:     card.OracleText,
		Colors:         card.Colors,
		ColorIdentity:  card.ColorIdentity,
		Rarity:         card.Rarity,
		SetCode:        card.SetCode,
		MetadataSource: unified.SourceCache,
		DraftStats: &unified.DraftStatistics{
			GIHWR:       rating.GIHWR,
			OHWR:        rating.OHWR,
			ALSA:        rating.ALSA,
			ATA:         rating.ATA,
			GIH:         rating.GIH,
			GamesPlayed: rating.GamesPlayed,
			Format:      rating.Format,
			LastUpdated: rating.LastUpdated,
		},
		StatsAge:    time.Since(rating.LastUpdated),
		StatsSource: unified.SourceCache,
	}

	return uc, nil
}

// getManyCache retrieves multiple cards from cache.
func (q *cardQuery) getManyCache(ctx context.Context, arenaIDs []int, opts QueryOptions) (cached []*unified.UnifiedCard, missing []int) {
	for _, arenaID := range arenaIDs {
		card, err := q.getCached(ctx, arenaID, opts.Format)
		if err != nil || !q.isFresh(card, opts.MaxStaleAge) {
			missing = append(missing, arenaID)
			continue
		}
		cached = append(cached, card)
	}
	return cached, missing
}

// fetchFresh fetches fresh data from APIs.
func (q *cardQuery) fetchFresh(ctx context.Context, arenaID int, opts QueryOptions) (*unified.UnifiedCard, error) {
	// Delegate to unified service
	return q.unifiedService.GetCard(ctx, arenaID, "", opts.Format)
}

// saveCached saves a card to cache.
func (q *cardQuery) saveCached(ctx context.Context, card *unified.UnifiedCard, format string) error {
	// Save metadata if not already cached
	// For now, we assume the unified service already caches through its providers
	// This is a placeholder for explicit caching logic
	return nil
}

// isFresh checks if a card is fresh enough based on max stale age.
func (q *cardQuery) isFresh(card *unified.UnifiedCard, maxStaleAge time.Duration) bool {
	if maxStaleAge == 0 {
		return false // Require fresh data
	}
	// Check metadata age
	if card.MetadataAge > maxStaleAge {
		return false
	}
	// Check stats age if stats are present
	if card.HasDraftStats() && card.StatsAge > maxStaleAge {
		return false
	}
	return true
}

// queueRefresh queues a background refresh for a card.
func (q *cardQuery) queueRefresh(arenaID int, opts QueryOptions) {
	select {
	case q.refreshQueue <- refreshJob{arenaID: arenaID, opts: opts}:
		q.logger.Debug("Queued background refresh", "arenaID", arenaID)
	default:
		q.logger.Warn("Refresh queue full, skipping background refresh", "arenaID", arenaID)
	}
}

// refreshWorker processes background refresh jobs.
func (q *cardQuery) refreshWorker() {
	defer q.refreshWg.Done()

	for {
		select {
		case <-q.closed:
			return
		case job, ok := <-q.refreshQueue:
			if !ok {
				return
			}

			// Rate limit background refreshes
			time.Sleep(100 * time.Millisecond)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			fresh, err := q.fetchFresh(ctx, job.arenaID, job.opts)
			cancel()

			if err != nil {
				q.logger.Warn("Background refresh failed",
					"arenaID", job.arenaID,
					"error", err)
				continue
			}

			if err := q.saveCached(context.Background(), fresh, job.opts.Format); err != nil {
				q.logger.Warn("Failed to cache refreshed data",
					"arenaID", job.arenaID,
					"error", err)
				continue
			}

			q.logger.Debug("Background refresh complete",
				"arenaID", job.arenaID,
				"name", fresh.Name)
		}
	}
}
