package embeddings

import (
	"testing"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

func TestGenerator_GenerateEmbedding(t *testing.T) {
	generator := NewGenerator()

	tests := []struct {
		name    string
		card    *CardData
		checkFn func(t *testing.T, emb *models.CardEmbedding)
	}{
		{
			name: "basic creature with colors",
			card: &CardData{
				ArenaID:    12345,
				Name:       "Llanowar Elves",
				ManaCost:   "{G}",
				CMC:        1,
				TypeLine:   "Creature — Elf Druid",
				Colors:     []string{"G"},
				OracleText: "{T}: Add {G}.",
				Power:      "1",
				Toughness:  "1",
				Rarity:     "common",
			},
			checkFn: func(t *testing.T, emb *models.CardEmbedding) {
				// Check color encoding (position 4 = Green)
				if emb.Embedding[4] == 0 {
					t.Error("expected green color to be encoded")
				}
				// Check CMC bucket (position 5+1 = 6 for CMC 1)
				if emb.Embedding[6] == 0 {
					t.Error("expected CMC 1 to be encoded at position 6")
				}
				// Check creature type (position 13)
				if emb.Embedding[13] == 0 {
					t.Error("expected creature type to be encoded")
				}
				// Check common rarity (position 21)
				if emb.Embedding[21] == 0 {
					t.Error("expected common rarity to be encoded")
				}
			},
		},
		{
			name: "multicolor creature with keywords",
			card: &CardData{
				ArenaID:    67890,
				Name:       "Vampire Nighthawk",
				ManaCost:   "{1}{B}{B}",
				CMC:        3,
				TypeLine:   "Creature — Vampire Shaman",
				Colors:     []string{"B"},
				OracleText: "Flying, deathtouch, lifelink",
				Power:      "2",
				Toughness:  "3",
				Rarity:     "uncommon",
			},
			checkFn: func(t *testing.T, emb *models.CardEmbedding) {
				// Check black color (position 2)
				if emb.Embedding[2] == 0 {
					t.Error("expected black color to be encoded")
				}
				// Check CMC bucket (position 5+3 = 8 for CMC 3)
				if emb.Embedding[8] == 0 {
					t.Error("expected CMC 3 to be encoded at position 8")
				}
				// Check flying keyword (position 35)
				if emb.Embedding[35] == 0 {
					t.Error("expected flying keyword to be encoded")
				}
				// Check uncommon rarity (position 22)
				if emb.Embedding[22] == 0 {
					t.Error("expected uncommon rarity to be encoded")
				}
			},
		},
		{
			name: "instant spell",
			card: &CardData{
				ArenaID:    11111,
				Name:       "Lightning Bolt",
				ManaCost:   "{R}",
				CMC:        1,
				TypeLine:   "Instant",
				Colors:     []string{"R"},
				OracleText: "Lightning Bolt deals 3 damage to any target.",
				Rarity:     "common",
			},
			checkFn: func(t *testing.T, emb *models.CardEmbedding) {
				// Check red color (position 3)
				if emb.Embedding[3] == 0 {
					t.Error("expected red color to be encoded")
				}
				// Check instant type (position 14)
				if emb.Embedding[14] == 0 {
					t.Error("expected instant type to be encoded")
				}
			},
		},
		{
			name: "colorless artifact",
			card: &CardData{
				ArenaID:    22222,
				Name:       "Sol Ring",
				ManaCost:   "{1}",
				CMC:        1,
				TypeLine:   "Artifact",
				Colors:     []string{},
				OracleText: "{T}: Add {C}{C}.",
				Rarity:     "uncommon",
			},
			checkFn: func(t *testing.T, emb *models.CardEmbedding) {
				// Check no colors encoded
				for i := 0; i < 5; i++ {
					if emb.Embedding[i] != 0 {
						t.Errorf("expected no color at position %d, got %f", i, emb.Embedding[i])
					}
				}
				// Check artifact type (position 17)
				if emb.Embedding[17] == 0 {
					t.Error("expected artifact type to be encoded")
				}
			},
		},
		{
			name: "high CMC creature",
			card: &CardData{
				ArenaID:    33333,
				Name:       "Ulamog, the Ceaseless Hunger",
				ManaCost:   "{10}",
				CMC:        10,
				TypeLine:   "Legendary Creature — Eldrazi",
				Colors:     []string{},
				OracleText: "When you cast this spell, exile two target permanents. Indestructible. Whenever Ulamog attacks, defending player exiles the top twenty cards of their library.",
				Power:      "10",
				Toughness:  "10",
				Rarity:     "mythic",
			},
			checkFn: func(t *testing.T, emb *models.CardEmbedding) {
				// Check CMC 7+ bucket (position 12)
				if emb.Embedding[12] == 0 {
					t.Error("expected CMC 7+ to be encoded at position 12")
				}
				// Check mythic rarity (position 24)
				if emb.Embedding[24] == 0 {
					t.Error("expected mythic rarity to be encoded")
				}
				// Check indestructible keyword
				if emb.Embedding[47] == 0 {
					t.Error("expected indestructible keyword to be encoded")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			emb := generator.GenerateEmbedding(tt.card)

			// Basic checks
			if emb == nil {
				t.Fatal("expected embedding to not be nil")
			}
			if emb.ArenaID != tt.card.ArenaID {
				t.Errorf("expected ArenaID %d, got %d", tt.card.ArenaID, emb.ArenaID)
			}
			if emb.CardName != tt.card.Name {
				t.Errorf("expected CardName %s, got %s", tt.card.Name, emb.CardName)
			}
			if len(emb.Embedding) != models.EmbeddingDimensions {
				t.Errorf("expected %d dimensions, got %d", models.EmbeddingDimensions, len(emb.Embedding))
			}
			if emb.EmbeddingVersion != models.EmbeddingVersion {
				t.Errorf("expected version %d, got %d", models.EmbeddingVersion, emb.EmbeddingVersion)
			}
			if emb.Source != models.EmbeddingSourceCharacteristics {
				t.Errorf("expected source %s, got %s", models.EmbeddingSourceCharacteristics, emb.Source)
			}

			// Check embedding is normalized (L2 norm should be approximately 1)
			var sumSquares float64
			for _, v := range emb.Embedding {
				sumSquares += v * v
			}
			// Allow small floating point error
			if sumSquares < 0.99 || sumSquares > 1.01 {
				t.Errorf("expected normalized embedding (L2 norm ~1), got %f", sumSquares)
			}

			// Run test-specific checks
			if tt.checkFn != nil {
				tt.checkFn(t, emb)
			}
		})
	}
}

func TestCosineSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
		epsilon  float64
	}{
		{
			name:     "identical vectors",
			a:        []float64{1, 0, 0, 0},
			b:        []float64{1, 0, 0, 0},
			expected: 1.0,
			epsilon:  0.001,
		},
		{
			name:     "orthogonal vectors",
			a:        []float64{1, 0, 0, 0},
			b:        []float64{0, 1, 0, 0},
			expected: 0.0,
			epsilon:  0.001,
		},
		{
			name:     "opposite vectors",
			a:        []float64{1, 0, 0, 0},
			b:        []float64{-1, 0, 0, 0},
			expected: -1.0,
			epsilon:  0.001,
		},
		{
			name:     "similar vectors",
			a:        []float64{1, 1, 0, 0},
			b:        []float64{1, 0.5, 0, 0},
			expected: 0.949,
			epsilon:  0.01,
		},
		{
			name:     "different length vectors",
			a:        []float64{1, 0, 0},
			b:        []float64{1, 0, 0, 0},
			expected: 0.0,
			epsilon:  0.001,
		},
		{
			name:     "zero vectors",
			a:        []float64{0, 0, 0, 0},
			b:        []float64{0, 0, 0, 0},
			expected: 0.0,
			epsilon:  0.001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CosineSimilarity(tt.a, tt.b)
			diff := result - tt.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.epsilon {
				t.Errorf("expected %f, got %f (diff: %f)", tt.expected, result, diff)
			}
		})
	}
}

