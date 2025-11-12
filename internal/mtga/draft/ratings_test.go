package draft

import (
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

func createTestSetFile() *seventeenlands.SetFile {
	return &seventeenlands.SetFile{
		Meta: seventeenlands.SetMeta{
			SetCode:     "TST",
			DraftFormat: "PremierDraft",
		},
		CardRatings: map[string]*seventeenlands.CardRatingData{
			"100": {
				ArenaID:  100,
				Name:     "Strong Black Card",
				ManaCost: "{2}{B}{B}",
				CMC:      4.0,
				Rarity:   "uncommon",
				Types:    []string{"Creature"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 65.0, ALSA: 5.0, ATA: 3.0, IWD: 2.0, GIH: 500},
					"B":   {GIHWR: 68.0, ALSA: 4.5, ATA: 2.5, IWD: 2.5, GIH: 400},
					"BR":  {GIHWR: 70.0, ALSA: 4.0, ATA: 2.0, IWD: 3.0, GIH: 350},
				},
			},
			"200": {
				ArenaID:  200,
				Name:     "Medium Red Card",
				ManaCost: "{1}{R}",
				CMC:      2.0,
				Rarity:   "common",
				Types:    []string{"Creature"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 55.0, ALSA: 8.0, ATA: 6.0, IWD: 1.0, GIH: 600},
					"R":   {GIHWR: 58.0, ALSA: 7.5, ATA: 5.5, IWD: 1.5, GIH: 500},
					"BR":  {GIHWR: 60.0, ALSA: 7.0, ATA: 5.0, IWD: 2.0, GIH: 450},
				},
			},
			"300": {
				ArenaID:  300,
				Name:     "Weak White Card",
				ManaCost: "{W}",
				CMC:      1.0,
				Rarity:   "common",
				Types:    []string{"Creature"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 48.0, ALSA: 10.0, ATA: 9.0, IWD: 0.5, GIH: 400},
					"W":   {GIHWR: 50.0, ALSA: 9.5, ATA: 8.5, IWD: 0.8, GIH: 350},
				},
			},
			"400": {
				ArenaID:  400,
				Name:     "Low Sample Black Card",
				ManaCost: "{3}{B}",
				CMC:      4.0,
				Rarity:   "rare",
				Types:    []string{"Creature"},
				DeckColors: map[string]*seventeenlands.DeckColorRatings{
					"ALL": {GIHWR: 75.0, ALSA: 2.0, ATA: 1.5, IWD: 4.0, GIH: 50}, // Low GIH
					"B":   {GIHWR: 78.0, ALSA: 1.8, ATA: 1.2, IWD: 4.5, GIH: 40},
				},
			},
		},
	}
}

func TestNewRatingsProvider(t *testing.T) {
	setFile := createTestSetFile()
	config := DefaultBayesianConfig()

	rp := NewRatingsProvider(setFile, config, nil)

	if rp == nil {
		t.Fatal("NewRatingsProvider() returned nil")
	}

	if rp.setFile != setFile {
		t.Error("SetFile not stored correctly")
	}

	if rp.config.MinGIH != config.MinGIH {
		t.Errorf("Config.MinGIH = %d, want %d", rp.config.MinGIH, config.MinGIH)
	}
}

func TestDefaultBayesianConfig(t *testing.T) {
	config := DefaultBayesianConfig()

	if config.MinGIH != 300 {
		t.Errorf("MinGIH = %d, want 300", config.MinGIH)
	}

	if config.PriorMean != 50.0 {
		t.Errorf("PriorMean = %.1f, want 50.0", config.PriorMean)
	}

	if config.PriorWeight != 100 {
		t.Errorf("PriorWeight = %d, want 100", config.PriorWeight)
	}
}

