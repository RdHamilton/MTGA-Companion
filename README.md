# MTGA-Companion

A modern companion application for Magic: The Gathering Arena (MTGA). Track your matches, analyze your performance, and enhance your MTGA experience with real-time statistics, ML-powered recommendations, and metagame insights.

## Features

### Modern Web UI (v1.4)
- **Browser-Based Interface**: React SPA with REST API backend - opens in your default browser
- **Real-Time Updates**: Live statistics via WebSocket while you play MTGA
- **Dark Theme**: Easy on the eyes during long gaming sessions
- **Responsive Design**: Works on any screen size

### Match Tracking & Analytics
- **Match History**: View all your matches with filtering and sorting
- **Win Rate Trends**: Visualize performance over time
- **Deck Performance**: Track win rates by deck
- **Format Distribution**: See your play patterns across formats
- **Result Breakdown**: Detailed statistics by format and time period

### Data Management
- **Log Reading**: Automatically locates and reads MTGA Player.log files
- **Auto-Detection**: Cross-platform support for macOS and Windows log locations
- **Real-Time Monitoring**: Poll-based log watching for instant updates
- **Database Storage**: Local SQLite database with migration support
- **Export System**: Export statistics in CSV or JSON formats

### Draft Assistant (v1.1)
- **Real-Time Draft Assistant**: Live card recommendations and analysis during drafts
- **Type Synergy Detection**: Automatic detection of card type synergies with visual indicators
- **Card Suggestions**: Context-aware recommendations based on your picked cards
- **Draft Deck Win Rate Predictor**: AI-powered prediction with letter grades (A/B/C/D/F)
- **Format Meta Insights**: Archetype performance data, best color pairs, overdrafted colors
- **Archetype Performance Dashboard**: Interactive archetype selection with top cards per archetype
- **Draft Statistics Dashboard**: Real-time mana curve and color distribution
- **Missing Cards Detection**: Track cards you don't own from draft packs
- **Historical Draft Replay**: View and replay past draft pick sequences
- **Card Data Integration**: 17Lands public datasets and Scryfall metadata

### Deck Builder (v1.3)
- **Deck Creation & Management**: Build constructed and draft-based decks with full CRUD operations
- **AI-Powered Recommendations**: Intelligent card suggestions based on color fit, mana curve, synergy, and card quality
- **Multiple Import Formats**: Import from Arena, plain text, and other common formats
- **Multiple Export Formats**: Export to Arena, MTGO, MTGGoldfish, and plain text formats
- **Comprehensive Statistics**: Mana curve, color distribution, type breakdown, and land recommendations
- **Draft Pool Validation**: Draft decks restricted to cards from the associated draft
- **Format Legality Checking**: Validate decks for Standard, Historic, Explorer, Alchemy, Brawl, and Commander
- **Performance Tracking**: Automatic win rate tracking and performance metrics
- **Tagging & Organization**: Categorize and filter decks with custom tags
- **Deck Library**: Advanced filtering by format, source, tags, and performance

### Quest & Statistics Tracking (v1.3)
- **Quest Tracking**: Monitor daily and weekly quest progress
- **Accurate Gold Calculation**: Parses actual quest rewards instead of estimates
- **Real-Time Updates**: Live draft updates via `draft:updated` events
- **Set Symbol Display**: Card displays now show set symbols/icons

### ML-Powered Recommendations (v1.4)
- **Machine Learning Engine**: Intelligent card recommendations using trained ML models
- **Personal Play Style Learning**: Adapts recommendations based on your deck building history and preferences
- **Meta-Aware Suggestions**: Incorporates tournament data and metagame trends into recommendations
- **Ollama Integration**: Optional local LLM support for natural language explanations of recommendations
- **Feedback Collection**: Records your card acceptance/rejection to improve future recommendations
- **Hybrid Scoring**: Combines ML predictions with rule-based analysis for best results

### Metagame Dashboard (v1.4)
- **Live Meta Data**: Real-time metagame data from MTGGoldfish and MTGTop8
- **Archetype Tier Lists**: View Tier 1-4 archetypes with meta share and tournament performance
- **Archetype Detail View**: Click any archetype for detailed stats, trend analysis, and tier explanations
- **Format Support**: Standard, Historic, Explorer, Pioneer, and Modern formats
- **Tournament Tracking**: Recent tournament results with top decks and winner information

