package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

func TestCommunityComparisonAnalyzer_CompareToCommunity(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	// Create sessions with match results
	sessions := []*models.DraftSession{
		{ID: "session-1", SetCode: "FDN", Status: "completed", StartTime: now.Add(-24 * time.Hour)},
		{ID: "session-2", SetCode: "FDN", Status: "completed", StartTime: now.Add(-48 * time.Hour)},
	}

	draftRepo := &mockDraftRepository{
		sessions: sessions,
		picks:    make(map[string][]*models.DraftPickSession),
	}

	analyticsRepo := newMockAnalyticsRepository()
	analyticsRepo.matchResults["session-1"] = []*models.DraftMatchResult{
		{SessionID: "session-1", MatchID: "m1", Result: "win", MatchTimestamp: now.Add(-20 * time.Hour)},
		{SessionID: "session-1", MatchID: "m2", Result: "win", MatchTimestamp: now.Add(-19 * time.Hour)},
		{SessionID: "session-1", MatchID: "m3", Result: "loss", MatchTimestamp: now.Add(-18 * time.Hour)},
	}
	analyticsRepo.matchResults["session-2"] = []*models.DraftMatchResult{
		{SessionID: "session-2", MatchID: "m4", Result: "win", MatchTimestamp: now.Add(-44 * time.Hour)},
		{SessionID: "session-2", MatchID: "m5", Result: "win", MatchTimestamp: now.Add(-43 * time.Hour)},
	}

	ratingsProvider := NewDefault17LandsProvider()

	analyzer := NewCommunityComparisonAnalyzer(draftRepo, analyticsRepo, ratingsProvider)

	comparison, err := analyzer.CompareToCommunity(ctx, "FDN", "QuickDraft")
	if err != nil {
		t.Fatalf("CompareToCommunity failed: %v", err)
	}

	if comparison == nil {
		t.Fatal("expected comparison to not be nil")
	}

	if comparison.SetCode != "FDN" {
		t.Errorf("expected set code 'FDN', got '%s'", comparison.SetCode)
	}

	// User has 4 wins out of 5 matches = 80% win rate
	expectedWinRate := 0.8
	if comparison.UserWinRate != expectedWinRate {
		t.Errorf("expected user win rate %f, got %f", expectedWinRate, comparison.UserWinRate)
	}

	// Community average should be around 0.52 for FDN
	if comparison.CommunityAvgWinRate < 0.50 || comparison.CommunityAvgWinRate > 0.55 {
		t.Errorf("expected community win rate around 0.52, got %f", comparison.CommunityAvgWinRate)
	}

	// With 80% win rate, percentile should be very high
	if comparison.PercentileRank == nil {
		t.Error("expected percentile rank to be set")
	} else if *comparison.PercentileRank < 80 {
		t.Errorf("expected high percentile for 80%% win rate, got %f", *comparison.PercentileRank)
	}

	if comparison.SampleSize != 5 {
		t.Errorf("expected sample size 5, got %d", comparison.SampleSize)
	}
}

func TestCommunityComparisonAnalyzer_CompareToCommunity_NoData(t *testing.T) {
	ctx := context.Background()

	draftRepo := &mockDraftRepository{
		sessions: []*models.DraftSession{},
		picks:    make(map[string][]*models.DraftPickSession),
	}

	analyticsRepo := newMockAnalyticsRepository()
	ratingsProvider := NewDefault17LandsProvider()

	analyzer := NewCommunityComparisonAnalyzer(draftRepo, analyticsRepo, ratingsProvider)

	comparison, err := analyzer.CompareToCommunity(ctx, "FDN", "QuickDraft")
	if err != nil {
		t.Fatalf("CompareToCommunity failed: %v", err)
	}

	// Should return nil when no data
	if comparison != nil {
		t.Error("expected nil comparison when no data")
	}
}

