package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// MLSuggestionRepository handles ML suggestion data operations.
type MLSuggestionRepository struct {
	db *sql.DB
}

// NewMLSuggestionRepository creates a new MLSuggestionRepository.
func NewMLSuggestionRepository(db *sql.DB) *MLSuggestionRepository {
	return &MLSuggestionRepository{db: db}
}

// ============================================================================
// Individual Card Stats
// ============================================================================

// UpsertIndividualCardStats inserts or updates individual card statistics.
func (r *MLSuggestionRepository) UpsertIndividualCardStats(ctx context.Context, stats *models.CardIndividualStats) error {
	query := `
		INSERT INTO card_individual_stats (card_id, format, total_games, wins, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(card_id, format) DO UPDATE SET
			total_games = total_games + excluded.total_games,
			wins = wins + excluded.wins,
			updated_at = excluded.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		stats.CardID, stats.Format, stats.TotalGames, stats.Wins, time.Now(),
	)
	return err
}

// GetIndividualCardStats retrieves stats for a specific card.
func (r *MLSuggestionRepository) GetIndividualCardStats(ctx context.Context, cardID int, format string) (*models.CardIndividualStats, error) {
	query := `
		SELECT card_id, format, total_games, wins, updated_at
		FROM card_individual_stats
		WHERE card_id = ? AND format = ?
	`

	var stats models.CardIndividualStats
	err := r.db.QueryRowContext(ctx, query, cardID, format).Scan(
		&stats.CardID, &stats.Format, &stats.TotalGames, &stats.Wins, &stats.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// UpdateSeparateStatsFromIndividual calculates and updates games_card1_only, wins_card1_only, etc.
// for all card combinations based on individual card stats.
// This should be called after processing all matches.
func (r *MLSuggestionRepository) UpdateSeparateStatsFromIndividual(ctx context.Context, format string) error {
	// For each card pair, calculate separate stats:
	// games_card1_only = individual_games_card1 - games_together
	// wins_card1_only = individual_wins_card1 - wins_together (proportionally)
	// Note: This is an approximation since we track total individual stats, not per-pair

	query := `
		UPDATE card_combination_stats
		SET
			games_card1_only = COALESCE(
				(SELECT i1.total_games FROM card_individual_stats i1 WHERE i1.card_id = card_combination_stats.card_id_1 AND i1.format = card_combination_stats.format),
				0
			) - games_together,
			games_card2_only = COALESCE(
				(SELECT i2.total_games FROM card_individual_stats i2 WHERE i2.card_id = card_combination_stats.card_id_2 AND i2.format = card_combination_stats.format),
				0
			) - games_together,
			wins_card1_only = CASE
				WHEN COALESCE(
					(SELECT i1.total_games FROM card_individual_stats i1 WHERE i1.card_id = card_combination_stats.card_id_1 AND i1.format = card_combination_stats.format),
					0
				) - games_together <= 0 THEN 0
				ELSE CAST(
					(COALESCE(
						(SELECT i1.wins FROM card_individual_stats i1 WHERE i1.card_id = card_combination_stats.card_id_1 AND i1.format = card_combination_stats.format),
						0
					) - wins_together) AS INTEGER
				)
			END,
			wins_card2_only = CASE
				WHEN COALESCE(
					(SELECT i2.total_games FROM card_individual_stats i2 WHERE i2.card_id = card_combination_stats.card_id_2 AND i2.format = card_combination_stats.format),
					0
				) - games_together <= 0 THEN 0
				ELSE CAST(
					(COALESCE(
						(SELECT i2.wins FROM card_individual_stats i2 WHERE i2.card_id = card_combination_stats.card_id_2 AND i2.format = card_combination_stats.format),
						0
					) - wins_together) AS INTEGER
				)
			END,
			updated_at = ?
		WHERE format = ?
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), format)
	return err
}

// ============================================================================
// Card Combination Stats
// ============================================================================

