package repository

import (
	"database/sql"
	"strconv"
	"time"
)

// FrequencyRepo handles frequency/heatmap database operations.
type FrequencyRepo struct {
	db *sql.DB
}

// NewFrequencyRepo creates a new FrequencyRepo.
func NewFrequencyRepo(db *sql.DB) *FrequencyRepo {
	return &FrequencyRepo{db: db}
}

// HourlyHeatmap returns a 7x24 heatmap grid for the given number of days.
func (r *FrequencyRepo) GetHourlyHeatmap(days int) ([][]int64, error) {
	from := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	rows, err := r.db.Query(`
		SELECT
			strftime('%w', timestamp) as day_of_week,
			strftime('%H', timestamp) as hour,
			COUNT(*) as count
		FROM usage_records
		WHERE timestamp > ?
		GROUP BY day_of_week, hour
		ORDER BY day_of_week, hour
	`, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Initialize 7x24 grid
	heatmap := make([][]int64, 7)
	for i := range heatmap {
		heatmap[i] = make([]int64, 24)
	}

	for rows.Next() {
		var dowStr, hourStr string
		var count int64
		if err := rows.Scan(&dowStr, &hourStr, &count); err != nil {
			continue
		}
		dow, _ := strconv.Atoi(dowStr)
		hour, _ := strconv.Atoi(hourStr)
		if dow >= 0 && dow < 7 && hour >= 0 && hour < 24 {
			heatmap[dow][hour] = count
		}
	}
	return heatmap, nil
}

// PeakQPS contains peak queries-per-second data.
type PeakQPS struct {
	PeakQPS    float64 `json:"peak_qps"`
	PeakMinute string  `json:"peak_minute"`
	AvgQPS     float64 `json:"avg_qps"`
	PeakCount  int64   `json:"peak_count"`
}

// GetPeakQPS returns peak QPS data for the given number of days.
func (r *FrequencyRepo) GetPeakQPS(days int) (*PeakQPS, error) {
	from := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	rows, err := r.db.Query(`
		SELECT
			strftime('%Y-%m-%d %H:%M', timestamp) as minute,
			COUNT(*) as count
		FROM usage_records
		WHERE timestamp > ?
		GROUP BY minute
		ORDER BY count DESC
		LIMIT 1
	`, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var peakMinute string
	var peakCount int64
	if rows.Next() {
		rows.Scan(&peakMinute, &peakCount)
	}

	var avgQPS float64
	r.db.QueryRow(`
		SELECT CAST(COUNT(*) AS REAL) / (3600.0 * ?)
		FROM usage_records
		WHERE timestamp > ?
	`, days, from).Scan(&avgQPS)

	return &PeakQPS{
		PeakQPS:    float64(peakCount) / 60.0,
		PeakMinute: peakMinute,
		AvgQPS:     avgQPS,
		PeakCount:  peakCount,
	}, nil
}

// HourlyDistribution returns hourly request counts for today.
func (r *FrequencyRepo) GetHourlyDistribution() ([]int64, string, error) {
	now := time.Now()
	from := now.Truncate(24 * time.Hour).Format("2006-01-02")

	rows, err := r.db.Query(`
		SELECT
			strftime('%H', timestamp) as hour,
			COUNT(*) as count
		FROM usage_records
		WHERE timestamp > ?
		GROUP BY hour
		ORDER BY hour
	`, from)
	if err != nil {
		return nil, from, err
	}
	defer rows.Close()

	hourly := make([]int64, 24)
	for rows.Next() {
		var hourStr string
		var count int64
		if err := rows.Scan(&hourStr, &count); err != nil {
			continue
		}
		hour, _ := strconv.Atoi(hourStr)
		if hour >= 0 && hour < 24 {
			hourly[hour] = count
		}
	}
	return hourly, from, nil
}
