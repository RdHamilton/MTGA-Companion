package analysis

import (
	"context"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// AnalysisResult contains findings from play pattern analysis.
type AnalysisResult struct {
	DeckID           string
	TotalGames       int
	Wins             int
	Losses           int
	WinRate          float64
	AvgGameLength    float64 // Average turns per game
	MulliganRate     float64 // Percentage of games with mulligan
	LandDropMissRate float64 // Percentage of games with missed land drops
	CurveAnalysis    CurveAnalysis
	ManaAnalysis     ManaAnalysis
	SequencingIssues []SequencingIssue
}

// CurveAnalysis examines mana curve performance.
type CurveAnalysis struct {
	AvgFirstPlay     float64     // Average turn of first non-land spell
	EmptyTurns       float64     // Average turns with no plays
	FloodedGames     int         // Games with 7+ lands and few spells in hand
	ScrewedGames     int         // Games stuck on 2-3 lands by turn 5
	CMCDistribution  map[int]int // CMC -> count of plays at that CMC
	CurveSuggestions []string    // Generated curve improvement suggestions
}

// ManaAnalysis examines mana base performance.
type ManaAnalysis struct {
	ColorScrew    int     // Games where needed color wasn't available
	ManaFlood     int     // Games with 7+ lands in play
	ManaScrew     int     // Games stuck below 3 lands by turn 5
	AvgLandDrops  float64 // Average lands played per game
	DualLandUsage float64 // Percentage of games where dual lands helped fix colors
}

// SequencingIssue represents a potential play order mistake.
type SequencingIssue struct {
	MatchID     string
	Turn        int
	Description string
	CardID      *int
	CardName    *string
}

// PlayAnalyzer analyzes game_plays data to find patterns and generate improvement suggestions.
type PlayAnalyzer struct {
	playRepo  repository.GamePlayRepository
	matchRepo repository.MatchRepository
}

// NewPlayAnalyzer creates a new play analyzer.
func NewPlayAnalyzer(playRepo repository.GamePlayRepository, matchRepo repository.MatchRepository) *PlayAnalyzer {
	return &PlayAnalyzer{
		playRepo:  playRepo,
		matchRepo: matchRepo,
	}
}

// AnalyzeDeck analyzes play patterns for a deck across multiple matches.
// minGames specifies the minimum number of games required for meaningful analysis.
func (a *PlayAnalyzer) AnalyzeDeck(ctx context.Context, deckID string, minGames int) (*AnalysisResult, error) {
	// Get all matches for this deck using StatsFilter
	filter := models.StatsFilter{
		DeckID: &deckID,
	}
	matches, err := a.matchRepo.GetMatches(ctx, filter)
	if err != nil {
		return nil, err
	}

	if len(matches) < minGames {
		return &AnalysisResult{
			DeckID:     deckID,
			TotalGames: len(matches),
		}, nil
	}

	result := &AnalysisResult{
		DeckID:           deckID,
		TotalGames:       len(matches),
		SequencingIssues: []SequencingIssue{},
		CurveAnalysis: CurveAnalysis{
			CMCDistribution: make(map[int]int),
		},
	}

	var totalTurns int
	var gamesWithMulligan int
	var gamesWithMissedLandDrop int
	var firstPlayTurns []int
	var matchesWithPlayData int

	for _, match := range matches {
		// Count wins/losses
		if match.Result == "win" {
			result.Wins++
		} else {
			result.Losses++
		}

		// Analyze plays for this match
		plays, err := a.playRepo.GetPlaysByMatch(ctx, match.ID)
		if err != nil {
			// Log but continue - play data may not exist for all matches
			continue
		}

		snapshots, err := a.playRepo.GetSnapshotsByMatch(ctx, match.ID)
		if err != nil {
			// Snapshots are optional - continue with plays only
			snapshots = nil
		}

		// Track that we have play data for this match
		if len(plays) > 0 {
			matchesWithPlayData++
		}

		matchAnalysis := a.analyzeMatch(plays, snapshots)

		// Collect sequencing issues from this match
		for _, issue := range a.detectSequencingIssues(plays, snapshots) {
			issue.MatchID = match.ID
			result.SequencingIssues = append(result.SequencingIssues, issue)
		}

		// Aggregate results
		totalTurns += matchAnalysis.maxTurn

		if matchAnalysis.hadMulligan {
			gamesWithMulligan++
		}

		if matchAnalysis.missedLandDrop {
			gamesWithMissedLandDrop++
		}

		if matchAnalysis.firstPlayTurn > 0 {
			firstPlayTurns = append(firstPlayTurns, matchAnalysis.firstPlayTurn)
		}

		if matchAnalysis.manaScrew {
			result.ManaAnalysis.ManaScrew++
		}

		if matchAnalysis.manaFlood {
			result.ManaAnalysis.ManaFlood++
		}

		// Track CMC distribution
		for cmc, count := range matchAnalysis.cmcPlays {
			result.CurveAnalysis.CMCDistribution[cmc] += count
		}

		result.ManaAnalysis.AvgLandDrops += float64(matchAnalysis.landDrops)
	}

	// Calculate averages
	if result.TotalGames > 0 {
		result.WinRate = float64(result.Wins) / float64(result.TotalGames) * 100
		result.AvgGameLength = float64(totalTurns) / float64(result.TotalGames)
		result.MulliganRate = float64(gamesWithMulligan) / float64(result.TotalGames) * 100
		result.LandDropMissRate = float64(gamesWithMissedLandDrop) / float64(result.TotalGames) * 100
		result.ManaAnalysis.AvgLandDrops /= float64(result.TotalGames)
	}

	// Calculate average first play turn
	if len(firstPlayTurns) > 0 {
		var sum int
		for _, t := range firstPlayTurns {
			sum += t
		}
		result.CurveAnalysis.AvgFirstPlay = float64(sum) / float64(len(firstPlayTurns))
	}

	// Generate curve suggestions based on analysis
	result.CurveAnalysis.CurveSuggestions = a.generateCurveSuggestions(result)

	return result, nil
}

// matchAnalysisResult holds analysis for a single match.
type matchAnalysisResult struct {
	maxTurn        int
	hadMulligan    bool
	missedLandDrop bool
	firstPlayTurn  int
	manaScrew      bool
	manaFlood      bool
	landDrops      int
	cmcPlays       map[int]int
	emptyTurns     int
}

// analyzeMatch analyzes plays and snapshots for a single match.
func (a *PlayAnalyzer) analyzeMatch(plays []*models.GamePlay, snapshots []*models.GameStateSnapshot) *matchAnalysisResult {
	result := &matchAnalysisResult{
		cmcPlays: make(map[int]int),
	}

	// Track turns and plays
	playsByTurn := make(map[int][]*models.GamePlay)
	for _, play := range plays {
		if play.TurnNumber > result.maxTurn {
			result.maxTurn = play.TurnNumber
		}
		playsByTurn[play.TurnNumber] = append(playsByTurn[play.TurnNumber], play)

		// Check for mulligan
		if play.ActionType == "mulligan" && play.PlayerType == "player" {
			result.hadMulligan = true
		}

		// Track land drops
		if play.ActionType == "land_drop" && play.PlayerType == "player" {
			result.landDrops++
		}

		// Track first spell (non-land play)
		if play.ActionType == "play_card" && play.PlayerType == "player" && result.firstPlayTurn == 0 {
			result.firstPlayTurn = play.TurnNumber
		}
	}

	// Analyze snapshots for mana issues
	for _, snapshot := range snapshots {
		// Skip if lands data not available
		if snapshot.PlayerLandsInPlay == nil {
			continue
		}

		// Mana screw: stuck below 3 lands by turn 5
		if snapshot.TurnNumber >= 5 && *snapshot.PlayerLandsInPlay < 3 {
			result.manaScrew = true
		}

		// Mana flood: 7+ lands in play
		if *snapshot.PlayerLandsInPlay >= 7 {
			result.manaFlood = true
		}
	}

	// Check for missed land drops (turn without land drop when not at 7+ lands)
	for turn := 1; turn <= result.maxTurn; turn++ {
		turnPlays := playsByTurn[turn]
		hadLandDrop := false
		for _, play := range turnPlays {
			if play.ActionType == "land_drop" && play.PlayerType == "player" {
				hadLandDrop = true
				break
			}
		}

		// If no land drop and we're below 7 lands, might be missed
		if !hadLandDrop && turn <= 6 && result.landDrops < turn {
			result.missedLandDrop = true
		}

		// Track empty turns (no player actions)
		playerPlays := 0
		for _, play := range turnPlays {
			if play.PlayerType == "player" {
				playerPlays++
			}
		}
		if playerPlays == 0 {
			result.emptyTurns++
		}
	}

	return result
}

// generateCurveSuggestions generates curve improvement suggestions based on analysis.
func (a *PlayAnalyzer) generateCurveSuggestions(result *AnalysisResult) []string {
	var suggestions []string

	// High curve suggestion
	if result.CurveAnalysis.AvgFirstPlay > 2.5 {
		suggestions = append(suggestions, "Consider adding more 1-2 mana spells to improve early game presence")
	}

	// Check CMC distribution
	lowCMCPlays := result.CurveAnalysis.CMCDistribution[1] + result.CurveAnalysis.CMCDistribution[2]
	highCMCPlays := result.CurveAnalysis.CMCDistribution[5] + result.CurveAnalysis.CMCDistribution[6]

	if lowCMCPlays < highCMCPlays && result.TotalGames >= 5 {
		suggestions = append(suggestions, "Deck curve appears top-heavy; early game presence is limited")
	}

	return suggestions
}

// AnalyzeMatch provides detailed analysis for a single match.
func (a *PlayAnalyzer) AnalyzeMatch(ctx context.Context, matchID string) (*MatchAnalysis, error) {
	plays, err := a.playRepo.GetPlaysByMatch(ctx, matchID)
	if err != nil {
		return nil, err
	}

	snapshots, err := a.playRepo.GetSnapshotsByMatch(ctx, matchID)
	if err != nil {
		return nil, err
	}

	summary, err := a.playRepo.GetPlaySummary(ctx, matchID)
	if err != nil {
		return nil, err
	}

	result := &MatchAnalysis{
		MatchID:     matchID,
		TotalTurns:  0,
		TotalPlays:  len(plays),
		PlayerPlays: 0,
		LandDrops:   0,
		Attacks:     0,
		Blocks:      0,
	}

	if summary != nil {
		result.TotalTurns = summary.TotalTurns
		result.PlayerPlays = summary.PlayerPlays
		result.LandDrops = summary.LandDrops
		result.Attacks = summary.Attacks
		result.Blocks = summary.Blocks
	}

	// Analyze plays for sequencing issues
	result.SequencingIssues = a.detectSequencingIssues(plays, snapshots)

	return result, nil
}

// MatchAnalysis contains detailed analysis for a single match.
type MatchAnalysis struct {
	MatchID          string
	TotalTurns       int
	TotalPlays       int
	PlayerPlays      int
	LandDrops        int
	Attacks          int
	Blocks           int
	SequencingIssues []SequencingIssue
}

// detectSequencingIssues identifies potential play order mistakes.
func (a *PlayAnalyzer) detectSequencingIssues(plays []*models.GamePlay, snapshots []*models.GameStateSnapshot) []SequencingIssue {
	var issues []SequencingIssue

	// Build snapshot lookup
	snapshotByTurn := make(map[int]*models.GameStateSnapshot)
	for _, s := range snapshots {
		snapshotByTurn[s.TurnNumber] = s
	}

	// Group plays by turn
	playsByTurn := make(map[int][]*models.GamePlay)
	for _, play := range plays {
		playsByTurn[play.TurnNumber] = append(playsByTurn[play.TurnNumber], play)
	}

	// Look for common sequencing issues
	for turn, turnPlays := range playsByTurn {
		// Check for playing spells before land drop
		landDropIdx, firstSpellIdx := -1, -1
		for i, play := range turnPlays {
			if play.PlayerType != "player" {
				continue
			}
			if play.ActionType == "land_drop" && landDropIdx == -1 {
				landDropIdx = i
			}
			if play.ActionType == "play_card" && firstSpellIdx == -1 {
				firstSpellIdx = i
			}
		}

		// If spell played before land drop
		if firstSpellIdx != -1 && landDropIdx != -1 && firstSpellIdx < landDropIdx {
			issues = append(issues, SequencingIssue{
				Turn:        turn,
				Description: "Spell played before land drop - consider playing land first to have mana available for responses",
			})
		}
	}

	return issues
}
