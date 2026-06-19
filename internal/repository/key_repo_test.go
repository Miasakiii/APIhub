package repository

import (
	"testing"
)

func TestKeyRepo_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewKeyRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	err := repo.Create("k1", "p1", "hash123", []byte("encrypted"), "My Key")
	if err != nil {
		t.Fatalf("create key: %v", err)
	}

	count, err := repo.CountByHash("hash123")
	if err != nil {
		t.Fatalf("count by hash: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1, got %d", count)
	}
}

func TestKeyRepo_CreateWithSource(t *testing.T) {
	db := setupTestDB(t)
	repo := NewKeyRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	err := repo.CreateWithSource("k1", "p1", "hash456", []byte("enc"), "Auto Key", "ccswitch")
	if err != nil {
		t.Fatalf("create with source: %v", err)
	}

	keys, err := repo.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if keys[0].Source != "ccswitch" {
		t.Fatalf("expected source 'ccswitch', got %q", keys[0].Source)
	}
}

func TestKeyRepo_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewKeyRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	// Empty
	keys, err := repo.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(keys) != 0 {
		t.Fatalf("expected 0, got %d", len(keys))
	}

	// With keys
	repo.Create("k1", "p1", "h1", nil, "Key 1")
	repo.Create("k2", "p1", "h2", nil, "Key 2")

	keys, err = repo.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2, got %d", len(keys))
	}
}

func TestKeyRepo_CountActive(t *testing.T) {
	db := setupTestDB(t)
	repo := NewKeyRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	repo.Create("k1", "p1", "h1", nil, "Key 1")
	repo.Create("k2", "p1", "h2", nil, "Key 2")

	count, err := repo.CountActive()
	if err != nil {
		t.Fatalf("count active: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2, got %d", count)
	}
}

func TestKeyRepo_Revoke(t *testing.T) {
	db := setupTestDB(t)
	repo := NewKeyRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	repo.Create("k1", "p1", "h1", nil, "Key 1")

	if err := repo.Revoke("k1"); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	count, err := repo.CountActive()
	if err != nil {
		t.Fatalf("count active: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 active after revoke, got %d", count)
	}
}

func TestKeyRepo_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewKeyRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	repo.Create("k1", "p1", "h1", nil, "Key 1")

	if err := repo.Delete("k1"); err != nil {
		t.Fatalf("delete: %v", err)
	}

	keys, _ := repo.List()
	if len(keys) != 0 {
		t.Fatalf("expected 0 keys after delete, got %d", len(keys))
	}
}

func TestKeyRepo_GetEncryptedKey(t *testing.T) {
	db := setupTestDB(t)
	repo := NewKeyRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	encrypted := []byte("super-secret-encrypted-data")
	repo.Create("k1", "p1", "h1", encrypted, "Key 1")

	got, err := repo.GetEncryptedKey("k1")
	if err != nil {
		t.Fatalf("get encrypted: %v", err)
	}
	if string(got) != string(encrypted) {
		t.Fatalf("encrypted data mismatch")
	}
}

func TestKeyRepo_ListByProvider(t *testing.T) {
	db := setupTestDB(t)
	repo := NewKeyRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")
	seedProvider(t, db, "p2", "Anthropic", "anthropic")

	repo.Create("k1", "p1", "h1", nil, "OpenAI Key")
	repo.Create("k2", "p1", "h2", nil, "OpenAI Key 2")
	repo.Create("k3", "p2", "h3", nil, "Anthropic Key")

	keys, err := repo.ListByProvider("p1")
	if err != nil {
		t.Fatalf("list by provider: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2, got %d", len(keys))
	}
}

func TestKeyRepo_GetByKeyID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewKeyRepo(db.DB)

	// Insert provider with syncer and base_url
	db.Exec(`INSERT INTO providers (id, name, type, syncer, base_url, enabled) VALUES ('p1', 'OpenAI', 'openai', 'openai', 'https://api.openai.com/v1', 1)`)
	repo.Create("k1", "p1", "h1", []byte("enc"), "Key 1")

	detail, err := repo.GetByKeyID("k1")
	if err != nil {
		t.Fatalf("get by key id: %v", err)
	}
	if detail.ProviderID != "p1" {
		t.Fatalf("expected provider_id 'p1', got %q", detail.ProviderID)
	}
	if detail.Syncer != "openai" {
		t.Fatalf("expected syncer 'openai', got %q", detail.Syncer)
	}
	if detail.BaseURL != "https://api.openai.com/v1" {
		t.Fatalf("unexpected base_url: %q", detail.BaseURL)
	}
}
