package repository

import (
	apihubDB "apihub/internal/db"
	"testing"
	"time"
)

func seedUsageRecord(t *testing.T, db *apihubDB.DB, id, providerID, model, source string, inputTokens, outputTokens int64, cost float64, ts time.Time) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO usage_records (id, provider_id, model, input_tokens, output_tokens, cost_usd, source, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		id, providerID, model, inputTokens, outputTokens, cost, source, ts.Format("2006-01-02 15:04:05"))
	if err != nil {
		t.Fatalf("seed usage: %v", err)
	}
}

func TestUsageRepo_List_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUsageRepo(db.DB)

	records, total, err := repo.List(UsageFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected 0 total, got %d", total)
	}
	if len(records) != 0 {
		t.Fatalf("expected 0 records, got %d", len(records))
	}
}

func TestUsageRepo_List_WithRecords(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUsageRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	now := time.Now()
	seedUsageRecord(t, db, "u1", "p1", "gpt-4", "syncer", 1000, 500, 0.05, now)
	seedUsageRecord(t, db, "u2", "p1", "gpt-4", "syncer", 2000, 1000, 0.10, now)

	records, total, err := repo.List(UsageFilter{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected 2 total, got %d", total)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
}

func TestUsageRepo_List_FilterByProvider(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUsageRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")
	seedProvider(t, db, "p2", "Anthropic", "anthropic")

	now := time.Now()
	seedUsageRecord(t, db, "u1", "p1", "gpt-4", "syncer", 1000, 500, 0.05, now)
	seedUsageRecord(t, db, "u2", "p2", "claude-3", "syncer", 2000, 1000, 0.10, now)

	records, total, err := repo.List(UsageFilter{ProviderID: "p1", Page: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if total != 1 {
		t.Fatalf("expected 1 total, got %d", total)
	}
	if records[0].ProviderID != "p1" {
		t.Fatalf("expected provider p1, got %q", records[0].ProviderID)
	}
}

func TestUsageRepo_GetSummary_Empty(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUsageRepo(db.DB)

	s, err := repo.GetSummary()
	if err != nil {
		t.Fatalf("get summary: %v", err)
	}
	if s.TotalCost != 0 {
		t.Fatalf("expected 0 cost, got %f", s.TotalCost)
	}
	if s.TotalRequests != 0 {
		t.Fatalf("expected 0 requests, got %d", s.TotalRequests)
	}
}

func TestUsageRepo_GetSummary_WithData(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUsageRepo(db.DB)
	seedProvider(t, db, "p1", "OpenAI", "openai")

	now := time.Now()
	seedUsageRecord(t, db, "u1", "p1", "gpt-4", "syncer", 1000, 500, 0.05, now)
	seedUsageRecord(t, db, "u2", "p1", "gpt-4", "syncer", 2000, 1000, 0.10, now)

	s, err := repo.GetSummary()
	if err != nil {
		t.Fatalf("get summary: %v", err)
	}
	if s.TotalCost < 0.149 || s.TotalCost > 0.151 {
		t.Fatalf("expected cost ~0.15, got %f", s.TotalCost)
	}
	if s.TotalRequests != 2 {
		t.Fatalf("expected 2 requests, got %d", s.TotalRequests)
	}
	if s.TotalTokens != 4500 {
		t.Fatalf("expected 4500 tokens, got %d", s.TotalTokens)
	}
}
