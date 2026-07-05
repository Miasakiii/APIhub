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
		c.Header("Access-Control-Max-Age", "86400") // 24h preflight cache
		// Only enable credentials for trusted origins.
		// When using a wildcard ("*"), credentials must be false.
		if origin != "*" {
			c.Header("Access-Control-Allow-Credentials", "true")
		}
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// SecurityHeaders adds recommended security HTTP response headers.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("X-Permitted-Cross-Domain-Policies", "none")
		c.Header("X-Download-Options", "noopen")
		c.Next()
	}
}
