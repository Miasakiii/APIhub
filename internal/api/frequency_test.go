package api

import (
	"net/http"
	"testing"
)

func TestFrequencyHandler_Hourly_Empty(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/frequency/hourly", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestFrequencyHandler_Hourly_WithDays(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/frequency/hourly?days=3", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestFrequencyHandler_PeakQPS_Empty(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/frequency/peak-qps", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestFrequencyHandler_Today_Empty(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/frequency/today", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
