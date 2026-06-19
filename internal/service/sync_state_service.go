package service

import "apihub/internal/model"

// SyncStateRepository defines the interface for sync state data access.
type SyncStateRepository interface {
	ListAll() ([]model.SyncState, error)
}

// SyncStateService handles sync state business logic.
type SyncStateService struct {
	repo SyncStateRepository
}

// NewSyncStateService creates a new SyncStateService.
func NewSyncStateService(repo SyncStateRepository) *SyncStateService {
	return &SyncStateService{repo: repo}
}

// ListAll returns all sync states.
func (s *SyncStateService) ListAll() ([]model.SyncState, error) {
	return s.repo.ListAll()
}
