# UMG Bus Tracker — Developer Documentation

## What Is This?

A campus bus monitoring backend API. Bus **pilots** (drivers) install a mobile app that streams their GPS position via WebSocket. **Students** connect via WebSocket to receive real-time bus locations and decide whether to walk or wait.

---

## Architecture Overview

```
┌─────────────┐         ┌──────────────────────────────────────────────┐
│  Pilot App  │─── WS ──│  WebSocket Server (net/http :8085)           │
│  (GPS)      │         │    /ws/pilot   → PublishLocation()           │
└─────────────┘         │    /ws/student → Subscribe()                 │
                        └──────────┬───────────────────────────────────┘
                                   │
┌─────────────┐         ┌──────────▼───────────────────────────────────┐
│ Student App │─── WS ──│  Redis                                       │
│             │         │    Hash:   campus:live:{campusID}:{pilotID}  │
└─────────────┘         │    PubSub: campus:channel:{campusID}         │
                        │    Auth:   refresh:{sha256(token)}           │
                        └──────────────────────────────────────────────┘
                                   │
┌─────────────┐         ┌──────────▼───────────────────────────────────┐
│ Any Client  │── HTTP ──│  Fiber HTTP Server (:8084)                   │
│             │         │    REST:    /auth/*, /campus/*, /health-check│
│             │         │    GraphQL: /graphql, /playground            │
│             │         │    MCP:     /mcp                             │
└─────────────┘         └──────────┬───────────────────────────────────┘
                                   │
                        ┌──────────▼───────────────────────────────────┐
                        │  PostgreSQL                                   │
                        │    Tables: campus, pilots, students,         │
                        │            sessions, student_campus_changes, │
                        │            location_log                      │
                        └──────────────────────────────────────────────┘
```

### Dual Server Design

The app runs **two servers concurrently**:

| Server | Port | Framework | Purpose |
|--------|------|-----------|---------|
| HTTP | `APP_PORT` (8084) | Fiber v2 (fasthttp) | REST, GraphQL, MCP |
| WebSocket | `WS_PORT` (8085) | net/http stdlib | Pilot & Student WebSocket streams |

**Why two?** Fiber uses `fasthttp` under the hood, which doesn't support standard `net/http` WebSocket hijacking. The WS server uses `coder/websocket` on stdlib `net/http`.

### Database Split

| Store | Purpose | TTL |
|-------|---------|-----|
| **PostgreSQL** | Users, campuses, sessions, location history (audit) | Permanent |
| **Redis Hash** | Current pilot positions (`campus:live:{cid}:{pid}`) | 30 seconds |
| **Redis Pub/Sub** | Broadcast location updates (`campus:channel:{cid}`) | Ephemeral |
| **Redis KV** | Refresh token sessions (`refresh:{sha256}`) | 7 days |

---

## Quick Start

```bash
# 1. Start infrastructure
docker compose up -d   # PostgreSQL :5436, Redis :6381

# 2. Configure environment
cp backend/.env.example backend/.env
# Edit .env if needed (defaults work for local dev)

# 3. Run the server
cd backend
go mod tidy
go run ./cmd/server

# Server outputs:
#   "Migrations aplicadas"
#   "WebSocket server en puerto 8085"
#   "Servidor en el puerto 8084"
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_HOST` | `localhost` | PostgreSQL host |
| `POSTGRES_PORT` | `5436` | PostgreSQL port |
| `POSTGRES_USER` | `admin` | PostgreSQL user |
| `POSTGRES_PASSWORD` | `postgresumgtracker` | PostgreSQL password |
| `POSTGRES_DB` | `transportation-tracker-db` | Database name |
| `REDIS_ADDR` | `localhost:6381` | Redis address |
| `REDIS_PASS` | *(empty)* | Redis password |
| `APP_PORT` | `8084` | Fiber HTTP port |
| `WS_PORT` | `8085` | WebSocket port |
| `JWT_SECRET` | `dev-secret-cambia-en-produccion` | JWT signing secret |

### Running Tests

```bash
cd backend
go test ./...
```

Tests use **testcontainers** — they spin up real PostgreSQL and Redis containers automatically. Docker must be running. 56 tests across 6 files.

### Building for Production

```bash
cd backend
go build -o bin/server ./cmd/server

# Or via Docker:
docker build -t umg-bus-app ./backend
docker run -p 8084:8084 umg-bus-app
```

---

## Server Initialization Flow

