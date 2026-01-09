package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// OpponentRepository handles database operations for opponent deck analysis.
type OpponentRepository interface {
	// Opponent deck profiles
	CreateOrUpdateProfile(ctx context.Context, profile *models.OpponentDeckProfile) error
	GetProfileByMatchID(ctx context.Context, matchID string) (*models.OpponentDeckProfile, error)
	ListProfiles(ctx context.Context, filter *OpponentProfileFilter) ([]*models.OpponentDeckProfile, error)
	DeleteProfile(ctx context.Context, matchID string) error

	// Matchup statistics
	RecordMatchup(ctx context.Context, stat *models.MatchupStatistic) error
	GetMatchupStats(ctx context.Context, accountID int, playerArchetype, opponentArchetype, format string) (*models.MatchupStatistic, error)
	ListMatchupStats(ctx context.Context, accountID int, format *string) ([]*models.MatchupStatistic, error)
	GetTopMatchups(ctx context.Context, accountID int, format string, limit int) ([]*models.MatchupStatistic, error)

	// Expected cards
	UpsertExpectedCard(ctx context.Context, card *models.ArchetypeExpectedCard) error
	GetExpectedCards(ctx context.Context, archetypeName, format string) ([]*models.ArchetypeExpectedCard, error)
	DeleteExpectedCards(ctx context.Context, archetypeName, format string) error

	// Opponent history summary
	GetOpponentHistorySummary(ctx context.Context, accountID int, format *string) (*models.OpponentHistorySummary, error)
}

// OpponentProfileFilter provides filtering options for opponent profiles.
type OpponentProfileFilter struct {
	Archetype     *string
	Format        *string
	MinConfidence *float64
	Limit         int
	Offset        int
}

// opponentRepository implements OpponentRepository.
type opponentRepository struct {
	db *sql.DB
}

// NewOpponentRepository creates a new opponent repository.
func NewOpponentRepository(db *sql.DB) OpponentRepository {
	return &opponentRepository{db: db}
}

// CreateOrUpdateProfile creates or updates an opponent deck profile.
func (r *opponentRepository) CreateOrUpdateProfile(ctx context.Context, profile *models.OpponentDeckProfile) error {
	query := `
		INSERT INTO opponent_deck_profiles (
			match_id, detected_archetype, archetype_confidence, color_identity,
			deck_style, cards_observed, estimated_deck_size, observed_card_ids,
			inferred_card_ids, signature_cards, format, meta_archetype_id,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(match_id) DO UPDATE SET
			detected_archetype = excluded.detected_archetype,
			archetype_confidence = excluded.archetype_confidence,
			color_identity = excluded.color_identity,
			deck_style = excluded.deck_style,
			cards_observed = excluded.cards_observed,
			estimated_deck_size = excluded.estimated_deck_size,
			observed_card_ids = excluded.observed_card_ids,
			inferred_card_ids = excluded.inferred_card_ids,
			signature_cards = excluded.signature_cards,
			format = excluded.format,
			meta_archetype_id = excluded.meta_archetype_id,
			updated_at = excluded.updated_at
	`

	now := time.Now()
	if profile.CreatedAt.IsZero() {
		profile.CreatedAt = now
	}
	profile.UpdatedAt = now

	result, err := r.db.ExecContext(ctx, query,
		profile.MatchID, profile.DetectedArchetype, profile.ArchetypeConfidence,
		profile.ColorIdentity, profile.DeckStyle, profile.CardsObserved,
		profile.EstimatedDeckSize, profile.ObservedCardIDs, profile.InferredCardIDs,
		profile.SignatureCards, profile.Format, profile.MetaArchetypeID,
		profile.CreatedAt, profile.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create/update opponent profile: %w", err)
	}

	if profile.ID == 0 {
		id, err := result.LastInsertId()
		if err == nil {
			profile.ID = int(id)
		}
	}

	return nil
}

// GetProfileByMatchID retrieves an opponent profile by match ID.
func (r *opponentRepository) GetProfileByMatchID(ctx context.Context, matchID string) (*models.OpponentDeckProfile, error) {
	query := `
		SELECT id, match_id, detected_archetype, archetype_confidence, color_identity,
			deck_style, cards_observed, estimated_deck_size, observed_card_ids,
			inferred_card_ids, signature_cards, format, meta_archetype_id,
			created_at, updated_at
		FROM opponent_deck_profiles
		WHERE match_id = ?
	`

	var profile models.OpponentDeckProfile
	err := r.db.QueryRowContext(ctx, query, matchID).Scan(
		&profile.ID, &profile.MatchID, &profile.DetectedArchetype,
		&profile.ArchetypeConfidence, &profile.ColorIdentity, &profile.DeckStyle,
		&profile.CardsObserved, &profile.EstimatedDeckSize, &profile.ObservedCardIDs,
		&profile.InferredCardIDs, &profile.SignatureCards, &profile.Format,
		&profile.MetaArchetypeID, &profile.CreatedAt, &profile.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get opponent profile: %w", err)
	}

	return &profile, nil
}

