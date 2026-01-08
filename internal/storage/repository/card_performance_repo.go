package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// CardPerformanceRepository handles card performance analysis queries.
type CardPerformanceRepository interface {
	// GetCardPerformance calculates performance metrics for all cards in a deck.
	GetCardPerformance(ctx context.Context, filter models.CardPerformanceFilter) ([]*models.CardPerformance, error)

	// GetDeckPerformanceAnalysis returns a complete analysis for a deck.
	GetDeckPerformanceAnalysis(ctx context.Context, deckID string) (*models.DeckPerformanceAnalysis, error)

	// GetCardPlayEvents retrieves all play events for a specific card in a deck.
	GetCardPlayEvents(ctx context.Context, deckID string, cardID int) ([]*models.CardPlayEvent, error)

	// GetUnderperformingCards identifies cards that hurt deck performance.
	GetUnderperformingCards(ctx context.Context, deckID string, threshold float64) ([]*models.CardPerformance, error)

	// GetOverperformingCards identifies cards with high win impact.
	GetOverperformingCards(ctx context.Context, deckID string, threshold float64) ([]*models.CardPerformance, error)
}

// cardPerformanceRepository is the concrete implementation.
type cardPerformanceRepository struct {
	db *sql.DB
}

// NewCardPerformanceRepository creates a new card performance repository.
func NewCardPerformanceRepository(db *sql.DB) CardPerformanceRepository {
	return &cardPerformanceRepository{db: db}
}

