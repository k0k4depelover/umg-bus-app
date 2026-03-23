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

### Key libraries

- `github.com/jackc/pgx/v5` — PostgreSQL driver (pgxpool for concurrency)
- `github.com/redis/go-redis/v9` — Redis client
- `github.com/joho/godotenv` — `.env` loading
