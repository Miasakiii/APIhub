package service

import "apihub/internal/model"

// WebhookRepository defines the interface for webhook data access.
type WebhookRepository interface {
	List() ([]model.WebhookSetting, error)
	Create(s model.WebhookSetting) (model.WebhookSetting, error)
	Delete(id string) error
}

// WebhookService handles webhook business logic.
type WebhookService struct {
	repo WebhookRepository
}

// NewWebhookService creates a new WebhookService.
func NewWebhookService(repo WebhookRepository) *WebhookService {
	return &WebhookService{repo: repo}
}

// List returns all webhook settings.
func (s *WebhookService) List() ([]model.WebhookSetting, error) {
	return s.repo.List()
}

// Create adds a new webhook setting.
func (s *WebhookService) Create(setting model.WebhookSetting) (model.WebhookSetting, error) {
	return s.repo.Create(setting)
}

// Delete removes a webhook setting by ID.
func (s *WebhookService) Delete(id string) error {
	return s.repo.Delete(id)
}