// GetCardPerformance calculates performance metrics for all cards in a deck.
func (r *cardPerformanceRepository) GetCardPerformance(ctx context.Context, filter models.CardPerformanceFilter) ([]*models.CardPerformance, error) {
	if filter.DeckID == "" {
		return nil, fmt.Errorf("deck_id is required")
	}

	// Get deck win rate for comparison
	deckWinRate, totalGames, err := r.getDeckWinRate(ctx, filter.DeckID)
	if err != nil {
		return nil, fmt.Errorf("failed to get deck win rate: %w", err)
	}

	if totalGames < filter.MinGames {
		return nil, nil // Not enough data
	}

	// Query for card play statistics
	// This joins game_plays with matches to correlate card plays with match results
	query := `
		WITH card_plays AS (
			SELECT
				gp.card_id,
				gp.card_name,
				gp.match_id,
				gp.game_id,
				gp.turn_number,
				m.result as match_result
			FROM game_plays gp
			INNER JOIN matches m ON gp.match_id = m.id
			WHERE m.deck_id = ?
			AND gp.player_type = 'player'
			AND gp.action_type = 'play_card'
			AND gp.card_id IS NOT NULL
		),
		card_draws AS (
			SELECT
				gp.card_id,
				gp.card_name,
				gp.match_id,
				gp.game_id,
				gp.turn_number,
				m.result as match_result
			FROM game_plays gp
			INNER JOIN matches m ON gp.match_id = m.id
			WHERE m.deck_id = ?
			AND gp.player_type = 'player'
			AND gp.zone_to = 'hand'
			AND gp.card_id IS NOT NULL
		),
		card_mulligans AS (
			SELECT
				gp.card_id,
				gp.match_id,
				gp.game_id
			FROM game_plays gp
			INNER JOIN matches m ON gp.match_id = m.id
			WHERE m.deck_id = ?
			AND gp.player_type = 'player'
			AND gp.action_type = 'mulligan'
			AND gp.card_id IS NOT NULL
		)
		SELECT
			COALESCE(cp.card_id, cd.card_id) as card_id,
			COALESCE(cp.card_name, cd.card_name) as card_name,
			COUNT(DISTINCT cd.game_id) as games_drawn,
			COUNT(DISTINCT cp.game_id) as games_played,
			SUM(CASE WHEN cd.match_result = 'win' THEN 1 ELSE 0 END) as wins_when_drawn,
			SUM(CASE WHEN cp.match_result = 'win' THEN 1 ELSE 0 END) as wins_when_played,
			AVG(CASE WHEN cp.turn_number IS NOT NULL THEN cp.turn_number END) as avg_turn_played,
			COUNT(DISTINCT cm.game_id) as mulliganed_games
		FROM card_draws cd
		LEFT JOIN card_plays cp ON cd.card_id = cp.card_id AND cd.game_id = cp.game_id
		LEFT JOIN card_mulligans cm ON cd.card_id = cm.card_id AND cd.game_id = cm.game_id
		GROUP BY COALESCE(cp.card_id, cd.card_id), COALESCE(cp.card_name, cd.card_name)
		HAVING games_drawn >= ?
		ORDER BY games_drawn DESC
	`

	minGames := filter.MinGames
	if minGames < models.MinGamesForAnalysis {
		minGames = models.MinGamesForAnalysis
	}

	rows, err := r.db.QueryContext(ctx, query, filter.DeckID, filter.DeckID, filter.DeckID, minGames)
	if err != nil {
		return nil, fmt.Errorf("failed to query card performance: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var performances []*models.CardPerformance
	for rows.Next() {
		var perf models.CardPerformance
		var avgTurnPlayed sql.NullFloat64
		var winsWhenDrawn, winsWhenPlayed int

		err := rows.Scan(
			&perf.CardID,
			&perf.CardName,
			&perf.GamesDrawn,
			&perf.GamesPlayed,
			&winsWhenDrawn,
			&winsWhenPlayed,
			&avgTurnPlayed,
			&perf.MulliganedAway,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card performance: %w", err)
		}

		// Skip basic lands unless requested
		if !filter.IncludeLands && isBasicLand(perf.CardName) {
			continue
		}

		// Calculate metrics
		perf.GamesWithCard = totalGames
		perf.DeckWinRate = deckWinRate

		if perf.GamesDrawn > 0 {
			perf.WinRateWhenDrawn = float64(winsWhenDrawn) / float64(perf.GamesDrawn)
			perf.MulliganRate = float64(perf.MulliganedAway) / float64(perf.GamesDrawn)
		}

		if perf.GamesPlayed > 0 {
			perf.WinRateWhenPlayed = float64(winsWhenPlayed) / float64(perf.GamesPlayed)
		}

		if perf.GamesDrawn > 0 {
			perf.PlayRate = float64(perf.GamesPlayed) / float64(perf.GamesDrawn)
		}

		if avgTurnPlayed.Valid {
			perf.AvgTurnPlayed = avgTurnPlayed.Float64
		}

		// Calculate win contribution (how much better/worse than deck average)
		perf.WinContribution = perf.WinRateWhenDrawn - deckWinRate

		// Calculate impact score (-1 to +1 scale)
		perf.ImpactScore = calculateImpactScore(perf.WinContribution, perf.GamesDrawn)

		// Assign confidence level
		perf.ConfidenceLevel = getConfidenceLevel(perf.GamesDrawn)
		perf.SampleSize = perf.GamesDrawn

		// Assign performance grade
		perf.PerformanceGrade = getPerformanceGrade(perf.WinContribution)

		performances = append(performances, &perf)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating card performance: %w", err)
	}

	// Sort by impact score (best performers first)
	sort.Slice(performances, func(i, j int) bool {
		return performances[i].ImpactScore > performances[j].ImpactScore
	})

	return performances, nil
}

// GetDeckPerformanceAnalysis returns a complete analysis for a deck.
func (r *cardPerformanceRepository) GetDeckPerformanceAnalysis(ctx context.Context, deckID string) (*models.DeckPerformanceAnalysis, error) {
	// Get deck info
	var deckName string
	err := r.db.QueryRowContext(ctx, `SELECT name FROM decks WHERE id = ?`, deckID).Scan(&deckName)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("deck not found: %s", deckID)
		}
		return nil, fmt.Errorf("failed to get deck: %w", err)
	}

	// Get deck statistics
	var totalMatches, matchesWon int
	err = r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN result = 'win' THEN 1 ELSE 0 END) as wins
		FROM matches
		WHERE deck_id = ?
	`, deckID).Scan(&totalMatches, &matchesWon)
	if err != nil {
		return nil, fmt.Errorf("failed to get match stats: %w", err)
	}

	// Get total games
	var totalGames int
	err = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM games g
		INNER JOIN matches m ON g.match_id = m.id
		WHERE m.deck_id = ?
	`, deckID).Scan(&totalGames)
	if err != nil {
		return nil, fmt.Errorf("failed to get game count: %w", err)
	}

	// Get card performance
	filter := models.CardPerformanceFilter{
		DeckID:       deckID,
		MinGames:     models.MinGamesForAnalysis,
		IncludeLands: false,
	}
	cardPerf, err := r.GetCardPerformance(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Identify best and worst performers
	var bestPerformers, worstPerformers []string
	for _, perf := range cardPerf {
		switch perf.PerformanceGrade {
		case models.PerformanceGradeExcellent, models.PerformanceGradeGood:
			bestPerformers = append(bestPerformers, perf.CardName)
		case models.PerformanceGradePoor, models.PerformanceGradeBad:
			worstPerformers = append(worstPerformers, perf.CardName)
		}
	}

	// Limit to top 5 each
	if len(bestPerformers) > 5 {
		bestPerformers = bestPerformers[:5]
	}
	if len(worstPerformers) > 5 {
		worstPerformers = worstPerformers[:5]
	}

	analysis := &models.DeckPerformanceAnalysis{
		DeckID:          deckID,
		DeckName:        deckName,
		TotalMatches:    totalMatches,
		TotalGames:      totalGames,
		OverallWinRate:  0,
		CardPerformance: cardPerf,
		BestPerformers:  bestPerformers,
		WorstPerformers: worstPerformers,
	}

	if totalMatches > 0 {
		analysis.OverallWinRate = float64(matchesWon) / float64(totalMatches)
	}

	return analysis, nil
}

