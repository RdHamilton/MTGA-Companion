package prediction

import (
	"testing"
)

func TestPredictWinRate_EmptyDeck(t *testing.T) {
	_, err := PredictWinRate([]Card{})
	if err == nil {
		t.Error("Expected error for empty deck")
	}
}

func TestPredictWinRate_BasicDeck(t *testing.T) {
	cards := []Card{
		{Name: "Card 1", CMC: 2, Color: "R", GIHWR: 0.52, Rarity: "common"},
		{Name: "Card 2", CMC: 3, Color: "R", GIHWR: 0.50, Rarity: "common"},
		{Name: "Card 3", CMC: 4, Color: "R", GIHWR: 0.48, Rarity: "common"},
	}

	prediction, err := PredictWinRate(cards)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Check that prediction is in valid range
	if prediction.PredictedWinRate < 0.30 || prediction.PredictedWinRate > 0.70 {
		t.Errorf("Prediction out of range: %f", prediction.PredictedWinRate)
	}

	// Check that min < predicted < max
	if prediction.PredictedWinRateMin > prediction.PredictedWinRate {
		t.Error("Min should be less than predicted")
	}
	if prediction.PredictedWinRateMax < prediction.PredictedWinRate {
		t.Error("Max should be greater than predicted")
	}
}

func TestPredictWinRate_BombDeck(t *testing.T) {
	cards := []Card{
		{Name: "Bomb 1", CMC: 4, Color: "W", GIHWR: 0.65, Rarity: "mythic"},
		{Name: "Bomb 2", CMC: 5, Color: "W", GIHWR: 0.62, Rarity: "rare"},
		{Name: "Filler", CMC: 3, Color: "W", GIHWR: 0.50, Rarity: "common"},
	}

	prediction, err := PredictWinRate(cards)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Deck with bombs should have higher prediction
	if prediction.Factors.BombBonus <= 0 {
		t.Error("Expected bomb bonus for deck with high GIHWR cards")
	}

	if len(prediction.Factors.HighPerformers) == 0 {
		t.Error("Expected high performers to be identified")
	}
}

func TestPredictWinRate_TwoColorDeck(t *testing.T) {
	cards := []Card{
		{Name: "Card 1", CMC: 2, Color: "W", GIHWR: 0.52},
		{Name: "Card 2", CMC: 3, Color: "U", GIHWR: 0.50},
	}

	prediction, err := PredictWinRate(cards)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if prediction.Factors.ColorAdjustment != twoColorBonus {
		t.Errorf("Expected two-color bonus %f, got %f", twoColorBonus, prediction.Factors.ColorAdjustment)
	}
}

func TestPredictWinRate_ThreeColorDeck(t *testing.T) {
	cards := []Card{
		{Name: "Card 1", CMC: 2, Color: "W", GIHWR: 0.52},
		{Name: "Card 2", CMC: 3, Color: "U", GIHWR: 0.50},
		{Name: "Card 3", CMC: 4, Color: "B", GIHWR: 0.48},
	}

	prediction, err := PredictWinRate(cards)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if prediction.Factors.ColorAdjustment != threeColorPenalty {
		t.Errorf("Expected three-color penalty %f, got %f", threeColorPenalty, prediction.Factors.ColorAdjustment)
	}
}

func TestPredictWinRate_SynergyFactors(t *testing.T) {
	cards := []Card{
		{Name: "Card 1", CMC: 2, Color: "R", GIHWR: 0.52},
		{Name: "Card 2", CMC: 3, Color: "R", GIHWR: 0.50},
	}

	prediction, err := PredictWinRate(cards)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Synergy score should be calculated
	if prediction.Factors.SynergyScore < 0 || prediction.Factors.SynergyScore > 1 {
		t.Errorf("Synergy score out of range: %f", prediction.Factors.SynergyScore)
	}

	// Synergy details should be populated
	if prediction.Factors.SynergyDetails == nil {
		t.Error("Expected synergy details to be populated")
	}
}

