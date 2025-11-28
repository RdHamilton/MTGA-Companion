package ml

import (
	"context"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// mockRatingsRepo is a mock implementation of DraftRatingsRepository.
type mockRatingsRepo struct {
	cardRatings  []seventeenlands.CardRating
	colorRatings []seventeenlands.ColorRating
	snapshots    []*mockSnapshotInfo
}

type mockSnapshotInfo struct {
	ID        int
	ArenaID   int
	Expansion string
	Format    string
}

func (m *mockRatingsRepo) SaveSetRatings(ctx context.Context, setCode, draftFormat string, cardRatings []seventeenlands.CardRating, colorRatings []seventeenlands.ColorRating, dataSource string) error {
	return nil
}

func (m *mockRatingsRepo) GetCardRatings(ctx context.Context, setCode, draftFormat string) ([]seventeenlands.CardRating, time.Time, error) {
	return m.cardRatings, time.Now(), nil
}

func (m *mockRatingsRepo) GetColorRatings(ctx context.Context, setCode, draftFormat string) ([]seventeenlands.ColorRating, time.Time, error) {
	return m.colorRatings, time.Now(), nil
}

func (m *mockRatingsRepo) GetCardRatingByArenaID(ctx context.Context, setCode, draftFormat, arenaID string) (*seventeenlands.CardRating, error) {
	return nil, nil
}

func (m *mockRatingsRepo) IsSetRatingsCached(ctx context.Context, setCode, draftFormat string) (bool, error) {
	return len(m.cardRatings) > 0, nil
}

func (m *mockRatingsRepo) DeleteSetRatings(ctx context.Context, setCode, draftFormat string) error {
	return nil
}

func (m *mockRatingsRepo) GetAllSnapshots(ctx context.Context) ([]*struct {
	ID          int
	ArenaID     int
	Expansion   string
	Format      string
	Colors      string
	StartDate   string
	EndDate     string
	CachedAt    time.Time
	LastUpdated time.Time
}, error,
) {
	// Convert internal mock type to expected type
	result := make([]*struct {
		ID          int
		ArenaID     int
		Expansion   string
		Format      string
		Colors      string
		StartDate   string
		EndDate     string
		CachedAt    time.Time
		LastUpdated time.Time
	}, len(m.snapshots))
	for i, s := range m.snapshots {
		result[i] = &struct {
			ID          int
			ArenaID     int
			Expansion   string
			Format      string
			Colors      string
			StartDate   string
			EndDate     string
			CachedAt    time.Time
			LastUpdated time.Time
		}{
			ID:        s.ID,
			ArenaID:   s.ArenaID,
			Expansion: s.Expansion,
			Format:    s.Format,
		}
	}
	return result, nil
}

func (m *mockRatingsRepo) DeleteSnapshotsBatch(ctx context.Context, ids []int) error {
	return nil
}

func (m *mockRatingsRepo) GetSnapshotCountByExpansion(ctx context.Context) (map[string]int, error) {
	return nil, nil
}

func (m *mockRatingsRepo) GetOldestSnapshotDate(ctx context.Context) (time.Time, error) {
	return time.Time{}, nil
}

func (m *mockRatingsRepo) GetNewestSnapshotDate(ctx context.Context) (time.Time, error) {
	return time.Time{}, nil
}

func (m *mockRatingsRepo) GetCardWinRateTrend(ctx context.Context, arenaID int, expansion string, days int) ([]*struct {
	Date        time.Time
	GIHWR       float64
	OHWR        float64
	ALSA        float64
	ATA         float64
	SampleSize  int
	GamesPlayed int
}, error,
) {
	return nil, nil
}

func (m *mockRatingsRepo) GetExpansionCardIDs(ctx context.Context, expansion string, days int) ([]int, error) {
	return nil, nil
}

func (m *mockRatingsRepo) GetCardRatingHistory(ctx context.Context, arenaID int, expansion string) ([]*struct {
	ID          int
	ArenaID     int
	Expansion   string
	Format      string
	Colors      string
	GIHWR       float64
	OHWR        float64
	ALSA        float64
	ATA         float64
	GIH         int
	OH          int
	GamesPlayed int
	StartDate   string
	EndDate     string
	CachedAt    time.Time
	LastUpdated time.Time
}, error,
) {
	return nil, nil
}

func (m *mockRatingsRepo) GetPeriodAverages(ctx context.Context, expansion string, startDate, endDate time.Time) (map[int]*struct {
	GIHWR       float64
	OHWR        float64
	ALSA        float64
	ATA         float64
	GamesPlayed int
	SampleCount int
}, error,
) {
	return nil, nil
}

func (m *mockRatingsRepo) GetSetCodeByArenaID(ctx context.Context, arenaID string) (string, error) {
	return "", nil
}

func (m *mockRatingsRepo) GetStatisticsStaleness(ctx context.Context, staleAgeSeconds int) (*struct {
	Total     int
	Fresh     int
	Stale     int
	StaleSets []string
}, error,
) {
	return nil, nil
}

func (m *mockRatingsRepo) GetStaleSets(ctx context.Context, staleAgeSeconds int) ([]string, error) {
	return nil, nil
}

func (m *mockRatingsRepo) GetStaleStats(ctx context.Context, staleAgeSeconds int) ([]*struct {
	SetCode     string
	Format      string
	LastUpdated string
}, error,
) {
	return nil, nil
}

func TestNewTrainingPipeline(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}
	ratingsRepo := &mockRatingsRepo{}

	config := DefaultModelConfig()
	model := NewModel(feedbackRepo, perfRepo, config)

	pipeline := NewTrainingPipeline(model, nil, feedbackRepo, perfRepo, nil)

	if pipeline == nil {
		t.Fatal("expected pipeline to be created")
	}

	if pipeline.model != model {
		t.Error("expected model to be set")
	}

	_ = ratingsRepo // Use ratingsRepo to avoid unused variable
}

