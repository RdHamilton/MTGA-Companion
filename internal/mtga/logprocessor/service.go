package logprocessor

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage/models"
)

// Service handles processing of MTGA log entries and storing results.
// This service encapsulates all log processing logic to avoid duplication
// between CLI and GUI implementations.
type Service struct {
	storage    *storage.Service
	dryRun     bool // When true, parse entries but don't store to database (for replay testing)
	replayMode bool // When true, keep draft sessions as "in_progress" for UI testing
}

// NewService creates a new log processor service.
func NewService(storage *storage.Service) *Service {
	return &Service{
		storage:    storage,
		dryRun:     false,
		replayMode: false,
	}
}

// SetDryRun enables or disables dry run mode.
// In dry run mode, entries are parsed but not stored to the database.
// This is used for replay testing to avoid polluting the database.
func (s *Service) SetDryRun(enabled bool) {
	s.dryRun = enabled
	if enabled {
		log.Println("⚠️  Log processor in DRY RUN mode - data will NOT be stored to database")
	} else {
		log.Println("✓ Log processor in NORMAL mode - data will be stored to database")
	}
}

// SetReplayMode enables or disables replay mode.
// In replay mode, draft sessions are kept as "in_progress" to enable UI testing of Active Draft view.
func (s *Service) SetReplayMode(enabled bool) {
	s.replayMode = enabled
	if enabled {
		log.Println("🎬 Log processor in REPLAY MODE - draft sessions will remain active for UI testing")
	}
}

// ProcessResult contains the results of processing log entries.
type ProcessResult struct {
	MatchesStored      int
	GamesStored        int
	DecksStored        int
	RanksStored        int
	QuestsStored       int
	QuestsCompleted    int
	AchievementsStored int
	DraftsStored       int
	DraftPicksStored   int
	Errors             []error
}

// ProcessLogEntries processes a batch of log entries and stores all extracted data.
// This is the main entry point for both initial log reads and incremental updates.
func (s *Service) ProcessLogEntries(ctx context.Context, entries []*logreader.LogEntry) (*ProcessResult, error) {
	result := &ProcessResult{
		Errors: []error{},
	}

	// Process arena stats (matches and games)
	if err := s.processArenaStats(ctx, entries, result); err != nil {
		result.Errors = append(result.Errors, err)
	}

	// Process decks
	if err := s.processDecks(ctx, entries, result); err != nil {
		result.Errors = append(result.Errors, err)
	}

	// Process rank updates
	if err := s.processRankUpdates(ctx, entries, result); err != nil {
		result.Errors = append(result.Errors, err)
	}

	// Process quests
	if err := s.processQuests(ctx, entries, result); err != nil {
		result.Errors = append(result.Errors, err)
	}

	// Process graph state for progress tracking (daily wins, weekly wins, etc.)
	// Note: We don't use this for quest COMPLETION anymore - that's handled automatically
	if err := s.processGraphState(ctx, entries, result); err != nil {
		result.Errors = append(result.Errors, err)
	}

	// Process achievements
	if err := s.processAchievements(ctx, entries, result); err != nil {
		result.Errors = append(result.Errors, err)
	}

	// Process draft sessions
	if err := s.processDrafts(ctx, entries, result); err != nil {
		result.Errors = append(result.Errors, err)
	}

	return result, nil
}

// processArenaStats parses and stores arena statistics from log entries.
func (s *Service) processArenaStats(ctx context.Context, entries []*logreader.LogEntry, result *ProcessResult) error {
	arenaStats, err := logreader.ParseArenaStats(entries)
	if err != nil {
		log.Printf("Warning: Failed to parse arena stats: %v", err)
		return err
	}

	// Store stats if we have match data
	if arenaStats != nil && (arenaStats.TotalMatches > 0 || arenaStats.TotalGames > 0) {
		if s.dryRun {
			// Dry run mode: count what would be stored but don't actually store
			log.Printf("[DRY RUN] Would store arena stats: %d matches, %d games", arenaStats.TotalMatches, arenaStats.TotalGames)
			result.MatchesStored = arenaStats.TotalMatches
			result.GamesStored = arenaStats.TotalGames
			return nil
		}

		if err := s.storage.StoreArenaStats(ctx, arenaStats, entries); err != nil {
			log.Printf("Warning: Failed to store arena stats: %v", err)
			return err
		}

		result.MatchesStored = arenaStats.TotalMatches
		result.GamesStored = arenaStats.TotalGames

		log.Printf("✓ Stored statistics: %d matches, %d games", arenaStats.TotalMatches, arenaStats.TotalGames)

		// Note: We don't infer deck IDs here anymore - we wait until AFTER decks are processed
		// to ensure we have the most up-to-date last_played timestamps
	}

	return nil
}

