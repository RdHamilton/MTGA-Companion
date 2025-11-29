package ml

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/seventeenlands"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// TrainingPipeline handles ML model training from 17Lands data and feedback.
type TrainingPipeline struct {
	model           *Model
	ratingsRepo     repository.DraftRatingsRepository
	feedbackRepo    repository.RecommendationFeedbackRepository
	performanceRepo repository.DeckPerformanceRepository
	dataDir         string
	config          *PipelineConfig
	mu              sync.Mutex

	// Progress tracking
	progress     *TrainingProgress
	progressChan chan *TrainingProgress
}

// PipelineConfig configures the training pipeline.
type PipelineConfig struct {
	// MinGamesThreshold is the minimum games for a card to be included in training.
	MinGamesThreshold int

	// DataValidationEnabled enables data quality checks.
	DataValidationEnabled bool

	// IncrementalTraining enables incremental updates vs full retraining.
	IncrementalTraining bool

	// MaxCardAge is the maximum age of card ratings to use (in days).
	MaxCardAge int

	// BatchSize is the number of records to process at once.
	BatchSize int

	// ParallelWorkers is the number of parallel processing workers.
	ParallelWorkers int

	// CheckpointInterval is how often to save checkpoints (in records).
	CheckpointInterval int

	// CheckpointDir is the directory for saving checkpoints.
	CheckpointDir string
}

// DefaultPipelineConfig returns default configuration.
func DefaultPipelineConfig() *PipelineConfig {
	return &PipelineConfig{
		MinGamesThreshold:     100,
		DataValidationEnabled: true,
		IncrementalTraining:   true,
		MaxCardAge:            90,
		BatchSize:             1000,
		ParallelWorkers:       4,
		CheckpointInterval:    5000,
		CheckpointDir:         "",
	}
}

// TrainingProgress tracks the progress of a training run.
type TrainingProgress struct {
	Stage           string    `json:"stage"`
	CurrentStep     int       `json:"current_step"`
	TotalSteps      int       `json:"total_steps"`
	Percent         float64   `json:"percent"`
	CardsProcessed  int       `json:"cards_processed"`
	SetsProcessed   int       `json:"sets_processed"`
	ErrorCount      int       `json:"error_count"`
	StartTime       time.Time `json:"start_time"`
	EstimatedTimeMs int64     `json:"estimated_time_ms"`
	Complete        bool      `json:"complete"`
	Failed          bool      `json:"failed"`
	Error           string    `json:"error,omitempty"`
}

// TrainingMetrics contains metrics from a training run.
type TrainingMetrics struct {
	TotalCards          int       `json:"total_cards"`
	TotalSets           int       `json:"total_sets"`
	TotalFeedback       int       `json:"total_feedback"`
	TotalPerformance    int       `json:"total_performance"`
	SkippedLowGames     int       `json:"skipped_low_games"`
	SkippedInvalid      int       `json:"skipped_invalid"`
	ProcessingTimeMs    int64     `json:"processing_time_ms"`
	DataQualityScore    float64   `json:"data_quality_score"`
	TrainedAt           time.Time `json:"trained_at"`
	ModelVersion        string    `json:"model_version"`
	ArchetypesLearned   int       `json:"archetypes_learned"`
	CardFeaturesLearned int       `json:"card_features_learned"`
}

// CardTrainingData represents processed card data for training.
type CardTrainingData struct {
	CardID   int
	ArenaID  string
	Name     string
	SetCode  string
	Colors   []string
	CMC      float64
	Rarity   string
	Types    []string
	Keywords []string

	// Performance metrics from 17Lands
	GIHWR       float64
	OHWR        float64
	ATA         float64
	ALSA        float64
	GamesPlayed int

	// Normalized scores (0.0-1.0)
	QualityScore float64
	PickScore    float64
	WinRateScore float64

	// Archetype affinities
	ArchetypeAffinities map[string]float64
}

// ArchetypeTrainingData represents archetype patterns for training.
type ArchetypeTrainingData struct {
	Name          string
	SetCode       string
	Format        string
	ColorIdentity string
	WinRate       float64
	Popularity    float64 // Based on games played
	KeyCards      []int   // Card IDs of signature cards
	CardWeights   map[int]float64
}

