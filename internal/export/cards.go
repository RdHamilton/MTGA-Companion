package export

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/unified"
)

// CardExportOptions configures card data export.
type CardExportOptions struct {
	Format        Format   // CSV, JSON, Markdown
	IncludeStats  bool     // Include draft statistics
	TopN          int      // Export only top N cards (0 = all)
	SortBy        string   // Sort field: gihwr, alsa, ata, name, cmc
	FilterRarity  []string // Filter by rarity (common, uncommon, rare, mythic)
	FilterColors  []string // Filter by color (W, U, B, R, G)
	IncludeFaces  bool     // Include card faces for MDFCs
	PrettyJSON    bool     // Pretty-print JSON
	ShowDataAge   bool     // Show data freshness indicators
	MinSampleSize int      // Minimum sample size for stats (0 = no filter)
	OnlyWithStats bool     // Export only cards with draft stats
}

// CardExportData represents the complete export data structure.
type CardExportData struct {
	SetCode    string              `json:"set_code,omitempty"`
	Format     string              `json:"format,omitempty"`
	ExportedAt time.Time           `json:"exported_at"`
	TotalCards int                 `json:"total_cards"`
	Cards      []*CardExportRow    `json:"cards"`
	Metadata   *CardExportMetadata `json:"metadata,omitempty"`
}

// CardExportMetadata provides context about the export.
type CardExportMetadata struct {
	DataAge            string `json:"data_age,omitempty"`
	MetadataSource     string `json:"metadata_source,omitempty"`
	StatsSource        string `json:"stats_source,omitempty"`
	CardsWithStats     int    `json:"cards_with_stats"`
	CardsWithoutStats  int    `json:"cards_without_stats"`
	AverageMetadataAge string `json:"average_metadata_age,omitempty"`
	AverageStatsAge    string `json:"average_stats_age,omitempty"`
}

// CardExportRow represents a single card for export (flattened for CSV).
type CardExportRow struct {
	// Core identity
	ArenaID int    `json:"arena_id" csv:"Arena ID"`
	Name    string `json:"name" csv:"Name"`

	// Card metadata
	ManaCost        string  `json:"mana_cost" csv:"Mana Cost"`
	CMC             float64 `json:"cmc" csv:"CMC"`
	TypeLine        string  `json:"type_line" csv:"Type"`
	OracleText      string  `json:"oracle_text,omitempty" csv:"-"`
	Colors          string  `json:"colors" csv:"Colors"`
	ColorIdentity   string  `json:"color_identity,omitempty" csv:"Color Identity"`
	Rarity          string  `json:"rarity" csv:"Rarity"`
	SetCode         string  `json:"set_code" csv:"Set"`
	CollectorNumber string  `json:"collector_number,omitempty" csv:"Number"`
	Power           string  `json:"power,omitempty" csv:"Power"`
	Toughness       string  `json:"toughness,omitempty" csv:"Toughness"`

	// Draft statistics (nil if unavailable)
	GIHWR            *float64 `json:"gihwr,omitempty" csv:"GIHWR"`
	OHWR             *float64 `json:"ohwr,omitempty" csv:"OHWR"`
	ALSA             *float64 `json:"alsa,omitempty" csv:"ALSA"`
	ATA              *float64 `json:"ata,omitempty" csv:"ATA"`
	SampleSize       *int     `json:"sample_size,omitempty" csv:"Sample Size"`
	GamesPlayed      *int     `json:"games_played,omitempty" csv:"Games Played"`
	NumDecks         *int     `json:"num_decks,omitempty" csv:"Num Decks"`
	StatsFormat      string   `json:"stats_format,omitempty" csv:"Stats Format"`
	StatsLastUpdated string   `json:"stats_last_updated,omitempty" csv:"Stats Updated"`

	// Data freshness
	MetadataAge    string `json:"metadata_age,omitempty" csv:"Metadata Age"`
	StatsAge       string `json:"stats_age,omitempty" csv:"Stats Age"`
	MetadataSource string `json:"metadata_source,omitempty" csv:"Metadata Source"`
	StatsSource    string `json:"stats_source,omitempty" csv:"Stats Source"`
}

