# Skills — internal/location

## What This Package Does
Manages real-time bus location tracking via WebSocket + Redis Pub/Sub. Pilots push GPS, students subscribe to campus channels.

## Files
- **hub.go** — Core `Hub` struct. Publishes locations to Redis, subscribes students to Pub/Sub, batches writes to `location_log`.
- **handler.go** — `HandlePilot()`: WebSocket handler that reads `PilotPing` from pilots and calls `hub.PublishLocation()`.
- **student_handler.go** — `HandleStudent()`: WebSocket handler that sends initial position then subscribes to live updates.

## Key Types
```go
PilotPing    { Lat, Lng, Bearing, Speed float64 }                    // Incoming from pilot
LiveLocation { PilotoID, CampusID string; Lat, Lng, Bearing, Speed float64; UpdatedAt time.Time }  // Broadcast to students
LocationLogEntry { PilotID, CampusID string; Lat, Lng, Bearing, Speed float64 }  // Queued for DB batch
```

## Redis Data Model
- **Position Hash**: `campus:live:{campusID}:{pilotID}` → hash fields (lat, lng, bearing, speed, updated_at). TTL: 30 seconds.
- **Pub/Sub Channel**: `campus:channel:{campusID}` → JSON-encoded LiveLocation messages.

## PublishLocation() Flow
1. Redis pipeline: `HSet` position hash + `Expire` 30s (atomic)
2. Marshal LiveLocation to JSON
3. `Publish` to campus channel
4. Send `LocationLogEntry` to `logCh` (non-blocking, drops if channel full at 500)

## Batch Logging (runLogWriter)
- Background goroutine reads from `logCh` (buffered channel, capacity 500)
- Flushes to `location_log` table when batch reaches 100 entries OR every 10 seconds
- Uses `pgx.CopyFrom()` for efficient bulk insert

## GetLiveLocation()
- Scans Redis keys matching `campus:live:{campusID}:*`
- Returns the first match (single pilot per campus assumption)
- Called when student first connects to provide initial state

## Subscribe()
- Creates Redis Pub/Sub subscription to `campus:channel:{campusID}`
- Select loop: forwards messages to WebSocket until context cancels
- Each student has independent subscription

## When Adding Features
- To add new data to pilot pings: update `PilotPing` + `LiveLocation` + Redis hash fields in `PublishLocation()`
- To add a new subscriber type: create a new handler similar to `student_handler.go`, reuse `hub.Subscribe()`
- To change batch behavior: modify `runLogWriter()` constants (100 batch size, 10s ticker)
- Warning: `GetLiveLocation()` uses `KEYS` command — fine for prototype, replace with `SCAN` for production scale

## WebSocket Server Note
These handlers run on the stdlib `net/http` server (port 8085), NOT on Fiber. They use `coder/websocket` library. Auth is via `?token=` query param, not Authorization header.
