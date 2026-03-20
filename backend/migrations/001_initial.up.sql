-- ============================================================
--  Bus Tracking System — PostgreSQL Schema (Final)
--  3NF · Integridad referencial · Consistencia de datos
-- Paso no.4 -> Debemos tener definida la base de datos en el
-- contenedor, o por el gestor.
-- ============================================================

CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "citext";

CREATE TYPE user_role AS ENUM ('pilot', 'student');

-- ============================================================
--  1. campus
-- ============================================================

CREATE TABLE campus (
    campus_id        UUID             PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT             NOT NULL,
    city             TEXT             NOT NULL,

    bound_sw_lat     DOUBLE PRECISION NOT NULL,
    bound_sw_lng     DOUBLE PRECISION NOT NULL,
    bound_ne_lat     DOUBLE PRECISION NOT NULL,
    bound_ne_lng     DOUBLE PRECISION NOT NULL,

    route_geojson    JSONB            NULL,
    active           BOOLEAN          NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ      NOT NULL DEFAULT now(),  -- FIX: era "create_at"

    CONSTRAINT chk_campus_bounds_lat CHECK (bound_sw_lat < bound_ne_lat),
    CONSTRAINT chk_campus_bounds_lng CHECK (bound_sw_lng < bound_ne_lng),
    CONSTRAINT chk_campus_lat_range  CHECK (
        bound_sw_lat BETWEEN -90  AND 90  AND
        bound_ne_lat BETWEEN -90  AND 90
    ),
    CONSTRAINT chk_campus_lng_range  CHECK (
        bound_sw_lng BETWEEN -180 AND 180 AND
        bound_ne_lng BETWEEN -180 AND 180
    )
);

COMMENT ON TABLE  campus               IS 'Entidad raíz. Campus universitarios con límites geográficos de mapa.';
COMMENT ON COLUMN campus.route_geojson IS 'GeoJSON LineString de la ruta predefinida. NULL si no aplica aún.';
COMMENT ON COLUMN campus.bound_sw_lat  IS 'Esquina suroeste del bounding box (latitud). Usado por el cliente para fitBounds.';

-- ============================================================
--  2. pilots
-- ============================================================

CREATE TABLE pilots (
    pilot_id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(), -- FIX: faltaba PK y DEFAULT
    campus_id        UUID        NOT NULL REFERENCES campus(campus_id)  -- FIX: faltaba NOT NULL y FK
                                 ON UPDATE CASCADE
                                 ON DELETE RESTRICT,

    username         CITEXT      NOT NULL,                              -- FIX: faltaba NOT NULL
    password_hash    TEXT        NOT NULL,                              -- FIX: faltaba NOT NULL

    secret_code      UUID        NOT NULL DEFAULT gen_random_uuid(),
    secret_used_at   TIMESTAMPTZ NULL,

    full_name        TEXT        NOT NULL,
    phone            TEXT        NULL,
    active           BOOLEAN     NOT NULL DEFAULT TRUE,                 -- FIX: faltaba DEFAULT TRUE
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_seen_at     TIMESTAMPTZ NULL,

    CONSTRAINT uq_pilot_username     UNIQUE (username),
    CONSTRAINT uq_pilot_secret_code  UNIQUE (secret_code),
    CONSTRAINT chk_pilot_username_len CHECK (char_length(username) BETWEEN 4 AND 50),
    CONSTRAINT chk_pilot_phone        CHECK (phone IS NULL OR phone ~ '^\+?[0-9\s\-]{7,20}$')
);

COMMENT ON TABLE  pilots              IS 'Conductores autorizados por campus. El celular del piloto actúa como GPS del bus.';
COMMENT ON COLUMN pilots.secret_code  IS 'UUID prerregistrado por admin. secret_used_at NOT NULL indica que ya fue activado.';
COMMENT ON COLUMN pilots.last_seen_at IS 'Actualizado solo en logout o expiración de TTL Redis. No en cada ping GPS.';

-- ============================================================
--  3. students
--     Sin carnet — adopción más fácil para estudiantes.
-- ============================================================

CREATE TABLE students (
    student_id       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    campus_id        UUID        NOT NULL REFERENCES campus(campus_id)
                                 ON UPDATE CASCADE
                                 ON DELETE RESTRICT,

    username         CITEXT      NOT NULL,
    password_hash    TEXT        NOT NULL,

    full_name        TEXT        NOT NULL,

    active           BOOLEAN     NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT uq_student_username      UNIQUE (username),
    CONSTRAINT chk_student_username_len CHECK (char_length(username) BETWEEN 4 AND 50)
);

COMMENT ON TABLE students IS 'Usuarios estudiantes. campus_id editable vía endpoint protegido.';

-- ============================================================
--  4. student_campus_changes
-- ============================================================

CREATE TABLE student_campus_changes (
    change_id       UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id      UUID        NOT NULL REFERENCES students(student_id)
                                ON UPDATE CASCADE
                                ON DELETE CASCADE,
    from_campus_id  UUID        NOT NULL REFERENCES campus(campus_id)
                                ON UPDATE CASCADE
                                ON DELETE RESTRICT,
    to_campus_id    UUID        NOT NULL REFERENCES campus(campus_id)
                                ON UPDATE CASCADE
                                ON DELETE RESTRICT,
    changed_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_campus_change_diff CHECK (from_campus_id <> to_campus_id)
);

COMMENT ON TABLE student_campus_changes IS 'Log inmutable. Permite detectar cambios de campus anómalos o abusivos.';

-- ============================================================
--  5. sessions
--     Log histórico. Tokens activos viven en Redis con TTL.
--     Solo se escribe en login y logout.
-- ============================================================

CREATE TABLE sessions (
    session_id    UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id       UUID        NOT NULL,                                 -- FIX: faltaba NOT NULL
    user_role     user_role   NOT NULL,
    token_hash    TEXT        NOT NULL,
    issued_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at    TIMESTAMPTZ NOT NULL,
    revoked_at    TIMESTAMPTZ NULL,
    revoke_reason TEXT        NULL,

    CONSTRAINT uq_session_token_hash UNIQUE (token_hash),
    CONSTRAINT chk_session_expiry    CHECK (expires_at > issued_at),
    CONSTRAINT chk_revoke_reason     CHECK (
        revoke_reason IS NULL OR
        revoke_reason IN ('logout', 'password_change', 'admin_revoke')
    )
);

COMMENT ON TABLE  sessions            IS 'Log de sesiones. Tokens activos viven en Redis. Esta tabla es solo auditoría.';
COMMENT ON COLUMN sessions.token_hash IS 'SHA-256 del refresh token. Nunca el token raw.';
COMMENT ON COLUMN sessions.user_id    IS 'pilot_id o student_id según user_role. Sin FK por diseño (herencia de tabla).';

-- ============================================================
--  ÍNDICES
-- ============================================================

CREATE INDEX idx_campus_active
    ON campus (active)
    WHERE active = TRUE;

CREATE INDEX idx_pilots_campus
    ON pilots (campus_id);

CREATE INDEX idx_pilots_active
    ON pilots (active)
    WHERE active = TRUE;

CREATE INDEX idx_students_campus
    ON students (campus_id);

CREATE INDEX idx_students_active
    ON students (active)
    WHERE active = TRUE;

CREATE INDEX idx_campus_changes_student
    ON student_campus_changes (student_id, changed_at DESC);

CREATE INDEX idx_sessions_user
    ON sessions (user_id, user_role)
    WHERE revoked_at IS NULL;

CREATE INDEX idx_sessions_token_hash
    ON sessions (token_hash);
