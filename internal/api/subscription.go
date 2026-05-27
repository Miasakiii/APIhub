package api

import (
	"apihub/internal/model"
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterSubscriptions registers subscription-related endpoints.
func RegisterSubscriptions(g *gin.RouterGroup, db *sql.DB) {
	// List subscriptions
	g.GET("", func(c *gin.Context) {
		rows, err := db.Query(`
			SELECT s.id, s.provider_id, s.plan_name, s.price, s.currency, s.billing_cycle,
			       s.quota_type, s.quota_total, s.quota_used, s.start_date, s.renew_date,
			       s.auto_renew, s.status, s.notes, s.created_at, s.updated_at,
			       p.name as provider_name
			FROM subscriptions s
			LEFT JOIN providers p ON s.provider_id = p.id
			ORDER BY s.renew_date ASC
		`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var subs []model.Subscription
		for rows.Next() {
			var s model.Subscription
			var startDate, renewDate sql.NullString
			var createdAt, updatedAt sql.NullString
			var providerName string
			var autoRenew int
			if err := rows.Scan(&s.ID, &s.ProviderID, &s.PlanName, &s.Price, &s.Currency, &s.BillingCycle,
				&s.QuotaType, &s.QuotaTotal, &s.QuotaUsed, &startDate, &renewDate,
				&autoRenew, &s.Status, &s.Notes, &createdAt, &updatedAt, &providerName); err != nil {
				continue
			}
			s.AutoRenew = autoRenew == 1
			if startDate.Valid {
				t, _ := time.Parse("2006-01-02", startDate.String)
				s.StartDate = &t
			}
			if renewDate.Valid {
				t, _ := time.Parse("2006-01-02", renewDate.String)
				s.RenewDate = &t
			}
			if createdAt.Valid {
				s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
			}
			if updatedAt.Valid {
				s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt.String)
			}
			s.Provider = &model.Provider{Name: providerName}
			subs = append(subs, s)
		}
		if subs == nil {
			subs = []model.Subscription{}
		}
		c.JSON(http.StatusOK, subs)
	})

	// Create subscription
	g.POST("", func(c *gin.Context) {
		var req model.Subscription
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		req.ID = generateID()
		_, err := db.Exec(`
			INSERT INTO subscriptions (id, provider_id, plan_name, price, currency, billing_cycle,
				quota_type, quota_total, quota_used, start_date, renew_date, auto_renew, status, notes)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, req.ID, req.ProviderID, req.PlanName, req.Price, req.Currency, req.BillingCycle,
			req.QuotaType, req.QuotaTotal, req.QuotaUsed,
			toDateStr(req.StartDate), toDateStr(req.RenewDate),
			boolToInt(req.AutoRenew), req.Status, req.Notes)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, req)
	})

	// Get subscription detail
	g.GET("/:id", func(c *gin.Context) {
		var s model.Subscription
		var startDate, renewDate sql.NullString
		var createdAt, updatedAt sql.NullString
		var providerName string
		var autoRenew int

		err := db.QueryRow(`
			SELECT s.id, s.provider_id, s.plan_name, s.price, s.currency, s.billing_cycle,
			       s.quota_type, s.quota_total, s.quota_used, s.start_date, s.renew_date,
			       s.auto_renew, s.status, s.notes, s.created_at, s.updated_at,
			       p.name as provider_name
			FROM subscriptions s
			LEFT JOIN providers p ON s.provider_id = p.id
			WHERE s.id = ?
		`, c.Param("id")).Scan(&s.ID, &s.ProviderID, &s.PlanName, &s.Price, &s.Currency, &s.BillingCycle,
			&s.QuotaType, &s.QuotaTotal, &s.QuotaUsed, &startDate, &renewDate,
			&autoRenew, &s.Status, &s.Notes, &createdAt, &updatedAt, &providerName)
		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		s.AutoRenew = autoRenew == 1
		if startDate.Valid {
			t, _ := time.Parse("2006-01-02", startDate.String)
			s.StartDate = &t
		}
		if renewDate.Valid {
			t, _ := time.Parse("2006-01-02", renewDate.String)
			s.RenewDate = &t
		}
		if createdAt.Valid {
			s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
		}
		if updatedAt.Valid {
			s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt.String)
		}
		s.Provider = &model.Provider{Name: providerName}
		c.JSON(http.StatusOK, s)
	})

	// Update subscription
	g.PUT("/:id", func(c *gin.Context) {
		var req model.Subscription
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		_, err := db.Exec(`
			UPDATE subscriptions SET provider_id=?, plan_name=?, price=?, currency=?, billing_cycle=?,
				quota_type=?, quota_total=?, quota_used=?, start_date=?, renew_date=?, auto_renew=?, status=?, notes=?
			WHERE id=?
		`, req.ProviderID, req.PlanName, req.Price, req.Currency, req.BillingCycle,
			req.QuotaType, req.QuotaTotal, req.QuotaUsed,
			toDateStr(req.StartDate), toDateStr(req.RenewDate),
			boolToInt(req.AutoRenew), req.Status, req.Notes, c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	// Delete subscription
	g.DELETE("/:id", func(c *gin.Context) {
		_, err := db.Exec("DELETE FROM subscriptions WHERE id=?", c.Param("id"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})
}

func toDateStr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format("2006-01-02")
	return &s
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
