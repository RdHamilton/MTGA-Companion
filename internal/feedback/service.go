package feedback

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// Service handles recommendation feedback collection and analysis.
type Service struct {
	feedbackRepo repository.RecommendationFeedbackRepository
	accountID    int
}

// NewService creates a new feedback service.
func NewService(feedbackRepo repository.RecommendationFeedbackRepository, accountID int) *Service {
	return &Service{
		feedbackRepo: feedbackRepo,
		accountID:    accountID,
	}
}

// RecommendationContext captures the state when a recommendation is made.
type RecommendationContext struct {
	DeckID            string    `json:"deckID,omitempty"`
	DraftEventID      string    `json:"draftEventID,omitempty"`
	Format            string    `json:"format,omitempty"`
	SetCode           string    `json:"setCode,omitempty"`
	DeckCardCount     int       `json:"deckCardCount"`
	DeckColorIdentity string    `json:"deckColorIdentity,omitempty"`
	PackNumber        int       `json:"packNumber,omitempty"`
	PickNumber        int       `json:"pickNumber,omitempty"`
	AvailableCards    []int     `json:"availableCards,omitempty"`
	CurrentArchetype  string    `json:"currentArchetype,omitempty"`
	RecommendedCards  []int     `json:"recommendedCards,omitempty"`
	Timestamp         time.Time `json:"timestamp"`
}

// RecordRecommendationRequest contains data for recording a new recommendation.
type RecordRecommendationRequest struct {
	RecommendationType   string                 // "card_pick", "deck_card", "archetype", "sideboard"
	RecommendedCardID    *int                   // Card that was recommended
	RecommendedArchetype *string                // Or archetype that was recommended
	Context              *RecommendationContext // State at time of recommendation
	Score                *float64               // Recommendation score
	Rank                 *int                   // Position in recommendation list
}

// RecordRecommendation records a new recommendation event.
func (s *Service) RecordRecommendation(ctx context.Context, req *RecordRecommendationRequest) (string, error) {
	// Serialize context to JSON
	contextData := "{}"
	if req.Context != nil {
		data, err := json.Marshal(req.Context)
		if err != nil {
			return "", fmt.Errorf("failed to serialize context: %w", err)
		}
		contextData = string(data)
	}

	recommendationID := uuid.New().String()

	feedback := &models.RecommendationFeedback{
		AccountID:            s.accountID,
		RecommendationType:   req.RecommendationType,
		RecommendationID:     recommendationID,
		RecommendedCardID:    req.RecommendedCardID,
		RecommendedArchetype: req.RecommendedArchetype,
		ContextData:          contextData,
		Action:               "ignored", // Default until user responds
		RecommendationScore:  req.Score,
		RecommendationRank:   req.Rank,
		RecommendedAt:        time.Now(),
	}

	err := s.feedbackRepo.Create(ctx, feedback)
	if err != nil {
		return "", fmt.Errorf("failed to record recommendation: %w", err)
	}

	return recommendationID, nil
}

// RecordAction records the user's action on a recommendation.
func (s *Service) RecordAction(ctx context.Context, recommendationID, action string, alternateChoiceID *int) error {
	// Validate action
	validActions := map[string]bool{
		"accepted":  true,
		"rejected":  true,
		"ignored":   true,
		"alternate": true,
	}
	if !validActions[action] {
		return fmt.Errorf("invalid action: %s", action)
	}

	// Get feedback by recommendation ID
	feedback, err := s.feedbackRepo.GetByRecommendationID(ctx, recommendationID)
	if err != nil {
		return fmt.Errorf("failed to get recommendation: %w", err)
	}
	if feedback == nil {
		return fmt.Errorf("recommendation not found: %s", recommendationID)
	}

	// Update action
	err = s.feedbackRepo.UpdateAction(ctx, feedback.ID, action, alternateChoiceID)
	if err != nil {
		return fmt.Errorf("failed to update action: %w", err)
	}

	return nil
}

// RecordOutcome records the match outcome for a recommendation.
func (s *Service) RecordOutcome(ctx context.Context, recommendationID, matchID, result string) error {
	// Validate result
	if result != "win" && result != "loss" {
		return fmt.Errorf("invalid result: %s (must be 'win' or 'loss')", result)
	}

	feedback, err := s.feedbackRepo.GetByRecommendationID(ctx, recommendationID)
	if err != nil {
		return fmt.Errorf("failed to get recommendation: %w", err)
	}
	if feedback == nil {
		return fmt.Errorf("recommendation not found: %s", recommendationID)
	}

	err = s.feedbackRepo.UpdateOutcome(ctx, feedback.ID, matchID, result)
	if err != nil {
		return fmt.Errorf("failed to update outcome: %w", err)
	}

	return nil
}

