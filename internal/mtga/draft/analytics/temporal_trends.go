package analytics

import (
	"context"
	"sort"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// TemporalTrendAnalyzer analyzes draft performance over time.
type TemporalTrendAnalyzer struct {
	draftRepo     repository.DraftRepository
	analyticsRepo repository.DraftAnalyticsRepository
}

// NewTemporalTrendAnalyzer creates a new temporal trend analyzer.
func NewTemporalTrendAnalyzer(
	draftRepo repository.DraftRepository,
	analyticsRepo repository.DraftAnalyticsRepository,
) *TemporalTrendAnalyzer {
	return &TemporalTrendAnalyzer{
		draftRepo:     draftRepo,
		analyticsRepo: analyticsRepo,
	}
}

// CalculateWeeklyTrends calculates weekly performance trends.
func (t *TemporalTrendAnalyzer) CalculateWeeklyTrends(ctx context.Context, numWeeks int, setCode *string) ([]*models.DraftTemporalTrend, error) {
	return t.calculateTrends(ctx, models.PeriodTypeWeek, numWeeks, setCode)
}

// CalculateMonthlyTrends calculates monthly performance trends.
func (t *TemporalTrendAnalyzer) CalculateMonthlyTrends(ctx context.Context, numMonths int, setCode *string) ([]*models.DraftTemporalTrend, error) {
	return t.calculateTrends(ctx, models.PeriodTypeMonth, numMonths, setCode)
}

func (t *TemporalTrendAnalyzer) calculateTrends(ctx context.Context, periodType string, numPeriods int, setCode *string) ([]*models.DraftTemporalTrend, error) {
	// Get all completed draft sessions
	sessions, err := t.draftRepo.GetCompletedSessions(ctx, 10000)
	if err != nil {
		return nil, err
	}

	// Filter by set code if provided
	var filteredSessions []*models.DraftSession
	for _, s := range sessions {
		if setCode == nil || s.SetCode == *setCode {
			filteredSessions = append(filteredSessions, s)
		}
	}

	// Get all match results for these sessions
	sessionResults := make(map[string][]*models.DraftMatchResult)
	for _, session := range filteredSessions {
		results, err := t.analyticsRepo.GetDraftMatchResults(ctx, session.ID)
		if err == nil {
			sessionResults[session.ID] = results
		}
	}

	// Calculate period boundaries
	now := time.Now()
	periods := t.calculatePeriodBoundaries(periodType, numPeriods, now)

	// Aggregate data by period
	trends := make([]*models.DraftTemporalTrend, 0, len(periods))
	for _, period := range periods {
		trend := &models.DraftTemporalTrend{
			PeriodType:   periodType,
			PeriodStart:  period.start,
			PeriodEnd:    period.end,
			SetCode:      setCode,
			CalculatedAt: now,
		}

		var gradeSum float64
		var gradeCount int

		// Count drafts and matches in this period
		for _, session := range filteredSessions {
			if !session.StartTime.Before(period.start) && session.StartTime.Before(period.end) {
				trend.DraftsCount++

				// Track draft grades
				if session.OverallScore != nil {
					gradeSum += float64(*session.OverallScore)
					gradeCount++
				}

				// Count match results
				if results, ok := sessionResults[session.ID]; ok {
					for _, r := range results {
						trend.MatchesPlayed++
						if r.Result == "win" {
							trend.MatchesWon++
						}
					}
				}
			}
		}

		// Calculate average draft grade
		if gradeCount > 0 {
			avgGrade := gradeSum / float64(gradeCount)
			trend.AvgDraftGrade = &avgGrade
		}

		// Only include periods with data
		if trend.DraftsCount > 0 || trend.MatchesPlayed > 0 {
			trends = append(trends, trend)
		}
	}

	// Save trends to database
	for _, trend := range trends {
		if err := t.analyticsRepo.SaveTemporalTrend(ctx, trend); err != nil {
			// Continue even if one save fails
			continue
		}
	}

	// Sort by period start (ascending for trend analysis)
	sort.Slice(trends, func(i, j int) bool {
		return trends[i].PeriodStart.Before(trends[j].PeriodStart)
	})

	return trends, nil
}

type periodBoundary struct {
	start time.Time
	end   time.Time
}

func (t *TemporalTrendAnalyzer) calculatePeriodBoundaries(periodType string, numPeriods int, now time.Time) []periodBoundary {
	var periods []periodBoundary

	switch periodType {
	case models.PeriodTypeWeek:
		// Start from the beginning of the current week (Sunday)
		weekStart := now.Truncate(24 * time.Hour)
		for weekStart.Weekday() != time.Sunday {
			weekStart = weekStart.Add(-24 * time.Hour)
		}

		for i := 0; i < numPeriods; i++ {
			end := weekStart
			start := weekStart.Add(-7 * 24 * time.Hour)
			periods = append(periods, periodBoundary{start: start, end: end})
			weekStart = start
		}

	case models.PeriodTypeMonth:
		// Start from the beginning of the current month using consistent timezone
		loc := now.Location()
		monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc)

		for i := 0; i < numPeriods; i++ {
			end := monthStart
			// Go back one month
			if monthStart.Month() == time.January {
				monthStart = time.Date(monthStart.Year()-1, time.December, 1, 0, 0, 0, 0, loc)
			} else {
				monthStart = time.Date(monthStart.Year(), monthStart.Month()-1, 1, 0, 0, 0, 0, loc)
			}
			periods = append(periods, periodBoundary{start: monthStart, end: end})
		}
	}

	return periods
}

