package repository

import (
	"apihub/internal/model"
	"database/sql"
	"fmt"
	"time"
)

// SessionFilter defines filters for querying sessions.
type SessionFilter struct {
	ProviderID string
	Model      string
	Source     string
	AgentID    string
	From       time.Time
	To         time.Time
	Page       int
	PageSize   int
}

// SessionStats holds aggregate statistics about sessions.
type SessionStats struct {
	TotalSessions int64   `json:"total_sessions"`
	AvgDurationMs float64 `json:"avg_duration_ms"`
	AvgCost       float64 `json:"avg_cost_usd"`
	TotalCost     float64 `json:"total_cost_usd"`
	TotalRequests int64   `json:"total_requests"`
}

// SessionRepo handles database operations for usage_sessions.
type SessionRepo struct {
	db *sql.DB
}

// NewSessionRepo creates a new SessionRepo.
func NewSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

// List returns paginated sessions with optional filters.
func (r *SessionRepo) List(f SessionFilter) ([]model.UsageSession, int, error) {
	where, args := buildSessionWhere(f)

	// Count total
	var total int
	countQuery := "SELECT COUNT(*) FROM usage_sessions" + where
	if err := r.db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count sessions: %w", err)
	}

	// Paginate
	page, pageSize := normalizePage(f.Page, f.PageSize)
	offset := (page - 1) * pageSize

	query := `SELECT id, provider_id, model, source, agent_id, started_at, ended_at, duration_ms,
		request_count, input_tokens, output_tokens, cache_read, cache_create, cost_usd
		FROM usage_sessions` + where +
		` ORDER BY started_at DESC LIMIT ? OFFSET ?`
	args = append(args, pageSize, offset)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []model.UsageSession
	for rows.Next() {
		var s model.UsageSession
		var agentID sql.NullString
		if err := rows.Scan(&s.ID, &s.ProviderID, &s.Model, &s.Source, &agentID,
			&s.StartedAt, &s.EndedAt, &s.DurationMs,
			&s.RequestCount, &s.InputTokens, &s.OutputTokens,
			&s.CacheRead, &s.CacheCreate, &s.CostUSD); err != nil {
			return nil, 0, fmt.Errorf("scan session: %w", err)
		}
		if agentID.Valid {
			s.AgentID = agentID.String
		}
		sessions = append(sessions, s)
	}

	return sessions, total, nil
}

// GetStats returns aggregate session statistics.
func (r *SessionRepo) GetStats() (*SessionStats, error) {
	var stats SessionStats
	err := r.db.QueryRow(`
		SELECT COUNT(*), COALESCE(AVG(duration_ms), 0), COALESCE(AVG(cost_usd), 0),
			COALESCE(SUM(cost_usd), 0), COALESCE(SUM(request_count), 0)
		FROM usage_sessions
	`).Scan(&stats.TotalSessions, &stats.AvgDurationMs, &stats.AvgCost,
		&stats.TotalCost, &stats.TotalRequests)
	if err != nil {
		return nil, fmt.Errorf("session stats: %w", err)
	}
	return &stats, nil
}

func buildSessionWhere(f SessionFilter) (string, []any) {
	var conds []string
	var args []any

	if f.ProviderID != "" {
		conds = append(conds, "provider_id = ?")
		args = append(args, f.ProviderID)
	}
	if f.Model != "" {
		conds = append(conds, "model = ?")
		args = append(args, f.Model)
	}
	if f.Source != "" {
		conds = append(conds, "source = ?")
		args = append(args, f.Source)
	}
	if f.AgentID != "" {
		conds = append(conds, "agent_id = ?")
		args = append(args, f.AgentID)
	}
	if !f.From.IsZero() {
		conds = append(conds, "started_at >= ?")
		args = append(args, f.From)
	}
	if !f.To.IsZero() {
		conds = append(conds, "started_at <= ?")
		args = append(args, f.To)
	}

	if len(conds) == 0 {
		return "", args
	}
	return " WHERE " + joinConds(conds, " AND "), args
}

