package model

import "time"

// Alert represents a user-defined alert rule.
type Alert struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Type            string     `json:"type"` // balance_low | quota_low | subscription_expiring | key_expired | abnormal_frequency
	ProviderID      string     `json:"provider_id,omitempty"`
	APIKeyID        string     `json:"api_key_id,omitempty"`
	Threshold       float64    `json:"threshold"`
	Unit            string     `json:"unit,omitempty"` // usd | percent | days | count
	Enabled         bool       `json:"enabled"`
	LastTriggeredAt *time.Time `json:"last_triggered_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at,omitempty"`
}

// AlertHistory represents a triggered alert event.
type AlertHistory struct {
	ID        string    `json:"id"`
	AlertID   string    `json:"alert_id"`
	Message   string    `json:"message"`
	Level     string    `json:"level"` // info | warning | critical
	CreatedAt time.Time `json:"created_at,omitempty"`
}
