package repository

import (
	apihubDB "apihub/internal/db"
	"apihub/internal/model"
	"testing"
)

func seedAlert(t *testing.T, db *apihubDB.DB, id, name, alertType, providerID string, threshold float64) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO alerts (id, name, type, provider_id, threshold, unit, enabled) VALUES (?, ?, ?, ?, ?, 'usd', 1)`,
		id, name, alertType, providerID, threshold)
	if err != nil {
		t.Fatalf("seed alert: %v", err)
	}
}

func TestAlertRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAlertRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	err := repo.Create(model.Alert{
		ID:         "a1",
		Name:       "Low Balance",
		Type:       "balance_low",
		ProviderID: "p1",
		Threshold:  10.0,
		Unit:       "usd",
	})
	if err != nil {
		t.Fatalf("create alert: %v", err)
	}

	alerts, err := repo.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(alerts) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(alerts))
	}
	if alerts[0].Name != "Low Balance" {
		t.Fatalf("expected name 'Low Balance', got %q", alerts[0].Name)
	}
	if !alerts[0].Enabled {
		t.Fatal("expected alert to be enabled")
	}
}

func TestAlertRepo_List_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAlertRepo(db.DB)

	alerts, err := repo.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected 0, got %d", len(alerts))
	}
}

func TestAlertRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAlertRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	seedAlert(t, db, "a1", "Old Name", "balance_low", "p1", 10)

	err := repo.Update("a1", model.Alert{
		Name:      "New Name",
		Type:      "balance_low",
		Threshold: 20,
		Enabled:   true,
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	alerts, err := repo.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(alerts) == 0 {
		t.Fatal("expected at least 1 alert after update")
	}
	if alerts[0].Name != "New Name" {
		t.Fatalf("expected 'New Name', got %q", alerts[0].Name)
	}
	if alerts[0].Threshold != 20 {
		t.Fatalf("expected threshold 20, got %f", alerts[0].Threshold)
	}
}

func TestAlertRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAlertRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	seedAlert(t, db, "a1", "Test", "balance_low", "p1", 10)
	repo.Delete("a1")

	alerts, _ := repo.List()
	if len(alerts) != 0 {
		t.Fatalf("expected 0 after delete, got %d", len(alerts))
	}
}

func TestAlertRepo_ListHistory(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAlertRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	seedAlert(t, db, "a1", "Test", "balance_low", "p1", 10)

	// Insert history
	db.Exec(`INSERT INTO alert_history (id, alert_id, message, level) VALUES ('h1', 'a1', 'Balance low', 'warning')`)
	db.Exec(`INSERT INTO alert_history (id, alert_id, message, level) VALUES ('h2', 'a1', 'Balance critical', 'critical')`)

	history, err := repo.ListHistory("")
	if err != nil {
		t.Fatalf("list history: %v", err)
	}
	if len(history) != 2 {
		t.Fatalf("expected 2, got %d", len(history))
	}

	// Filter by alert ID
	filtered, err := repo.ListHistory("a1")
	if err != nil {
		t.Fatalf("list history filtered: %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2, got %d", len(filtered))
	}
}
