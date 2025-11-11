package storage

import (
	"context"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

func TestSaveCardRatings(t *testing.T) {
	svc := setupTestService(t)

	ctx := context.Background()

	// Prepare test data
	ratings := []seventeenlands.CardRating{
		{
			Name:        "Test Card 1",
			MTGAID:      12345,
			GIHWR:       0.58,
			OHWR:        0.55,
			GPWR:        0.60,
			GDWR:        0.62,
			IHDWR:       0.59,
			GIHWRDelta:  0.02,
			OHWRDelta:   0.01,
			GDWRDelta:   0.03,
			IHDWRDelta:  0.02,
			ALSA:        3.5,
			ATA:         4.2,
			GIH:         1000,
			OH:          500,
			GP:          1200,
			GD:          800,
			IHD:         600,
			GamesPlayed: 1000,
			NumberDecks: 200,
		},
		{
			Name:        "Test Card 2",
			MTGAID:      67890,
			GIHWR:       0.45,
			OHWR:        0.42,
			GPWR:        0.48,
			ALSA:        6.8,
			ATA:         8.1,
			GIH:         500,
			OH:          250,
			GP:          600,
			GamesPlayed: 500,
			NumberDecks: 100,
		},
	}

	// Test initial save
	err := svc.SaveCardRatings(ctx, ratings, "BLB", "PremierDraft", "", "2024-08-01", "2024-09-01")
	if err != nil {
		t.Fatalf("SaveCardRatings failed: %v", err)
	}

	// Verify saved data
	saved, err := svc.GetCardRating(ctx, 12345, "BLB", "PremierDraft", "")
	if err != nil {
		t.Fatalf("GetCardRating failed: %v", err)
	}
	if saved == nil {
		t.Fatal("Expected rating to be saved, got nil")
	}
	if saved.ArenaID != 12345 {
		t.Errorf("Expected arena_id 12345, got %d", saved.ArenaID)
	}
	if saved.GIHWR != 0.58 {
		t.Errorf("Expected GIHWR 0.58, got %f", saved.GIHWR)
	}

	// Test UPSERT (update existing)
	updatedRatings := []seventeenlands.CardRating{
		{
			MTGAID:      12345,
			GIHWR:       0.62, // Changed
			OHWR:        0.55,
			GPWR:        0.60,
			ALSA:        3.5,
			ATA:         4.2,
			GIH:         1500, // Changed
			OH:          500,
			GP:          1200,
			GamesPlayed: 1500,
			NumberDecks: 300,
		},
	}

	// Allow small delay to ensure different last_updated
	time.Sleep(10 * time.Millisecond)

	err = svc.SaveCardRatings(ctx, updatedRatings, "BLB", "PremierDraft", "", "2024-08-01", "2024-09-01")
	if err != nil {
		t.Fatalf("SaveCardRatings update failed: %v", err)
	}

	// Verify updated data
	updated, err := svc.GetCardRating(ctx, 12345, "BLB", "PremierDraft", "")
	if err != nil {
		t.Fatalf("GetCardRating after update failed: %v", err)
	}
	if updated.GIHWR != 0.62 {
		t.Errorf("Expected updated GIHWR 0.62, got %f", updated.GIHWR)
	}
	if updated.GIH != 1500 {
		t.Errorf("Expected updated GIH 1500, got %d", updated.GIH)
	}
}

func TestSaveCardRatings_EmptySlice(t *testing.T) {
	svc := setupTestService(t)

	ctx := context.Background()

	// Test empty ratings slice
	err := svc.SaveCardRatings(ctx, []seventeenlands.CardRating{}, "BLB", "PremierDraft", "", "2024-08-01", "2024-09-01")
	if err != nil {
		t.Errorf("SaveCardRatings with empty slice should not error, got: %v", err)
	}
}

func TestSaveCardRatings_SkipNoArenaID(t *testing.T) {
	svc := setupTestService(t)

	ctx := context.Background()

	// Test rating without Arena ID (should be skipped)
	ratings := []seventeenlands.CardRating{
		{
			Name:        "Card Without ID",
			MTGAID:      0, // No Arena ID
			GIHWR:       0.50,
			GamesPlayed: 100,
		},
		{
			Name:        "Card With ID",
			MTGAID:      99999,
			GIHWR:       0.55,
			GamesPlayed: 200,
		},
	}

	err := svc.SaveCardRatings(ctx, ratings, "BLB", "PremierDraft", "", "2024-08-01", "2024-09-01")
	if err != nil {
		t.Fatalf("SaveCardRatings failed: %v", err)
	}

	// Verify only card with ID was saved
	saved, err := svc.GetCardRating(ctx, 99999, "BLB", "PremierDraft", "")
	if err != nil {
		t.Fatalf("GetCardRating failed: %v", err)
	}
	if saved == nil {
		t.Fatal("Expected card with ID to be saved")
	}
}

func TestGetCardRating_NotFound(t *testing.T) {
	svc := setupTestService(t)

	ctx := context.Background()

	// Try to get non-existent rating
	rating, err := svc.GetCardRating(ctx, 99999, "BLB", "PremierDraft", "")
	if err != nil {
		t.Fatalf("GetCardRating should not error on not found, got: %v", err)
	}
	if rating != nil {
		t.Error("Expected nil for non-existent rating")
	}
}

func TestGetCardRatingsForSet(t *testing.T) {
	svc := setupTestService(t)

	ctx := context.Background()

	// Save multiple ratings
	ratings := []seventeenlands.CardRating{
		{MTGAID: 1, GIHWR: 0.60, GamesPlayed: 100},
		{MTGAID: 2, GIHWR: 0.55, GamesPlayed: 200},
		{MTGAID: 3, GIHWR: 0.65, GamesPlayed: 150},
	}

	err := svc.SaveCardRatings(ctx, ratings, "BLB", "PremierDraft", "", "2024-08-01", "2024-09-01")
	if err != nil {
		t.Fatalf("SaveCardRatings failed: %v", err)
	}

	// Get all ratings for set
	allRatings, err := svc.GetCardRatingsForSet(ctx, "BLB", "PremierDraft", "")
	if err != nil {
		t.Fatalf("GetCardRatingsForSet failed: %v", err)
	}

	if len(allRatings) != 3 {
		t.Fatalf("Expected 3 ratings, got %d", len(allRatings))
	}

	// Verify ordered by GIHWR DESC
	if allRatings[0].GIHWR < allRatings[1].GIHWR {
		t.Error("Expected ratings ordered by GIHWR DESC")
	}
	if allRatings[0].ArenaID != 3 {
		t.Errorf("Expected highest GIHWR card (ID 3) first, got ID %d", allRatings[0].ArenaID)
	}
}

func TestGetStaleCardRatings(t *testing.T) {
	svc := setupTestService(t)

	ctx := context.Background()

	// Save ratings with different timestamps
	ratings1 := []seventeenlands.CardRating{
		{MTGAID: 1, GIHWR: 0.50, GamesPlayed: 100},
	}
	err := svc.SaveCardRatings(ctx, ratings1, "BLB", "PremierDraft", "", "2024-08-01", "2024-09-01")
	if err != nil {
		t.Fatalf("SaveCardRatings failed: %v", err)
	}

	// Wait for timestamps to differ (SQLite CURRENT_TIMESTAMP has 1-second resolution)
	time.Sleep(2 * time.Second)

	ratings2 := []seventeenlands.CardRating{
		{MTGAID: 2, GIHWR: 0.55, GamesPlayed: 200},
	}
	err = svc.SaveCardRatings(ctx, ratings2, "MKM", "PremierDraft", "", "2024-09-01", "2024-10-01")
	if err != nil {
		t.Fatalf("SaveCardRatings failed: %v", err)
	}

	// Check for stale ratings (older than 1 second - should find first one)
	stale, err := svc.GetStaleCardRatings(ctx, 1*time.Second)
	if err != nil {
		t.Fatalf("GetStaleCardRatings failed: %v", err)
	}

	if len(stale) != 1 {
		t.Fatalf("Expected 1 stale rating, got %d", len(stale))
	}

	if stale[0].Expansion != "BLB" {
		t.Errorf("Expected BLB to be stale, got %s", stale[0].Expansion)
	}

	// Check for very recent data (nothing should be stale)
	recent, err := svc.GetStaleCardRatings(ctx, 10*time.Hour)
	if err != nil {
		t.Fatalf("GetStaleCardRatings failed: %v", err)
	}

	if len(recent) != 0 {
		t.Errorf("Expected no stale ratings for 10 hour threshold, got %d", len(recent))
	}
}

func TestSaveColorRatings(t *testing.T) {
	svc := setupTestService(t)

	ctx := context.Background()

	// Prepare test data
	ratings := []seventeenlands.ColorRating{
		{ColorName: "WU", WinRate: 0.58, GamesPlayed: 1000, NumberDecks: 200},
		{ColorName: "BR", WinRate: 0.52, GamesPlayed: 800, NumberDecks: 150},
		{ColorName: "UBG", WinRate: 0.55, GamesPlayed: 600, NumberDecks: 100},
	}

	err := svc.SaveColorRatings(ctx, ratings, "BLB", "PremierDraft", "2024-08-01", "2024-09-01")
	if err != nil {
		t.Fatalf("SaveColorRatings failed: %v", err)
	}

	// Verify saved data
	saved, err := svc.GetColorRatings(ctx, "BLB", "PremierDraft")
	if err != nil {
		t.Fatalf("GetColorRatings failed: %v", err)
	}

	if len(saved) != 3 {
		t.Fatalf("Expected 3 color ratings, got %d", len(saved))
	}

	// Verify ordered by win rate DESC
	if saved[0].WinRate < saved[1].WinRate {
		t.Error("Expected color ratings ordered by win_rate DESC")
	}

	// Test UPSERT
	updatedRatings := []seventeenlands.ColorRating{
		{ColorName: "WU", WinRate: 0.62, GamesPlayed: 1500, NumberDecks: 300}, // Updated
	}

	time.Sleep(10 * time.Millisecond)

	err = svc.SaveColorRatings(ctx, updatedRatings, "BLB", "PremierDraft", "2024-08-01", "2024-09-01")
	if err != nil {
		t.Fatalf("SaveColorRatings update failed: %v", err)
	}

	updated, err := svc.GetColorRatings(ctx, "BLB", "PremierDraft")
	if err != nil {
		t.Fatalf("GetColorRatings after update failed: %v", err)
	}

	// Find WU rating (normalized to WUBRG order: W before U)
	var wuRating *DraftColorRating
	for _, r := range updated {
		if r.ColorCombination == "WU" {
			wuRating = r
			break
		}
	}

	if wuRating == nil {
		t.Fatal("Expected to find WU rating")
	}
	if wuRating.WinRate != 0.62 {
		t.Errorf("Expected updated win rate 0.62, got %f", wuRating.WinRate)
	}
}

func TestSaveColorRatings_EmptySlice(t *testing.T) {
	svc := setupTestService(t)

	ctx := context.Background()

	err := svc.SaveColorRatings(ctx, []seventeenlands.ColorRating{}, "BLB", "PremierDraft", "2024-08-01", "2024-09-01")
	if err != nil {
		t.Errorf("SaveColorRatings with empty slice should not error, got: %v", err)
	}
}

func TestGetColorRatings_NotFound(t *testing.T) {
	svc := setupTestService(t)

	ctx := context.Background()

	ratings, err := svc.GetColorRatings(ctx, "NONEXISTENT", "PremierDraft")
	if err != nil {
		t.Fatalf("GetColorRatings should not error on not found, got: %v", err)
	}
	if len(ratings) != 0 {
		t.Errorf("Expected empty slice for non-existent expansion, got %d ratings", len(ratings))
	}
}

func TestGetStaleColorRatings(t *testing.T) {
	svc := setupTestService(t)

	ctx := context.Background()

	// Save color ratings with different timestamps
	ratings1 := []seventeenlands.ColorRating{
		{ColorName: "WU", WinRate: 0.55, GamesPlayed: 100},
	}
	err := svc.SaveColorRatings(ctx, ratings1, "BLB", "PremierDraft", "2024-08-01", "2024-09-01")
	if err != nil {
		t.Fatalf("SaveColorRatings failed: %v", err)
	}

	// Wait for timestamps to differ (SQLite CURRENT_TIMESTAMP has 1-second resolution)
	time.Sleep(2 * time.Second)

	ratings2 := []seventeenlands.ColorRating{
		{ColorName: "BR", WinRate: 0.52, GamesPlayed: 200},
	}
	err = svc.SaveColorRatings(ctx, ratings2, "MKM", "QuickDraft", "2024-09-01", "2024-10-01")
	if err != nil {
		t.Fatalf("SaveColorRatings failed: %v", err)
	}

	// Check for stale ratings (older than 1 second - should find first one)
	stale, err := svc.GetStaleColorRatings(ctx, 1*time.Second)
	if err != nil {
		t.Fatalf("GetStaleColorRatings failed: %v", err)
	}

	if len(stale) != 1 {
		t.Fatalf("Expected 1 stale color rating, got %d", len(stale))
	}

	if stale[0].Expansion != "BLB" {
		t.Errorf("Expected BLB to be stale, got %s", stale[0].Expansion)
	}
}

func TestNormalizeColorCombination(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"W", "W"},
		{"U", "U"},
		{"WU", "WU"},
		{"UW", "WU"}, // Should be normalized to WUBRG order
		{"RW", "WR"},
		{"BR", "BR"},
		{"RB", "BR"},
		{"WUG", "WUG"},
		{"GUW", "WUG"},
		{"UGW", "WUG"},
		{"WUBRG", "WUBRG"},
		{"GRUBW", "WUBRG"},
		{"", ""}, // Empty should return empty
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeColorCombination(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeColorCombination(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDraftStatistics_ColorFiltering(t *testing.T) {
	svc := setupTestService(t)

	ctx := context.Background()

	// Save card ratings with different color filters
	ratingsAll := []seventeenlands.CardRating{
		{MTGAID: 1, GIHWR: 0.60, GamesPlayed: 1000},
	}
	ratingsWU := []seventeenlands.CardRating{
		{MTGAID: 1, GIHWR: 0.62, GamesPlayed: 500},
	}

	err := svc.SaveCardRatings(ctx, ratingsAll, "BLB", "PremierDraft", "", "2024-08-01", "2024-09-01")
	if err != nil {
		t.Fatalf("SaveCardRatings (all colors) failed: %v", err)
	}

	err = svc.SaveCardRatings(ctx, ratingsWU, "BLB", "PremierDraft", "W,U", "2024-08-01", "2024-09-01")
	if err != nil {
		t.Fatalf("SaveCardRatings (W,U) failed: %v", err)
	}

	// Verify both are stored separately
	ratingAll, err := svc.GetCardRating(ctx, 1, "BLB", "PremierDraft", "")
	if err != nil {
		t.Fatalf("GetCardRating (all) failed: %v", err)
	}
	if ratingAll.GIHWR != 0.60 {
		t.Errorf("Expected GIHWR 0.60 for all colors, got %f", ratingAll.GIHWR)
	}

	ratingWU, err := svc.GetCardRating(ctx, 1, "BLB", "PremierDraft", "W,U")
	if err != nil {
		t.Fatalf("GetCardRating (W,U) failed: %v", err)
	}
	if ratingWU.GIHWR != 0.62 {
		t.Errorf("Expected GIHWR 0.62 for W,U filter, got %f", ratingWU.GIHWR)
	}
}
