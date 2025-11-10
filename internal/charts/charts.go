package charts

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

// ChartConfig holds configuration for charts.
type ChartConfig struct {
	Title      string   // Chart title
	Subtitle   string   // Chart subtitle
	YAxisLabel string   // Y-axis label
	XAxisLabel string   // X-axis label
	Width      string   // Chart width (e.g., "900px")
	Height     string   // Chart height (e.g., "500px")
	Theme      string   // Chart theme
	ShowLegend bool     // Show legend
	Smooth     bool     // Smooth line (for line charts)
	Colors     []string // Custom colors
}

// DefaultChartConfig returns default chart configuration.
func DefaultChartConfig() ChartConfig {
	return ChartConfig{
		Title:      "",
		Subtitle:   "",
		YAxisLabel: "",
		XAxisLabel: "",
		Width:      "900px",
		Height:     "500px",
		Theme:      "light",
		ShowLegend: true,
		Smooth:     true,
		Colors:     []string{"#5470C6", "#91CC75", "#FAC858", "#EE6666", "#73C0DE", "#3BA272", "#FC8452", "#9A60B4", "#EA7CCC"},
	}
}

// DataPoint represents a single data point in a chart.
type DataPoint struct {
	Label string
	Value float64
}

// SeriesData represents a data series for multi-series charts.
type SeriesData struct {
	Name   string
	Points []DataPoint
}

// RenderLineChart creates an interactive line chart HTML file.
func RenderLineChart(data []DataPoint, config ChartConfig, outputPath string) error {
	line := charts.NewLine()

	// Set global options
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  config.Width,
			Height: config.Height,
			Theme:  config.Theme,
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    config.Title,
			Subtitle: config.Subtitle,
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    opts.Bool(true),
			Trigger: "axis",
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: opts.Bool(config.ShowLegend),
		}),
		charts.WithColorsOpts(opts.Colors{
			config.Colors[0],
		}),
	)

	// Prepare X-axis labels
	xLabels := make([]string, len(data))
	for i, point := range data {
		xLabels[i] = point.Label
	}

	// Prepare Y-axis data
	yData := make([]opts.LineData, len(data))
	for i, point := range data {
		yData[i] = opts.LineData{Value: point.Value}
	}

	// Add data to chart
	line.SetXAxis(xLabels).
		AddSeries("Win Rate", yData).
		SetSeriesOptions(
			charts.WithLineChartOpts(opts.LineChart{
				Smooth: opts.Bool(config.Smooth),
			}),
			charts.WithLabelOpts(opts.Label{
				Show: opts.Bool(false),
			}),
		)

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create chart file: %w", err)
	}
	defer f.Close()

	if err := line.Render(f); err != nil {
		return fmt.Errorf("failed to render chart: %w", err)
	}

	return nil
}

// RenderBarChart creates an interactive bar chart HTML file.
func RenderBarChart(data []DataPoint, config ChartConfig, outputPath string) error {
	bar := charts.NewBar()

	// Set global options
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  config.Width,
			Height: config.Height,
			Theme:  config.Theme,
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    config.Title,
			Subtitle: config.Subtitle,
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    opts.Bool(true),
			Trigger: "axis",
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: opts.Bool(config.ShowLegend),
		}),
		charts.WithColorsOpts(opts.Colors{
			config.Colors[0],
		}),
	)

	// Prepare X-axis labels
	xLabels := make([]string, len(data))
	for i, point := range data {
		xLabels[i] = point.Label
	}

	// Prepare Y-axis data
	yData := make([]opts.BarData, len(data))
	for i, point := range data {
		yData[i] = opts.BarData{Value: point.Value}
	}

	// Add data to chart
	bar.SetXAxis(xLabels).
		AddSeries("Win Rate", yData).
		SetSeriesOptions(
			charts.WithLabelOpts(opts.Label{
				Show: opts.Bool(false),
			}),
		)

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create chart file: %w", err)
	}
	defer f.Close()

	if err := bar.Render(f); err != nil {
		return fmt.Errorf("failed to render chart: %w", err)
	}

	return nil
}

// RenderMultiLineChart creates a multi-series line chart HTML file.
func RenderMultiLineChart(series []SeriesData, config ChartConfig, outputPath string) error {
	line := charts.NewLine()

	// Set global options
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			Width:  config.Width,
			Height: config.Height,
			Theme:  config.Theme,
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    config.Title,
			Subtitle: config.Subtitle,
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show:    opts.Bool(true),
			Trigger: "axis",
		}),
		charts.WithLegendOpts(opts.Legend{
			Show: opts.Bool(config.ShowLegend),
		}),
	)

	// Use first series for X-axis labels
	if len(series) == 0 {
		return fmt.Errorf("no data series provided")
	}

	xLabels := make([]string, len(series[0].Points))
	for i, point := range series[0].Points {
		xLabels[i] = point.Label
	}

	line.SetXAxis(xLabels)

	// Add each series
	for i, s := range series {
		yData := make([]opts.LineData, len(s.Points))
		for j, point := range s.Points {
			yData[j] = opts.LineData{Value: point.Value}
		}

		color := config.Colors[i%len(config.Colors)]
		line.AddSeries(s.Name, yData).
			SetSeriesOptions(
				charts.WithLineChartOpts(opts.LineChart{
					Smooth: opts.Bool(config.Smooth),
				}),
				charts.WithLabelOpts(opts.Label{
					Show: opts.Bool(false),
				}),
				charts.WithItemStyleOpts(opts.ItemStyle{
					Color: color,
				}),
			)
	}

	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create chart file: %w", err)
	}
	defer f.Close()

	if err := line.Render(f); err != nil {
		return fmt.Errorf("failed to render chart: %w", err)
	}

	return nil
}

// OpenInBrowser opens the given file path in the default web browser.
func OpenInBrowser(filePath string) error {
	// Get absolute path
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", absPath)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", absPath)
	case "linux":
		cmd = exec.Command("xdg-open", absPath)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return cmd.Start()
}