func TestDefaultPipelineConfig(t *testing.T) {
	config := DefaultPipelineConfig()

	if config.MinGamesThreshold <= 0 {
		t.Error("expected positive MinGamesThreshold")
	}
	if config.BatchSize <= 0 {
		t.Error("expected positive BatchSize")
	}
	if config.ParallelWorkers <= 0 {
		t.Error("expected positive ParallelWorkers")
	}
}

func TestNormalizeWinRate(t *testing.T) {
	pipeline := &TrainingPipeline{config: DefaultPipelineConfig()}

	tests := []struct {
		name     string
		winRate  float64
		expected float64
	}{
		{"low win rate", 45.0, 0.0},
		{"mid win rate", 55.0, 0.5},
		{"high win rate", 65.0, 1.0},
		{"very high", 70.0, 1.0}, // Capped at 1.0
		{"very low", 40.0, 0.0},  // Capped at 0.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pipeline.normalizeWinRate(tt.winRate)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestNormalizePickPosition(t *testing.T) {
	pipeline := &TrainingPipeline{config: DefaultPipelineConfig()}

	tests := []struct {
		name     string
		ata      float64
		expected float64
	}{
		{"first pick", 1.0, 1.0},
		{"mid pick", 7.0, 0.5384615384615384}, // (1 - (7-1)/13)
		{"late pick", 14.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pipeline.normalizePickPosition(tt.ata)
			// Allow small floating point differences
			if result < tt.expected-0.01 || result > tt.expected+0.01 {
				t.Errorf("expected ~%f, got %f", tt.expected, result)
			}
		})
	}
}

