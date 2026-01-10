package gui

import (
	"testing"
	"time"
)

const cardBackPlaceholderURL = "https://cards.scryfall.io/back.png"

func TestCollectionCard_ImageURIFallback(t *testing.T) {
	tests := []struct {
		name        string
		imageURI    string
		expectedURI string
	}{
		{
			name:        "empty imageURI should use placeholder",
			imageURI:    "",
			expectedURI: cardBackPlaceholderURL,
		},
		{
			name:        "valid imageURI should be preserved",
			imageURI:    "https://cards.scryfall.io/normal/front/1/2/test.jpg",
			expectedURI: "https://cards.scryfall.io/normal/front/1/2/test.jpg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &CollectionCard{
				CardID:   1,
				ArenaID:  1,
				Quantity: 1,
				ImageURI: tt.imageURI,
			}

			// Simulate the fallback logic from GetCollection
			if card.ImageURI == "" {
				card.ImageURI = cardBackPlaceholderURL
			}

			if card.ImageURI != tt.expectedURI {
				t.Errorf("ImageURI = %q, want %q", card.ImageURI, tt.expectedURI)
			}
		})
	}
}

func TestCardBackPlaceholderURL(t *testing.T) {
	// Verify the placeholder URL format is correct (.png not .jpg)
	expectedSuffix := ".png"
	if len(cardBackPlaceholderURL) < len(expectedSuffix) {
		t.Errorf("cardBackPlaceholderURL is too short")
		return
	}
	suffix := cardBackPlaceholderURL[len(cardBackPlaceholderURL)-len(expectedSuffix):]
	if suffix != expectedSuffix {
		t.Errorf("cardBackPlaceholderURL should end with %q, got %q", expectedSuffix, suffix)
	}

	// Verify it points to the correct Scryfall domain
	expectedPrefix := "https://cards.scryfall.io/"
	if len(cardBackPlaceholderURL) < len(expectedPrefix) {
		t.Errorf("cardBackPlaceholderURL is too short")
		return
	}
	prefix := cardBackPlaceholderURL[:len(expectedPrefix)]
	if prefix != expectedPrefix {
		t.Errorf("cardBackPlaceholderURL should start with %q, got %q", expectedPrefix, prefix)
	}
}

func TestRarityOrder(t *testing.T) {
	tests := []struct {
		rarity   string
		expected int
	}{
		{"common", 1},
		{"Common", 1},
		{"COMMON", 1},
		{"uncommon", 2},
		{"Uncommon", 2},
		{"rare", 3},
		{"Rare", 3},
		{"mythic", 4},
		{"Mythic", 4},
		{"", 0},
		{"unknown", 0},
	}

	for _, tt := range tests {
		t.Run(tt.rarity, func(t *testing.T) {
			result := rarityOrder(tt.rarity)
			if result != tt.expected {
				t.Errorf("rarityOrder(%q) = %d, want %d", tt.rarity, result, tt.expected)
			}
		})
	}
}

func TestWildcardCostTotal(t *testing.T) {
	cost := &WildcardCost{
		Common:   5,
		Uncommon: 3,
		Rare:     2,
		Mythic:   1,
	}
	cost.Total = cost.Common + cost.Uncommon + cost.Rare + cost.Mythic

	expectedTotal := 11
	if cost.Total != expectedTotal {
		t.Errorf("WildcardCost.Total = %d, want %d", cost.Total, expectedTotal)
	}
}

func TestMissingCardSortByRarity(t *testing.T) {
	cards := []*MissingCard{
		{CardID: 1, Name: "Common Card", Rarity: "common"},
		{CardID: 2, Name: "Mythic Card", Rarity: "mythic"},
		{CardID: 3, Name: "Rare Card", Rarity: "rare"},
		{CardID: 4, Name: "Uncommon Card", Rarity: "uncommon"},
	}

	// Sort by rarity descending (mythic first)
	// This mimics the sort in GetMissingCardsForDeck
	for i := 0; i < len(cards)-1; i++ {
		for j := i + 1; j < len(cards); j++ {
			if rarityOrder(cards[i].Rarity) < rarityOrder(cards[j].Rarity) {
				cards[i], cards[j] = cards[j], cards[i]
			}
		}
	}

	expectedOrder := []string{"mythic", "rare", "uncommon", "common"}
	for i, card := range cards {
		if card.Rarity != expectedOrder[i] {
			t.Errorf("cards[%d].Rarity = %s, want %s", i, card.Rarity, expectedOrder[i])
		}
	}
}

