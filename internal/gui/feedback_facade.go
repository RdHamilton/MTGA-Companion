package gui

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/feedback"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// FeedbackFacade handles recommendation feedback operations.
type FeedbackFacade struct {
	services *Services
}

// NewFeedbackFacade creates a new FeedbackFacade.
func NewFeedbackFacade(services *Services) *FeedbackFacade {
	return &FeedbackFacade{
		services: services,
	}
}

// getFeedbackService creates a feedback service with the current account.
func (f *FeedbackFacade) getFeedbackService() (*feedback.Service, error) {
	if f.services.Storage == nil {
		return nil, &AppError{Message: "Database not initialized"}
	}

	accountID := f.services.Storage.GetCurrentAccountID()
	if accountID == 0 {
		return nil, &AppError{Message: "No active account"}
	}

	return feedback.NewService(f.services.Storage.RecommendationFeedbackRepo(), accountID), nil
}

// RecordRecommendationRequest represents a request to record a recommendation.
type RecordRecommendationRequest struct {
	RecommendationType   string                     `json:"recommendationType"`
	RecommendedCardID    *int                       `json:"recommendedCardID,omitempty"`
	RecommendedArchetype *string                    `json:"recommendedArchetype,omitempty"`
	Context              *RecommendationContextData `json:"context,omitempty"`
	Score                *float64                   `json:"score,omitempty"`
	Rank                 *int                       `json:"rank,omitempty"`
}

// RecommendationContextData represents context when a recommendation is made.
type RecommendationContextData struct {
	DeckID            string `json:"deckID,omitempty"`
	DraftEventID      string `json:"draftEventID,omitempty"`
	Format            string `json:"format,omitempty"`
	SetCode           string `json:"setCode,omitempty"`
	DeckCardCount     int    `json:"deckCardCount"`
	DeckColorIdentity string `json:"deckColorIdentity,omitempty"`
	PackNumber        int    `json:"packNumber,omitempty"`
	PickNumber        int    `json:"pickNumber,omitempty"`
	AvailableCards    []int  `json:"availableCards,omitempty"`
	CurrentArchetype  string `json:"currentArchetype,omitempty"`
	RecommendedCards  []int  `json:"recommendedCards,omitempty"`
}

// RecordRecommendationResponse contains the result of recording a recommendation.
type RecordRecommendationResponse struct {
	RecommendationID string `json:"recommendationID"`
	Success          bool   `json:"success"`
	Error            string `json:"error,omitempty"`
}

