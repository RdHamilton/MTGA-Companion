package datasets

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
)

// CSVParser parses 17Lands S3 CSV datasets and calculates card metrics.
type CSVParser struct {
	// Column indices (populated during header parsing)
	colWon         int            // game outcome column
	colOpeningHand map[string]int // opening_hand_{cardname} columns
	colDrawn       map[string]int // drawn_{cardname} columns
	colDeck        map[string]int // deck_{cardname} columns
	colGameInHand  map[string]int // Alternative: in_hand_{cardname}

	// Card statistics tracking
	cardStats map[string]*CardStats
}

// CardStats tracks aggregated statistics for a single card.
type CardStats struct {
	Name   string
	Color  string
	Rarity string

	// Game counts
	GamesInHand    int // Games where card was in hand
	GamesWonInHand int // Games won when card was in hand

	GamesOpening    int // Games with card in opening hand
	GamesWonOpening int // Games won with card in opening hand

	GamesDrawn    int // Games where card was drawn
	GamesWonDrawn int // Games won when card was drawn

	GamesDeck    int // Games with card in deck
	GamesWonDeck int // Games won with card in deck

	// Pick statistics (if available in CSV)
	TotalPicks    int   // Number of times picked
	PickPositions []int // All pick positions for ALSA calculation
}

// NewCSVParser creates a new CSV parser.
func NewCSVParser() *CSVParser {
	return &CSVParser{
		colOpeningHand: make(map[string]int),
		colDrawn:       make(map[string]int),
		colDeck:        make(map[string]int),
		colGameInHand:  make(map[string]int),
		cardStats:      make(map[string]*CardStats),
	}
}

// ParseCSV parses a 17Lands CSV file and returns card ratings.
func (p *CSVParser) ParseCSV(csvPath string) ([]seventeenlands.CardRating, error) {
	// Open CSV file
	file, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Create CSV reader
	reader := csv.NewReader(file)
	reader.LazyQuotes = true // Handle malformed quotes
	reader.TrimLeadingSpace = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Parse header to find column indices
	if err := p.parseHeader(header); err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	log.Printf("[CSVParser] Parsing CSV with %d card columns", len(p.colDrawn))

	// Read and process each row (game)
	lineNum := 1
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("[CSVParser] Warning: error reading row %d: %v", lineNum, err)
			lineNum++
			continue
		}

		if err := p.processRow(row); err != nil {
			log.Printf("[CSVParser] Warning: error processing row %d: %v", lineNum, err)
		}

		lineNum++

		// Progress logging every 10k rows
		if lineNum%10000 == 0 {
			log.Printf("[CSVParser] Processed %d games...", lineNum)
		}
	}

	log.Printf("[CSVParser] Finished parsing %d games, calculating metrics...", lineNum)

	// Calculate final card ratings
	ratings := p.calculateRatings()

	log.Printf("[CSVParser] Calculated ratings for %d cards", len(ratings))
	return ratings, nil
}

// parseHeader parses the CSV header and identifies card-related columns.
func (p *CSVParser) parseHeader(header []string) error {
	// Find outcome column (won)
	p.colWon = -1
	for i, col := range header {
		colLower := strings.ToLower(col)

		if colLower == "won" || colLower == "won_game" || colLower == "game_result" {
			p.colWon = i
		}

		// Card presence columns
		// Typical format: drawn_CardName, opening_hand_CardName, deck_CardName
		if strings.HasPrefix(colLower, "drawn_") {
			cardName := strings.TrimPrefix(col, "drawn_")
			p.colDrawn[cardName] = i
		} else if strings.HasPrefix(colLower, "opening_hand_") {
			cardName := strings.TrimPrefix(col, "opening_hand_")
			p.colOpeningHand[cardName] = i
		} else if strings.HasPrefix(colLower, "deck_") {
			cardName := strings.TrimPrefix(col, "deck_")
			p.colDeck[cardName] = i
		} else if strings.HasPrefix(colLower, "in_hand_") {
			// Alternative column name
			cardName := strings.TrimPrefix(col, "in_hand_")
			p.colGameInHand[cardName] = i
		}
	}

	if p.colWon == -1 {
		return fmt.Errorf("could not find 'won' column in CSV header")
	}

	if len(p.colDrawn) == 0 && len(p.colGameInHand) == 0 {
		return fmt.Errorf("could not find any card columns in CSV header")
	}

	return nil
}

