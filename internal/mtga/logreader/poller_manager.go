package logreader

import (
	"context"
	"fmt"
	"sync"
)

// PollerManager manages multiple log file pollers simultaneously.
type PollerManager struct {
	pollers   map[string]*Poller
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	updates   chan *LogEntry
	errChan   chan error
	notifier  *Notifier
	running   bool
	runningMu sync.RWMutex
	wg        sync.WaitGroup
}

// PollerManagerConfig holds configuration for a PollerManager.
type PollerManagerConfig struct {
	// BufferSize is the size of the aggregated updates channel buffer.
	// Default: 1000
	BufferSize int

	// EnableNotifications enables the notification system.
	// Default: false
	EnableNotifications bool

	// NotificationConfig is the configuration for notifications.
	NotificationConfig *NotificationConfig
}

// DefaultPollerManagerConfig returns a PollerManagerConfig with sensible defaults.
func DefaultPollerManagerConfig() *PollerManagerConfig {
	return &PollerManagerConfig{
		BufferSize:          1000,
		EnableNotifications: false,
		NotificationConfig:  DefaultNotificationConfig(),
	}
}

// NewPollerManager creates a new PollerManager.
func NewPollerManager(config *PollerManagerConfig) *PollerManager {
	if config == nil {
		config = DefaultPollerManagerConfig()
	}

	if config.BufferSize == 0 {
		config.BufferSize = 1000
	}

	ctx, cancel := context.WithCancel(context.Background())

	var notifier *Notifier
	if config.EnableNotifications {
		notifier = NewNotifier(config.NotificationConfig)
	}

	return &PollerManager{
		pollers:  make(map[string]*Poller),
		ctx:      ctx,
		cancel:   cancel,
		updates:  make(chan *LogEntry, config.BufferSize),
		errChan:  make(chan error, 100),
		notifier: notifier,
	}
}

// AddPoller adds a new poller to the manager.
// The key is used to identify the poller (e.g., file path or instance name).
func (pm *PollerManager) AddPoller(key string, config *PollerConfig) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.pollers[key]; exists {
		return fmt.Errorf("poller with key %q already exists", key)
	}

	poller, err := NewPoller(config)
	if err != nil {
		return fmt.Errorf("create poller: %w", err)
	}

	pm.pollers[key] = poller

	// If manager is running, start the new poller immediately
	if pm.isRunningLocked() {
		pm.startPoller(key, poller)
	}

	return nil
}

// RemovePoller removes and stops a poller from the manager.
func (pm *PollerManager) RemovePoller(key string) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	poller, exists := pm.pollers[key]
	if !exists {
		return fmt.Errorf("poller with key %q not found", key)
	}

	poller.Stop()
	delete(pm.pollers, key)

	return nil
}

// Start starts all pollers in the manager.
func (pm *PollerManager) Start() <-chan *LogEntry {
	pm.runningMu.Lock()
	if pm.running {
		pm.runningMu.Unlock()
		return pm.updates
	}
	pm.running = true
	pm.runningMu.Unlock()

	pm.mu.RLock()
	for key, poller := range pm.pollers {
		pm.startPoller(key, poller)
	}
	pm.mu.RUnlock()

	return pm.updates
}

// startPoller starts a single poller and aggregates its output.
func (pm *PollerManager) startPoller(key string, poller *Poller) {
	updates := poller.Start()
	errChan := poller.Errors()

	// Capture key by value for the goroutine
	pollerKey := key

	// Aggregate updates from this poller (Go 1.25: WaitGroup.Go handles Add/Done)
	pm.wg.Go(func() {
		for {
			select {
			case <-pm.ctx.Done():
				return
			case entry, ok := <-updates:
				if !ok {
					// Poller stopped
					return
				}
				// Send to aggregated channel
				select {
				case pm.updates <- entry:
					// Process notifications if enabled
					if pm.notifier != nil {
						pm.notifier.ProcessEntry(entry)
					}
				case <-pm.ctx.Done():
					return
				}
			case err, ok := <-errChan:
				if !ok {
					return
				}
				// Forward error with context
				select {
				case pm.errChan <- fmt.Errorf("poller %q: %w", pollerKey, err):
				default:
					// Error channel full, skip
				}
			}
		}
	})
}

// Stop stops all pollers in the manager.
func (pm *PollerManager) Stop() {
	pm.runningMu.Lock()
	if !pm.running {
		pm.runningMu.Unlock()
		return
	}
	pm.running = false
	pm.runningMu.Unlock()

	pm.mu.Lock()
	for _, poller := range pm.pollers {
		poller.Stop()
	}
	pm.mu.Unlock()

	pm.cancel()

	// Wait for all aggregator goroutines to finish
	pm.wg.Wait()

	close(pm.updates)
}

// Errors returns a channel that receives errors from all pollers.
func (pm *PollerManager) Errors() <-chan error {
	return pm.errChan
}

// IsRunning returns whether the manager is currently running.
func (pm *PollerManager) IsRunning() bool {
	pm.runningMu.RLock()
	defer pm.runningMu.RUnlock()
	return pm.running
}

// isRunningLocked checks if running without acquiring lock (must hold lock).
func (pm *PollerManager) isRunningLocked() bool {
	pm.runningMu.RLock()
	defer pm.runningMu.RUnlock()
	return pm.running
}

// AggregateMetrics returns aggregated metrics from all pollers.
func (pm *PollerManager) AggregateMetrics() *PollerMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	aggregate := &PollerMetrics{}
	count := 0

	for _, poller := range pm.pollers {
		metrics := poller.Metrics()
		if metrics == nil {
			continue
		}

		aggregate.PollCount += metrics.PollCount
		aggregate.EntriesProcessed += metrics.EntriesProcessed
		aggregate.ErrorCount += metrics.ErrorCount
		aggregate.TotalProcessingTime += metrics.TotalProcessingTime

		// Track latest poll time
		if metrics.LastPollTime.After(aggregate.LastPollTime) {
			aggregate.LastPollTime = metrics.LastPollTime
			aggregate.LastPollDuration = metrics.LastPollDuration
		}

		count++
	}

	// Calculate average
	if aggregate.PollCount > 0 {
		aggregate.AverageEntriesPerPoll = float64(aggregate.EntriesProcessed) / float64(aggregate.PollCount)
	}

	return aggregate
}

// PollerCount returns the number of pollers being managed.
func (pm *PollerManager) PollerCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return len(pm.pollers)
}

// GetNotifier returns the notification system if enabled.
func (pm *PollerManager) GetNotifier() *Notifier {
	return pm.notifier
}
