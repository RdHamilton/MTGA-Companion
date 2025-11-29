package ml

import (
	"context"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// mockFeedbackRepo is a mock implementation of RecommendationFeedbackRepository.
type mockFeedbackRepo struct {
	feedbackData []*models.RecommendationFeedback
}

func (m *mockFeedbackRepo) Create(ctx context.Context, feedback *models.RecommendationFeedback) error {
	m.feedbackData = append(m.feedbackData, feedback)
	return nil
}

func (m *mockFeedbackRepo) GetByID(ctx context.Context, id int) (*models.RecommendationFeedback, error) {
	for _, fb := range m.feedbackData {
		if fb.ID == id {
			return fb, nil
		}
	}
	return nil, nil
}

func (m *mockFeedbackRepo) GetByRecommendationID(ctx context.Context, recommendationID string) (*models.RecommendationFeedback, error) {
	for _, fb := range m.feedbackData {
		if fb.RecommendationID == recommendationID {
			return fb, nil
		}
	}
	return nil, nil
}

func (m *mockFeedbackRepo) GetByAccount(ctx context.Context, accountID int, limit int) ([]*models.RecommendationFeedback, error) {
	return m.feedbackData, nil
}

func (m *mockFeedbackRepo) GetByType(ctx context.Context, accountID int, recType string, limit int) ([]*models.RecommendationFeedback, error) {
	return m.feedbackData, nil
}

func (m *mockFeedbackRepo) UpdateAction(ctx context.Context, id int, action string, alternateChoiceID *int) error {
	return nil
}

func (m *mockFeedbackRepo) UpdateOutcome(ctx context.Context, id int, matchID string, result string) error {
	return nil
}

func (m *mockFeedbackRepo) GetStats(ctx context.Context, accountID int, recType *string) (*models.RecommendationStats, error) {
	return &models.RecommendationStats{}, nil
}

func (m *mockFeedbackRepo) GetStatsByDateRange(ctx context.Context, accountID int, start, end time.Time) (*models.RecommendationStats, error) {
	return &models.RecommendationStats{}, nil
}

func (m *mockFeedbackRepo) GetPendingFeedback(ctx context.Context, accountID int) ([]*models.RecommendationFeedback, error) {
	return nil, nil
}

func (m *mockFeedbackRepo) GetForMLTraining(ctx context.Context, limit int) ([]*models.RecommendationFeedback, error) {
	return m.feedbackData, nil
}

// mockPerformanceRepo is a mock implementation of DeckPerformanceRepository.
type mockPerformanceRepo struct {
	history    []*models.DeckPerformanceHistory
	archetypes []*models.DeckArchetype
	weights    map[int][]*models.ArchetypeCardWeight
}

func (m *mockPerformanceRepo) CreateHistory(ctx context.Context, history *models.DeckPerformanceHistory) error {
	m.history = append(m.history, history)
	return nil
}

func (m *mockPerformanceRepo) GetHistoryByDeck(ctx context.Context, deckID string) ([]*models.DeckPerformanceHistory, error) {
	return m.history, nil
}

func (m *mockPerformanceRepo) GetHistoryByArchetype(ctx context.Context, archetype string, format string) ([]*models.DeckPerformanceHistory, error) {
	return m.history, nil
}

func (m *mockPerformanceRepo) GetHistoryByAccount(ctx context.Context, accountID int, limit int) ([]*models.DeckPerformanceHistory, error) {
	return m.history, nil
}

func (m *mockPerformanceRepo) GetPerformanceByDateRange(ctx context.Context, accountID int, start, end time.Time) ([]*models.DeckPerformanceHistory, error) {
	return m.history, nil
}

func (m *mockPerformanceRepo) GetArchetypePerformance(ctx context.Context, archetype string, format string) (*models.ArchetypePerformanceStats, error) {
	return &models.ArchetypePerformanceStats{}, nil
}

func (m *mockPerformanceRepo) CreateArchetype(ctx context.Context, archetype *models.DeckArchetype) error {
	m.archetypes = append(m.archetypes, archetype)
	return nil
}

func (m *mockPerformanceRepo) GetArchetypeByID(ctx context.Context, id int) (*models.DeckArchetype, error) {
	for _, a := range m.archetypes {
		if a.ID == id {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockPerformanceRepo) GetArchetypeByName(ctx context.Context, name string, setCode *string, format string) (*models.DeckArchetype, error) {
	for _, a := range m.archetypes {
		if a.Name == name {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockPerformanceRepo) ListArchetypes(ctx context.Context, setCode *string, format *string) ([]*models.DeckArchetype, error) {
	return m.archetypes, nil
}

func (m *mockPerformanceRepo) UpdateArchetypeStats(ctx context.Context, archetypeID int, totalMatches, totalWins int) error {
	return nil
}

func (m *mockPerformanceRepo) CreateCardWeight(ctx context.Context, weight *models.ArchetypeCardWeight) error {
	if m.weights == nil {
		m.weights = make(map[int][]*models.ArchetypeCardWeight)
	}
	m.weights[weight.ArchetypeID] = append(m.weights[weight.ArchetypeID], weight)
	return nil
}

func (m *mockPerformanceRepo) GetCardWeights(ctx context.Context, archetypeID int) ([]*models.ArchetypeCardWeight, error) {
	if m.weights == nil {
		return nil, nil
	}
	return m.weights[archetypeID], nil
}

func (m *mockPerformanceRepo) GetCardWeightsByCard(ctx context.Context, cardID int) ([]*models.ArchetypeCardWeight, error) {
	return nil, nil
}

func (m *mockPerformanceRepo) UpsertCardWeight(ctx context.Context, weight *models.ArchetypeCardWeight) error {
	return m.CreateCardWeight(ctx, weight)
}

func (m *mockPerformanceRepo) DeleteCardWeight(ctx context.Context, archetypeID, cardID int) error {
	return nil
}

func TestNewModel(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	model := NewModel(feedbackRepo, perfRepo, nil)

	if model == nil {
		t.Fatal("expected model to be created")
	}

	info := model.GetModelInfo()
	if info.Version == "" {
		t.Error("expected version to be set")
	}
	if info.IsReady {
		t.Error("expected model to not be ready without training")
	}
}

func TestDefaultModelConfig(t *testing.T) {
	config := DefaultModelConfig()

	if config.MinTrainingSamples <= 0 {
		t.Error("expected positive MinTrainingSamples")
	}
	if config.CollaborativeWeight < 0 || config.CollaborativeWeight > 1 {
		t.Error("expected CollaborativeWeight between 0 and 1")
	}
	if config.PersonalWeight < 0 || config.PersonalWeight > 1 {
		t.Error("expected PersonalWeight between 0 and 1")
	}
	if config.MetaWeight < 0 || config.MetaWeight > 1 {
		t.Error("expected MetaWeight between 0 and 1")
	}
	if config.LearningRate <= 0 {
		t.Error("expected positive LearningRate")
	}
}

func TestScoreCards(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	// Configure with low minimum samples for testing
	config := DefaultModelConfig()
	config.MinTrainingSamples = 1

	model := NewModel(feedbackRepo, perfRepo, config)

	// Register some card features
	model.RegisterCardFeatures(1, &CardFeatures{
		CardID:     1,
		Name:       "Test Creature",
		CMC:        3,
		Colors:     []string{"W", "U"},
		Types:      []string{"Creature"},
		IsCreature: true,
	})
	model.RegisterCardFeatures(2, &CardFeatures{
		CardID:    2,
		Name:      "Test Spell",
		CMC:       2,
		Colors:    []string{"U"},
		Types:     []string{"Instant"},
		IsInstant: true,
	})

	deck := &DeckContext{
		DeckID:           "test-deck",
		Cards:            []int{},
		ColorIdentity:    []string{"W", "U"},
		CMCDistribution:  map[int]int{2: 3, 3: 2},
		TypeDistribution: map[string]int{"Creature": 10, "Instant": 5},
		Keywords:         map[string]int{"Flying": 3},
		CreatureTypes:    map[string]int{"Human": 4},
		Archetype:        "UW Flyers",
	}

	ctx := context.Background()
	scores, err := model.ScoreCards(ctx, []int{1, 2}, deck, 1)
	if err != nil {
		t.Fatalf("ScoreCards failed: %v", err)
	}

	if len(scores) != 2 {
		t.Errorf("expected 2 scores, got %d", len(scores))
	}

	for _, score := range scores {
		if score.Score < 0 || score.Score > 1 {
			t.Errorf("score %f out of range [0, 1]", score.Score)
		}
		if score.Confidence < 0 || score.Confidence > 1 {
			t.Errorf("confidence %f out of range [0, 1]", score.Confidence)
		}
	}
}

func TestColorFitScoring(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}
	model := NewModel(feedbackRepo, perfRepo, nil)

	tests := []struct {
		name       string
		cardColors []string
		deckColors []string
		expected   float64
	}{
		{
			name:       "colorless card always fits",
			cardColors: []string{},
			deckColors: []string{"W", "U"},
			expected:   1.0,
		},
		{
			name:       "perfect match",
			cardColors: []string{"W"},
			deckColors: []string{"W", "U"},
			expected:   1.0,
		},
		{
			name:       "no match",
			cardColors: []string{"R"},
			deckColors: []string{"W", "U"},
			expected:   0.0,
		},
		{
			name:       "partial match",
			cardColors: []string{"W", "R"},
			deckColors: []string{"W", "U"},
			expected:   0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := model.scoreColorFit(tt.cardColors, tt.deckColors)
			if score != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, score)
			}
		})
	}
}

func TestCMCFitScoring(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}
	model := NewModel(feedbackRepo, perfRepo, nil)

	tests := []struct {
		name     string
		cardCMC  float64
		cmcDist  map[int]int
		wantHigh bool
	}{
		{
			name:     "empty slot gets high score",
			cardCMC:  2,
			cmcDist:  map[int]int{1: 2, 3: 3},
			wantHigh: true,
		},
		{
			name:     "full slot gets lower score",
			cardCMC:  2,
			cmcDist:  map[int]int{2: 10},
			wantHigh: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := model.scoreCMCFit(tt.cardCMC, tt.cmcDist)
			if tt.wantHigh && score < 0.6 {
				t.Errorf("expected high score, got %f", score)
			}
			if !tt.wantHigh && score > 0.6 {
				t.Errorf("expected lower score, got %f", score)
			}
		})
	}
}

