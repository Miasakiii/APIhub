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

// APIKeyDetail contains key metadata with encrypted data for decryption.
// Used by playground and sync handlers to access provider info and encrypted key.
type APIKeyDetail struct {
	ProviderID   string
	Syncer       string
	Encrypted    []byte
	BaseURL      string
	ProviderType string
}

// Agent represents a tool/application that makes API calls (e.g. Claude Code, Cursor).
type Agent struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"` // cli, ide, api, proxy
	Icon      string    `json:"icon,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// UsageRecord represents a single API call's token usage.
type UsageRecord struct {
	ID           string    `json:"id"`
	APIKeyID     string    `json:"api_key_id,omitempty"`
	ProviderID   string    `json:"provider_id"`
	Model        string    `json:"model"`
	AgentID      string    `json:"agent_id,omitempty"`
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
	AgentID      string    `json:"agent_id,omitempty"`
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

// WebhookSetting represents a webhook configuration.
type WebhookSetting struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	Headers string `json:"headers"`
	Enabled bool   `json:"enabled"`
}

// ProviderDetail contains a provider with its keys and usage summary.
type ProviderDetail struct {
	Provider      Provider  `json:"provider"`
	Keys          []APIKey  `json:"keys"`
	TotalCost     float64   `json:"total_cost"`
	TotalRequests int64     `json:"total_requests"`
}

// ModelPricing stores per-million-token costs for a model.
type ModelPricing struct {
	ModelID               string  `json:"model_id"`
	DisplayName           string  `json:"display_name"`
	InputCostPerM         float64 `json:"input_cost_per_million"`
	OutputCostPerM        float64 `json:"output_cost_per_million"`
	CacheReadCostPerM     float64 `json:"cache_read_cost_per_million"`
	CacheCreationCostPerM float64 `json:"cache_creation_cost_per_million"`
	IsCustom              bool    `json:"is_custom"`
}

// UsageSession represents a group of consecutive API calls within a time window.
type UsageSession struct {
	ID           string    `json:"id"`
	ProviderID   string    `json:"provider_id"`
	Model        string    `json:"model"`
	Source       string    `json:"source"`
	AgentID      string    `json:"agent_id,omitempty"`
	StartedAt    time.Time `json:"started_at"`
	EndedAt      time.Time `json:"ended_at"`
	DurationMs   int64     `json:"duration_ms"`
	RequestCount int64     `json:"request_count"`
	InputTokens  int64     `json:"input_tokens"`
	OutputTokens int64     `json:"output_tokens"`
	CacheRead    int64     `json:"cache_read"`
	CacheCreate  int64     `json:"cache_create"`
	CostUSD      float64   `json:"cost_usd"`
	CreatedAt    time.Time `json:"created_at,omitempty"`
}

// KeyAuditLog records an action performed on an API key for audit trail.
type KeyAuditLog struct {
	ID        string    `json:"id"`
	KeyID     string    `json:"key_id"`
	Action    string    `json:"action"` // created, revoked, deleted, auto_imported
	Detail    string    `json:"detail,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
}

// ActivityBucket represents aggregated usage within a single hour for a provider+model.
type ActivityBucket struct {
	ID           string    `json:"id"`
	BucketStart  time.Time `json:"bucket_start"`
	ProviderID   string    `json:"provider_id"`
	Model        string    `json:"model"`
	AgentID      string    `json:"agent_id,omitempty"`
	RequestCount int64     `json:"request_count"`
	InputTokens  int64     `json:"input_tokens"`
	OutputTokens int64     `json:"output_tokens"`
	CacheRead    int64     `json:"cache_read"`
	CacheCreate  int64     `json:"cache_create"`
	CostUSD      float64   `json:"cost_usd"`
	UpdatedAt    time.Time `json:"updated_at,omitempty"`
}