func TestGetCardRating(t *testing.T) {
	setFile := createTestSetFile()
	config := DefaultBayesianConfig()
	rp := NewRatingsProvider(setFile, config, nil)

	tests := []struct {
		name         string
		cardID       int
		colorFilter  string
		wantName     string
		wantGIHWR    float64
		wantBayesian bool
		wantError    bool
	}{
		{
			name:         "strong card with ALL filter",
			cardID:       100,
			colorFilter:  "ALL",
			wantName:     "Strong Black Card",
			wantGIHWR:    65.0,
			wantBayesian: false, // GIH=500 > 300
		},
		{
			name:         "strong card with B filter",
			cardID:       100,
			colorFilter:  "B",
			wantName:     "Strong Black Card",
			wantGIHWR:    68.0,
			wantBayesian: false, // GIH=400 > 300
		},
		{
			name:         "strong card with BR filter",
			cardID:       100,
			colorFilter:  "BR",
			wantName:     "Strong Black Card",
			wantGIHWR:    70.0,
			wantBayesian: false, // GIH=350 > 300
		},
		{
			name:         "low sample card needs Bayesian",
			cardID:       400,
			colorFilter:  "ALL",
			wantName:     "Low Sample Black Card",
			wantGIHWR:    75.0,
			wantBayesian: true, // GIH=50 < 300
		},
		{
			name:        "card not found",
			cardID:      999,
			colorFilter: "ALL",
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rating, err := rp.GetCardRating(tt.cardID, tt.colorFilter)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetCardRating() error = %v", err)
			}

			if rating.Name != tt.wantName {
				t.Errorf("Name = %s, want %s", rating.Name, tt.wantName)
			}

			if rating.GIHWR != tt.wantGIHWR {
				t.Errorf("GIHWR = %.1f, want %.1f", rating.GIHWR, tt.wantGIHWR)
			}

			if rating.IsBayesianAdjust != tt.wantBayesian {
				t.Errorf("IsBayesianAdjust = %v, want %v", rating.IsBayesianAdjust, tt.wantBayesian)
			}

			// Verify Bayesian adjustment was applied correctly
			if tt.wantBayesian {
				expected := rp.calculateBayesianGIHWR(rating.GIHWR, rating.GIH)
				if rating.BayesianGIHWR != expected {
					t.Errorf("BayesianGIHWR = %.2f, want %.2f", rating.BayesianGIHWR, expected)
				}
				// Bayesian should pull toward prior (50.0)
				if rating.BayesianGIHWR >= rating.GIHWR {
					t.Errorf("Bayesian adjustment should reduce GIHWR for low sample, got %.2f >= %.2f",
						rating.BayesianGIHWR, rating.GIHWR)
				}
			} else {
				if rating.BayesianGIHWR != rating.GIHWR {
					t.Errorf("BayesianGIHWR should equal GIHWR for high sample")
				}
			}
		})
	}
}

func TestGetPackRatings(t *testing.T) {
	setFile := createTestSetFile()
	config := DefaultBayesianConfig()
	rp := NewRatingsProvider(setFile, config, nil)

	pack := &Pack{
		PackNumber: 1,
		PickNumber: 1,
		CardIDs:    []int{100, 200, 300, 999}, // 999 doesn't exist
		Timestamp:  time.Now(),
	}

	packRatings, err := rp.GetPackRatings(pack, "ALL")
	if err != nil {
		t.Fatalf("GetPackRatings() error = %v", err)
	}

	// Should have 3 cards (999 is skipped)
	if len(packRatings.CardRatings) != 3 {
		t.Errorf("CardRatings length = %d, want 3", len(packRatings.CardRatings))
	}

	if packRatings.Pack != pack {
		t.Error("Pack reference not stored correctly")
	}

	if packRatings.ColorFilter != "ALL" {
		t.Errorf("ColorFilter = %s, want ALL", packRatings.ColorFilter)
	}
}

func TestGetPackRatingsWithColorFilter(t *testing.T) {
	setFile := createTestSetFile()
	config := DefaultBayesianConfig()
	rp := NewRatingsProvider(setFile, config, nil)

	pack := &Pack{
		PackNumber: 1,
		PickNumber: 1,
		CardIDs:    []int{100, 200},
		Timestamp:  time.Now(),
	}

	packRatings, err := rp.GetPackRatings(pack, "BR")
	if err != nil {
		t.Fatalf("GetPackRatings() error = %v", err)
	}

	// Verify BR-specific ratings are used
	for _, rating := range packRatings.CardRatings {
		if rating.CardID == 100 {
			if rating.GIHWR != 70.0 { // BR rating for card 100
				t.Errorf("Card 100 GIHWR = %.1f, want 70.0 (BR rating)", rating.GIHWR)
			}
		} else if rating.CardID == 200 {
			if rating.GIHWR != 60.0 { // BR rating for card 200
				t.Errorf("Card 200 GIHWR = %.1f, want 60.0 (BR rating)", rating.GIHWR)
			}
		}
	}
}

