package api

import (
	"apihub/internal/repository"
	"apihub/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// registerStats registers statistics endpoints.
func registerStats(g *gin.RouterGroup, svc *service.StatsService) {
	g.GET("/daily", func(c *gin.Context) {
		stats, err := svc.ListDaily(repository.DailyStatsFilter{
			ProviderID: c.Query("provider_id"),
			Model:      c.Query("model"),
			AgentID:    c.Query("agent_id"),
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, stats)
	})

	g.GET("/cost-trend", func(c *gin.Context) {
		trend, err := svc.GetCostTrend()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, trend)
	})

	g.GET("/model-breakdown", func(c *gin.Context) {
		breakdown, err := svc.GetModelBreakdown()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, breakdown)
	})
}
