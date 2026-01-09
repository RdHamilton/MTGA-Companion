package analysis

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// MLEngine processes match history to build card synergy models and generate suggestions.
type MLEngine struct {
	mlRepo        *repository.MLSuggestionRepository
	matchRepo     repository.MatchRepository
	deckRepo      repository.DeckRepository
	cardRepo      repository.SetCardRepository
	analyzer      *PlayAnalyzer
	minGames      int
	minConfidence float64
}

// NewMLEngine creates a new ML engine.
func NewMLEngine(
	mlRepo *repository.MLSuggestionRepository,
	matchRepo repository.MatchRepository,
	deckRepo repository.DeckRepository,
	cardRepo repository.SetCardRepository,
	analyzer *PlayAnalyzer,
) *MLEngine {
	return &MLEngine{
		mlRepo:        mlRepo,
		matchRepo:     matchRepo,
		deckRepo:      deckRepo,
		cardRepo:      cardRepo,
		analyzer:      analyzer,
		minGames:      5,
		minConfidence: 0.3,
	}
}

// MLSuggestionResult contains a suggestion with additional context.
type MLSuggestionResult struct {
	Suggestion  *models.MLSuggestion        `json:"suggestion"`
	SynergyData []*CardSynergyInfo          `json:"synergyData,omitempty"`
	Reasons     []models.MLSuggestionReason `json:"reasons"`
}

// CardSynergyInfo provides synergy information for a card pair.
type CardSynergyInfo struct {
	CardID          int     `json:"cardId"`
	CardName        string  `json:"cardName"`
	SynergyScore    float64 `json:"synergyScore"`
	WinRateTogether float64 `json:"winRateTogether"`
	GamesTogether   int     `json:"gamesTogether"`
}

// ProcessMatchHistory analyzes match history to build card combination statistics.
// This should be called periodically or after matches are recorded.
func (e *MLEngine) ProcessMatchHistory(ctx context.Context, format string, lookbackDays int) error {
	// Get recent matches
	since := time.Now().AddDate(0, 0, -lookbackDays)
	filter := models.StatsFilter{
		StartDate: &since,
	}

	matches, err := e.matchRepo.GetMatches(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to get matches: %w", err)
	}

	// Process each match to extract card combinations
	for _, match := range matches {
		if match.DeckID == nil || *match.DeckID == "" {
			continue
		}

		// Get deck
		deck, err := e.deckRepo.GetByID(ctx, *match.DeckID)
		if err != nil || deck == nil {
			continue
		}

		// Skip if format doesn't match (if specified)
		if format != "" && deck.Format != format {
			continue
		}

		// Get deck cards
		cardIDs, err := e.getDeckCardIDs(ctx, deck.ID)
		if err != nil || len(cardIDs) < 2 {
			continue
		}

		// Record combinations
		isWin := match.Result == "win"
		if err := e.recordCombinations(ctx, cardIDs, deck.Format, isWin); err != nil {
			// Log but continue processing
			continue
		}
	}

	// Recalculate synergy scores after processing
	return e.mlRepo.CalculateAndUpdateSynergyScores(ctx, e.minGames)
}

// getDeckCardIDs extracts unique card IDs from a deck's mainboard.
func (e *MLEngine) getDeckCardIDs(ctx context.Context, deckID string) ([]int, error) {
	cards, err := e.deckRepo.GetCards(ctx, deckID)
	if err != nil {
		return nil, err
	}

	seen := make(map[int]bool)
	var cardIDs []int
	for _, card := range cards {
		if card.Board == "main" && !seen[card.CardID] {
			cardIDs = append(cardIDs, card.CardID)
			seen[card.CardID] = true
		}
	}

	return cardIDs, nil
}