// UpsertCombinationStats inserts or updates card combination statistics.
func (r *MLSuggestionRepository) UpsertCombinationStats(ctx context.Context, stats *models.CardCombinationStats) error {
	// Ensure card_id_1 < card_id_2 for uniqueness
	if stats.CardID1 > stats.CardID2 {
		stats.CardID1, stats.CardID2 = stats.CardID2, stats.CardID1
	}

	query := `
		INSERT INTO card_combination_stats (
			card_id_1, card_id_2, deck_id, format,
			games_together, games_card1_only, games_card2_only,
			wins_together, wins_card1_only, wins_card2_only,
			synergy_score, confidence_score, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(card_id_1, card_id_2, deck_id, format) DO UPDATE SET
			games_together = games_together + excluded.games_together,
			games_card1_only = games_card1_only + excluded.games_card1_only,
			games_card2_only = games_card2_only + excluded.games_card2_only,
			wins_together = wins_together + excluded.wins_together,
			wins_card1_only = wins_card1_only + excluded.wins_card1_only,
			wins_card2_only = wins_card2_only + excluded.wins_card2_only,
			updated_at = excluded.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		stats.CardID1, stats.CardID2, stats.DeckID, stats.Format,
		stats.GamesTogether, stats.GamesCard1Only, stats.GamesCard2Only,
		stats.WinsTogether, stats.WinsCard1Only, stats.WinsCard2Only,
		stats.SynergyScore, stats.ConfidenceScore, time.Now(),
	)
	return err
}

// GetCombinationStats retrieves stats for a specific card pair.
func (r *MLSuggestionRepository) GetCombinationStats(ctx context.Context, cardID1, cardID2 int, format string) (*models.CardCombinationStats, error) {
	// Ensure proper ordering
	if cardID1 > cardID2 {
		cardID1, cardID2 = cardID2, cardID1
	}

	query := `
		SELECT id, card_id_1, card_id_2, deck_id, format,
			games_together, games_card1_only, games_card2_only,
			wins_together, wins_card1_only, wins_card2_only,
			synergy_score, confidence_score, created_at, updated_at
		FROM card_combination_stats
		WHERE card_id_1 = ? AND card_id_2 = ? AND format = ?
	`

	var stats models.CardCombinationStats
	var deckID sql.NullString
	err := r.db.QueryRowContext(ctx, query, cardID1, cardID2, format).Scan(
		&stats.ID, &stats.CardID1, &stats.CardID2, &deckID, &stats.Format,
		&stats.GamesTogether, &stats.GamesCard1Only, &stats.GamesCard2Only,
		&stats.WinsTogether, &stats.WinsCard1Only, &stats.WinsCard2Only,
		&stats.SynergyScore, &stats.ConfidenceScore, &stats.CreatedAt, &stats.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if deckID.Valid {
		stats.DeckID = deckID.String
	}
	return &stats, nil
}

// GetTopSynergiesForCard returns the top synergistic cards for a given card.
func (r *MLSuggestionRepository) GetTopSynergiesForCard(ctx context.Context, cardID int, format string, limit int) ([]*models.CardCombinationStats, error) {
	query := `
		SELECT id, card_id_1, card_id_2, deck_id, format,
			games_together, games_card1_only, games_card2_only,
			wins_together, wins_card1_only, wins_card2_only,
			synergy_score, confidence_score, created_at, updated_at
		FROM card_combination_stats
		WHERE (card_id_1 = ? OR card_id_2 = ?) AND format = ?
			AND games_together >= 5
		ORDER BY synergy_score DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, cardID, cardID, format, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []*models.CardCombinationStats
	for rows.Next() {
		var stats models.CardCombinationStats
		var deckID sql.NullString
		if err := rows.Scan(
			&stats.ID, &stats.CardID1, &stats.CardID2, &deckID, &stats.Format,
			&stats.GamesTogether, &stats.GamesCard1Only, &stats.GamesCard2Only,
			&stats.WinsTogether, &stats.WinsCard1Only, &stats.WinsCard2Only,
			&stats.SynergyScore, &stats.ConfidenceScore, &stats.CreatedAt, &stats.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if deckID.Valid {
			stats.DeckID = deckID.String
		}
		results = append(results, &stats)
	}
	return results, rows.Err()
}

// CalculateAndUpdateSynergyScores recalculates synergy scores for all combinations.
func (r *MLSuggestionRepository) CalculateAndUpdateSynergyScores(ctx context.Context, minGames int) error {
	// Synergy score formula:
	// synergy = (win_rate_together - avg(win_rate_separate)) * confidence
	// confidence = 1 - 1/sqrt(sample_size)

	query := `
		UPDATE card_combination_stats
		SET synergy_score = CASE
			WHEN games_together = 0 THEN 0
			ELSE (
				(CAST(wins_together AS REAL) / games_together) -
				(
					CASE WHEN games_card1_only > 0 THEN CAST(wins_card1_only AS REAL) / games_card1_only ELSE 0.5 END +
					CASE WHEN games_card2_only > 0 THEN CAST(wins_card2_only AS REAL) / games_card2_only ELSE 0.5 END
				) / 2.0
			) * (1.0 - 1.0 / (1.0 + SQRT(games_together)))
		END,
		confidence_score = 1.0 - 1.0 / (1.0 + SQRT(games_together)),
		updated_at = ?
		WHERE games_together >= ?
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), minGames)
	return err
}

// ============================================================================
// ML Suggestions
// ============================================================================

// CreateSuggestion creates a new ML suggestion.
func (r *MLSuggestionRepository) CreateSuggestion(ctx context.Context, suggestion *models.MLSuggestion) error {
	query := `
		INSERT INTO ml_suggestions (
			deck_id, suggestion_type, card_id, card_name,
			swap_for_card_id, swap_for_card_name,
			confidence, expected_win_rate_change,
			title, description, reasoning, evidence
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		suggestion.DeckID, suggestion.SuggestionType,
		suggestion.CardID, suggestion.CardName,
		suggestion.SwapForCardID, suggestion.SwapForCardName,
		suggestion.Confidence, suggestion.ExpectedWinRateChange,
		suggestion.Title, suggestion.Description,
		suggestion.Reasoning, suggestion.Evidence,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	suggestion.ID = id
	return nil
}

// GetSuggestionsByDeck returns all suggestions for a deck.
func (r *MLSuggestionRepository) GetSuggestionsByDeck(ctx context.Context, deckID string) ([]*models.MLSuggestion, error) {
	query := `
		SELECT id, deck_id, suggestion_type, card_id, card_name,
			swap_for_card_id, swap_for_card_name,
			confidence, expected_win_rate_change,
			title, description, reasoning, evidence,
			is_dismissed, was_applied, outcome_win_rate_change,
			created_at, applied_at, outcome_recorded_at
		FROM ml_suggestions
		WHERE deck_id = ?
		ORDER BY confidence DESC, created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, deckID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanSuggestions(rows)
}

// GetActiveSuggestions returns non-dismissed suggestions for a deck.
func (r *MLSuggestionRepository) GetActiveSuggestions(ctx context.Context, deckID string) ([]*models.MLSuggestion, error) {
	query := `
		SELECT id, deck_id, suggestion_type, card_id, card_name,
			swap_for_card_id, swap_for_card_name,
			confidence, expected_win_rate_change,
			title, description, reasoning, evidence,
			is_dismissed, was_applied, outcome_win_rate_change,
			created_at, applied_at, outcome_recorded_at
		FROM ml_suggestions
		WHERE deck_id = ? AND is_dismissed = FALSE AND was_applied = FALSE
		ORDER BY confidence DESC
	`

	rows, err := r.db.QueryContext(ctx, query, deckID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return r.scanSuggestions(rows)
}

// DismissSuggestion marks a suggestion as dismissed.
func (r *MLSuggestionRepository) DismissSuggestion(ctx context.Context, id int64) error {
	query := `UPDATE ml_suggestions SET is_dismissed = TRUE WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// ApplySuggestion marks a suggestion as applied.
func (r *MLSuggestionRepository) ApplySuggestion(ctx context.Context, id int64) error {
	query := `UPDATE ml_suggestions SET was_applied = TRUE, applied_at = ? WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, time.Now(), id)
	return err
}

// RecordSuggestionOutcome records the actual win rate change after applying a suggestion.
func (r *MLSuggestionRepository) RecordSuggestionOutcome(ctx context.Context, id int64, winRateChange float64) error {
	query := `
		UPDATE ml_suggestions
		SET outcome_win_rate_change = ?, outcome_recorded_at = ?
		WHERE id = ?
	`
	_, err := r.db.ExecContext(ctx, query, winRateChange, time.Now(), id)
	return err
}

// DeleteSuggestionsByDeck removes all suggestions for a deck.
func (r *MLSuggestionRepository) DeleteSuggestionsByDeck(ctx context.Context, deckID string) error {
	query := `DELETE FROM ml_suggestions WHERE deck_id = ?`
	_, err := r.db.ExecContext(ctx, query, deckID)
	return err
}

func (r *MLSuggestionRepository) scanSuggestions(rows *sql.Rows) ([]*models.MLSuggestion, error) {
	var results []*models.MLSuggestion
	for rows.Next() {
		var s models.MLSuggestion
		var cardID, swapForCardID sql.NullInt64
		var cardName, swapForCardName, description, reasoning, evidence sql.NullString
		var outcomeWinRateChange sql.NullFloat64
		var appliedAt, outcomeRecordedAt sql.NullTime

		if err := rows.Scan(
			&s.ID, &s.DeckID, &s.SuggestionType, &cardID, &cardName,
			&swapForCardID, &swapForCardName,
			&s.Confidence, &s.ExpectedWinRateChange,
			&s.Title, &description, &reasoning, &evidence,
			&s.IsDismissed, &s.WasApplied, &outcomeWinRateChange,
			&s.CreatedAt, &appliedAt, &outcomeRecordedAt,
		); err != nil {
			return nil, err
		}

		if cardID.Valid {
			s.CardID = int(cardID.Int64)
		}
		if cardName.Valid {
			s.CardName = cardName.String
		}
		if swapForCardID.Valid {
			s.SwapForCardID = int(swapForCardID.Int64)
		}
		if swapForCardName.Valid {
			s.SwapForCardName = swapForCardName.String
		}
		if description.Valid {
			s.Description = description.String
		}
		if reasoning.Valid {
			s.Reasoning = reasoning.String
		}
		if evidence.Valid {
			s.Evidence = evidence.String
		}
		if outcomeWinRateChange.Valid {
			s.OutcomeWinRateChange = &outcomeWinRateChange.Float64
		}
		if appliedAt.Valid {
			s.AppliedAt = &appliedAt.Time
		}
		if outcomeRecordedAt.Valid {
			s.OutcomeRecordedAt = &outcomeRecordedAt.Time
		}

		results = append(results, &s)
	}
	return results, rows.Err()
}

// ============================================================================
// Card Affinity
// ============================================================================

// UpsertCardAffinity inserts or updates card affinity score.
func (r *MLSuggestionRepository) UpsertCardAffinity(ctx context.Context, affinity *models.CardAffinity) error {
	// Ensure card_id_1 < card_id_2
	if affinity.CardID1 > affinity.CardID2 {
		affinity.CardID1, affinity.CardID2 = affinity.CardID2, affinity.CardID1
	}

	query := `
		INSERT INTO card_affinity (
			card_id_1, card_id_2, format,
			affinity_score, sample_size, confidence, source, computed_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(card_id_1, card_id_2, format) DO UPDATE SET
			affinity_score = excluded.affinity_score,
			sample_size = excluded.sample_size,
			confidence = excluded.confidence,
			source = excluded.source,
			computed_at = excluded.computed_at
	`

	_, err := r.db.ExecContext(ctx, query,
		affinity.CardID1, affinity.CardID2, affinity.Format,
		affinity.AffinityScore, affinity.SampleSize, affinity.Confidence,
		affinity.Source, time.Now(),
	)
	return err
}

// GetCardAffinity retrieves affinity between two cards.
func (r *MLSuggestionRepository) GetCardAffinity(ctx context.Context, cardID1, cardID2 int, format string) (*models.CardAffinity, error) {
	if cardID1 > cardID2 {
		cardID1, cardID2 = cardID2, cardID1
	}

	query := `
		SELECT id, card_id_1, card_id_2, format,
			affinity_score, sample_size, confidence, source, computed_at
		FROM card_affinity
		WHERE card_id_1 = ? AND card_id_2 = ? AND format = ?
	`

	var a models.CardAffinity
	err := r.db.QueryRowContext(ctx, query, cardID1, cardID2, format).Scan(
		&a.ID, &a.CardID1, &a.CardID2, &a.Format,
		&a.AffinityScore, &a.SampleSize, &a.Confidence, &a.Source, &a.ComputedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// GetTopAffinities returns the top affinity cards for a given card.
func (r *MLSuggestionRepository) GetTopAffinities(ctx context.Context, cardID int, format string, limit int) ([]*models.CardAffinity, error) {
	query := `
		SELECT id, card_id_1, card_id_2, format,
			affinity_score, sample_size, confidence, source, computed_at
		FROM card_affinity
		WHERE (card_id_1 = ? OR card_id_2 = ?) AND format = ?
		ORDER BY affinity_score DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, cardID, cardID, format, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []*models.CardAffinity
	for rows.Next() {
		var a models.CardAffinity
		if err := rows.Scan(
			&a.ID, &a.CardID1, &a.CardID2, &a.Format,
			&a.AffinityScore, &a.SampleSize, &a.Confidence, &a.Source, &a.ComputedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, &a)
	}
	return results, rows.Err()
}

// ============================================================================
// User Play Patterns
// ============================================================================

// UpsertUserPlayPatterns inserts or updates user play patterns.
func (r *MLSuggestionRepository) UpsertUserPlayPatterns(ctx context.Context, patterns *models.UserPlayPatterns) error {
	query := `
		INSERT INTO user_play_patterns (
			account_id, preferred_archetype,
			aggro_affinity, midrange_affinity, control_affinity, combo_affinity,
			color_preferences, avg_game_length, aggression_score, interaction_score,
			total_matches, total_decks, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(account_id) DO UPDATE SET
			preferred_archetype = excluded.preferred_archetype,
			aggro_affinity = excluded.aggro_affinity,
			midrange_affinity = excluded.midrange_affinity,
			control_affinity = excluded.control_affinity,
			combo_affinity = excluded.combo_affinity,
			color_preferences = excluded.color_preferences,
			avg_game_length = excluded.avg_game_length,
			aggression_score = excluded.aggression_score,
			interaction_score = excluded.interaction_score,
			total_matches = excluded.total_matches,
			total_decks = excluded.total_decks,
			updated_at = excluded.updated_at
	`

	_, err := r.db.ExecContext(ctx, query,
		patterns.AccountID, patterns.PreferredArchetype,
		patterns.AggroAffinity, patterns.MidrangeAffinity,
		patterns.ControlAffinity, patterns.ComboAffinity,
		patterns.ColorPreferences, patterns.AvgGameLength,
		patterns.AggressionScore, patterns.InteractionScore,
		patterns.TotalMatches, patterns.TotalDecks, time.Now(),
	)
	return err
}

// GetUserPlayPatterns retrieves play patterns for a user.
func (r *MLSuggestionRepository) GetUserPlayPatterns(ctx context.Context, accountID string) (*models.UserPlayPatterns, error) {
	query := `
		SELECT id, account_id, preferred_archetype,
			aggro_affinity, midrange_affinity, control_affinity, combo_affinity,
			color_preferences, avg_game_length, aggression_score, interaction_score,
			total_matches, total_decks, created_at, updated_at
		FROM user_play_patterns
		WHERE account_id = ?
	`

	var p models.UserPlayPatterns
	var preferredArchetype, colorPrefs sql.NullString
	err := r.db.QueryRowContext(ctx, query, accountID).Scan(
		&p.ID, &p.AccountID, &preferredArchetype,
		&p.AggroAffinity, &p.MidrangeAffinity, &p.ControlAffinity, &p.ComboAffinity,
		&colorPrefs, &p.AvgGameLength, &p.AggressionScore, &p.InteractionScore,
		&p.TotalMatches, &p.TotalDecks, &p.CreatedAt, &p.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if preferredArchetype.Valid {
		p.PreferredArchetype = preferredArchetype.String
	}
	if colorPrefs.Valid {
		p.ColorPreferences = colorPrefs.String
	}
	return &p, nil
}

// ============================================================================
// ML Model Metadata
// ============================================================================

// SaveModelMetadata saves model metadata.
func (r *MLSuggestionRepository) SaveModelMetadata(ctx context.Context, meta *models.MLModelMetadata) error {
	query := `
		INSERT INTO ml_model_metadata (
			model_name, model_version, training_samples, training_date,
			accuracy, precision_score, recall, f1_score, is_active, model_data
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(model_name, model_version) DO UPDATE SET
			training_samples = excluded.training_samples,
			training_date = excluded.training_date,
			accuracy = excluded.accuracy,
			precision_score = excluded.precision_score,
			recall = excluded.recall,
			f1_score = excluded.f1_score,
			is_active = excluded.is_active,
			model_data = excluded.model_data
	`

	_, err := r.db.ExecContext(ctx, query,
		meta.ModelName, meta.ModelVersion, meta.TrainingSamples, meta.TrainingDate,
		meta.Accuracy, meta.PrecisionScore, meta.Recall, meta.F1Score,
		meta.IsActive, meta.ModelData,
	)
	if err != nil {
		return err
	}

	// Query for the actual ID since LastInsertId doesn't work with ON CONFLICT DO UPDATE
	var id int64
	err = r.db.QueryRowContext(ctx,
		"SELECT id FROM ml_model_metadata WHERE model_name = ? AND model_version = ?",
		meta.ModelName, meta.ModelVersion,
	).Scan(&id)
	if err != nil {
		return err
	}
	meta.ID = id
	return nil
}

// GetActiveModel returns the active model for a given name.
func (r *MLSuggestionRepository) GetActiveModel(ctx context.Context, modelName string) (*models.MLModelMetadata, error) {
	query := `
		SELECT id, model_name, model_version, training_samples, training_date,
			accuracy, precision_score, recall, f1_score, is_active, model_data, created_at
		FROM ml_model_metadata
		WHERE model_name = ? AND is_active = TRUE
		ORDER BY created_at DESC
		LIMIT 1
	`

	var m models.MLModelMetadata
	var trainingDate sql.NullTime
	var accuracy, precision, recall, f1 sql.NullFloat64
	err := r.db.QueryRowContext(ctx, query, modelName).Scan(
		&m.ID, &m.ModelName, &m.ModelVersion, &m.TrainingSamples, &trainingDate,
		&accuracy, &precision, &recall, &f1, &m.IsActive, &m.ModelData, &m.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if trainingDate.Valid {
		m.TrainingDate = &trainingDate.Time
	}
	if accuracy.Valid {
		m.Accuracy = &accuracy.Float64
	}
	if precision.Valid {
		m.PrecisionScore = &precision.Float64
	}
	if recall.Valid {
		m.Recall = &recall.Float64
	}
	if f1.Valid {
		m.F1Score = &f1.Float64
	}
	return &m, nil
}

// ============================================================================
// Analysis Helpers
// ============================================================================

// CalculateSynergyScore calculates the synergy score between two cards.
func CalculateSynergyScore(stats *models.CardCombinationStats) float64 {
	if stats.GamesTogether < 5 {
		return 0 // Not enough data
	}

	winRateTogether := stats.WinRateTogether()
	winRateSeparate := (stats.WinRateCard1Only() + stats.WinRateCard2Only()) / 2.0

	// Synergy is the improvement in win rate when cards are together
	rawSynergy := winRateTogether - winRateSeparate

	// Apply confidence weighting based on sample size
	confidence := 1.0 - 1.0/(1.0+math.Sqrt(float64(stats.GamesTogether)))

	return rawSynergy * confidence
}

// CalculateConfidenceScore calculates confidence based on sample size.
func CalculateConfidenceScore(sampleSize int) float64 {
	if sampleSize == 0 {
		return 0
	}
	return 1.0 - 1.0/(1.0+math.Sqrt(float64(sampleSize)))
}

// GetPairedCardID returns the other card ID in a combination stat.
func GetPairedCardID(stats *models.CardCombinationStats, cardID int) int {
	if stats.CardID1 == cardID {
		return stats.CardID2
	}
	return stats.CardID1
}

// GenerateMLSuggestion creates a suggestion with reasoning.
func GenerateMLSuggestion(
	deckID string,
	suggestionType string,
	cardID int,
	cardName string,
	confidence float64,
	expectedChange float64,
	reasons []models.MLSuggestionReason,
) (*models.MLSuggestion, error) {
	suggestion := &models.MLSuggestion{
		DeckID:                deckID,
		SuggestionType:        suggestionType,
		CardID:                cardID,
		CardName:              cardName,
		Confidence:            confidence,
		ExpectedWinRateChange: expectedChange,
		CreatedAt:             time.Now(),
	}

	// Generate title based on type
	switch suggestionType {
	case models.MLSuggestionTypeAdd:
		suggestion.Title = fmt.Sprintf("Consider adding %s", cardName)
	case models.MLSuggestionTypeRemove:
		suggestion.Title = fmt.Sprintf("Consider removing %s", cardName)
	case models.MLSuggestionTypeSwap:
		suggestion.Title = fmt.Sprintf("Consider swapping %s", cardName)
	}

	// Generate description from top reasons
	if len(reasons) > 0 {
		suggestion.Description = reasons[0].Description
	}

	if err := suggestion.SetReasons(reasons); err != nil {
		return nil, err
	}

	return suggestion, nil
}
