package export

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// ExportCardTrend exports a card's rating trend to the specified format.
func ExportCardTrend(w io.Writer, trend *storage.CardTrend, format string) error {
	switch format {
	case "json":
		return exportCardTrendJSON(w, trend)
	case "csv":
		return exportCardTrendCSV(w, trend)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// exportCardTrendJSON exports trend data as JSON.
func exportCardTrendJSON(w io.Writer, trend *storage.CardTrend) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(trend)
}

// exportCardTrendCSV exports trend data as CSV.
func exportCardTrendCSV(w io.Writer, trend *storage.CardTrend) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"Date",
		"GIHWR",
		"OHWR",
		"ALSA",
		"ATA",
		"SampleSize",
		"GamesPlayed",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write data points
	for _, point := range trend.Points {
		record := []string{
			point.Date.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%.4f", point.GIHWR),
			fmt.Sprintf("%.4f", point.OHWR),
			fmt.Sprintf("%.2f", point.ALSA),
			fmt.Sprintf("%.2f", point.ATA),
			fmt.Sprintf("%d", point.SampleSize),
			fmt.Sprintf("%d", point.GamesPlayed),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// ExportMetaComparison exports a meta comparison to the specified format.
func ExportMetaComparison(w io.Writer, comp *storage.MetaComparison, format string) error {
	switch format {
	case "json":
		return exportMetaComparisonJSON(w, comp)
	case "csv":
		return exportMetaComparisonCSV(w, comp)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// exportMetaComparisonJSON exports meta comparison as JSON.
func exportMetaComparisonJSON(w io.Writer, comp *storage.MetaComparison) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(comp)
}

// exportMetaComparisonCSV exports meta comparison as CSV.
func exportMetaComparisonCSV(w io.Writer, comp *storage.MetaComparison) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"ArenaID",
		"CardName",
		"EarlyGIHWR",
		"LateGIHWR",
		"GIHWRChange",
		"ChangeType",
		"EarlySampleSize",
		"LateSampleSize",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write top improvers
	for _, improvement := range comp.TopImprovers {
		record := []string{
			fmt.Sprintf("%d", improvement.ArenaID),
			improvement.CardName,
			fmt.Sprintf("%.4f", improvement.EarlyGIHWR),
			fmt.Sprintf("%.4f", improvement.LateGIHWR),
			fmt.Sprintf("%.4f", improvement.GIHWRChange),
			"Improver",
			fmt.Sprintf("%d", improvement.EarlySampleSize),
			fmt.Sprintf("%d", improvement.LateSampleSize),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	// Write top decliners
	for _, decline := range comp.TopDecliners {
		record := []string{
			fmt.Sprintf("%d", decline.ArenaID),
			decline.CardName,
			fmt.Sprintf("%.4f", decline.EarlyGIHWR),
			fmt.Sprintf("%.4f", decline.LateGIHWR),
			fmt.Sprintf("%.4f", decline.GIHWRChange),
			"Decliner",
			fmt.Sprintf("%d", decline.EarlySampleSize),
			fmt.Sprintf("%d", decline.LateSampleSize),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// ExportRatingHistory exports complete rating history to the specified format.
func ExportRatingHistory(w io.Writer, history []*storage.DraftCardRating, format string) error {
	switch format {
	case "json":
		return exportRatingHistoryJSON(w, history)
	case "csv":
		return exportRatingHistoryCSV(w, history)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// exportRatingHistoryJSON exports rating history as JSON.
func exportRatingHistoryJSON(w io.Writer, history []*storage.DraftCardRating) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(history)
}

// exportRatingHistoryCSV exports rating history as CSV.
func exportRatingHistoryCSV(w io.Writer, history []*storage.DraftCardRating) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"CachedAt",
		"GIHWR",
		"OHWR",
		"ALSA",
		"ATA",
		"GIH",
		"GamesPlayed",
		"NumDecks",
		"StartDate",
		"EndDate",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Write history records
	for _, rating := range history {
		record := []string{
			rating.CachedAt.Format("2006-01-02 15:04:05"),
			fmt.Sprintf("%.4f", rating.GIHWR),
			fmt.Sprintf("%.4f", rating.OHWR),
			fmt.Sprintf("%.2f", rating.ALSA),
			fmt.Sprintf("%.2f", rating.ATA),
			fmt.Sprintf("%d", rating.GIH),
			fmt.Sprintf("%d", rating.GamesPlayed),
			fmt.Sprintf("%d", rating.NumDecks),
			rating.StartDate,
			rating.EndDate,
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// ExportCleanupResult exports cleanup result to the specified format.
func ExportCleanupResult(w io.Writer, result *storage.CleanupResult, format string) error {
	switch format {
	case "json":
		return exportCleanupResultJSON(w, result)
	case "csv":
		return exportCleanupResultCSV(w, result)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// exportCleanupResultJSON exports cleanup result as JSON.
func exportCleanupResultJSON(w io.Writer, result *storage.CleanupResult) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

// exportCleanupResultCSV exports cleanup result as CSV.
func exportCleanupResultCSV(w io.Writer, result *storage.CleanupResult) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write summary
	if err := writer.Write([]string{"Summary", "Value"}); err != nil {
		return err
	}

	rows := [][]string{
		{"Total Snapshots", fmt.Sprintf("%d", result.TotalSnapshots)},
		{"Removed Snapshots", fmt.Sprintf("%d", result.RemovedSnapshots)},
		{"Retained Snapshots", fmt.Sprintf("%d", result.RetainedSnapshots)},
		{"Oldest Snapshot", formatTime(result.OldestSnapshot)},
		{"Newest Snapshot", formatTime(result.NewestSnapshot)},
		{"Dry Run", fmt.Sprintf("%t", result.DryRun)},
	}

	for _, row := range rows {
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	// Write per-set breakdown
	if err := writer.Write([]string{""}); err != nil { // Empty row
		return err
	}
	if err := writer.Write([]string{"Expansion", "Removed", "Retained"}); err != nil {
		return err
	}

	// Combine removed and retained maps
	sets := make(map[string]struct{})
	for set := range result.RemovedBySet {
		sets[set] = struct{}{}
	}
	for set := range result.RetainedBySet {
		sets[set] = struct{}{}
	}

	for set := range sets {
		removed := result.RemovedBySet[set]
		retained := result.RetainedBySet[set]
		if err := writer.Write([]string{
			set,
			fmt.Sprintf("%d", removed),
			fmt.Sprintf("%d", retained),
		}); err != nil {
			return err
		}
	}

	return nil
}

// formatTime formats a time.Time for CSV output.
func formatTime(t time.Time) string {
	if t.IsZero() {
		return "N/A"
	}
	return t.Format("2006-01-02 15:04:05")
}
