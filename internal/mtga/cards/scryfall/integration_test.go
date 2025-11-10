//go:build integration
// +build integration

package scryfall

import (
	"context"
	"testing"
	"time"
)

// These tests make real API calls to Scryfall.
// Run with: go test -tags=integration

func TestIntegration_GetCardByArenaID(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Test with a known Arena ID (Lightning Bolt from a recent set)
	// Arena ID 89019 is from Bloomburrow (BLB)
	card, err := client.GetCardByArenaID(ctx, 89019)
	if err != nil {
		t.Fatalf("GetCardByArenaID failed: %v", err)
	}

	if card == nil {
		t.Fatal("Card is nil")
	}

	if card.ArenaID == nil || *card.ArenaID != 89019 {
		t.Errorf("Expected Arena ID 89019, got %v", card.ArenaID)
	}

	if card.Name == "" {
		t.Error("Card name is empty")
	}

	if card.SetCode == "" {
		t.Error("Set code is empty")
	}

	t.Logf("Successfully retrieved card: %s (Arena ID: %d)", card.Name, *card.ArenaID)
}

func TestIntegration_GetCard(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Test with a known Scryfall ID
	// This is Lightning Bolt from Alpha (always exists)
	card, err := client.GetCard(ctx, "1d72ab16-c3dd-4b92-ba1f-7a490a61f36f")
	if err != nil {
		t.Fatalf("GetCard failed: %v", err)
	}

	if card == nil {
		t.Fatal("Card is nil")
	}

	if card.ID != "1d72ab16-c3dd-4b92-ba1f-7a490a61f36f" {
		t.Errorf("Unexpected card ID: %s", card.ID)
	}

	if card.Name != "Lightning Bolt" {
		t.Errorf("Expected 'Lightning Bolt', got '%s'", card.Name)
	}

	t.Logf("Successfully retrieved card: %s from set %s", card.Name, card.SetCode)
}

func TestIntegration_GetSet(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Test with Bloomburrow (BLB) set code
	set, err := client.GetSet(ctx, "blb")
	if err != nil {
		t.Fatalf("GetSet failed: %v", err)
	}

	if set == nil {
		t.Fatal("Set is nil")
	}

	if set.Code != "blb" {
		t.Errorf("Expected set code 'blb', got '%s'", set.Code)
	}

	if set.Name == "" {
		t.Error("Set name is empty")
	}

	if set.CardCount == 0 {
		t.Error("Set has 0 cards")
	}

	t.Logf("Successfully retrieved set: %s (%d cards)", set.Name, set.CardCount)
}

func TestIntegration_SearchCards(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Search for Lightning Bolt
	result, err := client.SearchCards(ctx, "!\"Lightning Bolt\"")
	if err != nil {
		t.Fatalf("SearchCards failed: %v", err)
	}

	if result == nil {
		t.Fatal("Search result is nil")
	}

	if result.TotalCards == 0 {
		t.Error("Search returned no cards")
	}

	if len(result.Data) == 0 {
		t.Error("Search result data is empty")
	}

	// Verify first card is Lightning Bolt
	if result.Data[0].Name != "Lightning Bolt" {
		t.Errorf("Expected first card to be 'Lightning Bolt', got '%s'", result.Data[0].Name)
	}

	t.Logf("Successfully searched and found %d results", result.TotalCards)
}

func TestIntegration_GetBulkData(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	bulkData, err := client.GetBulkData(ctx)
	if err != nil {
		t.Fatalf("GetBulkData failed: %v", err)
	}

	if bulkData == nil {
		t.Fatal("Bulk data is nil")
	}

	if len(bulkData.Data) == 0 {
		t.Error("No bulk data files available")
	}

	// Find the "Default Cards" bulk data
	var defaultCards *BulkData
	for i := range bulkData.Data {
		if bulkData.Data[i].Type == "default_cards" {
			defaultCards = &bulkData.Data[i]
			break
		}
	}

	if defaultCards == nil {
		t.Fatal("Could not find 'default_cards' bulk data")
	}

	if defaultCards.DownloadURI == "" {
		t.Error("Bulk data download URI is empty")
	}

	if defaultCards.CompressedSize == 0 {
		t.Error("Bulk data compressed size is 0")
	}

	t.Logf("Successfully retrieved bulk data: %s (%.2f MB)",
		defaultCards.Name,
		float64(defaultCards.CompressedSize)/(1024*1024))
}

func TestIntegration_GetSets(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	sets, err := client.GetSets(ctx)
	if err != nil {
		t.Fatalf("GetSets failed: %v", err)
	}

	if sets == nil {
		t.Fatal("Sets list is nil")
	}

	if len(sets.Data) == 0 {
		t.Error("No sets returned")
	}

	// Should have hundreds of sets
	if len(sets.Data) < 100 {
		t.Errorf("Expected at least 100 sets, got %d", len(sets.Data))
	}

	t.Logf("Successfully retrieved %d sets", len(sets.Data))
}

func TestIntegration_RateLimitingRealAPI(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Make 5 requests and measure time
	start := time.Now()
	for i := 0; i < 5; i++ {
		_, err := client.GetSet(ctx, "blb")
		if err != nil {
			t.Fatalf("Request %d failed: %v", i+1, err)
		}
	}
	elapsed := time.Since(start)

	// Should take at least 400ms (4 delays of 100ms each between 5 requests)
	minDuration := 400 * time.Millisecond
	if elapsed < minDuration {
		t.Errorf("Rate limiting may not be working: completed 5 requests in %v (expected >= %v)", elapsed, minDuration)
	}

	t.Logf("Rate limiting verified: 5 requests took %v", elapsed)
}

func TestIntegration_NotFoundError(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Try to get a card with an invalid Arena ID
	_, err := client.GetCardByArenaID(ctx, 99999999)
	if err == nil {
		t.Fatal("Expected error for invalid Arena ID, got nil")
	}

	if !IsNotFound(err) {
		t.Errorf("Expected NotFoundError, got: %T - %v", err, err)
	}

	t.Logf("NotFoundError correctly returned for invalid Arena ID")
}

func TestIntegration_CardFields(t *testing.T) {
	client := NewClient()
	ctx := context.Background()

	// Get a well-known card and verify all important fields are populated
	card, err := client.GetCardByArenaID(ctx, 89019)
	if err != nil {
		t.Fatalf("GetCardByArenaID failed: %v", err)
	}

	// Verify core fields
	if card.ID == "" {
		t.Error("Card ID is empty")
	}
	if card.Name == "" {
		t.Error("Card name is empty")
	}
	if card.TypeLine == "" {
		t.Error("Type line is empty")
	}
	if card.SetCode == "" {
		t.Error("Set code is empty")
	}
	if card.Rarity == "" {
		t.Error("Rarity is empty")
	}

	// Verify image URIs
	if card.ImageURIs == nil {
		t.Error("Image URIs are nil")
	} else {
		if card.ImageURIs.Small == "" {
			t.Error("Small image URI is empty")
		}
		if card.ImageURIs.Normal == "" {
			t.Error("Normal image URI is empty")
		}
		if card.ImageURIs.Large == "" {
			t.Error("Large image URI is empty")
		}
	}

	// Verify legalities
	if card.Legalities.Standard == "" {
		t.Error("Standard legality is empty")
	}

	t.Logf("Card fields verified for: %s", card.Name)
}
