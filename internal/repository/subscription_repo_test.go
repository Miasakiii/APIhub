package repository

import (
	"apihub/internal/model"
	"testing"
)

func TestSubscriptionRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSubscriptionRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	err := repo.Create(model.Subscription{
		ID:           "s1",
		ProviderID:   "p1",
		PlanName:     "Pay-as-you-go",
		Price:        0,
		Currency:     "USD",
		BillingCycle: "pay-as-go",
		QuotaType:    "credits",
		QuotaTotal:   100,
		QuotaUsed:    25,
		Status:       "active",
		Source:       "manual",
	})
	if err != nil {
		t.Fatalf("create subscription: %v", err)
	}

	s, err := repo.GetByID("s1")
	if err != nil {
		t.Fatalf("get by id: %v", err)
	}
	if s.PlanName != "Pay-as-you-go" {
		t.Fatalf("expected plan 'Pay-as-you-go', got %q", s.PlanName)
	}
	if s.Provider == nil || s.Provider.Name != "OpenAI" {
		t.Fatal("expected provider name 'OpenAI'")
	}
}

func TestSubscriptionRepo_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSubscriptionRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	// Empty
	subs, err := repo.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(subs) != 0 {
		t.Fatalf("expected 0, got %d", len(subs))
	}

	// With data
	repo.Create(model.Subscription{ID: "s1", ProviderID: "p1", PlanName: "Plan A", Status: "active"})
	repo.Create(model.Subscription{ID: "s2", ProviderID: "p1", PlanName: "Plan B", Status: "active"})

	subs, err = repo.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("expected 2, got %d", len(subs))
	}
}

func TestSubscriptionRepo_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSubscriptionRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	repo.Create(model.Subscription{ID: "s1", ProviderID: "p1", PlanName: "Old Plan", Status: "active"})

	err := repo.Update("s1", model.Subscription{
		ProviderID: "p1",
		PlanName:   "New Plan",
		Status:     "active",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	s, _ := repo.GetByID("s1")
	if s.PlanName != "New Plan" {
		t.Fatalf("expected 'New Plan', got %q", s.PlanName)
	}
}

func TestSubscriptionRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSubscriptionRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	repo.Create(model.Subscription{ID: "s1", ProviderID: "p1", PlanName: "Test"})
	repo.Delete("s1")

	subs, _ := repo.List()
	if len(subs) != 0 {
		t.Fatalf("expected 0 after delete, got %d", len(subs))
	}
}

func TestSubscriptionRepo_UpsertAutoSubscription(t *testing.T) {
	db := setupTestDB(t)
	repo := NewSubscriptionRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	// First insert
	err := repo.UpsertAutoSubscription("p1", "Pay-as-you-go", "USD", 100, 10)
	if err != nil {
		t.Fatalf("upsert (insert): %v", err)
	}

	subs, _ := repo.List()
	if len(subs) != 1 {
		t.Fatalf("expected 1, got %d", len(subs))
	}
	if subs[0].Source != "auto" {
		t.Fatalf("expected source 'auto', got %q", subs[0].Source)
	}

	// Update
	err = repo.UpsertAutoSubscription("p1", "Pay-as-you-go", "USD", 200, 50)
	if err != nil {
		t.Fatalf("upsert (update): %v", err)
	}

	subs, _ = repo.List()
	if len(subs) != 1 {
		t.Fatalf("expected 1 after upsert, got %d", len(subs))
	}
	if subs[0].QuotaTotal != 200 {
		t.Fatalf("expected quota 200, got %f", subs[0].QuotaTotal)
	}
}
