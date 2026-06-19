package aggregator

import (
	"apihub/internal/model"
	"apihub/internal/ws"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log"
	"strings"
	"sync"
	"time"
)

const (
	// sessionWindow is the max gap between consecutive records to be considered
	// the same session. Records with the same (provider, model, source) within
	// this window are merged into one session.
	sessionWindow = 30 * time.Minute
)

// pricingCache holds model pricing data loaded from the database.
type pricingCache struct {
	mu     sync.RWMutex
	prices map[string]model.ModelPricing // keyed by normalized model_id
	loaded bool
	db     *sql.DB
}

var globalPricing = &pricingCache{}

// LoadPricing loads model_pricing into the global cache. Call once at startup.
func LoadPricing(db *sql.DB) error {
	return globalPricing.load(db)
}

func (pc *pricingCache) load(db *sql.DB) error {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	rows, err := db.Query(`SELECT model_id, display_name,
		input_cost_per_million, output_cost_per_million,
		cache_read_cost_per_million, cache_creation_cost_per_million
		FROM model_pricing`)
	if err != nil {
		return err
	}
	defer rows.Close()

	pc.prices = make(map[string]model.ModelPricing)
	for rows.Next() {
		var p model.ModelPricing
		if err := rows.Scan(&p.ModelID, &p.DisplayName,
			&p.InputCostPerM, &p.OutputCostPerM,
			&p.CacheReadCostPerM, &p.CacheCreationCostPerM); err != nil {
			return err
		}
		pc.prices[normalizeModel(p.ModelID)] = p
	}
	pc.loaded = true
	pc.db = db
	log.Printf("[aggregator] loaded %d model pricing entries", len(pc.prices))
	return nil
}

// Lookup returns the pricing for a model, trying normalized name first then prefix matching.
func (pc *pricingCache) Lookup(modelName string) (model.ModelPricing, bool) {
	pc.mu.RLock()
	defer pc.mu.RUnlock()

	if !pc.loaded {
		return model.ModelPricing{}, false
	}

	normalized := normalizeModel(modelName)

	// Exact match
	if p, ok := pc.prices[normalized]; ok {
		return p, true
	}

	// Prefix match: try stripping suffixes like "-thinking-high", "[1M]", "-ultraspeed"
	// Try progressively shorter prefixes
	parts := strings.Split(normalized, "-")
	for i := len(parts) - 1; i >= 1; i-- {
		prefix := strings.Join(parts[:i], "-")
		if p, ok := pc.prices[prefix]; ok {
			return p, true
		}
	}

	return model.ModelPricing{}, false
}

// normalizeModel strips common suffixes and normalizes a model name for pricing lookup.
func normalizeModel(name string) string {
	// Strip bracketed suffixes like [1M], [128k]
	if idx := strings.Index(name, "["); idx > 0 {
		name = name[:idx]
	}
	// Strip thinking/effort suffixes
	for _, suffix := range []string{"-thinking-high", "-thinking-xhigh", "-thinking-max", "-ultraspeed"} {
		if strings.HasSuffix(name, suffix) {
			name = name[:len(name)-len(suffix)]
			break
		}
	}
	return strings.TrimSpace(name)
}

// computeCost calculates cost in USD from token counts using the pricing cache.
// Returns 0 if no pricing is found for the model.
func computeCost(modelName string, input, output, cacheRead, cacheCreate int64) float64 {
	p, ok := globalPricing.Lookup(modelName)
	if !ok {
		return 0
	}
	cost := float64(input)/1_000_000*p.InputCostPerM +
		float64(output)/1_000_000*p.OutputCostPerM +
		float64(cacheRead)/1_000_000*p.CacheReadCostPerM +
		float64(cacheCreate)/1_000_000*p.CacheCreationCostPerM
	return cost
}

// Aggregator consumes new UsageRecords and updates DailyStats, Sessions, and Buckets.
// Uses a single goroutine + batch upserts to prevent conflicts.
type Aggregator struct {
	db    *sql.DB
	queue chan model.UsageRecord
	done  chan struct{}
	hub   *ws.Hub
}

// New creates and starts the aggregator goroutine.
func New(db *sql.DB) *Aggregator {
	a := &Aggregator{
		db:    db,
		queue: make(chan model.UsageRecord, 500),
		done:  make(chan struct{}),
	}
	go a.run()
	return a
}

// SetHub sets the WebSocket hub for real-time usage broadcasting.
func (a *Aggregator) SetHub(h *ws.Hub) {
	a.hub = h
}

