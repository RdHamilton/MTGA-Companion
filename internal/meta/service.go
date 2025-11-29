package meta

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// Service aggregates meta data from multiple sources.
type Service struct {
	goldfishClient *GoldfishClient
	top8Client     *Top8Client
	mu             sync.RWMutex
}

// ServiceConfig configures the meta service.
type ServiceConfig struct {
	GoldfishConfig *GoldfishConfig
	Top8Config     *Top8Config
}

// AggregatedMeta combines meta data from all sources.
type AggregatedMeta struct {
	Format          string                 `json:"format"`
	Archetypes      []*AggregatedArchetype `json:"archetypes"`
	TopDecks        []*MetaDeck            `json:"top_decks"`
	Tournaments     []*Tournament          `json:"tournaments,omitempty"`
	TotalArchetypes int                    `json:"total_archetypes"`
	LastUpdated     time.Time              `json:"last_updated"`
	Sources         []string               `json:"sources"`
}

// AggregatedArchetype combines archetype data from multiple sources.
type AggregatedArchetype struct {
	Name            string    `json:"name"`
	NormalizedName  string    `json:"normalized_name"`
	Colors          []string  `json:"colors"`
	MetaShare       float64   `json:"meta_share"`       // From MTGGoldfish
	TournamentTop8s int       `json:"tournament_top8s"` // From MTGTop8
	TournamentWins  int       `json:"tournament_wins"`  // From MTGTop8
	Tier            int       `json:"tier"`             // 1-4 based on combined data
	ConfidenceScore float64   `json:"confidence_score"` // How reliable the data is
	TrendDirection  string    `json:"trend_direction"`  // "up", "down", "stable"
	LastSeenInMeta  time.Time `json:"last_seen_in_meta"`
	LastSeenInTop8  time.Time `json:"last_seen_in_top8,omitempty"`
}

// NewService creates a new meta service.
func NewService(config *ServiceConfig) *Service {
	if config == nil {
		config = &ServiceConfig{}
	}

	return &Service{
		goldfishClient: NewGoldfishClient(config.GoldfishConfig),
		top8Client:     NewTop8Client(config.Top8Config),
	}
}

// GetAggregatedMeta returns combined meta data from all sources.
func (s *Service) GetAggregatedMeta(ctx context.Context, format string) (*AggregatedMeta, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Fetch from both sources concurrently
	type goldfishResult struct {
		meta *FormatMeta
		err  error
	}
	type top8Result struct {
		meta *TournamentMeta
		err  error
	}

	goldfishCh := make(chan goldfishResult, 1)
	top8Ch := make(chan top8Result, 1)

	go func() {
		meta, err := s.goldfishClient.GetMeta(ctx, format)
		goldfishCh <- goldfishResult{meta, err}
	}()

	go func() {
		meta, err := s.top8Client.GetTournamentMeta(ctx, format)
		top8Ch <- top8Result{meta, err}
	}()

	// Collect results
	var goldfishMeta *FormatMeta
	var top8Meta *TournamentMeta
	var sources []string

	gfResult := <-goldfishCh
	if gfResult.err == nil && gfResult.meta != nil {
		goldfishMeta = gfResult.meta
		sources = append(sources, "mtggoldfish")
	}

	t8Result := <-top8Ch
	if t8Result.err == nil && t8Result.meta != nil {
		top8Meta = t8Result.meta
		sources = append(sources, "mtgtop8")
	}

	if goldfishMeta == nil && top8Meta == nil {
		return nil, fmt.Errorf("failed to fetch meta from any source")
	}

	// Aggregate the data
	aggregated := s.aggregateData(format, goldfishMeta, top8Meta)
	aggregated.Sources = sources
	aggregated.LastUpdated = time.Now()

	return aggregated, nil
}