// NewTrainingPipeline creates a new training pipeline.
func NewTrainingPipeline(
	model *Model,
	ratingsRepo repository.DraftRatingsRepository,
	feedbackRepo repository.RecommendationFeedbackRepository,
	performanceRepo repository.DeckPerformanceRepository,
	config *PipelineConfig,
) *TrainingPipeline {
	if config == nil {
		config = DefaultPipelineConfig()
	}

	return &TrainingPipeline{
		model:           model,
		ratingsRepo:     ratingsRepo,
		feedbackRepo:    feedbackRepo,
		performanceRepo: performanceRepo,
		config:          config,
		progress: &TrainingProgress{
			Stage: "idle",
		},
	}
}

// SetDataDir sets the directory for set file data.
func (p *TrainingPipeline) SetDataDir(dir string) {
	p.dataDir = dir
}

// SetProgressChannel sets a channel to receive progress updates.
func (p *TrainingPipeline) SetProgressChannel(ch chan *TrainingProgress) {
	p.progressChan = ch
}

// Train runs the full training pipeline.
func (p *TrainingPipeline) Train(ctx context.Context) (*TrainingMetrics, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	metrics := &TrainingMetrics{
		TrainedAt: time.Now(),
	}

	startTime := time.Now()

	// Stage 1: Extract 17Lands data
	p.updateProgress("extracting_17lands", 1, 5, 0)

	cardData, err := p.extract17LandsData(ctx)
	if err != nil {
		p.failProgress(fmt.Sprintf("failed to extract 17Lands data: %v", err))
		return nil, fmt.Errorf("extraction failed: %w", err)
	}
	metrics.TotalCards = len(cardData)

	// Stage 2: Load feedback data
	p.updateProgress("loading_feedback", 2, 5, 20)

	feedbackData, err := p.loadFeedbackData(ctx)
	if err != nil {
		p.failProgress(fmt.Sprintf("failed to load feedback: %v", err))
		return nil, fmt.Errorf("feedback loading failed: %w", err)
	}
	metrics.TotalFeedback = len(feedbackData)

	// Stage 3: Load performance data
	p.updateProgress("loading_performance", 3, 5, 40)

	performanceData, err := p.loadPerformanceData(ctx)
	if err != nil {
		p.failProgress(fmt.Sprintf("failed to load performance: %v", err))
		return nil, fmt.Errorf("performance loading failed: %w", err)
	}
	metrics.TotalPerformance = len(performanceData)

	// Stage 4: Transform and validate
	p.updateProgress("transforming", 4, 5, 60)

	transformedCards, skippedLow, skippedInvalid := p.transformData(cardData)
	metrics.SkippedLowGames = skippedLow
	metrics.SkippedInvalid = skippedInvalid

	archetypes := p.extractArchetypes(ctx, performanceData)
	metrics.ArchetypesLearned = len(archetypes)

	// Stage 5: Train model
	p.updateProgress("training", 5, 5, 80)

	if err := p.trainModel(ctx, transformedCards, archetypes, feedbackData); err != nil {
		p.failProgress(fmt.Sprintf("model training failed: %v", err))
		return nil, fmt.Errorf("training failed: %w", err)
	}

	// Calculate data quality
	metrics.DataQualityScore = p.calculateDataQuality(transformedCards, feedbackData)
	metrics.CardFeaturesLearned = len(transformedCards)
	metrics.ProcessingTimeMs = time.Since(startTime).Milliseconds()

	if p.model != nil {
		info := p.model.GetModelInfo()
		metrics.ModelVersion = info.Version
	}

	p.completeProgress()

	return metrics, nil
}

// TrainIncremental performs incremental training with new data.
func (p *TrainingPipeline) TrainIncremental(ctx context.Context, since time.Time) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Load only new feedback since last training
	feedbackData, err := p.feedbackRepo.GetForMLTraining(ctx, 10000)
	if err != nil {
		return fmt.Errorf("failed to load feedback: %w", err)
	}

	// Filter to feedback after 'since'
	newFeedback := make([]*models.RecommendationFeedback, 0)
	for _, fb := range feedbackData {
		if fb.CreatedAt.After(since) {
			newFeedback = append(newFeedback, fb)
		}
	}

	if len(newFeedback) == 0 {
		return nil // Nothing to train on
	}

	// Update model with new feedback
	for _, fb := range newFeedback {
		if err := p.model.UpdateFromFeedback(ctx, fb); err != nil {
			// Log but continue
			continue
		}
	}

	return nil
}

