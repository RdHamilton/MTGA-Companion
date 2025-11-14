# MTGA-Companion

A modern desktop companion application for Magic: The Gathering Arena (MTGA). Track your matches, analyze your performance, and enhance your MTGA experience with real-time statistics and insights.

## Features

### Desktop GUI
- **Modern Interface**: Cross-platform desktop application with React UI
- **Real-Time Updates**: Live statistics while you play MTGA
- **Dark Theme**: Easy on the eyes during long gaming sessions
- **Responsive Design**: Adapts to different window sizes

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

### Draft Support
- **Draft Tracking**: Record and analyze all your draft picks
- **Card Data Integration**: 17Lands draft statistics and Scryfall metadata
- **Draft Overlay**: In-game overlay with pick recommendations (CLI mode)

## Documentation

📚 **[Complete Documentation Wiki →](https://github.com/RdHamilton/MTGA-Companion/wiki)**

- **[Installation Guide](https://github.com/RdHamilton/MTGA-Companion/wiki/Installation)** - Setup instructions for macOS and Windows
- **[Usage Guide](https://github.com/RdHamilton/MTGA-Companion/wiki/Usage-Guide)** - How to use all features
- **[CLI Commands](https://github.com/RdHamilton/MTGA-Companion/wiki/CLI-Commands)** - Complete command reference
- **[Configuration](https://github.com/RdHamilton/MTGA-Companion/wiki/Configuration)** - Configuration options
- **[Troubleshooting](https://github.com/RdHamilton/MTGA-Companion/wiki/Troubleshooting)** - Common issues and solutions

### Technical Documentation

- **[Architecture](https://github.com/RdHamilton/MTGA-Companion/wiki/Architecture)** - System design and architecture
- **[Database Schema](https://github.com/RdHamilton/MTGA-Companion/wiki/Database-Schema)** - Database structure
- **[Development](https://github.com/RdHamilton/MTGA-Companion/wiki/Development)** - Development setup and guidelines

## Prerequisites

- **MTG Arena** must be installed and configured to enable detailed logging
- **Go 1.21+** (for building from source)

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

Download the latest release for your platform from the [Releases page](https://github.com/RdHamilton/MTGA-Companion/releases):

#### Windows

1. Download `MTGA-Companion-windows-amd64.exe`
2. Run the executable - no installation required!
3. **(Optional)** Create a shortcut to your desktop or taskbar

**First Run**: Windows may show a security warning. Click "More info" → "Run anyway"

#### macOS

1. Download `MTGA-Companion.app.zip`
2. Extract and drag `MTGA-Companion.app` to your Applications folder
3. **First launch**: Right-click the app → "Open" (to bypass Gatekeeper)
4. Grant permissions if macOS requests access to files

**Subsequent launches**: Double-click the app normally

#### Linux

1. Download `MTGA-Companion-linux-amd64`
2. Make executable and run:
   ```bash
   chmod +x MTGA-Companion-linux-amd64
   ./MTGA-Companion-linux-amd64
   ```

### Build From Source

**Prerequisites**:
- [Go 1.23+](https://go.dev/dl/)
- [Node.js 20+](https://nodejs.org/) (for frontend)
- [Wails CLI](https://wails.io/docs/gettingstarted/installation)

**Install Wails**:
```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

**Clone and Build**:
```bash
# Clone repository
git clone https://github.com/RdHamilton/MTGA-Companion.git
cd MTGA-Companion

# Install frontend dependencies
cd frontend
npm install
cd ..

# Build with Wails
wails build

# Built app will be in build/bin/
```

**Development Mode** (with hot reload):
```bash
wails dev
```

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
- **Charts**: Visualize your performance data
  - Win Rate Trend: Performance over time
  - Deck Performance: Win rates by deck
  - Rank Progression: Track your ladder climbing
  - Format Distribution: Play patterns across formats
  - Result Breakdown: Detailed statistics
- **Settings**: Configure database path and other options

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
├── main.go                  # Wails entry point
├── app.go                   # Go backend API for frontend
├── frontend/                # React + TypeScript frontend
│   ├── src/
│   │   ├── components/     # Reusable UI components
│   │   ├── pages/          # Page components (routes)
│   │   ├── App.tsx         # Root component
│   │   └── main.tsx        # Frontend entry point
│   ├── wailsjs/            # Auto-generated Wails bindings
│   ├── package.json
│   └── vite.config.ts
├── internal/                # Private application code
│   ├── gui/                # GUI-specific backend code
│   ├── mtga/               # MTGA-specific logic
│   │   ├── logreader/     # Log parsing
│   │   └── draft/         # Draft overlay
│   └── storage/            # Database and persistence
│       ├── models/        # Data models
│       └── repository/    # Data access layer
├── cmd/                     # CLI application (legacy)
│   └── mtga-companion/
├── scripts/                 # Development scripts
└── CLAUDE.md               # AI assistant guidance
```

### Development Workflow

**Wails Development**:
```bash
# Run in development mode with hot reload
wails dev

# Build production version
wails build

# Generate Go ↔ TypeScript bindings (after changing app.go)
wails generate module
```

**Go Development** (backend):
```bash
# Format, lint, and build
./scripts/dev.sh

# Run specific checks
./scripts/dev.sh fmt       # Format code
./scripts/dev.sh vet       # Run go vet
./scripts/dev.sh lint      # Run golangci-lint
./scripts/dev.sh check     # Run all checks

# Run tests
./scripts/test.sh          # Run tests with race detection
./scripts/test.sh coverage # Generate coverage report
```

**Frontend Development**:
```bash
# Install dependencies
cd frontend
npm install

# Run frontend dev server (standalone)
npm run dev

# Build frontend for production
npm run build

# Type checking
npm run type-check

# Linting
npm run lint
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

### Desktop Application

- **[Wails v2](https://wails.io/)** - Go + Web frontend framework for desktop apps
  - Native webview (WebKit on macOS, WebView2 on Windows, WebKitGTK on Linux)
  - No Electron overhead - smaller binary, faster startup
  - Type-safe Go ↔ JavaScript bindings

### Backend (Go)

- **[Go 1.23+](https://go.dev/)** - Programming language
- **[SQLite 3](https://www.sqlite.org/)** - Local database storage
- **[modernc.org/sqlite](https://gitlab.com/cznic/sqlite)** - Pure Go SQLite driver (no CGo required)
- **[golang-migrate/migrate](https://github.com/golang-migrate/migrate)** - Database migration management
- **[fsnotify](https://github.com/fsnotify/fsnotify)** - Cross-platform file system notifications

### Frontend (React + TypeScript)

- **[React 18](https://react.dev/)** - UI library with hooks
- **[TypeScript](https://www.typescriptlang.org/)** - Type-safe JavaScript
- **[React Router](https://reactrouter.com/)** - Client-side routing
- **[Recharts](https://recharts.org/)** - Data visualization and charting library
- **[Vite](https://vite.dev/)** - Fast build tool and dev server

### Data Sources

- **[17Lands](https://www.17lands.com/)** - Draft statistics and card ratings
- **[Scryfall](https://scryfall.com/)** - Card metadata and images

### Legacy CLI

- **[Fyne](https://fyne.io/)** - GUI framework for draft overlay (CLI mode only)

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