// aggregateData combines data from both sources.
func (s *Service) aggregateData(format string, goldfish *FormatMeta, top8 *TournamentMeta) *AggregatedMeta {
	archetypeMap := make(map[string]*AggregatedArchetype)

	// Process MTGGoldfish data
	if goldfish != nil {
		for _, deck := range goldfish.Decks {
			normalized := strings.ToLower(deck.ArchetypeName)
			if _, exists := archetypeMap[normalized]; !exists {
				archetypeMap[normalized] = &AggregatedArchetype{
					Name:           deck.Name,
					NormalizedName: normalized,
					Colors:         deck.Colors,
					Tier:           deck.Tier,
					LastSeenInMeta: deck.LastUpdated,
				}
			}
			archetypeMap[normalized].MetaShare = deck.MetaShare
			archetypeMap[normalized].LastSeenInMeta = deck.LastUpdated
		}
	}

	// Process MTGTop8 data
	if top8 != nil {
		for name, stats := range top8.ArchetypeStats {
			normalized := strings.ToLower(name)
			if existing, exists := archetypeMap[normalized]; exists {
				existing.TournamentTop8s = stats.Top8Count
				existing.TournamentWins = stats.WinCount
				existing.LastSeenInTop8 = stats.LastSeen
				if stats.TrendDirection != "" {
					existing.TrendDirection = stats.TrendDirection
				}
				// Merge colors if needed
				if len(existing.Colors) == 0 {
					existing.Colors = stats.Colors
				}
			} else {
				archetypeMap[normalized] = &AggregatedArchetype{
					Name:            stats.ArchetypeName,
					NormalizedName:  normalized,
					Colors:          stats.Colors,
					TournamentTop8s: stats.Top8Count,
					TournamentWins:  stats.WinCount,
					TrendDirection:  stats.TrendDirection,
					LastSeenInTop8:  stats.LastSeen,
				}
			}
		}
	}

	// Calculate confidence scores and finalize tiers
	archetypes := make([]*AggregatedArchetype, 0, len(archetypeMap))
	for _, arch := range archetypeMap {
		arch.ConfidenceScore = s.calculateConfidence(arch)
		if arch.Tier == 0 {
			arch.Tier = s.calculateTier(arch)
		}
		if arch.TrendDirection == "" {
			arch.TrendDirection = "stable"
		}
		archetypes = append(archetypes, arch)
	}

	// Sort by combined relevance (meta share + tournament presence)
	sort.Slice(archetypes, func(i, j int) bool {
		scoreI := archetypes[i].MetaShare + float64(archetypes[i].TournamentTop8s)*0.5
		scoreJ := archetypes[j].MetaShare + float64(archetypes[j].TournamentTop8s)*0.5
		return scoreI > scoreJ
	})

	// Build top decks list from goldfish
	var topDecks []*MetaDeck
	if goldfish != nil {
		topDecks = goldfish.Decks
	}

	// Build tournaments list from top8
	var tournaments []*Tournament
	if top8 != nil {
		tournaments = top8.Tournaments
	}

	return &AggregatedMeta{
		Format:          format,
		Archetypes:      archetypes,
		TopDecks:        topDecks,
		Tournaments:     tournaments,
		TotalArchetypes: len(archetypes),
	}
}

// calculateConfidence calculates how confident we are in the archetype data.
func (s *Service) calculateConfidence(arch *AggregatedArchetype) float64 {
	confidence := 0.0

	// Has meta share data
	if arch.MetaShare > 0 {
		confidence += 0.4
	}

	// Has tournament data
	if arch.TournamentTop8s > 0 {
		confidence += 0.3
		// More top 8s = more confidence
		if arch.TournamentTop8s >= 10 {
			confidence += 0.1
		}
		if arch.TournamentTop8s >= 20 {
			confidence += 0.1
		}
	}

	// Has color data
	if len(arch.Colors) > 0 {
		confidence += 0.1
	}

	return confidence
}

