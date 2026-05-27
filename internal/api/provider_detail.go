package api

import (
	"apihub/internal/model"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterProviderDetail registers provider detail endpoints.
func RegisterProviderDetail(g *gin.RouterGroup, db *sql.DB) {
	// Provider detail with keys, usage, and stats
	g.GET("/:id", func(c *gin.Context) {
		providerID := c.Param("id")

		// Get provider info
		var p model.Provider
		var enabled int
		var baseURL, consoleURL, topupURL, docsURL sql.NullString
		err := db.QueryRow(`
			SELECT id, name, type, base_url, console_url, topup_url, docs_url, enabled
			FROM providers WHERE id = ?
		`, providerID).Scan(&p.ID, &p.Name, &p.Type, &baseURL, &consoleURL, &topupURL, &docsURL, &enabled)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "provider not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		p.Enabled = enabled == 1
		if baseURL.Valid {
			p.BaseURL = baseURL.String
		}
		if consoleURL.Valid {
			p.ConsoleURL = consoleURL.String
		}
		if topupURL.Valid {
			p.TopUpURL = topupURL.String
		}
		if docsURL.Valid {
			p.DocsURL = docsURL.String
		}

		// Get keys
		var keys []model.APIKey
		keyRows, err := db.Query(`
			SELECT id, key_hash, name, source, status, balance_usd, last_checked, created_at
			FROM api_keys WHERE provider_id = ? ORDER BY created_at DESC
		`, providerID)
		if err == nil {
			defer keyRows.Close()
			for keyRows.Next() {
				var k model.APIKey
				var lastChecked sql.NullString
				var createdAt sql.NullString
				if err := keyRows.Scan(&k.ID, &k.KeyHash, &k.Name, &k.Source, &k.Status, &k.BalanceUSD, &lastChecked, &createdAt); err != nil {
					continue
				}
				if lastChecked.Valid {
					t, _ := time.Parse("2006-01-02 15:04:05", lastChecked.String)
					k.LastChecked = &t
				}
				if createdAt.Valid {
					k.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
				}
				keys = append(keys, k)
			}
		}

		// Get recent usage
		var totalCost float64
		db.QueryRow(`
			SELECT COALESCE(SUM(cost_usd), 0) FROM usage_records WHERE provider_id = ?
		`, providerID).Scan(&totalCost)

		var totalRequests int64
		db.QueryRow(`
			SELECT COUNT(*) FROM usage_records WHERE provider_id = ?
		`, providerID).Scan(&totalRequests)

		c.JSON(http.StatusOK, gin.H{
			"provider":       p,
			"keys":           keys,
			"total_cost":     totalCost,
			"total_requests": totalRequests,
		})
	})
}