// recordCombinations records all pairwise card combinations from a game.
func (e *MLEngine) recordCombinations(ctx context.Context, cardIDs []int, format string, isWin bool) error {
	// Sort card IDs for consistent ordering
	sort.Ints(cardIDs)

	// Process all pairs
	for i := 0; i < len(cardIDs)-1; i++ {
		for j := i + 1; j < len(cardIDs); j++ {
			stats := &models.CardCombinationStats{
				CardID1:       cardIDs[i],
				CardID2:       cardIDs[j],
				Format:        format,
				GamesTogether: 1,
			}
			if isWin {
				stats.WinsTogether = 1
			}

			if err := e.mlRepo.UpsertCombinationStats(ctx, stats); err != nil {
				return err
			}
		}
	}

	return nil
}

// GenerateMLSuggestions creates ML-powered suggestions for a deck.
func (e *MLEngine) GenerateMLSuggestions(ctx context.Context, deckID string) ([]*MLSuggestionResult, error) {
	// Get deck
	deck, err := e.deckRepo.GetByID(ctx, deckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deck: %w", err)
	}
	if deck == nil {
		return nil, fmt.Errorf("deck not found: %s", deckID)
	}

	// Get current card IDs in deck
	currentCards, err := e.getDeckCardIDs(ctx, deck.ID)
	if err != nil || len(currentCards) == 0 {
		return nil, nil
	}

	var results []*MLSuggestionResult

	// Generate suggestions for cards that could be added based on synergy
	addSuggestions, err := e.generateAddSuggestions(ctx, deck, currentCards)
	if err == nil {
		results = append(results, addSuggestions...)
	}

	// Generate suggestions for cards that might be underperforming
	removeSuggestions, err := e.generateRemoveSuggestions(ctx, deck, currentCards)
	if err == nil {
		results = append(results, removeSuggestions...)
	}

	// Generate swap suggestions
	swapSuggestions, err := e.generateSwapSuggestions(ctx, deck, currentCards)
	if err == nil {
		results = append(results, swapSuggestions...)
	}

	// Store suggestions in database
	for _, result := range results {
		if err := e.mlRepo.CreateSuggestion(ctx, result.Suggestion); err != nil {
			// Log but continue
			continue
		}
	}

	return results, nil
}

// generateAddSuggestions finds cards with high synergy to existing deck cards.
func (e *MLEngine) generateAddSuggestions(ctx context.Context, deck *models.Deck, currentCards []int) ([]*MLSuggestionResult, error) {
	var results []*MLSuggestionResult
	candidateScores := make(map[int]float64)
	candidateSynergies := make(map[int][]*CardSynergyInfo)

	currentCardSet := make(map[int]bool)
	for _, id := range currentCards {
		currentCardSet[id] = true
	}

	// Find cards with high synergy to multiple cards in the deck
	for _, cardID := range currentCards {
		synergies, err := e.mlRepo.GetTopSynergiesForCard(ctx, cardID, deck.Format, 20)
		if err != nil {
			continue
		}

		for _, syn := range synergies {
			partnerID := repository.GetPairedCardID(syn, cardID)

			// Skip if card is already in deck
			if currentCardSet[partnerID] {
				continue
			}

			// Skip low confidence results
			confidence := repository.CalculateConfidenceScore(syn.GamesTogether)
			if confidence < e.minConfidence {
				continue
			}

			// Accumulate score for this candidate
			candidateScores[partnerID] += syn.SynergyScore

			// Track synergy info
			candidateSynergies[partnerID] = append(candidateSynergies[partnerID], &CardSynergyInfo{
				CardID:          cardID,
				SynergyScore:    syn.SynergyScore,
				WinRateTogether: syn.WinRateTogether(),
				GamesTogether:   syn.GamesTogether,
			})
		}
	}

	// Sort candidates by total synergy score
	type candidateScore struct {
		cardID int
		score  float64
	}
	var sorted []candidateScore
	for id, score := range candidateScores {
		sorted = append(sorted, candidateScore{id, score})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].score > sorted[j].score
	})

	// Take top 3 candidates
	limit := 3
	if len(sorted) < limit {
		limit = len(sorted)
	}

	for i := 0; i < limit; i++ {
		candidate := sorted[i]

		// Get card name
		cardName := fmt.Sprintf("Card #%d", candidate.cardID)
		card, err := e.cardRepo.GetCardByArenaID(ctx, strconv.Itoa(candidate.cardID))
		if err == nil && card != nil {
			cardName = card.Name
		}

		// Build reasons from synergy data
		reasons := e.buildAddReasons(candidateSynergies[candidate.cardID], cardName)

		// Calculate confidence based on number of synergies and their confidence
		synCount := len(candidateSynergies[candidate.cardID])
		avgConfidence := 0.0
		for _, syn := range candidateSynergies[candidate.cardID] {
			avgConfidence += repository.CalculateConfidenceScore(syn.GamesTogether)
		}
		if synCount > 0 {
			avgConfidence /= float64(synCount)
		}

		suggestion, err := repository.GenerateMLSuggestion(
			deck.ID,
			models.MLSuggestionTypeAdd,
			candidate.cardID,
			cardName,
			avgConfidence,
			candidate.score*100, // Convert to percentage
			reasons,
		)
		if err != nil {
			continue
		}

		results = append(results, &MLSuggestionResult{
			Suggestion:  suggestion,
			SynergyData: candidateSynergies[candidate.cardID],
			Reasons:     reasons,
		})
	}

	return results, nil
}

