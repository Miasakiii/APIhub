package api

import (
	"github.com/gin-gonic/gin"
)

// CORSMiddleware sets CORS headers based on auth configuration.
func CORSMiddleware(cfg AuthConfig) gin.HandlerFunc {
	origin := cfg.CORSOrigin
	if origin == "" {
		origin = "http://localhost:5173"
	}

	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
