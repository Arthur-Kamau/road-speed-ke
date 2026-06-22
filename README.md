# Kenya Speed Limits Database & Map Overlay

An open-source database of Kenyan road speed limits with a Go API and Chrome extension that overlays color-coded speed zones on Google Maps.

## The Problem

Kenya's National Transport Safety Authority (NTSA) has deployed mobile speed detectors with instant fines across the country. However, road signage for speed limits is poor or missing entirely. Drivers often have no way to know the legal speed limit on the road they're traveling — until they're fined.

## The Solution

1. **Open Speed Limit Database** — Scraped from Kenya Law (Traffic Act Cap 403), Kenya Gazette notices, and NTSA publications. Stored as GeoJSON in this repo so anyone can use, verify, and contribute.
2. **REST API** — A Go server backed by PostgreSQL with PostGIS for spatial queries. Query speed limits by bounding box, route, or road ID.
3. **Chrome Extension** — Injects color-coded speed limit overlays directly onto Google Maps. No separate app needed — works where people already plan their trips.

## Architecture

```
┌─────────────────────────────────────────────────┐
│  Chrome Extension (overlay on Google Maps)       │
│  - Color-coded polylines on route segments       │
│  - Speed limit popup on click                    │
└──────────────────────┬──────────────────────────┘
                       │ REST API
┌──────────────────────▼──────────────────────────┐
│  Go API Server (gin)                             │
│  GET /api/v1/speeds?bbox=lat1,lng1,lat2,lng2    │
│  GET /api/v1/speeds/route?points=...            │
│  GET /api/v1/roads/:id                          │
│  GET /api/v1/stats                              │
└──────────────────────┬──────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────┐
│  PostgreSQL + PostGIS                            │
│  - Spatial indexing for bbox/proximity queries   │
│  - Road segments with speed limit metadata       │
└─────────────────────────────────────────────────┘
```

## Speed Limit Color Coding

| Color  | Speed (km/h) | Typical Context         |
|--------|-------------- |-------------------------|
| 🟢 Green  | ≤ 50          | Urban / town centers    |
| 🟡 Yellow | 51–80         | Peri-urban / county roads |
| 🟠 Orange | 81–100        | Highways                |
| 🔴 Red    | 101–110       | Expressways             |

## Data Sources

- **Kenya Traffic Act (Cap 403)** — Default speed limits by road classification
- **Kenya Gazette Notices** — Specific speed limit declarations for individual roads
- **NTSA Publications** — Official speed zone maps and announcements
- **OpenStreetMap** — Road geometries (CC-BY-SA), paired with legal speed data

## Project Structure

```
speed/
├── cmd/
│   ├── api/            # API server entrypoint
│   └── scraper/        # Data scraper entrypoint
├── internal/
│   ├── api/            # HTTP handlers and routes
│   ├── db/             # Database connection and queries
│   ├── models/         # Data structures
│   └── scraper/        # Scraping logic for Kenya Law
├── data/
│   └── geojson/        # Open speed limit data (GeoJSON)
├── migrations/         # PostgreSQL migration files
├── extension/          # Chrome extension (Manifest V3)
│   ├── src/            # Extension JavaScript
│   ├── icons/          # Extension icons
│   └── styles/         # Extension CSS
├── scripts/            # Utility scripts
└── docker-compose.yml  # Local dev setup (PostgreSQL + PostGIS)
```

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL 16+ with PostGIS extension
- Node.js 18+ and npm (for the frontend)
- Chrome or Chromium (for the extension)

### Option A: Without Docker (local PostgreSQL)

If you already have PostgreSQL running locally:

#### 1. Create the database and enable PostGIS

```bash
psql -U your_user -d postgres -c "CREATE DATABASE speed_limit_ke;"
psql -U your_user -d speed_limit_ke -c "CREATE EXTENSION IF NOT EXISTS postgis;"
```

> **Note:** PostGIS must be installed on your system. On Ubuntu/Debian: `sudo apt install postgresql-17-postgis-3` (match your PG version).

#### 2. Configure environment

```bash
cp .env.example .env
```

Edit `.env` to match your local PostgreSQL credentials:

```
DATABASE_URL=postgres://your_user:your_password@localhost:5432/speed_limit_ke?sslmode=disable
PORT=8080
GIN_MODE=debug
```

#### 3. Run migrations

```bash
psql "$DATABASE_URL" -f migrations/001_create_road_segments.up.sql
```

#### 4. Seed the data

```bash
go run cmd/scraper/main.go --seed
```

#### 5. Start the API server

```bash
go run cmd/api/main.go serve
```

The API will be available at `http://localhost:8080`.

#### 6. Start the frontend

```bash
cd frontend
npm install
npm run dev
```

The frontend will be available at `http://localhost:5173`.

### Option B: With Docker

If you prefer an isolated setup with Docker:

#### 1. Start the database

```bash
docker compose up -d
```

#### 2. Configure environment

```bash
cp .env.example .env
```

The default `.env.example` credentials match the Docker container — no edits needed.

#### 3. Run migrations, seed, and start

```bash
psql "$DATABASE_URL" -f migrations/001_create_road_segments.up.sql
go run cmd/scraper/main.go --seed
go run cmd/api/main.go serve
```

Or use the Makefile shortcut:

```bash
make dev
```

### Load the Chrome Extension

1. Open `chrome://extensions/`
2. Enable "Developer mode"
3. Click "Load unpacked" and select the `extension/` directory
4. Open Google Maps — you should see speed limit overlays

## API Endpoints

### Get speed limits by bounding box

```
GET /api/v1/speeds?bbox=-1.35,36.70,-1.20,36.90
```

Returns all road segments with speed limits within the bounding box.

### Get speed limits along a route

```
GET /api/v1/speeds/route?points=-1.286,36.817,-1.163,36.955
```

Returns speed limit segments along a series of coordinates.

### Get a specific road

```
GET /api/v1/roads/:id
```

### Get database stats

```
GET /api/v1/stats
```

## Contributing

We need help from Kenyan drivers, developers, and legal researchers:

1. **Add missing roads** — Submit GeoJSON with speed limit data via pull request
2. **Verify existing data** — Cross-check against gazette notices or physical signage
3. **Report errors** — Open an issue if a speed limit is wrong
4. **Improve the scraper** — Help extract data from more Kenya Law sources

### GeoJSON Format

Each road segment in `data/geojson/` follows this schema:

```json
{
  "type": "Feature",
  "properties": {
    "road_name": "Mombasa Road (A109)",
    "speed_limit_kmh": 80,
    "road_class": "highway",
    "direction": "both",
    "source": "Kenya Gazette Vol. CXVII No. 42",
    "verified": true,
    "last_updated": "2025-01-15"
  },
  "geometry": {
    "type": "LineString",
    "coordinates": [[36.8219, -1.3191], [36.8432, -1.3054]]
  }
}
```

## Legal

Speed limit data is sourced from publicly available Kenyan law and government publications. This project is for informational purposes — always obey posted signs and exercise safe driving judgment.

## License

MIT
