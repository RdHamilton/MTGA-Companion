package llm

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/ml"
)

// ExplanationConfig configures the explanation generator.
type ExplanationConfig struct {
	// UseLLM enables LLM-powered explanations when available.
	UseLLM bool

	// MaxExplanationLength is the maximum length of generated explanations.
	MaxExplanationLength int

	// Temperature controls creativity in LLM responses.
	Temperature float64

	// CacheTTL is how long to cache explanations.
	CacheTTL time.Duration

	// FallbackToTemplate uses templates when LLM is unavailable.
	FallbackToTemplate bool
}

// DefaultExplanationConfig returns sensible defaults.
func DefaultExplanationConfig() *ExplanationConfig {
	return &ExplanationConfig{
		UseLLM:               true,
		MaxExplanationLength: 200,
		Temperature:          0.7,
		CacheTTL:             1 * time.Hour,
		FallbackToTemplate:   true,
	}
}

// ExplanationGenerator generates natural language explanations for recommendations.
type ExplanationGenerator struct {
	config       *ExplanationConfig
	ollamaClient *OllamaClient
	cache        map[string]*CachedExplanation
	cacheMu      sync.RWMutex
}

// CachedExplanation represents a cached explanation.
type CachedExplanation struct {
	Text      string
	CreatedAt time.Time
}

// CardExplanation represents an explanation for a card recommendation.
type CardExplanation struct {
	CardID      int      `json:"card_id"`
	CardName    string   `json:"card_name"`
	Summary     string   `json:"summary"`     // Brief one-line summary
	Explanation string   `json:"explanation"` // Full explanation
	Factors     []string `json:"factors"`     // Contributing factors
	Source      string   `json:"source"`      // "llm" or "template"
	Confidence  float64  `json:"confidence"`  // How confident in the explanation
}

// CardContext provides context about a card for explanation generation.
type CardContext struct {
	CardID        int
	CardName      string
	Colors        []string
	Types         []string
	CMC           float64
	Keywords      []string
	CreatureTypes []string
	Rarity        string
}

// DeckExplanationContext provides deck context for explanations.
type DeckExplanationContext struct {
	DeckName      string
	Format        string
	Colors        []string
	Archetype     string
	Strategy      string // aggro, control, midrange, tempo
	CardCount     int
	CreatureCount int
	SpellCount    int
	LandCount     int
}

// NewExplanationGenerator creates a new explanation generator.
func NewExplanationGenerator(ollamaClient *OllamaClient, config *ExplanationConfig) *ExplanationGenerator {
	if config == nil {
		config = DefaultExplanationConfig()
	}

	return &ExplanationGenerator{
		config:       config,
		ollamaClient: ollamaClient,
		cache:        make(map[string]*CachedExplanation),
	}
}

// GenerateCardExplanation generates an explanation for why a card was recommended.
func (g *ExplanationGenerator) GenerateCardExplanation(
	ctx context.Context,
	card *CardContext,
	deck *DeckExplanationContext,
	score *ml.CardScore,
) (*CardExplanation, error) {
	explanation := &CardExplanation{
		CardID:   card.CardID,
		CardName: card.CardName,
		Factors:  score.Factors,
	}

	// Try cache first
	cacheKey := g.makeCacheKey(card.CardID, deck.Archetype, deck.Format)
	if cached := g.getFromCache(cacheKey); cached != nil {
		explanation.Explanation = cached.Text
		explanation.Source = "cached"
		explanation.Summary = g.extractSummary(cached.Text)
		return explanation, nil
	}

	// Try LLM if enabled and available
	if g.config.UseLLM && g.ollamaClient != nil && g.ollamaClient.IsAvailable() {
		llmExplanation, err := g.generateLLMExplanation(ctx, card, deck, score)
		if err == nil {
			explanation.Explanation = llmExplanation
			explanation.Source = "llm"
			explanation.Summary = g.extractSummary(llmExplanation)
			explanation.Confidence = 0.9
			g.setCache(cacheKey, llmExplanation)
			return explanation, nil
		}
		// Fall through to template if LLM fails
	}

	// Use template fallback
	if g.config.FallbackToTemplate {
		templateExplanation := g.generateTemplateExplanation(card, deck, score)
		explanation.Explanation = templateExplanation
		explanation.Source = "template"
		explanation.Summary = g.extractSummary(templateExplanation)
		explanation.Confidence = 0.7
		return explanation, nil
	}

	return nil, fmt.Errorf("no explanation method available")
}

