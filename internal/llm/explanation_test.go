package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/ml"
)

func TestDefaultExplanationConfig(t *testing.T) {
	config := DefaultExplanationConfig()

	if !config.UseLLM {
		t.Error("expected UseLLM to be true")
	}
	if config.MaxExplanationLength != 200 {
		t.Errorf("unexpected MaxExplanationLength: %d", config.MaxExplanationLength)
	}
	if config.Temperature != 0.7 {
		t.Errorf("unexpected Temperature: %f", config.Temperature)
	}
	if !config.FallbackToTemplate {
		t.Error("expected FallbackToTemplate to be true")
	}
}

func TestNewExplanationGenerator(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		gen := NewExplanationGenerator(nil, nil)
		if gen == nil {
			t.Fatal("expected non-nil generator")
		}
		if !gen.config.UseLLM {
			t.Error("expected default config")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		config := &ExplanationConfig{
			UseLLM:      false,
			Temperature: 0.5,
		}
		gen := NewExplanationGenerator(nil, config)
		if gen.config.UseLLM {
			t.Error("expected custom config")
		}
		if gen.config.Temperature != 0.5 {
			t.Errorf("unexpected temperature: %f", gen.config.Temperature)
		}
	})
}

func TestExplanationGenerator_GenerateCardExplanation_TemplateOnly(t *testing.T) {
	config := &ExplanationConfig{
		UseLLM:             false,
		FallbackToTemplate: true,
		CacheTTL:           1 * time.Hour,
	}
	gen := NewExplanationGenerator(nil, config)

	ctx := context.Background()
	card := &CardContext{
		CardID:   12345,
		CardName: "Lightning Bolt",
		Colors:   []string{"R"},
		Types:    []string{"Instant"},
		CMC:      1,
		Keywords: []string{"damage"},
	}
	deck := &DeckExplanationContext{
		DeckName:  "Mono Red Aggro",
		Format:    "standard",
		Colors:    []string{"R"},
		Archetype: "aggro",
		Strategy:  "aggro",
	}
	score := &ml.CardScore{
		CardID:  12345,
		Score:   0.85,
		Factors: []string{"good synergy", "tournament proven"},
	}

	explanation, err := gen.GenerateCardExplanation(ctx, card, deck, score)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if explanation == nil {
		t.Fatal("expected non-nil explanation")
	}
	if explanation.Source != "template" {
		t.Errorf("expected template source, got %s", explanation.Source)
	}
	if explanation.Explanation == "" {
		t.Error("expected non-empty explanation")
	}
	if explanation.CardName != "Lightning Bolt" {
		t.Errorf("unexpected card name: %s", explanation.CardName)
	}
}

func TestExplanationGenerator_GenerateCardExplanation_WithLLM(t *testing.T) {
	// Create mock Ollama server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/version":
			_ = json.NewEncoder(w).Encode(VersionResponse{Version: "0.1.0"})
		case "/api/tags":
			_ = json.NewEncoder(w).Encode(ListModelsResponse{
				Models: []ModelInfo{{Name: "qwen3:8b"}},
			})
		case "/api/generate":
			resp := GenerateResponse{
				Model:    "qwen3:8b",
				Response: "Lightning Bolt is an excellent choice for your aggressive deck. It provides efficient removal and direct damage, helping you close out games quickly.",
				Done:     true,
			}
			_ = json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	ollamaClient := NewOllamaClient(&OllamaConfig{
		BaseURL:          server.URL,
		Model:            "qwen3:8b",
		RequestTimeout:   5 * time.Second,
		InferenceTimeout: 30 * time.Second,
		AutoPullModel:    false,
	})

	// Check availability
	ctx := context.Background()
	status := ollamaClient.CheckAvailability(ctx)
	if !status.Available {
		t.Skip("Ollama not available for test")
	}

	config := &ExplanationConfig{
		UseLLM:             true,
		Temperature:        0.7,
		FallbackToTemplate: true,
		CacheTTL:           1 * time.Hour,
	}
	gen := NewExplanationGenerator(ollamaClient, config)

	card := &CardContext{
		CardID:   12345,
		CardName: "Lightning Bolt",
		Colors:   []string{"R"},
		Types:    []string{"Instant"},
		CMC:      1,
	}
	deck := &DeckExplanationContext{
		DeckName:  "Mono Red Aggro",
		Format:    "standard",
		Colors:    []string{"R"},
		Archetype: "aggro",
		Strategy:  "aggro",
	}
	score := &ml.CardScore{
		CardID:  12345,
		Score:   0.9,
		Factors: []string{"efficient removal"},
	}

	explanation, err := gen.GenerateCardExplanation(ctx, card, deck, score)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if explanation.Source != "llm" {
		t.Errorf("expected llm source, got %s", explanation.Source)
	}
	if explanation.Explanation == "" {
		t.Error("expected non-empty explanation")
	}
}

