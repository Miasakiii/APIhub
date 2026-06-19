package repository

import (
	"apihub/internal/model"
	"database/sql"
	"fmt"
	"strings"
)

// StatsRepo handles statistics database operations.
type StatsRepo struct {
	db *sql.DB
}

// NewStatsRepo creates a new StatsRepo.
func NewStatsRepo(db *sql.DB) *StatsRepo {
	return &StatsRepo{db: db}
}

// DailyStatsFilter contains filters for daily stats queries.
type DailyStatsFilter struct {
	ProviderID string
	Model      string
	AgentID    string
}

// ListDaily returns the last 30 days of daily stats matching the filter.
func (r *StatsRepo) ListDaily(f DailyStatsFilter) ([]model.DailyStats, error) {
	q := `SELECT id, provider_id, model, source, agent_id, date, request_count,
		         input_tokens, output_tokens, cache_read, cache_create, cost_usd, updated_at
		  FROM daily_stats`

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
	if f.AgentID != "" {
		conditions = append(conditions, "agent_id = ?")
		args = append(args, f.AgentID)
	}
	if len(conditions) > 0 {
		q += " WHERE " + strings.Join(conditions, " AND ")
	}
	q += " ORDER BY date DESC LIMIT 30"

	rows, err := r.db.Query(q, args...)
	if err != nil {
		return nil, fmt.Errorf("query daily stats: %w", err)
	}
	defer rows.Close()

	var stats []model.DailyStats
	for rows.Next() {
		var s model.DailyStats
		var agentID, updatedStr sql.NullString
		if err := rows.Scan(&s.ID, &s.ProviderID, &s.Model, &s.Source, &agentID, &s.Date,
			&s.RequestCount, &s.InputTokens, &s.OutputTokens, &s.CacheRead, &s.CacheCreate,
			&s.CostUSD, &updatedStr); err != nil {
			return nil, fmt.Errorf("scan daily stat: %w", err)
		}
		if agentID.Valid {
			s.AgentID = agentID.String
		}
		if updatedStr.Valid {
			s.UpdatedAt, _ = parseTime(updatedStr.String)
		}
		stats = append(stats, s)
	}
	if stats == nil {
		stats = []model.DailyStats{}
	}
	return stats, nil
}

// TrendRow represents a cost trend data point.
type TrendRow struct {
	Date      string  `json:"date"`
	Cost      float64 `json:"cost_usd"`
	Tokens    int64   `json:"tokens"`
	ReqCount  int64   `json:"request_count"`
}

// GetCostTrend returns the last 30 days of cost trend.
func (r *StatsRepo) GetCostTrend() ([]TrendRow, error) {
	rows, err := r.db.Query(`
		SELECT date, SUM(cost_usd) as cost, SUM(input_tokens + output_tokens) as tokens, SUM(request_count) as req_count
		FROM daily_stats
		GROUP BY date
		ORDER BY date DESC
		LIMIT 30
	`)
	if err != nil {
		return nil, fmt.Errorf("query cost trend: %w", err)
	}
	defer rows.Close()

	var result []TrendRow
	for rows.Next() {
		var t TrendRow
		rows.Scan(&t.Date, &t.Cost, &t.Tokens, &t.ReqCount)
		result = append(result, t)
	}
	if result == nil {
		result = []TrendRow{}
	}
	return result, nil
}

// BreakdownRow represents a model cost breakdown.
type BreakdownRow struct {
	Model       string  `json:"model"`
	TotalCost   float64 `json:"total_cost_usd"`
	TotalTokens int64   `json:"total_tokens"`
	ReqCount    int64   `json:"request_count"`
}

// GetModelBreakdown returns cost breakdown by model.
func (r *StatsRepo) GetModelBreakdown() ([]BreakdownRow, error) {
	rows, err := r.db.Query(`
		SELECT model, SUM(cost_usd) as total_cost,
		       SUM(input_tokens + output_tokens + cache_read + cache_create) as total_tokens,
		       SUM(request_count) as req_count
		FROM daily_stats
		GROUP BY model
		ORDER BY total_cost DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("query model breakdown: %w", err)
	}
	defer rows.Close()

	var result []BreakdownRow
	for rows.Next() {
		var b BreakdownRow
		rows.Scan(&b.Model, &b.TotalCost, &b.TotalTokens, &b.ReqCount)
		result = append(result, b)
	}
	if result == nil {
		result = []BreakdownRow{}
	}
	return result, nil
}