func TestCommunityComparisonAnalyzer_CompareToCommunity_NilProvider(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	sessions := []*models.DraftSession{
		{ID: "session-1", SetCode: "FDN", Status: "completed", StartTime: now.Add(-24 * time.Hour)},
	}

	draftRepo := &mockDraftRepository{
		sessions: sessions,
		picks:    make(map[string][]*models.DraftPickSession),
	}

	analyticsRepo := newMockAnalyticsRepository()
	analyticsRepo.matchResults["session-1"] = []*models.DraftMatchResult{
		{SessionID: "session-1", MatchID: "m1", Result: "win", MatchTimestamp: now.Add(-20 * time.Hour)},
		{SessionID: "session-1", MatchID: "m2", Result: "loss", MatchTimestamp: now.Add(-19 * time.Hour)},
	}

	// Test with nil ratings provider
	analyzer := NewCommunityComparisonAnalyzer(draftRepo, analyticsRepo, nil)

	comparison, err := analyzer.CompareToCommunity(ctx, "FDN", "QuickDraft")
	if err != nil {
		t.Fatalf("CompareToCommunity failed: %v", err)
	}

	if comparison == nil {
		t.Fatal("expected comparison to not be nil")
	}

	// Should use default community rate of 0.52
	if comparison.CommunityAvgWinRate != 0.52 {
		t.Errorf("expected default community win rate 0.52, got %f", comparison.CommunityAvgWinRate)
	}
}

func TestCommunityComparisonAnalyzer_CompareArchetypePerformance(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	analyticsRepo := newMockAnalyticsRepository()

	// Pre-populate with archetype stats
	stats := []*models.DraftArchetypeStats{
		{ID: 1, SetCode: "FDN", ColorCombination: "WG", ArchetypeName: "Selesnya", MatchesPlayed: 10, MatchesWon: 7, DraftsCount: 3, UpdatedAt: now},
		{ID: 2, SetCode: "FDN", ColorCombination: "UB", ArchetypeName: "Dimir", MatchesPlayed: 8, MatchesWon: 3, DraftsCount: 2, UpdatedAt: now},
	}

	for _, s := range stats {
		analyticsRepo.UpsertArchetypeStats(ctx, s)
	}

	draftRepo := &mockDraftRepository{}
	ratingsProvider := NewDefault17LandsProvider()

	analyzer := NewCommunityComparisonAnalyzer(draftRepo, analyticsRepo, ratingsProvider)

	entries, err := analyzer.CompareArchetypePerformance(ctx, "FDN", "QuickDraft")
	if err != nil {
		t.Fatalf("CompareArchetypePerformance failed: %v", err)
	}

	if len(entries) != 2 {
		t.Errorf("expected 2 archetype entries, got %d", len(entries))
	}

	// Find Selesnya entry and verify
	var selesnya *ArchetypeComparisonEntry
	for _, e := range entries {
		if e.ColorCombination == "WG" {
			selesnya = e
			break
		}
	}

	if selesnya == nil {
		t.Fatal("expected to find WG archetype")
	}

	// 7/10 = 70% win rate
	if selesnya.UserWinRate != 0.7 {
		t.Errorf("expected user win rate 0.7, got %f", selesnya.UserWinRate)
	}

	// Should be above community (70% > ~52%)
	if !selesnya.IsAboveCommunity {
		t.Error("expected Selesnya to be above community average")
	}

	if selesnya.WinRateDelta <= 0 {
		t.Errorf("expected positive win rate delta, got %f", selesnya.WinRateDelta)
	}
}

func TestCommunityComparisonAnalyzer_CalculatePercentile(t *testing.T) {
	analyzer := &CommunityComparisonAnalyzer{}

	tests := []struct {
		name         string
		userWinRate  float64
		communityAvg float64
		minExpected  float64
		maxExpected  float64
	}{
		{
			name:         "average player",
			userWinRate:  0.50,
			communityAvg: 0.52,
			minExpected:  40,
			maxExpected:  60,
		},
		{
			name:         "strong player",
			userWinRate:  0.60,
			communityAvg: 0.52,
			minExpected:  70,
			maxExpected:  90,
		},
		{
			name:         "exceptional player",
			userWinRate:  0.70,
			communityAvg: 0.52,
			minExpected:  90,
			maxExpected:  99,
		},
		{
			name:         "struggling player",
			userWinRate:  0.40,
			communityAvg: 0.52,
			minExpected:  1,  // (0.40 - 0.52) * 300 + 50 = 14, rounds to ~14
			maxExpected:  20, // Allow some tolerance
		},
		{
			name:         "very low win rate",
			userWinRate:  0.25,
			communityAvg: 0.52,
			minExpected:  1,
			maxExpected:  10,
		},
		{
			name:         "very high win rate - clamped",
			userWinRate:  0.90,
			communityAvg: 0.52,
			minExpected:  99,
			maxExpected:  99,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			percentile := analyzer.calculatePercentile(tc.userWinRate, tc.communityAvg)
			if percentile < tc.minExpected || percentile > tc.maxExpected {
				t.Errorf("expected percentile between %f and %f, got %f", tc.minExpected, tc.maxExpected, percentile)
			}
		})
	}
}

