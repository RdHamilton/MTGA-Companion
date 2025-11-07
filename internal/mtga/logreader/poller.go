package logreader

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Poller monitors a log file for new entries and sends them through a channel.
// It tracks file position to only read new entries and handles log file rotation.
type Poller struct {
	path       string
	interval   time.Duration
	lastPos    int64
	lastSize   int64
	lastMod    time.Time
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	updates    chan *LogEntry
	errChan    chan error
	done       chan struct{}
	running    bool
	runningMu  sync.RWMutex
}

// PollerConfig holds configuration for a Poller.
type PollerConfig struct {
	// Path is the path to the log file to monitor.
	Path string

	// Interval is how often to check for new entries.
	// Default: 2 seconds
	Interval time.Duration

	// BufferSize is the size of the updates channel buffer.
	// Default: 100
	BufferSize int
}

// DefaultPollerConfig returns a PollerConfig with sensible defaults.
func DefaultPollerConfig(path string) *PollerConfig {
	return &PollerConfig{
		Path:       path,
		Interval:   2 * time.Second,
		BufferSize: 100,
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
		path:     config.Path,
		interval: config.Interval,
		ctx:      ctx,
		cancel:   cancel,
		updates:  make(chan *LogEntry, config.BufferSize),
		errChan:  make(chan error, 1),
		done:     make(chan struct{}),
	}

	// Initialize position tracking
	if err := poller.initializePosition(); err != nil {
		cancel()
		return nil, fmt.Errorf("initialize position: %w", err)
	}

	return poller, nil
}

// initializePosition initializes the poller's position tracking by reading
// to the end of the file if it exists.
func (p *Poller) initializePosition() error {
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

	// Seek to end of file
	pos, err := file.Seek(0, io.SeekEnd)
	if err != nil {
		return fmt.Errorf("seek to end: %w", err)
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

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			return
		case <-ticker.C:
			if err := p.checkForUpdates(); err != nil {
				// Send error through error channel (non-blocking)
				select {
				case p.errChan <- err:
				default:
					// Error channel is full, skip
				}
			}
		}
	}
}

// checkForUpdates checks the log file for new entries and sends them through the updates channel.
func (p *Poller) checkForUpdates() error {
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
		return fmt.Errorf("open file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
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
	var newPos int64 = lastPos

	for scanner.Scan() {
		line := scanner.Text()
		entry := &LogEntry{
			Raw: line,
		}
		entry.parseJSON()

		// Only send JSON entries
		if entry.IsJSON {
			newEntries = append(newEntries, entry)
		}

		// Update position (line length + newline)
		newPos += int64(len(line)) + 1
	}

	if err := scanner.Err(); err != nil {
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

