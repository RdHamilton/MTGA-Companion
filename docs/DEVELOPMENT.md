# MTGA-Companion Development Guide

This guide covers setting up your development environment, understanding the codebase, and contributing to MTGA-Companion.

## Table of Contents

- [Development Setup](#development-setup)
- [Project Structure](#project-structure)
- [Development Workflow](#development-workflow)
- [Running the Application](#running-the-application)
- [Debugging](#debugging)
- [Testing](#testing)
- [Code Organization](#code-organization)
- [Adding New Features](#adding-new-features)
- [Contributing Guidelines](#contributing-guidelines)

## Development Setup

### Prerequisites

**Required**:
- **Go 1.23+** - [Download](https://go.dev/dl/)
- **Node.js 20+** - [Download](https://nodejs.org/)
- **Wails CLI** - Desktop app framework
- **Git** - Version control

**Optional but Recommended**:
- **GoLand** or **VS Code** - IDEs with Go support
- **golangci-lint** - Go linter
- **gofumpt** - Go code formatter

### Installing Wails

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Verify installation:
```bash
wails version
# Should show: v2.11.0 or higher
```

### Installing Development Tools

```bash
# golangci-lint (linter)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# gofumpt (formatter)
go install mvdan.cc/gofumpt@latest
```

### Cloning the Repository

```bash
git clone https://github.com/RdHamilton/MTGA-Companion.git
cd MTGA-Companion
```

### Installing Dependencies

**Go dependencies**:
```bash
go mod download
go mod tidy
```

**Frontend dependencies**:
```bash
cd frontend
npm install
cd ..
```

### Verifying Setup

```bash
# Check Wails can build the project
wails doctor

# Should show all green checks for:
# - Go installation
# - Node installation
# - Platform-specific dependencies
```

## Project Structure

```
MTGA-Companion/
├── main.go                  # Wails entry point (GUI app)
├── app.go                   # Wails backend API
├── frontend/                # React + TypeScript frontend
│   ├── src/
│   │   ├── components/     # Reusable UI components
│   │   ├── pages/          # Page components (routes)
│   │   ├── App.tsx         # Root component
│   │   └── main.tsx        # Frontend entry point
│   ├── wailsjs/            # Auto-generated Wails bindings
│   ├── package.json
│   └── vite.config.ts
├── cmd/                     # CLI commands
│   └── mtga-companion/
│       ├── main.go         # CLI entry point
│       ├── daemon.go       # Daemon mode implementation
│       └── service.go      # Service management commands
├── internal/                # Private application code
│   ├── gui/                # GUI-specific backend code
│   ├── mtga/               # MTGA-specific logic
│   │   ├── logreader/     # Log parsing
│   │   ├── poller/        # Log file monitoring
│   │   ├── logprocessor/  # Shared log processing
│   │   └── draft/         # Draft overlay
│   ├── storage/            # Database and persistence
│   │   ├── models/        # Data models
│   │   └── repository/    # Data access layer
│   └── ipc/                # WebSocket IPC (client/server)
├── docs/                    # Documentation
├── scripts/                 # Development scripts
└── CLAUDE.md               # AI assistant guidance
```

### Key Files

**Wails Application**:
- `main.go` - Entry point for GUI app, initializes Wails
- `app.go` - Go backend methods callable from TypeScript frontend
- `frontend/` - React frontend code

**CLI Application**:
- `cmd/mtga-companion/main.go` - Entry point for CLI
- `cmd/mtga-companion/daemon.go` - Daemon mode (background service)
- `cmd/mtga-companion/service.go` - Service installation/management

**Shared Code**:
- `internal/mtga/logprocessor/` - Log parsing (used by both GUI and daemon)
- `internal/storage/` - Database access (shared)
- `internal/ipc/` - WebSocket IPC (client in GUI, server in daemon)

## Development Workflow

### Development Scripts

**Backend Development** (`./scripts/dev.sh`):
```bash
./scripts/dev.sh           # Run all checks and build
./scripts/dev.sh fmt       # Format code
./scripts/dev.sh vet       # Run go vet
./scripts/dev.sh lint      # Run golangci-lint
./scripts/dev.sh check     # Run fmt, vet, and lint
./scripts/dev.sh build     # Build CLI application
```

**Testing** (`./scripts/test.sh`):
```bash
./scripts/test.sh          # Run all tests with race detection
./scripts/test.sh coverage # Generate coverage report
./scripts/test.sh verbose  # Run with verbose output
```

### Git Workflow

**Branch naming**:
- Feature: `feature/feature-name`
- Bug fix: `fix/bug-description`
- Documentation: `docs/topic`

**Commit messages**:
- Follow conventional commits
- Examples:
  - `feat: Add draft overlay support`
  - `fix: Resolve database lock on GUI startup`
  - `docs: Update installation guide`
  - `refactor: Extract log processor into shared package`

**PR process**:
1. Create feature branch from `main`
2. Implement changes with tests
3. Run `./scripts/dev.sh check` - ensure all checks pass
4. Run `./scripts/test.sh` - ensure all tests pass
5. Push branch and create PR
6. Address review comments
7. Merge when CI passes and approved

## Running the Application

### GUI Development Mode (Recommended)

**Run with hot reload**:
```bash
wails dev
```

This starts:
- Wails development server
- Frontend dev server (Vite) with hot reload
- Go backend with live reload
- Opens app in development mode

**Access**:
- App window opens automatically
- Frontend runs on http://localhost:34115 (internal)
- Changes to frontend files reload instantly
- Changes to Go files rebuild and restart

### GUI in Standalone Mode

When daemon is not running, GUI automatically falls back to standalone mode with embedded log poller.

**To develop standalone mode**:
```bash
# Just run wails dev - it will start in standalone mode
# if no daemon is detected
wails dev
```

### Daemon Development Mode

**Run daemon separately**:
```bash
# Build CLI first
go build -o bin/mtga-companion ./cmd/mtga-companion

# Run daemon
./bin/mtga-companion daemon

# With debug logging
./bin/mtga-companion daemon --debug-mode
```

**Run daemon + GUI together**:

Terminal 1 (Daemon):
```bash
./bin/mtga-companion daemon --debug-mode
```

Terminal 2 (GUI):
```bash
wails dev
```

GUI will connect to daemon via WebSocket.

### CLI Development

**Build and run CLI**:
```bash
go build -o bin/mtga-companion ./cmd/mtga-companion
./bin/mtga-companion read
./bin/mtga-companion export stats -json
```

### GoLand/IDE Configuration

**For Wails GUI Development**:

1. **Create Run Configuration**:
   - Name: "Wails Dev"
   - Type: Shell Script
   - Execute: `wails dev`
   - Working directory: `$PROJECT_DIR$`

2. **For Daemon Development**:
   - Name: "Daemon"
   - Type: Go Build
   - Package path: `github.com/ramonehamilton/MTGA-Companion/cmd/mtga-companion`
   - Program arguments: `daemon --debug-mode`
   - Working directory: `$PROJECT_DIR$`

3. **Compound Configuration** (Run Both):
   - Create compound configuration
   - Add "Daemon" and "Wails Dev"
   - Run both simultaneously

## Debugging

### Backend Debugging (Go)

**Add debug logging**:
```go
import "log"

log.Printf("[DEBUG] Variable value: %+v", variable)
```

**Use debugger in IDE**:
- Set breakpoints in Go code
- Run `wails dev` in debug mode
- Debugger attaches to Go backend process

**Enable debug mode**:
```bash
# Daemon
./bin/mtga-companion daemon --debug-mode

# GUI (set environment variable)
DEBUG=true wails dev
```

### Frontend Debugging (React)

**Browser DevTools**:
- Right-click in app → "Inspect Element"
- Console tab shows `console.log()` output
- React DevTools available

**Console logging**:
```typescript
console.log('Debug info:', data);
console.error('Error occurred:', error);
```

**Wails runtime events**:
```typescript
import { EventsOn, EventsOff } from '../wailsjs/runtime';

EventsOn('debug:event', (data) => {
  console.log('Event received:', data);
});
```

### WebSocket Debugging

**Test daemon WebSocket**:
```bash
# Check if daemon is listening
curl http://localhost:9999/status

# Expected: {"status":"ok"}
```

**Monitor WebSocket traffic**:
```go
// In daemon (internal/ipc/server.go)
log.Printf("[WS] Broadcasting event: %s", eventType)

// In GUI (internal/ipc/client.go)
log.Printf("[WS] Received event: %s", event.Type)
```

**Frontend WebSocket debugging**:
```typescript
// Check connection status
const status = await GetConnectionStatus();
console.log('Daemon connection:', status);
```

### Database Debugging

**Inspect SQLite database**:
```bash
# Open database
sqlite3 ~/.mtga-companion/data.db

# List tables
.tables

# Query matches
SELECT * FROM matches ORDER BY created_at DESC LIMIT 10;

# Exit
.quit
```

**Enable SQL query logging**:
```go
// In storage/repository code
log.Printf("[SQL] Query: %s, Args: %v", query, args)
```

## Testing

### Running Tests

**All tests**:
```bash
go test ./...
```

**With coverage**:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

**Specific package**:
```bash
go test ./internal/storage
go test ./internal/mtga/logprocessor -v
```

**With race detection**:
```bash
go test -race ./...
```

### Writing Tests

**Unit test example**:
```go
package storage

import (
    "testing"
)

func TestMatchRepository_GetMatches(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    defer db.Close()

    repo := NewMatchRepository(db)

    // Test
    matches, err := repo.GetMatches(Filter{})

    // Assert
    if err != nil {
        t.Fatalf("Expected no error, got: %v", err)
    }
    if len(matches) != 5 {
        t.Errorf("Expected 5 matches, got: %d", len(matches))
    }
}
```

**Table-driven tests**:
```go
func TestParseLogEntry(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *LogEntry
        wantErr bool
    }{
        {
            name:  "valid entry",
            input: `{"type":"MatchCreated","data":{...}}`,
            want:  &LogEntry{Type: "MatchCreated"},
            wantErr: false,
        },
        {
            name:    "invalid JSON",
            input:   `{invalid`,
            want:    nil,
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParseLogEntry(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("ParseLogEntry() error = %v, wantErr %v", err, tt.wantErr)
            }
            // ... more assertions
        })
    }
}
```

### Frontend Testing

Frontend testing uses **Vitest** for component tests and **Playwright** for E2E tests.

**Component Tests (Vitest + React Testing Library)**:
```bash
cd frontend

# Run in watch mode (development)
npm run test

# Run once (CI)
npm run test:run

# Run with UI
npm run test:ui

# Generate coverage report
npm run test:coverage
```

**E2E Tests (Playwright)**:
```bash
# Start the Wails dev server first
wails dev

# In another terminal:
cd frontend

# Run E2E tests (headless)
npm run test:e2e

# Run with interactive UI
npm run test:e2e:ui

# Run in debug mode
npm run test:e2e:debug

# View test report
npx playwright show-report
```

**Component Test Example**:
```typescript
import { describe, it, expect, vi } from 'vitest';
import { screen, waitFor } from '@testing-library/react';
import { render } from '../test/utils/testUtils';
import { mockWailsApp } from '../test/mocks/wailsApp';
import Footer from './Footer';

describe('Footer', () => {
  it('should display match statistics', async () => {
    // Mock the backend response
    mockWailsApp.GetStats.mockResolvedValue({
      TotalMatches: 100,
      WinRate: 0.6,
    });

    render(<Footer />);

    // Wait for async data to load
    await waitFor(() => {
      expect(screen.getByText('100')).toBeInTheDocument();
      expect(screen.getByText(/60%/)).toBeInTheDocument();
    });
  });
});
```

**E2E Test Example**:
```typescript
import { test, expect } from '@playwright/test';

test('should navigate to draft view', async ({ page }) => {
  await page.goto('/');
  await expect(page.locator('.app-container')).toBeVisible();

  await page.getByText('Draft').click();
  await expect(page).toHaveURL(/\/draft/);
});
```

**Testing Best Practices**:
- Test user behavior, not implementation details
- Use meaningful test descriptions: "should [do X] when [Y condition]"
- Mock all Wails backend calls using `mockWailsApp`
- Use `waitFor` for async operations, not fixed timeouts
- Test loading states, error states, and empty states
- Keep tests isolated and independent

For comprehensive testing documentation, see [docs/TESTING.md](./TESTING.md).

## Code Organization

### Backend Patterns

**Repository Pattern**:
```go
// Define interface
type MatchRepository interface {
    GetMatches(filter Filter) ([]*Match, error)
    GetMatchByID(id string) (*Match, error)
    SaveMatch(match *Match) error
}

// Implement interface
type matchRepository struct {
    db *sql.DB
}

func NewMatchRepository(db *sql.DB) MatchRepository {
    return &matchRepository{db: db}
}
```

**Dependency Injection**:
```go
// App struct holds dependencies
type App struct {
    db         *storage.DB
    ipcClient  *ipc.Client
    poller     *poller.Poller
}

// Injected via constructor
func NewApp(db *storage.DB) *App {
    return &App{
        db: db,
    }
}
```

**Error Handling**:
```go
// Return errors, don't panic
func GetMatch(id string) (*Match, error) {
    match, err := repo.GetMatchByID(id)
    if err != nil {
        return nil, fmt.Errorf("failed to get match: %w", err)
    }
    return match, nil
}

// Wrap errors for context
if err != nil {
    return fmt.Errorf("parsing log entry: %w", err)
}
```

### Frontend Patterns

**Hooks for data fetching**:
```typescript
const [matches, setMatches] = useState<Match[]>([]);
const [loading, setLoading] = useState(true);

useEffect(() => {
    loadMatches();
}, []);

const loadMatches = async () => {
    try {
        const data = await GetMatches();
        setMatches(data);
    } catch (error) {
        console.error('Failed to load matches:', error);
    } finally {
        setLoading(false);
    }
};
```

**Event listeners**:
```typescript
useEffect(() => {
    // Subscribe to events
    EventsOn('match:new', () => {
        loadMatches(); // Refresh data
    });

    // Cleanup
    return () => {
        EventsOff('match:new');
    };
}, []);
```

## Adding New Features

### Adding a New WebSocket Event

**1. Emit event from daemon** (`cmd/mtga-companion/daemon.go`):
```go
server.Broadcast("inventory:updated", map[string]interface{}{
    "gems": 1500,
    "gold": 5000,
})
```

**2. Handle in GUI backend** (`app.go`):
```go
func (a *App) handleInventoryUpdate(data map[string]interface{}) {
    // Process data
    log.Printf("Inventory updated: %+v", data)

    // Forward to frontend
    runtime.EventsEmit(a.ctx, "inventory:updated", data)
}
```

**3. Listen in frontend** (`frontend/src/pages/Inventory.tsx`):
```typescript
useEffect(() => {
    EventsOn('inventory:updated', (data: any) => {
        setInventory(data);
    });

    return () => {
        EventsOff('inventory:updated');
    };
}, []);
```

### Adding a New Database Table

**1. Create migration** (`internal/storage/migrations/0005_add_inventory.up.sql`):
```sql
CREATE TABLE inventory (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    gems INTEGER NOT NULL,
    gold INTEGER NOT NULL,
    wildcards_common INTEGER NOT NULL,
    updated_at DATETIME NOT NULL
);
```

**2. Create down migration** (`0005_add_inventory.down.sql`):
```sql
DROP TABLE inventory;
```

**3. Create model** (`internal/storage/models/inventory.go`):
```go
type Inventory struct {
    ID              int
    Gems            int
    Gold            int
    WildcardsCommon int
    UpdatedAt       time.Time
}
```

**4. Create repository** (`internal/storage/inventory_repository.go`):
```go
type InventoryRepository interface {
    GetLatest() (*models.Inventory, error)
    Save(inv *models.Inventory) error
}

type inventoryRepository struct {
    db *sql.DB
}

func (r *inventoryRepository) GetLatest() (*models.Inventory, error) {
    // Implementation
}
```

**5. Run migration**:
```bash
./bin/mtga-companion migrate up
```

### Adding a New Wails Backend Method

**1. Add method to App** (`app.go`):
```go
// GetInventory returns current inventory
func (a *App) GetInventory() (*models.Inventory, error) {
    return a.db.Inventory.GetLatest()
}
```

**2. Generate TypeScript bindings**:
```bash
wails generate module
```

**3. Use in frontend**:
```typescript
import { GetInventory } from '../wailsjs/go/main/App';

const inventory = await GetInventory();
```

## Contributing Guidelines

### Before Submitting PR

**Checklist**:
- [ ] Code follows Go best practices (see `CLAUDE.md`)
- [ ] All tests pass (`./scripts/test.sh`)
- [ ] Code is formatted (`./scripts/dev.sh fmt`)
- [ ] Linter passes (`./scripts/dev.sh lint`)
- [ ] No new compiler warnings
- [ ] Documentation updated if needed
- [ ] PR references issue number (e.g., "Closes #123")

### Code Style

**Go**:
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofumpt` for formatting
- Pass `golangci-lint` checks
- Keep functions small and focused
- Write tests for new functionality

**TypeScript/React**:
- Use TypeScript strict mode
- Functional components with hooks
- Props interfaces for all components
- Meaningful variable names
- Comments for complex logic

### Review Process

1. **Submit PR** - Create PR with clear description
2. **Automated Checks** - CI runs tests, linter, build
3. **Code Review** - Maintainer reviews code
4. **Address Feedback** - Make requested changes
5. **Approval** - PR approved by maintainer
6. **Merge** - PR merged to main

### Getting Help

**Resources**:
- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture
- [DAEMON_API.md](DAEMON_API.md) - WebSocket API reference
- [CLAUDE.md](../CLAUDE.md) - Project conventions

**Support**:
- GitHub Issues - Report bugs or request features
- GitHub Discussions - Ask questions
- Discord - (Future: Community chat)

## Common Development Tasks

### Regenerate Wails Bindings

After changing Go backend methods in `app.go`:
```bash
wails generate module
```

This regenerates:
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/go/main/App.d.ts`

### Build Production Binary

**Wails app** (GUI):
```bash
wails build
# Output: build/bin/MTGA-Companion.app (macOS)
# Output: build/bin/MTGA-Companion.exe (Windows)
```

**CLI binary**:
```bash
go build -o bin/mtga-companion ./cmd/mtga-companion
```

### Update Dependencies

**Go modules**:
```bash
go get -u ./...
go mod tidy
```

**Frontend**:
```bash
cd frontend
npm update
cd ..
```

### Database Migrations

**Create new migration**:
```bash
# Manually create files:
# internal/storage/migrations/0006_description.up.sql
# internal/storage/migrations/0006_description.down.sql
```

**Apply migrations**:
```bash
./bin/mtga-companion migrate up
```

**Rollback**:
```bash
./bin/mtga-companion migrate down
```

**Check status**:
```bash
./bin/mtga-companion migrate status
```

## Performance Profiling

### CPU Profiling

```bash
go test -cpuprofile=cpu.prof -bench=. ./internal/mtga/logprocessor
go tool pprof cpu.prof
```

### Memory Profiling

```bash
go test -memprofile=mem.prof -bench=. ./internal/storage
go tool pprof mem.prof
```

### Race Detection

```bash
go test -race ./...
```

## Security Considerations

### Database

- SQLite file permissions (user-only access)
- Prepared statements to prevent SQL injection
- No sensitive data in database

### WebSocket

- Daemon listens on localhost only (not network-accessible)
- No authentication required (local-only access)
- Future: TLS for network access

### Log Files

- Read-only access to MTGA Player.log
- Never write to game files
- No sensitive data logged

## Troubleshooting Development Issues

### "Wails not found"

```bash
# Ensure Go bin is in PATH
export PATH=$PATH:$(go env GOPATH)/bin

# Or add to ~/.zshrc or ~/.bashrc
```

### "Frontend dependencies missing"

```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
cd ..
```

### "Database is locked"

Stop all running instances:
```bash
./bin/mtga-companion service stop
killall mtga-companion
```

### "Port 9999 already in use"

Find and kill process using port:
```bash
# macOS/Linux
lsof -ti:9999 | xargs kill -9

# Windows
netstat -ano | findstr :9999
taskkill /PID <PID> /F
```

## Next Steps

- Read [ARCHITECTURE.md](ARCHITECTURE.md) to understand the system design
- Review [DAEMON_API.md](DAEMON_API.md) for WebSocket events
- Check [CLAUDE.md](../CLAUDE.md) for coding principles
- Start contributing! Check [good first issues](https://github.com/RdHamilton/MTGA-Companion/labels/good%20first%20issue)
