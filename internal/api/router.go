package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/ramonehamilton/MTGA-Companion/internal/api/handlers"
	"github.com/ramonehamilton/MTGA-Companion/internal/api/response"
)

// setupRoutes configures all API routes.
func (s *Server) setupRoutes() {
	// Health check endpoint (no versioning)
	s.router.Get("/health", s.healthCheck)

	// WebSocket endpoint (no JSON content-type requirement)
	s.router.Get("/ws", s.wsHub.ServeWs)

	// API v1 routes
	s.router.Route("/api/v1", func(r chi.Router) {
		// Match routes
		matchHandler := handlers.NewMatchHandler(s.matchFacade)
		r.Route("/matches", func(r chi.Router) {
			r.Post("/", matchHandler.GetMatches)       // POST for complex filters
			r.Get("/{matchID}", matchHandler.GetMatch) // Get single match
			r.Get("/{matchID}/games", matchHandler.GetMatchGames)
			r.Post("/stats", matchHandler.GetStats) // POST for complex filters
			r.Post("/trends", matchHandler.GetTrendAnalysis)
			r.Get("/formats", matchHandler.GetFormats)
			r.Get("/archetypes", matchHandler.GetArchetypes)
			r.Post("/format-distribution", matchHandler.GetFormatDistribution)
			r.Post("/win-rate-over-time", matchHandler.GetWinRateOverTime)
			r.Post("/performance-by-hour", matchHandler.GetPerformanceByHour)
			r.Post("/matchup-matrix", matchHandler.GetMatchupMatrix)
		})

		// Draft routes
		draftHandler := handlers.NewDraftHandler(s.draftFacade)
		r.Route("/drafts", func(r chi.Router) {
			r.Post("/", draftHandler.GetDraftSessions) // POST for complex filters
			r.Get("/{sessionID}", draftHandler.GetDraftSession)
			r.Get("/{sessionID}/picks", draftHandler.GetDraftPicks)
			r.Get("/{sessionID}/pool", draftHandler.GetDraftPool)
			r.Get("/{sessionID}/analysis", draftHandler.GetDraftAnalysis)
			r.Get("/{sessionID}/curve", draftHandler.GetDraftCurve)
			r.Get("/{sessionID}/colors", draftHandler.GetDraftColors)
			r.Post("/stats", draftHandler.GetDraftStats)
			r.Get("/formats", draftHandler.GetDraftFormats)
			r.Get("/recent", draftHandler.GetRecentDrafts)
			r.Post("/grade-pick", draftHandler.GradePick)
			r.Post("/insights", draftHandler.GetDraftInsights)
			r.Post("/win-probability", draftHandler.PredictWinProbability)
		})

		// Deck routes
		deckHandler := handlers.NewDeckHandler(s.deckFacade)
		r.Route("/decks", func(r chi.Router) {
			r.Get("/", deckHandler.GetDecks)
			r.Post("/", deckHandler.CreateDeck)
			r.Get("/{deckID}", deckHandler.GetDeck)
			r.Put("/{deckID}", deckHandler.UpdateDeck)
			r.Delete("/{deckID}", deckHandler.DeleteDeck)
			r.Get("/{deckID}/stats", deckHandler.GetDeckStats)
			r.Get("/{deckID}/matches", deckHandler.GetDeckMatches)
			r.Get("/{deckID}/curve", deckHandler.GetDeckCurve)
			r.Get("/{deckID}/colors", deckHandler.GetDeckColors)
			r.Post("/{deckID}/export", deckHandler.ExportDeck)
			r.Post("/import", deckHandler.ImportDeck)
			r.Post("/parse", deckHandler.ParseDeckList)
			r.Post("/suggest", deckHandler.SuggestDecks)
			r.Post("/analyze", deckHandler.AnalyzeDeck)
		})

		// Card routes
		cardHandler := handlers.NewCardHandler(s.cardFacade)
		r.Route("/cards", func(r chi.Router) {
			r.Get("/", cardHandler.SearchCards)
			r.Get("/{cardID}", cardHandler.GetCard)
			r.Get("/name/{name}", cardHandler.GetCardByName)
			r.Get("/sets", cardHandler.GetSets)
			r.Get("/sets/{setCode}", cardHandler.GetSetCards)
			r.Get("/ratings/{setCode}", cardHandler.GetRatings)
			r.Post("/bulk", cardHandler.GetCardsBulk)
		})

		// Collection routes
		collectionHandler := handlers.NewCollectionHandler(s.collectionFacade)
		r.Route("/collection", func(r chi.Router) {
			r.Get("/", collectionHandler.GetCollection)
			r.Get("/stats", collectionHandler.GetCollectionStats)
			r.Get("/sets", collectionHandler.GetCollectionBySets)
			r.Get("/rarity", collectionHandler.GetCollectionByRarity)
			r.Get("/missing/{setCode}", collectionHandler.GetMissingCards)
			r.Post("/search", collectionHandler.SearchCollection)
		})

		// System routes
		systemHandler := handlers.NewSystemHandler(s.systemFacade)
		r.Route("/system", func(r chi.Router) {
			r.Get("/status", systemHandler.GetStatus)
			r.Get("/daemon/status", systemHandler.GetDaemonStatus)
			r.Post("/daemon/connect", systemHandler.ConnectDaemon)
			r.Post("/daemon/disconnect", systemHandler.DisconnectDaemon)
			r.Get("/version", systemHandler.GetVersion)
			r.Get("/database/path", systemHandler.GetDatabasePath)
			r.Post("/database/path", systemHandler.SetDatabasePath)
		})

		// Settings routes
		settingsHandler := handlers.NewSettingsHandler(s.settingsFacade)
		r.Route("/settings", func(r chi.Router) {
			r.Get("/", settingsHandler.GetSettings)
			r.Put("/", settingsHandler.UpdateSettings)
			r.Get("/{key}", settingsHandler.GetSetting)
			r.Put("/{key}", settingsHandler.UpdateSetting)
		})

		// Export routes
		exportHandler := handlers.NewExportHandler(s.exportFacade)
		r.Route("/export", func(r chi.Router) {
			r.Post("/matches", exportHandler.ExportMatches)
			r.Post("/drafts", exportHandler.ExportDrafts)
			r.Post("/collection", exportHandler.ExportCollection)
			r.Post("/deck", exportHandler.ExportDeck)
			r.Get("/formats", exportHandler.GetExportFormats)
		})

		// Quest routes (from system facade)
		questHandler := handlers.NewQuestHandler(s.systemFacade)
		r.Route("/quests", func(r chi.Router) {
			r.Get("/active", questHandler.GetActiveQuests)
			r.Get("/history", questHandler.GetQuestHistory)
			r.Get("/wins/daily", questHandler.GetDailyWins)
			r.Get("/wins/weekly", questHandler.GetWeeklyWins)
		})

		// Meta routes
		metaHandler := handlers.NewMetaHandler(s.metaFacade)
		r.Route("/meta", func(r chi.Router) {
			r.Get("/archetypes", metaHandler.GetMetaArchetypes)
			r.Get("/deck-analysis", metaHandler.GetDeckAnalysis)
			r.Post("/identify-archetype", metaHandler.IdentifyArchetype)
		})

		// Feedback routes
		feedbackHandler := handlers.NewFeedbackHandler(s.feedbackFacade)
		r.Route("/feedback", func(r chi.Router) {
			r.Post("/", feedbackHandler.SubmitFeedback)
			r.Post("/bug", feedbackHandler.SubmitBugReport)
			r.Post("/feature", feedbackHandler.SubmitFeatureRequest)
		})
	})
}

// healthCheck returns server health status.
func (s *Server) healthCheck(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "mtga-companion-api",
	})
}
