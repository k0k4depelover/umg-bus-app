# Skills — internal/auth

## What This Package Does
Handles JWT authentication with a hybrid stateless/stateful architecture. Access tokens are stateless (verified by signature). Refresh tokens are stateful (stored in Redis + PostgreSQL).

## Files
- **jwt.go** — `JWTService` with `GenerateAccess()`, `GenerateRefresh()`, `Verify()`. HS256 signing.
- **handler.go** — `Handler` with `Login()`, `Refresh()`, `Logout()` HTTP endpoints.
- **middleware.go** — `Middleware()` (extracts Bearer token, stores Claims in Locals) + `RequireRole()`.

## Token Architecture
- **Access Token** (15 min): Contains `user_id`, `campus_id`, `role`. No DB lookup needed.
- **Refresh Token** (7 days): Contains `user_id`, `role` (no campus_id). Stored in Redis as `refresh:{sha256(token)}` with value `userID|role`. Also persisted in PostgreSQL `sessions` table for audit.

## Critical Security Details
- Refresh tokens are NEVER stored raw — only SHA-256 hashes
- Passwords verified with `bcrypt.CompareHashAndPassword()`
- On refresh, `campus_id` is fetched fresh from DB (not from old token) — this ensures campus changes take effect on next token refresh
- Refresh tokens are NOT rotated — same token reused until TTL expires or logout

## Redis Key Pattern
```
refresh:{sha256hex(refreshToken)} → "userID|role"  TTL: 7 days
```

## Login Flow
1. Parse body → validate role ("pilot" or "student")
2. Query DB for user by username + active=TRUE
3. Compare bcrypt hash
4. Generate access + refresh tokens
5. Store refresh hash in Redis (7d TTL) + insert into `sessions` table
6. Return both tokens

## When Modifying Auth
- If adding a new role: update `Login()` switch, add DB query, update `RequireRole()` callers
- If changing token claims: update `Claims` struct in jwt.go + all `GenerateAccess()` callers
- If adding token rotation: modify `Refresh()` to delete old + create new Redis entry
- Tests are comprehensive (32 total) — run them after any change

## WebSocket Auth
WS endpoints don't use Fiber middleware. Instead, they extract `?token=` from the query string and call `jwtSvc.Verify()` directly in the handler.
