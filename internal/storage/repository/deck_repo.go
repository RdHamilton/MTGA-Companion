package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// DeckFilter provides filtering options for deck queries.
type DeckFilter struct {
	AccountID int      // Required: filter by account
	Format    *string  // Optional: filter by format (Standard, Historic, etc.)
	Source    *string  // Optional: filter by source (draft, constructed, imported)
	Tags      []string // Optional: filter by tags (must have ALL specified tags)
	SortBy    string   // Sort field: "modified", "created", "name", "performance"
	SortDesc  bool     // Sort descending (default: true for dates, false for names)
}

// DeckRepository handles database operations for decks.
type DeckRepository interface {
	// Create inserts a new deck into the database.
	Create(ctx context.Context, deck *models.Deck) error

	// Update updates an existing deck.
	Update(ctx context.Context, deck *models.Deck) error

	// GetByID retrieves a deck by its ID.
	GetByID(ctx context.Context, id string) (*models.Deck, error)

	// List retrieves all decks for an account.
	List(ctx context.Context, accountID int) ([]*models.Deck, error)

	// GetByFormat retrieves all decks for a specific format and account.
	GetByFormat(ctx context.Context, accountID int, format string) ([]*models.Deck, error)

	// GetBySource retrieves all decks for a specific source (draft/constructed/imported).
	GetBySource(ctx context.Context, accountID int, source string) ([]*models.Deck, error)

	// GetByDraftEvent retrieves the deck associated with a draft event.
	GetByDraftEvent(ctx context.Context, draftEventID string) (*models.Deck, error)

	// Delete deletes a deck by its ID.
	Delete(ctx context.Context, id string) error

	// DeleteBySourceExcluding deletes all decks with the specified source that are NOT in the exclusion list.
	// Returns the number of decks deleted.
	DeleteBySourceExcluding(ctx context.Context, accountID int, source string, excludeIDs []string) (int, error)

	// UpdatePerformance updates deck performance metrics after a match.
	UpdatePerformance(ctx context.Context, deckID string, matchWon bool, gamesWon, gamesLost int) error

	// GetPerformance calculates and returns deck performance metrics.
	GetPerformance(ctx context.Context, deckID string) (*models.DeckPerformance, error)

	// AddCard adds a card to a deck.
	AddCard(ctx context.Context, card *models.DeckCard) error

	// GetCards retrieves all cards in a deck.
	GetCards(ctx context.Context, deckID string) ([]*models.DeckCard, error)

	// RemoveCard decrements the quantity of a card in a deck by 1.
	RemoveCard(ctx context.Context, deckID string, cardID int, board string) error

	// RemoveAllCopies removes all copies of a card from a deck.
	RemoveAllCopies(ctx context.Context, deckID string, cardID int, board string) error

	// ClearCards removes all cards from a deck.
	ClearCards(ctx context.Context, deckID string) error

	// GetDraftCards retrieves all cards picked during a draft event.
	GetDraftCards(ctx context.Context, draftEventID string) ([]int, error)

	// ValidateDraftDeck validates that all cards in a deck are from the associated draft.
	ValidateDraftDeck(ctx context.Context, deckID string) (bool, error)

	// AddTag adds a tag to a deck.
	AddTag(ctx context.Context, tag *models.DeckTag) error

	// GetTags retrieves all tags for a deck.
	GetTags(ctx context.Context, deckID string) ([]*models.DeckTag, error)

	// RemoveTag removes a tag from a deck.
	RemoveTag(ctx context.Context, deckID string, tag string) error

	// Clone creates a copy of a deck with a new ID and name.
	Clone(ctx context.Context, deckID, newName string) (*models.Deck, error)

	// GetByTags retrieves all decks that have ALL specified tags.
	GetByTags(ctx context.Context, accountID int, tags []string) ([]*models.Deck, error)

	// GetByFilters retrieves decks matching multiple filter criteria.
	GetByFilters(ctx context.Context, filter *DeckFilter) ([]*models.Deck, error)

	// GetCardCountsByAccount returns aggregated card counts across all decks for an account.
	// Returns a map of card ID to total quantity across all decks.
	GetCardCountsByAccount(ctx context.Context, accountID int) (map[int]int, error)
}