// Submit pushes a record into the queue (non-blocking if queue not full).
func (a *Aggregator) Submit(r model.UsageRecord) {
	select {
	case a.queue <- r:
	default:
		a.drain()
	}
}

// SubmitBatch submits multiple records then drains.
func (a *Aggregator) SubmitBatch(records []model.UsageRecord) {
	for _, r := range records {
		a.Submit(r)
	}
	a.drain()
}

// Stop signals the aggregator to shut down and waits for pending work.
func (a *Aggregator) Stop() {
	close(a.queue)
	<-a.done
}

func (a *Aggregator) run() {
	defer close(a.done)

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case _, ok := <-a.queue:
			if !ok {
				a.drain()
				return
			}
			a.drain()
		case <-ticker.C:
			a.drain()
		}
	}
}

func (a *Aggregator) drain() {
	var batch []model.UsageRecord
	for {
		select {
		case r := <-a.queue:
			batch = append(batch, r)
			if len(batch) >= 100 {
				a.upsert(batch)
				batch = nil
			}
		default:
			if len(batch) > 0 {
				a.upsert(batch)
			}
			return
		}
	}
}

// backfillCost fills in CostUSD from model_pricing when the source didn't provide it.
func backfillCost(records []model.UsageRecord) {
	for i := range records {
		if records[i].CostUSD == 0 {
			records[i].CostUSD = computeCost(records[i].Model,
				records[i].InputTokens, records[i].OutputTokens,
				records[i].CacheRead, records[i].CacheCreate)
		}
	}
}

// truncateHour truncates a timestamp to the start of its hour.
func truncateHour(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
}

