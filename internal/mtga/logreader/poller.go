package logreader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// Poller monitors a log file for new entries and sends them through a channel.
// It tracks file position to only read new entries and handles log file rotation.
type Poller struct {
	path          string
	interval      time.Duration
	useFileEvents bool
	watcher       *fsnotify.Watcher
	lastPos       int64
	lastSize      int64
	lastMod       time.Time
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	updates       chan *LogEntry
	errChan       chan error
	done          chan struct{}
	running       bool
	runningMu     sync.RWMutex
	metrics       *PollerMetrics
	enableMetrics bool
}

// PollerMetrics tracks performance metrics for the poller.
type PollerMetrics struct {
	mu                    sync.RWMutex
	PollCount             uint64
	EntriesProcessed      uint64
	ErrorCount            uint64
	TotalProcessingTime   time.Duration
	LastPollTime          time.Time
	LastPollDuration      time.Duration
	AverageEntriesPerPoll float64
}

// PollerConfig holds configuration for a Poller.
type PollerConfig struct {
	// Path is the path to the log file to monitor.
	Path string

	// Interval is how often to check for new entries when using polling,
	// or how often to perform fallback checks when using file events.
	// Default: 2 seconds
	Interval time.Duration

	// BufferSize is the size of the updates channel buffer.
	// Default: 100
	BufferSize int

	// UseFileEvents enables file system event monitoring (fsnotify) for more
	// efficient log file monitoring. Falls back to periodic polling if file
	// events are unavailable or fail.
	// Default: true
	UseFileEvents bool

	// EnableMetrics enables collection of performance metrics.
	// Default: false
	EnableMetrics bool

	// ReadFromStart if true, reads the entire log file from the beginning
	// on first start. If false, only reads new entries added after start.
	// Default: false (only monitor new entries)
	ReadFromStart bool
}

// DefaultPollerConfig returns a PollerConfig with sensible defaults.
func DefaultPollerConfig(path string) *PollerConfig {
	return &PollerConfig{
		Path:          path,
		Interval:      2 * time.Second,
		BufferSize:    100,
		UseFileEvents: true,
	}
}

// NewPoller creates a new Poller with the given configuration.
func NewPoller(config *PollerConfig) (*Poller, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if config.Path == "" {
		return nil, fmt.Errorf("path cannot be empty")
	}

	if config.Interval == 0 {
		config.Interval = 2 * time.Second
	}
	if config.BufferSize == 0 {
		config.BufferSize = 100
	}

	ctx, cancel := context.WithCancel(context.Background())

	poller := &Poller{
		path:          config.Path,
		interval:      config.Interval,
		useFileEvents: config.UseFileEvents,
		enableMetrics: config.EnableMetrics,
		ctx:           ctx,
		cancel:        cancel,
		updates:       make(chan *LogEntry, config.BufferSize),
		errChan:       make(chan error, 1),
		done:          make(chan struct{}),
		metrics:       &PollerMetrics{},
	}

	// Initialize position tracking
	if err := poller.initializePosition(config.ReadFromStart); err != nil {
		cancel()
		return nil, fmt.Errorf("initialize position: %w", err)
	}

	return poller, nil
}

// initializePosition initializes the poller's position tracking.
// If readFromStart is true, starts at beginning of file to read all existing entries.
// If readFromStart is false, starts at end of file to only read new entries.
func (p *Poller) initializePosition(readFromStart bool) error {
	file, err := os.Open(p.path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist yet, start from position 0
			p.mu.Lock()
			p.lastPos = 0
			p.lastSize = 0
			p.lastMod = time.Time{}
			p.mu.Unlock()
			return nil
		}
		return fmt.Errorf("open file: %w", err)
	}
	defer func() {
		_ = file.Close() //nolint:errcheck // Ignore error on cleanup
	}()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("stat file: %w", err)
	}

	var pos int64
	if readFromStart {
		// Start at beginning to read entire log file
		pos = 0
	} else {
		// Seek to end of file to only monitor new entries
		pos, err = file.Seek(0, io.SeekEnd)
		if err != nil {
			return fmt.Errorf("seek to end: %w", err)
		}
	}

	p.mu.Lock()
	p.lastPos = pos
	p.lastSize = stat.Size()
	p.lastMod = stat.ModTime()
	p.mu.Unlock()

	return nil
}

// Start begins polling the log file for new entries.
// It returns a channel that receives new log entries.
// The poller runs in a separate goroutine and can be stopped with Stop().
func (p *Poller) Start() <-chan *LogEntry {
	p.runningMu.Lock()
	if p.running {
		p.runningMu.Unlock()
		return p.updates
	}
	p.running = true
	p.runningMu.Unlock()

	go p.poll()

	return p.updates
}

