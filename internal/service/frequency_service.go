package service

import (
	"apihub/internal/repository"
)

// FrequencyRepository defines the interface for frequency data access.
type FrequencyRepository interface {
	GetHourlyHeatmap(days int) ([][]int64, error)
	GetPeakQPS(days int) (*repository.PeakQPS, error)
	GetHourlyDistribution() ([]int64, string, error)
}

// FrequencyService handles frequency/heatmap business logic.
type FrequencyService struct {
	repo FrequencyRepository
}

// NewFrequencyService creates a new FrequencyService.
func NewFrequencyService(repo FrequencyRepository) *FrequencyService {
	return &FrequencyService{repo: repo}
}

// HeatmapResult contains heatmap data.
type HeatmapResult struct {
	Heatmap [][]int64 `json:"heatmap"`
	Days    int       `json:"days"`
}

// GetHourlyHeatmap returns a 7x24 heatmap grid.
func (s *FrequencyService) GetHourlyHeatmap(days int) (*HeatmapResult, error) {
	heatmap, err := s.repo.GetHourlyHeatmap(days)
	if err != nil {
		return nil, err
	}
	return &HeatmapResult{
		Heatmap: heatmap,
		Days:    days,
	}, nil
}

// GetPeakQPS returns peak QPS data.
func (s *FrequencyService) GetPeakQPS(days int) (*repository.PeakQPS, error) {
	return s.repo.GetPeakQPS(days)
}

// TodayResult contains today's hourly distribution.
type TodayResult struct {
	Hourly []int64 `json:"hourly"`
	Date   string  `json:"date"`
}

// GetTodayDistribution returns hourly distribution for today.
func (s *FrequencyService) GetTodayDistribution() (*TodayResult, error) {
	hourly, date, err := s.repo.GetHourlyDistribution()
	if err != nil {
		return nil, err
	}
	return &TodayResult{
		Hourly: hourly,
		Date:   date,
	}, nil
}
