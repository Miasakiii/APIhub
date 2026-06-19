package service

import (
	"apihub/internal/model"
)

// AlertRepository defines the interface for alert data access.
type AlertRepository interface {
	List() ([]model.Alert, error)
	Create(a model.Alert) error
	Update(id string, a model.Alert) error
	Delete(id string) error
	ListHistory(alertID string) ([]model.AlertHistory, error)
}

// AlertService handles alert business logic.
type AlertService struct {
	repo AlertRepository
}

// NewAlertService creates a new AlertService.
func NewAlertService(repo AlertRepository) *AlertService {
	return &AlertService{repo: repo}
}

// List returns all alerts.
func (s *AlertService) List() ([]model.Alert, error) {
	return s.repo.List()
}

// Create creates a new alert with a generated ID.
func (s *AlertService) Create(a model.Alert) (model.Alert, error) {
	a.ID = generateID()
	if err := s.repo.Create(a); err != nil {
		return a, err
	}
	return a, nil
}

// Update updates an existing alert.
func (s *AlertService) Update(id string, a model.Alert) error {
	return s.repo.Update(id, a)
}

// Delete removes an alert by ID.
func (s *AlertService) Delete(id string) error {
	return s.repo.Delete(id)
}

// ListHistory returns alert history, optionally filtered by alert ID.
func (s *AlertService) ListHistory(alertID string) ([]model.AlertHistory, error) {
	return s.repo.ListHistory(alertID)
}