// processDecks parses and stores decks from log entries.
func (s *Service) processDecks(ctx context.Context, entries []*logreader.LogEntry, result *ProcessResult) error {
	deckLibrary, err := logreader.ParseDecks(entries)
	if err != nil {
		log.Printf("Warning: Failed to parse decks: %v", err)
		return err
	}

	if deckLibrary == nil || len(deckLibrary.Decks) == 0 {
		return nil
	}

	log.Printf("Found %d deck(s) in entries", len(deckLibrary.Decks))

	storedCount := 0
	processedCount := 0

	for _, deck := range deckLibrary.Decks {
		// Small delay between deck operations to avoid database lock contention
		if processedCount > 0 {
			time.Sleep(50 * time.Millisecond)
		}
		processedCount++

		// Convert card slices to storage format
		mainDeck := make([]struct {
			CardID   int
			Quantity int
		}, len(deck.MainDeck))
		for i, card := range deck.MainDeck {
			mainDeck[i].CardID = card.CardID
			mainDeck[i].Quantity = card.Quantity
		}

		sideboard := make([]struct {
			CardID   int
			Quantity int
		}, len(deck.Sideboard))
		for i, card := range deck.Sideboard {
			sideboard[i].CardID = card.CardID
			sideboard[i].Quantity = card.Quantity
		}

		// Handle timestamps
		created := deck.Created
		if created.IsZero() && !deck.Modified.IsZero() {
			created = deck.Modified
		} else if created.IsZero() {
			created = time.Now()
		}

		modified := deck.Modified
		if modified.IsZero() {
			modified = time.Now()
		}

		if s.dryRun {
			// Dry run mode: just count, don't store
			storedCount++
		} else {
			err := s.storage.StoreDeckFromParser(
				ctx,
				deck.DeckID,
				deck.Name,
				deck.Format,
				deck.Description,
				created,
				modified,
				deck.LastPlayed,
				mainDeck,
				sideboard,
			)
			if err != nil {
				log.Printf("Warning: Failed to store deck %s: %v", deck.Name, err)
			} else {
				storedCount++
			}
		}
	}

	if storedCount > 0 {
		result.DecksStored = storedCount
		if s.dryRun {
			log.Printf("[DRY RUN] Would store %d/%d deck(s)", storedCount, len(deckLibrary.Decks))
		} else {
			log.Printf("✓ Stored %d/%d deck(s)", storedCount, len(deckLibrary.Decks))

			// Infer deck IDs for matches after storing decks
			inferredCount, err := s.storage.InferDeckIDsForMatches(ctx)
			if err != nil {
				log.Printf("Warning: Failed to infer deck IDs: %v", err)
			} else if inferredCount > 0 {
				log.Printf("✓ Linked %d match(es) to decks", inferredCount)
			}
		}
	}

	return nil
}

// processRankUpdates parses and stores rank updates from log entries.
func (s *Service) processRankUpdates(ctx context.Context, entries []*logreader.LogEntry, result *ProcessResult) error {
	rankUpdates, err := logreader.ParseRankUpdates(entries)
	if err != nil {
		log.Printf("Warning: Failed to parse rank updates: %v", err)
		return err
	}

	if len(rankUpdates) == 0 {
		return nil
	}

	log.Printf("Found %d rank update(s) in entries", len(rankUpdates))

	storedCount := 0
	for _, update := range rankUpdates {
		// Small delay between operations to avoid database lock contention
		if storedCount > 0 && !s.dryRun {
			time.Sleep(25 * time.Millisecond)
		}

		if s.dryRun {
			// Dry run mode: just count, don't store
			storedCount++
		} else {
			if err := s.storage.StoreRankUpdate(ctx, update); err != nil {
				log.Printf("Warning: Failed to store rank update: %v", err)
			} else {
				storedCount++
			}
		}
	}

	if storedCount > 0 {
		result.RanksStored = storedCount
		if s.dryRun {
			log.Printf("[DRY RUN] Would store %d/%d rank update(s)", storedCount, len(rankUpdates))
		} else {
			log.Printf("✓ Stored %d/%d rank update(s)", storedCount, len(rankUpdates))
		}
	}

	return nil
}

