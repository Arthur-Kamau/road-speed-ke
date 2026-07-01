# Kenya Speed Limits — CLAUDE.md

## What This Project Is

Open-source database of Kenyan road speed limits and road hazards. Five delivery surfaces: a Go REST API backed by PostgreSQL+PostGIS, a SvelteKit web frontend with Leaflet maps, a Chrome extension that overlays speed zones on Google Maps, an Android Auto app (Car App Library), and a Kotlin Multiplatform mobile app. The canonical speed data lives as GeoJSON files in `data/geojson/` — the database is populated from these files. Hazard data (bumps, rumble strips, speed cameras) lives in `data/geojson/hazards/`.

The Android Auto app and KMP mobile app live in a separate sibling repo, [speed-ke-mobile](https://github.com/Arthur-Kamau/speed-ke-mobile), not in this repo. Both apps consume the same deployed API. `kmp/` and `android-auto/` are gitignored here in case a local checkout is placed alongside this repo for convenience, but they are not tracked or pushed from this repo.

## Architecture

```
cmd/api/              → API server entrypoint (gin, port 8080)
cmd/scraper/          → Data scraper / seed loader CLI
internal/api/         → HTTP handlers + router (gin + cors)
internal/db/          → PostgreSQL connection pool (pgx) + spatial queries
internal/models/      → Go structs (RoadSegment, RoadHazard, BBoxQuery, Stats)
internal/scraper/     → Kenya Law scraper (colly) + GeoJSON seed loader
data/geojson/         → Source-of-truth speed limit data (GeoJSON FeatureCollections)
data/geojson/hazards/ → Road hazard data (bumps, rumble strips, speed cameras)
migrations/           → PostgreSQL + PostGIS DDL (golang-migrate format)
frontend/             → SvelteKit 2 + Svelte 5 (runes mode) + Leaflet + TypeScript
extension/            → Chrome Manifest V3 extension
```

Mobile apps (separate repo, [speed-ke-mobile](https://github.com/Arthur-Kamau/speed-ke-mobile)):

```
android-auto/         → Android Auto app (Car App Library 1.4.0, phone projection)
kmp/                  → Kotlin Multiplatform app (Compose, OSMDroid map, Android target)
```

## Commands

### Go backend
```bash
docker compose up -d                    # Start PostgreSQL+PostGIS
migrate -path migrations -database "$DATABASE_URL" up  # Run migrations
go run cmd/scraper/main.go --seed       # Load GeoJSON into database
go run cmd/api/main.go serve            # Start API on :8080
go build ./...                          # Build all packages
go vet ./...                            # Lint
go test ./...                           # Test
make dev                                # All-in-one: db + migrate + seed + serve
```

### Frontend (SvelteKit)
```bash
cd frontend
pnpm install                             # Install deps
pnpm run dev                             # Dev server on :5173 (proxies /api to :8080)
pnpm run build                           # Production build
npx svelte-check --tsconfig ./tsconfig.json  # Type check
```

### Data management
```bash
go run cmd/scraper/main.go --scrape --output data/scraped.json  # Scrape Kenya Law
# Regenerate static fallback after editing GeoJSON:
python3 -c "import json,glob; ... " > frontend/static/speeds.json
```

## Code Conventions

- **Go**: Standard library style. `internal/` for non-exported packages. pgx for Postgres (not database/sql). Gin for HTTP. No ORM.
- **Frontend**: Svelte 5 runes mode (`$state`, `$derived`, `$effect`, `$props`). No Svelte 4 stores or reactive statements. TypeScript strict. Components in `src/lib/components/`, services in `src/lib/services/`, types in `src/lib/types/`.
- **GeoJSON**: Each file is a FeatureCollection. Coordinates are `[longitude, latitude]` (GeoJSON standard). Properties must include: `road_name`, `speed_limit_kmh`, `road_class` (urban|peri_urban|highway|expressway), `direction`, `source`, `verified`, `county`, `last_updated`.
- **road_class** values are constrained by a CHECK in the database: `urban`, `peri_urban`, `highway`, `expressway`.

## Key Design Decisions

- **GeoJSON is source of truth**, not the database. The database is derived via the seed command, which **truncates and fully reloads `road_segments` inside one transaction on every run** (see `internal/scraper/seed.go`) — this is what it means for the DB to be "derived": renamed/edited/removed segments in GeoJSON are reflected exactly, with no stale or duplicate rows left behind from previous deploys. `road_segments` has no other write path, so this is safe. Edit `data/geojson/*.geojson` to change speed data.
- **Frontend works offline** — if the Go API is unreachable, it falls back to `frontend/static/speeds.json` (a bundled copy of all GeoJSON data).
- **Routing/geocoding provider is switchable.** Default is Google (Places Autocomplete + Directions API with alternative routes) when `VITE_GOOGLE_MAPS_API_KEY` is set in `frontend/.env`; falls back automatically to the free stack (Nominatim geocoding + OSRM routing, no key, no billing) if the key is absent, or if `VITE_MAP_PROVIDER=free` is set explicitly. See `frontend/src/lib/services/mapConfig.ts`. Map tiles stay OpenStreetMap/Leaflet either way — only geocoding and routing switch providers.
- **Route-to-speed matching** happens client-side in `frontend/src/lib/services/matcher.ts` — it finds the nearest speed limit segment within 200m of each route point. This means GeoJSON coordinates must actually follow the real road (see `/speed-data` skill's coordinate-verification step) — segments placed more than ~200m off the real alignment will silently never match on a live route, even though they show up fine in bbox queries.
- **Nairobi Expressway** is 80 km/h (not the dual carriageway default of 110), per NTSA directive. This is a deliberate exception.
- **Speed limits come from Kenya Traffic Act Cap 403 Section 42 and Legal Notice 62/1975**. Legal sources are documented in `data/LEGAL_SOURCES.md`.
- **Feedback email**: kamaukenn11@gmail.com — shown in the webapp sidebar footer and feedback section.

## Database

PostgreSQL 16 + PostGIS 3.4. Connection via `DATABASE_URL` env var.
Default dev credentials: `speed:speed_dev@localhost:5432/speed_limits` (see docker-compose.yml).

Single table: `road_segments` with a `geometry GEOMETRY(LineString, 4326)` column and a GIST spatial index. Queries use `ST_MakeEnvelope` for bbox and `ST_DWithin` for route proximity.

## Environment Variables

```
DATABASE_URL=postgres://speed:speed_dev@localhost:5432/speed_limits?sslmode=disable
PORT=8080
GIN_MODE=debug
```

Copy `.env.example` to `.env` for local development.

## Hazard Types

- `bump` — physical speed bump
- `rumble_strip` — transverse rumble strips
- `speed_camera` — fixed or frequent mobile NTSA camera position
- `pothole` — severe pothole reported by driver

Hazards are stored in `road_hazards` table (Point geometry) and served at `GET /api/v1/hazards?bbox=...`.
Community-submitted hazards via `POST /api/v1/hazards` go in unverified state.
Community-submitted speed observations via `POST /api/v1/speeds/report` go into `speed_reports` (reviewed before promoting to `road_segments`).

## Android Auto App & KMP Mobile App

Both live in the separate [speed-ke-mobile](https://github.com/Arthur-Kamau/speed-ke-mobile) repo — clone it as a sibling directory (e.g. `../speed-ke-mobile`) to work on them. They are not part of this repo's git history; see `.gitignore` for the local `kmp/`/`android-auto/` exclusions.

**Android Auto app** — Car App Library 1.4.0. Open in Android Studio: `File → Open → android-auto/`.
Test with [Desktop Head Unit (DHU)](https://developer.android.com/training/cars/testing/dhu).
- `mobile/` module = Android Auto phone projection (primary)
- `automotive/` module = Automotive OS skeleton (future)

**KMP mobile app** — Compose Multiplatform, Android target. Open in Android Studio: `File → Open → kmp/`.
Features: OSMDroid map with speed overlays + hazard markers, GPS proximity alerts (same 2km/1km logic),
add speed limit report, add hazard report (both POST to deployed API).

## Deployment

`deploy.sh` is gitignored — copy `deploy.example.sh` to `deploy.sh` and fill in server details.
Never commit `deploy.sh` — it contains server credentials.

## Speed Limit Rules (Kenya)

- 50 km/h — all built-up areas (towns, cities, trading centres)
- 30 km/h — school zones, health facilities, playgrounds
- 110 km/h — dual carriageway highways (private vehicles)
- 100 km/h — single carriageway highways (private vehicles)
- 80 km/h — all PSVs/commercial vehicles on any road
- 65 km/h — vehicles towing trailers