func TestPredictWinRate_ConfidenceLevels(t *testing.T) {
	// Small deck -> low confidence
	smallDeck := make([]Card, 20)
	for i := 0; i < 20; i++ {
		smallDeck[i] = Card{Name: "Card", CMC: 3, Color: "R", GIHWR: 0.50}
	}

	prediction, _ := PredictWinRate(smallDeck)
	if prediction.Factors.ConfidenceLevel != "low" {
		t.Errorf("Expected low confidence for small deck, got %s", prediction.Factors.ConfidenceLevel)
	}

	// Medium deck -> medium confidence
	mediumDeck := make([]Card, 35)
	for i := 0; i < 35; i++ {
		mediumDeck[i] = Card{Name: "Card", CMC: 3, Color: "R", GIHWR: 0.50}
	}

	prediction, _ = PredictWinRate(mediumDeck)
	if prediction.Factors.ConfidenceLevel != "medium" {
		t.Errorf("Expected medium confidence for medium deck, got %s", prediction.Factors.ConfidenceLevel)
	}

	// Large deck -> high confidence
	largeDeck := make([]Card, 45)
	for i := 0; i < 45; i++ {
		largeDeck[i] = Card{Name: "Card", CMC: 3, Color: "R", GIHWR: 0.50}
	}

	prediction, _ = PredictWinRate(largeDeck)
	if prediction.Factors.ConfidenceLevel != "high" {
		t.Errorf("Expected high confidence for large deck, got %s", prediction.Factors.ConfidenceLevel)
	}
}

func TestPredictWinRate_LowPerformers(t *testing.T) {
	cards := []Card{
		{Name: "Bad Card 1", CMC: 5, Color: "R", GIHWR: 0.40, Rarity: "common"},
		{Name: "Bad Card 2", CMC: 6, Color: "R", GIHWR: 0.42, Rarity: "common"},
		{Name: "OK Card", CMC: 3, Color: "R", GIHWR: 0.50, Rarity: "common"},
	}

	prediction, err := PredictWinRate(cards)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(prediction.Factors.LowPerformers) == 0 {
		t.Error("Expected low performers to be identified")
	}
}

func TestEvaluateCurve(t *testing.T) {
	// Good curve (lots of 2-4 drops)
	goodCurve := map[int]int{
		1: 2,
		2: 5,
		3: 5,
		4: 3,
		5: 1,
	}

	score := evaluateCurve(goodCurve, 16)
	if score < 0.5 {
		t.Errorf("Expected above-average score for good curve, got %f", score)
	}

	// Bad curve (too many expensive cards)
	badCurve := map[int]int{
		1: 1,
		2: 2,
		5: 5,
		6: 4,
		7: 3,
	}

	score = evaluateCurve(badCurve, 15)
	if score > 0.5 {
		t.Errorf("Expected below-average score for heavy curve, got %f", score)
	}
}

func TestGenerateExplanation(t *testing.T) {
	factors := PredictionFactors{
		DeckAverageGIHWR:  0.55,
		ColorDistribution: map[string]int{"W": 5, "U": 5},
		CurveScore:        0.65,
		HighPerformers:    []string{"Bomb 1", "Bomb 2"},
		SynergyDetails: &SynergyResult{
			TribalSynergies: 3,
			MechSynergies:   2,
		},
	}

	explanation := generateExplanation(factors, 0.56)

	if explanation == "" {
		t.Error("Expected non-empty explanation")
	}

	// Check that it contains win rate
	if !containsIgnoreCase(explanation, "56") {
		t.Error("Expected explanation to contain win rate percentage")
	}
}

func TestPredictionFactorsToJSON(t *testing.T) {
	factors := &PredictionFactors{
		DeckAverageGIHWR:  0.52,
		ColorAdjustment:   0.02,
		TotalCards:        40,
		ColorDistribution: map[string]int{"W": 20, "U": 20},
	}

	json, err := factors.ToJSON()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if json == "" {
		t.Error("Expected non-empty JSON")
	}

	// Parse it back
	parsed, err := FromJSON(json)
	if err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	if parsed.DeckAverageGIHWR != factors.DeckAverageGIHWR {
		t.Error("JSON round-trip failed")
	}
}

func TestFromJSON_Invalid(t *testing.T) {
	_, err := FromJSON("invalid json")
	if err == nil {
		t.Error("Expected error for invalid JSON")
	}
}
