package embeddings

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// Service provides card embedding operations.
type Service struct {
	repo      repository.EmbeddingRepository
	generator *Generator
	cache     map[int]*models.CardEmbedding // In-memory cache for fast similarity lookups
	cacheMu   sync.RWMutex
}

// NewService creates a new embedding service.
func NewService(repo repository.EmbeddingRepository) *Service {
	return &Service{
		repo:      repo,
		generator: NewGenerator(),
		cache:     make(map[int]*models.CardEmbedding),
	}
}

// GenerateAndStore generates an embedding for a card and stores it.
func (s *Service) GenerateAndStore(ctx context.Context, card *CardData) (*models.CardEmbedding, error) {
	embedding := s.generator.GenerateEmbedding(card)

	if err := s.repo.UpsertEmbedding(ctx, embedding); err != nil {
		return nil, fmt.Errorf("failed to store embedding: %w", err)
	}

	// Update cache
	s.cacheMu.Lock()
	s.cache[card.ArenaID] = embedding
	s.cacheMu.Unlock()

	return embedding, nil
}

// GenerateAndStoreBatch generates and stores embeddings for multiple cards.
func (s *Service) GenerateAndStoreBatch(ctx context.Context, cards []*CardData) error {
	for _, card := range cards {
		embedding := s.generator.GenerateEmbedding(card)
		if err := s.repo.UpsertEmbedding(ctx, embedding); err != nil {
			return fmt.Errorf("failed to store embedding for %s: %w", card.Name, err)
		}

		// Update cache
		s.cacheMu.Lock()
		s.cache[card.ArenaID] = embedding
		s.cacheMu.Unlock()
	}

	return nil
}

// GetEmbedding retrieves an embedding by arena ID.
func (s *Service) GetEmbedding(ctx context.Context, arenaID int) (*models.CardEmbedding, error) {
	// Check cache first
	s.cacheMu.RLock()
	if emb, ok := s.cache[arenaID]; ok {
		s.cacheMu.RUnlock()
		return emb, nil
	}
	s.cacheMu.RUnlock()

	// Fall back to database
	emb, err := s.repo.GetEmbedding(ctx, arenaID)
	if err != nil {
		return nil, err
	}

	if emb != nil {
		s.cacheMu.Lock()
		s.cache[arenaID] = emb
		s.cacheMu.Unlock()
	}

	return emb, nil
}

// LoadAllToCache loads all embeddings into the in-memory cache.
func (s *Service) LoadAllToCache(ctx context.Context) error {
	embeddings, err := s.repo.GetAllEmbeddings(ctx)
	if err != nil {
		return fmt.Errorf("failed to load embeddings: %w", err)
	}

	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	for _, emb := range embeddings {
		s.cache[emb.ArenaID] = emb
	}

	return nil
}

// GetSimilarCards finds the most similar cards to a given card.
// If useCache is true, it will try to use the pre-computed similarity cache.
// Otherwise, it computes similarities on-the-fly.
func (s *Service) GetSimilarCards(ctx context.Context, arenaID int, limit int, useCache bool) ([]*models.SimilarCard, error) {
	// Try cache first
	if useCache {
		cached, err := s.repo.GetSimilarCards(ctx, arenaID, limit)
		if err == nil && len(cached) > 0 {
			return cached, nil
		}
	}

	// Compute on-the-fly
	return s.computeSimilarCards(ctx, arenaID, limit)
}

// computeSimilarCards computes similar cards using cosine similarity.
func (s *Service) computeSimilarCards(ctx context.Context, arenaID int, limit int) ([]*models.SimilarCard, error) {
	// Get the target embedding
	targetEmb, err := s.GetEmbedding(ctx, arenaID)
	if err != nil {
		return nil, fmt.Errorf("failed to get target embedding: %w", err)
	}
	if targetEmb == nil {
		return nil, fmt.Errorf("no embedding found for arena ID %d", arenaID)
	}

	// Load all embeddings if cache is empty
	s.cacheMu.RLock()
	cacheSize := len(s.cache)
	s.cacheMu.RUnlock()

	if cacheSize == 0 {
		if err := s.LoadAllToCache(ctx); err != nil {
			return nil, err
		}
	}

	// Compute similarities
	type scoredCard struct {
		arenaID int
		name    string
		score   float64
	}

	var scores []scoredCard

	s.cacheMu.RLock()
	for id, emb := range s.cache {
		if id == arenaID {
			continue // Skip self
		}
		score := CosineSimilarity(targetEmb.Embedding, emb.Embedding)
		scores = append(scores, scoredCard{
			arenaID: id,
			name:    emb.CardName,
			score:   score,
		})
	}
	s.cacheMu.RUnlock()

	// Sort by similarity score (descending)
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	// Take top N
	if len(scores) > limit {
		scores = scores[:limit]
	}

	// Convert to result format
	result := make([]*models.SimilarCard, len(scores))
	for i, sc := range scores {
		result[i] = &models.SimilarCard{
			ArenaID:         sc.arenaID,
			CardName:        sc.name,
			SimilarityScore: sc.score,
			Rank:            i + 1,
		}
	}

	return result, nil
}

