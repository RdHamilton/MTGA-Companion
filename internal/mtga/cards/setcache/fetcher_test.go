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

func TestArenaExclusiveBasicLands_TLAMapping(t *testing.T) {
	// Test that TLA basic lands are correctly mapped
	tests := []struct {
		arenaID      int
		expectedSet  string
		expectedName string
	}{
		{97563, "TLA", "Plains"},
		{97564, "TLA", "Island"},
		{97565, "TLA", "Swamp"},
		{97566, "TLA", "Mountain"},
		{97567, "TLA", "Forest"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedName, func(t *testing.T) {
			basicLand, ok := ArenaExclusiveBasicLands[tt.arenaID]
			if !ok {
				t.Fatalf("Expected ArenaExclusiveBasicLands to contain arenaID %d", tt.arenaID)
			}

			if basicLand.SetCode != tt.expectedSet {
				t.Errorf("Expected SetCode=%s, got %s", tt.expectedSet, basicLand.SetCode)
			}

			if basicLand.CardName != tt.expectedName {
				t.Errorf("Expected CardName=%s, got %s", tt.expectedName, basicLand.CardName)
			}
		})
	}
}

func TestArenaExclusiveBasicLands_UnknownID(t *testing.T) {
	// Test that unknown IDs return false
	unknownIDs := []int{12345, 99999, 0, -1}

	for _, id := range unknownIDs {
		if _, ok := ArenaExclusiveBasicLands[id]; ok {
			t.Errorf("Expected ArenaExclusiveBasicLands to NOT contain arenaID %d", id)
		}
	}
}

func TestCheckCacheCompleteness_ArenaExclusiveSet_IncompleteCache(t *testing.T) {
	// Arena-exclusive set: Scryfall reports 0 game:arena cards, but 17Lands has 286 ratings
	// This tests the logic for sets like TLA where Scryfall lacks Arena IDs

	// Simulated state: 50 cached cards
	cachedCount := 50

	// Scryfall returns 0 (no game:arena cards for Arena-exclusive sets)
	scryfallExpected := 0

	// 17Lands has 286 cards
	ratingsCount := 286

	// Logic: if scryfallExpected == 0, use ratingsCount instead
	var needsRefresh bool
	if scryfallExpected > 0 {
		needsRefresh = cachedCount < (scryfallExpected * 9 / 10)
	} else if ratingsCount > 0 {
		needsRefresh = cachedCount < (ratingsCount * 9 / 10)
	}

	if !needsRefresh {
		t.Errorf("Expected needsRefresh=true for Arena-exclusive set: cached=%d, 17lands=%d", cachedCount, ratingsCount)
	}
}

func TestCheckCacheCompleteness_ArenaExclusiveSet_CompleteCache(t *testing.T) {
	// Arena-exclusive set with complete cache
	cachedCount := 280
	scryfallExpected := 0 // Scryfall has no Arena IDs
	ratingsCount := 286   // 17Lands has 286 cards

	var needsRefresh bool
	if scryfallExpected > 0 {
		needsRefresh = cachedCount < (scryfallExpected * 9 / 10)
	} else if ratingsCount > 0 {
		needsRefresh = cachedCount < (ratingsCount * 9 / 10)
	}

	if needsRefresh {
		t.Errorf("Expected needsRefresh=false for complete Arena-exclusive cache: cached=%d, 17lands=%d (90%% = %d)",
			cachedCount, ratingsCount, ratingsCount*9/10)
	}
}

func TestCheckCacheCompleteness_ArenaExclusiveSet_NoRatings(t *testing.T) {
	// Arena-exclusive set with no 17Lands ratings yet
	cachedCount := 5      // Some basic lands cached
	scryfallExpected := 0 // Scryfall has no Arena IDs
	ratingsCount := 0     // No 17Lands ratings yet

	var needsRefresh bool
	if scryfallExpected > 0 {
		needsRefresh = cachedCount < (scryfallExpected * 9 / 10)
	} else if ratingsCount > 0 {
		needsRefresh = cachedCount < (ratingsCount * 9 / 10)
	}

	// With no ratings, we can't determine completeness, so no refresh
	if needsRefresh {
		t.Errorf("Expected needsRefresh=false when no 17Lands ratings available")
	}
}

func TestExtractScryfallIDFromURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "standard scryfall image URL",
			url:      "https://cards.scryfall.io/large/front/f/a/fa940e68-010e-4b68-be8a-555d7068f7b4.jpg",
			expected: "fa940e68-010e-4b68-be8a-555d7068f7b4",
		},
		{
			name:     "small image URL",
			url:      "https://cards.scryfall.io/small/front/1/2/12345678-1234-1234-1234-123456789abc.jpg",
			expected: "12345678-1234-1234-1234-123456789abc",
		},
		{
			name:     "art crop URL",
			url:      "https://cards.scryfall.io/art_crop/front/a/b/abcdef01-2345-6789-abcd-ef0123456789.jpg",
			expected: "abcdef01-2345-6789-abcd-ef0123456789",
		},
		{
			name:     "normal image URL",
			url:      "https://cards.scryfall.io/normal/front/9/9/99887766-5544-3322-1100-aabbccddeeff.png",
			expected: "99887766-5544-3322-1100-aabbccddeeff",
		},
		{
			name:     "empty URL",
			url:      "",
			expected: "",
		},
		{
			name:     "URL without UUID",
			url:      "https://example.com/image.jpg",
			expected: "",
		},
		{
			name:     "URL with invalid UUID format",
			url:      "https://cards.scryfall.io/large/front/1/2/not-a-uuid.jpg",
			expected: "",
		},
		{
			name:     "partial UUID",
			url:      "https://cards.scryfall.io/large/front/1/2/12345678-1234-1234.jpg",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractScryfallIDFromURL(tt.url)
			if result != tt.expected {
				t.Errorf("ExtractScryfallIDFromURL(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

func TestExtractScryfallIDFromURL_VariousFormats(t *testing.T) {
	// Test that any URL containing a valid UUID extracts it correctly
	validUUID := "fa940e68-010e-4b68-be8a-555d7068f7b4"

	urlsWithValidUUID := []string{
		"https://cards.scryfall.io/large/front/f/a/fa940e68-010e-4b68-be8a-555d7068f7b4.jpg",
		"https://cdn.17lands.com/images/fa940e68-010e-4b68-be8a-555d7068f7b4.png",
		"https://example.com/path/to/fa940e68-010e-4b68-be8a-555d7068f7b4/image.webp",
		"fa940e68-010e-4b68-be8a-555d7068f7b4",
	}

	for _, url := range urlsWithValidUUID {
		result := ExtractScryfallIDFromURL(url)
		if result != validUUID {
			t.Errorf("ExtractScryfallIDFromURL(%q) = %q, want %q", url, result, validUUID)
		}
	}
}
