package deckimport

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
)

// ParsedCard represents a single card in a deck import.
type ParsedCard struct {
	Quantity int
	Name     string
	SetCode  string // Optional, extracted from formats like "4 Lightning Bolt (M21) 123"
	Board    string // "main" or "sideboard"
}

// ParsedDeck represents a deck parsed from an import string.
type ParsedDeck struct {
	Name      string
	Format    string
	Mainboard []*ParsedCard
	Sideboard []*ParsedCard
	ParsedOK  bool
	Errors    []string
	Warnings  []string
}

// Parser handles deck import parsing from various formats.
type Parser struct {
	cardService *cards.Service
}

// NewParser creates a new deck import parser.
func NewParser(cardService *cards.Service) *Parser {
	return &Parser{
		cardService: cardService,
	}
}

// ParseResult contains the result of parsing a deck import.
type ParseResult struct {
	Deck     *ParsedDeck
	CardIDs  map[string]int // Map of card name to card ID for validation
	Errors   []error
	Warnings []string
}

// Parse attempts to parse deck import text from multiple formats.
// It tries Arena format first, then falls back to plain text.
func (p *Parser) Parse(input string) (*ParseResult, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, fmt.Errorf("empty import string")
	}

	// Try Arena format first
	if result, err := p.ParseArenaFormat(input); err == nil && result.Deck.ParsedOK {
		return result, nil
	}

	// Try plain text format
	if result, err := p.ParsePlainText(input); err == nil && result.Deck.ParsedOK {
		return result, nil
	}

	return nil, fmt.Errorf("unable to parse deck format")
}

// ParseArenaFormat parses MTGA Arena deck export format.
// Format example:
//
//	Deck
//	4 Lightning Bolt (M21) 123
//	2 Shock (M21) 124
//
//	2 Duress (M21) 95
//
// The empty line separates mainboard from sideboard.
func (p *Parser) ParseArenaFormat(input string) (*ParseResult, error) {
	result := &ParseResult{
		Deck: &ParsedDeck{
			Mainboard: make([]*ParsedCard, 0),
			Sideboard: make([]*ParsedCard, 0),
			ParsedOK:  true,
			Errors:    make([]string, 0),
			Warnings:  make([]string, 0),
		},
		CardIDs:  make(map[string]int),
		Errors:   make([]error, 0),
		Warnings: make([]string, 0),
	}

	lines := strings.Split(input, "\n")
	board := "main"
	foundEmptyLine := false

	// Arena format regex: "4 Lightning Bolt (M21) 123" or "4 Lightning Bolt"
	// Group 1: quantity, Group 2: card name, Group 3: set code (optional), Group 4: collector number (optional)
	arenaRegex := regexp.MustCompile(`^(\d+)\s+([^(]+?)(?:\s+\(([A-Z0-9]+)\)(?:\s+(\d+))?)?$`)

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Skip "Deck" header
		if line == "Deck" {
			continue
		}

		// Empty line switches to sideboard
		if line == "" {
			if !foundEmptyLine && board == "main" {
				board = "sideboard"
				foundEmptyLine = true
			}
			continue
		}

		// Parse card line
		matches := arenaRegex.FindStringSubmatch(line)
		if matches == nil {
			result.Deck.Warnings = append(result.Deck.Warnings,
				fmt.Sprintf("Line %d: Could not parse '%s'", i+1, line))
			continue
		}

		quantity, err := strconv.Atoi(matches[1])
		if err != nil {
			result.Deck.Errors = append(result.Deck.Errors,
				fmt.Sprintf("Line %d: Invalid quantity '%s'", i+1, matches[1]))
			result.Deck.ParsedOK = false
			continue
		}

		cardName := strings.TrimSpace(matches[2])
		setCode := ""
		if len(matches) > 3 && matches[3] != "" {
			setCode = matches[3]
		}

		parsedCard := &ParsedCard{
			Quantity: quantity,
			Name:     cardName,
			SetCode:  setCode,
			Board:    board,
		}

		if board == "main" {
			result.Deck.Mainboard = append(result.Deck.Mainboard, parsedCard)
		} else {
			result.Deck.Sideboard = append(result.Deck.Sideboard, parsedCard)
		}

		// Try to resolve card name to ID
		if p.cardService != nil {
			if cardID, err := p.resolveCardID(cardName, setCode); err == nil {
				result.CardIDs[cardName] = cardID
			} else {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Card '%s' not found in database", cardName))
			}
		}
	}

	if len(result.Deck.Mainboard) == 0 && len(result.Deck.Sideboard) == 0 {
		result.Deck.ParsedOK = false
		result.Deck.Errors = append(result.Deck.Errors, "No cards found in import")
	}

	return result, nil
}

