package unified

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// mockMetadataProvider implements CardMetadataProvider for testing.
type mockMetadataProvider struct {
	cards    map[int]*models.SetCard
	setCards map[string][]*models.SetCard
	err      error
}

func (m *mockMetadataProvider) GetCard(ctx context.Context, arenaID int) (*models.SetCard, error) {
	if m.err != nil {
		return nil, m.err
	}
	card, ok := m.cards[arenaID]
	if !ok {
		return nil, nil
	}
	return card, nil
}

func (m *mockMetadataProvider) GetCards(ctx context.Context, arenaIDs []int) ([]*models.SetCard, error) {
	if m.err != nil {
		return nil, m.err
	}
	var cards []*models.SetCard
	for _, id := range arenaIDs {
		if card, ok := m.cards[id]; ok {
			cards = append(cards, card)
		}
	}
	return cards, nil
}

func (m *mockMetadataProvider) GetSetCards(ctx context.Context, setCode string) ([]*models.SetCard, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.setCards[setCode], nil
}

// mockDraftStatsProvider implements DraftStatsProvider for testing.
type mockDraftStatsProvider struct {
	ratings    map[int]*storage.DraftCardRating
	setRatings map[string][]*storage.DraftCardRating
	err        error
}

func (m *mockDraftStatsProvider) GetCardRating(ctx context.Context, arenaID int, expansion, format, colors string) (*storage.DraftCardRating, error) {
	if m.err != nil {
		return nil, m.err
	}
	rating, ok := m.ratings[arenaID]
	if !ok {
		return nil, nil
	}
	return rating, nil
}

func (m *mockDraftStatsProvider) GetCardRatingsForSet(ctx context.Context, expansion, format, colors string) ([]*storage.DraftCardRating, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.setRatings[expansion], nil
}

// TestGetCard_WithStats tests getting a card with draft statistics.
func TestGetCard_WithStats(t *testing.T) {
	arenaID := 12345
	metadata := &mockMetadataProvider{
		cards: map[int]*models.SetCard{
			arenaID: {
				ID:       1,
				ArenaID:  fmt.Sprintf("%d", arenaID),
				Name:     "Test Card",
				ManaCost: "{2}{R}",
				CMC:      3,
				Types:    []string{"Creature"},
				Rarity:   "rare",
				SetCode:  "BLB",
			},
		},
	}

	draftstats := &mockDraftStatsProvider{
		ratings: map[int]*storage.DraftCardRating{
			arenaID: {
				ArenaID:     arenaID,
				Expansion:   "BLB",
				Format:      "PremierDraft",
				GIHWR:       0.585,
				ALSA:        4.2,
				ATA:         2.1,
				GIH:         500,
				GamesPlayed: 450,
				LastUpdated: time.Now().Add(-1 * time.Hour),
			},
		},
	}

	service := NewService(metadata, draftstats)
	ctx := context.Background()

	card, err := service.GetCard(ctx, arenaID, "BLB", "PremierDraft")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if card.Name != "Test Card" {
		t.Errorf("Expected name 'Test Card', got: %s", card.Name)
	}

	if !card.HasDraftStats() {
		t.Fatal("Expected card to have draft stats")
	}

	if card.DraftStats.GIHWR != 0.585 {
		t.Errorf("Expected GIHWR 0.585, got: %f", card.DraftStats.GIHWR)
	}

	if card.DraftStats.GIH != 500 {
		t.Errorf("Expected GIH 500, got: %d", card.DraftStats.GIH)
	}
}

// TestGetCard_WithoutStats tests getting a card without draft statistics.
func TestGetCard_WithoutStats(t *testing.T) {
	arenaID := 12345
	metadata := &mockMetadataProvider{
		cards: map[int]*models.SetCard{
			arenaID: {
				ID:      1,
				ArenaID: fmt.Sprintf("%d", arenaID),
				Name:    "Test Card",
				SetCode: "BLB",
			},
		},
	}

	draftstats := &mockDraftStatsProvider{
		ratings: map[int]*storage.DraftCardRating{}, // Empty - no stats
	}

	service := NewService(metadata, draftstats)
	ctx := context.Background()

	card, err := service.GetCard(ctx, arenaID, "BLB", "PremierDraft")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if card.Name != "Test Card" {
		t.Errorf("Expected name 'Test Card', got: %s", card.Name)
	}

	if card.HasDraftStats() {
		t.Error("Expected card to have no draft stats")
	}

	if card.IsComplete() {
		t.Error("Expected card to be incomplete (no stats)")
	}
}

