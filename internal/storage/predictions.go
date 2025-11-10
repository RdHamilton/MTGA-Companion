package storage

import (
	"context"
	"fmt"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// WinRatePrediction represents a win rate prediction with confidence.
type WinRatePrediction struct {
	CurrentWinRate    float64 `json:"current_win_rate"`   // Current win rate percentage
	PredictedWinRate  float64 `json:"predicted_win_rate"` // Predicted future win rate
	Trend             string  `json:"trend"`              // "improving", "declining", "stable"
	TrendStrength     float64 `json:"trend_strength"`     // Absolute change per match
	Confidence        float64 `json:"confidence"`         // 0-100, based on sample size
	SampleSize        int     `json:"sample_size"`        // Number of recent matches analyzed
	RecentMatches     int     `json:"recent_matches"`     // Matches used for trend analysis
	ProjectionMatches int     `json:"projection_matches"` // Matches ahead being projected
}

// FormatPrediction represents predictions for a specific format.
type FormatPrediction struct {
	Format            string            `json:"format"`
	Prediction        WinRatePrediction `json:"prediction"`
	RecentPerformance string            `json:"recent_performance"` // "strong", "weak", "average"
}

// PredictionSummary provides overall prediction insights.
type PredictionSummary struct {
	Overall         WinRatePrediction  `json:"overall"`
	ByFormat        []FormatPrediction `json:"by_format"`
	StrongestFormat string             `json:"strongest_format"` // Format with best predicted win rate
	WeakestFormat   string             `json:"weakest_format"`   // Format with worst predicted win rate
	MostImproving   string             `json:"most_improving"`   // Format with strongest upward trend
	MostDeclining   string             `json:"most_declining"`   // Format with strongest downward trend
}

// PredictWinRate calculates win rate prediction based on recent performance.
// recentMatchCount determines how many recent matches to analyze (default: 30).
// projectionMatches determines how many matches ahead to project (default: 10).
func (s *Service) PredictWinRate(ctx context.Context, filter models.StatsFilter, recentMatchCount, projectionMatches int) (*WinRatePrediction, error) {
	// Set defaults
	if recentMatchCount <= 0 {
		recentMatchCount = 30
	}
	if projectionMatches <= 0 {
		projectionMatches = 10
	}

	// Get all matches
	matches, err := s.GetMatches(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get matches: %w", err)
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found for prediction")
	}

	// Calculate current overall win rate
	totalMatches := len(matches)
	totalWins := 0
	for _, match := range matches {
		if match.Result == "win" {
			totalWins++
		}
	}
	currentWinRate := float64(totalWins) / float64(totalMatches) * 100

	// Use recent matches for trend analysis (limit to available matches)
	trendSampleSize := recentMatchCount
	if trendSampleSize > len(matches) {
		trendSampleSize = len(matches)
	}

	// Calculate moving window win rates to detect trend
	recentMatches := matches[:trendSampleSize]             // Already sorted DESC by timestamp
	winRates := calculateMovingWinRates(recentMatches, 10) // 10-match windows

	// Calculate trend using linear regression on recent win rates
	trend, trendStrength := calculateTrend(winRates)

	// Project future win rate
	predictedWinRate := currentWinRate + (trendStrength * float64(projectionMatches))

	// Clamp to 0-100 range
	if predictedWinRate < 0 {
		predictedWinRate = 0
	} else if predictedWinRate > 100 {
		predictedWinRate = 100
	}

	// Calculate confidence based on sample size and variance
	confidence := calculateConfidence(trendSampleSize, winRates)

	prediction := &WinRatePrediction{
		CurrentWinRate:    currentWinRate,
		PredictedWinRate:  predictedWinRate,
		Trend:             trend,
		TrendStrength:     trendStrength,
		Confidence:        confidence,
		SampleSize:        totalMatches,
		RecentMatches:     trendSampleSize,
		ProjectionMatches: projectionMatches,
	}

	return prediction, nil
}

// PredictByFormat calculates predictions for each format separately.
func (s *Service) PredictByFormat(ctx context.Context, filter models.StatsFilter, recentMatchCount, projectionMatches int) (*PredictionSummary, error) {
	// Get overall prediction
	overallPrediction, err := s.PredictWinRate(ctx, filter, recentMatchCount, projectionMatches)
	if err != nil {
		return nil, fmt.Errorf("failed to get overall prediction: %w", err)
	}

	// Get predictions for each format
	statsByFormat, err := s.GetStatsByFormat(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get format stats: %w", err)
	}

	var formatPredictions []FormatPrediction
	var bestFormat, worstFormat string
	var bestWinRate, worstWinRate float64 = -1, 101
	var mostImproving, mostDeclining string
	var strongestUptrend, strongestDowntrend float64 = -1000, 1000

	for format := range statsByFormat {
		// Create format-specific filter
		formatFilter := filter
		formatStr := format
		formatFilter.Format = &formatStr

		pred, err := s.PredictWinRate(ctx, formatFilter, recentMatchCount, projectionMatches)
		if err != nil {
			continue // Skip formats with insufficient data
		}

		// Determine performance category
		performance := "average"
		if pred.CurrentWinRate >= 55 {
			performance = "strong"
		} else if pred.CurrentWinRate <= 45 {
			performance = "weak"
		}

		formatPred := FormatPrediction{
			Format:            format,
			Prediction:        *pred,
			RecentPerformance: performance,
		}
		formatPredictions = append(formatPredictions, formatPred)

		// Track best/worst predicted formats
		if pred.PredictedWinRate > bestWinRate {
			bestWinRate = pred.PredictedWinRate
			bestFormat = format
		}
		if pred.PredictedWinRate < worstWinRate {
			worstWinRate = pred.PredictedWinRate
			worstFormat = format
		}

		// Track strongest trends
		if pred.TrendStrength > strongestUptrend {
			strongestUptrend = pred.TrendStrength
			mostImproving = format
		}
		if pred.TrendStrength < strongestDowntrend {
			strongestDowntrend = pred.TrendStrength
			mostDeclining = format
		}
	}

	summary := &PredictionSummary{
		Overall:         *overallPrediction,
		ByFormat:        formatPredictions,
		StrongestFormat: bestFormat,
		WeakestFormat:   worstFormat,
		MostImproving:   mostImproving,
		MostDeclining:   mostDeclining,
	}

	return summary, nil
}

// calculateMovingWinRates calculates win rates for moving windows.
func calculateMovingWinRates(matches []*models.Match, windowSize int) []float64 {
	if len(matches) < windowSize {
		// Not enough data for windows, return overall win rate
		wins := 0
		for _, m := range matches {
			if m.Result == "win" {
				wins++
			}
		}
		if len(matches) > 0 {
			return []float64{float64(wins) / float64(len(matches)) * 100}
		}
		return []float64{}
	}

	var winRates []float64
	for i := 0; i <= len(matches)-windowSize; i++ {
		window := matches[i : i+windowSize]
		wins := 0
		for _, m := range window {
			if m.Result == "win" {
				wins++
			}
		}
		winRate := float64(wins) / float64(windowSize) * 100
		winRates = append(winRates, winRate)
	}

	return winRates
}

// calculateTrend performs simple linear regression on win rates to detect trend.
func calculateTrend(winRates []float64) (string, float64) {
	if len(winRates) < 2 {
		return "stable", 0.0
	}

	// Simple linear regression: y = mx + b
	n := float64(len(winRates))
	var sumX, sumY, sumXY, sumX2 float64

	for i, y := range winRates {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope (m)
	slope := (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Determine trend category
	var trend string
	switch {
	case slope > 0.5:
		trend = "improving"
	case slope < -0.5:
		trend = "declining"
	default:
		trend = "stable"
	}

	return trend, slope
}

// calculateConfidence returns confidence score (0-100) based on sample size and variance.
func calculateConfidence(sampleSize int, winRates []float64) float64 {
	// Base confidence on sample size
	var sizeConfidence float64
	switch {
	case sampleSize >= 50:
		sizeConfidence = 100
	case sampleSize >= 30:
		sizeConfidence = 80
	case sampleSize >= 20:
		sizeConfidence = 60
	case sampleSize >= 10:
		sizeConfidence = 40
	default:
		sizeConfidence = 20
	}

	// Adjust for variance (lower variance = higher confidence)
	if len(winRates) >= 2 {
		variance := calculateVariance(winRates)
		// High variance (>400) reduces confidence, low variance (<100) maintains it
		varianceAdjustment := 0.0
		if variance > 400 {
			varianceAdjustment = -20
		} else if variance > 200 {
			varianceAdjustment = -10
		}
		sizeConfidence += varianceAdjustment
	}

	// Clamp to 0-100
	if sizeConfidence < 0 {
		sizeConfidence = 0
	} else if sizeConfidence > 100 {
		sizeConfidence = 100
	}

	return sizeConfidence
}

// calculateVariance calculates variance of a dataset.
func calculateVariance(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	// Calculate mean
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))

	// Calculate variance
	var squaredDiff float64
	for _, v := range values {
		diff := v - mean
		squaredDiff += diff * diff
	}

	return squaredDiff / float64(len(values))
}
