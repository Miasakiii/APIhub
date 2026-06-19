package api

import (
	"apihub/internal/model"
	"apihub/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterAlerts registers alert-related endpoints.
func RegisterAlerts(g *gin.RouterGroup, svc *service.AlertService) {
	// List alert rules
	g.GET("", func(c *gin.Context) {
		alerts, err := svc.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, alerts)
	})

	// Create alert rule
	g.POST("", func(c *gin.Context) {
		var req model.Alert
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		alert, err := svc.Create(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, alert)
	})

	// Update alert rule
	g.PUT("/:id", func(c *gin.Context) {
		var req model.Alert
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := svc.Update(c.Param("id"), req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	// Delete alert rule
	g.DELETE("/:id", func(c *gin.Context) {
		if err := svc.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	// Get alert history
	g.GET("/history", func(c *gin.Context) {
		history, err := svc.ListHistory(c.Query("alert_id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, history)
	})
}
