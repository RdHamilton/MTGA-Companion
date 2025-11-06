# MTGA-Companion

A companion application and overlay system for Magic: The Gathering Arena (MTGA). This tool helps enhance your MTGA experience by reading and analyzing game data from MTGA's log files.

## Features

- **Log Reading**: Automatically locates and reads MTGA Player.log files on both macOS and Windows
- **JSON Parsing**: Parses JSON game events from the log file for analysis
- **Cross-Platform**: Works on both Mac and PC where MTGA is supported

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

### From Source

1. Clone the repository:
   ```bash
   git clone https://github.com/ramonehamilton/MTGA-Companion.git
   cd MTGA-Companion
   ```

2. Build the application:
   ```bash
   go build -o bin/mtga-companion ./cmd/mtga-companion
   ```

   Or use the development script:
   ```bash
   ./scripts/dev.sh build
   ```

3. Run the application:
   ```bash
   ./bin/mtga-companion
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

## Contributing

Contributions are welcome! Please ensure:
- Code follows Go best practices (see `CLAUDE.md`)
- All tests pass (`./scripts/test.sh`)
- Code is formatted (`./scripts/dev.sh fmt`)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Disclaimer

MTGA-Companion is not affiliated with, endorsed by, or sponsored by Wizards of the Coast. Magic: The Gathering Arena and its associated trademarks are property of Wizards of the Coast LLC.

## Acknowledgments

- Wizards of the Coast for MTG Arena and its detailed logging support
- The MTGA community for documentation on log formats and structure