// generateRemoveSuggestions finds cards with negative synergy in the deck.
func (e *MLEngine) generateRemoveSuggestions(ctx context.Context, deck *models.Deck, currentCards []int) ([]*MLSuggestionResult, error) {
	var results []*MLSuggestionResult
	cardSynergySums := make(map[int]float64)
	cardSynergyCount := make(map[int]int)

	// Calculate average synergy for each card in the deck
	for i := 0; i < len(currentCards)-1; i++ {
		for j := i + 1; j < len(currentCards); j++ {
			stats, err := e.mlRepo.GetCombinationStats(ctx, currentCards[i], currentCards[j], deck.Format)
			if err != nil || stats == nil {
				continue
			}

			synergy := repository.CalculateSynergyScore(stats)
			cardSynergySums[currentCards[i]] += synergy
			cardSynergySums[currentCards[j]] += synergy
			cardSynergyCount[currentCards[i]]++
			cardSynergyCount[currentCards[j]]++
		}
	}

	// Find cards with negative average synergy
	type cardScore struct {
		cardID   int
		avgSyn   float64
		pairings int
	}
	var negatives []cardScore
	for id, sum := range cardSynergySums {
		count := cardSynergyCount[id]
		if count >= 3 { // Need at least 3 pairings
			avg := sum / float64(count)
			if avg < -0.02 { // Threshold for suggesting removal
				negatives = append(negatives, cardScore{id, avg, count})
			}
		}
	}

	// Sort by worst synergy
	sort.Slice(negatives, func(i, j int) bool {
		return negatives[i].avgSyn < negatives[j].avgSyn
	})

	// Take top 2 candidates for removal
	limit := 2
	if len(negatives) < limit {
		limit = len(negatives)
	}

	for i := 0; i < limit; i++ {
		candidate := negatives[i]

		cardName := fmt.Sprintf("Card #%d", candidate.cardID)
		card, err := e.cardRepo.GetCardByArenaID(ctx, strconv.Itoa(candidate.cardID))
		if err == nil && card != nil {
			cardName = card.Name
		}

		reasons := []models.MLSuggestionReason{
			{
				Type:        "synergy",
				Description: fmt.Sprintf("%s has negative synergy with %d other cards in your deck", cardName, candidate.pairings),
				Impact:      candidate.avgSyn,
				Confidence:  repository.CalculateConfidenceScore(candidate.pairings * 5),
			},
		}

		suggestion, err := repository.GenerateMLSuggestion(
			deck.ID,
			models.MLSuggestionTypeRemove,
			candidate.cardID,
			cardName,
			repository.CalculateConfidenceScore(candidate.pairings*5),
			candidate.avgSyn*100,
			reasons,
		)
		if err != nil {
			continue
		}

		results = append(results, &MLSuggestionResult{
			Suggestion: suggestion,
			Reasons:    reasons,
		})
	}

	return results, nil
}

