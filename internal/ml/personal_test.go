package ml

import (
	"context"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

func TestNewPersonalLearner(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	learner := NewPersonalLearner(perfRepo, feedbackRepo, nil)

	if learner == nil {
		t.Fatal("expected learner to be created")
	}

	if learner.config == nil {
		t.Error("expected config to be set")
	}
}

func TestDefaultPersonalLearnerConfig(t *testing.T) {
	config := DefaultPersonalLearnerConfig()

	if config.MinMatchesForPersonalization <= 0 {
		t.Error("expected positive MinMatchesForPersonalization")
	}
	if config.LearningRate <= 0 || config.LearningRate > 1 {
		t.Error("expected learning rate between 0 and 1")
	}
	if config.MaxHistorySize <= 0 {
		t.Error("expected positive MaxHistorySize")
	}
}

func TestGetProfile(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	learner := NewPersonalLearner(perfRepo, feedbackRepo, nil)
	ctx := context.Background()

	profile, err := learner.GetProfile(ctx, 123)
	if err != nil {
		t.Fatalf("GetProfile failed: %v", err)
	}

	if profile == nil {
		t.Fatal("expected profile to be created")
	}

	if profile.AccountID != 123 {
		t.Errorf("expected account ID 123, got %d", profile.AccountID)
	}

	// Getting same profile should return cached version
	profile2, _ := learner.GetProfile(ctx, 123)
	if profile != profile2 {
		t.Error("expected same profile instance to be returned")
	}
}

func TestLearnFromMatch(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	config := DefaultPersonalLearnerConfig()
	config.LearningRate = 0.5 // Higher learning rate for testing

	learner := NewPersonalLearner(perfRepo, feedbackRepo, config)
	ctx := context.Background()

	matchData := &MatchLearningData{
		DeckID:     "deck-1",
		DeckColors: []string{"W", "U"},
		TypeDistribution: map[string]int{
			"Creature": 15,
			"Instant":  5,
			"Sorcery":  3,
		},
		TotalCards: 23,
		AverageCMC: 2.5,
		Archetype:  "UW Flyers",
		Result:     "win",
	}

	err := learner.LearnFromMatch(ctx, 123, matchData)
	if err != nil {
		t.Fatalf("LearnFromMatch failed: %v", err)
	}

	profile, _ := learner.GetProfile(ctx, 123)

	// Check color preferences were updated
	if profile.ColorPreferences["W"] <= 0 {
		t.Error("expected W preference to increase")
	}
	if profile.ColorPreferences["U"] <= 0 {
		t.Error("expected U preference to increase")
	}

	// Check archetype preference was updated
	if profile.ArchetypePreferences["UW Flyers"] <= 0 {
		t.Error("expected archetype preference to increase")
	}

	// Check stats were updated
	if profile.TotalMatches != 1 {
		t.Errorf("expected 1 match, got %d", profile.TotalMatches)
	}
	if profile.TotalWins != 1 {
		t.Errorf("expected 1 win, got %d", profile.TotalWins)
	}
}

func TestLearnFromMatchLoss(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	config := DefaultPersonalLearnerConfig()
	config.LearningRate = 0.5

	learner := NewPersonalLearner(perfRepo, feedbackRepo, config)
	ctx := context.Background()

	// First a win to establish baseline
	_ = learner.LearnFromMatch(ctx, 123, &MatchLearningData{
		DeckColors: []string{"R"},
		Archetype:  "Aggro",
		Result:     "win",
	})

	// Then a loss
	err := learner.LearnFromMatch(ctx, 123, &MatchLearningData{
		DeckColors: []string{"R"},
		Archetype:  "Aggro",
		Result:     "loss",
	})
	if err != nil {
		t.Fatalf("LearnFromMatch failed: %v", err)
	}

	profile, _ := learner.GetProfile(ctx, 123)

	// Preference should decrease on loss
	if profile.ColorPreferences["R"] >= 1.0 {
		t.Error("expected R preference to not be maxed after loss")
	}

	if profile.TotalMatches != 2 {
		t.Errorf("expected 2 matches, got %d", profile.TotalMatches)
	}
	if profile.TotalWins != 1 {
		t.Errorf("expected 1 win, got %d", profile.TotalWins)
	}
}

func TestLearnFromFeedback(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	config := DefaultPersonalLearnerConfig()
	config.LearningRate = 0.5

	learner := NewPersonalLearner(perfRepo, feedbackRepo, config)
	ctx := context.Background()

	cardID := 12345

	// Accepted feedback
	feedback := &models.RecommendationFeedback{
		AccountID:         123,
		RecommendedCardID: &cardID,
		Action:            "accepted",
	}

	err := learner.LearnFromFeedback(ctx, 123, feedback)
	if err != nil {
		t.Fatalf("LearnFromFeedback failed: %v", err)
	}

	profile, _ := learner.GetProfile(ctx, 123)

	if profile.CardPreferences[cardID] <= 0 {
		t.Error("expected card preference to increase on accept")
	}

	// Rejected feedback
	feedback.Action = "rejected"
	_ = learner.LearnFromFeedback(ctx, 123, feedback)

	// Preference should decrease
	if profile.CardPreferences[cardID] >= 1.0 {
		t.Error("expected card preference to decrease on reject")
	}
}

func TestLearnFromFeedbackWithOutcome(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	config := DefaultPersonalLearnerConfig()
	config.LearningRate = 0.5

	learner := NewPersonalLearner(perfRepo, feedbackRepo, config)
	ctx := context.Background()

	cardID := 12345
	winResult := "win"

	feedback := &models.RecommendationFeedback{
		AccountID:         123,
		RecommendedCardID: &cardID,
		Action:            "accepted",
		OutcomeResult:     &winResult,
	}

	_ = learner.LearnFromFeedback(ctx, 123, feedback)

	profile, _ := learner.GetProfile(ctx, 123)

	// Should have higher preference due to win outcome
	if profile.CardPreferences[cardID] < 0.5 {
		t.Error("expected higher card preference with win outcome")
	}
}

func TestGetPersonalScore(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	learner := NewPersonalLearner(perfRepo, feedbackRepo, nil)
	ctx := context.Background()

	// Build up profile with some data
	for i := 0; i < 15; i++ {
		_ = learner.LearnFromMatch(ctx, 123, &MatchLearningData{
			DeckColors: []string{"W", "U"},
			TypeDistribution: map[string]int{
				"Creature": 15,
			},
			TotalCards: 20,
			AverageCMC: 2.5,
			Archetype:  "UW Flyers",
			Result:     "win",
		})
	}

	card := &CardFeatures{
		CardID: 100,
		Colors: []string{"W", "U"},
		Types:  []string{"Creature"},
		CMC:    2.0,
	}

	score, err := learner.GetPersonalScore(ctx, 123, card)
	if err != nil {
		t.Fatalf("GetPersonalScore failed: %v", err)
	}

	// Should have high score since card matches learned preferences
	if score < 0.5 {
		t.Errorf("expected score > 0.5 for matching card, got %f", score)
	}

	// Card that doesn't match should have lower score
	unmatchedCard := &CardFeatures{
		CardID: 101,
		Colors: []string{"R", "G"},
		Types:  []string{"Sorcery"},
		CMC:    6.0,
	}

	unmatchedScore, _ := learner.GetPersonalScore(ctx, 123, unmatchedCard)
	if unmatchedScore >= score {
		t.Error("expected unmatched card to have lower score")
	}
}

func TestIsPersonalizationReady(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	config := DefaultPersonalLearnerConfig()
	config.MinMatchesForPersonalization = 10

	learner := NewPersonalLearner(perfRepo, feedbackRepo, config)
	ctx := context.Background()

	// Initially not ready
	ready, confidence := learner.IsPersonalizationReady(ctx, 123)
	if ready {
		t.Error("expected not ready with no matches")
	}
	if confidence > 0 {
		t.Error("expected zero confidence with no matches")
	}

	// Add some matches
	for i := 0; i < 5; i++ {
		_ = learner.LearnFromMatch(ctx, 123, &MatchLearningData{
			DeckColors: []string{"W"},
			Result:     "win",
		})
	}

	ready, _ = learner.IsPersonalizationReady(ctx, 123)
	if ready {
		t.Error("expected not ready with only 5 matches")
	}

	// Add more matches to reach threshold
	for i := 0; i < 6; i++ {
		_ = learner.LearnFromMatch(ctx, 123, &MatchLearningData{
			DeckColors: []string{"W"},
			Result:     "win",
		})
	}

	ready, confidence = learner.IsPersonalizationReady(ctx, 123)
	if !ready {
		t.Error("expected ready with 11 matches")
	}
	if confidence <= 0 {
		t.Error("expected positive confidence")
	}
}

func TestResetProfile(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	learner := NewPersonalLearner(perfRepo, feedbackRepo, nil)
	ctx := context.Background()

	// Build profile
	_ = learner.LearnFromMatch(ctx, 123, &MatchLearningData{
		DeckColors: []string{"W"},
		Result:     "win",
	})

	profile, _ := learner.GetProfile(ctx, 123)
	if profile.TotalMatches == 0 {
		t.Error("expected profile to have data")
	}

	// Reset
	err := learner.ResetProfile(ctx, 123)
	if err != nil {
		t.Fatalf("ResetProfile failed: %v", err)
	}

	// Get profile again - should be fresh
	profile, _ = learner.GetProfile(ctx, 123)
	if profile.TotalMatches != 0 {
		t.Error("expected fresh profile after reset")
	}
}

func TestGetProfileStats(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	learner := NewPersonalLearner(perfRepo, feedbackRepo, nil)
	ctx := context.Background()

	// Build profile with variety
	for i := 0; i < 15; i++ {
		_ = learner.LearnFromMatch(ctx, 123, &MatchLearningData{
			DeckColors:       []string{"W", "U"},
			TypeDistribution: map[string]int{"Creature": 15},
			TotalCards:       20,
			AverageCMC:       2.3,
			Archetype:        "UW Flyers",
			Result:           "win",
		})
	}

	stats, err := learner.GetProfileStats(ctx, 123)
	if err != nil {
		t.Fatalf("GetProfileStats failed: %v", err)
	}

	if stats.TotalMatches != 15 {
		t.Errorf("expected 15 matches, got %d", stats.TotalMatches)
	}
	if stats.TotalWins != 15 {
		t.Errorf("expected 15 wins, got %d", stats.TotalWins)
	}
	if !stats.IsReady {
		t.Error("expected profile to be ready")
	}
	if len(stats.PreferredColors) == 0 {
		t.Error("expected preferred colors to be detected")
	}
}

func TestCalculateConfidence(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	config := DefaultPersonalLearnerConfig()
	config.MinMatchesForPersonalization = 10

	learner := NewPersonalLearner(perfRepo, feedbackRepo, config)

	tests := []struct {
		name       string
		matchCount int
		expectZero bool
		expectHigh bool
	}{
		{"below threshold", 5, true, false},
		{"at threshold", 10, false, false},
		{"moderate data", 50, false, false},
		{"lots of data", 200, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			confidence := learner.calculateConfidence(tt.matchCount)

			if tt.expectZero && confidence != 0 {
				t.Errorf("expected zero confidence, got %f", confidence)
			}
			if tt.expectHigh && confidence < 0.9 {
				t.Errorf("expected high confidence, got %f", confidence)
			}
		})
	}
}