func TestUpdateFromFeedback(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}
	model := NewModel(feedbackRepo, perfRepo, nil)

	ctx := context.Background()
	cardID := 123
	winResult := "win"

	feedback := &models.RecommendationFeedback{
		RecommendedCardID: &cardID,
		Action:            "accepted",
		OutcomeResult:     &winResult,
	}

	err := model.UpdateFromFeedback(ctx, feedback)
	if err != nil {
		t.Fatalf("UpdateFromFeedback failed: %v", err)
	}

	// Check that stats were updated
	model.acceptanceMu.RLock()
	stats, exists := model.cardAcceptanceRates[cardID]
	model.acceptanceMu.RUnlock()

	if !exists {
		t.Fatal("expected stats to be created")
	}
	if stats.Accepted != 1 {
		t.Errorf("expected 1 accepted, got %d", stats.Accepted)
	}
	if stats.WinsOnAccept != 1 {
		t.Errorf("expected 1 win, got %d", stats.WinsOnAccept)
	}
}

func TestUpdatePersonalPreferences(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}
	model := NewModel(feedbackRepo, perfRepo, nil)

	ctx := context.Background()
	accountID := 42

	deck := &DeckContext{
		DeckID:           "test-deck",
		ColorIdentity:    []string{"W", "U"},
		TypeDistribution: map[string]int{"Creature": 15, "Instant": 5},
		Archetype:        "UW Flyers",
	}

	// Only learns from wins
	err := model.UpdatePersonalPreferences(ctx, accountID, deck, "win")
	if err != nil {
		t.Fatalf("UpdatePersonalPreferences failed: %v", err)
	}

	model.personalMu.RLock()
	prefs, exists := model.personalPreferences[accountID]
	model.personalMu.RUnlock()

	if !exists {
		t.Fatal("expected preferences to be created")
	}
	if prefs.PreferredColors["W"] <= 0 {
		t.Error("expected W preference to increase")
	}
	if prefs.PreferredColors["U"] <= 0 {
		t.Error("expected U preference to increase")
	}
	if prefs.ArchetypePrefs["UW Flyers"] <= 0 {
		t.Error("expected archetype preference to increase")
	}
}

