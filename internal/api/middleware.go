package api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kamau/speed/internal/db"
)

func OptionalAuth(queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.Next()
			return
		}

		idToken := strings.TrimPrefix(auth, "Bearer ")
		info, err := verifyGoogleIDToken(idToken)
		if err != nil {
			c.Next()
			return
		}

		user, err := queries.GetUserByGoogleID(c.Request.Context(), info.Sub)
		if err != nil {
			c.Next()
			return
		}

		c.Set("user_id", user.ID)
		c.Next()
	}
}

func RequireAuth(queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(401, gin.H{"error": "authorization required"})
			c.Abort()
			return
		}

		idToken := strings.TrimPrefix(auth, "Bearer ")
		info, err := verifyGoogleIDToken(idToken)
		if err != nil {
			c.JSON(401, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		user, err := queries.GetUserByGoogleID(c.Request.Context(), info.Sub)
		if err != nil {
			c.JSON(401, gin.H{"error": "user not found"})
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Next()
	}
}
