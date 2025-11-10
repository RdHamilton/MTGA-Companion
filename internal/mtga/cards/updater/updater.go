package updater

import (
	"context"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/importer"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// Updater handles incremental updates of card data.
type Updater struct {
	client  *scryfall.Client
	storage *storage.Service
	options UpdateOptions
}

// UpdateOptions configures the update process.
type UpdateOptions struct {
	// StaleSetThreshold is the age after which sets are considered stale.
	StaleSetThreshold time.Duration

	// StaleCardThreshold is the age after which cards are considered stale.
	StaleCardThreshold time.Duration

	// Force forces update of all sets regardless of staleness.
	Force bool

	// SpecificSet limits update to a specific set code.
	SpecificSet string

	// Verbose enables verbose logging.
	Verbose bool

	// Progress callback for progress reporting.
	Progress func(message string)
}

// DefaultUpdateOptions returns sensible default options.
func DefaultUpdateOptions() UpdateOptions {
	return UpdateOptions{
		StaleSetThreshold:  30 * 24 * time.Hour, // 30 days
		StaleCardThreshold: 7 * 24 * time.Hour,  // 7 days
		Force:              false,
		Verbose:            false,
	}
}

// NewUpdater creates a new card data updater.
func NewUpdater(client *scryfall.Client, storage *storage.Service, options UpdateOptions) *Updater {
	return &Updater{
		client:  client,
		storage: storage,
		options: options,
	}
}

// UpdateInfo contains information about available updates.
type UpdateInfo struct {
	NewSets        []string
	StaleSets      []string
	TotalNewCards  int
	TotalStaleSets int
	LastUpdate     time.Time
}

// CheckUpdates checks for available updates without applying them.
func (u *Updater) CheckUpdates(ctx context.Context) (*UpdateInfo, error) {
	info := &UpdateInfo{}

	// Get all sets from Scryfall
	if u.options.Progress != nil {
		u.options.Progress("Fetching sets from Scryfall...")
	}

	scryfallSets, err := u.client.GetSets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sets from Scryfall: %w", err)
	}

	// Get local sets
	localSets, err := u.storage.GetAllSets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get local sets: %w", err)
	}

	// Build map of local sets for quick lookup
	localSetMap := make(map[string]*storage.Set)
	for _, set := range localSets {
		localSetMap[set.Code] = set
		if set.LastUpdated.After(info.LastUpdate) {
			info.LastUpdate = set.LastUpdated
		}
	}

	// Check each Scryfall set
	for _, scryfallSet := range scryfallSets.Data {
		// Filter to only Arena-available sets
		if scryfallSet.Digital && scryfallSet.CardCount > 0 {
			localSet, exists := localSetMap[scryfallSet.Code]

			if !exists {
				// New set
				info.NewSets = append(info.NewSets, scryfallSet.Code)
				info.TotalNewCards += scryfallSet.CardCount
			} else {
				// Check if set is stale
				age := time.Since(localSet.LastUpdated)
				if u.options.Force || age > u.options.StaleSetThreshold {
					info.StaleSets = append(info.StaleSets, scryfallSet.Code)
					info.TotalStaleSets++
				}
			}
		}
	}

	return info, nil
}

// UpdateStats contains statistics about an update operation.
type UpdateStats struct {
	SetsProcessed int
	SetsUpdated   int
	CardsAdded    int
	CardsUpdated  int
	Errors        int
	Duration      time.Duration
	StartTime     time.Time
	EndTime       time.Time
}