// GetTrends retrieves cached temporal trends.
func (t *TemporalTrendAnalyzer) GetTrends(ctx context.Context, periodType string, limit int) ([]*models.DraftTemporalTrend, error) {
	return t.analyticsRepo.GetTemporalTrends(ctx, periodType, limit)
}

// GetTrendsBySet retrieves cached temporal trends for a specific set.
func (t *TemporalTrendAnalyzer) GetTrendsBySet(ctx context.Context, setCode, periodType string, limit int) ([]*models.DraftTemporalTrend, error) {
	return t.analyticsRepo.GetTemporalTrendsBySet(ctx, setCode, periodType, limit)
}

// AnalyzeTrendDirection determines if performance is improving, stable, or declining.
func (t *TemporalTrendAnalyzer) AnalyzeTrendDirection(trends []*models.DraftTemporalTrend) models.TrendDirection {
	// Convert from pointer slice to value slice, filtering nil entries
	valueTrends := make([]models.DraftTemporalTrend, 0, len(trends))
	for _, tr := range trends {
		if tr == nil {
			continue
		}
		valueTrends = append(valueTrends, *tr)
	}
	return models.AnalyzeTrendDirection(valueTrends)
}

// TrendAnalysisResponse is the API response for temporal trend analysis.
type TrendAnalysisResponse struct {
	PeriodType string                `json:"periodType"`
	SetCode    *string               `json:"setCode,omitempty"`
	Trends     []*TrendEntry         `json:"trends"`
	Direction  models.TrendDirection `json:"direction"`
	Summary    *TrendSummary         `json:"summary"`
}

// TrendEntry represents a single period's trend data.
type TrendEntry struct {
	PeriodStart   string   `json:"periodStart"`
	PeriodEnd     string   `json:"periodEnd"`
	DraftsCount   int      `json:"draftsCount"`
	MatchesPlayed int      `json:"matchesPlayed"`
	MatchesWon    int      `json:"matchesWon"`
	WinRate       float64  `json:"winRate"`
	AvgDraftGrade *float64 `json:"avgDraftGrade,omitempty"`
}

// TrendSummary provides summary statistics across all periods.
type TrendSummary struct {
	TotalDrafts        int     `json:"totalDrafts"`
	TotalMatches       int     `json:"totalMatches"`
	TotalWins          int     `json:"totalWins"`
	OverallWinRate     float64 `json:"overallWinRate"`
	BestPeriodWinRate  float64 `json:"bestPeriodWinRate"`
	WorstPeriodWinRate float64 `json:"worstPeriodWinRate"`
	WinRateImprovement float64 `json:"winRateImprovement"` // Change from first to last period
}