// calculateTier determines the tier based on available data.
func (s *Service) calculateTier(arch *AggregatedArchetype) int {
	// Use meta share if available
	if arch.MetaShare >= 5.0 {
		return 1
	}
	if arch.MetaShare >= 2.0 {
		return 2
	}
	if arch.MetaShare >= 0.5 {
		return 3
	}

	// Fall back to tournament data
	if arch.TournamentTop8s >= 20 {
		return 1
	}
	if arch.TournamentTop8s >= 10 {
		return 2
	}
	if arch.TournamentTop8s >= 5 {
		return 3
	}

	return 4 // Untiered
}

// GetTopArchetypes returns the top N archetypes for a format.
func (s *Service) GetTopArchetypes(ctx context.Context, format string, limit int) ([]*AggregatedArchetype, error) {
	meta, err := s.GetAggregatedMeta(ctx, format)
	if err != nil {
		return nil, err
	}

	if limit <= 0 || limit > len(meta.Archetypes) {
		return meta.Archetypes, nil
	}

	return meta.Archetypes[:limit], nil
}

// GetArchetypeByName finds an archetype by name.
func (s *Service) GetArchetypeByName(ctx context.Context, format, name string) (*AggregatedArchetype, error) {
	meta, err := s.GetAggregatedMeta(ctx, format)
	if err != nil {
		return nil, err
	}

	nameLower := strings.ToLower(name)
	for _, arch := range meta.Archetypes {
		if arch.NormalizedName == nameLower || strings.Contains(arch.NormalizedName, nameLower) {
			return arch, nil
		}
	}

	return nil, fmt.Errorf("archetype not found: %s", name)
}

// GetTier1Archetypes returns all tier 1 archetypes for a format.
func (s *Service) GetTier1Archetypes(ctx context.Context, format string) ([]*AggregatedArchetype, error) {
	meta, err := s.GetAggregatedMeta(ctx, format)
	if err != nil {
		return nil, err
	}

	tier1 := make([]*AggregatedArchetype, 0)
	for _, arch := range meta.Archetypes {
		if arch.Tier == 1 {
			tier1 = append(tier1, arch)
		}
	}

	return tier1, nil
}

// GetArchetypesByColors returns archetypes matching the given colors.
func (s *Service) GetArchetypesByColors(ctx context.Context, format string, colors []string) ([]*AggregatedArchetype, error) {
	meta, err := s.GetAggregatedMeta(ctx, format)
	if err != nil {
		return nil, err
	}

	matching := make([]*AggregatedArchetype, 0)
	for _, arch := range meta.Archetypes {
		if s.colorsMatch(arch.Colors, colors) {
			matching = append(matching, arch)
		}
	}

	return matching, nil
}

// colorsMatch checks if archetype colors match the given colors.
func (s *Service) colorsMatch(archetypeColors, targetColors []string) bool {
	if len(targetColors) == 0 {
		return true
	}

	// Check if all target colors are in archetype colors
	for _, target := range targetColors {
		found := false
		for _, arch := range archetypeColors {
			if arch == target {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}

// GetRecentTournaments returns recent tournaments for a format.
func (s *Service) GetRecentTournaments(ctx context.Context, format string, limit int) ([]*Tournament, error) {
	return s.top8Client.GetRecentTournaments(ctx, format, limit)
}

// RefreshAll forces a refresh of all meta data for a format.
func (s *Service) RefreshAll(ctx context.Context, format string) (*AggregatedMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear caches
	s.goldfishClient.ClearCache()
	s.top8Client.ClearCache()

	// Re-read lock for GetAggregatedMeta
	s.mu.RUnlock()
	defer s.mu.RLock()

	return s.GetAggregatedMeta(ctx, format)
}

// GetSupportedFormats returns the list of supported formats.
func (s *Service) GetSupportedFormats() []string {
	return []string{
		"standard",
		"historic",
		"explorer",
		"pioneer",
		"modern",
		"legacy",
		"vintage",
		"pauper",
		"alchemy",
		"timeless",
	}
}

// IsFormatSupported checks if a format is supported.
func (s *Service) IsFormatSupported(format string) bool {
	formatLower := strings.ToLower(format)
	for _, f := range s.GetSupportedFormats() {
		if f == formatLower {
			return true
		}
	}
	return false
}
