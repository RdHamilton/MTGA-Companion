package importer

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// BulkImporter handles importing card data from Scryfall bulk files.
type BulkImporter struct {
	client  *scryfall.Client
	storage *storage.Service
	options BulkImportOptions
}

// BulkImportOptions configures the bulk import process.
type BulkImportOptions struct {
	// BatchSize is the number of cards to insert per transaction.
	BatchSize int

	// DataDir is the directory to store downloaded bulk files.
	DataDir string

	// MaxAge is the maximum age of a bulk file before re-downloading.
	MaxAge time.Duration

	// Progress is an optional callback for progress reporting.
	// Receives (cardsProcessed, totalCards) on each batch.
	Progress func(processed, total int)

	// ForceDownload forces re-download even if file exists and is fresh.
	ForceDownload bool

	// Verbose enables verbose logging.
	Verbose bool
}

// DefaultBulkImportOptions returns sensible default options.
func DefaultBulkImportOptions() BulkImportOptions {
	return BulkImportOptions{
		BatchSize:     500,
		DataDir:       filepath.Join(os.TempDir(), "mtga-companion", "bulk"),
		MaxAge:        24 * time.Hour,
		ForceDownload: false,
		Verbose:       false,
	}
}

// NewBulkImporter creates a new bulk importer.
func NewBulkImporter(client *scryfall.Client, storage *storage.Service, options BulkImportOptions) *BulkImporter {
	return &BulkImporter{
		client:  client,
		storage: storage,
		options: options,
	}
}

// ImportStats contains statistics about the import process.
type ImportStats struct {
	TotalCards     int
	ImportedCards  int
	SkippedCards   int
	ErrorCards     int
	Duration       time.Duration
	BulkFileURL    string
	BulkFileSize   int64
	DownloadTime   time.Duration
	ProcessingTime time.Duration
}