// TestGetCards_Batch tests batch fetching of multiple cards.
func TestGetCards_Batch(t *testing.T) {
	arenaID1 := 1111
	arenaID2 := 2222
	arenaID3 := 3333

	metadata := &mockMetadataProvider{
		cards: map[int]*models.SetCard{
			arenaID1: {
				ID:      1,
				ArenaID: fmt.Sprintf("%d", arenaID1),
				Name:    "Card 1",
				SetCode: "BLB",
			},
			arenaID2: {
				ID:      2,
				ArenaID: fmt.Sprintf("%d", arenaID2),
				Name:    "Card 2",
				SetCode: "BLB",
			},
			arenaID3: {
				ID:      3,
				ArenaID: fmt.Sprintf("%d", arenaID3),
				Name:    "Card 3",
				SetCode: "MKM",
			},
		},
	}

	draftstats := &mockDraftStatsProvider{
		setRatings: map[string][]*storage.DraftCardRating{
			"BLB": {
				{ArenaID: arenaID1, GIHWR: 0.55, GIH: 200, LastUpdated: time.Now()},
				{ArenaID: arenaID2, GIHWR: 0.60, GIH: 300, LastUpdated: time.Now()},
			},
			"MKM": {
				{ArenaID: arenaID3, GIHWR: 0.52, GIH: 150, LastUpdated: time.Now()},
			},
		},
	}

	service := NewService(metadata, draftstats)
	ctx := context.Background()

	cards, err := service.GetCards(ctx, []int{arenaID1, arenaID2, arenaID3}, "PremierDraft")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(cards) != 3 {
		t.Fatalf("Expected 3 cards, got: %d", len(cards))
	}

	// All cards should have stats
	for i, card := range cards {
		if !card.HasDraftStats() {
			t.Errorf("Card %d (%s) missing draft stats", i, card.Name)
		}
	}

	// Verify specific cards
	if cards[0].Name != "Card 1" || cards[0].DraftStats.GIHWR != 0.55 {
		t.Error("Card 1 data incorrect")
	}
	if cards[1].Name != "Card 2" || cards[1].DraftStats.GIHWR != 0.60 {
		t.Error("Card 2 data incorrect")
	}
	if cards[2].Name != "Card 3" || cards[2].DraftStats.GIHWR != 0.52 {
		t.Error("Card 3 data incorrect")
	}
}

// TestGetSetCards tests fetching all cards for a set.
func TestGetSetCards(t *testing.T) {
	arenaID1 := 1111
	arenaID2 := 2222

	metadata := &mockMetadataProvider{
		setCards: map[string][]*models.SetCard{
			"BLB": {
				{ID: 1, ArenaID: fmt.Sprintf("%d", arenaID1), Name: "Card 1", SetCode: "BLB"},
				{ID: 2, ArenaID: fmt.Sprintf("%d", arenaID2), Name: "Card 2", SetCode: "BLB"},
			},
		},
	}

	draftstats := &mockDraftStatsProvider{
		setRatings: map[string][]*storage.DraftCardRating{
			"BLB": {
				{ArenaID: arenaID1, GIHWR: 0.55, GIH: 200, LastUpdated: time.Now()},
			},
		},
	}

	service := NewService(metadata, draftstats)
	ctx := context.Background()

	cards, err := service.GetSetCards(ctx, "BLB", "PremierDraft")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(cards) != 2 {
		t.Fatalf("Expected 2 cards, got: %d", len(cards))
	}

	// First card should have stats
	if !cards[0].HasDraftStats() {
		t.Error("Card 1 missing draft stats")
	}

	// Second card should not have stats
	if cards[1].HasDraftStats() {
		t.Error("Card 2 should not have draft stats")
	}
}

// TestFilterByRarity tests filtering cards by rarity.
func TestFilterByRarity(t *testing.T) {
	cards := []*UnifiedCard{
		{Name: "Common 1", Rarity: "common"},
		{Name: "Rare 1", Rarity: "rare"},
		{Name: "Common 2", Rarity: "common"},
		{Name: "Mythic 1", Rarity: "mythic"},
	}

	filtered := FilterByRarity(cards, "common")
	if len(filtered) != 2 {
		t.Errorf("Expected 2 common cards, got: %d", len(filtered))
	}

	filtered = FilterByRarity(cards, "rare")
	if len(filtered) != 1 {
		t.Errorf("Expected 1 rare card, got: %d", len(filtered))
	}
}

// TestFilterByStats tests filtering cards with draft statistics.
func TestFilterByStats(t *testing.T) {
	cards := []*UnifiedCard{
		{Name: "Card 1", DraftStats: &DraftStatistics{GIHWR: 0.55}},
		{Name: "Card 2", DraftStats: nil},
		{Name: "Card 3", DraftStats: &DraftStatistics{GIHWR: 0.60}},
	}

	filtered := FilterByStats(cards)
	if len(filtered) != 2 {
		t.Errorf("Expected 2 cards with stats, got: %d", len(filtered))
	}
}

