# Implementation Status

## Rate Limiter Implementation Status

**File:** `backend/internal/middleware/rate_limiter.go`

Uses `github.com/gofiber/fiber/v2/middleware/limiter` (built into Fiber, no extra dependency needed).

Four limiters, all using in-memory storage with 1-minute expiration windows:

| Limiter             | Max req/min | Key strategy                          | Applied to          |
|---------------------|-------------|---------------------------------------|----------------------|
| `AuthLimiter()`     | 20          | IP (`c.IP()`)                         | `/auth/*`            |
| `PublicLimiter()`   | 120         | IP (`c.IP()`)                         | `/`, `/campus*`      |
| `ProtectedLimiter()`| 200         | `claims.UserID`, fallback to `c.IP()` | `/graphql`           |
| `MCPLimiter()`      | 60          | IP (`c.IP()`)                         | `/mcp/*`             |

All return HTTP 429 with `{"error": "too many requests, try again later"}` on limit breach.

Imports `auth.Claims` to extract UserID from Fiber locals in `ProtectedLimiter`.

## Test Suite Implementation Status

No test files exist yet (`*_test.go` count: 0).

### Test files to create

#### 1. `backend/internal/auth/jwt_test.go` (unit tests)

| Function | Validates |
|----------|-----------|
| `TestGenerateAccess_ValidToken` | GenerateAccess returns a non-empty string; Verify parses it back with correct UserID, CampusID, Role |
| `TestGenerateAccess_ExpiresIn15Min` | ExpiresAt claim is ~15 minutes from now |
| `TestGenerateRefresh_ValidToken` | GenerateRefresh returns valid token; Verify returns correct UserID, Role; CampusID is empty |
| `TestGenerateRefresh_ExpiresIn7Days` | ExpiresAt claim is ~7 days from now |
| `TestVerify_InvalidSignature` | Token signed with wrong secret returns error |
| `TestVerify_ExpiredToken` | Manually crafted expired token returns error |
| `TestVerify_MalformedToken` | Random string returns error |
| `TestVerify_WrongSigningMethod` | Token using RS256 (or non-HMAC) is rejected |

#### 2. `backend/internal/auth/middleware_test.go` (unit tests)

| Function | Validates |
|----------|-----------|
| `TestMiddleware_NoHeader` | Returns 401 when Authorization header is missing |
| `TestMiddleware_InvalidPrefix` | Returns 401 when header doesn't start with "Bearer" |
| `TestMiddleware_InvalidToken` | Returns 401 when token fails verification |
| `TestMiddleware_ValidToken` | Calls `c.Next()`; claims are stored in `c.Locals("claims")` |
| `TestRequireRole_CorrectRole` | Calls `c.Next()` when role matches |
| `TestRequireRole_WrongRole` | Returns 403 when role doesn't match |
| `TestRequireRole_NoClaims` | Returns 403 when no claims in locals |

#### 3. `backend/internal/auth/handler_test.go` (integration tests)

Requires: PostgreSQL (testcontainers), Redis (testcontainers), seeded user data with bcrypt-hashed passwords.

| Function | Validates |
|----------|-----------|
| `TestLogin_Success_Pilot` | Valid pilot credentials return access_token + refresh_token; refresh token stored in Redis |
| `TestLogin_Success_Student` | Valid student credentials return tokens |
| `TestLogin_InvalidRole` | Role other than "pilot"/"student" returns 400 |
| `TestLogin_WrongPassword` | Returns 401 |
| `TestLogin_NonexistentUser` | Returns 401 |
| `TestLogin_InvalidBody` | Malformed JSON returns 400 |
| `TestRefresh_ValidToken` | Returns new access_token |
| `TestRefresh_ExpiredOrRevoked` | Returns 401 when key missing from Redis |
| `TestRefresh_InvalidBody` | Returns 400 |
| `TestLogout_Success` | Returns `{"ok": true}`; Redis key deleted |
| `TestLogout_InvalidBody` | Returns 400 |

#### 4. `backend/internal/middleware/rate_limiter_test.go` (unit tests)

| Function | Validates |
|----------|-----------|
| `TestAuthLimiter_AllowsUnderLimit` | 20 requests succeed |
| `TestAuthLimiter_BlocksOverLimit` | 21st request returns 429 |
| `TestPublicLimiter_AllowsUnderLimit` | 120 requests succeed |
| `TestPublicLimiter_BlocksOverLimit` | 121st request returns 429 |
| `TestProtectedLimiter_KeysByUserID` | Two users with same IP get independent limits |
| `TestProtectedLimiter_FallbackToIP` | No claims in locals falls back to IP-based key |
| `TestMCPLimiter_AllowsUnderLimit` | 60 requests succeed |
| `TestMCPLimiter_BlocksOverLimit` | 61st request returns 429 |

#### 5. `backend/internal/mcp/server_test.go` (integration tests)

Requires: PostgreSQL (testcontainers), Redis (testcontainers), seeded campus data, active pilot location in Redis.

| Function | Validates |
|----------|-----------|
| `TestHandle_GetCampusList` | Returns campus list as MCP response |
| `TestHandle_GetLiveLocation_ActivePilot` | Returns location JSON in MCP content |
| `TestHandle_GetLiveLocation_NoPilot` | Returns "No hay piloto activo" message |
| `TestHandle_GetRouteHistory` | Returns points from location_log table |
| `TestHandle_GetRouteHistory_DefaultLimit` | Limit 0 defaults to 100 |
| `TestHandle_UnknownTool` | Returns 400 |
| `TestHandle_InvalidBody` | Returns 400 |

