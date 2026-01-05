package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// DeckPermutationRepository handles database operations for deck version tracking.
type DeckPermutationRepository interface {
	// Create inserts a new permutation for a deck.
	Create(ctx context.Context, perm *models.DeckPermutation) error

	// GetByID retrieves a permutation by its ID.
	GetByID(ctx context.Context, id int) (*models.DeckPermutation, error)

	// GetByDeckID retrieves all permutations for a deck, ordered by version number.
	GetByDeckID(ctx context.Context, deckID string) ([]*models.DeckPermutation, error)

	// GetLatest retrieves the most recent permutation for a deck.
	GetLatest(ctx context.Context, deckID string) (*models.DeckPermutation, error)

	// GetCurrent retrieves the currently active permutation for a deck.
	GetCurrent(ctx context.Context, deckID string) (*models.DeckPermutation, error)

	// GetByCardHash finds an existing permutation by its card hash (for detecting duplicates).
	GetByCardHash(ctx context.Context, deckID, cardHash string) (*models.DeckPermutation, error)

	// SetCurrentPermutation sets which permutation is the active one for a deck.
	SetCurrentPermutation(ctx context.Context, deckID string, permutationID int) error

	// UpdatePerformance updates a permutation's performance stats after a match.
	UpdatePerformance(ctx context.Context, permutationID int, matchWon bool, gamesWon, gamesLost int) error

	// GetPerformance retrieves performance metrics for a permutation.
	GetPerformance(ctx context.Context, permutationID int) (*models.DeckPermutationPerformance, error)

	// GetAllPerformance retrieves performance metrics for all permutations of a deck.
	GetAllPerformance(ctx context.Context, deckID string) ([]*models.DeckPermutationPerformance, error)

	// GetDiff calculates the difference between two permutations.
	GetDiff(ctx context.Context, fromPermID, toPermID int) (*models.DeckPermutationDiff, error)

	// CreateFromCurrentDeck creates a new permutation from a deck's current cards.
	// If a permutation with the same card_hash already exists, returns it instead.
	CreateFromCurrentDeck(ctx context.Context, deckID string, versionName, changeSummary *string) (*models.DeckPermutation, error)

	// Delete removes a permutation (cascading from deck deletion handled by FK).
	Delete(ctx context.Context, id int) error

	// GetNextVersionNumber returns the next version number for a deck.
	GetNextVersionNumber(ctx context.Context, deckID string) (int, error)
}

// deckPermutationRepository is the concrete implementation.
type deckPermutationRepository struct {
	db *sql.DB
}

// NewDeckPermutationRepository creates a new deck permutation repository.
func NewDeckPermutationRepository(db *sql.DB) DeckPermutationRepository {
	return &deckPermutationRepository{db: db}
}

// computeCardHash generates a deterministic hash from cards sorted by card_id and board.
func computeCardHash(cards []models.DeckPermutationCard) string {
	// Sort cards by card_id, then by board
	sorted := make([]models.DeckPermutationCard, len(cards))
	copy(sorted, cards)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].CardID != sorted[j].CardID {
			return sorted[i].CardID < sorted[j].CardID
		}
		return sorted[i].Board < sorted[j].Board
	})

	// Build hash string: card_id:quantity:board|card_id:quantity:board|...
	var parts []string
	for _, card := range sorted {
		parts = append(parts, fmt.Sprintf("%d:%d:%s", card.CardID, card.Quantity, card.Board))
	}
	return strings.Join(parts, "|")
}

