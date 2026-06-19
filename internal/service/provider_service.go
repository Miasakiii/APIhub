package service

import (
	"apihub/internal/model"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
)

// ErrProviderNotFound is returned when a provider does not exist.
var ErrProviderNotFound = errors.New("provider not found")

// ProviderRepository defines the interface for provider data access.
type ProviderRepository interface {
	List() ([]model.Provider, error)
	Create(p model.Provider) error
	Delete(id string) error
	GetByID(id string) (*model.Provider, error)
	GetByType(pType string) (*model.Provider, error)
}

// ProviderKeyRepository defines the interface for key operations needed by provider service.
type ProviderKeyRepository interface {
	ListByProvider(providerID string) ([]model.APIKey, error)
}

// ProviderUsageRepository defines the interface for usage operations needed by provider service.
type ProviderUsageRepository interface {
	GetProviderSummary(providerID string) (float64, int64, error)
}

// ProviderService handles provider business logic.
type ProviderService struct {
	repo    ProviderRepository
	keyRepo ProviderKeyRepository
	usageRepo ProviderUsageRepository
}

// NewProviderService creates a new ProviderService.
func NewProviderService(repo ProviderRepository, keyRepo ProviderKeyRepository, usageRepo ProviderUsageRepository) *ProviderService {
	return &ProviderService{repo: repo, keyRepo: keyRepo, usageRepo: usageRepo}
}

// List returns all providers.
func (s *ProviderService) List() ([]model.Provider, error) {
	return s.repo.List()
}

// Create creates a new provider with a generated ID.
func (s *ProviderService) Create(p model.Provider) (model.Provider, error) {
	p.ID = generateID()
	p.Enabled = true
	if err := s.repo.Create(p); err != nil {
		return p, err
	}
	return p, nil
}

// Delete removes a provider by ID.
func (s *ProviderService) Delete(id string) error {
	return s.repo.Delete(id)
}

// GetByID returns a provider by ID.
func (s *ProviderService) GetByID(id string) (*model.Provider, error) {
	return s.repo.GetByID(id)
}

// GetByType returns the first provider matching the given type.
func (s *ProviderService) GetByType(pType string) (*model.Provider, error) {
	return s.repo.GetByType(pType)
}

// GetDetail returns a provider with its keys and usage summary.
func (s *ProviderService) GetDetail(id string) (*model.ProviderDetail, error) {
	provider, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrProviderNotFound
		}
		return nil, err
	}

	keys, _ := s.keyRepo.ListByProvider(id)
	if keys == nil {
		keys = []model.APIKey{}
	}

	totalCost, totalRequests, _ := s.usageRepo.GetProviderSummary(id)

	return &model.ProviderDetail{
		Provider:      *provider,
		Keys:          keys,
		TotalCost:     totalCost,
		TotalRequests: totalRequests,
	}, nil
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