```
main()
  ├─ config.Load()                    → Read env vars
  ├─ db.RunMigrations(dsn)            → Apply pending SQL migrations
  ├─ db.NewPostgres(dsn)              → Connect pgxpool (connection pool)
  ├─ db.NewRedis(addr, pass)          → Connect Redis client
  ├─ auth.NewJWTService(secret)       → Create JWT signer/verifier
  ├─ SetupApp(pg, rdb, jwt)           → Build Fiber app with all routes
  │    ├─ Create repositories (Campus, Pilot, Student)
  │    ├─ Create location Hub
  │    ├─ Register routes with rate limiters
  │    └─ Return *fiber.App
  ├─ Start WebSocket server (goroutine, :8085)
  └─ Start Fiber HTTP server (blocking, :8084)
```

---

## API Reference

### Authentication Endpoints

All auth endpoints are rate-limited to **20 req/min per IP**.

#### `POST /auth/login`

Authenticates a pilot or student. Returns JWT access + refresh tokens.

```json
// Request
{
  "username": "testpilot",
  "password": "pilotpass123",
  "role": "pilot"           // "pilot" or "student"
}

// Response 200
{
  "access_token": "eyJhbGci...",
  "refresh_token": "eyJhbGci..."
}
```

| Status | Meaning |
|--------|---------|
| 200 | Success — tokens returned |
| 400 | Bad request — missing fields or invalid role |
| 401 | Unauthorized — wrong credentials or inactive user |

**SQL executed (pilot):**
```sql
SELECT pilot_id, campus_id, password_hash FROM pilots WHERE username = $1 AND active=TRUE
```

#### `POST /auth/refresh`

Generates a new access token from a valid refresh token. The refresh token itself is NOT rotated.

```json
// Request
{ "refresh_token": "eyJhbGci..." }

// Response 200
{ "access_token": "eyJhbGci..." }
```

**Flow:** Hash token → lookup in Redis (`refresh:{hash}`) → extract userID|role → query fresh campusID from DB → generate new access token.

#### `POST /auth/logout`

Revokes the refresh token. Idempotent — calling twice returns 200 both times.

```json
// Request
{ "refresh_token": "eyJhbGci..." }

// Response 200
{ "ok": true }
```

**Actions:** Deletes from Redis + marks `revoked_at` in PostgreSQL `sessions` table.

### JWT Token Details

| Token | Lifetime | Claims | Stored? |
|-------|----------|--------|---------|
| Access | 15 min | `user_id`, `campus_id`, `role` | No (stateless) |
| Refresh | 7 days | `user_id`, `role` (no campus_id) | Redis + PostgreSQL |

- Algorithm: **HS256** (HMAC-SHA256)
- Refresh tokens stored as **SHA-256 hash** — never raw
- Access tokens validated by signature only — no DB/Redis lookup

### Authentication Middleware

Protected routes require: `Authorization: Bearer <access_token>`

- **`auth.Middleware(jwtSvc)`** — Verifies token, stores `*Claims` in `c.Locals("claims")`
- **`auth.RequireRole("pilot")`** — Checks `claims.Role` matches; returns 403 if not

WebSocket endpoints use query param: `?token=<access_token>`

---

### Public REST Endpoints

Rate-limited to **120 req/min per IP**.

#### `GET /health-check`
```json
{ "status": "online!" }
```

#### `GET /campus`
Returns all active campuses.

#### `GET /campus/:id`
Returns a single campus by UUID.

#### `GET /campus/:id/pilot`
Returns the active pilot for a given campus.

---

### GraphQL API

**Endpoint:** `POST /graphql` (requires JWT auth, **200 req/min per user**)
**Playground:** `GET /playground` (no auth, dev tool)

```graphql
type Campus {
  campusID: ID!
  name: String!
  city: String!
  bounds: MapBounds!
  activePilot: PilotStatus
}

type PilotStatus {
  pilotId: ID!
  fullName: String!
  isOnline: Boolean!
  lastSeen: String!
}

type LiveLocation {
  pilotId: ID!
  campusID: ID!
  lat: Float!
  lng: Float!
  bearing: Float!
  speedKmh: Float!
  updatedAt: String!
}

type Query {
  campuses: [Campus!]!
  campus(id: ID!): Campus
  liveLocation(CampusId: ID!): LiveLocation
}

type Mutation {
  changeCampus(studentId: ID!, toCampusId: ID!): Boolean!
}
```

**Example query:**
```graphql
query {
  campuses {
    campusID
    name
    activePilot { fullName isOnline }
  }
}
```

---

### WebSocket Endpoints

