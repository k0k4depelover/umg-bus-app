package testutil

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
	"github.com/umg-bus-app/backend/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

const TestJWTSecret = "integration-test-secret"

// TestEnv holds all dependencies needed for integration tests.
type TestEnv struct {
	PG        *pgxpool.Pool
	Redis     *redis.Client
	JWT       *auth.JWTService
	CampusID  string
	PilotID   string
	StudentID string
}

var (
	sharedEnv  *TestEnv
	sharedOnce sync.Once
	sharedErr  error
)

// Setup returns a shared TestEnv. Containers are created once and reused
// across all tests in the same process. Each test gets a clean Redis state.
func Setup(t *testing.T) *TestEnv {
	t.Helper()

	sharedOnce.Do(func() {
		sharedEnv, sharedErr = createEnv()
	})

	if sharedErr != nil {
		t.Fatalf("failed to create test environment: %v", sharedErr)
	}

	// Flush Redis between tests for isolation
	sharedEnv.Redis.FlushAll(context.Background())

	return sharedEnv
}

func createEnv() (*TestEnv, error) {
	ctx := context.Background()

	// --- Postgres container ---
	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("testuser"),
		tcpostgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("postgres container: %w", err)
	}

	pgConnStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("postgres conn string: %w", err)
	}

	pool, err := pgxpool.New(ctx, pgConnStr)
	if err != nil {
		return nil, fmt.Errorf("pgxpool: %w", err)
	}

	// --- Redis container ---
	redisContainer, err := tcredis.Run(ctx, "redis:7-alpine")
	if err != nil {
		return nil, fmt.Errorf("redis container: %w", err)
	}

	redisEndpoint, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("redis endpoint: %w", err)
	}

	rdb := redis.NewClient(&redis.Options{Addr: redisEndpoint})
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	// --- Run schema ---
	if err := runSchema(pool); err != nil {
		return nil, fmt.Errorf("schema: %w", err)
	}

	env := &TestEnv{
		PG:    pool,
		Redis: rdb,
		JWT:   auth.NewJWTService(TestJWTSecret),
	}

	if err := seedData(env); err != nil {
		return nil, fmt.Errorf("seed: %w", err)
	}

	fmt.Printf("Test containers ready: campus=%s pilot=%s student=%s\n", env.CampusID, env.PilotID, env.StudentID)
	return env, nil
}

