package gui

import (
	"context"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// LoadingIndicator creates a loading spinner with a message.
func LoadingIndicator(message string) fyne.CanvasObject {
	progress := widget.NewProgressBarInfinite()
	label := widget.NewLabel(message)
	label.Alignment = fyne.TextAlignCenter

	return container.NewVBox(
		container.NewPadded(progress),
		container.NewPadded(label),
	)
}

// LoadingView creates a centered loading view with a message.
func (a *App) LoadingView(message string) fyne.CanvasObject {
	return container.NewCenter(
		container.NewVBox(
			widget.NewProgressBarInfinite(),
			widget.NewLabel(message),
		),
	)
}

// WithLoading executes a function asynchronously and shows a loading indicator.
// Returns a container that shows loading state initially, then the result.
func (a *App) WithLoading(message string, loadFunc func() (fyne.CanvasObject, error)) fyne.CanvasObject {
	// Container that will hold either loading or result
	content := container.NewStack()

	// Initially show loading
	loadingView := a.LoadingView(message)
	content.Objects = []fyne.CanvasObject{loadingView}

	// Load data asynchronously
	go func() {
		result, err := loadFunc()

		// Update UI - Fyne handles thread safety for Refresh()
		time.Sleep(50 * time.Millisecond) // Brief delay so loading indicator is visible
		if err != nil {
			content.Objects = []fyne.CanvasObject{
				a.ErrorView("Error Loading Data", err, func() fyne.CanvasObject {
					return a.WithLoading(message, loadFunc)
				}),
			}
		} else {
			content.Objects = []fyne.CanvasObject{result}
		}
		content.Refresh()
	}()

	return content
}

// LoadingOverlay creates a loading overlay on top of existing content.
// Useful for showing loading state without replacing the entire view.
func (a *App) LoadingOverlay(baseContent fyne.CanvasObject, message string) fyne.CanvasObject {
	// Simple overlay with loading indicator
	spinner := widget.NewProgressBarInfinite()
	label := widget.NewLabel(message)
	label.Alignment = fyne.TextAlignCenter

	loadingContent := container.NewCenter(
		container.NewVBox(
			spinner,
			label,
		),
	)

	return container.NewStack(
		baseContent,
		loadingContent,
	)
}

// AsyncRefresh refreshes content asynchronously with a loading indicator.
// Returns a function that triggers the refresh.
func (a *App) AsyncRefresh(
	ctx context.Context,
	containerRef *fyne.Container,
	message string,
	refreshFunc func() (fyne.CanvasObject, error),
) func() {
	return func() {
		// Show loading
		containerRef.Objects = []fyne.CanvasObject{a.LoadingView(message)}
		containerRef.Refresh()

		// Load data asynchronously
		go func() {
			result, err := refreshFunc()

			// Update UI - Fyne handles thread safety
			time.Sleep(50 * time.Millisecond)
			if err != nil {
				containerRef.Objects = []fyne.CanvasObject{
					a.ErrorView("Error Refreshing Data", err, func() fyne.CanvasObject {
						// Retry
						go func() {
							time.Sleep(100 * time.Millisecond)
							result, err := refreshFunc()
							if err == nil {
								containerRef.Objects = []fyne.CanvasObject{result}
								containerRef.Refresh()
							}
						}()
						return a.LoadingView(message)
					}),
				}
			} else {
				containerRef.Objects = []fyne.CanvasObject{result}
			}
			containerRef.Refresh()
		}()
	}
}