func TestSerializeDeserialize(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}
	model := NewModel(feedbackRepo, perfRepo, nil)

	// Add some data
	cardID := 123
	model.acceptanceMu.Lock()
	model.cardAcceptanceRates[cardID] = &acceptanceStats{
		Accepted:       5,
		Rejected:       2,
		WinsOnAccept:   3,
		LossesOnAccept: 2,
	}
	model.acceptanceMu.Unlock()

	model.affinityMu.Lock()
	model.archetypeAffinities["UW Flyers"] = map[int]float64{
		100: 0.8,
		101: 0.6,
	}
	model.affinityMu.Unlock()

	// Serialize
	data, err := model.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Create new model and deserialize
	model2 := NewModel(feedbackRepo, perfRepo, nil)
	err = model2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify data was restored
	model2.acceptanceMu.RLock()
	stats, exists := model2.cardAcceptanceRates[cardID]
	model2.acceptanceMu.RUnlock()

	if !exists {
		t.Fatal("expected acceptance stats to be restored")
	}
	if stats.Accepted != 5 {
		t.Errorf("expected 5 accepted, got %d", stats.Accepted)
	}

	model2.affinityMu.RLock()
	affin, exists := model2.archetypeAffinities["UW Flyers"]
	model2.affinityMu.RUnlock()

	if !exists {
		t.Fatal("expected archetype affinities to be restored")
	}
	if affin[100] != 0.8 {
		t.Errorf("expected affinity 0.8, got %f", affin[100])
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
	}{
		{
			name:     "identical vectors",
			a:        []float64{1, 0, 0},
			b:        []float64{1, 0, 0},
			expected: 1.0,
		},
		{
			name:     "orthogonal vectors",
			a:        []float64{1, 0, 0},
			b:        []float64{0, 1, 0},
			expected: 0.0,
		},
		{
			name:     "empty vectors",
			a:        []float64{},
			b:        []float64{},
			expected: 0.0,
		},
		{
			name:     "different lengths",
			a:        []float64{1, 0},
			b:        []float64{1, 0, 0},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CosineSimilarity(tt.a, tt.b)
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestBuildDeckEmbedding(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}
	model := NewModel(feedbackRepo, perfRepo, nil)

	deck := &DeckContext{
		DeckID:        "test-deck",
		ColorIdentity: []string{"W", "U", "B"},
		CMCDistribution: map[int]int{
			1: 2,
			2: 5,
			3: 5,
			4: 4,
			5: 3,
		},
		TypeDistribution: map[string]int{
			"Creature":    15,
			"Instant":     5,
			"Enchantment": 3,
		},
		Keywords:  map[string]int{"Flying": 5, "Lifelink": 3},
		Archetype: "Esper Control",
	}

	embed := model.buildDeckEmbedding(deck)

	if embed.DeckID != "test-deck" {
		t.Errorf("expected deck ID 'test-deck', got '%s'", embed.DeckID)
	}
	if embed.Archetype != "Esper Control" {
		t.Errorf("expected archetype 'Esper Control', got '%s'", embed.Archetype)
	}

	// Check color profile
	if embed.ColorProfile[0] != 1.0 { // W
		t.Error("expected W color profile to be 1.0")
	}
	if embed.ColorProfile[1] != 1.0 { // U
		t.Error("expected U color profile to be 1.0")
	}
	if embed.ColorProfile[2] != 1.0 { // B
		t.Error("expected B color profile to be 1.0")
	}

	// Check CMC profile is normalized
	total := 0.0
	for _, v := range embed.CMCProfile {
		total += v
	}
	if total > 1.01 || total < 0.99 { // Allow small float error
		t.Errorf("expected CMC profile sum ~1.0, got %f", total)
	}

	// Check keyword freq
	if embed.KeywordFreq["Flying"] != 5 {
		t.Errorf("expected Flying freq 5, got %f", embed.KeywordFreq["Flying"])
	}
}

func TestGetModelInfo(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}
	config := DefaultModelConfig()
	config.MinTrainingSamples = 100

	model := NewModel(feedbackRepo, perfRepo, config)

	info := model.GetModelInfo()

	if info.Version == "" {
		t.Error("expected version to be set")
	}
	if info.IsReady {
		t.Error("expected model to not be ready without sufficient training")
	}
	if info.CardFeaturesCount != 0 {
		t.Error("expected 0 card features initially")
	}

	// Add some features
	model.RegisterCardFeatures(1, &CardFeatures{CardID: 1, Name: "Test"})
	model.RegisterCardFeatures(2, &CardFeatures{CardID: 2, Name: "Test2"})

	info = model.GetModelInfo()
	if info.CardFeaturesCount != 2 {
		t.Errorf("expected 2 card features, got %d", info.CardFeaturesCount)
	}
}
