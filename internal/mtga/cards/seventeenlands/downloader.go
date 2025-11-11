package seventeenlands

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
)

// ColorCombinations defines all deck color combinations to fetch ratings for.
var ColorCombinations = []string{
	"ALL", // Aggregate across all decks
	// Mono-color
	"W", "U", "B", "R", "G",
	// Two-color
	"WU", "WB", "WR", "WG",
	"UB", "UR", "UG",
	"BR", "BG",
	"RG",
	// Three-color (optional - can be enabled based on data availability)
	// "WUB", "WUR", "WUG", "WBR", "WBG", "WRG",
	// "UBR", "UBG", "URG",
	// "BRG",
}

// Downloader handles downloading and combining set file data from multiple sources.
type Downloader struct {
	client          *Client
	scryfallClient  *scryfall.Client
	setsDir         string
	progressHandler func(DownloadProgress)
}

// DownloaderOptions configures the set file downloader.
type DownloaderOptions struct {
	// Client for 17Lands API
	Client *Client

	// ScryfallClient for card metadata
	ScryfallClient *scryfall.Client

	// SetsDir is the directory to save set files (default: "Sets")
	SetsDir string

	// ProgressHandler is called with download progress updates
	ProgressHandler func(DownloadProgress)
}

// NewDownloader creates a new set file downloader.
func NewDownloader(options DownloaderOptions) (*Downloader, error) {
	if options.Client == nil {
		return nil, fmt.Errorf("17Lands client is required")
	}
	if options.ScryfallClient == nil {
		return nil, fmt.Errorf("scryfall client is required")
	}

	setsDir := options.SetsDir
	if setsDir == "" {
		setsDir = "Sets"
	}

	// Ensure sets directory exists
	if err := os.MkdirAll(setsDir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create sets directory: %w", err)
	}

	return &Downloader{
		client:          options.Client,
		scryfallClient:  options.ScryfallClient,
		setsDir:         setsDir,
		progressHandler: options.ProgressHandler,
	}, nil
}

// DownloadSetFile downloads and combines data into a complete set file.
func (d *Downloader) DownloadSetFile(ctx context.Context, params DownloadParams) (*SetFile, error) {
	progress := DownloadProgress{
		TotalSteps: 4,
		StartTime:  time.Now(),
	}

	// Step 1: Fetch card metadata from Scryfall
	progress.CurrentStep = 1
	progress.StepName = "Fetching card metadata"
	progress.Message = fmt.Sprintf("Fetching card metadata for %s from Scryfall...", params.SetCode)
	progress.Percent = 0
	d.updateProgress(progress)

	cardMetadata, err := d.fetchCardMetadata(ctx, params.SetCode)
	if err != nil {
		progress.Failed = true
		progress.Errors = append(progress.Errors, fmt.Sprintf("Failed to fetch card metadata: %v", err))
		d.updateProgress(progress)
		return nil, fmt.Errorf("failed to fetch card metadata: %w", err)
	}

	progress.Message = fmt.Sprintf("Fetched metadata for %d cards", len(cardMetadata))
	progress.Percent = 25
	d.updateProgress(progress)

	// Step 2: Fetch 17Lands card ratings for all color combinations
	progress.CurrentStep = 2
	progress.StepName = "Downloading 17Lands ratings"
	progress.Message = "Fetching card ratings for all color combinations..."
	progress.SubTaskTotal = len(ColorCombinations)
	d.updateProgress(progress)

	cardRatingsMap, totalGames, err := d.fetchCardRatings(ctx, params, &progress)
	if err != nil {
		progress.Failed = true
		progress.Errors = append(progress.Errors, fmt.Sprintf("Failed to fetch card ratings: %v", err))
		d.updateProgress(progress)
		return nil, fmt.Errorf("failed to fetch card ratings: %w", err)
	}

	progress.Message = fmt.Sprintf("Fetched ratings for %d cards", len(cardRatingsMap))
	progress.Percent = 60
	progress.SubTask = ""
	progress.SubTaskCurrent = 0
	progress.SubTaskTotal = 0
	d.updateProgress(progress)

	// Step 3: Fetch color combination win rates
	progress.CurrentStep = 3
	progress.StepName = "Downloading color ratings"
	progress.Message = "Fetching color combination win rates..."
	progress.Percent = 65
	d.updateProgress(progress)

	colorRatings, err := d.fetchColorRatings(ctx, params)
	if err != nil {
		// Non-fatal error - log but continue
		log.Printf("Warning: Failed to fetch color ratings: %v", err)
		progress.Errors = append(progress.Errors, fmt.Sprintf("Failed to fetch color ratings: %v", err))
		colorRatings = make(map[string]float64)
	}

	progress.Message = fmt.Sprintf("Fetched ratings for %d color combinations", len(colorRatings))
	progress.Percent = 80
	d.updateProgress(progress)

	// Step 4: Combine and create set file
	progress.CurrentStep = 4
	progress.StepName = "Combining data"
	progress.Message = "Combining card metadata with ratings..."
	progress.Percent = 85
	d.updateProgress(progress)

	setFile := d.combineData(params, cardMetadata, cardRatingsMap, colorRatings, totalGames)

	progress.Message = fmt.Sprintf("Set file ready with %d cards", len(setFile.CardRatings))
	progress.Percent = 95
	d.updateProgress(progress)

	// Save to disk
	if err := d.SaveSetFile(setFile, params.SetCode, params.Format); err != nil {
		progress.Failed = true
		progress.Errors = append(progress.Errors, fmt.Sprintf("Failed to save set file: %v", err))
		d.updateProgress(progress)
		return nil, fmt.Errorf("failed to save set file: %w", err)
	}

	progress.Complete = true
	progress.Percent = 100
	progress.Message = "Download complete!"
	d.updateProgress(progress)

	return setFile, nil
}

