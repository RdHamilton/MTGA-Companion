package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/archetype"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// OpponentAnalyzer provides opponent deck analysis capabilities.
type OpponentAnalyzer struct {
	gamePlayRepo repository.GamePlayRepository
	opponentRepo repository.OpponentRepository
	matchRepo    repository.MatchRepository
	cardService  *cards.Service
	classifier   *archetype.Classifier
}

// NewOpponentAnalyzer creates a new opponent analyzer.
func NewOpponentAnalyzer(
	gamePlayRepo repository.GamePlayRepository,
	opponentRepo repository.OpponentRepository,
	matchRepo repository.MatchRepository,
	cardService *cards.Service,
	classifier *archetype.Classifier,
) *OpponentAnalyzer {
	return &OpponentAnalyzer{
		gamePlayRepo: gamePlayRepo,
		opponentRepo: opponentRepo,
		matchRepo:    matchRepo,
		cardService:  cardService,
		classifier:   classifier,
	}
}

// AnalyzeOpponent analyzes the opponent's deck from observed cards in a match.
func (a *OpponentAnalyzer) AnalyzeOpponent(ctx context.Context, matchID string) (*models.OpponentAnalysis, error) {
	// Check if we already have a profile
	existingProfile, err := a.opponentRepo.GetProfileByMatchID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing profile: %w", err)
	}

	// Get observed cards from the match
	observedCards, err := a.gamePlayRepo.GetOpponentCardsByMatch(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get opponent cards: %w", err)
	}

	if len(observedCards) == 0 {
		return &models.OpponentAnalysis{
			ObservedCards: []models.ObservedCard{},
			ExpectedCards: []models.ExpectedCard{},
			StrategicInsights: []models.StrategicInsight{{
				Type:        "info",
				Description: "No opponent cards were observed during this match",
				Priority:    models.InsightPriorityLow,
			}},
		}, nil
	}

	// Convert observed cards to classification format
	cardIDs := make([]int, 0, len(observedCards))
	quantities := make(map[int]int)
	cardIDSet := make(map[int]bool)

	for _, card := range observedCards {
		if !cardIDSet[card.CardID] {
			cardIDs = append(cardIDs, card.CardID)
			cardIDSet[card.CardID] = true
		}
		// Each observed card counts as 1 for classification
		// (we can't know how many copies they have)
		quantities[card.CardID] = 1
	}

	// Classify the observed cards
	classification, err := a.classifier.ClassifyCards(cardIDs, quantities)
	if err != nil {
		return nil, fmt.Errorf("failed to classify opponent cards: %w", err)
	}

	// Get match info for format
	match, err := a.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return nil, fmt.Errorf("failed to get match: %w", err)
	}

	var format *string
	if match != nil && match.Format != "" {
		format = &match.Format
	}

	// Build or update opponent deck profile
	profile := a.buildProfile(matchID, classification, len(observedCards), cardIDs, format)
	if existingProfile != nil {
		profile.ID = existingProfile.ID
		profile.CreatedAt = existingProfile.CreatedAt
	}

	if err := a.opponentRepo.CreateOrUpdateProfile(ctx, profile); err != nil {
		return nil, fmt.Errorf("failed to save opponent profile: %w", err)
	}

	// Convert observed cards to response format
	responseCards := a.convertObservedCards(observedCards, classification)

	// Get expected cards based on archetype
	expectedCards, err := a.getExpectedCards(ctx, classification, cardIDSet, format)
	if err != nil {
		// Non-fatal error, continue without expected cards
		expectedCards = []models.ExpectedCard{}
	}

	// Generate strategic insights
	insights := a.generateInsights(classification, responseCards, expectedCards)

	return &models.OpponentAnalysis{
		Profile:           profile,
		ObservedCards:     responseCards,
		ExpectedCards:     expectedCards,
		StrategicInsights: insights,
	}, nil
}