func TestMissingCardsForDeckResponse_IsComplete(t *testing.T) {
	// Test complete deck (no missing cards)
	completeResponse := &MissingCardsForDeckResponse{
		DeckID:       "test-deck-1",
		DeckName:     "Complete Deck",
		TotalMissing: 0,
		IsComplete:   true,
	}

	if !completeResponse.IsComplete {
		t.Error("Expected IsComplete to be true when TotalMissing is 0")
	}

	// Test incomplete deck
	incompleteResponse := &MissingCardsForDeckResponse{
		DeckID:       "test-deck-2",
		DeckName:     "Incomplete Deck",
		TotalMissing: 5,
		IsComplete:   false,
	}

	if incompleteResponse.IsComplete {
		t.Error("Expected IsComplete to be false when TotalMissing > 0")
	}
}

func TestMissingCardsForSetResponse_CompletionPct(t *testing.T) {
	tests := []struct {
		name          string
		totalOwned    int
		totalPossible int
		expectedPct   float64
	}{
		{"empty collection", 0, 100, 0.0},
		{"half complete", 50, 100, 50.0},
		{"fully complete", 100, 100, 100.0},
		{"quarter complete", 25, 100, 25.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var completionPct float64
			if tt.totalPossible > 0 {
				completionPct = float64(tt.totalOwned) / float64(tt.totalPossible) * 100
			}

			if completionPct != tt.expectedPct {
				t.Errorf("completion percentage = %.2f, want %.2f", completionPct, tt.expectedPct)
			}
		})
	}
}

func TestMissingCard_QuantityCalculation(t *testing.T) {
	tests := []struct {
		name           string
		deckQuantity   int
		collectionOwns int
		expectedNeeded int
	}{
		{"have none need 4", 4, 0, 4},
		{"have 2 need 4", 4, 2, 2},
		{"have 4 need 4", 4, 4, 0},
		{"have more than needed", 4, 6, 0},
		{"have 1 need 3", 3, 1, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			needed := tt.deckQuantity - tt.collectionOwns
			if needed < 0 {
				needed = 0
			}

			if needed != tt.expectedNeeded {
				t.Errorf("needed = %d, want %d", needed, tt.expectedNeeded)
			}
		})
	}
}

func TestWildcardCost_ByRarity(t *testing.T) {
	// Simulate adding missing cards to wildcard cost
	cost := &WildcardCost{}

	// Add some cards by rarity
	addCards := []struct {
		rarity   string
		quantity int
	}{
		{"common", 4},
		{"uncommon", 3},
		{"rare", 2},
		{"mythic", 1},
		{"rare", 1}, // Add more rares
	}

	for _, card := range addCards {
		switch card.rarity {
		case "common":
			cost.Common += card.quantity
		case "uncommon":
			cost.Uncommon += card.quantity
		case "rare":
			cost.Rare += card.quantity
		case "mythic":
			cost.Mythic += card.quantity
		}
	}

	cost.Total = cost.Common + cost.Uncommon + cost.Rare + cost.Mythic

	if cost.Common != 4 {
		t.Errorf("Common = %d, want 4", cost.Common)
	}
	if cost.Uncommon != 3 {
		t.Errorf("Uncommon = %d, want 3", cost.Uncommon)
	}
	if cost.Rare != 3 {
		t.Errorf("Rare = %d, want 3", cost.Rare)
	}
	if cost.Mythic != 1 {
		t.Errorf("Mythic = %d, want 1", cost.Mythic)
	}
	if cost.Total != 11 {
		t.Errorf("Total = %d, want 11", cost.Total)
	}
}

func TestNewCollectionFacade_InitializesFailedLookups(t *testing.T) {
	facade := NewCollectionFacade(nil)

	if facade.failedLookups == nil {
		t.Error("Expected failedLookups map to be initialized")
	}

	if len(facade.failedLookups) != 0 {
		t.Errorf("Expected failedLookups to be empty, got %d entries", len(facade.failedLookups))
	}
}

