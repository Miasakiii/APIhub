package api

import (
	"apihub/internal/model"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterExport registers export-related endpoints.
func RegisterExport(g *gin.RouterGroup, db *sql.DB) {
	// Export usage records as CSV
	g.GET("/csv", func(c *gin.Context) {
		q := `SELECT id, api_key_id, provider_id, model, input_tokens, output_tokens,
			     cache_read, cache_create, cost_usd, source, timestamp, created_at
		  FROM usage_records ORDER BY timestamp DESC LIMIT 10000`

		rows, err := db.Query(q)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var records []model.UsageRecord
		for rows.Next() {
			var r model.UsageRecord
			var apiKeyID, tsStr, createdStr sql.NullString
			if err := rows.Scan(&r.ID, &apiKeyID, &r.ProviderID, &r.Model,
				&r.InputTokens, &r.OutputTokens, &r.CacheRead, &r.CacheCreate,
				&r.CostUSD, &r.Source, &tsStr, &createdStr); err != nil {
				continue
			}
			if apiKeyID.Valid {
				r.APIKeyID = apiKeyID.String
			}
			if tsStr.Valid {
				r.Timestamp, _ = time.Parse("2006-01-02 15:04:05", tsStr.String)
			}
			if createdStr.Valid {
				r.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdStr.String)
			}
			records = append(records, r)
		}

		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=usage_%s.csv", time.Now().Format("20060102")))

		var b strings.Builder
		b.WriteString("ID,Provider,Model,Input Tokens,Output Tokens,Cache Read,Cache Create,Cost (USD),Source,Timestamp\n")
		for _, r := range records {
			b.WriteString(fmt.Sprintf("%s,%s,%s,%d,%d,%d,%d,%.6f,%s,%s\n",
				r.ID, r.ProviderID, r.Model, r.InputTokens, r.OutputTokens,
				r.CacheRead, r.CacheCreate, r.CostUSD, r.Source,
				r.Timestamp.Format("2006-01-02 15:04:05")))
		}

		c.String(http.StatusOK, b.String())
	})
}