func TestExplanationGenerator_GenerateBatchExplanations(t *testing.T) {
	config := &ExplanationConfig{
		UseLLM:             false,
		FallbackToTemplate: true,
	}
	gen := NewExplanationGenerator(nil, config)

	ctx := context.Background()
	cards := []*CardContext{
		{CardID: 1, CardName: "Card One", Colors: []string{"R"}, Types: []string{"Creature"}},
		{CardID: 2, CardName: "Card Two", Colors: []string{"U"}, Types: []string{"Instant"}},
		{CardID: 3, CardName: "Card Three", Colors: []string{"G"}, Types: []string{"Sorcery"}},
	}
	deck := &DeckExplanationContext{
		DeckName:  "Test Deck",
		Format:    "standard",
		Colors:    []string{"R", "U", "G"},
		Archetype: "midrange",
		Strategy:  "midrange",
	}
	scores := []*ml.CardScore{
		{CardID: 1, Score: 0.8, Factors: []string{}},
		{CardID: 2, Score: 0.7, Factors: []string{}},
		{CardID: 3, Score: 0.6, Factors: []string{}},
	}

	explanations, err := gen.GenerateBatchExplanations(ctx, cards, deck, scores)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(explanations) != 3 {
		t.Errorf("expected 3 explanations, got %d", len(explanations))
	}

	for i, exp := range explanations {
		if exp.CardID != cards[i].CardID {
			t.Errorf("card ID mismatch at %d", i)
		}
		if exp.Explanation == "" {
			t.Errorf("empty explanation at %d", i)
		}
	}
}

func TestExplanationGenerator_Cache(t *testing.T) {
	config := &ExplanationConfig{
		UseLLM:             false,
		FallbackToTemplate: true,
		CacheTTL:           1 * time.Hour,
	}
	gen := NewExplanationGenerator(nil, config)

	ctx := context.Background()
	card := &CardContext{
		CardID:   12345,
		CardName: "Test Card",
		Colors:   []string{"R"},
		Types:    []string{"Creature"},
	}
	deck := &DeckExplanationContext{
		DeckName:  "Test Deck",
		Format:    "standard",
		Archetype: "aggro",
		Strategy:  "aggro",
	}
	score := &ml.CardScore{CardID: 12345, Score: 0.8}

	// First call
	exp1, _ := gen.GenerateCardExplanation(ctx, card, deck, score)

	// Cache the result manually since template doesn't cache
	gen.setCache(gen.makeCacheKey(card.CardID, deck.Archetype, deck.Format), exp1.Explanation)

	// Second call should use cache
	exp2, _ := gen.GenerateCardExplanation(ctx, card, deck, score)

	if exp2.Source != "cached" {
		t.Errorf("expected cached source, got %s", exp2.Source)
	}
}

