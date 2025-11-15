# Development Status

**Last Updated**: 2025-11-15

## Current Sprint/Focus

### v1.0.0 - Service-Based Architecture
- **Status**: Epic #329 in progress - documentation phase
- **Current Task**: #334 - Migration Guide and Documentation
- **Completed**:
  - ✅ #330 - Extract Shared Log Processing Service (PR #335 - merged)
  - ✅ #331 - Add Daemon Mode and WebSocket Server (PR #336 - merged)
  - ✅ #332 - Refactor GUI to Use IPC Client (PR #337 - merged)
  - ✅ #338 - Add Daemon Settings and Connection Status to GUI (PR #339 - merged)
  - ✅ #333 - Platform-Specific Service Installation Scripts (PR #340 - pending merge)
  - 🚧 #334 - Migration Guide and Documentation (current task)
- **Next**: Platform-specific testing (#346), v1.0 release preparation

## Recently Completed

### Epic #329 - Service-Based Architecture Refactor (November 2025) ✅ IN PROGRESS
- ✅ **PR #335** - Extract Shared Log Processing Service
  - Refactored log processing into shared package
  - Used by both daemon and standalone GUI
  - Single source of truth for log parsing
- ✅ **PR #336** - Add Daemon Mode and WebSocket Server
  - Background daemon service for 24/7 log monitoring
  - WebSocket server (port 9999) for real-time events
  - Automatic crash recovery via service manager
- ✅ **PR #337** - Refactor GUI to Use IPC Client
  - GUI connects to daemon via WebSocket
  - Automatic fallback to standalone mode
  - Real-time event handling for match updates
- ✅ **PR #339** - Add Daemon Settings and Connection Status to GUI
  - Connection status indicator in navigation
  - Daemon configuration in Settings page
  - Mode switching (daemon ↔ standalone)
- ✅ **PR #340** - Platform-Specific Service Installation
  - Service management commands (install, start, stop, status)
  - Cross-platform support (macOS/Windows/Linux)
  - Auto-start on system boot
- 🚧 **PR #341** (current) - Migration Guide and Documentation
  - MIGRATION_TO_SERVICE_ARCHITECTURE.md
  - ARCHITECTURE.md with diagrams
  - DEVELOPMENT.md for developers
  - DAEMON_API.md with WebSocket events

### PR #318 - Wails React Migration (November 2025) ✅ MERGED
- ✅ Complete desktop GUI with React + TypeScript + Wails v2
- ✅ Match History page with filtering and sorting
- ✅ Win Rate Trend chart (line chart over time)
- ✅ Deck Performance chart (bar chart)
- ✅ Rank Progression chart (line chart)
- ✅ Format Distribution pie chart
- ✅ Result Breakdown statistics page
- ✅ Settings page for database configuration
- ✅ Real-time updates via poller + events
- ✅ Persistent footer with at-a-glance stats
- ✅ Toast notifications for updates
- ✅ Responsive design (800x600 to 1920x1080+)
- ✅ Dark theme UI
- ✅ CI/CD updated to build frontend
- ✅ All documentation updated (CLAUDE.md, README.md, frontend/README.md, Wiki)
- ✅ All lint errors fixed
- ✅ All CI checks passing across Linux, macOS, and Windows

### PR #320 - Rank Progression Parsing (November 2025) ✅ IMPLEMENTED
- ✅ Parser for RankUpdated events from MTGA logs
- ✅ Extract player rank, season, class, level, step data
- ✅ Support for both Constructed and Limited formats
- ✅ Real-time rank tracking via log file poller
- ✅ Automatic storage with deduplication
- ✅ Comprehensive test suite (11 tests, all passing)
- ✅ Integration with existing log processing pipeline
- ✅ All code quality checks passing

### v0.2.0 - Statistics & Analytics (Prior)
- ✅ Comprehensive statistics engine
- ✅ Export system (CSV/JSON)
- ✅ Time pattern analysis (hour/day)
- ✅ Streak tracking
- ✅ Predictive analytics
- ✅ Season comparisons
- ✅ Database migrations
- ✅ Draft tracking

## In Progress

### 🚧 PR #320 Review - HIGH PRIORITY
**Status**: Awaiting review and testing

**What's left**:
- Review PR #320 code changes
- Test with real MTGA log data (if available)
- Verify rank progression chart displays correctly in GUI
- Merge to main

**Then**:
- Cut v0.3.0 release

### 🚧 E2E Testing Setup (#319) - MEDIUM PRIORITY
**Status**: Identified as needed during PR #318 review

**What's needed**:
- Automated end-to-end tests for GUI
- Test user flows (view matches, apply filters, check charts)
- Verify real-time updates work
- Cross-platform testing (macOS, Windows, Linux)

**Proposed approach**:
- Playwright for E2E testing
- Mock MTGA log file for consistent test data
- GitHub Actions integration

## Next Up (Priority Order)

### High Priority
1. **Rank Progression Parsing** (#317) - Needed for complete GUI feature set
2. **Bug Fixes** - Any issues found in PR #318 testing
3. **Performance Testing** - Verify app performs well with large datasets

### Medium Priority
4. **Enhanced Deck Features**:
   - Deck builder UI
   - Import/export deck lists
   - Deck recommendations based on meta

5. **Draft Overlay GUI**:
   - Move draft overlay from CLI to GUI window
   - In-app draft recommendations
   - Pack simulator

### Low Priority
6. **Collection Tracking** (if data becomes available):
   - Track owned cards
   - Missing cards for decks
   - Wildcard optimization

7. **Advanced Analytics**:
   - Mulligan analysis
   - Play/draw win rates
   - Opponent deck detection
   - Meta-game analysis

## Known Issues

### Critical
- None currently

### Important
- **Rank progression chart empty** (#317) - Needs parser implementation
- **Collection data unavailable** - MTGA doesn't log full collection in Player.log

### Minor
- None currently

## Technical Debt

### High Priority
- [ ] Add frontend TypeScript tests (unit + integration)
- [ ] Add E2E tests for GUI workflows
- [ ] Improve error messages in GUI (more user-friendly)

### Medium Priority
- [ ] Optimize deck inference algorithm for large match counts
- [ ] Add frontend performance monitoring
- [ ] Consider lazy loading for chart pages

### Low Priority
- [ ] Refactor some shared CSS into CSS modules
- [ ] Add loading skeletons instead of spinner text
- [ ] Consider adding a global state manager (if needed)

## Performance Metrics

### Current (as of v0.3.0)
- **Startup time**: ~500ms - 1s cold start
- **Memory usage**: ~50-60 MB (GUI + backend + database)
- **CPU usage (idle)**: ~0-1%
- **CPU usage (active)**: ~2-5% (log parsing + chart rendering)
- **Database size**: ~1-50 MB (varies by match count)

### Target
- Startup: <1s
- Memory: <100 MB
- Idle CPU: <1%
- Active CPU: <10%

## Test Coverage

### Backend (Go)
- **Coverage**: ~85-90%
- **Test count**: 150+ tests
- **CI**: All tests passing on Linux, macOS, Windows

### Frontend (TypeScript/React)
- **Coverage**: 0% (not yet implemented)
- **Test count**: 0
- **CI**: Linting and type-checking only

**Goal**: 80% coverage for frontend by next release

## Dependencies Status

### Backend
- ✅ All dependencies up to date
- ✅ No security vulnerabilities (govulncheck passing)
- ⚠️ GO-2025-4010 (net/url) - known, requires Go 1.24.8 (beyond stable)

### Frontend
- ✅ All dependencies up to date
- ✅ No security vulnerabilities
- ✅ Using React 18, TypeScript 5, Vite 6

## Deployment Status

### v0.3.0-rc1 (Current PR #318)
- **Status**: Ready for merge
- **CI**: All checks passing ✅
- **Platforms**: macOS, Windows, Linux
- **Distribution**: Wails build creates native apps

### Next Release: v0.3.0
- **Planned**: After PR #318 merge
- **Blockers**: None (rank progression can be v0.3.1)
- **Release notes**: To be written

## Community & Contributions

### Recent PRs
- #318 - Wails React Migration (in review)
- #316 - Fyne footer (closed, superseded by #318)

### Open Issues by Priority
- **High**: None (rank progression complete in PR #320)
- **Medium**: #319 (E2E testing), PR #320 (rank progression - needs review)
- **Low**: Various enhancement requests

### Active Contributors
- @RdHamilton (maintainer)
- Community PRs welcome!

## Notes for Next Session

### What we just finished:
- ✅ PR #319 merged - Persistent context documentation
- ✅ PR #320 created - Rank progression parsing implementation
- ✅ RankUpdated event parser with comprehensive tests (11 tests)
- ✅ Integration with log processing pipeline
- ✅ All code quality checks passing

### What to do next:
1. **Review PR #320** - Rank progression parsing (ready for review)
2. **Test with real MTGA logs** - Verify rank data is captured correctly (if MTGA is available)
3. **Merge PR #320** - Complete rank progression feature
4. **Cut v0.3.0 release** - Major release with Wails GUI + rank progression
5. **Add E2E tests** (#319) - Playwright for GUI testing (future enhancement)

### Context for Claude:
- We use Wails v2 (Go backend + React frontend)
- **Service-based architecture**: Daemon (background) + GUI (frontend)
- **Daemon mode (recommended)**: 24/7 log monitoring, WebSocket events
- **Standalone mode (fallback)**: GUI with embedded poller
- Follow responsive design principles (see CLAUDE.md)
- Material Design-inspired dark theme
- All UI must work 800x600 to 1920x1080+
- Real-time updates via WebSocket events (daemon) or EventsEmit (standalone)
- Deck inference links matches to decks via timestamp proximity

### Architecture:
- **Daemon** (`cmd/mtga-companion/daemon.go`): Background service, WebSocket server
- **GUI** (`main.go`, `app.go`): Wails app, IPC client
- **Shared** (`internal/mtga/logprocessor/`): Log parsing logic
- **IPC** (`internal/ipc/`): WebSocket client/server
- **Storage** (`internal/storage/`): SQLite database, repositories

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for complete architecture documentation.