// BuildTrendAnalysisResponse builds the full API response for temporal trends.
func BuildTrendAnalysisResponse(periodType string, setCode *string, trends []*models.DraftTemporalTrend, direction models.TrendDirection) *TrendAnalysisResponse {
	response := &TrendAnalysisResponse{
		PeriodType: periodType,
		SetCode:    setCode,
		Trends:     make([]*TrendEntry, 0, len(trends)),
		Direction:  direction,
	}

	summary := &TrendSummary{}
	var bestWinRate, worstWinRate float64 = 0, 1

	for _, t := range trends {
		entry := &TrendEntry{
			PeriodStart:   t.PeriodStart.Format("2006-01-02"),
			PeriodEnd:     t.PeriodEnd.Format("2006-01-02"),
			DraftsCount:   t.DraftsCount,
			MatchesPlayed: t.MatchesPlayed,
			MatchesWon:    t.MatchesWon,
			WinRate:       t.WinRate(),
			AvgDraftGrade: t.AvgDraftGrade,
		}
		response.Trends = append(response.Trends, entry)

		// Update summary
		summary.TotalDrafts += t.DraftsCount
		summary.TotalMatches += t.MatchesPlayed
		summary.TotalWins += t.MatchesWon

		winRate := t.WinRate()
		if t.MatchesPlayed > 0 {
			if winRate > bestWinRate {
				bestWinRate = winRate
			}
			if winRate < worstWinRate {
				worstWinRate = winRate
			}
		}
	}

	// Calculate overall stats
	if summary.TotalMatches > 0 {
		summary.OverallWinRate = float64(summary.TotalWins) / float64(summary.TotalMatches)
	}
	summary.BestPeriodWinRate = bestWinRate
	summary.WorstPeriodWinRate = worstWinRate

	// Calculate improvement (last - first period win rate)
	if len(trends) >= 2 {
		firstWinRate := trends[0].WinRate()
		lastWinRate := trends[len(trends)-1].WinRate()
		summary.WinRateImprovement = lastWinRate - firstWinRate
	}

	response.Summary = summary
	return response
}

// LearningCurveResponse shows learning progression for a set.
type LearningCurveResponse struct {
	SetCode     string                 `json:"setCode"`
	Periods     []*LearningPeriodEntry `json:"periods"`
	Improvement float64                `json:"improvement"`
	IsMastered  bool                   `json:"isMastered"` // True if consistently above 55% win rate
}

// LearningPeriodEntry represents a period in the learning curve.
type LearningPeriodEntry struct {
	DraftNumber int     `json:"draftNumber"` // Nth draft in the set
	WinRate     float64 `json:"winRate"`
	Cumulative  float64 `json:"cumulative"` // Cumulative win rate up to this point
}

// BuildLearningCurve builds a learning curve showing improvement over drafts.
func (t *TemporalTrendAnalyzer) BuildLearningCurve(ctx context.Context, setCode string) (*LearningCurveResponse, error) {
	// Get completed sessions for this set
	sessions, err := t.draftRepo.GetCompletedSessions(ctx, 1000)
	if err != nil {
		return nil, err
	}

	// Filter to this set and sort by date
	var setSession []*models.DraftSession
	for _, s := range sessions {
		if s.SetCode == setCode {
			setSession = append(setSession, s)
		}
	}

	sort.Slice(setSession, func(i, j int) bool {
		return setSession[i].StartTime.Before(setSession[j].StartTime)
	})

	response := &LearningCurveResponse{
		SetCode: setCode,
		Periods: make([]*LearningPeriodEntry, 0),
	}

	var totalMatches, totalWins int
	var consecutiveGoodPerformance int

	for i, session := range setSession {
		results, err := t.analyticsRepo.GetDraftMatchResults(ctx, session.ID)
		if err != nil || len(results) == 0 {
			continue
		}

		draftMatches := 0
		draftWins := 0
		for _, r := range results {
			draftMatches++
			if r.Result == "win" {
				draftWins++
			}
		}

		totalMatches += draftMatches
		totalWins += draftWins

		draftWinRate := 0.0
		if draftMatches > 0 {
			draftWinRate = float64(draftWins) / float64(draftMatches)
		}

		cumulative := 0.0
		if totalMatches > 0 {
			cumulative = float64(totalWins) / float64(totalMatches)
		}

		response.Periods = append(response.Periods, &LearningPeriodEntry{
			DraftNumber: i + 1,
			WinRate:     draftWinRate,
			Cumulative:  cumulative,
		})

		// Track consecutive good performance
		if draftWinRate >= 0.55 {
			consecutiveGoodPerformance++
		} else {
			consecutiveGoodPerformance = 0
		}
	}

	// Calculate improvement
	if len(response.Periods) >= 2 {
		firstHalf := response.Periods[:len(response.Periods)/2]
		secondHalf := response.Periods[len(response.Periods)/2:]

		firstHalfWinRate := calculateAverageWinRate(firstHalf)
		secondHalfWinRate := calculateAverageWinRate(secondHalf)
		response.Improvement = secondHalfWinRate - firstHalfWinRate
	}

	// Consider mastered if last 3+ drafts are all >= 55% win rate
	response.IsMastered = consecutiveGoodPerformance >= 3 && len(response.Periods) >= 5

	return response, nil
}

func calculateAverageWinRate(periods []*LearningPeriodEntry) float64 {
	if len(periods) == 0 {
		return 0
	}
	sum := 0.0
	for _, p := range periods {
		sum += p.WinRate
	}
	return sum / float64(len(periods))
}
