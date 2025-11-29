package ml

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/meta"
)

func createTestMetaServers() (*httptest.Server, *httptest.Server) {
	goldfishServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
		<body>
		<div class="archetype-tile">
			<span class="archetype-tile-title">Mono Red Aggro</span>
			<span class="archetype-tile-statistic">15.5%</span>
		</div>
		<div class="archetype-tile">
			<span class="archetype-tile-title">Azorius Control</span>
			<span class="archetype-tile-statistic">12.3%</span>
		</div>
		<div class="archetype-tile">
			<span class="archetype-tile-title">Golgari Midrange</span>
			<span class="archetype-tile-statistic">3.7%</span>
		</div>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))

	top8Server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
		<html>
		<body>
		<table>
		<tr>
			<td><a href="/archetype?a=123">Mono Red Aggro</a></td>
			<td>25</td>
		</tr>
		<tr>
			<td><a href="/archetype?a=456">Azorius Control</a></td>
			<td>18</td>
		</tr>
		<tr>
			<td><a href="/archetype?a=789">Dimir Control</a></td>
			<td>12</td>
		</tr>
		</table>
		<div class="event_title">Grand Prix Test</div>
		</body>
		</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))

	return goldfishServer, top8Server
}

func createTestMetaService(t *testing.T) (*meta.Service, func()) {
	goldfishServer, top8Server := createTestMetaServers()

	config := &meta.ServiceConfig{
		GoldfishConfig: &meta.GoldfishConfig{
			BaseURL:        goldfishServer.URL,
			CacheTTL:       1 * time.Hour,
			RequestTimeout: 5 * time.Second,
			RateLimitMs:    10,
		},
		Top8Config: &meta.Top8Config{
			BaseURL:        top8Server.URL,
			CacheTTL:       1 * time.Hour,
			RequestTimeout: 5 * time.Second,
			RateLimitMs:    10,
		},
	}

	service := meta.NewService(config)

	cleanup := func() {
		goldfishServer.Close()
		top8Server.Close()
	}

	return service, cleanup
}

func TestDefaultMetaWeightingConfig(t *testing.T) {
	config := DefaultMetaWeightingConfig()

	if config.Tier1Weight != 1.0 {
		t.Errorf("unexpected Tier1Weight: %f", config.Tier1Weight)
	}
	if config.Tier2Weight != 0.75 {
		t.Errorf("unexpected Tier2Weight: %f", config.Tier2Weight)
	}
	if config.TournamentBonus != 0.1 {
		t.Errorf("unexpected TournamentBonus: %f", config.TournamentBonus)
	}
	if config.CacheTTL != 30*time.Minute {
		t.Errorf("unexpected CacheTTL: %v", config.CacheTTL)
	}
}

func TestNewMetaWeighter(t *testing.T) {
	t.Run("with nil config uses defaults", func(t *testing.T) {
		metaService, cleanup := createTestMetaService(t)
		defer cleanup()

		weighter := NewMetaWeighter(metaService, nil)
		if weighter == nil {
			t.Fatal("expected non-nil weighter")
		}
		if weighter.config.Tier1Weight != 1.0 {
			t.Error("expected default config")
		}
	})

	t.Run("with custom config", func(t *testing.T) {
		metaService, cleanup := createTestMetaService(t)
		defer cleanup()

		config := &MetaWeightingConfig{
			Tier1Weight: 0.8,
			CacheTTL:    1 * time.Hour,
		}
		weighter := NewMetaWeighter(metaService, config)
		if weighter.config.Tier1Weight != 0.8 {
			t.Errorf("expected custom config, got %f", weighter.config.Tier1Weight)
		}
	})
}

func TestMetaWeighter_GetArchetypeScore(t *testing.T) {
	metaService, cleanup := createTestMetaService(t)
	defer cleanup()

	weighter := NewMetaWeighter(metaService, nil)
	ctx := context.Background()

	t.Run("found archetype", func(t *testing.T) {
		score, err := weighter.GetArchetypeScore(ctx, "standard", "mono red")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if score == nil {
			t.Fatal("expected non-nil score")
		}
		if score.OverallScore <= 0 {
			t.Error("expected positive overall score")
		}
	})

	t.Run("cache hit", func(t *testing.T) {
		// First call
		_, _ = weighter.GetArchetypeScore(ctx, "standard", "mono red")

		// Second call should use cache
		score, err := weighter.GetArchetypeScore(ctx, "standard", "mono red")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if score == nil {
			t.Fatal("expected cached score")
		}
	})
}

