package api

import (
	"apihub/internal/repository"
	"apihub/internal/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// registerSessions mounts session and bucket routes.
func registerSessions(rg *gin.RouterGroup, svc *service.SessionService) {
	g := rg.Group("/sessions")

	// GET /sessions — paginated session list
	g.GET("", func(c *gin.Context) {
		f := repository.SessionFilter{
			ProviderID: c.Query("provider_id"),
			Model:      c.Query("model"),
			Source:     c.Query("source"),
			AgentID:    c.Query("agent_id"),
		}
		if v := c.Query("page"); v != "" {
			f.Page, _ = strconv.Atoi(v)
		}
		if v := c.Query("page_size"); v != "" {
			f.PageSize, _ = strconv.Atoi(v)
		}
		if v := c.Query("from"); v != "" {
			if t, err := time.Parse("2006-01-02", v); err == nil {
				f.From = t
			}
		}
		if v := c.Query("to"); v != "" {
			if t, err := time.Parse("2006-01-02", v); err == nil {
				f.To = t.Add(24*time.Hour - time.Second) // end of day
			}
		}

		result, err := svc.ListSessions(f)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
	})

	// GET /sessions/stats — aggregate session statistics
	g.GET("/stats", func(c *gin.Context) {
		stats, err := svc.GetStats()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, stats)
	})

	// GET /sessions/buckets — activity bucket list
	g.GET("/buckets", func(c *gin.Context) {
		var from, to time.Time
		if v := c.Query("from"); v != "" {
			if t, err := time.Parse("2006-01-02", v); err == nil {
				from = t
			}
		}
		if v := c.Query("to"); v != "" {
			if t, err := time.Parse("2006-01-02", v); err == nil {
				to = t.Add(24*time.Hour - time.Second)
			}
		}

		buckets, err := svc.ListBuckets(from, to, c.Query("provider_id"), c.Query("model"), c.Query("agent_id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"buckets": buckets})
	})

	// GET /sessions/hourly?date=2026-06-19 — 24-hour bucket distribution
	g.GET("/hourly", func(c *gin.Context) {
		date := c.Query("date")
		if date == "" {
			date = time.Now().Format("2006-01-02")
		}

		buckets, err := svc.GetHourlyBuckets(date)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"date": date, "buckets": buckets})
	})
}