// processRow processes a single CSV row (one game).
func (p *CSVParser) processRow(row []string) error {
	if len(row) <= p.colWon {
		return fmt.Errorf("row too short")
	}

	// Parse game outcome
	won := row[p.colWon] == "1" || row[p.colWon] == "True" || row[p.colWon] == "true"

	// Process each card column
	for cardName, colIdx := range p.colDrawn {
		if colIdx >= len(row) {
			continue
		}

		// Check if card was drawn (1 = present, 0 = absent)
		if row[colIdx] == "1" || row[colIdx] == "True" || row[colIdx] == "true" {
			stats := p.getOrCreateCardStats(cardName)
			stats.GamesDrawn++
			if won {
				stats.GamesWonDrawn++
			}
		}
	}

	// Process opening hand columns
	for cardName, colIdx := range p.colOpeningHand {
		if colIdx >= len(row) {
			continue
		}

		if row[colIdx] == "1" || row[colIdx] == "True" || row[colIdx] == "true" {
			stats := p.getOrCreateCardStats(cardName)
			stats.GamesOpening++
			if won {
				stats.GamesWonOpening++
			}
		}
	}

	// Process deck columns
	for cardName, colIdx := range p.colDeck {
		if colIdx >= len(row) {
			continue
		}

		if row[colIdx] == "1" || row[colIdx] == "True" || row[colIdx] == "true" {
			stats := p.getOrCreateCardStats(cardName)
			stats.GamesDeck++
			if won {
				stats.GamesWonDeck++
			}
		}
	}

	// Process in_hand columns (use as GIH if drawn not available)
	for cardName, colIdx := range p.colGameInHand {
		if colIdx >= len(row) {
			continue
		}

		if row[colIdx] == "1" || row[colIdx] == "True" || row[colIdx] == "true" {
			stats := p.getOrCreateCardStats(cardName)
			stats.GamesInHand++
			if won {
				stats.GamesWonInHand++
			}
		}
	}

	return nil
}

// getOrCreateCardStats retrieves or creates card statistics.
func (p *CSVParser) getOrCreateCardStats(cardName string) *CardStats {
	if stats, exists := p.cardStats[cardName]; exists {
		return stats
	}

	stats := &CardStats{
		Name:          cardName,
		PickPositions: []int{},
	}
	p.cardStats[cardName] = stats
	return stats
}

// calculateRatings converts card statistics to CardRating structs.
func (p *CSVParser) calculateRatings() []seventeenlands.CardRating {
	ratings := make([]seventeenlands.CardRating, 0, len(p.cardStats))

	for _, stats := range p.cardStats {
		// Calculate GIHWR (Games In Hand Win Rate)
		// Use drawn stats as proxy for "in hand" if GIH not available
		gihwr := 0.0
		gih := stats.GamesInHand
		if gih == 0 {
			gih = stats.GamesDrawn
		}
		if gih > 0 {
			gihWins := stats.GamesWonInHand
			if gihWins == 0 {
				gihWins = stats.GamesWonDrawn
			}
			gihwr = (float64(gihWins) / float64(gih)) * 100.0
		}

		// Calculate OHWR (Opening Hand Win Rate)
		ohwr := 0.0
		if stats.GamesOpening > 0 {
			ohwr = (float64(stats.GamesWonOpening) / float64(stats.GamesOpening)) * 100.0
		}

		// Calculate GPWR (Game Present Win Rate) from deck stats
		gpwr := 0.0
		if stats.GamesDeck > 0 {
			gpwr = (float64(stats.GamesWonDeck) / float64(stats.GamesDeck)) * 100.0
		}

		// Calculate ALSA if pick data available
		alsa := 0.0
		if len(stats.PickPositions) > 0 {
			sum := 0
			for _, pos := range stats.PickPositions {
				sum += pos
			}
			alsa = float64(sum) / float64(len(stats.PickPositions))
		}

		rating := seventeenlands.CardRating{
			Name:   stats.Name,
			Color:  stats.Color,
			Rarity: stats.Rarity,

			// Win rates (as percentages)
			GIHWR: gihwr,
			OHWR:  ohwr,
			GPWR:  gpwr,

			// Game counts
			GIH: gih,
			OH:  stats.GamesOpening,
			GP:  stats.GamesDeck,
			GD:  stats.GamesDrawn,

			// Pick metrics
			ALSA: alsa,
		}

		ratings = append(ratings, rating)
	}

	return ratings
}
