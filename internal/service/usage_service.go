package service

import (
	"apihub/internal/model"
	"apihub/internal/repository"
	"fmt"
	"strings"
	"time"
)

// UsageRepository defines the interface for usage data access.
type UsageRepository interface {
	List(f repository.UsageFilter) ([]model.UsageRecord, int, error)
	GetSummary() (*repository.Summary, error)
	ListAll() ([]model.UsageRecord, error)
}

// UsageKeyRepository defines the interface for key operations needed by usage service.
type UsageKeyRepository interface {
	CountActive() (int64, error)
}

// UsageService handles usage record business logic.
type UsageService struct {
	repo    UsageRepository
	keyRepo UsageKeyRepository
}

// NewUsageService creates a new UsageService.
func NewUsageService(repo UsageRepository, keyRepo UsageKeyRepository) *UsageService {
	return &UsageService{repo: repo, keyRepo: keyRepo}
}

// ListResult contains paginated usage records.
type ListResult struct {
	Records  []model.UsageRecord `json:"records"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// List returns paginated usage records matching the filter.
func (s *UsageService) List(f repository.UsageFilter) (*ListResult, error) {
	records, total, err := s.repo.List(f)
	if err != nil {
		return nil, err
	}
	return &ListResult{
		Records:  records,
		Total:    total,
		Page:     f.Page,
		PageSize: f.PageSize,
	}, nil
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
func (s *UsageService) GetSummary() (*Summary, error) {
	repoSummary, err := s.repo.GetSummary()
	if err != nil {
		return nil, err
	}

	// Get active key count from key repo
	activeKeys, err := s.keyRepo.CountActive()
	if err != nil {
		activeKeys = 0 // Non-critical, continue
	}

	return &Summary{
		TotalCost:     repoSummary.TotalCost,
		TotalTokens:   repoSummary.TotalTokens,
		TotalRequests: repoSummary.TotalRequests,
		UniqueModels:  repoSummary.UniqueModels,
		UniqueKeys:    activeKeys,
	}, nil
}

// ExportCSV returns usage records formatted as a CSV string.
func (s *UsageService) ExportCSV() (string, string, error) {
	records, err := s.repo.ListAll()
	if err != nil {
		return "", "", err
	}

	filename := fmt.Sprintf("usage_%s.csv", time.Now().Format("20060102"))

	var b strings.Builder
	b.WriteString("ID,Provider,Model,Input Tokens,Output Tokens,Cache Read,Cache Create,Cost (USD),Source,Timestamp\n")
	for _, r := range records {
		b.WriteString(fmt.Sprintf("%s,%s,%s,%d,%d,%d,%d,%.6f,%s,%s\n",
			r.ID, r.ProviderID, r.Model, r.InputTokens, r.OutputTokens,
			r.CacheRead, r.CacheCreate, r.CostUSD, r.Source,
			r.Timestamp.Format("2006-01-02 15:04:05")))
	}

	return b.String(), filename, nil
}
