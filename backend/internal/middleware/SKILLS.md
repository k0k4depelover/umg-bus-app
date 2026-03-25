# Skills — internal/middleware

## What This Package Does
Rate limiters for different route groups. Uses Fiber's built-in `limiter` middleware with in-memory storage.

## Available Limiters

| Function | Limit | Key | Used On |
|----------|-------|-----|---------|
| `AuthLimiter()` | 20 req/min | IP | `/auth/*` |
| `PublicLimiter()` | 120 req/min | IP | `/campus/*`, `/health-check` |
| `ProtectedLimiter()` | 200 req/min | UserID (fallback IP) | `/graphql` |
| `MCPLimiter()` | 60 req/min | IP | `/mcp` |

## How They Work
- Each returns a `fiber.Handler` middleware
- KeyGenerator extracts the rate-limit key (IP or UserID from JWT claims)
- On limit exceeded: returns 429 `{"error": "too many requests, try again later"}`
- Window: 1 minute sliding window

## ProtectedLimiter Special Behavior
Uses `claims.UserID` from `c.Locals("claims")` if available, falls back to `c.IP()` if no claims found. This means authenticated users get per-user limits, not per-IP.

## When Adding a New Limiter
1. Add a new function returning `fiber.Handler` with desired max/window
2. Apply it as middleware on the route group in `cmd/server/main.go`