// extract17LandsData extracts training data from 17Lands ratings.
func (p *TrainingPipeline) extract17LandsData(ctx context.Context) ([]*CardTrainingData, error) {
	cards := make([]*CardTrainingData, 0)

	// Get available sets from data directory if specified
	if p.dataDir != "" {
		setFiles, err := p.loadSetFiles(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load set files: %w", err)
		}

		for _, sf := range setFiles {
			setCards := p.extractCardsFromSetFile(sf)
			cards = append(cards, setCards...)
		}
	}

	// Also load from ratings repository
	if p.ratingsRepo != nil {
		// Get all snapshots to find available sets
		snapshots, err := p.ratingsRepo.GetAllSnapshots(ctx)
		if err == nil && len(snapshots) > 0 {
			// Group by expansion
			setMap := make(map[string]bool)
			for _, s := range snapshots {
				setMap[s.Expansion] = true
			}

			for setCode := range setMap {
				ratings, _, err := p.ratingsRepo.GetCardRatings(ctx, setCode, "PremierDraft")
				if err != nil {
					continue
				}

				for _, r := range ratings {
					card := p.convertRatingToTrainingData(r, setCode)
					if card != nil {
						cards = append(cards, card)
					}
				}
			}
		}
	}

	return cards, nil
}

// loadSetFiles loads set files from the data directory.
func (p *TrainingPipeline) loadSetFiles(ctx context.Context) ([]*seventeenlands.SetFile, error) {
	if p.dataDir == "" {
		return nil, nil
	}

	files := make([]*seventeenlands.SetFile, 0)

	// Look for JSON files in data directory
	pattern := filepath.Join(p.dataDir, "*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to glob set files: %w", err)
	}

	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var sf seventeenlands.SetFile
		if err := json.Unmarshal(data, &sf); err != nil {
			continue
		}

		// Validate it's a proper set file
		if sf.Meta.SetCode != "" {
			files = append(files, &sf)
		}
	}

	return files, nil
}

// extractCardsFromSetFile extracts training data from a set file.
func (p *TrainingPipeline) extractCardsFromSetFile(sf *seventeenlands.SetFile) []*CardTrainingData {
	cards := make([]*CardTrainingData, 0)

	for arenaID, cardData := range sf.CardRatings {
		// Get "ALL" deck color ratings if available
		var ratings *seventeenlands.DeckColorRatings
		if cardData.DeckColors != nil {
			if allRatings, ok := cardData.DeckColors["ALL"]; ok {
				ratings = allRatings
			}
		}

		if ratings == nil {
			continue
		}

		// Skip cards with too few games
		if ratings.GIH < p.config.MinGamesThreshold {
			continue
		}

		card := &CardTrainingData{
			ArenaID:             arenaID,
			Name:                cardData.Name,
			SetCode:             sf.Meta.SetCode,
			Colors:              cardData.Colors,
			CMC:                 cardData.CMC,
			Rarity:              cardData.Rarity,
			Types:               cardData.Types,
			GIHWR:               ratings.GIHWR,
			OHWR:                ratings.OHWR,
			ATA:                 ratings.ATA,
			ALSA:                ratings.ALSA,
			GamesPlayed:         ratings.GIH,
			ArchetypeAffinities: make(map[string]float64),
		}

		// Calculate normalized scores
		card.QualityScore = p.normalizeWinRate(ratings.GIHWR)
		card.PickScore = p.normalizePickPosition(ratings.ATA)
		card.WinRateScore = p.normalizeWinRate(ratings.GIHWR)

		// Calculate archetype affinities from color ratings
		if cardData.DeckColors != nil {
			for colorCombo, colorRatings := range cardData.DeckColors {
				if colorCombo == "ALL" {
					continue
				}
				// Calculate affinity as relative performance vs ALL
				if ratings.GIHWR > 0 {
					affinity := colorRatings.GIHWR / ratings.GIHWR
					if affinity > 1.0 {
						affinity = 1.0 + (affinity-1.0)*0.5 // Dampen high affinities
					}
					card.ArchetypeAffinities[colorCombo] = affinity
				}
			}
		}

		cards = append(cards, card)
	}

	return cards
}

