package service

import (
	"apihub/internal/model"
	"apihub/internal/repository"
	"time"
)

// SessionRepoInterface defines the session repository operations.
type SessionRepoInterface interface {
	List(f repository.SessionFilter) ([]model.UsageSession, int, error)
	GetStats() (*repository.SessionStats, error)
}

// BucketRepoInterface defines the bucket repository operations.
type BucketRepoInterface interface {
	List(f repository.BucketFilter) ([]model.ActivityBucket, error)
	GetHourlyBuckets(date string) ([]model.ActivityBucket, error)
}

// SessionService provides business logic for sessions and activity buckets.
type SessionService struct {
	sessionRepo SessionRepoInterface
	bucketRepo  BucketRepoInterface
}

// NewSessionService creates a new SessionService.
func NewSessionService(sessionRepo SessionRepoInterface, bucketRepo BucketRepoInterface) *SessionService {
	return &SessionService{sessionRepo: sessionRepo, bucketRepo: bucketRepo}
}

// SessionListResult wraps paginated session results.
type SessionListResult struct {
	Sessions []model.UsageSession `json:"sessions"`
	Total    int                  `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"page_size"`
}

// ListSessions returns paginated sessions with filters.
func (s *SessionService) ListSessions(f repository.SessionFilter) (*SessionListResult, error) {
	sessions, total, err := s.sessionRepo.List(f)
	if err != nil {
		return nil, err
	}
	page, pageSize := normalizePage(f.Page, f.PageSize)
	return &SessionListResult{
		Sessions: sessions,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetStats returns aggregate session statistics.
func (s *SessionService) GetStats() (*repository.SessionStats, error) {
	return s.sessionRepo.GetStats()
}

// ListBuckets returns activity buckets filtered by time range.
func (s *SessionService) ListBuckets(from, to time.Time, providerID, model, agentID string) ([]model.ActivityBucket, error) {
	f := repository.BucketFilter{
		ProviderID: providerID,
		Model:      model,
		AgentID:    agentID,
		From:       from,
		To:         to,
	}
	return s.bucketRepo.List(f)
}

// GetHourlyBuckets returns 24 hourly buckets for a specific date.
func (s *SessionService) GetHourlyBuckets(date string) ([]model.ActivityBucket, error) {
	return s.bucketRepo.GetHourlyBuckets(date)
}

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
