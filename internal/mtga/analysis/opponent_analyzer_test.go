package analysis

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/ramonehamilton/MTGA-Companion/internal/archetype"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// Mock repositories for testing
type mockGamePlayRepo struct {
	mock.Mock
}

func (m *mockGamePlayRepo) GetOpponentCardsByMatch(ctx context.Context, matchID string) ([]*models.OpponentCardObserved, error) {
	args := m.Called(ctx, matchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.OpponentCardObserved), args.Error(1)
}

type mockOpponentRepo struct {
	mock.Mock
}

func (m *mockOpponentRepo) CreateOrUpdateProfile(ctx context.Context, profile *models.OpponentDeckProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *mockOpponentRepo) GetProfileByMatchID(ctx context.Context, matchID string) (*models.OpponentDeckProfile, error) {
	args := m.Called(ctx, matchID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OpponentDeckProfile), args.Error(1)
}

func (m *mockOpponentRepo) ListProfiles(ctx context.Context, filter *repository.OpponentProfileFilter) ([]*models.OpponentDeckProfile, error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.OpponentDeckProfile), args.Error(1)
}

func (m *mockOpponentRepo) GetExpectedCards(ctx context.Context, archetype, format string) ([]*models.ArchetypeExpectedCard, error) {
	args := m.Called(ctx, archetype, format)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.ArchetypeExpectedCard), args.Error(1)
}

func (m *mockOpponentRepo) RecordMatchup(ctx context.Context, stat *models.MatchupStatistic) error {
	args := m.Called(ctx, stat)
	return args.Error(0)
}

func (m *mockOpponentRepo) ListMatchupStats(ctx context.Context, accountID int, format *string) ([]*models.MatchupStatistic, error) {
	args := m.Called(ctx, accountID, format)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.MatchupStatistic), args.Error(1)
}

func (m *mockOpponentRepo) GetMatchupStats(ctx context.Context, accountID int, playerArchetype, opponentArchetype, format string) (*models.MatchupStatistic, error) {
	args := m.Called(ctx, accountID, playerArchetype, opponentArchetype, format)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.MatchupStatistic), args.Error(1)
}

func (m *mockOpponentRepo) GetOpponentHistorySummary(ctx context.Context, accountID int, format *string) (*models.OpponentHistorySummary, error) {
	args := m.Called(ctx, accountID, format)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.OpponentHistorySummary), args.Error(1)
}

func (m *mockOpponentRepo) DeleteProfile(ctx context.Context, matchID string) error {
	args := m.Called(ctx, matchID)
	return args.Error(0)
}

func (m *mockOpponentRepo) GetTopMatchups(ctx context.Context, accountID int, format string, limit int) ([]*models.MatchupStatistic, error) {
	args := m.Called(ctx, accountID, format, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.MatchupStatistic), args.Error(1)
}

func (m *mockOpponentRepo) UpsertExpectedCard(ctx context.Context, card *models.ArchetypeExpectedCard) error {
	args := m.Called(ctx, card)
	return args.Error(0)
}

func (m *mockOpponentRepo) DeleteExpectedCards(ctx context.Context, archetypeName, format string) error {
	args := m.Called(ctx, archetypeName, format)
	return args.Error(0)
}

type mockMatchRepo struct {
	mock.Mock
}

func (m *mockMatchRepo) GetByID(ctx context.Context, id string) (*models.Match, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Match), args.Error(1)
}

