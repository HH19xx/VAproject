CREATE TABLE IF NOT EXISTS world_signals (
    id SERIAL PRIMARY KEY,
    source VARCHAR(64) NOT NULL DEFAULT 'open_meteo',
    location_key VARCHAR(128) NOT NULL,
    latitude NUMERIC(9,6) NOT NULL,
    longitude NUMERIC(9,6) NOT NULL,
    observed_at TIMESTAMPTZ NOT NULL,
    temperature_c DOUBLE PRECISION NULL,
    precipitation_mm DOUBLE PRECISION NULL,
    wind_speed_ms DOUBLE PRECISION NULL,
    weather_code INTEGER NULL,
    raw_json JSONB NULL,
    create_user VARCHAR(255) NOT NULL,
    update_user VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_world_signals_source_location_observed
ON world_signals(source, location_key, observed_at)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_world_signals_lookup
ON world_signals(source, location_key, observed_at DESC)
WHERE deleted_at IS NULL;