Both require JWT via query param: `ws://host:8085/ws/pilot?token=<JWT>`

#### `WS /ws/pilot` (role: pilot)

Pilot sends GPS pings continuously. Server stores in Redis and broadcasts to students.

```json
// Client → Server (repeated)
{
  "lat": 14.5833,
  "lng": -90.5519,
  "bearing": 180.5,
  "speed": 25.3
}
// No server response — fire and forget
```

#### `WS /ws/student` (role: student)

Student connects and receives live location updates. First message is the last known position.

```json
// Server → Client (initial + continuous)
{
  "pilot_id": "uuid",
  "campus_id": "uuid",
  "lat": 14.5833,
  "lng": -90.5519,
  "bearing": 180.5,
  "speed": 25.3,
  "updated_at": "2026-03-25T10:30:45Z"
}
```

**Data flow:**
```
Pilot GPS → WS /ws/pilot → hub.PublishLocation()
  ├─ Redis HSet (campus:live:{cid}:{pid}, TTL 30s)
  ├─ Redis Publish (campus:channel:{cid})
  └─ logCh → batch → PostgreSQL location_log (async, every 100 entries or 10s)
                          ↓
Student ← WS /ws/student ← hub.Subscribe() ← Redis Pub/Sub
```

---

### MCP Endpoint

**`POST /mcp`** — Model Context Protocol tool dispatcher. Rate-limited to **60 req/min per IP**.

```json
// Request
{ "tool": "get_campus_list", "input": {} }

// Response
{
  "content": [{ "type": "text", "text": "[{...campus data...}]" }]
}
```

**Available tools:**

| Tool | Input | Description |
|------|-------|-------------|
| `get_campus_list` | *(none)* | Returns all active campuses |
| `get_live_location` | `{ "campus_id": "uuid" }` | Current bus position for a campus |
| `get_route_history` | `{ "campus_id": "uuid", "limit": 100 }` | Recent location_log entries (max 500) |

---

## Database Schema

### Tables

#### `campus`
| Column | Type | Notes |
|--------|------|-------|
| campus_id | UUID PK | auto-generated |
| name | TEXT | NOT NULL |
| city | TEXT | NOT NULL |
| bound_sw_lat / bound_sw_lng | DOUBLE PRECISION | Southwest map bound |
| bound_ne_lat / bound_ne_lng | DOUBLE PRECISION | Northeast map bound |
| route_geojson | JSONB | Optional GeoJSON LineString |
| active | BOOLEAN | DEFAULT TRUE |
| created_at | TIMESTAMPTZ | DEFAULT now() |

#### `pilots`
| Column | Type | Notes |
|--------|------|-------|
| pilot_id | UUID PK | auto-generated |
| campus_id | UUID FK | → campus |
| username | CITEXT | UNIQUE, 4-50 chars |
| password_hash | TEXT | bcrypt |
| secret_code | UUID | one-time activation code |
| full_name | TEXT | NOT NULL |
| phone | TEXT | optional, regex-validated |
| active | BOOLEAN | DEFAULT TRUE |
| last_seen_at | TIMESTAMPTZ | updated on activity |

#### `students`
| Column | Type | Notes |
|--------|------|-------|
| student_id | UUID PK | auto-generated |
| campus_id | UUID FK | → campus |
| username | CITEXT | UNIQUE, 4-50 chars |
| password_hash | TEXT | bcrypt |
| full_name | TEXT | NOT NULL |
| active | BOOLEAN | DEFAULT TRUE |

#### `sessions`
| Column | Type | Notes |
|--------|------|-------|
| session_id | UUID PK | auto-generated |
| user_id | UUID | pilot_id or student_id |
| user_role | ENUM | 'pilot' or 'student' |
| token_hash | TEXT | SHA-256 of refresh token, UNIQUE |
| expires_at | TIMESTAMPTZ | 7 days from login |
| revoked_at | TIMESTAMPTZ | NULL until logout |
| revoke_reason | TEXT | 'logout', 'password_change', 'admin_revoke' |

#### `student_campus_changes`
Audit log for campus switches.

| Column | Type | Notes |
|--------|------|-------|
| change_id | UUID PK | auto-generated |
| student_id | UUID FK | → students |
| from_campus_id | UUID FK | → campus |
| to_campus_id | UUID FK | → campus, must differ from `from` |
| changed_at | TIMESTAMPTZ | DEFAULT now() |

#### `location_log`
Persistent GPS history (batched async writes from WebSocket).