func TestMetaWeighter_GetCardMetaScore(t *testing.T) {
	metaService, cleanup := createTestMetaService(t)
	defer cleanup()

	weighter := NewMetaWeighter(metaService, nil)
	ctx := context.Background()

	t.Run("with matching archetype", func(t *testing.T) {
		score, err := weighter.GetCardMetaScore(ctx, "standard", 12345, []string{"R"}, "aggro")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if score == nil {
			t.Fatal("expected non-nil score")
		}
		if score.Score < 0 || score.Score > 1 {
			t.Errorf("score out of range: %f", score.Score)
		}
	})

	t.Run("with color match only", func(t *testing.T) {
		score, err := weighter.GetCardMetaScore(ctx, "standard", 67890, []string{"W", "U"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if score == nil {
			t.Fatal("expected non-nil score")
		}
	})

	t.Run("no color match", func(t *testing.T) {
		// Use colors that don't match any archetype in our mock data
		score, err := weighter.GetCardMetaScore(ctx, "standard", 11111, []string{"W", "B", "G", "R", "U"}, "unknown deck xyz")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if score == nil {
			t.Fatal("expected non-nil score")
		}
		// Score should be valid
		if score.Score < 0 || score.Score > 1 {
			t.Errorf("score out of range: %f", score.Score)
		}
	})
}

func TestMetaWeighter_GetTopArchetypes(t *testing.T) {
	metaService, cleanup := createTestMetaService(t)
	defer cleanup()

	weighter := NewMetaWeighter(metaService, nil)
	ctx := context.Background()

	archetypes, err := weighter.GetTopArchetypes(ctx, "standard", 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(archetypes) > 2 {
		t.Errorf("expected at most 2 archetypes, got %d", len(archetypes))
	}

	// Should be sorted by score descending
	for i := 1; i < len(archetypes); i++ {
		if archetypes[i].OverallScore > archetypes[i-1].OverallScore {
			t.Error("archetypes should be sorted by score descending")
		}
	}
}

func TestMetaWeighter_GetMetaScoreForDeck(t *testing.T) {
	metaService, cleanup := createTestMetaService(t)
	defer cleanup()

	weighter := NewMetaWeighter(metaService, nil)
	ctx := context.Background()

	t.Run("with archetype match", func(t *testing.T) {
		score, err := weighter.GetMetaScoreForDeck(ctx, "standard", []string{"R"}, "aggro")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if score < 0 || score > 1 {
			t.Errorf("score out of range: %f", score)
		}
	})

	t.Run("with color match only", func(t *testing.T) {
		score, err := weighter.GetMetaScoreForDeck(ctx, "standard", []string{"W", "U"}, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if score < 0 || score > 1 {
			t.Errorf("score out of range: %f", score)
		}
	})
}

func TestMetaWeighter_ClearCache(t *testing.T) {
	metaService, cleanup := createTestMetaService(t)
	defer cleanup()

	weighter := NewMetaWeighter(metaService, nil)
	ctx := context.Background()

	// Populate cache
	_, _ = weighter.GetArchetypeScore(ctx, "standard", "mono red")

	// Clear cache
	weighter.ClearCache()

	// Verify cache is empty
	if len(weighter.archetypeScores) != 0 {
		t.Error("expected empty archetype scores cache")
	}
	if len(weighter.cardScores) != 0 {
		t.Error("expected empty card scores cache")
	}
}

func TestMetaWeighter_UpdateConfig(t *testing.T) {
	metaService, cleanup := createTestMetaService(t)
	defer cleanup()

	weighter := NewMetaWeighter(metaService, nil)
	ctx := context.Background()

	// Populate cache
	_, _ = weighter.GetArchetypeScore(ctx, "standard", "mono red")

	// Update config
	newConfig := &MetaWeightingConfig{
		Tier1Weight: 0.5,
		CacheTTL:    1 * time.Hour,
	}
	weighter.UpdateConfig(newConfig)

	// Cache should be cleared
	if len(weighter.archetypeScores) != 0 {
		t.Error("expected cache to be cleared after config update")
	}

	if weighter.config.Tier1Weight != 0.5 {
		t.Errorf("expected updated config, got %f", weighter.config.Tier1Weight)
	}
}

func TestMetaWeighter_ColorsMatch(t *testing.T) {
	weighter := &MetaWeighter{}

	tests := []struct {
		name            string
		archetypeColors []string
		cardColors      []string
		expected        bool
	}{
		{"exact match", []string{"W", "U"}, []string{"W", "U"}, true},
		{"subset match", []string{"W", "U", "B"}, []string{"W", "U"}, true},
		{"no match", []string{"R", "G"}, []string{"W", "U"}, false},
		{"empty card colors", []string{"W", "U"}, []string{}, true},
		{"empty archetype colors", []string{}, []string{"W"}, false},
		{"single color match", []string{"R"}, []string{"R"}, true},
		{"partial match fails", []string{"W"}, []string{"W", "U"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := weighter.colorsMatch(tt.archetypeColors, tt.cardColors)
			if result != tt.expected {
				t.Errorf("colorsMatch(%v, %v) = %v, expected %v",
					tt.archetypeColors, tt.cardColors, result, tt.expected)
			}
		})
	}
}

func TestMetaWeighter_CalculateArchetypeScore(t *testing.T) {
	weighter := NewMetaWeighter(nil, nil)

	tests := []struct {
		name     string
		arch     *meta.AggregatedArchetype
		minScore float64
		maxScore float64
	}{
		{
			"tier 1 with tournament wins",
			&meta.AggregatedArchetype{
				Name:            "Top Deck",
				Tier:            1,
				MetaShare:       15.0,
				TournamentTop8s: 25,
				TournamentWins:  5,
				TrendDirection:  "up",
			},
			0.7, 1.0,
		},
		{
			"tier 2",
			&meta.AggregatedArchetype{
				Name:            "Mid Deck",
				Tier:            2,
				MetaShare:       5.0,
				TournamentTop8s: 10,
				TrendDirection:  "stable",
			},
			0.4, 0.8,
		},
		{
			"tier 4 trending down",
			&meta.AggregatedArchetype{
				Name:            "Low Deck",
				Tier:            4,
				MetaShare:       0.5,
				TournamentTop8s: 2,
				TrendDirection:  "down",
			},
			0.1, 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := weighter.calculateArchetypeScore(tt.arch)
			if score.OverallScore < tt.minScore || score.OverallScore > tt.maxScore {
				t.Errorf("overall score %f outside expected range [%f, %f]",
					score.OverallScore, tt.minScore, tt.maxScore)
			}
		})
	}
}

func TestArchetypeMetaScore_Fields(t *testing.T) {
	score := &ArchetypeMetaScore{
		ArchetypeName:   "Test",
		MetaShare:       15.5,
		TournamentScore: 0.8,
		TierScore:       1.0,
		TrendScore:      0.1,
		OverallScore:    0.85,
		Confidence:      0.9,
		Colors:          []string{"R"},
		LastUpdated:     time.Now(),
	}

	if score.ArchetypeName != "Test" {
		t.Error("unexpected archetype name")
	}
	if score.MetaShare != 15.5 {
		t.Error("unexpected meta share")
	}
}

func TestCardMetaScore_Fields(t *testing.T) {
	score := &CardMetaScore{
		CardID:           12345,
		Score:            0.75,
		ArchetypeMatches: []string{"Aggro", "Tempo"},
		TopArchetype:     "Aggro",
		Confidence:       0.8,
		Factors:          []string{"Tournament proven"},
		LastUpdated:      time.Now(),
	}

	if score.CardID != 12345 {
		t.Error("unexpected card ID")
	}
	if len(score.ArchetypeMatches) != 2 {
		t.Error("unexpected archetype matches length")
	}
}

func TestMetaWeightedScorer(t *testing.T) {
	metaService, cleanup := createTestMetaService(t)
	defer cleanup()

	weighter := NewMetaWeighter(metaService, nil)
	scorer := NewMetaWeightedScorer(nil, weighter)

	if scorer == nil {
		t.Fatal("expected non-nil scorer")
	}
	if scorer.metaWeighter != weighter {
		t.Error("weighter not set correctly")
	}
	if scorer.GetMetaWeighter() != weighter {
		t.Error("GetMetaWeighter returned wrong weighter")
	}
	if scorer.GetModel() != nil {
		t.Error("expected nil model")
	}
}

func TestMetaWeightedScorer_ScoreCardsMetaOnly(t *testing.T) {
	metaService, cleanup := createTestMetaService(t)
	defer cleanup()

	weighter := NewMetaWeighter(metaService, nil)
	scorer := NewMetaWeightedScorer(nil, weighter) // No model, meta only

	ctx := context.Background()
	deck := &DeckContext{
		DeckID:        "test-deck",
		Format:        "standard",
		ColorIdentity: []string{"R"},
		Archetype:     "aggro",
	}

	scores, err := scorer.ScoreCardsWithMeta(ctx, "standard", []int{1, 2, 3}, deck, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(scores) != 3 {
		t.Errorf("expected 3 scores, got %d", len(scores))
	}

	// All scores should have meta-based values
	for _, score := range scores {
		if score.Score < 0 || score.Score > 1 {
			t.Errorf("score out of range: %f", score.Score)
		}
	}
}

func TestMetaWeighter_GetConfig(t *testing.T) {
	config := &MetaWeightingConfig{
		Tier1Weight: 0.9,
		CacheTTL:    2 * time.Hour,
	}
	weighter := NewMetaWeighter(nil, config)

	retrievedConfig := weighter.GetConfig()
	if retrievedConfig.Tier1Weight != 0.9 {
		t.Errorf("expected Tier1Weight 0.9, got %f", retrievedConfig.Tier1Weight)
	}
	if retrievedConfig.CacheTTL != 2*time.Hour {
		t.Errorf("expected CacheTTL 2h, got %v", retrievedConfig.CacheTTL)
	}
}
