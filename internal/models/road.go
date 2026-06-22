package models

import "time"

type RoadSegment struct {
	ID            int64     `json:"id"`
	RoadName      string    `json:"road_name"`
	RoadClass     string    `json:"road_class"`
	SpeedLimitKmh int       `json:"speed_limit_kmh"`
	Direction     string    `json:"direction"`
	Source        string    `json:"source"`
	Verified      bool      `json:"verified"`
	County        string    `json:"county"`
	LastUpdated   time.Time `json:"last_updated"`
	Geometry      GeoJSON   `json:"geometry"`
}

type GeoJSON struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

type BBoxQuery struct {
	MinLat float64
	MinLng float64
	MaxLat float64
	MaxLng float64
}

type RouteQuery struct {
	Points [][]float64
}

type Stats struct {
	TotalSegments  int64 `json:"total_segments"`
	TotalRoads     int64 `json:"total_roads"`
	VerifiedCount  int64 `json:"verified_count"`
	CountyCoverage int64 `json:"county_coverage"`
}