// GetStats returns aggregated statistics for recommendations.
func (s *Service) GetStats(ctx context.Context, recType *string) (*models.RecommendationStats, error) {
	return s.feedbackRepo.GetStats(ctx, s.accountID, recType)
}

// GetStatsByDateRange returns statistics for a date range.
func (s *Service) GetStatsByDateRange(ctx context.Context, start, end time.Time) (*models.RecommendationStats, error) {
	return s.feedbackRepo.GetStatsByDateRange(ctx, s.accountID, start, end)
}

// GetRecentFeedback returns recent feedback entries.
func (s *Service) GetRecentFeedback(ctx context.Context, limit int) ([]*models.RecommendationFeedback, error) {
	return s.feedbackRepo.GetByAccount(ctx, s.accountID, limit)
}

// GetFeedbackByType returns feedback filtered by type.
func (s *Service) GetFeedbackByType(ctx context.Context, recType string, limit int) ([]*models.RecommendationFeedback, error) {
	return s.feedbackRepo.GetByType(ctx, s.accountID, recType, limit)
}

// GetPendingFeedback returns recommendations that haven't been responded to.
func (s *Service) GetPendingFeedback(ctx context.Context) ([]*models.RecommendationFeedback, error) {
	return s.feedbackRepo.GetPendingFeedback(ctx, s.accountID)
}

// MLTrainingData represents a single training data point for ML.
type MLTrainingData struct {
	RecommendationType   string                 `json:"recommendationType"`
	RecommendedCardID    *int                   `json:"recommendedCardID,omitempty"`
	RecommendedArchetype *string                `json:"recommendedArchetype,omitempty"`
	Context              *RecommendationContext `json:"context"`
	Action               string                 `json:"action"`
	AlternateChoiceID    *int                   `json:"alternateChoiceID,omitempty"`
	OutcomeResult        *string                `json:"outcomeResult,omitempty"`
	RecommendationScore  *float64               `json:"recommendationScore,omitempty"`
	RecommendationRank   *int                   `json:"recommendationRank,omitempty"`
	RecommendedAt        time.Time              `json:"recommendedAt"`
	RespondedAt          *time.Time             `json:"respondedAt,omitempty"`
}

// ExportForMLTraining exports feedback data formatted for ML training.
func (s *Service) ExportForMLTraining(ctx context.Context, limit int) ([]*MLTrainingData, error) {
	feedbacks, err := s.feedbackRepo.GetForMLTraining(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get training data: %w", err)
	}

	results := make([]*MLTrainingData, 0, len(feedbacks))
	for _, fb := range feedbacks {
		// Parse context from JSON
		var ctx *RecommendationContext
		if fb.ContextData != "" && fb.ContextData != "{}" {
			ctx = &RecommendationContext{}
			if err := json.Unmarshal([]byte(fb.ContextData), ctx); err != nil {
				// Skip entries with invalid context
				continue
			}
		}

		results = append(results, &MLTrainingData{
			RecommendationType:   fb.RecommendationType,
			RecommendedCardID:    fb.RecommendedCardID,
			RecommendedArchetype: fb.RecommendedArchetype,
			Context:              ctx,
			Action:               fb.Action,
			AlternateChoiceID:    fb.AlternateChoiceID,
			OutcomeResult:        fb.OutcomeResult,
			RecommendationScore:  fb.RecommendationScore,
			RecommendationRank:   fb.RecommendationRank,
			RecommendedAt:        fb.RecommendedAt,
			RespondedAt:          fb.RespondedAt,
		})
	}

	return results, nil
}

// DashboardMetrics provides high-level metrics for the feedback dashboard.
type DashboardMetrics struct {
	TotalRecommendations int                     `json:"totalRecommendations"`
	AcceptanceRate       float64                 `json:"acceptanceRate"`    // 0.0-1.0
	RejectionRate        float64                 `json:"rejectionRate"`     // 0.0-1.0
	WinRateOnAccepted    *float64                `json:"winRateOnAccepted"` // Win rate when recommendation was accepted
	WinRateOnRejected    *float64                `json:"winRateOnRejected"` // Win rate when recommendation was rejected
	WinRateDifference    *float64                `json:"winRateDifference"` // Difference (positive means accepted performed better)
	ByType               map[string]*TypeMetrics `json:"byType"`
	Last7Days            *PeriodMetrics          `json:"last7Days"`
	Last30Days           *PeriodMetrics          `json:"last30Days"`
}

