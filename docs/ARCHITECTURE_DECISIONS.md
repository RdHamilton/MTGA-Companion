# Architecture Decision Records (ADR)

This document records significant architectural decisions made during the development of MTGA-Companion, including the rationale and consequences of each decision.

---

## ADR-001: SQLite for Local Database (2024-01)

**Status**: ✅ Accepted

**Context**:
- Need local database for match history, drafts, and statistics
- Users want privacy (no cloud uploads)
- Application is single-user, desktop-focused
- Data size: thousands of matches, hundreds of drafts

**Decision**: Use SQLite 3 as the local database

**Rationale**:
- **Zero-configuration**: No database server to install or configure
- **Portable**: Single file database, easy to backup
- **Performance**: Sufficient for single-user workloads
- **Mature**: Battle-tested, stable, well-documented
- **Cross-platform**: Works identically on macOS, Windows, Linux
- **Small footprint**: ~1-100 MB for typical usage
- **ACID compliance**: Reliable transactions

**Alternatives Considered**:
- **PostgreSQL**: Overkill for single-user, requires server
- **BoltDB**: Less flexible query capabilities
- **JSON files**: No query optimization, difficult to maintain consistency

**Consequences**:
- ✅ Simple deployment (just copy the .db file)
- ✅ Easy backups (copy the file)
- ✅ No network security concerns
- ⚠️ Limited to single-user concurrent access
- ⚠️ Scalability limit around ~100GB (not a concern for this use case)

---

## ADR-002: Pure Go SQLite Driver (2024-01)

**Status**: ✅ Accepted

**Context**:
- Need SQLite driver for Go
- Two main options: mattn/go-sqlite3 (CGo) vs modernc.org/sqlite (Pure Go)
- Cross-compilation with CGo is complex
- Windows requires MinGW for CGo

**Decision**: Use modernc.org/sqlite (pure Go driver)

**Rationale**:
- **No CGo**: Simplifies cross-compilation significantly
- **No external dependencies**: No need for gcc/MinGW on Windows
- **Faster builds**: No C compilation step
- **Same SQL interface**: Drop-in replacement for database/sql
- **Performance**: Comparable to CGo version for our use case

**Alternatives Considered**:
- **mattn/go-sqlite3**: Most popular, but CGo complexity not worth it

**Consequences**:
- ✅ Easy cross-compilation (GOOS=windows go build works)
- ✅ Simpler CI/CD (no build tools needed)
- ✅ Faster build times
- ⚠️ Slightly slower than CGo version (not noticeable in practice)

---

## ADR-003: Golang-migrate for Database Migrations (2024-02)

**Status**: ✅ Accepted

**Context**:
- Need schema versioning for database evolution
- Users should get automatic migrations on app updates
- Need both up and down migrations
- Must be reliable and well-tested

**Decision**: Use golang-migrate/migrate for database migrations

**Rationale**:
- **Industry standard**: Widely used, well-maintained
- **Version control**: Migrations are numbered SQL files
- **Reversible**: Support for up and down migrations
- **CLI + library**: Can run from code or command line
- **Multiple sources**: Supports embedded files (go:embed)
- **Safe**: Tracks applied migrations, prevents re-running

**Consequences**:
- ✅ Reliable schema evolution
- ✅ Easy to add new migrations
- ✅ Version control friendly (SQL files in git)
- ⚠️ Must never modify existing migrations after release
- ⚠️ Must provide down migrations for reversibility

---

## ADR-004: Wails v2 for Desktop GUI (2024-11)

**Status**: ✅ Accepted

**Context**:
- Need cross-platform desktop GUI
- Initially used Fyne for draft overlay
- Want richer UI capabilities and modern UX
- React ecosystem provides excellent UI libraries

**Decision**: Migrate from Fyne to Wails v2 (Go backend + React frontend)

**Rationale**:
- **Separation of concerns**: Go handles data/logic, React handles UI
- **Rich UI ecosystem**: Access to React, Recharts, Material-UI, etc.
- **Native webview**: No Electron overhead, smaller binaries
- **Type safety**: Auto-generated TypeScript bindings for Go methods
- **Developer experience**: Hot reload, familiar React patterns
- **Responsive design**: CSS Grid/Flexbox easier than Fyne layouts
- **Larger talent pool**: React developers more common than Fyne

