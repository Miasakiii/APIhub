package api

import (
	"apihub/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterProviderDetail registers provider detail endpoints.
func RegisterProviderDetail(g *gin.RouterGroup, svc *service.ProviderService) {
	// Provider detail with keys, usage, and stats
	g.GET("/:id", func(c *gin.Context) {
		detail, err := svc.GetDetail(c.Param("id"))
		if err != nil {
			if errors.Is(err, service.ErrProviderNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, detail)
	})
}
