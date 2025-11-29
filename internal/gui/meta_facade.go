package gui

import (
	"context"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/meta"
)

// MetaFacade handles meta/metagame data for the frontend.
type MetaFacade struct {
	metaService *meta.Service
}

// NewMetaFacade creates a new MetaFacade.
func NewMetaFacade(metaService *meta.Service) *MetaFacade {
	return &MetaFacade{
		metaService: metaService,
	}
}

// MetaDashboardResponse represents the response for the meta dashboard.
type MetaDashboardResponse struct {
	Format          string            `json:"format"`
	Archetypes      []*ArchetypeInfo  `json:"archetypes"`
	Tournaments     []*TournamentInfo `json:"tournaments,omitempty"`
	TotalArchetypes int               `json:"totalArchetypes"`
	LastUpdated     time.Time         `json:"lastUpdated"`
	Sources         []string          `json:"sources"`
	Error           string            `json:"error,omitempty"`
}

// ArchetypeInfo represents archetype information for the frontend.
type ArchetypeInfo struct {
	Name            string   `json:"name"`
	Colors          []string `json:"colors"`
	MetaShare       float64  `json:"metaShare"`
	TournamentTop8s int      `json:"tournamentTop8s"`
	TournamentWins  int      `json:"tournamentWins"`
	Tier            int      `json:"tier"`
	ConfidenceScore float64  `json:"confidenceScore"`
	TrendDirection  string   `json:"trendDirection"`
}

// TournamentInfo represents tournament information for the frontend.
type TournamentInfo struct {
	Name      string    `json:"name"`
	Date      time.Time `json:"date"`
	Players   int       `json:"players"`
	Format    string    `json:"format"`
	TopDecks  []string  `json:"topDecks"`
	SourceURL string    `json:"sourceUrl,omitempty"`
}

// GetMetaDashboard returns meta information for the dashboard.
func (m *MetaFacade) GetMetaDashboard(ctx context.Context, format string) (*MetaDashboardResponse, error) {
	if m.metaService == nil {
		return &MetaDashboardResponse{
			Format: format,
			Error:  "Meta service not available",
		}, nil
	}

	if format == "" {
		format = "standard"
	}

	aggregated, err := m.metaService.GetAggregatedMeta(ctx, format)
	if err != nil {
		return &MetaDashboardResponse{
			Format: format,
			Error:  err.Error(),
		}, nil
	}

	// Convert archetypes
	archetypes := make([]*ArchetypeInfo, 0, len(aggregated.Archetypes))
	for _, arch := range aggregated.Archetypes {
		archetypes = append(archetypes, &ArchetypeInfo{
			Name:            arch.Name,
			Colors:          arch.Colors,
			MetaShare:       arch.MetaShare,
			TournamentTop8s: arch.TournamentTop8s,
			TournamentWins:  arch.TournamentWins,
			Tier:            arch.Tier,
			ConfidenceScore: arch.ConfidenceScore,
			TrendDirection:  arch.TrendDirection,
		})
	}

	// Convert tournaments
	var tournaments []*TournamentInfo
	if aggregated.Tournaments != nil {
		tournaments = make([]*TournamentInfo, 0, len(aggregated.Tournaments))
		for _, t := range aggregated.Tournaments {
			topDecks := make([]string, 0, len(t.TopDecks))
			for _, d := range t.TopDecks {
				topDecks = append(topDecks, d.ArchetypeName)
			}
			tournaments = append(tournaments, &TournamentInfo{
				Name:      t.Name,
				Date:      t.Date,
				Players:   t.Players,
				Format:    t.Format,
				TopDecks:  topDecks,
				SourceURL: t.URL,
			})
		}
	}

	return &MetaDashboardResponse{
		Format:          format,
		Archetypes:      archetypes,
		Tournaments:     tournaments,
		TotalArchetypes: aggregated.TotalArchetypes,
		LastUpdated:     aggregated.LastUpdated,
		Sources:         aggregated.Sources,
	}, nil
}

// RefreshMetaData forces a refresh of meta data for a format.
func (m *MetaFacade) RefreshMetaData(ctx context.Context, format string) (*MetaDashboardResponse, error) {
	if m.metaService == nil {
		return &MetaDashboardResponse{
			Format: format,
			Error:  "Meta service not available",
		}, nil
	}

	if format == "" {
		format = "standard"
	}

	aggregated, err := m.metaService.RefreshAll(ctx, format)
	if err != nil {
		return &MetaDashboardResponse{
			Format: format,
			Error:  err.Error(),
		}, nil
	}

	// Convert archetypes
	archetypes := make([]*ArchetypeInfo, 0, len(aggregated.Archetypes))
	for _, arch := range aggregated.Archetypes {
		archetypes = append(archetypes, &ArchetypeInfo{
			Name:            arch.Name,
			Colors:          arch.Colors,
			MetaShare:       arch.MetaShare,
			TournamentTop8s: arch.TournamentTop8s,
			TournamentWins:  arch.TournamentWins,
			Tier:            arch.Tier,
			ConfidenceScore: arch.ConfidenceScore,
			TrendDirection:  arch.TrendDirection,
		})
	}

	// Convert tournaments
	var tournaments []*TournamentInfo
	if aggregated.Tournaments != nil {
		tournaments = make([]*TournamentInfo, 0, len(aggregated.Tournaments))
		for _, t := range aggregated.Tournaments {
			topDecks := make([]string, 0, len(t.TopDecks))
			for _, d := range t.TopDecks {
				topDecks = append(topDecks, d.ArchetypeName)
			}
			tournaments = append(tournaments, &TournamentInfo{
				Name:      t.Name,
				Date:      t.Date,
				Players:   t.Players,
				Format:    t.Format,
				TopDecks:  topDecks,
				SourceURL: t.URL,
			})
		}
	}

	return &MetaDashboardResponse{
		Format:          format,
		Archetypes:      archetypes,
		Tournaments:     tournaments,
		TotalArchetypes: aggregated.TotalArchetypes,
		LastUpdated:     aggregated.LastUpdated,
		Sources:         aggregated.Sources,
	}, nil
}

// GetSupportedFormats returns the list of supported formats.
func (m *MetaFacade) GetSupportedFormats() []string {
	if m.metaService == nil {
		return []string{"standard", "historic", "explorer", "pioneer", "modern"}
	}
	return m.metaService.GetSupportedFormats()
}

// GetTierArchetypes returns archetypes for a specific tier.
func (m *MetaFacade) GetTierArchetypes(ctx context.Context, format string, tier int) ([]*ArchetypeInfo, error) {
	if m.metaService == nil {
		return nil, nil
	}

	aggregated, err := m.metaService.GetAggregatedMeta(ctx, format)
	if err != nil {
		return nil, err
	}

	archetypes := make([]*ArchetypeInfo, 0)
	for _, arch := range aggregated.Archetypes {
		if arch.Tier == tier {
			archetypes = append(archetypes, &ArchetypeInfo{
				Name:            arch.Name,
				Colors:          arch.Colors,
				MetaShare:       arch.MetaShare,
				TournamentTop8s: arch.TournamentTop8s,
				TournamentWins:  arch.TournamentWins,
				Tier:            arch.Tier,
				ConfidenceScore: arch.ConfidenceScore,
				TrendDirection:  arch.TrendDirection,
			})
		}
	}

	return archetypes, nil
}
