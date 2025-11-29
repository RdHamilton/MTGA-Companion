# Changelog

All notable changes to MTGA-Companion will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.4.0] - 2025-11-29

### Added

**ML-Powered Recommendations (Complete)**
- **Machine Learning Engine** - Intelligent card recommendations using trained ML models (#576, #577, #578)
- **Personal Play Style Learning** - Adapts recommendations based on deck building history and preferences
- **Meta-Aware Suggestions** - Incorporates tournament data and metagame trends into scoring
- **Hybrid Scoring System** - Combines ML predictions with rule-based analysis for optimal results
- **Recommendation Feedback Collection** - Records acceptance/rejection to improve future recommendations (#575)
- **Deck Performance Tracking** - Historical match data for ML training (#573)
- **Deck Archetype Classification** - Automatic archetype detection for decks and draft pools (#574)

**Ollama LLM Integration**
- **Local LLM Support** - Optional Ollama integration for natural language explanations (#579-#584)
- **Automatic Model Pull** - App automatically downloads required models if not present
- **Configurable Model Selection** - Choose from qwen3:8b, llama3.2, mistral, or any Ollama model
- **ML/AI Settings UI** - New settings section for configuring Ollama and ML features (#585)
- **Explanation Generator** - Natural language explanations for why cards are recommended (#586)
- **Template Fallback** - Works without Ollama using template-based explanations

**Metagame Dashboard**
- **Live Meta Data** - Real-time metagame data from MTGGoldfish and MTGTop8 (#587)
- **Archetype Tier Lists** - View Tier 1-4 archetypes with meta share percentages
- **Archetype Detail View** - Click archetypes to see detailed stats, trends, and tier explanations (#681)
- **Tournament Performance** - Track Top 8 placements and tournament wins per archetype
- **Trend Analysis** - See which archetypes are trending up, down, or stable
- **Multi-Format Support** - Standard, Historic, Explorer, Pioneer, and Modern
- **Recent Tournaments** - View recent tournament results with top decks

**Draft Enhancements**
- **Enhanced Synergy Scoring** - Improved draft prediction with better synergy detection (#639)
- **Color Pair Archetypes** - Automatic archetype detection based on top color pairs (#632)
- **Card Type Categorization** - Set guide uses card types for better organization (#633)
- **Keyword Extraction** - Sophisticated keyword analysis for card recommendations (#634)

**WebSocket Improvements**
- **Selective Event Subscription** - Subscribe only to events you need for better performance (#631)
- **Configurable CORS** - WebSocket CORS settings for production deployments (#630)

**E2E Testing**
- **Page-Specific E2E Tests** - Comprehensive Playwright tests for all pages (#679)
- **Charts Tests** - E2E tests for chart components
- **Collection Tests** - E2E tests for collection page
- **Decks Tests** - E2E tests for deck management
- **Draft Tests** - E2E tests for draft functionality
- **Match History Tests** - E2E tests for match history
- **Meta Tests** - E2E tests for metagame dashboard
- **Settings Tests** - E2E tests for settings page

### Fixed

- **Meta Page Data Loading** - Fixed MTGGoldfish HTML parsing for updated site structure (#680)
- **RWMutex Deadlock** - Fixed fatal panic in RefreshAll method due to incorrect mutex usage (#681)
- **Flaky Format Selection Test** - Fixed intermittent test failure in Meta page tests (#681)

### Changed

- **Settings Navigation** - Updated navigation to include ML/AI settings section
- **Card Recommendations** - Enhanced with ML-powered scoring alongside rule-based analysis

### Technical

**Code Quality**
- **Frontend Test Coverage**: 1485+ tests passing
- **Go Tests**: All passing
- **Linter**: 0 issues (golangci-lint)

**New Packages**
- `internal/ml/` - Machine learning engine with model, pipeline, and meta-weighting
- `internal/llm/` - Ollama client and explanation generator
- `internal/archetype/` - Deck archetype classification system
- `internal/feedback/` - Recommendation feedback service
- `internal/meta/` - Metagame data aggregation from multiple sources

### Documentation

- **README Updates** - Added ML features, Ollama setup guide, and metagame dashboard docs
- **Changelog** - Complete v1.4.0 changelog entry

## [1.3.1] - 2025-11-28

### Added

**Collection Tracking Feature (Complete)**
- **Automatic Collection Tracking** - Pure log-based collection tracking from MTGA game data (#617)
- **Collection Page UI** - Visual card browser with full card images, quantity badges, and rarity indicators (#602)
- **Set Completion Tracking** - Track progress toward completing each MTG set with rarity breakdown (#603)
- **Missing Cards Analysis** - Analyze missing cards for decks and sets with wildcard cost breakdown (#604)
- **Collection Change Notifications** - Real-time notifications when collection changes are detected (#605)
- **Card Search in Deck Builder** - Search cards with collection filter to see only owned cards (#626)
- **Collection Component Tests** - Comprehensive test coverage for all collection UI components (#607)
- **Collection User Documentation** - Complete user guide for collection features (#608)

**Cross-Platform Compatibility**
- **Removed .NET Daemon Dependency** - Collection tracking now works natively without the Windows-only .NET daemon (#619)
- **Native Log-Based Tracking** - Cross-platform solution that works on Windows, macOS, and Linux

### Fixed

- **Decks with NULL account_id** - Fixed decks stored with NULL account_id not appearing in deck list (#618)
- **Collection Card Images** - Fixed card images displaying as slivers instead of full images (#625)
- **Collection Card Metadata** - Fixed all cards showing as "Ambush Viper" with broken images (#625)
- **SET Dropdown Empty** - Fixed SET filter dropdown being empty on Collection page (#625)
- **DFC Image Handling** - Fixed double-faced card images using correct card_faces[0].ImageURIs (#625)
- **Card Back Placeholder** - Fixed card back placeholder URL from .jpg to .png (#625)

### Changed

- **Collection Architecture** - Simplified collection tracking to work without external daemon

### Technical

**Code Quality**
- **Frontend Test Coverage**: 1337 tests passing
- **All Go Tests**: Passing
- **Linter**: 0 issues (golangci-lint)

### Documentation

- **COLLECTION.md** - Comprehensive collection features user guide
- **Wiki Updates** - Updated Usage Guide with collection features

## [1.3.0] - 2025-11-25

### Added

**Deck Builder Feature (Complete)**
- **Deck Creation & Management** - Build constructed and draft-based decks with full CRUD operations
- **AI-Powered Card Recommendations** - Intelligent suggestions based on color fit, mana curve, synergy, and card quality (#504)
- **Multiple Import Formats** - Import from Arena, plain text, and other common formats (#503)
- **Multiple Export Formats** - Export to Arena, MTGO, MTGGoldfish, and plain text (#505)
- **Comprehensive Statistics** - Mana curve, color distribution, type breakdown, and land recommendations (#507)
- **Draft Pool Validation** - Draft decks restricted to cards from the associated draft
- **Format Legality Checking** - Validate decks for Standard, Historic, Explorer, Alchemy, Brawl, and Commander
- **Performance Tracking** - Automatic win rate tracking and performance metrics (#506)
- **Deck Library UI** - Advanced filtering by format, source, tags, and performance (#512)
- **Complete Documentation** - Comprehensive deck builder guide with API reference (#509)

**Quest & Statistics Tracking**
- **Quest Gold Calculation** - Accurate gold calculation by parsing actual reward values instead of estimates (#572)
- **Real-Time Draft Updates** - Live draft updates via `draft:updated` events (#524)
- **Set Symbol Display** - Card displays now show set symbols/icons (#548)

**Settings Page Improvements**
- **Collapsible Accordion Navigation** - Organized settings into 7 collapsible sections (#569)
- **URL Hash Navigation** - Direct links to settings sections (e.g., `/settings#connection`, `/settings#17lands`)
- **Keyboard Navigation** - Full accessibility support with ArrowUp/Down, Home/End, Enter/Space
- **LoadingButton Component** - Consistent loading states with spinner animation (#564)
- **Shared Form Components** - Reusable FormGroup, FormSelect, FormInput components (#565)
- **Settings Hooks** - Custom hooks for state management (useSettingsAccordion, useDaemonConnection, etc.) (#566)
- **CSS Extraction** - Moved inline styles to CSS classes for maintainability (#563)

**Testing Infrastructure (Expanded)**
- **Settings.tsx Tests** - 61 comprehensive tests covering all settings functionality (#570)
- **Page Tests** - MatchHistory (48 tests), Quests (52 tests), Decks (34 tests) (#551, #552, #553)
- **Chart Tests** - WinRateTrend, DeckPerformance, RankProgression, FormatDistribution, ResultBreakdown (#550)
- **Core Infrastructure Tests** - App.test.tsx, Layout.test.tsx, context tests (#549)
- **Typed Event Payloads** - Strongly-typed event handlers replacing inline type casts (#547)
- **Unified ConnectionStatus** - Single type definition across Go and TypeScript (#539)
- **LogReplayProgress Type** - Proper typed structure for daemon log replay events (#537)
- **Mock Type Improvements** - Replaced `any` types with proper model types in test mocks (#538)
- **Frontend Test Coverage** - Increased from 62% to 90%+ with 400+ new tests

**Type Safety Improvements**
- **Typed Event Payloads** - StatsUpdatedEvent, RankUpdatedEvent, QuestUpdatedEvent, DraftUpdatedEvent, etc. (#547)
- **ConnectionStatus Type** - Unified type definition generated from Go struct (#539)
- **LogReplayProgress Type** - Proper typed structure for daemon log replay events (#537)

### Fixed

- **Draft Grade Bug** - Fixed draft grade stuck at 59/100 when cards missing ratings (#523)
- **Deck Statistics** - Fixed showing 0 values for multi-colored draft cards (#522)
- **Card Recommendations** - Fixed returning 0 suggestions due to nil card services (#521, #511)
- **Quest Timestamps** - Fixed SQLite timestamp formatting for quest completion tracking (#500)
- **CardSearch Tests** - Fixed async test timing issues (#531)
- **ESLint Errors** - Fixed all ESLint errors and warnings across frontend codebase (#532)

### Changed

- **Removed Events Feature** - Removed unused Events tab and related code (#571)
- **Rank Progression** - Added format selector (Constructed/Limited) to Rank Progression page (#571)
- **Settings Refactoring** - Complete refactoring of Settings.tsx from monolithic to modular components (#554-#568)
- **Debug Code Removal** - Removed all console.log statements and debug panels from Settings (#562)
- **CI Optimization** - Reduced Linux test time from 11min to ~1min (#501)

### Technical

**Code Quality**
- **Frontend Test Coverage**: 90%+ (up from 62%)
- **400+ New Tests**: Comprehensive coverage across all pages and components
- **Type Safety**: Eliminated `any` types in favor of proper model types
- **Settings Refactoring**: Extracted 7 section components, 5 custom hooks, shared form components

**Performance**
- Same performance metrics as v1.2
- Draft analysis: <100ms
- Memory: ~60-80 MB
- Startup: ~500ms-1s

### Documentation

- **DECK_BUILDER.md** - Comprehensive deck builder documentation with API reference (#509)
- **Deck Statistics Testing** - Complete test coverage for deck statistics (#508)

## [1.2.0] - 2025-11-21

### Added

**Testing Infrastructure (Phase 5)**
- **Component Testing Framework** - Vitest + React Testing Library for frontend component tests
- **122 Component Tests** - Comprehensive test coverage for all UI components (Layout, Footer, ToastContainer, Draft pages, Charts)
- **End-to-End Testing** - Playwright E2E testing framework with smoke tests, draft workflow tests, and match tracking tests
- **CI/CD Integration** - Automated frontend testing in GitHub Actions with coverage reporting
- **Coverage Reporting** - Codecov integration with PR comments showing coverage metrics
- **Required Status Checks** - Frontend tests and security scans required for PR merges
- **Test Mocking System** - Comprehensive mocks for Wails runtime (EventsOn, EventsEmit) and app bindings
- **Testing Documentation** - Complete testing guide (docs/TESTING.md) with examples and best practices

### Changed

**Architecture Refactoring**
- **Phase 1: Facade Pattern** - Refactored app.go from 2,814 lines to ~300 lines with domain-specific facades (MatchFacade, DraftFacade, DeckFacade, CardFacade, ExportFacade, SystemFacade)
- **Phase 2: Strategy Pattern** - Pluggable draft format analysis strategies for Premier Draft vs Quick Draft
- **Phase 3: Builder Pattern** - Fluent API for export operations with simplified code
- **Phase 4: Observer Pattern** - EventDispatcher for decoupled event management across frontend, IPC, and logging
- **Phase 4: Command Pattern** - Encapsulated daemon operations as reusable, testable command objects

### Technical

**Code Quality**
- **Frontend Test Coverage**: 62% (up from 0%)
- **Backend Test Coverage**: 85-90% (maintained)
- **Pattern Implementation**: 2,745 lines added (Facades, Strategies, Builders, Observers, Commands)
- **Code Reduction**: app.go reduced from 2,814 lines to ~300 lines
- **All Phases**: Passed linting without issues
- **CI/CD**: Frontend tests, security scans, and coverage reporting integrated

**Performance**
- Same performance metrics as v1.1 (refactoring didn't impact performance)
- Draft analysis: <100ms
- Memory: ~60-80 MB
- Startup: ~500ms-1s

### Documentation

- **TESTING.md** - 478 lines of comprehensive testing guide
- **ARCHITECTURE_DECISIONS.md** - ADR-011 documenting design pattern refactoring
- **CI_README.md** - E2E testing documentation for CI environments
- **Pattern Documentation** - Detailed usage guides for all implemented patterns

## [1.1.0] - 2025-11-21

### Added

**Draft Assistant Features**
- **Type Synergy Detection** - Real-time analysis of card type synergies (Creatures, Instants, Sorceries, etc.) with visual indicators
- **Card Suggestions** - Context-aware card recommendations based on your picked cards
- **Missing Cards Detection** - Automatic identification of cards missing from draft packs for collection tracking
- **Draft Statistics Dashboard** - Real-time mana curve visualization and color distribution as picks are made
- **Draft Deck Win Rate Predictor** - AI-powered prediction of draft deck performance with letter grades (A/B/C/D/F)
- **Draft Grade Breakdown** - Detailed modal showing grade calculation and statistics

**Format Meta Analysis**
- **Format Meta Insights** - Format-wide archetype performance data, best color pairs, and overdrafted colors analysis
- **Archetype Performance Dashboard** - Interactive archetype selection with top cards per archetype
- **Archetype Filtering** - Win rate and popularity-based filtering and sorting
- **Archetype-Specific Card Lists** - View best overall cards, removal, and commons for each archetype

**Historical Draft & Replay**
- **Historical Draft Replay** - View and replay past draft pick sequences
- **Draft Replay Engine** - Infrastructure for simulating MTGA log events for testing
- **GUI Replay Controls** - Play/pause/reset functionality with progress tracking
- **CLI Replay Command** - `replay` command for development and testing

**Log Management**
- **Startup Recovery** - Automatic recovery and import of historical data on application startup
- **Runtime Log Rotation Handling** - Seamless handling of MTGA's UTC_Log rotation at runtime
- **Manual Log Import UI** - Import historical MTGA logs via GUI with batch processing
- **Automatic Log Archival** - Opt-in configurable automatic preservation of historical data
- **Initial Installation Import** - Automatic import of existing logs on first run

**Match Tracking**
- **Match Details View** - Expandable match details with game-by-game breakdown
- **Game Win/Loss Tracking** - Per-game statistics within matches

**UI Enhancements**
- **Card Tier List Visualization** - Visual tier list for draft cards
- **Enhanced Card Images** - Improved card image display in draft interface
- **N/A Grades for Missing Data** - Graceful handling of cards without rating data
- **Performance Monitoring** - Draft overlay performance metrics and monitoring infrastructure

**Data Integration**
- **17Lands Public Datasets** - Migration to 17Lands public API for more reliable card ratings
- **Improved Draft Grading** - Fixed bugs in draft grading algorithm

### Fixed
- **Database Locking** - Fixed database locking issues during rapid draft replay events (#431)
- **Duplicate Replay Controls** - Fixed duplicate replay controls and toast notification spam (#428, #427)
- **Tier List Scrolling** - Fixed scrolling behavior in Draft view tier list (#426)
- **Replay Tool Detection** - Fixed active draft detection in replay testing tool
- **Win Rate Modal Overlay** - Fixed z-index and overlay issues in win rate prediction modal
- **Draft Card Images** - Fixed card image display and N/A grade handling

### Changed
- **Documentation Structure** - Reorganized all documentation into `docs/` directory with comprehensive index
- **Code Cleanup** - Removed 8,488 lines of obsolete Fyne UI and daemon display code

### Technical
- **Performance**: Draft analysis latency <100ms
- **Memory Usage**: ~60-80 MB (increased due to draft assistant features)
- **Backend Tests**: 180+ tests with 85-90% coverage
- **CI/CD**: All checks passing on Linux, macOS, Windows

## [1.0.0] - 2025-10-01

### Added
- **Wails Desktop GUI** - Complete cross-platform desktop application with React + TypeScript
- **Match History Page** - View all matches with filtering and sorting
- **Win Rate Trend Charts** - Line charts visualizing performance over time
- **Deck Performance Charts** - Bar charts showing win rates by deck
- **Rank Progression Tracking** - Real-time rank tracking for Constructed and Limited
- **Format Distribution Charts** - Pie charts showing play patterns across formats
- **Result Breakdown Statistics** - Detailed statistics by format and time period
- **Settings Page** - Database configuration and application settings
- **Real-Time Updates** - Live statistics while playing MTGA
- **Toast Notifications** - Non-intrusive update notifications
- **Persistent Footer** - At-a-glance statistics always visible
- **Service-Based Architecture** - Daemon mode for 24/7 log monitoring with WebSocket events
- **Standalone Mode** - Fallback mode with embedded log poller
- **Responsive Design** - UI adapts from 800x600 to 1920x1080+
- **Dark Theme** - Material Design-inspired dark theme

### Changed
- **Architecture**: Complete migration from CLI to desktop GUI
- **Log Monitoring**: Enhanced with IPC client/server communication
- **UI Framework**: Migrated from Fyne to Wails v2 + React

### Technical
- Wails v2 framework (Go + React)
- React 18, TypeScript 5, Vite 6
- Recharts for data visualization
- WebSocket-based IPC (port 9999)

## [0.1.0] - 2025-01-12

### Added
- **Core Features**
  - Log reading and parsing for MTGA Player.log files
  - Cross-platform support (macOS and Windows)
  - Platform-aware log path detection
  - JSON event parsing

- **Draft Tracking**
  - Record all draft picks with pack context
  - Store draft event information
  - Track draft results (wins/losses)
  - Draft statistics and history

- **Database Storage**
  - SQLite database with auto-migration
  - Schema versioning with golang-migrate
  - Backup and restore functionality
  - Database integrity checks

- **Card Data Integration**
  - 17Lands draft statistics and ratings
  - Scryfall card metadata
  - Unified card data model
  - Automatic data updates for active sets
  - Offline caching with staleness tracking
  - Graceful fallback between sources

- **Statistics & Analytics**
  - Match win rate tracking
  - Format-specific statistics
  - Time-based pattern analysis (hour-of-day, day-of-week)
  - Performance streak tracking (win/loss streaks)
  - Season-over-season comparisons
  - Trend analysis
  - Predictive analytics based on performance trends

- **Export System**
  - CSV and JSON export formats
  - Draft picks export
  - Draft history export
  - Statistics export
  - Streak analysis export
  - Time pattern export
  - Predictive analytics export
  - Flexible filtering (date range, format, event)

- **Set File Management**
  - Download 17Lands set files for any format
  - Support for PremierDraft, QuickDraft, TradDraft
  - Automatic set code validation
  - Local caching and organization

- **Log File Monitoring**
  - File system event monitoring with fsnotify
  - Automatic log rotation detection and handling
  - Incremental log reading (only new entries)
  - Fallback to polling if fsnotify unavailable
  - Performance metrics tracking

- **CLI Commands**
  - `draft-stats` - View draft statistics
  - `export` - Export data in various formats (13+ export types)
  - `sets download` - Download 17Lands set files
  - `sets list` - List downloaded sets
  - `migrate` - Database migration operations
  - `backup` - Backup and restore operations
  - `cards` - Card data operations
  - `deck` - Deck management operations

- **Development Tools**
  - Development script (`scripts/dev.sh`) for builds and checks
  - Testing script (`scripts/test.sh`) for comprehensive testing
  - Code formatting with gofmt and gofumpt
  - Linting with golangci-lint
  - Race detection in tests
  - Coverage reporting

### Technical Details
- Go 1.23+ support
- SQLite 3 database
- Pure Go SQLite driver (modernc.org/sqlite)
- Cross-platform file system monitoring (fsnotify)
- Database migration system (golang-migrate)
- Thread-safe operations with sync.RWMutex
- Prepared statements for SQL injection prevention
- Indexed database queries for performance
- Connection pooling via database/sql

### Performance
- Efficient log monitoring (<1% CPU idle)
- Incremental log reading (only new bytes)
- Indexed database queries
- Streaming exports for large datasets
- Memory footprint: ~25-30 MB active

### Security
- Local-only storage (no cloud uploads)
- Read-only log file access
- SQL injection prevention with prepared statements
- Input validation on all CLI flags
- No PII in logs

## Project Status

### Production Ready ✅
- Log reading and parsing
- Database storage and migrations
- Card data integration
- Statistics tracking
- Export functionality
- Set file management

### In Development 🚧
- Draft overlay with real-time card ratings
- GUI application with Fyne

## Links

- [Documentation Wiki](https://github.com/RdHamilton/MTGA-Companion/wiki)
- [GitHub Repository](https://github.com/RdHamilton/MTGA-Companion)
- [Issue Tracker](https://github.com/RdHamilton/MTGA-Companion/issues)
- [Discussions](https://github.com/RdHamilton/MTGA-Companion/discussions)

---

**Note**: MTGA-Companion is not affiliated with, endorsed by, or sponsored by Wizards of the Coast. Magic: The Gathering Arena and its associated trademarks are property of Wizards of the Coast LLC.
