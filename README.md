# MTGA-Companion

A companion application and overlay system for Magic: The Gathering Arena (MTGA). This tool helps enhance your MTGA experience by reading and analyzing game data from MTGA's log files.

## Features

- **Log Reading**: Automatically locates and reads MTGA Player.log files on both macOS and Windows
- **JSON Parsing**: Parses JSON game events from the log file for analysis
- **Draft Tracking**: Record and analyze all your draft picks
- **Statistics**: Comprehensive win rate tracking and analytics
- **Export System**: Export your data in CSV or JSON formats
- **Card Data Integration**: 17Lands draft statistics and Scryfall metadata
- **Database Storage**: Local SQLite database with migration support
- **Cross-Platform**: Works on both Mac and PC where MTGA is supported

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

### Option 1: Download Pre-built Binary (Recommended)

Download the latest release for your platform from the [Releases page](https://github.com/RdHamilton/MTGA-Companion/releases):

#### Windows

1. Download `mtga-companion-windows-amd64.zip`
2. Extract the ZIP file to a folder of your choice (e.g., `C:\Program Files\MTGA-Companion\`)
3. Run `mtga-companion.exe` to launch the application
4. **(Optional)** Right-click `mtga-companion.exe` → "Send to" → "Desktop (create shortcut)" for easy access

#### macOS

1. Download `mtga-companion-macos-universal.tar.gz`
2. Extract the archive (double-click in Finder)
3. Drag `MTGA Companion.app` to your Applications folder
4. **First launch**: Right-click the app → "Open" (to bypass Gatekeeper since the app is not code-signed)
5. Subsequent launches: Double-click the app normally

**Note**: You may need to grant permissions in System Preferences → Security & Privacy if macOS blocks the app.

#### Linux

1. Download `mtga-companion-linux-amd64.tar.gz`
2. Extract the archive:
   ```bash
   tar -xzf mtga-companion-linux-amd64.tar.gz
   ```
3. Make the binary executable:
   ```bash
   chmod +x mtga-companion
   ```
4. Run the application:
   ```bash
   ./mtga-companion
   ```
5. **(Optional)** Move to system path for global access:
   ```bash
   sudo mv mtga-companion /usr/local/bin/
   ```

### Option 2: Build From Source

If you want to build the application yourself or contribute to development:

1. **Prerequisites**: Install [Go 1.22+](https://go.dev/dl/)

2. Clone the repository:
   ```bash
   git clone https://github.com/RdHamilton/MTGA-Companion.git
   cd MTGA-Companion
   ```

3. Build the application:
   ```bash
   go build -o bin/mtga-companion ./cmd/mtga-companion
   ```

   Or use the development script:
   ```bash
   ./scripts/dev.sh build
   ```

4. Run the application:
   ```bash
   ./bin/mtga-companion
   ```

5. **(Optional)** Create a distribution package:
   ```bash
   ./scripts/package.sh
   ```
   Packaged files will be in the `dist/` directory.

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

Run the companion app after ensuring detailed logging is enabled in MTGA:

```bash
./bin/mtga-companion
```

The application will:
1. Locate your Player.log file
2. Read and parse JSON entries
3. Display information about the log contents

## Development

### Project Structure

```
MTGA-Companion/
├── cmd/
│   └── mtga-companion/      # Application entry point
├── internal/
│   └── mtga/
│       └── logreader/       # Log reading and parsing logic
├── pkg/                     # Public libraries (future)
├── scripts/                 # Development and testing scripts
└── CLAUDE.md               # AI assistant guidance
```

### Development Workflow

Use the provided scripts for common development tasks:

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

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run tests with race detection
go test -race ./...
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

MTGA-Companion is built with:

### Core Technologies

- **[Go 1.23+](https://go.dev/)** - Programming language
- **[SQLite 3](https://www.sqlite.org/)** - Local database storage
- **[modernc.org/sqlite](https://gitlab.com/cznic/sqlite)** - Pure Go SQLite driver

### Libraries & Tools

- **[golang-migrate/migrate](https://github.com/golang-migrate/migrate)** - Database migration management
- **[fsnotify](https://github.com/fsnotify/fsnotify)** - Cross-platform file system notifications

### Data Sources

- **[17Lands](https://www.17lands.com/)** - Draft statistics and card ratings
- **[Scryfall](https://scryfall.com/)** - Card metadata and images

### Future/In Development

- **[Fyne](https://fyne.io/)** - GUI framework (in development)

For a complete list of dependencies, see [`go.mod`](go.mod).

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