func TestEuclideanDistance(t *testing.T) {
	tests := []struct {
		name     string
		a        []float64
		b        []float64
		expected float64
		epsilon  float64
	}{
		{
			name:     "identical vectors",
			a:        []float64{1, 0, 0, 0},
			b:        []float64{1, 0, 0, 0},
			expected: 0.0,
			epsilon:  0.001,
		},
		{
			name:     "unit distance",
			a:        []float64{0, 0, 0, 0},
			b:        []float64{1, 0, 0, 0},
			expected: 1.0,
			epsilon:  0.001,
		},
		{
			name:     "diagonal distance",
			a:        []float64{0, 0, 0},
			b:        []float64{1, 1, 1},
			expected: 1.732,
			epsilon:  0.01,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EuclideanDistance(tt.a, tt.b)
			diff := result - tt.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > tt.epsilon {
				t.Errorf("expected %f, got %f (diff: %f)", tt.expected, result, diff)
			}
		})
	}
}

func TestExtractKeywords(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected []string
	}{
		{
			name:     "single keyword",
			text:     "Flying",
			expected: []string{"flying"},
		},
		{
			name:     "multiple keywords",
			text:     "Flying, deathtouch, lifelink",
			expected: []string{"flying", "deathtouch", "lifelink"},
		},
		{
			name:     "keyword in sentence",
			text:     "This creature has flying and first strike.",
			expected: []string{"flying", "first strike"},
		},
		{
			name:     "no keywords",
			text:     "Draw a card.",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ExtractKeywords(tt.text)

			if len(tt.expected) == 0 && len(result) == 0 {
				return // Both empty, test passes
			}

			// Check all expected keywords are found
			for _, exp := range tt.expected {
				found := false
				for _, res := range result {
					if res == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected keyword %q not found in result %v", exp, result)
				}
			}
		})
	}
}

