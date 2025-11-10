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
			label := canvas.NewText(fmt.Sprintf("%.1f", value), color.Black)
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
			label := canvas.NewText(point.Label, color.Black)
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

	// Add title
	if config.Title != "" {
		title := canvas.NewText(config.Title, color.Black)
		title.TextSize = 16
		title.Alignment = fyne.TextAlignCenter
		title.Move(fyne.NewPos(chartWidth/2-100, 10))
		objects = append(objects, title)
	}

	// Y-axis label
	yAxisLabel := canvas.NewText("Win Rate (%)", color.Black)
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
			label := canvas.NewText(fmt.Sprintf("%.1f", value), color.Black)
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
		label := canvas.NewText(point.Label, color.Black)
		label.TextSize = 8
		label.Alignment = fyne.TextAlignCenter
		label.Move(fyne.NewPos(x-10, topMargin+plotHeight+10))
		label.Resize(fyne.NewSize(barWidth+20, 20))
		objects = append(objects, label)

		// Value label on top of bar
		valueLabel := canvas.NewText(fmt.Sprintf("%.1f", point.Value), color.Black)
		valueLabel.TextSize = 9
		valueLabel.Alignment = fyne.TextAlignCenter
		valueLabel.Move(fyne.NewPos(x, y-15))
		objects = append(objects, valueLabel)
	}

	// Add title
	if config.Title != "" {
		title := canvas.NewText(config.Title, color.Black)
		title.TextSize = 16
		title.Alignment = fyne.TextAlignCenter
		title.Move(fyne.NewPos(chartWidth/2-100, 10))
		objects = append(objects, title)
	}

	// Y-axis label
	yAxisLabel := canvas.NewText("Win Rate (%)", color.Black)
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
	leftMargin := float32(150)
	rightMargin := float32(100)
	topMargin := float32(50)
	bottomMargin := float32(20)

	plotWidth := chartWidth - leftMargin - rightMargin
	plotHeight := chartHeight - topMargin - bottomMargin

	barHeight := plotHeight / float32(len(data))
	if barHeight > 40 {
		barHeight = 40
	}

	// Container for all chart elements
	objects := []fyne.CanvasObject{}

	// Add title
	if config.Title != "" {
		title := canvas.NewText(config.Title, color.Black)
		title.TextSize = 16
		title.Alignment = fyne.TextAlignCenter
		title.Move(fyne.NewPos(chartWidth/2-100, 10))
		objects = append(objects, title)
	}

	// Define colors for segments
	colors := []color.Color{
		color.RGBA{R: 84, G: 112, B: 198, A: 255},
		color.RGBA{R: 145, G: 204, B: 117, A: 255},
		color.RGBA{R: 250, G: 200, B: 88, A: 255},
		color.RGBA{R: 238, G: 102, B: 102, A: 255},
		color.RGBA{R: 115, G: 192, B: 222, A: 255},
		color.RGBA{R: 59, G: 162, B: 114, A: 255},
		color.RGBA{R: 252, G: 132, B: 82, A: 255},
		color.RGBA{R: 154, G: 96, B: 180, A: 255},
		color.RGBA{R: 234, G: 124, B: 204, A: 255},
	}

	// Draw bars for each category
	for i, point := range data {
		percentage := (point.Value / total) * 100
		barWidth := plotWidth * float32(point.Value/total)

		y := topMargin + (barHeight+10)*float32(i)

		// Label
		label := canvas.NewText(point.Label, color.Black)
		label.TextSize = 11
		label.Move(fyne.NewPos(10, y+barHeight/2-7))
		objects = append(objects, label)

		// Bar
		bar := canvas.NewRectangle(colors[i%len(colors)])
		bar.Resize(fyne.NewSize(barWidth, barHeight))
		bar.Move(fyne.NewPos(leftMargin, y))
		objects = append(objects, bar)

		// Value and percentage
		valueText := fmt.Sprintf("%.0f (%.1f%%)", point.Value, percentage)
		valueLabel := canvas.NewText(valueText, color.Black)
		valueLabel.TextSize = 10
		valueLabel.Move(fyne.NewPos(leftMargin+barWidth+10, y+barHeight/2-7))
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
