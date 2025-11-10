package export

import (
	"context"
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// WinRatePredictionExportRow represents a win rate prediction for CSV export.
type WinRatePredictionExportRow struct {
	CurrentWinRate    float64 `csv:"current_win_rate" json:"current_win_rate"`
	PredictedWinRate  float64 `csv:"predicted_win_rate" json:"predicted_win_rate"`
	Trend             string  `csv:"trend" json:"trend"`
	TrendStrength     float64 `csv:"trend_strength" json:"trend_strength"`
	Confidence        float64 `csv:"confidence" json:"confidence"`
	SampleSize        int     `csv:"sample_size" json:"sample_size"`
	RecentMatches     int     `csv:"recent_matches" json:"recent_matches"`
	ProjectionMatches int     `csv:"projection_matches" json:"projection_matches"`
}

// FormatPredictionExportRow represents format-specific predictions for CSV export.
type FormatPredictionExportRow struct {
	Format            string  `csv:"format" json:"format"`
	CurrentWinRate    float64 `csv:"current_win_rate" json:"current_win_rate"`
	PredictedWinRate  float64 `csv:"predicted_win_rate" json:"predicted_win_rate"`
	Trend             string  `csv:"trend" json:"trend"`
	TrendStrength     float64 `csv:"trend_strength" json:"trend_strength"`
	Confidence        float64 `csv:"confidence" json:"confidence"`
	RecentPerformance string  `csv:"recent_performance" json:"recent_performance"`
	SampleSize        int     `csv:"sample_size" json:"sample_size"`
}

// PredictionSummaryExportRow represents prediction summary insights for CSV export.
type PredictionSummaryExportRow struct {
	OverallCurrentWR   float64 `csv:"overall_current_win_rate" json:"overall_current_win_rate"`
	OverallPredictedWR float64 `csv:"overall_predicted_win_rate" json:"overall_predicted_win_rate"`
	OverallTrend       string  `csv:"overall_trend" json:"overall_trend"`
	OverallConfidence  float64 `csv:"overall_confidence" json:"overall_confidence"`
	StrongestFormat    string  `csv:"strongest_format" json:"strongest_format"`
	WeakestFormat      string  `csv:"weakest_format" json:"weakest_format"`
	MostImproving      string  `csv:"most_improving" json:"most_improving"`
	MostDeclining      string  `csv:"most_declining" json:"most_declining"`
}

// ExportWinRatePrediction exports overall win rate prediction.
func ExportWinRatePrediction(ctx context.Context, service *storage.Service, filter models.StatsFilter, recentMatches, projectionMatches int, opts Options) error {
	prediction, err := service.PredictWinRate(ctx, filter, recentMatches, projectionMatches)
	if err != nil {
		return fmt.Errorf("failed to get win rate prediction: %w", err)
	}

	if prediction == nil {
		return fmt.Errorf("no prediction data available")
	}

	// Convert to export row
	row := WinRatePredictionExportRow{
		CurrentWinRate:    prediction.CurrentWinRate,
		PredictedWinRate:  prediction.PredictedWinRate,
		Trend:             prediction.Trend,
		TrendStrength:     prediction.TrendStrength,
		Confidence:        prediction.Confidence,
		SampleSize:        prediction.SampleSize,
		RecentMatches:     prediction.RecentMatches,
		ProjectionMatches: prediction.ProjectionMatches,
	}

	switch opts.Format {
	case FormatCSV:
		exporter := NewExporter(opts)
		return exporter.Export([]WinRatePredictionExportRow{row})
	case FormatJSON:
		exporter := NewExporter(opts)
		return exporter.Export(prediction)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// ExportPredictionsByFormat exports predictions for all formats.
func ExportPredictionsByFormat(ctx context.Context, service *storage.Service, filter models.StatsFilter, recentMatches, projectionMatches int, opts Options) error {
	summary, err := service.PredictByFormat(ctx, filter, recentMatches, projectionMatches)
	if err != nil {
		return fmt.Errorf("failed to get format predictions: %w", err)
	}

	if summary == nil {
		return fmt.Errorf("no prediction data available")
	}

	switch opts.Format {
	case FormatCSV:
		// Export format predictions as rows
		rows := make([]FormatPredictionExportRow, len(summary.ByFormat))
		for i, fp := range summary.ByFormat {
			rows[i] = FormatPredictionExportRow{
				Format:            fp.Format,
				CurrentWinRate:    fp.Prediction.CurrentWinRate,
				PredictedWinRate:  fp.Prediction.PredictedWinRate,
				Trend:             fp.Prediction.Trend,
				TrendStrength:     fp.Prediction.TrendStrength,
				Confidence:        fp.Prediction.Confidence,
				RecentPerformance: fp.RecentPerformance,
				SampleSize:        fp.Prediction.SampleSize,
			}
		}
		exporter := NewExporter(opts)
		return exporter.Export(rows)
	case FormatJSON:
		// Export full summary in JSON
		exporter := NewExporter(opts)
		return exporter.Export(summary)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// ExportPredictionSummary exports prediction summary insights.
func ExportPredictionSummary(ctx context.Context, service *storage.Service, filter models.StatsFilter, recentMatches, projectionMatches int, opts Options) error {
	summary, err := service.PredictByFormat(ctx, filter, recentMatches, projectionMatches)
	if err != nil {
		return fmt.Errorf("failed to get prediction summary: %w", err)
	}

	if summary == nil {
		return fmt.Errorf("no prediction summary available")
	}

	// Convert to export row
	row := PredictionSummaryExportRow{
		OverallCurrentWR:   summary.Overall.CurrentWinRate,
		OverallPredictedWR: summary.Overall.PredictedWinRate,
		OverallTrend:       summary.Overall.Trend,
		OverallConfidence:  summary.Overall.Confidence,
		StrongestFormat:    summary.StrongestFormat,
		WeakestFormat:      summary.WeakestFormat,
		MostImproving:      summary.MostImproving,
		MostDeclining:      summary.MostDeclining,
	}

	switch opts.Format {
	case FormatCSV:
		exporter := NewExporter(opts)
		return exporter.Export([]PredictionSummaryExportRow{row})
	case FormatJSON:
		exporter := NewExporter(opts)
		return exporter.Export(summary)
	default:
		return fmt.Errorf("unsupported format: %s", opts.Format)
	}
}
