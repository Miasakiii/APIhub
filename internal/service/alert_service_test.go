package service

import (
	"apihub/internal/model"
	"errors"
	"testing"
)

// mockAlertRepo is a mock implementation of alert repository.
type mockAlertRepo struct {
	alerts   []model.Alert
	history  []model.AlertHistory
	listErr  error
	createErr error
	updateErr error
	deleteErr error
}

func (m *mockAlertRepo) List() ([]model.Alert, error) {
	return m.alerts, m.listErr
}

func (m *mockAlertRepo) Create(a model.Alert) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.alerts = append(m.alerts, a)
	return nil
}

func (m *mockAlertRepo) Update(id string, a model.Alert) error {
	if m.updateErr != nil {
		return m.updateErr
	}
	for i, alert := range m.alerts {
		if alert.ID == id {
			m.alerts[i] = a
			m.alerts[i].ID = id
			return nil
		}
	}
	return errors.New("not found")
}

func (m *mockAlertRepo) Delete(id string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	for i, a := range m.alerts {
		if a.ID == id {
			m.alerts = append(m.alerts[:i], m.alerts[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockAlertRepo) ListHistory(alertID string) ([]model.AlertHistory, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if alertID == "" {
		return m.history, nil
	}
	var filtered []model.AlertHistory
	for _, h := range m.history {
		if h.AlertID == alertID {
			filtered = append(filtered, h)
		}
	}
	return filtered, nil
}

func TestAlertService_List(t *testing.T) {
	repo := &mockAlertRepo{
		alerts: []model.Alert{
			{ID: "1", Name: "Low Balance", Type: "balance_low"},
			{ID: "2", Name: "Key Expired", Type: "key_expired"},
		},
	}
	svc := NewAlertService(repo)

	alerts, err := svc.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(alerts) != 2 {
		t.Fatalf("expected 2 alerts, got %d", len(alerts))
	}
}

func TestAlertService_Create(t *testing.T) {
	repo := &mockAlertRepo{}
	svc := NewAlertService(repo)

	alert, err := svc.Create(model.Alert{
		Name:      "Low Balance",
		Type:      "balance_low",
		Threshold: 10.0,
		Unit:      "usd",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if alert.ID == "" {
		t.Error("expected ID to be generated")
	}
	if len(repo.alerts) != 1 {
		t.Fatalf("expected 1 alert in repo, got %d", len(repo.alerts))
	}
}

func TestAlertService_Update(t *testing.T) {
	repo := &mockAlertRepo{
		alerts: []model.Alert{
			{ID: "1", Name: "Old Name", Type: "balance_low"},
		},
	}
	svc := NewAlertService(repo)

	err := svc.Update("1", model.Alert{
		Name:      "New Name",
		Type:      "balance_low",
		Threshold: 20.0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.alerts[0].Name != "New Name" {
		t.Errorf("expected name 'New Name', got '%s'", repo.alerts[0].Name)
	}
}

func TestAlertService_Delete(t *testing.T) {
	repo := &mockAlertRepo{
		alerts: []model.Alert{
			{ID: "1", Name: "Test Alert"},
		},
	}
	svc := NewAlertService(repo)

	err := svc.Delete("1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(repo.alerts) != 0 {
		t.Errorf("expected 0 alerts after delete, got %d", len(repo.alerts))
	}
}

func TestAlertService_ListHistory(t *testing.T) {
	repo := &mockAlertRepo{
		history: []model.AlertHistory{
			{ID: "1", AlertID: "a1", Message: "Balance low", Level: "warning"},
			{ID: "2", AlertID: "a2", Message: "Key expired", Level: "critical"},
			{ID: "3", AlertID: "a1", Message: "Balance very low", Level: "critical"},
		},
	}
	svc := NewAlertService(repo)

	// List all
	all, err := svc.ListHistory("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 history entries, got %d", len(all))
	}

	// Filter by alert ID
	filtered, err := svc.ListHistory("a1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 history entries for alert 'a1', got %d", len(filtered))
	}
}