**Alternatives Considered**:
- **Fyne**: Good for simple UIs, but complex layouts are difficult
- **Electron**: Too heavy (~200MB), slow startup
- **Flutter Desktop**: Less mature, different language (Dart)
- **Qt/QML**: C++ complexity, licensing concerns

**Consequences**:
- ✅ Beautiful, responsive UI with modern design
- ✅ Easy to add new visualizations (Recharts library)
- ✅ Better developer experience (React ecosystem)
- ✅ Type-safe backend/frontend communication
- ⚠️ Requires Node.js for frontend development
- ⚠️ Larger learning curve (Go + TypeScript + React)
- ⚠️ Fyne draft overlay moved to legacy CLI mode

**Migration**: Completed in PR #318 (November 2024)

---

## ADR-005: React + TypeScript for Frontend (2024-11)

**Status**: ✅ Accepted

**Context**:
- Chosen Wails v2, need to select frontend framework
- Could use vanilla JS, React, Vue, Svelte, etc.
- Want type safety and good tooling

**Decision**: Use React 18 + TypeScript for the frontend

**Rationale**:
- **Type safety**: TypeScript prevents runtime errors
- **Component model**: React's component model fits our page-based UI
- **Hooks**: useState/useEffect perfect for our data-fetching patterns
- **Ecosystem**: Excellent libraries (React Router, Recharts)
- **Documentation**: Extensive resources and community support
- **Team familiarity**: React is widely known
- **Wails support**: First-class TypeScript bindings generation

**Alternatives Considered**:
- **Vue 3**: Comparable, but React has better charting libraries
- **Svelte**: Less mature ecosystem
- **Vanilla JS**: No type safety, error-prone

**Consequences**:
- ✅ Type-safe frontend code
- ✅ Auto-completion in IDE
- ✅ Catch errors at compile time
- ✅ Rich component libraries available
- ⚠️ Slightly larger bundle size than vanilla JS

---

## ADR-006: Dark Theme as Primary UI (2024-11)

**Status**: ✅ Accepted

**Context**:
- Need to choose color scheme for GUI
- Users play MTGA (and use companion) for hours
- Gaming applications typically use dark themes

**Decision**: Dark theme as primary (only) theme

**Rationale**:
- **Eye strain**: Dark theme reduces eye strain during long sessions
- **Gaming context**: MTGA itself uses dark theme
- **Modern aesthetic**: Dark themes are currently preferred
- **Battery savings**: On OLED screens, dark pixels use less power
- **Focus on content**: Dark backgrounds make charts/data stand out

**Color Palette**:
- Background: `#1e1e1e`
- Secondary background: `#2d2d2d`
- Primary accent: `#4a9eff`
- Text: `#ffffff`
- Success (win): `#4caf50`
- Error (loss): `#f44336`

**Alternatives Considered**:
- **Light theme**: Too bright for long gaming sessions
- **Theme toggle**: Adds complexity, decided to focus on one theme done well

**Consequences**:
- ✅ Reduced eye strain
- ✅ Consistent aesthetic
- ✅ Simpler CSS (no theme switching)
- ⚠️ Some users may prefer light theme (future enhancement)

---

## ADR-007: Responsive Design over Fixed Layouts (2024-11)

**Status**: ✅ Accepted

**Context**:
- Desktop application with resizable window
- Users have different monitor sizes (1080p, 1440p, 4K)
- Users may want to run app alongside MTGA

**Decision**: Fully responsive design with minimum 800x600, optimal 1024x768-1920x1080

**Rationale**:
- **Flexibility**: Works on any screen size
- **Side-by-side**: Users can run companion next to MTGA
- **Future-proof**: Adapts to new screen sizes
- **CSS Grid/Flexbox**: Modern CSS makes this easy
- **Better UX**: Content adapts instead of being cut off

