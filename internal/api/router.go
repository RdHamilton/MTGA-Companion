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
			r.Post("/stats", draftHandler.GetDraftStats)
			r.Post("/stats/reset", draftHandler.ResetStats)
			r.Get("/formats", draftHandler.GetDraftFormats)
			r.Get("/recent", draftHandler.GetRecentDrafts)
			r.Post("/grade-pick", draftHandler.GradePick)
			r.Post("/insights", draftHandler.GetDraftInsights)
			r.Post("/archetype-cards", draftHandler.GetArchetypeCards)
			r.Post("/win-probability", draftHandler.PredictWinProbability)
			r.Get("/{sessionID}", draftHandler.GetDraftSession)
			r.Get("/{sessionID}/picks", draftHandler.GetDraftPicks)
			r.Get("/{sessionID}/packs", draftHandler.GetDraftPacks)
			r.Get("/{sessionID}/pool", draftHandler.GetDraftPool)
			r.Get("/{sessionID}/analysis", draftHandler.GetDraftAnalysis)
			r.Get("/{sessionID}/curve", draftHandler.GetDraftCurve)
			r.Get("/{sessionID}/colors", draftHandler.GetDraftColors)
			r.Get("/{sessionID}/current-pack", draftHandler.GetCurrentPack)
			r.Post("/{sessionID}/missing-cards", draftHandler.GetMissingCards)
			r.Post("/{sessionID}/analyze-picks", draftHandler.AnalyzePickQuality)
			r.Post("/{sessionID}/calculate-grade", draftHandler.CalculateGrade)
			r.Post("/{sessionID}/calculate-prediction", draftHandler.CalculatePrediction)
			r.Post("/{sessionID}/repair", draftHandler.RepairSession)
		})

		// Deck routes
		deckHandler := handlers.NewDeckHandler(s.deckFacade)
		r.Route("/decks", func(r chi.Router) {
			r.Get("/", deckHandler.GetDecks)
			r.Post("/", deckHandler.CreateDeck)
			r.Post("/import", deckHandler.ImportDeck)
			r.Post("/parse", deckHandler.ParseDeckList)
			r.Post("/suggest", deckHandler.SuggestDecks)
			r.Post("/analyze", deckHandler.AnalyzeDeck)
			r.Post("/by-tags", deckHandler.GetDecksByTags)
			r.Post("/library", deckHandler.GetDeckLibrary)
			r.Post("/recommendations", deckHandler.GetRecommendations)
			r.Post("/explain-recommendation", deckHandler.ExplainRecommendation)
			r.Post("/classify-draft-pool", deckHandler.ClassifyDraftPoolArchetype)
			r.Post("/apply-suggestion", deckHandler.ApplySuggestedDeck)
			r.Post("/export-suggestion", deckHandler.ExportSuggestedDeck)
			r.Get("/by-draft/{draftEventID}", deckHandler.GetDeckByDraftEvent)
			r.Get("/{deckID}", deckHandler.GetDeck)
			r.Put("/{deckID}", deckHandler.UpdateDeck)
			r.Delete("/{deckID}", deckHandler.DeleteDeck)
			r.Get("/{deckID}/stats", deckHandler.GetDeckStats)
			r.Get("/{deckID}/matches", deckHandler.GetDeckMatches)
			r.Get("/{deckID}/curve", deckHandler.GetDeckCurve)
			r.Get("/{deckID}/colors", deckHandler.GetDeckColors)
			r.Post("/{deckID}/export", deckHandler.ExportDeck)
			r.Post("/{deckID}/clone", deckHandler.CloneDeck)
			r.Post("/{deckID}/cards", deckHandler.AddCard)
			r.Delete("/{deckID}/cards", deckHandler.RemoveCard)
			r.Post("/{deckID}/tags", deckHandler.AddTag)
			r.Delete("/{deckID}/tags/{tag}", deckHandler.RemoveTag)
			r.Get("/{deckID}/validate-draft", deckHandler.ValidateDraftDeck)
		})

		// Card routes
		cardHandler := handlers.NewCardHandler(s.cardFacade)
		r.Route("/cards", func(r chi.Router) {
			r.Get("/", cardHandler.SearchCards)
			r.Post("/search-with-collection", cardHandler.SearchCardsWithCollection)
			r.Get("/dataset-source", cardHandler.GetDatasetSource)
			r.Post("/clear-cache", cardHandler.ClearDatasetCache)
			r.Post("/bulk", cardHandler.GetCardsBulk)
			r.Get("/{cardID}", cardHandler.GetCard)
			r.Get("/name/{name}", cardHandler.GetCardByName)
			r.Get("/sets", cardHandler.GetSets)
			r.Get("/sets/{setCode}", cardHandler.GetSetCards)
			r.Get("/sets/{setCode}/info", cardHandler.GetSetInfo)
			r.Post("/sets/{setCode}/fetch", cardHandler.FetchSetCards)
			r.Post("/sets/{setCode}/refresh", cardHandler.RefreshSetCards)
			r.Get("/ratings/{setCode}", cardHandler.GetRatings)
			r.Get("/ratings/{setCode}/colors", cardHandler.GetColorRatings)
			r.Get("/ratings/{setCode}/{arenaID}", cardHandler.GetCardRatingByArenaID)
			r.Post("/ratings/{setCode}/fetch", cardHandler.FetchSetRatings)
			r.Post("/ratings/{setCode}/refresh", cardHandler.RefreshSetRatings)
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
			r.Get("/version", systemHandler.GetVersion)
			r.Get("/database/path", systemHandler.GetDatabasePath)
			r.Post("/database/path", systemHandler.SetDatabasePath)
			// Daemon routes
			r.Get("/daemon/status", systemHandler.GetDaemonStatus)
			r.Post("/daemon/connect", systemHandler.ConnectDaemon)
			r.Post("/daemon/disconnect", systemHandler.DisconnectDaemon)
			r.Post("/daemon/port", systemHandler.SetDaemonPort)
			r.Post("/daemon/mode/daemon", systemHandler.SwitchToDaemonMode)
			r.Post("/daemon/mode/standalone", systemHandler.SwitchToStandaloneMode)
			// Replay routes
			r.Get("/replay/status", systemHandler.GetReplayStatus)
			r.Get("/replay/progress", systemHandler.GetReplayProgress)
			r.Post("/replay/trigger", systemHandler.TriggerReplay)
			r.Post("/replay/pause", systemHandler.PauseReplay)
			r.Post("/replay/resume", systemHandler.ResumeReplay)
			r.Post("/replay/stop", systemHandler.StopReplay)
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
			r.Post("/import/matches", exportHandler.ImportMatches)
			r.Post("/import/log", exportHandler.ImportLogFile)
			r.Post("/clear", exportHandler.ClearAllData)
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
			r.Get("/dashboard", metaHandler.GetMetaDashboard)
			r.Post("/refresh", metaHandler.RefreshMetaData)
			r.Get("/formats", metaHandler.GetSupportedFormats)
			r.Get("/tier", metaHandler.GetTierArchetypes)
		})

		// Feedback routes
		feedbackHandler := handlers.NewFeedbackHandler(s.feedbackFacade)
		r.Route("/feedback", func(r chi.Router) {
			r.Post("/", feedbackHandler.SubmitFeedback)
			r.Post("/bug", feedbackHandler.SubmitBugReport)
			r.Post("/feature", feedbackHandler.SubmitFeatureRequest)
			r.Post("/recommendation", feedbackHandler.RecordRecommendation)
			r.Post("/action", feedbackHandler.RecordAction)
			r.Post("/outcome", feedbackHandler.RecordOutcome)
			r.Get("/stats", feedbackHandler.GetRecommendationStats)
			r.Get("/dashboard", feedbackHandler.GetDashboardMetrics)
			r.Get("/ml-training", feedbackHandler.ExportMLTrainingData)
		})

		// LLM routes
		llmHandler := handlers.NewLLMHandler(s.llmFacade)
		r.Route("/llm", func(r chi.Router) {
			r.Post("/status", llmHandler.CheckOllamaStatus)
			r.Get("/models", llmHandler.GetAvailableModels)
			r.Post("/models/pull", llmHandler.PullModel)
			r.Post("/test", llmHandler.TestGeneration)
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