// ExportCards exports unified card data to the specified format.
func ExportCards(w io.Writer, cards []*unified.UnifiedCard, opts CardExportOptions) error {
	// Apply filters and sorting
	filtered := filterCards(cards, opts)
	sorted := sortCards(filtered, opts.SortBy)

	// Limit to top N if specified
	if opts.TopN > 0 && len(sorted) > opts.TopN {
		sorted = sorted[:opts.TopN]
	}

	// Convert to export rows
	rows := make([]*CardExportRow, 0, len(sorted))
	for _, card := range sorted {
		rows = append(rows, convertToExportRow(card, opts))
	}

	// Build export data
	exportData := &CardExportData{
		ExportedAt: time.Now(),
		TotalCards: len(rows),
		Cards:      rows,
	}

	// Detect set code from cards
	if len(sorted) > 0 && sorted[0].SetCode != "" {
		exportData.SetCode = sorted[0].SetCode
	}

	// Detect format from draft stats
	if len(sorted) > 0 && sorted[0].HasDraftStats() {
		exportData.Format = sorted[0].DraftStats.Format
	}

	// Add metadata if requested
	if opts.ShowDataAge {
		exportData.Metadata = calculateMetadata(sorted)
	}

	// Export in requested format
	switch opts.Format {
	case FormatCSV:
		return exportCardsCSV(w, rows, opts)
	case FormatJSON:
		return exportCardsJSON(w, exportData, opts)
	case FormatMarkdown, "md":
		return exportCardsMarkdown(w, exportData, opts)
	case FormatArena:
		return exportCardsArena(w, rows, opts)
	default:
		return fmt.Errorf("unsupported export format: %s", opts.Format)
	}
}

// convertToExportRow converts a UnifiedCard to an export row.
func convertToExportRow(card *unified.UnifiedCard, opts CardExportOptions) *CardExportRow {
	row := &CardExportRow{
		ArenaID:         card.ArenaID,
		Name:            card.Name,
		ManaCost:        card.ManaCost,
		CMC:             card.CMC,
		TypeLine:        card.TypeLine,
		OracleText:      card.OracleText,
		Colors:          strings.Join(card.Colors, ""),
		ColorIdentity:   strings.Join(card.ColorIdentity, ""),
		Rarity:          card.Rarity,
		SetCode:         card.SetCode,
		CollectorNumber: card.CollectorNumber,
		Power:           card.Power,
		Toughness:       card.Toughness,
	}

	// Add draft stats if available and requested
	if opts.IncludeStats && card.HasDraftStats() {
		stats := card.DraftStats
		row.GIHWR = &stats.GIHWR
		row.OHWR = &stats.OHWR
		row.ALSA = &stats.ALSA
		row.ATA = &stats.ATA
		sampleSize := stats.GetSampleSize()
		row.SampleSize = &sampleSize
		row.GamesPlayed = &stats.GamesPlayed
		row.NumDecks = &stats.NumDecks
		row.StatsFormat = stats.Format
		if !stats.LastUpdated.IsZero() {
			row.StatsLastUpdated = stats.LastUpdated.Format(time.RFC3339)
		}
	}

	// Add data freshness if requested
	if opts.ShowDataAge {
		row.MetadataAge = formatDuration(card.MetadataAge)
		row.StatsAge = formatDuration(card.StatsAge)
		row.MetadataSource = card.MetadataSource.String()
		row.StatsSource = card.StatsSource.String()
	}

	return row
}

// filterCards applies filters to the card list.
func filterCards(cards []*unified.UnifiedCard, opts CardExportOptions) []*unified.UnifiedCard {
	filtered := make([]*unified.UnifiedCard, 0, len(cards))

	for _, card := range cards {
		// Filter by stats availability
		if opts.OnlyWithStats && !card.HasDraftStats() {
			continue
		}

		// Filter by minimum sample size
		if opts.MinSampleSize > 0 {
			if !card.HasDraftStats() || card.DraftStats.GetSampleSize() < opts.MinSampleSize {
				continue
			}
		}

		// Filter by rarity
		if len(opts.FilterRarity) > 0 {
			if !contains(opts.FilterRarity, strings.ToLower(card.Rarity)) {
				continue
			}
		}

		// Filter by colors
		if len(opts.FilterColors) > 0 {
			if !hasAnyColor(card.Colors, opts.FilterColors) {
				continue
			}
		}

		filtered = append(filtered, card)
	}

	return filtered
}