// DownloadParams specifies parameters for downloading a set file.
type DownloadParams struct {
	SetCode   string // e.g., "BLB", "DSK"
	Format    string // e.g., "PremierDraft", "QuickDraft"
	StartDate string // YYYY-MM-DD (optional)
	EndDate   string // YYYY-MM-DD (optional)

	// Options
	Include17Lands  bool // Include 17Lands ratings
	IncludeScryfall bool // Include Scryfall metadata
	DownloadImages  bool // Download card images (not yet implemented)
}

// fetchCardMetadata fetches card metadata from Scryfall.
func (d *Downloader) fetchCardMetadata(ctx context.Context, setCode string) (map[int]*scryfall.Card, error) {
	// Search for all cards in the set using the 'e:' (expansion) query
	// Note: We don't use unique:prints as it's not valid Scryfall syntax
	// Instead we filter by ArenaID presence to get MTGA-relevant cards
	query := fmt.Sprintf("e:%s", setCode)
	result, err := d.scryfallClient.SearchCards(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to search cards from Scryfall: %w", err)
	}

	// Index by Arena ID
	cardMap := make(map[int]*scryfall.Card)
	for i := range result.Data {
		card := &result.Data[i]
		// ArenaID is a pointer, so check if it's not nil and > 0
		if card.ArenaID != nil && *card.ArenaID > 0 {
			cardMap[*card.ArenaID] = card
		}
	}

	// TODO: Handle pagination if HasMore is true
	// For now, we assume one page is sufficient for most sets
	if result.HasMore {
		log.Printf("Warning: More cards available for set %s, but pagination not yet implemented", setCode)
	}

	return cardMap, nil
}