// processQuests parses and stores quests from log entries.
func (s *Service) processQuests(ctx context.Context, entries []*logreader.LogEntry, result *ProcessResult) error {
	quests, err := logreader.ParseQuests(entries)
	if err != nil {
		log.Printf("Warning: Failed to parse quests: %v", err)
		return err
	}

	if len(quests) == 0 {
		return nil
	}

	log.Printf("Found %d quest(s) in entries", len(quests))

	storedCount := 0
	for _, questData := range quests {
		// Small delay between operations to avoid database lock contention
		if storedCount > 0 && !s.dryRun {
			time.Sleep(25 * time.Millisecond)
		}

		if s.dryRun {
			// Dry run mode: just count, don't store
			storedCount++
			if questData.Completed {
				result.QuestsCompleted++
			}
		} else {
			// Convert QuestData to storage model
			quest := &models.Quest{
				QuestID:          questData.QuestID,
				QuestType:        questData.QuestType,
				Goal:             questData.Goal,
				StartingProgress: questData.StartingProgress,
				EndingProgress:   questData.EndingProgress,
				Completed:        questData.Completed,
				CanSwap:          questData.CanSwap,
				Rewards:          questData.Rewards,
				AssignedAt:       questData.AssignedAt,
				CompletedAt:      questData.CompletedAt,
				LastSeenAt:       questData.LastSeenAt,
				Rerolled:         questData.Rerolled,
			}

			// Save quest to database
			if err := s.storage.Quests().Save(quest); err != nil {
				log.Printf("Warning: Failed to store quest %s: %v", questData.QuestID, err)
			} else {
				storedCount++
			}
		}
	}

	if storedCount > 0 {
		result.QuestsStored = storedCount
		if s.dryRun {
			log.Printf("[DRY RUN] Would store %d/%d quest(s), %d completed", storedCount, len(quests), result.QuestsCompleted)
		} else {
			log.Printf("✓ Stored %d/%d quest(s)", storedCount, len(quests))
		}
	}

	return nil
}

// processGraphState parses GraphGetGraphState events for progress tracking data.
// Note: We don't use this for quest COMPLETION (handled automatically via ending_progress >= goal).
// Instead, we use it to discover and track other progress data (daily wins, weekly wins, etc.).
func (s *Service) processGraphState(ctx context.Context, entries []*logreader.LogEntry, result *ProcessResult) error {
	graphStates, err := logreader.ParseGraphState(entries)
	if err != nil {
		log.Printf("Warning: Failed to parse graph state: %v", err)
		return err
	}

	if len(graphStates) == 0 {
		return nil
	}

	if s.dryRun {
		// Dry run mode: parse but don't update
		log.Println("[DRY RUN] Would process graph state (mastery pass, daily/weekly wins) but skipping in dry run mode")
		return nil
	}

	// Parse and update mastery pass progression
	masteryPass, _ := logreader.ParseMasteryPass(entries)
	if masteryPass != nil {
		account, err := s.storage.GetCurrentAccount(ctx)
		if err == nil && account != nil {
			// Update mastery pass data if changed
			updated := false
			if masteryPass.CurrentLevel != account.MasteryLevel {
				account.MasteryLevel = masteryPass.CurrentLevel
				updated = true
			}
			if masteryPass.PassType != "" && masteryPass.PassType != account.MasteryPass {
				account.MasteryPass = masteryPass.PassType
				updated = true
			}
			if masteryPass.MaxLevel != 0 && masteryPass.MaxLevel != account.MasteryMax {
				account.MasteryMax = masteryPass.MaxLevel
				updated = true
			}
			if updated {
				if err := s.storage.UpdateAccount(ctx, account); err != nil {
					log.Printf("Warning: Failed to update mastery pass data: %v", err)
				}
			}
		}
	}

	// Parse and update daily/weekly wins
	periodicRewards, _ := logreader.ParsePeriodicRewards(entries)
	if periodicRewards != nil {
		account, err := s.storage.GetCurrentAccount(ctx)
		if err == nil && account != nil {
			// Update daily/weekly wins if changed
			updated := false
			if periodicRewards.DailyWins != account.DailyWins {
				account.DailyWins = periodicRewards.DailyWins
				updated = true
			}
			if periodicRewards.WeeklyWins != account.WeeklyWins {
				account.WeeklyWins = periodicRewards.WeeklyWins
				updated = true
			}
			if updated {
				if err := s.storage.UpdateAccount(ctx, account); err != nil {
					log.Printf("Warning: Failed to update daily/weekly wins: %v", err)
				}
			}
		}
	}

	return nil
}

