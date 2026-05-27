package api

import (
	"apihub/internal/crypto"
	"apihub/internal/model"
	"apihub/internal/syncer"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Register sets up all HTTP routes.
func Register(r *gin.Engine, db *sql.DB, store *crypto.Store, cfg AuthConfig) {
	api := r.Group("/api/v1")
	authMW := OptionalAuthMiddleware(cfg)
	sensitiveMW := SensitiveAuthMiddleware(cfg)

	auth := api.Group("/auth")
	RegisterAuth(auth, db, cfg)

	protected := api.Group("")
	protected.Use(authMW)
	registerProviders(protected.Group("/providers"), db)
	registerKeys(protected.Group("/keys"), db, store, sensitiveMW)
	registerUsage(protected.Group("/usage"), db)
	registerStats(protected.Group("/stats"), db)
	RegisterAlerts(protected.Group("/alerts"), db)
	RegisterSubscriptions(protected.Group("/subscriptions"), db)
	RegisterFrequency(protected.Group("/frequency"), db)
	RegisterExport(protected.Group("/export"), db)
	RegisterProviderDetail(protected.Group("/providers"), db)
	RegisterPlayground(protected.Group("/playground"), db, store, sensitiveMW)
	RegisterWebhook(protected.Group("/webhooks"), db)
}

// RegisterSyncRoutes registers sync endpoints on the protected group.
func RegisterSyncRoutes(g *gin.RouterGroup, db *sql.DB, registry *syncer.Registry, mgr *syncer.Manager, store *crypto.Store, trigger CCSwitchTrigger) {
	RegisterSync(g, db, registry, mgr, store)
	g.POST("/sync/ccswitch", TriggerCCSwitchSync(trigger))
}

func registerProviders(g *gin.RouterGroup, db *sql.DB) {
	g.GET("", func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT id, name, type, base_url, console_url, topup_url, docs_url, COALESCE(syncer, ''), enabled, created_at, updated_at
			FROM providers ORDER BY created_at DESC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var providers []model.Provider
		for rows.Next() {
			var p model.Provider
			var enabled int
			var baseURL, consoleURL, topupURL, docsURL, createdAt, updatedAt sql.NullString
			if err := rows.Scan(&p.ID, &p.Name, &p.Type, &baseURL, &consoleURL,
				&topupURL, &docsURL, &p.Syncer, &enabled, &createdAt, &updatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
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
			p.Enabled = enabled == 1
			if createdAt.Valid {
				p.CreatedAt, _ = parseTime(createdAt.String)
			}
			if updatedAt.Valid {
				p.UpdatedAt, _ = parseTime(updatedAt.String)
			}
			providers = append(providers, p)
		}
		if providers == nil {
			providers = []model.Provider{}
		}
		c.JSON(http.StatusOK, providers)
	})

	g.POST("", func(c *gin.Context) {
		var req model.Provider
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		req.ID = generateID()
		req.Enabled = true

		_, err := db.Exec(`
			INSERT INTO providers (id, name, type, base_url, console_url, topup_url, docs_url, api_key, syncer, enabled)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 1)
		`, req.ID, req.Name, req.Type, req.BaseURL, req.ConsoleURL, req.TopUpURL, req.DocsURL, req.Syncer, "")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, req)
	})

	g.DELETE("/:id", func(c *gin.Context) {
		_, err := db.Exec("DELETE FROM providers WHERE id = ?", c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})
}

func registerKeys(g *gin.RouterGroup, db *sql.DB, store *crypto.Store, sensitiveMW gin.HandlerFunc) {
	g.POST("", func(c *gin.Context) {
		var req struct {
			ProviderID string `json:"provider_id" binding:"required"`
			Key        string `json:"key" binding:"required"`
			Name       string `json:"name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		kh := crypto.KeyHash([]byte(req.Key))
		var existing int
		if err := db.QueryRow("SELECT COUNT(*) FROM api_keys WHERE key_hash = ?", kh).Scan(&existing); err == nil && existing > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "key already exists"})
			return
		}

		encrypted, err := store.Encrypt([]byte(req.Key))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		id := generateID()
		_, err = db.Exec(`
			INSERT INTO api_keys (id, provider_id, key_hash, key_encrypted, name, source, status)
			VALUES (?, ?, ?, ?, ?, 'manual', 'active')
		`, id, req.ProviderID, kh, encrypted, req.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"id": id, "key_hash": kh, "name": req.Name, "source": "manual", "status": "active"})
	})

	g.GET("", func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT id, provider_id, key_hash, name, source, status, balance_usd, last_checked, created_at
			FROM api_keys ORDER BY created_at DESC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var keys []model.APIKey
		for rows.Next() {
			var k model.APIKey
			var lastChecked, createdAt sql.NullString
			if err := rows.Scan(&k.ID, &k.ProviderID, &k.KeyHash, &k.Name, &k.Source,
				&k.Status, &k.BalanceUSD, &lastChecked, &createdAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if lastChecked.Valid {
				t, _ := parseTime(lastChecked.String)
				k.LastChecked = &t
			}
			if createdAt.Valid {
				k.CreatedAt, _ = parseTime(createdAt.String)
			}
			keys = append(keys, k)
		}
		if keys == nil {
			keys = []model.APIKey{}
		}
		c.JSON(http.StatusOK, keys)
	})

	g.GET("/:id/decrypt", sensitiveMW, func(c *gin.Context) {
		var encrypted []byte
		if err := db.QueryRow("SELECT key_encrypted FROM api_keys WHERE id = ?", c.Param("id")).Scan(&encrypted); err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
			return
		}
		plain, err := store.Decrypt(encrypted)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "decrypt failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"key": string(plain)})
	})

	g.POST("/:id/revoke", func(c *gin.Context) {
		db.Exec("UPDATE api_keys SET status = 'revoked' WHERE id = ?", c.Param("id"))
		c.Status(http.StatusNoContent)
	})

	g.DELETE("/:id", func(c *gin.Context) {
		db.Exec("DELETE FROM api_keys WHERE id = ?", c.Param("id"))
		c.Status(http.StatusNoContent)
	})
}

func registerUsage(g *gin.RouterGroup, db *sql.DB) {
	g.GET("", func(c *gin.Context) {
		// Parse pagination params
		page := 1
		pageSize := 50
		if p, err := strconv.Atoi(c.Query("page")); err == nil && p > 0 {
			page = p
		}
		if ps, err := strconv.Atoi(c.Query("page_size")); err == nil && ps > 0 && ps <= 200 {
			pageSize = ps
		}
		offset := (page - 1) * pageSize

		var args []any
		conditions := []string{}
		if pid := c.Query("provider_id"); pid != "" {
			conditions = append(conditions, "provider_id = ?")
			args = append(args, pid)
		}
		if model := c.Query("model"); model != "" {
			conditions = append(conditions, "model = ?")
			args = append(args, model)
		}
		if source := c.Query("source"); source != "" {
			conditions = append(conditions, "source = ?")
			args = append(args, source)
		}
		if date := c.Query("date"); date != "" {
			conditions = append(conditions, "DATE(timestamp) = ?")
			args = append(args, date)
		}

		where := ""
		if len(conditions) > 0 {
			where = " WHERE " + join(" AND ", conditions)
		}

		// Get total count
		var total int
		countQ := "SELECT COUNT(*) FROM usage_records" + where
		if err := db.QueryRow(countQ, args...).Scan(&total); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get paginated records
		q := `SELECT id, api_key_id, provider_id, model, input_tokens, output_tokens,
			       cache_read, cache_create, cost_usd, source, timestamp, created_at
			  FROM usage_records` + where + ` ORDER BY timestamp DESC LIMIT ? OFFSET ?`

		rows, err := db.Query(q, append(args, pageSize, offset)...)
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
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if apiKeyID.Valid {
				r.APIKeyID = apiKeyID.String
			}
			if tsStr.Valid {
				r.Timestamp, _ = parseTime(tsStr.String)
			}
			if createdStr.Valid {
				r.CreatedAt, _ = parseTime(createdStr.String)
			}
			records = append(records, r)
		}
		if records == nil {
			records = []model.UsageRecord{}
		}
		c.JSON(http.StatusOK, gin.H{
			"records":   records,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		})
	})

	g.GET("/summary", func(c *gin.Context) {
		var summary struct {
			TotalCost     float64 `json:"total_cost_usd"`
			TotalTokens   int64   `json:"total_tokens"`
			TotalRequests int64   `json:"total_requests"`
			UniqueModels  int64   `json:"unique_models"`
			UniqueKeys    int64   `json:"unique_keys"`
		}
		db.QueryRow(`
			SELECT COALESCE(SUM(cost_usd), 0),
			       COALESCE(SUM(input_tokens + output_tokens + cache_read + cache_create), 0),
			       COUNT(*)
			FROM usage_records
		`).Scan(&summary.TotalCost, &summary.TotalTokens, &summary.TotalRequests)

		db.QueryRow(`SELECT COUNT(DISTINCT model) FROM usage_records`).Scan(&summary.UniqueModels)
		db.QueryRow(`SELECT COUNT(*) FROM api_keys WHERE status = 'active'`).Scan(&summary.UniqueKeys)

		c.JSON(http.StatusOK, summary)
	})
}

func registerStats(g *gin.RouterGroup, db *sql.DB) {
	g.GET("/daily", func(c *gin.Context) {
		q := `SELECT id, provider_id, model, source, date, request_count,
			         input_tokens, output_tokens, cache_read, cache_create, cost_usd, updated_at
			  FROM daily_stats ORDER BY date DESC LIMIT 30`

		var args []any
		conditions := []string{}
		if pid := c.Query("provider_id"); pid != "" {
			conditions = append(conditions, "provider_id = ?")
			args = append(args, pid)
		}
		if model := c.Query("model"); model != "" {
			conditions = append(conditions, "model = ?")
			args = append(args, model)
		}
		if len(conditions) > 0 {
			q = `SELECT id, provider_id, model, source, date, request_count,
				     input_tokens, output_tokens, cache_read, cache_create, cost_usd, updated_at
			  FROM daily_stats WHERE ` + join(" AND ", conditions) + ` ORDER BY date DESC LIMIT 30`
		}

		rows, err := db.Query(q, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var stats []model.DailyStats
		for rows.Next() {
			var s model.DailyStats
			var updatedStr sql.NullString
			if err := rows.Scan(&s.ID, &s.ProviderID, &s.Model, &s.Source, &s.Date,
				&s.RequestCount, &s.InputTokens, &s.OutputTokens, &s.CacheRead, &s.CacheCreate,
				&s.CostUSD, &updatedStr); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if updatedStr.Valid {
				s.UpdatedAt, _ = parseTime(updatedStr.String)
			}
			stats = append(stats, s)
		}
		if stats == nil {
			stats = []model.DailyStats{}
		}
		c.JSON(http.StatusOK, stats)
	})

	g.GET("/cost-trend", func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT date, SUM(cost_usd) as cost, SUM(input_tokens + output_tokens) as tokens, SUM(request_count) as req_count
			FROM daily_stats
			GROUP BY date
			ORDER BY date DESC
			LIMIT 30
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		type TrendRow struct {
			Date    string  `json:"date"`
			Cost    float64 `json:"cost_usd"`
			Tokens  int64   `json:"tokens"`
			ReqCount int64  `json:"request_count"`
		}
		var result []TrendRow
		for rows.Next() {
			var t TrendRow
			rows.Scan(&t.Date, &t.Cost, &t.Tokens, &t.ReqCount)
			result = append(result, t)
		}
		if result == nil {
			result = []TrendRow{}
		}
		c.JSON(http.StatusOK, result)
	})

	g.GET("/model-breakdown", func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT model, SUM(cost_usd) as total_cost,
			       SUM(input_tokens + output_tokens + cache_read + cache_create) as total_tokens,
			       SUM(request_count) as req_count
			FROM daily_stats
			GROUP BY model
			ORDER BY total_cost DESC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		type BreakdownRow struct {
			Model      string  `json:"model"`
			TotalCost  float64 `json:"total_cost_usd"`
			TotalTokens int64  `json:"total_tokens"`
			ReqCount   int64   `json:"request_count"`
		}
		var result []BreakdownRow
		for rows.Next() {
			var t BreakdownRow
			rows.Scan(&t.Model, &t.TotalCost, &t.TotalTokens, &t.ReqCount)
			result = append(result, t)
		}
		if result == nil {
			result = []BreakdownRow{}
		}
		c.JSON(http.StatusOK, result)
	})
}

// SyncStatus returns the sync status for all sources.
func SyncStatus(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT id, source, last_sync, offset_val, status, error, updated_at
			FROM sync_state ORDER BY updated_at DESC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var states []model.SyncState
		for rows.Next() {
			var s model.SyncState
			var lastSync, errStr, updatedAt sql.NullString
			if err := rows.Scan(&s.ID, &s.Source, &lastSync, &s.Offset, &s.Status, &errStr, &updatedAt); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if lastSync.Valid {
				t, _ := parseTime(lastSync.String)
				s.LastSync = &t
			}
			if errStr.Valid {
				s.Error = errStr.String
			}
			if updatedAt.Valid {
				s.UpdatedAt, _ = parseTime(updatedAt.String)
			}
			states = append(states, s)
		}
		if states == nil {
			states = []model.SyncState{}
		}
		c.JSON(http.StatusOK, states)
	}
}

func parseTime(s string) (time.Time, error) {
	for _, layout := range []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.000Z",
		time.RFC3339,
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time: %s", s)
}

func join(sep string, parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += sep
		}
		result += p
	}
	return result
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
