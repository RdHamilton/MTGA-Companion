package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// DraftAnalyticsRepository provides methods for managing draft analytics data.
type DraftAnalyticsRepository interface {
	// Match results
	SaveDraftMatchResult(ctx context.Context, result *models.DraftMatchResult) error
	GetDraftMatchResults(ctx context.Context, sessionID string) ([]*models.DraftMatchResult, error)
	GetDraftMatchResultsByTimeRange(ctx context.Context, start, end time.Time) ([]*models.DraftMatchResult, error)
	GetDraftMatchResultCount(ctx context.Context) (int, error)

	// Archetype stats
	GetArchetypeStats(ctx context.Context, setCode string) ([]*models.DraftArchetypeStats, error)
	GetAllArchetypeStats(ctx context.Context) ([]*models.DraftArchetypeStats, error)
	UpsertArchetypeStats(ctx context.Context, stats *models.DraftArchetypeStats) error
	GetBestArchetypes(ctx context.Context, minMatches, limit int) ([]*models.DraftArchetypeStats, error)
	GetWorstArchetypes(ctx context.Context, minMatches, limit int) ([]*models.DraftArchetypeStats, error)

	// Temporal trends
	SaveTemporalTrend(ctx context.Context, trend *models.DraftTemporalTrend) error
	GetTemporalTrends(ctx context.Context, periodType string, limit int) ([]*models.DraftTemporalTrend, error)
	GetTemporalTrendsBySet(ctx context.Context, setCode, periodType string, limit int) ([]*models.DraftTemporalTrend, error)
	ClearTemporalTrends(ctx context.Context, periodType string) error

	// Pattern analysis
	SavePatternAnalysis(ctx context.Context, analysis *models.DraftPatternAnalysis) error
	GetPatternAnalysis(ctx context.Context, setCode *string) (*models.DraftPatternAnalysis, error)

	// Community comparison
	SaveCommunityComparison(ctx context.Context, comparison *models.DraftCommunityComparison) error
	GetCommunityComparison(ctx context.Context, setCode, draftFormat string) (*models.DraftCommunityComparison, error)
	GetAllCommunityComparisons(ctx context.Context) ([]*models.DraftCommunityComparison, error)
}

type draftAnalyticsRepository struct {
	db *sql.DB
}

// NewDraftAnalyticsRepository creates a new draft analytics repository.
func NewDraftAnalyticsRepository(db *sql.DB) DraftAnalyticsRepository {
	return &draftAnalyticsRepository{db: db}
}

// SaveDraftMatchResult saves a draft match result.
func (r *draftAnalyticsRepository) SaveDraftMatchResult(ctx context.Context, result *models.DraftMatchResult) error {
	query := `
		INSERT OR REPLACE INTO draft_match_results (session_id, match_id, result, opponent_colors, game_wins, game_losses, match_timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		result.SessionID,
		result.MatchID,
		result.Result,
		result.OpponentColors,
		result.GameWins,
		result.GameLosses,
		result.MatchTimestamp,
	)
	return err
}

// GetDraftMatchResults retrieves all match results for a draft session.
func (r *draftAnalyticsRepository) GetDraftMatchResults(ctx context.Context, sessionID string) ([]*models.DraftMatchResult, error) {
	query := `
		SELECT id, session_id, match_id, result, opponent_colors, game_wins, game_losses, match_timestamp
		FROM draft_match_results
		WHERE session_id = ?
		ORDER BY match_timestamp ASC
	`
	rows, err := r.db.QueryContext(ctx, query, sessionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanDraftMatchResults(rows)
}

// GetDraftMatchResultsByTimeRange retrieves all match results within a time range.
func (r *draftAnalyticsRepository) GetDraftMatchResultsByTimeRange(ctx context.Context, start, end time.Time) ([]*models.DraftMatchResult, error) {
	query := `
		SELECT id, session_id, match_id, result, opponent_colors, game_wins, game_losses, match_timestamp
		FROM draft_match_results
		WHERE match_timestamp >= ? AND match_timestamp <= ?
		ORDER BY match_timestamp ASC
	`
	rows, err := r.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanDraftMatchResults(rows)
}

// GetDraftMatchResultCount returns the total number of draft match results.
func (r *draftAnalyticsRepository) GetDraftMatchResultCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM draft_match_results").Scan(&count)
	return count, err
}

func scanDraftMatchResults(rows *sql.Rows) ([]*models.DraftMatchResult, error) {
	var results []*models.DraftMatchResult
	for rows.Next() {
		result := &models.DraftMatchResult{}
		var opponentColors sql.NullString
		err := rows.Scan(
			&result.ID,
			&result.SessionID,
			&result.MatchID,
			&result.Result,
			&opponentColors,
			&result.GameWins,
			&result.GameLosses,
			&result.MatchTimestamp,
		)
		if err != nil {
			return nil, err
		}
		if opponentColors.Valid {
			result.OpponentColors = opponentColors.String
		}
		results = append(results, result)
	}
	return results, rows.Err()
}

// GetArchetypeStats retrieves archetype statistics for a specific set.
func (r *draftAnalyticsRepository) GetArchetypeStats(ctx context.Context, setCode string) ([]*models.DraftArchetypeStats, error) {
	query := `
		SELECT id, set_code, color_combination, archetype_name, matches_played, matches_won,
			drafts_count, avg_draft_grade, last_played_at, updated_at
		FROM draft_archetype_stats
		WHERE set_code = ?
		ORDER BY matches_won * 1.0 / CASE WHEN matches_played = 0 THEN 1 ELSE matches_played END DESC
	`
	rows, err := r.db.QueryContext(ctx, query, setCode)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanArchetypeStats(rows)
}

// GetAllArchetypeStats retrieves all archetype statistics.
func (r *draftAnalyticsRepository) GetAllArchetypeStats(ctx context.Context) ([]*models.DraftArchetypeStats, error) {
	query := `
		SELECT id, set_code, color_combination, archetype_name, matches_played, matches_won,
			drafts_count, avg_draft_grade, last_played_at, updated_at
		FROM draft_archetype_stats
		ORDER BY set_code, matches_won * 1.0 / CASE WHEN matches_played = 0 THEN 1 ELSE matches_played END DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanArchetypeStats(rows)
}

