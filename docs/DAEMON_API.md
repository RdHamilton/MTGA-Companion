# MTGA-Companion Daemon WebSocket API

This document describes the WebSocket API provided by the MTGA-Companion daemon for real-time communication between the daemon and client applications (GUI, custom tools, etc.).

## Overview

The daemon provides a WebSocket server for clients to:
- Receive real-time events about match updates, statistics changes, draft progress
- Subscribe to specific event types
- Build custom integrations and tools

## Connection

### WebSocket Endpoint

```
ws://localhost:9999
```

**Protocol**: WebSocket (ws://)
**Host**: localhost (127.0.0.1)
**Port**: 9999 (default, configurable)

### Connection Example

**JavaScript/TypeScript**:
```typescript
const ws = new WebSocket('ws://localhost:9999');

ws.onopen = () => {
    console.log('Connected to daemon');
};

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log('Received event:', message);
};

ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};

ws.onclose = () => {
    console.log('Disconnected from daemon');
};
```

**Go**:
```go
import (
    "github.com/gorilla/websocket"
)

conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:9999", nil)
if err != nil {
    log.Fatal("Connection failed:", err)
}
defer conn.Close()

for {
    _, message, err := conn.ReadMessage()
    if err != nil {
        log.Println("Read error:", err)
        break
    }
    log.Printf("Received: %s", message)
}
```

**Python**:
```python
import websocket
import json

def on_message(ws, message):
    data = json.loads(message)
    print(f"Received event: {data}")

def on_open(ws):
    print("Connected to daemon")

ws = websocket.WebSocketApp(
    "ws://localhost:9999",
    on_message=on_message,
    on_open=on_open
)

ws.run_forever()
```

## Event Format

All events follow this structure:

```json
{
  "type": "event:category",
  "data": { ... },
  "timestamp": "2025-11-15T10:30:00Z"
}
```

**Fields**:
- `type` (string) - Event type identifier (e.g., "stats:updated", "match:new")
- `data` (object) - Event-specific payload
- `timestamp` (string) - ISO 8601 timestamp when event was emitted

## Event Types

### Statistics Events

#### `stats:updated`

Emitted when overall statistics are recalculated.

**When triggered**:
- After a match completes
- After database updates
- Periodically (e.g., every 5 minutes)

**Payload**:
```json
{
  "type": "stats:updated",
  "data": {
    "totalMatches": 150,
    "totalGames": 300,
    "matchWinRate": 0.63,
    "gameWinRate": 0.58,
    "currentStreak": {
      "type": "win",
      "length": 5
    }
  },
  "timestamp": "2025-11-15T10:30:00Z"
}
```

**Data fields**:
- `totalMatches` (integer) - Total matches recorded
- `totalGames` (integer) - Total games recorded
- `matchWinRate` (float) - Win rate for matches (0.0-1.0)
- `gameWinRate` (float) - Win rate for games (0.0-1.0)
- `currentStreak` (object) - Current win/loss streak
  - `type` (string) - "win" or "loss"
  - `length` (integer) - Streak length

---

### Match Events

#### `match:new`

Emitted when a new match is recorded.

**When triggered**:
- Match completion detected in log file
- Match data successfully stored in database

**Payload**:
```json
{
  "type": "match:new",
  "data": {
    "matchID": "abc-123-def-456",
    "result": "Win",
    "format": "ConstructedRanked",
    "eventName": "Ranked Traditional Standard",
    "startTime": "2025-11-15T10:20:00Z",
    "endTime": "2025-11-15T10:28:00Z",
    "duration": 480,
    "games": [
      {
        "gameNumber": 1,
        "result": "Win",
        "onPlay": true
      },
      {
        "gameNumber": 2,
        "result": "Win",
        "onPlay": false
      }
    ]
  },
  "timestamp": "2025-11-15T10:30:00Z"
}
```

**Data fields**:
- `matchID` (string) - Unique match identifier
- `result` (string) - "Win", "Loss", or "Draw"
- `format` (string) - Match format (e.g., "ConstructedRanked", "Draft", "Sealed")
- `eventName` (string) - MTGA event name
- `startTime` (string) - ISO 8601 timestamp
- `endTime` (string) - ISO 8601 timestamp
- `duration` (integer) - Match duration in seconds
- `games` (array) - Array of game objects

#### `match:updated`

Emitted when match data is modified.

**When triggered**:
- Match result corrected
- Additional game data added

**Payload**: Same as `match:new`

---

### Draft Events

#### `draft:started`

Emitted when a draft session begins.

**When triggered**:
- Player joins a draft event
- First pack is opened

**Payload**:
```json
{
  "type": "draft:started",
  "data": {
    "draftID": "draft-789-xyz",
    "eventID": "PremierDraft_ONE_20251115",
    "setCode": "ONE",
    "setName": "Phyrexia: All Will Be One",
    "startTime": "2025-11-15T10:30:00Z"
  },
  "timestamp": "2025-11-15T10:30:00Z"
}
```

**Data fields**:
- `draftID` (string) - Unique draft session ID
- `eventID` (string) - MTGA event ID
- `setCode` (string) - 3-letter set code (e.g., "ONE", "MOM")
- `setName` (string) - Full set name
- `startTime` (string) - ISO 8601 timestamp

#### `draft:pick`

Emitted each time a card is picked in the draft.

**When triggered**:
- Player selects a card from a pack

**Payload**:
```json
{
  "type": "draft:pick",
  "data": {
    "draftID": "draft-789-xyz",
    "pack": 1,
    "pick": 3,
    "cardID": 89765,
    "cardName": "Elesh Norn, Mother of Machines",
    "packCards": [89765, 89123, 89456, ...],
    "timestamp": "2025-11-15T10:32:00Z"
  },
  "timestamp": "2025-11-15T10:32:00Z"
}
```

**Data fields**:
- `draftID` (string) - Draft session ID
- `pack` (integer) - Pack number (1-3)
- `pick` (integer) - Pick number within pack (1-14)
- `cardID` (integer) - Arena card ID of picked card
- `cardName` (string) - Card name (if metadata available)
- `packCards` (array) - Array of card IDs in the pack
- `timestamp` (string) - When pick was made

#### `draft:completed`

Emitted when draft is finished.

**When triggered**:
- All 3 packs drafted (45 cards picked)
- Draft session ends

**Payload**:
```json
{
  "type": "draft:completed",
  "data": {
    "draftID": "draft-789-xyz",
    "setCode": "ONE",
    "totalPicks": 45,
    "completedAt": "2025-11-15T10:45:00Z"
  },
  "timestamp": "2025-11-15T10:45:00Z"
}
```

---

### Connection Events

#### `connection:status`

Emitted when daemon status changes.

**When triggered**:
- Client connects
- Daemon startup complete
- Daemon shutting down

**Payload**:
```json
{
  "type": "connection:status",
  "data": {
    "status": "connected",
    "version": "1.0.0",
    "connectedClients": 2,
    "uptime": 3600
  },
  "timestamp": "2025-11-15T10:30:00Z"
}
```

**Data fields**:
- `status` (string) - "connected", "ready", "shutting_down"
- `version` (string) - Daemon version
- `connectedClients` (integer) - Number of active WebSocket clients
- `uptime` (integer) - Daemon uptime in seconds

#### `ping`

Keepalive event sent periodically.

**When triggered**:
- Every 30 seconds (keepalive)

**Payload**:
```json
{
  "type": "ping",
  "data": {},
  "timestamp": "2025-11-15T10:30:00Z"
}
```

**Expected client response**: None required (optional pong)

---

## HTTP Endpoints

### Status Check

**Endpoint**: `GET http://localhost:9999/status`

**Response**:
```json
{
  "status": "running",
  "clients": 2,
  "time": "2025-11-17T09:30:00-05:00"
}
```

**Usage**:
```bash
curl http://localhost:9999/status
```

**Purpose**: Quick check to verify daemon is running and responsive

---

### Health Check

**Endpoint**: `GET http://localhost:9999/health`

**Description**: Comprehensive health status endpoint for monitoring daemon health, including database connectivity, log monitoring, WebSocket status, and performance metrics.

**HTTP Status Codes**:
- `200 OK` - Service is healthy or degraded (but operational)
- `503 Service Unavailable` - Service not fully initialized or unhealthy

**Response (Healthy)**:
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "uptime": 3600.5,
  "database": {
    "status": "ok",
    "lastWrite": "2025-11-17T09:29:45Z"
  },
  "logMonitor": {
    "status": "ok",
    "lastRead": "2025-11-17T09:29:50Z"
  },
  "websocket": {
    "status": "ok",
    "connectedClients": 2
  },
  "metrics": {
    "totalProcessed": 12543,
    "totalErrors": 0
  }
}
```

**Response (Degraded)**:
```json
{
  "status": "degraded",
  "version": "1.0.0",
  "uptime": 7200.3,
  "database": {
    "status": "ok",
    "lastWrite": "2025-11-17T09:15:30Z"
  },
  "logMonitor": {
    "status": "warning",
    "lastRead": "2025-11-17T09:20:15Z"
  },
  "websocket": {
    "status": "ok",
    "connectedClients": 1
  },
  "metrics": {
    "totalProcessed": 8500,
    "totalErrors": 950
  }
}
```

**Response (Unavailable)**:
```json
{
  "status": "unavailable",
  "message": "Service not fully initialized"
}
```

**Response Fields**:
- `status` (string) - Overall health status: `"healthy"`, `"degraded"`, or `"unavailable"`
- `version` (string) - Daemon version
- `uptime` (float) - Daemon uptime in seconds
- `database` (object) - Database health status
  - `status` (string) - `"ok"` or error state
  - `lastWrite` (string, optional) - ISO 8601 timestamp of last successful database write
- `logMonitor` (object) - Log monitoring health status
  - `status` (string) - `"ok"` or `"warning"`
  - `lastRead` (string, optional) - ISO 8601 timestamp of last successful log read
- `websocket` (object) - WebSocket server health status
  - `status` (string) - `"ok"` or error state
  - `connectedClients` (integer) - Number of active WebSocket connections
- `metrics` (object) - Performance metrics
  - `totalProcessed` (integer) - Total log entries processed
  - `totalErrors` (integer) - Total errors encountered

**Health Status Determination**:
- `healthy` - All components operational, error rate <10%, log reads within last 5 minutes
- `degraded` - Service operational but with issues:
  - No log reads for >5 minutes (log monitor status: `"warning"`)
  - Error rate >10% of processed entries
  - Still returns HTTP 200 to indicate service is available
- `unavailable` - Service not initialized (returns HTTP 503)

**Usage Examples**:

Basic health check:
```bash
curl http://localhost:9999/health
```

Check HTTP status code:
```bash
curl -w "%{http_code}" -o /dev/null -s http://localhost:9999/health
```

Parse with jq:
```bash
curl -s http://localhost:9999/health | jq '.status'
```

Monitor in shell script:
```bash
#!/bin/bash
response=$(curl -s http://localhost:9999/health)
status=$(echo $response | jq -r '.status')

if [ "$status" = "healthy" ]; then
    echo "✅ Daemon is healthy"
    exit 0
elif [ "$status" = "degraded" ]; then
    echo "⚠️  Daemon is degraded"
    exit 1
else
    echo "❌ Daemon is unavailable"
    exit 2
fi
```

Kubernetes liveness probe:
```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 9999
  initialDelaySeconds: 10
  periodSeconds: 30
  failureThreshold: 3
```

Docker health check:
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:9999/health || exit 1
```

**Purpose**:
- Monitor daemon health in production
- Integrate with monitoring systems (Prometheus, Datadog, etc.)
- Container orchestration health checks (Docker, Kubernetes)
- Automated alerting on degraded performance

---

## Connection Management

### Automatic Reconnection

Clients should implement automatic reconnection with exponential backoff:

```typescript
class DaemonClient {
    private reconnectDelay = 1000; // Start with 1 second
    private maxReconnectDelay = 30000; // Max 30 seconds

    connect() {
        this.ws = new WebSocket('ws://localhost:9999');

        this.ws.onclose = () => {
            console.log('Disconnected, reconnecting...');
            setTimeout(() => {
                this.connect();
                this.reconnectDelay = Math.min(
                    this.reconnectDelay * 2,
                    this.maxReconnectDelay
                );
            }, this.reconnectDelay);
        };

        this.ws.onopen = () => {
            this.reconnectDelay = 1000; // Reset delay
        };
    }
}
```

### Graceful Disconnection

```typescript
// Close connection cleanly
ws.close(1000, 'Client shutting down');
```

**Close codes**:
- `1000` - Normal closure
- `1001` - Going away (e.g., browser closing)
- `1006` - Abnormal closure (no close frame)

---

## Error Handling

### Connection Errors

**Daemon not running**:
```
Error: connect ECONNREFUSED 127.0.0.1:9999
```

**Solution**: Start daemon with `./mtga-companion daemon`

**Wrong port**:
```
Error: connect ECONNREFUSED 127.0.0.1:8888
```

**Solution**: Ensure daemon port matches client configuration

### Message Parsing Errors

**Invalid JSON**:
```typescript
ws.onmessage = (event) => {
    try {
        const message = JSON.parse(event.data);
        handleEvent(message);
    } catch (error) {
        console.error('Failed to parse message:', error);
    }
};
```

---

## Example Implementations

### Simple Event Logger

```typescript
const ws = new WebSocket('ws://localhost:9999');

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    console.log(`[${message.timestamp}] ${message.type}:`, message.data);
};
```

### Match Counter

```typescript
let matchCount = 0;

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);

    if (message.type === 'match:new') {
        matchCount++;
        console.log(`Total matches: ${matchCount}`);
        console.log(`Result: ${message.data.result}`);
    }
};
```

### Draft Tracker

```typescript
const draftPicks = [];

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);

    if (message.type === 'draft:pick') {
        draftPicks.push({
            pack: message.data.pack,
            pick: message.data.pick,
            cardID: message.data.cardID
        });

        console.log(`P${message.data.pack}P${message.data.pick}: ${message.data.cardName}`);
    }

    if (message.type === 'draft:completed') {
        console.log(`Draft complete! ${draftPicks.length} cards picked`);
    }
};
```

### Discord Bot Integration

```javascript
const Discord = require('discord.js');
const WebSocket = require('ws');

const bot = new Discord.Client();
const ws = new WebSocket('ws://localhost:9999');

ws.on('message', (data) => {
    const event = JSON.parse(data);

    if (event.type === 'match:new') {
        const channel = bot.channels.cache.get('CHANNEL_ID');
        channel.send(
            `🎮 Match ${event.data.result}! ` +
            `Format: ${event.data.format}`
        );
    }
});

bot.login('DISCORD_TOKEN');
```

### OBS Overlay

```html
<!DOCTYPE html>
<html>
<head>
    <title>MTGA Stats Overlay</title>
    <style>
        body { font-family: Arial; color: white; background: transparent; }
        #stats { font-size: 24px; text-shadow: 2px 2px 4px black; }
    </style>
</head>
<body>
    <div id="stats">Connecting...</div>

    <script>
        const ws = new WebSocket('ws://localhost:9999');

        ws.onmessage = (event) => {
            const message = JSON.parse(event.data);

            if (message.type === 'stats:updated') {
                const winRate = (message.data.matchWinRate * 100).toFixed(1);
                document.getElementById('stats').textContent =
                    `Matches: ${message.data.totalMatches} | Win Rate: ${winRate}%`;
            }
        };

        ws.onopen = () => {
            document.getElementById('stats').textContent = 'Connected';
        };
    </script>
</body>
</html>
```

---

## Security Considerations

### Local-Only Access

The daemon WebSocket server binds to `localhost` (127.0.0.1) only:
- Not accessible from network
- Only same-machine connections allowed
- No authentication required (local trust model)

### Future: Network Access

If daemon needs to be accessible over network:
- **Use TLS** (wss://) for encryption
- **Implement authentication** (API keys, tokens)
- **Configure CORS** appropriately
- **Add rate limiting** to prevent abuse

---

## Advanced Usage

### Custom Client Libraries

**Go Client** (`internal/ipc/client.go`):
```go
type Client struct {
    conn *websocket.Conn
    handlers map[string]func(map[string]interface{})
}

func (c *Client) On(eventType string, handler func(map[string]interface{})) {
    c.handlers[eventType] = handler
}

func (c *Client) Listen() {
    for {
        var event Event
        err := c.conn.ReadJSON(&event)
        if err != nil {
            log.Println("Read error:", err)
            break
        }

        if handler, ok := c.handlers[event.Type]; ok {
            handler(event.Data)
        }
    }
}
```

### Event Filtering

```typescript
const eventFilter = {
    allowedEvents: ['match:new', 'stats:updated'],

    handleEvent(event) {
        if (this.allowedEvents.includes(event.type)) {
            console.log('Processing event:', event.type);
            // Handle event
        } else {
            console.log('Ignoring event:', event.type);
        }
    }
};

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    eventFilter.handleEvent(message);
};
```

### Batch Processing

```typescript
const eventQueue = [];
const batchSize = 10;

ws.onmessage = (event) => {
    const message = JSON.parse(event.data);
    eventQueue.push(message);

    if (eventQueue.length >= batchSize) {
        processBatch(eventQueue.splice(0, batchSize));
    }
};

function processBatch(events) {
    console.log(`Processing ${events.length} events`);
    // Batch processing logic
}
```

---

## Debugging

### Enable Debug Logging

**Daemon**:
```bash
./mtga-companion daemon --debug-mode
```

**Client**:
```typescript
ws.onmessage = (event) => {
    console.log('[WS] Raw message:', event.data);
    const message = JSON.parse(event.data);
    console.log('[WS] Parsed event:', message);
};
```

### Monitor WebSocket Traffic

**Chrome DevTools**:
1. Open DevTools (F12)
2. Network tab
3. Filter: WS
4. Click WebSocket connection
5. Messages tab shows all traffic

**wireshark/tcpdump**:
```bash
tcpdump -i lo0 -A 'port 9999'
```

---

## Troubleshooting

### "Connection refused"

**Cause**: Daemon not running
**Solution**: Start daemon with `./mtga-companion daemon`

### "Connection timeout"

**Cause**: Firewall blocking port 9999
**Solution**: Allow port 9999 in firewall settings

### "Messages not received"

**Cause**: Event handler not registered
**Solution**: Ensure `ws.onmessage` is set before connection opens

### "Unexpected disconnections"

**Cause**: Daemon crash or restart
**Solution**: Implement auto-reconnect with exponential backoff

---

## Version History

### v1.0.0 (Current)

Initial WebSocket API release.

**Events**:
- `stats:updated`
- `match:new`
- `match:updated`
- `draft:started`
- `draft:pick`
- `draft:completed`
- `connection:status`
- `ping`

---

## References

- [ARCHITECTURE.md](ARCHITECTURE.md) - System architecture
- [DEVELOPMENT.md](DEVELOPMENT.md) - Development guide
- [WebSocket Protocol RFC](https://tools.ietf.org/html/rfc6455)
- [gorilla/websocket Documentation](https://pkg.go.dev/github.com/gorilla/websocket)
