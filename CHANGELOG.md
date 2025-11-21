# Changelog

All notable changes to MTGA-Companion will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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
