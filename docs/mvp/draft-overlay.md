# Draft Overlay Features

This document covers the three draft-specific MVP features: draft end detection, resume functionality, and log rotation handling.

## Table of Contents

- [Draft End Detection](#draft-end-detection)
- [Resume Functionality](#resume-functionality)
- [Log Rotation Handling](#log-rotation-handling)
- [Integration](#integration)

## Draft End Detection

**Issue:** [#261](https://github.com/RdHamilton/MTGA-Companion/issues/261) | **PR:** [#277](https://github.com/RdHamilton/MTGA-Companion/pull/277)

### Overview

Automatically detects when a draft ends and cleanly shuts down the overlay window, ensuring proper resource cleanup and a smooth user experience.

### How It Works

The overlay monitors for `DraftStatus` events in the MTGA log that indicate draft completion:

```go
type DraftStatus struct {
    DraftID    string
    EventName  string
    Status     string  // "Complete", "Draft_Complete", etc.
    InProgress bool    // false when draft ends
}
```

### Detection Logic

Located in `internal/mtga/draft/overlay.go:200-230`:

```go
func (o *Overlay) handleDraftStatus(status DraftStatus) {
    o.mu.Lock()
    defer o.mu.Unlock()

    // Update current state
    if o.currentState != nil && o.currentState.Event != nil {
        o.currentState.Event.InProgress = status.InProgress
    }

    // Draft ended - trigger cleanup
    if !status.InProgress {
        o.logger.Info("Draft completed, cleaning up overlay")
        o.cleanup()
    }
}

func (o *Overlay) cleanup() {
    // Close channels
    if o.updates != nil {
        close(o.updates)
        o.updates = nil
    }

    // Stop poller
    if o.poller != nil {
        o.poller.Stop()
        o.poller = nil
    }

    // Trigger window close callback
    if o.onClose != nil {
        o.onClose()
    }
}
```

### Event Flow

```
MTGA Draft Completed
    ↓
Log Entry: {"DraftStatus": {"InProgress": false}}
    ↓
Overlay Detects Event
    ↓
handleDraftStatus() Called
    ↓
cleanup() Executed
    ↓
Window Closes
```

### User Experience

**Before MVP:**
- Overlay window stayed open after draft ended
- User had to manually close window
- Resources not properly released

**After MVP:**
- Overlay automatically closes when draft ends
- Clean shutdown with proper resource cleanup
- Seamless transition back to MTGA

### Test Coverage

Located in `internal/mtga/draft/overlay_test.go:200-250`:

```go
func TestDraftEndDetection(t *testing.T) {
    overlay := createTestOverlay()

    // Start draft
    overlay.handleDraftStatus(DraftStatus{
        DraftID: "test-123",
        InProgress: true,
    })

    // End draft
    overlay.handleDraftStatus(DraftStatus{
        DraftID: "test-123",
        InProgress: false,
    })

    // Verify cleanup
    if overlay.currentState.Event.InProgress {
        t.Error("Draft should be marked as complete")
    }

    // Verify callback triggered
    if !overlay.onCloseCalled {
        t.Error("Cleanup callback should be triggered")
    }
}
```

## Resume Functionality

**Issue:** [#260](https://github.com/RdHamilton/MTGA-Companion/issues/260) | **PR:** [#279](https://github.com/RdHamilton/MTGA-Companion/pull/279)

### Overview

Allows the overlay to resume in-progress drafts by scanning recent log history, so users can restart the overlay without losing their draft state.

### How It Works

When resume is enabled (default), the overlay performs a "lookback scan" on startup:

1. **Calculate Lookback Window** (default: 24 hours)
2. **Scan Log Entries** from that time period
3. **Find Draft Events** (Draft.Notify, DraftStatus, Draft.MakePick)
4. **Filter Non-Draft Events** (exclude sealed)
5. **Restore Draft State** if found
6. **Continue Monitoring** for new events

### Scan Logic

Located in `internal/mtga/draft/overlay.go:450-550`:

```go
func (o *Overlay) scanForActiveDraft(lookbackHours int) error {
    o.logger.Debug("Scanning for active draft (lookback: %d hours)", lookbackHours)

    cutoff := time.Now().Add(-time.Duration(lookbackHours) * time.Hour)

    // Read log entries from cutoff time
    entries, err := o.scanner.ScanSince(cutoff)
    if err != nil {
        return fmt.Errorf("scan log history: %w", err)
    }

    o.logger.Debug("Found %d entries in lookback window", len(entries))

    // Process entries to find draft state
    for _, entry := range entries {
        if entry.Timestamp.Before(cutoff) {
            continue
        }

        // Check for draft events
        if draftNotify, ok := entry.ParseDraftNotify(); ok {
            // Only process draft events (not sealed)
            if strings.Contains(strings.ToLower(draftNotify.EventName), "sealed") {
                o.logger.Debug("Skipping sealed event: %s", draftNotify.EventName)
                continue
            }

            o.handleDraftNotify(draftNotify)
            o.logger.Info("Found active draft: %s", draftNotify.EventName)
        }

        if draftStatus, ok := entry.ParseDraftStatus(); ok {
            if draftStatus.InProgress {
                o.handleDraftStatus(draftStatus)
                o.logger.Debug("Draft in progress: Pack %d, Pick %d",
                    draftStatus.CurrentPack, draftStatus.CurrentPick)
            }
        }
    }

    return nil
}
```

### Configuration

**CLI Flags:**
```bash
# Enable resume (default)
--overlay-resume=true

# Disable resume
--overlay-resume=false

# Configure lookback window
--overlay-lookback 48  # 48 hours
```

**In Code:**
```go
config := draft.OverlayConfig{
    ResumeEnabled: true,
    LookbackHours: 24,
}

overlay := draft.NewOverlay(config)
```

### Use Cases

#### Use Case 1: Overlay Crash

```
User is drafting: Pack 2, Pick 5
    ↓
Overlay crashes (bug, system restart, etc.)
    ↓
User restarts overlay with resume enabled
    ↓
Overlay scans log, finds draft state
    ↓
Overlay resumes at Pack 2, Pick 5
```

#### Use Case 2: Fresh Start

```
User finished draft yesterday
    ↓
Starting new draft today
    ↓
User starts overlay with --overlay-resume=false
    ↓
Overlay ignores previous draft
    ↓
Only processes new events
```

### Sealed Event Filtering

The resume functionality **only works for draft events**, not sealed:

```go
// Check event name for "sealed"
if strings.Contains(strings.ToLower(event.EventName), "sealed") {
    logger.Debug("Skipping sealed event: %s", event.EventName)
    continue
}
```

**Sealed events are skipped because:**
- Sealed doesn't have packs/picks (different structure)
- Sealed deck building is not draft-based
- Overlay is designed specifically for draft

### Test Coverage

Located in `internal/mtga/draft/overlay_test.go:50-175`:

#### Test 1: Successful Resume
```go
func TestScanForActiveDraft_Success(t *testing.T) {
    // Create test log with draft events
    logData := createTestLogWithDraft()

    overlay := createTestOverlay()
    err := overlay.scanForActiveDraft(24)

    if err != nil {
        t.Fatalf("scanForActiveDraft() failed: %v", err)
    }

    // Verify draft state restored
    if overlay.currentState == nil {
        t.Fatal("Expected draft state to be restored")
    }

    if overlay.currentState.Event.CurrentPack != 1 {
        t.Errorf("Expected pack 1, got %d", overlay.currentState.Event.CurrentPack)
    }
}
```

#### Test 2: Draft Complete
```go
func TestScanForActiveDraft_DraftComplete(t *testing.T) {
    // Create test log with completed draft
    logData := createTestLogWithCompletedDraft()

    overlay := createTestOverlay()
    err := overlay.scanForActiveDraft(24)

    // Should not fail, just not restore state
    if err != nil {
        t.Fatalf("scanForActiveDraft() failed: %v", err)
    }

    // Verify no active draft
    t.Logf("Draft state after scan: %v", overlay.currentState != nil)
}
```

#### Test 3: Sealed Event Filtering
```go
func TestScanForActiveDraft_FiltersSealedEvents(t *testing.T) {
    // Create test log with sealed event
    logData := createTestLogWithSealedEvent()

    overlay := createTestOverlay()
    err := overlay.scanForActiveDraft(24)

    if err != nil {
        t.Fatalf("scanForActiveDraft() failed: %v", err)
    }

    // Verify sealed event was skipped
    if overlay.currentState != nil {
        // If state exists, verify it's from a draft, not sealed
        if !overlay.currentState.Event.InProgress {
            t.Error("Expected draft to be in progress")
        }
    }
}
```

## Log Rotation Handling

**Issue:** [#264](https://github.com/RdHamilton/MTGA-Companion/issues/264) | **PR:** [#278](https://github.com/RdHamilton/MTGA-Companion/pull/278)

### Overview

MTGA periodically rotates its log files (archives old logs, creates new ones). The overlay must handle this gracefully to maintain continuous monitoring.

### The Problem

**Without rotation handling:**
```
Overlay monitoring Player.log at position 50,000 bytes
    ↓
MTGA rotates log (renames to Player-old.log, creates new Player.log)
    ↓
Overlay tries to read from position 50,000 in NEW file
    ↓
ERROR: Invalid position (new file is 0 bytes)
    ↓
Overlay crashes or misses events
```

**With rotation handling:**
```
Overlay monitoring Player.log at position 50,000 bytes
    ↓
MTGA rotates log
    ↓
Overlay detects REMOVE/RENAME event via fsnotify
    ↓
Overlay resets position to 0
    ↓
Overlay waits for CREATE event (new log file)
    ↓
Overlay continues monitoring from position 0 in new file
```

### Implementation

Located in `internal/mtga/logreader/poller.go:203-296`:

#### File System Event Monitoring

The poller uses `fsnotify` to watch for file system events:

```go
func (p *Poller) setupWatcher() error {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return fmt.Errorf("create watcher: %w", err)
    }

    p.watcher = watcher

    // Watch parent directory (catches file recreation)
    dir := filepath.Dir(p.path)
    if err := p.watcher.Add(dir); err != nil {
        return fmt.Errorf("watch directory: %w", err)
    }

    return nil
}
```

#### Event Processing

```go
func (p *Poller) pollWithEvents() {
    ticker := time.NewTicker(p.interval * 5) // Fallback polling
    defer ticker.Stop()

    for {
        select {
        case event := <-p.watcher.Events:
            switch {
            case event.Has(fsnotify.Write):
                // File written to - check for updates
                p.checkForUpdates()

            case event.Has(fsnotify.Create):
                // File created (after rotation)
                if event.Name == p.path {
                    fmt.Printf("[INFO] Log file recreated: %s\n", event.Name)
                    p.checkForUpdates()
                }

            case event.Has(fsnotify.Remove), event.Has(fsnotify.Rename):
                // File removed/renamed (rotation)
                fmt.Printf("[INFO] Log rotation detected: %s\n", event.Op)
                p.resetPosition()
                fmt.Println("[INFO] Waiting for new log file...")
            }

        case <-ticker.C:
            // Fallback periodic check
            p.checkForUpdates()
        }
    }
}
```

#### Position Reset

```go
func (p *Poller) resetPosition() {
    p.mu.Lock()
    defer p.mu.Unlock()

    p.lastPos = 0
    p.lastSize = 0
    p.lastMod = time.Time{}
}
```

### Event Flow

```
MTGA Running, Logging to Player.log
    ↓
Log file grows to 50 MB
    ↓
MTGA Rotates: Player.log → Player-2025-01-12.log
    ↓
fsnotify: RENAME event detected
    ↓
Poller resets position tracking
    ↓
MTGA Creates: New empty Player.log
    ↓
fsnotify: CREATE event detected
    ↓
Poller resumes monitoring from position 0
    ↓
Continuous monitoring maintained
```

### Fallback Mechanisms

The poller has two fallback mechanisms for robustness:

#### 1. Polling Fallback

If `fsnotify` fails to initialize, the poller falls back to timer-based polling:

```go
if p.useFileEvents {
    if err := p.setupWatcher(); err != nil {
        // Failed to setup watcher, fall back to polling
        p.logger.Error("File watcher failed: %v", err)
        p.pollWithTimer()
        return
    }
    p.pollWithEvents()
} else {
    p.pollWithTimer()
}
```

#### 2. Size-Based Detection

Even without file events, the poller can detect rotation by checking file size:

```go
func (p *Poller) checkForUpdates() error {
    stat, err := file.Stat()
    if err != nil {
        return err
    }

    // If file size decreased, rotation occurred
    if stat.Size() < p.lastPos || stat.Size() < p.lastSize {
        fmt.Printf("[INFO] Rotation detected (size: %d → %d)\n",
            p.lastSize, stat.Size())
        p.resetPosition()
    }

    // Continue reading...
}
```

### Configuration

**File Event Monitoring:**
```go
config := &logreader.PollerConfig{
    Path:          logPath,
    UseFileEvents: true,  // Enable fsnotify (default)
    Interval:      2 * time.Second,
}

poller, err := logreader.NewPoller(config)
```

**Timer-Based Polling:**
```go
config := &logreader.PollerConfig{
    Path:          logPath,
    UseFileEvents: false,  // Disable fsnotify
    Interval:      2 * time.Second,
}
```

### Performance Metrics

When metrics are enabled, the poller tracks rotation events:

```go
type PollerMetrics struct {
    PollCount       uint64
    EntriesProcessed uint64
    ErrorCount       uint64
    // ... other metrics
}
```

### Debug Output

With `--debug`, you'll see rotation events:

```
[INFO] Log file rotation detected (REMOVE event): /path/to/Player.log
[INFO] Position tracking reset, waiting for new log file...
[INFO] Log file recreated after rotation: /path/to/Player.log
[DEBUG] Resuming monitoring from position 0
[DEBUG] Processing 15 new entries
```

## Integration

All three features work together seamlessly in the overlay:

### Initialization Flow

```go
// Create overlay configuration
config := draft.OverlayConfig{
    LogPath:        "/path/to/Player.log",
    SetFile:        setFile,
    ResumeEnabled:  true,   // Resume functionality
    LookbackHours:  24,
    DebugMode:      true,   // Debug logging
}

// Create overlay
overlay := draft.NewOverlay(config)

// Internally, overlay creates:
// 1. Poller with file event monitoring (rotation handling)
// 2. Logger for debug output
// 3. Scanner for resume functionality

// Start monitoring
updates := overlay.Start()

// Handle events
for update := range updates {
    // Process draft updates

    // Draft end detection automatically triggers cleanup
}
```

### Combined User Experience

```
User starts overlay with resume enabled
    ↓
Overlay scans log history (resume)
    ↓
[DEBUG] Found active draft: Pack 2, Pick 3
    ↓
Overlay restores draft state
    ↓
User continues drafting
    ↓
[INFO] Log file rotation detected
    ↓
Overlay resets position, waits for new log
    ↓
[INFO] Log file recreated
    ↓
Overlay continues monitoring
    ↓
User finishes draft
    ↓
MTGA logs: {"DraftStatus": {"InProgress": false}}
    ↓
[INFO] Draft completed, cleaning up overlay
    ↓
Overlay closes automatically
```

## Next Steps

- **[Developer Tools](developer-tools.md)** - Debug mode and API caching
- **[Architecture](architecture.md)** - Technical design details
- **[Testing](testing.md)** - Test coverage and strategies
- **[Usage Guide](usage-guide.md)** - Practical examples
