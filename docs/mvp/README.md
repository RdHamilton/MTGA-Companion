# MTGA-Companion MVP Documentation

## Overview

This directory contains comprehensive documentation for the MVP (Minimum Viable Product) release of the MTGA-Companion Draft Overlay system. The MVP includes five critical features that make the draft overlay production-ready.

## MVP Features

### 1. Draft End Detection & Overlay Cleanup ([#261](https://github.com/RdHamilton/MTGA-Companion/issues/261))
**Status:** ✅ Completed (PR [#277](https://github.com/RdHamilton/MTGA-Companion/pull/277))

Automatically detects when a draft ends and properly cleans up the overlay window, ensuring a smooth user experience.

**Key Improvements:**
- Detects `DraftStatus` events with draft completion
- Automatically closes overlay window
- Cleans up resources properly
- Comprehensive test coverage

**Related Files:**
- `internal/mtga/draft/overlay.go:200-230` - Detection logic
- `internal/mtga/draft/overlay_test.go:200-250` - Test coverage

### 2. Log Rotation Handling ([#264](https://github.com/RdHamilton/MTGA-Companion/issues/264))
**Status:** ✅ Completed (PR [#278](https://github.com/RdHamilton/MTGA-Companion/pull/278))

Handles MTGA log file rotation gracefully, ensuring continuous monitoring even when MTGA creates new log files.

**Key Improvements:**
- File system event monitoring (fsnotify)
- Automatic detection of CREATE/REMOVE/RENAME events
- Position tracking reset on rotation
- Fallback to polling if file events unavailable
- Performance metrics tracking

**Related Files:**
- `internal/mtga/logreader/poller.go:203-296` - Event-based monitoring
- `internal/mtga/logreader/poller_test.go` - Test coverage

### 3. Resume Functionality Tests ([#260](https://github.com/RdHamilton/MTGA-Companion/issues/260))
**Status:** ✅ Completed (PR [#279](https://github.com/RdHamilton/MTGA-Companion/pull/279))

Comprehensive test coverage for the resume functionality that allows the overlay to pick up in-progress drafts.

**Key Improvements:**
- Tests for successful draft resume
- Tests for completed draft detection
- Tests for sealed event filtering
- Tests for no active draft scenario
- Edge case coverage

**Related Files:**
- `internal/mtga/draft/overlay_test.go:50-175` - Resume test suite

### 4. Debug/Verbose Mode ([#262](https://github.com/RdHamilton/MTGA-Companion/issues/262))
**Status:** ✅ Completed (PR [#280](https://github.com/RdHamilton/MTGA-Companion/pull/280))

Leveled logging system with debug mode for troubleshooting and development.

**Key Improvements:**
- Three log levels: Debug, Info, Error
- CLI flag: `--debug`
- Conditional debug output
- Integration throughout overlay system
- Comprehensive test coverage

**Related Files:**
- `internal/mtga/draft/logger.go` - Logger implementation
- `internal/mtga/draft/logger_test.go` - Test coverage
- `cmd/mtga-companion/main.go:45` - CLI flag

### 5. API Response Caching ([#263](https://github.com/RdHamilton/MTGA-Companion/issues/263))
**Status:** ✅ Completed (PR [#281](https://github.com/RdHamilton/MTGA-Companion/pull/281))

In-memory caching system for card ratings to reduce redundant API lookups during draft sessions.

**Key Improvements:**
- Thread-safe in-memory cache (RWMutex)
- TTL-based expiration (default 24h)
- FIFO eviction policy
- Cache statistics (hits, misses, evictions, hit rate)
- CLI flags: `--cache`, `--cache-ttl`, `--cache-max-size`
- Comprehensive test coverage (13 tests)

**Related Files:**
- `internal/mtga/draft/cache.go` - Cache implementation (227 lines)
- `internal/mtga/draft/cache_test.go` - Test coverage (453 lines)
- `internal/mtga/draft/ratings.go:45-70` - Cache integration
- `cmd/mtga-companion/main.go:48-50` - CLI flags

## Documentation Structure

- **[Usage Guide](usage-guide.md)** - How to use MVP features with CLI examples
- **[Draft Overlay Features](draft-overlay.md)** - Detailed draft overlay functionality
- **[Developer Tools](developer-tools.md)** - Debug mode and caching details
- **[Architecture](architecture.md)** - Technical design and implementation
- **[Testing](testing.md)** - Test strategy and coverage

## Quick Start

```bash
# Run draft overlay with all MVP features enabled (defaults)
./mtga-companion draft-overlay --set MKM --format PremierDraft

# Enable debug logging
./mtga-companion draft-overlay --set MKM --format PremierDraft --debug

# Configure cache settings
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --cache-ttl 12h \
  --cache-max-size 1000

# Disable resume functionality (start fresh)
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --overlay-resume=false

# Disable caching
./mtga-companion draft-overlay --set MKM --format PremierDraft \
  --cache=false
```

## CI/CD Status

All MVP features are tested in CI/CD:
- ✅ Unit tests pass on Linux, macOS, Windows
- ✅ Integration tests pass
- ✅ Code formatting (gofumpt)
- ✅ Linting (golangci-lint)
- ✅ Security scanning (govulncheck)
- ✅ Race detection

## Project Links

- **Main Repository:** https://github.com/RdHamilton/MTGA-Companion
- **Project Board:** https://github.com/users/RdHamilton/projects/1
- **Issues:** https://github.com/RdHamilton/MTGA-Companion/issues

## Contributing

See [CLAUDE.md](../../CLAUDE.md) for development guidelines and contribution process.
