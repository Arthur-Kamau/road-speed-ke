-- Road hazards: bumps, rumble strips, speed cameras, potholes
CREATE TABLE IF NOT EXISTS road_hazards (
    id          BIGSERIAL PRIMARY KEY,
    hazard_type TEXT NOT NULL CHECK (hazard_type IN ('bump', 'rumble_strip', 'speed_camera', 'pothole')),
    description TEXT NOT NULL DEFAULT '',
    geometry    GEOMETRY(Point, 4326) NOT NULL,
    source      TEXT NOT NULL DEFAULT 'user_report',
    verified    BOOLEAN NOT NULL DEFAULT false,
    reported_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_road_hazards_geometry ON road_hazards USING GIST (geometry);
CREATE INDEX IF NOT EXISTS idx_road_hazards_type ON road_hazards (hazard_type);

-- Unverified community speed limit reports (manually reviewed before adding to road_segments)
CREATE TABLE IF NOT EXISTS speed_reports (
    id              BIGSERIAL PRIMARY KEY,
    road_name       TEXT NOT NULL,
    speed_limit_kmh INTEGER NOT NULL CHECK (speed_limit_kmh > 0 AND speed_limit_kmh <= 200),
    road_class      TEXT NOT NULL DEFAULT 'urban',
    geometry        GEOMETRY(Point, 4326) NOT NULL,
    source          TEXT NOT NULL DEFAULT 'user_report',
    reviewed        BOOLEAN NOT NULL DEFAULT false,
    reported_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_speed_reports_geometry ON speed_reports USING GIST (geometry);
