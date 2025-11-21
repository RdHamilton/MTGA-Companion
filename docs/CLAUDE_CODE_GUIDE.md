# Claude Code Guide

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Documentation Organization

**IMPORTANT**: All technical documentation (`.md` files) MUST be placed in the `docs/` directory, with the following exceptions:

**Root-level documentation (required by GitHub):**
- `README.md` - Project overview and quick start
- `CHANGELOG.md` - Version history and release notes
- `CONTRIBUTING.md` - Contribution guidelines
- `CODE_OF_CONDUCT.md` - Community code of conduct
- `SECURITY.md` - Security policies

**Private/local documentation (in `.gitignore`):**
- `CLAUDE.md` - Your local copy of development notes (NOT tracked in git)
- `.reference-notes.md` - Private reference notes (NOT tracked in git)

**All other documentation goes in `docs/`:**
- Architecture documentation
- Development guides
- API specifications
- Research notes
- Migration guides
- Technical specifications

**When creating new documentation:**
1. Place it in `docs/` directory
2. Add it to `docs/README.md` index
3. Link to it from relevant documentation
4. Use clear, descriptive filenames

## Documentation Maintenance Instructions

**IMPORTANT**: As you work with the user, you MUST proactively maintain these documentation files:

### 1. Update docs/DEVELOPMENT_STATUS.md
**When to update**: After completing any significant work

**What to update**:
- Move completed items from "In Progress" to "Recently Completed"
- Update "In Progress" section when starting new work
- Update "Next Up" if priorities change
- Add new issues to "Known Issues" when discovered
- Update "Notes for Next Session" at the end of each session
- Update "Last Updated" date at the top

**How to update**: Use the Edit tool to modify `docs/DEVELOPMENT_STATUS.md` with the latest status

**Example scenarios**:
- ✅ Just merged a PR → Move from "In Progress" to "Recently Completed"
- ✅ Started implementing a feature → Add to "In Progress" section
- ✅ Found a bug → Add to "Known Issues"
- ✅ Ending session → Update "Notes for Next Session"

### 2. Update docs/ARCHITECTURE_DECISIONS.md
**When to update**: When making or discussing any architectural decision

**What to update**:
- Add new ADR when you make a significant architectural choice
- Use the template at the bottom of the file
- Increment the ADR number appropriately
- Update the index at the bottom
- Change status from "Proposed" to "Accepted" when decision is finalized

**Example scenarios**:
- ✅ Choosing a new library → Document why in new ADR
- ✅ Changing architecture pattern → Create ADR explaining rationale
- ✅ Deciding on UI approach → Record decision with alternatives considered
- ✅ Database schema change approach → Document reasoning

**What qualifies as "architectural"**:
- Technology choices (libraries, frameworks, databases)
- Design patterns adopted
- Major refactoring decisions
- UI/UX paradigm shifts
- Data flow changes
- Security decisions
- Performance trade-offs

### 3. Update docs/CLAUDE_CODE_GUIDE.md (this file)
**When to update**: When the architecture, workflow, or standards change

**What to update**:
- Documentation Organization section when documentation structure changes
- Technology Stack section when dependencies change
- Project Structure when files/folders reorganize
- Architecture section when patterns change
- Coding Principles if new standards adopted
- Development Commands if workflow changes

**Example scenarios**:
- ✅ Added new npm script → Update Development Commands
- ✅ Changed folder structure → Update Project Structure
- ✅ Adopted new coding pattern → Update Coding Principles
- ✅ Changed build process → Update Development Commands

### 4. Update README.md
**When to update**: When user-facing features or setup changes

**What to update**:
- Features section when new capabilities added
- Installation if setup process changes
- Usage if commands change
- Technology Stack if major dependencies change

