package insights

import (
	"context"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

// InsightsStrategy defines the interface for format-specific draft insights analysis.
// Different draft formats (Premier, Quick, Traditional) may require different analysis strategies
// due to differences in player behavior, meta-game patterns, and card value assessments.
type InsightsStrategy interface {
	// AnalyzeColorRankings ranks colors by win rate and popularity.
	// Different formats may weight these factors differently based on format characteristics.
	AnalyzeColorRankings(colorRatings []seventeenlands.ColorRating) []ColorPowerRank

	// FindTopBombs identifies the best rare/mythic cards for the format.
	// Premier vs Quick drafts may have different bomb valuations due to opponent behavior.
	FindTopBombs(cardRatings []seventeenlands.CardRating) []TopCard

	// FindTopRemoval identifies the best removal spells for the format.
	// Removal value can vary based on whether opponents are bots or humans.
	FindTopRemoval(cardRatings []seventeenlands.CardRating) []TopCard

	// FindTopCreatures identifies the best creatures for the format.
	// Creature evaluation may differ based on format speed and opponent tendencies.
	FindTopCreatures(cardRatings []seventeenlands.CardRating) []TopCard

	// FindTopCommons identifies the best common cards for the format.
	// Commons are the backbone of drafts, but their value can vary by format.
	FindTopCommons(cardRatings []seventeenlands.CardRating) []TopCard

	// AnalyzeFormatSpeed determines if the format is fast, medium, or slow.
	// Bot behavior in Quick Draft may lead to different format speeds than human drafts.
	AnalyzeFormatSpeed(cardRatings []seventeenlands.CardRating) FormatSpeed

	// AnalyzeColors provides insights about color depth and overdrafted colors.
	// Quick Draft bots may overdraft certain colors differently than humans.
	AnalyzeColors(cardRatings []seventeenlands.CardRating, colorRatings []seventeenlands.ColorRating) *ColorAnalysis

	// FilterCardsByColor filters cards based on color combination for archetype analysis.
	// Some formats may require different color filtering logic (e.g., splash considerations).
	FilterCardsByColor(cards []seventeenlands.CardRating, colors string) []seventeenlands.CardRating

	// GetStrategyName returns a human-readable name for this strategy.
	GetStrategyName() string
}

// StrategyFactory creates the appropriate insights strategy based on draft format.
func StrategyFactory(ctx context.Context, draftFormat string) InsightsStrategy {
	switch draftFormat {
	case "PremierDraft":
		return NewPremierDraftStrategy()
	case "QuickDraft":
		return NewQuickDraftStrategy()
	default:
		// Default to Premier Draft strategy for unknown formats
		return NewPremierDraftStrategy()
	}
}