// processAchievements parses and stores achievements from log entries.
// NOTE: Achievement system temporarily disabled - removed from UI
func (s *Service) processAchievements(ctx context.Context, entries []*logreader.LogEntry, result *ProcessResult) error {
	// Achievement system removed - skip processing
	return nil
}

// processDrafts parses and stores draft sessions from log entries.
func (s *Service) processDrafts(ctx context.Context, entries []*logreader.LogEntry, result *ProcessResult) error {
	// Parse all draft events from entries
	var draftEvents []*logreader.DraftSessionEvent
	for _, entry := range entries {
		event, err := logreader.ParseDraftSessionEvent(entry)
		if err != nil {
			log.Printf("Warning: Failed to parse draft event: %v", err)
			continue
		}
		if event != nil {
			draftEvents = append(draftEvents, event)
		}
	}

	if len(draftEvents) == 0 {
		return nil
	}

	log.Printf("Found %d draft event(s) in entries", len(draftEvents))

	// Group events into sessions
	sessions := s.groupDraftEvents(draftEvents)

	if len(sessions) == 0 {
		return nil
	}

	log.Printf("Grouped into %d draft session(s)", len(sessions))

	// Store each session with its picks and packs
	storedCount := 0
	pickCount := 0
	for _, session := range sessions {
		if s.dryRun {
			// Dry run mode: just count, don't store
			storedCount++
			pickCount += len(session.Picks)
		} else {
			if err := s.storeDraftSession(ctx, session); err != nil {
				log.Printf("Warning: Failed to store draft session: %v", err)
				continue
			}
			storedCount++
			pickCount += len(session.Picks)
		}
	}

	if storedCount > 0 {
		result.DraftsStored = storedCount
		result.DraftPicksStored = pickCount
		if s.dryRun {
			log.Printf("[DRY RUN] Would store %d draft session(s) with %d pick(s)", storedCount, pickCount)
		} else {
			log.Printf("✓ Stored %d draft session(s) with %d pick(s)", storedCount, pickCount)
		}
	}

	return nil
}

// isUUID checks if a string looks like a UUID (e.g., "73e1c7a3-75ee-4b38-b32b-d6854e5c6c9c")
func isUUID(s string) bool {
	// UUID format: 8-4-4-4-12 hexadecimal characters separated by hyphens
	if len(s) != 36 {
		return false
	}
	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return false
	}
	return true
}

// draftSessionData holds all data for a complete draft session.
type draftSessionData struct {
	SessionID string
	EventName string
	SetCode   string
	DraftType string
	StartTime time.Time
	EndTime   *time.Time
	Status    string
	Picks     []*models.DraftPickSession
	Packs     []*models.DraftPackSession
}

