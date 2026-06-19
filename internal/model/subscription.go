package model

import "time"

// Subscription represents a user's subscription to a service.
type Subscription struct {
	ID           string     `json:"id"`
	ProviderID   string     `json:"provider_id"`
	PlanName     string     `json:"plan_name"`
	Price        float64    `json:"price"`
	Currency     string     `json:"currency"`
	BillingCycle string     `json:"billing_cycle"` // monthly | yearly | one-time | pay-as-go
	QuotaType    string     `json:"quota_type"`    // tokens | requests | credits | none
	QuotaTotal   float64    `json:"quota_total"`
	QuotaUsed    float64    `json:"quota_used"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	RenewDate    *time.Time `json:"renew_date,omitempty"`
	AutoRenew    bool       `json:"auto_renew"`
	Status       string     `json:"status"` // active | expired | cancelled | paused
	Source       string     `json:"source"`  // manual | auto
	Notes        string     `json:"notes,omitempty"`
	CreatedAt    time.Time  `json:"created_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at,omitempty"`

	Provider *Provider `json:"provider,omitempty"`
}
