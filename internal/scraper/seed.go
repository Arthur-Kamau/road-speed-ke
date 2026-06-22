package scraper

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
)

type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONFeature struct {
	Type       string          `json:"type"`
	Properties Properties      `json:"properties"`
	Geometry   json.RawMessage `json:"geometry"`
}

type Properties struct {
	RoadName      string `json:"road_name"`
	SpeedLimitKmh int    `json:"speed_limit_kmh"`
	RoadClass     string `json:"road_class"`
	Direction     string `json:"direction"`
	Source        string `json:"source"`
	Verified      bool   `json:"verified"`
	County        string `json:"county"`
	LastUpdated   string `json:"last_updated"`
}

func SeedFromGeoJSON(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	files, err := filepath.Glob(filepath.Join(dir, "*.geojson"))
	if err != nil {
		return fmt.Errorf("listing geojson files: %w", err)
	}

	for _, file := range files {
		if err := loadGeoJSONFile(ctx, pool, file); err != nil {
			return fmt.Errorf("loading %s: %w", file, err)
		}
	}

	return nil
}

func loadGeoJSONFile(ctx context.Context, pool *pgxpool.Pool, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var fc GeoJSONFeatureCollection
	if err := json.Unmarshal(data, &fc); err != nil {
		return fmt.Errorf("parsing geojson: %w", err)
	}

	for _, f := range fc.Features {
		geomStr := string(f.Geometry)

		_, err := pool.Exec(ctx, `
			INSERT INTO road_segments (road_name, road_class, speed_limit_kmh, direction, source, verified, county, last_updated, geometry)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8::date, ST_SetSRID(ST_GeomFromGeoJSON($9), 4326))
			ON CONFLICT DO NOTHING
		`,
			f.Properties.RoadName,
			f.Properties.RoadClass,
			f.Properties.SpeedLimitKmh,
			f.Properties.Direction,
			f.Properties.Source,
			f.Properties.Verified,
			f.Properties.County,
			f.Properties.LastUpdated,
			geomStr,
		)
		if err != nil {
			return fmt.Errorf("inserting segment %s: %w", f.Properties.RoadName, err)
		}
	}

	fmt.Printf("Loaded %d segments from %s\n", len(fc.Features), filepath.Base(path))
	return nil
}