// GetCardPlayEvents retrieves all play events for a specific card in a deck.
func (r *cardPerformanceRepository) GetCardPlayEvents(ctx context.Context, deckID string, cardID int) ([]*models.CardPlayEvent, error) {
	query := `
		SELECT
			gp.match_id,
			gp.game_id,
			gp.card_id,
			gp.card_name,
			gp.turn_number,
			gp.phase,
			m.result as match_result
		FROM game_plays gp
		INNER JOIN matches m ON gp.match_id = m.id
		WHERE m.deck_id = ?
		AND gp.card_id = ?
		AND gp.player_type = 'player'
		AND gp.action_type = 'play_card'
		ORDER BY m.timestamp DESC, gp.sequence_number ASC
	`

	rows, err := r.db.QueryContext(ctx, query, deckID, cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to query card play events: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var events []*models.CardPlayEvent
	for rows.Next() {
		event := &models.CardPlayEvent{}
		var cardName sql.NullString
		err := rows.Scan(
			&event.MatchID,
			&event.GameID,
			&event.CardID,
			&cardName,
			&event.TurnNumber,
			&event.Phase,
			&event.MatchResult,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan card play event: %w", err)
		}
		if cardName.Valid {
			event.CardName = cardName.String
		}
		events = append(events, event)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating card play events: %w", err)
	}

	return events, nil
}

// GetUnderperformingCards identifies cards that hurt deck performance.
func (r *cardPerformanceRepository) GetUnderperformingCards(ctx context.Context, deckID string, threshold float64) ([]*models.CardPerformance, error) {
	filter := models.CardPerformanceFilter{
		DeckID:       deckID,
		MinGames:     models.MinGamesForMediumConfidence,
		IncludeLands: false,
	}

	allCards, err := r.GetCardPerformance(ctx, filter)
	if err != nil {
		return nil, err
	}

	var underperformers []*models.CardPerformance
	for _, card := range allCards {
		if card.WinContribution < -threshold {
			underperformers = append(underperformers, card)
		}
	}

	// Sort by worst performers first
	sort.Slice(underperformers, func(i, j int) bool {
		return underperformers[i].WinContribution < underperformers[j].WinContribution
	})

	return underperformers, nil
}

// GetOverperformingCards identifies cards with high win impact.
func (r *cardPerformanceRepository) GetOverperformingCards(ctx context.Context, deckID string, threshold float64) ([]*models.CardPerformance, error) {
	filter := models.CardPerformanceFilter{
		DeckID:       deckID,
		MinGames:     models.MinGamesForMediumConfidence,
		IncludeLands: false,
	}

	allCards, err := r.GetCardPerformance(ctx, filter)
	if err != nil {
		return nil, err
	}

	var overperformers []*models.CardPerformance
	for _, card := range allCards {
		if card.WinContribution > threshold {
			overperformers = append(overperformers, card)
		}
	}

	// Sort by best performers first
	sort.Slice(overperformers, func(i, j int) bool {
		return overperformers[i].WinContribution > overperformers[j].WinContribution
	})

	return overperformers, nil
}

// getDeckWinRate calculates the overall win rate for a deck.
func (r *cardPerformanceRepository) getDeckWinRate(ctx context.Context, deckID string) (float64, int, error) {
	var total, wins int
	err := r.db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) as total,
			SUM(CASE WHEN result = 'win' THEN 1 ELSE 0 END) as wins
		FROM matches
		WHERE deck_id = ?
	`, deckID).Scan(&total, &wins)
	if err != nil {
		return 0, 0, err
	}

	if total == 0 {
		return 0, 0, nil
	}

	return float64(wins) / float64(total), total, nil
}

// isBasicLand checks if a card name is a basic land.
func isBasicLand(cardName string) bool {
	basicLands := []string{"Plains", "Island", "Swamp", "Mountain", "Forest"}
	for _, land := range basicLands {
		if cardName == land {
			return true
		}
	}
	return false
}

// calculateImpactScore converts win contribution to a -1 to +1 scale.
func calculateImpactScore(winContribution float64, sampleSize int) float64 {
	// Base score is the win contribution (already in decimal form)
	// We normalize it to approximately -1 to +1 range
	// A 20% swing is considered maximum impact
	score := winContribution / 0.20

	// Dampen score for low sample sizes
	if sampleSize < models.MinGamesForHighConfidence {
		dampening := float64(sampleSize) / float64(models.MinGamesForHighConfidence)
		score *= dampening
	}

	// Clamp to -1 to +1
	if score > 1.0 {
		score = 1.0
	}
	if score < -1.0 {
		score = -1.0
	}

	return score
}

// getConfidenceLevel returns the confidence level based on sample size.
func getConfidenceLevel(gamesDrawn int) string {
	if gamesDrawn >= models.MinGamesForHighConfidence {
		return models.ConfidenceHigh
	}
	if gamesDrawn >= models.MinGamesForMediumConfidence {
		return models.ConfidenceMedium
	}
	return models.ConfidenceLow
}

// getPerformanceGrade returns the performance grade based on win contribution.
func getPerformanceGrade(winContribution float64) string {
	switch {
	case winContribution > 0.10:
		return models.PerformanceGradeExcellent
	case winContribution > 0.05:
		return models.PerformanceGradeGood
	case winContribution > -0.05:
		return models.PerformanceGradeAverage
	case winContribution > -0.10:
		return models.PerformanceGradePoor
	default:
		return models.PerformanceGradeBad
	}
}

// GetTurnPlayedDistribution returns the distribution of turns a card was played.
func (r *cardPerformanceRepository) GetTurnPlayedDistribution(ctx context.Context, deckID string, cardID int) (map[int]int, error) {
	query := `
		SELECT
			gp.turn_number,
			COUNT(*) as count
		FROM game_plays gp
		INNER JOIN matches m ON gp.match_id = m.id
		WHERE m.deck_id = ?
		AND gp.card_id = ?
		AND gp.player_type = 'player'
		AND gp.action_type = 'play_card'
		GROUP BY gp.turn_number
		ORDER BY gp.turn_number ASC
	`

	rows, err := r.db.QueryContext(ctx, query, deckID, cardID)
	if err != nil {
		return nil, fmt.Errorf("failed to query turn distribution: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	distribution := make(map[int]int)
	for rows.Next() {
		var turn, count int
		if err := rows.Scan(&turn, &count); err != nil {
			return nil, fmt.Errorf("failed to scan turn distribution: %w", err)
		}
		distribution[turn] = count
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating turn distribution: %w", err)
	}

	return distribution, nil
}

// GetCardRecommendations generates add/remove recommendations for a deck.
func (r *cardPerformanceRepository) GetCardRecommendations(ctx context.Context, req models.RecommendationsRequest) (*models.RecommendationsResponse, error) {
	// Get deck info and current performance
	analysis, err := r.GetDeckPerformanceAnalysis(ctx, req.DeckID)
	if err != nil {
		return nil, err
	}

	response := &models.RecommendationsResponse{
		DeckID:         req.DeckID,
		DeckName:       analysis.DeckName,
		CurrentWinRate: analysis.OverallWinRate,
	}

	// Generate remove recommendations from underperforming cards
	underperformers, err := r.GetUnderperformingCards(ctx, req.DeckID, 0.05)
	if err != nil {
		return nil, err
	}

	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = 5
	}

	for i, card := range underperformers {
		if i >= maxResults {
			break
		}

		rec := &models.CardRecommendation{
			Type:           "remove",
			CardID:         card.CardID,
			CardName:       card.CardName,
			Reason:         generateRemoveReason(card),
			ImpactEstimate: -card.WinContribution, // Removing bad card improves win rate
			Confidence:     card.ConfidenceLevel,
			Priority:       i + 1,
			BasedOnGames:   card.SampleSize,
		}
		response.RemoveRecommendations = append(response.RemoveRecommendations, rec)
	}

	// Generate add recommendations from similar successful decks
	addRecs, err := r.getAddRecommendations(ctx, req)
	if err != nil {
		// Log but don't fail - add recommendations are supplementary
		addRecs = nil
	}
	response.AddRecommendations = addRecs

	// Generate swap recommendations if requested
	if req.IncludeSwaps && len(underperformers) > 0 && len(addRecs) > 0 {
		response.SwapRecommendations = generateSwapRecommendations(underperformers, addRecs, maxResults)
	}

	// Calculate projected win rate
	response.ProjectedWinRate = calculateProjectedWinRate(analysis.OverallWinRate, response)

	return response, nil
}

// getAddRecommendations finds cards from similar successful decks that could improve this deck.
func (r *cardPerformanceRepository) getAddRecommendations(ctx context.Context, req models.RecommendationsRequest) ([]*models.CardRecommendation, error) {
	// Get the deck's format and color identity
	var format, colorIdentity string
	err := r.db.QueryRowContext(ctx, `
		SELECT format, color_identity
		FROM decks
		WHERE id = ?
	`, req.DeckID).Scan(&format, &colorIdentity)
	if err != nil {
		return nil, fmt.Errorf("failed to get deck info: %w", err)
	}

	if req.Format != "" {
		format = req.Format
	}

	// Find cards that appear in high-win-rate decks of similar colors but not in this deck
	query := `
		WITH current_deck_cards AS (
			SELECT DISTINCT card_id
			FROM game_plays gp
			INNER JOIN matches m ON gp.match_id = m.id
			WHERE m.deck_id = ?
			AND gp.player_type = 'player'
			AND gp.card_id IS NOT NULL
		),
		similar_decks AS (
			SELECT d.id, d.name
			FROM decks d
			INNER JOIN matches m ON m.deck_id = d.id
			WHERE d.format = ?
			AND d.id != ?
			GROUP BY d.id
			HAVING AVG(CASE WHEN m.result = 'win' THEN 1.0 ELSE 0.0 END) > 0.55
			AND COUNT(*) >= 10
		),
		candidate_cards AS (
			SELECT
				gp.card_id,
				gp.card_name,
				COUNT(DISTINCT m.id) as games_played,
				AVG(CASE WHEN m.result = 'win' THEN 1.0 ELSE 0.0 END) as win_rate
			FROM game_plays gp
			INNER JOIN matches m ON gp.match_id = m.id
			INNER JOIN similar_decks sd ON m.deck_id = sd.id
			WHERE gp.player_type = 'player'
			AND gp.action_type = 'play_card'
			AND gp.card_id IS NOT NULL
			AND gp.card_id NOT IN (SELECT card_id FROM current_deck_cards)
			GROUP BY gp.card_id, gp.card_name
			HAVING games_played >= 10
		)
		SELECT card_id, card_name, games_played, win_rate
		FROM candidate_cards
		ORDER BY win_rate DESC, games_played DESC
		LIMIT ?
	`

	maxResults := req.MaxResults
	if maxResults <= 0 {
		maxResults = 5
	}

	rows, err := r.db.QueryContext(ctx, query, req.DeckID, format, req.DeckID, maxResults)
	if err != nil {
		return nil, fmt.Errorf("failed to query add recommendations: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var recommendations []*models.CardRecommendation
	priority := 1
	for rows.Next() {
		var cardID int
		var cardName sql.NullString
		var gamesPlayed int
		var winRate float64

		if err := rows.Scan(&cardID, &cardName, &gamesPlayed, &winRate); err != nil {
			return nil, fmt.Errorf("failed to scan add recommendation: %w", err)
		}

		name := ""
		if cardName.Valid {
			name = cardName.String
		}

		// Skip basic lands
		if isBasicLand(name) {
			continue
		}

		rec := &models.CardRecommendation{
			Type:           "add",
			CardID:         cardID,
			CardName:       name,
			Reason:         fmt.Sprintf("High win rate (%.1f%%) in similar decks", winRate*100),
			ImpactEstimate: winRate - 0.50, // Estimate impact vs baseline
			Confidence:     getConfidenceLevel(gamesPlayed),
			Priority:       priority,
			BasedOnGames:   gamesPlayed,
		}
		recommendations = append(recommendations, rec)
		priority++
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating add recommendations: %w", err)
	}

	return recommendations, nil
}

// generateRemoveReason creates a human-readable reason for removing a card.
func generateRemoveReason(card *models.CardPerformance) string {
	reasons := []string{}

	if card.WinContribution < -0.10 {
		reasons = append(reasons, fmt.Sprintf("significantly lowers win rate (%.1f%% below deck average)", card.WinContribution*100))
	} else if card.WinContribution < -0.05 {
		reasons = append(reasons, fmt.Sprintf("lowers win rate (%.1f%% below deck average)", card.WinContribution*100))
	}

	if card.PlayRate < 0.50 && card.GamesDrawn >= 10 {
		reasons = append(reasons, fmt.Sprintf("often not played when drawn (%.0f%% play rate)", card.PlayRate*100))
	}

	if card.MulliganRate > 0.30 && card.GamesDrawn >= 10 {
		reasons = append(reasons, fmt.Sprintf("frequently mulliganed (%.0f%% of games)", card.MulliganRate*100))
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "underperforming compared to deck average")
	}

	return strings.Join(reasons, "; ")
}

// generateSwapRecommendations creates swap recommendations from remove and add lists.
func generateSwapRecommendations(removes []*models.CardPerformance, adds []*models.CardRecommendation, maxResults int) []*models.CardRecommendation {
	var swaps []*models.CardRecommendation

	for i := 0; i < len(removes) && i < len(adds) && i < maxResults; i++ {
		remove := removes[i]
		add := adds[i]

		swap := &models.CardRecommendation{
			Type:            "swap",
			CardID:          remove.CardID,
			CardName:        remove.CardName,
			SwapForCardID:   &add.CardID,
			SwapForCardName: &add.CardName,
			Reason:          fmt.Sprintf("Replace underperformer with %s", add.CardName),
			ImpactEstimate:  add.ImpactEstimate - remove.WinContribution,
			Confidence:      add.Confidence,
			Priority:        i + 1,
			BasedOnGames:    add.BasedOnGames,
		}
		swaps = append(swaps, swap)
	}

	return swaps
}

// calculateProjectedWinRate estimates win rate after applying recommendations.
func calculateProjectedWinRate(currentWinRate float64, response *models.RecommendationsResponse) float64 {
	projected := currentWinRate

	// Add estimated impact from top recommendations
	for i, rec := range response.RemoveRecommendations {
		if i >= 3 {
			break // Only consider top 3
		}
		projected += rec.ImpactEstimate * 0.5 // Dampen estimate
	}

	for i, rec := range response.AddRecommendations {
		if i >= 3 {
			break
		}
		projected += rec.ImpactEstimate * 0.3 // More conservative for adds
	}

	// Clamp to valid range
	if projected > 1.0 {
		projected = 1.0
	}
	if projected < 0 {
		projected = 0
	}

	return projected
}
