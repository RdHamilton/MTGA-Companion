package deckexport

import (
	"fmt"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// ExportFormat represents the format to export the deck in.
type ExportFormat string

const (
	FormatArena       ExportFormat = "arena"       // MTGA Arena format
	FormatPlainText   ExportFormat = "plaintext"   // Simple text list (4x Card Name)
	FormatMTGO        ExportFormat = "mtgo"        // MTGO format
	FormatMTGGoldfish ExportFormat = "mtggoldfish" // MTGGoldfish format
)

// ExportOptions controls deck export behavior.
type ExportOptions struct {
	Format         ExportFormat
	IncludeStats   bool // Include deck statistics in export (as comments)
	IncludeHeaders bool // Include section headers (Deck, Sideboard, etc.)
}

// DeckExport represents an exported deck.
type DeckExport struct {
	Content  string       // The exported deck text
	Format   ExportFormat // The format used
	Filename string       // Suggested filename for download
}

// CardProvider is an interface for getting card information.
type CardProvider interface {
	GetCard(id int) (*cards.Card, error)
}

// Exporter handles deck export to various formats.
type Exporter struct {
	cardProvider CardProvider
}

// NewExporter creates a new deck exporter.
func NewExporter(cardProvider CardProvider) *Exporter {
	return &Exporter{
		cardProvider: cardProvider,
	}
}

// Export exports a deck to the specified format.
func (e *Exporter) Export(deck *models.Deck, deckCards []*models.DeckCard, options *ExportOptions) (*DeckExport, error) {
	if deck == nil {
		return nil, fmt.Errorf("deck is nil")
	}

	if options == nil {
		options = &ExportOptions{
			Format:         FormatArena,
			IncludeHeaders: true,
			IncludeStats:   false,
		}
	}

	// Get card metadata for all cards
	cardMetadata := make(map[int]*cards.Card)
	for _, deckCard := range deckCards {
		if _, exists := cardMetadata[deckCard.CardID]; !exists && e.cardProvider != nil {
			card, err := e.cardProvider.GetCard(deckCard.CardID)
			if err != nil {
				return nil, fmt.Errorf("failed to get card %d: %w", deckCard.CardID, err)
			}
			cardMetadata[deckCard.CardID] = card
		}
	}

	var content string
	var filename string

	switch options.Format {
	case FormatArena:
		content = e.exportArena(deck, deckCards, cardMetadata, options)
		filename = fmt.Sprintf("%s.txt", sanitizeFilename(deck.Name))
	case FormatPlainText:
		content = e.exportPlainText(deck, deckCards, cardMetadata, options)
		filename = fmt.Sprintf("%s.txt", sanitizeFilename(deck.Name))
	case FormatMTGO:
		content = e.exportMTGO(deck, deckCards, cardMetadata, options)
		filename = fmt.Sprintf("%s.dek", sanitizeFilename(deck.Name))
	case FormatMTGGoldfish:
		content = e.exportMTGGoldfish(deck, deckCards, cardMetadata, options)
		filename = fmt.Sprintf("%s.txt", sanitizeFilename(deck.Name))
	default:
		return nil, fmt.Errorf("unsupported export format: %s", options.Format)
	}

	return &DeckExport{
		Content:  content,
		Format:   options.Format,
		Filename: filename,
	}, nil
}

// exportArena exports deck in MTGA Arena format.
// Format: "4 Lightning Bolt (M21) 123"
func (e *Exporter) exportArena(deck *models.Deck, deckCards []*models.DeckCard, cardMetadata map[int]*cards.Card, options *ExportOptions) string {
	var sb strings.Builder

	// Add deck header
	if options.IncludeHeaders {
		sb.WriteString("Deck\n")
	}

	// Add mainboard cards
	mainboard := filterCardsByBoard(deckCards, "main")
	for _, deckCard := range mainboard {
		card, ok := cardMetadata[deckCard.CardID]
		if !ok {
			continue
		}

		// Format: "4 Card Name (SET) 123"
		line := fmt.Sprintf("%d %s", deckCard.Quantity, card.Name)
		if card.SetCode != "" && card.CollectorNumber != "" {
			line += fmt.Sprintf(" (%s) %s", strings.ToUpper(card.SetCode), card.CollectorNumber)
		}
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	// Add sideboard if present
	sideboard := filterCardsByBoard(deckCards, "sideboard")
	if len(sideboard) > 0 {
		sb.WriteString("\n") // Empty line before sideboard
		for _, deckCard := range sideboard {
			card, ok := cardMetadata[deckCard.CardID]
			if !ok {
				continue
			}

			line := fmt.Sprintf("%d %s", deckCard.Quantity, card.Name)
			if card.SetCode != "" && card.CollectorNumber != "" {
				line += fmt.Sprintf(" (%s) %s", strings.ToUpper(card.SetCode), card.CollectorNumber)
			}
			sb.WriteString(line)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// exportPlainText exports deck in simple plain text format.
// Format: "4x Card Name" or "4 Card Name"
func (e *Exporter) exportPlainText(deck *models.Deck, deckCards []*models.DeckCard, cardMetadata map[int]*cards.Card, options *ExportOptions) string {
	var sb strings.Builder

	// Add deck name as comment
	if options.IncludeHeaders {
		sb.WriteString(fmt.Sprintf("// %s\n", deck.Name))
		if deck.Format != "" {
			sb.WriteString(fmt.Sprintf("// Format: %s\n", deck.Format))
		}
		sb.WriteString("\n")
	}

	// Add mainboard
	if options.IncludeHeaders {
		sb.WriteString("Mainboard:\n")
	}

	mainboard := filterCardsByBoard(deckCards, "main")
	for _, deckCard := range mainboard {
		card, ok := cardMetadata[deckCard.CardID]
		if !ok {
			continue
		}

		sb.WriteString(fmt.Sprintf("%dx %s\n", deckCard.Quantity, card.Name))
	}

	// Add sideboard if present
	sideboard := filterCardsByBoard(deckCards, "sideboard")
	if len(sideboard) > 0 {
		sb.WriteString("\n")
		if options.IncludeHeaders {
			sb.WriteString("Sideboard:\n")
		}
		for _, deckCard := range sideboard {
			card, ok := cardMetadata[deckCard.CardID]
			if !ok {
				continue
			}

			sb.WriteString(fmt.Sprintf("%dx %s\n", deckCard.Quantity, card.Name))
		}
	}

	return sb.String()
}

// exportMTGO exports deck in MTGO format.
// MTGO uses quantity on the left, no 'x', and sideboard is marked with "SB:" prefix
func (e *Exporter) exportMTGO(deck *models.Deck, deckCards []*models.DeckCard, cardMetadata map[int]*cards.Card, options *ExportOptions) string {
	var sb strings.Builder

	// Add mainboard cards
	mainboard := filterCardsByBoard(deckCards, "main")
	for _, deckCard := range mainboard {
		card, ok := cardMetadata[deckCard.CardID]
		if !ok {
			continue
		}

		// MTGO format: "4 Card Name"
		sb.WriteString(fmt.Sprintf("%d %s\n", deckCard.Quantity, card.Name))
	}

	// Add sideboard with "SB:" prefix
	sideboard := filterCardsByBoard(deckCards, "sideboard")
	if len(sideboard) > 0 {
		sb.WriteString("\n") // Empty line before sideboard
		for _, deckCard := range sideboard {
			card, ok := cardMetadata[deckCard.CardID]
			if !ok {
				continue
			}

			// MTGO sideboard format: "SB: 2 Card Name"
			sb.WriteString(fmt.Sprintf("SB: %d %s\n", deckCard.Quantity, card.Name))
		}
	}

	return sb.String()
}

// exportMTGGoldfish exports deck in MTGGoldfish format.
// Similar to plain text but with specific formatting
func (e *Exporter) exportMTGGoldfish(deck *models.Deck, deckCards []*models.DeckCard, cardMetadata map[int]*cards.Card, options *ExportOptions) string {
	var sb strings.Builder

	// Add mainboard cards
	mainboard := filterCardsByBoard(deckCards, "main")
	for _, deckCard := range mainboard {
		card, ok := cardMetadata[deckCard.CardID]
		if !ok {
			continue
		}

		// MTGGoldfish format: "4 Card Name"
		sb.WriteString(fmt.Sprintf("%d %s\n", deckCard.Quantity, card.Name))
	}

	// Add sideboard
	sideboard := filterCardsByBoard(deckCards, "sideboard")
	if len(sideboard) > 0 {
		sb.WriteString("\n") // Empty line before sideboard
		sb.WriteString("Sideboard\n")
		for _, deckCard := range sideboard {
			card, ok := cardMetadata[deckCard.CardID]
			if !ok {
				continue
			}

			sb.WriteString(fmt.Sprintf("%d %s\n", deckCard.Quantity, card.Name))
		}
	}

	return sb.String()
}

// filterCardsByBoard returns cards from a specific board.
func filterCardsByBoard(deckCards []*models.DeckCard, board string) []*models.DeckCard {
	filtered := make([]*models.DeckCard, 0)
	for _, card := range deckCards {
		if card.Board == board {
			filtered = append(filtered, card)
		}
	}
	return filtered
}

// sanitizeFilename removes invalid characters from filename.
func sanitizeFilename(name string) string {
	// Replace invalid filename characters with underscore
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}
	// Trim spaces and limit length
	result = strings.TrimSpace(result)
	if len(result) > 100 {
		result = result[:100]
	}
	if result == "" {
		result = "deck"
	}
	return result
}
