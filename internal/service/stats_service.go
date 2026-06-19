package service

import (
	"apihub/internal/model"
	"apihub/internal/repository"
)

// StatsRepository defines the interface for stats data access.
type StatsRepository interface {
	ListDaily(f repository.DailyStatsFilter) ([]model.DailyStats, error)
	GetCostTrend() ([]repository.TrendRow, error)
	GetModelBreakdown() ([]repository.BreakdownRow, error)
}

// StatsService handles statistics business logic.
type StatsService struct {
	repo StatsRepository
}

// NewStatsService creates a new StatsService.
func NewStatsService(repo StatsRepository) *StatsService {
	return &StatsService{repo: repo}
}

// ListDaily returns daily stats matching the filter.
func (s *StatsService) ListDaily(f repository.DailyStatsFilter) ([]model.DailyStats, error) {
	return s.repo.ListDaily(f)
}

// GetCostTrend returns the cost trend for the last 30 days.
func (s *StatsService) GetCostTrend() ([]repository.TrendRow, error) {
	return s.repo.GetCostTrend()
}

// GetModelBreakdown returns cost breakdown by model.
func (s *StatsService) GetModelBreakdown() ([]repository.BreakdownRow, error) {
	return s.repo.GetModelBreakdown()
}