// generateSwapSuggestions finds cards that could replace underperforming cards.
func (e *MLEngine) generateSwapSuggestions(ctx context.Context, deck *models.Deck, currentCards []int) ([]*MLSuggestionResult, error) {
	var results []*MLSuggestionResult

	// Find underperforming cards
	cardSynergySums := make(map[int]float64)
	cardSynergyCount := make(map[int]int)

	for i := 0; i < len(currentCards)-1; i++ {
		for j := i + 1; j < len(currentCards); j++ {
			stats, err := e.mlRepo.GetCombinationStats(ctx, currentCards[i], currentCards[j], deck.Format)
			if err != nil || stats == nil {
				continue
			}

			synergy := repository.CalculateSynergyScore(stats)
			cardSynergySums[currentCards[i]] += synergy
			cardSynergySums[currentCards[j]] += synergy
			cardSynergyCount[currentCards[i]]++
			cardSynergyCount[currentCards[j]]++
		}
	}

	// Find worst performing card
	var worstCard int
	worstAvg := 0.0
	for id, sum := range cardSynergySums {
		count := cardSynergyCount[id]
		if count >= 3 {
			avg := sum / float64(count)
			if avg < worstAvg || worstCard == 0 {
				worstCard = id
				worstAvg = avg
			}
		}
	}

	if worstCard == 0 || worstAvg >= -0.01 {
		return results, nil
	}

	// Find best replacement candidate
	currentCardSet := make(map[int]bool)
	for _, id := range currentCards {
		currentCardSet[id] = true
	}

	candidateScores := make(map[int]float64)
	for _, cardID := range currentCards {
		if cardID == worstCard {
			continue
		}

		synergies, err := e.mlRepo.GetTopSynergiesForCard(ctx, cardID, deck.Format, 10)
		if err != nil {
			continue
		}

		for _, syn := range synergies {
			partnerID := repository.GetPairedCardID(syn, cardID)
			if !currentCardSet[partnerID] && syn.SynergyScore > 0 {
				candidateScores[partnerID] += syn.SynergyScore
			}
		}
	}

	// Find best candidate
	var bestCandidate int
	bestScore := 0.0
	for id, score := range candidateScores {
		if score > bestScore {
			bestCandidate = id
			bestScore = score
		}
	}

	if bestCandidate == 0 {
		return results, nil
	}

	// Get card names
	worstName := fmt.Sprintf("Card #%d", worstCard)
	if card, err := e.cardRepo.GetCardByArenaID(ctx, strconv.Itoa(worstCard)); err == nil && card != nil {
		worstName = card.Name
	}

	bestName := fmt.Sprintf("Card #%d", bestCandidate)
	if card, err := e.cardRepo.GetCardByArenaID(ctx, strconv.Itoa(bestCandidate)); err == nil && card != nil {
		bestName = card.Name
	}

	reasons := []models.MLSuggestionReason{
		{
			Type:        "synergy",
			Description: fmt.Sprintf("%s has negative synergy (%.1f%%) with your deck", worstName, worstAvg*100),
			Impact:      worstAvg,
			Confidence:  0.6,
		},
		{
			Type:        "synergy",
			Description: fmt.Sprintf("%s has strong synergy (%.1f%%) with your other cards", bestName, bestScore*100),
			Impact:      bestScore,
			Confidence:  0.6,
		},
	}

	suggestion := &models.MLSuggestion{
		DeckID:                deck.ID,
		SuggestionType:        models.MLSuggestionTypeSwap,
		CardID:                worstCard,
		CardName:              worstName,
		SwapForCardID:         bestCandidate,
		SwapForCardName:       bestName,
		Confidence:            0.5,
		ExpectedWinRateChange: (bestScore - worstAvg) * 100,
		Title:                 fmt.Sprintf("Swap %s for %s", worstName, bestName),
		Description:           fmt.Sprintf("Replace underperforming %s with %s which has better synergy with your deck", worstName, bestName),
		CreatedAt:             time.Now(),
	}
	if err := suggestion.SetReasons(reasons); err != nil {
		return results, nil
	}

	results = append(results, &MLSuggestionResult{
		Suggestion: suggestion,
		Reasons:    reasons,
	})

	return results, nil
}

