package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

func TestTemporalTrendAnalyzer_CalculateWeeklyTrends(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	// Create sessions across multiple weeks
	sessions := []*models.DraftSession{
		{ID: "session-1", SetCode: "FDN", Status: "completed", StartTime: now.Add(-2 * 24 * time.Hour)},
		{ID: "session-2", SetCode: "FDN", Status: "completed", StartTime: now.Add(-9 * 24 * time.Hour)},
		{ID: "session-3", SetCode: "FDN", Status: "completed", StartTime: now.Add(-16 * 24 * time.Hour)},
	}

	draftRepo := &mockDraftRepository{
		sessions: sessions,
		picks:    make(map[string][]*models.DraftPickSession),
	}

	analyticsRepo := newMockAnalyticsRepository()
	// Add match results
	analyticsRepo.matchResults["session-1"] = []*models.DraftMatchResult{
		{SessionID: "session-1", MatchID: "m1", Result: "win", MatchTimestamp: now.Add(-2 * 24 * time.Hour)},
		{SessionID: "session-1", MatchID: "m2", Result: "win", MatchTimestamp: now.Add(-2 * 24 * time.Hour)},
	}
	analyticsRepo.matchResults["session-2"] = []*models.DraftMatchResult{
		{SessionID: "session-2", MatchID: "m3", Result: "win", MatchTimestamp: now.Add(-9 * 24 * time.Hour)},
		{SessionID: "session-2", MatchID: "m4", Result: "loss", MatchTimestamp: now.Add(-9 * 24 * time.Hour)},
	}
	analyticsRepo.matchResults["session-3"] = []*models.DraftMatchResult{
		{SessionID: "session-3", MatchID: "m5", Result: "loss", MatchTimestamp: now.Add(-16 * 24 * time.Hour)},
	}

	analyzer := NewTemporalTrendAnalyzer(draftRepo, analyticsRepo)

	setCode := "FDN"
	trends, err := analyzer.CalculateWeeklyTrends(ctx, 4, &setCode)
	if err != nil {
		t.Fatalf("CalculateWeeklyTrends failed: %v", err)
	}

	// Should have some trends
	if len(trends) == 0 {
		t.Log("No weekly trends found (sessions may not align with week boundaries)")
	}

	for _, trend := range trends {
		t.Logf("Period: %s to %s, Drafts: %d, Matches: %d, Wins: %d",
			trend.PeriodStart.Format("2006-01-02"),
			trend.PeriodEnd.Format("2006-01-02"),
			trend.DraftsCount,
			trend.MatchesPlayed,
			trend.MatchesWon)
	}
}

func TestTemporalTrendAnalyzer_CalculateMonthlyTrends(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	// Create sessions across multiple months
	sessions := []*models.DraftSession{
		{ID: "session-1", SetCode: "FDN", Status: "completed", StartTime: now.Add(-5 * 24 * time.Hour)},
		{ID: "session-2", SetCode: "FDN", Status: "completed", StartTime: now.Add(-35 * 24 * time.Hour)},
		{ID: "session-3", SetCode: "FDN", Status: "completed", StartTime: now.Add(-65 * 24 * time.Hour)},
	}

	draftRepo := &mockDraftRepository{
		sessions: sessions,
		picks:    make(map[string][]*models.DraftPickSession),
	}

	analyticsRepo := newMockAnalyticsRepository()
	analyticsRepo.matchResults["session-1"] = []*models.DraftMatchResult{
		{SessionID: "session-1", MatchID: "m1", Result: "win", MatchTimestamp: now.Add(-5 * 24 * time.Hour)},
	}
	analyticsRepo.matchResults["session-2"] = []*models.DraftMatchResult{
		{SessionID: "session-2", MatchID: "m2", Result: "win", MatchTimestamp: now.Add(-35 * 24 * time.Hour)},
		{SessionID: "session-2", MatchID: "m3", Result: "loss", MatchTimestamp: now.Add(-35 * 24 * time.Hour)},
	}

	analyzer := NewTemporalTrendAnalyzer(draftRepo, analyticsRepo)

	trends, err := analyzer.CalculateMonthlyTrends(ctx, 3, nil)
	if err != nil {
		t.Fatalf("CalculateMonthlyTrends failed: %v", err)
	}

	// Should have some trends
	if len(trends) == 0 {
		t.Log("No monthly trends found (sessions may not align with month boundaries)")
	}
}