func TestDetectStyle(t *testing.T) {
	analyzer := &OpponentAnalyzer{}

	tests := []struct {
		name     string
		analysis *archetype.DeckAnalysis
		expected string
	}{
		{
			name: "Aggro - low curve with many creatures",
			analysis: &archetype.DeckAnalysis{
				AvgCMC:            2.0,
				CreatureCount:     20,
				InstantCount:      4,
				SorceryCount:      4,
				ArtifactCount:     0,
				EnchantmentCount:  0,
				PlaneswalkerCount: 0,
			},
			expected: models.DeckStyleAggro,
		},
		{
			name: "Control - high curve with many spells",
			analysis: &archetype.DeckAnalysis{
				AvgCMC:            4.0,
				CreatureCount:     4,
				InstantCount:      12,
				SorceryCount:      8,
				ArtifactCount:     0,
				EnchantmentCount:  2,
				PlaneswalkerCount: 4,
			},
			expected: models.DeckStyleControl,
		},
		{
			name: "Midrange - balanced creatures and curve",
			analysis: &archetype.DeckAnalysis{
				AvgCMC:            3.0,
				CreatureCount:     16,
				InstantCount:      8,
				SorceryCount:      4,
				ArtifactCount:     0,
				EnchantmentCount:  2,
				PlaneswalkerCount: 2,
			},
			expected: models.DeckStyleMidrange,
		},
		{
			name: "Tempo - mixed creatures and spells with low curve",
			analysis: &archetype.DeckAnalysis{
				AvgCMC:            2.5,
				CreatureCount:     10,
				InstantCount:      12,
				SorceryCount:      4,
				ArtifactCount:     0,
				EnchantmentCount:  2,
				PlaneswalkerCount: 0,
			},
			expected: models.DeckStyleTempo,
		},
		{
			name: "Empty deck - returns empty string",
			analysis: &archetype.DeckAnalysis{
				AvgCMC:            0,
				CreatureCount:     0,
				InstantCount:      0,
				SorceryCount:      0,
				ArtifactCount:     0,
				EnchantmentCount:  0,
				PlaneswalkerCount: 0,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.detectStyle(tt.analysis)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateInsights(t *testing.T) {
	analyzer := &OpponentAnalyzer{}

	archType := "Mono Red Aggro"
	classification := &archetype.ClassificationResult{
		PrimaryArchetype: archType,
		Confidence:       0.85,
		Analysis: &archetype.DeckAnalysis{
			AvgCMC:            2.0,
			CreatureCount:     20,
			InstantCount:      4,
			SorceryCount:      4,
			ArtifactCount:     0,
			EnchantmentCount:  0,
			PlaneswalkerCount: 0,
		},
	}

	removalCategory := models.CardCategoryRemoval
	observed := []models.ObservedCard{
		{CardID: 1, CardName: "Lightning Bolt", Category: &removalCategory},
		{CardID: 2, CardName: "Monastery Swiftspear", Category: nil},
	}

	expected := []models.ExpectedCard{
		{CardID: 3, CardName: "Goblin Guide", InclusionRate: 0.95, WasSeen: false},
		{CardID: 1, CardName: "Lightning Bolt", InclusionRate: 0.99, WasSeen: true},
	}

	insights := analyzer.generateInsights(classification, observed, expected)

	// Should have archetype insight
	hasArchetypeInsight := false
	for _, i := range insights {
		if i.Type == "archetype" && i.Priority == models.InsightPriorityHigh {
			hasArchetypeInsight = true
			break
		}
	}
	assert.True(t, hasArchetypeInsight, "should have archetype insight")

	// Should have strategy insight (aggro deck)
	hasStrategyInsight := false
	for _, i := range insights {
		if i.Type == "strategy" {
			hasStrategyInsight = true
			break
		}
	}
	assert.True(t, hasStrategyInsight, "should have strategy insight for aggro deck")

	// Should have removal insight
	hasRemovalInsight := false
	for _, i := range insights {
		if i.Type == models.CardCategoryRemoval {
			hasRemovalInsight = true
			break
		}
	}
	assert.True(t, hasRemovalInsight, "should have removal insight")

	// Should have expected cards insight (1 unseen high-inclusion card)
	hasExpectedInsight := false
	for _, i := range insights {
		if i.Type == "expected" {
			hasExpectedInsight = true
			break
		}
	}
	assert.True(t, hasExpectedInsight, "should have expected cards insight")
}

func TestGetPlayAroundAdvice(t *testing.T) {
	analyzer := &OpponentAnalyzer{}

	tests := []struct {
		name     string
		card     *models.ArchetypeExpectedCard
		expected string
	}{
		{
			name: "removal card",
			card: &models.ArchetypeExpectedCard{
				CardName: "Lightning Bolt",
				Category: ptrString(models.CardCategoryRemoval),
			},
			expected: "Hold back threats; Lightning Bolt may remove key creatures",
		},
		{
			name: "interaction card",
			card: &models.ArchetypeExpectedCard{
				CardName: "Counterspell",
				Category: ptrString(models.CardCategoryInteraction),
			},
			expected: "Leave mana open for protection against Counterspell",
		},
		{
			name: "wincon card",
			card: &models.ArchetypeExpectedCard{
				CardName: "Teferi, Hero of Dominaria",
				Category: ptrString(models.CardCategoryWincon),
			},
			expected: "Prepare answers for Teferi, Hero of Dominaria - key win condition",
		},
		{
			name: "no category",
			card: &models.ArchetypeExpectedCard{
				CardName: "Some Card",
				Category: nil,
			},
			expected: "",
		},
		{
			name: "utility card returns empty",
			card: &models.ArchetypeExpectedCard{
				CardName: "Some Utility",
				Category: ptrString(models.CardCategoryUtility),
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := analyzer.getPlayAroundAdvice(tt.card)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestUpdateMatchupStats(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOpponentRepo)

	analyzer := &OpponentAnalyzer{
		opponentRepo: mockRepo,
	}

	// Test win case
	mockRepo.On("RecordMatchup", ctx, mock.MatchedBy(func(s *models.MatchupStatistic) bool {
		return s.Wins == 1 && s.Losses == 0 && s.TotalMatches == 1
	})).Return(nil).Once()

	err := analyzer.UpdateMatchupStats(ctx, 1, "UW Control", "Mono Red", "Standard", "win", nil)
	assert.NoError(t, err)

	// Test loss case
	mockRepo.On("RecordMatchup", ctx, mock.MatchedBy(func(s *models.MatchupStatistic) bool {
		return s.Wins == 0 && s.Losses == 1 && s.TotalMatches == 1
	})).Return(nil).Once()

	err = analyzer.UpdateMatchupStats(ctx, 1, "UW Control", "Mono Red", "Standard", "loss", nil)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestGetMatchupSummary(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOpponentRepo)

	analyzer := &OpponentAnalyzer{
		opponentRepo: mockRepo,
	}

	expectedStats := []*models.MatchupStatistic{
		{
			ID:                1,
			AccountID:         1,
			PlayerArchetype:   "UW Control",
			OpponentArchetype: "Mono Red",
			Format:            "Standard",
			TotalMatches:      5,
			Wins:              3,
			Losses:            2,
		},
	}

	mockRepo.On("ListMatchupStats", ctx, 1, (*string)(nil)).Return(expectedStats, nil)

	stats, err := analyzer.GetMatchupSummary(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Len(t, stats, 1)
	assert.Equal(t, "UW Control", stats[0].PlayerArchetype)

	mockRepo.AssertExpectations(t)
}

func TestGetOpponentHistory(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(mockOpponentRepo)

	analyzer := &OpponentAnalyzer{
		opponentRepo: mockRepo,
	}

	expectedSummary := &models.OpponentHistorySummary{
		TotalOpponents:      20,
		UniqueArchetypes:    8,
		MostCommonArchetype: "Mono Red Aggro",
		MostCommonCount:     5,
		ArchetypeBreakdown: []models.ArchetypeBreakdownEntry{
			{Archetype: "Mono Red Aggro", Count: 5, Percentage: 25.0, WinRate: 0.6},
		},
		ColorIdentityStats: []models.ColorIdentityStatsEntry{
			{ColorIdentity: "R", Count: 5, Percentage: 25.0, WinRate: 0.6},
		},
	}

	mockRepo.On("GetOpponentHistorySummary", ctx, 1, (*string)(nil)).Return(expectedSummary, nil)

	summary, err := analyzer.GetOpponentHistory(ctx, 1, nil)
	assert.NoError(t, err)
	assert.Equal(t, 20, summary.TotalOpponents)
	assert.Equal(t, "Mono Red Aggro", summary.MostCommonArchetype)

	mockRepo.AssertExpectations(t)
}

func TestBuildProfile(t *testing.T) {
	analyzer := &OpponentAnalyzer{}

	arch := "Mono Red Aggro"
	style := models.DeckStyleAggro
	classification := &archetype.ClassificationResult{
		PrimaryArchetype: arch,
		Confidence:       0.85,
		ColorIdentity:    "R",
		SignatureCards:   []int{12345},
		Analysis: &archetype.DeckAnalysis{
			AvgCMC:            2.0,
			CreatureCount:     20,
			InstantCount:      4,
			SorceryCount:      4,
			ArtifactCount:     0,
			EnchantmentCount:  0,
			PlaneswalkerCount: 0,
		},
	}

	format := "Standard"
	cardIDs := []int{12345, 12346}

	profile := analyzer.buildProfile("match-123", classification, 12, cardIDs, &format)

	assert.Equal(t, "match-123", profile.MatchID)
	assert.Equal(t, &arch, profile.DetectedArchetype)
	assert.Equal(t, 0.85, profile.ArchetypeConfidence)
	assert.Equal(t, "R", profile.ColorIdentity)
	assert.Equal(t, &style, profile.DeckStyle)
	assert.Equal(t, 12, profile.CardsObserved)
	assert.Equal(t, 60, profile.EstimatedDeckSize)
	assert.Equal(t, &format, profile.Format)
	assert.NotNil(t, profile.ObservedCardIDs)
	assert.NotNil(t, profile.SignatureCards)
}

// Helper function
func ptrString(s string) *string {
	return &s
}

// Ensure test file passes when run
func TestDummy(t *testing.T) {
	_ = time.Now()
	_ = cards.Service{}
	assert.True(t, true)
}
