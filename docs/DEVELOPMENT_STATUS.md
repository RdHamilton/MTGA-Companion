# Development Status

**Last Updated**: 2025-11-21

## Current Sprint/Focus

### v1.3 - Deck Builder
- **Status**: Planning phase - 10 issues created
- **Current Task**: Ready to begin Phase 1 (Core Infrastructure)
- **Next**: Start with Issue #497 (Deck Infrastructure) and #498 (AI Recommendations)

### v1.2 - Application Refactor & Testing
- **Status**: ✅ COMPLETED - Ready for release
- **Release Date**: 2025-11-21

### v1.1 - Draft Assistant & UI Enhancements
- **Status**: ✅ COMPLETED
- **Release Date**: 2025-11-21

## Recently Completed

### v1.2 - Application Refactor & Testing (November 2025) ✅ COMPLETED

#### Architecture Refactoring (5 PRs - Phases 1-4)

**Phase 1: Facade Pattern**
- ✅ **PR #480** - Implement Facade Pattern for app.go refactoring
  - Reduced app.go from 2,814 lines to ~300 lines
  - Created MatchFacade, DraftFacade, DeckFacade, CardFacade, ExportFacade, SystemFacade
  - Clean separation of concerns with domain-specific facades
  - Single Responsibility Principle for each facade

**Phase 2: Strategy Pattern**
- ✅ **PR #481** - Implement Strategy Pattern for draft insights
  - Pluggable draft format analysis strategies
  - PremierStrategy for human drafters vs QuickStrategy for bot drafters
  - Easy to extend with new formats (Traditional Draft, Sealed, etc.)
  - Independently testable strategies

**Phase 3: Builder Pattern**
- ✅ **PR #482** - Implement Builder Pattern for export operations
  - Fluent API for export configuration
  - Simplified export code with method chaining
  - Centralized validation logic
  - Reduced boilerplate in export methods

**Phase 4: Observer & Command Patterns**
- ✅ **PR #483** - Implement Observer & Command Patterns
  - EventDispatcher for centralized event management
  - WailsObserver, IPCObserver, LoggingObserver for decoupled event handling
  - Command pattern for daemon operations (ReplayCommand, StartupCommand)
  - Reusable, testable, composable operations

#### Testing Infrastructure (4 PRs - Phase 5)

**Component Testing**
- ✅ **PR #485** - Phase 5: Add UI Testing Infrastructure (Issues #469-#471)
  - Vitest + React Testing Library setup
  - Test utilities and mocking system
  - Comprehensive Wails runtime mocks
  - 15+ tests for Footer component
- ✅ **PR #486** - Phase 5: Draft Component Tests (#472)
  - 16 tests for Draft.tsx
  - 17 tests for DraftGrade.tsx
  - 20 tests for DraftStatistics.tsx
  - 19 tests for FormatInsights.tsx
- ✅ **PR #487** - Phase 5: UI Component Testing Infrastructure (#472, #473)
  - 19 tests for Layout.tsx
  - 16 tests for ToastContainer.tsx
  - Fixed TypeScript type errors in mocks
- ✅ **PR #488** - Phase 5: Complete Testing Infrastructure (#469-#478)
  - CI/CD integration with GitHub Actions
  - Coverage reporting via Codecov
  - PR comment coverage summaries
  - E2E testing with Playwright
  - Required status checks for PR merges
  - 478-line comprehensive testing guide (docs/TESTING.md)

**E2E Testing**
- ✅ **Playwright Setup** - E2E testing framework configured
  - Smoke tests for basic application functionality
  - Draft workflow tests
  - Match tracking tests
  - Local development focus (CI integration deferred)

#### Documentation (1 PR)
- ✅ **PR #484** - Document v1.2 design pattern refactoring
  - ADR-011 for design pattern decisions
  - Pattern usage documentation
  - Architecture documentation updates