// UpsertArchetypeStats inserts or updates archetype statistics.
func (r *draftAnalyticsRepository) UpsertArchetypeStats(ctx context.Context, stats *models.DraftArchetypeStats) error {
	query := `
		INSERT INTO draft_archetype_stats (set_code, color_combination, archetype_name, matches_played, matches_won,
			drafts_count, avg_draft_grade, last_played_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(set_code, color_combination) DO UPDATE SET
			archetype_name = excluded.archetype_name,
			matches_played = excluded.matches_played,
			matches_won = excluded.matches_won,
			drafts_count = excluded.drafts_count,
			avg_draft_grade = excluded.avg_draft_grade,
			last_played_at = excluded.last_played_at,
			updated_at = excluded.updated_at
	`
	_, err := r.db.ExecContext(ctx, query,
		stats.SetCode,
		stats.ColorCombination,
		stats.ArchetypeName,
		stats.MatchesPlayed,
		stats.MatchesWon,
		stats.DraftsCount,
		stats.AvgDraftGrade,
		stats.LastPlayedAt,
		stats.UpdatedAt,
	)
	return err
}

// GetBestArchetypes retrieves the best performing archetypes by win rate.
func (r *draftAnalyticsRepository) GetBestArchetypes(ctx context.Context, minMatches, limit int) ([]*models.DraftArchetypeStats, error) {
	query := `
		SELECT id, set_code, color_combination, archetype_name, matches_played, matches_won,
			drafts_count, avg_draft_grade, last_played_at, updated_at
		FROM draft_archetype_stats
		WHERE matches_played >= ?
		ORDER BY matches_won * 1.0 / matches_played DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, minMatches, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanArchetypeStats(rows)
}

// GetWorstArchetypes retrieves the worst performing archetypes by win rate.
func (r *draftAnalyticsRepository) GetWorstArchetypes(ctx context.Context, minMatches, limit int) ([]*models.DraftArchetypeStats, error) {
	query := `
		SELECT id, set_code, color_combination, archetype_name, matches_played, matches_won,
			drafts_count, avg_draft_grade, last_played_at, updated_at
		FROM draft_archetype_stats
		WHERE matches_played >= ?
		ORDER BY matches_won * 1.0 / matches_played ASC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, minMatches, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanArchetypeStats(rows)
}

