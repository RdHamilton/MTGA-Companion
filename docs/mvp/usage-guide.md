# Usage Guide - MVP Features

This guide covers how to use all MVP features of the MTGA-Companion Draft Overlay.

## Table of Contents

- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [CLI Flags](#cli-flags)
- [Resume Functionality](#resume-functionality)
- [Debug Mode](#debug-mode)
- [Cache Configuration](#cache-configuration)
- [Log Rotation](#log-rotation)
- [Troubleshooting](#troubleshooting)

## Installation

```bash
# Build from source
go build -o bin/mtga-companion ./cmd/mtga-companion

# Or use the development script
./scripts/dev.sh build
```

## Basic Usage

### Starting the Draft Overlay

The most basic command to start the draft overlay:

```bash
./mtga-companion draft-overlay --set MKM --format PremierDraft
```

This will:
1. Auto-detect your MTGA log file location
2. Load the MKM (Murders at Karlov Manor) PremierDraft ratings
3. Enable resume functionality (scan for active drafts)
4. Enable API response caching (24h TTL)
5. Start monitoring for draft events

### Required vs Optional Flags

**Required:**
- `--set` or `--set-file`: Specify which set to use for ratings

**Optional but Recommended:**
- `--format`: Draft format (PremierDraft, QuickDraft, TradDraft)
- `--log-path`: Path to Player.log (if auto-detection fails)

## CLI Flags

### Set Configuration

```bash
# Using set code (auto-loads from ~/.mtga-companion/sets/)
--set MKM
--format PremierDraft

# Using explicit set file path
--set-file "/path/to/MKM_PremierDraft_data.json"
```

**Available Formats:**
- `PremierDraft` (default)
- `QuickDraft`
- `TradDraft`

### Resume Functionality

Resume functionality scans recent log history to detect and resume in-progress drafts.

```bash
# Enable resume (default)
--overlay-resume=true

# Disable resume (start fresh, only process new events)
--overlay-resume=false

# Configure lookback window (default: 24 hours)
--overlay-lookback 48  # Look back 48 hours
```

**When to use resume:**
- ✅ Restarting overlay during a draft
- ✅ Overlay crashed during a draft
- ✅ Want to pick up where you left off

**When to disable resume:**
- ❌ Starting a brand new draft
- ❌ Testing/development (want clean state)
- ❌ Draft is completed and starting a new one

### Debug Mode

Enable verbose debug logging for troubleshooting:

```bash
# Enable debug mode
--debug

# Example with debug enabled
./mtga-companion draft-overlay --set MKM --format PremierDraft --debug
```

**Debug Output Includes:**
- Draft event detection
- Pack processing details
- Resume scan progress
- Cache hit/miss information
- File rotation events
- Error details

**Example Debug Output:**
```
[DEBUG] Scanning log file from 2025-01-10 12:00:00
[DEBUG] Found DraftStatus event: InProgress=true, Pack=1, Pick=1
[DEBUG] Processing pack with 15 cards
[DEBUG] Cache hit for card 12345 (filter: ALL)
[INFO] Draft resumed: Pack 1, Pick 1
```

### Cache Configuration

Control the API response cache behavior:

```bash
# Enable cache with defaults (24h TTL, unlimited size)
--cache=true

# Disable cache
--cache=false

# Configure TTL (time-to-live)
--cache-ttl 12h   # 12 hours
--cache-ttl 48h   # 48 hours
--cache-ttl 1h    # 1 hour

# Configure max size (0 = unlimited)
--cache-max-size 1000  # Max 1000 entries
--cache-max-size 500   # Max 500 entries
```

**Cache Benefits:**
- Reduces redundant card rating lookups
- Faster overlay response times
- Less memory usage with size limit

**Example Configurations:**

```bash
# Short draft session (minimal caching)
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --cache-ttl 1h \
  --cache-max-size 500

# Full day of drafting (aggressive caching)
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --cache-ttl 24h \
  --cache-max-size 0

# Development/testing (no caching)
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --cache=false
```

### Log Path

Override auto-detected log path:

```bash
# macOS default
--log-path "$HOME/Library/Application Support/com.wizards.mtga/Logs/Logs/Player.log"

# Windows default
--log-path "C:\Users\YourName\AppData\LocalLow\Wizards Of The Coast\MTGA\Player.log"

# Custom location
--log-path "/path/to/custom/Player.log"
```

## Resume Functionality

### How Resume Works

When resume is enabled (`--overlay-resume=true`), the overlay:

1. **Scans recent log history** (default: 24 hours back)
2. **Searches for draft events**:
   - `Draft.Notify` (draft start)
   - `DraftStatus` (in-progress check)
   - `Draft.MakePick` (previous picks)
3. **Filters out sealed events** (only supports draft)
4. **Restores draft state** if found
5. **Continues monitoring** for new events

### Resume Scenarios

#### Scenario 1: Overlay Crash During Draft

```bash
# Draft in progress: Pack 2, Pick 5
# Overlay crashes
# Restart with resume enabled
./mtga-companion draft-overlay --set MKM --format PremierDraft

# Result: Overlay scans logs, finds draft state, resumes at Pack 2, Pick 5
```

#### Scenario 2: Starting New Draft

```bash
# Previous draft finished yesterday
# Starting fresh draft today
./mtga-companion draft-overlay --set MKM --format PremierDraft --overlay-resume=false

# Result: Overlay ignores previous draft, only processes new events
```

#### Scenario 3: Extended Lookback

```bash
# Draft started 36 hours ago
# Need to resume with longer lookback
./mtga-companion draft-overlay --set MKM --format PremierDraft --overlay-lookback 48

# Result: Scans 48 hours back to find draft state
```

## Debug Mode

### When to Use Debug Mode

Use `--debug` when:
- Troubleshooting overlay issues
- Verifying draft detection
- Checking cache performance
- Understanding log parsing
- Reporting bugs

### Debug Output Categories

**1. Draft Events:**
```
[DEBUG] Draft event detected: type=Draft.Notify
[DEBUG] Draft started: Pack 1, Pick 1
[DEBUG] Pack contains 15 cards
```

**2. Resume Scanning:**
```
[DEBUG] Scanning log history: lookback=24h
[DEBUG] Found 3 draft events in history
[DEBUG] Most recent draft: Pack 2, Pick 3, InProgress=true
[INFO] Draft resumed successfully
```

**3. Cache Operations:**
```
[DEBUG] Cache miss: card=12345, filter=ALL
[DEBUG] Cache hit: card=12345, filter=BR (hit rate: 67.5%)
[INFO] Card ratings cache enabled (TTL: 24h0m0s, MaxSize: 0)
```

**4. Log Rotation:**
```
[INFO] Log file rotation detected (size decreased from 50000 to 0 bytes)
[INFO] Position tracking reset, waiting for new log file...
[INFO] Log file recreated after rotation: /path/to/Player.log
```

## Cache Configuration

### Understanding Cache Behavior

The cache stores card ratings in memory to avoid redundant lookups.

**Cache Key Format:** `{cardID}_{colorFilter}`

**Example:**
- Card 12345 with filter "ALL": `12345_ALL`
- Card 12345 with filter "BR": `12345_BR`

### Cache Statistics

With debug mode enabled, you can see cache statistics:

```
[DEBUG] Cache stats: hits=150, misses=50, hit_rate=75.0%, size=50, evictions=0
```

**Metrics:**
- **Hits**: Number of cache hits (found in cache)
- **Misses**: Number of cache misses (not in cache)
- **Hit Rate**: Percentage of hits (higher is better)
- **Size**: Current number of cached entries
- **Evictions**: Number of entries removed due to max size

### Cache Tuning

**For short draft sessions (1-2 drafts):**
```bash
--cache-ttl 2h --cache-max-size 500
```
- Low TTL: Ratings won't be stale
- Small size: Minimal memory usage

**For full day of drafting (5+ drafts):**
```bash
--cache-ttl 24h --cache-max-size 0
```
- Long TTL: Maximize cache hits
- Unlimited size: Cache all seen cards

**For development/testing:**
```bash
--cache=false
```
- No cache: Always fetch fresh ratings
- Easier to test rating changes

## Log Rotation

### What is Log Rotation?

MTGA periodically creates new log files and archives old ones. This is called "log rotation."

**How the overlay handles it:**
1. Detects file system events (REMOVE/RENAME/CREATE)
2. Resets position tracking
3. Waits for new log file
4. Continues monitoring seamlessly

**You don't need to do anything** - it's automatic!

### Log Rotation Debug Output

With `--debug`, you'll see rotation events:

```
[INFO] Log file rotation detected (REMOVE event): /path/to/Player.log
[INFO] Position tracking reset, waiting for new log file...
[INFO] Log file recreated after rotation: /path/to/Player.log
[DEBUG] Resuming monitoring from position 0
```

## Troubleshooting

### Overlay Not Detecting Draft

**Problem:** Overlay running but not showing draft info

**Solutions:**
1. Enable debug mode: `--debug`
2. Check log path is correct: `--log-path "/path/to/Player.log"`
3. Verify MTGA is running and logging
4. Check set code matches current draft: `--set MKM`

### Draft Not Resuming

**Problem:** Restarted overlay but draft didn't resume

**Solutions:**
1. Increase lookback window: `--overlay-lookback 48`
2. Check debug output for "Found X draft events"
3. Verify draft is actually in progress in MTGA
4. Try disabling resume and starting fresh: `--overlay-resume=false`

### Cache Not Working

**Problem:** Cache shows 0% hit rate

**Solutions:**
1. Enable debug mode to see cache operations
2. Check cache is enabled: `--cache=true`
3. Verify TTL isn't too short: `--cache-ttl 24h`
4. Check max size isn't too small: `--cache-max-size 0`

### Overlay Crashes

**Problem:** Overlay exits unexpectedly

**Solutions:**
1. Enable debug mode to see error details
2. Check for panic messages in terminal
3. Verify set file exists and is valid JSON
4. Report issue with debug output: https://github.com/RdHamilton/MTGA-Companion/issues

### Performance Issues

**Problem:** Overlay is slow or laggy

**Solutions:**
1. Enable caching if disabled: `--cache=true`
2. Reduce cache max size: `--cache-max-size 1000`
3. Disable debug mode (adds overhead): remove `--debug`
4. Check CPU usage in system monitor

## Examples

### Example 1: Basic Usage

```bash
# Start overlay for Murders at Karlov Manor Premier Draft
./mtga-companion draft-overlay --set MKM --format PremierDraft
```

### Example 2: Debugging

```bash
# Start with debug output and resume disabled
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --debug \
  --overlay-resume=false
```

### Example 3: Production Use

```bash
# Optimized for all-day drafting
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --cache-ttl 24h \
  --cache-max-size 0 \
  --overlay-resume=true \
  --overlay-lookback 24
```

### Example 4: Development

```bash
# Fresh state, no caching, debug output
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --debug \
  --cache=false \
  --overlay-resume=false
```

### Example 5: Custom Log Path

```bash
# macOS with custom log location
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --log-path "/Volumes/GameDrive/MTGA/Player.log"
```

## Next Steps

- **[Draft Overlay Features](draft-overlay.md)** - Learn about draft detection and cleanup
- **[Developer Tools](developer-tools.md)** - Deep dive into debug mode and caching
- **[Architecture](architecture.md)** - Understand the technical design
- **[Testing](testing.md)** - Learn about test coverage and strategy