func TestPersonalLearnerSerializeDeserialize(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	learner := NewPersonalLearner(perfRepo, feedbackRepo, nil)
	ctx := context.Background()

	// Build profile
	for i := 0; i < 10; i++ {
		_ = learner.LearnFromMatch(ctx, 123, &MatchLearningData{
			DeckColors: []string{"W", "U"},
			Archetype:  "UW Flyers",
			Result:     "win",
		})
	}

	// Serialize
	data, err := learner.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	// Create new learner and deserialize
	learner2 := NewPersonalLearner(perfRepo, feedbackRepo, nil)
	err = learner2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	// Verify data was restored
	profile, _ := learner2.GetProfile(ctx, 123)
	if profile.TotalMatches != 10 {
		t.Errorf("expected 10 matches after deserialize, got %d", profile.TotalMatches)
	}
	if profile.ColorPreferences["W"] <= 0 {
		t.Error("expected W preference to be restored")
	}
}

func TestPlayStyleDetection(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	config := DefaultPersonalLearnerConfig()
	config.StyleDetectionEnabled = true

	learner := NewPersonalLearner(perfRepo, feedbackRepo, config)
	ctx := context.Background()

	// Play lots of aggro games
	for i := 0; i < 20; i++ {
		_ = learner.LearnFromMatch(ctx, 123, &MatchLearningData{
			DeckColors: []string{"R"},
			TypeDistribution: map[string]int{
				"Creature": 18,
				"Instant":  3,
			},
			TotalCards: 21,
			AverageCMC: 2.0,
			Archetype:  "Mono Red Aggro",
			Result:     "win",
		})
	}

	profile, _ := learner.GetProfile(ctx, 123)

	if profile.Style.Aggro < profile.Style.Control {
		t.Error("expected aggro style to be higher than control")
	}
	if !profile.Style.PreferCreatureHeavy {
		t.Error("expected creature-heavy preference")
	}
}