// ListProfiles lists opponent profiles with filtering.
func (r *opponentRepository) ListProfiles(ctx context.Context, filter *OpponentProfileFilter) ([]*models.OpponentDeckProfile, error) {
	query := `
		SELECT id, match_id, detected_archetype, archetype_confidence, color_identity,
			deck_style, cards_observed, estimated_deck_size, observed_card_ids,
			inferred_card_ids, signature_cards, format, meta_archetype_id,
			created_at, updated_at
		FROM opponent_deck_profiles
		WHERE 1=1
	`
	args := make([]interface{}, 0)

	if filter != nil {
		if filter.Archetype != nil {
			query += " AND detected_archetype = ?"
			args = append(args, *filter.Archetype)
		}
		if filter.Format != nil {
			query += " AND format = ?"
			args = append(args, *filter.Format)
		}
		if filter.MinConfidence != nil {
			query += " AND archetype_confidence >= ?"
			args = append(args, *filter.MinConfidence)
		}
	}

	query += " ORDER BY created_at DESC"

	if filter != nil && filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list opponent profiles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var profiles []*models.OpponentDeckProfile
	for rows.Next() {
		var p models.OpponentDeckProfile
		err := rows.Scan(
			&p.ID, &p.MatchID, &p.DetectedArchetype, &p.ArchetypeConfidence,
			&p.ColorIdentity, &p.DeckStyle, &p.CardsObserved, &p.EstimatedDeckSize,
			&p.ObservedCardIDs, &p.InferredCardIDs, &p.SignatureCards,
			&p.Format, &p.MetaArchetypeID, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan opponent profile: %w", err)
		}
		profiles = append(profiles, &p)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate opponent profiles: %w", err)
	}

	return profiles, nil
}

// DeleteProfile deletes an opponent profile.
func (r *opponentRepository) DeleteProfile(ctx context.Context, matchID string) error {
	query := `DELETE FROM opponent_deck_profiles WHERE match_id = ?`
	_, err := r.db.ExecContext(ctx, query, matchID)
	if err != nil {
		return fmt.Errorf("failed to delete opponent profile: %w", err)
	}
	return nil
}

// RecordMatchup records or updates matchup statistics.
func (r *opponentRepository) RecordMatchup(ctx context.Context, stat *models.MatchupStatistic) error {
	query := `
		INSERT INTO matchup_statistics (
			account_id, player_archetype, opponent_archetype, format,
			total_matches, wins, losses, avg_game_duration, last_match_at,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(account_id, player_archetype, opponent_archetype, format) DO UPDATE SET
			total_matches = matchup_statistics.total_matches + excluded.total_matches,
			wins = matchup_statistics.wins + excluded.wins,
			losses = matchup_statistics.losses + excluded.losses,
			avg_game_duration = excluded.avg_game_duration,
			last_match_at = excluded.last_match_at,
			updated_at = excluded.updated_at
	`

	now := time.Now()
	if stat.CreatedAt.IsZero() {
		stat.CreatedAt = now
	}
	stat.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		stat.AccountID, stat.PlayerArchetype, stat.OpponentArchetype, stat.Format,
		stat.TotalMatches, stat.Wins, stat.Losses, stat.AvgGameDuration,
		stat.LastMatchAt, stat.CreatedAt, stat.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to record matchup: %w", err)
	}

	return nil
}

// GetMatchupStats retrieves matchup stats for a specific matchup.
func (r *opponentRepository) GetMatchupStats(ctx context.Context, accountID int, playerArchetype, opponentArchetype, format string) (*models.MatchupStatistic, error) {
	query := `
		SELECT id, account_id, player_archetype, opponent_archetype, format,
			total_matches, wins, losses, avg_game_duration, last_match_at,
			created_at, updated_at
		FROM matchup_statistics
		WHERE account_id = ? AND player_archetype = ? AND opponent_archetype = ? AND format = ?
	`

	var stat models.MatchupStatistic
	err := r.db.QueryRowContext(ctx, query, accountID, playerArchetype, opponentArchetype, format).Scan(
		&stat.ID, &stat.AccountID, &stat.PlayerArchetype, &stat.OpponentArchetype,
		&stat.Format, &stat.TotalMatches, &stat.Wins, &stat.Losses,
		&stat.AvgGameDuration, &stat.LastMatchAt, &stat.CreatedAt, &stat.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get matchup stats: %w", err)
	}

	// Calculate win rate
	if stat.TotalMatches > 0 {
		stat.WinRate = float64(stat.Wins) / float64(stat.TotalMatches)
	}

	return &stat, nil
}

