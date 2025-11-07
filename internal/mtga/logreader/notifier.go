package logreader

import (
	"fmt"
	"sync"
	"time"
)

// EventType represents the type of event for notifications.
type EventType string

const (
	EventMatchComplete EventType = "match_complete"
	EventRankChange    EventType = "rank_change"
	EventDraftComplete EventType = "draft_complete"
	EventMilestone     EventType = "milestone"
	EventCollection    EventType = "collection_update"
)

// EventImportance represents the importance level of an event.
type EventImportance int

const (
	ImportanceLow EventImportance = iota
	ImportanceMedium
	ImportanceHigh
)

// Event represents a notification event.
type Event struct {
	Type       EventType
	Importance EventImportance
	Message    string
	Timestamp  time.Time
	Data       map[string]interface{}
}

// NotificationConfig holds configuration for the notification system.
type NotificationConfig struct {
	// EnabledEvents specifies which event types to notify on.
	// If empty, all events are enabled.
	EnabledEvents []EventType

	// MinImportance is the minimum importance level to notify.
	// Default: ImportanceMedium
	MinImportance EventImportance

	// RateLimit is the minimum duration between notifications of the same type.
	// Default: 5 seconds
	RateLimit time.Duration

	// MaxHistory is the maximum number of events to keep in history.
	// Default: 100
	MaxHistory int

	// EnableConsole enables console notifications.
	// Default: true
	EnableConsole bool

	// EnableSound enables sound notifications.
	// Default: false
	EnableSound bool
}

// DefaultNotificationConfig returns a NotificationConfig with sensible defaults.
func DefaultNotificationConfig() *NotificationConfig {
	return &NotificationConfig{
		EnabledEvents: []EventType{
			EventMatchComplete,
			EventRankChange,
			EventDraftComplete,
			EventMilestone,
		},
		MinImportance: ImportanceMedium,
		RateLimit:     5 * time.Second,
		MaxHistory:    100,
		EnableConsole: true,
		EnableSound:   false,
	}
}

// Notifier manages event notifications.
type Notifier struct {
	config        *NotificationConfig
	history       []*Event
	lastNotified  map[EventType]time.Time
	mu            sync.RWMutex
	eventHandlers []func(*Event)
}

// NewNotifier creates a new Notifier.
func NewNotifier(config *NotificationConfig) *Notifier {
	if config == nil {
		config = DefaultNotificationConfig()
	}

	n := &Notifier{
		config:       config,
		history:      make([]*Event, 0, config.MaxHistory),
		lastNotified: make(map[EventType]time.Time),
	}

	// Add default console handler if enabled
	if config.EnableConsole {
		n.AddHandler(consoleNotificationHandler)
	}

	return n
}

// AddHandler adds a notification event handler.
func (n *Notifier) AddHandler(handler func(*Event)) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.eventHandlers = append(n.eventHandlers, handler)
}

// Notify sends a notification event.
func (n *Notifier) Notify(event *Event) {
	if event == nil {
		return
	}

	// Check if event type is enabled
	if !n.isEventEnabled(event.Type) {
		return
	}

	// Check importance level
	if event.Importance < n.config.MinImportance {
		return
	}

	// Check rate limit
	if !n.checkRateLimit(event.Type) {
		return
	}

	event.Timestamp = time.Now()

	// Add to history
	n.addToHistory(event)

	// Call handlers
	n.mu.RLock()
	handlers := n.eventHandlers
	n.mu.RUnlock()

	for _, handler := range handlers {
		handler(event)
	}

	// Update last notified time
	n.mu.Lock()
	n.lastNotified[event.Type] = event.Timestamp
	n.mu.Unlock()
}

// isEventEnabled checks if an event type is enabled.
func (n *Notifier) isEventEnabled(eventType EventType) bool {
	if len(n.config.EnabledEvents) == 0 {
		return true
	}

	for _, enabled := range n.config.EnabledEvents {
		if enabled == eventType {
			return true
		}
	}

	return false
}

// checkRateLimit checks if enough time has passed since the last notification of this type.
func (n *Notifier) checkRateLimit(eventType EventType) bool {
	n.mu.RLock()
	lastTime, exists := n.lastNotified[eventType]
	n.mu.RUnlock()

	if !exists {
		return true
	}

	return time.Since(lastTime) >= n.config.RateLimit
}

// addToHistory adds an event to the notification history.
func (n *Notifier) addToHistory(event *Event) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.history = append(n.history, event)

	// Trim history if it exceeds max size
	if len(n.history) > n.config.MaxHistory {
		n.history = n.history[len(n.history)-n.config.MaxHistory:]
	}
}

// GetHistory returns a copy of the notification history.
func (n *Notifier) GetHistory() []*Event {
	n.mu.RLock()
	defer n.mu.RUnlock()

	history := make([]*Event, len(n.history))
	copy(history, n.history)
	return history
}

// ProcessEntry processes a log entry and generates notifications based on its content.
func (n *Notifier) ProcessEntry(entry *LogEntry) {
	if entry == nil || !entry.IsJSON {
		return
	}

	// Check for match completion
	if eventType, ok := entry.JSON["CurrentEventState"].(string); ok {
		if eventType == "MatchCompleted" {
			result := "completed"
			if outcome, ok := entry.JSON["outcome"].(string); ok {
				result = outcome
			}

			n.Notify(&Event{
				Type:       EventMatchComplete,
				Importance: ImportanceMedium,
				Message:    fmt.Sprintf("Match %s", result),
				Data: map[string]interface{}{
					"result": result,
				},
			})
		}
	}

	// Check for rank changes
	if rankClass, ok := entry.JSON["rankClass"].(string); ok {
		if rankTier, ok := entry.JSON["rankTier"].(float64); ok {
			n.Notify(&Event{
				Type:       EventRankChange,
				Importance: ImportanceHigh,
				Message:    fmt.Sprintf("Rank updated: %s Tier %d", rankClass, int(rankTier)),
				Data: map[string]interface{}{
					"class": rankClass,
					"tier":  int(rankTier),
				},
			})
		}
	}

	// Check for draft completion
	if draftStatus, ok := entry.JSON["draftStatus"].(string); ok {
		if draftStatus == "Complete" {
			n.Notify(&Event{
				Type:       EventDraftComplete,
				Importance: ImportanceMedium,
				Message:    "Draft completed",
				Data:       map[string]interface{}{},
			})
		}
	}
}

// consoleNotificationHandler is the default console notification handler.
func consoleNotificationHandler(event *Event) {
	prefix := "📢"
	switch event.Importance {
	case ImportanceHigh:
		prefix = "🔴"
	case ImportanceMedium:
		prefix = "🟡"
	case ImportanceLow:
		prefix = "🔵"
	}

	fmt.Printf("\n%s [%s] %s\n", prefix, event.Type, event.Message)
}
