package charts

import (
	"fmt"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// FyneChartConfig holds configuration for Fyne charts.
type FyneChartConfig struct {
	Title      string
	Width      float32
	Height     float32
	ShowGrid   bool
	GridColor  color.Color
	LineColor  color.Color
	PointColor color.Color
	BarColor   color.Color
}

// DefaultFyneChartConfig returns default Fyne chart configuration.
func DefaultFyneChartConfig() FyneChartConfig {
	return FyneChartConfig{
		Title:      "",
		Width:      800,
		Height:     400,
		ShowGrid:   true,
		GridColor:  color.RGBA{R: 200, G: 200, B: 200, A: 255},
		LineColor:  color.RGBA{R: 84, G: 112, B: 198, A: 255},
		PointColor: color.RGBA{R: 84, G: 112, B: 198, A: 255},
		BarColor:   color.RGBA{R: 84, G: 112, B: 198, A: 255},
	}
}

// CreateFyneLineChart creates a Fyne line chart widget.
func CreateFyneLineChart(data []DataPoint, config FyneChartConfig) fyne.CanvasObject {
	if len(data) == 0 {
		return widget.NewLabel("No data available")
	}

	// Calculate data bounds
	minVal, maxVal := findMinMaxValues(data)
	valueRange := maxVal - minVal
	if valueRange == 0 {
		valueRange = 1
	}

	// Add padding
	padding := valueRange * 0.1
	minVal -= padding
	maxVal += padding
	valueRange = maxVal - minVal

	// Chart dimensions
	chartWidth := config.Width
	chartHeight := config.Height
	leftMargin := float32(60)
	rightMargin := float32(40)
	topMargin := float32(40)
	bottomMargin := float32(60)

	plotWidth := chartWidth - leftMargin - rightMargin
	plotHeight := chartHeight - topMargin - bottomMargin

	// Container for all chart elements
	objects := []fyne.CanvasObject{}

	// Text color for all labels
	textColor := color.RGBA{R: 66, G: 66, B: 66, A: 255}

	// Draw grid lines if enabled
	if config.ShowGrid {
		// Horizontal grid lines
		for i := 0; i <= 5; i++ {
			y := topMargin + (plotHeight / 5 * float32(i))
			line := canvas.NewLine(config.GridColor)
			line.Position1 = fyne.NewPos(leftMargin, y)
			line.Position2 = fyne.NewPos(leftMargin+plotWidth, y)
			line.StrokeWidth = 1
			objects = append(objects, line)

			// Y-axis label
			value := maxVal - (valueRange / 5 * float64(i))
			label := canvas.NewText(fmt.Sprintf("%.1f", value), textColor)
			label.TextSize = 10
			label.Move(fyne.NewPos(5, y-7))
			objects = append(objects, label)
		}

		// Vertical grid lines
		gridStep := int(math.Ceil(float64(len(data)) / 10.0))
		if gridStep < 1 {
			gridStep = 1
		}
		for i := 0; i < len(data); i += gridStep {
			x := leftMargin + (plotWidth / float32(len(data)-1) * float32(i))
			line := canvas.NewLine(config.GridColor)
			line.Position1 = fyne.NewPos(x, topMargin)
			line.Position2 = fyne.NewPos(x, topMargin+plotHeight)
			line.StrokeWidth = 1
			objects = append(objects, line)
		}
	}

	// Draw chart border
	border := canvas.NewRectangle(color.Transparent)
	border.StrokeColor = color.Black
	border.StrokeWidth = 2
	border.Resize(fyne.NewSize(plotWidth, plotHeight))
	border.Move(fyne.NewPos(leftMargin, topMargin))
	objects = append(objects, border)

	// Normalize data and draw lines/points
	points := make([]fyne.Position, len(data))
	for i, point := range data {
		x := leftMargin + (plotWidth / float32(len(data)-1) * float32(i))
		normalizedValue := (point.Value - minVal) / valueRange
		y := topMargin + plotHeight - (plotHeight * float32(normalizedValue))
		points[i] = fyne.NewPos(x, y)

		// Draw point
		circle := canvas.NewCircle(config.PointColor)
		circle.Resize(fyne.NewSize(6, 6))
		circle.Move(fyne.NewPos(x-3, y-3))
		objects = append(objects, circle)

		// X-axis labels (show subset)
		labelStep := int(math.Ceil(float64(len(data)) / 10.0))
		if labelStep < 1 {
			labelStep = 1
		}
		if i%labelStep == 0 || i == len(data)-1 {
			label := canvas.NewText(point.Label, textColor)
			label.TextSize = 9
			label.Alignment = fyne.TextAlignCenter
			label.Move(fyne.NewPos(x-30, topMargin+plotHeight+10))
			objects = append(objects, label)
		}
	}

	// Draw connecting lines
	for i := 0; i < len(points)-1; i++ {
		line := canvas.NewLine(config.LineColor)
		line.Position1 = points[i]
		line.Position2 = points[i+1]
		line.StrokeWidth = 2
		objects = append(objects, line)
	}

	// Add title with better contrast
	if config.Title != "" {
		titleColor := color.RGBA{R: 66, G: 66, B: 66, A: 255}
		title := canvas.NewText(config.Title, titleColor)
		title.TextSize = 16
		title.Alignment = fyne.TextAlignCenter
		title.Move(fyne.NewPos(chartWidth/2-100, 10))
		objects = append(objects, title)
	}

	// Y-axis label with better visibility
	yAxisLabel := canvas.NewText("Win Rate (%)", textColor)
	yAxisLabel.TextSize = 12
	yAxisLabel.Move(fyne.NewPos(10, chartHeight/2))
	objects = append(objects, yAxisLabel)

	// Create container with fixed size
	chart := container.NewWithoutLayout(objects...)
	chart.Resize(fyne.NewSize(chartWidth, chartHeight))

	return chart
}

// CreateFyneBarChart creates a Fyne bar chart widget.
func CreateFyneBarChart(data []DataPoint, config FyneChartConfig) fyne.CanvasObject {
	if len(data) == 0 {
		return widget.NewLabel("No data available")
	}

	// Calculate data bounds
	minVal, maxVal := findMinMaxValues(data)
	valueRange := maxVal - minVal
	if valueRange == 0 {
		valueRange = 1
	}

	// Add padding
	padding := valueRange * 0.1
	minVal -= padding
	maxVal += padding
	valueRange = maxVal - minVal

	// Chart dimensions
	chartWidth := config.Width
	chartHeight := config.Height
	leftMargin := float32(60)
	rightMargin := float32(40)
	topMargin := float32(40)
	bottomMargin := float32(80)

	plotWidth := chartWidth - leftMargin - rightMargin
	plotHeight := chartHeight - topMargin - bottomMargin

	// Container for all chart elements
	objects := []fyne.CanvasObject{}

	// Text color for all labels
	textColor := color.RGBA{R: 66, G: 66, B: 66, A: 255}

	// Draw grid lines if enabled
	if config.ShowGrid {
		// Horizontal grid lines
		for i := 0; i <= 5; i++ {
			y := topMargin + (plotHeight / 5 * float32(i))
			line := canvas.NewLine(config.GridColor)
			line.Position1 = fyne.NewPos(leftMargin, y)
			line.Position2 = fyne.NewPos(leftMargin+plotWidth, y)
			line.StrokeWidth = 1
			objects = append(objects, line)

			// Y-axis label
			value := maxVal - (valueRange / 5 * float64(i))
			label := canvas.NewText(fmt.Sprintf("%.1f", value), textColor)
			label.TextSize = 10
			label.Move(fyne.NewPos(5, y-7))
			objects = append(objects, label)
		}
	}

	// Draw chart border
	border := canvas.NewRectangle(color.Transparent)
	border.StrokeColor = color.Black
	border.StrokeWidth = 2
	border.Resize(fyne.NewSize(plotWidth, plotHeight))
	border.Move(fyne.NewPos(leftMargin, topMargin))
	objects = append(objects, border)

	// Calculate bar width
	barWidth := plotWidth / float32(len(data)) * 0.8
	barSpacing := plotWidth / float32(len(data))

	// Draw bars
	for i, point := range data {
		normalizedValue := (point.Value - minVal) / valueRange
		barHeight := plotHeight * float32(normalizedValue)

		x := leftMargin + (barSpacing * float32(i)) + (barSpacing-barWidth)/2
		y := topMargin + plotHeight - barHeight

		// Draw bar
		bar := canvas.NewRectangle(config.BarColor)
		bar.Resize(fyne.NewSize(barWidth, barHeight))
		bar.Move(fyne.NewPos(x, y))
		objects = append(objects, bar)

		// X-axis label
		label := canvas.NewText(point.Label, textColor)
		label.TextSize = 8
		label.Alignment = fyne.TextAlignCenter
		label.Move(fyne.NewPos(x-10, topMargin+plotHeight+10))
		label.Resize(fyne.NewSize(barWidth+20, 20))
		objects = append(objects, label)

		// Value label on top of bar
		valueLabel := canvas.NewText(fmt.Sprintf("%.1f", point.Value), textColor)
		valueLabel.TextSize = 9
		valueLabel.Alignment = fyne.TextAlignCenter
		valueLabel.Move(fyne.NewPos(x, y-15))
		objects = append(objects, valueLabel)
	}

	// Add title with better contrast
	if config.Title != "" {
		titleColor := color.RGBA{R: 66, G: 66, B: 66, A: 255}
		title := canvas.NewText(config.Title, titleColor)
		title.TextSize = 16
		title.Alignment = fyne.TextAlignCenter
		title.Move(fyne.NewPos(chartWidth/2-100, 10))
		objects = append(objects, title)
	}

	// Y-axis label with better visibility
	yAxisLabel := canvas.NewText("Win Rate (%)", textColor)
	yAxisLabel.TextSize = 12
	yAxisLabel.Move(fyne.NewPos(10, chartHeight/2))
	objects = append(objects, yAxisLabel)

	// Create container with fixed size
	chart := container.NewWithoutLayout(objects...)
	chart.Resize(fyne.NewSize(chartWidth, chartHeight))

	return chart
}

// CreateFynePieChartBreakdown creates a horizontal bar breakdown view (pie chart alternative).
// This provides a simpler visualization for categorical data compared to drawing actual pie slices.
func CreateFynePieChartBreakdown(data []DataPoint, config FyneChartConfig) fyne.CanvasObject {
	if len(data) == 0 {
		return widget.NewLabel("No data available")
	}

	// Calculate total
	total := 0.0
	for _, point := range data {
		total += point.Value
	}

	if total == 0 {
		return widget.NewLabel("No data to display")
	}

	// Chart dimensions
	chartWidth := config.Width
	chartHeight := config.Height
	leftMargin := float32(100)  // Space for labels
	rightMargin := float32(150) // Space for percentages
	topMargin := float32(100)   // Space for title

	plotWidth := chartWidth - leftMargin - rightMargin

	// Calculate bar height with spacing
	barHeight := float32(35)  // Fixed bar height for consistency
	barSpacing := float32(20) // Spacing between bars

	// Container for all chart elements
	objects := []fyne.CanvasObject{}

	// Add title centered above the chart with proper spacing
	if config.Title != "" {
		titleColor := color.RGBA{R: 66, G: 66, B: 66, A: 255}
		title := canvas.NewText(config.Title, titleColor)
		title.TextSize = 18
		title.Alignment = fyne.TextAlignCenter
		title.Move(fyne.NewPos(0, 30))
		title.Resize(fyne.NewSize(chartWidth, 30))
		objects = append(objects, title)
	}

	// Define darker, more saturated colors for better visibility on light backgrounds
	colors := []color.Color{
		color.RGBA{R: 30, G: 96, B: 215, A: 255},  // Darker Blue
		color.RGBA{R: 34, G: 139, B: 34, A: 255},  // Forest Green
		color.RGBA{R: 255, G: 140, B: 0, A: 255},  // Dark Orange
		color.RGBA{R: 178, G: 34, B: 34, A: 255},  // Firebrick Red
		color.RGBA{R: 75, G: 0, B: 130, A: 255},   // Indigo
		color.RGBA{R: 0, G: 128, B: 128, A: 255},  // Teal
		color.RGBA{R: 255, G: 69, B: 0, A: 255},   // Red Orange
		color.RGBA{R: 128, G: 0, B: 128, A: 255},  // Purple
		color.RGBA{R: 199, G: 21, B: 133, A: 255}, // Medium Violet Red
		color.RGBA{R: 0, G: 100, B: 100, A: 255},  // Dark Cyan
	}

	// Draw bars for each category
	for i, point := range data {
		percentage := (point.Value / total) * 100
		barWidth := plotWidth * float32(point.Value/total)

		y := topMargin + (barHeight+barSpacing)*float32(i)

		// Label text color (dark gray for better visibility)
		labelColor := color.RGBA{R: 66, G: 66, B: 66, A: 255}

		// Label on the left (format name)
		label := canvas.NewText(point.Label, labelColor)
		label.TextSize = 14
		label.Alignment = fyne.TextAlignTrailing
		label.Move(fyne.NewPos(10, y+barHeight/2-7))
		label.Resize(fyne.NewSize(leftMargin-20, 20))
		objects = append(objects, label)

		// Bar in the middle
		bar := canvas.NewRectangle(colors[i%len(colors)])
		bar.Resize(fyne.NewSize(barWidth, barHeight))
		bar.Move(fyne.NewPos(leftMargin, y))
		objects = append(objects, bar)

		// Percentage on the right
		percentageText := fmt.Sprintf("%.1f%%", percentage)
		valueLabel := canvas.NewText(percentageText, labelColor)
		valueLabel.TextSize = 14
		valueLabel.Alignment = fyne.TextAlignLeading
		valueLabel.Move(fyne.NewPos(leftMargin+plotWidth+15, y+barHeight/2-7))
		objects = append(objects, valueLabel)
	}

	// Create container with fixed size
	chart := container.NewWithoutLayout(objects...)
	chart.Resize(fyne.NewSize(chartWidth, chartHeight))

	return chart
}

// findMinMaxValues finds the minimum and maximum values in the data.
func findMinMaxValues(data []DataPoint) (float64, float64) {
	if len(data) == 0 {
		return 0, 100
	}

	min := data[0].Value
	max := data[0].Value

	for _, point := range data {
		if point.Value < min {
			min = point.Value
		}
		if point.Value > max {
			max = point.Value
		}
	}

	return min, max
}
