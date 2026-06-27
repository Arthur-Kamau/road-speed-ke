# Kenya Speed Limits — CLAUDE.md

## What This Project Is

Open-source database of Kenyan road speed limits. Three delivery surfaces: a Go REST API backed by PostgreSQL+PostGIS, a SvelteKit web frontend with Leaflet maps, and a Chrome extension that overlays speed zones on Google Maps. The canonical speed data lives as GeoJSON files in `data/geojson/` — the database is populated from these files.

## Architecture

```
cmd/api/          → API server entrypoint (gin, port 8080)
cmd/scraper/      → Data scraper / seed loader CLI
internal/api/     → HTTP handlers + router (gin + cors)
internal/db/      → PostgreSQL connection pool (pgx) + spatial queries
internal/models/  → Go structs (RoadSegment, BBoxQuery, Stats)
internal/scraper/ → Kenya Law scraper (colly) + GeoJSON seed loader
data/geojson/     → Source-of-truth speed limit data (GeoJSON FeatureCollections)
migrations/       → PostgreSQL + PostGIS DDL (golang-migrate format)
frontend/         → SvelteKit 2 + Svelte 5 (runes mode) + Leaflet + TypeScript
extension/        → Chrome Manifest V3 extension
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

- **GeoJSON is source of truth**, not the database. The database is derived via the seed command. Edit `data/geojson/*.geojson` to change speed data.
- **Frontend works offline** — if the Go API is unreachable, it falls back to `frontend/static/speeds.json` (a bundled copy of all GeoJSON data).
- **Routing uses OSRM** (free, no API key). Geocoding uses Nominatim. Map tiles are OpenStreetMap. No paid API keys required.
- **Route-to-speed matching** happens client-side in `frontend/src/lib/services/matcher.ts` — it finds the nearest speed limit segment within 200m of each route point.
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

## Speed Limit Rules (Kenya)

- 50 km/h — all built-up areas (towns, cities, trading centres)
- 30 km/h — school zones, health facilities, playgrounds
- 110 km/h — dual carriageway highways (private vehicles)
- 100 km/h — single carriageway highways (private vehicles)
- 80 km/h — all PSVs/commercial vehicles on any road
- 65 km/h — vehicles towing trailers
