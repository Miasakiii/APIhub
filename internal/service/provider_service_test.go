package service

import (
	"apihub/internal/model"
	"errors"
	"testing"
)

// mockProviderRepo is a mock implementation of provider repository.
type mockProviderRepo struct {
	providers []model.Provider
	listErr   error
	createErr error
	deleteErr error
}

func (m *mockProviderRepo) List() ([]model.Provider, error) {
	return m.providers, m.listErr
}

func (m *mockProviderRepo) Create(p model.Provider) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.providers = append(m.providers, p)
	return nil
}

func (m *mockProviderRepo) Delete(id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for i, p := range m.providers {
		if p.ID == id {
			m.providers = append(m.providers[:i], m.providers[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockProviderRepo) GetByID(id string) (*model.Provider, error) {
	for _, p := range m.providers {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, errors.New("not found")
}

func (m *mockProviderRepo) GetByType(pType string) (*model.Provider, error) {
	for _, p := range m.providers {
		if p.Type == pType {
			return &p, nil
		}
	}
	return nil, errors.New("not found")
}

// mockProviderKeyRepo is a mock for ProviderKeyRepository.
type mockProviderKeyRepo struct {
	keys []model.APIKey
}

func (m *mockProviderKeyRepo) ListByProvider(providerID string) ([]model.APIKey, error) {
	return m.keys, nil
}

// mockProviderUsageRepo is a mock for ProviderUsageRepository.
type mockProviderUsageRepo struct {
	cost     float64
	requests int64
}

func (m *mockProviderUsageRepo) GetProviderSummary(providerID string) (float64, int64, error) {
	return m.cost, m.requests, nil
}

func newTestProviderService(repo *mockProviderRepo) *ProviderService {
	return NewProviderService(repo, &mockProviderKeyRepo{}, &mockProviderUsageRepo{})
}

func TestProviderService_List(t *testing.T) {
	repo := &mockProviderRepo{
		providers: []model.Provider{
			{ID: "1", Name: "OpenAI", Type: "openai"},
			{ID: "2", Name: "Anthropic", Type: "anthropic"},
		},
	}
	svc := newTestProviderService(repo)

	providers, err := svc.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}
	if providers[0].Name != "OpenAI" {
		t.Errorf("expected first provider name 'OpenAI', got '%s'", providers[0].Name)
	}
}

func TestProviderService_List_Error(t *testing.T) {
	repo := &mockProviderRepo{
		listErr: errors.New("database error"),
	}
	svc := newTestProviderService(repo)

	_, err := svc.List()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProviderService_Create(t *testing.T) {
	repo := &mockProviderRepo{}
	svc := newTestProviderService(repo)

	provider, err := svc.Create(model.Provider{
		Name: "TestProvider",
		Type: "openai",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider.ID == "" {
		t.Error("expected ID to be generated")
	}
	if !provider.Enabled {
		t.Error("expected provider to be enabled")
	}
	if len(repo.providers) != 1 {
		t.Fatalf("expected 1 provider in repo, got %d", len(repo.providers))
	}
}

func TestProviderService_Create_Error(t *testing.T) {
	repo := &mockProviderRepo{
		createErr: errors.New("database error"),
	}
	svc := newTestProviderService(repo)

	_, err := svc.Create(model.Provider{Name: "Test"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProviderService_Delete(t *testing.T) {
	repo := &mockProviderRepo{
		providers: []model.Provider{
			{ID: "1", Name: "OpenAI"},
		},
	}
	svc := newTestProviderService(repo)

	err := svc.Delete("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.providers) != 0 {
		t.Errorf("expected 0 providers after delete, got %d", len(repo.providers))
	}
}

func TestProviderService_GetByID(t *testing.T) {
	repo := &mockProviderRepo{
		providers: []model.Provider{
			{ID: "1", Name: "OpenAI"},
		},
	}
	svc := newTestProviderService(repo)

	provider, err := svc.GetByID("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider.Name != "OpenAI" {
		t.Errorf("expected name 'OpenAI', got '%s'", provider.Name)
	}
}

func TestProviderService_GetByID_NotFound(t *testing.T) {
	repo := &mockProviderRepo{}
	svc := newTestProviderService(repo)

	_, err := svc.GetByID("nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestProviderService_GetDetail(t *testing.T) {
	repo := &mockProviderRepo{
		providers: []model.Provider{
			{ID: "1", Name: "OpenAI", Type: "openai"},
		},
	}
	keyRepo := &mockProviderKeyRepo{
		keys: []model.APIKey{
			{ID: "k1", ProviderID: "1", Name: "key1"},
		},
	}
	usageRepo := &mockProviderUsageRepo{cost: 1.5, requests: 100}
	svc := NewProviderService(repo, keyRepo, usageRepo)

	detail, err := svc.GetDetail("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.Provider.Name != "OpenAI" {
		t.Errorf("expected provider name 'OpenAI', got '%s'", detail.Provider.Name)
	}
	if len(detail.Keys) != 1 {
		t.Errorf("expected 1 key, got %d", len(detail.Keys))
	}
	if detail.TotalCost != 1.5 {
		t.Errorf("expected cost 1.5, got %f", detail.TotalCost)
	}
	if detail.TotalRequests != 100 {
		t.Errorf("expected 100 requests, got %d", detail.TotalRequests)
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	if id1 == "" {
		t.Error("expected non-empty ID")
	}
	if id1 == id2 {
		t.Error("expected unique IDs")
	}
	if len(id1) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("expected 32 char hex ID, got %d chars", len(id1))
	}
}