| Column | Type | Notes |
|--------|------|-------|
| log_id | UUID PK | auto-generated |
| pilot_id | UUID FK | → pilots |
| campus_id | UUID FK | → campus |
| lat / lng | DOUBLE PRECISION | constrained ±90/±180 |
| bearing | DOUBLE PRECISION | DEFAULT 0 |
| speed_kmh | DOUBLE PRECISION | DEFAULT 0 |
| recorded_at | TIMESTAMPTZ | DEFAULT now() |

---

## Rate Limiting

| Limiter | Scope | Limit | Key |
|---------|-------|-------|-----|
| AuthLimiter | `/auth/*` | 20 req/min | IP |
| PublicLimiter | `/campus/*`, `/health-check` | 120 req/min | IP |
| ProtectedLimiter | `/graphql` | 200 req/min | UserID (fallback: IP) |
| MCPLimiter | `/mcp` | 60 req/min | IP |

All return `429` with `{"error": "too many requests, try again later"}`.

---

## Project Structure

```
backend/
├── cmd/server/
│   ├── main.go                    # Entry point, server setup, route registration
│   └── main_test.go               # Integration tests (13 tests)
├── internal/
│   ├── auth/
│   │   ├── jwt.go                 # JWTService: GenerateAccess, GenerateRefresh, Verify
│   │   ├── handler.go             # Login, Refresh, Logout HTTP handlers
│   │   ├── middleware.go          # Auth middleware + RequireRole
│   │   ├── jwt_test.go            # 8 tests
│   │   ├── handler_test.go        # 13 tests
│   │   └── middleware_test.go     # 7 tests
│   ├── config/
│   │   └── config.go              # Env var loader with defaults
│   ├── db/
│   │   ├── postgres.go            # pgxpool connection
│   │   ├── redis.go               # Redis client
│   │   └── migrate.go             # SQL migration runner
│   ├── domain/
│   │   ├── campus.go              # Campus struct
│   │   ├── pilot.go               # Pilot struct
│   │   └── student.go             # Student struct
│   ├── location/
│   │   ├── hub.go                 # Location Hub: publish, subscribe, batch logging
│   │   ├── handler.go             # Pilot WS handler
│   │   └── student_handler.go     # Student WS handler
│   ├── mcp/
│   │   ├── server.go              # MCP tool dispatcher
│   │   └── server_test.go         # 7 tests
│   ├── middleware/
│   │   ├── rate_limiter.go        # 4 rate limiters
│   │   └── rate_limiter_test.go   # 8 tests
│   ├── repository/
│   │   ├── campus_repo.go         # Campus SQL queries
│   │   ├── pilot_repo.go          # Pilot SQL queries
│   │   └── student_repo.go        # Student SQL queries
│   └── testutil/
│       └── containers.go          # Testcontainers setup + seed data
├── graph/
│   ├── schema.graphqls            # GraphQL schema
│   ├── resolver.go                # Resolver struct
│   ├── schema.resolvers.go        # Resolver implementations
│   ├── generated.go               # gqlgen auto-generated
│   └── model/models_gen.go        # gqlgen auto-generated models
├── migrations/
│   ├── 001_initial.up.sql         # Core tables
│   ├── 001_initial.down.sql       # Rollback
│   ├── 002_location_log.up.sql    # Location history table
│   └── 002_location_log.down.sql  # Rollback
├── Dockerfile                     # Multi-stage build (golang:1.22 → alpine:3.19)
├── go.mod                         # Go 1.25, all dependencies
├── gqlgen.yml                     # GraphQL codegen config
└── .env.example                   # Environment template
```

---

## How to Add New Components

### Adding a New REST Endpoint

1. **Create a handler** in the appropriate `internal/` package (or create a new one under `internal/`):
   ```go
   // internal/myfeature/handler.go
   type Handler struct {
       db *pgxpool.Pool
   }

   func NewHandler(db *pgxpool.Pool) *Handler {
       return &Handler{db: db}
   }

   func (h *Handler) GetSomething(c *fiber.Ctx) error {
       // your logic
       return c.JSON(fiber.Map{"data": result})
   }
   ```

2. **Register the route** in `cmd/server/main.go` inside `SetupApp()`:
   ```go
   myHandler := myfeature.NewHandler(pg)
   // Public route
   pub.Get("/my-endpoint", myHandler.GetSomething)
   // Or protected route (requires JWT)
   app.Get("/my-endpoint", auth.Middleware(jwtSvc), myHandler.GetSomething)
   ```

