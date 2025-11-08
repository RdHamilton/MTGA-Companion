package viewer

import (
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// DraftViewer provides draft viewing with card metadata.
type DraftViewer struct {
	cardService *cards.Service
}

// NewDraftViewer creates a new draft viewer.
func NewDraftViewer(cardService *cards.Service) *DraftViewer {
	return &DraftViewer{
		cardService: cardService,
	}
}

// DraftPickView represents a single draft pick with metadata.
type DraftPickView struct {
	CourseID         string
	PackNumber       int
	PickNumber       int
	SelectedCard     *models.DraftCardView
	AvailableCards   []*models.DraftCardView
	TotalAvailable   int
	CardsWithMeta    int // Number of cards that have metadata
}

// DraftView represents a complete draft with all picks and metadata.
type DraftView struct {
	CourseID      string
	Picks         []*DraftPickView
	TotalPicks    int
	DeckCards     []*models.DraftCardView // All picked cards
	ColorAnalysis *DraftColorAnalysis
	CurveAnalysis map[int]int // CMC -> count
}

// DraftColorAnalysis provides analysis of colors in a draft pool.
type DraftColorAnalysis struct {
	White int
	Blue  int
	Black int
	Red   int
	Green int
	Gold  int // Multi-color cards
	Colorless int
}

// EnhanceDraftPicks adds card metadata to draft picks.
func (dv *DraftViewer) EnhanceDraftPicks(draftPicks *logreader.DraftPicks) (*DraftView, error) {
	if draftPicks == nil || len(draftPicks.Picks) == 0 {
		return nil, nil
	}

	view := &DraftView{
		CourseID:      draftPicks.CourseID,
		Picks:         make([]*DraftPickView, 0, len(draftPicks.Picks)),
		TotalPicks:    len(draftPicks.Picks),
		DeckCards:     make([]*models.DraftCardView, 0),
		ColorAnalysis: &DraftColorAnalysis{},
		CurveAnalysis: make(map[int]int),
	}

	// Collect all unique card IDs
	cardIDSet := make(map[int]bool)
	for _, pick := range draftPicks.Picks {
		if pick.SelectedCard > 0 {
			cardIDSet[pick.SelectedCard] = true
		}
		for _, cardID := range pick.AvailableCards {
			if cardID > 0 {
				cardIDSet[cardID] = true
			}
		}
	}

	// Convert set to slice
	cardIDs := make([]int, 0, len(cardIDSet))
	for cardID := range cardIDSet {
		cardIDs = append(cardIDs, cardID)
	}

	// Fetch all metadata
	cardMetadata, err := dv.cardService.GetCards(cardIDs)
	if err != nil {
		cardMetadata = make(map[int]*cards.Card)
	}

	// Process each pick
	for i, pick := range draftPicks.Picks {
		pickView := &DraftPickView{
			CourseID:       pick.CourseID,
			PackNumber:     pick.PackNumber,
			PickNumber:     pick.PickNumber,
			AvailableCards: make([]*models.DraftCardView, 0, len(pick.AvailableCards)),
			TotalAvailable: len(pick.AvailableCards),
		}

		// Add selected card
		if pick.SelectedCard > 0 {
			selectedMeta := cardMetadata[pick.SelectedCard]
			pickView.SelectedCard = &models.DraftCardView{
				CardID:   pick.SelectedCard,
				Pack:     pick.PackNumber,
				Pick:     pick.PickNumber,
				Round:    i + 1,
				Metadata: selectedMeta,
			}

			// Add to deck cards
			view.DeckCards = append(view.DeckCards, pickView.SelectedCard)

			// Update color analysis
			if selectedMeta != nil {
				dv.updateColorAnalysis(view.ColorAnalysis, selectedMeta)
				// Update curve analysis
				cmc := int(selectedMeta.CMC)
				view.CurveAnalysis[cmc]++
			}
		}

		// Add available cards
		for _, cardID := range pick.AvailableCards {
			meta := cardMetadata[cardID]
			cardView := &models.DraftCardView{
				CardID:   cardID,
				Pack:     pick.PackNumber,
				Pick:     pick.PickNumber,
				Round:    i + 1,
				Metadata: meta,
			}
			pickView.AvailableCards = append(pickView.AvailableCards, cardView)
			if meta != nil {
				pickView.CardsWithMeta++
			}
		}

		view.Picks = append(view.Picks, pickView)
	}

	return view, nil
}

// EnhanceDraftPicksList enhances multiple draft events with metadata.
func (dv *DraftViewer) EnhanceDraftPicksList(allDraftPicks []*logreader.DraftPicks) ([]*DraftView, error) {
	if allDraftPicks == nil {
		return nil, nil
	}

	views := make([]*DraftView, 0, len(allDraftPicks))
	for _, draftPicks := range allDraftPicks {
		view, err := dv.EnhanceDraftPicks(draftPicks)
		if err != nil {
			continue // Skip drafts that fail to enhance
		}
		if view != nil {
			views = append(views, view)
		}
	}

	return views, nil
}

// GetDraftDeck returns all cards picked in a draft as a deck-like structure.
func (dv *DraftViewer) GetDraftDeck(draftView *DraftView) []*models.DraftCardView {
	if draftView == nil {
		return nil
	}
	return draftView.DeckCards
}

// GetPicksByPack returns all picks from a specific pack.
func (dv *DraftViewer) GetPicksByPack(draftView *DraftView, packNumber int) []*DraftPickView {
	if draftView == nil {
		return nil
	}

	picks := make([]*DraftPickView, 0)
	for _, pick := range draftView.Picks {
		if pick.PackNumber == packNumber {
			picks = append(picks, pick)
		}
	}

	return picks
}

// AnalyzeDraftColors provides detailed color analysis for a draft.
func (dv *DraftViewer) AnalyzeDraftColors(draftView *DraftView) *DraftColorAnalysis {
	if draftView == nil || draftView.ColorAnalysis == nil {
		return &DraftColorAnalysis{}
	}
	return draftView.ColorAnalysis
}

// GetTopColors returns the most-picked colors in order.
func (dv *DraftViewer) GetTopColors(draftView *DraftView) []string {
	if draftView == nil || draftView.ColorAnalysis == nil {
		return nil
	}

	analysis := draftView.ColorAnalysis
	colorCounts := map[string]int{
		"W": analysis.White,
		"U": analysis.Blue,
		"B": analysis.Black,
		"R": analysis.Red,
		"G": analysis.Green,
	}

	// Sort colors by count
	colors := []string{"W", "U", "B", "R", "G"}
	for i := 0; i < len(colors); i++ {
		for j := i + 1; j < len(colors); j++ {
			if colorCounts[colors[i]] < colorCounts[colors[j]] {
				colors[i], colors[j] = colors[j], colors[i]
			}
		}
	}

	// Filter out colors with 0 count
	topColors := make([]string, 0)
	for _, color := range colors {
		if colorCounts[color] > 0 {
			topColors = append(topColors, color)
		}
	}

	return topColors
}

// updateColorAnalysis updates the color analysis based on a card.
func (dv *DraftViewer) updateColorAnalysis(analysis *DraftColorAnalysis, card *cards.Card) {
	if card == nil {
		return
	}

	colorCount := len(card.Colors)

	if colorCount == 0 {
		analysis.Colorless++
		return
	}

	if colorCount > 1 {
		analysis.Gold++
		return
	}

	// Single color card
	for _, color := range card.Colors {
		switch color {
		case "W":
			analysis.White++
		case "U":
			analysis.Blue++
		case "B":
			analysis.Black++
		case "R":
			analysis.Red++
		case "G":
			analysis.Green++
		}
	}
}

// GetManaCurve returns the mana curve for a draft pool.
func (dv *DraftViewer) GetManaCurve(draftView *DraftView) map[int]int {
	if draftView == nil {
		return make(map[int]int)
	}
	return draftView.CurveAnalysis
}
