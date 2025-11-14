# Development Status

**Last Updated**: 2025-11-14

## Current Sprint/Focus

### Post-v0.3.0 Preparation
- **Status**: PR #318 merged successfully to main ✅
- **Next Priority**: Implement rank progression parsing (#317)
- **Alternative**: Cut v0.3.0 release (rank progression can be v0.3.1)

## Recently Completed

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

### 🚧 Rank Progression Parsing (#317) - HIGH PRIORITY
**Status**: Data source research complete, needs implementation

**Current State**:
- Rank progression chart exists in GUI
- Returns empty timeline (rank_history table is empty)
- MTGA logs DO contain rank data in RankUpdated events
- Parser needed to extract rank changes from logs

**Tasks**:
1. Implement RankUpdated event parser in `internal/mtga/logreader/`
2. Store rank snapshots in `rank_history` table
3. Test with real MTGA rank changes
4. Verify chart populates correctly

**Acceptance Criteria**:
- Parse RankUpdated events from MTGA logs
- Store rank changes with timestamp
- Support all formats (constructed, limited)
- Chart displays rank progression over time

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
- **High**: #317 (Rank progression parsing)
- **Medium**: #319 (E2E testing)
- **Low**: Various enhancement requests

### Active Contributors
- @RdHamilton (maintainer)
- Community PRs welcome!

## Notes for Next Session

### What we just finished:
- ✅ PR #318 successfully merged to main
- ✅ Wails React migration complete
- ✅ All documentation updated
- ✅ All CI checks passing
- ✅ Persistent context documentation added (DEVELOPMENT_STATUS.md, ARCHITECTURE_DECISIONS.md)

### What to do next:
1. **Implement rank progression parsing** (#317) - Extract RankUpdated events from MTGA logs
2. **Add E2E tests** (#319) - Playwright for GUI testing
3. **Cut v0.3.0 release** - Consider releasing now or wait for rank progression
4. **Performance testing** - Verify app performs well with large datasets

### Context for Claude:
- We use Wails v2 (Go backend + React frontend)
- Follow responsive design principles (see CLAUDE.md)
- Material Design-inspired dark theme
- All UI must work 800x600 to 1920x1080+
- Real-time updates via EventsEmit/EventsOn pattern
- Deck inference links matches to decks via timestamp proximity