// ListMatchupStats lists all matchup stats for an account.
func (r *opponentRepository) ListMatchupStats(ctx context.Context, accountID int, format *string) ([]*models.MatchupStatistic, error) {
	query := `
		SELECT id, account_id, player_archetype, opponent_archetype, format,
			total_matches, wins, losses, avg_game_duration, last_match_at,
			created_at, updated_at
		FROM matchup_statistics
		WHERE account_id = ?
	`
	args := []interface{}{accountID}

	if format != nil {
		query += " AND format = ?"
		args = append(args, *format)
	}

	query += " ORDER BY total_matches DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list matchup stats: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var stats []*models.MatchupStatistic
	for rows.Next() {
		var s models.MatchupStatistic
		err := rows.Scan(
			&s.ID, &s.AccountID, &s.PlayerArchetype, &s.OpponentArchetype,
			&s.Format, &s.TotalMatches, &s.Wins, &s.Losses,
			&s.AvgGameDuration, &s.LastMatchAt, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan matchup stats: %w", err)
		}
		if s.TotalMatches > 0 {
			s.WinRate = float64(s.Wins) / float64(s.TotalMatches)
		}
		stats = append(stats, &s)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate matchup stats: %w", err)
	}

	return stats, nil
}

// GetTopMatchups gets the most common matchups for an account.
func (r *opponentRepository) GetTopMatchups(ctx context.Context, accountID int, format string, limit int) ([]*models.MatchupStatistic, error) {
	query := `
		SELECT id, account_id, player_archetype, opponent_archetype, format,
			total_matches, wins, losses, avg_game_duration, last_match_at,
			created_at, updated_at
		FROM matchup_statistics
		WHERE account_id = ? AND format = ?
		ORDER BY total_matches DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, accountID, format, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get top matchups: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var stats []*models.MatchupStatistic
	for rows.Next() {
		var s models.MatchupStatistic
		err := rows.Scan(
			&s.ID, &s.AccountID, &s.PlayerArchetype, &s.OpponentArchetype,
			&s.Format, &s.TotalMatches, &s.Wins, &s.Losses,
			&s.AvgGameDuration, &s.LastMatchAt, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan matchup stats: %w", err)
		}
		if s.TotalMatches > 0 {
			s.WinRate = float64(s.Wins) / float64(s.TotalMatches)
		}
		stats = append(stats, &s)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate top matchups: %w", err)
	}

	return stats, nil
}

// UpsertExpectedCard creates or updates an expected card for an archetype.
func (r *opponentRepository) UpsertExpectedCard(ctx context.Context, card *models.ArchetypeExpectedCard) error {
	query := `
		INSERT INTO archetype_expected_cards (
			archetype_name, format, card_id, card_name, inclusion_rate,
			avg_copies, is_signature, category, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(archetype_name, format, card_id) DO UPDATE SET
			card_name = excluded.card_name,
			inclusion_rate = excluded.inclusion_rate,
			avg_copies = excluded.avg_copies,
			is_signature = excluded.is_signature,
			category = excluded.category
	`

	if card.CreatedAt.IsZero() {
		card.CreatedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx, query,
		card.ArchetypeName, card.Format, card.CardID, card.CardName,
		card.InclusionRate, card.AvgCopies, card.IsSignature, card.Category,
		card.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert expected card: %w", err)
	}

	return nil
}

// GetExpectedCards retrieves expected cards for an archetype.
func (r *opponentRepository) GetExpectedCards(ctx context.Context, archetypeName, format string) ([]*models.ArchetypeExpectedCard, error) {
	query := `
		SELECT id, archetype_name, format, card_id, card_name, inclusion_rate,
			avg_copies, is_signature, category, created_at
		FROM archetype_expected_cards
		WHERE archetype_name = ? AND format = ?
		ORDER BY inclusion_rate DESC, is_signature DESC
	`

	rows, err := r.db.QueryContext(ctx, query, archetypeName, format)
	if err != nil {
		return nil, fmt.Errorf("failed to get expected cards: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var cards []*models.ArchetypeExpectedCard
	for rows.Next() {
		var c models.ArchetypeExpectedCard
		err := rows.Scan(
			&c.ID, &c.ArchetypeName, &c.Format, &c.CardID, &c.CardName,
			&c.InclusionRate, &c.AvgCopies, &c.IsSignature, &c.Category,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan expected card: %w", err)
		}
		cards = append(cards, &c)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate expected cards: %w", err)
	}

	return cards, nil
}

// DeleteExpectedCards deletes expected cards for an archetype.
func (r *opponentRepository) DeleteExpectedCards(ctx context.Context, archetypeName, format string) error {
	query := `DELETE FROM archetype_expected_cards WHERE archetype_name = ? AND format = ?`
	_, err := r.db.ExecContext(ctx, query, archetypeName, format)
	if err != nil {
		return fmt.Errorf("failed to delete expected cards: %w", err)
	}
	return nil
}

// GetOpponentHistorySummary retrieves aggregated opponent statistics.
func (r *opponentRepository) GetOpponentHistorySummary(ctx context.Context, accountID int, format *string) (*models.OpponentHistorySummary, error) {
	// Get archetype breakdown
	archetypeQuery := `
		SELECT
			COALESCE(odp.detected_archetype, 'Unknown') as archetype,
			COUNT(*) as count,
			SUM(CASE WHEN m.result = 'win' THEN 1 ELSE 0 END) as wins,
			SUM(CASE WHEN m.result = 'loss' THEN 1 ELSE 0 END) as losses
		FROM opponent_deck_profiles odp
		JOIN matches m ON m.id = odp.match_id
		WHERE m.account_id = ?
	`
	args := []interface{}{accountID}

	if format != nil {
		archetypeQuery += " AND odp.format = ?"
		args = append(args, *format)
	}

	archetypeQuery += " GROUP BY COALESCE(odp.detected_archetype, 'Unknown') ORDER BY count DESC"

	rows, err := r.db.QueryContext(ctx, archetypeQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get archetype breakdown: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var breakdown []models.ArchetypeBreakdownEntry
	totalOpponents := 0
	for rows.Next() {
		var entry models.ArchetypeBreakdownEntry
		var wins, losses int
		if err := rows.Scan(&entry.Archetype, &entry.Count, &wins, &losses); err != nil {
			return nil, fmt.Errorf("failed to scan archetype breakdown: %w", err)
		}
		totalOpponents += entry.Count
		if wins+losses > 0 {
			entry.WinRate = float64(wins) / float64(wins+losses)
		}
		breakdown = append(breakdown, entry)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate archetype breakdown: %w", err)
	}

	// Calculate percentages
	for i := range breakdown {
		if totalOpponents > 0 {
			breakdown[i].Percentage = float64(breakdown[i].Count) / float64(totalOpponents) * 100
		}
	}

	// Get color identity stats
	colorQuery := `
		SELECT
			odp.color_identity,
			COUNT(*) as count,
			SUM(CASE WHEN m.result = 'win' THEN 1 ELSE 0 END) as wins,
			SUM(CASE WHEN m.result = 'loss' THEN 1 ELSE 0 END) as losses
		FROM opponent_deck_profiles odp
		JOIN matches m ON m.id = odp.match_id
		WHERE m.account_id = ?
	`
	args = []interface{}{accountID}

	if format != nil {
		colorQuery += " AND odp.format = ?"
		args = append(args, *format)
	}

	colorQuery += " GROUP BY odp.color_identity ORDER BY count DESC"

	colorRows, err := r.db.QueryContext(ctx, colorQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get color identity stats: %w", err)
	}
	defer func() { _ = colorRows.Close() }()

	var colorStats []models.ColorIdentityStatsEntry
	for colorRows.Next() {
		var entry models.ColorIdentityStatsEntry
		var wins, losses int
		if err := colorRows.Scan(&entry.ColorIdentity, &entry.Count, &wins, &losses); err != nil {
			return nil, fmt.Errorf("failed to scan color stats: %w", err)
		}
		if totalOpponents > 0 {
			entry.Percentage = float64(entry.Count) / float64(totalOpponents) * 100
		}
		if wins+losses > 0 {
			entry.WinRate = float64(wins) / float64(wins+losses)
		}
		colorStats = append(colorStats, entry)
	}
	if err = colorRows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate color stats: %w", err)
	}

	summary := &models.OpponentHistorySummary{
		TotalOpponents:     totalOpponents,
		UniqueArchetypes:   len(breakdown),
		ArchetypeBreakdown: breakdown,
		ColorIdentityStats: colorStats,
	}

	if len(breakdown) > 0 {
		summary.MostCommonArchetype = breakdown[0].Archetype
		summary.MostCommonCount = breakdown[0].Count
	}

	return summary, nil
}

// Helper function to convert card IDs slice to JSON string.
func CardIDsToJSON(cardIDs []int) string {
	data, _ := json.Marshal(cardIDs)
	return string(data)
}

// Helper function to parse JSON string to card IDs slice.
func JSONToCardIDs(jsonStr *string) []int {
	if jsonStr == nil || *jsonStr == "" {
		return nil
	}
	var cardIDs []int
	if err := json.Unmarshal([]byte(*jsonStr), &cardIDs); err != nil {
		return nil
	}
	return cardIDs
}
