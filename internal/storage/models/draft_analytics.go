package models

import (
	"encoding/json"
	"time"
)

// DraftMatchResult links a draft session to its match results.
type DraftMatchResult struct {
	ID             int64     `json:"id" db:"id"`
	SessionID      string    `json:"sessionId" db:"session_id"`
	MatchID        string    `json:"matchId" db:"match_id"`
	Result         string    `json:"result" db:"result"` // win, loss
	OpponentColors string    `json:"opponentColors,omitempty" db:"opponent_colors"`
	GameWins       int       `json:"gameWins" db:"game_wins"`
	GameLosses     int       `json:"gameLosses" db:"game_losses"`
	MatchTimestamp time.Time `json:"matchTimestamp" db:"match_timestamp"`
}

// WinRate returns the match win rate (1.0 for win, 0.0 for loss).
func (d *DraftMatchResult) WinRate() float64 {
	if d.Result == "win" {
		return 1.0
	}
	return 0.0
}

// DraftArchetypeStats tracks performance aggregation by archetype/color combo.
type DraftArchetypeStats struct {
	ID               int64      `json:"id" db:"id"`
	SetCode          string     `json:"setCode" db:"set_code"`
	ColorCombination string     `json:"colorCombination" db:"color_combination"`
	ArchetypeName    string     `json:"archetypeName" db:"archetype_name"`
	MatchesPlayed    int        `json:"matchesPlayed" db:"matches_played"`
	MatchesWon       int        `json:"matchesWon" db:"matches_won"`
	DraftsCount      int        `json:"draftsCount" db:"drafts_count"`
	AvgDraftGrade    *float64   `json:"avgDraftGrade,omitempty" db:"avg_draft_grade"`
	LastPlayedAt     *time.Time `json:"lastPlayedAt,omitempty" db:"last_played_at"`
	UpdatedAt        time.Time  `json:"updatedAt" db:"updated_at"`
}

// WinRate returns the match win rate for this archetype.
func (d *DraftArchetypeStats) WinRate() float64 {
	if d.MatchesPlayed == 0 {
		return 0
	}
	return float64(d.MatchesWon) / float64(d.MatchesPlayed)
}

// DraftCommunityComparison compares user performance to 17Lands community averages.
type DraftCommunityComparison struct {
	ID                  int64     `json:"id" db:"id"`
	SetCode             string    `json:"setCode" db:"set_code"`
	DraftFormat         string    `json:"draftFormat" db:"draft_format"`
	UserWinRate         float64   `json:"userWinRate" db:"user_win_rate"`
	CommunityAvgWinRate float64   `json:"communityAvgWinRate" db:"community_avg_win_rate"`
	PercentileRank      *float64  `json:"percentileRank,omitempty" db:"percentile_rank"`
	SampleSize          int       `json:"sampleSize" db:"sample_size"`
	CalculatedAt        time.Time `json:"calculatedAt" db:"calculated_at"`
}

// WinRateDelta returns the difference between user and community win rates.
func (d *DraftCommunityComparison) WinRateDelta() float64 {
	return d.UserWinRate - d.CommunityAvgWinRate
}

// DraftTemporalTrend tracks performance over time periods.
type DraftTemporalTrend struct {
	ID            int64     `json:"id" db:"id"`
	PeriodType    string    `json:"periodType" db:"period_type"` // week, month
	PeriodStart   time.Time `json:"periodStart" db:"period_start"`
	PeriodEnd     time.Time `json:"periodEnd" db:"period_end"`
	SetCode       *string   `json:"setCode,omitempty" db:"set_code"`
	DraftsCount   int       `json:"draftsCount" db:"drafts_count"`
	MatchesPlayed int       `json:"matchesPlayed" db:"matches_played"`
	MatchesWon    int       `json:"matchesWon" db:"matches_won"`
	AvgDraftGrade *float64  `json:"avgDraftGrade,omitempty" db:"avg_draft_grade"`
	CalculatedAt  time.Time `json:"calculatedAt" db:"calculated_at"`
}

// WinRate returns the match win rate for this period.
func (d *DraftTemporalTrend) WinRate() float64 {
	if d.MatchesPlayed == 0 {
		return 0
	}
	return float64(d.MatchesWon) / float64(d.MatchesPlayed)
}