func TestGenerator_SimilarCardsSimilarity(t *testing.T) {
	generator := NewGenerator()

	// Create two similar creatures (both green creatures with similar stats)
	card1 := &CardData{
		ArenaID:    1,
		Name:       "Grizzly Bears",
		ManaCost:   "{1}{G}",
		CMC:        2,
		TypeLine:   "Creature — Bear",
		Colors:     []string{"G"},
		OracleText: "",
		Power:      "2",
		Toughness:  "2",
		Rarity:     "common",
	}

	card2 := &CardData{
		ArenaID:    2,
		Name:       "Runeclaw Bear",
		ManaCost:   "{1}{G}",
		CMC:        2,
		TypeLine:   "Creature — Bear",
		Colors:     []string{"G"},
		OracleText: "",
		Power:      "2",
		Toughness:  "2",
		Rarity:     "common",
	}

	// Create a very different card (blue instant, high CMC)
	card3 := &CardData{
		ArenaID:    3,
		Name:       "Time Walk",
		ManaCost:   "{1}{U}",
		CMC:        2,
		TypeLine:   "Sorcery",
		Colors:     []string{"U"},
		OracleText: "Take an extra turn after this one.",
		Rarity:     "rare",
	}

	emb1 := generator.GenerateEmbedding(card1)
	emb2 := generator.GenerateEmbedding(card2)
	emb3 := generator.GenerateEmbedding(card3)

	sim12 := CosineSimilarity(emb1.Embedding, emb2.Embedding)
	sim13 := CosineSimilarity(emb1.Embedding, emb3.Embedding)

	// Similar cards should have higher similarity
	if sim12 <= sim13 {
		t.Errorf("expected similar cards (%.4f) to have higher similarity than different cards (%.4f)", sim12, sim13)
	}

	// Similar cards should have high similarity (> 0.8)
	if sim12 < 0.8 {
		t.Errorf("expected similar cards to have similarity > 0.8, got %.4f", sim12)
	}
}