// Import performs the bulk import of cards from Scryfall.
func (bi *BulkImporter) Import(ctx context.Context) (*ImportStats, error) {
	startTime := time.Now()
	stats := &ImportStats{}

	// Ensure data directory exists
	if err := os.MkdirAll(bi.options.DataDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Get bulk data info from Scryfall
	if bi.options.Verbose {
		fmt.Println("Fetching bulk data information from Scryfall...")
	}

	bulkData, err := bi.client.GetBulkData(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get bulk data info: %w", err)
	}

	// Find "Default Cards" bulk data
	var defaultCards *scryfall.BulkData
	for i := range bulkData.Data {
		if bulkData.Data[i].Type == "default_cards" {
			defaultCards = &bulkData.Data[i]
			break
		}
	}

	if defaultCards == nil {
		return nil, fmt.Errorf("default_cards bulk data not found")
	}

	stats.BulkFileURL = defaultCards.DownloadURI
	stats.BulkFileSize = int64(defaultCards.CompressedSize)

	if bi.options.Verbose {
		fmt.Printf("Found bulk file: %s (%.2f MB)\n", defaultCards.Name, float64(defaultCards.CompressedSize)/(1024*1024))
	}

	// Download bulk file
	downloadStart := time.Now()
	filePath, err := bi.downloadBulkFile(ctx, defaultCards)
	if err != nil {
		return nil, fmt.Errorf("failed to download bulk file: %w", err)
	}
	defer func() {
		if !bi.options.Verbose {
			// Clean up the downloaded file after import
			os.Remove(filePath)
		}
	}()
	stats.DownloadTime = time.Since(downloadStart)

	if bi.options.Verbose {
		fmt.Printf("Downloaded in %.2fs\n", stats.DownloadTime.Seconds())
		fmt.Println("Processing cards...")
	}

	// Process the bulk file
	processStart := time.Now()
	processed, imported, skipped, errors, err := bi.processBulkFile(ctx, filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to process bulk file: %w", err)
	}
	stats.ProcessingTime = time.Since(processStart)

	stats.TotalCards = processed
	stats.ImportedCards = imported
	stats.SkippedCards = skipped
	stats.ErrorCards = errors
	stats.Duration = time.Since(startTime)

	if bi.options.Verbose {
		fmt.Printf("\nImport complete:\n")
		fmt.Printf("  Total cards: %d\n", stats.TotalCards)
		fmt.Printf("  Imported: %d\n", stats.ImportedCards)
		fmt.Printf("  Skipped (no Arena ID): %d\n", stats.SkippedCards)
		fmt.Printf("  Errors: %d\n", stats.ErrorCards)
		fmt.Printf("  Download time: %.2fs\n", stats.DownloadTime.Seconds())
		fmt.Printf("  Processing time: %.2fs\n", stats.ProcessingTime.Seconds())
		fmt.Printf("  Total time: %.2fs\n", stats.Duration.Seconds())
	}

	return stats, nil
}

// downloadBulkFile downloads the bulk file if needed.
func (bi *BulkImporter) downloadBulkFile(ctx context.Context, bulkInfo *scryfall.BulkData) (string, error) {
	fileName := filepath.Base(bulkInfo.DownloadURI)
	filePath := filepath.Join(bi.options.DataDir, fileName)

	// Check if file exists and is fresh
	if !bi.options.ForceDownload {
		if info, err := os.Stat(filePath); err == nil {
			age := time.Since(info.ModTime())
			if age < bi.options.MaxAge {
				if bi.options.Verbose {
					fmt.Printf("Using existing bulk file (age: %s)\n", age.Round(time.Minute))
				}
				return filePath, nil
			}
		}
	}

	// Download the file
	if bi.options.Verbose {
		fmt.Printf("Downloading bulk file: %s\n", fileName)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, bulkInfo.DownloadURI, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Create temporary file
	tmpFile, err := os.CreateTemp(bi.options.DataDir, "bulk-*.tmp")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Copy with progress
	written, err := io.Copy(tmpFile, resp.Body)
	if err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	tmpFile.Close()

	// Rename to final name
	if err := os.Rename(tmpPath, filePath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to rename file: %w", err)
	}

	if bi.options.Verbose {
		fmt.Printf("Downloaded %.2f MB\n", float64(written)/(1024*1024))
	}

	return filePath, nil
}

// processBulkFile processes the bulk file and imports cards.
func (bi *BulkImporter) processBulkFile(ctx context.Context, filePath string) (processed, imported, skipped, errors int, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Create gzip reader for compressed file
	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return 0, 0, 0, 0, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()

	// Create buffered scanner for line-by-line reading
	scanner := bufio.NewScanner(gzReader)
	// Increase buffer size for large lines (some cards have very long oracle text)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024) // 1MB max line size

	batch := make([]*storage.Card, 0, bi.options.BatchSize)

	for scanner.Scan() {
		// Check context cancellation
		if err := ctx.Err(); err != nil {
			return processed, imported, skipped, errors, err
		}

		processed++

		// Parse JSON line
		var card scryfall.Card
		if err := json.Unmarshal(scanner.Bytes(), &card); err != nil {
			errors++
			if bi.options.Verbose {
				fmt.Printf("Error parsing card %d: %v\n", processed, err)
			}
			continue
		}

		// Skip cards without Arena ID
		if card.ArenaID == nil || *card.ArenaID == 0 {
			skipped++
			continue
		}

		// Convert to storage card
		storageCard := ConvertToStorageCard(&card)
		batch = append(batch, storageCard)

		// Insert batch if full
		if len(batch) >= bi.options.BatchSize {
			if err := bi.insertBatch(ctx, batch); err != nil {
				return processed, imported, skipped, errors, fmt.Errorf("failed to insert batch: %w", err)
			}
			imported += len(batch)

			if bi.options.Progress != nil {
				bi.options.Progress(imported, -1) // Total unknown
			}

			if bi.options.Verbose {
				fmt.Printf("\rProcessed: %d  Imported: %d  Skipped: %d  Errors: %d", processed, imported, skipped, errors)
			}

			batch = batch[:0] // Clear batch
		}
	}

	// Insert final batch
	if len(batch) > 0 {
		if err := bi.insertBatch(ctx, batch); err != nil {
			return processed, imported, skipped, errors, fmt.Errorf("failed to insert final batch: %w", err)
		}
		imported += len(batch)
	}

	if err := scanner.Err(); err != nil {
		return processed, imported, skipped, errors, fmt.Errorf("scanner error: %w", err)
	}

	return processed, imported, skipped, errors, nil
}

// insertBatch inserts a batch of cards into the database.
func (bi *BulkImporter) insertBatch(ctx context.Context, cards []*storage.Card) error {
	for _, card := range cards {
		if err := bi.storage.SaveCard(ctx, card); err != nil {
			return fmt.Errorf("failed to save card %s: %w", card.Name, err)
		}
	}
	return nil
}

// ConvertToStorageCard converts a Scryfall card to a storage card.
func ConvertToStorageCard(card *scryfall.Card) *storage.Card {
	// Convert Legalities struct to map
	legalities := legalitiesToMap(card.Legalities)

	return &storage.Card{
		ID:              card.ID,
		ArenaID:         card.ArenaID,
		Name:            card.Name,
		ManaCost:        card.ManaCost,
		CMC:             card.CMC,
		TypeLine:        card.TypeLine,
		OracleText:      card.OracleText,
		Colors:          card.Colors,
		ColorIdentity:   card.ColorIdentity,
		Rarity:          card.Rarity,
		SetCode:         card.SetCode,
		CollectorNumber: card.CollectorNumber,
		Power:           card.Power,
		Toughness:       card.Toughness,
		Loyalty:         card.Loyalty,
		ImageURIs:       card.ImageURIs,
		Layout:          card.Layout,
		CardFaces:       card.CardFaces,
		Legalities:      legalities,
		ReleasedAt:      card.ReleasedAt,
	}
}

// legalitiesToMap converts a Legalities struct to a map.
func legalitiesToMap(l scryfall.Legalities) map[string]string {
	return map[string]string{
		"standard":        l.Standard,
		"future":          l.Future,
		"historic":        l.Historic,
		"gladiator":       l.Gladiator,
		"pioneer":         l.Pioneer,
		"explorer":        l.Explorer,
		"modern":          l.Modern,
		"legacy":          l.Legacy,
		"pauper":          l.Pauper,
		"vintage":         l.Vintage,
		"penny":           l.Penny,
		"commander":       l.Commander,
		"oathbreaker":     l.Oathbreaker,
		"brawl":           l.Brawl,
		"historicbrawl":   l.HistoricBrawl,
		"alchemy":         l.Alchemy,
		"paupercommander": l.PauperCommander,
		"duel":            l.Duel,
		"oldschool":       l.OldSchool,
	}
}
