package updater

import (
	"context"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/importer"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// BulkDataUpdater handles automatic updates of card bulk data.
// It checks remote timestamps and delegates actual downloading to BulkImporter.
type BulkDataUpdater struct {
	scryfall        *scryfall.Client
	storage         *storage.Service
	importer        *importer.BulkImporter
	options         importer.BulkImportOptions
	verboseProgress func(string)
}

// NewBulkDataUpdater creates a new bulk data updater.
func NewBulkDataUpdater(scryfallClient *scryfall.Client, storageService *storage.Service, importOptions importer.BulkImportOptions) *BulkDataUpdater {
	bulkImporter := importer.NewBulkImporter(scryfallClient, storageService, importOptions)

	return &BulkDataUpdater{
		scryfall: scryfallClient,
		storage:  storageService,
		importer: bulkImporter,
		options:  importOptions,
	}
}

// SetProgressCallback sets a progress callback for verbose output.
func (u *BulkDataUpdater) SetProgressCallback(callback func(string)) {
	u.verboseProgress = callback
}

// CheckForUpdates checks if bulk data updates are available.
// Returns true if updates are available, along with the update timestamp.
func (u *BulkDataUpdater) CheckForUpdates(ctx context.Context) (bool, time.Time, error) {
	// Get remote bulk data info
	bulkDataList, err := u.scryfall.GetBulkData(ctx)
	if err != nil {
		return false, time.Time{}, fmt.Errorf("failed to get bulk data info: %w", err)
	}

	// Find default_cards bulk data
	var remoteUpdatedAt time.Time
	for _, bulk := range bulkDataList.Data {
		if bulk.Type == "default_cards" {
			remoteUpdatedAt = bulk.UpdatedAt
			break
		}
	}

	if remoteUpdatedAt.IsZero() {
		return false, time.Time{}, fmt.Errorf("default_cards bulk data not found")
	}

	// Get last update time from database
	lastUpdate, err := u.storage.GetLastBulkDataUpdate(ctx)
	if err != nil {
		// If error getting last update, assume updates are available
		return true, remoteUpdatedAt, nil
	}

	// Compare timestamps
	needsUpdate := remoteUpdatedAt.After(lastUpdate)
	return needsUpdate, remoteUpdatedAt, nil
}

// UpdateIfNeeded checks for updates and performs import if needed.
// Returns true if an update was performed.
func (u *BulkDataUpdater) UpdateIfNeeded(ctx context.Context) (bool, error) {
	needsUpdate, updateTime, err := u.CheckForUpdates(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check for updates: %w", err)
	}

	if !needsUpdate {
		return false, nil
	}

	// Perform update
	if err := u.PerformUpdate(ctx); err != nil {
		return false, fmt.Errorf("failed to perform update: %w", err)
	}

	// Record update time
	if err := u.storage.SetLastBulkDataUpdate(ctx, updateTime); err != nil {
		// Log error but don't fail - update was successful
		return true, fmt.Errorf("update successful but failed to record timestamp: %w", err)
	}

	return true, nil
}

// PerformUpdate downloads and imports bulk data using the BulkImporter.
func (u *BulkDataUpdater) PerformUpdate(ctx context.Context) error {
	if u.verboseProgress != nil {
		u.verboseProgress("Starting bulk data update...")
	}

	// Delegate to the bulk importer
	stats, err := u.importer.Import(ctx)
	if err != nil {
		return fmt.Errorf("failed to perform bulk update: %w", err)
	}

	if u.verboseProgress != nil {
		u.verboseProgress(fmt.Sprintf("Bulk update complete: %d cards imported, %d skipped, %d errors",
			stats.ImportedCards, stats.SkippedCards, stats.ErrorCards))
	}

	// Check if any cards were imported
	if stats.ImportedCards == 0 && stats.ErrorCards > 0 {
		return fmt.Errorf("no cards were imported, encountered %d errors", stats.ErrorCards)
	}

	return nil
}