func scanArchetypeStats(rows *sql.Rows) ([]*models.DraftArchetypeStats, error) {
	var stats []*models.DraftArchetypeStats
	for rows.Next() {
		s := &models.DraftArchetypeStats{}
		var avgGrade sql.NullFloat64
		var lastPlayed sql.NullTime
		err := rows.Scan(
			&s.ID,
			&s.SetCode,
			&s.ColorCombination,
			&s.ArchetypeName,
			&s.MatchesPlayed,
			&s.MatchesWon,
			&s.DraftsCount,
			&avgGrade,
			&lastPlayed,
			&s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		if avgGrade.Valid {
			s.AvgDraftGrade = &avgGrade.Float64
		}
		if lastPlayed.Valid {
			s.LastPlayedAt = &lastPlayed.Time
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}

// SaveTemporalTrend saves a temporal trend record.
func (r *draftAnalyticsRepository) SaveTemporalTrend(ctx context.Context, trend *models.DraftTemporalTrend) error {
	query := `
		INSERT OR REPLACE INTO draft_temporal_trends (period_type, period_start, period_end, set_code,
			drafts_count, matches_played, matches_won, avg_draft_grade, calculated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		trend.PeriodType,
		trend.PeriodStart,
		trend.PeriodEnd,
		trend.SetCode,
		trend.DraftsCount,
		trend.MatchesPlayed,
		trend.MatchesWon,
		trend.AvgDraftGrade,
		trend.CalculatedAt,
	)
	return err
}

// GetTemporalTrends retrieves temporal trends for a period type.
func (r *draftAnalyticsRepository) GetTemporalTrends(ctx context.Context, periodType string, limit int) ([]*models.DraftTemporalTrend, error) {
	query := `
		SELECT id, period_type, period_start, period_end, set_code, drafts_count,
			matches_played, matches_won, avg_draft_grade, calculated_at
		FROM draft_temporal_trends
		WHERE period_type = ? AND set_code IS NULL
		ORDER BY period_start DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, periodType, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanTemporalTrends(rows)
}

// GetTemporalTrendsBySet retrieves temporal trends for a specific set.
func (r *draftAnalyticsRepository) GetTemporalTrendsBySet(ctx context.Context, setCode, periodType string, limit int) ([]*models.DraftTemporalTrend, error) {
	query := `
		SELECT id, period_type, period_start, period_end, set_code, drafts_count,
			matches_played, matches_won, avg_draft_grade, calculated_at
		FROM draft_temporal_trends
		WHERE period_type = ? AND set_code = ?
		ORDER BY period_start DESC
		LIMIT ?
	`
	rows, err := r.db.QueryContext(ctx, query, periodType, setCode, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	return scanTemporalTrends(rows)
}

// ClearTemporalTrends clears all temporal trends of a specific type.
func (r *draftAnalyticsRepository) ClearTemporalTrends(ctx context.Context, periodType string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM draft_temporal_trends WHERE period_type = ?", periodType)
	return err
}

func scanTemporalTrends(rows *sql.Rows) ([]*models.DraftTemporalTrend, error) {
	var trends []*models.DraftTemporalTrend
	for rows.Next() {
		t := &models.DraftTemporalTrend{}
		var setCode sql.NullString
		var avgGrade sql.NullFloat64
		err := rows.Scan(
			&t.ID,
			&t.PeriodType,
			&t.PeriodStart,
			&t.PeriodEnd,
			&setCode,
			&t.DraftsCount,
			&t.MatchesPlayed,
			&t.MatchesWon,
			&avgGrade,
			&t.CalculatedAt,
		)
		if err != nil {
			return nil, err
		}
		if setCode.Valid {
			t.SetCode = &setCode.String
		}
		if avgGrade.Valid {
			t.AvgDraftGrade = &avgGrade.Float64
		}
		trends = append(trends, t)
	}
	return trends, rows.Err()
}

// SavePatternAnalysis saves a pattern analysis record.
func (r *draftAnalyticsRepository) SavePatternAnalysis(ctx context.Context, analysis *models.DraftPatternAnalysis) error {
	query := `
		INSERT OR REPLACE INTO draft_pattern_analysis (set_code, color_preference_json, type_preference_json,
			pick_order_pattern_json, archetype_affinity_json, sample_size, calculated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		analysis.SetCode,
		analysis.ColorPreferenceJSON,
		analysis.TypePreferenceJSON,
		analysis.PickOrderPatternJSON,
		analysis.ArchetypeAffinityJSON,
		analysis.SampleSize,
		analysis.CalculatedAt,
	)
	return err
}

// GetPatternAnalysis retrieves pattern analysis for a set (or overall if setCode is nil).
func (r *draftAnalyticsRepository) GetPatternAnalysis(ctx context.Context, setCode *string) (*models.DraftPatternAnalysis, error) {
	var query string
	var args []interface{}
	if setCode != nil {
		query = `
			SELECT id, set_code, color_preference_json, type_preference_json,
				pick_order_pattern_json, archetype_affinity_json, sample_size, calculated_at
			FROM draft_pattern_analysis
			WHERE set_code = ?
		`
		args = append(args, *setCode)
	} else {
		query = `
			SELECT id, set_code, color_preference_json, type_preference_json,
				pick_order_pattern_json, archetype_affinity_json, sample_size, calculated_at
			FROM draft_pattern_analysis
			WHERE set_code IS NULL
		`
	}
	row := r.db.QueryRowContext(ctx, query, args...)

	analysis := &models.DraftPatternAnalysis{}
	var setCodeVal sql.NullString
	var colorPref, typePref, pickOrder, archAffinity sql.NullString
	err := row.Scan(
		&analysis.ID,
		&setCodeVal,
		&colorPref,
		&typePref,
		&pickOrder,
		&archAffinity,
		&analysis.SampleSize,
		&analysis.CalculatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if setCodeVal.Valid {
		analysis.SetCode = &setCodeVal.String
	}
	if colorPref.Valid {
		analysis.ColorPreferenceJSON = colorPref.String
	}
	if typePref.Valid {
		analysis.TypePreferenceJSON = typePref.String
	}
	if pickOrder.Valid {
		analysis.PickOrderPatternJSON = pickOrder.String
	}
	if archAffinity.Valid {
		analysis.ArchetypeAffinityJSON = archAffinity.String
	}

	return analysis, nil
}

// SaveCommunityComparison saves a community comparison record.
func (r *draftAnalyticsRepository) SaveCommunityComparison(ctx context.Context, comparison *models.DraftCommunityComparison) error {
	query := `
		INSERT OR REPLACE INTO draft_community_comparison (set_code, draft_format, user_win_rate,
			community_avg_win_rate, percentile_rank, sample_size, calculated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.ExecContext(ctx, query,
		comparison.SetCode,
		comparison.DraftFormat,
		comparison.UserWinRate,
		comparison.CommunityAvgWinRate,
		comparison.PercentileRank,
		comparison.SampleSize,
		comparison.CalculatedAt,
	)
	return err
}

// GetCommunityComparison retrieves community comparison for a set and format.
func (r *draftAnalyticsRepository) GetCommunityComparison(ctx context.Context, setCode, draftFormat string) (*models.DraftCommunityComparison, error) {
	query := `
		SELECT id, set_code, draft_format, user_win_rate, community_avg_win_rate,
			percentile_rank, sample_size, calculated_at
		FROM draft_community_comparison
		WHERE set_code = ? AND draft_format = ?
	`
	row := r.db.QueryRowContext(ctx, query, setCode, draftFormat)

	comparison := &models.DraftCommunityComparison{}
	var percentile sql.NullFloat64
	err := row.Scan(
		&comparison.ID,
		&comparison.SetCode,
		&comparison.DraftFormat,
		&comparison.UserWinRate,
		&comparison.CommunityAvgWinRate,
		&percentile,
		&comparison.SampleSize,
		&comparison.CalculatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if percentile.Valid {
		comparison.PercentileRank = &percentile.Float64
	}

	return comparison, nil
}

// GetAllCommunityComparisons retrieves all community comparison records.
func (r *draftAnalyticsRepository) GetAllCommunityComparisons(ctx context.Context) ([]*models.DraftCommunityComparison, error) {
	query := `
		SELECT id, set_code, draft_format, user_win_rate, community_avg_win_rate,
			percentile_rank, sample_size, calculated_at
		FROM draft_community_comparison
		ORDER BY calculated_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var comparisons []*models.DraftCommunityComparison
	for rows.Next() {
		c := &models.DraftCommunityComparison{}
		var percentile sql.NullFloat64
		err := rows.Scan(
			&c.ID,
			&c.SetCode,
			&c.DraftFormat,
			&c.UserWinRate,
			&c.CommunityAvgWinRate,
			&percentile,
			&c.SampleSize,
			&c.CalculatedAt,
		)
		if err != nil {
			return nil, err
		}
		if percentile.Valid {
			c.PercentileRank = &percentile.Float64
		}
		comparisons = append(comparisons, c)
	}

	return comparisons, rows.Err()
}
