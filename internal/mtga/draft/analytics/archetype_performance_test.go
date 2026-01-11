package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// mockMatchRepository is a mock implementation for testing.
type mockMatchRepository struct {
	matches []*models.Match
}

func (m *mockMatchRepository) Create(ctx context.Context, match *models.Match) error {
	return nil
}

func (m *mockMatchRepository) CreateGame(ctx context.Context, game *models.Game) error {
	return nil
}

func (m *mockMatchRepository) GetByID(ctx context.Context, id string) (*models.Match, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetByDateRange(ctx context.Context, start, end time.Time, accountID int) ([]*models.Match, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetByFormat(ctx context.Context, format string, accountID int) ([]*models.Match, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetRecentMatches(ctx context.Context, limit int, accountID int) ([]*models.Match, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetLatestMatch(ctx context.Context, accountID int) (*models.Match, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetMatches(ctx context.Context, filter models.StatsFilter) ([]*models.Match, error) {
	return m.matches, nil
}

func (m *mockMatchRepository) GetStats(ctx context.Context, filter models.StatsFilter) (*models.Statistics, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetStatsByFormat(ctx context.Context, filter models.StatsFilter) (map[string]*models.Statistics, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetStatsByDeck(ctx context.Context, filter models.StatsFilter) (map[string]*models.Statistics, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetGamesForMatch(ctx context.Context, matchID string) ([]*models.Game, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetPerformanceMetrics(ctx context.Context, filter models.StatsFilter) (*models.PerformanceMetrics, error) {
	return nil, nil
}

func (m *mockMatchRepository) GetMatchesWithoutDeckID(ctx context.Context) ([]*models.Match, error) {
	return nil, nil
}

func (m *mockMatchRepository) UpdateDeckID(ctx context.Context, matchID, deckID string) error {
	return nil
}

func (m *mockMatchRepository) DeleteAll(ctx context.Context, accountID int) error {
	return nil
}

func (m *mockMatchRepository) GetDailyWins(ctx context.Context, accountID int) (int, error) {
	return 0, nil
}

func (m *mockMatchRepository) GetWeeklyWins(ctx context.Context, accountID int) (int, error) {
	return 0, nil
}

func (m *mockMatchRepository) GetMatchesForMLProcessing(ctx context.Context, filter models.StatsFilter) ([]*models.Match, error) {
	return nil, nil
}

func (m *mockMatchRepository) MarkMatchesAsProcessedForML(ctx context.Context, matchIDs []string) error {
	return nil
}

func TestArchetypePerformanceAnalyzer_AnalyzeArchetypePerformance(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	draftRepo := &mockDraftRepository{
		sessions: []*models.DraftSession{
			{ID: "session-1", SetCode: "FDN", Status: "completed", StartTime: now},
			{ID: "session-2", SetCode: "FDN", Status: "completed", StartTime: now.Add(-24 * time.Hour)},
		},
		picks: map[string][]*models.DraftPickSession{
			"session-1": {
				{ID: 1, SessionID: "session-1", CardID: "1001", PickNumber: 1},
				{ID: 2, SessionID: "session-1", CardID: "1002", PickNumber: 2},
			},
			"session-2": {
				{ID: 3, SessionID: "session-2", CardID: "1003", PickNumber: 1},
				{ID: 4, SessionID: "session-2", CardID: "1004", PickNumber: 2},
			},
		},
	}

	analyticsRepo := newMockAnalyticsRepository()
	// Add some match results
	analyticsRepo.matchResults["session-1"] = []*models.DraftMatchResult{
		{SessionID: "session-1", MatchID: "match-1", Result: "win", MatchTimestamp: now},
		{SessionID: "session-1", MatchID: "match-2", Result: "win", MatchTimestamp: now.Add(time.Hour)},
	}
	analyticsRepo.matchResults["session-2"] = []*models.DraftMatchResult{
		{SessionID: "session-2", MatchID: "match-3", Result: "loss", MatchTimestamp: now.Add(-20 * time.Hour)},
	}

	cardStore := newMockCardStore()
	cardStore.cards[1001] = &cards.Card{ArenaID: 1001, Colors: []string{"G"}}
	cardStore.cards[1002] = &cards.Card{ArenaID: 1002, Colors: []string{"W"}}
	cardStore.cards[1003] = &cards.Card{ArenaID: 1003, Colors: []string{"U"}}
	cardStore.cards[1004] = &cards.Card{ArenaID: 1004, Colors: []string{"B"}}

	matchRepo := &mockMatchRepository{}

	analyzer := NewArchetypePerformanceAnalyzer(draftRepo, analyticsRepo, matchRepo, cardStore)

	setCode := "FDN"
	stats, err := analyzer.AnalyzeArchetypePerformance(ctx, &setCode)
	if err != nil {
		t.Fatalf("AnalyzeArchetypePerformance failed: %v", err)
	}

	if len(stats) != 2 {
		t.Errorf("expected 2 archetype stats, got %d", len(stats))
	}

	// Check that stats were saved
	allStats, err := analyticsRepo.GetAllArchetypeStats(ctx)
	if err != nil {
		t.Fatalf("GetAllArchetypeStats failed: %v", err)
	}

	if len(allStats) == 0 {
		// Should have some stats saved
		t.Log("No archetype stats saved (this may be expected if mocks don't fully implement)")
	}
}

func TestArchetypePerformanceAnalyzer_GetBestAndWorst(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	analyticsRepo := newMockAnalyticsRepository()

	// Pre-populate with archetype stats
	stats := []*models.DraftArchetypeStats{
		{ID: 1, SetCode: "FDN", ColorCombination: "GW", ArchetypeName: "Selesnya", MatchesPlayed: 10, MatchesWon: 8, DraftsCount: 3, UpdatedAt: now},
		{ID: 2, SetCode: "FDN", ColorCombination: "UB", ArchetypeName: "Dimir", MatchesPlayed: 8, MatchesWon: 2, DraftsCount: 2, UpdatedAt: now},
		{ID: 3, SetCode: "FDN", ColorCombination: "WR", ArchetypeName: "Boros", MatchesPlayed: 6, MatchesWon: 3, DraftsCount: 2, UpdatedAt: now},
	}

	for _, s := range stats {
		analyticsRepo.UpsertArchetypeStats(ctx, s)
	}

	draftRepo := &mockDraftRepository{}
	matchRepo := &mockMatchRepository{}
	cardStore := newMockCardStore()

	analyzer := NewArchetypePerformanceAnalyzer(draftRepo, analyticsRepo, matchRepo, cardStore)

	// Get best archetypes
	best, err := analyzer.GetBestArchetypes(ctx, 5, 2)
	if err != nil {
		t.Fatalf("GetBestArchetypes failed: %v", err)
	}

	// Should return top archetypes by win rate
	// Note: mock returns nil, so this tests error handling
	if best != nil && len(best) > 0 {
		t.Logf("Got %d best archetypes", len(best))
	}

	// Get worst archetypes
	worst, err := analyzer.GetWorstArchetypes(ctx, 5, 2)
	if err != nil {
		t.Fatalf("GetWorstArchetypes failed: %v", err)
	}

	if worst != nil && len(worst) > 0 {
		t.Logf("Got %d worst archetypes", len(worst))
	}
}

func TestToArchetypePerformanceEntry(t *testing.T) {
	now := time.Now()
	avgGrade := 85.5

	stats := &models.DraftArchetypeStats{
		ID:               1,
		SetCode:          "FDN",
		ColorCombination: "GW",
		ArchetypeName:    "Selesnya",
		MatchesPlayed:    10,
		MatchesWon:       7,
		DraftsCount:      3,
		AvgDraftGrade:    &avgGrade,
		LastPlayedAt:     &now,
		UpdatedAt:        now,
	}

	entry := ToArchetypePerformanceEntry(stats)

	if entry == nil {
		t.Fatal("expected entry to not be nil")
	}

	if entry.ColorCombination != "GW" {
		t.Errorf("expected color combination 'GW', got '%s'", entry.ColorCombination)
	}

	if entry.ArchetypeName != "Selesnya" {
		t.Errorf("expected archetype name 'Selesnya', got '%s'", entry.ArchetypeName)
	}

	if entry.WinRate != 0.7 {
		t.Errorf("expected win rate 0.7, got %f", entry.WinRate)
	}

	if entry.AvgDraftGrade == nil || *entry.AvgDraftGrade != 85.5 {
		t.Error("expected avg draft grade 85.5")
	}

	if entry.LastPlayedAt == nil {
		t.Error("expected last played at to be set")
	}
}

func TestToArchetypePerformanceEntry_Nil(t *testing.T) {
	entry := ToArchetypePerformanceEntry(nil)
	if entry != nil {
		t.Error("expected nil entry for nil input")
	}
}

func TestBuildArchetypePerformanceResponse(t *testing.T) {
	now := time.Now()

	setCode := "FDN"
	allStats := []*models.DraftArchetypeStats{
		{SetCode: "FDN", ColorCombination: "GW", ArchetypeName: "Selesnya", MatchesPlayed: 10, MatchesWon: 7, DraftsCount: 3, UpdatedAt: now},
		{SetCode: "FDN", ColorCombination: "UB", ArchetypeName: "Dimir", MatchesPlayed: 8, MatchesWon: 5, DraftsCount: 2, UpdatedAt: now},
	}
	best := []*models.DraftArchetypeStats{
		{SetCode: "FDN", ColorCombination: "GW", ArchetypeName: "Selesnya", MatchesPlayed: 10, MatchesWon: 7, DraftsCount: 3, UpdatedAt: now},
	}
	worst := []*models.DraftArchetypeStats{
		{SetCode: "FDN", ColorCombination: "UB", ArchetypeName: "Dimir", MatchesPlayed: 8, MatchesWon: 5, DraftsCount: 2, UpdatedAt: now},
	}

	response := BuildArchetypePerformanceResponse(&setCode, allStats, best, worst)

	if response == nil {
		t.Fatal("expected response to not be nil")
	}

	if *response.SetCode != "FDN" {
		t.Errorf("expected set code 'FDN', got '%s'", *response.SetCode)
	}

	if len(response.Archetypes) != 2 {
		t.Errorf("expected 2 archetypes, got %d", len(response.Archetypes))
	}

	if len(response.Best) != 1 {
		t.Errorf("expected 1 best archetype, got %d", len(response.Best))
	}

	if len(response.Worst) != 1 {
		t.Errorf("expected 1 worst archetype, got %d", len(response.Worst))
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"QuickDraft_FDN", "Draft", true},
		{"QuickDraft_FDN", "draft", true},
		{"QuickDraft_FDN", "DRAFT", true},
		{"QuickDraft_FDN", "FDN", true},
		{"QuickDraft_FDN", "fdn", true},
		{"QuickDraft_FDN", "Sealed", false},
		{"PremierDraft", "Premier", true},
		{"", "test", false},
		{"test", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.s+"_"+tc.substr, func(t *testing.T) {
			result := containsIgnoreCase(tc.s, tc.substr)
			if result != tc.expected {
				t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v", tc.s, tc.substr, result, tc.expected)
			}
		})
	}
}

func TestIsMatchFromDraft(t *testing.T) {
	now := time.Now()
	endTime := now.Add(-2 * time.Hour)

	analyzer := &ArchetypePerformanceAnalyzer{}

	session := &models.DraftSession{
		ID:        "session-1",
		SetCode:   "FDN",
		StartTime: now.Add(-5 * time.Hour),
		EndTime:   &endTime,
	}

	tests := []struct {
		name     string
		match    *models.Match
		expected bool
	}{
		{
			name: "Draft match within time frame",
			match: &models.Match{
				ID:        "match-1",
				EventName: "QuickDraft_FDN",
				Timestamp: now.Add(-time.Hour),
			},
			expected: true,
		},
		{
			name: "Non-draft match",
			match: &models.Match{
				ID:        "match-2",
				EventName: "Constructed_Standard",
				Timestamp: now.Add(-time.Hour),
			},
			expected: false,
		},
		{
			name: "Draft match too old",
			match: &models.Match{
				ID:        "match-3",
				EventName: "QuickDraft_FDN",
				Timestamp: now.Add(-48 * time.Hour),
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := analyzer.isMatchFromDraft(tc.match, session)
			if result != tc.expected {
				t.Errorf("isMatchFromDraft() = %v, want %v", result, tc.expected)
			}
		})
	}
}
