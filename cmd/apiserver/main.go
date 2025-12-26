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
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

var (
	port   = flag.Int("port", 8080, "API server port")
	dbPath = flag.String("db-path", "", "Database path (default: ~/.mtga-companion/data.db)")
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
		finalDBPath = filepath.Join(home, ".mtga-companion", "data.db")
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

	// Create context
	ctx := context.Background()

	// Initialize meta service
	metaService := meta.NewService(nil)

	// Initialize shared services
	services := &gui.Services{
		Context:      ctx,
		Storage:      storageService,
		DaemonPort:   9999,
		DraftMetrics: metrics.NewDraftMetrics(),
		MetaService:  metaService,
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
		Port: *port,
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