// convertRatingToTrainingData converts a 17Lands CardRating to training data.
func (p *TrainingPipeline) convertRatingToTrainingData(r seventeenlands.CardRating, setCode string) *CardTrainingData {
	// Skip cards with too few games
	if r.GIH < p.config.MinGamesThreshold {
		return nil
	}

	card := &CardTrainingData{
		ArenaID:             fmt.Sprintf("%d", r.MTGAID),
		Name:                r.Name,
		SetCode:             setCode,
		Colors:              []string{r.Color},
		Rarity:              r.Rarity,
		GIHWR:               r.GIHWR,
		OHWR:                r.OHWR,
		ATA:                 r.ATA,
		ALSA:                r.ALSA,
		GamesPlayed:         r.GIH,
		ArchetypeAffinities: make(map[string]float64),
	}

	// Calculate normalized scores
	card.QualityScore = p.normalizeWinRate(r.GIHWR)
	card.PickScore = p.normalizePickPosition(r.ATA)
	card.WinRateScore = p.normalizeWinRate(r.GIHWR)

	return card
}

// loadFeedbackData loads user feedback for training.
func (p *TrainingPipeline) loadFeedbackData(ctx context.Context) ([]*models.RecommendationFeedback, error) {
	if p.feedbackRepo == nil {
		return nil, nil
	}

	return p.feedbackRepo.GetForMLTraining(ctx, 50000)
}

// loadPerformanceData loads deck performance history for training.
func (p *TrainingPipeline) loadPerformanceData(ctx context.Context) ([]*models.DeckPerformanceHistory, error) {
	if p.performanceRepo == nil {
		return nil, nil
	}

	// Get performance data from the last year
	startDate := time.Now().AddDate(-1, 0, 0)
	endDate := time.Now()

	return p.performanceRepo.GetPerformanceByDateRange(ctx, 0, startDate, endDate)
}

// transformData transforms and validates card data.
func (p *TrainingPipeline) transformData(cards []*CardTrainingData) ([]*CardTrainingData, int, int) {
	valid := make([]*CardTrainingData, 0, len(cards))
	skippedLowGames := 0
	skippedInvalid := 0

	for _, card := range cards {
		// Validate
		if !p.isValidCard(card) {
			skippedInvalid++
			continue
		}

		// Check minimum games threshold
		if card.GamesPlayed < p.config.MinGamesThreshold {
			skippedLowGames++
			continue
		}

		// Extract keywords from types
		card.Keywords = p.extractKeywordsFromTypes(card.Types)

		valid = append(valid, card)
	}

	return valid, skippedLowGames, skippedInvalid
}

// isValidCard checks if card data is valid for training.
func (p *TrainingPipeline) isValidCard(card *CardTrainingData) bool {
	if card == nil {
		return false
	}

	// Must have name
	if card.Name == "" {
		return false
	}

	// Win rate must be valid (0-100%)
	if card.GIHWR < 0 || card.GIHWR > 100 {
		return false
	}

	// ATA must be valid (1-15 typically)
	if card.ATA < 0 || card.ATA > 20 {
		return false
	}

	return true
}

// extractKeywordsFromTypes extracts keyword-like features from card types.
func (p *TrainingPipeline) extractKeywordsFromTypes(types []string) []string {
	keywords := make([]string, 0)

	// Common creature types that indicate synergies
	creatureTypes := map[string]bool{
		"Human": true, "Elf": true, "Goblin": true, "Zombie": true,
		"Vampire": true, "Angel": true, "Dragon": true, "Beast": true,
		"Wizard": true, "Warrior": true, "Rogue": true, "Cleric": true,
		"Elemental": true, "Spirit": true, "Merfolk": true, "Soldier": true,
	}

	for _, t := range types {
		if creatureTypes[t] {
			keywords = append(keywords, t)
		}
	}

	return keywords
}