// poll is the main polling loop that runs in a goroutine.
func (p *Poller) poll() {
	defer close(p.done)
	defer close(p.updates)

	// Try to use file system events if enabled
	if p.useFileEvents {
		if err := p.setupWatcher(); err != nil {
			// Failed to setup watcher, fall back to polling
			p.sendError(fmt.Errorf("failed to setup file watcher, falling back to polling: %w", err))
			p.pollWithTimer()
			return
		}
		defer p.cleanupWatcher()
		p.pollWithEvents()
	} else {
		// Use timer-based polling
		p.pollWithTimer()
	}
}

// setupWatcher initializes the fsnotify watcher and adds the log file.
func (p *Poller) setupWatcher() error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("create watcher: %w", err)
	}

	p.watcher = watcher

	// Watch the parent directory to catch file creation events after rotation
	// This is necessary because watching a file directly won't catch CREATE events
	// after the file is removed and recreated
	dir := filepath.Dir(p.path)
	if err := p.watcher.Add(dir); err != nil {
		_ = p.watcher.Close() //nolint:errcheck // Ignore error on cleanup
		p.watcher = nil
		return fmt.Errorf("watch directory: %w", err)
	}

	return nil
}

// cleanupWatcher closes the fsnotify watcher.
func (p *Poller) cleanupWatcher() {
	if p.watcher != nil {
		_ = p.watcher.Close()
		p.watcher = nil
	}
}

// pollWithEvents uses file system events for monitoring.
func (p *Poller) pollWithEvents() {
	// Use a ticker for fallback periodic checks (less frequent than pure polling)
	// This ensures we don't miss events
	ticker := time.NewTicker(p.interval * 5) // 5x less frequent
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return

		case event, ok := <-p.watcher.Events:
			if !ok {
				// Watcher closed
				return
			}

			// Handle different event types
			switch {
			case event.Has(fsnotify.Write):
				// File was written to - check for updates
				if err := p.checkForUpdates(); err != nil {
					p.sendError(err)
				}

			case event.Has(fsnotify.Create):
				// File was created (possibly after rotation)
				// Check if it's our target file and read any new content
				if event.Name == p.path {
					fmt.Printf("[INFO] Log file recreated after rotation: %s\n", event.Name)
					// Check for updates immediately
					if err := p.checkForUpdates(); err != nil {
						p.sendError(err)
					}
				}

			case event.Has(fsnotify.Remove), event.Has(fsnotify.Rename):
				// File was removed or renamed (log rotation)
				fmt.Printf("[INFO] Log file rotation detected (%s event): %s\n", event.Op, event.Name)
				// Reset position tracking and wait for file to be recreated
				p.mu.Lock()
				p.lastPos = 0
				p.lastSize = 0
				p.lastMod = time.Time{}
				p.mu.Unlock()
				fmt.Println("[INFO] Position tracking reset, waiting for new log file...")
			}

		case err, ok := <-p.watcher.Errors:
			if !ok {
				// Watcher closed
				return
			}
			p.sendError(fmt.Errorf("watcher error: %w", err))

		case <-ticker.C:
			// Fallback periodic check to ensure we don't miss anything
			if err := p.checkForUpdates(); err != nil {
				p.sendError(err)
			}
		}
	}
}

// pollWithTimer uses timer-based polling (original implementation).
func (p *Poller) pollWithTimer() {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			if err := p.checkForUpdates(); err != nil {
				p.sendError(err)
			}
		}
	}
}

// sendError sends an error through the error channel (non-blocking).
func (p *Poller) sendError(err error) {
	select {
	case p.errChan <- err:
	default:
		// Error channel is full, skip
	}
}