// buildProfile creates an OpponentDeckProfile from classification results.
func (a *OpponentAnalyzer) buildProfile(matchID string, classification *archetype.ClassificationResult, cardsObserved int, cardIDs []int, format *string) *models.OpponentDeckProfile {
	profile := &models.OpponentDeckProfile{
		MatchID:             matchID,
		ArchetypeConfidence: classification.Confidence,
		ColorIdentity:       classification.ColorIdentity,
		CardsObserved:       cardsObserved,
		EstimatedDeckSize:   60, // Default for constructed
		Format:              format,
	}

	if classification.PrimaryArchetype != "" {
		profile.DetectedArchetype = &classification.PrimaryArchetype
	}

	// Detect deck style from analysis
	if classification.Analysis != nil {
		style := a.detectStyle(classification.Analysis)
		if style != "" {
			profile.DeckStyle = &style
		}
	}

	// Store observed card IDs as JSON
	observedJSON, _ := json.Marshal(cardIDs)
	observedStr := string(observedJSON)
	profile.ObservedCardIDs = &observedStr

	// Store signature cards as JSON
	if len(classification.SignatureCards) > 0 {
		sigJSON, _ := json.Marshal(classification.SignatureCards)
		sigStr := string(sigJSON)
		profile.SignatureCards = &sigStr
	}

	return profile
}

// detectStyle determines the deck style from analysis.
func (a *OpponentAnalyzer) detectStyle(analysis *archetype.DeckAnalysis) string {
	totalNonLand := analysis.CreatureCount + analysis.InstantCount + analysis.SorceryCount +
		analysis.ArtifactCount + analysis.EnchantmentCount + analysis.PlaneswalkerCount

	if totalNonLand == 0 {
		return ""
	}

	creatureRatio := float64(analysis.CreatureCount) / float64(totalNonLand)
	spellRatio := float64(analysis.InstantCount+analysis.SorceryCount) / float64(totalNonLand)

	// Low curve with many creatures = Aggro
	if analysis.AvgCMC < 2.5 && creatureRatio > 0.6 {
		return models.DeckStyleAggro
	}

	// High curve with many spells = Control
	if analysis.AvgCMC > 3.5 && spellRatio > 0.4 {
		return models.DeckStyleControl
	}

	// Balanced = Midrange
	if creatureRatio > 0.4 && analysis.AvgCMC >= 2.5 && analysis.AvgCMC <= 3.5 {
		return models.DeckStyleMidrange
	}

	// Mixed with lower curve = Tempo
	if creatureRatio > 0.3 && spellRatio > 0.3 && analysis.AvgCMC < 3.0 {
		return models.DeckStyleTempo
	}

	return ""
}

// convertObservedCards converts repository models to response models.
func (a *OpponentAnalyzer) convertObservedCards(observed []*models.OpponentCardObserved, classification *archetype.ClassificationResult) []models.ObservedCard {
	signatureSet := make(map[int]bool)
	for _, id := range classification.SignatureCards {
		signatureSet[id] = true
	}

	result := make([]models.ObservedCard, 0, len(observed))
	for _, card := range observed {
		name := ""
		if card.CardName != nil {
			name = *card.CardName
		}

		oc := models.ObservedCard{
			CardID:        card.CardID,
			CardName:      name,
			Zone:          card.ZoneObserved,
			TurnFirstSeen: card.TurnFirstSeen,
			TimesSeen:     card.TimesSeen,
			IsSignature:   signatureSet[card.CardID],
		}

		// Categorize the card
		category := a.categorizeCard(card.CardID)
		if category != "" {
			oc.Category = &category
		}

		result = append(result, oc)
	}

	return result
}

