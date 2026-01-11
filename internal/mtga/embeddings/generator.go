package embeddings

import (
	"math"
	"regexp"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// CardData represents the card data needed for embedding generation.
type CardData struct {
	ArenaID    int
	Name       string
	ManaCost   string
	CMC        float64
	TypeLine   string
	Colors     []string
	OracleText string
	Power      string
	Toughness  string
	Rarity     string
}

// Generator creates card embeddings from card characteristics.
type Generator struct{}

// NewGenerator creates a new embedding generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateEmbedding creates a 64-dimensional embedding from card data.
// Dimensions breakdown:
// - [0-4]: Color identity (W, U, B, R, G) - 5D
// - [5-12]: CMC bucketed (0, 1, 2, 3, 4, 5, 6, 7+) - 8D
// - [13-20]: Card types (creature, instant, sorcery, enchantment, artifact, planeswalker, land, other) - 8D
// - [21-24]: Rarity (common, uncommon, rare, mythic) - 4D
// - [25-34]: Power/Toughness buckets (0, 1, 2, 3, 4, 5, 6, 7+, X, N/A) - 10D each but compressed to 5D
// - [35-63]: Keywords (29 common keywords) - 29D
func (g *Generator) GenerateEmbedding(card *CardData) *models.CardEmbedding {
	embedding := make([]float64, models.EmbeddingDimensions)

	// Color identity (0-4)
	g.encodeColors(embedding[0:5], card.Colors)

	// CMC buckets (5-12)
	g.encodeCMC(embedding[5:13], card.CMC)

	// Card types (13-20)
	g.encodeTypes(embedding[13:21], card.TypeLine)

	// Rarity (21-24)
	g.encodeRarity(embedding[21:25], card.Rarity)

	// Power/Toughness (25-34)
	g.encodePowerToughness(embedding[25:35], card.Power, card.Toughness)

	// Keywords from oracle text (35-63)
	g.encodeKeywords(embedding[35:64], card.OracleText)

	// Normalize the embedding
	g.normalize(embedding)

	return &models.CardEmbedding{
		ArenaID:          card.ArenaID,
		CardName:         card.Name,
		Embedding:        embedding,
		EmbeddingVersion: models.EmbeddingVersion,
		Source:           models.EmbeddingSourceCharacteristics,
	}
}

// encodeColors sets color identity flags.
func (g *Generator) encodeColors(vec []float64, colors []string) {
	colorMap := map[string]int{"W": 0, "U": 1, "B": 2, "R": 3, "G": 4}
	for _, c := range colors {
		if idx, ok := colorMap[strings.ToUpper(c)]; ok {
			vec[idx] = 1.0
		}
	}
}

// encodeCMC buckets the converted mana cost.
func (g *Generator) encodeCMC(vec []float64, cmc float64) {
	idx := int(cmc)
	if idx > 7 {
		idx = 7
	}
	if idx < 0 {
		idx = 0
	}
	vec[idx] = 1.0
}

// encodeTypes extracts card types from the type line.
func (g *Generator) encodeTypes(vec []float64, typeLine string) {
	lower := strings.ToLower(typeLine)

	typeChecks := []struct {
		keyword string
		index   int
	}{
		{"creature", 0},
		{"instant", 1},
		{"sorcery", 2},
		{"enchantment", 3},
		{"artifact", 4},
		{"planeswalker", 5},
		{"land", 6},
	}

	foundType := false
	for _, tc := range typeChecks {
		if strings.Contains(lower, tc.keyword) {
			vec[tc.index] = 1.0
			foundType = true
		}
	}

	// "Other" type (tribal, battle, etc.)
	if !foundType {
		vec[7] = 1.0
	}
}

// encodeRarity sets rarity flag.
func (g *Generator) encodeRarity(vec []float64, rarity string) {
	rarityMap := map[string]int{
		"common":   0,
		"uncommon": 1,
		"rare":     2,
		"mythic":   3,
	}
	if idx, ok := rarityMap[strings.ToLower(rarity)]; ok {
		vec[idx] = 1.0
	}
}

// encodePowerToughness encodes creature stats.
func (g *Generator) encodePowerToughness(vec []float64, power, toughness string) {
	// Power (0-4)
	g.encodeStatValue(vec[0:5], power)
	// Toughness (5-9)
	g.encodeStatValue(vec[5:10], toughness)
}

// encodeStatValue converts a power/toughness value to bucketed encoding.
func (g *Generator) encodeStatValue(vec []float64, value string) {
	if value == "" || value == "*" {
		// Variable or N/A - spread across buckets
		for i := range vec {
			vec[i] = 0.2
		}
		return
	}

	// Handle X values
	if strings.Contains(value, "X") || strings.Contains(value, "x") {
		vec[4] = 1.0 // Last bucket for variable
		return
	}

	// Parse numeric value
	val := 0
	for _, c := range value {
		if c >= '0' && c <= '9' {
			val = val*10 + int(c-'0')
		}
	}

	// Bucket: 0-1, 2-3, 4-5, 6-7, 8+
	bucket := val / 2
	if bucket > 4 {
		bucket = 4
	}
	vec[bucket] = 1.0
}

// Common MTG keywords for embedding.
var commonKeywords = []string{
	"flying", "trample", "haste", "vigilance", "lifelink",
	"deathtouch", "first strike", "double strike", "menace", "reach",
	"flash", "hexproof", "indestructible", "defender", "protection",
	"ward", "prowess", "scry", "surveil", "draw",
	"counter", "destroy", "exile", "return", "sacrifice",
	"token", "enters", "dies", "graveyard",
}

// encodeKeywords extracts keyword abilities from oracle text.
func (g *Generator) encodeKeywords(vec []float64, oracleText string) {
	lower := strings.ToLower(oracleText)

	for i, keyword := range commonKeywords {
		if i >= len(vec) {
			break
		}
		if strings.Contains(lower, keyword) {
			vec[i] = 1.0
		}
	}
}

// normalize applies L2 normalization to the embedding.
func (g *Generator) normalize(vec []float64) {
	var sumSquares float64
	for _, v := range vec {
		sumSquares += v * v
	}

	if sumSquares == 0 {
		return
	}

	norm := math.Sqrt(sumSquares)
	for i := range vec {
		vec[i] /= norm
	}
}

// CosineSimilarity computes the cosine similarity between two embeddings.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float64
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// EuclideanDistance computes the Euclidean distance between two embeddings.
func EuclideanDistance(a, b []float64) float64 {
	if len(a) != len(b) {
		return math.MaxFloat64
	}

	var sumSquares float64
	for i := range a {
		diff := a[i] - b[i]
		sumSquares += diff * diff
	}

	return math.Sqrt(sumSquares)
}

// ExtractKeywords extracts keyword abilities from oracle text.
func ExtractKeywords(oracleText string) []string {
	lower := strings.ToLower(oracleText)
	var found []string

	// Pattern for keyword abilities (usually at the start or as standalone words)
	keywordPattern := regexp.MustCompile(`\b(flying|trample|haste|vigilance|lifelink|deathtouch|first strike|double strike|menace|reach|flash|hexproof|indestructible|defender|protection|ward|prowess)\b`)

	matches := keywordPattern.FindAllString(lower, -1)
	seen := make(map[string]bool)
	for _, m := range matches {
		if !seen[m] {
			found = append(found, m)
			seen[m] = true
		}
	}

	return found
}
