package display

import (
	"context"
	"fmt"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// DraftPicksDisplayer handles displaying draft picks in a readable format.
type DraftPicksDisplayer struct {
	cardService *cards.Service
}

// NewDraftPicksDisplayer creates a new draft picks displayer.
func NewDraftPicksDisplayer(cardService *cards.Service) *DraftPicksDisplayer {
	return &DraftPicksDisplayer{
		cardService: cardService,
	}
}

// DisplayPicks displays all picks for a draft event.
func (d *DraftPicksDisplayer) DisplayPicks(ctx context.Context, picks []*models.DraftPick) error {
	if len(picks) == 0 {
		fmt.Println("No picks found for this draft event.")
		return nil
	}

	// Group picks by pack
	picksByPack := make(map[int][]*models.DraftPick)
	for _, pick := range picks {
		picksByPack[pick.PackNumber] = append(picksByPack[pick.PackNumber], pick)
	}

	// Display header
	fmt.Printf("\n")
	fmt.Printf("Draft Picks Summary\n")
	fmt.Printf("===================\n")
	fmt.Printf("Total Picks: %d\n", len(picks))
	fmt.Printf("Packs: %d\n\n", len(picksByPack))

	// Display each pack
	for packNum := 1; packNum <= len(picksByPack); packNum++ {
		packPicks := picksByPack[packNum]
		if len(packPicks) == 0 {
			continue
		}

		fmt.Printf("═══════════════════════════════════════════════════════════════\n")
		fmt.Printf("Pack %d (%d picks)\n", packNum, len(packPicks))
		fmt.Printf("═══════════════════════════════════════════════════════════════\n\n")

		for _, pick := range packPicks {
			if err := d.displaySinglePick(ctx, pick); err != nil {
				// Log error but continue
				fmt.Printf("Error displaying pick %d: %v\n", pick.PickNumber, err)
			}
			fmt.Println()
		}
	}

	return nil
}

// displaySinglePick displays a single pick with pack contents.
func (d *DraftPicksDisplayer) displaySinglePick(ctx context.Context, pick *models.DraftPick) error {
	fmt.Printf("Pick %d:\n", pick.PickNumber)
	fmt.Printf("├─ Time: %s\n", pick.Timestamp.Format("2006-01-02 15:04:05"))
	fmt.Printf("├─ Pack Contents (%d cards):\n", len(pick.AvailableCards))

	// Display available cards
	for i, cardID := range pick.AvailableCards {
		isSelected := cardID == pick.SelectedCard
		prefix := "│  "
		if i == len(pick.AvailableCards)-1 {
			if isSelected {
				prefix = "└─►"
			} else {
				prefix = "└─ "
			}
		} else if isSelected {
			prefix = "├─►"
		} else {
			prefix = "├─ "
		}

		cardName := d.getCardName(cardID)
		if isSelected {
			fmt.Printf("%s [PICKED] %s\n", prefix, cardName)
		} else {
			fmt.Printf("%s %s\n", prefix, cardName)
		}
	}

	return nil
}

// DisplayPicksCompact displays picks in a more compact format.
func (d *DraftPicksDisplayer) DisplayPicksCompact(ctx context.Context, picks []*models.DraftPick) error {
	if len(picks) == 0 {
		fmt.Println("No picks found for this draft event.")
		return nil
	}

	// Group picks by pack
	picksByPack := make(map[int][]*models.DraftPick)
	for _, pick := range picks {
		picksByPack[pick.PackNumber] = append(picksByPack[pick.PackNumber], pick)
	}

	fmt.Printf("\nDraft Picks Summary (%d total picks, %d packs)\n\n", len(picks), len(picksByPack))

	// Display header
	fmt.Printf("%-6s %-6s %-30s %-10s\n", "Pack", "Pick", "Card Picked", "Pack Size")
	fmt.Printf("%s\n", strings.Repeat("─", 60))

	for packNum := 1; packNum <= len(picksByPack); packNum++ {
		packPicks := picksByPack[packNum]
		for _, pick := range packPicks {
			cardName := d.getCardName(pick.SelectedCard)
			fmt.Printf("%-6d %-6d %-30s %-10d\n",
				pick.PackNumber,
				pick.PickNumber,
				truncateString(cardName, 28),
				len(pick.AvailableCards),
			)
		}
	}

	fmt.Println()
	return nil
}

// DisplayPicksSummary displays a high-level summary of the draft.
func (d *DraftPicksDisplayer) DisplayPicksSummary(ctx context.Context, picks []*models.DraftPick) error {
	if len(picks) == 0 {
		fmt.Println("No picks found for this draft event.")
		return nil
	}

	// Count picks by pack
	picksByPack := make(map[int]int)
	for _, pick := range picks {
		picksByPack[pick.PackNumber]++
	}

	fmt.Printf("\nDraft Summary\n")
	fmt.Printf("═════════════\n\n")
	fmt.Printf("Total Picks: %d\n", len(picks))
	fmt.Printf("Total Packs: %d\n\n", len(picksByPack))

	fmt.Printf("Picks by Pack:\n")
	for packNum := 1; packNum <= len(picksByPack); packNum++ {
		count := picksByPack[packNum]
		fmt.Printf("  Pack %d: %d picks\n", packNum, count)
	}

	// Show first and last picks
	if len(picks) > 0 {
		fmt.Printf("\nFirst Pick: %s (Pack %d, Pick %d)\n",
			d.getCardName(picks[0].SelectedCard),
			picks[0].PackNumber,
			picks[0].PickNumber,
		)

		lastPick := picks[len(picks)-1]
		fmt.Printf("Last Pick:  %s (Pack %d, Pick %d)\n",
			d.getCardName(lastPick.SelectedCard),
			lastPick.PackNumber,
			lastPick.PickNumber,
		)
	}

	fmt.Println()
	return nil
}

// getCardName retrieves the card name for a given Arena ID.
// Returns the Arena ID as a string if the name cannot be found.
func (d *DraftPicksDisplayer) getCardName(arenaID int) string {
	if d.cardService == nil {
		return fmt.Sprintf("Card #%d", arenaID)
	}

	card, err := d.cardService.GetCard(arenaID)
	if err != nil || card == nil {
		return fmt.Sprintf("Card #%d", arenaID)
	}

	return card.Name
}

// truncateString truncates a string to the specified length, adding "..." if truncated.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