// GenerateBatchExplanations generates explanations for multiple cards.
func (g *ExplanationGenerator) GenerateBatchExplanations(
	ctx context.Context,
	cards []*CardContext,
	deck *DeckExplanationContext,
	scores []*ml.CardScore,
) ([]*CardExplanation, error) {
	explanations := make([]*CardExplanation, 0, len(cards))

	for i, card := range cards {
		var score *ml.CardScore
		if i < len(scores) {
			score = scores[i]
		} else {
			score = &ml.CardScore{CardID: card.CardID, Factors: []string{}}
		}

		explanation, err := g.GenerateCardExplanation(ctx, card, deck, score)
		if err != nil {
			// Create a basic explanation on error
			explanation = &CardExplanation{
				CardID:      card.CardID,
				CardName:    card.CardName,
				Explanation: "Recommended based on deck composition.",
				Source:      "default",
				Confidence:  0.5,
			}
		}
		explanations = append(explanations, explanation)
	}

	return explanations, nil
}

// generateLLMExplanation uses the LLM to generate an explanation.
func (g *ExplanationGenerator) generateLLMExplanation(
	ctx context.Context,
	card *CardContext,
	deck *DeckExplanationContext,
	score *ml.CardScore,
) (string, error) {
	systemPrompt := `You are a Magic: The Gathering deck building assistant.
Explain card recommendations concisely and helpfully.
Focus on synergies, strategy fit, and metagame considerations.
Keep explanations under 150 words.
Don't use markdown formatting.`

	// Build the user prompt
	var factorsStr string
	if len(score.Factors) > 0 {
		factorsStr = "Key factors: " + strings.Join(score.Factors, ", ")
	}

	prompt := fmt.Sprintf(`Explain why %s is recommended for this deck:

Deck: %s (%s format)
Colors: %s
Archetype: %s
Strategy: %s

Card: %s
Colors: %s
Types: %s
CMC: %.0f
%s

Write a brief, helpful explanation for why this card fits the deck.`,
		card.CardName,
		deck.DeckName,
		deck.Format,
		strings.Join(deck.Colors, ""),
		deck.Archetype,
		deck.Strategy,
		card.CardName,
		strings.Join(card.Colors, ""),
		strings.Join(card.Types, " "),
		card.CMC,
		factorsStr,
	)

	options := &GenerateOptions{
		Temperature: g.config.Temperature,
		NumPredict:  g.config.MaxExplanationLength * 2, // Tokens, not chars
	}

	resp, err := g.ollamaClient.GenerateWithSystem(ctx, systemPrompt, prompt, options)
	if err != nil {
		return "", fmt.Errorf("LLM generation failed: %w", err)
	}

	// Clean up the response
	explanation := strings.TrimSpace(resp.Response)

	// Remove any thinking tags that Qwen might add
	if idx := strings.Index(explanation, "</think>"); idx != -1 {
		explanation = strings.TrimSpace(explanation[idx+8:])
	}

	return explanation, nil
}

