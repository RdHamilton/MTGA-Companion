package daemon

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logprocessor"
	"github.com/ramonehamilton/MTGA-Companion/internal/mtga/logreader"
	"github.com/ramonehamilton/MTGA-Companion/internal/storage"
)

// Service represents the daemon service that runs continuously.
type Service struct {
	config       *Config
	storage      *storage.Service
	logProcessor *logprocessor.Service
	poller       *logreader.Poller
	wsServer     *WebSocketServer
	ctx          context.Context
	cancel       context.CancelFunc
	startTime    time.Time
}

// New creates a new daemon service.
func New(config *Config, storage *storage.Service) *Service {
	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		config:       config,
		storage:      storage,
		logProcessor: logprocessor.NewService(storage),
		wsServer:     NewWebSocketServer(config.Port),
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start starts the daemon service.
func (s *Service) Start() error {
	s.startTime = time.Now()
	log.Println("Starting MTGA Companion daemon...")

	// Determine log path
	logPath := s.config.LogPath
	if logPath == "" {
		detected, err := logreader.DefaultLogPath()
		if err != nil {
			return fmt.Errorf("failed to detect log path: %w", err)
		}
		logPath = detected
		log.Printf("Auto-detected log path: %s", logPath)
	}

	// Create and start log poller
	pollerConfig := logreader.DefaultPollerConfig(logPath)
	pollerConfig.Interval = s.config.PollInterval
	pollerConfig.UseFileEvents = s.config.UseFSNotify
	pollerConfig.EnableMetrics = s.config.EnableMetrics

	poller, err := logreader.NewPoller(pollerConfig)
	if err != nil {
		return fmt.Errorf("failed to create log poller: %w", err)
	}

	s.poller = poller

	// Start WebSocket server in background
	go func() {
		if err := s.wsServer.Start(); err != nil && err != http.ErrServerClosed {
			log.Printf("WebSocket server error: %v", err)
		}
	}()

	// Wait for WebSocket server to start
	time.Sleep(100 * time.Millisecond)

	// Start log poller
	updates := s.poller.Start()
	errChan := s.poller.Errors()

	log.Printf("Daemon started successfully")
	log.Printf("WebSocket server: ws://localhost:%d", s.config.Port)
	log.Printf("Status endpoint: http://localhost:%d/status", s.config.Port)

	// Send status event
	s.wsServer.Broadcast(Event{
		Type: "daemon:status",
		Data: map[string]interface{}{
			"status":  "running",
			"port":    s.config.Port,
			"logPath": logPath,
		},
	})

	// Process log updates
	go s.processUpdates(updates, errChan)

	// Send periodic status updates
	go s.sendPeriodicStatus()

	return nil
}

// Stop gracefully stops the daemon service.
func (s *Service) Stop() error {
	log.Println("Stopping daemon...")

	// Cancel context
	s.cancel()

	// Stop poller
	if s.poller != nil {
		s.poller.Stop()
	}

	// Stop WebSocket server
	if s.wsServer != nil {
		if err := s.wsServer.Stop(); err != nil {
			log.Printf("Error stopping WebSocket server: %v", err)
		}
	}

	log.Println("Daemon stopped")
	return nil
}

// processUpdates processes log updates and broadcasts events.
func (s *Service) processUpdates(updates <-chan *logreader.LogEntry, errChan <-chan error) {
	var entryBuffer []*logreader.LogEntry
	ticker := time.NewTicker(5 * time.Second) // Batch process every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case entry, ok := <-updates:
			if !ok {
				return
			}
			// Buffer entries for batch processing
			entryBuffer = append(entryBuffer, entry)
		case err, ok := <-errChan:
			if !ok {
				return
			}
			log.Printf("Poller error: %v", err)
			// Broadcast error event
			s.wsServer.Broadcast(Event{
				Type: "daemon:error",
				Data: map[string]interface{}{
					"error": err.Error(),
				},
			})
		case <-ticker.C:
			// Process buffered entries
			if len(entryBuffer) > 0 {
				s.processEntries(entryBuffer)
				entryBuffer = nil // Clear buffer
			}
		}
	}
}

// processEntries processes a batch of log entries.
func (s *Service) processEntries(entries []*logreader.LogEntry) {
	log.Printf("Processing %d log entries...", len(entries))
	result, err := s.logProcessor.ProcessLogEntries(s.ctx, entries)
	if err != nil {
		log.Printf("Error processing log entries: %v", err)
		return
	}

	// Broadcast events for updates
	if result.MatchesStored > 0 || result.GamesStored > 0 {
		log.Printf("Stored %d matches, %d games", result.MatchesStored, result.GamesStored)
		s.wsServer.Broadcast(Event{
			Type: "stats:updated",
			Data: map[string]interface{}{
				"matches": result.MatchesStored,
				"games":   result.GamesStored,
			},
		})
	}

	if result.DecksStored > 0 {
		log.Printf("Stored %d deck(s)", result.DecksStored)
		s.wsServer.Broadcast(Event{
			Type: "deck:updated",
			Data: map[string]interface{}{
				"count": result.DecksStored,
			},
		})
	}

	if result.RanksStored > 0 {
		log.Printf("Stored %d rank update(s)", result.RanksStored)
		s.wsServer.Broadcast(Event{
			Type: "rank:updated",
			Data: map[string]interface{}{
				"count": result.RanksStored,
			},
		})
	}

	if result.QuestsStored > 0 {
		log.Printf("Stored %d quest(s)", result.QuestsStored)
		s.wsServer.Broadcast(Event{
			Type: "quest:updated",
			Data: map[string]interface{}{
				"count": result.QuestsStored,
			},
		})
	}
}

// sendPeriodicStatus sends periodic status updates to clients.
func (s *Service) sendPeriodicStatus() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			uptime := time.Since(s.startTime).Seconds()
			s.wsServer.Broadcast(Event{
				Type: "daemon:status",
				Data: map[string]interface{}{
					"status":  "running",
					"uptime":  uptime,
					"clients": s.wsServer.ClientCount(),
				},
			})
		}
	}
}

// GetUptime returns the daemon uptime in seconds.
func (s *Service) GetUptime() float64 {
	return time.Since(s.startTime).Seconds()
}

// GetClientCount returns the number of connected WebSocket clients.
func (s *Service) GetClientCount() int {
	return s.wsServer.ClientCount()
}
