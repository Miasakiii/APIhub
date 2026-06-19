package repository

import (
	"apihub/internal/model"
	"database/sql"
	"fmt"
	"strings"
)

// UsageRepo handles usage record database operations.
type UsageRepo struct {
	db *sql.DB
}

// NewUsageRepo creates a new UsageRepo.
func NewUsageRepo(db *sql.DB) *UsageRepo {
	return &UsageRepo{db: db}
}

// UsageFilter contains filters for querying usage records.
type UsageFilter struct {
	ProviderID string
	Model      string
	Source     string
	AgentID    string
	Date       string
	Page       int
	PageSize   int
}

// List returns paginated usage records matching the filter.
func (r *UsageRepo) List(f UsageFilter) ([]model.UsageRecord, int, error) {
	var args []any
	conditions := []string{}
	if f.ProviderID != "" {
		conditions = append(conditions, "provider_id = ?")
		args = append(args, f.ProviderID)
	}
	if f.Model != "" {
		conditions = append(conditions, "model = ?")
		args = append(args, f.Model)
	}
	if f.Source != "" {
		conditions = append(conditions, "source = ?")
		args = append(args, f.Source)
	}
	if f.AgentID != "" {
		conditions = append(conditions, "agent_id = ?")
		args = append(args, f.AgentID)
	}
	if f.Date != "" {
		conditions = append(conditions, "DATE(timestamp) = ?")
		args = append(args, f.Date)
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Get total count
	var total int
	countQ := "SELECT COUNT(*) FROM usage_records" + where
	if err := r.db.QueryRow(countQ, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count usage records: %w", err)
	}

	// Get paginated records
	q := `SELECT id, api_key_id, provider_id, model, agent_id, input_tokens, output_tokens,
		       cache_read, cache_create, cost_usd, source, timestamp, created_at
		  FROM usage_records` + where + ` ORDER BY timestamp DESC LIMIT ? OFFSET ?`

	rows, err := r.db.Query(q, append(args, f.PageSize, (f.Page-1)*f.PageSize)...)
	if err != nil {
		return nil, 0, fmt.Errorf("query usage records: %w", err)
	}
	defer rows.Close()

	var records []model.UsageRecord
	for rows.Next() {
		var rec model.UsageRecord
		var apiKeyID, agentID, tsStr, createdStr sql.NullString
		if err := rows.Scan(&rec.ID, &apiKeyID, &rec.ProviderID, &rec.Model, &agentID,
			&rec.InputTokens, &rec.OutputTokens, &rec.CacheRead, &rec.CacheCreate,
			&rec.CostUSD, &rec.Source, &tsStr, &createdStr); err != nil {
			return nil, 0, fmt.Errorf("scan usage record: %w", err)
		}
		if apiKeyID.Valid {
			rec.APIKeyID = apiKeyID.String
		}
		if agentID.Valid {
			rec.AgentID = agentID.String
		}
		if tsStr.Valid {
			rec.Timestamp, _ = parseTime(tsStr.String)
		}
		if createdStr.Valid {
			rec.CreatedAt, _ = parseTime(createdStr.String)
		}
		records = append(records, rec)
	}
	if records == nil {
		records = []model.UsageRecord{}
	}
	return records, total, nil
}

// Summary contains aggregated usage statistics.
type Summary struct {
	TotalCost     float64 `json:"total_cost_usd"`
	TotalTokens   int64   `json:"total_tokens"`
	TotalRequests int64   `json:"total_requests"`
	UniqueModels  int64   `json:"unique_models"`
	UniqueKeys    int64   `json:"unique_keys"`
}

// GetSummary returns aggregated usage statistics.
func (r *UsageRepo) GetSummary() (*Summary, error) {
	var s Summary
	err := r.db.QueryRow(`
		SELECT COALESCE(SUM(cost_usd), 0),
		       COALESCE(SUM(input_tokens + output_tokens + cache_read + cache_create), 0),
		       COUNT(*)
		FROM usage_records
	`).Scan(&s.TotalCost, &s.TotalTokens, &s.TotalRequests)
	if err != nil {
		return nil, err
	}

	r.db.QueryRow(`SELECT COUNT(DISTINCT model) FROM usage_records`).Scan(&s.UniqueModels)
	// UniqueKeys is set externally by KeyRepo
	return &s, nil
}

// GetProviderSummary returns aggregated cost and request count for a specific provider.
func (r *UsageRepo) GetProviderSummary(providerID string) (float64, int64, error) {
	var totalCost float64
	var totalRequests int64
	err := r.db.QueryRow(`
		SELECT COALESCE(SUM(cost_usd), 0), COUNT(*)
		FROM usage_records WHERE provider_id = ?
	`, providerID).Scan(&totalCost, &totalRequests)
	if err != nil {
		return 0, 0, err
	}
	return totalCost, totalRequests, nil
}

// ListAll returns all usage records (for CSV export, capped at 10000).
func (r *UsageRepo) ListAll() ([]model.UsageRecord, error) {
	rows, err := r.db.Query(`
		SELECT id, api_key_id, provider_id, model, agent_id, input_tokens, output_tokens,
		       cache_read, cache_create, cost_usd, source, timestamp, created_at
		FROM usage_records ORDER BY timestamp DESC LIMIT 10000
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []model.UsageRecord
	for rows.Next() {
		var rec model.UsageRecord
		var apiKeyID, agentID, tsStr, createdStr sql.NullString
		if err := rows.Scan(&rec.ID, &apiKeyID, &rec.ProviderID, &rec.Model, &agentID,
			&rec.InputTokens, &rec.OutputTokens, &rec.CacheRead, &rec.CacheCreate,
			&rec.CostUSD, &rec.Source, &tsStr, &createdStr); err != nil {
			return nil, err
		}
		if apiKeyID.Valid {
			rec.APIKeyID = apiKeyID.String
		}
		if agentID.Valid {
			rec.AgentID = agentID.String
		}
		if tsStr.Valid {
			rec.Timestamp, _ = parseTime(tsStr.String)
		}
		if createdStr.Valid {
			rec.CreatedAt, _ = parseTime(createdStr.String)
		}
		records = append(records, rec)
	}
	if records == nil {
		records = []model.UsageRecord{}
	}
	return records, nil
}