// checkForUpdates checks the log file for new entries and sends them through the updates channel.
func (p *Poller) checkForUpdates() error {
	start := time.Now()
	var entriesProcessed uint64
	var hadError bool

	defer func() {
		if p.enableMetrics {
			duration := time.Since(start)
			p.metrics.mu.Lock()
			p.metrics.PollCount++
			p.metrics.EntriesProcessed += entriesProcessed
			p.metrics.TotalProcessingTime += duration
			p.metrics.LastPollTime = start
			p.metrics.LastPollDuration = duration
			if p.metrics.PollCount > 0 {
				p.metrics.AverageEntriesPerPoll = float64(p.metrics.EntriesProcessed) / float64(p.metrics.PollCount)
			}
			if hadError {
				p.metrics.ErrorCount++
			}
			p.metrics.mu.Unlock()
		}
	}()

	file, err := os.Open(p.path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, reset position tracking
			p.mu.Lock()
			p.lastPos = 0
			p.lastSize = 0
			p.lastMod = time.Time{}
			p.mu.Unlock()
			return nil
		}
		hadError = true
		return fmt.Errorf("open file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		hadError = true
		_ = file.Close() //nolint:errcheck // Ignore error on cleanup
		return fmt.Errorf("stat file: %w", err)
	}

	p.mu.RLock()
	lastPos := p.lastPos
	lastSize := p.lastSize
	lastMod := p.lastMod
	p.mu.RUnlock()

	// Check for log rotation (file size decreased or modification time changed significantly)
	// If file size is less than last position, assume rotation occurred
	if stat.Size() < lastPos || (stat.Size() < lastSize && !stat.ModTime().Equal(lastMod)) {
		// Log file was rotated, reset position
		fmt.Printf("[INFO] Log file rotation detected (size decreased from %d to %d bytes)\n", lastSize, stat.Size())
		p.mu.Lock()
		p.lastPos = 0
		p.mu.Unlock()
		lastPos = 0
	}

	// If file hasn't grown, nothing to do
	if stat.Size() <= lastPos {
		_ = file.Close() //nolint:errcheck // Ignore error on cleanup
		return nil
	}

	// Seek to last read position
	if _, err := file.Seek(lastPos, io.SeekStart); err != nil {
		hadError = true
		_ = file.Close() //nolint:errcheck // Ignore error on cleanup
		return fmt.Errorf("seek to position %d: %w", lastPos, err)
	}

	// Read new entries
	scanner := bufio.NewScanner(file)
	// Increase buffer size to handle very long JSON lines
	const maxScanTokenSize = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxScanTokenSize)
	scanner.Buffer(buf, maxScanTokenSize)

	var newEntries []*LogEntry
	newPos := lastPos

	for scanner.Scan() {
		line := scanner.Text()
		entry := &LogEntry{
			Raw: line,
		}
		entry.parseJSON()

		// Only send JSON entries
		if entry.IsJSON {
			newEntries = append(newEntries, entry)
			entriesProcessed++
		}

		// Update position (line length + newline)
		newPos += int64(len(line)) + 1
	}

	if err := scanner.Err(); err != nil {
		hadError = true
		_ = file.Close() //nolint:errcheck // Ignore error on cleanup
		return fmt.Errorf("scan file: %w", err)
	}

	// Get current position (in case we didn't read to the end)
	currentPos, err := file.Seek(0, io.SeekCurrent)
	if err == nil {
		newPos = currentPos
	}

	_ = file.Close() //nolint:errcheck // Ignore error on cleanup

	// Update position tracking
	p.mu.Lock()
	p.lastPos = newPos
	p.lastSize = stat.Size()
	p.lastMod = stat.ModTime()
	p.mu.Unlock()

	// Send new entries through channel
	for _, entry := range newEntries {
		select {
		case p.updates <- entry:
		case <-p.ctx.Done():
			return p.ctx.Err()
		}
	}

	return nil
}

// Stop stops the poller and closes the updates channel.
// It blocks until the poller has fully stopped.
func (p *Poller) Stop() {
	p.runningMu.Lock()
	if !p.running {
		p.runningMu.Unlock()
		return
	}
	p.running = false
	p.runningMu.Unlock()

	p.cancel()
	<-p.done
}

// Errors returns a channel that receives errors encountered during polling.
func (p *Poller) Errors() <-chan error {
	return p.errChan
}

// IsRunning returns whether the poller is currently running.
func (p *Poller) IsRunning() bool {
	p.runningMu.RLock()
	defer p.runningMu.RUnlock()
	return p.running
}

// Metrics returns a copy of the current poller metrics.
// Returns nil if metrics are not enabled.
func (p *Poller) Metrics() *PollerMetrics {
	if !p.enableMetrics {
		return nil
	}

	p.metrics.mu.RLock()
	defer p.metrics.mu.RUnlock()

	// Return a copy to prevent external modification
	return &PollerMetrics{
		PollCount:             p.metrics.PollCount,
		EntriesProcessed:      p.metrics.EntriesProcessed,
		ErrorCount:            p.metrics.ErrorCount,
		TotalProcessingTime:   p.metrics.TotalProcessingTime,
		LastPollTime:          p.metrics.LastPollTime,
		LastPollDuration:      p.metrics.LastPollDuration,
		AverageEntriesPerPoll: p.metrics.AverageEntriesPerPoll,
	}
}

// LogMetrics logs the current metrics to stdout if metrics are enabled.
func (p *Poller) LogMetrics() {
	metrics := p.Metrics()
	if metrics == nil {
		return
	}

	fmt.Printf("\n=== Poller Metrics ===\n")
	fmt.Printf("Poll Count: %d\n", metrics.PollCount)
	fmt.Printf("Entries Processed: %d\n", metrics.EntriesProcessed)
	fmt.Printf("Error Count: %d\n", metrics.ErrorCount)
	fmt.Printf("Average Entries/Poll: %.2f\n", metrics.AverageEntriesPerPoll)
	fmt.Printf("Total Processing Time: %v\n", metrics.TotalProcessingTime)
	if !metrics.LastPollTime.IsZero() {
		fmt.Printf("Last Poll: %v (duration: %v)\n", metrics.LastPollTime.Format(time.RFC3339), metrics.LastPollDuration)
	}
	fmt.Printf("======================\n\n")
}