// extractArchetypes extracts archetype patterns from performance data.
func (p *TrainingPipeline) extractArchetypes(ctx context.Context, perfData []*models.DeckPerformanceHistory) []*ArchetypeTrainingData {
	archetypes := make(map[string]*ArchetypeTrainingData)

	// Group by archetype
	for _, perf := range perfData {
		if perf.Archetype == nil || *perf.Archetype == "" {
			continue
		}

		arch := *perf.Archetype
		if _, exists := archetypes[arch]; !exists {
			archetypes[arch] = &ArchetypeTrainingData{
				Name:          arch,
				Format:        perf.Format,
				ColorIdentity: perf.ColorIdentity,
				CardWeights:   make(map[int]float64),
			}
		}

		// Update win rate
		data := archetypes[arch]
		if perf.Result == "win" {
			data.WinRate = (data.WinRate*data.Popularity + 1.0) / (data.Popularity + 1)
		} else {
			data.WinRate = (data.WinRate * data.Popularity) / (data.Popularity + 1)
		}
		data.Popularity++
	}

	// Convert to slice and load card weights from repository
	result := make([]*ArchetypeTrainingData, 0, len(archetypes))
	for _, arch := range archetypes {
		// Load card weights from performance repo if available
		if p.performanceRepo != nil {
			dbArch, err := p.performanceRepo.GetArchetypeByName(ctx, arch.Name, nil, arch.Format)
			if err == nil && dbArch != nil {
				weights, err := p.performanceRepo.GetCardWeights(ctx, dbArch.ID)
				if err == nil {
					for _, w := range weights {
						arch.CardWeights[w.CardID] = w.Weight / 10.0 // Normalize to 0-1
						if w.IsSignature {
							arch.KeyCards = append(arch.KeyCards, w.CardID)
						}
					}
				}
			}
		}

		result = append(result, arch)
	}

	// Sort by popularity
	sort.Slice(result, func(i, j int) bool {
		return result[i].Popularity > result[j].Popularity
	})

	return result
}

// trainModel trains the ML model with processed data.
func (p *TrainingPipeline) trainModel(
	ctx context.Context,
	cards []*CardTrainingData,
	archetypes []*ArchetypeTrainingData,
	feedback []*models.RecommendationFeedback,
) error {
	if p.model == nil {
		return fmt.Errorf("model not initialized")
	}

	// Register card features
	for _, card := range cards {
		arenaID := 0
		_, _ = fmt.Sscanf(card.ArenaID, "%d", &arenaID)

		features := &CardFeatures{
			CardID:         arenaID,
			ArenaID:        card.ArenaID,
			Name:           card.Name,
			CMC:            card.CMC,
			Colors:         card.Colors,
			Types:          card.Types,
			Keywords:       card.Keywords,
			Rarity:         card.Rarity,
			SetCode:        card.SetCode,
			ColorCount:     len(card.Colors),
			IsCreature:     containsType(card.Types, "Creature"),
			IsInstant:      containsType(card.Types, "Instant"),
			IsSorcery:      containsType(card.Types, "Sorcery"),
			IsEnchantment:  containsType(card.Types, "Enchantment"),
			IsArtifact:     containsType(card.Types, "Artifact"),
			IsLand:         containsType(card.Types, "Land"),
			IsPlaneswalker: containsType(card.Types, "Planeswalker"),
		}

		p.model.RegisterCardFeatures(arenaID, features)
	}

	// Update archetype affinities
	p.model.affinityMu.Lock()
	for _, arch := range archetypes {
		if len(arch.CardWeights) > 0 {
			p.model.archetypeAffinities[arch.Name] = arch.CardWeights
		}
	}
	p.model.affinityMu.Unlock()

	// Process feedback to update acceptance rates
	for _, fb := range feedback {
		if err := p.model.UpdateFromFeedback(ctx, fb); err != nil {
			// Log but continue
			continue
		}
	}

	// Update model metadata
	p.model.mu.Lock()
	p.model.trainingSamples = len(feedback)
	p.model.lastTrainedAt = time.Now()
	p.model.mu.Unlock()

	return nil
}