// RecordRecommendation records a new recommendation event.
func (f *FeedbackFacade) RecordRecommendation(ctx context.Context, req *RecordRecommendationRequest) (*RecordRecommendationResponse, error) {
	svc, err := f.getFeedbackService()
	if err != nil {
		return &RecordRecommendationResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	// Convert context
	var recContext *feedback.RecommendationContext
	if req.Context != nil {
		recContext = &feedback.RecommendationContext{
			DeckID:            req.Context.DeckID,
			DraftEventID:      req.Context.DraftEventID,
			Format:            req.Context.Format,
			SetCode:           req.Context.SetCode,
			DeckCardCount:     req.Context.DeckCardCount,
			DeckColorIdentity: req.Context.DeckColorIdentity,
			PackNumber:        req.Context.PackNumber,
			PickNumber:        req.Context.PickNumber,
			AvailableCards:    req.Context.AvailableCards,
			CurrentArchetype:  req.Context.CurrentArchetype,
			RecommendedCards:  req.Context.RecommendedCards,
		}
	}

	var id string
	err = storage.RetryOnBusy(func() error {
		var err error
		id, err = svc.RecordRecommendation(ctx, &feedback.RecordRecommendationRequest{
			RecommendationType:   req.RecommendationType,
			RecommendedCardID:    req.RecommendedCardID,
			RecommendedArchetype: req.RecommendedArchetype,
			Context:              recContext,
			Score:                req.Score,
			Rank:                 req.Rank,
		})
		return err
	})
	if err != nil {
		return &RecordRecommendationResponse{
			Success: false,
			Error:   err.Error(),
		}, nil
	}

	log.Printf("Recorded recommendation %s (type: %s)", id, req.RecommendationType)
	return &RecordRecommendationResponse{
		RecommendationID: id,
		Success:          true,
	}, nil
}

// RecordActionRequest represents a request to record user action on a recommendation.
type RecordActionRequest struct {
	RecommendationID  string `json:"recommendationID"`
	Action            string `json:"action"` // "accepted", "rejected", "ignored", "alternate"
	AlternateChoiceID *int   `json:"alternateChoiceID,omitempty"`
}

// RecordAction records the user's action on a recommendation.
func (f *FeedbackFacade) RecordAction(ctx context.Context, req *RecordActionRequest) error {
	svc, err := f.getFeedbackService()
	if err != nil {
		return err
	}

	err = storage.RetryOnBusy(func() error {
		return svc.RecordAction(ctx, req.RecommendationID, req.Action, req.AlternateChoiceID)
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to record action: %v", err)}
	}

	log.Printf("Recorded action '%s' for recommendation %s", req.Action, req.RecommendationID)
	return nil
}

// RecordOutcomeRequest represents a request to record match outcome for a recommendation.
type RecordOutcomeRequest struct {
	RecommendationID string `json:"recommendationID"`
	MatchID          string `json:"matchID"`
	Result           string `json:"result"` // "win" or "loss"
}

// RecordOutcome records the match outcome for a recommendation.
func (f *FeedbackFacade) RecordOutcome(ctx context.Context, req *RecordOutcomeRequest) error {
	svc, err := f.getFeedbackService()
	if err != nil {
		return err
	}

	err = storage.RetryOnBusy(func() error {
		return svc.RecordOutcome(ctx, req.RecommendationID, req.MatchID, req.Result)
	})
	if err != nil {
		return &AppError{Message: fmt.Sprintf("Failed to record outcome: %v", err)}
	}

	log.Printf("Recorded outcome '%s' for recommendation %s", req.Result, req.RecommendationID)
	return nil
}

// RecommendationStatsResponse contains recommendation statistics.
type RecommendationStatsResponse struct {
	TotalRecommendations int      `json:"totalRecommendations"`
	AcceptedCount        int      `json:"acceptedCount"`
	RejectedCount        int      `json:"rejectedCount"`
	IgnoredCount         int      `json:"ignoredCount"`
	AlternateCount       int      `json:"alternateCount"`
	AcceptanceRate       float64  `json:"acceptanceRate"`    // 0.0-1.0
	AcceptancePercent    int      `json:"acceptancePercent"` // 0-100
	WinRateOnAccepted    *float64 `json:"winRateOnAccepted,omitempty"`
	WinRateOnRejected    *float64 `json:"winRateOnRejected,omitempty"`
}

// GetRecommendationStats returns aggregated recommendation statistics.
func (f *FeedbackFacade) GetRecommendationStats(ctx context.Context, recType *string) (*RecommendationStatsResponse, error) {
	svc, err := f.getFeedbackService()
	if err != nil {
		return nil, err
	}

	var stats *feedback.Service
	_ = stats // Unused, using svc instead

	modelStats, err := svc.GetStats(ctx, recType)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get stats: %v", err)}
	}

	return &RecommendationStatsResponse{
		TotalRecommendations: modelStats.TotalRecommendations,
		AcceptedCount:        modelStats.AcceptedCount,
		RejectedCount:        modelStats.RejectedCount,
		IgnoredCount:         modelStats.IgnoredCount,
		AlternateCount:       modelStats.AlternateCount,
		AcceptanceRate:       modelStats.AcceptanceRate,
		AcceptancePercent:    int(modelStats.AcceptanceRate * 100),
		WinRateOnAccepted:    modelStats.WinRateOnAccepted,
		WinRateOnRejected:    modelStats.WinRateOnRejected,
	}, nil
}

// DashboardMetricsResponse contains comprehensive feedback metrics.
type DashboardMetricsResponse struct {
	TotalRecommendations int                     `json:"totalRecommendations"`
	AcceptanceRate       float64                 `json:"acceptanceRate"`
	AcceptancePercent    int                     `json:"acceptancePercent"`
	RejectionRate        float64                 `json:"rejectionRate"`
	RejectionPercent     int                     `json:"rejectionPercent"`
	WinRateOnAccepted    *float64                `json:"winRateOnAccepted,omitempty"`
	WinRateOnRejected    *float64                `json:"winRateOnRejected,omitempty"`
	WinRateDifference    *float64                `json:"winRateDifference,omitempty"`
	ByType               map[string]*TypeMetrics `json:"byType"`
	Last7Days            *PeriodMetrics          `json:"last7Days,omitempty"`
	Last30Days           *PeriodMetrics          `json:"last30Days,omitempty"`
}

// TypeMetrics provides metrics for a specific recommendation type.
type TypeMetrics struct {
	Total             int      `json:"total"`
	AcceptanceRate    float64  `json:"acceptanceRate"`
	AcceptancePercent int      `json:"acceptancePercent"`
	WinRateOnAccepted *float64 `json:"winRateOnAccepted,omitempty"`
}

// PeriodMetrics provides metrics for a time period.
type PeriodMetrics struct {
	Total             int     `json:"total"`
	AcceptanceRate    float64 `json:"acceptanceRate"`
	AcceptancePercent int     `json:"acceptancePercent"`
}

// GetDashboardMetrics returns comprehensive feedback metrics for the dashboard.
func (f *FeedbackFacade) GetDashboardMetrics(ctx context.Context) (*DashboardMetricsResponse, error) {
	svc, err := f.getFeedbackService()
	if err != nil {
		return nil, err
	}

	metrics, err := svc.GetDashboardMetrics(ctx)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to get dashboard metrics: %v", err)}
	}

	response := &DashboardMetricsResponse{
		TotalRecommendations: metrics.TotalRecommendations,
		AcceptanceRate:       metrics.AcceptanceRate,
		AcceptancePercent:    int(metrics.AcceptanceRate * 100),
		RejectionRate:        metrics.RejectionRate,
		RejectionPercent:     int(metrics.RejectionRate * 100),
		WinRateOnAccepted:    metrics.WinRateOnAccepted,
		WinRateOnRejected:    metrics.WinRateOnRejected,
		WinRateDifference:    metrics.WinRateDifference,
		ByType:               make(map[string]*TypeMetrics),
	}

	// Convert type metrics
	for typeName, typeMetrics := range metrics.ByType {
		response.ByType[typeName] = &TypeMetrics{
			Total:             typeMetrics.Total,
			AcceptanceRate:    typeMetrics.AcceptanceRate,
			AcceptancePercent: int(typeMetrics.AcceptanceRate * 100),
			WinRateOnAccepted: typeMetrics.WinRateOnAccepted,
		}
	}

	// Convert period metrics
	if metrics.Last7Days != nil {
		response.Last7Days = &PeriodMetrics{
			Total:             metrics.Last7Days.Total,
			AcceptanceRate:    metrics.Last7Days.AcceptanceRate,
			AcceptancePercent: int(metrics.Last7Days.AcceptanceRate * 100),
		}
	}

	if metrics.Last30Days != nil {
		response.Last30Days = &PeriodMetrics{
			Total:             metrics.Last30Days.Total,
			AcceptanceRate:    metrics.Last30Days.AcceptanceRate,
			AcceptancePercent: int(metrics.Last30Days.AcceptanceRate * 100),
		}
	}

	return response, nil
}

