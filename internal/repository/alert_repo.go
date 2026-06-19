package repository

import (
	"apihub/internal/model"
	"database/sql"
	"time"
)

// AlertRepo handles alert database operations.
type AlertRepo struct {
	db *sql.DB
}

// NewAlertRepo creates a new AlertRepo.
func NewAlertRepo(db *sql.DB) *AlertRepo {
	return &AlertRepo{db: db}
}

// List returns all alerts ordered by creation date.
func (r *AlertRepo) List() ([]model.Alert, error) {
	rows, err := r.db.Query(`
		SELECT id, name, type, provider_id, api_key_id, threshold, unit, enabled, last_triggered_at, created_at
		FROM alerts ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
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
	return alerts, nil
}

// Create inserts a new alert.
func (r *AlertRepo) Create(a model.Alert) error {
	_, err := r.db.Exec(`
		INSERT INTO alerts (id, name, type, provider_id, api_key_id, threshold, unit, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, a.ID, a.Name, a.Type, a.ProviderID, a.APIKeyID, a.Threshold, a.Unit, 1)
	return err
}

// Update updates an existing alert.
func (r *AlertRepo) Update(id string, a model.Alert) error {
	var enabled int
	if a.Enabled {
		enabled = 1
	}
	_, err := r.db.Exec(`
		UPDATE alerts SET name=?, type=?, provider_id=?, api_key_id=?, threshold=?, unit=?, enabled=?
		WHERE id=?
	`, a.Name, a.Type, a.ProviderID, a.APIKeyID, a.Threshold, a.Unit, enabled, id)
	return err
}

// Delete removes an alert by ID.
func (r *AlertRepo) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM alerts WHERE id=?", id)
	return err
}

// ListHistory returns alert history, optionally filtered by alert ID.
func (r *AlertRepo) ListHistory(alertID string) ([]model.AlertHistory, error) {
	var query string
	var args []any
	if alertID != "" {
		query = `SELECT id, alert_id, message, level, created_at FROM alert_history WHERE alert_id = ? ORDER BY created_at DESC LIMIT 100`
		args = []any{alertID}
	} else {
		query = `SELECT id, alert_id, message, level, created_at FROM alert_history ORDER BY created_at DESC LIMIT 100`
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
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
	return history, nil
}
