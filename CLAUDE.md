# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Stack

- **Backend:** Go — `backend/`
- **Mobile:** React Native — `mobile/` (not yet scaffolded)
- **Database:** PostgreSQL (persistent data) + Redis (real-time bus positions)

## Project Purpose

Campus bus monitoring prototype. Bus drivers install the mobile app; it streams GPS position to the backend. Students query the backend to decide whether to walk or wait.

## Backend Commands

```bash
cd backend

# Install dependencies
go mod tidy

# Run
go run main.go

# Build
go build -o bin/server .

# Test
go test ./...
```

## Environment Setup

```bash
cp backend/.env.example backend/.env
# Fill in POSTGRES_PASSWORD and any other values
```

## Architecture

### Database split
- **PostgreSQL** — routes, stops, users, trip history
- **Redis** — current position of each bus driver, key format: `bus:position:<driverID>`
  - Use `SET`/`GET` with short TTLs for positions; consumers poll or subscribe via Pub/Sub

### Backend layout
```
backend/
├── config/config.go   # Loads env vars → Config struct
├── db/
│   ├── postgres.go    # pgxpool connection (thread-safe pool)
│   └── redis.go       # go-redis client + BusPositionKey helper
└── main.go            # Wires everything together
```

### Key libraries
- `github.com/jackc/pgx/v5` — PostgreSQL driver (pgxpool for concurrency)
- `github.com/redis/go-redis/v9` — Redis client
- `github.com/joho/godotenv` — `.env` loading
