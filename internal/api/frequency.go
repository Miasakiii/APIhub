package api

import (
	"apihub/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// RegisterFrequency registers frequency-related endpoints.
func RegisterFrequency(g *gin.RouterGroup, svc *service.FrequencyService) {
	// Hourly heatmap data (last 7 days)
	g.GET("/hourly", func(c *gin.Context) {
		days := 7
		if d := c.Query("days"); d != "" {
			if n, err := strconv.Atoi(d); err == nil && n > 0 {
				days = n
			}
		}

		result, err := svc.GetHourlyHeatmap(days)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})

	// Peak QPS (queries per second) for the last N days
	g.GET("/peak-qps", func(c *gin.Context) {
		days := 1
		if d := c.Query("days"); d != "" {
			if n, err := strconv.Atoi(d); err == nil && n > 0 {
				days = n
			}
		}

		result, err := svc.GetPeakQPS(days)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})

	// Hourly distribution for today
	g.GET("/today", func(c *gin.Context) {
		result, err := svc.GetTodayDistribution()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})
}
