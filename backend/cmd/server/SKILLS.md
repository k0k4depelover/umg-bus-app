# Skills — cmd/server

## What This Directory Does
Entry point for the application. `main.go` initializes all dependencies and starts both servers (Fiber HTTP on APP_PORT, net/http WebSocket on WS_PORT).

## Key Function: `SetupApp()`
- Creates repositories, location hub, auth handler, GraphQL resolver
- Registers ALL routes with their middleware and rate limiters
- Returns `*fiber.App` — also used by tests to create an isolated app instance

## Key Function: `main()`
- Loads config → runs migrations → connects PG + Redis → creates JWT service
- Calls `SetupApp()` for the Fiber app
- Starts WS server in a goroutine, then blocks on Fiber `.Listen()`

## Route Registration Pattern
Routes are grouped by middleware:
- `pub` group: PublicLimiter → `/health-check`, `/campus/*`
- `authGroup`: AuthLimiter → `/auth/login`, `/auth/refresh`, `/auth/logout`
- Protected: `auth.Middleware(jwtSvc)` + ProtectedLimiter → `/graphql`
- MCP: MCPLimiter → `/mcp`
- WebSocket (separate mux): `/ws/pilot`, `/ws/student`

## When Adding New Routes
1. Decide if public, authenticated, or role-specific
2. Add to the appropriate group in `SetupApp()`
3. Apply matching rate limiter from `internal/middleware/`
4. If WebSocket: add to the `wsMux` in `main()`, NOT to Fiber

## Testing
`main_test.go` uses `SetupApp()` with testcontainer-backed DB/Redis. It tests route existence, auth enforcement, and rate limiting at the integration level.

## Gotcha: Dual Hub Instances
The location Hub is created twice — once in `SetupApp()` (for GraphQL resolvers) and once in `main()` (for WS handlers). Both share the same Redis/PG connections but are separate in-memory instances.
