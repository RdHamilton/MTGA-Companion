package recommendations

import (
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
)

// Archetype represents a deck archetype category.
type Archetype string

const (
	ArchetypeAggro     Archetype = "Aggro"
	ArchetypeControl   Archetype = "Control"
	ArchetypeMidrange  Archetype = "Midrange"
	ArchetypeCombo     Archetype = "Combo"
	ArchetypeTempo     Archetype = "Tempo"
	ArchetypeRamp      Archetype = "Ramp"
	ArchetypeTribal    Archetype = "Tribal"
	ArchetypeTokens    Archetype = "Tokens"
	ArchetypeArtifacts Archetype = "Artifacts"
)

// ArchetypeScore represents how well a deck matches an archetype.
type ArchetypeScore struct {
	Archetype   Archetype
	Score       float64  // 0.0-1.0
	Confidence  float64  // How confident we are in this classification
	Signals     []string // What indicates this archetype
	Description string   // Human-readable description
}

// ArchetypeSignal defines a signal that indicates an archetype.
type ArchetypeSignal struct {
	Name      string   // Signal name for debugging
	Patterns  []string // Oracle text patterns
	TypeLines []string // Type line patterns
	Keywords  []string // Keyword abilities
	MinCMC    float64  // Minimum average CMC (0 = no min)
	MaxCMC    float64  // Maximum average CMC (0 = no max)
	MinCount  int      // Minimum count of matching cards
	Weight    float64  // How much this signal contributes
}

