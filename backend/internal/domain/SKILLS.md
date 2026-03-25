# Skills — internal/domain

## What This Package Does
Pure data models (structs) representing database entities. No logic, no methods — just field definitions.

## Models
- **Campus** — campus_id, name, city, map bounds (SW/NE lat/lng), optional route_geojson, active, created_at
- **Pilot** — pilot_id, campus_id, username, full_name, phone, active, last_seen_at, created_at (password_hash excluded)
- **Student** — student_id, campus_id, username, full_name, active, created_at (password_hash excluded)

## Design Notes
- `password_hash` is intentionally NOT in domain structs — it's only accessed directly in `auth/handler.go` login queries
- UUIDs use `github.com/google/uuid` type
- Optional fields use pointers (`*string`, `*time.Time`)
- These structs are returned by repositories and consumed by handlers/resolvers

## When Adding a New Model
1. Create `my_entity.go` in this directory
2. Match fields to your migration table columns
3. Exclude sensitive fields (passwords, secrets) from the struct