func TestCalculateBayesianGIHWR(t *testing.T) {
	config := DefaultBayesianConfig()
	rp := NewRatingsProvider(createTestSetFile(), config, nil)

	tests := []struct {
		name         string
		observedWR   float64
		gih          int
		wantInRange  [2]float64 // min, max
		wantBehavior string
	}{
		{
			name:         "high sample maintains value",
			observedWR:   70.0,
			gih:          1000,
			wantInRange:  [2]float64{68.0, 69.0}, // (100*50 + 1000*70)/(100+1000) = 68.18
			wantBehavior: "should be close to observed",
		},
		{
			name:         "low sample pulls toward prior",
			observedWR:   80.0,
			gih:          50,
			wantInRange:  [2]float64{55.0, 70.0}, // Between prior (50) and observed (80)
			wantBehavior: "should be between prior and observed",
		},
		{
			name:         "zero sample equals prior",
			observedWR:   90.0,
			gih:          0,
			wantInRange:  [2]float64{50.0, 50.0},
			wantBehavior: "should equal prior",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rp.calculateBayesianGIHWR(tt.observedWR, tt.gih)

			if result < tt.wantInRange[0] || result > tt.wantInRange[1] {
				t.Errorf("BayesianGIHWR = %.2f, want in range [%.2f, %.2f] (%s)",
					result, tt.wantInRange[0], tt.wantInRange[1], tt.wantBehavior)
			}
		})
	}
}

func TestPackRatingsSortByRating(t *testing.T) {
	setFile := createTestSetFile()
	config := DefaultBayesianConfig()
	rp := NewRatingsProvider(setFile, config, nil)

	pack := &Pack{
		PackNumber: 1,
		PickNumber: 1,
		CardIDs:    []int{300, 100, 200}, // Weak, Strong, Medium
		Timestamp:  time.Now(),
	}

	packRatings, err := rp.GetPackRatings(pack, "ALL")
	if err != nil {
		t.Fatalf("GetPackRatings() error = %v", err)
	}

	packRatings.SortByRating()

	// Should be sorted: 100 (65.0), 200 (55.0), 300 (48.0)
	if packRatings.CardRatings[0].CardID != 100 {
		t.Errorf("First card = %d, want 100", packRatings.CardRatings[0].CardID)
	}

	if packRatings.CardRatings[1].CardID != 200 {
		t.Errorf("Second card = %d, want 200", packRatings.CardRatings[1].CardID)
	}

	if packRatings.CardRatings[2].CardID != 300 {
		t.Errorf("Third card = %d, want 300", packRatings.CardRatings[2].CardID)
	}

	// Verify descending order
	for i := 0; i < len(packRatings.CardRatings)-1; i++ {
		if packRatings.CardRatings[i].BayesianGIHWR < packRatings.CardRatings[i+1].BayesianGIHWR {
			t.Errorf("Not sorted descending at index %d", i)
		}
	}
}

func TestPackRatingsTopN(t *testing.T) {
	setFile := createTestSetFile()
	config := DefaultBayesianConfig()
	rp := NewRatingsProvider(setFile, config, nil)

	pack := &Pack{
		PackNumber: 1,
		PickNumber: 1,
		CardIDs:    []int{100, 200, 300},
		Timestamp:  time.Now(),
	}

	packRatings, err := rp.GetPackRatings(pack, "ALL")
	if err != nil {
		t.Fatalf("GetPackRatings() error = %v", err)
	}

	// Get top 2
	top2 := packRatings.TopN(2)
	if len(top2) != 2 {
		t.Errorf("TopN(2) returned %d cards, want 2", len(top2))
	}

	// Should be 100 and 200
	if top2[0].CardID != 100 {
		t.Errorf("Top card = %d, want 100", top2[0].CardID)
	}
	if top2[1].CardID != 200 {
		t.Errorf("Second card = %d, want 200", top2[1].CardID)
	}

	// Test TopN with N > pack size
	topAll := packRatings.TopN(10)
	if len(topAll) != 3 {
		t.Errorf("TopN(10) returned %d cards, want 3", len(topAll))
	}
}

