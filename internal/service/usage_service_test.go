package service

import (
	"apihub/internal/model"
	"apihub/internal/repository"
	"errors"
	"testing"
)

// mockUsageRepo is a mock implementation of usage repository.
type mockUsageRepo struct {
	records   []model.UsageRecord
	total     int
	summary   *repository.Summary
	listErr   error
	summaryErr error
}

func (m *mockUsageRepo) List(f repository.UsageFilter) ([]model.UsageRecord, int, error) {
	return m.records, m.total, m.listErr
}

func (m *mockUsageRepo) GetSummary() (*repository.Summary, error) {
	return m.summary, m.summaryErr
}

func (m *mockUsageRepo) ListAll() ([]model.UsageRecord, error) {
	return m.records, m.listErr
}

// mockKeyRepoForUsage is a mock key repo for usage service tests.
type mockKeyRepoForUsage struct {
	activeCount int64
	countErr    error
}

func (m *mockKeyRepoForUsage) CountActive() (int64, error) {
	return m.activeCount, m.countErr
}

func TestUsageService_List(t *testing.T) {
	repo := &mockUsageRepo{
		records: []model.UsageRecord{
			{ID: "1", Model: "gpt-4", CostUSD: 0.05},
			{ID: "2", Model: "claude-3", CostUSD: 0.03},
		},
		total: 10,
	}
	keyRepo := &mockKeyRepoForUsage{activeCount: 5}
	svc := NewUsageService(repo, keyRepo)

	result, err := svc.List(repository.UsageFilter{
		Page:     1,
		PageSize: 50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 10 {
		t.Errorf("expected total=10, got %d", result.Total)
	}
	if len(result.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(result.Records))
	}
	if result.Page != 1 {
		t.Errorf("expected page=1, got %d", result.Page)
	}
}

func TestUsageService_List_Error(t *testing.T) {
	repo := &mockUsageRepo{
		listErr: errors.New("database error"),
	}
	keyRepo := &mockKeyRepoForUsage{}
	svc := NewUsageService(repo, keyRepo)

	_, err := svc.List(repository.UsageFilter{Page: 1, PageSize: 50})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUsageService_GetSummary(t *testing.T) {
	repo := &mockUsageRepo{
		summary: &repository.Summary{
			TotalCost:     123.45,
			TotalTokens:   100000,
			TotalRequests: 500,
			UniqueModels:  5,
		},
	}
	keyRepo := &mockKeyRepoForUsage{activeCount: 3}
	svc := NewUsageService(repo, keyRepo)

	summary, err := svc.GetSummary()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.TotalCost != 123.45 {
		t.Errorf("expected total cost 123.45, got %f", summary.TotalCost)
	}
	if summary.UniqueKeys != 3 {
		t.Errorf("expected unique keys 3, got %d", summary.UniqueKeys)
	}
	if summary.UniqueModels != 5 {
		t.Errorf("expected unique models 5, got %d", summary.UniqueModels)
	}
}
