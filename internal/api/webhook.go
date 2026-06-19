package api

import (
	"apihub/internal/model"
	"apihub/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterWebhook registers webhook management endpoints.
func RegisterWebhook(g *gin.RouterGroup, svc *service.WebhookService) {
	g.GET("", func(c *gin.Context) {
		settings, err := svc.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, settings)
	})

	g.POST("", func(c *gin.Context) {
		var req model.WebhookSetting
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := svc.Create(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, result)
	})

	g.DELETE("/:id", func(c *gin.Context) {
		if err := svc.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})
}