// fetchCardRatings fetches card ratings for all color combinations.
func (d *Downloader) fetchCardRatings(
	ctx context.Context,
	params DownloadParams,
	progress *DownloadProgress,
) (map[int]map[string]*DeckColorRatings, int, error) {
	// Map: ArenaID -> ColorCombo -> Ratings
	ratingsMap := make(map[int]map[string]*DeckColorRatings)
	totalGames := 0

	for i, colorCombo := range ColorCombinations {
		progress.SubTask = colorCombo
		progress.SubTaskCurrent = i + 1
		progress.SubTaskPercent = float64(i) / float64(len(ColorCombinations)) * 100
		progress.Message = fmt.Sprintf("Fetching ratings for %s decks...", colorCombo)
		progress.Percent = 25 + (35 * float64(i) / float64(len(ColorCombinations)))
		d.updateProgress(*progress)

		// Prepare query
		queryParams := QueryParams{
			Expansion: params.SetCode,
			Format:    params.Format,
			StartDate: params.StartDate,
			EndDate:   params.EndDate,
		}

		// Set colors filter (empty for ALL)
		if colorCombo != "ALL" {
			queryParams.Colors = []string{colorCombo}
		}

		// Fetch ratings
		ratings, err := d.client.GetCardRatings(ctx, queryParams)
		if err != nil {
			// Log error but continue with other color combinations
			log.Printf("Warning: Failed to fetch ratings for %s: %v", colorCombo, err)
			progress.Errors = append(progress.Errors, fmt.Sprintf("Failed to fetch %s: %v", colorCombo, err))
			progress.FailedTasks = append(progress.FailedTasks, colorCombo)
			continue
		}

		// Process ratings
		for _, rating := range ratings {
			if rating.MTGAID == 0 {
				continue // Skip cards without Arena ID
			}

			// Initialize card entry if needed
			if ratingsMap[rating.MTGAID] == nil {
				ratingsMap[rating.MTGAID] = make(map[string]*DeckColorRatings)
			}

			// Store ratings for this color combination
			ratingsMap[rating.MTGAID][colorCombo] = &DeckColorRatings{
				GIHWR:       rating.GIHWR,
				OHWR:        rating.OHWR,
				GPWR:        rating.GPWR,
				GDWR:        rating.GDWR,
				GIH:         rating.GIH,
				OH:          rating.OH,
				GP:          rating.GP,
				GD:          rating.GD,
				ALSA:        rating.ALSA,
				ATA:         rating.ATA,
				IWD:         rating.GIHWRDelta, // Improvement when drawn
				PickRate:    rating.PickRate,
				GamesPlayed: rating.GamesPlayed,
				NumberDecks: rating.NumberDecks,
			}

			// Track total games (use ALL aggregate)
			if colorCombo == "ALL" && rating.GIH > totalGames {
				totalGames = rating.GIH
			}
		}

		log.Printf("Fetched %d card ratings for %s", len(ratings), colorCombo)
	}

	return ratingsMap, totalGames, nil
}

// fetchColorRatings fetches color combination win rates.
func (d *Downloader) fetchColorRatings(ctx context.Context, params DownloadParams) (map[string]float64, error) {
	queryParams := QueryParams{
		Expansion:     params.SetCode,
		EventType:     params.Format,
		StartDate:     params.StartDate,
		EndDate:       params.EndDate,
		CombineSplash: false,
	}

	ratings, err := d.client.GetColorRatings(ctx, queryParams)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch color ratings: %w", err)
	}

	// Convert to map
	colorMap := make(map[string]float64)
	for _, rating := range ratings {
		colorMap[rating.ColorName] = rating.WinRate
	}

	return colorMap, nil
}

// combineData combines card metadata with ratings into a unified SetFile.
func (d *Downloader) combineData(
	params DownloadParams,
	cardMetadata map[int]*scryfall.Card,
	ratingsMap map[int]map[string]*DeckColorRatings,
	colorRatings map[string]float64,
	totalGames int,
) *SetFile {
	// Create set file
	setFile := &SetFile{
		Meta: SetMeta{
			Version:               2,
			SetCode:               params.SetCode,
			DraftFormat:           params.Format,
			StartDate:             params.StartDate,
			EndDate:               params.EndDate,
			CollectionDate:        time.Now(),
			SeventeenLandsEnabled: params.Include17Lands,
			ScryfallEnabled:       params.IncludeScryfall,
			TotalCards:            len(cardMetadata),
			TotalGames:            totalGames,
		},
		CardRatings:  make(map[string]*CardRatingData),
		ColorRatings: colorRatings,
	}

	// Combine card metadata with ratings
	for arenaID, card := range cardMetadata {
		// Extract images
		var images []string
		if card.ImageURIs != nil && card.ImageURIs.Normal != "" {
			images = append(images, card.ImageURIs.Normal)
		}

		// Create card rating data
		cardData := &CardRatingData{
			Name:       card.Name,
			ManaCost:   card.ManaCost,
			CMC:        card.CMC,
			Types:      parseTypes(card.TypeLine),
			Colors:     card.Colors,
			Rarity:     card.Rarity,
			Images:     images,
			OracleID:   card.OracleID,
			ScryfallID: card.ID,
			ArenaID:    arenaID,
		}

		// Add 17Lands ratings if available
		if ratings, ok := ratingsMap[arenaID]; ok {
			cardData.DeckColors = ratings
		}

		// Store by Arena ID (as string for JSON)
		setFile.CardRatings[fmt.Sprintf("%d", arenaID)] = cardData
	}

	return setFile
}

