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
	storage *storage.Service
}

// NewService creates a new log processor service.
func NewService(storage *storage.Service) *Service {
	return &Service{
		storage: storage,
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

	if storedCount > 0 {
		result.DecksStored = storedCount
		log.Printf("✓ Stored %d/%d deck(s)", storedCount, len(deckLibrary.Decks))

		// Infer deck IDs for matches after storing decks
		inferredCount, err := s.storage.InferDeckIDsForMatches(ctx)
		if err != nil {
			log.Printf("Warning: Failed to infer deck IDs: %v", err)
		} else if inferredCount > 0 {
			log.Printf("✓ Linked %d match(es) to decks", inferredCount)
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
		if storedCount > 0 {
			time.Sleep(25 * time.Millisecond)
		}

		if err := s.storage.StoreRankUpdate(ctx, update); err != nil {
			log.Printf("Warning: Failed to store rank update: %v", err)
		} else {
			storedCount++
		}
	}

	if storedCount > 0 {
		result.RanksStored = storedCount
		log.Printf("✓ Stored %d/%d rank update(s)", storedCount, len(rankUpdates))
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
		if storedCount > 0 {
			time.Sleep(25 * time.Millisecond)
		}

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
			Rerolled:         questData.Rerolled,
		}

		// Save quest to database
		if err := s.storage.Quests().Save(quest); err != nil {
			log.Printf("Warning: Failed to store quest %s: %v", questData.QuestID, err)
		} else {
			storedCount++
		}
	}

	if storedCount > 0 {
		result.QuestsStored = storedCount
		log.Printf("✓ Stored %d/%d quest(s)", storedCount, len(quests))
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
		if err := s.storeDraftSession(ctx, session); err != nil {
			log.Printf("Warning: Failed to store draft session: %v", err)
			continue
		}
		storedCount++
		pickCount += len(session.Picks)
	}

	if storedCount > 0 {
		result.DraftsStored = storedCount
		result.DraftPicksStored = pickCount
		log.Printf("✓ Stored %d draft session(s) with %d pick(s)", storedCount, pickCount)
	}

	return nil
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
	// Group events by event name (each draft run has same event name)
	eventGroups := make(map[string][]*logreader.DraftSessionEvent)
	for _, event := range events {
		if event.EventName == "" {
			continue
		}
		eventGroups[event.EventName] = append(eventGroups[event.EventName], event)
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

		// Extract set code and draft type from event name
		setCode := ""
		draftType := "QuickDraft"
		for _, event := range eventList {
			if event.SetCode != "" {
				setCode = event.SetCode
			}
			if event.Context == "PremierDraft" {
				draftType = "PremierDraft"
			}
		}

		// Create session ID from event name
		sessionID := eventName

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

		// Determine status
		status := "in_progress"
		if hasEnd {
			status = "completed"
		}

		// Count total picks expected (typically 45 for a full draft: 3 packs * 15 picks)
		totalPicks := len(picks)
		if totalPicks == 0 {
			totalPicks = 45 // Default expectation
		}

		session := &draftSessionData{
			SessionID: sessionID,
			EventName: eventName,
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
	// Create draft session
	session := &models.DraftSession{
		ID:         data.SessionID,
		EventName:  data.EventName,
		SetCode:    data.SetCode,
		DraftType:  data.DraftType,
		StartTime:  data.StartTime,
		EndTime:    data.EndTime,
		Status:     data.Status,
		TotalPicks: len(data.Picks),
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
