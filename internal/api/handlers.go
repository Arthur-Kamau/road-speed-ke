package api

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kamau/speed/internal/db"
	"github.com/kamau/speed/internal/models"
)

type Handler struct {
	queries *db.Queries
}

func NewHandler(queries *db.Queries) *Handler {
	return &Handler{queries: queries}
}

func (h *Handler) GetSpeedsByBBox(c *gin.Context) {
	bboxStr := c.Query("bbox")
	if bboxStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bbox parameter required (minLat,minLng,maxLat,maxLng)"})
		return
	}

	parts := strings.Split(bboxStr, ",")
	if len(parts) != 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bbox must have 4 values: minLat,minLng,maxLat,maxLng"})
		return
	}

	coords := make([]float64, 4)
	for i, p := range parts {
		val, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coordinate value"})
			return
		}
		coords[i] = val
	}

	bbox := models.BBoxQuery{
		MinLat: coords[0],
		MinLng: coords[1],
		MaxLat: coords[2],
		MaxLng: coords[3],
	}

	segments, err := h.queries.GetByBBox(c.Request.Context(), bbox)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query speed limits"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"type":     "FeatureCollection",
		"features": toFeatureCollection(segments),
	})
}

func (h *Handler) GetSpeedsByRoute(c *gin.Context) {
	pointsStr := c.Query("points")
	if pointsStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "points parameter required (lat1,lng1,lat2,lng2,...)"})
		return
	}

	parts := strings.Split(pointsStr, ",")
	if len(parts)%2 != 0 || len(parts) < 4 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "points must be pairs of lat,lng (minimum 2 points)"})
		return
	}

	var points [][]float64
	for i := 0; i < len(parts); i += 2 {
		lat, err1 := strconv.ParseFloat(strings.TrimSpace(parts[i]), 64)
		lng, err2 := strconv.ParseFloat(strings.TrimSpace(parts[i+1]), 64)
		if err1 != nil || err2 != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid coordinate value"})
			return
		}
		points = append(points, []float64{lat, lng})
	}

	segments, err := h.queries.GetNearRoute(c.Request.Context(), points, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query speed limits"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"type":     "FeatureCollection",
		"features": toFeatureCollection(segments),
	})
}

func (h *Handler) GetRoadByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid road ID"})
		return
	}

	segment, err := h.queries.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "road not found"})
		return
	}

	c.JSON(http.StatusOK, toFeature(*segment))
}

func (h *Handler) GetStats(c *gin.Context) {
	stats, err := h.queries.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

func toFeature(seg models.RoadSegment) gin.H {
	return gin.H{
		"type": "Feature",
		"properties": gin.H{
			"id":              seg.ID,
			"road_name":       seg.RoadName,
			"road_class":      seg.RoadClass,
			"speed_limit_kmh": seg.SpeedLimitKmh,
			"direction":       seg.Direction,
			"source":          seg.Source,
			"verified":        seg.Verified,
			"county":          seg.County,
			"last_updated":    seg.LastUpdated,
		},
		"geometry": seg.Geometry,
	}
}

func toFeatureCollection(segments []models.RoadSegment) []gin.H {
	features := make([]gin.H, 0, len(segments))
	for _, seg := range segments {
		features = append(features, toFeature(seg))
	}
	return features
}
