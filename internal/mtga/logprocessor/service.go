package logprocessor

import (
	"context"
	"log"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// Service handles processing of MTGA log entries and storing results.
// This service encapsulates all log processing logic to avoid duplication
// between CLI and GUI implementations.
type Service struct {
	storage *storage.Service
}

// NewService creates a new log processor service.
func NewService(storage *storage.Service) *Service {
	return &Service{
		storage: storage,
	}
}

// ProcessResult contains the results of processing log entries.
type ProcessResult struct {
	MatchesStored int
	GamesStored   int
	DecksStored   int
	RanksStored   int
	Errors        []error
}

// ProcessLogEntries processes a batch of log entries and stores all extracted data.
// This is the main entry point for both initial log reads and incremental updates.
func (s *Service) ProcessLogEntries(ctx context.Context, entries []*logreader.LogEntry) (*ProcessResult, error) {
	result := &ProcessResult{
		Errors: []error{},
	}

	// Process arena stats (matches and games)
	if err := s.processArenaStats(ctx, entries, result); err != nil {
		result.Errors = append(result.Errors, err)
	}

	// Process decks
	if err := s.processDecks(ctx, entries, result); err != nil {
		result.Errors = append(result.Errors, err)
	}

	// Process rank updates
	if err := s.processRankUpdates(ctx, entries, result); err != nil {
		result.Errors = append(result.Errors, err)
	}

	return result, nil
}

// processArenaStats parses and stores arena statistics from log entries.
func (s *Service) processArenaStats(ctx context.Context, entries []*logreader.LogEntry, result *ProcessResult) error {
	arenaStats, err := logreader.ParseArenaStats(entries)
	if err != nil {
		log.Printf("Warning: Failed to parse arena stats: %v", err)
		return err
	}

	// Store stats if we have match data
	if arenaStats != nil && (arenaStats.TotalMatches > 0 || arenaStats.TotalGames > 0) {
		if err := s.storage.StoreArenaStats(ctx, arenaStats, entries); err != nil {
			log.Printf("Warning: Failed to store arena stats: %v", err)
			return err
		}

		result.MatchesStored = arenaStats.TotalMatches
		result.GamesStored = arenaStats.TotalGames

		log.Printf("✓ Stored statistics: %d matches, %d games", arenaStats.TotalMatches, arenaStats.TotalGames)

		// Try to infer deck IDs for the new matches
		inferredCount, err := s.storage.InferDeckIDsForMatches(ctx)
		if err != nil {
			log.Printf("Warning: Failed to infer deck IDs: %v", err)
		} else if inferredCount > 0 {
			log.Printf("✓ Linked %d match(es) to decks", inferredCount)
		}
	}

	return nil
}

// processDecks parses and stores decks from log entries.
func (s *Service) processDecks(ctx context.Context, entries []*logreader.LogEntry, result *ProcessResult) error {
	deckLibrary, err := logreader.ParseDecks(entries)
	if err != nil {
		log.Printf("Warning: Failed to parse decks: %v", err)
		return err
	}

	if deckLibrary == nil || len(deckLibrary.Decks) == 0 {
		return nil
	}

	log.Printf("Found %d deck(s) in entries", len(deckLibrary.Decks))

	storedCount := 0
	processedCount := 0

	for _, deck := range deckLibrary.Decks {
		// Small delay between deck operations to avoid database lock contention
		if processedCount > 0 {
			time.Sleep(50 * time.Millisecond)
		}
		processedCount++

		// Convert card slices to storage format
		mainDeck := make([]struct {
			CardID   int
			Quantity int
		}, len(deck.MainDeck))
		for i, card := range deck.MainDeck {
			mainDeck[i].CardID = card.CardID
			mainDeck[i].Quantity = card.Quantity
		}

		sideboard := make([]struct {
			CardID   int
			Quantity int
		}, len(deck.Sideboard))
		for i, card := range deck.Sideboard {
			sideboard[i].CardID = card.CardID
			sideboard[i].Quantity = card.Quantity
		}

		// Handle timestamps
		created := deck.Created
		if created.IsZero() && !deck.Modified.IsZero() {
			created = deck.Modified
		} else if created.IsZero() {
			created = time.Now()
		}

		modified := deck.Modified
		if modified.IsZero() {
			modified = time.Now()
		}

		err := s.storage.StoreDeckFromParser(
			ctx,
			deck.DeckID,
			deck.Name,
			deck.Format,
			deck.Description,
			created,
			modified,
			deck.LastPlayed,
			mainDeck,
			sideboard,
		)
		if err != nil {
			log.Printf("Warning: Failed to store deck %s: %v", deck.Name, err)
		} else {
			storedCount++
		}
	}

	if storedCount > 0 {
		result.DecksStored = storedCount
		log.Printf("✓ Stored %d/%d deck(s)", storedCount, len(deckLibrary.Decks))

		// Infer deck IDs for matches after storing decks
		inferredCount, err := s.storage.InferDeckIDsForMatches(ctx)
		if err != nil {
			log.Printf("Warning: Failed to infer deck IDs: %v", err)
		} else if inferredCount > 0 {
			log.Printf("✓ Linked %d match(es) to decks", inferredCount)
		}
	}

	return nil
}

// processRankUpdates parses and stores rank updates from log entries.
func (s *Service) processRankUpdates(ctx context.Context, entries []*logreader.LogEntry, result *ProcessResult) error {
	rankUpdates, err := logreader.ParseRankUpdates(entries)
	if err != nil {
		log.Printf("Warning: Failed to parse rank updates: %v", err)
		return err
	}

	if len(rankUpdates) == 0 {
		return nil
	}

	log.Printf("Found %d rank update(s) in entries", len(rankUpdates))

	storedCount := 0
	for _, update := range rankUpdates {
		// Small delay between operations to avoid database lock contention
		if storedCount > 0 {
			time.Sleep(25 * time.Millisecond)
		}

		if err := s.storage.StoreRankUpdate(ctx, update); err != nil {
			log.Printf("Warning: Failed to store rank update: %v", err)
		} else {
			storedCount++
		}
	}

	if storedCount > 0 {
		result.RanksStored = storedCount
		log.Printf("✓ Stored %d/%d rank update(s)", storedCount, len(rankUpdates))
	}

	return nil
}
