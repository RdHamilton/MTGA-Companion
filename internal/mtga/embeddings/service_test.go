package embeddings

import (
	"context"
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// mockEmbeddingRepo is a mock implementation of the EmbeddingRepository interface.
type mockEmbeddingRepo struct {
	embeddings   map[int]*models.CardEmbedding
	similarities map[int][]*models.SimilarCard
	upsertErr    error
	getErr       error
}

func newMockEmbeddingRepo() *mockEmbeddingRepo {
	return &mockEmbeddingRepo{
		embeddings:   make(map[int]*models.CardEmbedding),
		similarities: make(map[int][]*models.SimilarCard),
	}
}

func (m *mockEmbeddingRepo) UpsertEmbedding(ctx context.Context, embedding *models.CardEmbedding) error {
	if m.upsertErr != nil {
		return m.upsertErr
	}
	m.embeddings[embedding.ArenaID] = embedding
	return nil
}

func (m *mockEmbeddingRepo) GetEmbedding(ctx context.Context, arenaID int) (*models.CardEmbedding, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.embeddings[arenaID], nil
}

func (m *mockEmbeddingRepo) GetEmbeddings(ctx context.Context, arenaIDs []int) ([]*models.CardEmbedding, error) {
	var result []*models.CardEmbedding
	for _, id := range arenaIDs {
		if emb, ok := m.embeddings[id]; ok {
			result = append(result, emb)
		}
	}
	return result, nil
}

func (m *mockEmbeddingRepo) GetAllEmbeddings(ctx context.Context) ([]*models.CardEmbedding, error) {
	result := make([]*models.CardEmbedding, 0, len(m.embeddings))
	for _, emb := range m.embeddings {
		result = append(result, emb)
	}
	return result, nil
}

func (m *mockEmbeddingRepo) DeleteEmbedding(ctx context.Context, arenaID int) error {
	delete(m.embeddings, arenaID)
	return nil
}

func (m *mockEmbeddingRepo) GetEmbeddingCount(ctx context.Context) (int, error) {
	return len(m.embeddings), nil
}

func (m *mockEmbeddingRepo) GetOutdatedEmbeddings(ctx context.Context, version int) ([]int, error) {
	var outdated []int
	for id, emb := range m.embeddings {
		if emb.EmbeddingVersion < version {
			outdated = append(outdated, id)
		}
	}
	return outdated, nil
}

func (m *mockEmbeddingRepo) UpsertSimilarity(ctx context.Context, similarity *models.CardSimilarity) error {
	return nil
}

func (m *mockEmbeddingRepo) BulkUpsertSimilarities(ctx context.Context, similarities []*models.CardSimilarity) error {
	return nil
}

func (m *mockEmbeddingRepo) GetSimilarCards(ctx context.Context, arenaID int, limit int) ([]*models.SimilarCard, error) {
	if sims, ok := m.similarities[arenaID]; ok {
		if len(sims) > limit {
			return sims[:limit], nil
		}
		return sims, nil
	}
	return nil, nil
}

func (m *mockEmbeddingRepo) GetSimilarityBetween(ctx context.Context, arenaID1, arenaID2 int) (float64, error) {
	return 0, nil
}

func (m *mockEmbeddingRepo) ClearSimilarityCache(ctx context.Context) error {
	m.similarities = make(map[int][]*models.SimilarCard)
	return nil
}

func (m *mockEmbeddingRepo) ClearSimilarityCacheForCard(ctx context.Context, arenaID int) error {
	delete(m.similarities, arenaID)
	return nil
}

func TestService_GenerateAndStore(t *testing.T) {
	repo := newMockEmbeddingRepo()
	service := NewService(repo)

	card := &CardData{
		ArenaID:    12345,
		Name:       "Test Card",
		ManaCost:   "{1}{U}",
		CMC:        2,
		TypeLine:   "Creature — Human",
		Colors:     []string{"U"},
		OracleText: "Flying",
		Power:      "2",
		Toughness:  "1",
		Rarity:     "common",
	}

	ctx := context.Background()
	emb, err := service.GenerateAndStore(ctx, card)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if emb == nil {
		t.Fatal("expected embedding to not be nil")
	}
	if emb.ArenaID != card.ArenaID {
		t.Errorf("expected ArenaID %d, got %d", card.ArenaID, emb.ArenaID)
	}

	// Check that it was stored in the repository
	stored, err := repo.GetEmbedding(ctx, card.ArenaID)
	if err != nil {
		t.Fatalf("unexpected error getting stored embedding: %v", err)
	}
	if stored == nil {
		t.Fatal("expected embedding to be stored in repository")
	}
}

func TestService_GetEmbedding_FromCache(t *testing.T) {
	repo := newMockEmbeddingRepo()
	service := NewService(repo)

	// Store an embedding
	card := &CardData{
		ArenaID:  12345,
		Name:     "Test Card",
		CMC:      2,
		TypeLine: "Creature",
		Colors:   []string{"U"},
		Rarity:   "common",
	}

	ctx := context.Background()
	_, err := service.GenerateAndStore(ctx, card)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get embedding (should come from cache)
	emb, err := service.GetEmbedding(ctx, card.ArenaID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if emb == nil {
		t.Fatal("expected embedding to not be nil")
	}
	if emb.ArenaID != card.ArenaID {
		t.Errorf("expected ArenaID %d, got %d", card.ArenaID, emb.ArenaID)
	}
}

func TestService_GetSimilarCards(t *testing.T) {
	repo := newMockEmbeddingRepo()
	service := NewService(repo)

	ctx := context.Background()

	// Generate embeddings for several cards
	cards := []*CardData{
		{ArenaID: 1, Name: "Card A", CMC: 2, TypeLine: "Creature", Colors: []string{"U"}, Rarity: "common", OracleText: "Flying"},
		{ArenaID: 2, Name: "Card B", CMC: 2, TypeLine: "Creature", Colors: []string{"U"}, Rarity: "common", OracleText: "Flying"},
		{ArenaID: 3, Name: "Card C", CMC: 5, TypeLine: "Sorcery", Colors: []string{"R"}, Rarity: "rare"},
	}

	for _, card := range cards {
		_, err := service.GenerateAndStore(ctx, card)
		if err != nil {
			t.Fatalf("unexpected error storing card %s: %v", card.Name, err)
		}
	}

	// Get similar cards to Card A (should find Card B as most similar)
	similar, err := service.GetSimilarCards(ctx, 1, 2, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(similar) == 0 {
		t.Fatal("expected at least one similar card")
	}

	// Card B should be more similar to Card A than Card C
	foundB := false
	for _, sim := range similar {
		if sim.ArenaID == 2 {
			foundB = true
			break
		}
	}
	if !foundB {
		t.Error("expected Card B to be in similar cards for Card A")
	}
}

func TestService_ComputeSimilarity(t *testing.T) {
	repo := newMockEmbeddingRepo()
	service := NewService(repo)

	ctx := context.Background()

	// Generate embeddings for two similar cards
	card1 := &CardData{ArenaID: 1, Name: "Bear A", CMC: 2, TypeLine: "Creature — Bear", Colors: []string{"G"}, Rarity: "common"}
	card2 := &CardData{ArenaID: 2, Name: "Bear B", CMC: 2, TypeLine: "Creature — Bear", Colors: []string{"G"}, Rarity: "common"}

	_, _ = service.GenerateAndStore(ctx, card1)
	_, _ = service.GenerateAndStore(ctx, card2)

	similarity, err := service.ComputeSimilarity(ctx, 1, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Similar cards should have high similarity
	if similarity < 0.8 {
		t.Errorf("expected high similarity for similar cards, got %f", similarity)
	}
}

func TestService_GetSynergyBonus(t *testing.T) {
	repo := newMockEmbeddingRepo()
	service := NewService(repo)

	ctx := context.Background()

	// Generate embeddings for a deck of similar cards
	deckCards := []*CardData{
		{ArenaID: 1, Name: "Elf A", CMC: 1, TypeLine: "Creature — Elf", Colors: []string{"G"}, Rarity: "common"},
		{ArenaID: 2, Name: "Elf B", CMC: 2, TypeLine: "Creature — Elf", Colors: []string{"G"}, Rarity: "common"},
		{ArenaID: 3, Name: "Elf C", CMC: 3, TypeLine: "Creature — Elf", Colors: []string{"G"}, Rarity: "uncommon"},
	}

	for _, card := range deckCards {
		_, _ = service.GenerateAndStore(ctx, card)
	}

	// New card that synergizes with the deck
	newCard := &CardData{ArenaID: 4, Name: "Elf Lord", CMC: 3, TypeLine: "Creature — Elf", Colors: []string{"G"}, Rarity: "rare"}
	_, _ = service.GenerateAndStore(ctx, newCard)

	// Card that doesn't synergize
	otherCard := &CardData{ArenaID: 5, Name: "Counterspell", CMC: 2, TypeLine: "Instant", Colors: []string{"U"}, Rarity: "common", OracleText: "Counter target spell."}
	_, _ = service.GenerateAndStore(ctx, otherCard)

	deckIDs := []int{1, 2, 3}

	// Get synergy bonus for elf card (should be high)
	elfBonus := service.GetSynergyBonus(ctx, 4, deckIDs)

	// Get synergy bonus for counterspell (should be lower)
	counterBonus := service.GetSynergyBonus(ctx, 5, deckIDs)

	if elfBonus <= counterBonus {
		t.Errorf("expected elf card (%.4f) to have higher synergy than counterspell (%.4f)", elfBonus, counterBonus)
	}
}

func TestService_GetEmbeddingStats(t *testing.T) {
	repo := newMockEmbeddingRepo()
	service := NewService(repo)

	ctx := context.Background()

	// Store some embeddings
	for i := 1; i <= 5; i++ {
		card := &CardData{ArenaID: i, Name: "Card", CMC: 2, TypeLine: "Creature", Colors: []string{"U"}, Rarity: "common"}
		_, _ = service.GenerateAndStore(ctx, card)
	}

	stats, err := service.GetEmbeddingStats(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats["total_embeddings"].(int) != 5 {
		t.Errorf("expected 5 total embeddings, got %d", stats["total_embeddings"])
	}
	if stats["current_version"].(int) != models.EmbeddingVersion {
		t.Errorf("expected version %d, got %d", models.EmbeddingVersion, stats["current_version"])
	}
	if stats["dimensions"].(int) != models.EmbeddingDimensions {
		t.Errorf("expected %d dimensions, got %d", models.EmbeddingDimensions, stats["dimensions"])
	}
}

func TestService_GenerateAndStoreBatch(t *testing.T) {
	repo := newMockEmbeddingRepo()
	service := NewService(repo)

	ctx := context.Background()

	cards := []*CardData{
		{ArenaID: 1, Name: "Card A", CMC: 1, TypeLine: "Creature", Colors: []string{"W"}, Rarity: "common"},
		{ArenaID: 2, Name: "Card B", CMC: 2, TypeLine: "Instant", Colors: []string{"U"}, Rarity: "uncommon"},
		{ArenaID: 3, Name: "Card C", CMC: 3, TypeLine: "Sorcery", Colors: []string{"B"}, Rarity: "rare"},
	}

	err := service.GenerateAndStoreBatch(ctx, cards)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all cards were stored
	count, _ := repo.GetEmbeddingCount(ctx)
	if count != 3 {
		t.Errorf("expected 3 embeddings, got %d", count)
	}
}

func TestService_LoadAllToCache(t *testing.T) {
	repo := newMockEmbeddingRepo()
	service := NewService(repo)

	ctx := context.Background()

	// Store embeddings directly in repo
	cards := []*CardData{
		{ArenaID: 1, Name: "Card A", CMC: 1, TypeLine: "Creature", Colors: []string{"W"}, Rarity: "common"},
		{ArenaID: 2, Name: "Card B", CMC: 2, TypeLine: "Instant", Colors: []string{"U"}, Rarity: "uncommon"},
	}

	for _, card := range cards {
		emb := service.generator.GenerateEmbedding(card)
		_ = repo.UpsertEmbedding(ctx, emb)
	}

	// Load to cache
	err := service.LoadAllToCache(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get stats to check cache size
	stats, _ := service.GetEmbeddingStats(ctx)
	if stats["cache_size"].(int) != 2 {
		t.Errorf("expected cache size 2, got %d", stats["cache_size"])
	}
}
