package repository

import (
	"apihub/internal/model"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"
)

// SubscriptionRepo handles subscription database operations.
type SubscriptionRepo struct {
	db *sql.DB
}

// NewSubscriptionRepo creates a new SubscriptionRepo.
func NewSubscriptionRepo(db *sql.DB) *SubscriptionRepo {
	return &SubscriptionRepo{db: db}
}

// List returns all subscriptions with provider names.
func (r *SubscriptionRepo) List() ([]model.Subscription, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.provider_id, s.plan_name, s.price, s.currency, s.billing_cycle,
		       s.quota_type, s.quota_total, s.quota_used, s.start_date, s.renew_date,
		       s.auto_renew, s.status, s.source, s.notes, s.created_at, s.updated_at,
		       p.name as provider_name
		FROM subscriptions s
		LEFT JOIN providers p ON s.provider_id = p.id
		ORDER BY s.renew_date ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		s, err := scanSubscription(rows)
		if err != nil {
			continue
		}
		subs = append(subs, s)
	}
	if subs == nil {
		subs = []model.Subscription{}
	}
	return subs, nil
}

// GetByID returns a single subscription by ID.
func (r *SubscriptionRepo) GetByID(id string) (*model.Subscription, error) {
	var s model.Subscription
	var startDate, renewDate sql.NullString
	var createdAt, updatedAt sql.NullString
	var providerName string
	var autoRenew int

	err := r.db.QueryRow(`
		SELECT s.id, s.provider_id, s.plan_name, s.price, s.currency, s.billing_cycle,
		       s.quota_type, s.quota_total, s.quota_used, s.start_date, s.renew_date,
		       s.auto_renew, s.status, s.source, s.notes, s.created_at, s.updated_at,
		       p.name as provider_name
		FROM subscriptions s
		LEFT JOIN providers p ON s.provider_id = p.id
		WHERE s.id = ?
	`, id).Scan(&s.ID, &s.ProviderID, &s.PlanName, &s.Price, &s.Currency, &s.BillingCycle,
		&s.QuotaType, &s.QuotaTotal, &s.QuotaUsed, &startDate, &renewDate,
		&autoRenew, &s.Status, &s.Source, &s.Notes, &createdAt, &updatedAt, &providerName)
	if err != nil {
		return nil, err
	}

	s.AutoRenew = autoRenew == 1
	s.StartDate = parseDatePtr(startDate)
	s.RenewDate = parseDatePtr(renewDate)
	if createdAt.Valid {
		s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
	}
	if updatedAt.Valid {
		s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt.String)
	}
	s.Provider = &model.Provider{Name: providerName}
	return &s, nil
}

