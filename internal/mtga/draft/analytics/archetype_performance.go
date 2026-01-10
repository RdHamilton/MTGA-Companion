package analytics

import (
	"context"
	"strconv"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// ArchetypePerformanceAnalyzer analyzes win rates and performance by archetype/color combination.
type ArchetypePerformanceAnalyzer struct {
	draftRepo     repository.DraftRepository
	analyticsRepo repository.DraftAnalyticsRepository
	matchRepo     repository.MatchRepository
	cardStore     CardStore
}

// NewArchetypePerformanceAnalyzer creates a new archetype performance analyzer.
func NewArchetypePerformanceAnalyzer(
	draftRepo repository.DraftRepository,
	analyticsRepo repository.DraftAnalyticsRepository,
	matchRepo repository.MatchRepository,
	cardStore CardStore,
) *ArchetypePerformanceAnalyzer {
	return &ArchetypePerformanceAnalyzer{
		draftRepo:     draftRepo,
		analyticsRepo: analyticsRepo,
		matchRepo:     matchRepo,
		cardStore:     cardStore,
	}
}

// AnalyzeArchetypePerformance calculates and stores performance stats for all archetypes.
func (a *ArchetypePerformanceAnalyzer) AnalyzeArchetypePerformance(ctx context.Context, setCode *string) ([]*models.DraftArchetypeStats, error) {
	// Get completed draft sessions
	sessions, err := a.draftRepo.GetCompletedSessions(ctx, 1000)
	if err != nil {
		return nil, err
	}

	// Filter by set code if provided
	var filteredSessions []*models.DraftSession
	for _, s := range sessions {
		if setCode == nil || s.SetCode == *setCode {
			filteredSessions = append(filteredSessions, s)
		}
	}

	// Track stats by set and color combination
	statsMap := make(map[string]*models.DraftArchetypeStats)

	for _, session := range filteredSessions {
		// Determine the color combination for this draft
		colorPair, err := a.determineDraftColorPair(ctx, session.ID)
		if err != nil || colorPair == "" {
			continue
		}

		key := session.SetCode + "_" + colorPair
		if _, ok := statsMap[key]; !ok {
			statsMap[key] = &models.DraftArchetypeStats{
				SetCode:          session.SetCode,
				ColorCombination: colorPair,
				ArchetypeName:    getArchetypeName(colorPair),
				UpdatedAt:        time.Now(),
			}
		}

		stats := statsMap[key]
		stats.DraftsCount++

		// Get match results for this draft session
		results, err := a.analyticsRepo.GetDraftMatchResults(ctx, session.ID)
		if err == nil && len(results) > 0 {
			for _, r := range results {
				stats.MatchesPlayed++
				if r.Result == "win" {
					stats.MatchesWon++
				}
			}
			// Update last played
			if stats.LastPlayedAt == nil || results[len(results)-1].MatchTimestamp.After(*stats.LastPlayedAt) {
				lastPlayed := results[len(results)-1].MatchTimestamp
				stats.LastPlayedAt = &lastPlayed
			}
		}

		// Track draft grade if available
		if session.OverallScore != nil {
			if stats.AvgDraftGrade == nil {
				score := float64(*session.OverallScore)
				stats.AvgDraftGrade = &score
			} else {
				// Calculate running average
				currentAvg := *stats.AvgDraftGrade
				newAvg := (currentAvg*float64(stats.DraftsCount-1) + float64(*session.OverallScore)) / float64(stats.DraftsCount)
				stats.AvgDraftGrade = &newAvg
			}
		}
	}

	// Save all stats to database
	var allStats []*models.DraftArchetypeStats
	for _, stats := range statsMap {
		if err := a.analyticsRepo.UpsertArchetypeStats(ctx, stats); err != nil {
			continue
		}
		allStats = append(allStats, stats)
	}

	return allStats, nil
}

// GetArchetypeStats retrieves cached archetype stats for a set.
func (a *ArchetypePerformanceAnalyzer) GetArchetypeStats(ctx context.Context, setCode string) ([]*models.DraftArchetypeStats, error) {
	return a.analyticsRepo.GetArchetypeStats(ctx, setCode)
}

// GetAllArchetypeStats retrieves all cached archetype stats.
func (a *ArchetypePerformanceAnalyzer) GetAllArchetypeStats(ctx context.Context) ([]*models.DraftArchetypeStats, error) {
	return a.analyticsRepo.GetAllArchetypeStats(ctx)
}

// GetBestArchetypes retrieves the best performing archetypes.
func (a *ArchetypePerformanceAnalyzer) GetBestArchetypes(ctx context.Context, minMatches, limit int) ([]*models.DraftArchetypeStats, error) {
	return a.analyticsRepo.GetBestArchetypes(ctx, minMatches, limit)
}

// GetWorstArchetypes retrieves the worst performing archetypes.
func (a *ArchetypePerformanceAnalyzer) GetWorstArchetypes(ctx context.Context, minMatches, limit int) ([]*models.DraftArchetypeStats, error) {
	return a.analyticsRepo.GetWorstArchetypes(ctx, minMatches, limit)
}

// LinkDraftToMatches links a draft session to its corresponding matches.
func (a *ArchetypePerformanceAnalyzer) LinkDraftToMatches(ctx context.Context, sessionID string, session *models.DraftSession) error {
	if session == nil {
		var err error
		session, err = a.draftRepo.GetSession(ctx, sessionID)
		if err != nil || session == nil {
			return err
		}
	}

	// Find matches that occurred during or after the draft
	startTime := session.StartTime
	if session.EndTime != nil {
		startTime = *session.EndTime
	}

	filter := models.StatsFilter{
		StartDate: &startTime,
		Formats:   []string{"Ladder", "Play"},
	}

	matches, err := a.matchRepo.GetMatches(ctx, filter)
	if err != nil {
		return err
	}

	// Filter to matches that appear to be from this draft
	// (within reasonable time frame and matching event)
	for _, match := range matches {
		// Check if this match is likely from our draft
		if !a.isMatchFromDraft(match, session) {
			continue
		}

		result := &models.DraftMatchResult{
			SessionID:      sessionID,
			MatchID:        match.ID,
			Result:         match.Result,
			GameWins:       match.PlayerWins,
			GameLosses:     match.OpponentWins,
			MatchTimestamp: match.Timestamp,
		}

		// Try to determine opponent colors from match data
		if match.OpponentName != nil {
			result.OpponentColors = "" // Would need opponent analysis data
		}

		if err := a.analyticsRepo.SaveDraftMatchResult(ctx, result); err != nil {
			// Continue even if one save fails
			continue
		}
	}

	return nil
}

func (a *ArchetypePerformanceAnalyzer) isMatchFromDraft(match *models.Match, session *models.DraftSession) bool {
	// Check if match event name contains draft-related terms
	eventName := match.EventName
	isDraftEvent := false
	draftTerms := []string{"Draft", "Sealed", "Limited", "QuickDraft", "PremierDraft"}
	for _, term := range draftTerms {
		if containsIgnoreCase(eventName, term) {
			isDraftEvent = true
			break
		}
	}

	if !isDraftEvent {
		return false
	}

	// Check if match is within reasonable time frame of draft
	// Drafts typically result in matches within 24 hours
	maxTimeDiff := 24 * time.Hour
	if session.EndTime != nil {
		timeDiff := match.Timestamp.Sub(*session.EndTime)
		if timeDiff < 0 || timeDiff > maxTimeDiff {
			return false
		}
	} else {
		timeDiff := match.Timestamp.Sub(session.StartTime)
		if timeDiff < 0 || timeDiff > maxTimeDiff {
			return false
		}
	}

	// Check if set code matches
	if containsIgnoreCase(eventName, session.SetCode) {
		return true
	}

	return true // Default to including if other checks pass
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(substr) == 0 ||
		(len(s) > 0 && containsLower(toLower(s), toLower(substr))))
}