// upsertAll writes records to usage_records, then populates daily_stats,
// usage_activity_buckets, and usage_sessions in a single transaction.
func upsertAll(tx *sql.Tx, records []model.UsageRecord) error {
	// 1. Insert raw usage_records
	stmt, err := tx.Prepare(`
		INSERT INTO usage_records (id, api_key_id, provider_id, model, agent_id,
			input_tokens, output_tokens, cache_read, cache_create,
			cost_usd, source, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range records {
		apiKeyID := nullableString(r.APIKeyID)
		if _, err := stmt.Exec(r.ID, apiKeyID, r.ProviderID, r.Model, r.AgentID,
			r.InputTokens, r.OutputTokens, r.CacheRead, r.CacheCreate,
			r.CostUSD, r.Source, r.Timestamp); err != nil {
			return err
		}
	}

	// 2. Aggregate into daily_stats
	type dailyKey struct {
		ProviderID, Model, Source, Date string
	}
	dailyStats := make(map[dailyKey]*model.DailyStats)
	for _, r := range records {
		date := r.Timestamp.Format("2006-01-02")
		k := dailyKey{r.ProviderID, r.Model, r.Source, date}
		s, ok := dailyStats[k]
		if !ok {
			s = &model.DailyStats{ProviderID: k.ProviderID, Model: k.Model, Source: k.Source, Date: k.Date, AgentID: r.AgentID}
			dailyStats[k] = s
		}
		s.RequestCount++
		s.InputTokens += r.InputTokens
		s.OutputTokens += r.OutputTokens
		s.CacheRead += r.CacheRead
		s.CacheCreate += r.CacheCreate
		s.CostUSD += r.CostUSD
	}

	dailyStmt, err := tx.Prepare(`
		INSERT INTO daily_stats (id, provider_id, model, source, agent_id, date,
			request_count, input_tokens, output_tokens, cache_read, cache_create, cost_usd, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(provider_id, model, source, date)
		DO UPDATE SET
			request_count = daily_stats.request_count + excluded.request_count,
			input_tokens = daily_stats.input_tokens + excluded.input_tokens,
			output_tokens = daily_stats.output_tokens + excluded.output_tokens,
			cache_read = daily_stats.cache_read + excluded.cache_read,
			cache_create = daily_stats.cache_create + excluded.cache_create,
			cost_usd = daily_stats.cost_usd + excluded.cost_usd,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer dailyStmt.Close()

	for _, s := range dailyStats {
		s.ID = generateID()
		if _, err := dailyStmt.Exec(s.ID, s.ProviderID, s.Model, s.Source, s.AgentID, s.Date,
			s.RequestCount, s.InputTokens, s.OutputTokens, s.CacheRead, s.CacheCreate, s.CostUSD); err != nil {
			return err
		}
	}

	// 3. Aggregate into usage_activity_buckets
	type bucketKey struct {
		BucketStart string
		ProviderID  string
		Model       string
	}
	buckets := make(map[bucketKey]*model.ActivityBucket)
	for _, r := range records {
		bs := truncateHour(r.Timestamp).Format("2006-01-02T15:04:05")
		k := bucketKey{bs, r.ProviderID, r.Model}
		b, ok := buckets[k]
		if !ok {
			b = &model.ActivityBucket{ProviderID: k.ProviderID, Model: k.Model, AgentID: r.AgentID}
			buckets[k] = b
		}
		b.RequestCount++
		b.InputTokens += r.InputTokens
		b.OutputTokens += r.OutputTokens
		b.CacheRead += r.CacheRead
		b.CacheCreate += r.CacheCreate
		b.CostUSD += r.CostUSD
	}

	bucketStmt, err := tx.Prepare(`
		INSERT INTO usage_activity_buckets (id, bucket_start, provider_id, model, agent_id,
			request_count, input_tokens, output_tokens, cache_read, cache_create, cost_usd, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(bucket_start, provider_id, model)
		DO UPDATE SET
			request_count = usage_activity_buckets.request_count + excluded.request_count,
			input_tokens = usage_activity_buckets.input_tokens + excluded.input_tokens,
			output_tokens = usage_activity_buckets.output_tokens + excluded.output_tokens,
			cache_read = usage_activity_buckets.cache_read + excluded.cache_read,
			cache_create = usage_activity_buckets.cache_create + excluded.cache_create,
			cost_usd = usage_activity_buckets.cost_usd + excluded.cost_usd,
			updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		return err
	}
	defer bucketStmt.Close()

	for _, b := range buckets {
		b.ID = generateID()
		if _, err := bucketStmt.Exec(b.ID, b.BucketStart, b.ProviderID, b.Model, b.AgentID,
			b.RequestCount, b.InputTokens, b.OutputTokens, b.CacheRead, b.CacheCreate, b.CostUSD); err != nil {
			return err
		}
	}

	// 4. Session detection: find or create sessions
	return upsertSessions(tx, records)
}

// upsertSessions handles session detection for incoming records.
// For each record, it finds the most recent session with matching (provider_id, model, source, agent_id)
// where ended_at is within sessionWindow of the record's timestamp. If found, it extends
// that session; otherwise it creates a new one.
func upsertSessions(tx *sql.Tx, records []model.UsageRecord) error {
	// Sort records by timestamp to ensure correct session detection
	// (records may arrive out of order from different sources)
	sorted := make([]model.UsageRecord, len(records))
	copy(sorted, records)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Timestamp.Before(sorted[i].Timestamp) {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	// Track sessions we've already extended in this batch to avoid re-querying
	type sessionKey struct {
		ProviderID, Model, Source, AgentID string
	}
	activeSessions := make(map[sessionKey]*model.UsageSession)

	// Prepare statements
	findSessionStmt, err := tx.Prepare(`
		SELECT id, started_at, ended_at, request_count,
			input_tokens, output_tokens, cache_read, cache_create, cost_usd
		FROM usage_sessions
		WHERE provider_id = ? AND model = ? AND source = ? AND agent_id = ?
		ORDER BY ended_at DESC
		LIMIT 1
	`)
	if err != nil {
		return err
	}
	defer findSessionStmt.Close()

	updateSessionStmt, err := tx.Prepare(`
		UPDATE usage_sessions
		SET ended_at = ?, duration_ms = ?,
			request_count = request_count + ?,
			input_tokens = input_tokens + ?,
			output_tokens = output_tokens + ?,
			cache_read = cache_read + ?,
			cache_create = cache_create + ?,
			cost_usd = cost_usd + ?
		WHERE id = ?
	`)
	if err != nil {
		return err
	}
	defer updateSessionStmt.Close()

	insertSessionStmt, err := tx.Prepare(`
		INSERT INTO usage_sessions (id, provider_id, model, source, agent_id,
			started_at, ended_at, duration_ms, request_count,
			input_tokens, output_tokens, cache_read, cache_create, cost_usd)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer insertSessionStmt.Close()

	for _, r := range sorted {
		k := sessionKey{r.ProviderID, r.Model, r.Source, r.AgentID}

		// Check if we already have an active session in this batch
		if s, ok := activeSessions[k]; ok {
			gap := r.Timestamp.Sub(s.EndedAt)
			if gap <= sessionWindow {
				// Extend existing session
				s.EndedAt = r.Timestamp
				s.DurationMs = s.EndedAt.Sub(s.StartedAt).Milliseconds()
				s.RequestCount++
				s.InputTokens += r.InputTokens
				s.OutputTokens += r.OutputTokens
				s.CacheRead += r.CacheRead
				s.CacheCreate += r.CacheCreate
				s.CostUSD += r.CostUSD
				if _, err := updateSessionStmt.Exec(s.EndedAt, s.DurationMs,
					1, r.InputTokens, r.OutputTokens, r.CacheRead, r.CacheCreate, r.CostUSD,
					s.ID); err != nil {
					return err
				}
				continue
			}
			// Gap too large, create new session
		}

		// Look up the latest session in DB
		var sessID string
		var sessStarted, sessEnded time.Time
		var sessReqCount, sessIn, sessOut, sessCR, sessCC int64
		var sessCost float64

		err := findSessionStmt.QueryRow(r.ProviderID, r.Model, r.Source, r.AgentID).Scan(
			&sessID, &sessStarted, &sessEnded,
			&sessReqCount, &sessIn, &sessOut, &sessCR, &sessCC, &sessCost)

		if err == nil {
			gap := r.Timestamp.Sub(sessEnded)
			if gap <= sessionWindow {
				// Extend this session
				newEnded := r.Timestamp
				newDuration := newEnded.Sub(sessStarted).Milliseconds()
				if _, err := updateSessionStmt.Exec(newEnded, newDuration,
					1, r.InputTokens, r.OutputTokens, r.CacheRead, r.CacheCreate, r.CostUSD,
					sessID); err != nil {
					return err
				}
				activeSessions[k] = &model.UsageSession{
					ID: sessID, ProviderID: r.ProviderID, Model: r.Model, Source: r.Source, AgentID: r.AgentID,
					StartedAt: sessStarted, EndedAt: newEnded, DurationMs: newDuration,
					RequestCount: sessReqCount + 1,
					InputTokens:  sessIn + r.InputTokens,
					OutputTokens: sessOut + r.OutputTokens,
					CacheRead:    sessCR + r.CacheRead,
					CacheCreate:  sessCC + r.CacheCreate,
					CostUSD:      sessCost + r.CostUSD,
				}
				continue
			}
		}

		// Create new session
		newSess := &model.UsageSession{
			ID:           generateID(),
			ProviderID:   r.ProviderID,
			Model:        r.Model,
			Source:       r.Source,
			AgentID:      r.AgentID,
			StartedAt:    r.Timestamp,
			EndedAt:      r.Timestamp,
			RequestCount: 1,
			InputTokens:  r.InputTokens,
			OutputTokens: r.OutputTokens,
			CacheRead:    r.CacheRead,
			CacheCreate:  r.CacheCreate,
			CostUSD:      r.CostUSD,
		}
		if _, err := insertSessionStmt.Exec(newSess.ID, newSess.ProviderID, newSess.Model, newSess.Source, newSess.AgentID,
			newSess.StartedAt, newSess.EndedAt, newSess.DurationMs, newSess.RequestCount,
			newSess.InputTokens, newSess.OutputTokens, newSess.CacheRead, newSess.CacheCreate,
			newSess.CostUSD); err != nil {
			return err
		}
		activeSessions[k] = newSess
	}

	return nil
}

func (a *Aggregator) upsert(records []model.UsageRecord) {
	backfillCost(records)

	tx, err := a.db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	if err := upsertAll(tx, records); err != nil {
		return
	}

	if err := tx.Commit(); err != nil {
		return
	}

	// Broadcast usage update to connected WebSocket clients
	if a.hub != nil {
		var totalInput, totalOutput int64
		var totalCost float64
		for _, r := range records {
			totalInput += r.InputTokens
			totalOutput += r.OutputTokens
			totalCost += r.CostUSD
		}
		a.hub.Broadcast(ws.NewMessage(ws.TypeUsageUpdate, ws.UsageUpdateData{
			RequestCount: len(records),
			InputTokens:  int(totalInput),
			OutputTokens: int(totalOutput),
			CostUSD:      totalCost,
		}))
	}
}

// ImportFromCCSwitch bulk-imports usage records and updates daily_stats, buckets,
// and sessions in one transaction.
func ImportFromCCSwitch(db *sql.DB, records []model.UsageRecord) error {
	if len(records) == 0 {
		return nil
	}

	backfillCost(records)

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := upsertAll(tx, records); err != nil {
		return err
	}

	return tx.Commit()
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