// DraftPatternAnalysis stores analyzed drafting patterns.
type DraftPatternAnalysis struct {
	ID                    int64     `json:"id" db:"id"`
	SetCode               *string   `json:"setCode,omitempty" db:"set_code"`
	ColorPreferenceJSON   string    `json:"-" db:"color_preference_json"`
	TypePreferenceJSON    string    `json:"-" db:"type_preference_json"`
	PickOrderPatternJSON  string    `json:"-" db:"pick_order_pattern_json"`
	ArchetypeAffinityJSON string    `json:"-" db:"archetype_affinity_json"`
	SampleSize            int       `json:"sampleSize" db:"sample_size"`
	CalculatedAt          time.Time `json:"calculatedAt" db:"calculated_at"`
}

// ColorPreference represents a color with its pick preference metrics.
type ColorPreference struct {
	Color         string  `json:"color"`
	TotalPicks    int     `json:"totalPicks"`
	PercentOfPool float64 `json:"percentOfPool"`
	AvgPickOrder  float64 `json:"avgPickOrder"`
}

// GetColorPreferences parses the color preferences from JSON.
func (d *DraftPatternAnalysis) GetColorPreferences() ([]ColorPreference, error) {
	if d.ColorPreferenceJSON == "" {
		return nil, nil
	}
	var prefs []ColorPreference
	if err := json.Unmarshal([]byte(d.ColorPreferenceJSON), &prefs); err != nil {
		return nil, err
	}
	return prefs, nil
}

// SetColorPreferences serializes color preferences to JSON.
func (d *DraftPatternAnalysis) SetColorPreferences(prefs []ColorPreference) error {
	data, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	d.ColorPreferenceJSON = string(data)
	return nil
}

// TypePreference represents a card type with its pick preference metrics.
type TypePreference struct {
	Type          string  `json:"type"` // Creature, Instant, Sorcery, etc.
	TotalPicks    int     `json:"totalPicks"`
	PercentOfPool float64 `json:"percentOfPool"`
	AvgPickOrder  float64 `json:"avgPickOrder"`
}

// GetTypePreferences parses the type preferences from JSON.
func (d *DraftPatternAnalysis) GetTypePreferences() ([]TypePreference, error) {
	if d.TypePreferenceJSON == "" {
		return nil, nil
	}
	var prefs []TypePreference
	if err := json.Unmarshal([]byte(d.TypePreferenceJSON), &prefs); err != nil {
		return nil, err
	}
	return prefs, nil
}

// SetTypePreferences serializes type preferences to JSON.
func (d *DraftPatternAnalysis) SetTypePreferences(prefs []TypePreference) error {
	data, err := json.Marshal(prefs)
	if err != nil {
		return err
	}
	d.TypePreferenceJSON = string(data)
	return nil
}

// PickOrderPattern represents pick order tendencies.
type PickOrderPattern struct {
	Phase       string  `json:"phase"` // early (1-5), mid (6-10), late (11-14)
	AvgRating   float64 `json:"avgRating"`
	TotalPicks  int     `json:"totalPicks"`
	RarePicks   int     `json:"rarePicks"`
	CommonPicks int     `json:"commonPicks"`
}

// GetPickOrderPatterns parses the pick order patterns from JSON.
func (d *DraftPatternAnalysis) GetPickOrderPatterns() ([]PickOrderPattern, error) {
	if d.PickOrderPatternJSON == "" {
		return nil, nil
	}
	var patterns []PickOrderPattern
	if err := json.Unmarshal([]byte(d.PickOrderPatternJSON), &patterns); err != nil {
		return nil, err
	}
	return patterns, nil
}

// SetPickOrderPatterns serializes pick order patterns to JSON.
func (d *DraftPatternAnalysis) SetPickOrderPatterns(patterns []PickOrderPattern) error {
	data, err := json.Marshal(patterns)
	if err != nil {
		return err
	}
	d.PickOrderPatternJSON = string(data)
	return nil
}

// ArchetypeAffinity represents affinity toward a specific archetype.
type ArchetypeAffinity struct {
	ColorPair     string  `json:"colorPair"`
	ArchetypeName string  `json:"archetypeName"`
	TimesBuilt    int     `json:"timesBuilt"`
	AvgWinRate    float64 `json:"avgWinRate"`
	AffinityScore float64 `json:"affinityScore"` // 0-1, how often player gravitates to this
}