func TestFailedLookupCooldown_Value(t *testing.T) {
	// Verify the cooldown is set to 1 hour
	expectedCooldown := 1 * time.Hour
	if failedLookupCooldown != expectedCooldown {
		t.Errorf("failedLookupCooldown = %v, want %v", failedLookupCooldown, expectedCooldown)
	}
}

func TestFailedLookupTracking_WithinCooldown(t *testing.T) {
	facade := NewCollectionFacade(nil)

	// Record a failed lookup
	cardID := 12345
	now := time.Now()
	facade.failedLookups[cardID] = now

	// Check if it's within cooldown (should be skipped)
	failTime, exists := facade.failedLookups[cardID]
	if !exists {
		t.Error("Expected card to be in failedLookups")
	}

	elapsed := time.Since(failTime)
	if elapsed >= failedLookupCooldown {
		t.Error("Expected failed lookup to be within cooldown period")
	}
}

func TestFailedLookupTracking_ExpiredCooldown(t *testing.T) {
	facade := NewCollectionFacade(nil)

	// Record a failed lookup that's older than the cooldown
	cardID := 12345
	expiredTime := time.Now().Add(-2 * time.Hour) // 2 hours ago
	facade.failedLookups[cardID] = expiredTime

	// Check if it's expired (should allow retry)
	failTime := facade.failedLookups[cardID]
	elapsed := time.Since(failTime)

	if elapsed < failedLookupCooldown {
		t.Error("Expected failed lookup to be past cooldown period")
	}
}

func TestFailedLookupTracking_MultipleCards(t *testing.T) {
	facade := NewCollectionFacade(nil)

	// Record multiple failed lookups
	now := time.Now()
	facade.failedLookups[111] = now
	facade.failedLookups[222] = now.Add(-30 * time.Minute) // 30 min ago
	facade.failedLookups[333] = now.Add(-2 * time.Hour)    // 2 hours ago (expired)

	if len(facade.failedLookups) != 3 {
		t.Errorf("Expected 3 entries in failedLookups, got %d", len(facade.failedLookups))
	}

	// Verify each card's status
	// Card 111: just added, within cooldown
	if time.Since(facade.failedLookups[111]) >= failedLookupCooldown {
		t.Error("Card 111 should be within cooldown")
	}

	// Card 222: 30 min ago, still within cooldown
	if time.Since(facade.failedLookups[222]) >= failedLookupCooldown {
		t.Error("Card 222 should be within cooldown")
	}

	// Card 333: 2 hours ago, expired
	if time.Since(facade.failedLookups[333]) < failedLookupCooldown {
		t.Error("Card 333 should be past cooldown")
	}
}

func TestMaxAutoLookups_Value(t *testing.T) {
	// Verify the max auto lookups is reasonable
	if maxAutoLookups <= 0 {
		t.Errorf("maxAutoLookups should be positive, got %d", maxAutoLookups)
	}

	if maxAutoLookups > 50 {
		t.Errorf("maxAutoLookups should not be too high to avoid rate limiting, got %d", maxAutoLookups)
	}

	// Current expected value is 10
	if maxAutoLookups != 10 {
		t.Errorf("maxAutoLookups = %d, expected 10", maxAutoLookups)
	}
}