// deckRepository is the concrete implementation of DeckRepository.
type deckRepository struct {
	db *sql.DB
}

// NewDeckRepository creates a new deck repository.
func NewDeckRepository(db *sql.DB) DeckRepository {
	return &deckRepository{db: db}
}

// Create inserts a new deck into the database.
func (r *deckRepository) Create(ctx context.Context, deck *models.Deck) error {
	query := `
		INSERT INTO decks (
			id, account_id, name, format, description, color_identity, source, draft_event_id,
			matches_played, matches_won, games_played, games_won,
			created_at, modified_at, last_played
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Format timestamps using UTC ISO 8601 without timezone suffixes (SQLite best practice)
	createdAtStr := deck.CreatedAt.UTC().Format("2006-01-02 15:04:05.999999")
	modifiedAtStr := deck.ModifiedAt.UTC().Format("2006-01-02 15:04:05.999999")

	var lastPlayedStr *string
	if deck.LastPlayed != nil {
		formatted := deck.LastPlayed.UTC().Format("2006-01-02 15:04:05.999999")
		lastPlayedStr = &formatted
	}

	_, err := r.db.ExecContext(ctx, query,
		deck.ID,
		deck.AccountID,
		deck.Name,
		deck.Format,
		deck.Description,
		deck.ColorIdentity,
		deck.Source,
		deck.DraftEventID,
		deck.MatchesPlayed,
		deck.MatchesWon,
		deck.GamesPlayed,
		deck.GamesWon,
		createdAtStr,
		modifiedAtStr,
		lastPlayedStr,
	)
	if err != nil {
		return fmt.Errorf("failed to create deck: %w", err)
	}

	return nil
}

// Update updates an existing deck.
func (r *deckRepository) Update(ctx context.Context, deck *models.Deck) error {
	query := `
		UPDATE decks
		SET name = ?, format = ?, description = ?, color_identity = ?, source = ?,
		    modified_at = ?, last_played = ?
		WHERE id = ?
	`

	modifiedAtStr := deck.ModifiedAt.UTC().Format("2006-01-02 15:04:05.999999")

	var lastPlayedStr *string
	if deck.LastPlayed != nil {
		formatted := deck.LastPlayed.UTC().Format("2006-01-02 15:04:05.999999")
		lastPlayedStr = &formatted
	}

	_, err := r.db.ExecContext(ctx, query,
		deck.Name,
		deck.Format,
		deck.Description,
		deck.ColorIdentity,
		deck.Source,
		modifiedAtStr,
		lastPlayedStr,
		deck.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update deck: %w", err)
	}

	return nil
}

// GetByID retrieves a deck by its ID.
func (r *deckRepository) GetByID(ctx context.Context, id string) (*models.Deck, error) {
	query := `
		SELECT id, account_id, name, format, description, color_identity, source, draft_event_id,
		       matches_played, matches_won, games_played, games_won,
		       created_at, modified_at, last_played
		FROM decks
		WHERE id = ?
	`

	deck := &models.Deck{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&deck.ID,
		&deck.AccountID,
		&deck.Name,
		&deck.Format,
		&deck.Description,
		&deck.ColorIdentity,
		&deck.Source,
		&deck.DraftEventID,
		&deck.MatchesPlayed,
		&deck.MatchesWon,
		&deck.GamesPlayed,
		&deck.GamesWon,
		&deck.CreatedAt,
		&deck.ModifiedAt,
		&deck.LastPlayed,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get deck by id: %w", err)
	}

	return deck, nil
}

// List retrieves all decks for an account.
func (r *deckRepository) List(ctx context.Context, accountID int) ([]*models.Deck, error) {
	query := `
		SELECT id, account_id, name, format, description, color_identity, source, draft_event_id,
		       matches_played, matches_won, games_played, games_won,
		       created_at, modified_at, last_played
		FROM decks
		WHERE account_id = ?
		ORDER BY modified_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to list decks: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanDecks(rows)
}