### How to Remember
At the **end of each significant task** or **end of session**, ask yourself:
1. "Did we complete something?" → Update `docs/DEVELOPMENT_STATUS.md`
2. "Did we make an architectural decision?" → Update `docs/ARCHITECTURE_DECISIONS.md`
3. "Did the architecture/workflow change?" → Update `docs/CLAUDE_CODE_GUIDE.md`
4. "Did user-facing features change?" → Update `README.md`
5. "Did we add new documentation?" → Update `docs/README.md` index

**Do this automatically without being prompted.** The user should not have to ask for documentation updates.

## Project Overview

MTGA-Companion is a **service-based application** with two components:

1. **Headless Daemon** (Go) - Background service that monitors MTGA logs, collects data, and stores in SQLite
2. **Desktop GUI** (Wails v2 + React) - Cross-platform UI that connects to daemon via WebSocket for real-time updates

**Architecture**: Service-oriented, NOT monolithic. Data collection (daemon) is separated from data presentation (GUI).

## Workflow and Issue Management

**IMPORTANT**: All work must be tracked through GitHub issues and the project board.

### Issue-Driven Development
1. **No Work Without Tickets**: Never implement features, fixes, or changes without a corresponding GitHub issue
2. **Issue First**: If a task doesn't have an issue, create one before starting work
3. **Link Everything**: All PRs must reference their associated issue (e.g., "Closes #42")

### Project Board Process
The project uses GitHub Projects for tracking work: https://github.com/users/RdHamilton/projects/1

**Issue Lifecycle**:
1. **Todo** - Issue is created and ready to be worked on
2. **In Progress** - Actively working on the issue (move when you start)
3. **Done** - Issue is completed and PR is merged (GitHub auto-moves when closed)

**Before Starting Work**:
- Check the project board for available issues
- Verify the issue has clear acceptance criteria
- Ensure you understand the requirements
- Move the issue to "In Progress"

**During Development**:
- Keep the issue updated with progress notes
- Reference the issue number in all commits (e.g., "#15: Implement poller")
- Update the issue if you discover new requirements or blockers

**Completing Work**:
- Ensure all acceptance criteria are met
- Create PR with "Closes #N" in description
- Issue automatically moves to "Done" when PR is merged

### Issue Priority and Phases

**Priority Levels**:
- **High**: Critical infrastructure or blocking work
- **Medium**: Core features and important improvements
- **Low**: Nice-to-have features and enhancements