func runSchema(pool *pgxpool.Pool) error {
	ctx := context.Background()

	schema := `
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

DO $$ BEGIN
    CREATE TYPE user_role AS ENUM ('pilot', 'student');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS campus (
    campus_id        UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT             NOT NULL,
    city             TEXT             NOT NULL,
    bound_sw_lat     DOUBLE PRECISION NOT NULL,
    bound_sw_lng     DOUBLE PRECISION NOT NULL,
    bound_ne_lat     DOUBLE PRECISION NOT NULL,
    bound_ne_lng     DOUBLE PRECISION NOT NULL,
    route_geojson    JSONB            NULL,
    active           BOOLEAN          NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ      NOT NULL DEFAULT now(),
    CONSTRAINT chk_campus_bounds_lat CHECK (bound_sw_lat < bound_ne_lat),
    CONSTRAINT chk_campus_bounds_lng CHECK (bound_sw_lng < bound_ne_lng)
);

CREATE TABLE IF NOT EXISTS pilots (
    pilot_id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    campus_id        UUID        NOT NULL REFERENCES campus(campus_id),
    username         CITEXT      NOT NULL,
    password_hash    TEXT        NOT NULL,
    secret_code      UUID        NOT NULL DEFAULT gen_random_uuid(),
    secret_used_at   TIMESTAMPTZ NULL,
    full_name        TEXT        NOT NULL,
    phone            TEXT        NULL,
    active           BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at     TIMESTAMPTZ NULL,
    CONSTRAINT uq_pilot_username UNIQUE (username)
);

CREATE TABLE IF NOT EXISTS students (
    student_id       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    campus_id        UUID        NOT NULL REFERENCES campus(campus_id),
    username         CITEXT      NOT NULL,
    password_hash    TEXT        NOT NULL,
    full_name        TEXT        NOT NULL,
    active           BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_student_username UNIQUE (username)
);

CREATE TABLE IF NOT EXISTS student_campus_changes (
    change_id       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id      UUID        NOT NULL REFERENCES students(student_id),
    from_campus_id  UUID        NOT NULL REFERENCES campus(campus_id),
    to_campus_id    UUID        NOT NULL REFERENCES campus(campus_id),
    changed_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_campus_change_diff CHECK (from_campus_id <> to_campus_id)
);

CREATE TABLE IF NOT EXISTS sessions (
    session_id    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL,
    user_role     user_role   NOT NULL,
    token         TEXT        NOT NULL,
    issued_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at    TIMESTAMPTZ NOT NULL,
    revoked_at    TIMESTAMPTZ NULL,
    revoke_reason TEXT        NULL
);

CREATE TABLE IF NOT EXISTS location_log (
    log_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    pilot_id    UUID NOT NULL REFERENCES pilots(pilot_id) ON DELETE CASCADE,
    campus_id   UUID NOT NULL REFERENCES campus(campus_id) ON DELETE CASCADE,
    lat         DOUBLE PRECISION NOT NULL,
    lng         DOUBLE PRECISION NOT NULL,
    bearing     DOUBLE PRECISION NOT NULL DEFAULT 0,
    speed_kmh   DOUBLE PRECISION NOT NULL DEFAULT 0,
    recorded_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`
	_, err := pool.Exec(ctx, schema)
	return err
}

func seedData(env *TestEnv) error {
	ctx := context.Background()

	err := env.PG.QueryRow(ctx, `
		INSERT INTO campus (name, city, bound_sw_lat, bound_sw_lng, bound_ne_lat, bound_ne_lng)
		VALUES ('Campus Central', 'Guatemala City', 14.5, -90.6, 14.7, -90.4)
		RETURNING campus_id
	`).Scan(&env.CampusID)
	if err != nil {
		return fmt.Errorf("seed campus: %w", err)
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("pilotpass123"), bcrypt.DefaultCost)
	err = env.PG.QueryRow(ctx, `
		INSERT INTO pilots (campus_id, username, password_hash, full_name, phone)
		VALUES ($1, 'testpilot', $2, 'Test Pilot', '+50212345678')
		RETURNING pilot_id
	`, env.CampusID, string(hash)).Scan(&env.PilotID)
	if err != nil {
		return fmt.Errorf("seed pilot: %w", err)
	}

	hash, _ = bcrypt.GenerateFromPassword([]byte("studentpass123"), bcrypt.DefaultCost)
	err = env.PG.QueryRow(ctx, `
		INSERT INTO students (campus_id, username, password_hash, full_name)
		VALUES ($1, 'teststudent', $2, 'Test Student')
		RETURNING student_id
	`, env.CampusID, string(hash)).Scan(&env.StudentID)
	if err != nil {
		return fmt.Errorf("seed student: %w", err)
	}

	hash, _ = bcrypt.GenerateFromPassword([]byte("inactive123"), bcrypt.DefaultCost)
	_, err = env.PG.Exec(ctx, `
		INSERT INTO pilots (campus_id, username, password_hash, full_name, active)
		VALUES ($1, 'inactivepilot', $2, 'Inactive Pilot', FALSE)
	`, env.CampusID, string(hash))
	if err != nil {
		return fmt.Errorf("seed inactive pilot: %w", err)
	}

	_, err = env.PG.Exec(ctx, `
		INSERT INTO campus (name, city, bound_sw_lat, bound_sw_lng, bound_ne_lat, bound_ne_lng)
		VALUES ('Campus Norte', 'Quetzaltenango', 14.8, -91.6, 15.0, -91.4)
	`)
	if err != nil {
		return fmt.Errorf("seed second campus: %w", err)
	}

	return nil
}
