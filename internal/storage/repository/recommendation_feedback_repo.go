package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// RecommendationFeedbackRepository handles database operations for recommendation feedback.
type RecommendationFeedbackRepository interface {
	// Create records a new recommendation feedback entry
	Create(ctx context.Context, feedback *models.RecommendationFeedback) error

	// GetByID retrieves feedback by ID
	GetByID(ctx context.Context, id int) (*models.RecommendationFeedback, error)

	// GetByRecommendationID retrieves feedback by recommendation ID
	GetByRecommendationID(ctx context.Context, recommendationID string) (*models.RecommendationFeedback, error)

	// GetByAccount retrieves all feedback for an account
	GetByAccount(ctx context.Context, accountID int, limit int) ([]*models.RecommendationFeedback, error)

	// GetByType retrieves feedback filtered by recommendation type
	GetByType(ctx context.Context, accountID int, recType string, limit int) ([]*models.RecommendationFeedback, error)

	// UpdateAction updates the user action for a recommendation
	UpdateAction(ctx context.Context, id int, action string, alternateChoiceID *int) error

	// UpdateOutcome updates the match outcome for a recommendation
	UpdateOutcome(ctx context.Context, id int, matchID string, result string) error

	// GetStats calculates aggregated statistics for recommendations
	GetStats(ctx context.Context, accountID int, recType *string) (*models.RecommendationStats, error)

	// GetStatsByDateRange calculates stats within a date range
	GetStatsByDateRange(ctx context.Context, accountID int, start, end time.Time) (*models.RecommendationStats, error)

	// GetPendingFeedback retrieves recommendations without user response
	GetPendingFeedback(ctx context.Context, accountID int) ([]*models.RecommendationFeedback, error)

	// GetForMLTraining retrieves feedback suitable for ML training (with outcomes)
	GetForMLTraining(ctx context.Context, limit int) ([]*models.RecommendationFeedback, error)
}

type recommendationFeedbackRepository struct {
	db *sql.DB
}

// NewRecommendationFeedbackRepository creates a new recommendation feedback repository.
func NewRecommendationFeedbackRepository(db *sql.DB) RecommendationFeedbackRepository {
	return &recommendationFeedbackRepository{db: db}
}