// parseTypes extracts card types from type line.
func parseTypes(typeLine string) []string {
	// Simple implementation - could be enhanced
	// Example: "Creature — Elf Warrior" -> ["Creature", "Elf", "Warrior"]
	// For now, just return the full type line as a single entry
	if typeLine == "" {
		return []string{}
	}
	return []string{typeLine}
}

// SaveSetFile saves a set file to disk.
func (d *Downloader) SaveSetFile(setFile *SetFile, setCode, format string) error {
	filename := fmt.Sprintf("%s_%s_data.json", setCode, format)
	filepath := filepath.Join(d.setsDir, filename)

	data, err := json.MarshalIndent(setFile, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal set file: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write set file: %w", err)
	}

	log.Printf("Saved set file to %s (%d bytes)", filepath, len(data))
	return nil
}

// LoadSetFile loads a set file from disk.
func (d *Downloader) LoadSetFile(setCode, format string) (*SetFile, error) {
	filename := fmt.Sprintf("%s_%s_data.json", setCode, format)
	filepath := filepath.Join(d.setsDir, filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read set file: %w", err)
	}

	var setFile SetFile
	if err := json.Unmarshal(data, &setFile); err != nil {
		return nil, fmt.Errorf("failed to unmarshal set file: %w", err)
	}

	return &setFile, nil
}

// ListSetFiles returns information about all available set files.
func (d *Downloader) ListSetFiles() ([]SetFileInfo, error) {
	entries, err := os.ReadDir(d.setsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []SetFileInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read sets directory: %w", err)
	}

	var setFiles []SetFileInfo
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filepath := filepath.Join(d.setsDir, entry.Name())

		// Read file to extract metadata
		data, err := os.ReadFile(filepath)
		if err != nil {
			log.Printf("Warning: Failed to read %s: %v", entry.Name(), err)
			continue
		}

		var setFile SetFile
		if err := json.Unmarshal(data, &setFile); err != nil {
			log.Printf("Warning: Failed to unmarshal %s: %v", entry.Name(), err)
			continue
		}

		info, err := entry.Info()
		if err != nil {
			log.Printf("Warning: Failed to get file info for %s: %v", entry.Name(), err)
			continue
		}

		setFiles = append(setFiles, SetFileInfo{
			FilePath:       filepath,
			SetCode:        setFile.Meta.SetCode,
			DraftFormat:    setFile.Meta.DraftFormat,
			StartDate:      setFile.Meta.StartDate,
			EndDate:        setFile.Meta.EndDate,
			CollectionDate: setFile.Meta.CollectionDate,
			TotalCards:     setFile.Meta.TotalCards,
			FileSize:       info.Size(),
			Enabled:        true,  // TODO: Track enabled state
			Active:         false, // TODO: Track active state
		})
	}

	return setFiles, nil
}

// DeleteSetFile deletes a set file from disk.
func (d *Downloader) DeleteSetFile(setCode, format string) error {
	filename := fmt.Sprintf("%s_%s_data.json", setCode, format)
	filepath := filepath.Join(d.setsDir, filename)

	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("failed to delete set file: %w", err)
	}

	return nil
}

// updateProgress calls the progress handler if configured.
func (d *Downloader) updateProgress(progress DownloadProgress) {
	if d.progressHandler != nil {
		d.progressHandler(progress)
	}
}
