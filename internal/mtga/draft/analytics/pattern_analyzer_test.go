package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// Mock implementations

type mockDraftRepository struct {
	sessions []*models.DraftSession
	picks    map[string][]*models.DraftPickSession
}

func (m *mockDraftRepository) CreateSession(ctx context.Context, session *models.DraftSession) error {
	return nil
}

func (m *mockDraftRepository) GetSession(ctx context.Context, id string) (*models.DraftSession, error) {
	for _, s := range m.sessions {
		if s.ID == id {
			return s, nil
		}
	}
	return nil, nil
}

func (m *mockDraftRepository) GetActiveSessions(ctx context.Context) ([]*models.DraftSession, error) {
	return nil, nil
}

func (m *mockDraftRepository) GetActiveSessionByIDPrefix(ctx context.Context, prefix string) (*models.DraftSession, error) {
	return nil, nil
}

func (m *mockDraftRepository) GetCompletedSessions(ctx context.Context, limit int) ([]*models.DraftSession, error) {
	if limit > len(m.sessions) {
		return m.sessions, nil
	}
	return m.sessions[:limit], nil
}

func (m *mockDraftRepository) UpdateSessionStatus(ctx context.Context, id string, status string, endTime *time.Time) error {
	return nil
}

func (m *mockDraftRepository) UpdateSessionTotalPicks(ctx context.Context, id string, totalPicks int) error {
	return nil
}

func (m *mockDraftRepository) IncrementSessionPicks(ctx context.Context, id string) error {
	return nil
}

func (m *mockDraftRepository) SavePick(ctx context.Context, pick *models.DraftPickSession) error {
	return nil
}

func (m *mockDraftRepository) GetPicksBySession(ctx context.Context, sessionID string) ([]*models.DraftPickSession, error) {
	return m.picks[sessionID], nil
}

func (m *mockDraftRepository) GetPickByNumber(ctx context.Context, sessionID string, packNum, pickNum int) (*models.DraftPickSession, error) {
	return nil, nil
}

func (m *mockDraftRepository) UpdatePickQuality(ctx context.Context, pickID int, grade string, rank int, packBestGIHWR, pickedCardGIHWR float64, alternativesJSON string) error {
	return nil
}

func (m *mockDraftRepository) SavePack(ctx context.Context, pack *models.DraftPackSession) error {
	return nil
}

func (m *mockDraftRepository) GetPacksBySession(ctx context.Context, sessionID string) ([]*models.DraftPackSession, error) {
	return nil, nil
}

func (m *mockDraftRepository) GetPack(ctx context.Context, sessionID string, packNum, pickNum int) (*models.DraftPackSession, error) {
	return nil, nil
}

func (m *mockDraftRepository) UpdateSessionGrade(ctx context.Context, sessionID string, overallGrade string, overallScore int, pickQuality, colorDiscipline, deckComposition, strategic float64) error {
	return nil
}

func (m *mockDraftRepository) UpdateSessionPrediction(ctx context.Context, sessionID string, winRate, winRateMin, winRateMax float64, factorsJSON string, predictedAt time.Time) error {
	return nil
}

func (m *mockDraftRepository) ClearAllSessions(ctx context.Context) (sessionsDeleted, picksDeleted, packsDeleted int64, err error) {
	return 0, 0, 0, nil
}

func (m *mockDraftRepository) GetSessionCount(ctx context.Context) (int, error) {
	return len(m.sessions), nil
}

func (m *mockDraftRepository) GetPickCount(ctx context.Context) (int, error) {
	total := 0
	for _, picks := range m.picks {
		total += len(picks)
	}
	return total, nil
}

func (m *mockDraftRepository) GetPackCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockDraftRepository) GetAllPickCardCounts(ctx context.Context) (map[int]int, error) {
	return nil, nil
}

type mockAnalyticsRepository struct {
	patternAnalysis      map[string]*models.DraftPatternAnalysis
	matchResults         map[string][]*models.DraftMatchResult
	archetypeStats       map[string]*models.DraftArchetypeStats
	communityComparisons map[string]*models.DraftCommunityComparison
}

