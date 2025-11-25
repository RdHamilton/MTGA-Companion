package daemon

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/storage/repository"
)

// SyncEventType represents the type of sync event.
type SyncEventType string

const (
	SyncEventStarted   SyncEventType = "collection:sync_started"
	SyncEventCompleted SyncEventType = "collection:sync_completed"
	SyncEventChanged   SyncEventType = "collection:changed"
	SyncEventError     SyncEventType = "collection:sync_error"
)

// SyncEvent represents an event emitted during collection sync.
type SyncEvent struct {
	Type      SyncEventType `json:"type"`
	Timestamp time.Time     `json:"timestamp"`
	Data      interface{}   `json:"data,omitempty"`
}

// CollectionChange represents a change to a card's quantity.
type CollectionChange struct {
	CardID      int    `json:"cardId"`
	OldQuantity int    `json:"oldQuantity"`
	NewQuantity int    `json:"newQuantity"`
	Delta       int    `json:"delta"`
	Source      string `json:"source"` // "sync", "pack", "draft", etc.
}

// SyncResult contains the result of a collection sync operation.
type SyncResult struct {
	TotalCards   int                `json:"totalCards"`
	NewCards     int                `json:"newCards"`
	UpdatedCards int                `json:"updatedCards"`
	Changes      []CollectionChange `json:"changes"`
	Duration     time.Duration      `json:"duration"`
	Success      bool               `json:"success"`
	Error        string             `json:"error,omitempty"`
}

// EventEmitter is an interface for emitting sync events.
type EventEmitter interface {
	Emit(event SyncEvent)
}

// SyncServiceConfig holds configuration for the collection sync service.
type SyncServiceConfig struct {
	// SyncInterval is how often to automatically sync (0 = disabled)
	SyncInterval time.Duration

	// SyncTimeout is the maximum time to wait for a sync operation
	SyncTimeout time.Duration
}

// DefaultSyncServiceConfig returns a SyncServiceConfig with sensible defaults.
func DefaultSyncServiceConfig() *SyncServiceConfig {
	return &SyncServiceConfig{
		SyncInterval: 5 * time.Minute,
		SyncTimeout:  30 * time.Second,
	}
}

// CollectionSyncService handles synchronizing collection data from daemon.
type CollectionSyncService struct {
	client   *Client
	repo     repository.CollectionRepository
	emitter  EventEmitter
	config   *SyncServiceConfig
	mu       sync.RWMutex
	syncing  bool
	lastSync time.Time

	// For scheduled syncs
	stopChan chan struct{}
	doneChan chan struct{}
}

// NewCollectionSyncService creates a new collection sync service.
func NewCollectionSyncService(
	client *Client,
	repo repository.CollectionRepository,
	emitter EventEmitter,
	config *SyncServiceConfig,
) *CollectionSyncService {
	if config == nil {
		config = DefaultSyncServiceConfig()
	}
	return &CollectionSyncService{
		client:  client,
		repo:    repo,
		emitter: emitter,
		config:  config,
	}
}

