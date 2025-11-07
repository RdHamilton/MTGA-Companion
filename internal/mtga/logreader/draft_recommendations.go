package logreader

// DraftRecommendation represents a recommended pick for a draft.
type DraftRecommendation struct {
	CardID    int
	Priority  int    // 1-5, where 5 is highest priority
	Reason    string // Explanation for the recommendation
	Archetype string // Suggested archetype
}

// DraftRecommendations represents recommendations for a draft pick.
type DraftRecommendations struct {
	TopPicks     []DraftRecommendation // Top 3 recommended picks
	Alternatives []DraftRecommendation // Alternative picks for different archetypes
}

// GetDraftRecommendations provides basic draft recommendations based on pack contents and previous picks.
// This is a basic implementation that can be enhanced with card ratings and archetype analysis.
func GetDraftRecommendations(packCards []int, previousPicks []DraftPick) DraftRecommendations {
	recommendations := DraftRecommendations{
		TopPicks:     []DraftRecommendation{},
		Alternatives: []DraftRecommendation{},
	}

	if len(packCards) == 0 {
		return recommendations
	}

	// Basic recommendation: suggest first card in pack
	// In the future, this can be enhanced with:
	// - Card ratings (17lands data, tier lists)
	// - Archetype analysis
	// - Color commitment analysis
	// - Mana curve considerations

	for i, cardID := range packCards {
		if i >= 3 {
			break // Limit to top 3
		}

		priority := 5 - i // Higher priority for earlier cards (basic heuristic)
		reason := "Basic recommendation"
		if i == 0 {
			reason = "First pick in pack"
		}

		recommendation := DraftRecommendation{
			CardID:   cardID,
			Priority: priority,
			Reason:   reason,
		}

		if i < 3 {
			recommendations.TopPicks = append(recommendations.TopPicks, recommendation)
		} else {
			recommendations.Alternatives = append(recommendations.Alternatives, recommendation)
		}
	}

	return recommendations
}

