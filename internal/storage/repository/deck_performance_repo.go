package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// DeckPerformanceRepository handles database operations for deck performance tracking.
type DeckPerformanceRepository interface {
	// CreateHistory records a new deck performance history entry
	CreateHistory(ctx context.Context, history *models.DeckPerformanceHistory) error

	// GetHistoryByDeck retrieves all performance history for a deck
	GetHistoryByDeck(ctx context.Context, deckID string) ([]*models.DeckPerformanceHistory, error)

	// GetHistoryByArchetype retrieves performance history for an archetype
	GetHistoryByArchetype(ctx context.Context, archetype string, format string) ([]*models.DeckPerformanceHistory, error)

	// GetHistoryByAccount retrieves all performance history for an account
	GetHistoryByAccount(ctx context.Context, accountID int, limit int) ([]*models.DeckPerformanceHistory, error)

	// GetArchetypePerformance calculates aggregated performance stats for an archetype
	GetArchetypePerformance(ctx context.Context, archetype string, format string) (*models.ArchetypePerformanceStats, error)

	// GetPerformanceByDateRange retrieves history within a date range
	GetPerformanceByDateRange(ctx context.Context, accountID int, start, end time.Time) ([]*models.DeckPerformanceHistory, error)

	// CreateArchetype creates a new deck archetype definition
	CreateArchetype(ctx context.Context, archetype *models.DeckArchetype) error

	// GetArchetypeByID retrieves an archetype by ID
	GetArchetypeByID(ctx context.Context, id int) (*models.DeckArchetype, error)

	// GetArchetypeByName retrieves an archetype by name, set, and format
	GetArchetypeByName(ctx context.Context, name string, setCode *string, format string) (*models.DeckArchetype, error)

	// ListArchetypes retrieves all archetypes, optionally filtered
	ListArchetypes(ctx context.Context, setCode *string, format *string) ([]*models.DeckArchetype, error)

	// UpdateArchetypeStats updates the performance statistics for an archetype
	UpdateArchetypeStats(ctx context.Context, archetypeID int, totalMatches, totalWins int) error

	// CreateCardWeight creates a card weight association for an archetype
	CreateCardWeight(ctx context.Context, weight *models.ArchetypeCardWeight) error

	// GetCardWeights retrieves all card weights for an archetype
	GetCardWeights(ctx context.Context, archetypeID int) ([]*models.ArchetypeCardWeight, error)

	// GetCardWeightsByCard retrieves all archetype associations for a card
	GetCardWeightsByCard(ctx context.Context, cardID int) ([]*models.ArchetypeCardWeight, error)

	// UpsertCardWeight creates or updates a card weight
	UpsertCardWeight(ctx context.Context, weight *models.ArchetypeCardWeight) error

	// DeleteCardWeight removes a card weight association
	DeleteCardWeight(ctx context.Context, archetypeID, cardID int) error
}

type deckPerformanceRepository struct {
	db *sql.DB
}

// NewDeckPerformanceRepository creates a new deck performance repository.
func NewDeckPerformanceRepository(db *sql.DB) DeckPerformanceRepository {
	return &deckPerformanceRepository{db: db}
}