// FullSync performs a full synchronization of collection data.
func (s *CollectionSyncService) FullSync(ctx context.Context) (*SyncResult, error) {
	s.mu.Lock()
	if s.syncing {
		s.mu.Unlock()
		return nil, fmt.Errorf("sync already in progress")
	}
	s.syncing = true
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		s.syncing = false
		s.lastSync = time.Now()
		s.mu.Unlock()
	}()

	start := time.Now()
	result := &SyncResult{}

	// Emit sync started event
	if s.emitter != nil {
		s.emitter.Emit(SyncEvent{
			Type:      SyncEventStarted,
			Timestamp: start,
		})
	}

	// Apply timeout
	if s.config.SyncTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.config.SyncTimeout)
		defer cancel()
	}

	// Fetch cards from daemon
	cards, err := s.client.GetCards(ctx)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Duration = time.Since(start)
		s.emitError(err)
		return result, fmt.Errorf("failed to fetch cards from daemon: %w", err)
	}

	// Get current collection for change detection
	existingCollection, err := s.repo.GetAll(ctx)
	if err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Duration = time.Since(start)
		s.emitError(err)
		return result, fmt.Errorf("failed to get existing collection: %w", err)
	}

	// Detect changes and prepare entries
	entries := make([]repository.CollectionEntry, 0, len(cards.Cards))
	changes := make([]CollectionChange, 0)
	source := "sync"

	for cardID, newQty := range cards.Cards {
		entries = append(entries, repository.CollectionEntry{
			CardID:   cardID,
			Quantity: newQty,
		})

		oldQty, exists := existingCollection[cardID]
		if !exists {
			// New card
			result.NewCards++
			changes = append(changes, CollectionChange{
				CardID:      cardID,
				OldQuantity: 0,
				NewQuantity: newQty,
				Delta:       newQty,
				Source:      source,
			})
		} else if oldQty != newQty {
			// Updated card
			result.UpdatedCards++
			changes = append(changes, CollectionChange{
				CardID:      cardID,
				OldQuantity: oldQty,
				NewQuantity: newQty,
				Delta:       newQty - oldQty,
				Source:      source,
			})
		}
	}

	// Upsert all cards
	if err := s.repo.UpsertMany(ctx, entries); err != nil {
		result.Success = false
		result.Error = err.Error()
		result.Duration = time.Since(start)
		s.emitError(err)
		return result, fmt.Errorf("failed to save collection: %w", err)
	}

	// Record changes in history (collection already updated by UpsertMany)
	now := time.Now()
	for _, change := range changes {
		if err := s.repo.RecordHistoryEntry(ctx, change.CardID, change.Delta, change.NewQuantity, now, &source); err != nil {
			// Log but don't fail the sync
			continue
		}

		// Emit change event
		if s.emitter != nil {
			s.emitter.Emit(SyncEvent{
				Type:      SyncEventChanged,
				Timestamp: now,
				Data:      change,
			})
		}
	}

	result.TotalCards = len(cards.Cards)
	result.Changes = changes
	result.Duration = time.Since(start)
	result.Success = true

	// Emit sync completed event
	if s.emitter != nil {
		s.emitter.Emit(SyncEvent{
			Type:      SyncEventCompleted,
			Timestamp: time.Now(),
			Data:      result,
		})
	}

	return result, nil
}

// IsSyncing returns true if a sync is currently in progress.
func (s *CollectionSyncService) IsSyncing() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.syncing
}

// LastSyncTime returns the time of the last sync.
func (s *CollectionSyncService) LastSyncTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastSync
}

// StartScheduledSync starts periodic sync operations.
func (s *CollectionSyncService) StartScheduledSync(ctx context.Context) {
	if s.config.SyncInterval <= 0 {
		return
	}

	s.mu.Lock()
	if s.stopChan != nil {
		s.mu.Unlock()
		return // Already running
	}
	s.stopChan = make(chan struct{})
	s.doneChan = make(chan struct{})
	s.mu.Unlock()

	go func() {
		defer close(s.doneChan)
		ticker := time.NewTicker(s.config.SyncInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				_, _ = s.FullSync(ctx)
			case <-s.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// StopScheduledSync stops periodic sync operations.
func (s *CollectionSyncService) StopScheduledSync() {
	s.mu.Lock()
	if s.stopChan == nil {
		s.mu.Unlock()
		return
	}
	close(s.stopChan)
	doneChan := s.doneChan
	s.mu.Unlock()

	// Wait for goroutine to finish
	<-doneChan

	s.mu.Lock()
	s.stopChan = nil
	s.doneChan = nil
	s.mu.Unlock()
}

// emitError emits a sync error event.
func (s *CollectionSyncService) emitError(err error) {
	if s.emitter != nil {
		s.emitter.Emit(SyncEvent{
			Type:      SyncEventError,
			Timestamp: time.Now(),
			Data:      err.Error(),
		})
	}
}

// SetClient updates the daemon client.
func (s *CollectionSyncService) SetClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.client = client
}
