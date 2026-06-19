package syncer

import (
	"apihub/internal/crypto"
	"apihub/internal/ws"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Manager orchestrates syncer runs with rate limiting and retry logic.
type Manager struct {
	db       *sql.DB
	registry *Registry
	store    *crypto.Store
	client   *http.Client
	hub      *ws.Hub
}

// NewManager creates a new sync manager.
func NewManager(db *sql.DB, registry *Registry, store *crypto.Store) *Manager {
	return &Manager{
		db:       db,
		registry: registry,
		store:    store,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetHub sets the WebSocket hub for real-time sync broadcasting.
func (m *Manager) SetHub(h *ws.Hub) {
	m.hub = h
}

// SyncProvider runs sync for a single provider's active keys.
func (m *Manager) SyncProvider(ctx context.Context, providerID string, from, to time.Time) error {
	// Resolve syncer name from provider's syncer field, then fall back to type.
	syncerName, err := m.resolveSyncerName(providerID)
	if err != nil {
		return err
	}
	syncer, ok := m.registry.Get(syncerName)
	if !ok {
		return fmt.Errorf("no syncer registered for provider %s (syncer=%s)", providerID, syncerName)
	}

	// Fetch active keys for this provider
	rows, err := m.db.QueryContext(ctx,
		"SELECT id, key_encrypted, name FROM api_keys WHERE provider_id = ? AND status = 'active'",
		providerID)
	if err != nil {
		return fmt.Errorf("fetch keys: %w", err)
	}
	defer rows.Close()

	var keys []struct {
		ID        string
		Encrypted []byte
		Name      string
	}
	for rows.Next() {
		var k struct {
			ID        string
			Encrypted []byte
			Name      string
		}
		if err := rows.Scan(&k.ID, &k.Encrypted, &k.Name); err != nil {
			continue
		}
		keys = append(keys, k)
	}

	if len(keys) == 0 {
		return nil
	}

	// Fetch provider base URL
	var baseURL string
	m.db.QueryRowContext(ctx,
		"SELECT base_url FROM providers WHERE id = ?", providerID).Scan(&baseURL)

	// Sync each key
	totalKeys := len(keys)
	for i, key := range keys {
		// Broadcast sync progress
		if m.hub != nil {
			m.hub.Broadcast(ws.NewMessage(ws.TypeSyncProgress, ws.SyncProgressData{
				ProviderID:    providerID,
				Status:        "running",
				Progress:      float64(i) / float64(totalKeys),
				ProcessedKeys: i,
				TotalKeys:     totalKeys,
			}))
		}

		plain, err := m.store.Decrypt(key.Encrypted)
		if err != nil {
			log.Printf("decrypt key %s: %v", key.ID, err)
			continue
		}
		apiKey := string(plain)

		// Validate key
		if err := syncer.ValidateKey(ctx, apiKey, baseURL); err != nil {
			log.Printf("validate key %s: %v", key.ID, err)
			m.markKeyStatus(key.ID, "invalid", err.Error())
			continue
		}

		// Fetch balance if supported
		if syncer.SupportsBalance() {
			balance, err := syncer.FetchBalance(ctx, apiKey, baseURL)
			if err != nil {
				log.Printf("fetch balance %s: %v", key.ID, err)
			} else {
				m.updateBalance(key.ID, balance)
				m.upsertSubscription(providerID, balance)
			}
		}

		// Fetch usage if supported
		if syncer.SupportsUsage() {
			records, err := syncer.FetchUsage(ctx, apiKey, baseURL, from, to)
			if err != nil {
				log.Printf("fetch usage %s: %v", key.ID, err)
				continue
			}
			if err := m.importRecords(records, key.ID, providerID); err != nil {
				log.Printf("import records %s: %v", key.ID, err)
			}
		}
	}

	// Broadcast sync complete
	if m.hub != nil {
		m.hub.Broadcast(ws.NewMessage(ws.TypeSyncComplete, ws.SyncCompleteData{
			ProviderID: providerID,
		}))
	}

	return nil
}

// importRecords writes syncer records to usage_records.
func (m *Manager) importRecords(records []Record, apiKeyID, providerID string) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := m.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO usage_records (id, api_key_id, provider_id, model, agent_id,
			input_tokens, output_tokens, cache_read, cache_create,
			cost_usd, source, timestamp, created_at)
		VALUES (?, ?, ?, ?, '', ?, ?, 0, 0, ?, 'syncer', ?, CURRENT_TIMESTAMP)
		ON CONFLICT DO NOTHING
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range records {
		id := generateID()
		if _, err := stmt.Exec(id, apiKeyID, providerID, r.Model,
			r.InputTokens, r.OutputTokens, r.CostUSD, r.Date); err != nil {
			log.Printf("insert record: %v", err)
		}
	}

	if err := upsertDailyStats(tx, records, providerID); err != nil {
		return err
	}

	return tx.Commit()
}

// upsertDailyStats aggregates records and writes to daily_stats with accumulation.
func upsertDailyStats(tx *sql.Tx, records []Record, providerID string) error {
	type aggKey struct {
		Model string
		Date  string
	}
	type aggVal struct {
		requests int64
		input    int64
		output   int64
		cost     float64
	}
	aggs := make(map[aggKey]*aggVal)

	for _, r := range records {
		date := r.Date.Format("2006-01-02")
		k := aggKey{r.Model, date}
		a, ok := aggs[k]
		if !ok {
			a = &aggVal{}
			aggs[k] = a
		}
		a.requests += r.RequestCount
		a.input += r.InputTokens
		a.output += r.OutputTokens
		a.cost += r.CostUSD
	}

	stmt, err := tx.Prepare(`
			INSERT INTO daily_stats (id, provider_id, model, source, agent_id, date,
				request_count, input_tokens, output_tokens, cache_read, cache_create, cost_usd, updated_at)
			VALUES (?, ?, ?, 'syncer', '', ?, ?, ?, ?, 0, 0, ?, CURRENT_TIMESTAMP)
			ON CONFLICT(provider_id, model, source, date)
			DO UPDATE SET
				request_count = daily_stats.request_count + excluded.request_count,
				input_tokens = daily_stats.input_tokens + excluded.input_tokens,
				output_tokens = daily_stats.output_tokens + excluded.output_tokens,
				cost_usd = daily_stats.cost_usd + excluded.cost_usd,
				updated_at = CURRENT_TIMESTAMP
		`)
		if err != nil {
			return err
		}
		defer stmt.Close()

		for k, v := range aggs {
			id := generateID()
			if _, err := stmt.Exec(id, providerID, k.Model, k.Date,
				v.requests, v.input, v.output, v.cost); err != nil {
				log.Printf("upsert daily_stats: %v", err)
			}
		}
		return nil
	}

func (m *Manager) resolveSyncerName(providerID string) (string, error) {
	var syncer, ptype string
	if err := m.db.QueryRow(
		"SELECT COALESCE(syncer, ''), type FROM providers WHERE id = ?", providerID,
	).Scan(&syncer, &ptype); err != nil {
		return "", fmt.Errorf("provider %s not found: %w", providerID, err)
	}
	if syncer != "" {
		return syncer, nil
	}
	// Fall back to provider type as syncer name (backwards compat).
	if ptype != "" {
		return ptype, nil
	}
	return "", fmt.Errorf("provider %s has no syncer configured", providerID)
}

func (m *Manager) markKeyStatus(keyID, status, errMsg string) {
	_, _ = m.db.Exec(
		"UPDATE api_keys SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, keyID)
	log.Printf("key %s marked as %s: %s", keyID, status, errMsg)
}

func (m *Manager) updateBalance(keyID string, b *BalanceInfo) {
	_, _ = m.db.Exec(
		"UPDATE api_keys SET balance_usd = ?, last_checked = CURRENT_TIMESTAMP, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		b.Available, keyID)
}

// upsertSubscription creates or updates an auto-detected subscription for a provider.
func (m *Manager) upsertSubscription(providerID string, b *BalanceInfo) {
	if b.Total <= 0 {
		return // No quota info available
	}

	// Get provider name for plan_name
	var planName string
	if err := m.db.QueryRow("SELECT name FROM providers WHERE id = ?", providerID).Scan(&planName); err != nil {
		planName = providerID
	}

	currency := b.Currency
	if currency == "" {
		currency = "USD"
	}
	quotaUsed := b.Total - b.Available
	if quotaUsed < 0 {
		quotaUsed = 0
	}

	// Try to find existing auto subscription
	var existingID string
	err := m.db.QueryRow(
		"SELECT id FROM subscriptions WHERE provider_id = ? AND source = 'auto' LIMIT 1",
		providerID,
	).Scan(&existingID)

	if err == nil {
		// Update existing
		_, _ = m.db.Exec(`
			UPDATE subscriptions
			SET plan_name=?, currency=?, quota_total=?, quota_used=?, updated_at=CURRENT_TIMESTAMP
			WHERE id=?
		`, planName, currency, b.Total, quotaUsed, existingID)
		return
	}

	// Create new
	id := generateID()
	_, _ = m.db.Exec(`
		INSERT INTO subscriptions (id, provider_id, plan_name, currency, billing_cycle,
			quota_type, quota_total, quota_used, status, source)
		VALUES (?, ?, ?, ?, 'pay-as-go', 'credits', ?, ?, 'active', 'auto')
	`, id, providerID, planName, currency, b.Total, quotaUsed)
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