// ParsePlainText parses simple text format card lists.
// Format examples:
//   - "4 Lightning Bolt"
//   - "4x Lightning Bolt"
//   - "Lightning Bolt x4"
func (p *Parser) ParsePlainText(input string) (*ParseResult, error) {
	result := &ParseResult{
		Deck: &ParsedDeck{
			Mainboard: make([]*ParsedCard, 0),
			Sideboard: make([]*ParsedCard, 0),
			ParsedOK:  true,
			Errors:    make([]string, 0),
			Warnings:  make([]string, 0),
		},
		CardIDs:  make(map[string]int),
		Errors:   make([]error, 0),
		Warnings: make([]string, 0),
	}

	lines := strings.Split(input, "\n")
	board := "main"

	// Plain text regex patterns
	// Pattern 1: "4 Card Name" or "4x Card Name"
	pattern1 := regexp.MustCompile(`^(\d+)x?\s+(.+)$`)
	// Pattern 2: "Card Name x4"
	pattern2 := regexp.MustCompile(`^(.+?)\s+x(\d+)$`)

	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check for sideboard marker
		if strings.HasPrefix(strings.ToLower(line), "sideboard") {
			board = "sideboard"
			continue
		}

		var quantity int
		var cardName string
		var matched bool

		// Try pattern 1: "4 Card Name" or "4x Card Name"
		if matches := pattern1.FindStringSubmatch(line); matches != nil {
			q, err := strconv.Atoi(matches[1])
			if err == nil {
				quantity = q
				cardName = strings.TrimSpace(matches[2])
				matched = true
			}
		}

		// Try pattern 2: "Card Name x4"
		if !matched {
			if matches := pattern2.FindStringSubmatch(line); matches != nil {
				q, err := strconv.Atoi(matches[2])
				if err == nil {
					quantity = q
					cardName = strings.TrimSpace(matches[1])
					matched = true
				}
			}
		}

		if !matched {
			result.Deck.Warnings = append(result.Deck.Warnings,
				fmt.Sprintf("Line %d: Could not parse '%s'", i+1, line))
			continue
		}

		parsedCard := &ParsedCard{
			Quantity: quantity,
			Name:     cardName,
			Board:    board,
		}

		if board == "main" {
			result.Deck.Mainboard = append(result.Deck.Mainboard, parsedCard)
		} else {
			result.Deck.Sideboard = append(result.Deck.Sideboard, parsedCard)
		}

		// Try to resolve card name to ID
		if p.cardService != nil {
			if cardID, err := p.resolveCardID(cardName, ""); err == nil {
				result.CardIDs[cardName] = cardID
			} else {
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("Card '%s' not found in database", cardName))
			}
		}
	}

	if len(result.Deck.Mainboard) == 0 && len(result.Deck.Sideboard) == 0 {
		result.Deck.ParsedOK = false
		result.Deck.Errors = append(result.Deck.Errors, "No cards found in import")
	}

	return result, nil
}

// resolveCardID attempts to find the card ID for a given card name.
func (p *Parser) resolveCardID(cardName, setCode string) (int, error) {
	if p.cardService == nil {
		return 0, fmt.Errorf("card service not available")
	}

	// TODO: Implement card lookup by name
	// For now, return an error - this will be implemented when we integrate with card service
	return 0, fmt.Errorf("card lookup not yet implemented")
}

// ValidateDraftImport validates that all cards in the parsed deck exist in the draft pool.
func (p *Parser) ValidateDraftImport(result *ParseResult, draftCardIDs []int) []error {
	errors := make([]error, 0)

	// Create a set of draft card IDs for O(1) lookup
	draftCardSet := make(map[int]bool)
	for _, cardID := range draftCardIDs {
		draftCardSet[cardID] = true
	}

	// Check all mainboard cards
	for _, card := range result.Deck.Mainboard {
		if cardID, ok := result.CardIDs[card.Name]; ok {
			if !draftCardSet[cardID] {
				errors = append(errors, fmt.Errorf("card '%s' not found in draft pool", card.Name))
			}
		}
	}

	// Check all sideboard cards
	for _, card := range result.Deck.Sideboard {
		if cardID, ok := result.CardIDs[card.Name]; ok {
			if !draftCardSet[cardID] {
				errors = append(errors, fmt.Errorf("card '%s' not found in draft pool", card.Name))
			}
		}
	}

	return errors
}
