package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebhookSetting represents a webhook configuration.
type WebhookSetting struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	Headers string `json:"headers"`
	Enabled bool   `json:"enabled"`
}

// RegisterWebhook registers webhook management endpoints.
func RegisterWebhook(g *gin.RouterGroup, db *sql.DB) {
	g.GET("", func(c *gin.Context) {
		rows, err := db.Query(`SELECT id, name, url, headers, enabled FROM webhook_settings`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var settings []WebhookSetting
		for rows.Next() {
			var s WebhookSetting
			var enabled int
			if err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.Headers, &enabled); err != nil {
				continue
			}
			s.Enabled = enabled == 1
			settings = append(settings, s)
		}
		if settings == nil {
			settings = []WebhookSetting{}
		}
		c.JSON(http.StatusOK, settings)
	})

	g.POST("", func(c *gin.Context) {
		var req WebhookSetting
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		req.ID = generateID()
		_, err := db.Exec(`
			INSERT INTO webhook_settings (id, name, url, headers, enabled)
			VALUES (?, ?, ?, ?, ?)
		`, req.ID, req.Name, req.URL, req.Headers, 1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, req)
	})

	g.DELETE("/:id", func(c *gin.Context) {
		_, err := db.Exec("DELETE FROM webhook_settings WHERE id = ?", c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})
}
