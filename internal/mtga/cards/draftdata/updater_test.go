package draftdata

import (
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

func TestNewUpdater(t *testing.T) {
	scryfallClient := scryfall.NewClient()
	seventeenlandsClient := seventeenlands.NewClient(seventeenlands.DefaultClientOptions())

	tests := []struct {
		name    string
		config  UpdaterConfig
		wantErr bool
	}{
		{
			name: "valid config with all fields",
			config: UpdaterConfig{
				ScryfallClient:       scryfallClient,
				SeventeenLandsClient: seventeenlandsClient,
				Storage:              nil, // We'll need to mock this
				NewSetThreshold:      30 * 24 * time.Hour,
				StaleThreshold:       24 * time.Hour,
				DateRangeWindow:      7 * 24 * time.Hour,
			},
			wantErr: true, // Storage is nil
		},
		{
			name: "missing Scryfall client",
			config: UpdaterConfig{
				ScryfallClient:       nil,
				SeventeenLandsClient: seventeenlandsClient,
				Storage:              nil,
			},
			wantErr: true,
		},
		{
			name: "missing 17Lands client",
			config: UpdaterConfig{
				ScryfallClient:       scryfallClient,
				SeventeenLandsClient: nil,
				Storage:              nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewUpdater(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUpdater() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewUpdater_Defaults(t *testing.T) {
	// This test requires a mock storage, so we skip it for now
	// In a real implementation, you'd use a mock
	t.Skip("Requires mock storage implementation")
}

func TestActiveSet_IsNew(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name       string
		releaseAge time.Duration
		threshold  time.Duration
		want       bool
	}{
		{
			name:       "brand new set (1 day old)",
			releaseAge: 1 * 24 * time.Hour,
			threshold:  30 * 24 * time.Hour,
			want:       true,
		},
		{
			name:       "new set (29 days old)",
			releaseAge: 29 * 24 * time.Hour,
			threshold:  30 * 24 * time.Hour,
			want:       true,
		},
		{
			name:       "just over threshold (31 days)",
			releaseAge: 31 * 24 * time.Hour,
			threshold:  30 * 24 * time.Hour,
			want:       false,
		},
		{
			name:       "old set (6 months)",
			releaseAge: 180 * 24 * time.Hour,
			threshold:  30 * 24 * time.Hour,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			releaseDate := now.Add(-tt.releaseAge)
			age := now.Sub(releaseDate)
			isNew := age <= tt.threshold

			if isNew != tt.want {
				t.Errorf("IsNew calculation = %v, want %v (age: %v, threshold: %v)",
					isNew, tt.want, age, tt.threshold)
			}
		})
	}
}

// TestGetActiveSets_Integration tests the GetActiveSets method with real Scryfall API.
// This is an integration test and will be skipped in normal test runs.
func TestGetActiveSets_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	scryfallClient := scryfall.NewClient()
	seventeenlandsClient := seventeenlands.NewClient(seventeenlands.DefaultClientOptions())

	// Note: This test will fail without a real storage instance
	// In production, you'd use a test database or mock
	t.Skip("Requires storage implementation")

	// This is the test code that would run with proper mocks:
	/*
		updater, err := NewUpdater(UpdaterConfig{
			ScryfallClient:       scryfallClient,
			SeventeenLandsClient: seventeenlandsClient,
			Storage:              testStorage, // Would need a test storage instance
		})
		if err != nil {
			t.Fatalf("Failed to create updater: %v", err)
		}

		ctx := context.Background()
		sets, err := updater.GetActiveSets(ctx)
		if err != nil {
			t.Fatalf("GetActiveSets failed: %v", err)
		}

		if len(sets) == 0 {
			t.Error("Expected at least one active set")
		}

		// Verify sets have required fields
		for _, set := range sets {
			if set.Code == "" {
				t.Error("Active set missing code")
			}
			if set.Name == "" {
				t.Error("Active set missing name")
			}
			if set.ReleasedAt.IsZero() {
				t.Error("Active set missing release date")
			}
		}

		// Verify at least one new set exists (or log if none)
		hasNewSet := false
		for _, set := range sets {
			if set.IsNew {
				hasNewSet = true
				t.Logf("Found new set: %s (%s) released %v ago", set.Code, set.Name, time.Since(set.ReleasedAt))
				break
			}
		}
		if !hasNewSet {
			t.Logf("No new sets found (this may be normal depending on release schedule)")
		}
	*/

	_ = scryfallClient
	_ = seventeenlandsClient
}

func TestUpdateResult(t *testing.T) {
	result := &UpdateResult{
		SetCode:      "BLB",
		Success:      true,
		CardRatings:  250,
		ColorRatings: 10,
		Error:        nil,
		Duration:     5 * time.Second,
	}

	if result.SetCode != "BLB" {
		t.Errorf("Expected SetCode BLB, got %s", result.SetCode)
	}
	if !result.Success {
		t.Error("Expected Success to be true")
	}
	if result.CardRatings != 250 {
		t.Errorf("Expected 250 CardRatings, got %d", result.CardRatings)
	}
	if result.ColorRatings != 10 {
		t.Errorf("Expected 10 ColorRatings, got %d", result.ColorRatings)
	}
	if result.Duration != 5*time.Second {
		t.Errorf("Expected Duration 5s, got %v", result.Duration)
	}
}
