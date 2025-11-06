package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// Service provides high-level operations for storing and retrieving MTGA data.
type Service struct {
	db         *DB
	matches    repository.MatchRepository
	stats      repository.StatsRepository
	decks      repository.DeckRepository
	collection repository.CollectionRepository
}

// NewService creates a new storage service.
func NewService(db *DB) *Service {
	return &Service{
		db:         db,
		matches:    repository.NewMatchRepository(db.Conn()),
		stats:      repository.NewStatsRepository(db.Conn()),
		decks:      repository.NewDeckRepository(db.Conn()),
		collection: repository.NewCollectionRepository(db.Conn()),
	}
}

// StoreArenaStats stores arena statistics parsed from the log.
// It creates match and game records and updates daily stats.
func (s *Service) StoreArenaStats(ctx context.Context, arenaStats *logreader.ArenaStats) error {
	if arenaStats == nil {
		return nil
	}

	// Note: This is a simplified example. In a real implementation, you would need:
	// - Match IDs from the log entries
	// - Timestamps for each match
	// - Event names and formats
	// - Deck IDs if available
	//
	// For now, this demonstrates the pattern of using repositories.

	// Example: Store daily stats for each format
	today := time.Now().Truncate(24 * time.Hour)
	now := time.Now()

	for eventName, formatStat := range arenaStats.FormatStats {
		stats := &PlayerStats{
			Date:          today,
			Format:        eventName,
			MatchesPlayed: formatStat.MatchesPlayed,
			MatchesWon:    formatStat.MatchWins,
			GamesPlayed:   formatStat.GamesPlayed,
			GamesWon:      formatStat.GameWins,
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := s.stats.Upsert(ctx, stats); err != nil {
			return fmt.Errorf("failed to store stats for %s: %w", eventName, err)
		}
	}

	return nil
}

// StoreMatch stores a single match and its games.
// This is useful when processing match completion events from the log.
func (s *Service) StoreMatch(ctx context.Context, match *Match, games []*Game) error {
	// Use a transaction to ensure atomicity
	return s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Create the match
		if err := s.matches.Create(ctx, match); err != nil {
			return fmt.Errorf("failed to create match: %w", err)
		}

		// Create each game
		for _, game := range games {
			if err := s.matches.CreateGame(ctx, game); err != nil {
				return fmt.Errorf("failed to create game: %w", err)
			}
		}

		return nil
	})
}

// GetStats retrieves statistics with optional filtering.
func (s *Service) GetStats(ctx context.Context, filter StatsFilter) (*Statistics, error) {
	return s.matches.GetStats(ctx, filter)
}

// GetRecentMatches retrieves matches within a date range.
func (s *Service) GetRecentMatches(ctx context.Context, days int) ([]*Match, error) {
	end := time.Now()
	start := end.Add(-time.Duration(days) * 24 * time.Hour)
	return s.matches.GetByDateRange(ctx, start, end)
}

// StoreDeck stores a complete deck with its cards.
func (s *Service) StoreDeck(ctx context.Context, deck *Deck, cards []*DeckCard) error {
	return s.db.WithTransaction(ctx, func(tx *sql.Tx) error {
		// Create or update the deck
		existing, err := s.decks.GetByID(ctx, deck.ID)
		if err != nil {
			return fmt.Errorf("failed to check existing deck: %w", err)
		}

		if existing == nil {
			if err := s.decks.Create(ctx, deck); err != nil {
				return fmt.Errorf("failed to create deck: %w", err)
			}
		} else {
			if err := s.decks.Update(ctx, deck); err != nil {
				return fmt.Errorf("failed to update deck: %w", err)
			}
		}

		// Clear existing cards and add new ones
		if err := s.decks.ClearCards(ctx, deck.ID); err != nil {
			return fmt.Errorf("failed to clear deck cards: %w", err)
		}

		for _, card := range cards {
			if err := s.decks.AddCard(ctx, card); err != nil {
				return fmt.Errorf("failed to add card to deck: %w", err)
			}
		}

		return nil
	})
}

// UpdateCollection updates the collection based on changes detected in the log.
// This would typically be called when processing inventory updates.
func (s *Service) UpdateCollection(ctx context.Context, cardID int, newQuantity int, source string) error {
	// Get current quantity
	currentQty, err := s.collection.GetCard(ctx, cardID)
	if err != nil {
		return fmt.Errorf("failed to get current quantity: %w", err)
	}

	// Calculate delta
	delta := newQuantity - currentQty

	if delta != 0 {
		// Record the change
		if err := s.collection.RecordChange(ctx, cardID, delta, time.Now(), &source); err != nil {
			return fmt.Errorf("failed to record collection change: %w", err)
		}
	}

	return nil
}

// GetCollection retrieves the entire collection.
func (s *Service) GetCollection(ctx context.Context) (map[int]int, error) {
	return s.collection.GetAll(ctx)
}

// GetRecentCollectionChanges retrieves recent changes to the collection.
func (s *Service) GetRecentCollectionChanges(ctx context.Context, limit int) ([]*CollectionHistory, error) {
	return s.collection.GetRecentChanges(ctx, limit)
}

// Close closes the database connection.
func (s *Service) Close() error {
	return s.db.Close()
}