// buildAddReasons creates suggestion reasons from synergy data.
func (e *MLEngine) buildAddReasons(synergies []*CardSynergyInfo, cardName string) []models.MLSuggestionReason {
	var reasons []models.MLSuggestionReason

	if len(synergies) == 0 {
		return reasons
	}

	// Sort by synergy score
	sort.Slice(synergies, func(i, j int) bool {
		return synergies[i].SynergyScore > synergies[j].SynergyScore
	})

	// Add top synergy reason
	top := synergies[0]
	topCardName := top.CardName
	if topCardName == "" {
		topCardName = fmt.Sprintf("Card #%d", top.CardID)
	}
	reasons = append(reasons, models.MLSuggestionReason{
		Type:        "synergy",
		Description: fmt.Sprintf("%s has %.1f%% higher win rate when paired with %s", cardName, top.WinRateTogether*100, topCardName),
		Impact:      top.SynergyScore,
		Confidence:  repository.CalculateConfidenceScore(top.GamesTogether),
	})

	// Add count-based reason if multiple synergies
	if len(synergies) >= 3 {
		reasons = append(reasons, models.MLSuggestionReason{
			Type:        "performance",
			Description: fmt.Sprintf("Has positive synergy with %d cards already in your deck", len(synergies)),
			Impact:      0.5,
			Confidence:  0.7,
		})
	}

	return reasons
}

// GetSynergyReport returns a synergy analysis report for a deck.
func (e *MLEngine) GetSynergyReport(ctx context.Context, deckID string) (*SynergyReport, error) {
	deck, err := e.deckRepo.GetByID(ctx, deckID)
	if err != nil {
		return nil, err
	}
	if deck == nil {
		return nil, fmt.Errorf("deck not found: %s", deckID)
	}

	currentCards, err := e.getDeckCardIDs(ctx, deck.ID)
	if err != nil {
		return nil, err
	}

	report := &SynergyReport{
		DeckID:    deckID,
		CardCount: len(currentCards),
		Synergies: make([]CardPairSynergy, 0),
	}

	// Calculate synergies for all pairs
	for i := 0; i < len(currentCards)-1; i++ {
		for j := i + 1; j < len(currentCards); j++ {
			stats, err := e.mlRepo.GetCombinationStats(ctx, currentCards[i], currentCards[j], deck.Format)
			if err != nil || stats == nil {
				continue
			}

			synergy := repository.CalculateSynergyScore(stats)
			if stats.GamesTogether >= e.minGames {
				report.Synergies = append(report.Synergies, CardPairSynergy{
					Card1ID:       stats.CardID1,
					Card2ID:       stats.CardID2,
					SynergyScore:  synergy,
					GamesTogether: stats.GamesTogether,
					WinRate:       stats.WinRateTogether(),
				})
				report.TotalPairs++
				report.AvgSynergyScore += synergy
			}
		}
	}

	if report.TotalPairs > 0 {
		report.AvgSynergyScore /= float64(report.TotalPairs)
	}

	// Sort by synergy score
	sort.Slice(report.Synergies, func(i, j int) bool {
		return report.Synergies[i].SynergyScore > report.Synergies[j].SynergyScore
	})

	// Limit to top/bottom 10
	if len(report.Synergies) > 20 {
		top10 := report.Synergies[:10]
		bottom10 := report.Synergies[len(report.Synergies)-10:]
		report.Synergies = append(top10, bottom10...)
	}

	return report, nil
}

