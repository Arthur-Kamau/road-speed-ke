package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kamau/speed/internal/db"
)

func NewRouter(queries *db.Queries) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: false,
	}))

	h := NewHandler(queries)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/speeds", h.GetSpeedsByBBox)
		v1.GET("/speeds/route", h.GetSpeedsByRoute)
		v1.GET("/roads/:id", h.GetRoadByID)
		v1.GET("/stats", h.GetStats)

		v1.GET("/hazards", h.GetHazardsByBBox)
		v1.POST("/hazards", OptionalAuth(queries), h.ReportHazard)
		v1.POST("/speeds/report", OptionalAuth(queries), h.ReportSpeed)

		v1.POST("/auth/google", h.GoogleAuth)
		v1.GET("/auth/me", RequireAuth(queries), h.GetMe)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