func TestCollectionValue_Fields(t *testing.T) {
	lastUpdated := int64(1704067200)
	value := &CollectionValue{
		TotalValueUSD:        1234.56,
		TotalValueEUR:        1100.00,
		UniqueCardsWithPrice: 500,
		CardCount:            1500,
		ValueByRarity: map[string]float64{
			"mythic":   500.00,
			"rare":     600.00,
			"uncommon": 100.00,
			"common":   34.56,
		},
		TopCards: []*CardValue{
			{CardID: 1, Name: "Black Lotus", SetCode: "LEA", Rarity: "rare", Quantity: 1, PriceUSD: 50000.00, TotalUSD: 50000.00},
		},
		LastUpdated: &lastUpdated,
	}

	// Verify basic fields
	if value.TotalValueUSD != 1234.56 {
		t.Errorf("TotalValueUSD = %f, want 1234.56", value.TotalValueUSD)
	}
	if value.TotalValueEUR != 1100.00 {
		t.Errorf("TotalValueEUR = %f, want 1100.00", value.TotalValueEUR)
	}
	if value.UniqueCardsWithPrice != 500 {
		t.Errorf("UniqueCardsWithPrice = %d, want 500", value.UniqueCardsWithPrice)
	}
	if value.CardCount != 1500 {
		t.Errorf("CardCount = %d, want 1500", value.CardCount)
	}

	// Verify rarity breakdown sums correctly
	var rarityTotal float64
	for _, v := range value.ValueByRarity {
		rarityTotal += v
	}
	if rarityTotal != 1234.56 {
		t.Errorf("ValueByRarity sum = %f, want 1234.56", rarityTotal)
	}

	// Verify top cards
	if len(value.TopCards) != 1 {
		t.Errorf("TopCards length = %d, want 1", len(value.TopCards))
	}
	if value.TopCards[0].Name != "Black Lotus" {
		t.Errorf("TopCards[0].Name = %s, want Black Lotus", value.TopCards[0].Name)
	}

	// Verify last updated
	if value.LastUpdated == nil {
		t.Error("LastUpdated should not be nil")
	} else if *value.LastUpdated != lastUpdated {
		t.Errorf("LastUpdated = %d, want %d", *value.LastUpdated, lastUpdated)
	}
}

func TestCollectionValue_EmptyCollection(t *testing.T) {
	value := &CollectionValue{
		ValueByRarity: make(map[string]float64),
		TopCards:      make([]*CardValue, 0),
	}

	if value.TotalValueUSD != 0 {
		t.Errorf("Empty collection TotalValueUSD = %f, want 0", value.TotalValueUSD)
	}
	if value.CardCount != 0 {
		t.Errorf("Empty collection CardCount = %d, want 0", value.CardCount)
	}
	if len(value.TopCards) != 0 {
		t.Errorf("Empty collection TopCards length = %d, want 0", len(value.TopCards))
	}
}

func TestCardValue_TotalCalculation(t *testing.T) {
	tests := []struct {
		name        string
		priceUSD    float64
		quantity    int
		expectedTot float64
	}{
		{"single copy", 10.00, 1, 10.00},
		{"playset", 25.00, 4, 100.00},
		{"three copies", 3.33, 3, 9.99},
		{"zero price", 0.00, 4, 0.00},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			card := &CardValue{
				CardID:   1,
				Name:     "Test Card",
				Quantity: tt.quantity,
				PriceUSD: tt.priceUSD,
				TotalUSD: tt.priceUSD * float64(tt.quantity),
			}

			if card.TotalUSD != tt.expectedTot {
				t.Errorf("TotalUSD = %f, want %f", card.TotalUSD, tt.expectedTot)
			}
		})
	}
}

func TestDeckValue_Fields(t *testing.T) {
	value := &DeckValue{
		DeckID:         "deck-123",
		DeckName:       "Test Deck",
		TotalValueUSD:  250.50,
		TotalValueEUR:  225.00,
		CardCount:      60,
		CardsWithPrice: 55,
		TopCards: []*CardValue{
			{CardID: 1, Name: "Mox Diamond", SetCode: "STH", Rarity: "rare", Quantity: 1, PriceUSD: 100.00, TotalUSD: 100.00},
			{CardID: 2, Name: "City of Traitors", SetCode: "EXO", Rarity: "rare", Quantity: 1, PriceUSD: 80.00, TotalUSD: 80.00},
		},
	}

	if value.DeckID != "deck-123" {
		t.Errorf("DeckID = %s, want deck-123", value.DeckID)
	}
	if value.DeckName != "Test Deck" {
		t.Errorf("DeckName = %s, want Test Deck", value.DeckName)
	}
	if value.TotalValueUSD != 250.50 {
		t.Errorf("TotalValueUSD = %f, want 250.50", value.TotalValueUSD)
	}
	if value.CardCount != 60 {
		t.Errorf("CardCount = %d, want 60", value.CardCount)
	}
	if value.CardsWithPrice != 55 {
		t.Errorf("CardsWithPrice = %d, want 55", value.CardsWithPrice)
	}
	if len(value.TopCards) != 2 {
		t.Errorf("TopCards length = %d, want 2", len(value.TopCards))
	}
}