// archetypeSignals defines signals for each archetype.
var archetypeSignals = map[Archetype][]ArchetypeSignal{
	ArchetypeAggro: {
		{Name: "low curve", MaxCMC: 2.5, Weight: 0.25},
		{Name: "haste creatures", Keywords: []string{"haste"}, MinCount: 3, Weight: 0.15},
		{Name: "direct damage", Patterns: []string{"deals.*damage to any target", "deals.*damage to.*player", "deals.*damage to.*opponent"}, MinCount: 3, Weight: 0.2},
		{Name: "cheap creatures", TypeLines: []string{"Creature"}, MaxCMC: 2, MinCount: 12, Weight: 0.2},
		{Name: "burn spells", Patterns: []string{"deals.*damage"}, TypeLines: []string{"Instant", "Sorcery"}, MinCount: 4, Weight: 0.2},
	},
	ArchetypeControl: {
		{Name: "high curve", MinCMC: 3.0, Weight: 0.15},
		{Name: "counterspells", Patterns: []string{"counter target spell", "counter target.*spell"}, MinCount: 4, Weight: 0.25},
		{Name: "removal", Patterns: []string{"destroy target", "exile target", "deals.*damage to.*creature"}, MinCount: 6, Weight: 0.2},
		{Name: "card draw", Patterns: []string{"draw.*card", "draw two", "draw three"}, MinCount: 4, Weight: 0.15},
		{Name: "board wipes", Patterns: []string{"destroy all", "exile all", "deals.*damage to each"}, MinCount: 2, Weight: 0.15},
		{Name: "few creatures", TypeLines: []string{"Creature"}, MinCount: 0, Weight: 0.1}, // Handled specially
	},
	ArchetypeMidrange: {
		{Name: "balanced curve", MinCMC: 2.5, MaxCMC: 3.5, Weight: 0.2},
		{Name: "value creatures", Patterns: []string{"when.*enters the battlefield", "when.*enters,"}, TypeLines: []string{"Creature"}, MinCount: 6, Weight: 0.25},
		{Name: "removal suite", Patterns: []string{"destroy target", "exile target"}, MinCount: 4, Weight: 0.15},
		{Name: "card advantage", Patterns: []string{"draw a card", "draw cards"}, MinCount: 3, Weight: 0.15},
		{Name: "efficient threats", TypeLines: []string{"Creature"}, MinCount: 12, Weight: 0.15},
		{Name: "planeswalkers", TypeLines: []string{"Planeswalker"}, MinCount: 2, Weight: 0.1},
	},
	ArchetypeCombo: {
		{Name: "tutors", Patterns: []string{"search your library for"}, MinCount: 3, Weight: 0.3},
		{Name: "infinite mana", Patterns: []string{"untap.*land", "add.*mana.*for each"}, MinCount: 2, Weight: 0.2},
		{Name: "card selection", Patterns: []string{"scry", "surveil", "look at the top"}, MinCount: 4, Weight: 0.15},
		{Name: "protection", Patterns: []string{"hexproof", "can't be countered"}, MinCount: 2, Weight: 0.15},
		{Name: "recursion", Patterns: []string{"return.*from.*graveyard", "cast.*from.*graveyard"}, MinCount: 3, Weight: 0.2},
	},
	ArchetypeTempo: {
		{Name: "cheap threats", TypeLines: []string{"Creature"}, MaxCMC: 2, MinCount: 8, Weight: 0.2},
		{Name: "counterspells", Patterns: []string{"counter target"}, MinCount: 4, Weight: 0.25},
		{Name: "bounce", Patterns: []string{"return.*to.*owner's hand", "return target.*to its owner's hand"}, MinCount: 3, Weight: 0.2},
		{Name: "flash", Keywords: []string{"flash"}, MinCount: 4, Weight: 0.15},
		{Name: "card draw", Patterns: []string{"draw a card"}, MinCount: 3, Weight: 0.1},
		{Name: "low curve", MaxCMC: 2.8, Weight: 0.1},
	},
	ArchetypeRamp: {
		{Name: "mana dorks", Patterns: []string{"add.*mana", "add {g}", "add one mana"}, TypeLines: []string{"Creature"}, MinCount: 4, Weight: 0.25},
		{Name: "land search", Patterns: []string{"search your library for.*land", "search your library for a basic land"}, MinCount: 3, Weight: 0.2},
		{Name: "big finishers", TypeLines: []string{"Creature"}, MinCount: 4, Weight: 0.2}, // CMC >= 5, handled specially
		{Name: "treasure tokens", Patterns: []string{"create a treasure", "treasure token"}, MinCount: 3, Weight: 0.15},
		{Name: "high top end", MinCMC: 3.5, Weight: 0.2},
	},
	ArchetypeTribal: {
		{Name: "lords", Patterns: []string{"other.*you control get +", "creatures you control have"}, MinCount: 2, Weight: 0.3},
		{Name: "tribal synergy", Patterns: []string{"for each.*you control", "whenever.*you control"}, MinCount: 3, Weight: 0.25},
		{Name: "creature focus", TypeLines: []string{"Creature"}, MinCount: 20, Weight: 0.25},
		{Name: "type matters", Patterns: []string{"other.*creatures you control"}, MinCount: 2, Weight: 0.2},
	},
	ArchetypeTokens: {
		{Name: "token makers", Patterns: []string{"create a.*token", "create.*tokens"}, MinCount: 8, Weight: 0.3},
		{Name: "anthems", Patterns: []string{"creatures you control get +", "other creatures you control get +"}, MinCount: 3, Weight: 0.25},
		{Name: "token payoffs", Patterns: []string{"for each creature you control", "equal to the number of creatures"}, MinCount: 2, Weight: 0.2},
		{Name: "go wide", Patterns: []string{"whenever.*creature.*enters", "whenever a token"}, MinCount: 3, Weight: 0.15},
		{Name: "sacrifice", Patterns: []string{"sacrifice a creature", "sacrifice a token"}, MinCount: 2, Weight: 0.1},
	},
	ArchetypeArtifacts: {
		{Name: "artifact producers", Patterns: []string{"create.*artifact token", "create a treasure", "create a clue", "create a food"}, MinCount: 5, Weight: 0.25},
		{Name: "artifact payoffs", Patterns: []string{"whenever an artifact", "for each artifact"}, MinCount: 3, Weight: 0.25},
		{Name: "artifact creatures", TypeLines: []string{"Artifact Creature"}, MinCount: 6, Weight: 0.2},
		{Name: "artifact synergy", Patterns: []string{"artifact you control", "artifacts you control"}, MinCount: 3, Weight: 0.2},
		{Name: "affinity", Keywords: []string{"affinity"}, MinCount: 2, Weight: 0.1},
	},
}

