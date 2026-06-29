package models

import "time"

type RoadHazard struct {
	ID          int64     `json:"id"`
	HazardType  string    `json:"hazard_type"`
	Description string    `json:"description"`
	Lat         float64   `json:"lat"`
	Lng         float64   `json:"lng"`
	Source      string    `json:"source"`
	Verified    bool      `json:"verified"`
	ReportedAt  time.Time `json:"reported_at"`
}

type HazardReport struct {
	HazardType  string  `json:"hazard_type" binding:"required"`
	Description string  `json:"description"`
	Lat         float64 `json:"lat" binding:"required"`
	Lng         float64 `json:"lng" binding:"required"`
}

type SpeedReport struct {
	RoadName      string  `json:"road_name" binding:"required"`
	SpeedLimitKmh int     `json:"speed_limit_kmh" binding:"required,min=1,max=200"`
	RoadClass     string  `json:"road_class"`
	Lat           float64 `json:"lat" binding:"required"`
	Lng           float64 `json:"lng" binding:"required"`
}