### Draft Enhancements (v1.4)
- **Enhanced Synergy Scoring**: Improved draft prediction with better synergy detection
- **Color Pair Archetypes**: Automatic archetype detection based on top color pairs in your draft
- **Card Type Categorization**: Set guide uses card types for better organization
- **Keyword Extraction**: Sophisticated keyword analysis for card recommendations

### Enhanced Deck Builder (v1.4.1)
- **Undo/Redo Support**: Full undo/redo functionality with Ctrl+Z/Ctrl+Y keyboard shortcuts
- **Build Around Mode**: Generate complete decks around key cards with archetype selection (Aggro, Midrange, Control)
- **Iterative Building**: Add cards one at a time with live suggestions that update based on your choices
- **Quick Generate**: Instantly generate a complete 60-card deck with optimal land distribution
- **Budget Mode**: Filter suggestions to cards you already own
- **Score Breakdown**: See why cards are recommended (color fit, curve fit, synergy, card quality)

### In-App Documentation (v1.4.1)
- **Contextual Help Icons**: Click "?" icons throughout the app for detailed feature explanations
- **Enhanced Tooltips**: Hover over stats, badges, and buttons for quick help
- **ML Suggestions Help**: Learn how ML recommendations work and how to use confidence scores
- **Archetype Explanations**: Understand Aggro, Midrange, and Control playstyles
- **Play Pattern Insights**: Documentation for improvement suggestions based on your matches

### Settings Improvements (v1.3)
- **Collapsible Accordion Navigation**: Organized settings into collapsible sections
- **URL Hash Navigation**: Direct links to settings sections (e.g., `#connection`, `#17lands`)
- **Keyboard Navigation**: Full keyboard support for accessibility
- **LoadingButton Component**: Consistent loading states across all async operations

## Screenshots

### Match History
![Match History](docs/images/match-history.png)

### Draft History
![Draft History](docs/images/draft-history.png)

### Deck Management
![Decks](docs/images/decks.png)

### Collection Browser
![Collection](docs/images/collection.png)

### Meta Dashboard
![Meta Dashboard](docs/images/meta-dashboard.png)

### Charts & Analytics
![Deck Performance](docs/images/charts-deck-performance.png)

![Format Distribution](docs/images/charts-format-distribution.png)

> **Note**: Screenshots are generated automatically using Playwright. Run `npm run screenshots` from the project root to regenerate them with your local data.

## Documentation