// groupDraftEvents groups draft events into complete sessions.
func (s *Service) groupDraftEvents(events []*logreader.DraftSessionEvent) []*draftSessionData {
	// Group events by event name (Quick Draft) or session ID (Premier Draft)
	eventGroups := make(map[string][]*logreader.DraftSessionEvent)
	for _, event := range events {
		// Use EventName for Quick Draft, SessionID for Premier Draft
		groupKey := event.EventName
		if groupKey == "" {
			groupKey = event.SessionID
		}
		if groupKey == "" {
			continue // Skip events with neither EventName nor SessionID
		}
		eventGroups[groupKey] = append(eventGroups[groupKey], event)
	}

	var sessions []*draftSessionData

	for eventName, eventList := range eventGroups {
		// Find start and end times
		var startTime time.Time
		var endTime *time.Time
		hasStart := false
		hasEnd := false

		for _, event := range eventList {
			if event.Type == "started" && !hasStart {
				startTime = event.Timestamp
				hasStart = true
			}
			if event.Type == "ended" && !hasEnd {
				t := event.Timestamp
				endTime = &t
				hasEnd = true
			}
		}

		// Use first event time if no start event found
		if !hasStart && len(eventList) > 0 {
			startTime = eventList[0].Timestamp
		}

		// Extract set code, draft type, event name, and session ID from events
		setCode := ""
		// Default draft type detection:
		// - If grouped by UUID (SessionID), it's PremierDraft
		// - If grouped by EventName string, it's QuickDraft
		// - Context field (HumanDraft/BotDraft) overrides this
		draftType := "QuickDraft"
		if isUUID(eventName) {
			// Grouped by SessionID (UUID format) = Premier Draft
			draftType = "PremierDraft"
			log.Printf("[Draft Detection] Group key is UUID format - defaulting to PremierDraft")
		}
		sessionID := eventName // Default to group key (EventName or SessionID)
		actualEventName := eventName
		detectedContexts := []string{} // Track all contexts seen

		for _, event := range eventList {
			// session_info events (from EventJoin) have complete metadata
			if event.Type == "session_info" {
				if event.EventName != "" {
					actualEventName = event.EventName
				}
				if event.SetCode != "" {
					setCode = event.SetCode
				}
				if event.SessionID != "" {
					sessionID = event.SessionID
				}
			}
			if event.SetCode != "" {
				setCode = event.SetCode
			}
			// Track all contexts for debugging
			if event.Context != "" {
				detectedContexts = append(detectedContexts, event.Context)
			}
			// HumanDraft = Premier/Traditional Draft, BotDraft = Quick Draft
			if event.Context == "HumanDraft" {
				log.Printf("[Draft Detection] Found HumanDraft context - setting type to PremierDraft")
				draftType = "PremierDraft"
			} else if event.Context == "BotDraft" {
				log.Printf("[Draft Detection] Found BotDraft context - keeping type as QuickDraft")
			}
			// Use SessionID if available (Premier Draft)
			if event.SessionID != "" {
				sessionID = event.SessionID
			}
		}

		// Log draft type detection results
		if len(detectedContexts) > 0 {
			log.Printf("[Draft Detection] Group=%s, Contexts=%v, FinalType=%s", eventName, detectedContexts, draftType)
		} else {
			log.Printf("[Draft Detection] Group=%s, NO contexts found - defaulting to %s", eventName, draftType)
		}

		// Build picks and packs from events
		picks := []*models.DraftPickSession{}
		packs := []*models.DraftPackSession{}

		for _, event := range eventList {
			// Save pack contents
			if event.Type == "status_updated" && len(event.DraftPack) > 0 {
				pack := &models.DraftPackSession{
					SessionID:  sessionID,
					PackNumber: event.PackNumber,
					PickNumber: event.PickNumber,
					CardIDs:    event.DraftPack,
					Timestamp:  event.Timestamp,
				}
				packs = append(packs, pack)
			}

			// Save picks
			if event.Type == "pick_made" && len(event.SelectedCard) > 0 {
				for _, cardID := range event.SelectedCard {
					pick := &models.DraftPickSession{
						SessionID:  sessionID,
						PackNumber: event.PackNumber,
						PickNumber: event.PickNumber,
						CardID:     cardID,
						Timestamp:  event.Timestamp,
					}
					picks = append(picks, pick)
				}
			}
		}

		// Fallback: If set_code is still missing, try to infer from card IDs
		if setCode == "" && len(picks) > 0 {
			setCode = s.inferSetCodeFromCardID(picks[0].CardID)
			if setCode != "" {
				log.Printf("Inferred set code %s from card ID %s for session %s", setCode, picks[0].CardID, sessionID)
			}
		}

		// Determine status
		status := "in_progress"
		// In replay mode, keep sessions as "in_progress" for UI testing
		// This allows testers to see the Active Draft view populate in real-time
		if !s.replayMode {
			if hasEnd {
				status = "completed"
			} else {
				// Also mark as completed if all picks are done
				// Quick Draft = 42 picks (3 packs * 14 cards)
				// Premier Draft = 45 picks (3 packs * 15 cards)
				expectedPicks := 42 // Default for QuickDraft
				if draftType == "PremierDraft" {
					expectedPicks = 45
				}
				if len(picks) >= expectedPicks {
					status = "completed"
				}
			}
		}

		session := &draftSessionData{
			SessionID: sessionID,
			EventName: actualEventName,
			SetCode:   setCode,
			DraftType: draftType,
			StartTime: startTime,
			EndTime:   endTime,
			Status:    status,
			Picks:     picks,
			Packs:     packs,
		}

		sessions = append(sessions, session)
	}

	return sessions
}

