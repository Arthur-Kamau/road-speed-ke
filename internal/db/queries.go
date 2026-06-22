package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kamau/speed/internal/models"
)

type Queries struct {
	pool *pgxpool.Pool
}

func NewQueries(pool *pgxpool.Pool) *Queries {
	return &Queries{pool: pool}
}

func (q *Queries) GetByBBox(ctx context.Context, bbox models.BBoxQuery) ([]models.RoadSegment, error) {
	query := `
		SELECT id, road_name, road_class, speed_limit_kmh, direction, source, verified, county, last_updated,
			ST_AsGeoJSON(geometry)::jsonb
		FROM road_segments
		WHERE geometry && ST_MakeEnvelope($1, $2, $3, $4, 4326)
		ORDER BY road_name
	`

	rows, err := q.pool.Query(ctx, query, bbox.MinLng, bbox.MinLat, bbox.MaxLng, bbox.MaxLat)
	if err != nil {
		return nil, fmt.Errorf("querying by bbox: %w", err)
	}
	defer rows.Close()

	var segments []models.RoadSegment
	for rows.Next() {
		var seg models.RoadSegment
		var geojsonBytes []byte

		err := rows.Scan(
			&seg.ID, &seg.RoadName, &seg.RoadClass, &seg.SpeedLimitKmh,
			&seg.Direction, &seg.Source, &seg.Verified, &seg.County,
			&seg.LastUpdated, &geojsonBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		if err := json.Unmarshal(geojsonBytes, &seg.Geometry); err != nil {
			return nil, fmt.Errorf("unmarshaling geometry: %w", err)
		}

		segments = append(segments, seg)
	}

	return segments, nil
}

func (q *Queries) GetByID(ctx context.Context, id int64) (*models.RoadSegment, error) {
	query := `
		SELECT id, road_name, road_class, speed_limit_kmh, direction, source, verified, county, last_updated,
			ST_AsGeoJSON(geometry)::jsonb
		FROM road_segments
		WHERE id = $1
	`

	var seg models.RoadSegment
	var geojsonBytes []byte

	err := q.pool.QueryRow(ctx, query, id).Scan(
		&seg.ID, &seg.RoadName, &seg.RoadClass, &seg.SpeedLimitKmh,
		&seg.Direction, &seg.Source, &seg.Verified, &seg.County,
		&seg.LastUpdated, &geojsonBytes,
	)
	if err != nil {
		return nil, fmt.Errorf("querying by id: %w", err)
	}

	if err := json.Unmarshal(geojsonBytes, &seg.Geometry); err != nil {
		return nil, fmt.Errorf("unmarshaling geometry: %w", err)
	}

	return &seg, nil
}

func (q *Queries) GetNearRoute(ctx context.Context, points [][]float64, bufferMeters float64) ([]models.RoadSegment, error) {
	linestring := "LINESTRING("
	for i, p := range points {
		if i > 0 {
			linestring += ","
		}
		linestring += fmt.Sprintf("%f %f", p[1], p[0])
	}
	linestring += ")"

	query := `
		SELECT id, road_name, road_class, speed_limit_kmh, direction, source, verified, county, last_updated,
			ST_AsGeoJSON(geometry)::jsonb
		FROM road_segments
		WHERE ST_DWithin(
			geometry::geography,
			ST_GeomFromText($1, 4326)::geography,
			$2
		)
		ORDER BY road_name
	`

	rows, err := q.pool.Query(ctx, query, linestring, bufferMeters)
	if err != nil {
		return nil, fmt.Errorf("querying near route: %w", err)
	}
	defer rows.Close()

	var segments []models.RoadSegment
	for rows.Next() {
		var seg models.RoadSegment
		var geojsonBytes []byte

		err := rows.Scan(
			&seg.ID, &seg.RoadName, &seg.RoadClass, &seg.SpeedLimitKmh,
			&seg.Direction, &seg.Source, &seg.Verified, &seg.County,
			&seg.LastUpdated, &geojsonBytes,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}

		if err := json.Unmarshal(geojsonBytes, &seg.Geometry); err != nil {
			return nil, fmt.Errorf("unmarshaling geometry: %w", err)
		}

		segments = append(segments, seg)
	}

	return segments, nil
}

func (q *Queries) GetStats(ctx context.Context) (*models.Stats, error) {
	var stats models.Stats

	err := q.pool.QueryRow(ctx, `SELECT COUNT(*) FROM road_segments`).Scan(&stats.TotalSegments)
	if err != nil {
		return nil, err
	}

	err = q.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT road_name) FROM road_segments`).Scan(&stats.TotalRoads)
	if err != nil {
		return nil, err
	}

	err = q.pool.QueryRow(ctx, `SELECT COUNT(*) FROM road_segments WHERE verified = true`).Scan(&stats.VerifiedCount)
	if err != nil {
		return nil, err
	}

	err = q.pool.QueryRow(ctx, `SELECT COUNT(DISTINCT county) FROM road_segments`).Scan(&stats.CountyCoverage)
	if err != nil {
		return nil, err
	}

	return &stats, nil
}
