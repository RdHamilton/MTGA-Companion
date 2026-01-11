package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// CooccurrenceRepository handles card co-occurrence data operations.
type CooccurrenceRepository interface {
	// UpsertCooccurrence inserts or updates a card pair co-occurrence count.
	UpsertCooccurrence(ctx context.Context, cardAArenaID, cardBArenaID int, format string, count int) error

	// IncrementCooccurrence increments the count for a card pair.
	IncrementCooccurrence(ctx context.Context, cardAArenaID, cardBArenaID int, format string) error

	// GetCooccurrence gets the co-occurrence record for a card pair.
	GetCooccurrence(ctx context.Context, cardAArenaID, cardBArenaID int, format string) (*models.CardCooccurrence, error)

	// GetTopCooccurrences gets the top co-occurring cards for a given card.
	GetTopCooccurrences(ctx context.Context, cardArenaID int, format string, limit int) ([]*models.CardCooccurrence, error)

	// UpdatePMIScores updates PMI scores for all co-occurrences in a format.
	UpdatePMIScores(ctx context.Context, format string) error

	// GetCooccurrenceScore returns the PMI score for a card pair.
	GetCooccurrenceScore(ctx context.Context, cardAArenaID, cardBArenaID int, format string) (float64, error)

	// UpsertCardFrequency updates the frequency record for a card.
	UpsertCardFrequency(ctx context.Context, cardArenaID int, format string, deckCount, totalDecks int) error

	// GetCardFrequency gets the frequency record for a card.
	GetCardFrequency(ctx context.Context, cardArenaID int, format string) (*models.CardFrequency, error)

	// UpsertSource updates the source tracking record.
	UpsertSource(ctx context.Context, sourceType, sourceID, format string, deckCount, cardCount int) error

	// GetSource gets a source tracking record.
	GetSource(ctx context.Context, sourceType, sourceID, format string) (*models.CooccurrenceSource, error)

	// ClearFormat removes all co-occurrence data for a format.
	ClearFormat(ctx context.Context, format string) error
}

// cooccurrenceRepo implements CooccurrenceRepository.
type cooccurrenceRepo struct {
	db *sql.DB
}

// NewCooccurrenceRepository creates a new co-occurrence repository.
func NewCooccurrenceRepository(db *sql.DB) CooccurrenceRepository {
	return &cooccurrenceRepo{db: db}
}

// UpsertCooccurrence inserts or updates a card pair co-occurrence count.
func (r *cooccurrenceRepo) UpsertCooccurrence(ctx context.Context, cardAArenaID, cardBArenaID int, format string, count int) error {
	// Ensure consistent ordering (smaller ID first)
	if cardAArenaID > cardBArenaID {
		cardAArenaID, cardBArenaID = cardBArenaID, cardAArenaID
	}

	query := `
		INSERT INTO card_cooccurrence (card_a_arena_id, card_b_arena_id, format, count, last_updated)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(card_a_arena_id, card_b_arena_id, format) DO UPDATE SET
			count = excluded.count,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query, cardAArenaID, cardBArenaID, format, count)
	if err != nil {
		return fmt.Errorf("failed to upsert co-occurrence: %w", err)
	}

	return nil
}

// IncrementCooccurrence increments the count for a card pair.
func (r *cooccurrenceRepo) IncrementCooccurrence(ctx context.Context, cardAArenaID, cardBArenaID int, format string) error {
	// Ensure consistent ordering (smaller ID first)
	if cardAArenaID > cardBArenaID {
		cardAArenaID, cardBArenaID = cardBArenaID, cardAArenaID
	}

	query := `
		INSERT INTO card_cooccurrence (card_a_arena_id, card_b_arena_id, format, count, last_updated)
		VALUES (?, ?, ?, 1, CURRENT_TIMESTAMP)
		ON CONFLICT(card_a_arena_id, card_b_arena_id, format) DO UPDATE SET
			count = card_cooccurrence.count + 1,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query, cardAArenaID, cardBArenaID, format)
	if err != nil {
		return fmt.Errorf("failed to increment co-occurrence: %w", err)
	}

	return nil
}

