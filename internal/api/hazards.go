package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kamau/speed/internal/models"
)

func (h *Handler) GetHazardsByBBox(c *gin.Context) {
	bbox, ok := parseBBox(c)
	if !ok {
		return
	}

	hazards, err := h.queries.GetHazardsByBBox(c.Request.Context(), bbox)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to query hazards"})
		return
	}
	if hazards == nil {
		hazards = []models.RoadHazard{}
	}

	features := make([]gin.H, 0, len(hazards))
	for _, hz := range hazards {
		features = append(features, gin.H{
			"type": "Feature",
			"properties": gin.H{
				"id":          hz.ID,
				"hazard_type": hz.HazardType,
				"description": hz.Description,
				"verified":    hz.Verified,
				"reported_at": hz.ReportedAt,
			},
			"geometry": gin.H{
				"type":        "Point",
				"coordinates": []float64{hz.Lng, hz.Lat},
			},
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"type":     "FeatureCollection",
		"features": features,
	})
}

func (h *Handler) ReportHazard(c *gin.Context) {
	var report models.HazardReport
	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	validTypes := map[string]bool{"bump": true, "rumble_strip": true, "speed_camera": true, "pothole": true}
	if !validTypes[report.HazardType] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "hazard_type must be bump, rumble_strip, speed_camera, or pothole"})
		return
	}

	var userID *int64
	if uid, ok := c.Get("user_id"); ok {
		v := uid.(int64)
		userID = &v
	}

	id, err := h.queries.InsertHazard(c.Request.Context(), report, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save hazard report"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "hazard reported — thank you!"})
}

func (h *Handler) ReportSpeed(c *gin.Context) {
	var report models.SpeedReport
	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var userID *int64
	if uid, ok := c.Get("user_id"); ok {
		v := uid.(int64)
		userID = &v
	}

	id, err := h.queries.InsertSpeedReport(c.Request.Context(), report, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save speed report"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "speed limit reported — thank you!"})
}