// GetByFormat retrieves all decks for a specific format and account.
func (r *deckRepository) GetByFormat(ctx context.Context, accountID int, format string) ([]*models.Deck, error) {
	query := `
		SELECT id, account_id, name, format, description, color_identity, source, draft_event_id,
		       matches_played, matches_won, games_played, games_won,
		       created_at, modified_at, last_played
		FROM decks
		WHERE account_id = ? AND format = ?
		ORDER BY modified_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID, format)
	if err != nil {
		return nil, fmt.Errorf("failed to get decks by format: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanDecks(rows)
}

// Delete deletes a deck by its ID.
func (r *deckRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM decks WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete deck: %w", err)
	}

	return nil
}

// DeleteBySourceExcluding deletes all decks with the specified source that are NOT in the exclusion list.
// This is used to clean up stale arena decks that are no longer present in MTGA logs.
func (r *deckRepository) DeleteBySourceExcluding(ctx context.Context, accountID int, source string, excludeIDs []string) (int, error) {
	if len(excludeIDs) == 0 {
		// No exclusions - delete all decks with this source
		query := `DELETE FROM decks WHERE account_id = ? AND source = ?`
		result, err := r.db.ExecContext(ctx, query, accountID, source)
		if err != nil {
			return 0, fmt.Errorf("failed to delete decks by source: %w", err)
		}
		affected, _ := result.RowsAffected()
		return int(affected), nil
	}

	// Build query with exclusion list using placeholders
	// SQLite doesn't support arrays, so we build a parameterized IN clause
	placeholders := make([]string, len(excludeIDs))
	args := make([]interface{}, 0, len(excludeIDs)+2)
	args = append(args, accountID, source)

	for i, id := range excludeIDs {
		placeholders[i] = "?"
		args = append(args, id)
	}

	query := fmt.Sprintf(
		`DELETE FROM decks WHERE account_id = ? AND source = ? AND id NOT IN (%s)`,
		strings.Join(placeholders, ", "),
	)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to delete decks by source excluding: %w", err)
	}

	affected, _ := result.RowsAffected()
	return int(affected), nil
}

// AddCard adds a card to a deck. If the card already exists, increments the quantity.
func (r *deckRepository) AddCard(ctx context.Context, card *models.DeckCard) error {
	query := `
		INSERT INTO deck_cards (deck_id, card_id, quantity, board, from_draft_pick)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(deck_id, card_id, board) DO UPDATE SET
			quantity = quantity + excluded.quantity,
			from_draft_pick = excluded.from_draft_pick
	`

	fromDraftPickInt := 0
	if card.FromDraftPick {
		fromDraftPickInt = 1
	}

	result, err := r.db.ExecContext(ctx, query,
		card.DeckID,
		card.CardID,
		card.Quantity,
		card.Board,
		fromDraftPickInt,
	)
	if err != nil {
		return fmt.Errorf("failed to add card to deck: %w", err)
	}

	// If this is an insert, set the ID
	if card.ID == 0 {
		id, err := result.LastInsertId()
		if err == nil {
			card.ID = int(id)
		}
	}

	return nil
}

// GetCards retrieves all cards in a deck.
func (r *deckRepository) GetCards(ctx context.Context, deckID string) ([]*models.DeckCard, error) {
	query := `
		SELECT id, deck_id, card_id, quantity, board, from_draft_pick
		FROM deck_cards
		WHERE deck_id = ?
		ORDER BY board, card_id
	`

	rows, err := r.db.QueryContext(ctx, query, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deck cards: %w", err)
	}
	defer func() {
		//nolint:errcheck // Ignore error on cleanup - this is a defer cleanup operation
		_ = rows.Close()
	}()

	var cards []*models.DeckCard
	for rows.Next() {
		card := &models.DeckCard{}
		var fromDraftPickInt int
		err := rows.Scan(
			&card.ID,
			&card.DeckID,
			&card.CardID,
			&card.Quantity,
			&card.Board,
			&fromDraftPickInt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deck card: %w", err)
		}
		card.FromDraftPick = fromDraftPickInt == 1
		cards = append(cards, card)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deck cards: %w", err)
	}

	return cards, nil
}

// RemoveCard decrements the quantity of a card in a deck by 1.
// If the quantity reaches 0, the card is removed from the deck.
func (r *deckRepository) RemoveCard(ctx context.Context, deckID string, cardID int, board string) error {
	// First, decrement the quantity
	updateQuery := `
		UPDATE deck_cards
		SET quantity = quantity - 1
		WHERE deck_id = ? AND card_id = ? AND board = ? AND quantity > 0
	`

	result, err := r.db.ExecContext(ctx, updateQuery, deckID, cardID, board)
	if err != nil {
		return fmt.Errorf("failed to decrement card quantity: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("card not found in deck or quantity already 0")
	}

	// Delete any rows where quantity is now 0
	deleteQuery := `DELETE FROM deck_cards WHERE deck_id = ? AND card_id = ? AND board = ? AND quantity <= 0`
	_, err = r.db.ExecContext(ctx, deleteQuery, deckID, cardID, board)
	if err != nil {
		return fmt.Errorf("failed to clean up zero quantity card: %w", err)
	}

	return nil
}

// RemoveAllCopies removes all copies of a card from a deck.
func (r *deckRepository) RemoveAllCopies(ctx context.Context, deckID string, cardID int, board string) error {
	query := `DELETE FROM deck_cards WHERE deck_id = ? AND card_id = ? AND board = ?`

	_, err := r.db.ExecContext(ctx, query, deckID, cardID, board)
	if err != nil {
		return fmt.Errorf("failed to remove card from deck: %w", err)
	}

	return nil
}

// ClearCards removes all cards from a deck.
func (r *deckRepository) ClearCards(ctx context.Context, deckID string) error {
	query := `DELETE FROM deck_cards WHERE deck_id = ?`

	_, err := r.db.ExecContext(ctx, query, deckID)
	if err != nil {
		return fmt.Errorf("failed to clear deck cards: %w", err)
	}

	return nil
}

// GetBySource retrieves all decks for a specific source (draft/constructed/imported).
func (r *deckRepository) GetBySource(ctx context.Context, accountID int, source string) ([]*models.Deck, error) {
	query := `
		SELECT id, account_id, name, format, description, color_identity, source, draft_event_id,
		       matches_played, matches_won, games_played, games_won,
		       created_at, modified_at, last_played
		FROM decks
		WHERE account_id = ? AND source = ?
		ORDER BY modified_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID, source)
	if err != nil {
		return nil, fmt.Errorf("failed to get decks by source: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanDecks(rows)
}

// GetByDraftEvent retrieves the deck associated with a draft event.
func (r *deckRepository) GetByDraftEvent(ctx context.Context, draftEventID string) (*models.Deck, error) {
	query := `
		SELECT id, account_id, name, format, description, color_identity, source, draft_event_id,
		       matches_played, matches_won, games_played, games_won,
		       created_at, modified_at, last_played
		FROM decks
		WHERE draft_event_id = ?
	`

	deck := &models.Deck{}
	err := r.db.QueryRowContext(ctx, query, draftEventID).Scan(
		&deck.ID,
		&deck.AccountID,
		&deck.Name,
		&deck.Format,
		&deck.Description,
		&deck.ColorIdentity,
		&deck.Source,
		&deck.DraftEventID,
		&deck.MatchesPlayed,
		&deck.MatchesWon,
		&deck.GamesPlayed,
		&deck.GamesWon,
		&deck.CreatedAt,
		&deck.ModifiedAt,
		&deck.LastPlayed,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get deck by draft event: %w", err)
	}

	return deck, nil
}

// UpdatePerformance updates deck performance metrics after a match.
func (r *deckRepository) UpdatePerformance(ctx context.Context, deckID string, matchWon bool, gamesWon, gamesLost int) error {
	query := `
		UPDATE decks
		SET matches_played = matches_played + 1,
		    matches_won = matches_won + ?,
		    games_played = games_played + ?,
		    games_won = games_won + ?,
		    last_played = ?,
		    modified_at = ?
		WHERE id = ?
	`

	matchWonInt := 0
	if matchWon {
		matchWonInt = 1
	}

	now := time.Now().UTC().Format("2006-01-02 15:04:05.999999")
	totalGames := gamesWon + gamesLost

	_, err := r.db.ExecContext(ctx, query,
		matchWonInt,
		totalGames,
		gamesWon,
		now,
		now,
		deckID,
	)
	if err != nil {
		return fmt.Errorf("failed to update deck performance: %w", err)
	}

	return nil
}

// GetPerformance calculates and returns deck performance metrics.
func (r *deckRepository) GetPerformance(ctx context.Context, deckID string) (*models.DeckPerformance, error) {
	// Get basic performance data from deck
	deckQuery := `
		SELECT matches_played, matches_won, games_played, games_won, last_played
		FROM decks
		WHERE id = ?
	`

	perf := &models.DeckPerformance{DeckID: deckID}

	err := r.db.QueryRowContext(ctx, deckQuery, deckID).Scan(
		&perf.MatchesPlayed,
		&perf.MatchesWon,
		&perf.GamesPlayed,
		&perf.GamesWon,
		&perf.LastPlayed,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get deck performance: %w", err)
	}

	// Calculate derived metrics
	perf.MatchesLost = perf.MatchesPlayed - perf.MatchesWon
	perf.GamesLost = perf.GamesPlayed - perf.GamesWon

	if perf.MatchesPlayed > 0 {
		perf.MatchWinRate = float64(perf.MatchesWon) / float64(perf.MatchesPlayed)
	}
	if perf.GamesPlayed > 0 {
		perf.GameWinRate = float64(perf.GamesWon) / float64(perf.GamesPlayed)
	}

	// Calculate streaks and average duration from matches table
	streakQuery := `
		SELECT result, duration_seconds
		FROM matches
		WHERE deck_id = ?
		ORDER BY timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, streakQuery, deckID)
	if err != nil {
		// Don't fail the whole request if streak query fails, just return without streak data
		return perf, nil
	}
	defer func() {
		_ = rows.Close()
	}()

	var (
		results           []string
		longestWinStreak  int
		longestLossStreak int
		totalDuration     int64
		durationCount     int
	)

	for rows.Next() {
		var result string
		var duration *int
		if err := rows.Scan(&result, &duration); err != nil {
			continue
		}
		results = append(results, result)
		if duration != nil && *duration > 0 {
			totalDuration += int64(*duration)
			durationCount++
		}
	}

	if err = rows.Err(); err != nil {
		return perf, nil
	}

	// Calculate current streak (from most recent matches)
	if len(results) > 0 {
		firstResult := results[0]
		currentStreak := 0
		for _, result := range results {
			if result == firstResult {
				if firstResult == "win" {
					currentStreak++
				} else {
					currentStreak--
				}
			} else {
				break
			}
		}
		perf.CurrentWinStreak = currentStreak
	}

	// Calculate longest streaks by iterating through all matches (oldest to newest)
	winStreak := 0
	lossStreak := 0
	for i := len(results) - 1; i >= 0; i-- {
		if results[i] == "win" {
			winStreak++
			lossStreak = 0
			if winStreak > longestWinStreak {
				longestWinStreak = winStreak
			}
		} else {
			lossStreak++
			winStreak = 0
			if lossStreak > longestLossStreak {
				longestLossStreak = lossStreak
			}
		}
	}
	perf.LongestWinStreak = longestWinStreak
	perf.LongestLossStreak = longestLossStreak

	// Calculate average duration
	if durationCount > 0 {
		avgDuration := float64(totalDuration) / float64(durationCount)
		perf.AverageDuration = &avgDuration
	}

	return perf, nil
}

// GetDraftCards retrieves all cards picked during a draft event.
func (r *deckRepository) GetDraftCards(ctx context.Context, draftEventID string) ([]int, error) {
	log.Printf("[GetDraftCards] Looking for picks with session_id=%s", draftEventID)

	query := `
		SELECT DISTINCT CAST(card_id AS INTEGER) as card_id_int
		FROM draft_picks
		WHERE session_id = ?
		ORDER BY card_id_int
	`

	rows, err := r.db.QueryContext(ctx, query, draftEventID)
	if err != nil {
		log.Printf("[GetDraftCards] Query error: %v", err)
		return nil, fmt.Errorf("failed to get draft cards: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var cardIDs []int
	for rows.Next() {
		var cardID int
		if err := rows.Scan(&cardID); err != nil {
			return nil, fmt.Errorf("failed to scan card ID: %w", err)
		}
		cardIDs = append(cardIDs, cardID)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating draft cards: %w", err)
	}

	log.Printf("[GetDraftCards] Found %d cards for session_id=%s", len(cardIDs), draftEventID)

	// If no cards found, log the distinct session IDs in the database for debugging
	if len(cardIDs) == 0 {
		var distinctSessions []string
		sessionQuery := `SELECT DISTINCT session_id FROM draft_picks ORDER BY session_id LIMIT 10`
		sessionRows, err := r.db.QueryContext(ctx, sessionQuery)
		if err == nil {
			defer func() { _ = sessionRows.Close() }()
			for sessionRows.Next() {
				var sid string
				if sessionRows.Scan(&sid) == nil {
					distinctSessions = append(distinctSessions, sid)
				}
			}
		}
		log.Printf("[GetDraftCards] No cards found. Existing session_ids in draft_picks: %v", distinctSessions)
	}

	return cardIDs, nil
}

// ValidateDraftDeck validates that all cards in a deck are from the associated draft.
func (r *deckRepository) ValidateDraftDeck(ctx context.Context, deckID string) (bool, error) {
	// Get the deck and check if it's a draft deck
	deck, err := r.GetByID(ctx, deckID)
	if err != nil {
		return false, fmt.Errorf("failed to get deck: %w", err)
	}
	if deck == nil {
		return false, fmt.Errorf("deck not found")
	}

	// Only validate draft decks
	if deck.Source != "draft" || deck.DraftEventID == nil {
		return true, nil // Non-draft decks are always valid
	}

	// Get all cards in the draft
	draftCardIDs, err := r.GetDraftCards(ctx, *deck.DraftEventID)
	if err != nil {
		return false, fmt.Errorf("failed to get draft cards: %w", err)
	}

	// Create a set of draft card IDs for O(1) lookup
	draftCardSet := make(map[int]bool)
	for _, cardID := range draftCardIDs {
		draftCardSet[cardID] = true
	}

	// Get all cards in the deck
	deckCards, err := r.GetCards(ctx, deckID)
	if err != nil {
		return false, fmt.Errorf("failed to get deck cards: %w", err)
	}

	// Basic lands that can be added to any deck (standard Arena IDs)
	basicLands := map[int]bool{
		81716: true, // Plains
		81717: true, // Island
		81718: true, // Swamp
		81719: true, // Mountain
		81720: true, // Forest
	}

	// Validate that all non-basic-land deck cards are in the draft
	for _, card := range deckCards {
		// Skip basic lands - they can always be added
		if basicLands[card.CardID] {
			continue
		}
		if !draftCardSet[card.CardID] {
			return false, nil // Found a card not in the draft
		}
	}

	return true, nil
}

// AddTag adds a tag to a deck.
func (r *deckRepository) AddTag(ctx context.Context, tag *models.DeckTag) error {
	query := `
		INSERT INTO deck_tags (deck_id, tag, created_at)
		VALUES (?, ?, ?)
		ON CONFLICT(deck_id, tag) DO NOTHING
	`

	createdAtStr := tag.CreatedAt.UTC().Format("2006-01-02 15:04:05.999999")

	result, err := r.db.ExecContext(ctx, query, tag.DeckID, tag.Tag, createdAtStr)
	if err != nil {
		return fmt.Errorf("failed to add deck tag: %w", err)
	}

	if tag.ID == 0 {
		id, err := result.LastInsertId()
		if err == nil {
			tag.ID = int(id)
		}
	}

	return nil
}

// GetTags retrieves all tags for a deck.
func (r *deckRepository) GetTags(ctx context.Context, deckID string) ([]*models.DeckTag, error) {
	query := `
		SELECT id, deck_id, tag, created_at
		FROM deck_tags
		WHERE deck_id = ?
		ORDER BY tag
	`

	rows, err := r.db.QueryContext(ctx, query, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deck tags: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var tags []*models.DeckTag
	for rows.Next() {
		tag := &models.DeckTag{}
		err := rows.Scan(&tag.ID, &tag.DeckID, &tag.Tag, &tag.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deck tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deck tags: %w", err)
	}

	return tags, nil
}

// RemoveTag removes a tag from a deck.
func (r *deckRepository) RemoveTag(ctx context.Context, deckID string, tag string) error {
	query := `DELETE FROM deck_tags WHERE deck_id = ? AND tag = ?`

	_, err := r.db.ExecContext(ctx, query, deckID, tag)
	if err != nil {
		return fmt.Errorf("failed to remove deck tag: %w", err)
	}

	return nil
}

// Clone creates a copy of a deck with a new ID and name.
func (r *deckRepository) Clone(ctx context.Context, deckID, newName string) (*models.Deck, error) {
	// Get the original deck
	originalDeck, err := r.GetByID(ctx, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original deck: %w", err)
	}
	if originalDeck == nil {
		return nil, fmt.Errorf("original deck not found")
	}

	// Create a new deck with a new ID
	newDeck := &models.Deck{
		ID:            fmt.Sprintf("%s-clone-%d", deckID, time.Now().Unix()),
		AccountID:     originalDeck.AccountID,
		Name:          newName,
		Format:        originalDeck.Format,
		Description:   originalDeck.Description,
		ColorIdentity: originalDeck.ColorIdentity,
		Source:        "constructed", // Clones are always constructed, not draft
		DraftEventID:  nil,           // Don't copy draft event association
		MatchesPlayed: 0,             // Reset performance stats
		MatchesWon:    0,
		GamesPlayed:   0,
		GamesWon:      0,
		CreatedAt:     time.Now(),
		ModifiedAt:    time.Now(),
	}

	// Create the new deck
	if err := r.Create(ctx, newDeck); err != nil {
		return nil, fmt.Errorf("failed to create cloned deck: %w", err)
	}

	// Copy all cards
	originalCards, err := r.GetCards(ctx, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original deck cards: %w", err)
	}

	for _, card := range originalCards {
		newCard := &models.DeckCard{
			DeckID:        newDeck.ID,
			CardID:        card.CardID,
			Quantity:      card.Quantity,
			Board:         card.Board,
			FromDraftPick: false, // Don't copy draft pick flag
		}
		if err := r.AddCard(ctx, newCard); err != nil {
			return nil, fmt.Errorf("failed to copy card to cloned deck: %w", err)
		}
	}

	// Copy all tags
	originalTags, err := r.GetTags(ctx, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get original deck tags: %w", err)
	}

	for _, tag := range originalTags {
		newTag := &models.DeckTag{
			DeckID:    newDeck.ID,
			Tag:       tag.Tag,
			CreatedAt: time.Now(),
		}
		if err := r.AddTag(ctx, newTag); err != nil {
			return nil, fmt.Errorf("failed to copy tag to cloned deck: %w", err)
		}
	}

	return newDeck, nil
}

// GetByTags retrieves all decks that have ALL specified tags.
func (r *deckRepository) GetByTags(ctx context.Context, accountID int, tags []string) ([]*models.Deck, error) {
	if len(tags) == 0 {
		return r.List(ctx, accountID)
	}

	// Build query with JOIN for each tag to ensure deck has ALL tags
	query := `
		SELECT DISTINCT d.id, d.account_id, d.name, d.format, d.description, d.color_identity, d.source, d.draft_event_id,
		       d.matches_played, d.matches_won, d.games_played, d.games_won,
		       d.created_at, d.modified_at, d.last_played
		FROM decks d
	`

	// Add a JOIN for each tag
	for i := range tags {
		query += fmt.Sprintf(" INNER JOIN deck_tags dt%d ON d.id = dt%d.deck_id", i, i)
	}

	query += " WHERE d.account_id = ?"

	// Add conditions for each tag
	for i := range tags {
		query += fmt.Sprintf(" AND dt%d.tag = ?", i)
	}

	query += " ORDER BY d.modified_at DESC"

	// Build args: accountID + all tags
	args := make([]interface{}, 0, len(tags)+1)
	args = append(args, accountID)
	for _, tag := range tags {
		args = append(args, tag)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get decks by tags: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanDecks(rows)
}

// GetByFilters retrieves decks matching multiple filter criteria.
func (r *deckRepository) GetByFilters(ctx context.Context, filter *DeckFilter) ([]*models.Deck, error) {
	if filter == nil {
		return nil, fmt.Errorf("filter cannot be nil")
	}

	// Start with base query
	query := `
		SELECT DISTINCT d.id, d.account_id, d.name, d.format, d.description, d.color_identity, d.source, d.draft_event_id,
		       d.matches_played, d.matches_won, d.games_played, d.games_won,
		       d.created_at, d.modified_at, d.last_played
		FROM decks d
	`

	// Add JOINs for tag filtering
	if len(filter.Tags) > 0 {
		for i := range filter.Tags {
			query += fmt.Sprintf(" INNER JOIN deck_tags dt%d ON d.id = dt%d.deck_id", i, i)
		}
	}

	// WHERE clause
	query += " WHERE d.account_id = ?"
	args := []interface{}{filter.AccountID}

	// Add format filter
	if filter.Format != nil {
		query += " AND d.format = ?"
		args = append(args, *filter.Format)
	}

	// Add source filter
	if filter.Source != nil {
		query += " AND d.source = ?"
		args = append(args, *filter.Source)
	}

	// Add tag conditions
	for i, tag := range filter.Tags {
		query += fmt.Sprintf(" AND dt%d.tag = ?", i)
		args = append(args, tag)
	}

	// Add ORDER BY
	sortField := "d.modified_at"
	if filter.SortBy != "" {
		switch filter.SortBy {
		case "modified":
			sortField = "d.modified_at"
		case "created":
			sortField = "d.created_at"
		case "name":
			sortField = "d.name"
		case "performance":
			sortField = "(CAST(d.matches_won AS REAL) / NULLIF(d.matches_played, 0))"
		}
	}

	sortDir := "DESC"
	if !filter.SortDesc && filter.SortBy != "name" {
		sortDir = "ASC"
	} else if filter.SortBy == "name" && !filter.SortDesc {
		sortDir = "ASC"
	}

	query += fmt.Sprintf(" ORDER BY %s %s", sortField, sortDir)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get decks by filters: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanDecks(rows)
}

// scanDecks is a helper function to scan multiple decks from rows.
func (r *deckRepository) scanDecks(rows *sql.Rows) ([]*models.Deck, error) {
	var decks []*models.Deck
	for rows.Next() {
		deck := &models.Deck{}
		err := rows.Scan(
			&deck.ID,
			&deck.AccountID,
			&deck.Name,
			&deck.Format,
			&deck.Description,
			&deck.ColorIdentity,
			&deck.Source,
			&deck.DraftEventID,
			&deck.MatchesPlayed,
			&deck.MatchesWon,
			&deck.GamesPlayed,
			&deck.GamesWon,
			&deck.CreatedAt,
			&deck.ModifiedAt,
			&deck.LastPlayed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan deck: %w", err)
		}
		decks = append(decks, deck)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating decks: %w", err)
	}

	return decks, nil
}

// GetCardCountsByAccount returns aggregated card counts across all decks for an account.
func (r *deckRepository) GetCardCountsByAccount(ctx context.Context, accountID int) (map[int]int, error) {
	query := `
		SELECT dc.card_id, SUM(dc.quantity) as total_qty
		FROM deck_cards dc
		JOIN decks d ON dc.deck_id = d.id
		WHERE d.account_id = ?
		GROUP BY dc.card_id
	`

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query deck cards: %w", err)
	}
	defer func() { _ = rows.Close() }()

	cardCounts := make(map[int]int)
	for rows.Next() {
		var cardID, totalQty int
		if err := rows.Scan(&cardID, &totalQty); err != nil {
			return nil, fmt.Errorf("failed to scan deck card: %w", err)
		}
		cardCounts[cardID] = totalQty
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deck cards: %w", err)
	}

	return cardCounts, nil
}