// GetCooccurrence gets the co-occurrence record for a card pair.
func (r *cooccurrenceRepo) GetCooccurrence(ctx context.Context, cardAArenaID, cardBArenaID int, format string) (*models.CardCooccurrence, error) {
	// Ensure consistent ordering
	if cardAArenaID > cardBArenaID {
		cardAArenaID, cardBArenaID = cardBArenaID, cardAArenaID
	}

	query := `
		SELECT id, card_a_arena_id, card_b_arena_id, format, count, pmi_score, last_updated
		FROM card_cooccurrence
		WHERE card_a_arena_id = ? AND card_b_arena_id = ? AND format = ?
	`

	var cooc models.CardCooccurrence
	err := r.db.QueryRowContext(ctx, query, cardAArenaID, cardBArenaID, format).Scan(
		&cooc.ID, &cooc.CardAArenaID, &cooc.CardBArenaID, &cooc.Format,
		&cooc.Count, &cooc.PMIScore, &cooc.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get co-occurrence: %w", err)
	}

	return &cooc, nil
}

// GetTopCooccurrences gets the top co-occurring cards for a given card.
func (r *cooccurrenceRepo) GetTopCooccurrences(ctx context.Context, cardArenaID int, format string, limit int) ([]*models.CardCooccurrence, error) {
	query := `
		SELECT id, card_a_arena_id, card_b_arena_id, format, count, pmi_score, last_updated
		FROM card_cooccurrence
		WHERE (card_a_arena_id = ? OR card_b_arena_id = ?) AND format = ?
		ORDER BY pmi_score DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, cardArenaID, cardArenaID, format, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top co-occurrences: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var result []*models.CardCooccurrence
	for rows.Next() {
		var cooc models.CardCooccurrence
		if err := rows.Scan(
			&cooc.ID, &cooc.CardAArenaID, &cooc.CardBArenaID, &cooc.Format,
			&cooc.Count, &cooc.PMIScore, &cooc.LastUpdated,
		); err != nil {
			return nil, fmt.Errorf("failed to scan co-occurrence: %w", err)
		}
		result = append(result, &cooc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating co-occurrences: %w", err)
	}

	return result, nil
}

// UpdatePMIScores updates PMI scores for all co-occurrences in a format.
// PMI(A,B) = log(P(A,B) / (P(A) * P(B)))
// where P(A,B) is the probability of A and B appearing together,
// and P(A), P(B) are the individual probabilities.
func (r *cooccurrenceRepo) UpdatePMIScores(ctx context.Context, format string) error {
	// Get total deck count for this format from card_frequency
	var totalDecks int
	err := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(total_decks), 0) FROM card_frequency WHERE format = ?
	`, format).Scan(&totalDecks)
	if err != nil {
		return fmt.Errorf("failed to get total decks: %w", err)
	}

	if totalDecks == 0 {
		return nil // No data to calculate PMI
	}

	// Update PMI scores using the formula:
	// PMI = log2(P(A,B) / (P(A) * P(B)))
	// P(A,B) = count / totalDecks
	// P(A) = freq_a.deck_count / totalDecks
	// P(B) = freq_b.deck_count / totalDecks
	//
	// Simplified: PMI = log2(count * totalDecks / (freq_a * freq_b))
	query := `
		UPDATE card_cooccurrence
		SET pmi_score = (
			SELECT CASE
				WHEN freq_a.deck_count > 0 AND freq_b.deck_count > 0 THEN
					log(1.0 * card_cooccurrence.count * ? / (freq_a.deck_count * freq_b.deck_count)) / log(2)
				ELSE 0
			END
			FROM card_frequency freq_a, card_frequency freq_b
			WHERE freq_a.card_arena_id = card_cooccurrence.card_a_arena_id
			  AND freq_a.format = card_cooccurrence.format
			  AND freq_b.card_arena_id = card_cooccurrence.card_b_arena_id
			  AND freq_b.format = card_cooccurrence.format
		),
		last_updated = CURRENT_TIMESTAMP
		WHERE format = ?
	`

	_, err = r.db.ExecContext(ctx, query, totalDecks, format)
	if err != nil {
		return fmt.Errorf("failed to update PMI scores: %w", err)
	}

	return nil
}