// generateTemplateExplanation generates a template-based explanation.
func (g *ExplanationGenerator) generateTemplateExplanation(
	card *CardContext,
	deck *DeckExplanationContext,
	score *ml.CardScore,
) string {
	parts := make([]string, 0)

	// Opening based on score
	if score.Score >= 0.8 {
		parts = append(parts, fmt.Sprintf("%s is an excellent fit for your %s deck.", card.CardName, deck.Archetype))
	} else if score.Score >= 0.6 {
		parts = append(parts, fmt.Sprintf("%s is a good addition to your %s deck.", card.CardName, deck.Archetype))
	} else {
		parts = append(parts, fmt.Sprintf("%s could work in your %s deck.", card.CardName, deck.Archetype))
	}

	// Color synergy
	if len(card.Colors) > 0 && len(deck.Colors) > 0 {
		matchingColors := g.countMatchingColors(card.Colors, deck.Colors)
		if matchingColors == len(card.Colors) {
			parts = append(parts, "The mana cost fits perfectly within your color base.")
		} else if matchingColors > 0 {
			parts = append(parts, "It's mostly on-color for your deck.")
		}
	} else if len(card.Colors) == 0 {
		parts = append(parts, "As a colorless card, it slots easily into any deck.")
	}

	// Type-based explanation
	if g.containsString(card.Types, "Creature") {
		switch deck.Strategy {
		case "aggro":
			if card.CMC <= 3 {
				parts = append(parts, "This low-cost creature supports your aggressive game plan.")
			}
		case "control":
			if card.CMC >= 4 {
				parts = append(parts, "This can serve as a finisher for your control strategy.")
			}
		}
	}

	if g.containsString(card.Types, "Instant") {
		parts = append(parts, "Having instant-speed interaction gives you flexibility.")
	}

	// Factors from ML scoring
	for _, factor := range score.Factors {
		if strings.Contains(strings.ToLower(factor), "synergy") {
			parts = append(parts, "It has strong synergy with other cards in your deck.")
			break
		}
		if strings.Contains(strings.ToLower(factor), "meta") || strings.Contains(strings.ToLower(factor), "tournament") {
			parts = append(parts, "It's proven effective in competitive play.")
			break
		}
		if strings.Contains(strings.ToLower(factor), "style") || strings.Contains(strings.ToLower(factor), "preference") {
			parts = append(parts, "It matches your personal play style preferences.")
			break
		}
	}

	// Meta score explanation
	if score.MetaScore >= 0.7 {
		parts = append(parts, "This card is popular in the current metagame.")
	}

	return strings.Join(parts, " ")
}

// extractSummary extracts a brief summary from the full explanation.
func (g *ExplanationGenerator) extractSummary(explanation string) string {
	// Take first sentence
	if idx := strings.Index(explanation, "."); idx != -1 && idx < 100 {
		return explanation[:idx+1]
	}
	// Or first 80 chars
	if len(explanation) > 80 {
		return explanation[:80] + "..."
	}
	return explanation
}

// makeCacheKey creates a cache key for an explanation.
func (g *ExplanationGenerator) makeCacheKey(cardID int, archetype, format string) string {
	return fmt.Sprintf("%d:%s:%s", cardID, archetype, format)
}

// getFromCache retrieves an explanation from cache.
func (g *ExplanationGenerator) getFromCache(key string) *CachedExplanation {
	g.cacheMu.RLock()
	defer g.cacheMu.RUnlock()

	cached, exists := g.cache[key]
	if !exists {
		return nil
	}

	if time.Since(cached.CreatedAt) > g.config.CacheTTL {
		return nil
	}

	return cached
}

// setCache stores an explanation in cache.
func (g *ExplanationGenerator) setCache(key, explanation string) {
	g.cacheMu.Lock()
	defer g.cacheMu.Unlock()

	g.cache[key] = &CachedExplanation{
		Text:      explanation,
		CreatedAt: time.Now(),
	}
}

// ClearCache clears the explanation cache.
func (g *ExplanationGenerator) ClearCache() {
	g.cacheMu.Lock()
	defer g.cacheMu.Unlock()

	g.cache = make(map[string]*CachedExplanation)
}

// countMatchingColors counts how many colors match between two color lists.
func (g *ExplanationGenerator) countMatchingColors(colors1, colors2 []string) int {
	count := 0
	for _, c1 := range colors1 {
		for _, c2 := range colors2 {
			if c1 == c2 {
				count++
				break
			}
		}
	}
	return count
}

// containsString checks if a slice contains a string.
func (g *ExplanationGenerator) containsString(slice []string, s string) bool {
	for _, item := range slice {
		if strings.EqualFold(item, s) {
			return true
		}
	}
	return false
}

// GetConfig returns the current configuration.
func (g *ExplanationGenerator) GetConfig() *ExplanationConfig {
	return g.config
}

// UpdateConfig updates the configuration.
func (g *ExplanationGenerator) UpdateConfig(config *ExplanationConfig) {
	g.config = config
	g.ClearCache()
}

// IsLLMAvailable returns whether LLM explanations are available.
func (g *ExplanationGenerator) IsLLMAvailable() bool {
	if g.ollamaClient == nil {
		return false
	}
	return g.ollamaClient.IsAvailable()
}
