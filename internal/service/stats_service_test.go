package service

import (
	"apihub/internal/model"
	"apihub/internal/repository"
	"errors"
	"testing"
)

// mockStatsRepo is a mock implementation of stats repository.
type mockStatsRepo struct {
	daily     []model.DailyStats
	trend     []repository.TrendRow
	breakdown []repository.BreakdownRow
	dailyErr  error
	trendErr  error
	breakErr  error
}

func (m *mockStatsRepo) ListDaily(f repository.DailyStatsFilter) ([]model.DailyStats, error) {
	return m.daily, m.dailyErr
}

func (m *mockStatsRepo) GetCostTrend() ([]repository.TrendRow, error) {
	return m.trend, m.trendErr
}

func (m *mockStatsRepo) GetModelBreakdown() ([]repository.BreakdownRow, error) {
	return m.breakdown, m.breakErr
}

func TestStatsService_ListDaily(t *testing.T) {
	repo := &mockStatsRepo{
		daily: []model.DailyStats{
			{ID: "1", Date: "2024-01-15", CostUSD: 10.0},
			{ID: "2", Date: "2024-01-14", CostUSD: 8.5},
		},
	}
	svc := NewStatsService(repo)

	stats, err := svc.ListDaily(repository.DailyStatsFilter{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 stats, got %d", len(stats))
	}
}

func TestStatsService_ListDaily_WithFilter(t *testing.T) {
	repo := &mockStatsRepo{
		daily: []model.DailyStats{
			{ID: "1", Model: "gpt-4", Date: "2024-01-15", CostUSD: 10.0},
		},
	}
	svc := NewStatsService(repo)

	stats, err := svc.ListDaily(repository.DailyStatsFilter{
		Model: "gpt-4",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}
	if stats[0].Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got '%s'", stats[0].Model)
	}
}

func TestStatsService_GetCostTrend(t *testing.T) {
	repo := &mockStatsRepo{
		trend: []repository.TrendRow{
			{Date: "2024-01-15", Cost: 10.0, Tokens: 5000, ReqCount: 50},
			{Date: "2024-01-14", Cost: 8.5, Tokens: 4000, ReqCount: 40},
		},
	}
	svc := NewStatsService(repo)

	trend, err := svc.GetCostTrend()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(trend) != 2 {
		t.Fatalf("expected 2 trend entries, got %d", len(trend))
	}
	if trend[0].Cost != 10.0 {
		t.Errorf("expected cost 10.0, got %f", trend[0].Cost)
	}
}

func TestStatsService_GetModelBreakdown(t *testing.T) {
	repo := &mockStatsRepo{
		breakdown: []repository.BreakdownRow{
			{Model: "gpt-4", TotalCost: 50.0, TotalTokens: 25000, ReqCount: 100},
			{Model: "claude-3", TotalCost: 30.0, TotalTokens: 15000, ReqCount: 60},
		},
	}
	svc := NewStatsService(repo)

	breakdown, err := svc.GetModelBreakdown()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(breakdown) != 2 {
		t.Fatalf("expected 2 breakdown entries, got %d", len(breakdown))
	}
	if breakdown[0].Model != "gpt-4" {
		t.Errorf("expected first model 'gpt-4', got '%s'", breakdown[0].Model)
	}
	if breakdown[0].TotalCost != 50.0 {
		t.Errorf("expected total cost 50.0, got %f", breakdown[0].TotalCost)
	}
}

func TestStatsService_Error(t *testing.T) {
	repo := &mockStatsRepo{
		dailyErr: errors.New("database error"),
	}
	svc := NewStatsService(repo)

	_, err := svc.ListDaily(repository.DailyStatsFilter{})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