// MLTrainingDataExport represents exported training data for ML.
type MLTrainingDataExport struct {
	Data       []*MLTrainingEntry `json:"data"`
	TotalCount int                `json:"totalCount"`
	ExportedAt string             `json:"exportedAt"`
}

// MLTrainingEntry represents a single training data point.
type MLTrainingEntry struct {
	RecommendationType   string                     `json:"recommendationType"`
	RecommendedCardID    *int                       `json:"recommendedCardID,omitempty"`
	RecommendedArchetype *string                    `json:"recommendedArchetype,omitempty"`
	Context              *RecommendationContextData `json:"context,omitempty"`
	Action               string                     `json:"action"`
	AlternateChoiceID    *int                       `json:"alternateChoiceID,omitempty"`
	OutcomeResult        *string                    `json:"outcomeResult,omitempty"`
	RecommendationScore  *float64                   `json:"recommendationScore,omitempty"`
	RecommendationRank   *int                       `json:"recommendationRank,omitempty"`
	RecommendedAt        string                     `json:"recommendedAt"`
	RespondedAt          *string                    `json:"respondedAt,omitempty"`
}

// ExportMLTrainingData exports feedback data for ML training.
func (f *FeedbackFacade) ExportMLTrainingData(ctx context.Context, limit int) (*MLTrainingDataExport, error) {
	svc, err := f.getFeedbackService()
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 1000 // Default limit
	}

	trainingData, err := svc.ExportForMLTraining(ctx, limit)
	if err != nil {
		return nil, &AppError{Message: fmt.Sprintf("Failed to export training data: %v", err)}
	}

	entries := make([]*MLTrainingEntry, 0, len(trainingData))
	for _, td := range trainingData {
		entry := &MLTrainingEntry{
			RecommendationType:   td.RecommendationType,
			RecommendedCardID:    td.RecommendedCardID,
			RecommendedArchetype: td.RecommendedArchetype,
			Action:               td.Action,
			AlternateChoiceID:    td.AlternateChoiceID,
			OutcomeResult:        td.OutcomeResult,
			RecommendationScore:  td.RecommendationScore,
			RecommendationRank:   td.RecommendationRank,
			RecommendedAt:        td.RecommendedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		if td.Context != nil {
			entry.Context = &RecommendationContextData{
				DeckID:            td.Context.DeckID,
				DraftEventID:      td.Context.DraftEventID,
				Format:            td.Context.Format,
				SetCode:           td.Context.SetCode,
				DeckCardCount:     td.Context.DeckCardCount,
				DeckColorIdentity: td.Context.DeckColorIdentity,
				PackNumber:        td.Context.PackNumber,
				PickNumber:        td.Context.PickNumber,
				AvailableCards:    td.Context.AvailableCards,
				CurrentArchetype:  td.Context.CurrentArchetype,
				RecommendedCards:  td.Context.RecommendedCards,
			}
		}

		if td.RespondedAt != nil {
			respondedStr := td.RespondedAt.Format("2006-01-02T15:04:05Z07:00")
			entry.RespondedAt = &respondedStr
		}

		entries = append(entries, entry)
	}

	return &MLTrainingDataExport{
		Data:       entries,
		TotalCount: len(entries),
		ExportedAt: time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
