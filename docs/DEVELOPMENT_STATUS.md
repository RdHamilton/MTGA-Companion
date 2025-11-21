# Development Status

**Last Updated**: 2025-11-21

## Current Sprint/Focus

### v1.2 - Application Refactor & Testing
- **Status**: Planning phase - 33 issues created across 5 phases
- **Current Task**: Preparing v1.1 release documentation
- **Next**: Begin Phase 1 (Facade Pattern) after v1.1 release

### v1.1 - Draft Assistant & UI Enhancements
- **Status**: ✅ COMPLETED - Ready for release
- **Release Date**: Pending (documentation finalization in progress)

## Recently Completed

### v1.1 - Draft Assistant & UI Enhancements (November 2025) ✅ COMPLETED

#### Major Features (15 PRs)

**Draft Assistant & Analysis**
- ✅ **PR #424** - Type synergy detection and card suggestions for draft
  - Real-time type synergy analysis (Creatures, Instants, Sorceries, etc.)
  - Context-aware card suggestions based on picked cards
  - Synergy badges and visual indicators
- ✅ **PR #432** - Missing cards detection from draft packs (Issue #179)
  - Automatic detection of missing cards from packs
  - Collection tracking integration
  - Visual indicators for missing cards
- ✅ **PR #435** - Draft statistics and mana curve visualization (Issue #181)
  - Real-time mana curve as picks are made
  - Draft statistics dashboard
  - Color distribution analysis
- ✅ **PR #441** - Format meta insights (Issue #392)
  - Format-wide archetype performance data
  - Best color pairs and combinations
  - Overdrafted colors analysis
- ✅ **PR #442** - Archetype Performance Dashboard (Issue #391)
  - Interactive archetype selection
  - Top cards per archetype
  - Win rate and popularity filtering
  - Best removal and commons by archetype

**Match Tracking & Display**
- ✅ **PR #434** - Match details view with game breakdown (Issue #365)
  - Expandable match details
  - Game-by-game breakdown
  - Win/loss tracking per game

**Draft Grading & Prediction**
- ✅ **PR #390, #404** - Draft deck win rate predictor
  - Predictive model for draft deck performance
  - Grade breakdown (A/B/C/D/F)
  - Expected win rate estimation
- ✅ **PR #393, #403** - Draft grade UI with breakdown modal
  - Visual grade display
  - Detailed breakdown of grade calculation
  - Interactive modal with statistics

**Historical Draft & Replay**
- ✅ **PR #402** - Draft UI improvements and historical draft replay
  - View past drafts
  - Replay draft pick sequences
  - Historical performance analysis
- ✅ **PR #413** - Phase 1: Core Replay Engine for log simulation
  - Infrastructure for replaying MTGA logs
  - Event simulation engine
  - Testing framework foundation
- ✅ **PR #414** - Phase 2: CLI command for log replay testing
  - `replay` CLI command
  - Log file processing and simulation
  - Development testing tools
- ✅ **PR #415** - Phase 3: GUI replay controls
  - Replay controls in GUI
  - Play/pause/reset functionality
  - Progress tracking
- ✅ **PR #419, #421** - Draft UI improvements with tier lists and controls
  - Card tier list visualization
  - Enhanced card images
  - N/A grades for missing data
  - Improved replay controls

**Data Integration & Performance**
- ✅ **PR #423** - Migrate to 17Lands public datasets
  - Updated to use 17Lands public API
  - Fixed draft grading bugs
  - More reliable card ratings
- ✅ **PR #436** - Performance monitoring and metrics (Issue #266)
  - Draft overlay performance metrics
  - Monitoring infrastructure
  - Performance optimization

**Log Management**
- ✅ **PR #408** - Phase 1: Startup recovery for log archival
  - Automatic recovery on application startup
  - Historical data preservation
- ✅ **PR #410** - Phase 2: UTC_Log monitoring for runtime log rotation
  - Handle MTGA log rotation at runtime
  - Seamless transition between log files
- ✅ **PR #411** - Phase 3: Manual log file import UI
  - Import historical MTGA logs via GUI
  - Batch processing support
- ✅ **PR #412** - Phase 4: Automatic log archival (opt-in)
  - Configurable automatic archival
  - Preservation of historical data
- ✅ **PR #400, #401** - Historical log import on initial installation
  - Automatic import of existing logs on first run
  - User onboarding improvements

