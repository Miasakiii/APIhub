package aggregator

import (
	"apihub/internal/model"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"
)

// Aggregator consumes new UsageRecords and updates DailyStats.
// Uses a single goroutine + batch upserts to prevent conflicts.
type Aggregator struct {
	db    *sql.DB
	queue chan model.UsageRecord
	done  chan struct{}
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

func (a *Aggregator) upsert(records []model.UsageRecord) {
	tx, err := a.db.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO usage_records (id, api_key_id, provider_id, model,
			input_tokens, output_tokens, cache_read, cache_create,
			cost_usd, source, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return
	}
	defer stmt.Close()

	for _, r := range records {
		apiKeyID := nullableString(r.APIKeyID)
		if _, err := stmt.Exec(r.ID, apiKeyID, r.ProviderID, r.Model,
			r.InputTokens, r.OutputTokens, r.CacheRead, r.CacheCreate,
			r.CostUSD, r.Source, r.Timestamp); err != nil {
			return
		}
	}

	// Aggregate into daily_stats
	type aggKey struct {
		ProviderID string
		Model      string
		Source     string
		Date       string
	}
	stats := make(map[aggKey]*model.DailyStats)

	for _, r := range records {
		date := r.Timestamp.Format("2006-01-02")
		k := aggKey{r.ProviderID, r.Model, r.Source, date}
		s, ok := stats[k]
		if !ok {
			s = &model.DailyStats{
				ProviderID: k.ProviderID,
				Model:      k.Model,
				Source:     k.Source,
				Date:       k.Date,
			}
			stats[k] = s
		}
		s.RequestCount++
		s.InputTokens += r.InputTokens
		s.OutputTokens += r.OutputTokens
		s.CacheRead += r.CacheRead
		s.CacheCreate += r.CacheCreate
		s.CostUSD += r.CostUSD
	}

	upsertStmt, err := tx.Prepare(`
		INSERT INTO daily_stats (id, provider_id, model, source, date,
			request_count, input_tokens, output_tokens, cache_read, cache_create, cost_usd, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
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
		return
	}
	defer upsertStmt.Close()

	for _, s := range stats {
		s.ID = generateID()
		if _, err := upsertStmt.Exec(s.ID, s.ProviderID, s.Model, s.Source, s.Date,
			s.RequestCount, s.InputTokens, s.OutputTokens, s.CacheRead, s.CacheCreate, s.CostUSD); err != nil {
			return
		}
	}

	tx.Commit()
}

// ImportFromCCSwitch bulk-imports usage records and updates daily_stats in one transaction.
func ImportFromCCSwitch(db *sql.DB, records []model.UsageRecord) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO usage_records (id, api_key_id, provider_id, model,
			input_tokens, output_tokens, cache_read, cache_create,
			cost_usd, source, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range records {
		apiKeyID := nullableString(r.APIKeyID)
		if _, err := stmt.Exec(r.ID, apiKeyID, r.ProviderID, r.Model,
			r.InputTokens, r.OutputTokens, r.CacheRead, r.CacheCreate,
			r.CostUSD, r.Source, r.Timestamp); err != nil {
			return err
		}
	}

	type aggKey struct {
		ProviderID string
		Model      string
		Source     string
		Date       string
	}
	stats := make(map[aggKey]*model.DailyStats)

	for _, r := range records {
		date := r.Timestamp.Format("2006-01-02")
		k := aggKey{r.ProviderID, r.Model, r.Source, date}
		s, ok := stats[k]
		if !ok {
			s = &model.DailyStats{
				ProviderID: k.ProviderID,
				Model:      k.Model,
				Source:     k.Source,
				Date:       k.Date,
			}
			stats[k] = s
		}
		s.RequestCount++
		s.InputTokens += r.InputTokens
		s.OutputTokens += r.OutputTokens
		s.CacheRead += r.CacheRead
		s.CacheCreate += r.CacheCreate
		s.CostUSD += r.CostUSD
	}

	upsertStmt, err := tx.Prepare(`
		INSERT INTO daily_stats (id, provider_id, model, source, date,
			request_count, input_tokens, output_tokens, cache_read, cache_create, cost_usd, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
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
	defer upsertStmt.Close()

	for _, s := range stats {
		s.ID = generateID()
		if _, err := upsertStmt.Exec(s.ID, s.ProviderID, s.Model, s.Source, s.Date,
			s.RequestCount, s.InputTokens, s.OutputTokens, s.CacheRead, s.CacheCreate, s.CostUSD); err != nil {
			return err
		}
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
