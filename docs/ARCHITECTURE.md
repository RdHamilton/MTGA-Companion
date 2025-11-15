# MTGA-Companion Architecture

## Overview

MTGA-Companion uses a service-based architecture that separates data collection from data display. This document describes the system design, component responsibilities, data flow, and extension points.

## Architecture Diagram

```
┌──────────────────────────────────────────────────────────────┐
│                    MTGA (Game Client)                        │
│                                                               │
│  Plays matches, generates game events, writes detailed logs  │
└──────────────────────┬───────────────────────────────────────┘
                       │ writes game events
                       ↓
┌──────────────────────────────────────────────────────────────┐
│                      Player.log File                         │
│                                                               │
│  JSON-formatted log entries (matches, drafts, inventory)     │
└──────────────────────┬───────────────────────────────────────┘
                       │ monitors (fsnotify or polling)
                       ↓
┌──────────────────────────────────────────────────────────────┐
│                    CLI Daemon (Backend)                      │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                    Log Monitoring                      │ │
│  │  ┌─────────────┐                ┌─────────────┐       │ │
│  │  │   Poller    │────monitors───▶│ File Events │       │ │
│  │  │  (Goroutine)│                │  (fsnotify) │       │ │
│  │  └──────┬──────┘                └──────┬──────┘       │ │
│  │         │                               │              │ │
│  │         └──────────────┬────────────────┘              │ │
│  │                        │ new entries                   │ │
│  │                        ↓                               │ │
│  │              ┌────────────────┐                        │ │
│  │              │  Log Processor │                        │ │
│  │              │                │                        │ │
│  │              │  - Parses JSON │                        │ │
│  │              │  - Validates   │                        │ │
│  │              │  - Routes data │                        │ │
│  │              └────────┬───────┘                        │ │
│  └─────────────────────────────────────────────────────────┘ │
│                         │ parsed data                       │
│                         ↓                                    │
│  ┌──────────────────────────────────────────────────────────┤
│  │                   Data Storage                           │
│  │  ┌─────────────────────────────────────────────────────┐ │
│  │  │                Repository Layer                     │ │
│  │  │  ┌───────────┐  ┌──────────┐  ┌────────────┐      │ │
│  │  │  │  Matches  │  │  Drafts  │  │  Settings  │      │ │
│  │  │  │Repository │  │Repository│  │ Repository │      │ │
│  │  │  └─────┬─────┘  └─────┬────┘  └─────┬──────┘      │ │
│  │  │        │               │             │              │ │
│  │  │        └───────────────┴─────────────┘              │ │
│  │  │                        │                            │ │
│  │  │                        ↓                            │ │
│  │  │            ┌────────────────────────┐              │ │
│  │  │            │  SQLite Database       │              │ │
│  │  │            │  ~/.mtga-companion/    │              │ │
│  │  │            │  data.db               │              │ │
│  │  │            └────────────────────────┘              │ │
│  │  └─────────────────────────────────────────────────────┘ │
│  └──────────────────────────────────────────────────────────┘
│                         │ data stored                       │
│                         ↓                                    │
│  ┌──────────────────────────────────────────────────────────┤
│  │                  Event Broadcasting                      │
│  │  ┌─────────────────────────────────────────────────────┐ │
│  │  │              WebSocket Server                       │ │
│  │  │                                                     │ │
│  │  │  Listen on: ws://localhost:9999                    │ │
│  │  │                                                     │ │
│  │  │  Events:                                            │ │
│  │  │  - stats:updated                                    │ │
│  │  │  - match:new                                        │ │
│  │  │  - draft:started                                    │ │
│  │  │  - draft:pick                                       │ │
│  │  └─────────────────────────────────────────────────────┘ │
│  └──────────────────────┬───────────────────────────────────┘
└─────────────────────────┼───────────────────────────────────┘
                          │ WebSocket events
                          ↓
         ┌────────────────────────────────┐
         │     WebSocket Clients          │
         │  (Any client can connect)      │
         └────────────────┬───────────────┘
                          │ connects
                          ↓
┌──────────────────────────────────────────────────────────────┐
│                    GUI (Frontend - Wails)                    │
│  ┌────────────────────────────────────────────────────────┐ │
│  │                   IPC Client Layer                     │ │
│  │  ┌─────────────────────────────────────────────────── ││ │
│  │  │         WebSocket Connection Handler              ││ │
│  │  │                                                    ││ │
│  │  │  - Connect to daemon (ws://localhost:9999)        ││ │
│  │  │  - Subscribe to events                            ││ │
│  │  │  - Handle reconnection                            ││ │
│  │  │  - Automatic fallback to standalone               ││ │
│  │  └────────────────────────────────────────────────────┘│ │
│  │                        │                               │ │
│  │                        │ events                        │ │
│  │                        ↓                               │ │
│  │  ┌────────────────────────────────────────────────────┐│ │
│  │  │              Event Handlers                        ││ │
│  │  │                                                    ││ │
│  │  │  - stats:updated → Refresh statistics display     ││ │
│  │  │  - match:new → Update match history               ││ │
│  │  │  - draft:pick → Update draft overlay              ││ │
│  │  └──────────────────────┬─────────────────────────────┘│ │
│  └─────────────────────────┼──────────────────────────────┘ │
│                            │ trigger UI updates             │
│                            ↓                                 │
│  ┌──────────────────────────────────────────────────────────┤
│  │                   React Frontend                         │
│  │  ┌─────────────────────────────────────────────────────┐ │
│  │  │                   Pages/Views                       │ │
│  │  │  ┌──────────┐  ┌───────────┐  ┌────────────┐      │ │
│  │  │  │  Match   │  │  Charts   │  │  Settings  │      │ │
│  │  │  │ History  │  │  & Stats  │  │            │      │ │
│  │  │  └──────────┘  └───────────┘  └────────────┘      │ │
│  │  │                                                     │ │
│  │  │  Components fetch data via:                        │ │
│  │  │  - Go backend methods (via Wails bindings)         │ │
│  │  │  - Real-time updates from WebSocket events         │ │
│  │  └─────────────────────────────────────────────────────┘ │
│  └──────────────────────────────────────────────────────────┘
└──────────────────────────────────────────────────────────────┘

Standalone Mode (Fallback):
┌──────────────────────────────────────────────────────────────┐
│                    GUI (Standalone Mode)                     │
│  ┌────────────────────────────────────────────────────────┐ │
│  │              Embedded Log Poller                       │ │
│  │  (Same functionality as daemon, runs in GUI process)   │ │
│  └────────────────────┬───────────────────────────────────┘ │
│                       │                                      │
│                       ↓                                      │
│  ┌────────────────────────────────────────────────────────┐ │
│  │              Direct Database Access                    │ │
│  │  (No WebSocket, direct repository calls)               │ │
│  └────────────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

## Component Responsibilities

### 1. CLI Daemon (Backend Service)

**Location**: `cmd/mtga-companion/daemon.go`

**Responsibilities**:
- Monitor MTGA `Player.log` file for changes
- Parse JSON log entries into structured data
- Store data in SQLite database
- Broadcast events to WebSocket clients
- Run as background service (24/7 operation)
- Automatic crash recovery via service manager

**Key Components**:

**Log Poller** (`internal/mtga/poller/poller.go`):
- Monitors log file for changes using fsnotify or polling
- Detects new entries and log rotation
- Handles file system events (create, write, rename, remove)
- Configurable poll interval

**Log Processor** (`internal/mtga/logprocessor/processor.go`):
- Shared component used by both daemon and standalone GUI
- Parses JSON log entries
- Routes data to appropriate storage repositories
- Handles match tracking, draft tracking, inventory updates

**WebSocket Server** (`internal/ipc/server.go`):
- Listens on port 9999 (configurable)
- Manages client connections
- Broadcasts events to all connected clients
- Handles client disconnection gracefully

**Storage Layer** (`internal/storage/`):
- Repository pattern for data access
- SQLite database with migration support
- Repositories: matches, drafts, statistics, settings

### 2. GUI (Frontend Application)

**Location**: `main.go`, `app.go`, `frontend/`

**Responsibilities**:
- Connect to daemon via WebSocket
- Display match history, statistics, charts
- Handle user interactions and settings
- Automatic fallback to standalone mode if daemon unavailable
- Real-time UI updates via event listeners

**Key Components**:

**IPC Client** (`internal/ipc/client.go`):
- WebSocket client that connects to daemon
- Subscribes to events (stats:updated, match:new, etc.)
- Handles reconnection with exponential backoff
- Detects daemon availability and falls back to standalone

**Wails Backend** (`app.go`):
- Go backend methods callable from TypeScript frontend
- Example: `GetMatches()`, `GetStatistics()`, `SetDaemonPort()`
- Manages IPC client connection
- Controls standalone poller when daemon unavailable

**React Frontend** (`frontend/src/`):
- TypeScript + React 18
- Pages: Match History, Charts, Settings
- Components: Layout, tables, charts, status indicators
- Uses Wails bindings to call Go backend methods
- Listens for WebSocket events via `EventsOn()`

### 3. Shared Components

**Log Processor** (`internal/mtga/logprocessor/`):
- Shared by both daemon and standalone GUI
- Single source of truth for log parsing logic
- Parses matches, drafts, inventory, rank progression

**Storage Repositories** (`internal/storage/`):
- Direct database access layer
- Used by both daemon and standalone GUI
- Consistent data access patterns

## Data Flow

### Normal Operation (Daemon Mode)

```
1. MTGA writes to Player.log
   │
   ↓
