package api

import (
	"net/http"
	"testing"
)

func TestUsageHandler_List_Empty(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/usage", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	result := parseJSON(t, w)
	// Usage list returns paginated result
	if result["total"] != float64(0) {
		t.Fatalf("expected total 0, got %v", result["total"])
	}
}

func TestUsageHandler_List_WithPagination(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/usage?page=1&page_size=10", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestUsageHandler_Summary_Empty(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/usage/summary", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	result := parseJSON(t, w)
	if result["total_cost_usd"] != float64(0) {
		t.Fatalf("expected cost 0, got %v", result["total_cost_usd"])
	}
}

func TestStatsHandler_Daily_Empty(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/stats/daily", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	arr := parseJSONArray(t, w)
	if len(arr) != 0 {
		t.Fatalf("expected 0, got %d", len(arr))
	}
}

func TestStatsHandler_CostTrend_Empty(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/stats/cost-trend", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestStatsHandler_ModelBreakdown_Empty(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/stats/model-breakdown", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