📚 **[Complete Documentation Wiki →](https://github.com/RdHamilton/MTGA-Companion/wiki)**

- **[Installation Guide](https://github.com/RdHamilton/MTGA-Companion/wiki/Installation)** - Setup instructions for macOS and Windows
- **[Usage Guide](https://github.com/RdHamilton/MTGA-Companion/wiki/Usage-Guide)** - How to use all features
- **[CLI Commands](https://github.com/RdHamilton/MTGA-Companion/wiki/CLI-Commands)** - Complete command reference
- **[Configuration](https://github.com/RdHamilton/MTGA-Companion/wiki/Configuration)** - Configuration options
- **[Troubleshooting](https://github.com/RdHamilton/MTGA-Companion/wiki/Troubleshooting)** - Common issues and solutions

### Technical Documentation

- **[Architecture](docs/ARCHITECTURE.md)** - Service-based system design and architecture
- **[Deck Builder Guide](docs/DECK_BUILDER.md)** - Comprehensive deck builder documentation with API reference
- **[Daemon API](docs/DAEMON_API.md)** - WebSocket API reference for daemon integration
- **[Development Guide](docs/DEVELOPMENT.md)** - Development setup and contributing guidelines
- **[Migration Guide](docs/MIGRATION_TO_SERVICE_ARCHITECTURE.md)** - Upgrading to service-based architecture
- **[Daemon Installation](docs/DAEMON_INSTALLATION.md)** - Complete daemon service installation guide
- **[Database Schema](https://github.com/RdHamilton/MTGA-Companion/wiki/Database-Schema)** - Database structure

## Prerequisites

- **MTG Arena** must be installed and configured to enable detailed logging
- **Go 1.25+** (for building from source)
- **Ollama** (optional) - For AI-powered natural language explanations

## Ollama Setup (Optional)

MTGA Companion can use [Ollama](https://ollama.ai/) to provide natural language explanations for card recommendations. This feature is completely optional - the app works fully without it.

### Installing Ollama

**macOS**:
```bash
brew install ollama
# Or download from https://ollama.ai/download
```

**Windows**:
Download the installer from https://ollama.ai/download

**Linux**:
```bash
curl -fsSL https://ollama.ai/install.sh | sh
```

### Starting Ollama

```bash
# Start Ollama server (runs on port 11434 by default)
ollama serve
```

### Configuring in MTGA Companion

1. Open MTGA Companion
2. Go to **Settings** → **ML/AI Settings**
3. Enable **Ollama Integration**
4. Configure:
   - **Ollama URL**: `http://localhost:11434` (default)
   - **Model**: `qwen3:8b` (recommended) or any compatible model
5. Click **Test Connection** to verify

The app will automatically pull the model if it's not already downloaded.

### Supported Models

Any Ollama model works, but these are recommended:
- `qwen3:8b` - Default, good balance of quality and speed
- `llama3.2:3b` - Faster, smaller, good for older hardware
- `mistral:7b` - Alternative with different response style

**Note**: Without Ollama, MTGA Companion uses template-based explanations which work well for most use cases.

## Enabling Detailed Logging in MTG Arena

**IMPORTANT**: You must enable detailed logging in MTG Arena for this companion app to work properly.

### Steps to Enable Detailed Logging:

1. Launch **Magic: The Gathering Arena**
2. Click the **Adjust Options** gear icon ⚙️ at the top right of the home screen
3. In the Options menu, click **View Account**
4. Find and check the **Detailed Logs** checkbox (may also be labeled "Enable Detailed Logs" or "Plugin Support")
5. **Restart** MTG Arena for the changes to take effect

### Why Enable Detailed Logging?

Detailed logging allows MTG Arena to output game events and data in JSON format to the Player.log file. This enables companion applications like MTGA-Companion to:
- Track your game statistics
- Analyze your collection
- Display deck information
- Monitor game state in real-time

**Note**: Detailed logging has no impact on game performance and is specifically designed to support third-party companion tools.

## Installation

### Quick Start (Recommended)

Download the latest release from the [Releases page](https://github.com/RdHamilton/MTGA-Companion/releases):

#### macOS (Currently Supported)

1. Download `MTGA-Companion-vX.X.X-macOS.dmg`
2. Open the DMG and drag `MTGA Companion.app` to your Applications folder
3. **First launch**: Right-click the app → "Open" (to bypass Gatekeeper)
4. The app will start the API server and open your default browser

**What happens on launch:**
- The app starts a local REST API server (port 8080)
- Your default browser opens to the MTGA Companion UI
- The app monitors MTGA logs in the background via the daemon service

#### Windows / Linux

Windows and Linux builds are planned for future releases. Currently, you can build from source (see below).

### Daemon Mode (Recommended)

**What is Daemon Mode?**

MTGA Companion can run as a background service (daemon) that continuously monitors your MTGA log file and provides data to the GUI. This is the **recommended setup** because:

✅ **Always Running** - Data collection continues even when GUI is closed
✅ **Auto-Start** - Daemon starts automatically on system boot
✅ **Reliable** - Automatic restart if it crashes
✅ **Cleaner** - Separation of data collection (daemon) and display (GUI)

**Platform Support Status:**

- ✅ **macOS**: Service installation fully tested and verified
- ⚠️ **Windows**: Service code implemented but not yet verified on Windows
- ⚠️ **Linux**: Service code implemented but not yet verified on Linux

> **Note**: The service installation code uses the cross-platform [kardianos/service](https://github.com/kardianos/service) library which supports Windows, macOS, and Linux. While the implementation is complete for all platforms, testing has only been performed on macOS. Windows and Linux service installation should work but has not been verified yet.

**Installation**:

1. Download and extract MTGA Companion for your platform (see Quick Start above)

2. Install the daemon service:

   **macOS/Linux**:
   ```bash
   cd /path/to/MTGA-Companion
   ./mtga-companion service install
   ./mtga-companion service start
   ```

   **Windows (as Administrator)**:
   ```powershell
   cd C:\Path\To\MTGA-Companion
   .\mtga-companion.exe service install
   .\mtga-companion.exe service start
   ```

3. Verify daemon is running:
   ```bash
   ./mtga-companion service status
   ```

   Expected output:
   ```
   Service Status:
     Status: ✓ Running
   ```

4. Launch the GUI normally - it will automatically connect to the daemon

**Service Management**:

```bash
# Check status
./mtga-companion service status

# Start/Stop
./mtga-companion service start
./mtga-companion service stop

# Restart
./mtga-companion service restart

# Uninstall
./mtga-companion service uninstall
```

📚 **For detailed daemon installation and troubleshooting, see [docs/DAEMON_INSTALLATION.md](docs/DAEMON_INSTALLATION.md)**

**Alternative: Standalone Mode**

If you prefer not to use daemon mode, the GUI includes an embedded log poller that works standalone. Simply launch the app and it will monitor logs automatically.

### Build From Source

**Prerequisites**:
- [Go 1.25+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/) (for frontend)

**Clone and Build**:
```bash
# Clone repository
git clone https://github.com/RdHamilton/MTGA-Companion.git
cd MTGA-Companion

# Build the Go backend (API server + daemon)
go build -o bin/mtga-companion ./cmd/mtga-companion
go build -o bin/apiserver ./cmd/apiserver

# Install and build frontend
cd frontend
npm install
npm run build
cd ..
```

**Development Mode** (with hot reload):
```bash
# Terminal 1: Start API server
go run ./cmd/apiserver

# Terminal 2: Start frontend dev server
cd frontend
npm run dev
```

The frontend dev server runs at `http://localhost:3000` and proxies API requests to the backend at `http://localhost:8080`.

## Player.log File Locations

The application automatically detects the Player.log location based on your platform:

### macOS
```
~/Library/Application Support/com.wizards.mtga/Logs/Logs/Player.log
```

**Tip**: If you can't see your Library folder, press `Command + Shift + .` (dot) to show hidden files in Finder.

### Windows
```
C:\Users\{username}\AppData\LocalLow\Wizards Of The Coast\MTGA\Player.log
```

**Tip**: You can paste this path directly into Windows Explorer's address bar (replace `{username}` with your Windows username).

### Previous Session Logs

MTGA also saves the previous session's log as `Player-prev.log` in the same directory, which can be useful for reviewing past games.

### Log File Rotation

MTGA may rotate log files during long gaming sessions when the log becomes large. MTGA-Companion automatically handles log rotation:

- **Detection**: Monitors for file size decreases, file removal/rename events (via fsnotify)
- **Recovery**: Automatically reopens the new log file and continues monitoring
- **State Preservation**: Maintains draft state and game tracking across rotation events
- **Logging**: Rotation events are logged with `[INFO]` messages for visibility

**Rotation scenarios handled:**
- Size-based rotation (when Player.log exceeds MTGA's size limit)
- File removal and recreation
- Manual log deletion/archival

The overlay and tracking features continue working seamlessly during and after log rotation.

## Usage

### GUI Application

Launch the MTGA Companion desktop app:

**Windows**: Double-click `MTGA-Companion.exe`
**macOS**: Double-click `MTGA-Companion.app` from Applications
**Linux**: Run `./MTGA-Companion-linux-amd64`

The application will:
1. Automatically locate your MTGA Player.log file
2. Initialize the database (first run creates `~/.mtga-companion/data.db`)
3. Start monitoring the log file for new matches
4. Display your statistics and match history in real-time

### Navigation

- **Match History**: View and filter all your matches
- **Draft**: Real-time draft assistant with recommendations, synergy detection, and format insights
- **Decks**: Deck library with builder, import/export, and AI recommendations (v1.3)
- **Collection**: Browse and track your card collection (v1.3.1)
- **Meta**: Metagame dashboard with archetype tier lists and tournament data (v1.4)
- **Charts**: Visualize your performance data
  - Win Rate Trend: Performance over time
  - Deck Performance: Win rates by deck
  - Rank Progression: Track your ladder climbing
  - Format Distribution: Play patterns across formats
  - Result Breakdown: Detailed statistics
- **Settings**: Configure database path, ML settings, and Ollama integration

### Real-Time Updates

While MTGA is running and you're playing games:
- New matches are automatically detected and added
- Statistics update in real-time
- Footer shows at-a-glance stats (total matches, win rate, streak)
- Toast notifications confirm when data is updated

### CLI Mode (Advanced)

The CLI is still available for automation and advanced users:

```bash
# Read log and display basic info
./mtga-companion read

# Export statistics
./mtga-companion export stats -json

# Run draft overlay
./mtga-companion -draft-overlay-mode
```

See the [CLI Commands Wiki](https://github.com/RdHamilton/MTGA-Companion/wiki/CLI-Commands) for complete reference.

## Development

### Project Structure

```
MTGA-Companion/
├── cmd/                     # Application entry points
│   ├── apiserver/          # REST API server (v1.4+)
│   └── mtga-companion/     # CLI daemon for log monitoring
├── frontend/                # React + TypeScript SPA
│   ├── src/
│   │   ├── components/     # Reusable UI components
│   │   ├── pages/          # Page components (routes)
│   │   ├── services/api/   # REST API client modules
│   │   ├── App.tsx         # Root component
│   │   └── main.tsx        # Frontend entry point
│   ├── package.json
│   └── vite.config.ts
├── internal/                # Private application code
│   ├── api/                # REST API handlers & router (v1.4+)
│   │   ├── handlers/      # HTTP request handlers
│   │   ├── websocket/     # WebSocket for real-time updates
│   │   └── router.go      # API route definitions
│   ├── gui/                # Facade layer (business logic)
│   ├── ml/                 # Machine learning engine (v1.4+)
│   ├── llm/                # Ollama LLM client (v1.4+)
│   ├── meta/               # Metagame data service (v1.4+)
│   ├── mtga/               # MTGA-specific logic
│   │   ├── logreader/     # Log parsing
│   │   ├── draft/         # Draft overlay
│   │   └── recommendations/ # Card recommendations
│   └── storage/            # Database and persistence
│       ├── models/        # Data models
│       └── repository/    # Data access layer
├── docs/                    # Documentation
├── scripts/                 # Development scripts
└── CLAUDE.md               # AI assistant guidance
```

### Development Workflow

**Full Stack Development** (recommended):
```bash
# Terminal 1: Start API server with hot reload
go run ./cmd/apiserver

# Terminal 2: Start frontend dev server
cd frontend
npm run dev

# Open browser to http://localhost:3000
```

**Go Development** (backend):
```bash
# Format code
gofumpt -w .

# Run linter
golangci-lint run --timeout=5m

# Run tests
go test ./...

# Run tests with race detection
go test -race ./...

# Build binaries
go build -o bin/apiserver ./cmd/apiserver
go build -o bin/mtga-companion ./cmd/mtga-companion
```

**Frontend Development**:
```bash
# Install dependencies
cd frontend
npm install

# Run frontend dev server
npm run dev

# Build frontend for production
npm run build

# Type checking
npm run tsc

# Linting
npm run lint

# Run tests
npm run test:run
```

### Running Tests

**Go Tests**:
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests with race detection
go test -race ./...
```

**Frontend Tests** (when added):
```bash
cd frontend
npm test
```

## Troubleshooting

### "Player.log not found!"

If you see this error:
1. Verify MTG Arena is installed
2. Ensure you've enabled detailed logging (see instructions above)
3. Run MTG Arena at least once after enabling detailed logging
4. Check that the log file exists at the expected location for your platform

### macOS: Cannot Find Library Folder

Press `Command + Shift + .` in Finder to show hidden files and folders.

### Windows: Access Denied

Ensure you have read permissions for the MTGA log directory. Try running as administrator if needed.

## Technology Stack

MTGA-Companion is built with modern technologies for performance and cross-platform compatibility:

### Architecture (v1.4+)

- **REST API + Browser SPA** - Decoupled architecture for flexibility
  - Go REST API server with WebSocket support
  - React SPA served via Vite or static files
  - Opens in your default browser - no native app required

### Go 1.25 Features

MTGA Companion leverages Go 1.25's new features for improved performance and debugging:

**Flight Recorder** (`internal/daemon/flight_recorder.go`)
- Uses `runtime/trace.FlightRecorder` for low-overhead execution tracing
- Automatically captures traces when errors exceed threshold
- Configurable trace buffer size and retention
- Manual trace capture via daemon API

**Benchmark Suite** (`benchmarks/`)
- **GC Benchmarks**: Compare default GC vs experimental `greenteagc` garbage collector
  ```bash
  # Run with default GC
  go test -bench=. -benchmem ./benchmarks/...

  # Run with greenteagc (Go 1.25+ required)
  GOEXPERIMENT=greenteagc go test -bench=. -benchmem ./benchmarks/...

  # Compare results
  ./benchmarks/run_gc_comparison.sh
  ```
- **JSON Benchmarks**: Compare `encoding/json` (v1) vs experimental `encoding/json/v2`
  ```bash
  # Run comparison (requires GOEXPERIMENT=jsonv2)
  ./benchmarks/run_json_comparison.sh
  ```

### Backend (Go)

- **[Go 1.25+](https://go.dev/)** - Programming language
- **[Chi Router](https://github.com/go-chi/chi)** - Lightweight HTTP router
- **[SQLite 3](https://www.sqlite.org/)** - Local database storage
- **[modernc.org/sqlite](https://gitlab.com/cznic/sqlite)** - Pure Go SQLite driver (no CGo required)
- **[golang-migrate/migrate](https://github.com/golang-migrate/migrate)** - Database migration management
- **[gorilla/websocket](https://github.com/gorilla/websocket)** - WebSocket implementation
- **[fsnotify](https://github.com/fsnotify/fsnotify)** - Cross-platform file system notifications
- **[kardianos/service](https://github.com/kardianos/service)** - Cross-platform service management

### Frontend (React + TypeScript)

- **[React 18](https://react.dev/)** - UI library with hooks
- **[TypeScript](https://www.typescriptlang.org/)** - Type-safe JavaScript
- **[React Router](https://reactrouter.com/)** - Client-side routing
- **[Recharts](https://recharts.org/)** - Data visualization and charting library
- **[Vite](https://vite.dev/)** - Fast build tool and dev server
- **[Vitest](https://vitest.dev/)** - Unit testing framework
- **[Playwright](https://playwright.dev/)** - E2E testing framework

### ML/AI Features (v1.4+)

- **Machine Learning Engine** - Custom ML model for card recommendations
- **[Ollama](https://ollama.ai/)** - Local LLM integration for natural language explanations

### Data Sources

- **[17Lands](https://www.17lands.com/)** - Draft statistics and card ratings
- **[Scryfall](https://scryfall.com/)** - Card metadata and images
- **[MTGGoldfish](https://www.mtggoldfish.com/)** - Metagame data
- **[MTGTop8](https://www.mtgtop8.com/)** - Tournament results

For a complete list of dependencies, see [`go.mod`](go.mod) and [`frontend/package.json`](frontend/package.json).

## Contributing

Contributions are welcome! Please ensure:
- Code follows Go best practices (see `CLAUDE.md`)
- All tests pass (`./scripts/test.sh`)
- Code is formatted (`./scripts/dev.sh fmt`)

See the [Development Guide](https://github.com/RdHamilton/MTGA-Companion/wiki/Development) for detailed contribution guidelines.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Disclaimer

MTGA-Companion is not affiliated with, endorsed by, or sponsored by Wizards of the Coast. Magic: The Gathering Arena and its associated trademarks are property of Wizards of the Coast LLC.

## Acknowledgments

- Wizards of the Coast for MTG Arena and its detailed logging support
- The MTGA community for documentation on log formats and structure

## CLI Flag Migration (v0.2.0)

As of v0.2.0, CLI flags have been standardized for consistency. Old flags are still supported but deprecated.

### Quick Reference

| Old Flag (Deprecated) | New Flag | Shorthand |
|-----------------------|----------|-----------|
| `-gui` | `-gui-mode` | `-g` |
| `-debug` | `-debug-mode` | `-d` |
| `-cache` | `-cache-enabled` | |
| `-poll-interval` | `-log-poll-interval` | |
| `-use-file-events` | `-log-use-fsnotify` | |
| `-draft-overlay` | `-draft-overlay-mode` | |
| `-set-file` | `-overlay-set-file` | |
| `-log-path` | `-log-file-path` | |
| `-overlay-set` | `-overlay-set-code` | |
| `-overlay-lookback` | `-overlay-lookback-hours` | |

**Note:** Deprecated flags will show a warning and will be removed in v2.0.0. See `FLAG_MIGRATION.md` for complete details.

### Examples

```bash
# Old syntax (still works, shows warning)
./bin/mtga-companion -debug -gui

# New syntax (recommended)
./bin/mtga-companion -debug-mode -gui-mode

# New syntax with shortcuts
./bin/mtga-companion -d -g
```