// GetArchetypeAffinities parses the archetype affinities from JSON.
func (d *DraftPatternAnalysis) GetArchetypeAffinities() ([]ArchetypeAffinity, error) {
	if d.ArchetypeAffinityJSON == "" {
		return nil, nil
	}
	var affinities []ArchetypeAffinity
	if err := json.Unmarshal([]byte(d.ArchetypeAffinityJSON), &affinities); err != nil {
		return nil, err
	}
	return affinities, nil
}

// SetArchetypeAffinities serializes archetype affinities to JSON.
func (d *DraftPatternAnalysis) SetArchetypeAffinities(affinities []ArchetypeAffinity) error {
	data, err := json.Marshal(affinities)
	if err != nil {
		return err
	}
	d.ArchetypeAffinityJSON = string(data)
	return nil
}

// TrendDirection represents the direction of a performance trend.
type TrendDirection string

const (
	TrendDirectionImproving TrendDirection = "improving"
	TrendDirectionStable    TrendDirection = "stable"
	TrendDirectionDeclining TrendDirection = "declining"
)

// AnalyzeTrendDirection analyzes win rate trends over multiple periods.
func AnalyzeTrendDirection(trends []DraftTemporalTrend) TrendDirection {
	if len(trends) < 2 {
		return TrendDirectionStable
	}

	// Sort by period start (ascending)
	// Calculate win rate change from first to last
	firstWinRate := trends[0].WinRate()
	lastWinRate := trends[len(trends)-1].WinRate()
	delta := lastWinRate - firstWinRate

	const threshold = 0.03 // 3% change threshold
	if delta > threshold {
		return TrendDirectionImproving
	} else if delta < -threshold {
		return TrendDirectionDeclining
	}
	return TrendDirectionStable
}

// PeriodType constants for temporal trends.
const (
	PeriodTypeWeek  = "week"
	PeriodTypeMonth = "month"
)

// DraftMatchResultType constants.
const (
	DraftMatchResultWin  = "win"
	DraftMatchResultLoss = "loss"
)

// ArchetypeTarget defines deck composition targets for different archetypes.
type ArchetypeTarget struct {
	Name           string  `json:"name"`
	CreatureMin    int     `json:"creatureMin"`
	CreatureMax    int     `json:"creatureMax"`
	MaxAvgCMC      float64 `json:"maxAvgCmc"`
	LandCount      int     `json:"landCount"`
	PreferredCurve []int   `json:"preferredCurve"` // [0, 1-drop, 2-drop, 3-drop, 4-drop, 5+]
	Description    string  `json:"description"`
}

// StandardArchetypeTargets defines the default archetype targets.
var StandardArchetypeTargets = map[string]ArchetypeTarget{
	"aggro": {
		Name:           "Aggro",
		CreatureMin:    16,
		CreatureMax:    18,
		MaxAvgCMC:      2.5,
		LandCount:      16,
		PreferredCurve: []int{0, 4, 8, 5, 1, 0},
		Description:    "Fast, creature-heavy decks that aim to win early",
	},
	"midrange": {
		Name:           "Midrange",
		CreatureMin:    14,
		CreatureMax:    16,
		MaxAvgCMC:      3.0,
		LandCount:      17,
		PreferredCurve: []int{0, 2, 5, 5, 3, 2},
		Description:    "Balanced decks with good creatures and removal",
	},
	"control": {
		Name:           "Control",
		CreatureMin:    10,
		CreatureMax:    12,
		MaxAvgCMC:      3.5,
		LandCount:      18,
		PreferredCurve: []int{0, 1, 4, 4, 3, 3},
		Description:    "Slower decks focused on removal and card advantage",
	},
}

// DraftAnalyticsSummary provides a summary view of all analytics.
type DraftAnalyticsSummary struct {
	TotalDrafts     int                   `json:"totalDrafts"`
	TotalMatches    int                   `json:"totalMatches"`
	OverallWinRate  float64               `json:"overallWinRate"`
	BestArchetype   *DraftArchetypeStats  `json:"bestArchetype,omitempty"`
	WorstArchetype  *DraftArchetypeStats  `json:"worstArchetype,omitempty"`
	RecentTrend     TrendDirection        `json:"recentTrend"`
	CommunityRank   *float64              `json:"communityRank,omitempty"`
	PatternAnalysis *DraftPatternAnalysis `json:"patternAnalysis,omitempty"`
}