// CreateHistory records a new deck performance history entry.
func (r *deckPerformanceRepository) CreateHistory(ctx context.Context, history *models.DeckPerformanceHistory) error {
	query := `
		INSERT INTO deck_performance_history (
			account_id, deck_id, match_id, archetype, secondary_archetype,
			archetype_confidence, color_identity, card_count, result,
			games_won, games_lost, duration_seconds, format, event_type,
			opponent_archetype, rank_tier, match_timestamp, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	matchTimestampStr := history.MatchTimestamp.UTC().Format("2006-01-02 15:04:05.999999")
	createdAtStr := time.Now().UTC().Format("2006-01-02 15:04:05.999999")

	result, err := r.db.ExecContext(ctx, query,
		history.AccountID, history.DeckID, history.MatchID, history.Archetype,
		history.SecondaryArchetype, history.ArchetypeConfidence, history.ColorIdentity,
		history.CardCount, history.Result, history.GamesWon, history.GamesLost,
		history.DurationSeconds, history.Format, history.EventType,
		history.OpponentArchetype, history.RankTier, matchTimestampStr, createdAtStr,
	)
	if err != nil {
		return fmt.Errorf("failed to create deck performance history: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get inserted ID: %w", err)
	}
	history.ID = int(id)

	return nil
}

// GetHistoryByDeck retrieves all performance history for a deck.
func (r *deckPerformanceRepository) GetHistoryByDeck(ctx context.Context, deckID string) ([]*models.DeckPerformanceHistory, error) {
	query := `
		SELECT id, account_id, deck_id, match_id, archetype, secondary_archetype,
			   archetype_confidence, color_identity, card_count, result,
			   games_won, games_lost, duration_seconds, format, event_type,
			   opponent_archetype, rank_tier, match_timestamp, created_at
		FROM deck_performance_history
		WHERE deck_id = ?
		ORDER BY match_timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to query deck performance history: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanHistoryRows(rows)
}

