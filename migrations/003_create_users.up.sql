CREATE TABLE IF NOT EXISTS users (
    id          BIGSERIAL PRIMARY KEY,
    google_id   TEXT NOT NULL UNIQUE,
    email       TEXT NOT NULL,
    name        TEXT NOT NULL DEFAULT '',
    picture_url TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_google_id ON users (google_id);

ALTER TABLE road_hazards ADD COLUMN IF NOT EXISTS user_id BIGINT REFERENCES users(id);
ALTER TABLE speed_reports ADD COLUMN IF NOT EXISTS user_id BIGINT REFERENCES users(id);