// sortCards sorts the card list by the specified field.
func sortCards(cards []*unified.UnifiedCard, sortBy string) []*unified.UnifiedCard {
	sorted := make([]*unified.UnifiedCard, len(cards))
	copy(sorted, cards)

	switch strings.ToLower(sortBy) {
	case "gihwr":
		sort.Slice(sorted, func(i, j int) bool {
			if !sorted[i].HasDraftStats() {
				return false
			}
			if !sorted[j].HasDraftStats() {
				return true
			}
			return sorted[i].DraftStats.GIHWR > sorted[j].DraftStats.GIHWR
		})
	case "alsa":
		sort.Slice(sorted, func(i, j int) bool {
			if !sorted[i].HasDraftStats() {
				return false
			}
			if !sorted[j].HasDraftStats() {
				return true
			}
			return sorted[i].DraftStats.ALSA < sorted[j].DraftStats.ALSA
		})
	case "ata":
		sort.Slice(sorted, func(i, j int) bool {
			if !sorted[i].HasDraftStats() {
				return false
			}
			if !sorted[j].HasDraftStats() {
				return true
			}
			return sorted[i].DraftStats.ATA < sorted[j].DraftStats.ATA
		})
	case "cmc":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CMC < sorted[j].CMC
		})
	case "name":
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Name < sorted[j].Name
		})
	default:
		// Default: sort by collector number if available, otherwise by name
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].SetCode == sorted[j].SetCode && sorted[i].CollectorNumber != "" && sorted[j].CollectorNumber != "" {
				return sorted[i].CollectorNumber < sorted[j].CollectorNumber
			}
			return sorted[i].Name < sorted[j].Name
		})
	}

	return sorted
}

// exportCardsCSV exports cards to CSV format.
func exportCardsCSV(w io.Writer, rows []*CardExportRow, opts CardExportOptions) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Build header based on options
	header := []string{
		"Arena ID", "Name", "Mana Cost", "CMC", "Type", "Rarity", "Colors", "Set",
	}

	if opts.IncludeStats {
		header = append(header, "GIHWR", "OHWR", "ALSA", "ATA", "Sample Size", "Games Played")
	}

	if opts.ShowDataAge {
		header = append(header, "Metadata Age", "Stats Age")
	}

	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write rows
	for _, row := range rows {
		record := []string{
			fmt.Sprintf("%d", row.ArenaID),
			row.Name,
			row.ManaCost,
			fmt.Sprintf("%.1f", row.CMC),
			row.TypeLine,
			row.Rarity,
			row.Colors,
			row.SetCode,
		}

		if opts.IncludeStats {
			record = append(record,
				formatFloatPtr(row.GIHWR, "%.1f"),
				formatFloatPtr(row.OHWR, "%.1f"),
				formatFloatPtr(row.ALSA, "%.1f"),
				formatFloatPtr(row.ATA, "%.1f"),
				formatIntPtr(row.SampleSize),
				formatIntPtr(row.GamesPlayed),
			)
		}

		if opts.ShowDataAge {
			record = append(record, row.MetadataAge, row.StatsAge)
		}

		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	return nil
}

// exportCardsJSON exports cards to JSON format.
func exportCardsJSON(w io.Writer, data *CardExportData, opts CardExportOptions) error {
	encoder := json.NewEncoder(w)
	if opts.PrettyJSON {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(data)
}

// exportCardsMarkdown exports cards to Markdown format.
func exportCardsMarkdown(w io.Writer, data *CardExportData, opts CardExportOptions) error {
	var buf bytes.Buffer

	// Title
	if data.SetCode != "" {
		buf.WriteString(fmt.Sprintf("# %s Draft Statistics\n\n", strings.ToUpper(data.SetCode)))
	} else {
		buf.WriteString("# Card Statistics\n\n")
	}

	// Export metadata
	buf.WriteString(fmt.Sprintf("**Exported:** %s\n\n", data.ExportedAt.Format("2006-01-02 15:04:05")))
	buf.WriteString(fmt.Sprintf("**Total Cards:** %d\n\n", data.TotalCards))

	// Data freshness info
	if data.Metadata != nil {
		buf.WriteString("## Data Freshness\n\n")
		if data.Metadata.AverageMetadataAge != "" {
			buf.WriteString(fmt.Sprintf("- **Average Metadata Age:** %s\n", data.Metadata.AverageMetadataAge))
		}
		if data.Metadata.AverageStatsAge != "" {
			buf.WriteString(fmt.Sprintf("- **Average Stats Age:** %s\n", data.Metadata.AverageStatsAge))
		}
		buf.WriteString(fmt.Sprintf("- **Cards With Stats:** %d\n", data.Metadata.CardsWithStats))
		buf.WriteString(fmt.Sprintf("- **Cards Without Stats:** %d\n\n", data.Metadata.CardsWithoutStats))
	}

	// Card table
	if opts.IncludeStats {
		buf.WriteString("## Cards with Draft Statistics\n\n")
		buf.WriteString("| Card | Mana Cost | Type | Rarity | GIHWR | ALSA | ATA | Sample |\n")
		buf.WriteString("|------|-----------|------|--------|-------|------|-----|--------|\n")

		for _, row := range data.Cards {
			if row.GIHWR != nil {
				buf.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %.1f%% | %.1f | %.1f | %s |\n",
					row.Name,
					row.ManaCost,
					truncate(row.TypeLine, 20),
					row.Rarity,
					*row.GIHWR,
					*row.ALSA,
					*row.ATA,
					formatSampleSize(row.SampleSize),
				))
			}
		}
	} else {
		buf.WriteString("## All Cards\n\n")
		buf.WriteString("| Card | Mana Cost | Type | Rarity | Colors |\n")
		buf.WriteString("|------|-----------|------|--------|--------|\n")

		for _, row := range data.Cards {
			buf.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				row.Name,
				row.ManaCost,
				truncate(row.TypeLine, 25),
				row.Rarity,
				row.Colors,
			))
		}
	}

	_, err := w.Write(buf.Bytes())
	return err
}

