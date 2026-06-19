package service

import (
	"apihub/internal/model"
	"database/sql"
	"errors"
	"testing"
)

// mockSubscriptionRepo is a mock implementation of subscription repository.
type mockSubscriptionRepo struct {
	subs     []model.Subscription
	listErr  error
	createErr error
	getErr    error
	updateErr error
	deleteErr error
}

func (m *mockSubscriptionRepo) List() ([]model.Subscription, error) {
	return m.subs, m.listErr
}

func (m *mockSubscriptionRepo) GetByID(id string) (*model.Subscription, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, s := range m.subs {
		if s.ID == id {
			return &s, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockSubscriptionRepo) Create(s model.Subscription) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.subs = append(m.subs, s)
	return nil
}

func (m *mockSubscriptionRepo) Update(id string, s model.Subscription) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	for i, sub := range m.subs {
		if sub.ID == id {
			m.subs[i] = s
			m.subs[i].ID = id
			return nil
		}
	}
	return errors.New("not found")
}

func (m *mockSubscriptionRepo) Delete(id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for i, s := range m.subs {
		if s.ID == id {
			m.subs = append(m.subs[:i], m.subs[i+1:]...)
			return nil
		}
	}
	return nil
}

func TestSubscriptionService_List(t *testing.T) {
	repo := &mockSubscriptionRepo{
		subs: []model.Subscription{
			{ID: "1", PlanName: "Pro", Price: 20.0},
			{ID: "2", PlanName: "Enterprise", Price: 100.0},
		},
	}
	svc := NewSubscriptionService(repo)

	subs, err := svc.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("expected 2 subscriptions, got %d", len(subs))
	}
}

func TestSubscriptionService_GetByID(t *testing.T) {
	repo := &mockSubscriptionRepo{
		subs: []model.Subscription{
			{ID: "1", PlanName: "Pro", Price: 20.0},
		},
	}
	svc := NewSubscriptionService(repo)

	sub, err := svc.GetByID("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.PlanName != "Pro" {
		t.Errorf("expected plan name 'Pro', got '%s'", sub.PlanName)
	}
}

func TestSubscriptionService_GetByID_NotFound(t *testing.T) {
	repo := &mockSubscriptionRepo{}
	svc := NewSubscriptionService(repo)

	_, err := svc.GetByID("nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != ErrSubscriptionNotFound {
		t.Errorf("expected ErrSubscriptionNotFound, got %v", err)
	}
}

func TestSubscriptionService_Create(t *testing.T) {
	repo := &mockSubscriptionRepo{}
	svc := NewSubscriptionService(repo)

	sub, err := svc.Create(model.Subscription{
		ProviderID: "p1",
		PlanName:   "Pro",
		Price:      20.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.ID == "" {
		t.Error("expected ID to be generated")
	}
	if len(repo.subs) != 1 {
		t.Fatalf("expected 1 subscription in repo, got %d", len(repo.subs))
	}
}

func TestSubscriptionService_Update(t *testing.T) {
	repo := &mockSubscriptionRepo{
		subs: []model.Subscription{
			{ID: "1", PlanName: "Old Plan", Price: 10.0},
		},
	}
	svc := NewSubscriptionService(repo)

	err := svc.Update("1", model.Subscription{
		PlanName: "New Plan",
		Price:    25.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.subs[0].PlanName != "New Plan" {
		t.Errorf("expected plan name 'New Plan', got '%s'", repo.subs[0].PlanName)
	}
}

func TestSubscriptionService_Delete(t *testing.T) {
	repo := &mockSubscriptionRepo{
		subs: []model.Subscription{
			{ID: "1", PlanName: "Pro"},
		},
	}
	svc := NewSubscriptionService(repo)

	err := svc.Delete("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.subs) != 0 {
		t.Errorf("expected 0 subscriptions after delete, got %d", len(repo.subs))
	}
}