#### 6. `backend/cmd/server/main_test.go` (integration tests)

Requires: PostgreSQL (testcontainers), Redis (testcontainers), migrations run.

| Function | Validates |
|----------|-----------|
| `TestSetupApp_HealthCheck` | GET /health-check returns 200 with `{"status":"online!"}` |
| `TestSetupApp_AuthRoutes_Registered` | POST /auth/login, /auth/refresh, /auth/logout return non-404 |
| `TestSetupApp_GraphQL_RequiresAuth` | POST /graphql without token returns 401 |
| `TestSetupApp_MCP_Registered` | POST /mcp/ returns non-404 |
| `TestSetupApp_RateLimiting_Auth` | 21 rapid POST /auth/login returns 429 |

#### 7. `backend/internal/location/hub_test.go` (integration tests)

Requires: Redis (testcontainers), PostgreSQL (testcontainers) with location_log table.

| Function | Validates |
|----------|-----------|
| `TestPublishLocation_StoresInRedis` | After publish, HGetAll returns correct fields |
| `TestPublishLocation_SetsExpiry` | Key TTL is ~30 seconds |
| `TestPublishLocation_PublishesToChannel` | Subscriber on `campus:channel:{id}` receives message |
| `TestGetLiveLocation_Found` | Returns LiveLocation with correct fields |
| `TestGetLiveLocation_NoPilot` | Returns error when no keys match |
| `TestFlushBatch_InsertsToDb` | After flush, location_log table has correct rows |

## Continuation Guide

### Immediate next steps

1. **Add testcontainers dependency:**
   ```bash
   cd backend
   go get github.com/testcontainers/testcontainers-go
   go get github.com/testcontainers/testcontainers-go/modules/postgres
   go get github.com/testcontainers/testcontainers-go/modules/redis
   go mod tidy
   ```

2. **Create a shared test helper** at `backend/internal/testutil/containers.go` with functions to spin up PostgreSQL and Redis containers, run migrations, and return connected clients.

3. **Start with unit tests** (no external dependencies):
   - `auth/jwt_test.go` — pure logic, no DB/Redis
   - `auth/middleware_test.go` — uses `fiber.Test()` with mock JWTService or real JWTService with known secret
   - `middleware/rate_limiter_test.go` — uses `fiber.Test()` with in-memory Fiber app

4. **Then integration tests** (require testcontainers + Docker):
   - `auth/handler_test.go`
   - `mcp/server_test.go`
   - `location/hub_test.go`
   - `cmd/server/main_test.go`

### Current blockers

- **No test dependencies in go.mod** — `testcontainers-go` not yet added.
- **Docker required** for integration tests (testcontainers spins up real PG/Redis).
- **No test seed data** — need helper to insert pilots/students with bcrypt-hashed passwords.
- **`godotenv` listed in CLAUDE.md** but not in `go.mod` — `config.Load()` may use it internally; verify `backend/internal/config/` before running.

### Verification commands

```bash
cd backend

# After writing unit tests
go test ./internal/auth/ -run "TestGenerate|TestVerify" -v
go test ./internal/middleware/ -v
go test ./internal/auth/ -run "TestMiddleware|TestRequireRole" -v

# After writing integration tests (requires Docker)
go test ./internal/auth/ -run "TestLogin|TestRefresh|TestLogout" -v
go test ./internal/mcp/ -v
go test ./internal/location/ -v
go test ./cmd/server/ -v

# Full suite
go test ./... -v
```

## Architecture Decisions

### Why SetupApp was extracted

`SetupApp` accepts pre-built `*pgxpool.Pool`, `*redis.Client`, and `*auth.JWTService` as parameters and returns a `*fiber.App`. This lets integration tests call `SetupApp` with testcontainer-backed connections and use `app.Test()` to exercise the full HTTP routing stack (middleware, rate limiters, handlers) without starting a real TCP listener. `main()` remains thin: load config, open connections, call `SetupApp`, listen.

### Why rate limiter uses in-memory storage

Fiber's built-in limiter stores counters in process memory by default. This is acceptable because: (1) the backend runs as a single instance (no horizontal scaling yet), (2) it avoids adding Redis round-trips to every request, and (3) counters reset on restart which is a non-issue for a prototype. If the app scales to multiple instances, switch to `limiter.Config.Storage` backed by Redis.

### Why ProtectedLimiter keys by UserID with IP fallback

Keying by UserID means authenticated users get their own 200 req/min budget regardless of shared IP (e.g., campus NAT). The IP fallback handles the edge case where `c.Locals("claims")` is nil — this shouldn't happen because `auth.Middleware` runs first on `/graphql`, but defensive coding prevents a nil-pointer panic if middleware ordering changes.

### Why MCP endpoint path changed from /mcp to /mcp/ (group prefix)

Fiber's `app.Group("/mcp")` sets `/mcp` as the prefix. The handler is registered as `mcpGroup.Post("/")`, making the full path `POST /mcp/`. This is standard Fiber group behavior — the group owns the prefix and sub-routes are relative to it. A `POST /mcp` (no trailing slash) would 404; clients must use `/mcp/`. This is a consequence of using groups for middleware scoping (the `MCPLimiter` applies to all routes in the group).
