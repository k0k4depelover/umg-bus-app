# Skills — internal/db

## What This Package Does
Database connection setup and migration runner. Three files, three functions.

## Files
- **postgres.go** — `NewPostgres(dsn string) (*pgxpool.Pool, error)` — Creates a connection pool using pgx v5, pings to verify.
- **redis.go** — `NewRedis(addr, password string) (*redis.Client, error)` — Creates Redis client, pings to verify.
- **migrate.go** — `RunMigrations(dsn string) error` — Runs pending SQL migrations from `./migrations/` using golang-migrate.

## DSN Format
```
postgres://user:password@host:port/dbname?sslmode=disable
```
Built in `config.Load()`.

## Migration Details
- Uses `golang-migrate/migrate/v4` with `file://./migrations` source
- Runs `m.Up()` on every server start — idempotent (handles `ErrNoChange` gracefully)
- Migration files follow pattern: `NNN_description.up.sql` / `NNN_description.down.sql`

## When Adding Migrations
1. Create `backend/migrations/NNN_feature.up.sql` and `.down.sql`
2. They auto-apply on next server start
3. Down migrations: `m.Down()` is not called automatically — manual rollback only