// archetypeDescriptions provides human-readable descriptions.
var archetypeDescriptions = map[Archetype]string{
	ArchetypeAggro:     "Fast, aggressive deck that wins by dealing damage quickly",
	ArchetypeControl:   "Reactive deck that answers threats and wins in the late game",
	ArchetypeMidrange:  "Flexible deck with efficient threats and removal",
	ArchetypeCombo:     "Deck built around specific card combinations",
	ArchetypeTempo:     "Deck that deploys threats while disrupting the opponent",
	ArchetypeRamp:      "Deck that accelerates mana to cast big spells early",
	ArchetypeTribal:    "Deck focused on creature type synergies",
	ArchetypeTokens:    "Deck that creates many creature tokens",
	ArchetypeArtifacts: "Deck built around artifact synergies",
}

// DeckAnalysisForArchetype contains deck statistics for archetype classification.
type DeckAnalysisForArchetype struct {
	AverageCMC        float64
	CreatureCount     int
	InstantCount      int
	SorceryCount      int
	ArtifactCount     int
	EnchantmentCount  int
	PlaneswalkerCount int
	LandCount         int
	TotalCards        int
	HighCMCCount      int // Cards with CMC >= 5
	LowCMCCount       int // Cards with CMC <= 2
}

// ClassifyDeck analyzes a deck and returns archetype scores.
func ClassifyDeck(deckCards []*cards.Card) []ArchetypeScore {
	if len(deckCards) == 0 {
		return nil
	}

	// Analyze deck composition
	analysis := analyzeDeckComposition(deckCards)

	// Score each archetype
	var scores []ArchetypeScore
	for archetype, signals := range archetypeSignals {
		score, matchedSignals := scoreArchetype(deckCards, analysis, archetype, signals)
		if score > 0.2 { // Only include if there's meaningful signal
			scores = append(scores, ArchetypeScore{
				Archetype:   archetype,
				Score:       score,
				Confidence:  calculateArchetypeConfidence(score, len(matchedSignals)),
				Signals:     matchedSignals,
				Description: archetypeDescriptions[archetype],
			})
		}
	}

	// Sort by score descending
	for i := 0; i < len(scores)-1; i++ {
		for j := i + 1; j < len(scores); j++ {
			if scores[j].Score > scores[i].Score {
				scores[i], scores[j] = scores[j], scores[i]
			}
		}
	}

	return scores
}

// analyzeDeckComposition calculates deck statistics.
func analyzeDeckComposition(deckCards []*cards.Card) *DeckAnalysisForArchetype {
	analysis := &DeckAnalysisForArchetype{}
	totalCMC := 0.0
	nonLandCount := 0

	for _, card := range deckCards {
		typeLine := strings.ToLower(card.TypeLine)
		analysis.TotalCards++

		if strings.Contains(typeLine, "land") {
			analysis.LandCount++
			continue
		}

		nonLandCount++
		totalCMC += card.CMC

		if card.CMC >= 5 {
			analysis.HighCMCCount++
		}
		if card.CMC <= 2 {
			analysis.LowCMCCount++
		}

		if strings.Contains(typeLine, "creature") {
			analysis.CreatureCount++
		}
		if strings.Contains(typeLine, "instant") {
			analysis.InstantCount++
		}
		if strings.Contains(typeLine, "sorcery") {
			analysis.SorceryCount++
		}
		if strings.Contains(typeLine, "artifact") && !strings.Contains(typeLine, "creature") {
			analysis.ArtifactCount++
		}
		if strings.Contains(typeLine, "enchantment") {
			analysis.EnchantmentCount++
		}
		if strings.Contains(typeLine, "planeswalker") {
			analysis.PlaneswalkerCount++
		}
	}

	if nonLandCount > 0 {
		analysis.AverageCMC = totalCMC / float64(nonLandCount)
	}

	return analysis
}

// scoreArchetype calculates how well a deck matches an archetype.
func scoreArchetype(deckCards []*cards.Card, analysis *DeckAnalysisForArchetype, archetype Archetype, signals []ArchetypeSignal) (float64, []string) {
	totalWeight := 0.0
	weightedScore := 0.0
	var matchedSignals []string

	for _, signal := range signals {
		totalWeight += signal.Weight
		signalScore := evaluateSignal(deckCards, analysis, &signal, archetype)
		if signalScore > 0 {
			weightedScore += signalScore * signal.Weight
			matchedSignals = append(matchedSignals, signal.Name)
		}
	}

	if totalWeight == 0 {
		return 0, nil
	}

	return weightedScore / totalWeight, matchedSignals
}