**Implementation Phases**:
- **Phase 1: Foundation** - Database, migrations, core infrastructure (#18, #19)
- **Phase 2: Core Features** - Main user-facing features (#11, #15, #16, #17)
- **Phase 3: Advanced** - Polish, analytics, and advanced features (#12, #13, #14)

**Work Order**:
- Prioritize Phase 1 (Foundation) issues first - everything depends on these
- Complete database (#18) and migrations (#19) before persistent storage features
- Phase 2 features can be worked on in parallel after Phase 1 completes
- Phase 3 features require Phase 2 completion

### Database and Migrations

**Technology Stack**:
- **Database**: SQLite3
- **Migrations**: `golang-migrate/migrate` (gomigrate)

**Migration Guidelines**:
- All schema changes must use gomigrate migrations
- Never modify existing migrations after they're merged
- Always provide both up and down migrations
- Test migrations on a copy of production data
- See issue #19 for detailed migration practices

## Development Commands

### Workflow Scripts

Two helper scripts are available to streamline development and testing:

**Development Script** (`./scripts/dev.sh`)
```bash
./scripts/dev.sh           # Run all checks and build
./scripts/dev.sh fmt       # Format code
./scripts/dev.sh vet       # Run go vet
./scripts/dev.sh lint      # Run golangci-lint
./scripts/dev.sh check     # Run fmt, vet, and lint
./scripts/dev.sh build     # Build the application
```

**Testing Script** (`./scripts/test.sh`)
```bash
./scripts/test.sh                    # Run all tests with race detection
./scripts/test.sh unit               # Run unit tests
./scripts/test.sh coverage           # Run tests with coverage report
./scripts/test.sh race               # Run tests with race detection
./scripts/test.sh verbose            # Run tests with verbose output
./scripts/test.sh bench              # Run benchmarks
./scripts/test.sh specific -name TestName -pkg ./internal/mtga
```

### Initial Setup
```bash
# Initialize Go module (if not already done)
go mod init github.com/ramonehamilton/MTGA-Companion

# Download dependencies
go mod download

# Tidy up dependencies
go mod tidy
```

### Building
```bash
# Build the application
go build -o bin/mtga-companion ./cmd/mtga-companion

# Build for specific platforms
GOOS=windows GOARCH=amd64 go build -o bin/mtga-companion.exe ./cmd/mtga-companion
GOOS=darwin GOARCH=amd64 go build -o bin/mtga-companion ./cmd/mtga-companion
GOOS=linux GOARCH=amd64 go build -o bin/mtga-companion ./cmd/mtga-companion
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests with verbose output
go test -v ./...

# Run specific test
go test -v -run TestName ./path/to/package

# Run tests with race detection
go test -race ./...
```

### Running
```bash
# Run without building
go run ./cmd/mtga-companion

# Run with flags
go run ./cmd/mtga-companion -flag=value
```

### Code Quality
```bash
# Format code
go fmt ./...

# Run linter (requires golangci-lint)
golangci-lint run

# Vet code
go vet ./...
```

## Architecture

### Project Context
This is a companion application for MTGA with a modern desktop GUI, which requires:
- **Log file parsing**: Reading and interpreting MTGA log files to track game state
- **Desktop GUI**: Cross-platform desktop application with Wails + React
- **Real-time updates**: Live data updates while MTGA is running
- **Data aggregation**: Tracking statistics, match history, decks, and game analytics

### Technology Stack

**Backend (Go)**:
- **Wails v2** - Desktop application framework (Go + Web UI)
- **SQLite** - Local database storage
- **Log polling** - Real-time MTGA log file monitoring

**Frontend (React + TypeScript)**:
- **React 18** - UI library with hooks
- **TypeScript** - Type-safe JavaScript
- **React Router** - Client-side routing
- **Recharts** - Data visualization and charting
- **Vite** - Build tool and dev server

### Project Structure
```
MTGA-Companion/
├── main.go                     # Wails entry point
├── app.go                      # Go backend methods exposed to frontend
├── frontend/                   # React frontend application
│   ├── src/
│   │   ├── components/        # Reusable React components
│   │   │   ├── Layout.tsx    # App layout with navigation
│   │   │   ├── Footer.tsx    # Statistics footer
│   │   │   └── ToastContainer.tsx
│   │   ├── pages/            # Page components (routes)
│   │   │   ├── MatchHistory.tsx
│   │   │   ├── WinRateTrend.tsx
│   │   │   ├── DeckPerformance.tsx
│   │   │   └── Settings.tsx
│   │   ├── App.tsx           # Root component with routing
│   │   └── main.tsx          # Frontend entry point
│   ├── wailsjs/              # Auto-generated Wails bindings
│   │   ├── go/              # Go method bindings
│   │   └── runtime/         # Wails runtime functions
│   ├── package.json
│   └── vite.config.ts
├── internal/                  # Private application code
│   ├── gui/                  # GUI-specific backend code
│   ├── mtga/                 # MTGA-specific logic
│   │   ├── logreader/       # Log parsing
│   │   └── draft/           # Draft overlay
│   └── storage/             # Database and persistence
│       ├── models/          # Data models
│       └── repository/      # Data access layer
├── cmd/                      # CLI application (legacy)
│   └── mtga-companion/
└── scripts/                  # Build and development scripts
```

### Platform Considerations
MTGA runs on both macOS and Windows. This application:
- **Cross-platform GUI**: Wails compiles to native apps on macOS, Windows, and Linux
- **Platform-agnostic**: Most code is platform-independent
- **Platform-specific**: Log file paths and file system operations use platform detection
- **Native performance**: Wails uses the system's native webview (no Electron overhead)

### Service-Based Architecture

**CRITICAL: This is NOT a monolithic architecture.** MTGA-Companion uses a **service-oriented design** with clear separation of concerns.

**Architecture Decision**: Separation of data collection (daemon) from data presentation (GUI)

**Design Principle**:
- **Daemon = Data Collection** - Polls logs, stores data, emits events
- **GUI = Data Presentation** - Displays data, handles UI logic, user interaction
- **Never mix these concerns** - The daemon must remain headless with no terminal display/charts

---

**Components**:

**1. Headless Daemon** (`cmd/mtga-companion/`):
- Runs as background service (24/7 operation)
- Monitors MTGA `Player.log` file continuously
- Parses log entries and stores in database
- Broadcasts events to connected clients via WebSocket
- Auto-starts on system boot (macOS: launchd, Windows: Service, Linux: systemd)
- Lightweight (~10-20 MB RAM)
- **HEADLESS**: Outputs structured logs only, NO terminal display, charts, or formatted output
- **Core commands**: `daemon`, `service`, `migrate`, `backup`, `replay` (testing)

**2. GUI Application** (`main.go`, `app.go`):
- Wails desktop app (Go backend + React frontend)
- Connects to daemon via WebSocket (port 9999)
- Receives real-time events for data updates
- Falls back to standalone mode if daemon unavailable
- Memory: ~50-100 MB (includes WebView)
- **ALL presentation logic** - Charts, tables, statistics, draft recommendations
- **UI-only**: No log polling or database writes (daemon handles that)

**3. Shared Components**:
- **Log Processor** (`internal/mtga/logprocessor/`): Shared log parsing logic
- **Storage Layer** (`internal/storage/`): Database access (SQLite)
- **IPC Layer** (`internal/ipc/`): WebSocket client (GUI) and server (daemon)

**Operating Modes**:

**Daemon Mode (Recommended)**:
```
MTGA → Player.log → Daemon (poller + parser) → Database
                         ↓
                   WebSocket Server (:9999)
                         ↓
                    GUI (IPC Client) → User
```
- Data collection runs 24/7, even when GUI closed
- GUI receives real-time updates via WebSocket events
- Best for users who want complete match tracking

**Standalone Mode (Fallback)**:
```
MTGA → Player.log → GUI (embedded poller + parser) → Database → User
```
- GUI has embedded log poller (same as daemon)
- No WebSocket communication
- Only collects data when GUI is running
- Good for development and casual use

**Benefits**:
- ✅ Zero data loss (daemon runs continuously)
- ✅ Better resource usage (lightweight daemon vs full GUI)
- ✅ Crash-resistant (service manager auto-restarts daemon)
- ✅ Multiple clients can connect to same daemon
- ✅ Backward compatible (standalone mode still works)

**Service Management**:
```bash
# Install daemon as system service
./mtga-companion service install

# Start/stop/status
./mtga-companion service start
./mtga-companion service stop
./mtga-companion service status

# Uninstall service
./mtga-companion service uninstall
```

**WebSocket Events**:
- `stats:updated` - Statistics recalculated
- `match:new` - New match recorded
- `draft:started` - Draft session began
- `draft:pick` - Card picked in draft
- See [docs/DAEMON_API.md](docs/DAEMON_API.md) for complete event reference

**Daemon Code Constraints**:

✅ **Daemon SHOULD use**:
- `internal/daemon` - Daemon server logic
- `internal/mtga/logprocessor` - Log parsing
- `internal/mtga/logreader` - Log file reading
- `internal/storage` - Database access
- `internal/ipc` - WebSocket server (NOT client)
- Standard library logging (`log`, `slog`)

❌ **Daemon MUST NOT use**:
- `internal/charts` - Terminal chart rendering (presentation layer)
- `internal/display` - Terminal formatting (presentation layer)
- `internal/config` - Old CLI configuration (unused)
- Any terminal display/formatting libraries
- Any code that renders charts, tables, or formatted output

**Why this matters**:
- Daemon runs as system service with no terminal
- All presentation belongs in the Wails GUI (React components)
- Mixing concerns makes the daemon bloated and defeats the architecture
- Display code in daemon = architectural violation

**For complete architecture details, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md)**

### Log Reader Architecture

The log reader is organized to parse different sections of MTGA data:

**Core Components** (`internal/mtga/logreader/`)
- `path.go` - Platform-aware log file location detection
- `reader.go` - Base JSON log entry parser

**Data Section Parsers**
Each data section has its own parser module:
- **Profile** - Player profile information
- **Arena Stats** - Game statistics and performance metrics
- **Win Rate** - Win/loss tracking and calculations
- **Draft History** - Draft picks and recommendations
- **Vault Progress** - Vault opening progress tracking

**Parser Design Pattern**
- Each parser should implement a consistent interface
- Parsers extract specific JSON event types from the log
- Follow single responsibility principle - one parser per data section
- All parsers must have comprehensive test coverage
- Use composition to build complex data from log entries

**MTGA Log Event Research**

For detailed analysis of MTGA log events, structures, and correlations discovered through manual log parsing, see:

📖 **[docs/MTGA_LOG_RESEARCH.md](docs/MTGA_LOG_RESEARCH.md)**

This document contains:
- Event structure documentation
- Timing and correlation analysis
- Edge cases and findings from manual testing
- Implementation status tracking
- Research notes organized by date

**When implementing log parsing features:** Always check the research document first to understand event timing, structure, and any edge cases that have been discovered.

### Frontend Architecture

The frontend is built with React and TypeScript, following modern best practices:

**Component Organization**:
- **Pages** (`frontend/src/pages/`) - Top-level route components
  - Each page is a complete view (MatchHistory, WinRateTrend, DeckPerformance, etc.)
  - Pages handle data fetching from backend via Wails bindings
  - Pages manage their own state and filters
- **Components** (`frontend/src/components/`) - Reusable UI components
  - Layout components (Layout, Footer, ToastContainer)
  - Shared components should be generic and reusable
  - Follow single responsibility principle

**Data Flow**:
1. **Frontend → Backend**: Call Go methods via `wailsjs/go/main/App`
   ```typescript
   import { GetMatches, GetStats } from '../../wailsjs/go/main/App';
   const matches = await GetMatches(filter);
   ```
2. **Backend → Frontend**: Event emission for real-time updates
   ```typescript
   import { EventsOn } from '../../wailsjs/runtime/runtime';
   EventsOn('stats:updated', () => { /* refresh data */ });
   ```
3. **State Management**: React hooks (useState, useEffect)
   - Local component state for UI state
   - No global state management (yet) - keep it simple

**Styling**:
- **CSS Modules** or **Component-scoped CSS** files
- Dark theme with consistent color palette:
  - Background: `#1e1e1e`
  - Secondary background: `#2d2d2d`
  - Primary accent: `#4a9eff`
  - Text: `#ffffff`
  - Muted text: `#aaaaaa`
- Use CSS Grid and Flexbox for layouts
- Avoid inline styles - prefer CSS classes

**TypeScript**:
- Use TypeScript for all frontend code
- Import models from `wailsjs/go/models`
- Avoid `any` types - use proper typing
- Use interfaces for component props

**Wails Bindings**:
- Auto-generated in `frontend/wailsjs/` - **DO NOT EDIT MANUALLY**
- Regenerate with `wails generate module`
- Go methods in `app.go` are automatically exposed to frontend
- Models from `internal/storage/models` are automatically converted to TypeScript

## Responsive Design Principles

**IMPORTANT**: All frontend UI must be responsive and adapt to different window sizes.

### Design Goals
- **Minimum window size**: 800x600 (configurable in `main.go`)
- **Optimal range**: 1024x768 to 1920x1080
- **Adapt gracefully**: UI should work at any size within reasonable bounds

### Implementation Guidelines

**1. Flexible Layouts**
- Use CSS Flexbox and Grid for responsive layouts
- Avoid fixed pixel widths - prefer percentages, `fr` units, or `min/max` constraints
- Use `flex-wrap` to allow content to reflow on smaller screens
- Example:
  ```css
  .filter-row {
    display: flex;
    gap: 16px;
    flex-wrap: wrap; /* Wraps on small screens */
  }
  ```

**2. Responsive Tables**
- Tables should scroll horizontally if needed
- Wrap table in a container with `overflow-x: auto`
- Consider hiding less important columns on smaller screens
- Example:
  ```css
  .table-container {
    overflow-x: auto;
    max-width: 100%;
  }
  ```

**3. Responsive Charts**
- Use `ResponsiveContainer` from Recharts
- Charts should scale with parent container
- Example:
  ```tsx
  <ResponsiveContainer width="100%" height={400}>
    <LineChart data={data}>
      {/* ... */}
    </LineChart>
  </ResponsiveContainer>
  ```

**4. Spacing and Typography**
- Use relative units (rem, em) for font sizes
- Maintain consistent spacing with CSS variables or Tailwind-style spacing scale
- Ensure minimum touch target size of 44x44px for interactive elements

**5. Container Management**
- Page containers should have `max-width` to prevent over-stretching on large screens
- Use `padding` instead of `margin` for internal spacing
- Example:
  ```css
  .page-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 16px;
  }
  ```

**6. Navigation and Footer**
- Navigation tabs should be horizontally scrollable if needed
- Footer should stick to bottom and adapt content based on available space
- Consider collapsing less important footer stats on small screens

**7. Forms and Filters**
- Filter rows should wrap on small screens
- Input fields should have `min-width` to remain usable
- Labels should be above inputs on mobile-style layouts

### Testing Responsive Design
- Test at minimum size (800x600)
- Test at common sizes (1024x768, 1280x720, 1920x1080)
- Resize window to ensure no layout breaking
- Check for horizontal scroll (usually indicates layout issue)

### Material Design Alignment
While we follow our own dark theme, we should adopt Material Design principles:
- **Elevation**: Use shadows and layering to create depth
- **Clear hierarchy**: Primary, secondary, and tertiary actions
- **Consistent spacing**: 4px/8px grid system (multiples of 4 or 8)
- **Transitions**: Smooth animations (200-300ms) for state changes
- **Feedback**: Visual feedback for all interactive elements (hover, active, focus states)

## Wails Development

### Building and Running

**Development mode** (with hot reload):
```bash
wails dev
```

**Production build**:
```bash
wails build
```

**Generate bindings** (after changing Go methods):
```bash
wails generate module
```

### Frontend Development

**Install dependencies**:
```bash
cd frontend
npm install
```

**Run frontend dev server** (standalone):
```bash
cd frontend
npm run dev
```

**Build frontend**:
```bash
cd frontend
npm run build
```

### Wails Project Configuration

Key files:
- `wails.json` - Wails project configuration
- `main.go` - Application entry point with window configuration
- `app.go` - Backend methods exposed to frontend

**Adding new backend methods**:
1. Add method to `App` struct in `app.go`
2. Method must be exported (capital first letter)
3. Run `wails generate module` to update bindings
4. Import from `wailsjs/go/main/App` in frontend

**Real-time events**:
```go
// Backend (Go)
runtime.EventsEmit(a.ctx, "stats:updated", data)

// Frontend (TypeScript)
EventsOn('stats:updated', (data) => { /* handle update */ });
```

## Design Patterns (v1.2 Refactoring)

**IMPORTANT**: The codebase follows several design patterns implemented during the v1.2 refactoring (2024-11). When adding new features or modifying existing code, leverage these patterns to maintain consistency and code quality.

### 1. Facade Pattern

**Location**: `internal/gui/*_facade.go`

**Purpose**: Simplify complex subsystem interactions and provide a clean interface between the Wails frontend and backend services.

**When to use**:
- Creating new major feature areas (e.g., if adding a tournament feature, create `TournamentFacade`)
- Exposing backend services to the frontend
- Grouping related operations into a cohesive interface

**Existing Facades**:
- `MatchFacade` - Match history and statistics
- `DraftFacade` - Draft sessions and insights
- `CardFacade` - Card data and metadata
- `ExportFacade` - Import/export operations
- `SystemFacade` - System initialization and daemon communication

### 2. Strategy Pattern

**Location**: `internal/mtga/draft/insights/*_strategy.go`

**Purpose**: Define a family of algorithms and make them interchangeable based on context (e.g., different analysis for Premier Draft vs Quick Draft).

**When to use**:
- Algorithms vary based on format/type/mode
- Need different behavior for similar operations
- Want to avoid complex if/else chains for type checking

**Existing Strategies**:
- `PremierDraftStrategy` - Analysis for human opponent drafts (10 bombs, 8 removal, 15 creatures, 20 commons)
- `QuickDraftStrategy` - Analysis for bot opponent drafts (12 bombs, 10 removal, 18 creatures, 25 commons)

**Example Usage**:
```go
// Use strategy factory to get the right analyzer
analyzer := insights.NewAnalyzerForFormat(storage, draftFormat)
results, err := analyzer.AnalyzeFormat(ctx, setCode, draftFormat)
```

### 3. Builder Pattern

**Location**: `internal/export/builder.go`

**Purpose**: Construct complex objects step-by-step with a fluent API, making configuration clear and readable.

**When to use**:
- Object has many configuration options
- Want to make construction intent clear
- Need to validate configuration before creating object

**Example Usage**:
```go
// Fluent API for export configuration
err := export.NewExportBuilder().
    WithFormat(export.FormatJSON).
    WithFilePath(filePath).
    WithPrettyJSON(true).
    WithOverwrite(true).
    Export(data)
```

### 4. Observer Pattern

**Location**: `internal/events/`

**Purpose**: Define one-to-many dependency between objects so when one object changes state, all dependents are notified automatically.

**When to use**:
- Multiple components need to react to the same event
- Want to decouple event producers from consumers
- Broadcasting events to multiple destinations (frontend, IPC, logging, analytics)

**Existing Observers**:
- `WailsObserver` - Forwards events to Wails frontend
- `IPCObserver` - Forwards events to IPC daemon
- `LoggingObserver` - Logs events for debugging

**Example Usage**:
```go
// Get the event dispatcher from SystemFacade
dispatcher := systemFacade.GetEventDispatcher()

// Dispatch events
dispatcher.Dispatch(events.Event{
    Type: "match:completed",
    Data: map[string]interface{}{
        "matchId": matchID,
        "result":  "victory",
    },
    Context: ctx,
})
```

**IMPORTANT**: Never use `wailsruntime.EventsEmit` directly. Always use the EventDispatcher.

### 5. Command Pattern

**Location**: `internal/commands/`

**Purpose**: Encapsulate operations as objects, enabling parameterization, queuing, logging, and undo support.

**When to use**:
- Operations need to be queued or scheduled
- Want to support undo/redo
- Need operation history for auditing
- Want retry logic for critical operations

**Existing Commands**:
- `ReplayCommand`, `PauseReplayCommand`, `ResumeReplayCommand`, `StopReplayCommand` - Log replay operations
- `StartupRecoveryCommand` - Initialize daemon with retry logic
- `ShutdownCommand` - Graceful shutdown

**Example Usage**:
```go
executor := commands.NewCommandExecutor(10)
cmd := commands.NewReplayCommand(ipcClient, filePaths, speed, filterType, pauseOnDraft, clearData)
err := executor.Execute(ctx, cmd)
```

### Pattern Implementation Guidelines

**For Event Emission** - Use EventDispatcher, not EventsEmit:
```go
// ❌ Don't do this
wailsruntime.EventsEmit(ctx, "stats:updated", data)

// ✅ Do this
facade.eventDispatcher.Dispatch(events.Event{
    Type:    "stats:updated",
    Data:    data,
    Context: ctx,
})
```

**For Export Operations** - Use Builder pattern:
```go
// ❌ Don't do this
exporter := export.NewExporter(export.Options{...})
err := exporter.Export(data)

// ✅ Do this
err := export.NewExportBuilder().
    WithFormat(export.FormatJSON).
    WithFilePath(filePath).
    Export(data)
```

**For Format-Specific Analysis** - Use Strategy pattern:
```go
// ❌ Don't do this
if draftFormat == "PremierDraft" {
    // Premier logic
} else if draftFormat == "QuickDraft" {
    // Quick logic
}

// ✅ Do this
analyzer := insights.NewAnalyzerForFormat(storage, draftFormat)
results, err := analyzer.AnalyzeFormat(ctx, setCode, draftFormat)
```

**For Daemon Operations** - Use Command pattern:
```go
// ❌ Don't do this
ipcClient.Send(map[string]interface{}{"type": "replay_logs", ...})

// ✅ Do this
cmd := commands.NewReplayCommand(ipcClient, filePaths, speed, filterType, pauseOnDraft, clearData)
err := executor.Execute(ctx, cmd)
```

### Adding New Features Checklist

When implementing new features, ask:

1. **Does it need a facade?** - Is this a major feature area that needs its own interface?
2. **Does behavior vary by type?** - Consider Strategy pattern
3. **Complex configuration?** - Consider Builder pattern
4. **Need event notifications?** - Use EventDispatcher (Observer pattern)
5. **Encapsulated operation?** - Consider Command pattern

For complete architectural details and rationale, see **ADR-011** in `docs/ARCHITECTURE_DECISIONS.md`.

---

## Coding Principles

### KISS (Keep It Simple, Stupid)
- Favor simple, straightforward solutions over clever or complex ones
- Avoid premature optimization and unnecessary abstractions
- Write code that is easy to understand and maintain
- Prefer clarity over brevity when they conflict

### Effective Go Standards
Follow the guidelines from https://go.dev/doc/effective_go:

**Naming**
- Use `MixedCaps` for exported identifiers, not underscores
- Keep package names lowercase, concise, single-word
- Use `-er` suffix for single-method interfaces (e.g., `Reader`, `Writer`)

**Formatting**
- Always run `gofmt` - trust the automated formatter
- Use tabs for indentation (gofmt default)
- Opening brace must be on the same line as control statements

**Error Handling**
- Return errors as additional return values, don't panic
- Check errors explicitly and handle them appropriately
- Use early returns to avoid deep nesting

**Concurrency**
- "Do not communicate by sharing memory; instead, share memory by communicating"
- Use channels to coordinate goroutines
- Prefer channels over mutex-protected shared variables

**Data Structures**
- Design types so their zero values are useful without initialization
- Use `make()` for slices, maps, and channels
- Embed types within structs for composition

**Interfaces**
- Keep interfaces small and focused
- Define interfaces where they're used, not where they're implemented
- Accept interfaces, return concrete types when appropriate

**Documentation**
- Write doc comments immediately preceding declarations
- Start comments with the name of the thing being described
- Keep comments clear and concise
- Run ./scripts/dev.sh before creating the PR
- Can you also try to make sure that we adhere to material design standards.  We should aim for simplicity but control design that is aesteticly pleasing.
- When creating issues always cleanly organize them by version and milestone.  If an appropriate project doesn't exist create it.
- We are using wails desktop for the UI not web