// categorizeCard determines the category of a card.
func (a *OpponentAnalyzer) categorizeCard(cardID int) string {
	cards, err := a.cardService.GetCards([]int{cardID})
	if err != nil || cards[cardID] == nil {
		return ""
	}

	card := cards[cardID]
	oracleText := ""
	if card.OracleText != nil {
		oracleText = strings.ToLower(*card.OracleText)
	}
	typeLine := strings.ToLower(card.TypeLine)

	// Check for removal
	if strings.Contains(oracleText, "destroy target") ||
		strings.Contains(oracleText, "exile target") ||
		strings.Contains(oracleText, "deals") && strings.Contains(oracleText, "damage to") {
		return models.CardCategoryRemoval
	}

	// Check for interaction (counters, bounce)
	if strings.Contains(oracleText, "counter target") ||
		strings.Contains(oracleText, "return target") && strings.Contains(oracleText, "to") {
		return models.CardCategoryInteraction
	}

	// Check for card draw
	if strings.Contains(oracleText, "draw") && strings.Contains(oracleText, "card") {
		return models.CardCategoryCardDraw
	}

	// Check for ramp
	if strings.Contains(oracleText, "search your library for") && strings.Contains(oracleText, "land") {
		return models.CardCategoryRamp
	}

	// Creatures and planeswalkers are threats
	if strings.Contains(typeLine, "creature") || strings.Contains(typeLine, "planeswalker") {
		return models.CardCategoryThreat
	}

	return models.CardCategoryUtility
}

// getExpectedCards returns expected cards based on detected archetype.
func (a *OpponentAnalyzer) getExpectedCards(ctx context.Context, classification *archetype.ClassificationResult, seenCards map[int]bool, format *string) ([]models.ExpectedCard, error) {
	if classification.PrimaryArchetype == "" || format == nil {
		return nil, nil
	}

	expected, err := a.opponentRepo.GetExpectedCards(ctx, classification.PrimaryArchetype, *format)
	if err != nil {
		return nil, err
	}

	result := make([]models.ExpectedCard, 0, len(expected))
	for _, exp := range expected {
		ec := models.ExpectedCard{
			CardID:        exp.CardID,
			CardName:      exp.CardName,
			InclusionRate: exp.InclusionRate,
			AvgCopies:     exp.AvgCopies,
			WasSeen:       seenCards[exp.CardID],
		}
		if exp.Category != nil {
			ec.Category = *exp.Category
		}
		ec.PlayAround = a.getPlayAroundAdvice(exp)
		result = append(result, ec)
	}

	return result, nil
}

// getPlayAroundAdvice generates advice for playing around a card.
func (a *OpponentAnalyzer) getPlayAroundAdvice(card *models.ArchetypeExpectedCard) string {
	if card.Category == nil {
		return ""
	}

	switch *card.Category {
	case models.CardCategoryRemoval:
		return fmt.Sprintf("Hold back threats; %s may remove key creatures", card.CardName)
	case models.CardCategoryInteraction:
		return fmt.Sprintf("Leave mana open for protection against %s", card.CardName)
	case models.CardCategoryWincon:
		return fmt.Sprintf("Prepare answers for %s - key win condition", card.CardName)
	default:
		return ""
	}
}