// evaluateSignal checks if a signal is present in the deck.
func evaluateSignal(deckCards []*cards.Card, analysis *DeckAnalysisForArchetype, signal *ArchetypeSignal, archetype Archetype) float64 {
	// Check CMC constraints
	if signal.MinCMC > 0 && analysis.AverageCMC < signal.MinCMC {
		return 0
	}
	if signal.MaxCMC > 0 && analysis.AverageCMC > signal.MaxCMC {
		return 0
	}

	// Special case: "few creatures" for control
	if signal.Name == "few creatures" {
		if analysis.CreatureCount <= 10 {
			return 1.0
		}
		return 0
	}

	// Special case: "big finishers" for ramp
	if signal.Name == "big finishers" {
		if analysis.HighCMCCount >= 4 {
			return 1.0
		}
		return 0
	}

	// Count matching cards
	matchCount := 0
	for _, card := range deckCards {
		if matchesSignal(card, signal) {
			matchCount++
		}
	}

	// Check minimum count
	if signal.MinCount > 0 && matchCount < signal.MinCount {
		return 0
	}

	// If we're just checking CMC and it passed, return 1.0
	if len(signal.Patterns) == 0 && len(signal.TypeLines) == 0 && len(signal.Keywords) == 0 {
		return 1.0
	}

	// Scale score based on how many matches we found
	if signal.MinCount > 0 {
		// More matches = higher score, up to 1.0
		return min(float64(matchCount)/float64(signal.MinCount*2), 1.0)
	}

	return 1.0
}

// matchesSignal checks if a card matches a signal's criteria.
func matchesSignal(card *cards.Card, signal *ArchetypeSignal) bool {
	oracleText := ""
	if card.OracleText != nil {
		oracleText = strings.ToLower(*card.OracleText)
	}
	typeLine := strings.ToLower(card.TypeLine)

	// Check type line requirements
	typeMatch := len(signal.TypeLines) == 0
	for _, requiredType := range signal.TypeLines {
		if strings.Contains(typeLine, strings.ToLower(requiredType)) {
			typeMatch = true
			break
		}
	}
	if !typeMatch {
		return false
	}

	// Check CMC requirements for type-based signals
	if signal.MaxCMC > 0 && len(signal.TypeLines) > 0 && card.CMC > signal.MaxCMC {
		return false
	}

	// Check pattern requirements
	if len(signal.Patterns) > 0 {
		patternMatch := false
		for _, pattern := range signal.Patterns {
			if containsPattern(oracleText, pattern) {
				patternMatch = true
				break
			}
		}
		if !patternMatch {
			return false
		}
	}

	// Check keyword requirements
	if len(signal.Keywords) > 0 {
		keywordMatch := false
		for _, keyword := range signal.Keywords {
			if strings.Contains(oracleText, strings.ToLower(keyword)) {
				keywordMatch = true
				break
			}
		}
		if !keywordMatch {
			return false
		}
	}

	return true
}

// calculateArchetypeConfidence determines how confident we are in the classification.
func calculateArchetypeConfidence(score float64, signalCount int) float64 {
	// Higher score and more signals = more confidence
	baseConfidence := score
	signalBonus := min(float64(signalCount)*0.1, 0.3)
	return min(baseConfidence+signalBonus, 1.0)
}

// GetPrimaryArchetype returns the most likely archetype for a deck.
func GetPrimaryArchetype(deckCards []*cards.Card) *ArchetypeScore {
	scores := ClassifyDeck(deckCards)
	if len(scores) == 0 {
		return nil
	}
	return &scores[0]
}

// GetArchetypeDescription returns a human-readable description of an archetype.
func GetArchetypeDescription(archetype Archetype) string {
	if desc, ok := archetypeDescriptions[archetype]; ok {
		return desc
	}
	return "Unknown archetype"
}

// IsAggroArchetype returns true if the archetype is aggressive.
func IsAggroArchetype(archetype Archetype) bool {
	return archetype == ArchetypeAggro || archetype == ArchetypeTempo
}

// IsControlArchetype returns true if the archetype is controlling.
func IsControlArchetype(archetype Archetype) bool {
	return archetype == ArchetypeControl
}

// min returns the smaller of two float64 values.
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