// Create inserts a new subscription.
func (r *SubscriptionRepo) Create(s model.Subscription) error {
	_, err := r.db.Exec(`
		INSERT INTO subscriptions (id, provider_id, plan_name, price, currency, billing_cycle,
			quota_type, quota_total, quota_used, start_date, renew_date, auto_renew, status, source, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, s.ID, s.ProviderID, s.PlanName, s.Price, s.Currency, s.BillingCycle,
		s.QuotaType, s.QuotaTotal, s.QuotaUsed,
		toDateStr(s.StartDate), toDateStr(s.RenewDate),
		boolToInt(s.AutoRenew), s.Status, s.Source, s.Notes)
	return err
}

// Update updates an existing subscription.
func (r *SubscriptionRepo) Update(id string, s model.Subscription) error {
	_, err := r.db.Exec(`
		UPDATE subscriptions SET provider_id=?, plan_name=?, price=?, currency=?, billing_cycle=?,
			quota_type=?, quota_total=?, quota_used=?, start_date=?, renew_date=?, auto_renew=?, status=?, source=?, notes=?
		WHERE id=?
	`, s.ProviderID, s.PlanName, s.Price, s.Currency, s.BillingCycle,
		s.QuotaType, s.QuotaTotal, s.QuotaUsed,
		toDateStr(s.StartDate), toDateStr(s.RenewDate),
		boolToInt(s.AutoRenew), s.Status, s.Source, s.Notes, id)
	return err
}

// Delete removes a subscription by ID.
func (r *SubscriptionRepo) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM subscriptions WHERE id=?", id)
	return err
}

// UpsertAutoSubscription creates or updates an auto-detected subscription for a provider.
// Matches by (provider_id, source='auto'). Updates quota and currency if found.
func (r *SubscriptionRepo) UpsertAutoSubscription(providerID, planName, currency string, quotaTotal, quotaUsed float64) error {
	// Try to find existing auto subscription
	var existingID string
	err := r.db.QueryRow(
		"SELECT id FROM subscriptions WHERE provider_id = ? AND source = 'auto' LIMIT 1",
		providerID,
	).Scan(&existingID)

	if err == nil {
		// Exists — update
		_, err = r.db.Exec(`
			UPDATE subscriptions
			SET plan_name=?, currency=?, quota_total=?, quota_used=?, updated_at=CURRENT_TIMESTAMP
			WHERE id=?
		`, planName, currency, quotaTotal, quotaUsed, existingID)
		return err
	}

	// Not found — create
	id := genSubID()
	_, err = r.db.Exec(`
		INSERT INTO subscriptions (id, provider_id, plan_name, price, currency, billing_cycle,
			quota_type, quota_total, quota_used, status, source)
		VALUES (?, ?, ?, 0, ?, 'pay-as-go', 'credits', ?, ?, 'active', 'auto')
	`, id, providerID, planName, currency, quotaTotal, quotaUsed)
	return err
}

// GetExpiring returns active subscriptions with renew_date within the given number of days.
func (r *SubscriptionRepo) GetExpiring(withinDays int) ([]model.Subscription, error) {
	rows, err := r.db.Query(`
		SELECT s.id, s.provider_id, s.plan_name, s.price, s.currency, s.billing_cycle,
		       s.quota_type, s.quota_total, s.quota_used, s.start_date, s.renew_date,
		       s.auto_renew, s.status, s.source, s.notes, s.created_at, s.updated_at,
		       p.name as provider_name
		FROM subscriptions s
		LEFT JOIN providers p ON s.provider_id = p.id
		WHERE s.status = 'active' AND s.renew_date IS NOT NULL
		  AND s.renew_date <= date('now', '+' || ? || ' days')
		  AND s.renew_date >= date('now')
		ORDER BY s.renew_date ASC
	`, withinDays)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		s, err := scanSubscription(rows)
		if err != nil {
			continue
		}
		subs = append(subs, s)
	}
	return subs, nil
}

func scanSubscription(rows *sql.Rows) (model.Subscription, error) {
	var s model.Subscription
	var startDate, renewDate sql.NullString
	var createdAt, updatedAt, notes sql.NullString
	var providerName string
	var autoRenew int
	if err := rows.Scan(&s.ID, &s.ProviderID, &s.PlanName, &s.Price, &s.Currency, &s.BillingCycle,
		&s.QuotaType, &s.QuotaTotal, &s.QuotaUsed, &startDate, &renewDate,
		&autoRenew, &s.Status, &s.Source, &notes, &createdAt, &updatedAt, &providerName); err != nil {
		return s, err
	}
	s.AutoRenew = autoRenew == 1
	s.Notes = notes.String
	s.StartDate = parseDatePtr(startDate)
	s.RenewDate = parseDatePtr(renewDate)
	if createdAt.Valid {
		s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
	}
	if updatedAt.Valid {
		s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt.String)
	}
	s.Provider = &model.Provider{Name: providerName}
	return s, nil
}

func parseDatePtr(ns sql.NullString) *time.Time {
	if !ns.Valid {
		return nil
	}
	t, err := time.Parse("2006-01-02", ns.String)
	if err != nil {
		return nil
	}
	return &t
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

func genSubID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
