CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS road_segments (
    id              BIGSERIAL PRIMARY KEY,
    road_name       TEXT NOT NULL,
    road_class      TEXT NOT NULL CHECK (road_class IN ('urban', 'peri_urban', 'highway', 'expressway')),
    speed_limit_kmh INTEGER NOT NULL CHECK (speed_limit_kmh > 0 AND speed_limit_kmh <= 200),
    direction       TEXT NOT NULL DEFAULT 'both' CHECK (direction IN ('both', 'forward', 'backward')),
    source          TEXT NOT NULL,
    verified        BOOLEAN NOT NULL DEFAULT false,
    county          TEXT NOT NULL DEFAULT '',
    last_updated    DATE NOT NULL DEFAULT CURRENT_DATE,
    geometry        GEOMETRY(LineString, 4326) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_road_segments_geometry ON road_segments USING GIST (geometry);
CREATE INDEX idx_road_segments_road_name ON road_segments (road_name);
CREATE INDEX idx_road_segments_county ON road_segments (county);
CREATE INDEX idx_road_segments_speed ON road_segments (speed_limit_kmh);
