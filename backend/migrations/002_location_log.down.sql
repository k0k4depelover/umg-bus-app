CREATE TABLE location_log(
  log_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  pilot_id UUID NOT NULL REFERENCES pilots(pilot_id) ON DELETE CASCADE,
  campus_id UUID NOT NULL REFERENCES campus(campus_id) ON DELETE CASCADE,
  lat DOUBLE PRECISION NOT NULL,
  lng DOUBLE PRECISION NOT NULL,
  bearing     DOUBLE PRECISION NOT NULL DEFAULT 0,
  speed_kmh   DOUBLE PRECISION NOT NULL DEFAULT 0,
  recorded_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT chk_lat_range CHECK (lat BETWEEN -90  AND 90),
    CONSTRAINT chk_lng_range CHECK (lng BETWEEN -180 AND 180)
)

-- Índice para queries ML: todos los puntos de un campus ordenados por tiempo
CREATE INDEX idx_location_log_campus_time
    ON location_log (campus_id, recorded_at DESC);

-- Índice para queries por piloto
CREATE INDEX idx_location_log_pilot
    ON location_log (pilot_id, recorded_at DESC);
