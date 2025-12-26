package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/websocket"
	"github.com/ramonehamilton/MTGA-Companion/internal/gui"
)

// Server represents the REST API server.
type Server struct {
	router     *chi.Mux
	httpServer *http.Server
	port       int

	// WebSocket hub for real-time events
	wsHub *websocket.Hub

	// Facades - same facades used by Wails
	matchFacade      *gui.MatchFacade
	draftFacade      *gui.DraftFacade
	cardFacade       *gui.CardFacade
	deckFacade       *gui.DeckFacade
	exportFacade     *gui.ExportFacade
	systemFacade     *gui.SystemFacade
	collectionFacade *gui.CollectionFacade
	settingsFacade   *gui.SettingsFacade
	feedbackFacade   *gui.FeedbackFacade
	llmFacade        *gui.LLMFacade
	metaFacade       *gui.MetaFacade

	services *gui.Services
}

// Config holds configuration for the API server.
type Config struct {
	Port int
}

// DefaultConfig returns the default API server configuration.
func DefaultConfig() *Config {
	return &Config{
		Port: 8080,
	}
}

// NewServer creates a new API server with the given facades.
func NewServer(cfg *Config, services *gui.Services, facades *Facades) *Server {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Create WebSocket hub
	wsHub := websocket.NewHub()

	s := &Server{
		router:           chi.NewRouter(),
		port:             cfg.Port,
		wsHub:            wsHub,
		services:         services,
		matchFacade:      facades.Match,
		draftFacade:      facades.Draft,
		cardFacade:       facades.Card,
		deckFacade:       facades.Deck,
		exportFacade:     facades.Export,
		systemFacade:     facades.System,
		collectionFacade: facades.Collection,
		settingsFacade:   facades.Settings,
		feedbackFacade:   facades.Feedback,
		llmFacade:        facades.LLM,
		metaFacade:       facades.Meta,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// Facades holds all the facade instances needed by the API server.
type Facades struct {
	Match      *gui.MatchFacade
	Draft      *gui.DraftFacade
	Card       *gui.CardFacade
	Deck       *gui.DeckFacade
	Export     *gui.ExportFacade
	System     *gui.SystemFacade
	Collection *gui.CollectionFacade
	Settings   *gui.SettingsFacade
	Feedback   *gui.FeedbackFacade
	LLM        *gui.LLMFacade
	Meta       *gui.MetaFacade
}

// setupMiddleware configures the middleware stack.
func (s *Server) setupMiddleware() {
	// Request ID for tracing
	s.router.Use(middleware.RequestID)

	// Real IP detection
	s.router.Use(middleware.RealIP)

	// Logging
	s.router.Use(middleware.Logger)

	// Panic recovery
	s.router.Use(middleware.Recoverer)

	// Request timeout
	s.router.Use(middleware.Timeout(60 * time.Second))

	// CORS configuration
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:*", "http://127.0.0.1:*", "https://localhost:*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"Link", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Content-Type enforcement for POST/PUT/PATCH
	s.router.Use(middleware.AllowContentType("application/json"))
}

// Start starts the API server in a goroutine.
func (s *Server) Start() error {
	// Start WebSocket hub
	go s.wsHub.Run()

	s.httpServer = &http.Server{
		Addr:              fmt.Sprintf(":%d", s.port),
		Handler:           s.router,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("API server starting on port %d", s.port)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("API server error: %v", err)
		}
	}()

	return nil
}

// Shutdown gracefully shuts down the API server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	log.Println("Shutting down API server...")
	return s.httpServer.Shutdown(ctx)
}

// Port returns the port the server is configured to listen on.
func (s *Server) Port() int {
	return s.port
}

// WebSocketHub returns the WebSocket hub for external integration.
// This can be used to create a WebSocketObserver for the EventDispatcher.
func (s *Server) WebSocketHub() *websocket.Hub {
	return s.wsHub
}

// NewWebSocketObserver creates a new WebSocket observer that can be registered
// with an EventDispatcher to forward events to WebSocket clients.
func (s *Server) NewWebSocketObserver() *websocket.WebSocketObserver {
	return websocket.NewWebSocketObserver(s.wsHub)
}
