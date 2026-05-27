package ccswitch

import (
	"fmt"
	"time"
)

// ProxyLog represents a single row from proxy_request_logs.
type ProxyLog struct {
	RequestID    string
	ProviderID   string
	AppType      string
	Model        string
	Status       int
	InputTokens  int64
	OutputTokens int64
	CacheRead    int64
	CacheCreate  int64
	TotalCostUSD float64
	LatencyMS    int64
	CreatedAt    time.Time
	DataSource   string
}

// FetchProxyLogs returns all proxy logs (for initial sync).
// For incremental sync, use FetchProxyLogsSince.
func (r *Reader) FetchProxyLogs() ([]ProxyLog, error) {
	return r.fetchProxyLogsWhere("", nil)
}

// FetchProxyLogsSince returns proxy logs after the given timestamp (incremental sync).
func (r *Reader) FetchProxyLogsSince(since time.Time) ([]ProxyLog, error) {
	return r.fetchProxyLogsWhere("WHERE created_at > ?", []any{since.Unix()})
}

func (r *Reader) fetchProxyLogsWhere(whereClause string, args []any) ([]ProxyLog, error) {
	query := fmt.Sprintf(`
		SELECT request_id, provider_id, app_type, model, status_code,
		       input_tokens, output_tokens,
		       cache_read_tokens, cache_creation_tokens,
		       total_cost_usd, latency_ms, created_at, data_source
		FROM proxy_request_logs
		%s
		ORDER BY created_at ASC
	`, whereClause)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []ProxyLog
	for rows.Next() {
		var l ProxyLog
		var statusInt, createdAtSec int64
		var costStr string
		if err := rows.Scan(
			&l.RequestID, &l.ProviderID, &l.AppType, &l.Model, &statusInt,
			&l.InputTokens, &l.OutputTokens,
			&l.CacheRead, &l.CacheCreate,
			&costStr, &l.LatencyMS, &createdAtSec, &l.DataSource,
		); err != nil {
			return nil, err
		}
		l.Status = int(statusInt)
		l.TotalCostUSD = parseFloat(costStr)
		l.CreatedAt = time.Unix(createdAtSec, 0)
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

// SyncStats is the summary after a sync run.
type SyncStats struct {
	Fetched     int
	DateRange   string
	TotalCost   float64
	Duration    time.Duration
}