// SynergyReport provides a summary of deck synergies.
type SynergyReport struct {
	DeckID          string            `json:"deckId"`
	CardCount       int               `json:"cardCount"`
	TotalPairs      int               `json:"totalPairs"`
	AvgSynergyScore float64           `json:"avgSynergyScore"`
	Synergies       []CardPairSynergy `json:"synergies"`
}

// CardPairSynergy represents synergy between two cards.
type CardPairSynergy struct {
	Card1ID       int     `json:"card1Id"`
	Card1Name     string  `json:"card1Name,omitempty"`
	Card2ID       int     `json:"card2Id"`
	Card2Name     string  `json:"card2Name,omitempty"`
	SynergyScore  float64 `json:"synergyScore"`
	GamesTogether int     `json:"gamesTogether"`
	WinRate       float64 `json:"winRate"`
}

// UpdateUserPlayPatterns analyzes user matches to build play pattern profile.
func (e *MLEngine) UpdateUserPlayPatterns(ctx context.Context, accountID string) error {
	// Get all matches for the user
	filter := models.StatsFilter{}
	matches, err := e.matchRepo.GetMatches(ctx, filter)
	if err != nil {
		return err
	}

	if len(matches) < 10 {
		return nil // Not enough data
	}

	patterns := &models.UserPlayPatterns{
		AccountID:    accountID,
		TotalMatches: len(matches),
	}

	// Analyze format preferences
	formatCounts := make(map[string]int)
	colorCounts := make(map[string]int)

	for _, match := range matches {
		if match.DeckID == nil || *match.DeckID == "" {
			continue
		}

		deck, err := e.deckRepo.GetByID(ctx, *match.DeckID)
		if err != nil || deck == nil {
			continue
		}

		// Count format
		formatCounts[deck.Format]++

		// Count colors from color identity
		if deck.ColorIdentity != nil {
			for _, c := range *deck.ColorIdentity {
				colorCounts[string(c)]++
			}
		}
	}

	// Determine preferred archetype based on format
	// Aggro-leaning formats vs control-leaning formats
	total := float64(len(matches))
	draftCount := float64(formatCounts["Draft"] + formatCounts["Limited"])
	constructedCount := float64(formatCounts["Standard"] + formatCounts["Historic"] + formatCounts["Explorer"])

	// Use format distribution as a proxy for archetype preference
	// Draft players tend to be more adaptive, constructed players more specialized
	if draftCount > constructedCount {
		patterns.MidrangeAffinity = 0.4
		patterns.AggroAffinity = 0.3
		patterns.ControlAffinity = 0.2
		patterns.ComboAffinity = 0.1
		patterns.PreferredArchetype = "Midrange"
	} else {
		patterns.AggroAffinity = 0.3
		patterns.MidrangeAffinity = 0.3
		patterns.ControlAffinity = 0.3
		patterns.ComboAffinity = 0.1
		patterns.PreferredArchetype = "Balanced"
	}

	// Set color preferences
	colorPrefs := make(map[string]float64)
	for color, count := range colorCounts {
		colorPrefs[color] = float64(count) / total
	}
	if err := patterns.SetColorPreferences(colorPrefs); err != nil {
		return err
	}

	// Get unique decks
	deckSet := make(map[string]bool)
	for _, match := range matches {
		if match.DeckID != nil && *match.DeckID != "" {
			deckSet[*match.DeckID] = true
		}
	}
	patterns.TotalDecks = len(deckSet)

	return e.mlRepo.UpsertUserPlayPatterns(ctx, patterns)
}
