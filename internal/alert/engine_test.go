package alert

import (
	apihubDB "apihub/internal/db"
	"apihub/internal/ws"
	"testing"
)

func setupTestDB(t *testing.T) *apihubDB.DB {
	t.Helper()
	db, err := apihubDB.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestNewEngine(t *testing.T) {
	db := setupTestDB(t)
	engine := NewEngine(db.DB)
	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}
}

func TestEngine_SetHub(t *testing.T) {
	db := setupTestDB(t)
	engine := NewEngine(db.DB)
	hub := ws.NewHub()
	engine.SetHub(hub)
	// Should not panic
}

func TestEngine_SetNotifier(t *testing.T) {
	db := setupTestDB(t)
	engine := NewEngine(db.DB)
	n := &Notifier{}
	engine.SetNotifier(n)
	// Should not panic
}

func TestEngine_RunOnce_NoRules(t *testing.T) {
	db := setupTestDB(t)
	engine := NewEngine(db.DB)

	err := engine.RunOnce()
	if err != nil {
		t.Fatalf("RunOnce with no rules: %v", err)
	}
}

func TestEngine_RunOnce_WithRules(t *testing.T) {
	db := setupTestDB(t)
	engine := NewEngine(db.DB)

	// Seed a provider and an alert rule
	db.Exec(`INSERT INTO providers (id, name, type, enabled) VALUES ('p1', 'OpenAI', 'openai', 1)`)
	db.Exec(`INSERT INTO alerts (id, name, type, provider_id, threshold, unit, enabled) VALUES ('a1', 'Low Balance', 'balance_low', 'p1', 10, 'usd', 1)`)

	err := engine.RunOnce()
	if err != nil {
		t.Fatalf("RunOnce: %v", err)
	}
}

func TestEngine_RunOnce_DisabledRule(t *testing.T) {
	db := setupTestDB(t)
	engine := NewEngine(db.DB)

	db.Exec(`INSERT INTO providers (id, name, type, enabled) VALUES ('p1', 'OpenAI', 'openai', 1)`)
	db.Exec(`INSERT INTO alerts (id, name, type, provider_id, threshold, unit, enabled) VALUES ('a1', 'Low Balance', 'balance_low', 'p1', 10, 'usd', 0)`)

	err := engine.RunOnce()
	if err != nil {
		t.Fatalf("RunOnce with disabled rule: %v", err)
	}
}

func TestEngine_listEnabledRules(t *testing.T) {
	db := setupTestDB(t)
	engine := NewEngine(db.DB)

	db.Exec(`INSERT INTO providers (id, name, type, enabled) VALUES ('p1', 'OpenAI', 'openai', 1)`)
	db.Exec(`INSERT INTO alerts (id, name, type, provider_id, threshold, unit, enabled) VALUES ('a1', 'Rule 1', 'balance_low', 'p1', 10, 'usd', 1)`)
	db.Exec(`INSERT INTO alerts (id, name, type, provider_id, threshold, unit, enabled) VALUES ('a2', 'Rule 2', 'balance_low', 'p1', 20, 'usd', 0)`)
	db.Exec(`INSERT INTO alerts (id, name, type, provider_id, threshold, unit, enabled) VALUES ('a3', 'Rule 3', 'balance_low', 'p1', 30, 'usd', 1)`)

	rules, err := engine.listEnabledRules()
	if err != nil {
		t.Fatalf("listEnabledRules: %v", err)
	}
	if len(rules) != 2 {
		t.Fatalf("expected 2 enabled rules, got %d", len(rules))
	}
}
