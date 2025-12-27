// Package main provides a standalone REST API server for E2E testing.
// This server starts the REST API without the Wails runtime, enabling
// frontend E2E tests to run against a real backend.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/api"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
	"github.com/ramonehamilton/MTGA-Companion/internal/meta"
	"github.com/ramonehamilton/MTGA-Companion/internal/metrics"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/datasets"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/scryfall"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/cards/setcache"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/deckexport"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/deckimport"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/recommendations"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

var (
	port         = flag.Int("port", 8080, "API server port")
	dbPath       = flag.String("db-path", "", "Database path (default: ~/.mtga-companion/mtga.db)")
	openBrowser  = flag.Bool("open-browser", false, "Open browser to frontend on startup")
	frontendURL  = flag.String("frontend-url", "http://localhost:3000", "Frontend URL to open in browser")
	loadFixtures = flag.String("load-fixtures", "", "Path to SQL fixtures file to load on startup")
)

func main() {
	flag.Parse()

	fmt.Println("MTGA Companion - REST API Server")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Printf("Starting API server on port %d...\n", *port)

	// Setup database path
	finalDBPath := *dbPath
	if finalDBPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalf("Failed to get home directory: %v", err)
		}
		finalDBPath = filepath.Join(home, ".mtga-companion", "mtga.db")
	}

	// Ensure directory exists
	dbDir := filepath.Dir(finalDBPath)
	if err := os.MkdirAll(dbDir, 0o755); err != nil {
		log.Fatalf("Failed to create database directory: %v", err)
	}

	fmt.Printf("Database: %s\n", finalDBPath)

	// Open database
	config := storage.DefaultConfig(finalDBPath)
	config.AutoMigrate = true
	db, err := storage.Open(config)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	storageService := storage.NewService(db)
	defer func() {
		if err := storageService.Close(); err != nil {
			log.Printf("Error closing storage service: %v", err)
		}
	}()

	// Load fixtures if specified
	if *loadFixtures != "" {
		fmt.Printf("Loading fixtures from: %s\n", *loadFixtures)
		if err := loadFixturesFromFile(db, *loadFixtures); err != nil {
			log.Fatalf("Failed to load fixtures: %v", err)
		}
		fmt.Println("Fixtures loaded successfully")
	}

	// Create context
	ctx := context.Background()

	// Initialize card services
	scryfallClient := scryfall.NewClient()

	// Initialize dataset service for 17Lands ratings
	datasetService, err := datasets.NewService(datasets.DefaultServiceOptions())
	if err != nil {
		log.Fatalf("Failed to initialize dataset service: %v", err)
	}

	// Initialize SetFetcher for card metadata
	setFetcher := setcache.NewFetcher(
		scryfallClient,
		storageService.SetCardRepo(),
		storageService.DraftRatingsRepo(),
	)

	// Initialize RatingsFetcher for draft ratings
	ratingsFetcher := setcache.NewRatingsFetcherWithDatasets(
		datasetService,
		storageService.DraftRatingsRepo(),
	)

	// Initialize CardService for card metadata with caching
	cardServiceConfig := cards.DefaultServiceConfig()
	cardServiceConfig.EnableDB = false
	cardService, err := cards.NewService(nil, cardServiceConfig)
	if err != nil {
		log.Fatalf("Failed to initialize card service: %v", err)
	}

	// Initialize DeckImportParser
	deckImportParser := deckimport.NewParser(cardService)

	// Initialize DeckExporter with a CardProvider
	deckExporter := deckexport.NewExporter(cardService)

	// Initialize RecommendationEngine
	ratingsRepo := storageService.DraftRatingsRepo()
	setCardRepo := storageService.SetCardRepo()
	recommendationEngine := recommendations.NewRuleBasedEngineWithSetRepo(cardService, setCardRepo, ratingsRepo)

	// Initialize meta service
	metaService := meta.NewService(nil)

	// Initialize shared services
	services := &gui.Services{
		Context:              ctx,
		Storage:              storageService,
		DaemonPort:           9999,
		DraftMetrics:         metrics.NewDraftMetrics(),
		MetaService:          metaService,
		SetFetcher:           setFetcher,
		RatingsFetcher:       ratingsFetcher,
		CardService:          cardService,
		DatasetService:       datasetService,
		DeckImportParser:     deckImportParser,
		DeckExporter:         deckExporter,
		RecommendationEngine: recommendationEngine,
	}

	// Create facades
	systemFacade := gui.NewSystemFacade(services)
	eventDispatcher := systemFacade.GetEventDispatcher()

	facades := &api.Facades{
		Match:      gui.NewMatchFacade(services),
		Draft:      gui.NewDraftFacade(services),
		Card:       gui.NewCardFacade(services),
		Deck:       gui.NewDeckFacade(services),
		Export:     gui.NewExportFacade(services, eventDispatcher),
		System:     systemFacade,
		Collection: gui.NewCollectionFacade(services),
		Settings:   gui.NewSettingsFacade(services),
		Feedback:   gui.NewFeedbackFacade(services),
		LLM:        gui.NewLLMFacade(services),
		Meta:       gui.NewMetaFacade(metaService),
	}

	// Create API server
	apiConfig := &api.Config{
		Port:        *port,
		OpenBrowser: *openBrowser,
		FrontendURL: *frontendURL,
	}
	server := api.NewServer(apiConfig, services, facades)

	// Start API server
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start API server: %v", err)
	}

	fmt.Println()
	fmt.Printf("API server running at http://localhost:%d\n", *port)
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	fmt.Println("Shutting down...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	fmt.Println("API server stopped.")
}

// loadFixturesFromFile reads and executes SQL statements from a fixture file.
func loadFixturesFromFile(db *storage.DB, filePath string) error {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read fixtures file: %w", err)
	}

	// Execute the SQL statements using the underlying connection
	_, err = db.Conn().Exec(string(content))
	if err != nil {
		return fmt.Errorf("failed to execute fixtures: %w", err)
	}

	return nil
}
