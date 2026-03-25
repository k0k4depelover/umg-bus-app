# Skills — internal/config

## What This Package Does
Single file (`config.go`) that loads configuration from environment variables with fallback defaults.

## Config Struct Fields
| Field | Env Var | Default |
|-------|---------|---------|
| PostgresDSN | built from POSTGRES_* vars | `postgres://admin:postgresumgtracker@localhost:5436/transportation-tracker-db?sslmode=disable` |
| RedisAddr | REDIS_ADDR | `localhost:6381` |
| RedisPass | REDIS_PASS | *(empty)* |
| AppPort | APP_PORT | `8084` |
| WSPort | WS_PORT | `8085` |
| JWTSecret | JWT_SECRET | `dev-secret-cambia-en-produccion` |

## Usage
```go
cfg := config.Load()  // reads env vars, returns Config struct
```

## When Adding New Config
1. Add field to `Config` struct
2. Read from env in `Load()` with `os.Getenv()` and a default
3. Update `.env.example` with the new variable