// BucketFilter defines filters for querying activity buckets.
type BucketFilter struct {
	ProviderID string
	Model      string
	AgentID    string
	From       time.Time
	To         time.Time
}

// BucketRepo handles database operations for usage_activity_buckets.
type BucketRepo struct {
	db *sql.DB
}

// NewBucketRepo creates a new BucketRepo.
func NewBucketRepo(db *sql.DB) *BucketRepo {
	return &BucketRepo{db: db}
}

// List returns buckets filtered by time range and optional provider/model.
func (r *BucketRepo) List(f BucketFilter) ([]model.ActivityBucket, error) {
	where, args := buildBucketWhere(f)

	query := `SELECT id, bucket_start, provider_id, model, agent_id,
		request_count, input_tokens, output_tokens, cache_read, cache_create, cost_usd
		FROM usage_activity_buckets` + where +
		` ORDER BY bucket_start ASC`

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list buckets: %w", err)
	}
	defer rows.Close()

	var buckets []model.ActivityBucket
	for rows.Next() {
		var b model.ActivityBucket
		var agentID sql.NullString
		if err := rows.Scan(&b.ID, &b.BucketStart, &b.ProviderID, &b.Model, &agentID,
			&b.RequestCount, &b.InputTokens, &b.OutputTokens,
			&b.CacheRead, &b.CacheCreate, &b.CostUSD); err != nil {
			return nil, fmt.Errorf("scan bucket: %w", err)
		}
		if agentID.Valid {
			b.AgentID = agentID.String
		}
		buckets = append(buckets, b)
	}
	return buckets, nil
}

// GetHourlyBuckets returns 24 hourly buckets for a specific date.
func (r *BucketRepo) GetHourlyBuckets(date string) ([]model.ActivityBucket, error) {
	rows, err := r.db.Query(`
		SELECT id, bucket_start, provider_id, model, agent_id,
			request_count, input_tokens, output_tokens, cache_read, cache_create, cost_usd
		FROM usage_activity_buckets
		WHERE DATE(bucket_start) = ?
		ORDER BY bucket_start ASC
	`, date)
	if err != nil {
		return nil, fmt.Errorf("hourly buckets: %w", err)
	}
	defer rows.Close()

	var buckets []model.ActivityBucket
	for rows.Next() {
		var b model.ActivityBucket
		var agentID sql.NullString
		if err := rows.Scan(&b.ID, &b.BucketStart, &b.ProviderID, &b.Model, &agentID,
			&b.RequestCount, &b.InputTokens, &b.OutputTokens,
			&b.CacheRead, &b.CacheCreate, &b.CostUSD); err != nil {
			return nil, fmt.Errorf("scan bucket: %w", err)
		}
		if agentID.Valid {
			b.AgentID = agentID.String
		}
		buckets = append(buckets, b)
	}
	return buckets, nil
}

func buildBucketWhere(f BucketFilter) (string, []any) {
	var conds []string
	var args []any

	if f.ProviderID != "" {
		conds = append(conds, "provider_id = ?")
		args = append(args, f.ProviderID)
	}
	if f.Model != "" {
		conds = append(conds, "model = ?")
		args = append(args, f.Model)
	}
	if f.AgentID != "" {
		conds = append(conds, "agent_id = ?")
		args = append(args, f.AgentID)
	}
	if !f.From.IsZero() {
		conds = append(conds, "bucket_start >= ?")
		args = append(args, f.From)
	}
	if !f.To.IsZero() {
		conds = append(conds, "bucket_start <= ?")
		args = append(args, f.To)
	}

	if len(conds) == 0 {
		return "", args
	}
	return " WHERE " + joinConds(conds, " AND "), args
}

// joinConds joins SQL conditions with the given separator.
func joinConds(conds []string, sep string) string {
	result := ""
	for i, c := range conds {
		if i > 0 {
			result += sep
		}
		result += c
	}
	return result
}

// normalizePage ensures page and pageSize are valid.
func normalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}
	if pageSize > 200 {
		pageSize = 200
	}
	return page, pageSize
}
