# Skills — migrations

## What This Directory Does
Versioned SQL migrations for PostgreSQL. Applied automatically on server start via `db.RunMigrations()`.

## Current Migrations

### 001_initial
- Extensions: `pgcrypto` (UUIDs), `citext` (case-insensitive usernames)
- ENUM: `user_role` = 'pilot' | 'student'
- Tables: `campus`, `pilots`, `students`, `student_campus_changes`, `sessions`
- Constraints: lat/lng range checks, username length, phone regex, session expiry > issued_at
- Indexes: active-record partial indexes, campus lookups, token hash

### 002_location_log
- Table: `location_log` (pilot GPS history)
- FKs to `pilots` and `campus` with CASCADE on delete
- Indexes: (campus_id, recorded_at DESC), (pilot_id, recorded_at DESC)

## File Naming Convention
```
NNN_description.up.sql    — applies the migration
NNN_description.down.sql  — rolls it back
```
Number must increment. golang-migrate tracks applied version in a `schema_migrations` table.

## When Adding a Migration
1. Create `NNN_feature.up.sql` and `NNN_feature.down.sql`
2. Write idempotent-safe SQL (use `IF NOT EXISTS` where appropriate)
3. The server applies pending migrations automatically on startup
4. Down migrations are for manual rollback only — never auto-applied
5. Always add CHECK constraints for data integrity (lat/lng ranges, enum values)
6. Add indexes for columns used in WHERE/ORDER BY clauses