func TestPackRatingsGetBestPick(t *testing.T) {
	setFile := createTestSetFile()
	config := DefaultBayesianConfig()
	rp := NewRatingsProvider(setFile, config, nil)

	pack := &Pack{
		PackNumber: 1,
		PickNumber: 1,
		CardIDs:    []int{300, 100, 200},
		Timestamp:  time.Now(),
	}

	packRatings, err := rp.GetPackRatings(pack, "ALL")
	if err != nil {
		t.Fatalf("GetPackRatings() error = %v", err)
	}

	best := packRatings.GetBestPick()
	if best == nil {
		t.Fatal("GetBestPick() returned nil")
	}

	if best.CardID != 100 {
		t.Errorf("Best pick = %d, want 100", best.CardID)
	}

	// Test empty pack
	emptyPack := &PackRatings{
		Pack:        pack,
		CardRatings: []*CardRating{},
	}
	if emptyPack.GetBestPick() != nil {
		t.Error("GetBestPick() on empty pack should return nil")
	}
}

func TestPackRatingsFilterByColors(t *testing.T) {
	setFile := createTestSetFile()
	config := DefaultBayesianConfig()
	rp := NewRatingsProvider(setFile, config, nil)

	pack := &Pack{
		PackNumber: 1,
		PickNumber: 1,
		CardIDs:    []int{100, 200, 300}, // B, R, W
		Timestamp:  time.Now(),
	}

	packRatings, err := rp.GetPackRatings(pack, "ALL")
	if err != nil {
		t.Fatalf("GetPackRatings() error = %v", err)
	}

	// Filter for BR colors
	brCards := packRatings.FilterByColors([]string{"B", "R"})
	if len(brCards) != 2 { // Should have B and R cards
		t.Errorf("FilterByColors(BR) returned %d cards, want 2", len(brCards))
	}

	// Verify only B and R cards
	for _, card := range brCards {
		colors := card.Colors
		isValid := false
		for _, c := range colors {
			if c == "B" || c == "R" {
				isValid = true
				break
			}
		}
		if !isValid {
			t.Errorf("Card %s has colors %v, expected B or R", card.Name, colors)
		}
	}

	// Filter with empty colors returns all
	allCards := packRatings.FilterByColors([]string{})
	if len(allCards) != 3 {
		t.Errorf("FilterByColors([]) returned %d cards, want 3", len(allCards))
	}
}

func TestPackRatingsGetCardByID(t *testing.T) {
	setFile := createTestSetFile()
	config := DefaultBayesianConfig()
	rp := NewRatingsProvider(setFile, config, nil)

	pack := &Pack{
		PackNumber: 1,
		PickNumber: 1,
		CardIDs:    []int{100, 200, 300},
		Timestamp:  time.Now(),
	}

	packRatings, err := rp.GetPackRatings(pack, "ALL")
	if err != nil {
		t.Fatalf("GetPackRatings() error = %v", err)
	}

	// Find existing card
	card := packRatings.GetCardByID(200)
	if card == nil {
		t.Fatal("GetCardByID(200) returned nil")
	}
	if card.CardID != 200 {
		t.Errorf("CardID = %d, want 200", card.CardID)
	}

	// Find non-existing card
	if packRatings.GetCardByID(999) != nil {
		t.Error("GetCardByID(999) should return nil")
	}
}

func TestRatingsProviderNilSetFile(t *testing.T) {
	config := DefaultBayesianConfig()
	rp := NewRatingsProvider(nil, config, nil)

	_, err := rp.GetCardRating(100, "ALL")
	if err == nil {
		t.Error("Expected error with nil set file")
	}
}

func TestGetPackRatingsNilPack(t *testing.T) {
	setFile := createTestSetFile()
	config := DefaultBayesianConfig()
	rp := NewRatingsProvider(setFile, config, nil)

	_, err := rp.GetPackRatings(nil, "ALL")
	if err == nil {
		t.Error("Expected error with nil pack")
	}
}

func TestGetPackRatingsEmptyColorFilter(t *testing.T) {
	setFile := createTestSetFile()
	config := DefaultBayesianConfig()
	rp := NewRatingsProvider(setFile, config, nil)

	pack := &Pack{
		PackNumber: 1,
		PickNumber: 1,
		CardIDs:    []int{100},
		Timestamp:  time.Now(),
	}

	// Empty color filter should default to "ALL"
	packRatings, err := rp.GetPackRatings(pack, "")
	if err != nil {
		t.Fatalf("GetPackRatings() error = %v", err)
	}

	if packRatings.ColorFilter != "ALL" {
		t.Errorf("ColorFilter = %s, want ALL", packRatings.ColorFilter)
	}
}
