package db

import (
	"context"
	"fmt"

	"github.com/kamau/speed/internal/models"
)

func (q *Queries) GetHazardsByBBox(ctx context.Context, bbox models.BBoxQuery) ([]models.RoadHazard, error) {
	query := `
		SELECT id, hazard_type, description,
			ST_Y(geometry::geometry) AS lat,
			ST_X(geometry::geometry) AS lng,
			source, verified, reported_at
		FROM road_hazards
		WHERE geometry && ST_MakeEnvelope($1, $2, $3, $4, 4326)
		ORDER BY reported_at DESC
	`
	rows, err := q.pool.Query(ctx, query, bbox.MinLng, bbox.MinLat, bbox.MaxLng, bbox.MaxLat)
	if err != nil {
		return nil, fmt.Errorf("querying hazards by bbox: %w", err)
	}
	defer rows.Close()

	var hazards []models.RoadHazard
	for rows.Next() {
		var h models.RoadHazard
		if err := rows.Scan(&h.ID, &h.HazardType, &h.Description, &h.Lat, &h.Lng,
			&h.Source, &h.Verified, &h.ReportedAt); err != nil {
			return nil, fmt.Errorf("scanning hazard row: %w", err)
		}
		hazards = append(hazards, h)
	}
	return hazards, nil
}

func (q *Queries) InsertHazard(ctx context.Context, r models.HazardReport) (int64, error) {
	var id int64
	err := q.pool.QueryRow(ctx, `
		INSERT INTO road_hazards (hazard_type, description, geometry, source)
		VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326), 'user_report')
		RETURNING id
	`, r.HazardType, r.Description, r.Lng, r.Lat).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("inserting hazard: %w", err)
	}
	return id, nil
}

func (q *Queries) InsertSpeedReport(ctx context.Context, r models.SpeedReport) (int64, error) {
	roadClass := r.RoadClass
	if roadClass == "" {
		roadClass = "urban"
	}
	var id int64
	err := q.pool.QueryRow(ctx, `
		INSERT INTO speed_reports (road_name, speed_limit_kmh, road_class, geometry, source)
		VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($4, $5), 4326), 'user_report')
		RETURNING id
	`, r.RoadName, r.SpeedLimitKmh, roadClass, r.Lng, r.Lat).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("inserting speed report: %w", err)
	}
	return id, nil
}
