package syncer

import (
	"context"
	"time"
)

// Record represents a single usage record from a provider.
type Record struct {
	ProviderID   string
	APIKeyID     string
	Model        string
	Date         time.Time
	InputTokens  int64
	OutputTokens int64
	RequestCount int64
	CostUSD      float64
}

// BalanceInfo holds account balance information.
type BalanceInfo struct {
	Available float64
	Total     float64
	Currency  string
	UpdatedAt time.Time
}

// Syncer is the interface that all provider syncers must implement.
type Syncer interface {
	// Name returns the syncer name (e.g. "openrouter", "one-api", "new-api").
	Name() string

	// FetchUsage fetches usage records for the given time range.
	// apiKey is the decrypted API key.
	FetchUsage(ctx context.Context, apiKey string, baseURL string, from, to time.Time) ([]Record, error)

	// FetchBalance fetches current balance.
	FetchBalance(ctx context.Context, apiKey string, baseURL string) (*BalanceInfo, error)

	// ValidateKey validates if the API key is valid.
	ValidateKey(ctx context.Context, apiKey string, baseURL string) error

	// SupportsUsage returns true if this syncer supports automatic usage sync.
	SupportsUsage() bool

	// SupportsBalance returns true if this syncer supports balance queries.
	SupportsBalance() bool
}

// Registry holds all registered syncers.
type Registry struct {
	syncers map[string]Syncer
}

// NewRegistry creates a new registry with all built-in syncers.
func NewRegistry() *Registry {
	r := &Registry{syncers: make(map[string]Syncer)}
	return r
}

// Register adds a syncer to the registry.
func (r *Registry) Register(s Syncer) {
	r.syncers[s.Name()] = s
}

// Get returns the syncer for the given name.
func (r *Registry) Get(name string) (Syncer, bool) {
	s, ok := r.syncers[name]
	return s, ok
}

// Names returns all registered syncer names.
func (r *Registry) Names() []string {
	names := make([]string, 0, len(r.syncers))
	for n := range r.syncers {
		names = append(names, n)
	}
	return names
}