#### Bug Fixes (6 PRs)
- ✅ **PR #433** - Fix database locking during rapid draft replay events (Issue #431)
- ✅ **PR #430** - Fix duplicate replay controls and toast spam (Issues #428, #427)
- ✅ **PR #429** - Fix tier list scrolling in Draft view (Issue #426)
- ✅ **PR #418** - Fix replay testing tool - active draft detection and UI improvements
- ✅ **PR #417** - Fix draft card images and add N/A grades for missing data
- ✅ **PR #405** - Fix win rate prediction modal overlay issues

#### Documentation & Cleanup (2 PRs)
- ✅ **PR #443** - Reorganize documentation structure for v1.1 release
  - Moved documentation to `docs/` directory
  - Created docs/README.md index
  - Updated all cross-references
  - Improved documentation discoverability
- ✅ **PR #445** - Remove obsolete Fyne files and daemon display code
  - Removed 8,488 lines of obsolete code
  - Cleaned up Fyne UI remnants
  - Removed unused daemon display files
  - Project cleanup for v1.1

### v1.0 - Wails Desktop GUI & Service Architecture (Prior)
- ✅ Complete desktop GUI with React + TypeScript + Wails v2
- ✅ Service-based architecture (daemon + GUI)
- ✅ Match History, Win Rate Trends, Deck Performance
- ✅ Rank Progression tracking
- ✅ Real-time updates and notifications
- ✅ Dark theme UI with responsive design
- ✅ Cross-platform support (macOS, Windows, Linux)

## In Progress

### 🚧 v1.1 Release Documentation - CURRENT TASK
**Status**: Finalizing documentation for v1.1 release

**What's left**:
- ✅ Update DEVELOPMENT_STATUS.md with v1.1 features
- 🚧 Update CHANGELOG.md with v1.1 release notes
- 🚧 Review and update README.md if needed
- 🚧 Create v1.1 git tag
- 🚧 Create GitHub release with release notes

**Then**:
- Start v1.2 Phase 1 (Facade Pattern)

## Next Up (Priority Order)

### v1.2 - Application Refactor (33 issues across 5 phases)