// Create inserts a new permutation for a deck.
func (r *deckPermutationRepository) Create(ctx context.Context, perm *models.DeckPermutation) error {
	query := `
		INSERT INTO deck_permutations (
			deck_id, parent_permutation_id, cards, card_hash, version_number, version_name, change_summary,
			matches_played, matches_won, games_played, games_won, created_at, last_played_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	createdAtStr := perm.CreatedAt.UTC().Format("2006-01-02 15:04:05.999999")

	var lastPlayedStr *string
	if perm.LastPlayedAt != nil {
		formatted := perm.LastPlayedAt.UTC().Format("2006-01-02 15:04:05.999999")
		lastPlayedStr = &formatted
	}

	result, err := r.db.ExecContext(ctx, query,
		perm.DeckID,
		perm.ParentPermutationID,
		perm.Cards,
		perm.CardHash,
		perm.VersionNumber,
		perm.VersionName,
		perm.ChangeSummary,
		perm.MatchesPlayed,
		perm.MatchesWon,
		perm.GamesPlayed,
		perm.GamesWon,
		createdAtStr,
		lastPlayedStr,
	)
	if err != nil {
		return fmt.Errorf("failed to create deck permutation: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}
	perm.ID = int(id)

	return nil
}

// GetByID retrieves a permutation by its ID.
func (r *deckPermutationRepository) GetByID(ctx context.Context, id int) (*models.DeckPermutation, error) {
	query := `
		SELECT id, deck_id, parent_permutation_id, cards, card_hash, version_number, version_name, change_summary,
		       matches_played, matches_won, games_played, games_won, created_at, last_played_at
		FROM deck_permutations
		WHERE id = ?
	`

	perm := &models.DeckPermutation{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&perm.ID,
		&perm.DeckID,
		&perm.ParentPermutationID,
		&perm.Cards,
		&perm.CardHash,
		&perm.VersionNumber,
		&perm.VersionName,
		&perm.ChangeSummary,
		&perm.MatchesPlayed,
		&perm.MatchesWon,
		&perm.GamesPlayed,
		&perm.GamesWon,
		&perm.CreatedAt,
		&perm.LastPlayedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get permutation by id: %w", err)
	}

	return perm, nil
}

// GetByDeckID retrieves all permutations for a deck, ordered by version number.
func (r *deckPermutationRepository) GetByDeckID(ctx context.Context, deckID string) ([]*models.DeckPermutation, error) {
	query := `
		SELECT id, deck_id, parent_permutation_id, cards, card_hash, version_number, version_name, change_summary,
		       matches_played, matches_won, games_played, games_won, created_at, last_played_at
		FROM deck_permutations
		WHERE deck_id = ?
		ORDER BY version_number ASC
	`

	rows, err := r.db.QueryContext(ctx, query, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permutations by deck id: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanPermutations(rows)
}

// GetLatest retrieves the most recent permutation for a deck.
func (r *deckPermutationRepository) GetLatest(ctx context.Context, deckID string) (*models.DeckPermutation, error) {
	query := `
		SELECT id, deck_id, parent_permutation_id, cards, card_hash, version_number, version_name, change_summary,
		       matches_played, matches_won, games_played, games_won, created_at, last_played_at
		FROM deck_permutations
		WHERE deck_id = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	perm := &models.DeckPermutation{}
	err := r.db.QueryRowContext(ctx, query, deckID).Scan(
		&perm.ID,
		&perm.DeckID,
		&perm.ParentPermutationID,
		&perm.Cards,
		&perm.CardHash,
		&perm.VersionNumber,
		&perm.VersionName,
		&perm.ChangeSummary,
		&perm.MatchesPlayed,
		&perm.MatchesWon,
		&perm.GamesPlayed,
		&perm.GamesWon,
		&perm.CreatedAt,
		&perm.LastPlayedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get latest permutation: %w", err)
	}

	return perm, nil
}