// GetCooccurrenceScore returns the PMI score for a card pair.
func (r *cooccurrenceRepo) GetCooccurrenceScore(ctx context.Context, cardAArenaID, cardBArenaID int, format string) (float64, error) {
	cooc, err := r.GetCooccurrence(ctx, cardAArenaID, cardBArenaID, format)
	if err != nil {
		return 0, err
	}
	if cooc == nil {
		return 0, nil
	}
	return cooc.PMIScore, nil
}

// UpsertCardFrequency updates the frequency record for a card.
func (r *cooccurrenceRepo) UpsertCardFrequency(ctx context.Context, cardArenaID int, format string, deckCount, totalDecks int) error {
	query := `
		INSERT INTO card_frequency (card_arena_id, format, deck_count, total_decks, last_updated)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(card_arena_id, format) DO UPDATE SET
			deck_count = excluded.deck_count,
			total_decks = excluded.total_decks,
			last_updated = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query, cardArenaID, format, deckCount, totalDecks)
	if err != nil {
		return fmt.Errorf("failed to upsert card frequency: %w", err)
	}

	return nil
}

// GetCardFrequency gets the frequency record for a card.
func (r *cooccurrenceRepo) GetCardFrequency(ctx context.Context, cardArenaID int, format string) (*models.CardFrequency, error) {
	query := `
		SELECT id, card_arena_id, format, deck_count, total_decks, frequency, last_updated
		FROM card_frequency
		WHERE card_arena_id = ? AND format = ?
	`

	var freq models.CardFrequency
	err := r.db.QueryRowContext(ctx, query, cardArenaID, format).Scan(
		&freq.ID, &freq.CardArenaID, &freq.Format, &freq.DeckCount,
		&freq.TotalDecks, &freq.Frequency, &freq.LastUpdated,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get card frequency: %w", err)
	}

	return &freq, nil
}

// UpsertSource updates the source tracking record.
func (r *cooccurrenceRepo) UpsertSource(ctx context.Context, sourceType, sourceID, format string, deckCount, cardCount int) error {
	query := `
		INSERT INTO cooccurrence_sources (source_type, source_id, format, deck_count, card_count, last_synced)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(source_type, source_id, format) DO UPDATE SET
			deck_count = excluded.deck_count,
			card_count = excluded.card_count,
			last_synced = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query, sourceType, sourceID, format, deckCount, cardCount)
	if err != nil {
		return fmt.Errorf("failed to upsert source: %w", err)
	}

	return nil
}

// GetSource gets a source tracking record.
func (r *cooccurrenceRepo) GetSource(ctx context.Context, sourceType, sourceID, format string) (*models.CooccurrenceSource, error) {
	query := `
		SELECT id, source_type, source_id, format, deck_count, card_count, last_synced
		FROM cooccurrence_sources
		WHERE source_type = ? AND source_id = ? AND format = ?
	`

	var source models.CooccurrenceSource
	err := r.db.QueryRowContext(ctx, query, sourceType, sourceID, format).Scan(
		&source.ID, &source.SourceType, &source.SourceID, &source.Format,
		&source.DeckCount, &source.CardCount, &source.LastSynced,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}

	return &source, nil
}

// ClearFormat removes all co-occurrence data for a format.
func (r *cooccurrenceRepo) ClearFormat(ctx context.Context, format string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, "DELETE FROM card_cooccurrence WHERE format = ?", format)
	if err != nil {
		return fmt.Errorf("failed to clear co-occurrences: %w", err)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM card_frequency WHERE format = ?", format)
	if err != nil {
		return fmt.Errorf("failed to clear frequencies: %w", err)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM cooccurrence_sources WHERE format = ?", format)
	if err != nil {
		return fmt.Errorf("failed to clear sources: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}
