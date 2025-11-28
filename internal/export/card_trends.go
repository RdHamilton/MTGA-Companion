package export

import (
	"context"
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// TrendExportRow represents a single data point in a card trend export.
type TrendExportRow struct {
	Date        string  `csv:"Date" json:"date"`
	GIHWR       float64 `csv:"GIH WR" json:"gihwr"`
	OHWR        float64 `csv:"OH WR" json:"ohwr"`
	ALSA        float64 `csv:"ALSA" json:"alsa"`
	ATA         float64 `csv:"ATA" json:"ata"`
	SampleSize  int     `csv:"Sample Size" json:"sample_size"`
	GamesPlayed int     `csv:"Games Played" json:"games_played"`
}

// CardTrendExport combines metadata with trend data.
type CardTrendExport struct {
	ArenaID     int               `json:"arena_id"`
	CardName    string            `json:"card_name"`
	Expansion   string            `json:"expansion"`
	Format      string            `json:"format"`
	Colors      string            `json:"colors"`
	TotalPoints int               `json:"total_points"`
	Points      []*TrendExportRow `json:"points"`
}

// ExportCardPerformanceTrend exports a card's performance trend over time.
func ExportCardPerformanceTrend(ctx context.Context, service *storage.Service, arenaID int, expansion string, days int, opts Options) error {
	// Get card trend data
	trend, err := service.GetCardWinRateTrend(ctx, arenaID, expansion, days)
	if err != nil {
		return fmt.Errorf("failed to get card trend: %w", err)
	}

	if trend == nil || len(trend.Points) == 0 {
		return fmt.Errorf("no trend data found for card %d in set %s", arenaID, expansion)
	}

	// Convert trend points to export rows
	rows := make([]*TrendExportRow, 0, len(trend.Points))
	for _, point := range trend.Points {
		rows = append(rows, &TrendExportRow{
			Date:        point.Date.Format("2006-01-02"),
			GIHWR:       point.GIHWR,
			OHWR:        point.OHWR,
			ALSA:        point.ALSA,
			ATA:         point.ATA,
			SampleSize:  point.SampleSize,
			GamesPlayed: point.GamesPlayed,
		})
	}

	// Get card name if available
	cardName := trend.CardName
	if cardName == "" {
		card, err := service.SetCardRepo().GetCardByArenaID(ctx, fmt.Sprintf("%d", arenaID))
		if err == nil && card != nil {
			cardName = card.Name
		}
	}

	// Export based on format
	exporter := NewExporter(opts)

	switch opts.Format {
	case FormatJSON:
		exportData := CardTrendExport{
			ArenaID:     arenaID,
			CardName:    cardName,
			Expansion:   expansion,
			Format:      trend.Format,
			Colors:      trend.Colors,
			TotalPoints: len(rows),
			Points:      rows,
		}
		return exporter.Export(exportData)
	case FormatCSV:
		return exporter.Export(rows)
	default:
		return fmt.Errorf("unsupported export format: %s", opts.Format)
	}
}

// MultipleCardTrendsExport exports multiple card trends for comparison.
type MultipleCardTrendsExport struct {
	Expansion string             `json:"expansion"`
	Format    string             `json:"format"`
	Days      int                `json:"days"`
	Cards     []*CardTrendExport `json:"cards"`
}

// ExportMultipleCardTrends exports trends for multiple cards for comparison.
func ExportMultipleCardTrends(ctx context.Context, service *storage.Service, arenaIDs []int, expansion string, days int, opts Options) error {
	trends := make([]*CardTrendExport, 0, len(arenaIDs))

	for _, arenaID := range arenaIDs {
		// Get card trend data
		trend, err := service.GetCardWinRateTrend(ctx, arenaID, expansion, days)
		if err != nil {
			// Skip cards with errors
			continue
		}

		if trend == nil || len(trend.Points) == 0 {
			continue
		}

		// Convert trend points to export rows
		rows := make([]*TrendExportRow, 0, len(trend.Points))
		for _, point := range trend.Points {
			rows = append(rows, &TrendExportRow{
				Date:        point.Date.Format("2006-01-02"),
				GIHWR:       point.GIHWR,
				OHWR:        point.OHWR,
				ALSA:        point.ALSA,
				ATA:         point.ATA,
				SampleSize:  point.SampleSize,
				GamesPlayed: point.GamesPlayed,
			})
		}

		// Get card name if available
		cardName := trend.CardName
		if cardName == "" {
			card, err := service.SetCardRepo().GetCardByArenaID(ctx, fmt.Sprintf("%d", arenaID))
			if err == nil && card != nil {
				cardName = card.Name
			}
		}

		trendExport := &CardTrendExport{
			ArenaID:     arenaID,
			CardName:    cardName,
			Expansion:   expansion,
			Format:      trend.Format,
			Colors:      trend.Colors,
			TotalPoints: len(rows),
			Points:      rows,
		}

		trends = append(trends, trendExport)
	}

	if len(trends) == 0 {
		return fmt.Errorf("no trend data found for any of the specified cards")
	}

	// Export based on format
	exporter := NewExporter(opts)

	switch opts.Format {
	case FormatJSON:
		// Get format from first trend
		format := ""
		if len(trends) > 0 && len(trends[0].Points) > 0 {
			format = trends[0].Format
		}

		exportData := MultipleCardTrendsExport{
			Expansion: expansion,
			Format:    format,
			Days:      days,
			Cards:     trends,
		}
		return exporter.Export(exportData)
	case FormatCSV:
		// For CSV, flatten all trends into rows with card identifier
		type FlatTrendRow struct {
			CardName    string  `csv:"Card Name"`
			ArenaID     int     `csv:"Arena ID"`
			Date        string  `csv:"Date"`
			GIHWR       float64 `csv:"GIH WR"`
			OHWR        float64 `csv:"OH WR"`
			ALSA        float64 `csv:"ALSA"`
			ATA         float64 `csv:"ATA"`
			SampleSize  int     `csv:"Sample Size"`
			GamesPlayed int     `csv:"Games Played"`
		}

		flatRows := make([]*FlatTrendRow, 0)
		for _, trend := range trends {
			for _, point := range trend.Points {
				flatRows = append(flatRows, &FlatTrendRow{
					CardName:    trend.CardName,
					ArenaID:     trend.ArenaID,
					Date:        point.Date,
					GIHWR:       point.GIHWR,
					OHWR:        point.OHWR,
					ALSA:        point.ALSA,
					ATA:         point.ATA,
					SampleSize:  point.SampleSize,
					GamesPlayed: point.GamesPlayed,
				})
			}
		}

		return exporter.Export(flatRows)
	default:
		return fmt.Errorf("unsupported export format: %s", opts.Format)
	}
}
