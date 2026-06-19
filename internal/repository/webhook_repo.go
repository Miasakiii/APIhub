package repository

import (
	"apihub/internal/model"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
)

// WebhookRepo handles webhook settings database operations.
type WebhookRepo struct {
	db *sql.DB
}

// NewWebhookRepo creates a new WebhookRepo.
func NewWebhookRepo(db *sql.DB) *WebhookRepo {
	return &WebhookRepo{db: db}
}

// List returns all webhook settings.
func (r *WebhookRepo) List() ([]model.WebhookSetting, error) {
	rows, err := r.db.Query(`SELECT id, name, url, headers, enabled FROM webhook_settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []model.WebhookSetting
	for rows.Next() {
		var s model.WebhookSetting
		var enabled int
		if err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.Headers, &enabled); err != nil {
			return nil, err
		}
		s.Enabled = enabled == 1
		settings = append(settings, s)
	}
	if settings == nil {
		settings = []model.WebhookSetting{}
	}
	return settings, nil
}

// Create inserts a new webhook setting.
func (r *WebhookRepo) Create(s model.WebhookSetting) (model.WebhookSetting, error) {
	s.ID = generateID()
	_, err := r.db.Exec(`
		INSERT INTO webhook_settings (id, name, url, headers, enabled)
		VALUES (?, ?, ?, ?, ?)
	`, s.ID, s.Name, s.URL, s.Headers, boolToInt(s.Enabled))
	return s, err
}

// Delete removes a webhook setting by ID.
func (r *WebhookRepo) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM webhook_settings WHERE id = ?", id)
	return err
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
