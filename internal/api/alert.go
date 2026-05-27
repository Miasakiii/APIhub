package api

import (
	"apihub/internal/model"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterAlerts registers alert-related endpoints.
func RegisterAlerts(g *gin.RouterGroup, db *sql.DB) {
	// List alert rules
	g.GET("", func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT id, name, type, provider_id, api_key_id, threshold, unit, enabled, last_triggered_at, created_at
			FROM alerts ORDER BY created_at DESC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var alerts []model.Alert
		for rows.Next() {
			var a model.Alert
			var enabled int
			var providerID, apiKeyID, lastTriggered, createdAt sql.NullString
			if err := rows.Scan(&a.ID, &a.Name, &a.Type, &providerID, &apiKeyID, &a.Threshold, &a.Unit, &enabled, &lastTriggered, &createdAt); err != nil {
				continue
			}
			if providerID.Valid {
				a.ProviderID = providerID.String
			}
			if apiKeyID.Valid {
				a.APIKeyID = apiKeyID.String
			}
			a.Enabled = enabled == 1
			if lastTriggered.Valid {
				t, _ := time.Parse("2006-01-02 15:04:05", lastTriggered.String)
				a.LastTriggeredAt = &t
			}
			if createdAt.Valid {
				a.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
			}
			alerts = append(alerts, a)
		}
		if alerts == nil {
			alerts = []model.Alert{}
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

		req.ID = generateID()
		_, err := db.Exec(`
			INSERT INTO alerts (id, name, type, provider_id, api_key_id, threshold, unit, enabled)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		`, req.ID, req.Name, req.Type, req.ProviderID, req.APIKeyID, req.Threshold, req.Unit, 1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, req)
	})

	// Update alert rule
	g.PUT("/:id", func(c *gin.Context) {
		var req model.Alert
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var enabled int
		if req.Enabled {
			enabled = 1
		}

		_, err := db.Exec(`
			UPDATE alerts SET name=?, type=?, provider_id=?, api_key_id=?, threshold=?, unit=?, enabled=?
			WHERE id=?
		`, req.Name, req.Type, req.ProviderID, req.APIKeyID, req.Threshold, req.Unit, enabled, c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	// Delete alert rule
	g.DELETE("/:id", func(c *gin.Context) {
		_, err := db.Exec("DELETE FROM alerts WHERE id=?", c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	// Get alert history
	g.GET("/history", func(c *gin.Context) {
		alertID := c.Query("alert_id")
		var query string
		var args []any
		if alertID != "" {
			query = `SELECT id, alert_id, message, level, created_at FROM alert_history WHERE alert_id = ? ORDER BY created_at DESC LIMIT 100`
			args = []any{alertID}
		} else {
			query = `SELECT id, alert_id, message, level, created_at FROM alert_history ORDER BY created_at DESC LIMIT 100`
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var history []model.AlertHistory
		for rows.Next() {
			var h model.AlertHistory
			var createdAt sql.NullString
			if err := rows.Scan(&h.ID, &h.AlertID, &h.Message, &h.Level, &createdAt); err != nil {
				continue
			}
			if createdAt.Valid {
				h.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
			}
			history = append(history, h)
		}
		if history == nil {
			history = []model.AlertHistory{}
		}
		c.JSON(http.StatusOK, history)
	})
}
