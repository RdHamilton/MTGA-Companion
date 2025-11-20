package datasets

import (
	"context"
	"testing"
	"time"
)

func TestGetCardRatings_WebAPIFallback(t *testing.T) {
	// Test that TLA uses web API fallback (not in S3 yet)
	service, err := NewService(DefaultServiceOptions())
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// TLA should use web API fallback
	ratings, err := service.GetCardRatings(ctx, "TLA", "PremierDraft")
	if err != nil {
		t.Fatalf("Failed to get TLA ratings: %v", err)
	}

	if len(ratings) == 0 {
		t.Error("Expected ratings for TLA, got 0")
	}

	t.Logf("Successfully fetched %d TLA card ratings", len(ratings))

	// Verify data source
	source := service.GetDataSource(ctx, "TLA", "PremierDraft")
	t.Logf("TLA data source: %s", source)

	// Check if ratings have game counts
	if len(ratings) > 0 {
		hasGameCounts := false
		for _, r := range ratings {
			if r.GIH > 0 {
				hasGameCounts = true
				t.Logf("Sample card with game count: %s - GIH=%d, GIHWR=%.2f%%, OHWR=%.2f%%",
					r.Name, r.GIH, r.GIHWR, r.OHWR)
				break
			}
		}
		if !hasGameCounts {
			t.Error("Expected ratings to have game counts (GIH > 0), but all were 0")
		}
	}
}

func TestGetCardRatings_S3Dataset(t *testing.T) {
	// Test that an older set uses S3 dataset (if available)
	service, err := NewService(DefaultServiceOptions())
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Try BLB (Bloomburrow) - should be in S3
	ratings, err := service.GetCardRatings(ctx, "BLB", "PremierDraft")
	if err != nil {
		t.Fatalf("Failed to get BLB ratings: %v", err)
	}

	if len(ratings) == 0 {
		t.Error("Expected ratings for BLB, got 0")
	}

	t.Logf("Successfully fetched %d BLB card ratings", len(ratings))

	// Verify data source
	source := service.GetDataSource(ctx, "BLB", "PremierDraft")
	t.Logf("BLB data source: %s", source)
}

func TestCheckDatasetAvailability(t *testing.T) {
	service, err := NewService(DefaultServiceOptions())
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Check if BLB is available in S3
	available, err := service.CheckDatasetAvailability(ctx, "BLB", "PremierDraft")
	if err != nil {
		t.Fatalf("Failed to check availability: %v", err)
	}

	t.Logf("BLB dataset available in S3: %v", available)

	// Check if TLA is available in S3 (should be false)
	available, err = service.CheckDatasetAvailability(ctx, "TLA", "PremierDraft")
	if err != nil {
		t.Fatalf("Failed to check availability: %v", err)
	}

	t.Logf("TLA dataset available in S3: %v", available)
}