// Update performs incremental updates of card data.
func (u *Updater) Update(ctx context.Context) (*UpdateStats, error) {
	startTime := time.Now()
	stats := &UpdateStats{
		StartTime: startTime,
	}

	// Get update info first
	updateInfo, err := u.CheckUpdates(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check updates: %w", err)
	}

	if len(updateInfo.NewSets) == 0 && len(updateInfo.StaleSets) == 0 {
		if u.options.Progress != nil {
			u.options.Progress("No updates available")
		}
		stats.EndTime = time.Now()
		stats.Duration = stats.EndTime.Sub(stats.StartTime)
		return stats, nil
	}

	// Process new sets
	if len(updateInfo.NewSets) > 0 {
		if u.options.Progress != nil {
			u.options.Progress(fmt.Sprintf("Processing %d new sets...", len(updateInfo.NewSets)))
		}

		for _, setCode := range updateInfo.NewSets {
			if err := u.updateSet(ctx, setCode, stats); err != nil {
				stats.Errors++
				if u.options.Verbose {
					fmt.Printf("Error updating set %s: %v\n", setCode, err)
				}
			}
		}
	}

	// Process stale sets
	if len(updateInfo.StaleSets) > 0 && !u.options.Force {
		if u.options.Progress != nil {
			u.options.Progress(fmt.Sprintf("Refreshing %d stale sets...", len(updateInfo.StaleSets)))
		}

		for _, setCode := range updateInfo.StaleSets {
			if err := u.updateSet(ctx, setCode, stats); err != nil {
				stats.Errors++
				if u.options.Verbose {
					fmt.Printf("Error refreshing set %s: %v\n", setCode, err)
				}
			}
		}
	}

	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime)

	return stats, nil
}

// updateSet updates cards for a specific set.
func (u *Updater) updateSet(ctx context.Context, setCode string, stats *UpdateStats) error {
	stats.SetsProcessed++

	if u.options.Progress != nil {
		u.options.Progress(fmt.Sprintf("Fetching set: %s", setCode))
	}

	// Get set info from Scryfall
	scryfallSet, err := u.client.GetSet(ctx, setCode)
	if err != nil {
		return fmt.Errorf("failed to get set info: %w", err)
	}

	// Save set to database
	storageSet := &storage.Set{
		Code:       scryfallSet.Code,
		Name:       scryfallSet.Name,
		ReleasedAt: scryfallSet.ReleasedAt,
		CardCount:  scryfallSet.CardCount,
		SetType:    scryfallSet.SetType,
		IconSVGURI: scryfallSet.IconSVGURI,
	}

	if err := u.storage.SaveSet(ctx, storageSet); err != nil {
		return fmt.Errorf("failed to save set: %w", err)
	}

	// Search for cards in this set with Arena IDs
	query := fmt.Sprintf("set:%s game:arena", setCode)

	if u.options.Progress != nil {
		u.options.Progress(fmt.Sprintf("Searching cards in set: %s", setCode))
	}

	searchResult, err := u.client.SearchCards(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to search cards: %w", err)
	}

	if len(searchResult.Data) == 0 {
		if u.options.Verbose {
			fmt.Printf("No Arena cards found in set %s\n", setCode)
		}
		return nil
	}

	// Save cards
	for i, scryfallCard := range searchResult.Data {
		// Skip cards without Arena ID
		if scryfallCard.ArenaID == nil || *scryfallCard.ArenaID == 0 {
			continue
		}

		storageCard := importer.ConvertToStorageCard(&scryfallCard)

		// Check if card exists
		existing, err := u.storage.GetCardByArenaID(ctx, *scryfallCard.ArenaID)
		if err != nil {
			return fmt.Errorf("failed to check existing card: %w", err)
		}

		if err := u.storage.SaveCard(ctx, storageCard); err != nil {
			return fmt.Errorf("failed to save card: %w", err)
		}

		if existing == nil {
			stats.CardsAdded++
		} else {
			stats.CardsUpdated++
		}

		if u.options.Verbose && (i+1)%10 == 0 {
			fmt.Printf("  Processed %d/%d cards\n", i+1, len(searchResult.Data))
		}
	}

	stats.SetsUpdated++

	if u.options.Progress != nil {
		u.options.Progress(fmt.Sprintf("Completed set: %s (%d cards)", setCode, len(searchResult.Data)))
	}

	return nil
}

// UpdateSpecificSet updates a specific set by code.
func (u *Updater) UpdateSpecificSet(ctx context.Context, setCode string) (*UpdateStats, error) {
	startTime := time.Now()
	stats := &UpdateStats{
		StartTime: startTime,
	}

	if err := u.updateSet(ctx, setCode, stats); err != nil {
		return stats, err
	}

	stats.EndTime = time.Now()
	stats.Duration = stats.EndTime.Sub(stats.StartTime)

	return stats, nil
}
