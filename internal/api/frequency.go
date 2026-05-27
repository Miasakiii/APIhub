package api

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterFrequency registers frequency-related endpoints.
func RegisterFrequency(g *gin.RouterGroup, db *sql.DB) {
	// Hourly heatmap data (last 7 days)
	g.GET("/hourly", func(c *gin.Context) {
		days := 7
		if d := c.Query("days"); d != "" {
			if n, err := strconv.Atoi(d); err == nil && n > 0 {
				days = n
			}
		}

		from := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

		rows, err := db.Query(`
			SELECT
				strftime('%w', timestamp) as day_of_week,
				strftime('%H', timestamp) as hour,
				COUNT(*) as count
			FROM usage_records
			WHERE timestamp > ?
			GROUP BY day_of_week, hour
			ORDER BY day_of_week, hour
		`, from)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		// Initialize 7x24 grid
		heatmap := make([][]int64, 7)
		for i := range heatmap {
			heatmap[i] = make([]int64, 24)
		}

		for rows.Next() {
			var dowStr, hourStr string
			var count int64
			if err := rows.Scan(&dowStr, &hourStr, &count); err != nil {
				continue
			}
			dow, _ := strconv.Atoi(dowStr)
			hour, _ := strconv.Atoi(hourStr)
			if dow >= 0 && dow < 7 && hour >= 0 && hour < 24 {
				heatmap[dow][hour] = count
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"heatmap": heatmap,
			"days":    days,
		})
	})

	// Peak QPS (queries per second) for the last N days
	g.GET("/peak-qps", func(c *gin.Context) {
		days := 1
		if d := c.Query("days"); d != "" {
			if n, err := strconv.Atoi(d); err == nil && n > 0 {
				days = n
			}
		}

		from := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

		// Peak QPS per minute
		rows, err := db.Query(`
			SELECT
				strftime('%Y-%m-%d %H:%M', timestamp) as minute,
				COUNT(*) as count
			FROM usage_records
			WHERE timestamp > ?
			GROUP BY minute
			ORDER BY count DESC
			LIMIT 1
		`, from)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var peakMinute string
		var peakCount int64
		if rows.Next() {
			rows.Scan(&peakMinute, &peakCount)
		}

		// Average QPS
		var avgQPS float64
		db.QueryRow(`
			SELECT CAST(COUNT(*) AS REAL) / (3600.0 * ?)
			FROM usage_records
			WHERE timestamp > ?
		`, days, from).Scan(&avgQPS)

		c.JSON(http.StatusOK, gin.H{
			"peak_qps":     float64(peakCount) / 60.0,
			"peak_minute":  peakMinute,
			"avg_qps":      avgQPS,
			"peak_count":   peakCount,
		})
	})

	// Hourly distribution for today
	g.GET("/today", func(c *gin.Context) {
		now := time.Now()
		from := now.Truncate(24 * time.Hour).Format("2006-01-02")

		rows, err := db.Query(`
			SELECT
				strftime('%H', timestamp) as hour,
				COUNT(*) as count
			FROM usage_records
			WHERE timestamp > ?
			GROUP BY hour
			ORDER BY hour
		`, from)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		hourly := make([]int64, 24)
		for rows.Next() {
			var hourStr string
			var count int64
			if err := rows.Scan(&hourStr, &count); err != nil {
				continue
			}
			hour, _ := strconv.Atoi(hourStr)
			if hour >= 0 && hour < 24 {
				hourly[hour] = count
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"hourly": hourly,
			"date":   from,
		})
	})
}