3. **Add rate limiting** if needed — use an existing limiter or create a new one in `internal/middleware/rate_limiter.go`.

### Adding a New Database Table

1. **Create a migration file** — increment the number:
   ```
   backend/migrations/003_my_feature.up.sql
   backend/migrations/003_my_feature.down.sql
   ```

2. **Write the SQL:**
   ```sql
   -- 003_my_feature.up.sql
   CREATE TABLE my_table (
       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
       name TEXT NOT NULL,
       created_at TIMESTAMPTZ NOT NULL DEFAULT now()
   );
   ```

3. **Create a domain model** in `internal/domain/`:
   ```go
   // internal/domain/my_entity.go
   type MyEntity struct {
       ID        uuid.UUID
       Name      string
       CreatedAt time.Time
   }
   ```

4. **Create a repository** in `internal/repository/`:
   ```go
   // internal/repository/my_entity_repo.go
   type MyEntityRepo struct { db *pgxpool.Pool }

   func NewMyEntityRepo(db *pgxpool.Pool) *MyEntityRepo {
       return &MyEntityRepo{db: db}
   }

   func (r *MyEntityRepo) GetByID(ctx context.Context, id string) (*domain.MyEntity, error) {
       row := r.db.QueryRow(ctx, "SELECT id, name, created_at FROM my_table WHERE id = $1", id)
       var e domain.MyEntity
       err := row.Scan(&e.ID, &e.Name, &e.CreatedAt)
       return &e, err
   }
   ```

5. Migrations run automatically on server start via `db.RunMigrations()`.

### Adding a New GraphQL Query or Mutation

1. **Edit the schema** in `graph/schema.graphqls`:
   ```graphql
   type MyType {
     id: ID!
     name: String!
   }

   extend type Query {
     myEntities: [MyType!]!
   }
   ```

2. **Regenerate code:**
   ```bash
   cd backend
   go run github.com/99designs/gqlgen generate
   ```

3. **Implement the resolver** in `graph/schema.resolvers.go` — gqlgen generates stubs with `panic("not implemented")`. Replace with your logic.

4. **Add repository to Resolver struct** in `graph/resolver.go` if needed:
   ```go
   type Resolver struct {
       // ...existing repos...
       MyEntityRepo *repository.MyEntityRepo
   }
   ```

5. **Wire it up** in `cmd/server/main.go` `SetupApp()` where the resolver is created.

### Adding a New MCP Tool

1. **Edit `internal/mcp/server.go`** — add a new case to the `Handle()` switch:
   ```go
   case "my_tool":
       var input struct {
           Param string `json:"param"`
       }
       json.Unmarshal(req.Input, &input)
       // your logic...
       return MCPResponse{Content: []MCPContent{{Type: "text", Text: result}}}
   ```

2. That's it — the tool is immediately available via `POST /mcp`.

### Adding a New WebSocket Handler

1. **Create a handler** in `internal/location/` (or a new package):
   ```go
   func HandleMyWS(hub *Hub, jwtSvc *auth.JWTService) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
           // Extract token from ?token= query param
           // Verify JWT, upgrade to WebSocket
           // Your read/write loop
       }
   }
   ```

2. **Register in `cmd/server/main.go`** where the WS mux is built:
   ```go
   wsMux.HandleFunc("/ws/my-endpoint", location.HandleMyWS(hub, jwtSvc))
   ```

### Writing Tests

- Use `testutil.Setup(t)` to get a `TestEnv` with real PostgreSQL + Redis + seeded data
- Seeded data includes: 1 campus, 1 active pilot (`pilotpass123`), 1 student (`studentpass123`), 1 inactive pilot
- Redis is flushed between tests for isolation
- Run with `go test ./...` (Docker must be running)

---

## Key Dependencies

| Package | Version | Purpose |
|---------|---------|---------|
| `gofiber/fiber/v2` | v2.52.12 | HTTP framework |
| `jackc/pgx/v5` | v5.8.0 | PostgreSQL driver (pgxpool) |
| `redis/go-redis/v9` | v9.18.0 | Redis client |
| `golang-jwt/jwt/v5` | v5.3.1 | JWT token handling |
| `golang-migrate/migrate/v4` | v4.19.1 | Database migrations |
| `coder/websocket` | v1.8.14 | WebSocket (RFC 6455) |
| `99designs/gqlgen` | v0.17.88 | GraphQL code generation |
| `testcontainers/testcontainers-go` | v0.41.0 | Docker-based test infra |
| `golang.org/x/crypto` | v0.49.0 | bcrypt password hashing |