// TypeMetrics provides metrics for a specific recommendation type.
type TypeMetrics struct {
	Total             int      `json:"total"`
	AcceptanceRate    float64  `json:"acceptanceRate"`
	WinRateOnAccepted *float64 `json:"winRateOnAccepted,omitempty"`
}

// PeriodMetrics provides metrics for a time period.
type PeriodMetrics struct {
	Total          int     `json:"total"`
	AcceptanceRate float64 `json:"acceptanceRate"`
}

// GetDashboardMetrics returns comprehensive metrics for the dashboard.
func (s *Service) GetDashboardMetrics(ctx context.Context) (*DashboardMetrics, error) {
	// Get overall stats
	overallStats, err := s.feedbackRepo.GetStats(ctx, s.accountID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get overall stats: %w", err)
	}

	metrics := &DashboardMetrics{
		TotalRecommendations: overallStats.TotalRecommendations,
		AcceptanceRate:       overallStats.AcceptanceRate,
		WinRateOnAccepted:    overallStats.WinRateOnAccepted,
		WinRateOnRejected:    overallStats.WinRateOnRejected,
		ByType:               make(map[string]*TypeMetrics),
	}

	// Calculate rejection rate
	if overallStats.TotalRecommendations > 0 {
		metrics.RejectionRate = float64(overallStats.RejectedCount) / float64(overallStats.TotalRecommendations)
	}

	// Calculate win rate difference
	if overallStats.WinRateOnAccepted != nil && overallStats.WinRateOnRejected != nil {
		diff := *overallStats.WinRateOnAccepted - *overallStats.WinRateOnRejected
		metrics.WinRateDifference = &diff
	}

	// Get stats by type
	types := []string{"card_pick", "deck_card", "archetype", "sideboard"}
	for _, recType := range types {
		typeStats, err := s.feedbackRepo.GetStats(ctx, s.accountID, &recType)
		if err != nil {
			continue
		}
		if typeStats.TotalRecommendations > 0 {
			metrics.ByType[recType] = &TypeMetrics{
				Total:             typeStats.TotalRecommendations,
				AcceptanceRate:    typeStats.AcceptanceRate,
				WinRateOnAccepted: typeStats.WinRateOnAccepted,
			}
		}
	}

	// Get 7-day stats
	now := time.Now()
	sevenDaysAgo := now.Add(-7 * 24 * time.Hour)
	weekStats, err := s.feedbackRepo.GetStatsByDateRange(ctx, s.accountID, sevenDaysAgo, now)
	if err == nil && weekStats.TotalRecommendations > 0 {
		metrics.Last7Days = &PeriodMetrics{
			Total:          weekStats.TotalRecommendations,
			AcceptanceRate: weekStats.AcceptanceRate,
		}
	}

	// Get 30-day stats
	thirtyDaysAgo := now.Add(-30 * 24 * time.Hour)
	monthStats, err := s.feedbackRepo.GetStatsByDateRange(ctx, s.accountID, thirtyDaysAgo, now)
	if err == nil && monthStats.TotalRecommendations > 0 {
		metrics.Last30Days = &PeriodMetrics{
			Total:          monthStats.TotalRecommendations,
			AcceptanceRate: monthStats.AcceptanceRate,
		}
	}

	return metrics, nil
}

// RecordBatchRecommendations records multiple recommendations at once (e.g., a list of suggestions).
func (s *Service) RecordBatchRecommendations(ctx context.Context, recType string, context *RecommendationContext, recommendations []int) ([]string, error) {
	ids := make([]string, 0, len(recommendations))

	for rank, cardID := range recommendations {
		cardIDCopy := cardID
		rankCopy := rank + 1 // 1-indexed

		req := &RecordRecommendationRequest{
			RecommendationType: recType,
			RecommendedCardID:  &cardIDCopy,
			Context:            context,
			Rank:               &rankCopy,
		}

		id, err := s.RecordRecommendation(ctx, req)
		if err != nil {
			// Log but continue with others
			continue
		}
		ids = append(ids, id)
	}

	return ids, nil
}