**Phase 1: Facade Pattern** (9 issues - Milestone #32)
- Break down app.go (2,814 lines → ~300 lines)
- Create domain-specific facades: Match, Draft, Deck, Card, Export, System
- Critical for deck builder preparation

**Phase 2: Strategy Pattern** (4 issues - Milestone #33)
- Pluggable draft format analysis strategies
- Support for Premier Draft, Quick Draft, and future formats

**Phase 3: Builder Pattern** (3 issues - Milestone #34)
- Fluent API for export operations
- Simplify export code

**Phase 4: Observer & Command Patterns** (6 issues - Milestone #35)
- Event dispatcher for decoupled event management
- Command pattern for daemon operations

**Phase 5: UI Testing Infrastructure** (10 issues - Milestone #36)
- Vitest + React Testing Library for component tests
- Playwright for E2E tests
- CI/CD integration for automated testing
- Target: 70-80% frontend coverage

### Future Features (Post-v1.2)
- **Deck Builder**: Full deck builder UI with import/export
- **Advanced Analytics**: Mulligan analysis, play/draw win rates
- **Collection Tracking**: Owned cards, missing cards, wildcard optimization

## Known Issues

### Critical
- None currently

### Important
- **Collection data incomplete** - MTGA doesn't log full collection in Player.log
  - Workaround: Missing cards detection from draft packs (implemented in v1.1)

### Minor
- None currently

## Technical Debt

### High Priority (Addressed in v1.2)
- [ ] Refactor app.go God Object (2,814 lines) → **Phase 1: Facade Pattern**
- [ ] Add frontend TypeScript tests → **Phase 5: UI Testing Infrastructure**
- [ ] Add E2E tests for GUI workflows → **Phase 5: UI Testing Infrastructure**
- [ ] Standardize event handling → **Phase 4: Observer Pattern**

### Medium Priority (Addressed in v1.2)
- [ ] Pluggable draft format strategies → **Phase 2: Strategy Pattern**
- [ ] Simplify export operations → **Phase 3: Builder Pattern**
- [ ] Daemon command structure → **Phase 4: Command Pattern**

### Low Priority
- [ ] Add loading skeletons instead of spinner text
- [ ] Consider adding a global state manager (Redux/Zustand) if complexity increases
- [ ] Refactor some shared CSS into CSS modules

## Performance Metrics

### Current (as of v1.1)
- **Startup time**: ~500ms - 1s cold start
- **Memory usage**: ~60-80 MB (GUI + backend + database)
- **CPU usage (idle)**: ~0-1%
- **CPU usage (active)**: ~3-8% (log parsing + chart rendering + draft analysis)
- **Database size**: ~5-100 MB (varies by match count and draft history)
- **Draft analysis latency**: <100ms for card recommendations

### Target (v1.2+)
- Startup: <1s
- Memory: <100 MB
- Idle CPU: <1%
- Active CPU: <10%
- Draft analysis: <50ms (after refactor)

## Test Coverage

### Backend (Go)
- **Coverage**: ~85-90%
- **Test count**: 180+ tests
- **CI**: All tests passing on Linux, macOS, Windows

### Frontend (TypeScript/React)
- **Coverage**: 0% (v1.2 Phase 5 will address this)
- **Test count**: 0
- **CI**: Linting and type-checking only

**Goal**: 70-80% coverage for frontend by end of v1.2

## Dependencies Status

### Backend
- ✅ All dependencies up to date
- ✅ No security vulnerabilities (govulncheck passing)
- ✅ Go 1.21+

### Frontend
- ✅ All dependencies up to date
- ✅ No security vulnerabilities
- ✅ Using React 18, TypeScript 5, Vite 6, Recharts

## Deployment Status

### v1.1 (Current - Ready for Release)
- **Status**: Documentation finalization in progress
- **CI**: All checks passing ✅
- **Platforms**: macOS, Windows, Linux
- **Distribution**: Wails build creates native apps
- **Release notes**: In progress

### Next Release: v1.2
- **Planned**: Q1 2026 (after refactor phases complete)
- **Focus**: Architecture refactor + UI testing
- **Blockers**: v1.1 release must be cut first

## Community & Contributions

### Recent PRs (v1.1)
- 23 PRs merged for v1.1 (November 2025)
- Major features: Draft Assistant, Archetype Dashboard, Format Insights
- 6 bug fixes, 2 documentation PRs

### Open Issues by Priority
- **v1.2 Planned**: 33 issues across 5 phases (Milestones #32-36)
- **High**: None blocking v1.1 release
- **Medium**: Various enhancement requests
- **Low**: Future feature ideas

### Active Contributors
- @RdHamilton (maintainer)
- Community PRs welcome!

## Notes for Next Session

### What we just finished:
- ✅ v1.1 feature development (23 PRs merged)
- ✅ v1.2 project planning (33 issues created across 5 phases)
- ✅ Phase 5 added for UI Testing Infrastructure
- ✅ DEVELOPMENT_STATUS.md updated with v1.1 summary
- 🚧 Documentation finalization for v1.1 release (current task)

### What to do next:
1. **Finalize v1.1 documentation** (current task)
   - ✅ Update DEVELOPMENT_STATUS.md
   - 🚧 Update CHANGELOG.md with v1.1 release notes
   - 🚧 Review README.md for accuracy
   - 🚧 Check ARCHITECTURE_DECISIONS.md for missing ADRs
2. **Cut v1.1 release**
   - Create git tag `v1.1`
   - Create GitHub release with release notes
   - Build and distribute binaries (if applicable)
3. **Begin v1.2 Phase 1** (Facade Pattern)
   - Issue #446: Create internal/gui package structure
   - Issue #447-452: Create domain facades

### Context for Claude:
- **Current Version**: v1.1 (ready for release)
- **Next Version**: v1.2 (architecture refactor + testing)
- We use Wails v2 (Go backend + React frontend)
- **Service-based architecture**: Daemon (background) + GUI (frontend)
- **Daemon mode (recommended)**: 24/7 log monitoring, WebSocket events
- **Standalone mode (fallback)**: GUI with embedded poller
- Follow responsive design principles (see docs/CLAUDE_CODE_GUIDE.md)
- Material Design-inspired dark theme
- All UI must work 800x600 to 1920x1080+
- Real-time updates via WebSocket events (daemon) or EventsEmit (standalone)

### Architecture:
- **Daemon** (`cmd/mtga-companion/daemon.go`): Background service, WebSocket server
- **GUI** (`main.go`, `app.go`): Wails app, IPC client (2,814 lines - target for refactor)
- **Shared** (`internal/mtga/logprocessor/`): Log parsing logic
- **IPC** (`internal/ipc/`): WebSocket client/server
- **Storage** (`internal/storage/`): SQLite database, repositories

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for complete architecture documentation.