func TestGetRankLabel(t *testing.T) {
	tests := []struct {
		percentile float64
		expected   string
	}{
		{99, "Top 5%"},
		{95, "Top 5%"},
		{92, "Top 10%"},
		{90, "Top 10%"},
		{85, "Top 20%"},
		{80, "Top 20%"},
		{70, "Above Average"},
		{60, "Above Average"},
		{55, "Average"},
		{50, "Average"},
		{40, "Average"},
		{35, "Below Average"},
		{20, "Below Average"},
		{15, "Needs Improvement"},
		{5, "Needs Improvement"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := getRankLabel(tc.percentile)
			if result != tc.expected {
				t.Errorf("getRankLabel(%f) = %s, want %s", tc.percentile, result, tc.expected)
			}
		})
	}
}

func TestBuildCommunityComparisonResponse(t *testing.T) {
	now := time.Now()
	percentile := 75.0

	comparison := &models.DraftCommunityComparison{
		SetCode:             "FDN",
		DraftFormat:         "QuickDraft",
		UserWinRate:         0.58,
		CommunityAvgWinRate: 0.52,
		PercentileRank:      &percentile,
		SampleSize:          25,
		CalculatedAt:        now,
	}

	archetypeComparison := []*ArchetypeComparisonEntry{
		{
			ColorCombination:  "WG",
			ArchetypeName:     "Selesnya",
			UserWinRate:       0.65,
			CommunityWinRate:  0.54,
			WinRateDelta:      0.11,
			UserMatchesPlayed: 10,
			PercentileRank:    80,
			IsAboveCommunity:  true,
		},
	}

	response := BuildCommunityComparisonResponse(comparison, archetypeComparison)

	if response == nil {
		t.Fatal("expected response to not be nil")
	}

	if response.SetCode != "FDN" {
		t.Errorf("expected set code 'FDN', got '%s'", response.SetCode)
	}

	if response.DraftFormat != "QuickDraft" {
		t.Errorf("expected draft format 'QuickDraft', got '%s'", response.DraftFormat)
	}

	if response.UserWinRate != 0.58 {
		t.Errorf("expected user win rate 0.58, got %f", response.UserWinRate)
	}

	if response.CommunityAvgWinRate != 0.52 {
		t.Errorf("expected community win rate 0.52, got %f", response.CommunityAvgWinRate)
	}

	// Win rate delta should be 0.06
	expectedDelta := 0.06
	if response.WinRateDelta < expectedDelta-0.01 || response.WinRateDelta > expectedDelta+0.01 {
		t.Errorf("expected win rate delta around %f, got %f", expectedDelta, response.WinRateDelta)
	}

	if response.PercentileRank != 75.0 {
		t.Errorf("expected percentile 75.0, got %f", response.PercentileRank)
	}

	if response.Rank != "Above Average" {
		t.Errorf("expected rank 'Above Average', got '%s'", response.Rank)
	}

	if len(response.ArchetypeComparison) != 1 {
		t.Errorf("expected 1 archetype comparison, got %d", len(response.ArchetypeComparison))
	}
}

func TestBuildCommunityComparisonResponse_NilInput(t *testing.T) {
	response := BuildCommunityComparisonResponse(nil, nil)
	if response != nil {
		t.Error("expected nil response for nil input")
	}
}

func TestBuildCommunityComparisonResponse_NilPercentile(t *testing.T) {
	comparison := &models.DraftCommunityComparison{
		SetCode:             "FDN",
		DraftFormat:         "QuickDraft",
		UserWinRate:         0.50,
		CommunityAvgWinRate: 0.52,
		PercentileRank:      nil, // nil percentile
		SampleSize:          5,
	}

	response := BuildCommunityComparisonResponse(comparison, nil)

	if response == nil {
		t.Fatal("expected response to not be nil")
	}

	// Should default to 50.0 when percentile is nil
	if response.PercentileRank != 50.0 {
		t.Errorf("expected default percentile 50.0, got %f", response.PercentileRank)
	}
}

