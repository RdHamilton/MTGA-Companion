package analysis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// SuggestionGenerator creates improvement suggestions from play analysis.
type SuggestionGenerator struct {
	analyzer *PlayAnalyzer
	suggRepo repository.SuggestionRepository
}

// NewSuggestionGenerator creates a new suggestion generator.
func NewSuggestionGenerator(analyzer *PlayAnalyzer, suggRepo repository.SuggestionRepository) *SuggestionGenerator {
	return &SuggestionGenerator{
		analyzer: analyzer,
		suggRepo: suggRepo,
	}
}

// GenerateSuggestions analyzes play patterns and generates improvement suggestions.
// Returns the generated suggestions and stores them in the database.
// minGames is the minimum number of games required for analysis (default: 5).
func (g *SuggestionGenerator) GenerateSuggestions(ctx context.Context, deckID string, minGames int) ([]*models.ImprovementSuggestion, error) {
	if minGames <= 0 {
		minGames = 5
	}

	// Run play analysis
	analysis, err := g.analyzer.AnalyzeDeck(ctx, deckID, minGames)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze deck: %w", err)
	}

	// Check if we have enough games
	if analysis.TotalGames < minGames {
		return nil, fmt.Errorf("insufficient games for analysis: %d/%d required", analysis.TotalGames, minGames)
	}

	// Delete existing active suggestions before generating new ones
	if err := g.suggRepo.DeleteActiveSuggestionsByDeck(ctx, deckID); err != nil {
		return nil, fmt.Errorf("failed to clear existing suggestions: %w", err)
	}

	var suggestions []*models.ImprovementSuggestion

	// Generate curve suggestions
	curveSuggestions := g.generateCurveSuggestions(deckID, analysis)
	suggestions = append(suggestions, curveSuggestions...)

	// Generate mana suggestions
	manaSuggestions := g.generateManaSuggestions(deckID, analysis)
	suggestions = append(suggestions, manaSuggestions...)

	// Generate sequencing suggestions
	sequencingSuggestions := g.generateSequencingSuggestions(deckID, analysis)
	suggestions = append(suggestions, sequencingSuggestions...)

	// Store suggestions in database
	for _, sugg := range suggestions {
		if err := g.suggRepo.CreateSuggestion(ctx, sugg); err != nil {
			return nil, fmt.Errorf("failed to create suggestion: %w", err)
		}
	}

	return suggestions, nil
}

// generateCurveSuggestions creates suggestions related to mana curve.
func (g *SuggestionGenerator) generateCurveSuggestions(deckID string, analysis *AnalysisResult) []*models.ImprovementSuggestion {
	var suggestions []*models.ImprovementSuggestion

	// High curve - late first plays
	if analysis.CurveAnalysis.AvgFirstPlay > 2.5 {
		evidence := map[string]interface{}{
			"avgFirstPlay": analysis.CurveAnalysis.AvgFirstPlay,
			"totalGames":   analysis.TotalGames,
		}
		evidenceJSON, _ := json.Marshal(evidence)
		evidenceStr := string(evidenceJSON)

		suggestions = append(suggestions, &models.ImprovementSuggestion{
			DeckID:         deckID,
			SuggestionType: models.SuggestionTypeCurve,
			Priority:       models.SuggestionPriorityHigh,
			Title:          "Curve May Be Too High",
			Description:    fmt.Sprintf("Your average first play is on turn %.1f. Consider adding more 1-2 mana spells to improve early game presence and avoid falling behind.", analysis.CurveAnalysis.AvgFirstPlay),
			Evidence:       &evidenceStr,
		})
	}

	// Very high curve - consistently late starts
	if analysis.CurveAnalysis.AvgFirstPlay > 3.0 {
		suggestions = append(suggestions, &models.ImprovementSuggestion{
			DeckID:         deckID,
			SuggestionType: models.SuggestionTypeCurve,
			Priority:       models.SuggestionPriorityHigh,
			Title:          "Severe Curve Issues",
			Description:    "You're consistently making your first play on turn 3 or later. This puts you significantly behind aggressive decks. Add cheap interaction or creatures.",
		})
	}

	// Check CMC distribution imbalance
	lowCMCPlays := analysis.CurveAnalysis.CMCDistribution[1] + analysis.CurveAnalysis.CMCDistribution[2]
	midCMCPlays := analysis.CurveAnalysis.CMCDistribution[3] + analysis.CurveAnalysis.CMCDistribution[4]
	highCMCPlays := analysis.CurveAnalysis.CMCDistribution[5] + analysis.CurveAnalysis.CMCDistribution[6]

	totalPlays := lowCMCPlays + midCMCPlays + highCMCPlays
	if totalPlays > 0 {
		lowRatio := float64(lowCMCPlays) / float64(totalPlays)
		if lowRatio < 0.3 && analysis.TotalGames >= 5 {
			suggestions = append(suggestions, &models.ImprovementSuggestion{
				DeckID:         deckID,
				SuggestionType: models.SuggestionTypeCurve,
				Priority:       models.SuggestionPriorityMedium,
				Title:          "Limited Early Game Options",
				Description:    fmt.Sprintf("Only %.0f%% of your plays are 1-2 mana cards. A healthy curve typically has 30-40%% of cards at low CMC.", lowRatio*100),
			})
		}
	}

	return suggestions
}