// Create records a new recommendation feedback entry.
func (r *recommendationFeedbackRepository) Create(ctx context.Context, feedback *models.RecommendationFeedback) error {
	query := `
		INSERT INTO recommendation_feedback (
			account_id, recommendation_type, recommendation_id, recommended_card_id,
			recommended_archetype, context_data, action, alternate_choice_id,
			outcome_match_id, outcome_result, recommendation_score, recommendation_rank,
			recommended_at, responded_at, outcome_recorded_at, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	recommendedAtStr := feedback.RecommendedAt.UTC().Format("2006-01-02 15:04:05.999999")
	createdAtStr := time.Now().UTC().Format("2006-01-02 15:04:05.999999")

	var respondedAtStr *string
	if feedback.RespondedAt != nil {
		s := feedback.RespondedAt.UTC().Format("2006-01-02 15:04:05.999999")
		respondedAtStr = &s
	}

	var outcomeRecordedAtStr *string
	if feedback.OutcomeRecordedAt != nil {
		s := feedback.OutcomeRecordedAt.UTC().Format("2006-01-02 15:04:05.999999")
		outcomeRecordedAtStr = &s
	}

	result, err := r.db.ExecContext(ctx, query,
		feedback.AccountID, feedback.RecommendationType, feedback.RecommendationID,
		feedback.RecommendedCardID, feedback.RecommendedArchetype, feedback.ContextData,
		feedback.Action, feedback.AlternateChoiceID, feedback.OutcomeMatchID,
		feedback.OutcomeResult, feedback.RecommendationScore, feedback.RecommendationRank,
		recommendedAtStr, respondedAtStr, outcomeRecordedAtStr, createdAtStr,
	)
	if err != nil {
		return fmt.Errorf("failed to create recommendation feedback: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get inserted feedback ID: %w", err)
	}
	feedback.ID = int(id)

	return nil
}

// GetByID retrieves feedback by ID.
func (r *recommendationFeedbackRepository) GetByID(ctx context.Context, id int) (*models.RecommendationFeedback, error) {
	query := `
		SELECT id, account_id, recommendation_type, recommendation_id, recommended_card_id,
			   recommended_archetype, context_data, action, alternate_choice_id,
			   outcome_match_id, outcome_result, recommendation_score, recommendation_rank,
			   recommended_at, responded_at, outcome_recorded_at, created_at
		FROM recommendation_feedback
		WHERE id = ?
	`

	feedback := &models.RecommendationFeedback{}
	var recommendedAt, createdAt string
	var respondedAt, outcomeRecordedAt *string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&feedback.ID, &feedback.AccountID, &feedback.RecommendationType,
		&feedback.RecommendationID, &feedback.RecommendedCardID, &feedback.RecommendedArchetype,
		&feedback.ContextData, &feedback.Action, &feedback.AlternateChoiceID,
		&feedback.OutcomeMatchID, &feedback.OutcomeResult, &feedback.RecommendationScore,
		&feedback.RecommendationRank, &recommendedAt, &respondedAt, &outcomeRecordedAt, &createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback: %w", err)
	}

	feedback.RecommendedAt, _ = time.Parse("2006-01-02 15:04:05.999999", recommendedAt)
	feedback.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", createdAt)
	if respondedAt != nil {
		t, _ := time.Parse("2006-01-02 15:04:05.999999", *respondedAt)
		feedback.RespondedAt = &t
	}
	if outcomeRecordedAt != nil {
		t, _ := time.Parse("2006-01-02 15:04:05.999999", *outcomeRecordedAt)
		feedback.OutcomeRecordedAt = &t
	}

	return feedback, nil
}

// GetByRecommendationID retrieves feedback by recommendation ID.
func (r *recommendationFeedbackRepository) GetByRecommendationID(ctx context.Context, recommendationID string) (*models.RecommendationFeedback, error) {
	query := `
		SELECT id, account_id, recommendation_type, recommendation_id, recommended_card_id,
			   recommended_archetype, context_data, action, alternate_choice_id,
			   outcome_match_id, outcome_result, recommendation_score, recommendation_rank,
			   recommended_at, responded_at, outcome_recorded_at, created_at
		FROM recommendation_feedback
		WHERE recommendation_id = ?
	`

	feedback := &models.RecommendationFeedback{}
	var recommendedAt, createdAt string
	var respondedAt, outcomeRecordedAt *string

	err := r.db.QueryRowContext(ctx, query, recommendationID).Scan(
		&feedback.ID, &feedback.AccountID, &feedback.RecommendationType,
		&feedback.RecommendationID, &feedback.RecommendedCardID, &feedback.RecommendedArchetype,
		&feedback.ContextData, &feedback.Action, &feedback.AlternateChoiceID,
		&feedback.OutcomeMatchID, &feedback.OutcomeResult, &feedback.RecommendationScore,
		&feedback.RecommendationRank, &recommendedAt, &respondedAt, &outcomeRecordedAt, &createdAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback by recommendation ID: %w", err)
	}

	feedback.RecommendedAt, _ = time.Parse("2006-01-02 15:04:05.999999", recommendedAt)
	feedback.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", createdAt)
	if respondedAt != nil {
		t, _ := time.Parse("2006-01-02 15:04:05.999999", *respondedAt)
		feedback.RespondedAt = &t
	}
	if outcomeRecordedAt != nil {
		t, _ := time.Parse("2006-01-02 15:04:05.999999", *outcomeRecordedAt)
		feedback.OutcomeRecordedAt = &t
	}

	return feedback, nil
}

// GetByAccount retrieves all feedback for an account.
func (r *recommendationFeedbackRepository) GetByAccount(ctx context.Context, accountID int, limit int) ([]*models.RecommendationFeedback, error) {
	query := `
		SELECT id, account_id, recommendation_type, recommendation_id, recommended_card_id,
			   recommended_archetype, context_data, action, alternate_choice_id,
			   outcome_match_id, outcome_result, recommendation_score, recommendation_rank,
			   recommended_at, responded_at, outcome_recorded_at, created_at
		FROM recommendation_feedback
		WHERE account_id = ?
		ORDER BY recommended_at DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, accountID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback by account: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanFeedbackRows(rows)
}

