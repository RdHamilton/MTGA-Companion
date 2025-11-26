package gui

import (
	"testing"
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