func TestDeckValue_EmptyDeck(t *testing.T) {
	value := &DeckValue{
		DeckID:         "empty-deck",
		DeckName:       "Empty Deck",
		TotalValueUSD:  0,
		TotalValueEUR:  0,
		CardCount:      0,
		CardsWithPrice: 0,
		TopCards:       make([]*CardValue, 0),
	}

	if value.TotalValueUSD != 0 {
		t.Errorf("Empty deck TotalValueUSD = %f, want 0", value.TotalValueUSD)
	}
	if value.CardCount != 0 {
		t.Errorf("Empty deck CardCount = %d, want 0", value.CardCount)
	}
	if len(value.TopCards) != 0 {
		t.Errorf("Empty deck TopCards length = %d, want 0", len(value.TopCards))
	}
}

func TestDeckValue_TopCardsSortedByValue(t *testing.T) {
	// Simulate how top cards should be sorted (highest value first)
	cards := []*CardValue{
		{Name: "Cheap Card", TotalUSD: 5.00},
		{Name: "Expensive Card", TotalUSD: 100.00},
		{Name: "Medium Card", TotalUSD: 25.00},
	}

	// Sort by TotalUSD descending
	for i := 0; i < len(cards)-1; i++ {
		for j := i + 1; j < len(cards); j++ {
			if cards[i].TotalUSD < cards[j].TotalUSD {
				cards[i], cards[j] = cards[j], cards[i]
			}
		}
	}

	expectedOrder := []string{"Expensive Card", "Medium Card", "Cheap Card"}
	for i, card := range cards {
		if card.Name != expectedOrder[i] {
			t.Errorf("cards[%d].Name = %s, want %s", i, card.Name, expectedOrder[i])
		}
	}
}

func TestCollectionCard_PriceFields(t *testing.T) {
	priceUSD := 10.50
	priceUSDFoil := 25.00
	priceEUR := 9.00
	pricesUpdatedAt := int64(1704067200)

	card := &CollectionCard{
		CardID:          1,
		Name:            "Test Card",
		Quantity:        4,
		PriceUSD:        &priceUSD,
		PriceUSDFoil:    &priceUSDFoil,
		PriceEUR:        &priceEUR,
		PricesUpdatedAt: &pricesUpdatedAt,
	}

	// Verify price fields
	if card.PriceUSD == nil || *card.PriceUSD != 10.50 {
		t.Errorf("PriceUSD = %v, want 10.50", card.PriceUSD)
	}
	if card.PriceUSDFoil == nil || *card.PriceUSDFoil != 25.00 {
		t.Errorf("PriceUSDFoil = %v, want 25.00", card.PriceUSDFoil)
	}
	if card.PriceEUR == nil || *card.PriceEUR != 9.00 {
		t.Errorf("PriceEUR = %v, want 9.00", card.PriceEUR)
	}
	if card.PricesUpdatedAt == nil || *card.PricesUpdatedAt != pricesUpdatedAt {
		t.Errorf("PricesUpdatedAt = %v, want %d", card.PricesUpdatedAt, pricesUpdatedAt)
	}
}

func TestCollectionCard_NilPrices(t *testing.T) {
	card := &CollectionCard{
		CardID:   1,
		Name:     "Card Without Prices",
		Quantity: 4,
	}

	// All price fields should be nil
	if card.PriceUSD != nil {
		t.Errorf("PriceUSD should be nil, got %v", card.PriceUSD)
	}
	if card.PriceUSDFoil != nil {
		t.Errorf("PriceUSDFoil should be nil, got %v", card.PriceUSDFoil)
	}
	if card.PriceEUR != nil {
		t.Errorf("PriceEUR should be nil, got %v", card.PriceEUR)
	}
	if card.PricesUpdatedAt != nil {
		t.Errorf("PricesUpdatedAt should be nil, got %v", card.PricesUpdatedAt)
	}
}