func TestDefault17LandsProvider_GetSetAverageWinRate(t *testing.T) {
	provider := NewDefault17LandsProvider()

	tests := []struct {
		setCode  string
		expected float64
	}{
		{"FDN", 0.52},
		{"TLA", 0.51},
		{"DSK", 0.52},
		{"MH3", 0.51},
		{"OTJ", 0.52},
		{"UNKNOWN", 0.52}, // Default
	}

	for _, tc := range tests {
		t.Run(tc.setCode, func(t *testing.T) {
			rate, err := provider.GetSetAverageWinRate(tc.setCode, "QuickDraft")
			if err != nil {
				t.Fatalf("GetSetAverageWinRate failed: %v", err)
			}
			if rate != tc.expected {
				t.Errorf("expected %f, got %f", tc.expected, rate)
			}
		})
	}
}

func TestDefault17LandsProvider_GetColorPairWinRates(t *testing.T) {
	provider := NewDefault17LandsProvider()

	// Test with known set
	rates, err := provider.GetColorPairWinRates("FDN", "QuickDraft")
	if err != nil {
		t.Fatalf("GetColorPairWinRates failed: %v", err)
	}

	if len(rates) == 0 {
		t.Error("expected non-empty color pair rates")
	}

	// Verify some expected color pairs
	if _, ok := rates["WU"]; !ok {
		t.Error("expected WU color pair rate")
	}

	// Test with unknown set - should return defaults
	unknownRates, err := provider.GetColorPairWinRates("UNKNOWN", "QuickDraft")
	if err != nil {
		t.Fatalf("GetColorPairWinRates for unknown set failed: %v", err)
	}

	if len(unknownRates) != 10 {
		t.Errorf("expected 10 default color pairs, got %d", len(unknownRates))
	}
}

func TestCommunityComparisonAnalyzer_GetCommunityComparison(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	analyticsRepo := newMockAnalyticsRepository()

	// Pre-save a comparison
	percentile := 65.0
	comparison := &models.DraftCommunityComparison{
		SetCode:             "FDN",
		DraftFormat:         "QuickDraft",
		UserWinRate:         0.55,
		CommunityAvgWinRate: 0.52,
		PercentileRank:      &percentile,
		SampleSize:          20,
		CalculatedAt:        now,
	}
	analyticsRepo.SaveCommunityComparison(ctx, comparison)

	analyzer := NewCommunityComparisonAnalyzer(nil, analyticsRepo, nil)

	retrieved, err := analyzer.GetCommunityComparison(ctx, "FDN", "QuickDraft")
	if err != nil {
		t.Fatalf("GetCommunityComparison failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected comparison to not be nil")
	}

	if retrieved.SetCode != "FDN" {
		t.Errorf("expected set code 'FDN', got '%s'", retrieved.SetCode)
	}

	if retrieved.UserWinRate != 0.55 {
		t.Errorf("expected user win rate 0.55, got %f", retrieved.UserWinRate)
	}
}

func TestCommunityComparisonAnalyzer_GetAllComparisons(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	analyticsRepo := newMockAnalyticsRepository()

	// Pre-save multiple comparisons
	percentile1 := 65.0
	percentile2 := 70.0
	comparisons := []*models.DraftCommunityComparison{
		{
			SetCode:             "FDN",
			DraftFormat:         "QuickDraft",
			UserWinRate:         0.55,
			CommunityAvgWinRate: 0.52,
			PercentileRank:      &percentile1,
			SampleSize:          20,
			CalculatedAt:        now,
		},
		{
			SetCode:             "TLA",
			DraftFormat:         "PremierDraft",
			UserWinRate:         0.58,
			CommunityAvgWinRate: 0.51,
			PercentileRank:      &percentile2,
			SampleSize:          15,
			CalculatedAt:        now,
		},
	}

	for _, c := range comparisons {
		analyticsRepo.SaveCommunityComparison(ctx, c)
	}

	analyzer := NewCommunityComparisonAnalyzer(nil, analyticsRepo, nil)

	all, err := analyzer.GetAllComparisons(ctx)
	if err != nil {
		t.Fatalf("GetAllComparisons failed: %v", err)
	}

	if len(all) != 2 {
		t.Errorf("expected 2 comparisons, got %d", len(all))
	}
}