// GetHistoryByArchetype retrieves performance history for an archetype.
func (r *deckPerformanceRepository) GetHistoryByArchetype(ctx context.Context, archetype string, format string) ([]*models.DeckPerformanceHistory, error) {
	query := `
		SELECT id, account_id, deck_id, match_id, archetype, secondary_archetype,
			   archetype_confidence, color_identity, card_count, result,
			   games_won, games_lost, duration_seconds, format, event_type,
			   opponent_archetype, rank_tier, match_timestamp, created_at
		FROM deck_performance_history
		WHERE archetype = ? AND format = ?
		ORDER BY match_timestamp DESC
	`

	rows, err := r.db.QueryContext(ctx, query, archetype, format)
	if err != nil {
		return nil, fmt.Errorf("failed to query archetype performance history: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanHistoryRows(rows)
}

// GetHistoryByAccount retrieves all performance history for an account.
func (r *deckPerformanceRepository) GetHistoryByAccount(ctx context.Context, accountID int, limit int) ([]*models.DeckPerformanceHistory, error) {
	query := `
		SELECT id, account_id, deck_id, match_id, archetype, secondary_archetype,
			   archetype_confidence, color_identity, card_count, result,
			   games_won, games_lost, duration_seconds, format, event_type,
			   opponent_archetype, rank_tier, match_timestamp, created_at
		FROM deck_performance_history
		WHERE account_id = ?
		ORDER BY match_timestamp DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, accountID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query account performance history: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanHistoryRows(rows)
}

// GetArchetypePerformance calculates aggregated performance stats for an archetype.
func (r *deckPerformanceRepository) GetArchetypePerformance(ctx context.Context, archetype string, format string) (*models.ArchetypePerformanceStats, error) {
	query := `
		SELECT
			COUNT(*) as total_matches,
			SUM(CASE WHEN result = 'win' THEN 1 ELSE 0 END) as total_wins,
			AVG(duration_seconds) as avg_duration
		FROM deck_performance_history
		WHERE archetype = ? AND format = ?
	`

	var totalMatches, totalWins int
	var avgDuration *float64

	err := r.db.QueryRowContext(ctx, query, archetype, format).Scan(&totalMatches, &totalWins, &avgDuration)
	if err != nil {
		return nil, fmt.Errorf("failed to get archetype performance: %w", err)
	}

	var winRate float64
	if totalMatches > 0 {
		winRate = float64(totalWins) / float64(totalMatches)
	}

	return &models.ArchetypePerformanceStats{
		ArchetypeName: archetype,
		Format:        format,
		TotalMatches:  totalMatches,
		TotalWins:     totalWins,
		WinRate:       winRate,
		AvgDuration:   avgDuration,
	}, nil
}

// GetPerformanceByDateRange retrieves history within a date range.
func (r *deckPerformanceRepository) GetPerformanceByDateRange(ctx context.Context, accountID int, start, end time.Time) ([]*models.DeckPerformanceHistory, error) {
	query := `
		SELECT id, account_id, deck_id, match_id, archetype, secondary_archetype,
			   archetype_confidence, color_identity, card_count, result,
			   games_won, games_lost, duration_seconds, format, event_type,
			   opponent_archetype, rank_tier, match_timestamp, created_at
		FROM deck_performance_history
		WHERE account_id = ? AND match_timestamp BETWEEN ? AND ?
		ORDER BY match_timestamp DESC
	`

	startStr := start.UTC().Format("2006-01-02 15:04:05.999999")
	endStr := end.UTC().Format("2006-01-02 15:04:05.999999")

	rows, err := r.db.QueryContext(ctx, query, accountID, startStr, endStr)
	if err != nil {
		return nil, fmt.Errorf("failed to query performance by date range: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanHistoryRows(rows)
}

// scanHistoryRows scans rows into DeckPerformanceHistory slice.
func (r *deckPerformanceRepository) scanHistoryRows(rows *sql.Rows) ([]*models.DeckPerformanceHistory, error) {
	var histories []*models.DeckPerformanceHistory

	for rows.Next() {
		h := &models.DeckPerformanceHistory{}
		var matchTimestamp, createdAt string

		err := rows.Scan(
			&h.ID, &h.AccountID, &h.DeckID, &h.MatchID, &h.Archetype,
			&h.SecondaryArchetype, &h.ArchetypeConfidence, &h.ColorIdentity,
			&h.CardCount, &h.Result, &h.GamesWon, &h.GamesLost,
			&h.DurationSeconds, &h.Format, &h.EventType,
			&h.OpponentArchetype, &h.RankTier, &matchTimestamp, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan history row: %w", err)
		}

		h.MatchTimestamp, _ = time.Parse("2006-01-02 15:04:05.999999", matchTimestamp)
		h.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", createdAt)

		histories = append(histories, h)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history rows: %w", err)
	}

	return histories, nil
}

// CreateArchetype creates a new deck archetype definition.
func (r *deckPerformanceRepository) CreateArchetype(ctx context.Context, archetype *models.DeckArchetype) error {
	query := `
		INSERT INTO deck_archetypes (
			name, set_code, format, color_identity, signature_cards,
			synergy_patterns, total_matches, total_wins, avg_win_rate,
			source, external_id, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().UTC().Format("2006-01-02 15:04:05.999999")

	result, err := r.db.ExecContext(ctx, query,
		archetype.Name, archetype.SetCode, archetype.Format, archetype.ColorIdentity,
		archetype.SignatureCards, archetype.SynergyPatterns, archetype.TotalMatches,
		archetype.TotalWins, archetype.AvgWinRate, archetype.Source,
		archetype.ExternalID, now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to create archetype: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get inserted archetype ID: %w", err)
	}
	archetype.ID = int(id)

	return nil
}

// GetArchetypeByID retrieves an archetype by ID.
func (r *deckPerformanceRepository) GetArchetypeByID(ctx context.Context, id int) (*models.DeckArchetype, error) {
	query := `
		SELECT id, name, set_code, format, color_identity, signature_cards,
			   synergy_patterns, total_matches, total_wins, avg_win_rate,
			   source, external_id, created_at, updated_at
		FROM deck_archetypes
		WHERE id = ?
	`

	archetype := &models.DeckArchetype{}
	var createdAt, updatedAt string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&archetype.ID, &archetype.Name, &archetype.SetCode, &archetype.Format,
		&archetype.ColorIdentity, &archetype.SignatureCards, &archetype.SynergyPatterns,
		&archetype.TotalMatches, &archetype.TotalWins, &archetype.AvgWinRate,
		&archetype.Source, &archetype.ExternalID, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get archetype: %w", err)
	}

	archetype.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", createdAt)
	archetype.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", updatedAt)

	return archetype, nil
}

// GetArchetypeByName retrieves an archetype by name, set, and format.
func (r *deckPerformanceRepository) GetArchetypeByName(ctx context.Context, name string, setCode *string, format string) (*models.DeckArchetype, error) {
	var query string
	var args []interface{}

	if setCode == nil {
		query = `
			SELECT id, name, set_code, format, color_identity, signature_cards,
				   synergy_patterns, total_matches, total_wins, avg_win_rate,
				   source, external_id, created_at, updated_at
			FROM deck_archetypes
			WHERE name = ? AND set_code IS NULL AND format = ?
		`
		args = []interface{}{name, format}
	} else {
		query = `
			SELECT id, name, set_code, format, color_identity, signature_cards,
				   synergy_patterns, total_matches, total_wins, avg_win_rate,
				   source, external_id, created_at, updated_at
			FROM deck_archetypes
			WHERE name = ? AND set_code = ? AND format = ?
		`
		args = []interface{}{name, *setCode, format}
	}

	archetype := &models.DeckArchetype{}
	var createdAt, updatedAt string

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&archetype.ID, &archetype.Name, &archetype.SetCode, &archetype.Format,
		&archetype.ColorIdentity, &archetype.SignatureCards, &archetype.SynergyPatterns,
		&archetype.TotalMatches, &archetype.TotalWins, &archetype.AvgWinRate,
		&archetype.Source, &archetype.ExternalID, &createdAt, &updatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get archetype by name: %w", err)
	}

	archetype.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", createdAt)
	archetype.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", updatedAt)

	return archetype, nil
}

// ListArchetypes retrieves all archetypes, optionally filtered.
func (r *deckPerformanceRepository) ListArchetypes(ctx context.Context, setCode *string, format *string) ([]*models.DeckArchetype, error) {
	query := `
		SELECT id, name, set_code, format, color_identity, signature_cards,
			   synergy_patterns, total_matches, total_wins, avg_win_rate,
			   source, external_id, created_at, updated_at
		FROM deck_archetypes
		WHERE 1=1
	`

	var args []interface{}

	if setCode != nil {
		query += " AND set_code = ?"
		args = append(args, *setCode)
	}

	if format != nil {
		query += " AND format = ?"
		args = append(args, *format)
	}

	query += " ORDER BY name"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list archetypes: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var archetypes []*models.DeckArchetype

	for rows.Next() {
		a := &models.DeckArchetype{}
		var createdAt, updatedAt string

		err := rows.Scan(
			&a.ID, &a.Name, &a.SetCode, &a.Format, &a.ColorIdentity,
			&a.SignatureCards, &a.SynergyPatterns, &a.TotalMatches,
			&a.TotalWins, &a.AvgWinRate, &a.Source, &a.ExternalID,
			&createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan archetype row: %w", err)
		}

		a.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", createdAt)
		a.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", updatedAt)

		archetypes = append(archetypes, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating archetype rows: %w", err)
	}

	return archetypes, nil
}

// UpdateArchetypeStats updates the performance statistics for an archetype.
func (r *deckPerformanceRepository) UpdateArchetypeStats(ctx context.Context, archetypeID int, totalMatches, totalWins int) error {
	var avgWinRate *float64
	if totalMatches > 0 {
		rate := float64(totalWins) / float64(totalMatches)
		avgWinRate = &rate
	}

	query := `
		UPDATE deck_archetypes
		SET total_matches = ?, total_wins = ?, avg_win_rate = ?,
			updated_at = ?
		WHERE id = ?
	`

	now := time.Now().UTC().Format("2006-01-02 15:04:05.999999")

	_, err := r.db.ExecContext(ctx, query, totalMatches, totalWins, avgWinRate, now, archetypeID)
	if err != nil {
		return fmt.Errorf("failed to update archetype stats: %w", err)
	}

	return nil
}

// CreateCardWeight creates a card weight association for an archetype.
func (r *deckPerformanceRepository) CreateCardWeight(ctx context.Context, weight *models.ArchetypeCardWeight) error {
	query := `
		INSERT INTO archetype_card_weights (
			archetype_id, card_id, weight, is_signature, source, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().UTC().Format("2006-01-02 15:04:05.999999")
	isSignature := 0
	if weight.IsSignature {
		isSignature = 1
	}

	result, err := r.db.ExecContext(ctx, query,
		weight.ArchetypeID, weight.CardID, weight.Weight, isSignature,
		weight.Source, now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to create card weight: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get inserted card weight ID: %w", err)
	}
	weight.ID = int(id)

	return nil
}

// GetCardWeights retrieves all card weights for an archetype.
func (r *deckPerformanceRepository) GetCardWeights(ctx context.Context, archetypeID int) ([]*models.ArchetypeCardWeight, error) {
	query := `
		SELECT id, archetype_id, card_id, weight, is_signature, source, created_at, updated_at
		FROM archetype_card_weights
		WHERE archetype_id = ?
		ORDER BY weight DESC
	`

	rows, err := r.db.QueryContext(ctx, query, archetypeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get card weights: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanCardWeightRows(rows)
}

// GetCardWeightsByCard retrieves all archetype associations for a card.
func (r *deckPerformanceRepository) GetCardWeightsByCard(ctx context.Context, cardID int) ([]*models.ArchetypeCardWeight, error) {
	query := `
		SELECT id, archetype_id, card_id, weight, is_signature, source, created_at, updated_at
		FROM archetype_card_weights
		WHERE card_id = ?
		ORDER BY weight DESC
	`

	rows, err := r.db.QueryContext(ctx, query, cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to get card weights by card: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanCardWeightRows(rows)
}

// UpsertCardWeight creates or updates a card weight.
func (r *deckPerformanceRepository) UpsertCardWeight(ctx context.Context, weight *models.ArchetypeCardWeight) error {
	query := `
		INSERT INTO archetype_card_weights (
			archetype_id, card_id, weight, is_signature, source, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(archetype_id, card_id) DO UPDATE SET
			weight = excluded.weight,
			is_signature = excluded.is_signature,
			source = excluded.source,
			updated_at = excluded.updated_at
	`

	now := time.Now().UTC().Format("2006-01-02 15:04:05.999999")
	isSignature := 0
	if weight.IsSignature {
		isSignature = 1
	}

	_, err := r.db.ExecContext(ctx, query,
		weight.ArchetypeID, weight.CardID, weight.Weight, isSignature,
		weight.Source, now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert card weight: %w", err)
	}

	return nil
}

// DeleteCardWeight removes a card weight association.
func (r *deckPerformanceRepository) DeleteCardWeight(ctx context.Context, archetypeID, cardID int) error {
	query := `DELETE FROM archetype_card_weights WHERE archetype_id = ? AND card_id = ?`

	_, err := r.db.ExecContext(ctx, query, archetypeID, cardID)
	if err != nil {
		return fmt.Errorf("failed to delete card weight: %w", err)
	}

	return nil
}

// scanCardWeightRows scans rows into ArchetypeCardWeight slice.
func (r *deckPerformanceRepository) scanCardWeightRows(rows *sql.Rows) ([]*models.ArchetypeCardWeight, error) {
	var weights []*models.ArchetypeCardWeight

	for rows.Next() {
		w := &models.ArchetypeCardWeight{}
		var isSignature int
		var createdAt, updatedAt string

		err := rows.Scan(
			&w.ID, &w.ArchetypeID, &w.CardID, &w.Weight, &isSignature,
			&w.Source, &createdAt, &updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card weight row: %w", err)
		}

		w.IsSignature = isSignature == 1
		w.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", createdAt)
		w.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", updatedAt)

		weights = append(weights, w)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating card weight rows: %w", err)
	}

	return weights, nil
}
