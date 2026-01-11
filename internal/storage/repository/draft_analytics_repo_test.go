package repository

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// setupDraftAnalyticsTestDB creates an in-memory database with draft analytics tables.
func setupDraftAnalyticsTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE draft_sessions (
			id TEXT PRIMARY KEY,
			event_name TEXT NOT NULL,
			set_code TEXT NOT NULL,
			draft_type TEXT DEFAULT 'quick_draft',
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP,
			status TEXT DEFAULT 'in_progress',
			total_picks INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE draft_match_results (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			match_id TEXT NOT NULL,
			result TEXT NOT NULL CHECK(result IN ('win', 'loss')),
			opponent_colors TEXT,
			game_wins INTEGER DEFAULT 0,
			game_losses INTEGER DEFAULT 0,
			match_timestamp TIMESTAMP NOT NULL,
			FOREIGN KEY (session_id) REFERENCES draft_sessions(id) ON DELETE CASCADE,
			UNIQUE(session_id, match_id)
		);

		CREATE TABLE draft_archetype_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			set_code TEXT NOT NULL,
			color_combination TEXT NOT NULL,
			archetype_name TEXT NOT NULL,
			matches_played INTEGER DEFAULT 0,
			matches_won INTEGER DEFAULT 0,
			drafts_count INTEGER DEFAULT 0,
			avg_draft_grade REAL,
			last_played_at TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(set_code, color_combination)
		);

		CREATE TABLE draft_community_comparison (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			set_code TEXT NOT NULL,
			draft_format TEXT NOT NULL,
			user_win_rate REAL NOT NULL,
			community_avg_win_rate REAL NOT NULL,
			percentile_rank REAL,
			sample_size INTEGER,
			calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(set_code, draft_format)
		);

		CREATE TABLE draft_temporal_trends (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			period_type TEXT NOT NULL CHECK(period_type IN ('week', 'month')),
			period_start DATE NOT NULL,
			period_end DATE NOT NULL,
			set_code TEXT,
			drafts_count INTEGER DEFAULT 0,
			matches_played INTEGER DEFAULT 0,
			matches_won INTEGER DEFAULT 0,
			avg_draft_grade REAL,
			calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(period_type, period_start, set_code)
		);

		CREATE TABLE draft_pattern_analysis (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			set_code TEXT,
			color_preference_json TEXT,
			type_preference_json TEXT,
			pick_order_pattern_json TEXT,
			archetype_affinity_json TEXT,
			sample_size INTEGER,
			calculated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(set_code)
		);

		CREATE INDEX idx_draft_match_results_session ON draft_match_results(session_id);
		CREATE INDEX idx_draft_match_results_timestamp ON draft_match_results(match_timestamp);
		CREATE INDEX idx_draft_archetype_stats_set ON draft_archetype_stats(set_code);
		CREATE INDEX idx_draft_temporal_trends_period ON draft_temporal_trends(period_type, period_start);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestDraftAnalyticsRepository_SaveDraftMatchResult(t *testing.T) {
	db := setupDraftAnalyticsTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftAnalyticsRepository(db)
	ctx := context.Background()
	now := time.Now()

	// Create a draft session first
	_, err := db.Exec(`INSERT INTO draft_sessions (id, event_name, set_code, start_time) VALUES (?, ?, ?, ?)`,
		"session-1", "Quick Draft FDN", "FDN", now)
	if err != nil {
		t.Fatalf("failed to create draft session: %v", err)
	}

	result := &models.DraftMatchResult{
		SessionID:      "session-1",
		MatchID:        "match-1",
		Result:         "win",
		OpponentColors: "UB",
		GameWins:       2,
		GameLosses:     1,
		MatchTimestamp: now,
	}

	err = repo.SaveDraftMatchResult(ctx, result)
	if err != nil {
		t.Fatalf("failed to save draft match result: %v", err)
	}

	// Verify it was saved
	results, err := repo.GetDraftMatchResults(ctx, "session-1")
	if err != nil {
		t.Fatalf("failed to get draft match results: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	if results[0].Result != "win" {
		t.Errorf("expected result 'win', got '%s'", results[0].Result)
	}

	if results[0].GameWins != 2 {
		t.Errorf("expected 2 game wins, got %d", results[0].GameWins)
	}
}

func TestDraftAnalyticsRepository_GetDraftMatchResultsByTimeRange(t *testing.T) {
	db := setupDraftAnalyticsTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftAnalyticsRepository(db)
	ctx := context.Background()

	// Create draft session
	now := time.Now()
	_, err := db.Exec(`INSERT INTO draft_sessions (id, event_name, set_code, start_time) VALUES (?, ?, ?, ?)`,
		"session-1", "Quick Draft", "FDN", now)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Save match results at different times
	yesterday := now.Add(-24 * time.Hour)
	lastWeek := now.Add(-7 * 24 * time.Hour)

	results := []*models.DraftMatchResult{
		{SessionID: "session-1", MatchID: "match-1", Result: "win", MatchTimestamp: now},
		{SessionID: "session-1", MatchID: "match-2", Result: "loss", MatchTimestamp: yesterday},
		{SessionID: "session-1", MatchID: "match-3", Result: "win", MatchTimestamp: lastWeek},
	}

	for _, r := range results {
		if err := repo.SaveDraftMatchResult(ctx, r); err != nil {
			t.Fatalf("failed to save result: %v", err)
		}
	}

	// Get results from last 3 days
	start := now.Add(-3 * 24 * time.Hour)
	recentResults, err := repo.GetDraftMatchResultsByTimeRange(ctx, start, now)
	if err != nil {
		t.Fatalf("failed to get results by time range: %v", err)
	}

	if len(recentResults) != 2 {
		t.Errorf("expected 2 results in last 3 days, got %d", len(recentResults))
	}
}

func TestDraftAnalyticsRepository_ArchetypeStats(t *testing.T) {
	db := setupDraftAnalyticsTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftAnalyticsRepository(db)
	ctx := context.Background()
	now := time.Now()

	stats := []*models.DraftArchetypeStats{
		{
			SetCode:          "FDN",
			ColorCombination: "WU",
			ArchetypeName:    "Azorius Flyers",
			MatchesPlayed:    10,
			MatchesWon:       7,
			DraftsCount:      3,
			UpdatedAt:        now,
		},
		{
			SetCode:          "FDN",
			ColorCombination: "BR",
			ArchetypeName:    "Rakdos Sacrifice",
			MatchesPlayed:    8,
			MatchesWon:       3,
			DraftsCount:      2,
			UpdatedAt:        now,
		},
		{
			SetCode:          "FDN",
			ColorCombination: "GW",
			ArchetypeName:    "Selesnya Tokens",
			MatchesPlayed:    15,
			MatchesWon:       12,
			DraftsCount:      5,
			UpdatedAt:        now,
		},
	}

	// Save all stats
	for _, s := range stats {
		if err := repo.UpsertArchetypeStats(ctx, s); err != nil {
			t.Fatalf("failed to upsert archetype stats: %v", err)
		}
	}

	// Get stats for FDN
	fdnStats, err := repo.GetArchetypeStats(ctx, "FDN")
	if err != nil {
		t.Fatalf("failed to get archetype stats: %v", err)
	}

	if len(fdnStats) != 3 {
		t.Errorf("expected 3 archetypes, got %d", len(fdnStats))
	}

	// Test best archetypes (min 5 matches)
	best, err := repo.GetBestArchetypes(ctx, 5, 2)
	if err != nil {
		t.Fatalf("failed to get best archetypes: %v", err)
	}

	if len(best) != 2 {
		t.Errorf("expected 2 best archetypes, got %d", len(best))
	}

	// GW should be first (80% win rate)
	if best[0].ColorCombination != "GW" {
		t.Errorf("expected GW to be best, got %s", best[0].ColorCombination)
	}

	// Test worst archetypes
	worst, err := repo.GetWorstArchetypes(ctx, 5, 1)
	if err != nil {
		t.Fatalf("failed to get worst archetypes: %v", err)
	}

	if len(worst) != 1 {
		t.Errorf("expected 1 worst archetype, got %d", len(worst))
	}

	// BR should be worst (37.5% win rate)
	if worst[0].ColorCombination != "BR" {
		t.Errorf("expected BR to be worst, got %s", worst[0].ColorCombination)
	}
}

func TestDraftAnalyticsRepository_TemporalTrends(t *testing.T) {
	db := setupDraftAnalyticsTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftAnalyticsRepository(db)
	ctx := context.Background()
	now := time.Now()

	// Create weekly trends
	trends := []*models.DraftTemporalTrend{
		{
			PeriodType:    models.PeriodTypeWeek,
			PeriodStart:   now.Add(-14 * 24 * time.Hour),
			PeriodEnd:     now.Add(-7 * 24 * time.Hour),
			DraftsCount:   5,
			MatchesPlayed: 15,
			MatchesWon:    8,
			CalculatedAt:  now,
		},
		{
			PeriodType:    models.PeriodTypeWeek,
			PeriodStart:   now.Add(-7 * 24 * time.Hour),
			PeriodEnd:     now,
			DraftsCount:   7,
			MatchesPlayed: 21,
			MatchesWon:    14,
			CalculatedAt:  now,
		},
	}

	for _, trend := range trends {
		if err := repo.SaveTemporalTrend(ctx, trend); err != nil {
			t.Fatalf("failed to save temporal trend: %v", err)
		}
	}

	// Get weekly trends
	retrieved, err := repo.GetTemporalTrends(ctx, models.PeriodTypeWeek, 10)
	if err != nil {
		t.Fatalf("failed to get temporal trends: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("expected 2 trends, got %d", len(retrieved))
	}

	// Most recent should be first (descending order)
	if retrieved[0].DraftsCount != 7 {
		t.Errorf("expected most recent trend to have 7 drafts, got %d", retrieved[0].DraftsCount)
	}

	// Test win rate calculation
	if retrieved[0].WinRate() < 0.66 || retrieved[0].WinRate() > 0.67 {
		t.Errorf("expected win rate ~66.67%%, got %.2f%%", retrieved[0].WinRate()*100)
	}
}

func TestDraftAnalyticsRepository_PatternAnalysis(t *testing.T) {
	db := setupDraftAnalyticsTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftAnalyticsRepository(db)
	ctx := context.Background()
	now := time.Now()
	setCode := "FDN"

	analysis := &models.DraftPatternAnalysis{
		SetCode:               &setCode,
		ColorPreferenceJSON:   `[{"color":"G","totalPicks":50,"percentOfPool":30.5,"avgPickOrder":4.2}]`,
		TypePreferenceJSON:    `[{"type":"Creature","totalPicks":80,"percentOfPool":60.0,"avgPickOrder":3.5}]`,
		PickOrderPatternJSON:  `[{"phase":"early","avgRating":4.5,"totalPicks":42}]`,
		ArchetypeAffinityJSON: `[{"colorPair":"GW","archetypeName":"Selesnya Tokens","timesBuilt":5}]`,
		SampleSize:            10,
		CalculatedAt:          now,
	}

	// Save pattern analysis
	err := repo.SavePatternAnalysis(ctx, analysis)
	if err != nil {
		t.Fatalf("failed to save pattern analysis: %v", err)
	}

	// Retrieve it
	retrieved, err := repo.GetPatternAnalysis(ctx, &setCode)
	if err != nil {
		t.Fatalf("failed to get pattern analysis: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected pattern analysis to be found")
	}

	if retrieved.SampleSize != 10 {
		t.Errorf("expected sample size 10, got %d", retrieved.SampleSize)
	}

	// Test JSON parsing
	colorPrefs, err := retrieved.GetColorPreferences()
	if err != nil {
		t.Fatalf("failed to parse color preferences: %v", err)
	}

	if len(colorPrefs) != 1 {
		t.Errorf("expected 1 color preference, got %d", len(colorPrefs))
	}

	if colorPrefs[0].Color != "G" {
		t.Errorf("expected color 'G', got '%s'", colorPrefs[0].Color)
	}
}

func TestDraftAnalyticsRepository_CommunityComparison(t *testing.T) {
	db := setupDraftAnalyticsTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftAnalyticsRepository(db)
	ctx := context.Background()
	now := time.Now()

	percentile := 72.5
	comparison := &models.DraftCommunityComparison{
		SetCode:             "FDN",
		DraftFormat:         "PremierDraft",
		UserWinRate:         0.58,
		CommunityAvgWinRate: 0.52,
		PercentileRank:      &percentile,
		SampleSize:          25,
		CalculatedAt:        now,
	}

	// Save comparison
	err := repo.SaveCommunityComparison(ctx, comparison)
	if err != nil {
		t.Fatalf("failed to save community comparison: %v", err)
	}

	// Retrieve it
	retrieved, err := repo.GetCommunityComparison(ctx, "FDN", "PremierDraft")
	if err != nil {
		t.Fatalf("failed to get community comparison: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected community comparison to be found")
	}

	if retrieved.UserWinRate != 0.58 {
		t.Errorf("expected user win rate 0.58, got %f", retrieved.UserWinRate)
	}

	if retrieved.PercentileRank == nil {
		t.Error("expected percentile rank to be set")
	} else if *retrieved.PercentileRank != 72.5 {
		t.Errorf("expected percentile 72.5, got %f", *retrieved.PercentileRank)
	}

	// Test delta calculation
	delta := retrieved.WinRateDelta()
	if delta < 0.05 || delta > 0.07 {
		t.Errorf("expected delta ~0.06, got %f", delta)
	}

	// Test get all
	all, err := repo.GetAllCommunityComparisons(ctx)
	if err != nil {
		t.Fatalf("failed to get all community comparisons: %v", err)
	}

	if len(all) != 1 {
		t.Errorf("expected 1 comparison, got %d", len(all))
	}
}

func TestDraftAnalyticsRepository_UpsertArchetypeStats(t *testing.T) {
	db := setupDraftAnalyticsTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := NewDraftAnalyticsRepository(db)
	ctx := context.Background()
	now := time.Now()

	// Initial save
	stats := &models.DraftArchetypeStats{
		SetCode:          "FDN",
		ColorCombination: "WU",
		ArchetypeName:    "Azorius Flyers",
		MatchesPlayed:    10,
		MatchesWon:       7,
		DraftsCount:      3,
		UpdatedAt:        now,
	}

	err := repo.UpsertArchetypeStats(ctx, stats)
	if err != nil {
		t.Fatalf("failed to save archetype stats: %v", err)
	}

	// Update with new stats
	stats.MatchesPlayed = 15
	stats.MatchesWon = 10
	stats.DraftsCount = 5
	stats.UpdatedAt = now.Add(time.Hour)

	err = repo.UpsertArchetypeStats(ctx, stats)
	if err != nil {
		t.Fatalf("failed to upsert archetype stats: %v", err)
	}

	// Verify update
	retrieved, err := repo.GetArchetypeStats(ctx, "FDN")
	if err != nil {
		t.Fatalf("failed to get archetype stats: %v", err)
	}

	if len(retrieved) != 1 {
		t.Fatalf("expected 1 archetype, got %d", len(retrieved))
	}

	if retrieved[0].MatchesPlayed != 15 {
		t.Errorf("expected 15 matches played, got %d", retrieved[0].MatchesPlayed)
	}

	if retrieved[0].DraftsCount != 5 {
		t.Errorf("expected 5 drafts, got %d", retrieved[0].DraftsCount)
	}
}

func TestModels_AnalyzeTrendDirection(t *testing.T) {
	tests := []struct {
		name     string
		trends   []models.DraftTemporalTrend
		expected models.TrendDirection
	}{
		{
			name: "improving trend",
			trends: []models.DraftTemporalTrend{
				{MatchesPlayed: 10, MatchesWon: 4}, // 40%
				{MatchesPlayed: 10, MatchesWon: 5}, // 50%
				{MatchesPlayed: 10, MatchesWon: 6}, // 60%
			},
			expected: models.TrendDirectionImproving,
		},
		{
			name: "declining trend",
			trends: []models.DraftTemporalTrend{
				{MatchesPlayed: 10, MatchesWon: 6}, // 60%
				{MatchesPlayed: 10, MatchesWon: 5}, // 50%
				{MatchesPlayed: 10, MatchesWon: 4}, // 40%
			},
			expected: models.TrendDirectionDeclining,
		},
		{
			name: "stable trend",
			trends: []models.DraftTemporalTrend{
				{MatchesPlayed: 10, MatchesWon: 5}, // 50%
				{MatchesPlayed: 10, MatchesWon: 5}, // 50%
				{MatchesPlayed: 10, MatchesWon: 5}, // 50%
			},
			expected: models.TrendDirectionStable,
		},
		{
			name:     "single period (stable)",
			trends:   []models.DraftTemporalTrend{{MatchesPlayed: 10, MatchesWon: 5}},
			expected: models.TrendDirectionStable,
		},
		{
			name:     "empty (stable)",
			trends:   []models.DraftTemporalTrend{},
			expected: models.TrendDirectionStable,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := models.AnalyzeTrendDirection(tc.trends)
			if result != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestModels_DraftArchetypeStats_WinRate(t *testing.T) {
	tests := []struct {
		name     string
		played   int
		won      int
		expected float64
	}{
		{"60% win rate", 10, 6, 0.6},
		{"zero matches", 0, 0, 0.0},
		{"100% win rate", 5, 5, 1.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stats := &models.DraftArchetypeStats{
				MatchesPlayed: tc.played,
				MatchesWon:    tc.won,
			}
			if stats.WinRate() != tc.expected {
				t.Errorf("expected win rate %f, got %f", tc.expected, stats.WinRate())
			}
		})
	}
}

func TestModels_DraftMatchResult_WinRate(t *testing.T) {
	winResult := &models.DraftMatchResult{Result: "win"}
	if winResult.WinRate() != 1.0 {
		t.Errorf("expected win rate 1.0 for win, got %f", winResult.WinRate())
	}

	lossResult := &models.DraftMatchResult{Result: "loss"}
	if lossResult.WinRate() != 0.0 {
		t.Errorf("expected win rate 0.0 for loss, got %f", lossResult.WinRate())
	}
}