func TestExplanationGenerator_ClearCache(t *testing.T) {
	config := &ExplanationConfig{
		UseLLM:             false,
		FallbackToTemplate: true,
		CacheTTL:           1 * time.Hour,
	}
	gen := NewExplanationGenerator(nil, config)

	// Add something to cache
	gen.setCache("test-key", "test explanation")

	// Verify it's there
	if cached := gen.getFromCache("test-key"); cached == nil {
		t.Error("expected cache entry")
	}

	// Clear cache
	gen.ClearCache()

	// Verify it's gone
	if cached := gen.getFromCache("test-key"); cached != nil {
		t.Error("expected cache to be cleared")
	}
}

func TestExplanationGenerator_ExtractSummary(t *testing.T) {
	gen := NewExplanationGenerator(nil, nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"short sentence",
			"This is a great card.",
			"This is a great card.",
		},
		{
			"multiple sentences",
			"This is a great card. It provides excellent value. Recommended for aggro.",
			"This is a great card.",
		},
		{
			"very long text",
			"This is a very long explanation that goes on and on and on and should be truncated because it exceeds the maximum length allowed for a summary which is typically around eighty characters.",
			"This is a very long explanation that goes on and on and on and should be truncat...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.extractSummary(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExplanationGenerator_CountMatchingColors(t *testing.T) {
	gen := NewExplanationGenerator(nil, nil)

	tests := []struct {
		name     string
		colors1  []string
		colors2  []string
		expected int
	}{
		{"all match", []string{"R", "U"}, []string{"R", "U", "B"}, 2},
		{"partial match", []string{"R", "G"}, []string{"R", "U"}, 1},
		{"no match", []string{"W", "B"}, []string{"R", "U", "G"}, 0},
		{"empty colors", []string{}, []string{"R", "U"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.countMatchingColors(tt.colors1, tt.colors2)
			if result != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, result)
			}
		})
	}
}

