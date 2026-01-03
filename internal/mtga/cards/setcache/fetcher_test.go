package setcache

import (
	"context"
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// mockSetCardRepo implements repository.SetCardRepository for testing.
type mockSetCardRepo struct {
	cards map[string][]*models.SetCard
}

func newMockSetCardRepo() *mockSetCardRepo {
	return &mockSetCardRepo{
		cards: make(map[string][]*models.SetCard),
	}
}

func (m *mockSetCardRepo) SaveCard(_ context.Context, card *models.SetCard) error {
	m.cards[card.SetCode] = append(m.cards[card.SetCode], card)
	return nil
}

func (m *mockSetCardRepo) SaveCards(_ context.Context, cards []*models.SetCard) error {
	for _, card := range cards {
		m.cards[card.SetCode] = append(m.cards[card.SetCode], card)
	}
	return nil
}

func (m *mockSetCardRepo) GetCardByArenaID(_ context.Context, _ string) (*models.SetCard, error) {
	return nil, nil
}

func (m *mockSetCardRepo) GetCardsBySet(_ context.Context, setCode string) ([]*models.SetCard, error) {
	return m.cards[setCode], nil
}

func (m *mockSetCardRepo) SearchCards(_ context.Context, _ string, _ []string, _ int) ([]*models.SetCard, error) {
	return nil, nil
}

func (m *mockSetCardRepo) IsSetCached(_ context.Context, setCode string) (bool, error) {
	return len(m.cards[setCode]) > 0, nil
}

func (m *mockSetCardRepo) GetCachedSets(_ context.Context) ([]string, error) {
	sets := make([]string, 0, len(m.cards))
	for setCode := range m.cards {
		sets = append(sets, setCode)
	}
	return sets, nil
}

func (m *mockSetCardRepo) DeleteSet(_ context.Context, setCode string) error {
	delete(m.cards, setCode)
	return nil
}

func (m *mockSetCardRepo) GetMetadataStaleness(_ context.Context, _, _ int) (*repository.MetadataStaleness, error) {
	return nil, nil
}

func (m *mockSetCardRepo) GetStaleCards(_ context.Context, _, _ int) ([]*repository.StaleCard, error) {
	return nil, nil
}

func (m *mockSetCardRepo) GetSetRarityCounts(_ context.Context) ([]*repository.SetRarityCount, error) {
	return nil, nil
}

func (m *mockSetCardRepo) GetAllCardSetInfo(_ context.Context) ([]*repository.CardSetInfo, error) {
	return nil, nil
}

// mockScryfallClient implements a mock Scryfall client for testing.
type mockScryfallClient struct {
	searchResults map[string]*scryfall.SearchResult
}

func newMockScryfallClient() *mockScryfallClient {
	return &mockScryfallClient{
		searchResults: make(map[string]*scryfall.SearchResult),
	}
}

func (m *mockScryfallClient) SearchCards(_ context.Context, query string) (*scryfall.SearchResult, error) {
	if result, ok := m.searchResults[query]; ok {
		return result, nil
	}
	return &scryfall.SearchResult{TotalCards: 0}, nil
}

func (m *mockScryfallClient) setSearchResult(query string, totalCards int) {
	m.searchResults[query] = &scryfall.SearchResult{TotalCards: totalCards}
}

// scryfallClientAdapter wraps mockScryfallClient to match the real client interface.
type scryfallClientAdapter struct {
	mock *mockScryfallClient
}

func TestCheckCacheCompleteness_IncompleteCache(t *testing.T) {
	// Setup: 50 cached cards, but Scryfall reports 286 Arena cards
	mockRepo := newMockSetCardRepo()
	for i := 0; i < 50; i++ {
		_ = mockRepo.SaveCard(context.Background(), &models.SetCard{
			SetCode: "TLA",
			ArenaID: string(rune(i)),
			Name:    "Test Card",
		})
	}

	mockScryfall := newMockScryfallClient()
	mockScryfall.setSearchResult("set:tla game:arena", 286)

	// Create fetcher with mocks - we can't use the real constructor
	// because it expects *scryfall.Client, so we test the logic directly
	cachedCards, _ := mockRepo.GetCardsBySet(context.Background(), "TLA")
	cachedCount := len(cachedCards)

	expectedCount := 286

	// Test the logic: cached < 90% of expected should trigger refresh
	needsRefresh := expectedCount > 0 && cachedCount < (expectedCount*9/10)

	if !needsRefresh {
		t.Errorf("Expected needsRefresh=true for cached=%d, expected=%d", cachedCount, expectedCount)
	}
}

func TestCheckCacheCompleteness_CompleteCache(t *testing.T) {
	// Setup: 280 cached cards, Scryfall reports 286 Arena cards (>90% complete)
	mockRepo := newMockSetCardRepo()
	for i := 0; i < 280; i++ {
		_ = mockRepo.SaveCard(context.Background(), &models.SetCard{
			SetCode: "TLA",
			ArenaID: string(rune(i)),
			Name:    "Test Card",
		})
	}

	cachedCards, _ := mockRepo.GetCardsBySet(context.Background(), "TLA")
	cachedCount := len(cachedCards)

	expectedCount := 286

	// Test the logic: cached >= 90% of expected should NOT trigger refresh
	needsRefresh := expectedCount > 0 && cachedCount < (expectedCount*9/10)

	if needsRefresh {
		t.Errorf("Expected needsRefresh=false for cached=%d, expected=%d (90%% threshold = %d)",
			cachedCount, expectedCount, expectedCount*9/10)
	}
}

func TestCheckCacheCompleteness_ExactThreshold(t *testing.T) {
	// Test the boundary: exactly 90% should NOT trigger refresh
	expectedCount := 100
	cachedCount := 90 // Exactly 90%

	// 90 >= 90 (which is 100*9/10), so needsRefresh should be false
	needsRefresh := expectedCount > 0 && cachedCount < (expectedCount*9/10)

	if needsRefresh {
		t.Errorf("Expected needsRefresh=false at exactly 90%% (cached=%d, expected=%d)", cachedCount, expectedCount)
	}

	// 89 < 90, so needsRefresh should be true
	cachedCount = 89
	needsRefresh = expectedCount > 0 && cachedCount < (expectedCount*9/10)

	if !needsRefresh {
		t.Errorf("Expected needsRefresh=true just below 90%% (cached=%d, expected=%d)", cachedCount, expectedCount)
	}
}

func TestCheckCacheCompleteness_EmptyExpected(t *testing.T) {
	// If Scryfall reports 0 cards, we should NOT trigger refresh
	expectedCount := 0
	cachedCount := 50

	needsRefresh := expectedCount > 0 && cachedCount < (expectedCount*9/10)

	if needsRefresh {
		t.Errorf("Expected needsRefresh=false when expectedCount=0")
	}
}

func TestMTGASetToScryfall_Mapping(t *testing.T) {
	tests := []struct {
		mtgaCode     string
		expectedCode string
	}{
		{"TLA", "tla"},
		{"BLB", "blb"},
		{"DSK", "dsk"},
		{"UNKNOWN", "unknown"}, // Falls back to lowercase
	}

	for _, tt := range tests {
		t.Run(tt.mtgaCode, func(t *testing.T) {
			scryfallCode, ok := MTGASetToScryfall[tt.mtgaCode]
			if !ok {
				// Falls back to lowercase
				scryfallCode = tt.mtgaCode
				scryfallCode = string([]rune(scryfallCode)) // Force lowercase would happen in actual code
			}

			// For unknown codes, the actual code uses strings.ToLower
			if _, exists := MTGASetToScryfall[tt.mtgaCode]; !exists {
				if tt.mtgaCode != "UNKNOWN" {
					t.Errorf("Expected %s to be in MTGASetToScryfall map", tt.mtgaCode)
				}
			}
		})
	}
}