**Guidelines**:
- All layouts use flexbox or CSS Grid
- Tables scroll horizontally if needed
- Charts use ResponsiveContainer
- Filter rows wrap on small screens
- Minimum 800x600 support

**Alternatives Considered**:
- **Fixed 1280x800**: Too rigid, doesn't adapt
- **Minimum 1920x1080**: Excludes smaller screens

**Consequences**:
- ✅ Works on any screen size
- ✅ Better user experience
- ✅ Future-proof
- ⚠️ More CSS complexity
- ⚠️ More testing needed (different sizes)

---

## ADR-008: Real-time Updates via Event System (2024-11)

**Status**: ✅ Accepted

**Context**:
- Need to update GUI when new matches are detected
- Polling from frontend would be inefficient
- Wails provides event system for backend→frontend communication

**Decision**: Use Wails event system for real-time updates

**Pattern**:
```go
// Backend emits event when data changes
runtime.EventsEmit(ctx, "stats:updated", data)
```

```typescript
// Frontend listens and refreshes
EventsOn('stats:updated', () => { loadData() })
```

**Rationale**:
- **Efficient**: Backend pushes updates only when needed
- **Decoupled**: Frontend doesn't need to poll
- **Reactive**: UI updates automatically
- **Simple**: Built into Wails, no additional libraries

**Alternatives Considered**:
- **Frontend polling**: Wasteful, adds latency
- **WebSockets**: Overkill for single-user desktop app

**Consequences**:
- ✅ Instant updates when matches detected
- ✅ No polling overhead
- ✅ Clean separation of concerns
- ✅ Easy to add new event types

---

## ADR-009: Deck Inference via Timestamp Proximity (2024-11)

**Status**: ✅ Accepted (with lessons learned)

**Context**:
- MTGA logs don't include deck_id in match events
- Need to link matches to decks for deck performance stats
- Deck data includes LastPlayed timestamp

**Decision**: Link matches to decks based on timestamp proximity (within 24 hours)