// ComputeSimilarity calculates the similarity between two cards.
func (s *Service) ComputeSimilarity(ctx context.Context, arenaID1, arenaID2 int) (float64, error) {
	// Try cache first
	cached, err := s.repo.GetSimilarityBetween(ctx, arenaID1, arenaID2)
	if err == nil && cached > 0 {
		return cached, nil
	}

	// Get embeddings
	emb1, err := s.GetEmbedding(ctx, arenaID1)
	if err != nil || emb1 == nil {
		return 0, fmt.Errorf("failed to get embedding for %d", arenaID1)
	}

	emb2, err := s.GetEmbedding(ctx, arenaID2)
	if err != nil || emb2 == nil {
		return 0, fmt.Errorf("failed to get embedding for %d", arenaID2)
	}

	return CosineSimilarity(emb1.Embedding, emb2.Embedding), nil
}

// PrecomputeSimilarities pre-computes and caches similar cards for all embeddings.
func (s *Service) PrecomputeSimilarities(ctx context.Context, topK int) error {
	// Clear existing cache
	if err := s.repo.ClearSimilarityCache(ctx); err != nil {
		return fmt.Errorf("failed to clear similarity cache: %w", err)
	}

	// Load all embeddings
	if err := s.LoadAllToCache(ctx); err != nil {
		return err
	}

	s.cacheMu.RLock()
	embeddings := make([]*models.CardEmbedding, 0, len(s.cache))
	for _, emb := range s.cache {
		embeddings = append(embeddings, emb)
	}
	s.cacheMu.RUnlock()

	// For each card, compute similarities and store top-K
	for _, targetEmb := range embeddings {
		similarCards, err := s.computeSimilarCards(ctx, targetEmb.ArenaID, topK)
		if err != nil {
			continue // Skip failures
		}

		// Convert to similarity records
		similarities := make([]*models.CardSimilarity, len(similarCards))
		for i, sc := range similarCards {
			similarities[i] = &models.CardSimilarity{
				CardArenaID:     targetEmb.ArenaID,
				SimilarArenaID:  sc.ArenaID,
				SimilarityScore: sc.SimilarityScore,
				Rank:            sc.Rank,
			}
		}

		if err := s.repo.BulkUpsertSimilarities(ctx, similarities); err != nil {
			return fmt.Errorf("failed to store similarities for %d: %w", targetEmb.ArenaID, err)
		}
	}

	return nil
}

// GetSynergyBonus returns a synergy score based on card similarity.
// This can be used to boost synergy scores in the recommendation engine.
func (s *Service) GetSynergyBonus(ctx context.Context, cardArenaID int, deckCardArenaIDs []int) float64 {
	if len(deckCardArenaIDs) == 0 {
		return 0
	}

	cardEmb, err := s.GetEmbedding(ctx, cardArenaID)
	if err != nil || cardEmb == nil {
		return 0
	}

	var totalSimilarity float64
	var count int

	for _, deckCardID := range deckCardArenaIDs {
		if deckCardID == cardArenaID {
			continue
		}

		deckEmb, err := s.GetEmbedding(ctx, deckCardID)
		if err != nil || deckEmb == nil {
			continue
		}

		similarity := CosineSimilarity(cardEmb.Embedding, deckEmb.Embedding)
		totalSimilarity += similarity
		count++
	}

	if count == 0 {
		return 0
	}

	// Return average similarity as synergy bonus
	return totalSimilarity / float64(count)
}

// GetEmbeddingStats returns statistics about the embedding database.
func (s *Service) GetEmbeddingStats(ctx context.Context) (map[string]interface{}, error) {
	count, err := s.repo.GetEmbeddingCount(ctx)
	if err != nil {
		return nil, err
	}

	outdated, err := s.repo.GetOutdatedEmbeddings(ctx, models.EmbeddingVersion)
	if err != nil {
		return nil, err
	}

	s.cacheMu.RLock()
	cacheSize := len(s.cache)
	s.cacheMu.RUnlock()

	return map[string]interface{}{
		"total_embeddings":    count,
		"outdated_embeddings": len(outdated),
		"cache_size":          cacheSize,
		"current_version":     models.EmbeddingVersion,
		"dimensions":          models.EmbeddingDimensions,
	}, nil
}