func TestAnalyzeHistory(t *testing.T) {
	feedbackRepo := &mockFeedbackRepo{}
	perfRepo := &mockPerformanceRepo{}

	learner := NewPersonalLearner(perfRepo, feedbackRepo, nil)

	profile := learner.newProfile(123)

	arch1 := "UW Flyers"
	arch2 := "BR Sacrifice"

	history := []*models.DeckPerformanceHistory{
		{ColorIdentity: "WU", Archetype: &arch1, Result: "win", MatchTimestamp: time.Now()},
		{ColorIdentity: "WU", Archetype: &arch1, Result: "win", MatchTimestamp: time.Now()},
		{ColorIdentity: "WU", Archetype: &arch1, Result: "loss", MatchTimestamp: time.Now()},
		{ColorIdentity: "BR", Archetype: &arch2, Result: "win", MatchTimestamp: time.Now()},
		{ColorIdentity: "BR", Archetype: &arch2, Result: "loss", MatchTimestamp: time.Now()},
		{ColorIdentity: "BR", Archetype: &arch2, Result: "loss", MatchTimestamp: time.Now()},
	}

	learner.analyzeHistory(profile, history)

	if profile.TotalMatches != 6 {
		t.Errorf("expected 6 matches, got %d", profile.TotalMatches)
	}
	if profile.TotalWins != 3 {
		t.Errorf("expected 3 wins, got %d", profile.TotalWins)
	}

	// UW Flyers should have higher preference (2/3 wins)
	if profile.ArchetypePreferences[arch1] <= profile.ArchetypePreferences[arch2] {
		t.Error("expected UW Flyers to have higher preference than BR Sacrifice")
	}
}