func toLower(s string) string {
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		result[i] = c
	}
	return string(result)
}

func containsLower(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (a *ArchetypePerformanceAnalyzer) determineDraftColorPair(ctx context.Context, sessionID string) (string, error) {
	picks, err := a.draftRepo.GetPicksBySession(ctx, sessionID)
	if err != nil || len(picks) == 0 {
		return "", err
	}

	colorCounts := make(map[string]int)
	for _, pick := range picks {
		cardID, err := strconv.Atoi(pick.CardID)
		if err != nil {
			continue
		}
		card, err := a.cardStore.GetCard(cardID)
		if err != nil || card == nil {
			continue
		}
		for _, color := range card.Colors {
			colorCounts[color]++
		}
	}

	return determinePrimaryColorPair(colorCounts), nil
}

// ArchetypePerformanceResponse is the API response for archetype performance.
type ArchetypePerformanceResponse struct {
	SetCode    *string                      `json:"setCode,omitempty"`
	Archetypes []*ArchetypePerformanceEntry `json:"archetypes"`
	Best       []*ArchetypePerformanceEntry `json:"best,omitempty"`
	Worst      []*ArchetypePerformanceEntry `json:"worst,omitempty"`
}

// ArchetypePerformanceEntry represents a single archetype's performance.
type ArchetypePerformanceEntry struct {
	ColorCombination string   `json:"colorCombination"`
	ArchetypeName    string   `json:"archetypeName"`
	SetCode          string   `json:"setCode"`
	MatchesPlayed    int      `json:"matchesPlayed"`
	MatchesWon       int      `json:"matchesWon"`
	WinRate          float64  `json:"winRate"`
	DraftsCount      int      `json:"draftsCount"`
	AvgDraftGrade    *float64 `json:"avgDraftGrade,omitempty"`
	LastPlayedAt     *string  `json:"lastPlayedAt,omitempty"`
}

// ToArchetypePerformanceEntry converts a DraftArchetypeStats to an API entry.
func ToArchetypePerformanceEntry(stats *models.DraftArchetypeStats) *ArchetypePerformanceEntry {
	if stats == nil {
		return nil
	}

	entry := &ArchetypePerformanceEntry{
		ColorCombination: stats.ColorCombination,
		ArchetypeName:    stats.ArchetypeName,
		SetCode:          stats.SetCode,
		MatchesPlayed:    stats.MatchesPlayed,
		MatchesWon:       stats.MatchesWon,
		WinRate:          stats.WinRate(),
		DraftsCount:      stats.DraftsCount,
		AvgDraftGrade:    stats.AvgDraftGrade,
	}

	if stats.LastPlayedAt != nil {
		formatted := stats.LastPlayedAt.Format(time.RFC3339)
		entry.LastPlayedAt = &formatted
	}

	return entry
}

// BuildArchetypePerformanceResponse builds the full API response.
func BuildArchetypePerformanceResponse(
	setCode *string,
	allStats []*models.DraftArchetypeStats,
	best []*models.DraftArchetypeStats,
	worst []*models.DraftArchetypeStats,
) *ArchetypePerformanceResponse {
	response := &ArchetypePerformanceResponse{
		SetCode:    setCode,
		Archetypes: make([]*ArchetypePerformanceEntry, 0),
		Best:       make([]*ArchetypePerformanceEntry, 0),
		Worst:      make([]*ArchetypePerformanceEntry, 0),
	}

	for _, s := range allStats {
		response.Archetypes = append(response.Archetypes, ToArchetypePerformanceEntry(s))
	}

	for _, s := range best {
		response.Best = append(response.Best, ToArchetypePerformanceEntry(s))
	}

	for _, s := range worst {
		response.Worst = append(response.Worst, ToArchetypePerformanceEntry(s))
	}

	return response
}

// CardWithPerformance pairs a card with performance data.
type CardWithPerformance struct {
	Card    *cards.Card `json:"card"`
	WinRate float64     `json:"winRate"`
	Picks   int         `json:"picks"`
}
