package cfb

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// Importer handles importing ChannelFireball card ratings.
type Importer struct {
	cfbRepo     repository.CFBRatingsRepository
	setCardRepo repository.SetCardRepository
}

// NewImporter creates a new CFB ratings importer.
func NewImporter(cfbRepo repository.CFBRatingsRepository, setCardRepo repository.SetCardRepository) *Importer {
	return &Importer{
		cfbRepo:     cfbRepo,
		setCardRepo: setCardRepo,
	}
}

// ImportFromJSON imports CFB ratings from a JSON file.
// The JSON file should contain an array of CFBRatingImport objects.
func (i *Importer) ImportFromJSON(ctx context.Context, filePath string) (int, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return 0, fmt.Errorf("failed to read file: %w", err)
	}

	var imports []models.CFBRatingImport
	if err := json.Unmarshal(data, &imports); err != nil {
		return 0, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return i.ImportRatings(ctx, imports)
}

// ImportRatings imports a slice of CFB ratings.
func (i *Importer) ImportRatings(ctx context.Context, imports []models.CFBRatingImport) (int, error) {
	if len(imports) == 0 {
		return 0, nil
	}

	log.Printf("[CFBImporter] Importing %d ratings", len(imports))

	imported := 0
	for _, imp := range imports {
		rating := &models.CFBRating{
			CardName:          imp.CardName,
			SetCode:           strings.ToUpper(imp.SetCode),
			LimitedRating:     imp.LimitedRating,
			LimitedScore:      models.LimitedGradeToScore(imp.LimitedRating),
			ConstructedRating: imp.ConstructedRating,
			ConstructedScore:  models.ConstructedRatingToScore(imp.ConstructedRating),
			ArchetypeFit:      imp.ArchetypeFit,
			Commentary:        imp.Commentary,
			SourceURL:         imp.SourceURL,
			Author:            imp.Author,
			ImportedAt:        time.Now(),
			UpdatedAt:         time.Now(),
		}

		if err := i.cfbRepo.UpsertRating(ctx, rating); err != nil {
			log.Printf("[CFBImporter] Warning: failed to import rating for %s: %v", imp.CardName, err)
			continue
		}
		imported++
	}

	log.Printf("[CFBImporter] Successfully imported %d/%d ratings", imported, len(imports))

	return imported, nil
}

// LinkArenaIDs links CFB ratings to Arena IDs by matching card names.
// This should be called after importing ratings for a set.
func (i *Importer) LinkArenaIDs(ctx context.Context, setCode string) (int, error) {
	setCode = strings.ToUpper(setCode)

	// Get all set cards for this set
	cards, err := i.setCardRepo.GetCardsBySet(ctx, setCode)
	if err != nil {
		return 0, fmt.Errorf("failed to get set cards: %w", err)
	}

	if len(cards) == 0 {
		log.Printf("[CFBImporter] No set cards found for %s - ratings won't be linked to Arena IDs", setCode)
		return 0, nil
	}

	// Build name to Arena ID map
	cardNameToArenaID := make(map[string]int)
	for _, card := range cards {
		// Normalize card name for matching (lowercase, trim whitespace)
		normalizedName := strings.TrimSpace(strings.ToLower(card.Name))
		// Convert arena_id string to int
		var arenaID int
		if _, err := fmt.Sscanf(card.ArenaID, "%d", &arenaID); err == nil {
			cardNameToArenaID[normalizedName] = arenaID
			// Also add the original case version
			cardNameToArenaID[strings.TrimSpace(card.Name)] = arenaID
		}
	}

	// Get CFB ratings for this set
	ratings, err := i.cfbRepo.GetRatingsForSet(ctx, setCode)
	if err != nil {
		return 0, fmt.Errorf("failed to get CFB ratings: %w", err)
	}

	linked := 0
	for _, rating := range ratings {
		// Skip if already linked
		if rating.ArenaID != nil {
			continue
		}

		// Try to find matching Arena ID
		normalizedName := strings.TrimSpace(strings.ToLower(rating.CardName))
		if arenaID, found := cardNameToArenaID[normalizedName]; found {
			rating.ArenaID = &arenaID
			if err := i.cfbRepo.UpsertRating(ctx, rating); err != nil {
				log.Printf("[CFBImporter] Warning: failed to link Arena ID for %s: %v", rating.CardName, err)
				continue
			}
			linked++
		} else if arenaID, found := cardNameToArenaID[rating.CardName]; found {
			rating.ArenaID = &arenaID
			if err := i.cfbRepo.UpsertRating(ctx, rating); err != nil {
				log.Printf("[CFBImporter] Warning: failed to link Arena ID for %s: %v", rating.CardName, err)
				continue
			}
			linked++
		}
	}

	log.Printf("[CFBImporter] Linked %d/%d ratings to Arena IDs for set %s", linked, len(ratings), setCode)

	return linked, nil
}

// GetRatingsForSet returns all CFB ratings for a set.
func (i *Importer) GetRatingsForSet(ctx context.Context, setCode string) ([]*models.CFBRating, error) {
	return i.cfbRepo.GetRatingsForSet(ctx, strings.ToUpper(setCode))
}

// GetRatingByCardName returns a CFB rating by card name and set code.
func (i *Importer) GetRatingByCardName(ctx context.Context, cardName, setCode string) (*models.CFBRating, error) {
	return i.cfbRepo.GetRating(ctx, cardName, strings.ToUpper(setCode))
}

// GetRatingByArenaID returns a CFB rating by Arena ID.
func (i *Importer) GetRatingByArenaID(ctx context.Context, arenaID int) (*models.CFBRating, error) {
	return i.cfbRepo.GetRatingByArenaID(ctx, arenaID)
}

// GetRatingsCount returns the number of CFB ratings for a set.
func (i *Importer) GetRatingsCount(ctx context.Context, setCode string) (int, error) {
	return i.cfbRepo.GetRatingsCount(ctx, strings.ToUpper(setCode))
}

// DeleteRatingsForSet deletes all CFB ratings for a set.
func (i *Importer) DeleteRatingsForSet(ctx context.Context, setCode string) error {
	return i.cfbRepo.DeleteRatingsForSet(ctx, strings.ToUpper(setCode))
}
