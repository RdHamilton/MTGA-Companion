package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// TrendPoint represents a single point in a trend analysis.
type TrendPoint struct {
	Date        time.Time
	GIHWR       float64
	OHWR        float64
	ALSA        float64
	ATA         float64
	SampleSize  int // GIH count
	GamesPlayed int
}

// CardTrend represents the complete trend data for a card.
type CardTrend struct {
	ArenaID     int
	CardName    string // Populated if available
	Expansion   string
	Format      string
	Colors      string
	Points      []TrendPoint
	StartDate   time.Time
	EndDate     time.Time
	TotalPoints int
}

// MetaComparison compares two time periods to show meta evolution.
type MetaComparison struct {
	Expansion string
	Format    string

	// Early period stats
	EarlyStartDate time.Time
	EarlyEndDate   time.Time
	EarlyCards     int

	// Late period stats
	LateStartDate time.Time
	LateEndDate   time.Time
	LateCards     int

	// Top improving cards
	TopImprovers []CardImprovement

	// Top declining cards
	TopDecliners []CardImprovement
}

// CardImprovement represents how a card's rating changed between periods.
type CardImprovement struct {
	ArenaID         int
	CardName        string // If available
	EarlyGIHWR      float64
	LateGIHWR       float64
	GIHWRChange     float64 // Late - Early
	EarlySampleSize int
	LateSampleSize  int
}

// GetCardWinRateTrend returns the win rate trend for a specific card over time.
func (s *Service) GetCardWinRateTrend(ctx context.Context, arenaID int, expansion string, days int) (*CardTrend, error) {
	repoPoints, err := s.draftRatings.GetCardWinRateTrend(ctx, arenaID, expansion, days)
	if err != nil {
		return nil, fmt.Errorf("failed to query trend: %w", err)
	}

	trend := &CardTrend{
		ArenaID:   arenaID,
		Expansion: expansion,
	}

	for _, rp := range repoPoints {
		trend.Points = append(trend.Points, convertRepoTrendPointToStorage(rp))
	}

	if len(trend.Points) > 0 {
		trend.StartDate = trend.Points[0].Date
		trend.EndDate = trend.Points[len(trend.Points)-1].Date
		trend.TotalPoints = len(trend.Points)
	}

	return trend, nil
}

// convertRepoTrendPointToStorage converts a repository TrendPoint to storage TrendPoint.
func convertRepoTrendPointToStorage(rp *repository.TrendPoint) TrendPoint {
	return TrendPoint{
		Date:        rp.Date,
		GIHWR:       rp.GIHWR,
		OHWR:        rp.OHWR,
		ALSA:        rp.ALSA,
		ATA:         rp.ATA,
		SampleSize:  rp.SampleSize,
		GamesPlayed: rp.GamesPlayed,
	}
}

// GetExpansionTrends returns trends for all cards in an expansion.
func (s *Service) GetExpansionTrends(ctx context.Context, expansion string, days int) (map[int]*CardTrend, error) {
	arenaIDs, err := s.draftRatings.GetExpansionCardIDs(ctx, expansion, days)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards: %w", err)
	}

	// Get trend for each card
	trends := make(map[int]*CardTrend)
	for _, arenaID := range arenaIDs {
		trend, err := s.GetCardWinRateTrend(ctx, arenaID, expansion, days)
		if err != nil {
			continue // Skip cards with errors
		}
		trends[arenaID] = trend
	}

	return trends, nil
}

// CompareMetaPeriods compares early and late draft meta for an expansion.
func (s *Service) CompareMetaPeriods(ctx context.Context, expansion string, earlyDays, lateDays int) (*MetaComparison, error) {
	now := time.Now()
	comp := &MetaComparison{
		Expansion: expansion,
		Format:    "PremierDraft",
	}

	// Define time periods
	// Early period: e.g., days 90-60 ago
	// Late period: e.g., days 30-0 ago
	comp.EarlyEndDate = now.AddDate(0, 0, -(lateDays + earlyDays))
	comp.EarlyStartDate = comp.EarlyEndDate.AddDate(0, 0, -earlyDays)
	comp.LateEndDate = now
	comp.LateStartDate = now.AddDate(0, 0, -lateDays)

	// Get early period data using repository
	earlyData, err := s.draftRatings.GetPeriodAverages(ctx, expansion, comp.EarlyStartDate, comp.EarlyEndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query early period: %w", err)
	}
	comp.EarlyCards = len(earlyData)

	// Get late period data using repository
	lateData, err := s.draftRatings.GetPeriodAverages(ctx, expansion, comp.LateStartDate, comp.LateEndDate)
	if err != nil {
		return nil, fmt.Errorf("failed to query late period: %w", err)
	}
	comp.LateCards = len(lateData)

	// Compare and find improvements/declines
	var improvements []CardImprovement
	for arenaID, earlyStats := range earlyData {
		lateStats, exists := lateData[arenaID]
		if !exists {
			continue // Card not in late period
		}

		improvement := CardImprovement{
			ArenaID:         arenaID,
			EarlyGIHWR:      earlyStats.AvgGIHWR,
			LateGIHWR:       lateStats.AvgGIHWR,
			GIHWRChange:     lateStats.AvgGIHWR - earlyStats.AvgGIHWR,
			EarlySampleSize: earlyStats.TotalGIH,
			LateSampleSize:  lateStats.TotalGIH,
		}

		improvements = append(improvements, improvement)
	}

	// Sort by change (descending for improvers, ascending for decliners)
	// Top improvers: biggest positive change
	// Top decliners: biggest negative change

	// Simple bubble sort for now (can optimize later)
	for i := 0; i < len(improvements); i++ {
		for j := i + 1; j < len(improvements); j++ {
			if improvements[j].GIHWRChange > improvements[i].GIHWRChange {
				improvements[i], improvements[j] = improvements[j], improvements[i]
			}
		}
	}

	// Top 10 improvers
	if len(improvements) > 10 {
		comp.TopImprovers = improvements[:10]
	} else {
		comp.TopImprovers = improvements
	}

	// Reverse for decliners (bottom 10)
	if len(improvements) > 10 {
		comp.TopDecliners = improvements[len(improvements)-10:]
	} else {
		comp.TopDecliners = improvements
	}

	return comp, nil
}

// GetRatingHistory returns the complete rating history for a card.
func (s *Service) GetRatingHistory(ctx context.Context, arenaID int, expansion string) ([]*DraftCardRating, error) {
	repoHistory, err := s.draftRatings.GetCardRatingHistory(ctx, arenaID, expansion)
	if err != nil {
		return nil, fmt.Errorf("failed to query rating history: %w", err)
	}

	// Convert repository snapshots to storage DraftCardRating
	history := make([]*DraftCardRating, len(repoHistory))
	for i, rs := range repoHistory {
		history[i] = &DraftCardRating{
			ID:          rs.ID,
			ArenaID:     rs.ArenaID,
			Expansion:   rs.Expansion,
			Format:      rs.Format,
			Colors:      rs.Colors,
			GIHWR:       rs.GIHWR,
			OHWR:        rs.OHWR,
			GPWR:        rs.GPWR,
			GDWR:        rs.GDWR,
			IHDWR:       rs.IHDWR,
			ALSA:        rs.ALSA,
			ATA:         rs.ATA,
			GIH:         rs.GIH,
			GamesPlayed: rs.GamesPlayed,
			NumDecks:    rs.NumDecks,
			StartDate:   rs.StartDate,
			EndDate:     rs.EndDate,
			CachedAt:    rs.CachedAt,
			LastUpdated: rs.LastUpdated,
		}
	}

	return history, nil
}
