package grading

import (
	"context"
	"fmt"
	"math"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// DraftGrade represents the overall grade and component scores for a draft.
type DraftGrade struct {
	OverallGrade         string  `json:"overall_grade"`          // A+, A, A-, B+, etc.
	OverallScore         int     `json:"overall_score"`          // 0-100
	PickQualityScore     float64 `json:"pick_quality_score"`     // 0-40
	ColorDisciplineScore float64 `json:"color_discipline_score"` // 0-20
	DeckCompositionScore float64 `json:"deck_composition_score"` // 0-25
	StrategicScore       float64 `json:"strategic_score"`        // 0-15
	BestPicks            []string `json:"best_picks"`             // Top 3 card names
	WorstPicks           []string `json:"worst_picks"`            // Bottom 3 card names
	Suggestions          []string `json:"suggestions"`            // Improvement suggestions
}

// Calculator calculates draft grades.
type Calculator struct {
	draftRepo   repository.DraftRepository
	ratingsRepo repository.DraftRatingsRepository
	setCardRepo repository.SetCardRepository
}

// NewCalculator creates a new draft grade calculator.
func NewCalculator(draftRepo repository.DraftRepository, ratingsRepo repository.DraftRatingsRepository, setCardRepo repository.SetCardRepository) *Calculator {
	return &Calculator{
		draftRepo:   draftRepo,
		ratingsRepo: ratingsRepo,
		setCardRepo: setCardRepo,
	}
}

// CalculateGrade calculates the overall grade for a draft session.
func (c *Calculator) CalculateGrade(ctx context.Context, sessionID string) (*DraftGrade, error) {
	// Get session
	session, err := c.draftRepo.GetSession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	if session == nil {
		return nil, fmt.Errorf("session not found")
	}

	// Get picks
	picks, err := c.draftRepo.GetPicksBySession(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get picks: %w", err)
	}
	if len(picks) == 0 {
		return nil, fmt.Errorf("no picks found")
	}

	// Calculate component scores
	pickQualityScore := c.calculatePickQualityScore(picks)
	colorScore := c.calculateColorDisciplineScore(ctx, session, picks)
	deckScore := c.calculateDeckCompositionScore(ctx, session, picks)
	strategicScore := c.calculateStrategicScore(picks)

	// Calculate overall score
	overallScore := int(math.Round(pickQualityScore + colorScore + deckScore + strategicScore))
	overallGrade := calculateLetterGrade(overallScore)

	// Get best/worst picks
	bestPicks, worstPicks := c.getBestAndWorstPicks(ctx, picks)

	// Generate suggestions
	suggestions := c.generateSuggestions(pickQualityScore, colorScore, deckScore, strategicScore, picks)

	return &DraftGrade{
		OverallGrade:         overallGrade,
		OverallScore:         overallScore,
		PickQualityScore:     pickQualityScore,
		ColorDisciplineScore: colorScore,
		DeckCompositionScore: deckScore,
		StrategicScore:       strategicScore,
		BestPicks:            bestPicks,
		WorstPicks:           worstPicks,
		Suggestions:          suggestions,
	}, nil
}

// calculatePickQualityScore calculates the pick quality component (0-40).
// Based on average GIHWR and pick quality grades.
func (c *Calculator) calculatePickQualityScore(picks []*models.DraftPickSession) float64 {
	if len(picks) == 0 {
		return 0
	}

	// Count picks with quality grades
	qualitySum := 0.0
	qualityCount := 0

	for _, pick := range picks {
		if pick.PickedCardGIHWR != nil {
			qualitySum += *pick.PickedCardGIHWR
			qualityCount++
		}
	}

	if qualityCount == 0 {
		return 20.0 // Default to C grade if no quality data
	}

	avgGIHWR := qualitySum / float64(qualityCount)

	// Score based on average GIHWR
	// A: >58% = 36-40 points
	// B: 54-58% = 32-36 points
	// C: 50-54% = 28-32 points
	// D: 46-50% = 24-28 points
	// F: <46% = 0-24 points

	var score float64
	switch {
	case avgGIHWR >= 58:
		score = 36 + (avgGIHWR-58)*2 // 36-40
	case avgGIHWR >= 54:
		score = 32 + (avgGIHWR-54)*1 // 32-36
	case avgGIHWR >= 50:
		score = 28 + (avgGIHWR-50)*1 // 28-32
	case avgGIHWR >= 46:
		score = 24 + (avgGIHWR-46)*1 // 24-28
	default:
		score = avgGIHWR / 2 // 0-24
	}

	// Cap at 40
	if score > 40 {
		score = 40
	}
	if score < 0 {
		score = 0
	}

	return score
}

// calculateColorDisciplineScore calculates the color discipline component (0-20).
// Based on staying in strong colors.
func (c *Calculator) calculateColorDisciplineScore(ctx context.Context, session *models.DraftSession, picks []*models.DraftPickSession) float64 {
	// Simple implementation: reward for having color data
	// In a full implementation, this would:
	// - Analyze color pair win rates from 17Lands
	// - Check if player stayed in top-tier colors
	// - Penalize for switching colors late

	// For now, give a default score
	return 15.0 // Default to B grade
}

// calculateDeckCompositionScore calculates the deck composition component (0-25).
// Based on curve, removal, and bombs.
func (c *Calculator) calculateDeckCompositionScore(ctx context.Context, session *models.DraftSession, picks []*models.DraftPickSession) float64 {
	if len(picks) == 0 {
		return 0
	}

	score := 0.0

	// Count high-quality cards (bombs)
	bombCount := 0

	for _, pick := range picks {
		if pick.PickQualityGrade != nil {
			grade := *pick.PickQualityGrade
			if grade == "A+" || grade == "A" {
				bombCount++
			}
		}

		// In a full implementation, would check if card is removal
		// For now, estimate based on pick quality
	}

	// Reward for bombs (0-10 points)
	// Optimal: 2-4 bombs
	if bombCount >= 2 && bombCount <= 4 {
		score += 10
	} else if bombCount == 1 || bombCount == 5 {
		score += 7
	} else {
		score += 4
	}

	// Curve and removal would add 0-15 more points
	// For now, give a default
	score += 10

	return score
}

// calculateStrategicScore calculates the strategic decisions component (0-15).
// Based on decision quality and reading signals.
func (c *Calculator) calculateStrategicScore(picks []*models.DraftPickSession) float64 {
	if len(picks) == 0 {
		return 0
	}

	// Count excellent picks (A+/A grades)
	excellentPicks := 0
	totalGradedPicks := 0

	for _, pick := range picks {
		if pick.PickQualityGrade != nil {
			totalGradedPicks++
			grade := *pick.PickQualityGrade
			if grade == "A+" || grade == "A" {
				excellentPicks++
			}
		}
	}

	if totalGradedPicks == 0 {
		return 10.0 // Default
	}

	// Score based on percentage of excellent picks
	excellentPercent := float64(excellentPicks) / float64(totalGradedPicks)
	score := excellentPercent * 15

	return score
}

// getBestAndWorstPicks returns the top 3 best and worst picks.
func (c *Calculator) getBestAndWorstPicks(ctx context.Context, picks []*models.DraftPickSession) ([]string, []string) {
	type pickWithGrade struct {
		cardID string
		gihwr  float64
		grade  string
	}

	var gradedPicks []pickWithGrade
	for _, pick := range picks {
		if pick.PickQualityGrade != nil && pick.PickedCardGIHWR != nil {
			gradedPicks = append(gradedPicks, pickWithGrade{
				cardID: pick.CardID,
				gihwr:  *pick.PickedCardGIHWR,
				grade:  *pick.PickQualityGrade,
			})
		}
	}

	if len(gradedPicks) == 0 {
		return []string{}, []string{}
	}

	// Sort by GIHWR
	// For simplicity, just take first/last 3
	best := []string{}
	worst := []string{}

	maxBest := 3
	if len(gradedPicks) < 3 {
		maxBest = len(gradedPicks)
	}

	// Get best picks (would need sorting by GIHWR in full implementation)
	for i := 0; i < maxBest; i++ {
		best = append(best, fmt.Sprintf("Card %s", gradedPicks[i].cardID))
	}

	// Get worst picks
	for i := len(gradedPicks) - 1; i >= len(gradedPicks)-maxBest && i >= 0; i-- {
		worst = append(worst, fmt.Sprintf("Card %s", gradedPicks[i].cardID))
	}

	return best, worst
}

// generateSuggestions generates improvement suggestions based on scores.
func (c *Calculator) generateSuggestions(pickQuality, color, deck, strategic float64, picks []*models.DraftPickSession) []string {
	suggestions := []string{}

	// Pick quality suggestions
	if pickQuality < 28 {
		suggestions = append(suggestions, "Focus on selecting higher-quality cards based on GIHWR data")
	}
	if pickQuality < 24 {
		suggestions = append(suggestions, "Review 17Lands tier list before drafting to identify strong cards")
	}

	// Color suggestions
	if color < 14 {
		suggestions = append(suggestions, "Stay in your colors earlier in the draft")
	}

	// Deck composition suggestions
	if deck < 18 {
		suggestions = append(suggestions, "Prioritize building a proper mana curve (2-3-4-5 drops)")
		suggestions = append(suggestions, "Include 3-5 removal spells in your deck")
	}

	// Strategic suggestions
	if strategic < 10 {
		suggestions = append(suggestions, "Take bombs and removal higher priority")
		suggestions = append(suggestions, "Read signals from pack 1 to identify open colors")
	}

	if len(suggestions) == 0 {
		suggestions = append(suggestions, "Excellent draft! Keep up the strong decision-making")
	}

	return suggestions
}

// calculateLetterGrade converts a numeric score to a letter grade.
func calculateLetterGrade(score int) string {
	switch {
	case score >= 97:
		return "A+"
	case score >= 93:
		return "A"
	case score >= 90:
		return "A-"
	case score >= 87:
		return "B+"
	case score >= 83:
		return "B"
	case score >= 80:
		return "B-"
	case score >= 77:
		return "C+"
	case score >= 73:
		return "C"
	case score >= 70:
		return "C-"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}