// GetByType retrieves feedback filtered by recommendation type.
func (r *recommendationFeedbackRepository) GetByType(ctx context.Context, accountID int, recType string, limit int) ([]*models.RecommendationFeedback, error) {
	query := `
		SELECT id, account_id, recommendation_type, recommendation_id, recommended_card_id,
			   recommended_archetype, context_data, action, alternate_choice_id,
			   outcome_match_id, outcome_result, recommendation_score, recommendation_rank,
			   recommended_at, responded_at, outcome_recorded_at, created_at
		FROM recommendation_feedback
		WHERE account_id = ? AND recommendation_type = ?
		ORDER BY recommended_at DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, accountID, recType, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback by type: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanFeedbackRows(rows)
}

// UpdateAction updates the user action for a recommendation.
func (r *recommendationFeedbackRepository) UpdateAction(ctx context.Context, id int, action string, alternateChoiceID *int) error {
	query := `
		UPDATE recommendation_feedback
		SET action = ?, alternate_choice_id = ?, responded_at = ?
		WHERE id = ?
	`

	now := time.Now().UTC().Format("2006-01-02 15:04:05.999999")

	_, err := r.db.ExecContext(ctx, query, action, alternateChoiceID, now, id)
	if err != nil {
		return fmt.Errorf("failed to update feedback action: %w", err)
	}

	return nil
}

// UpdateOutcome updates the match outcome for a recommendation.
func (r *recommendationFeedbackRepository) UpdateOutcome(ctx context.Context, id int, matchID string, result string) error {
	query := `
		UPDATE recommendation_feedback
		SET outcome_match_id = ?, outcome_result = ?, outcome_recorded_at = ?
		WHERE id = ?
	`

	now := time.Now().UTC().Format("2006-01-02 15:04:05.999999")

	_, err := r.db.ExecContext(ctx, query, matchID, result, now, id)
	if err != nil {
		return fmt.Errorf("failed to update feedback outcome: %w", err)
	}

	return nil
}

// GetStats calculates aggregated statistics for recommendations.
func (r *recommendationFeedbackRepository) GetStats(ctx context.Context, accountID int, recType *string) (*models.RecommendationStats, error) {
	query := `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN action = 'accepted' THEN 1 ELSE 0 END) as accepted,
			SUM(CASE WHEN action = 'rejected' THEN 1 ELSE 0 END) as rejected,
			SUM(CASE WHEN action = 'ignored' THEN 1 ELSE 0 END) as ignored,
			SUM(CASE WHEN action = 'alternate' THEN 1 ELSE 0 END) as alternate
		FROM recommendation_feedback
		WHERE account_id = ?
	`

	args := []interface{}{accountID}

	if recType != nil {
		query += " AND recommendation_type = ?"
		args = append(args, *recType)
	}

	var stats models.RecommendationStats

	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&stats.TotalRecommendations, &stats.AcceptedCount, &stats.RejectedCount,
		&stats.IgnoredCount, &stats.AlternateCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendation stats: %w", err)
	}

	if stats.TotalRecommendations > 0 {
		stats.AcceptanceRate = float64(stats.AcceptedCount) / float64(stats.TotalRecommendations)
	}

	// Calculate win rates for accepted vs rejected
	winRateQuery := `
		SELECT
			action,
			COUNT(*) as total,
			SUM(CASE WHEN outcome_result = 'win' THEN 1 ELSE 0 END) as wins
		FROM recommendation_feedback
		WHERE account_id = ? AND outcome_result IS NOT NULL
	`
	winRateArgs := []interface{}{accountID}

	if recType != nil {
		winRateQuery += " AND recommendation_type = ?"
		winRateArgs = append(winRateArgs, *recType)
	}

	winRateQuery += " GROUP BY action"

	rows, err := r.db.QueryContext(ctx, winRateQuery, winRateArgs...)
	if err != nil {
		return &stats, nil // Return stats without win rates on error
	}
	defer func() {
		_ = rows.Close()
	}()

	for rows.Next() {
		var action string
		var total, wins int
		if err := rows.Scan(&action, &total, &wins); err != nil {
			continue
		}
		if total > 0 {
			winRate := float64(wins) / float64(total)
			switch action {
			case "accepted":
				stats.WinRateOnAccepted = &winRate
			case "rejected":
				stats.WinRateOnRejected = &winRate
			}
		}
	}

	return &stats, nil
}

// GetStatsByDateRange calculates stats within a date range.
func (r *recommendationFeedbackRepository) GetStatsByDateRange(ctx context.Context, accountID int, start, end time.Time) (*models.RecommendationStats, error) {
	query := `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN action = 'accepted' THEN 1 ELSE 0 END) as accepted,
			SUM(CASE WHEN action = 'rejected' THEN 1 ELSE 0 END) as rejected,
			SUM(CASE WHEN action = 'ignored' THEN 1 ELSE 0 END) as ignored,
			SUM(CASE WHEN action = 'alternate' THEN 1 ELSE 0 END) as alternate
		FROM recommendation_feedback
		WHERE account_id = ? AND recommended_at BETWEEN ? AND ?
	`

	startStr := start.UTC().Format("2006-01-02 15:04:05.999999")
	endStr := end.UTC().Format("2006-01-02 15:04:05.999999")

	var stats models.RecommendationStats

	err := r.db.QueryRowContext(ctx, query, accountID, startStr, endStr).Scan(
		&stats.TotalRecommendations, &stats.AcceptedCount, &stats.RejectedCount,
		&stats.IgnoredCount, &stats.AlternateCount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recommendation stats by date range: %w", err)
	}

	if stats.TotalRecommendations > 0 {
		stats.AcceptanceRate = float64(stats.AcceptedCount) / float64(stats.TotalRecommendations)
	}

	return &stats, nil
}

// GetPendingFeedback retrieves recommendations without user response.
func (r *recommendationFeedbackRepository) GetPendingFeedback(ctx context.Context, accountID int) ([]*models.RecommendationFeedback, error) {
	query := `
		SELECT id, account_id, recommendation_type, recommendation_id, recommended_card_id,
			   recommended_archetype, context_data, action, alternate_choice_id,
			   outcome_match_id, outcome_result, recommendation_score, recommendation_rank,
			   recommended_at, responded_at, outcome_recorded_at, created_at
		FROM recommendation_feedback
		WHERE account_id = ? AND responded_at IS NULL
		ORDER BY recommended_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending feedback: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanFeedbackRows(rows)
}

