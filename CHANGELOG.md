# Changelog

All notable changes to MTGA-Companion will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive documentation wiki with 9 pages covering all features
  - Installation guide for macOS and Windows
  - Complete usage guide with examples
  - CLI commands reference
  - Configuration guide
  - Troubleshooting guide
  - Architecture documentation
  - Database schema documentation
  - Development guide
- Technology stack section in README with links to all dependencies
- Enhanced README with expanded feature list and documentation links

### Changed
- Improved README organization with clear sections
- Added direct links to all wiki pages from README

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
