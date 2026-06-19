package api

import (
	"apihub/internal/repository"
	"apihub/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// registerUsage registers usage record endpoints.
func registerUsage(g *gin.RouterGroup, svc *service.UsageService) {
	g.GET("", func(c *gin.Context) {
		page := 1
		pageSize := 50
		if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
			page = p
		}
		if ps, err := strconv.Atoi(c.Query("page_size")); err == nil && ps > 0 && ps <= 200 {
			pageSize = ps
		}

		result, err := svc.List(repository.UsageFilter{
			ProviderID: c.Query("provider_id"),
			Model:      c.Query("model"),
			Source:     c.Query("source"),
			AgentID:    c.Query("agent_id"),
			Date:       c.Query("date"),
			Page:       page,
			PageSize:   pageSize,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})

	g.GET("/summary", func(c *gin.Context) {
		summary, err := svc.GetSummary()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, summary)
	})
}