func newMockAnalyticsRepository() *mockAnalyticsRepository {
	return &mockAnalyticsRepository{
		patternAnalysis:      make(map[string]*models.DraftPatternAnalysis),
		matchResults:         make(map[string][]*models.DraftMatchResult),
		archetypeStats:       make(map[string]*models.DraftArchetypeStats),
		communityComparisons: make(map[string]*models.DraftCommunityComparison),
	}
}

func (m *mockAnalyticsRepository) SaveDraftMatchResult(ctx context.Context, result *models.DraftMatchResult) error {
	m.matchResults[result.SessionID] = append(m.matchResults[result.SessionID], result)
	return nil
}

func (m *mockAnalyticsRepository) GetDraftMatchResults(ctx context.Context, sessionID string) ([]*models.DraftMatchResult, error) {
	return m.matchResults[sessionID], nil
}

func (m *mockAnalyticsRepository) GetDraftMatchResultsByTimeRange(ctx context.Context, start, end time.Time) ([]*models.DraftMatchResult, error) {
	return nil, nil
}

func (m *mockAnalyticsRepository) GetDraftMatchResultCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockAnalyticsRepository) GetArchetypeStats(ctx context.Context, setCode string) ([]*models.DraftArchetypeStats, error) {
	var result []*models.DraftArchetypeStats
	for _, s := range m.archetypeStats {
		if s.SetCode == setCode {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockAnalyticsRepository) GetAllArchetypeStats(ctx context.Context) ([]*models.DraftArchetypeStats, error) {
	var result []*models.DraftArchetypeStats
	for _, s := range m.archetypeStats {
		result = append(result, s)
	}
	return result, nil
}

func (m *mockAnalyticsRepository) UpsertArchetypeStats(ctx context.Context, stats *models.DraftArchetypeStats) error {
	key := stats.SetCode + "_" + stats.ColorCombination
	m.archetypeStats[key] = stats
	return nil
}

func (m *mockAnalyticsRepository) GetBestArchetypes(ctx context.Context, minMatches, limit int) ([]*models.DraftArchetypeStats, error) {
	return nil, nil
}

func (m *mockAnalyticsRepository) GetWorstArchetypes(ctx context.Context, minMatches, limit int) ([]*models.DraftArchetypeStats, error) {
	return nil, nil
}

func (m *mockAnalyticsRepository) SaveTemporalTrend(ctx context.Context, trend *models.DraftTemporalTrend) error {
	return nil
}

func (m *mockAnalyticsRepository) GetTemporalTrends(ctx context.Context, periodType string, limit int) ([]*models.DraftTemporalTrend, error) {
	return nil, nil
}

func (m *mockAnalyticsRepository) GetTemporalTrendsBySet(ctx context.Context, setCode, periodType string, limit int) ([]*models.DraftTemporalTrend, error) {
	return nil, nil
}

func (m *mockAnalyticsRepository) ClearTemporalTrends(ctx context.Context, periodType string) error {
	return nil
}

func (m *mockAnalyticsRepository) SavePatternAnalysis(ctx context.Context, analysis *models.DraftPatternAnalysis) error {
	key := ""
	if analysis.SetCode != nil {
		key = *analysis.SetCode
	}
	m.patternAnalysis[key] = analysis
	return nil
}

func (m *mockAnalyticsRepository) GetPatternAnalysis(ctx context.Context, setCode *string) (*models.DraftPatternAnalysis, error) {
	key := ""
	if setCode != nil {
		key = *setCode
	}
	return m.patternAnalysis[key], nil
}

func (m *mockAnalyticsRepository) SaveCommunityComparison(ctx context.Context, comparison *models.DraftCommunityComparison) error {
	key := comparison.SetCode + "_" + comparison.DraftFormat
	m.communityComparisons[key] = comparison
	return nil
}

func (m *mockAnalyticsRepository) GetCommunityComparison(ctx context.Context, setCode, draftFormat string) (*models.DraftCommunityComparison, error) {
	key := setCode + "_" + draftFormat
	return m.communityComparisons[key], nil
}

func (m *mockAnalyticsRepository) GetAllCommunityComparisons(ctx context.Context) ([]*models.DraftCommunityComparison, error) {
	var result []*models.DraftCommunityComparison
	for _, c := range m.communityComparisons {
		result = append(result, c)
	}
	return result, nil
}

type mockCardStore struct {
	cards map[int]*cards.Card
}

func newMockCardStore() *mockCardStore {
	return &mockCardStore{
		cards: make(map[int]*cards.Card),
	}
}

func (m *mockCardStore) GetCard(arenaID int) (*cards.Card, error) {
	return m.cards[arenaID], nil
}

func (m *mockCardStore) GetCardByName(name string) (*cards.Card, error) {
	for _, c := range m.cards {
		if c.Name == name {
			return c, nil
		}
	}
	return nil, nil
}

// Tests

func TestPatternAnalyzer_CalculateColorPreferences(t *testing.T) {
	analyzer := &PatternAnalyzer{}

	picks := []*pickWithCard{
		{pick: &models.DraftPickSession{PickNumber: 1}, card: &cards.Card{Colors: []string{"G"}}},
		{pick: &models.DraftPickSession{PickNumber: 2}, card: &cards.Card{Colors: []string{"G"}}},
		{pick: &models.DraftPickSession{PickNumber: 3}, card: &cards.Card{Colors: []string{"W"}}},
		{pick: &models.DraftPickSession{PickNumber: 4}, card: &cards.Card{Colors: []string{"G", "W"}}},
		{pick: &models.DraftPickSession{PickNumber: 5}, card: &cards.Card{Colors: []string{}}}, // Colorless
	}

	prefs := analyzer.calculateColorPreferences(picks)

	// Should have 3 colors: G, W, C
	if len(prefs) != 3 {
		t.Errorf("expected 3 colors, got %d", len(prefs))
	}

	// Green should be first (3 picks)
	if prefs[0].Color != "G" {
		t.Errorf("expected green first, got %s", prefs[0].Color)
	}
	if prefs[0].TotalPicks != 3 {
		t.Errorf("expected 3 green picks, got %d", prefs[0].TotalPicks)
	}
}

func TestPatternAnalyzer_CalculateTypePreferences(t *testing.T) {
	analyzer := &PatternAnalyzer{}

	picks := []*pickWithCard{
		{pick: &models.DraftPickSession{PickNumber: 1}, card: &cards.Card{TypeLine: "Creature — Human Wizard"}},
		{pick: &models.DraftPickSession{PickNumber: 2}, card: &cards.Card{TypeLine: "Creature — Elf"}},
		{pick: &models.DraftPickSession{PickNumber: 3}, card: &cards.Card{TypeLine: "Instant"}},
		{pick: &models.DraftPickSession{PickNumber: 4}, card: &cards.Card{TypeLine: "Sorcery"}},
		{pick: &models.DraftPickSession{PickNumber: 5}, card: &cards.Card{TypeLine: "Artifact Creature — Golem"}},
	}

	prefs := analyzer.calculateTypePreferences(picks)

	// Creature should be first (3 picks)
	found := false
	for _, p := range prefs {
		if p.Type == "Creature" && p.TotalPicks == 3 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 3 creature picks")
	}
}

func TestExtractPrimaryType(t *testing.T) {
	tests := []struct {
		typeLine string
		expected string
	}{
		{"Creature — Human Wizard", "Creature"},
		{"Instant", "Instant"},
		{"Sorcery", "Sorcery"},
		{"Artifact Creature — Golem", "Creature"},
		{"Legendary Creature — Dragon", "Creature"},
		{"Enchantment — Aura", "Enchantment"},
		{"Artifact", "Artifact"},
		{"Planeswalker — Jace", "Planeswalker"},
		{"Basic Land — Forest", "Land"},
		{"Legendary Artifact", "Artifact"},
	}

	for _, tc := range tests {
		t.Run(tc.typeLine, func(t *testing.T) {
			result := extractPrimaryType(tc.typeLine)
			if result != tc.expected {
				t.Errorf("extractPrimaryType(%q) = %q, want %q", tc.typeLine, result, tc.expected)
			}
		})
	}
}

func TestPatternAnalyzer_CalculatePickOrderPatterns(t *testing.T) {
	analyzer := &PatternAnalyzer{}

	gihwr := 0.55
	picks := []*pickWithCard{
		{pick: &models.DraftPickSession{PickNumber: 1, PickedCardGIHWR: &gihwr}, card: &cards.Card{Rarity: "rare"}},
		{pick: &models.DraftPickSession{PickNumber: 2, PickedCardGIHWR: &gihwr}, card: &cards.Card{Rarity: "common"}},
		{pick: &models.DraftPickSession{PickNumber: 3, PickedCardGIHWR: &gihwr}, card: &cards.Card{Rarity: "common"}},
		{pick: &models.DraftPickSession{PickNumber: 6, PickedCardGIHWR: &gihwr}, card: &cards.Card{Rarity: "uncommon"}},
		{pick: &models.DraftPickSession{PickNumber: 7, PickedCardGIHWR: &gihwr}, card: &cards.Card{Rarity: "common"}},
		{pick: &models.DraftPickSession{PickNumber: 12, PickedCardGIHWR: &gihwr}, card: &cards.Card{Rarity: "common"}},
	}

	patterns := analyzer.calculatePickOrderPatterns(picks)

	if len(patterns) != 3 {
		t.Fatalf("expected 3 phases, got %d", len(patterns))
	}

	// Check early phase
	earlyFound := false
	for _, p := range patterns {
		if p.Phase == "early" {
			earlyFound = true
			if p.TotalPicks != 3 {
				t.Errorf("expected 3 early picks, got %d", p.TotalPicks)
			}
			if p.RarePicks != 1 {
				t.Errorf("expected 1 rare pick in early phase, got %d", p.RarePicks)
			}
		}
	}
	if !earlyFound {
		t.Error("early phase not found")
	}
}

func TestDeterminePrimaryColorPair(t *testing.T) {
	tests := []struct {
		name        string
		colorCounts map[string]int
		expected    string
	}{
		{"WG (sorted in WUBRG order)", map[string]int{"G": 10, "W": 8, "U": 2}, "WG"},
		{"UB", map[string]int{"U": 15, "B": 12, "R": 1}, "UB"},
		{"Mono Green", map[string]int{"G": 20}, "G"},
		{"Empty", map[string]int{}, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := determinePrimaryColorPair(tc.colorCounts)
			if result != tc.expected {
				t.Errorf("determinePrimaryColorPair() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestGetArchetypeName(t *testing.T) {
	tests := []struct {
		colorPair string
		expected  string
	}{
		{"WU", "Azorius"},
		{"UB", "Dimir"},
		{"BR", "Rakdos"},
		{"RG", "Gruul"},
		{"WG", "Selesnya"}, // WUBRG order
		{"GW", "Selesnya"}, // Also handles reverse
		{"WB", "Orzhov"},
		{"UR", "Izzet"},
		{"BG", "Golgari"},
		{"WR", "Boros"},
		{"UG", "Simic"},
		{"W", "Mono-White"},
		{"U", "Mono-Blue"},
		{"B", "Mono-Black"},
		{"R", "Mono-Red"},
		{"G", "Mono-Green"},
		{"XYZ", "XYZ"}, // Unknown
	}

	for _, tc := range tests {
		t.Run(tc.colorPair, func(t *testing.T) {
			result := getArchetypeName(tc.colorPair)
			if result != tc.expected {
				t.Errorf("getArchetypeName(%q) = %q, want %q", tc.colorPair, result, tc.expected)
			}
		})
	}
}

func TestPatternAnalyzer_AnalyzePatterns(t *testing.T) {
	ctx := context.Background()
	now := time.Now()

	// Set up mock data
	draftRepo := &mockDraftRepository{
		sessions: []*models.DraftSession{
			{ID: "session-1", SetCode: "FDN", Status: "completed", StartTime: now},
		},
		picks: map[string][]*models.DraftPickSession{
			"session-1": {
				{ID: 1, SessionID: "session-1", CardID: "1001", PickNumber: 1},
				{ID: 2, SessionID: "session-1", CardID: "1002", PickNumber: 2},
				{ID: 3, SessionID: "session-1", CardID: "1003", PickNumber: 3},
			},
		},
	}

	analyticsRepo := newMockAnalyticsRepository()

	cardStore := newMockCardStore()
	cardStore.cards[1001] = &cards.Card{ArenaID: 1001, Colors: []string{"G"}, TypeLine: "Creature — Elf", Rarity: "common"}
	cardStore.cards[1002] = &cards.Card{ArenaID: 1002, Colors: []string{"G"}, TypeLine: "Instant", Rarity: "uncommon"}
	cardStore.cards[1003] = &cards.Card{ArenaID: 1003, Colors: []string{"G", "W"}, TypeLine: "Creature — Human", Rarity: "rare"}

	analyzer := NewPatternAnalyzer(draftRepo, analyticsRepo, cardStore)

	setCode := "FDN"
	analysis, err := analyzer.AnalyzePatterns(ctx, &setCode)
	if err != nil {
		t.Fatalf("AnalyzePatterns failed: %v", err)
	}

	if analysis == nil {
		t.Fatal("expected analysis to not be nil")
	}

	if analysis.SampleSize != 1 {
		t.Errorf("expected sample size 1, got %d", analysis.SampleSize)
	}

	// Check that analysis was saved
	saved, err := analyticsRepo.GetPatternAnalysis(ctx, &setCode)
	if err != nil {
		t.Fatalf("GetPatternAnalysis failed: %v", err)
	}
	if saved == nil {
		t.Error("expected analysis to be saved in repository")
	}
}

func TestPatternAnalyzer_EmptyData(t *testing.T) {
	ctx := context.Background()

	draftRepo := &mockDraftRepository{
		sessions: []*models.DraftSession{},
		picks:    make(map[string][]*models.DraftPickSession),
	}

	analyticsRepo := newMockAnalyticsRepository()
	cardStore := newMockCardStore()

	analyzer := NewPatternAnalyzer(draftRepo, analyticsRepo, cardStore)

	analysis, err := analyzer.AnalyzePatterns(ctx, nil)
	if err != nil {
		t.Fatalf("AnalyzePatterns failed: %v", err)
	}

	if analysis.SampleSize != 0 {
		t.Errorf("expected sample size 0 for empty data, got %d", analysis.SampleSize)
	}
}

func TestToPatternAnalysisResponse(t *testing.T) {
	analysis := &models.DraftPatternAnalysis{
		SetCode:               nil,
		ColorPreferenceJSON:   `[{"color":"G","totalPicks":10,"percentOfPool":40.0,"avgPickOrder":3.5}]`,
		TypePreferenceJSON:    `[{"type":"Creature","totalPicks":15,"percentOfPool":60.0,"avgPickOrder":2.5}]`,
		PickOrderPatternJSON:  `[{"phase":"early","avgRating":0.55,"totalPicks":5,"rarePicks":2,"commonPicks":2}]`,
		ArchetypeAffinityJSON: `[{"colorPair":"GW","archetypeName":"Selesnya","timesBuilt":3,"avgWinRate":0.65,"affinityScore":0.3}]`,
		SampleSize:            10,
		CalculatedAt:          time.Now(),
	}

	response, err := ToPatternAnalysisResponse(analysis)
	if err != nil {
		t.Fatalf("ToPatternAnalysisResponse failed: %v", err)
	}

	if len(response.ColorPreferences) != 1 {
		t.Errorf("expected 1 color preference, got %d", len(response.ColorPreferences))
	}

	if response.ColorPreferences[0].Color != "G" {
		t.Errorf("expected color 'G', got '%s'", response.ColorPreferences[0].Color)
	}

	if len(response.TypePreferences) != 1 {
		t.Errorf("expected 1 type preference, got %d", len(response.TypePreferences))
	}

	if len(response.PickOrderPatterns) != 1 {
		t.Errorf("expected 1 pick order pattern, got %d", len(response.PickOrderPatterns))
	}

	if len(response.ArchetypeAffinities) != 1 {
		t.Errorf("expected 1 archetype affinity, got %d", len(response.ArchetypeAffinities))
	}
}

func TestToPatternAnalysisResponse_NilInput(t *testing.T) {
	response, err := ToPatternAnalysisResponse(nil)
	if err != nil {
		t.Fatalf("ToPatternAnalysisResponse failed: %v", err)
	}
	if response != nil {
		t.Error("expected nil response for nil input")
	}
}
