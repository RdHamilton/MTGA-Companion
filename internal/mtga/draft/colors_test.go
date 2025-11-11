package draft

import (
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

func TestParseManaCost(t *testing.T) {
	tests := []struct {
		name     string
		manaCost string
		want     []string
	}{
		{
			name:     "mono-white",
			manaCost: "{2}{W}{W}",
			want:     []string{"W"},
		},
		{
			name:     "two-color",
			manaCost: "{1}{W}{U}",
			want:     []string{"U", "W"}, // Sorted
		},
		{
			name:     "multi-color with generic",
			manaCost: "{2}{B}{R}{R}",
			want:     []string{"B", "R"},
		},
		{
			name:     "hybrid mana",
			manaCost: "{W/U}{W/U}",
			want:     []string{"U", "W"},
		},
		{
			name:     "colorless",
			manaCost: "{4}",
			want:     []string{},
		},
		{
			name:     "empty",
			manaCost: "",
			want:     []string{},
		},
		{
			name:     "five-color",
			manaCost: "{W}{U}{B}{R}{G}",
			want:     []string{"B", "G", "R", "U", "W"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseManaCost(tt.manaCost)
			if len(got) != len(tt.want) {
				t.Errorf("ParseManaCost() = %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ParseManaCost() = %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}

func TestCalculateDeckMetrics(t *testing.T) {
	cards := []*seventeenlands.CardRatingData{
		{
			Name: "Card 1",
			DeckColors: map[string]*seventeenlands.DeckColorRatings{
				"ALL": {GIHWR: 60.0, GIH: 100},
			},
		},
		{
			Name: "Card 2",
			DeckColors: map[string]*seventeenlands.DeckColorRatings{
				"ALL": {GIHWR: 50.0, GIH: 200},
			},
		},
		{
			Name: "Card 3",
			DeckColors: map[string]*seventeenlands.DeckColorRatings{
				"ALL": {GIHWR: 70.0, GIH: 150},
			},
		},
	}

	metrics := CalculateDeckMetrics(cards, "ALL")

	// Expected mean: (60 + 50 + 70) / 3 = 60
	if metrics.Mean != 60.0 {
		t.Errorf("Mean = %.2f, want 60.00", metrics.Mean)
	}

	// Expected std dev: sqrt(((0)^2 + (-10)^2 + (10)^2) / 3) = sqrt(200/3) ≈ 8.16
	expectedStdDev := 8.16
	if metrics.StandardDeviation < expectedStdDev-0.1 || metrics.StandardDeviation > expectedStdDev+0.1 {
		t.Errorf("StandardDeviation = %.2f, want ~%.2f", metrics.StandardDeviation, expectedStdDev)
	}
}

func TestCalculateColorAffinity(t *testing.T) {
	cards := []*seventeenlands.CardRatingData{
		{
			Name:     "Black Card 1",
			ManaCost: "{2}{B}{B}",
			DeckColors: map[string]*seventeenlands.DeckColorRatings{
				"ALL": {GIHWR: 65.0, GIH: 100},
			},
		},
		{
			Name:     "Red Card 1",
			ManaCost: "{1}{R}",
			DeckColors: map[string]*seventeenlands.DeckColorRatings{
				"ALL": {GIHWR: 62.0, GIH: 150},
			},
		},
		{
			Name:     "BR Card",
			ManaCost: "{B}{R}",
			DeckColors: map[string]*seventeenlands.DeckColorRatings{
				"ALL": {GIHWR: 58.0, GIH: 120},
			},
		},
		{
			Name:     "Weak White Card",
			ManaCost: "{W}",
			DeckColors: map[string]*seventeenlands.DeckColorRatings{
				"ALL": {GIHWR: 45.0, GIH: 80}, // Below threshold
			},
		},
	}

	threshold := 50.0
	affinities := CalculateColorAffinity(cards, "ALL", threshold)

	// Black should have highest affinity (2 cards above threshold)
	if affinities["B"].Count != 2 {
		t.Errorf("Black count = %d, want 2", affinities["B"].Count)
	}

	// Red should have 2 cards (both R cards above threshold)
	if affinities["R"].Count != 2 {
		t.Errorf("Red count = %d, want 2", affinities["R"].Count)
	}

	// White should have 0 cards (below threshold)
	if affinities["W"].Count != 0 {
		t.Errorf("White count = %d, want 0", affinities["W"].Count)
	}

	// Black score should be sum of strength above threshold
	// Card 1: 65 - 50 = 15
	// BR Card: 58 - 50 = 8
	// Total: 23
	expectedBlackScore := 15.0 + 8.0
	if affinities["B"].Score != expectedBlackScore {
		t.Errorf("Black score = %.2f, want %.2f", affinities["B"].Score, expectedBlackScore)
	}
}

func TestGenerateColorCombinations(t *testing.T) {
	tests := []struct {
		name      string
		colors    []string
		maxColors int
		wantCount int
		wantFirst string
	}{
		{
			name:      "2 colors, max 2",
			colors:    []string{"B", "R"},
			maxColors: 2,
			wantCount: 3, // B, R, BR
			wantFirst: "B",
		},
		{
			name:      "3 colors, max 2",
			colors:    []string{"W", "U", "B"},
			maxColors: 2,
			wantCount: 6, // W, U, B, WU, WB, UB
			wantFirst: "W",
		},
		{
			name:      "3 colors, max 3",
			colors:    []string{"B", "R", "G"},
			maxColors: 3,
			wantCount: 7, // B, R, G, BR, BG, RG, BRG
			wantFirst: "B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateColorCombinations(tt.colors, tt.maxColors)
			if len(got) != tt.wantCount {
				t.Errorf("GenerateColorCombinations() returned %d combinations, want %d", len(got), tt.wantCount)
				t.Logf("Got: %v", got)
			}
			if len(got) > 0 && got[0] != tt.wantFirst {
				t.Errorf("First combination = %s, want %s", got[0], tt.wantFirst)
			}
		})
	}
}

func TestIsSubsetOf(t *testing.T) {
	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{
			name: "subset",
			a:    []string{"B", "R"},
			b:    []string{"B", "R", "G"},
			want: true,
		},
		{
			name: "not subset",
			a:    []string{"B", "W"},
			b:    []string{"B", "R", "G"},
			want: false,
		},
		{
			name: "equal",
			a:    []string{"B", "R"},
			b:    []string{"B", "R"},
			want: true,
		},
		{
			name: "empty subset",
			a:    []string{},
			b:    []string{"B", "R"},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSubsetOf(tt.a, tt.b)
			if got != tt.want {
				t.Errorf("isSubsetOf(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestCalculateColorRating(t *testing.T) {
	cards := []*seventeenlands.CardRatingData{
		{
			Name:     "Black Card",
			ManaCost: "{2}{B}",
			DeckColors: map[string]*seventeenlands.DeckColorRatings{
				"B":   {GIHWR: 65.0, GIH: 100},
				"ALL": {GIHWR: 62.0, GIH: 150},
			},
		},
		{
			Name:     "Red Card",
			ManaCost: "{R}",
			DeckColors: map[string]*seventeenlands.DeckColorRatings{
				"R":   {GIHWR: 60.0, GIH: 120},
				"ALL": {GIHWR: 58.0, GIH: 130},
			},
		},
		{
			Name:     "BR Card",
			ManaCost: "{B}{R}",
			DeckColors: map[string]*seventeenlands.DeckColorRatings{
				"BR":  {GIHWR: 68.0, GIH: 90},
				"ALL": {GIHWR: 63.0, GIH: 110},
			},
		},
	}

	threshold := 50.0

	// Test BR rating
	rating := CalculateColorRating(cards, "BR", threshold)
	// Should include all 3 cards (all have B or R)
	// Uses "BR" specific ratings where available, falls back to "ALL"
	// Actual: (65 + 60 + 68) / 3 = 64.33, but with "ALL" fallback may differ
	if rating < 60.0 || rating > 70.0 {
		t.Errorf("BR rating = %.2f, want in range 60-70", rating)
	}

	// Test B only rating
	ratingB := CalculateColorRating(cards, "B", threshold)
	// Should include Black Card and BR Card
	// Uses "B" specific rating where available, falls back to "ALL"
	if ratingB < 60.0 || ratingB > 70.0 {
		t.Errorf("B rating = %.2f, want in range 60-70", ratingB)
	}
}

func TestCalculateCurveFactor(t *testing.T) {
	// Test with insufficient cards (should return 1.0)
	smallDeck := []*seventeenlands.CardRatingData{
		{Name: "Card", ManaCost: "{B}", CMC: 1.0},
	}
	factor := CalculateCurveFactor(smallDeck, "B")
	if factor != 1.0 {
		t.Errorf("Curve factor with <15 cards = %.2f, want 1.00", factor)
	}

	// Test with good curve
	goodCurve := make([]*seventeenlands.CardRatingData, 20)
	for i := 0; i < 20; i++ {
		cmc := 2.0
		if i < 10 {
			cmc = 2.0 // 2-drops
		} else if i < 16 {
			cmc = 3.0 // 3-drops
		} else {
			cmc = 4.0 // 4-drops
		}

		goodCurve[i] = &seventeenlands.CardRatingData{
			Name:     "Creature",
			ManaCost: "{B}",
			CMC:      cmc,
			Types:    []string{"Creature"},
		}
	}

	factorGood := CalculateCurveFactor(goodCurve, "B")
	// Should have positive factor (>= 1.0) for good curve
	if factorGood < 1.0 {
		t.Errorf("Curve factor with good curve = %.2f, want >= 1.00", factorGood)
	}
}

func TestRankDeckColors(t *testing.T) {
	cards := []*seventeenlands.CardRatingData{
		{
			Name:     "Black Card 1",
			ManaCost: "{2}{B}",
			DeckColors: map[string]*seventeenlands.DeckColorRatings{
				"ALL": {GIHWR: 65.0, GIH: 100},
			},
		},
		{
			Name:     "Black Card 2",
			ManaCost: "{B}{B}",
			DeckColors: map[string]*seventeenlands.DeckColorRatings{
				"ALL": {GIHWR: 70.0, GIH: 120},
			},
		},
		{
			Name:     "Red Card",
			ManaCost: "{1}{R}",
			DeckColors: map[string]*seventeenlands.DeckColorRatings{
				"ALL": {GIHWR: 62.0, GIH: 110},
			},
		},
		{
			Name:     "White Card",
			ManaCost: "{W}",
			DeckColors: map[string]*seventeenlands.DeckColorRatings{
				"ALL": {GIHWR: 48.0, GIH: 80},
			},
		},
	}

	config := DefaultColorAffinityConfig()
	metrics := CalculateDeckMetrics(cards, "ALL")

	ranked := RankDeckColors(cards, config, metrics)

	// Should have multiple color combinations
	if len(ranked) == 0 {
		t.Fatal("RankDeckColors() returned no combinations")
	}

	// First combination should be highest rated
	for i := 1; i < len(ranked); i++ {
		if ranked[i].Rating > ranked[0].Rating {
			t.Errorf("Color %s has higher rating than %s", ranked[i].Colors, ranked[0].Colors)
		}
	}

	// All ratings should have curve factor applied
	for _, dc := range ranked {
		if dc.CurveFactor == 0.0 {
			t.Errorf("Color %s has zero curve factor", dc.Colors)
		}
	}
}

func TestAutoSelectColors(t *testing.T) {
	tests := []struct {
		name      string
		cardCount int
		wantEmpty bool // true if should return empty (meaning ALL)
	}{
		{
			name:      "too few cards",
			cardCount: 10,
			wantEmpty: true,
		},
		{
			name:      "enough cards",
			cardCount: 20,
			wantEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test cards with mix of colors and varying GIHWR
			cards := make([]*seventeenlands.CardRatingData, tt.cardCount)
			for i := 0; i < tt.cardCount; i++ {
				// Alternate between B and R to create viable colors
				manaCost := "{B}"
				gihwr := 65.0 // Above average
				if i%2 == 1 {
					manaCost = "{R}"
					gihwr = 62.0
				}
				// Add some weaker cards
				if i%5 == 0 {
					gihwr = 50.0
				}

				cards[i] = &seventeenlands.CardRatingData{
					Name:     "Card",
					ManaCost: manaCost,
					DeckColors: map[string]*seventeenlands.DeckColorRatings{
						"ALL": {GIHWR: gihwr, GIH: 100},
					},
				}
			}

			config := DefaultColorAffinityConfig()
			metrics := CalculateDeckMetrics(cards, "ALL")

			result := AutoSelectColors(cards, config, metrics)

			isEmpty := len(result) == 0
			if isEmpty != tt.wantEmpty {
				t.Errorf("AutoSelectColors() returned empty=%v, want %v (got %v)", isEmpty, tt.wantEmpty, result)
			}
		})
	}
}

func TestFormatColorName(t *testing.T) {
	tests := []struct {
		colors string
		want   string
	}{
		{
			colors: "W",
			want:   "White",
		},
		{
			colors: "BR",
			want:   "Black-Red (BR)",
		},
		{
			colors: "WUG",
			want:   "White-Blue-Green (WUG)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.colors, func(t *testing.T) {
			got := FormatColorName(tt.colors)
			if got != tt.want {
				t.Errorf("FormatColorName(%s) = %s, want %s", tt.colors, got, tt.want)
			}
		})
	}
}

func TestDefaultColorAffinityConfig(t *testing.T) {
	config := DefaultColorAffinityConfig()

	if config.MinCards != 15 {
		t.Errorf("MinCards = %d, want 15", config.MinCards)
	}
	if config.MaxColors != 2 {
		t.Errorf("MaxColors = %d, want 2", config.MaxColors)
	}
	if !config.EnableAutoHighest {
		t.Error("EnableAutoHighest should be true by default")
	}
}