// GetForMLTraining retrieves feedback suitable for ML training (with outcomes).
func (r *recommendationFeedbackRepository) GetForMLTraining(ctx context.Context, limit int) ([]*models.RecommendationFeedback, error) {
	query := `
		SELECT id, account_id, recommendation_type, recommendation_id, recommended_card_id,
			   recommended_archetype, context_data, action, alternate_choice_id,
			   outcome_match_id, outcome_result, recommendation_score, recommendation_rank,
			   recommended_at, responded_at, outcome_recorded_at, created_at
		FROM recommendation_feedback
		WHERE outcome_result IS NOT NULL AND responded_at IS NOT NULL
		ORDER BY recommended_at DESC
		LIMIT ?
	`

	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get feedback for ML training: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	return r.scanFeedbackRows(rows)
}

// scanFeedbackRows scans rows into RecommendationFeedback slice.
func (r *recommendationFeedbackRepository) scanFeedbackRows(rows *sql.Rows) ([]*models.RecommendationFeedback, error) {
	var feedbacks []*models.RecommendationFeedback

	for rows.Next() {
		f := &models.RecommendationFeedback{}
		var recommendedAt, createdAt string
		var respondedAt, outcomeRecordedAt *string

		err := rows.Scan(
			&f.ID, &f.AccountID, &f.RecommendationType, &f.RecommendationID,
			&f.RecommendedCardID, &f.RecommendedArchetype, &f.ContextData,
			&f.Action, &f.AlternateChoiceID, &f.OutcomeMatchID, &f.OutcomeResult,
			&f.RecommendationScore, &f.RecommendationRank, &recommendedAt,
			&respondedAt, &outcomeRecordedAt, &createdAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feedback row: %w", err)
		}

		f.RecommendedAt, _ = time.Parse("2006-01-02 15:04:05.999999", recommendedAt)
		f.CreatedAt, _ = time.Parse("2006-01-02 15:04:05.999999", createdAt)
		if respondedAt != nil {
			t, _ := time.Parse("2006-01-02 15:04:05.999999", *respondedAt)
			f.RespondedAt = &t
		}
		if outcomeRecordedAt != nil {
			t, _ := time.Parse("2006-01-02 15:04:05.999999", *outcomeRecordedAt)
			f.OutcomeRecordedAt = &t
		}

		feedbacks = append(feedbacks, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating feedback rows: %w", err)
	}

	return feedbacks, nil
}