// storeDraftSession stores a complete draft session with picks and packs.
func (s *Service) storeDraftSession(ctx context.Context, data *draftSessionData) error {
	// Calculate expected total picks based on draft type
	// Quick Draft = 42 picks (3 packs * 14 cards)
	// Premier Draft = 45 picks (3 packs * 15 cards)
	expectedPicks := 42 // Default for QuickDraft
	if data.DraftType == "PremierDraft" {
		expectedPicks = 45
	}

	// Check if session already exists to avoid overwriting metadata
	// This applies to BOTH replay mode AND real-time mode because:
	// - Real-time: Log poller batches entries every 5 seconds, so picks come in multiple batches
	// - Replay: Events processed one at a time
	existingSession, err := s.storage.DraftRepo().GetSession(ctx, data.SessionID)
	if err == nil && existingSession != nil {
		// Session exists - only update picks/packs, don't recreate session
		if s.replayMode {
			log.Printf("DEBUG: [REPLAY] Session %s already exists, only adding new picks/packs", data.SessionID)
		}

		// Store new picks (INSERT OR REPLACE will handle duplicates)
		for _, pick := range data.Picks {
			if err := s.storage.DraftRepo().SavePick(ctx, pick); err != nil {
				log.Printf("Warning: Failed to save pick: %v", err)
			}
		}

		// Store new packs (INSERT OR REPLACE will handle duplicates)
		for _, pack := range data.Packs {
			if err := s.storage.DraftRepo().SavePack(ctx, pack); err != nil {
				log.Printf("Warning: Failed to save pack: %v", err)
			}
		}

		// Check if draft is complete (has all expected picks)
		// Get updated pick count
		picks, err := s.storage.DraftRepo().GetPicksBySession(ctx, data.SessionID)
		if err == nil && len(picks) >= expectedPicks && existingSession.Status == "in_progress" {
			// Draft is complete - mark it as completed
			endTime := time.Now()
			if err := s.storage.DraftRepo().UpdateSessionStatus(ctx, data.SessionID, "completed", &endTime); err != nil {
				log.Printf("Warning: Failed to mark draft session as completed: %v", err)
			} else {
				log.Printf("✓ Draft session %s marked as completed (%d/%d picks)", data.SessionID, len(picks), expectedPicks)
			}
		}

		return nil
	}

	// Session doesn't exist yet, create it
	if s.replayMode {
		log.Printf("DEBUG: [REPLAY] Session %s doesn't exist, creating new session", data.SessionID)
	}

	// Create draft session (first time or non-replay mode)
	session := &models.DraftSession{
		ID:         data.SessionID,
		EventName:  data.EventName,
		SetCode:    data.SetCode,
		DraftType:  data.DraftType,
		StartTime:  data.StartTime,
		EndTime:    data.EndTime,
		Status:     data.Status,
		TotalPicks: expectedPicks, // Use expected total, not current pick count
		CreatedAt:  data.StartTime,
		UpdatedAt:  time.Now(),
	}

	if err := s.storage.DraftRepo().CreateSession(ctx, session); err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	// Store all picks
	for _, pick := range data.Picks {
		if err := s.storage.DraftRepo().SavePick(ctx, pick); err != nil {
			log.Printf("Warning: Failed to save pick: %v", err)
		}
	}

	// Store all packs
	for _, pack := range data.Packs {
		if err := s.storage.DraftRepo().SavePack(ctx, pack); err != nil {
			log.Printf("Warning: Failed to save pack: %v", err)
		}
	}

	return nil
}

// inferSetCodeFromCardID attempts to determine the set code by looking up a card ID
// in the draft_card_ratings table. Returns empty string if not found.
func (s *Service) inferSetCodeFromCardID(cardID string) string {
	ctx := context.Background()

	// Query draft_card_ratings for this card ID
	query := `SELECT DISTINCT set_code FROM draft_card_ratings WHERE arena_id = ? LIMIT 1`
	var setCode string
	err := s.storage.GetDB().QueryRowContext(ctx, query, cardID).Scan(&setCode)
	if err != nil {
		// Card not found in ratings - this is expected if ratings haven't been fetched yet
		return ""
	}

	return setCode
}