// exportCardsArena exports cards in MTG Arena deck format.
func exportCardsArena(w io.Writer, rows []*CardExportRow, opts CardExportOptions) error {
	var buf bytes.Buffer

	// Write deck header
	buf.WriteString("Deck\n")

	// Write cards in Arena format: <quantity> <name> (<set>) <number>
	for _, row := range rows {
		quantity := 1 // Default quantity for card list exports

		// Handle missing collector number gracefully
		collectorNum := row.CollectorNumber
		if collectorNum == "" {
			collectorNum = "0" // Default if missing
		}

		// Format: 1 Lightning Bolt (BLB) 123
		buf.WriteString(fmt.Sprintf("%d %s (%s) %s\n",
			quantity,
			row.Name,
			strings.ToUpper(row.SetCode),
			collectorNum,
		))
	}

	_, err := w.Write(buf.Bytes())
	return err
}

// calculateMetadata generates metadata about the exported cards.
func calculateMetadata(cards []*unified.UnifiedCard) *CardExportMetadata {
	meta := &CardExportMetadata{}

	var totalMetadataAge time.Duration
	var totalStatsAge time.Duration
	var metadataCount int
	var statsCount int

	for _, card := range cards {
		if card.HasDraftStats() {
			meta.CardsWithStats++
			totalStatsAge += card.StatsAge
			statsCount++
		} else {
			meta.CardsWithoutStats++
		}

		totalMetadataAge += card.MetadataAge
		metadataCount++
	}

	if metadataCount > 0 {
		avgMetadataAge := totalMetadataAge / time.Duration(metadataCount)
		meta.AverageMetadataAge = formatDuration(avgMetadataAge)
	}

	if statsCount > 0 {
		avgStatsAge := totalStatsAge / time.Duration(statsCount)
		meta.AverageStatsAge = formatDuration(avgStatsAge)
	}

	return meta
}

// Helper functions

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "< 1 minute"
	}
	if d < time.Hour {
		return fmt.Sprintf("%.0f minutes", d.Minutes())
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1f hours", d.Hours())
	}
	return fmt.Sprintf("%.1f days", d.Hours()/24)
}

func formatFloatPtr(f *float64, format string) string {
	if f == nil {
		return "N/A"
	}
	return fmt.Sprintf(format, *f)
}

func formatIntPtr(i *int) string {
	if i == nil {
		return "N/A"
	}
	return fmt.Sprintf("%d", *i)
}

func formatSampleSize(size *int) string {
	if size == nil {
		return "N/A"
	}
	if *size >= 1000 {
		return fmt.Sprintf("%.1fK", float64(*size)/1000)
	}
	return fmt.Sprintf("%d", *size)
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func hasAnyColor(cardColors, filterColors []string) bool {
	if len(cardColors) == 0 {
		return false
	}
	for _, cc := range cardColors {
		for _, fc := range filterColors {
			if strings.EqualFold(cc, fc) {
				return true
			}
		}
	}
	return false
}