// TestSortByWinRate tests sorting cards by win rate.
func TestSortByWinRate(t *testing.T) {
	cards := []*UnifiedCard{
		{Name: "Card 1", DraftStats: &DraftStatistics{GIHWR: 0.55}},
		{Name: "Card 2", DraftStats: &DraftStatistics{GIHWR: 0.70}},
		{Name: "Card 3", DraftStats: &DraftStatistics{GIHWR: 0.60}},
		{Name: "Card 4", DraftStats: nil}, // No stats
	}

	SortByWinRate(cards)

	// Should be sorted descending: 0.70, 0.60, 0.55, 0.0
	if cards[0].Name != "Card 2" {
		t.Errorf("Expected Card 2 first, got: %s", cards[0].Name)
	}
	if cards[1].Name != "Card 3" {
		t.Errorf("Expected Card 3 second, got: %s", cards[1].Name)
	}
	if cards[2].Name != "Card 1" {
		t.Errorf("Expected Card 1 third, got: %s", cards[2].Name)
	}
	if cards[3].Name != "Card 4" {
		t.Errorf("Expected Card 4 last, got: %s", cards[3].Name)
	}
}

// TestDraftStatistics_Methods tests DraftStatistics helper methods.
func TestDraftStatistics_Methods(t *testing.T) {
	stats := &DraftStatistics{
		GIHWR: 0.585,
		GIH:   150,
	}

	if stats.GetPrimaryWinRate() != 0.585 {
		t.Errorf("Expected primary win rate 0.585, got: %f", stats.GetPrimaryWinRate())
	}

	if stats.GetSampleSize() != 150 {
		t.Errorf("Expected sample size 150, got: %d", stats.GetSampleSize())
	}

	if !stats.IsSignificant() {
		t.Error("Expected stats to be significant (>100 games)")
	}

	lowSampleStats := &DraftStatistics{GIH: 50}
	if lowSampleStats.IsSignificant() {
		t.Error("Expected stats NOT to be significant (<100 games)")
	}
}

// TestUnifiedCard_Methods tests UnifiedCard helper methods.
func TestUnifiedCard_Methods(t *testing.T) {
	card := &UnifiedCard{
		ID:   "test-id",
		Name: "Test Card",
		DraftStats: &DraftStatistics{
			GIHWR:       0.585,
			LastUpdated: time.Now().Add(-2 * time.Hour),
		},
		MetadataAge: 1 * time.Hour,
		StatsAge:    2 * time.Hour,
	}

	if !card.HasDraftStats() {
		t.Error("Expected card to have draft stats")
	}

	if !card.IsComplete() {
		t.Error("Expected card to be complete")
	}

	if !card.HasFreshMetadata(3 * time.Hour) {
		t.Error("Expected metadata to be fresh")
	}

	if card.HasFreshStats(1 * time.Hour) {
		t.Error("Expected stats NOT to be fresh")
	}

	if !card.HasFreshStats(3 * time.Hour) {
		t.Error("Expected stats to be fresh within 3 hours")
	}
}

// TestDataSource_String tests DataSource string representation.
func TestDataSource_String(t *testing.T) {
	tests := []struct {
		source DataSource
		want   string
	}{
		{SourceAPI, "API"},
		{SourceCache, "Cache"},
		{SourceFallback, "Fallback"},
		{SourceUnknown, "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.source.String(); got != tt.want {
			t.Errorf("DataSource.String() = %s, want %s", got, tt.want)
		}
	}
}

// TestFilterByColors tests filtering cards by color identity.
func TestFilterByColors(t *testing.T) {
	cards := []*UnifiedCard{
		{Name: "Red Card", ColorIdentity: []string{"R"}},
		{Name: "Blue Card", ColorIdentity: []string{"U"}},
		{Name: "Izzet Card", ColorIdentity: []string{"U", "R"}},
		{Name: "Colorless Card", ColorIdentity: []string{}},
	}

	// Filter for red cards
	filtered := FilterByColors(cards, []string{"R"})
	if len(filtered) != 2 { // Red Card and Izzet Card
		t.Errorf("Expected 2 red cards, got: %d", len(filtered))
	}

	// Filter for blue cards
	filtered = FilterByColors(cards, []string{"U"})
	if len(filtered) != 2 { // Blue Card and Izzet Card
		t.Errorf("Expected 2 blue cards, got: %d", len(filtered))
	}

	// Filter for both U and R (only Izzet card)
	filtered = FilterByColors(cards, []string{"U", "R"})
	if len(filtered) != 1 {
		t.Errorf("Expected 1 card with U and R, got: %d", len(filtered))
	}

	// Empty filter returns all
	filtered = FilterByColors(cards, []string{})
	if len(filtered) != len(cards) {
		t.Errorf("Expected all cards, got: %d", len(filtered))
	}
}
