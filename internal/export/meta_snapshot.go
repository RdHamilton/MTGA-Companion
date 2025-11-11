package export

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// MetaSnapshotRow represents a card in the meta snapshot export.
type MetaSnapshotRow struct {
	Rank        int     `csv:"Rank" json:"rank"`
	Name        string  `csv:"Name" json:"name"`
	SetCode     string  `csv:"Set" json:"set_code"`
	Rarity      string  `csv:"Rarity" json:"rarity"`
	Colors      string  `csv:"Colors" json:"colors"`
	CMC         float64 `csv:"CMC" json:"cmc"`
	GIHWR       float64 `csv:"GIH WR" json:"gihwr"`
	OHWR        float64 `csv:"OH WR" json:"ohwr"`
	ALSA        float64 `csv:"ALSA" json:"alsa"`
	ATA         float64 `csv:"ATA" json:"ata"`
	GIH         int     `csv:"GIH" json:"gih"`
	GamesPlayed int     `csv:"Games Played" json:"games_played"`
	GIHWRDelta  float64 `csv:"GIHWR Delta" json:"gihwr_delta"`
}

// MetaSnapshotExport combines metadata with snapshot data.
type MetaSnapshotExport struct {
	SetCode     string             `json:"set_code"`
	Format      string             `json:"format"`
	Colors      string             `json:"colors"`
	GeneratedAt time.Time          `json:"generated_at"`
	TopN        int                `json:"top_n"`
	TotalCards  int                `json:"total_cards"`
	Cards       []*MetaSnapshotRow `json:"cards"`
}

// ExportMetaSnapshot exports top cards for a set as a meta snapshot.
func ExportMetaSnapshot(ctx context.Context, service *storage.Service, setCode, format, colors string, topN int, opts Options) error {
	// Get all card ratings for the set
	ratings, err := service.GetCardRatingsForSet(ctx, setCode, format, colors)
	if err != nil {
		return fmt.Errorf("failed to get card ratings: %w", err)
	}

	if len(ratings) == 0 {
		return fmt.Errorf("no card ratings found for set %s (format: %s, colors: %s)", setCode, format, colors)
	}

	// Sort by GIHWR descending
	sort.Slice(ratings, func(i, j int) bool {
		return ratings[i].GIHWR > ratings[j].GIHWR
	})

	// Take top N cards
	if topN > 0 && topN < len(ratings) {
		ratings = ratings[:topN]
	}

	// Convert to export rows
	rows := make([]*MetaSnapshotRow, 0, len(ratings))
	for i, rating := range ratings {
		// Get card metadata
		card, err := service.GetCardByArenaID(ctx, rating.ArenaID)
		if err != nil || card == nil {
			// Skip cards we don't have metadata for
			continue
		}

		row := &MetaSnapshotRow{
			Rank:        i + 1,
			Name:        card.Name,
			SetCode:     card.SetCode,
			Rarity:      card.Rarity,
			Colors:      strings.Join(card.Colors, ""),
			CMC:         card.CMC,
			GIHWR:       rating.GIHWR,
			OHWR:        rating.OHWR,
			ALSA:        rating.ALSA,
			ATA:         rating.ATA,
			GIH:         rating.GIH,
			GamesPlayed: rating.GamesPlayed,
			GIHWRDelta:  rating.GIHWRDelta,
		}

		rows = append(rows, row)
	}

	// Export based on format
	exporter := NewExporter(opts)

	switch opts.Format {
	case FormatJSON:
		exportData := MetaSnapshotExport{
			SetCode:     setCode,
			Format:      format,
			Colors:      colors,
			GeneratedAt: time.Now(),
			TopN:        topN,
			TotalCards:  len(rows),
			Cards:       rows,
		}
		return exporter.Export(exportData)
	case FormatCSV:
		return exporter.Export(rows)
	default:
		return fmt.Errorf("unsupported export format: %s", opts.Format)
	}
}

// ExportMetaSnapshotByRarity exports meta snapshot grouped by rarity.
func ExportMetaSnapshotByRarity(ctx context.Context, service *storage.Service, setCode, format, colors string, topNPerRarity int, opts Options) error {
	// Get all card ratings for the set
	ratings, err := service.GetCardRatingsForSet(ctx, setCode, format, colors)
	if err != nil {
		return fmt.Errorf("failed to get card ratings: %w", err)
	}

	if len(ratings) == 0 {
		return fmt.Errorf("no card ratings found for set %s", setCode)
	}

	// Group by rarity
	byRarity := make(map[string][]*storage.DraftCardRating)
	for _, rating := range ratings {
		card, err := service.GetCardByArenaID(ctx, rating.ArenaID)
		if err != nil || card == nil {
			continue
		}
		byRarity[card.Rarity] = append(byRarity[card.Rarity], rating)
	}

	// Sort each rarity group by GIHWR and take top N
	allRows := make([]*MetaSnapshotRow, 0)
	rarityOrder := []string{"mythic", "rare", "uncommon", "common"}

	for _, rarity := range rarityOrder {
		rarityRatings, exists := byRarity[rarity]
		if !exists || len(rarityRatings) == 0 {
			continue
		}

		// Sort by GIHWR descending
		sort.Slice(rarityRatings, func(i, j int) bool {
			return rarityRatings[i].GIHWR > rarityRatings[j].GIHWR
		})

		// Take top N for this rarity
		if topNPerRarity > 0 && topNPerRarity < len(rarityRatings) {
			rarityRatings = rarityRatings[:topNPerRarity]
		}

		// Convert to export rows
		for i, rating := range rarityRatings {
			card, err := service.GetCardByArenaID(ctx, rating.ArenaID)
			if err != nil || card == nil {
				continue
			}

			row := &MetaSnapshotRow{
				Rank:        i + 1,
				Name:        card.Name,
				SetCode:     card.SetCode,
				Rarity:      card.Rarity,
				Colors:      strings.Join(card.Colors, ""),
				CMC:         card.CMC,
				GIHWR:       rating.GIHWR,
				OHWR:        rating.OHWR,
				ALSA:        rating.ALSA,
				ATA:         rating.ATA,
				GIH:         rating.GIH,
				GamesPlayed: rating.GamesPlayed,
				GIHWRDelta:  rating.GIHWRDelta,
			}

			allRows = append(allRows, row)
		}
	}

	// Export based on format
	exporter := NewExporter(opts)

	switch opts.Format {
	case FormatJSON:
		exportData := MetaSnapshotExport{
			SetCode:     setCode,
			Format:      format,
			Colors:      colors,
			GeneratedAt: time.Now(),
			TopN:        topNPerRarity,
			TotalCards:  len(allRows),
			Cards:       allRows,
		}
		return exporter.Export(exportData)
	case FormatCSV:
		return exporter.Export(allRows)
	default:
		return fmt.Errorf("unsupported export format: %s", opts.Format)
	}
}