func TestExplanationGenerator_ContainsString(t *testing.T) {
	gen := NewExplanationGenerator(nil, nil)

	tests := []struct {
		name     string
		slice    []string
		search   string
		expected bool
	}{
		{"found exact", []string{"Creature", "Instant"}, "Creature", true},
		{"found case insensitive", []string{"CREATURE", "Instant"}, "creature", true},
		{"not found", []string{"Creature", "Instant"}, "Sorcery", false},
		{"empty slice", []string{}, "Creature", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := gen.containsString(tt.slice, tt.search)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestExplanationGenerator_GetConfig(t *testing.T) {
	config := &ExplanationConfig{
		UseLLM:      false,
		Temperature: 0.5,
	}
	gen := NewExplanationGenerator(nil, config)

	retrieved := gen.GetConfig()
	if retrieved.Temperature != 0.5 {
		t.Errorf("unexpected temperature: %f", retrieved.Temperature)
	}
}

func TestExplanationGenerator_UpdateConfig(t *testing.T) {
	gen := NewExplanationGenerator(nil, nil)

	// Add something to cache
	gen.setCache("test-key", "test explanation")

	// Update config
	newConfig := &ExplanationConfig{
		UseLLM:      false,
		Temperature: 0.3,
	}
	gen.UpdateConfig(newConfig)

	// Config should be updated
	if gen.config.Temperature != 0.3 {
		t.Error("config not updated")
	}

	// Cache should be cleared
	if cached := gen.getFromCache("test-key"); cached != nil {
		t.Error("expected cache to be cleared after config update")
	}
}

func TestExplanationGenerator_IsLLMAvailable(t *testing.T) {
	t.Run("nil client", func(t *testing.T) {
		gen := NewExplanationGenerator(nil, nil)
		if gen.IsLLMAvailable() {
			t.Error("expected false with nil client")
		}
	})

	t.Run("unavailable client", func(t *testing.T) {
		client := NewOllamaClient(&OllamaConfig{
			BaseURL: "http://localhost:99999",
		})
		gen := NewExplanationGenerator(client, nil)
		if gen.IsLLMAvailable() {
			t.Error("expected false with unavailable client")
		}
	})
}

func TestCardExplanation_Fields(t *testing.T) {
	exp := &CardExplanation{
		CardID:      12345,
		CardName:    "Test Card",
		Summary:     "Brief summary.",
		Explanation: "Full explanation text.",
		Factors:     []string{"factor1", "factor2"},
		Source:      "llm",
		Confidence:  0.9,
	}

	if exp.CardID != 12345 {
		t.Errorf("unexpected CardID: %d", exp.CardID)
	}
	if exp.CardName != "Test Card" {
		t.Errorf("unexpected CardName: %s", exp.CardName)
	}
	if exp.Source != "llm" {
		t.Errorf("unexpected Source: %s", exp.Source)
	}
	if len(exp.Factors) != 2 {
		t.Errorf("unexpected Factors length: %d", len(exp.Factors))
	}
}

func TestCardContext_Fields(t *testing.T) {
	card := &CardContext{
		CardID:        12345,
		CardName:      "Test Card",
		Colors:        []string{"R", "U"},
		Types:         []string{"Creature", "Wizard"},
		CMC:           3,
		Keywords:      []string{"Flying"},
		CreatureTypes: []string{"Human", "Wizard"},
		Rarity:        "Rare",
	}

	if card.CardID != 12345 {
		t.Errorf("unexpected CardID: %d", card.CardID)
	}
	if len(card.Colors) != 2 {
		t.Errorf("unexpected Colors length: %d", len(card.Colors))
	}
	if card.CMC != 3 {
		t.Errorf("unexpected CMC: %f", card.CMC)
	}
}

func TestDeckExplanationContext_Fields(t *testing.T) {
	deck := &DeckExplanationContext{
		DeckName:      "Test Deck",
		Format:        "standard",
		Colors:        []string{"R", "U", "B"},
		Archetype:     "control",
		Strategy:      "control",
		CardCount:     60,
		CreatureCount: 15,
		SpellCount:    30,
		LandCount:     25,
	}

	if deck.DeckName != "Test Deck" {
		t.Errorf("unexpected DeckName: %s", deck.DeckName)
	}
	if deck.CardCount != 60 {
		t.Errorf("unexpected CardCount: %d", deck.CardCount)
	}
}

func TestExplanationGenerator_TemplateExplanation_HighScore(t *testing.T) {
	gen := NewExplanationGenerator(nil, nil)

	card := &CardContext{
		CardID:   1,
		CardName: "Great Card",
		Colors:   []string{"R"},
		Types:    []string{"Creature"},
		CMC:      2,
	}
	deck := &DeckExplanationContext{
		Colors:    []string{"R"},
		Archetype: "aggro",
		Strategy:  "aggro",
	}
	score := &ml.CardScore{Score: 0.9}

	explanation := gen.generateTemplateExplanation(card, deck, score)

	if explanation == "" {
		t.Error("expected non-empty explanation")
	}
	// High score should mention "excellent"
	if !containsIgnoreCase(explanation, "excellent") {
		t.Error("expected 'excellent' in high-score explanation")
	}
}

func TestExplanationGenerator_TemplateExplanation_MediumScore(t *testing.T) {
	gen := NewExplanationGenerator(nil, nil)

	card := &CardContext{
		CardID:   1,
		CardName: "Good Card",
		Colors:   []string{"R"},
		Types:    []string{"Creature"},
	}
	deck := &DeckExplanationContext{
		Colors:    []string{"R", "U"},
		Archetype: "midrange",
		Strategy:  "midrange",
	}
	score := &ml.CardScore{Score: 0.65}

	explanation := gen.generateTemplateExplanation(card, deck, score)

	if !containsIgnoreCase(explanation, "good") {
		t.Error("expected 'good' in medium-score explanation")
	}
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > 0 && containsIgnoreCase(s[1:], substr) ||
			len(s) >= len(substr) && equalFoldPrefix(s, substr))
}

func equalFoldPrefix(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := 0; i < len(prefix); i++ {
		if toLower(s[i]) != toLower(prefix[i]) {
			return false
		}
	}
	return true
}

func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}