2. Daemon's poller detects change (fsnotify event)
   │
   ↓
3. Daemon reads new log entries
   │
   ↓
4. Log processor parses JSON entries
   │
   ↓
5. Parsed data validated and routed
   │
   ↓
6. Repository stores data in SQLite
   │
   ↓
7. WebSocket server broadcasts event
   │  Example: {"type": "match:new", "data": {...}}
   │
   ↓
8. GUI's IPC client receives event
   │
   ↓
9. Event handler triggers data refresh
   │
   ↓
10. GUI fetches updated data from database
    │  (via Wails backend methods)
    │
    ↓
11. React components re-render with new data
    │
    ↓
12. User sees updated statistics/match history
```

### Standalone Mode (Fallback)

```
1. MTGA writes to Player.log
   │
   ↓
2. GUI's embedded poller detects change
   │
   ↓
3. GUI's log processor parses entries
   │
   ↓
4. GUI writes directly to database
   │
   ↓
5. GUI triggers internal event
   │
   ↓
6. React components refresh
```

### GUI Startup Flow

```
1. GUI starts, initializes Wails backend
   │
   ↓
2. Backend attempts to connect to daemon
   │  (ws://localhost:9999)
   │
   ├─ Success: Connected to daemon
   │  │
   │  ├─ Subscribe to WebSocket events
   │  ├─ Load initial data from database
   │  └─ Display "Connected" status
   │
   └─ Failure: Daemon not available
      │
      ├─ Log: "Daemon not available, falling back to standalone"
      ├─ Start embedded log poller
      ├─ Load data directly from database
      └─ Display "Standalone Mode" status
```

## WebSocket Event Protocol

### Connection

**URL**: `ws://localhost:9999`

**Connection handshake**:
1. Client connects via WebSocket
2. Server accepts connection
3. Client subscribes to events
4. Server broadcasts events to all clients

### Event Types

All events follow this structure:
```json
{
  "type": "event:name",
  "data": { ... },
  "timestamp": "2025-11-15T10:30:00Z"
}
```

**Available Events**:

**`stats:updated`** - Overall statistics changed
```json
{
  "type": "stats:updated",
  "data": {
    "totalMatches": 150,
    "totalGames": 300,
    "winRate": 0.63
  },
  "timestamp": "2025-11-15T10:30:00Z"
}
```

**`match:new`** - New match recorded
```json
{
  "type": "match:new",
  "data": {
    "matchID": "abc-123",
    "result": "Win",
    "format": "ConstructedRanked"
  },
  "timestamp": "2025-11-15T10:30:00Z"
}
```

**`draft:started`** - Draft session started
```json
{
  "type": "draft:started",
  "data": {
    "draftID": "draft-789",
    "setCode": "ONE"
  },
  "timestamp": "2025-11-15T10:30:00Z"
}
```

**`draft:pick`** - Card picked in draft
```json
{
  "type": "draft:pick",
  "data": {
    "draftID": "draft-789",
    "pack": 1,
    "pick": 3,
    "cardID": 89765
  },
  "timestamp": "2025-11-15T10:30:00Z"
}
```

**`connection:status`** - Connection state changed
```json
{
  "type": "connection:status",
  "data": {
    "status": "connected"
  },
  "timestamp": "2025-11-15T10:30:00Z"
}
```

### Error Handling

WebSocket errors are handled with automatic reconnection:

1. **Connection Lost**: Client attempts reconnection with exponential backoff
   - 1st retry: 1 second
   - 2nd retry: 2 seconds
   - 3rd retry: 4 seconds
   - Max: 30 seconds between retries

2. **Daemon Unavailable**: Client falls back to standalone mode
   - Embedded poller starts
   - No WebSocket connection maintained
   - GUI continues functioning normally

3. **Daemon Recovers**: Client automatically reconnects
   - Detects daemon is available again
   - Stops embedded poller
   - Reconnects to daemon WebSocket

## Database Schema

### Tables

**matches**
```sql
CREATE TABLE matches (
    id TEXT PRIMARY KEY,
    event_type TEXT,
    format TEXT,
    result TEXT,
    opponent_id TEXT,
    start_time DATETIME,
    end_time DATETIME,
    duration INTEGER,
    created_at DATETIME
);
```

**games**
```sql
CREATE TABLE games (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    match_id TEXT,
    game_number INTEGER,
    result TEXT,
    on_play BOOLEAN,
    created_at DATETIME,
    FOREIGN KEY (match_id) REFERENCES matches(id)
);
```

**drafts**
```sql
CREATE TABLE drafts (
    id TEXT PRIMARY KEY,
    event_id TEXT,
    set_code TEXT,
    status TEXT,
    created_at DATETIME,
    completed_at DATETIME
);
```

**draft_picks**
```sql
CREATE TABLE draft_picks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    draft_id TEXT,
    pack INTEGER,
    pick INTEGER,
    card_id INTEGER,
    created_at DATETIME,
    FOREIGN KEY (draft_id) REFERENCES drafts(id)
);
```

**settings**
```sql
CREATE TABLE settings (
    key TEXT PRIMARY KEY,
    value TEXT,
    updated_at DATETIME
);
```

### Migrations

Database migrations are managed with `golang-migrate/migrate`.

**Migration files**: `internal/storage/migrations/`

**Naming convention**: `NNNN_description.up.sql` / `NNNN_description.down.sql`

**Running migrations**:
```bash
# Apply all pending migrations
./mtga-companion migrate up

# Rollback last migration
./mtga-companion migrate down

# Check migration status
./mtga-companion migrate status
```

## Security Considerations

### WebSocket Security

**Current**: WebSocket listens on `localhost:9999` only
- Not network-accessible by default
- Only accepts connections from same machine
- No authentication required (local-only)

**Future Enhancement**: For network access, consider:
- TLS encryption (wss://)
- Authentication tokens
- CORS configuration
- Rate limiting

### Database Access

**Protection**: SQLite database is local file with file system permissions
- Located at `~/.mtga-companion/data.db`
- Only accessible by user who owns the file
- No network exposure

**Concurrent Access**: Database locking handled by SQLite
- Only one writer at a time (daemon OR standalone GUI)
- Multiple readers allowed
- Lock timeout: 5 seconds

### Log File Access

**Read-only**: Application only reads `Player.log`, never writes
- No risk of corrupting MTGA game state
- Detection of log rotation and recovery

## Extension Points

### Adding New Event Types

1. **Define event in daemon** (`cmd/mtga-companion/daemon.go`):
   ```go
   server.Broadcast("new:event", map[string]interface{}{
       "data": eventData,
   })
   ```

2. **Handle event in GUI** (`internal/gui/app.go`):
   ```go
   func (a *App) handleNewEvent(event map[string]interface{}) {
       // Process event, update UI
       runtime.EventsEmit(a.ctx, "new:event", event)
   }
   ```

3. **Listen in frontend** (`frontend/src/components/Component.tsx`):
   ```typescript
   EventsOn('new:event', (data) => {
       // Update React state
   });
   ```

### Adding New Data Sources

To track additional MTGA data (e.g., inventory, collection):

1. **Update log processor** (`internal/mtga/logprocessor/processor.go`):
   ```go
   func (p *Processor) ProcessInventoryUpdate(entry JSONEntry) {
       // Parse inventory data
       // Store in database
       // Broadcast event
   }
   ```

2. **Add repository method** (`internal/storage/inventory_repository.go`):
   ```go
   func (r *InventoryRepository) SaveInventory(inv *Inventory) error {
       // Database insert/update
   }
   ```

3. **Create migration** (`internal/storage/migrations/0004_add_inventory.up.sql`):
   ```sql
   CREATE TABLE inventory (...);
   ```

4. **Add GUI method** (`app.go`):
   ```go
   func (a *App) GetInventory() (*Inventory, error) {
       return a.db.Inventory.GetLatest()
   }
   ```

### Adding New Frontend Clients

The daemon can support multiple frontend types:

**Web Frontend**:
- Connect to `ws://localhost:9999` from browser
- Use same WebSocket event protocol
- Implement own UI (React, Vue, Angular, etc.)

**Mobile App**:
- Connect via WebSocket from mobile device
- Daemon would need network binding (not just localhost)
- Implement authentication for security

**Third-Party Tools**:
- Any WebSocket client can connect
- Subscribe to specific events
- Build custom integrations (Discord bot, OBS overlay, etc.)

## Technology Stack

### Backend (Go)

- **Language**: Go 1.23+
- **Database**: SQLite3 via `modernc.org/sqlite` (pure Go, no CGo)
- **Migrations**: `golang-migrate/migrate`
- **WebSocket**: `gorilla/websocket`
- **File Watching**: `fsnotify/fsnotify`
- **Service Management**: `kardianos/service`

### Frontend (React + Wails)

- **Framework**: Wails v2 (Go + Web)
- **UI Library**: React 18
- **Language**: TypeScript
- **Build Tool**: Vite
- **Routing**: React Router
- **Charts**: Recharts
- **Webview**: Native (WebKit on macOS, WebView2 on Windows)

### Platform Support

- **macOS**: Launch Agents (launchd)
- **Windows**: Windows Service (Service Control Manager)
- **Linux**: systemd units

## Performance Characteristics

### Resource Usage

**Daemon**:
- Memory: ~10-20 MB
- CPU: < 1% idle, ~5% during log processing
- Disk I/O: Minimal (reads log, writes database)

**GUI (Connected)**:
- Memory: ~50-100 MB (includes WebView)
- CPU: < 1% idle, ~10% during rendering
- Network: WebSocket only (localhost, negligible)

**GUI (Standalone)**:
- Memory: ~60-120 MB (includes WebView + poller)
- CPU: < 1% idle, ~10% during log processing + rendering

### Scalability

**Database**:
- SQLite handles millions of rows efficiently
- Indexed queries for fast lookups
- Database file size: ~1-5 MB per 1000 matches

**WebSocket**:
- Supports dozens of concurrent clients
- Broadcast overhead minimal (< 1ms per event)
- No performance degradation with multiple GUIs

## Monitoring and Debugging

### Daemon Logs

**macOS**: `~/Library/Logs/MTGACompanionDaemon.log`
**Windows**: Event Viewer → Application → MTGACompanionDaemon
**Linux**: `journalctl -u MTGACompanionDaemon -f`

### Debug Mode

Enable debug logging:
```bash
./mtga-companion daemon --debug-mode
```

Outputs:
- WebSocket connection events
- Database queries
- Log parsing details
- Error stack traces

### WebSocket Connection Testing

Test daemon connectivity:
```bash
curl http://localhost:9999/status
```

Expected response:
```json
{"status": "ok", "version": "1.0.0"}
```

## References

- [DAEMON_INSTALLATION.md](DAEMON_INSTALLATION.md) - Service installation guide
- [DAEMON_API.md](DAEMON_API.md) - WebSocket API reference
- [DEVELOPMENT.md](DEVELOPMENT.md) - Developer guide
- [MIGRATION_TO_SERVICE_ARCHITECTURE.md](MIGRATION_TO_SERVICE_ARCHITECTURE.md) - User migration guide
