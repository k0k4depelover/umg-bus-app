# Skills — internal/repository

## What This Package Does
Data access layer. Each repo wraps `*pgxpool.Pool` and exposes query methods that return domain structs.

## Repos and Methods

### CampusRepo
- `GetAll(ctx)` — All active campuses, ordered by name
- `GetByID(ctx, id)` — Single campus by UUID, active only

### PilotRepo
- `GetByCampus(ctx, campusID)` — Active pilot for a campus (LIMIT 1)
- `UpdateLastSeen(ctx, pilotID)` — Sets `last_seen_at = now()`

### StudentRepo
- `ChangeCampus(ctx, studentID, toCampusID)` — Updates student's campus in a transaction + inserts audit record into `student_campus_changes`

## Patterns
- All repos take `*pgxpool.Pool` in constructor
- Use `QueryRow` for single results, `Query` for lists
- Domain structs from `internal/domain/` are returned (not GraphQL models)
- Transactions used in `ChangeCampus()` via `pool.Begin()` → `tx.Commit()`/`tx.Rollback()`

## When Adding a New Repo
1. Create `my_entity_repo.go` with struct holding `*pgxpool.Pool`
2. Add constructor `NewMyEntityRepo(db *pgxpool.Pool)`
3. Write query methods returning `*domain.MyEntity`
4. Wire it in `cmd/server/main.go` `SetupApp()` and pass to handlers/resolvers that need it

## Note
Auth queries (login lookup by username) live in `auth/handler.go`, NOT here. The auth handler queries `pilots`/`students` tables directly for login because it needs `password_hash` which is excluded from domain structs.