// calculateDataQuality calculates a data quality score.
func (p *TrainingPipeline) calculateDataQuality(cards []*CardTrainingData, feedback []*models.RecommendationFeedback) float64 {
	score := 0.0
	factors := 0

	// Factor 1: Card count
	if len(cards) > 1000 {
		score += 1.0
	} else if len(cards) > 500 {
		score += 0.7
	} else if len(cards) > 100 {
		score += 0.4
	}
	factors++

	// Factor 2: Feedback count
	if len(feedback) > 1000 {
		score += 1.0
	} else if len(feedback) > 500 {
		score += 0.7
	} else if len(feedback) > 100 {
		score += 0.4
	}
	factors++

	// Factor 3: Win rate variance (higher variance = better signal)
	if len(cards) > 0 {
		sumWR := 0.0
		for _, c := range cards {
			sumWR += c.GIHWR
		}
		avgWR := sumWR / float64(len(cards))

		sumVar := 0.0
		for _, c := range cards {
			diff := c.GIHWR - avgWR
			sumVar += diff * diff
		}
		variance := sumVar / float64(len(cards))

		// Good variance is around 5-10 percentage points
		if variance > 25 && variance < 100 {
			score += 1.0
		} else if variance > 9 {
			score += 0.7
		} else {
			score += 0.3
		}
		factors++
	}

	if factors == 0 {
		return 0.5
	}

	return score / float64(factors)
}

// normalizeWinRate normalizes a win rate (45-65%) to 0.0-1.0.
func (p *TrainingPipeline) normalizeWinRate(wr float64) float64 {
	// Map 45% -> 0.0, 55% -> 0.5, 65% -> 1.0
	normalized := (wr - 45) / 20
	if normalized < 0 {
		return 0
	}
	if normalized > 1 {
		return 1
	}
	return normalized
}

// normalizePickPosition normalizes ATA (1-14) to 0.0-1.0 (lower is better).
func (p *TrainingPipeline) normalizePickPosition(ata float64) float64 {
	// Map 1 -> 1.0, 7 -> 0.5, 14 -> 0.0
	normalized := 1.0 - ((ata - 1) / 13)
	if normalized < 0 {
		return 0
	}
	if normalized > 1 {
		return 1
	}
	return normalized
}

// updateProgress updates the progress status.
func (p *TrainingPipeline) updateProgress(stage string, step, total int, percent float64) {
	p.progress = &TrainingProgress{
		Stage:       stage,
		CurrentStep: step,
		TotalSteps:  total,
		Percent:     percent,
		StartTime:   time.Now(),
	}

	if p.progressChan != nil {
		select {
		case p.progressChan <- p.progress:
		default:
		}
	}
}

// failProgress marks progress as failed.
func (p *TrainingPipeline) failProgress(errMsg string) {
	p.progress.Failed = true
	p.progress.Error = errMsg

	if p.progressChan != nil {
		select {
		case p.progressChan <- p.progress:
		default:
		}
	}
}

// completeProgress marks progress as complete.
func (p *TrainingPipeline) completeProgress() {
	p.progress.Complete = true
	p.progress.Percent = 100

	if p.progressChan != nil {
		select {
		case p.progressChan <- p.progress:
		default:
		}
	}
}

// GetProgress returns the current progress.
func (p *TrainingPipeline) GetProgress() *TrainingProgress {
	return p.progress
}

// SaveCheckpoint saves a training checkpoint.
func (p *TrainingPipeline) SaveCheckpoint(ctx context.Context) error {
	if p.config.CheckpointDir == "" || p.model == nil {
		return nil
	}

	data, err := p.model.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize model: %w", err)
	}

	filename := fmt.Sprintf("checkpoint_%s.json", time.Now().Format("20060102_150405"))
	path := filepath.Join(p.config.CheckpointDir, filename)

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write checkpoint: %w", err)
	}

	return nil
}

// LoadCheckpoint loads a model checkpoint.
func (p *TrainingPipeline) LoadCheckpoint(ctx context.Context, path string) error {
	if p.model == nil {
		return fmt.Errorf("model not initialized")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read checkpoint: %w", err)
	}

	if err := p.model.Deserialize(data); err != nil {
		return fmt.Errorf("failed to deserialize model: %w", err)
	}

	return nil
}

// Helper function
func containsType(types []string, target string) bool {
	for _, t := range types {
		if t == target {
			return true
		}
	}
	return false
}
