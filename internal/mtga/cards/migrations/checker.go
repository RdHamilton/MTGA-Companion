package migrations

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// Checker manages checking and applying Scryfall card migrations.
type Checker struct {
	scryfall        *scryfall.Client
	storage         *storage.Service
	checkInterval   time.Duration
	stopCh          chan struct{}
	running         bool
	verboseProgress func(string)
}

// CheckerOptions configures the migration checker.
type CheckerOptions struct {
	CheckInterval   time.Duration // How often to check for migrations
	VerboseProgress func(string)  // Optional progress callback
}

// DefaultCheckerOptions returns sensible default options.
func DefaultCheckerOptions() CheckerOptions {
	return CheckerOptions{
		CheckInterval: 24 * time.Hour, // Check daily
	}
}

// NewChecker creates a new migration checker.
func NewChecker(scryfall *scryfall.Client, storage *storage.Service, options CheckerOptions) *Checker {
	if options.CheckInterval == 0 {
		options.CheckInterval = 24 * time.Hour
	}

	return &Checker{
		scryfall:        scryfall,
		storage:         storage,
		checkInterval:   options.CheckInterval,
		stopCh:          make(chan struct{}),
		verboseProgress: options.VerboseProgress,
	}
}

// CheckResult represents the result of checking for migrations.
type CheckResult struct {
	TotalMigrations int
	NewMigrations   int
	MergeCount      int
	DeleteCount     int
	CardsUpdated    int
	CardsDeleted    int
	Errors          []error
}

// CheckAndApply checks for new migrations and applies them.
func (c *Checker) CheckAndApply(ctx context.Context) (*CheckResult, error) {
	c.progress("Checking for Scryfall migrations...")

	// Fetch migrations from Scryfall
	migrationList, err := c.scryfall.GetMigrations(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch migrations: %w", err)
	}

	result := &CheckResult{
		TotalMigrations: len(migrationList.Data),
	}

	c.progress(fmt.Sprintf("Found %d total migrations from Scryfall", result.TotalMigrations))

	// Get already processed migration IDs
	processedIDs, err := c.storage.GetProcessedMigrationIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get processed migrations: %w", err)
	}

	processedMap := make(map[string]bool)
	for _, id := range processedIDs {
		processedMap[id] = true
	}

	// Process each migration
	for _, migration := range migrationList.Data {
		// Skip if already processed
		if processedMap[migration.ID] {
			continue
		}

		result.NewMigrations++

		// Process based on strategy
		switch migration.MigrationStrategy {
		case "merge":
			result.MergeCount++
			if err := c.processMerge(ctx, &migration); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to process merge migration %s: %w", migration.ID, err))
				continue
			}
			result.CardsUpdated++

		case "delete":
			result.DeleteCount++
			if err := c.processDelete(ctx, &migration); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("failed to process delete migration %s: %w", migration.ID, err))
				continue
			}
			result.CardsDeleted++

		default:
			result.Errors = append(result.Errors, fmt.Errorf("unknown migration strategy: %s", migration.MigrationStrategy))
			continue
		}

		// Log the migration as processed
		if err := c.storage.LogMigration(ctx, &migration); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to log migration %s: %w", migration.ID, err))
		}
	}

	c.progress(fmt.Sprintf("Processed %d new migrations (%d merges, %d deletes)", result.NewMigrations, result.MergeCount, result.DeleteCount))

	if len(result.Errors) > 0 {
		c.progress(fmt.Sprintf("Encountered %d errors during processing", len(result.Errors)))
	}

	return result, nil
}

// processMerge handles a merge migration by updating the Scryfall ID.
func (c *Checker) processMerge(ctx context.Context, migration *scryfall.Migration) error {
	if migration.NewScryfallID == nil {
		return fmt.Errorf("merge migration missing new Scryfall ID")
	}

	c.progress(fmt.Sprintf("Merging card: %s -> %s", migration.OldScryfallID, *migration.NewScryfallID))

	// Update card's Scryfall ID
	return c.storage.UpdateCardScryfallID(ctx, migration.OldScryfallID, *migration.NewScryfallID)
}

// processDelete handles a delete migration by removing the card.
func (c *Checker) processDelete(ctx context.Context, migration *scryfall.Migration) error {
	c.progress(fmt.Sprintf("Deleting card: %s", migration.OldScryfallID))

	// Delete card from database
	return c.storage.DeleteCardByScryfallID(ctx, migration.OldScryfallID)
}

// Start begins the automatic migration checker on the configured interval.
func (c *Checker) Start(ctx context.Context) error {
	if c.running {
		return fmt.Errorf("migration checker already running")
	}

	c.running = true
	c.progress(fmt.Sprintf("Starting migration checker (interval: %v)", c.checkInterval))

	// Run initial check immediately
	go func() {
		if _, err := c.CheckAndApply(ctx); err != nil {
			log.Printf("Error during initial migration check: %v", err)
		}

		// Set up ticker for periodic checks
		ticker := time.NewTicker(c.checkInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if _, err := c.CheckAndApply(ctx); err != nil {
					log.Printf("Error during migration check: %v", err)
				}

			case <-c.stopCh:
				c.progress("Migration checker stopped")
				return
			}
		}
	}()

	return nil
}

// Stop stops the automatic migration checker.
func (c *Checker) Stop() {
	if !c.running {
		return
	}

	close(c.stopCh)
	c.running = false
}

// IsRunning returns whether the checker is currently running.
func (c *Checker) IsRunning() bool {
	return c.running
}

// progress logs progress messages if a callback is configured.
func (c *Checker) progress(message string) {
	if c.verboseProgress != nil {
		c.verboseProgress(message)
	}
}