#### Summary Stats
- **Total PRs**: 10 (5 refactoring + 4 testing + 1 documentation)
- **Total Issues**: 33 (Issues #446-#478)
- **Code Added**: 2,745 lines (patterns) + 1,200+ lines (tests)
- **Code Reduced**: app.go from 2,814 → 300 lines
- **Test Coverage**: Frontend 0% → 62%
- **Test Count**: 122 component tests + E2E suite

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

### 🚧 v1.3 Planning Complete - Ready for Implementation
**Status**: All 10 issues created and assigned to v1.3 Deck Builder project

**Ready to start**:
- ✅ Project #23 created (v1.3 Deck Builder)
- ✅ Milestone #37 created
- ✅ 10 issues created (#489-#498)
- ✅ All issues assigned to project
- 🚧 Begin Phase 1 implementation

## Next Up (Priority Order)

### v1.3 - Deck Builder (10 issues)

**Phase 1: Foundation** (3 issues - Infrastructure)
- **Issue #497**: Core Deck Infrastructure with Draft Integration
  - Database tables for decks, deck_cards, deck_performance, deck_tags
  - Deck model and repository with CRUD operations
  - Draft session linking and validation
  - ~6-8 hours
- **Issue #498**: AI-Powered Card Recommendation Engine
  - Progressive enhancement: rule-based → data-driven → ML
  - RecommendationEngine interface for future extensibility
  - Integration with 17Lands ratings
  - Core monetization feature
  - ~8-10 hours (Phase 1A: rule-based)
- **Issue #489**: Deck Import Parser with Draft Support
  - Parse MTGA Arena format, plain text lists
  - Validate against draft card pool for draft decks
  - ~4-5 hours

**Phase 2: UI Components** (2 issues)
- **Issue #490**: Deck List UI Component
  - Card grouping by type, mana curve, color distribution
  - ~6-8 hours
- **Issue #491**: Card Search with Draft Pool Filtering
  - Autocomplete search with draft mode filtering
  - ~5-6 hours

**Phase 3: Analytics** (1 issue)
- **Issue #492**: Deck Statistics and Analysis
  - Real-time stats, mana curve, format legality checker
  - ~5-6 hours

**Phase 4: Management & Export** (2 issues)
- **Issue #493**: Deck Management (Save/Load/Library)
  - Deck library with filters, draft tracking
  - ~6-7 hours
- **Issue #494**: Deck Export Functionality
  - Export to MTGA, MTGO, text formats
  - ~3-4 hours

**Phase 5: Quality Assurance** (2 issues)
- **Issue #495**: Deck Builder Testing
  - Unit, component, integration, E2E tests
  - ~8-10 hours
- **Issue #496**: Deck Builder Documentation
  - User guides, AI explanation, developer docs
  - ~4-5 hours

**Total Estimated Effort**: ~55-70 hours

### Future Features (Post-v1.3)
- **ML Enhancement** (v1.4.0): Machine learning for personalized deck recommendations
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

### High Priority (✅ Resolved in v1.2)
- ✅ Refactor app.go God Object (2,814 lines) → **Completed in Phase 1: Facade Pattern**
- ✅ Add frontend TypeScript tests → **Completed in Phase 5: UI Testing Infrastructure (62% coverage)**
- ✅ Add E2E tests for GUI workflows → **Completed in Phase 5: Playwright E2E tests**
- ✅ Standardize event handling → **Completed in Phase 4: Observer Pattern**

### Medium Priority (✅ Resolved in v1.2)
- ✅ Pluggable draft format strategies → **Completed in Phase 2: Strategy Pattern**
- ✅ Simplify export operations → **Completed in Phase 3: Builder Pattern**
- ✅ Daemon command structure → **Completed in Phase 4: Command Pattern**

### Low Priority
- [ ] Add loading skeletons instead of spinner text
- [ ] Consider adding a global state manager (Redux/Zustand) if complexity increases
- [ ] Refactor some shared CSS into CSS modules
- [ ] Improve E2E test coverage (currently local development only)

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

### Frontend (TypeScript/React) - ✅ Improved in v1.2
- **Coverage**: 62% (up from 0%)
- **Test count**: 122 component tests + E2E suite
- **Framework**: Vitest + React Testing Library + Playwright
- **CI**: Component tests run on every PR with coverage reporting

**v1.2 Achievement**: Exceeded 60% coverage goal, comprehensive test infrastructure in place

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

### v1.2 (Current - Ready for Release)
- **Status**: ✅ All phases complete, ready for release
- **CI**: All checks passing ✅ (frontend tests, security scans, backend tests)
- **Platforms**: macOS, Windows, Linux
- **Distribution**: Wails build creates native apps
- **Release Date**: 2025-11-21

### Next Release: v1.3
- **Planned**: Q1 2026
- **Focus**: Deck Builder with AI-powered recommendations
- **Key Features**: Draft-based deck building, recommendation engine, deck management

## Community & Contributions

### Recent PRs (v1.2)
- 10 PRs merged for v1.2 (November 2025)
- Major refactoring: Facade, Strategy, Builder, Observer, Command patterns
- Testing infrastructure: 122 component tests + E2E suite
- 1 documentation PR

### Recent PRs (v1.1)
- 23 PRs merged for v1.1 (November 2025)
- Major features: Draft Assistant, Archetype Dashboard, Format Insights
- 6 bug fixes, 2 documentation PRs

### Open Issues by Priority
- **v1.3 Planned**: 10 issues for Deck Builder (Milestone #37, Project #23)
- **High**: None currently
- **Medium**: Various enhancement requests
- **Low**: Future feature ideas

### Active Contributors
- @RdHamilton (maintainer)
- Community PRs welcome!

## Notes for Next Session

### What we just finished:
- ✅ v1.2 complete (10 PRs merged - all 5 phases)
  - Phase 1-4: Architecture refactoring (Facade, Strategy, Builder, Observer, Command)
  - Phase 5: Testing infrastructure (122 component tests + E2E suite)
- ✅ v1.3 project planning (10 issues created for Deck Builder)
- ✅ Documentation updated for v1.2 release
- ✅ CHANGELOG.md updated with v1.2 release notes
- ✅ DEVELOPMENT_STATUS.md updated with v1.2 completion
- 🚧 Ready to create v1.2 release

### What to do next:
1. **Create v1.2 release** (current task)
   - Create git tag `v1.2.0`
   - Create GitHub release with release notes
   - Build and distribute binaries (if applicable)
2. **Begin v1.3 Phase 1** (Deck Builder Foundation)
   - Issue #497: Core Deck Infrastructure with Draft Integration
   - Issue #498: AI-Powered Card Recommendation Engine (rule-based MVP)
   - Issue #489: Deck Import Parser with Draft Support

### Context for Claude:
- **Current Version**: v1.2.0 (ready for release)
- **Next Version**: v1.3 (Deck Builder with AI recommendations)
- We use Wails v2 (Go backend + React frontend)
- **Service-based architecture**: Daemon (background) + GUI (frontend)
- **Daemon mode (recommended)**: 24/7 log monitoring, WebSocket events
- **Standalone mode (fallback)**: GUI with embedded poller
- Follow responsive design principles (see docs/CLAUDE_CODE_GUIDE.md)
- Material Design-inspired dark theme
- All UI must work 800x600 to 1920x1080+
- Real-time updates via WebSocket events (daemon) or EventsEmit (standalone)

### Architecture (v1.2):
- **Daemon** (`cmd/mtga-companion/daemon.go`): Background service, WebSocket server
- **GUI** (`main.go`, `app.go`): Wails app (~300 lines after refactor), delegates to facades
- **Facades** (`internal/gui/*_facade.go`): Domain-specific facades (Match, Draft, Deck, Card, Export, System)
- **Patterns** (`internal/events/`, `internal/commands/`): Observer & Command patterns
- **Strategies** (`internal/mtga/draft/insights/*_strategy.go`): Format-specific analysis
- **Shared** (`internal/mtga/logprocessor/`): Log parsing logic
- **IPC** (`internal/ipc/`): WebSocket client/server
- **Storage** (`internal/storage/`): SQLite database, repositories

See [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md) for complete architecture documentation.
