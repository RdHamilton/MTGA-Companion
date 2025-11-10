package storage

import (
	"context"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
)

func TestSaveAndGetCard(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// Create test card
	arenaID := 12345
	card := &Card{
		ID:              "test-card-id",
		ArenaID:         &arenaID,
		Name:            "Lightning Bolt",
		ManaCost:        "{R}",
		CMC:             1.0,
		TypeLine:        "Instant",
		OracleText:      "Deal 3 damage to any target.",
		Colors:          []string{"R"},
		ColorIdentity:   []string{"R"},
		Rarity:          "common",
		SetCode:         "blb",
		CollectorNumber: "123",
		ImageURIs: &scryfall.ImageURIs{
			Small:  "https://example.com/small.jpg",
			Normal: "https://example.com/normal.jpg",
			Large:  "https://example.com/large.jpg",
		},
		Layout:     "normal",
		CardFaces:  []scryfall.CardFace{},
		Legalities: map[string]string{"standard": "legal"},
		ReleasedAt: "2024-08-02",
	}

	// Save card
	err := service.SaveCard(ctx, card)
	if err != nil {
		t.Fatalf("Failed to save card: %v", err)
	}

	// Get card by Arena ID
	retrieved, err := service.GetCardByArenaID(ctx, arenaID)
	if err != nil {
		t.Fatalf("Failed to get card by Arena ID: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved card is nil")
	}

	// Verify fields
	if retrieved.Name != card.Name {
		t.Errorf("Expected name %s, got %s", card.Name, retrieved.Name)
	}

	if retrieved.ManaCost != card.ManaCost {
		t.Errorf("Expected mana cost %s, got %s", card.ManaCost, retrieved.ManaCost)
	}

	if retrieved.CMC != card.CMC {
		t.Errorf("Expected CMC %.1f, got %.1f", card.CMC, retrieved.CMC)
	}

	if len(retrieved.Colors) != len(card.Colors) {
		t.Errorf("Expected %d colors, got %d", len(card.Colors), len(retrieved.Colors))
	}

	if retrieved.ImageURIs == nil {
		t.Error("ImageURIs is nil")
	} else if retrieved.ImageURIs.Normal != card.ImageURIs.Normal {
		t.Errorf("Expected image URI %s, got %s", card.ImageURIs.Normal, retrieved.ImageURIs.Normal)
	}
}

func TestGetCardByScryfallID(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	arenaID := 54321
	card := &Card{
		ID:            "scryfall-test-id",
		ArenaID:       &arenaID,
		Name:          "Test Card",
		Colors:        []string{},
		ColorIdentity: []string{},
		Legalities:    map[string]string{},
		CardFaces:     []scryfall.CardFace{},
	}

	// Save card
	err := service.SaveCard(ctx, card)
	if err != nil {
		t.Fatalf("Failed to save card: %v", err)
	}

	// Get card by Scryfall ID
	retrieved, err := service.GetCard(ctx, "scryfall-test-id")
	if err != nil {
		t.Fatalf("Failed to get card: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved card is nil")
	}

	if retrieved.Name != card.Name {
		t.Errorf("Expected name %s, got %s", card.Name, retrieved.Name)
	}
}

func TestUpdateCard(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	arenaID := 99999
	card := &Card{
		ID:            "update-test-id",
		ArenaID:       &arenaID,
		Name:          "Original Name",
		Colors:        []string{"U"},
		ColorIdentity: []string{"U"},
		Legalities:    map[string]string{},
		CardFaces:     []scryfall.CardFace{},
	}

	// Save initial card
	err := service.SaveCard(ctx, card)
	if err != nil {
		t.Fatalf("Failed to save card: %v", err)
	}

	// Update card
	card.Name = "Updated Name"
	card.Colors = []string{"U", "B"}
	err = service.SaveCard(ctx, card)
	if err != nil {
		t.Fatalf("Failed to update card: %v", err)
	}

	// Retrieve updated card
	retrieved, err := service.GetCardByArenaID(ctx, arenaID)
	if err != nil {
		t.Fatalf("Failed to get updated card: %v", err)
	}

	if retrieved.Name != "Updated Name" {
		t.Errorf("Expected updated name 'Updated Name', got %s", retrieved.Name)
	}

	if len(retrieved.Colors) != 2 {
		t.Errorf("Expected 2 colors, got %d", len(retrieved.Colors))
	}
}