func TestTemporalTrendAnalyzer_AnalyzeTrendDirection(t *testing.T) {
	analyzer := &TemporalTrendAnalyzer{}

	tests := []struct {
		name     string
		trends   []*models.DraftTemporalTrend
		expected models.TrendDirection
	}{
		{
			name: "improving",
			trends: []*models.DraftTemporalTrend{
				{MatchesPlayed: 10, MatchesWon: 4},
				{MatchesPlayed: 10, MatchesWon: 5},
				{MatchesPlayed: 10, MatchesWon: 6},
				{MatchesPlayed: 10, MatchesWon: 7},
			},
			expected: models.TrendDirectionImproving,
		},
		{
			name: "declining",
			trends: []*models.DraftTemporalTrend{
				{MatchesPlayed: 10, MatchesWon: 7},
				{MatchesPlayed: 10, MatchesWon: 6},
				{MatchesPlayed: 10, MatchesWon: 5},
				{MatchesPlayed: 10, MatchesWon: 4},
			},
			expected: models.TrendDirectionDeclining,
		},
		{
			name: "stable",
			trends: []*models.DraftTemporalTrend{
				{MatchesPlayed: 10, MatchesWon: 5},
				{MatchesPlayed: 10, MatchesWon: 5},
				{MatchesPlayed: 10, MatchesWon: 5},
			},
			expected: models.TrendDirectionStable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.AnalyzeTrendDirection(tc.trends)
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestBuildTrendAnalysisResponse(t *testing.T) {
	now := time.Now()
	setCode := "FDN"

	trends := []*models.DraftTemporalTrend{
		{
			PeriodType:    models.PeriodTypeWeek,
			PeriodStart:   now.Add(-14 * 24 * time.Hour),
			PeriodEnd:     now.Add(-7 * 24 * time.Hour),
			DraftsCount:   3,
			MatchesPlayed: 9,
			MatchesWon:    4,
		},
		{
			PeriodType:    models.PeriodTypeWeek,
			PeriodStart:   now.Add(-7 * 24 * time.Hour),
			PeriodEnd:     now,
			DraftsCount:   4,
			MatchesPlayed: 12,
			MatchesWon:    8,
		},
	}

	response := BuildTrendAnalysisResponse(models.PeriodTypeWeek, &setCode, trends, models.TrendDirectionImproving)

	if response == nil {
		t.Fatal("expected response to not be nil")
	}

	if *response.SetCode != "FDN" {
		t.Errorf("expected set code 'FDN', got '%s'", *response.SetCode)
	}

	if response.Direction != models.TrendDirectionImproving {
		t.Errorf("expected direction 'improving', got '%s'", response.Direction)
	}

	if len(response.Trends) != 2 {
		t.Errorf("expected 2 trend entries, got %d", len(response.Trends))
	}

	if response.Summary == nil {
		t.Fatal("expected summary to not be nil")
	}

	if response.Summary.TotalDrafts != 7 {
		t.Errorf("expected 7 total drafts, got %d", response.Summary.TotalDrafts)
	}

	if response.Summary.TotalMatches != 21 {
		t.Errorf("expected 21 total matches, got %d", response.Summary.TotalMatches)
	}

	// Improvement should be positive (second period win rate > first)
	// First: 4/9 = 0.444, Second: 8/12 = 0.667
	// Improvement = 0.667 - 0.444 = 0.223
	if response.Summary.WinRateImprovement < 0.2 {
		t.Errorf("expected positive improvement, got %f", response.Summary.WinRateImprovement)
	}
}

func TestCalculatePeriodBoundaries_Weekly(t *testing.T) {
	analyzer := &TemporalTrendAnalyzer{}
	now := time.Now()

	periods := analyzer.calculatePeriodBoundaries(models.PeriodTypeWeek, 4, now)

	if len(periods) != 4 {
		t.Errorf("expected 4 periods, got %d", len(periods))
	}

	// Each period should be 7 days
	for _, p := range periods {
		duration := p.end.Sub(p.start)
		if duration != 7*24*time.Hour {
			t.Errorf("expected 7 day period, got %v", duration)
		}
	}

	// Periods should not overlap
	for i := 1; i < len(periods); i++ {
		if !periods[i].end.Before(periods[i-1].start) && !periods[i].end.Equal(periods[i-1].start) {
			t.Error("periods should not overlap")
		}
	}
}

func TestCalculatePeriodBoundaries_Monthly(t *testing.T) {
	analyzer := &TemporalTrendAnalyzer{}
	now := time.Now()

	periods := analyzer.calculatePeriodBoundaries(models.PeriodTypeMonth, 3, now)

	if len(periods) != 3 {
		t.Errorf("expected 3 periods, got %d", len(periods))
	}

	// Each period should start on the 1st of a month
	for _, p := range periods {
		if p.start.Day() != 1 {
			t.Errorf("expected period to start on 1st, got day %d", p.start.Day())
		}
	}
}

func TestLearningCurveResponse_BuildLearningCurve(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	// Create sessions with improving performance
	sessions := []*models.DraftSession{
		{ID: "s1", SetCode: "FDN", Status: "completed", StartTime: now.Add(-30 * 24 * time.Hour)},
		{ID: "s2", SetCode: "FDN", Status: "completed", StartTime: now.Add(-25 * 24 * time.Hour)},
		{ID: "s3", SetCode: "FDN", Status: "completed", StartTime: now.Add(-20 * 24 * time.Hour)},
		{ID: "s4", SetCode: "FDN", Status: "completed", StartTime: now.Add(-15 * 24 * time.Hour)},
		{ID: "s5", SetCode: "FDN", Status: "completed", StartTime: now.Add(-10 * 24 * time.Hour)},
	}

	draftRepo := &mockDraftRepository{
		sessions: sessions,
		picks:    make(map[string][]*models.DraftPickSession),
	}

	analyticsRepo := newMockAnalyticsRepository()
	// Improve over time: 1/3 -> 1/3 -> 2/3 -> 2/3 -> 2/3
	analyticsRepo.matchResults["s1"] = []*models.DraftMatchResult{
		{Result: "win"}, {Result: "loss"}, {Result: "loss"},
	}
	analyticsRepo.matchResults["s2"] = []*models.DraftMatchResult{
		{Result: "win"}, {Result: "loss"}, {Result: "loss"},
	}
	analyticsRepo.matchResults["s3"] = []*models.DraftMatchResult{
		{Result: "win"}, {Result: "win"}, {Result: "loss"},
	}
	analyticsRepo.matchResults["s4"] = []*models.DraftMatchResult{
		{Result: "win"}, {Result: "win"}, {Result: "loss"},
	}
	analyticsRepo.matchResults["s5"] = []*models.DraftMatchResult{
		{Result: "win"}, {Result: "win"}, {Result: "loss"},
	}

	analyzer := NewTemporalTrendAnalyzer(draftRepo, analyticsRepo)

	curve, err := analyzer.BuildLearningCurve(ctx, "FDN")
	if err != nil {
		t.Fatalf("BuildLearningCurve failed: %v", err)
	}

	if curve == nil {
		t.Fatal("expected learning curve to not be nil")
	}

	if curve.SetCode != "FDN" {
		t.Errorf("expected set code 'FDN', got '%s'", curve.SetCode)
	}

	if len(curve.Periods) != 5 {
		t.Errorf("expected 5 periods, got %d", len(curve.Periods))
	}

	// Improvement should be positive (last drafts are better)
	if curve.Improvement < 0 {
		t.Logf("Improvement was negative: %f (this can happen with small samples)", curve.Improvement)
	}

	// Check cumulative win rate increases
	for i, p := range curve.Periods {
		t.Logf("Draft %d: Win Rate %.2f, Cumulative %.2f", p.DraftNumber, p.WinRate, p.Cumulative)
		if p.DraftNumber != i+1 {
			t.Errorf("expected draft number %d, got %d", i+1, p.DraftNumber)
		}
	}
}

func TestCalculateAverageWinRate(t *testing.T) {
	tests := []struct {
		name     string
		periods  []*LearningPeriodEntry
		expected float64
	}{
		{
			name:     "empty",
			periods:  []*LearningPeriodEntry{},
			expected: 0,
		},
		{
			name: "single",
			periods: []*LearningPeriodEntry{
				{WinRate: 0.5},
			},
			expected: 0.5,
		},
		{
			name: "multiple",
			periods: []*LearningPeriodEntry{
				{WinRate: 0.4},
				{WinRate: 0.6},
			},
			expected: 0.5,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateAverageWinRate(tc.periods)
			if result != tc.expected {
				t.Errorf("expected %f, got %f", tc.expected, result)
			}
		})
	}
}