// generateInsights creates strategic insights from the analysis.
func (a *OpponentAnalyzer) generateInsights(classification *archetype.ClassificationResult, observed []models.ObservedCard, expected []models.ExpectedCard) []models.StrategicInsight {
	insights := make([]models.StrategicInsight, 0)

	// Insight about detected archetype
	if classification.PrimaryArchetype != "" {
		confidence := "uncertain"
		if classification.Confidence >= 0.7 {
			confidence = "likely"
		} else if classification.Confidence >= 0.5 {
			confidence = "possibly"
		}

		insights = append(insights, models.StrategicInsight{
			Type:        "archetype",
			Description: fmt.Sprintf("Opponent is %s playing %s", confidence, classification.PrimaryArchetype),
			Priority:    models.InsightPriorityHigh,
		})
	}

	// Insights about deck style
	if classification.Analysis != nil {
		style := a.detectStyle(classification.Analysis)
		switch style {
		case models.DeckStyleAggro:
			insights = append(insights, models.StrategicInsight{
				Type:        "strategy",
				Description: "Aggressive deck - prioritize early blockers and stabilize life total",
				Priority:    models.InsightPriorityHigh,
			})
		case models.DeckStyleControl:
			insights = append(insights, models.StrategicInsight{
				Type:        "strategy",
				Description: "Control deck - apply early pressure and play around countermagic",
				Priority:    models.InsightPriorityHigh,
			})
		case models.DeckStyleMidrange:
			insights = append(insights, models.StrategicInsight{
				Type:        "strategy",
				Description: "Midrange deck - value-based strategy; card advantage is key",
				Priority:    models.InsightPriorityMedium,
			})
		}
	}

	// Insights about removal seen
	removalCount := 0
	var removalCards []int
	for _, card := range observed {
		if card.Category != nil && *card.Category == models.CardCategoryRemoval {
			removalCount++
			removalCards = append(removalCards, card.CardID)
		}
	}
	if removalCount > 0 {
		insights = append(insights, models.StrategicInsight{
			Type:        models.CardCategoryRemoval,
			Description: fmt.Sprintf("Opponent has shown %d removal spell(s) - be cautious with key threats", removalCount),
			Priority:    models.InsightPriorityMedium,
			Cards:       removalCards,
		})
	}

	// Insights about interaction (counters)
	interactionCount := 0
	var interactionCards []int
	for _, card := range observed {
		if card.Category != nil && *card.Category == models.CardCategoryInteraction {
			interactionCount++
			interactionCards = append(interactionCards, card.CardID)
		}
	}
	if interactionCount > 0 {
		insights = append(insights, models.StrategicInsight{
			Type:        models.CardCategoryInteraction,
			Description: fmt.Sprintf("Opponent has shown %d interaction spell(s) - bait with less important spells", interactionCount),
			Priority:    models.InsightPriorityMedium,
			Cards:       interactionCards,
		})
	}

	// Expected cards not yet seen
	unseenHighPriority := 0
	for _, exp := range expected {
		if !exp.WasSeen && exp.InclusionRate >= 0.7 {
			unseenHighPriority++
		}
	}
	if unseenHighPriority > 0 {
		insights = append(insights, models.StrategicInsight{
			Type:        "expected",
			Description: fmt.Sprintf("%d commonly-played cards not yet seen - check expected cards list", unseenHighPriority),
			Priority:    models.InsightPriorityLow,
		})
	}

	return insights
}

// UpdateMatchupStats updates matchup statistics after a match.
func (a *OpponentAnalyzer) UpdateMatchupStats(ctx context.Context, accountID int, playerArchetype, opponentArchetype, format, result string, durationSeconds *int) error {
	wins := 0
	losses := 0
	if result == "win" {
		wins = 1
	} else {
		losses = 1
	}

	stat := &models.MatchupStatistic{
		AccountID:         accountID,
		PlayerArchetype:   playerArchetype,
		OpponentArchetype: opponentArchetype,
		Format:            format,
		TotalMatches:      1,
		Wins:              wins,
		Losses:            losses,
		AvgGameDuration:   durationSeconds,
	}

	now := stat.UpdatedAt
	if now.IsZero() {
		now = stat.CreatedAt
	}
	stat.LastMatchAt = &now

	return a.opponentRepo.RecordMatchup(ctx, stat)
}

// GetMatchupSummary retrieves matchup statistics for an account.
func (a *OpponentAnalyzer) GetMatchupSummary(ctx context.Context, accountID int, format *string) ([]*models.MatchupStatistic, error) {
	return a.opponentRepo.ListMatchupStats(ctx, accountID, format)
}

// GetOpponentHistory retrieves opponent history summary for an account.
func (a *OpponentAnalyzer) GetOpponentHistory(ctx context.Context, accountID int, format *string) (*models.OpponentHistorySummary, error) {
	return a.opponentRepo.GetOpponentHistorySummary(ctx, accountID, format)
}