func TestSearchCards(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// Create test cards
	cards := []*Card{
		{
			ID:            "search-1",
			Name:          "Lightning Bolt",
			Colors:        []string{"R"},
			ColorIdentity: []string{"R"},
			Legalities:    map[string]string{},
			CardFaces:     []scryfall.CardFace{},
		},
		{
			ID:            "search-2",
			Name:          "Lightning Strike",
			Colors:        []string{"R"},
			ColorIdentity: []string{"R"},
			Legalities:    map[string]string{},
			CardFaces:     []scryfall.CardFace{},
		},
		{
			ID:            "search-3",
			Name:          "Counterspell",
			Colors:        []string{"U"},
			ColorIdentity: []string{"U"},
			Legalities:    map[string]string{},
			CardFaces:     []scryfall.CardFace{},
		},
	}

	// Save cards
	for _, card := range cards {
		err := service.SaveCard(ctx, card)
		if err != nil {
			t.Fatalf("Failed to save card: %v", err)
		}
	}

	// Search for "Lightning"
	results, err := service.SearchCards(ctx, "Lightning")
	if err != nil {
		t.Fatalf("Failed to search cards: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	// Verify results contain expected cards
	found := make(map[string]bool)
	for _, card := range results {
		found[card.Name] = true
	}

	if !found["Lightning Bolt"] {
		t.Error("Expected to find 'Lightning Bolt'")
	}

	if !found["Lightning Strike"] {
		t.Error("Expected to find 'Lightning Strike'")
	}
}

func TestGetCardsBySet(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// Create test cards in same set
	cards := []*Card{
		{
			ID:              "set-1",
			Name:            "Card A",
			SetCode:         "blb",
			CollectorNumber: "001",
			Colors:          []string{},
			ColorIdentity:   []string{},
			Legalities:      map[string]string{},
			CardFaces:       []scryfall.CardFace{},
		},
		{
			ID:              "set-2",
			Name:            "Card B",
			SetCode:         "blb",
			CollectorNumber: "002",
			Colors:          []string{},
			ColorIdentity:   []string{},
			Legalities:      map[string]string{},
			CardFaces:       []scryfall.CardFace{},
		},
		{
			ID:              "set-3",
			Name:            "Card C",
			SetCode:         "woe",
			CollectorNumber: "001",
			Colors:          []string{},
			ColorIdentity:   []string{},
			Legalities:      map[string]string{},
			CardFaces:       []scryfall.CardFace{},
		},
	}

	// Save cards
	for _, card := range cards {
		err := service.SaveCard(ctx, card)
		if err != nil {
			t.Fatalf("Failed to save card: %v", err)
		}
	}

	// Get cards for BLB set
	results, err := service.GetCardsBySet(ctx, "blb")
	if err != nil {
		t.Fatalf("Failed to get cards by set: %v", err)
	}

	if len(results) != 2 {
		t.Errorf("Expected 2 cards in BLB, got %d", len(results))
	}

	// Verify ordering by collector number
	if results[0].CollectorNumber != "001" {
		t.Errorf("Expected first card collector number '001', got %s", results[0].CollectorNumber)
	}
}

func TestGetStaleCards(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// Create test card
	arenaID := 77777
	card := &Card{
		ID:            "stale-test",
		ArenaID:       &arenaID,
		Name:          "Old Card",
		Colors:        []string{},
		ColorIdentity: []string{},
		Legalities:    map[string]string{},
		CardFaces:     []scryfall.CardFace{},
	}

	err := service.SaveCard(ctx, card)
	if err != nil {
		t.Fatalf("Failed to save card: %v", err)
	}

	// Check for stale cards (older than 1 second)
	time.Sleep(1100 * time.Millisecond)

	stale, err := service.GetStaleCards(ctx, 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to get stale cards: %v", err)
	}

	if len(stale) == 0 {
		t.Error("Expected at least 1 stale card")
	}

	// Fresh cards (older than 1 hour should return nothing)
	fresh, err := service.GetStaleCards(ctx, 1*time.Hour)
	if err != nil {
		t.Fatalf("Failed to get fresh cards: %v", err)
	}

	if len(fresh) != 0 {
		t.Errorf("Expected 0 fresh cards, got %d", len(fresh))
	}
}

func TestSaveAndGetSet(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	set := &Set{
		Code:       "blb",
		Name:       "Bloomburrow",
		ReleasedAt: "2024-08-02",
		CardCount:  398,
		SetType:    "expansion",
		IconSVGURI: "https://example.com/blb.svg",
	}

	// Save set
	err := service.SaveSet(ctx, set)
	if err != nil {
		t.Fatalf("Failed to save set: %v", err)
	}

	// Get set
	retrieved, err := service.GetSet(ctx, "blb")
	if err != nil {
		t.Fatalf("Failed to get set: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved set is nil")
	}

	if retrieved.Name != set.Name {
		t.Errorf("Expected name %s, got %s", set.Name, retrieved.Name)
	}

	if retrieved.CardCount != set.CardCount {
		t.Errorf("Expected card count %d, got %d", set.CardCount, retrieved.CardCount)
	}
}

func TestGetAllSets(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// Create test sets
	sets := []*Set{
		{Code: "blb", Name: "Bloomburrow", ReleasedAt: "2024-08-02", CardCount: 398, SetType: "expansion"},
		{Code: "woe", Name: "Wilds of Eldraine", ReleasedAt: "2023-09-08", CardCount: 276, SetType: "expansion"},
		{Code: "mkm", Name: "Murders at Karlov Manor", ReleasedAt: "2024-02-09", CardCount: 289, SetType: "expansion"},
	}

	// Save sets
	for _, set := range sets {
		err := service.SaveSet(ctx, set)
		if err != nil {
			t.Fatalf("Failed to save set: %v", err)
		}
	}

	// Get all sets
	all, err := service.GetAllSets(ctx)
	if err != nil {
		t.Fatalf("Failed to get all sets: %v", err)
	}

	if len(all) < 3 {
		t.Errorf("Expected at least 3 sets, got %d", len(all))
	}

	// Verify ordering (newest first)
	if all[0].Code != "blb" {
		t.Errorf("Expected first set to be BLB (newest), got %s", all[0].Code)
	}
}

func TestGetStaleSets(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	set := &Set{
		Code:       "old",
		Name:       "Old Set",
		CardCount:  100,
		SetType:    "expansion",
		ReleasedAt: "2020-01-01",
	}

	err := service.SaveSet(ctx, set)
	if err != nil {
		t.Fatalf("Failed to save set: %v", err)
	}

	// Check for stale sets (older than 1 second)
	time.Sleep(1100 * time.Millisecond)

	stale, err := service.GetStaleSets(ctx, 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to get stale sets: %v", err)
	}

	if len(stale) == 0 {
		t.Error("Expected at least 1 stale set")
	}
}

func TestCardWithDFC(t *testing.T) {
	service := setupTestService(t)
	ctx := context.Background()

	// Create a DFC card
	arenaID := 11111
	card := &Card{
		ID:            "dfc-test",
		ArenaID:       &arenaID,
		Name:          "Test DFC",
		Layout:        "modal_dfc",
		Colors:        []string{"U"},
		ColorIdentity: []string{"U"},
		Legalities:    map[string]string{},
		CardFaces: []scryfall.CardFace{
			{
				Name:     "Front Face",
				ManaCost: "{1}{U}",
				TypeLine: "Creature",
			},
			{
				Name:     "Back Face",
				TypeLine: "Land",
			},
		},
	}

	// Save card
	err := service.SaveCard(ctx, card)
	if err != nil {
		t.Fatalf("Failed to save DFC card: %v", err)
	}

	// Retrieve card
	retrieved, err := service.GetCardByArenaID(ctx, arenaID)
	if err != nil {
		t.Fatalf("Failed to get DFC card: %v", err)
	}

	if retrieved == nil {
		t.Fatal("Retrieved DFC card is nil")
	}

	if len(retrieved.CardFaces) != 2 {
		t.Errorf("Expected 2 card faces, got %d", len(retrieved.CardFaces))
	}

	if retrieved.CardFaces[0].Name != "Front Face" {
		t.Errorf("Expected front face name 'Front Face', got %s", retrieved.CardFaces[0].Name)
	}
}