func TestIsValidCard(t *testing.T) {
	pipeline := &TrainingPipeline{config: DefaultPipelineConfig()}

	tests := []struct {
		name     string
		card     *CardTrainingData
		expected bool
	}{
		{
			name:     "nil card",
			card:     nil,
			expected: false,
		},
		{
			name: "valid card",
			card: &CardTrainingData{
				Name:  "Test Card",
				GIHWR: 55.0,
				ATA:   5.0,
			},
			expected: true,
		},
		{
			name: "no name",
			card: &CardTrainingData{
				GIHWR: 55.0,
				ATA:   5.0,
			},
			expected: false,
		},
		{
			name: "invalid win rate",
			card: &CardTrainingData{
				Name:  "Test",
				GIHWR: 150.0, // Invalid
				ATA:   5.0,
			},
			expected: false,
		},
		{
			name: "invalid ATA",
			card: &CardTrainingData{
				Name:  "Test",
				GIHWR: 55.0,
				ATA:   25.0, // Invalid
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := pipeline.isValidCard(tt.card)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestTransformData(t *testing.T) {
	config := DefaultPipelineConfig()
	config.MinGamesThreshold = 50

	pipeline := &TrainingPipeline{config: config}

	cards := []*CardTrainingData{
		{
			Name:        "Good Card",
			GIHWR:       55.0,
			ATA:         3.0,
			GamesPlayed: 1000,
			Types:       []string{"Creature", "Elf", "Warrior"},
		},
		{
			Name:        "Low Games Card",
			GIHWR:       60.0,
			ATA:         2.0,
			GamesPlayed: 10, // Below threshold
			Types:       []string{"Instant"},
		},
		{
			Name:  "", // Invalid - no name
			GIHWR: 55.0,
			ATA:   5.0,
		},
	}

	valid, skippedLow, skippedInvalid := pipeline.transformData(cards)

	if len(valid) != 1 {
		t.Errorf("expected 1 valid card, got %d", len(valid))
	}
	if skippedLow != 1 {
		t.Errorf("expected 1 skipped low games, got %d", skippedLow)
	}
	if skippedInvalid != 1 {
		t.Errorf("expected 1 skipped invalid, got %d", skippedInvalid)
	}

	if len(valid) > 0 {
		// Check that keywords were extracted
		if len(valid[0].Keywords) == 0 {
			t.Error("expected keywords to be extracted from types")
		}
	}
}

func TestExtractKeywordsFromTypes(t *testing.T) {
	pipeline := &TrainingPipeline{config: DefaultPipelineConfig()}

	tests := []struct {
		name           string
		types          []string
		expectedLength int
	}{
		{
			name:           "creature types",
			types:          []string{"Creature", "Elf", "Warrior"},
			expectedLength: 2, // Elf and Warrior are creature types
		},
		{
			name:           "no creature types",
			types:          []string{"Instant", "Sorcery"},
			expectedLength: 0,
		},
		{
			name:           "mixed types",
			types:          []string{"Creature", "Human", "Wizard", "Enchantment"},
			expectedLength: 2, // Human and Wizard
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keywords := pipeline.extractKeywordsFromTypes(tt.types)
			if len(keywords) != tt.expectedLength {
				t.Errorf("expected %d keywords, got %d", tt.expectedLength, len(keywords))
			}
		})
	}
}

func TestCalculateDataQuality(t *testing.T) {
	pipeline := &TrainingPipeline{config: DefaultPipelineConfig()}

	// Test with good data
	goodCards := make([]*CardTrainingData, 1500)
	for i := range goodCards {
		// Create variance in win rates
		goodCards[i] = &CardTrainingData{
			GIHWR: 45.0 + float64(i%20),
		}
	}

	goodFeedback := make([]*models.RecommendationFeedback, 2000)
	for i := range goodFeedback {
		goodFeedback[i] = &models.RecommendationFeedback{}
	}

	qualityGood := pipeline.calculateDataQuality(goodCards, goodFeedback)
	if qualityGood < 0.5 {
		t.Errorf("expected high quality score, got %f", qualityGood)
	}

	// Test with minimal data
	minCards := []*CardTrainingData{{GIHWR: 55.0}}
	minFeedback := []*models.RecommendationFeedback{{}}

	qualityMin := pipeline.calculateDataQuality(minCards, minFeedback)
	if qualityMin > qualityGood {
		t.Error("minimal data should have lower quality than good data")
	}
}

func TestExtractArchetypes(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	model := NewModel(feedbackRepo, perfRepo, nil)
	pipeline := NewTrainingPipeline(model, nil, feedbackRepo, perfRepo, nil)

	ctx := context.Background()

	archetype1 := "UW Flyers"
	archetype2 := "BR Sacrifice"

	perfData := []*models.DeckPerformanceHistory{
		{Archetype: &archetype1, ColorIdentity: "WU", Format: "Draft", Result: "win"},
		{Archetype: &archetype1, ColorIdentity: "WU", Format: "Draft", Result: "win"},
		{Archetype: &archetype1, ColorIdentity: "WU", Format: "Draft", Result: "loss"},
		{Archetype: &archetype2, ColorIdentity: "BR", Format: "Draft", Result: "win"},
	}

	archetypes := pipeline.extractArchetypes(ctx, perfData)

	if len(archetypes) != 2 {
		t.Errorf("expected 2 archetypes, got %d", len(archetypes))
	}

	// First should be UW Flyers (3 games) due to sorting by popularity
	if len(archetypes) > 0 && archetypes[0].Name != "UW Flyers" {
		t.Errorf("expected first archetype to be UW Flyers, got %s", archetypes[0].Name)
	}
}

func TestProgressTracking(t *testing.T) {
	pipeline := &TrainingPipeline{
		config: DefaultPipelineConfig(),
		progress: &TrainingProgress{
			Stage: "idle",
		},
	}

	ch := make(chan *TrainingProgress, 10)
	pipeline.SetProgressChannel(ch)

	pipeline.updateProgress("testing", 1, 3, 33.3)

	progress := pipeline.GetProgress()
	if progress.Stage != "testing" {
		t.Errorf("expected stage 'testing', got '%s'", progress.Stage)
	}
	if progress.CurrentStep != 1 {
		t.Errorf("expected step 1, got %d", progress.CurrentStep)
	}

	// Check channel received progress
	select {
	case p := <-ch:
		if p.Stage != "testing" {
			t.Error("channel should receive progress update")
		}
	default:
		t.Error("expected progress on channel")
	}

	pipeline.failProgress("test error")
	if !pipeline.progress.Failed {
		t.Error("expected progress to be marked as failed")
	}

	pipeline.completeProgress()
	if !pipeline.progress.Complete {
		t.Error("expected progress to be marked as complete")
	}
}

func TestConvertRatingToTrainingData(t *testing.T) {
	config := DefaultPipelineConfig()
	config.MinGamesThreshold = 50

	pipeline := &TrainingPipeline{config: config}

	// Card with sufficient games
	rating := seventeenlands.CardRating{
		MTGAID: 12345,
		Name:   "Test Card",
		Color:  "W",
		Rarity: "uncommon",
		GIHWR:  55.5,
		OHWR:   54.0,
		ATA:    3.5,
		ALSA:   4.0,
		GIH:    1000,
	}

	result := pipeline.convertRatingToTrainingData(rating, "BLB")

	if result == nil {
		t.Fatal("expected card to be converted")
	}
	if result.Name != "Test Card" {
		t.Errorf("expected name 'Test Card', got '%s'", result.Name)
	}
	if result.SetCode != "BLB" {
		t.Errorf("expected set code 'BLB', got '%s'", result.SetCode)
	}
	if result.QualityScore < 0 || result.QualityScore > 1 {
		t.Errorf("quality score %f out of range", result.QualityScore)
	}

	// Card with insufficient games
	lowGamesRating := rating
	lowGamesRating.GIH = 10

	resultLow := pipeline.convertRatingToTrainingData(lowGamesRating, "BLB")
	if resultLow != nil {
		t.Error("expected nil for card with low games")
	}
}
