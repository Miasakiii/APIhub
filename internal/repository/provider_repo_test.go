package repository

import (
	"apihub/internal/model"
	"testing"
)

func TestProviderRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProviderRepo(db.DB)

	err := repo.Create(model.Provider{
		ID:      "p1",
		Name:    "OpenAI",
		Type:    "openai",
		BaseURL: "https://api.openai.com/v1",
	})
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	// Verify it exists
	p, err := repo.GetByID("p1")
	if err != nil {
		t.Fatalf("get provider: %v", err)
	}
	if p.Name != "OpenAI" {
		t.Fatalf("expected name 'OpenAI', got %q", p.Name)
	}
	if p.Type != "openai" {
		t.Fatalf("expected type 'openai', got %q", p.Type)
	}
	if !p.Enabled {
		t.Fatal("expected provider to be enabled")
	}
}

func TestProviderRepo_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProviderRepo(db.DB)

	// Empty list
	list, err := repo.List()
	if err != nil {
		t.Fatalf("list providers: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected 0 providers, got %d", len(list))
	}

	// Add providers
	seedProvider(t, db, "p1", "OpenAI", "openai")
	seedProvider(t, db, "p2", "Anthropic", "anthropic")

	list, err = repo.List()
	if err != nil {
		t.Fatalf("list providers: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(list))
	}
}

func TestProviderRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProviderRepo(db.DB)

	seedProvider(t, db, "p1", "OpenAI", "openai")

	if err := repo.Delete("p1"); err != nil {
		t.Fatalf("delete provider: %v", err)
	}

	_, err := repo.GetByID("p1")
	if err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestProviderRepo_GetByType(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProviderRepo(db.DB)

	seedProvider(t, db, "p1", "OpenAI", "openai")
	seedProvider(t, db, "p2", "Anthropic", "anthropic")

	p, err := repo.GetByType("anthropic")
	if err != nil {
		t.Fatalf("get by type: %v", err)
	}
	if p.ID != "p2" {
		t.Fatalf("expected ID 'p2', got %q", p.ID)
	}
}

func TestProviderRepo_GetByType_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewProviderRepo(db.DB)

	_, err := repo.GetByType("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent type")
	}
}