// GetCurrent retrieves the currently active permutation for a deck.
func (r *deckPermutationRepository) GetCurrent(ctx context.Context, deckID string) (*models.DeckPermutation, error) {
	query := `
		SELECT dp.id, dp.deck_id, dp.parent_permutation_id, dp.cards, dp.card_hash, dp.version_number,
		       dp.version_name, dp.change_summary, dp.matches_played, dp.matches_won,
		       dp.games_played, dp.games_won, dp.created_at, dp.last_played_at
		FROM deck_permutations dp
		JOIN decks d ON d.current_permutation_id = dp.id
		WHERE d.id = ?
	`

	perm := &models.DeckPermutation{}
	err := r.db.QueryRowContext(ctx, query, deckID).Scan(
		&perm.ID,
		&perm.DeckID,
		&perm.ParentPermutationID,
		&perm.Cards,
		&perm.CardHash,
		&perm.VersionNumber,
		&perm.VersionName,
		&perm.ChangeSummary,
		&perm.MatchesPlayed,
		&perm.MatchesWon,
		&perm.GamesPlayed,
		&perm.GamesWon,
		&perm.CreatedAt,
		&perm.LastPlayedAt,
	)

	if err == sql.ErrNoRows {
		// Fall back to latest if no current is set
		return r.GetLatest(ctx, deckID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get current permutation: %w", err)
	}

	return perm, nil
}

// GetByCardHash finds an existing permutation by its card hash.
func (r *deckPermutationRepository) GetByCardHash(ctx context.Context, deckID, cardHash string) (*models.DeckPermutation, error) {
	query := `
		SELECT id, deck_id, parent_permutation_id, cards, card_hash, version_number, version_name, change_summary,
		       matches_played, matches_won, games_played, games_won, created_at, last_played_at
		FROM deck_permutations
		WHERE deck_id = ? AND card_hash = ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	perm := &models.DeckPermutation{}
	err := r.db.QueryRowContext(ctx, query, deckID, cardHash).Scan(
		&perm.ID,
		&perm.DeckID,
		&perm.ParentPermutationID,
		&perm.Cards,
		&perm.CardHash,
		&perm.VersionNumber,
		&perm.VersionName,
		&perm.ChangeSummary,
		&perm.MatchesPlayed,
		&perm.MatchesWon,
		&perm.GamesPlayed,
		&perm.GamesWon,
		&perm.CreatedAt,
		&perm.LastPlayedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get permutation by card hash: %w", err)
	}

	return perm, nil
}

// SetCurrentPermutation sets which permutation is the active one for a deck.
// Validates that the permutation belongs to the specified deck.
func (r *deckPermutationRepository) SetCurrentPermutation(ctx context.Context, deckID string, permutationID int) error {
	// Validate permutation belongs to this deck and update atomically
	query := `
		UPDATE decks SET current_permutation_id = ?
		WHERE id = ? AND EXISTS (
			SELECT 1 FROM deck_permutations WHERE id = ? AND deck_id = ?
		)
	`

	result, err := r.db.ExecContext(ctx, query, permutationID, deckID, permutationID, deckID)
	if err != nil {
		return fmt.Errorf("failed to set current permutation: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("permutation %d does not belong to deck %s", permutationID, deckID)
	}

	return nil
}

// UpdatePerformance updates a permutation's performance stats after a match.
func (r *deckPermutationRepository) UpdatePerformance(ctx context.Context, permutationID int, matchWon bool, gamesWon, gamesLost int) error {
	query := `
		UPDATE deck_permutations
		SET matches_played = matches_played + 1,
		    matches_won = matches_won + ?,
		    games_played = games_played + ?,
		    games_won = games_won + ?,
		    last_played_at = ?
		WHERE id = ?
	`

	matchWonInt := 0
	if matchWon {
		matchWonInt = 1
	}

	now := time.Now().UTC().Format("2006-01-02 15:04:05.999999")
	totalGames := gamesWon + gamesLost

	_, err := r.db.ExecContext(ctx, query, matchWonInt, totalGames, gamesWon, now, permutationID)
	if err != nil {
		return fmt.Errorf("failed to update permutation performance: %w", err)
	}

	return nil
}

// GetPerformance retrieves performance metrics for a permutation.
func (r *deckPermutationRepository) GetPerformance(ctx context.Context, permutationID int) (*models.DeckPermutationPerformance, error) {
	query := `
		SELECT id, deck_id, version_number, version_name, matches_played, matches_won,
		       games_played, games_won, last_played_at, created_at
		FROM deck_permutations
		WHERE id = ?
	`

	perf := &models.DeckPermutationPerformance{}
	err := r.db.QueryRowContext(ctx, query, permutationID).Scan(
		&perf.PermutationID,
		&perf.DeckID,
		&perf.VersionNumber,
		&perf.VersionName,
		&perf.MatchesPlayed,
		&perf.MatchesWon,
		&perf.GamesPlayed,
		&perf.GamesWon,
		&perf.LastPlayedAt,
		&perf.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get permutation performance: %w", err)
	}

	// Calculate win rates
	if perf.MatchesPlayed > 0 {
		perf.MatchWinRate = float64(perf.MatchesWon) / float64(perf.MatchesPlayed)
	}
	if perf.GamesPlayed > 0 {
		perf.GameWinRate = float64(perf.GamesWon) / float64(perf.GamesPlayed)
	}

	return perf, nil
}

// GetAllPerformance retrieves performance metrics for all permutations of a deck.
func (r *deckPermutationRepository) GetAllPerformance(ctx context.Context, deckID string) ([]*models.DeckPermutationPerformance, error) {
	query := `
		SELECT id, deck_id, version_number, version_name, matches_played, matches_won,
		       games_played, games_won, last_played_at, created_at
		FROM deck_permutations
		WHERE deck_id = ?
		ORDER BY version_number ASC
	`

	rows, err := r.db.QueryContext(ctx, query, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get all permutation performance: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var perfs []*models.DeckPermutationPerformance
	for rows.Next() {
		perf := &models.DeckPermutationPerformance{}
		err := rows.Scan(
			&perf.PermutationID,
			&perf.DeckID,
			&perf.VersionNumber,
			&perf.VersionName,
			&perf.MatchesPlayed,
			&perf.MatchesWon,
			&perf.GamesPlayed,
			&perf.GamesWon,
			&perf.LastPlayedAt,
			&perf.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permutation performance: %w", err)
		}

		// Calculate win rates
		if perf.MatchesPlayed > 0 {
			perf.MatchWinRate = float64(perf.MatchesWon) / float64(perf.MatchesPlayed)
		}
		if perf.GamesPlayed > 0 {
			perf.GameWinRate = float64(perf.GamesWon) / float64(perf.GamesPlayed)
		}

		perfs = append(perfs, perf)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permutation performance: %w", err)
	}

	return perfs, nil
}

// GetDiff calculates the difference between two permutations.
func (r *deckPermutationRepository) GetDiff(ctx context.Context, fromPermID, toPermID int) (*models.DeckPermutationDiff, error) {
	// Get both permutations
	fromPerm, err := r.GetByID(ctx, fromPermID)
	if err != nil {
		return nil, fmt.Errorf("failed to get from permutation: %w", err)
	}
	if fromPerm == nil {
		return nil, fmt.Errorf("from permutation not found")
	}

	toPerm, err := r.GetByID(ctx, toPermID)
	if err != nil {
		return nil, fmt.Errorf("failed to get to permutation: %w", err)
	}
	if toPerm == nil {
		return nil, fmt.Errorf("to permutation not found")
	}

	// Parse card JSON
	var fromCards, toCards []models.DeckPermutationCard
	if err := json.Unmarshal([]byte(fromPerm.Cards), &fromCards); err != nil {
		return nil, fmt.Errorf("failed to parse from cards: %w", err)
	}
	if err := json.Unmarshal([]byte(toPerm.Cards), &toCards); err != nil {
		return nil, fmt.Errorf("failed to parse to cards: %w", err)
	}

	// Build maps for comparison (key: cardID-board)
	fromMap := make(map[string]models.DeckPermutationCard)
	for _, card := range fromCards {
		key := fmt.Sprintf("%d-%s", card.CardID, card.Board)
		fromMap[key] = card
	}

	toMap := make(map[string]models.DeckPermutationCard)
	for _, card := range toCards {
		key := fmt.Sprintf("%d-%s", card.CardID, card.Board)
		toMap[key] = card
	}

	diff := &models.DeckPermutationDiff{
		FromPermutationID: fromPermID,
		ToPermutationID:   toPermID,
	}

	// Find added and changed cards
	for key, toCard := range toMap {
		if fromCard, exists := fromMap[key]; exists {
			// Card exists in both - check if quantity changed
			if fromCard.Quantity != toCard.Quantity {
				diff.ChangedCards = append(diff.ChangedCards, models.DeckCardChange{
					CardID:      toCard.CardID,
					Board:       toCard.Board,
					OldQuantity: fromCard.Quantity,
					NewQuantity: toCard.Quantity,
				})
			}
		} else {
			// Card only in to - it's added
			diff.AddedCards = append(diff.AddedCards, toCard)
		}
	}

	// Find removed cards
	for key, fromCard := range fromMap {
		if _, exists := toMap[key]; !exists {
			diff.RemovedCards = append(diff.RemovedCards, fromCard)
		}
	}

	return diff, nil
}

// CreateFromCurrentDeck creates a new permutation from a deck's current cards.
// Uses a transaction to ensure atomicity and consistency.
// If a permutation with the same card_hash already exists, sets it as current and returns it.
func (r *deckPermutationRepository) CreateFromCurrentDeck(ctx context.Context, deckID string, versionName, changeSummary *string) (*models.DeckPermutation, error) {
	// Start transaction for atomicity
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback() // No-op if already committed
	}()

	// Get current deck cards with ordering for deterministic hash
	cardQuery := `
		SELECT card_id, quantity, board
		FROM deck_cards
		WHERE deck_id = ?
		ORDER BY card_id, board
	`

	rows, err := tx.QueryContext(ctx, cardQuery, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deck cards: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var cards []models.DeckPermutationCard
	for rows.Next() {
		var card models.DeckPermutationCard
		if err := rows.Scan(&card.CardID, &card.Quantity, &card.Board); err != nil {
			return nil, fmt.Errorf("failed to scan deck card: %w", err)
		}
		cards = append(cards, card)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating deck cards: %w", err)
	}

	// Compute card hash
	cardHash := computeCardHash(cards)

	// Check if a permutation with this hash already exists
	existingQuery := `
		SELECT id, deck_id, parent_permutation_id, cards, card_hash, version_number, version_name, change_summary,
		       matches_played, matches_won, games_played, games_won, created_at, last_played_at
		FROM deck_permutations
		WHERE deck_id = ? AND card_hash = ?
		LIMIT 1
	`
	existing := &models.DeckPermutation{}
	err = tx.QueryRowContext(ctx, existingQuery, deckID, cardHash).Scan(
		&existing.ID,
		&existing.DeckID,
		&existing.ParentPermutationID,
		&existing.Cards,
		&existing.CardHash,
		&existing.VersionNumber,
		&existing.VersionName,
		&existing.ChangeSummary,
		&existing.MatchesPlayed,
		&existing.MatchesWon,
		&existing.GamesPlayed,
		&existing.GamesWon,
		&existing.CreatedAt,
		&existing.LastPlayedAt,
	)
	if err == nil {
		// Permutation already exists - set it as current and return it
		updateQuery := `
			UPDATE decks SET current_permutation_id = ?
			WHERE id = ? AND EXISTS (SELECT 1 FROM deck_permutations WHERE id = ? AND deck_id = ?)
		`
		_, err = tx.ExecContext(ctx, updateQuery, existing.ID, deckID, existing.ID, deckID)
		if err != nil {
			return nil, fmt.Errorf("failed to set existing permutation as current: %w", err)
		}

		if err = tx.Commit(); err != nil {
			return nil, fmt.Errorf("failed to commit transaction: %w", err)
		}
		return existing, nil
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check for existing permutation: %w", err)
	}

	// Serialize cards to JSON
	cardsJSON, err := json.Marshal(cards)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal cards: %w", err)
	}

	// Get next version number
	versionQuery := `SELECT COALESCE(MAX(version_number), 0) + 1 FROM deck_permutations WHERE deck_id = ?`
	var nextVersion int
	if err = tx.QueryRowContext(ctx, versionQuery, deckID).Scan(&nextVersion); err != nil {
		return nil, fmt.Errorf("failed to get next version number: %w", err)
	}

	// Get current permutation to use as parent
	var parentID *int
	parentQuery := `
		SELECT dp.id FROM deck_permutations dp
		JOIN decks d ON d.current_permutation_id = dp.id
		WHERE d.id = ?
	`
	var parentIDVal int
	err = tx.QueryRowContext(ctx, parentQuery, deckID).Scan(&parentIDVal)
	if err == nil {
		parentID = &parentIDVal
	} else if err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get parent permutation: %w", err)
	}

	// Create the new permutation
	createdAt := time.Now()
	createdAtStr := createdAt.UTC().Format("2006-01-02 15:04:05.999999")

	insertQuery := `
		INSERT INTO deck_permutations (
			deck_id, parent_permutation_id, cards, card_hash, version_number, version_name, change_summary,
			matches_played, matches_won, games_played, games_won, created_at, last_played_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, 0, 0, 0, 0, ?, NULL)
	`
	result, err := tx.ExecContext(ctx, insertQuery,
		deckID, parentID, string(cardsJSON), cardHash, nextVersion, versionName, changeSummary, createdAtStr)
	if err != nil {
		// Handle UNIQUE constraint violation (concurrent insert race condition)
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			// Another transaction inserted this permutation - query and return it
			existing := &models.DeckPermutation{}
			err = tx.QueryRowContext(ctx, existingQuery, deckID, cardHash).Scan(
				&existing.ID, &existing.DeckID, &existing.ParentPermutationID,
				&existing.Cards, &existing.CardHash, &existing.VersionNumber,
				&existing.VersionName, &existing.ChangeSummary,
				&existing.MatchesPlayed, &existing.MatchesWon,
				&existing.GamesPlayed, &existing.GamesWon,
				&existing.CreatedAt, &existing.LastPlayedAt,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to get concurrent permutation: %w", err)
			}
			// Set existing as current
			_, err = tx.ExecContext(ctx, `UPDATE decks SET current_permutation_id = ? WHERE id = ?`,
				existing.ID, deckID)
			if err != nil {
				return nil, fmt.Errorf("failed to set concurrent permutation as current: %w", err)
			}
			if err = tx.Commit(); err != nil {
				return nil, fmt.Errorf("failed to commit transaction: %w", err)
			}
			return existing, nil
		}
		return nil, fmt.Errorf("failed to create permutation: %w", err)
	}

	permID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get permutation id: %w", err)
	}

	// Set as current permutation (with validation)
	updateQuery := `
		UPDATE decks SET current_permutation_id = ?
		WHERE id = ? AND EXISTS (SELECT 1 FROM deck_permutations WHERE id = ? AND deck_id = ?)
	`
	updateResult, err := tx.ExecContext(ctx, updateQuery, permID, deckID, permID, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to set current permutation: %w", err)
	}
	rowsAffected, err := updateResult.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, fmt.Errorf("failed to set current permutation: deck not found")
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	perm := &models.DeckPermutation{
		ID:                  int(permID),
		DeckID:              deckID,
		ParentPermutationID: parentID,
		Cards:               string(cardsJSON),
		CardHash:            cardHash,
		VersionNumber:       nextVersion,
		VersionName:         versionName,
		ChangeSummary:       changeSummary,
		CreatedAt:           createdAt,
	}

	return perm, nil
}

// Delete removes a permutation.
func (r *deckPermutationRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM deck_permutations WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete permutation: %w", err)
	}

	return nil
}

// GetNextVersionNumber returns the next version number for a deck.
func (r *deckPermutationRepository) GetNextVersionNumber(ctx context.Context, deckID string) (int, error) {
	query := `SELECT COALESCE(MAX(version_number), 0) + 1 FROM deck_permutations WHERE deck_id = ?`

	var nextVersion int
	err := r.db.QueryRowContext(ctx, query, deckID).Scan(&nextVersion)
	if err != nil {
		return 0, fmt.Errorf("failed to get next version number: %w", err)
	}

	return nextVersion, nil
}

// scanPermutations is a helper function to scan multiple permutations from rows.
func (r *deckPermutationRepository) scanPermutations(rows *sql.Rows) ([]*models.DeckPermutation, error) {
	var perms []*models.DeckPermutation
	for rows.Next() {
		perm := &models.DeckPermutation{}
		err := rows.Scan(
			&perm.ID,
			&perm.DeckID,
			&perm.ParentPermutationID,
			&perm.Cards,
			&perm.CardHash,
			&perm.VersionNumber,
			&perm.VersionName,
			&perm.ChangeSummary,
			&perm.MatchesPlayed,
			&perm.MatchesWon,
			&perm.GamesPlayed,
			&perm.GamesWon,
			&perm.CreatedAt,
			&perm.LastPlayedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permutation: %w", err)
		}
		perms = append(perms, perm)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permutations: %w", err)
	}

	return perms, nil
}
