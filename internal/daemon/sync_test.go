package daemon

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
	_ "modernc.org/sqlite"
)

// mockEventEmitter collects events for testing.
type mockEventEmitter struct {
	mu     sync.Mutex
	events []SyncEvent
}

func (m *mockEventEmitter) Emit(event SyncEvent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
}

func (m *mockEventEmitter) Events() []SyncEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]SyncEvent{}, m.events...)
}

func (m *mockEventEmitter) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = nil
}

// setupSyncTestDB creates an in-memory database with collection tables.
func setupSyncTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	schema := `
		CREATE TABLE collection (
			card_id INTEGER PRIMARY KEY,
			quantity INTEGER NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE collection_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			card_id INTEGER NOT NULL,
			quantity_delta INTEGER NOT NULL,
			quantity_after INTEGER NOT NULL,
			timestamp DATETIME NOT NULL,
			source TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestNewCollectionSyncService(t *testing.T) {
	config := DefaultClientConfig(9999)
	client := NewClient(config)
	emitter := &mockEventEmitter{}

	db := setupSyncTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := repository.NewCollectionRepository(db)
	syncConfig := DefaultSyncServiceConfig()

	service := NewCollectionSyncService(client, repo, emitter, syncConfig)

	if service == nil {
		t.Fatal("expected service to be created")
	}
	if service.config.SyncInterval != 5*time.Minute {
		t.Errorf("expected sync interval 5m, got %v", service.config.SyncInterval)
	}
}

func TestCollectionSyncService_FullSync(t *testing.T) {
	// Create mock daemon server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cards" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(CardCollection{
				Cards: map[int]int{
					12345: 4,
					67890: 2,
					11111: 1,
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Setup client
	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	client := NewClient(config)

	// Setup database
	db := setupSyncTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := repository.NewCollectionRepository(db)
	emitter := &mockEventEmitter{}

	syncConfig := DefaultSyncServiceConfig()
	syncConfig.SyncTimeout = 5 * time.Second

	service := NewCollectionSyncService(client, repo, emitter, syncConfig)

	// Perform sync
	ctx := context.Background()
	result, err := service.FullSync(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify result
	if result.TotalCards != 3 {
		t.Errorf("expected 3 total cards, got %d", result.TotalCards)
	}
	if result.NewCards != 3 {
		t.Errorf("expected 3 new cards, got %d", result.NewCards)
	}
	if !result.Success {
		t.Error("expected success to be true")
	}
	if len(result.Changes) != 3 {
		t.Errorf("expected 3 changes, got %d", len(result.Changes))
	}

	// Verify collection was saved
	collection, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("failed to get collection: %v", err)
	}

	if len(collection) != 3 {
		t.Errorf("expected 3 cards in collection, got %d", len(collection))
	}
	if collection[12345] != 4 {
		t.Errorf("expected card 12345 to have 4 copies, got %d", collection[12345])
	}

	// Verify events were emitted
	events := emitter.Events()
	if len(events) < 2 {
		t.Fatalf("expected at least 2 events (started + completed), got %d", len(events))
	}

	// First event should be started
	if events[0].Type != SyncEventStarted {
		t.Errorf("expected first event type %s, got %s", SyncEventStarted, events[0].Type)
	}

	// Last event should be completed
	lastEvent := events[len(events)-1]
	if lastEvent.Type != SyncEventCompleted {
		t.Errorf("expected last event type %s, got %s", SyncEventCompleted, lastEvent.Type)
	}
}

func TestCollectionSyncService_FullSync_DetectsChanges(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/cards" {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(CardCollection{
				Cards: map[int]int{
					12345: 4, // Will be updated (was 2)
					67890: 2, // Unchanged
					11111: 1, // New card
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	client := NewClient(config)

	db := setupSyncTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := repository.NewCollectionRepository(db)
	ctx := context.Background()

	// Pre-populate some cards
	_ = repo.UpsertCard(ctx, 12345, 2)
	_ = repo.UpsertCard(ctx, 67890, 2)

	emitter := &mockEventEmitter{}
	service := NewCollectionSyncService(client, repo, emitter, nil)

	// Perform sync
	result, err := service.FullSync(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify change detection
	if result.NewCards != 1 {
		t.Errorf("expected 1 new card, got %d", result.NewCards)
	}
	if result.UpdatedCards != 1 {
		t.Errorf("expected 1 updated card, got %d", result.UpdatedCards)
	}

	// Find the change for card 12345
	var found bool
	for _, change := range result.Changes {
		if change.CardID == 12345 {
			found = true
			if change.OldQuantity != 2 {
				t.Errorf("expected old quantity 2, got %d", change.OldQuantity)
			}
			if change.NewQuantity != 4 {
				t.Errorf("expected new quantity 4, got %d", change.NewQuantity)
			}
			if change.Delta != 2 {
				t.Errorf("expected delta 2, got %d", change.Delta)
			}
			break
		}
	}
	if !found {
		t.Error("expected to find change for card 12345")
	}
}

func TestCollectionSyncService_FullSync_AlreadyInProgress(t *testing.T) {
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(2 * time.Second)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CardCollection{Cards: map[int]int{}})
	}))
	defer slowServer.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = slowServer.URL
	config.Timeout = 10 * time.Second
	client := NewClient(config)

	db := setupSyncTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := repository.NewCollectionRepository(db)
	service := NewCollectionSyncService(client, repo, nil, nil)

	ctx := context.Background()

	// Start first sync in background
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = service.FullSync(ctx)
	}()

	// Wait a bit for sync to start
	time.Sleep(100 * time.Millisecond)

	// Try to start second sync
	_, err := service.FullSync(ctx)
	if err == nil {
		t.Error("expected error when sync already in progress")
	}

	wg.Wait()
}

func TestCollectionSyncService_IsSyncing(t *testing.T) {
	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(500 * time.Millisecond)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CardCollection{Cards: map[int]int{}})
	}))
	defer slowServer.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = slowServer.URL
	client := NewClient(config)

	db := setupSyncTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := repository.NewCollectionRepository(db)
	service := NewCollectionSyncService(client, repo, nil, nil)

	// Initially not syncing
	if service.IsSyncing() {
		t.Error("expected not syncing initially")
	}

	ctx := context.Background()

	// Start sync in background
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = service.FullSync(ctx)
	}()

	// Wait a bit and check if syncing
	time.Sleep(100 * time.Millisecond)
	if !service.IsSyncing() {
		t.Error("expected syncing to be true during sync")
	}

	wg.Wait()

	// Should not be syncing after completion
	if service.IsSyncing() {
		t.Error("expected not syncing after completion")
	}
}

func TestCollectionSyncService_LastSyncTime(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CardCollection{Cards: map[int]int{}})
	}))
	defer server.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	client := NewClient(config)

	db := setupSyncTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := repository.NewCollectionRepository(db)
	service := NewCollectionSyncService(client, repo, nil, nil)

	// Initially zero
	if !service.LastSyncTime().IsZero() {
		t.Error("expected last sync time to be zero initially")
	}

	ctx := context.Background()
	before := time.Now()
	_, _ = service.FullSync(ctx)
	after := time.Now()

	lastSync := service.LastSyncTime()
	if lastSync.Before(before) || lastSync.After(after) {
		t.Errorf("expected last sync time between %v and %v, got %v", before, after, lastSync)
	}
}

func TestCollectionSyncService_FullSync_DaemonError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("daemon error"))
	}))
	defer server.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	config.MaxRetries = 0 // No retries for faster test
	client := NewClient(config)

	db := setupSyncTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := repository.NewCollectionRepository(db)
	emitter := &mockEventEmitter{}
	service := NewCollectionSyncService(client, repo, emitter, nil)

	ctx := context.Background()
	result, err := service.FullSync(ctx)

	if err == nil {
		t.Error("expected error for daemon failure")
	}
	if result.Success {
		t.Error("expected success to be false")
	}
	if result.Error == "" {
		t.Error("expected error message to be set")
	}

	// Should have error event
	events := emitter.Events()
	hasErrorEvent := false
	for _, e := range events {
		if e.Type == SyncEventError {
			hasErrorEvent = true
			break
		}
	}
	if !hasErrorEvent {
		t.Error("expected error event to be emitted")
	}
}

func TestCollectionSyncService_ScheduledSync(t *testing.T) {
	var syncCount int
	var mu sync.Mutex

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		syncCount++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(CardCollection{Cards: map[int]int{}})
	}))
	defer server.Close()

	config := DefaultClientConfig(9999)
	config.BaseURL = server.URL
	client := NewClient(config)

	db := setupSyncTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := repository.NewCollectionRepository(db)
	syncConfig := &SyncServiceConfig{
		SyncInterval: 100 * time.Millisecond,
		SyncTimeout:  5 * time.Second,
	}
	service := NewCollectionSyncService(client, repo, nil, syncConfig)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduled sync
	service.StartScheduledSync(ctx)

	// Wait for a few syncs
	time.Sleep(350 * time.Millisecond)

	// Stop scheduled sync
	service.StopScheduledSync()

	mu.Lock()
	count := syncCount
	mu.Unlock()

	// Should have synced at least 2-3 times
	if count < 2 {
		t.Errorf("expected at least 2 syncs, got %d", count)
	}
}

func TestCollectionSyncService_StopScheduledSync_NotRunning(t *testing.T) {
	config := DefaultClientConfig(9999)
	client := NewClient(config)

	db := setupSyncTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := repository.NewCollectionRepository(db)
	service := NewCollectionSyncService(client, repo, nil, nil)

	// Should not panic when stopping without starting
	service.StopScheduledSync()
}

func TestCollectionSyncService_SetClient(t *testing.T) {
	config1 := DefaultClientConfig(9999)
	client1 := NewClient(config1)

	db := setupSyncTestDB(t)
	defer func() {
		if err := db.Close(); err != nil {
			t.Errorf("Error closing database: %v", err)
		}
	}()

	repo := repository.NewCollectionRepository(db)
	service := NewCollectionSyncService(client1, repo, nil, nil)

	// Create new client
	config2 := DefaultClientConfig(8888)
	client2 := NewClient(config2)

	service.SetClient(client2)

	// Verify client was updated (by checking the URL indirectly)
	if service.client.GetBaseURL() != "http://localhost:8888" {
		t.Errorf("expected client base URL to be updated, got %s", service.client.GetBaseURL())
	}
}

func TestSyncResult_Fields(t *testing.T) {
	result := &SyncResult{
		TotalCards:   100,
		NewCards:     10,
		UpdatedCards: 5,
		Changes: []CollectionChange{
			{CardID: 1, OldQuantity: 0, NewQuantity: 4, Delta: 4, Source: "sync"},
		},
		Duration: 500 * time.Millisecond,
		Success:  true,
	}

	if result.TotalCards != 100 {
		t.Errorf("expected total cards 100, got %d", result.TotalCards)
	}
	if result.NewCards != 10 {
		t.Errorf("expected new cards 10, got %d", result.NewCards)
	}
	if len(result.Changes) != 1 {
		t.Errorf("expected 1 change, got %d", len(result.Changes))
	}
}