// generateManaSuggestions creates suggestions related to mana base.
func (g *SuggestionGenerator) generateManaSuggestions(deckID string, analysis *AnalysisResult) []*models.ImprovementSuggestion {
	var suggestions []*models.ImprovementSuggestion

	// Frequent mana screw
	manaScrew := analysis.ManaAnalysis.ManaScrew
	screwRate := float64(manaScrew) / float64(analysis.TotalGames) * 100

	if screwRate > 25 {
		evidence := map[string]interface{}{
			"manaScrew":    manaScrew,
			"totalGames":   analysis.TotalGames,
			"screwRate":    screwRate,
			"avgLandDrops": analysis.ManaAnalysis.AvgLandDrops,
		}
		evidenceJSON, _ := json.Marshal(evidence)
		evidenceStr := string(evidenceJSON)

		priority := models.SuggestionPriorityMedium
		if screwRate > 40 {
			priority = models.SuggestionPriorityHigh
		}

		suggestions = append(suggestions, &models.ImprovementSuggestion{
			DeckID:         deckID,
			SuggestionType: models.SuggestionTypeMana,
			Priority:       priority,
			Title:          "Frequent Mana Screw",
			Description:    fmt.Sprintf("Mana screw occurred in %.0f%% of games (%d of %d). Consider adding 1-2 more lands or reducing high-cost spells.", screwRate, manaScrew, analysis.TotalGames),
			Evidence:       &evidenceStr,
		})
	}

	// Frequent mana flood
	manaFlood := analysis.ManaAnalysis.ManaFlood
	floodRate := float64(manaFlood) / float64(analysis.TotalGames) * 100

	if floodRate > 30 {
		evidence := map[string]interface{}{
			"manaFlood":    manaFlood,
			"totalGames":   analysis.TotalGames,
			"floodRate":    floodRate,
			"avgLandDrops": analysis.ManaAnalysis.AvgLandDrops,
		}
		evidenceJSON, _ := json.Marshal(evidence)
		evidenceStr := string(evidenceJSON)

		suggestions = append(suggestions, &models.ImprovementSuggestion{
			DeckID:         deckID,
			SuggestionType: models.SuggestionTypeMana,
			Priority:       models.SuggestionPriorityMedium,
			Title:          "Frequent Mana Flood",
			Description:    fmt.Sprintf("Mana flood occurred in %.0f%% of games (%d of %d). Consider removing 1-2 lands or adding card draw to utilize excess mana.", floodRate, manaFlood, analysis.TotalGames),
			Evidence:       &evidenceStr,
		})
	}

	// High land drop miss rate
	if analysis.LandDropMissRate > 30 {
		suggestions = append(suggestions, &models.ImprovementSuggestion{
			DeckID:         deckID,
			SuggestionType: models.SuggestionTypeMana,
			Priority:       models.SuggestionPriorityMedium,
			Title:          "Missed Land Drops",
			Description:    fmt.Sprintf("You're missing land drops in %.0f%% of games. Ensure you have enough lands for your curve, or add card filtering.", analysis.LandDropMissRate),
		})
	}

	return suggestions
}

// generateSequencingSuggestions creates suggestions related to play sequencing.
func (g *SuggestionGenerator) generateSequencingSuggestions(deckID string, analysis *AnalysisResult) []*models.ImprovementSuggestion {
	var suggestions []*models.ImprovementSuggestion

	// Check for sequencing issues
	if len(analysis.SequencingIssues) >= 3 {
		suggestions = append(suggestions, &models.ImprovementSuggestion{
			DeckID:         deckID,
			SuggestionType: models.SuggestionTypeSequencing,
			Priority:       models.SuggestionPriorityLow,
			Title:          "Play Sequencing",
			Description:    fmt.Sprintf("Detected %d potential sequencing issues. Consider playing lands before spells to keep mana open for responses.", len(analysis.SequencingIssues)),
		})
	}

	return suggestions
}

// GetDeckSuggestions retrieves existing suggestions for a deck without regenerating.
func (g *SuggestionGenerator) GetDeckSuggestions(ctx context.Context, deckID string, activeOnly bool) ([]*models.ImprovementSuggestion, error) {
	if activeOnly {
		return g.suggRepo.GetActiveSuggestions(ctx, deckID)
	}
	return g.suggRepo.GetSuggestionsByDeck(ctx, deckID)
}

// DismissSuggestion marks a suggestion as dismissed.
func (g *SuggestionGenerator) DismissSuggestion(ctx context.Context, suggestionID int64) error {
	return g.suggRepo.DismissSuggestion(ctx, suggestionID)
}
