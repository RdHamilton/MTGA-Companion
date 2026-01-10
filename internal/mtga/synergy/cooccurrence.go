package synergy

import (
	"context"
	"fmt"
	"log"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// CooccurrenceAnalyzer builds and queries card co-occurrence data.
type CooccurrenceAnalyzer struct {
	repo        repository.CooccurrenceRepository
	cardNameMap CardNameMapper
}

// CardNameMapper maps card names to Arena IDs.
// This interface allows the analyzer to work with different card databases.
type CardNameMapper interface {
	// GetArenaIDByName returns the Arena ID for a card name.
	GetArenaIDByName(ctx context.Context, name string) (int, error)
}

// NewCooccurrenceAnalyzer creates a new co-occurrence analyzer.
func NewCooccurrenceAnalyzer(repo repository.CooccurrenceRepository, cardNameMap CardNameMapper) *CooccurrenceAnalyzer {
	return &CooccurrenceAnalyzer{
		repo:        repo,
		cardNameMap: cardNameMap,
	}
}

// AnalysisResult contains the results of a co-occurrence analysis run.
type AnalysisResult struct {
	SourceName    string
	Format        string
	DecksAnalyzed int
	PairsCreated  int
	CardsTracked  int
}

// AnalyzeDecks analyzes a set of decks and updates co-occurrence data.
func (a *CooccurrenceAnalyzer) AnalyzeDecks(ctx context.Context, source DeckSource, format string, limit int) (*AnalysisResult, error) {
	log.Printf("[CooccurrenceAnalyzer] Fetching decks from %s for format %s...", source.SourceName(), format)

	decks, err := source.FetchDecks(ctx, format, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch decks: %w", err)
	}

	log.Printf("[CooccurrenceAnalyzer] Fetched %d decks, analyzing...", len(decks))

	result := &AnalysisResult{
		SourceName:    source.SourceName(),
		Format:        format,
		DecksAnalyzed: len(decks),
	}

	// Track card frequencies
	cardFrequencies := make(map[int]int) // cardArenaID -> deck count

	// Process each deck
	for _, deck := range decks {
		// Convert card names to Arena IDs
		arenaIDs := make([]int, 0, len(deck.CardNames))
		for _, name := range deck.CardNames {
			arenaID, err := a.cardNameMap.GetArenaIDByName(ctx, name)
			if err != nil || arenaID == 0 {
				continue // Skip unknown cards
			}
			arenaIDs = append(arenaIDs, arenaID)
		}

		if len(arenaIDs) < 2 {
			continue // Need at least 2 cards for co-occurrence
		}

		// Update card frequencies
		for _, arenaID := range arenaIDs {
			cardFrequencies[arenaID]++
		}

		// Generate all pairs and update co-occurrence counts
		for i := 0; i < len(arenaIDs); i++ {
			for j := i + 1; j < len(arenaIDs); j++ {
				if err := a.repo.IncrementCooccurrence(ctx, arenaIDs[i], arenaIDs[j], format); err != nil {
					log.Printf("[CooccurrenceAnalyzer] Failed to increment co-occurrence: %v", err)
					continue
				}
				result.PairsCreated++
			}
		}
	}

	// Update card frequencies
	for arenaID, deckCount := range cardFrequencies {
		if err := a.repo.UpsertCardFrequency(ctx, arenaID, format, deckCount, len(decks)); err != nil {
			log.Printf("[CooccurrenceAnalyzer] Failed to update card frequency: %v", err)
		}
	}
	result.CardsTracked = len(cardFrequencies)

	// Update source tracking
	if err := a.repo.UpsertSource(ctx, source.SourceName(), format, format, len(decks), len(cardFrequencies)); err != nil {
		log.Printf("[CooccurrenceAnalyzer] Failed to update source: %v", err)
	}

	// Calculate PMI scores
	log.Printf("[CooccurrenceAnalyzer] Calculating PMI scores...")
	if err := a.repo.UpdatePMIScores(ctx, format); err != nil {
		return nil, fmt.Errorf("failed to update PMI scores: %w", err)
	}

	log.Printf("[CooccurrenceAnalyzer] Analysis complete: %d decks, %d pairs, %d cards",
		result.DecksAnalyzed, result.PairsCreated, result.CardsTracked)

	return result, nil
}

// GetSynergyScore returns the co-occurrence synergy score for a card pair.
// Returns a normalized score from 0.0 to 1.0, where:
// - 0.0 means no synergy data or negative PMI (cards appear together less than chance)
// - 1.0 means very strong positive PMI (cards appear together much more than chance)
func (a *CooccurrenceAnalyzer) GetSynergyScore(ctx context.Context, cardAArenaID, cardBArenaID int, format string) (float64, error) {
	pmi, err := a.repo.GetCooccurrenceScore(ctx, cardAArenaID, cardBArenaID, format)
	if err != nil {
		return 0, err
	}

	// Normalize PMI to 0.0-1.0 range
	// Typical PMI range is -5 to +5, with 0 meaning independent
	// We map this to 0.0-1.0 with sigmoid-like scaling
	return normalizePMI(pmi), nil
}

// GetTopSynergies returns the top co-occurring cards for a given card.
func (a *CooccurrenceAnalyzer) GetTopSynergies(ctx context.Context, cardArenaID int, format string, limit int) ([]CardSynergy, error) {
	coocs, err := a.repo.GetTopCooccurrences(ctx, cardArenaID, format, limit)
	if err != nil {
		return nil, err
	}

	result := make([]CardSynergy, 0, len(coocs))
	for _, cooc := range coocs {
		// Determine which card is the "other" card
		otherCardID := cooc.CardBArenaID
		if cooc.CardAArenaID != cardArenaID {
			otherCardID = cooc.CardAArenaID
		}

		result = append(result, CardSynergy{
			CardArenaID: otherCardID,
			Score:       normalizePMI(cooc.PMIScore),
			RawPMI:      cooc.PMIScore,
			Count:       cooc.Count,
		})
	}

	return result, nil
}

// CardSynergy represents synergy between cards based on co-occurrence.
type CardSynergy struct {
	CardArenaID int     // The other card's Arena ID
	Score       float64 // Normalized score (0.0-1.0)
	RawPMI      float64 // Raw PMI value
	Count       int     // Number of decks containing both cards
}

// normalizePMI converts PMI values to a 0.0-1.0 scale.
// PMI typically ranges from -inf to +inf, with 0 meaning independence.
// We use a sigmoid-like function to squash this to 0.0-1.0.
func normalizePMI(pmi float64) float64 {
	// PMI of 0 = 0.5 (independent)
	// PMI of +5 ≈ 0.95 (strong positive)
	// PMI of -5 ≈ 0.05 (strong negative)

	// Negative PMI means cards appear together less than chance
	// For synergy purposes, treat this as 0
	if pmi <= 0 {
		return 0.0
	}

	// Positive PMI: scale 0 to ~5 to 0 to 1
	// Using tanh for smooth scaling
	// tanh(pmi/2) gives us ~0.46 at pmi=1, ~0.76 at pmi=2, ~0.96 at pmi=4
	normalized := (pmi / 5.0) // Simple linear scaling for now
	if normalized > 1.0 {
		normalized = 1.0
	}

	return normalized
}

// LocalDeckSource implements DeckSource using local deck data.
type LocalDeckSource struct {
	decks []*SimpleDeck
}

// NewLocalDeckSource creates a deck source from local decks.
func NewLocalDeckSource(decks []*SimpleDeck) *LocalDeckSource {
	return &LocalDeckSource{decks: decks}
}

// FetchDecks returns the local decks, optionally filtered by format.
func (s *LocalDeckSource) FetchDecks(ctx context.Context, format string, limit int) ([]*SimpleDeck, error) {
	if format == "" || format == "all" {
		if limit > 0 && limit < len(s.decks) {
			return s.decks[:limit], nil
		}
		return s.decks, nil
	}

	result := make([]*SimpleDeck, 0)
	for _, deck := range s.decks {
		if deck.Format == format {
			result = append(result, deck)
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}

	return result, nil
}

// SourceName returns the source name.
func (s *LocalDeckSource) SourceName() string {
	return "local"
}