**Algorithm**:
1. For each match without deck_id
2. Find deck with closest LastPlayed timestamp
3. If within 24 hours, link match to deck
4. Only link matches with NULL deck_id (don't re-link)

**Rationale**:
- **Best available heuristic**: Closest timestamp is most likely deck used
- **24-hour window**: Captures same-day play session
- **Simple**: No machine learning or complex logic needed
- **Good enough**: Works in practice for most users

**Lessons Learned** (PR #318):
- ⚠️ **Never re-link existing matches**: Originally tried to re-link all recent matches, caused matches to switch decks incorrectly
- ✅ **Only link NULL deck_id**: Prevents thrashing
- ✅ **Don't be too aggressive**: Simpler is better

**Consequences**:
- ✅ Deck performance stats work without manual linking
- ✅ Mostly accurate for single-deck sessions
- ⚠️ May mis-link if switching decks frequently
- ⚠️ Doesn't work if user plays multiple decks in quick succession

**Future**: Could improve with deck composition matching (compare decklist to cards played in match)

---

## ADR-010: Auto-start Poller in GUI Mode (2024-11)

**Status**: ✅ Accepted

**Context**:
- GUI application should "just work"
- Users shouldn't need to manually start monitoring
- Real-time updates are core value proposition

**Decision**: Auto-start log file poller when GUI launches

**Implementation**:
```go
func (a *App) startup(ctx context.Context) {
    // Auto-initialize database
    a.Initialize(defaultDBPath)

    // Auto-start poller for real-time updates
    a.StartPoller()
}
```

**Rationale**:
- **Better UX**: No configuration needed
- **Expectations**: Desktop apps should auto-detect and monitor
- **Core feature**: Real-time updates are main GUI benefit
- **Fallback**: Shows error if log file not found (user can configure)

**Alternatives Considered**:
- **Manual start**: Requires user action, poor UX
- **Settings toggle**: Extra complexity for core feature

**Consequences**:
- ✅ Works out-of-box for most users
- ✅ Real-time updates immediately available
- ⚠️ Shows warning if MTGA not installed (expected)
- ⚠️ Runs continuously (minimal CPU/battery impact)

---

## ADR-011: Design Pattern Refactoring (v1.2) (2024-11)

**Status**: ✅ Accepted

**Context**:
- Codebase grew organically with increasing complexity
- Event handling was scattered with manual `EventsEmit` calls
- No consistent patterns for complex object creation
- Daemon operations were difficult to test and reuse
- Format-specific logic mixed with general logic
- Need better separation of concerns and maintainability

**Decision**: Implement five design patterns and comprehensive testing in a phased refactoring:
1. **Facade Pattern** - Simplify frontend/backend interface
2. **Strategy Pattern** - Format-specific analysis algorithms
3. **Builder Pattern** - Complex object construction with fluent API
4. **Observer Pattern** - Decouple event emission from handlers
5. **Command Pattern** - Encapsulate operations as objects
6. **Testing Infrastructure** - Component and E2E testing with CI/CD integration

**Rationale**:

**Phase 1 - Facade Pattern**:
- Simplifies `app.go` from 2000+ lines to thin delegation layer
- Groups related operations into domain-specific facades
- Clear separation between frontend and backend
- Single Responsibility Principle for each facade

**Phase 2 - Strategy Pattern**:
- Premier Draft vs Quick Draft need different analysis (humans vs bots)
- Eliminates format-checking conditionals throughout code
- Easy to add new formats (Traditional Draft, Sealed, etc.)
- Strategies are independently testable

**Phase 3 - Builder Pattern**:
- Export operations have many configuration options
- Fluent API makes intent clear: `.WithFormat().WithPrettyJSON().Build()`
- Centralizes validation logic
- Reduces boilerplate in export methods

**Phase 4 - Observer Pattern**:
- 15+ manual `EventsEmit` calls scattered throughout codebase
- Need to broadcast same events to multiple destinations (frontend, IPC, logs)
- Adding analytics/metrics would require touching many files
- EventDispatcher centralizes all event handling

**Phase 4 - Command Pattern**:
- Daemon operations need retry logic, logging, history
- Hard to test IPC operations in isolation
- Want undo capability for certain operations
- Commands are reusable and composable

**Phase 5 - Testing Infrastructure**:
- Frontend had 0% test coverage (risky for refactoring)
- No automated testing for UI components
- Manual testing doesn't scale
- Need confidence in refactored code
- CI/CD should catch regressions early
- Component tests verify behavior, E2E tests verify workflows

**Implementation Details**:
```
internal/
├── commands/           # Command pattern (Phase 4)
│   ├── command.go     # Interface & executor
│   ├── replay_command.go
│   └── startup_command.go
├── events/            # Observer pattern (Phase 4)
│   ├── dispatcher.go  # EventDispatcher
│   └── observers.go   # WailsObserver, IPCObserver, LoggingObserver
├── export/
│   └── builder.go     # Builder pattern (Phase 3)
├── gui/               # Facade pattern (Phase 1)
│   ├── *_facade.go    # Domain-specific facades
│   └── services.go    # Shared services
└── mtga/draft/insights/
    ├── strategy.go            # Strategy interface (Phase 2)
    ├── premier_strategy.go    # Premier Draft strategy
    └── quick_strategy.go      # Quick Draft strategy

frontend/               # Testing infrastructure (Phase 5)
├── src/
│   ├── test/
│   │   ├── setup.ts           # Vitest configuration
│   │   ├── utils/
│   │   │   └── testUtils.tsx  # Custom render with router
│   │   └── mocks/
│   │       ├── wailsApp.ts    # Mock Go backend functions
│   │       └── wailsRuntime.ts # Mock Wails runtime
│   └── components/
│       ├── *.test.tsx         # 122 component tests
│       └── ...
├── tests/
│   └── e2e/
│       ├── smoke.spec.ts      # Basic E2E smoke tests
│       ├── draft-workflow.spec.ts  # Draft workflow tests
│       └── match-tracking.spec.ts  # Match history tests
├── vite.config.ts             # Vitest config with coverage
└── playwright.config.ts       # Playwright E2E config
```

**Alternatives Considered**:
- **No refactoring**: Technical debt would continue to accumulate
- **Big bang refactoring**: Too risky, prefer incremental approach
- **Different patterns**:
  - Factory pattern considered but Strategy is more flexible
  - Singleton for EventDispatcher considered but instance-based is more testable

**Consequences**:

✅ **Positive Outcomes**:
- Reduced `app.go` from 2000+ to ~200 lines
- Clear separation of concerns with facades
- Easy to add new formats with Strategy pattern
- Fluent export API improves code readability
- Centralized event handling via Observer pattern
- Testable, reusable operations via Command pattern
- **Frontend test coverage**: 0% → 62% (122 component tests)
- **CI/CD integration**: Automated testing on every PR
- **Comprehensive test infrastructure**: Vitest + Playwright + mocks
- 2,745+ lines of new pattern implementations
- 1,200+ lines of test code
- Better maintainability and extensibility

⚠️ **Trade-offs**:
- More files to navigate (organized by pattern/domain)
- Slight learning curve for contributors (documented in `docs/CLAUDE.md` and `docs/TESTING.md`)
- Pattern overhead for simple operations (but worth it for consistency)
- Test maintenance overhead (but prevents regressions)

✅ **Code Quality**:
- All phases passed linting without issues
- Consistent code formatting with `gofumpt`
- Comprehensive documentation of patterns and testing
- No breaking changes to external API
- Required status checks enforce quality

**Metrics**:
- Phase 1: 690 lines added, 314 removed
- Phase 2: 738 lines added, 308 removed
- Phase 3: 298 lines added, 17 removed
- Phase 4: 1,019 lines added, 67 removed
- Phase 5: 1,200+ lines of test code added
- **Total**: 3,945+ lines added, 706 removed (net +3,239)

**Related**:
- PR #480 (Phase 1: Facade Pattern)
- PR #481 (Phase 2: Strategy Pattern)
- PR #482 (Phase 3: Builder Pattern)
- PR #483 (Phase 4: Observer & Command Patterns)
- PR #485 (Phase 5: UI Testing Infrastructure - Foundation)
- PR #486 (Phase 5: Draft Component Tests)
- PR #487 (Phase 5: UI Component Testing)
- PR #488 (Phase 5: Complete Testing Infrastructure)
- PR #484 (Documentation)
- Issues #446-#478 (all refactoring and testing tasks)
- Documentation: `docs/CLAUDE.md` (pattern usage guide)
- Documentation: `docs/TESTING.md` (478-line testing guide)

**Future Enhancements**:
- Add more strategies for Traditional Draft, Sealed formats
- Implement analytics observer for usage metrics
- Add command history UI for debugging
- Create more builders for complex query construction

---

## Template for Future ADRs

```markdown
## ADR-XXX: [Decision Title] (YYYY-MM)

**Status**: 🚧 Proposed / ✅ Accepted / ❌ Rejected / ⚠️ Deprecated

**Context**:
- What is the issue/problem we're facing?
- Why does this decision need to be made?

**Decision**: [Clear statement of the decision]

**Rationale**:
- Why this decision?
- What are the key benefits?
- What makes this the best option?

**Alternatives Considered**:
- Option 1: Reason not chosen
- Option 2: Reason not chosen

**Consequences**:
- ✅ Positive outcomes
- ⚠️ Trade-offs
- ❌ Negative impacts (if any)

**Related**: [Links to PRs, issues, discussions]
```

---

## Index of Decisions

1. **ADR-001**: SQLite for Local Database
2. **ADR-002**: Pure Go SQLite Driver
3. **ADR-003**: Golang-migrate for Migrations
4. **ADR-004**: Wails v2 for Desktop GUI
5. **ADR-005**: React + TypeScript Frontend
6. **ADR-006**: Dark Theme as Primary UI
7. **ADR-007**: Responsive Design Principles
8. **ADR-008**: Real-time Updates via Events
9. **ADR-009**: Deck Inference Algorithm
10. **ADR-010**: Auto-start Poller in GUI
11. **ADR-011**: Design Pattern Refactoring (v1.2)

---

**Note**: When making new architectural decisions, add them to this file following the template above.
