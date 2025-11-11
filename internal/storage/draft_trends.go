package storage

import (
	"context"
	"fmt"
	"time"
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
	query := `
		SELECT
			cached_at, gihwr, ohwr, alsa, ata, gih, games_played
		FROM draft_card_ratings
		WHERE arena_id = ?
		  AND expansion = ?
		  AND cached_at >= datetime('now', '-' || ? || ' days')
		ORDER BY cached_at ASC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query, arenaID, expansion, days)
	if err != nil {
		return nil, fmt.Errorf("failed to query trend: %w", err)
	}
	defer func() { _ = rows.Close() }()

	trend := &CardTrend{
		ArenaID:   arenaID,
		Expansion: expansion,
	}

	for rows.Next() {
		var point TrendPoint
		var cachedAtStr string
		var gihwr, ohwr, alsa, ata *float64
		var gih, gamesPlayed *int

		err := rows.Scan(&cachedAtStr, &gihwr, &ohwr, &alsa, &ata, &gih, &gamesPlayed)
		if err != nil {
			return nil, fmt.Errorf("failed to scan trend point: %w", err)
		}

		// Parse date
		point.Date, _ = time.Parse("2006-01-02 15:04:05", cachedAtStr)

		// Set values (handle NULLs)
		if gihwr != nil {
			point.GIHWR = *gihwr
		}
		if ohwr != nil {
			point.OHWR = *ohwr
		}
		if alsa != nil {
			point.ALSA = *alsa
		}
		if ata != nil {
			point.ATA = *ata
		}
		if gih != nil {
			point.SampleSize = *gih
		}
		if gamesPlayed != nil {
			point.GamesPlayed = *gamesPlayed
		}

		trend.Points = append(trend.Points, point)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating trend points: %w", err)
	}

	if len(trend.Points) > 0 {
		trend.StartDate = trend.Points[0].Date
		trend.EndDate = trend.Points[len(trend.Points)-1].Date
		trend.TotalPoints = len(trend.Points)
	}

	return trend, nil
}

// GetExpansionTrends returns trends for all cards in an expansion.
func (s *Service) GetExpansionTrends(ctx context.Context, expansion string, days int) (map[int]*CardTrend, error) {
	query := `
		SELECT DISTINCT arena_id
		FROM draft_card_ratings
		WHERE expansion = ?
		  AND cached_at >= datetime('now', '-' || ? || ' days')
	`

	rows, err := s.db.Conn().QueryContext(ctx, query, expansion, days)
	if err != nil {
		return nil, fmt.Errorf("failed to query cards: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var arenaIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan arena ID: %w", err)
		}
		arenaIDs = append(arenaIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
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

	// Get early period data
	earlyQuery := `
		SELECT
			arena_id, AVG(gihwr) as avg_gihwr, SUM(gih) as total_gih
		FROM draft_card_ratings
		WHERE expansion = ?
		  AND cached_at BETWEEN ? AND ?
		GROUP BY arena_id
		HAVING total_gih > 100
	`

	earlyData := make(map[int]struct {
		GIHWR      float64
		SampleSize int
	})

	rows, err := s.db.Conn().QueryContext(ctx, earlyQuery,
		expansion,
		comp.EarlyStartDate.Format("2006-01-02"),
		comp.EarlyEndDate.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to query early period: %w", err)
	}

	for rows.Next() {
		var arenaID int
		var avgGIHWR *float64
		var totalGIH *int

		if err := rows.Scan(&arenaID, &avgGIHWR, &totalGIH); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("failed to scan early data: %w", err)
		}

		if avgGIHWR != nil && totalGIH != nil {
			earlyData[arenaID] = struct {
				GIHWR      float64
				SampleSize int
			}{
				GIHWR:      *avgGIHWR,
				SampleSize: *totalGIH,
			}
		}
	}
	_ = rows.Close()

	comp.EarlyCards = len(earlyData)

	// Get late period data
	lateQuery := `
		SELECT
			arena_id, AVG(gihwr) as avg_gihwr, SUM(gih) as total_gih
		FROM draft_card_ratings
		WHERE expansion = ?
		  AND cached_at BETWEEN ? AND ?
		GROUP BY arena_id
		HAVING total_gih > 100
	`

	lateData := make(map[int]struct {
		GIHWR      float64
		SampleSize int
	})

	rows, err = s.db.Conn().QueryContext(ctx, lateQuery,
		expansion,
		comp.LateStartDate.Format("2006-01-02"),
		comp.LateEndDate.Format("2006-01-02"))
	if err != nil {
		return nil, fmt.Errorf("failed to query late period: %w", err)
	}

	for rows.Next() {
		var arenaID int
		var avgGIHWR *float64
		var totalGIH *int

		if err := rows.Scan(&arenaID, &avgGIHWR, &totalGIH); err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("failed to scan late data: %w", err)
		}

		if avgGIHWR != nil && totalGIH != nil {
			lateData[arenaID] = struct {
				GIHWR      float64
				SampleSize int
			}{
				GIHWR:      *avgGIHWR,
				SampleSize: *totalGIH,
			}
		}
	}
	_ = rows.Close()

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
			EarlyGIHWR:      earlyStats.GIHWR,
			LateGIHWR:       lateStats.GIHWR,
			GIHWRChange:     lateStats.GIHWR - earlyStats.GIHWR,
			EarlySampleSize: earlyStats.SampleSize,
			LateSampleSize:  lateStats.SampleSize,
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
	query := `
		SELECT
			id, arena_id, expansion, format, colors,
			gihwr, ohwr, gpwr, gdwr, ihdwr,
			gihwr_delta, ohwr_delta, gdwr_delta, ihdwr_delta,
			alsa, ata,
			gih, oh, gp, gd, ihd,
			games_played, num_decks,
			start_date, end_date, cached_at, last_updated
		FROM draft_card_ratings
		WHERE arena_id = ?
		  AND expansion = ?
		ORDER BY cached_at ASC
	`

	rows, err := s.db.Conn().QueryContext(ctx, query, arenaID, expansion)
	if err != nil {
		return nil, fmt.Errorf("failed to query rating history: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var history []*DraftCardRating
	for rows.Next() {
		rating := &DraftCardRating{}
		var cachedAtStr, lastUpdatedStr string

		err := rows.Scan(
			&rating.ID,
			&rating.ArenaID,
			&rating.Expansion,
			&rating.Format,
			&rating.Colors,
			&rating.GIHWR,
			&rating.OHWR,
			&rating.GPWR,
			&rating.GDWR,
			&rating.IHDWR,
			&rating.GIHWRDelta,
			&rating.OHWRDelta,
			&rating.GDWRDelta,
			&rating.IHDWRDelta,
			&rating.ALSA,
			&rating.ATA,
			&rating.GIH,
			&rating.OH,
			&rating.GP,
			&rating.GD,
			&rating.IHD,
			&rating.GamesPlayed,
			&rating.NumDecks,
			&rating.StartDate,
			&rating.EndDate,
			&cachedAtStr,
			&lastUpdatedStr,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rating: %w", err)
		}

		// Parse timestamps
		rating.CachedAt, _ = time.Parse("2006-01-02 15:04:05", cachedAtStr)
		rating.LastUpdated, _ = time.Parse("2006-01-02 15:04:05", lastUpdatedStr)

		history = append(history, rating)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating history: %w", err)
	}

	return history, nil
}
