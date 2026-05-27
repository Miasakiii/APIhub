package model

import "time"

// Provider represents an API provider (OpenAI, Anthropic, etc.)
type Provider struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	BaseURL    string    `json:"base_url,omitempty"`
	ConsoleURL string    `json:"console_url,omitempty"`
	TopUpURL   string    `json:"topup_url,omitempty"`
	DocsURL    string    `json:"docs_url,omitempty"`
	Syncer     string    `json:"syncer,omitempty"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at,omitempty"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}

// APIKey represents a user-added or cc-switch-imported API key.
// KeyEncrypted is excluded from JSON serialization for security.
type APIKey struct {
	ID           string     `json:"id"`
	ProviderID   string     `json:"provider_id"`
	KeyHash      string     `json:"key_hash"`
	KeyEncrypted []byte     `json:"-"`
	Name         string     `json:"name"`
	Source       string     `json:"source"`
	Status       string     `json:"status"`
	BalanceUSD   float64    `json:"balance_usd"`
	LastChecked  *time.Time `json:"last_checked,omitempty"`
	CreatedAt    time.Time  `json:"created_at,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at,omitempty"`

	Provider *Provider `json:"provider,omitempty"`
}

// UsageRecord represents a single API call's token usage.
type UsageRecord struct {
	ID           string    `json:"id"`
	APIKeyID     string    `json:"api_key_id,omitempty"`
	ProviderID   string    `json:"provider_id"`
	Model        string    `json:"model"`
	InputTokens  int64     `json:"input_tokens"`
	OutputTokens int64     `json:"output_tokens"`
	CacheRead    int64     `json:"cache_read"`
	CacheCreate  int64     `json:"cache_create"`
	CostUSD      float64   `json:"cost_usd"`
	Source       string    `json:"source"`
	Timestamp    time.Time `json:"timestamp"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
}

// DailyStats is the aggregated daily statistics per provider+model+source.
type DailyStats struct {
	ID           string    `json:"id"`
	ProviderID   string    `json:"provider_id"`
	Model        string    `json:"model"`
	Source       string    `json:"source"`
	Date         string    `json:"date"`
	RequestCount int64     `json:"request_count"`
	InputTokens  int64     `json:"input_tokens"`
	OutputTokens int64     `json:"output_tokens"`
	CacheRead    int64     `json:"cache_read"`
	CacheCreate  int64     `json:"cache_create"`
	CostUSD      float64   `json:"cost_usd"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}

// SyncState tracks the last sync position for each data source.
type SyncState struct {
	ID        string    `json:"id"`
	Source    string    `json:"source"`
	LastSync  *time.Time `json:"last_sync,omitempty"`
	Offset    int64     `json:"offset"`
	Status    string    `json:"status"`
	Error     string    `json:"error,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SyncLog records each sync run's outcome.
type SyncLog struct {
	ID       string    `json:"id"`
	Source   string    `json:"source"`
	Fetched  int       `json:"fetched"`
	Inserted int       `json:"inserted"`
	Error    string    `json:"error,omitempty"`
	Duration int64     `json:"duration_ms"`
	CreatedAt time.Time `json:"created_at"`
}
