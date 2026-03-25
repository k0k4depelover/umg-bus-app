# Skills — internal/testutil

## What This Package Does
Test infrastructure using testcontainers. Spins up real PostgreSQL + Redis containers for integration tests.

## Key Function: `Setup(t *testing.T) *TestEnv`
Returns a `TestEnv` with:
- `Pool` — `*pgxpool.Pool` connected to test PostgreSQL
- `RDB` — `*redis.Client` connected to test Redis
- `JWT` — `*auth.JWTService` with test secret
- `CampusID` — UUID of seeded test campus
- `PilotID` — UUID of seeded active pilot
- `StudentID` — UUID of seeded student

## Seeded Test Data
- 1 campus
- 1 active pilot (username: varies, password: `pilotpass123`)
- 1 student (username: varies, password: `studentpass123`)
- 1 inactive pilot (for testing inactive user rejection)

## Container Reuse
Containers are reused across tests in the same `go test` run for speed. Redis is flushed (`FlushAll`) between tests to ensure isolation.

## Usage in Tests
```go
func TestSomething(t *testing.T) {
    env := testutil.Setup(t)
    // use env.Pool, env.RDB, env.JWT, env.CampusID, etc.
}
```

## When Adding Test Data
If your tests need additional seed data (e.g., a new entity type), add INSERT statements to the `Setup()` function and expose the IDs on `TestEnv`.
