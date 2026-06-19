package api

import (
	"net/http"
	"testing"
)

func TestProviderHandler_List_Empty(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/providers", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	arr := parseJSONArray(t, w)
	if len(arr) != 0 {
		t.Fatalf("expected 0 providers, got %d", len(arr))
	}
}

func TestProviderHandler_Create(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "POST", "/api/v1/providers", map[string]interface{}{
		"name": "OpenAI",
		"type": "openai",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	result := parseJSON(t, w)
	if result["name"] != "OpenAI" {
		t.Fatalf("expected name 'OpenAI', got %q", result["name"])
	}
}

func TestProviderHandler_Create_MissingFields(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "POST", "/api/v1/providers", map[string]interface{}{
		"name": "OpenAI",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProviderHandler_List_AfterCreate(t *testing.T) {
	env := setupTestEnv(t)
	doRequest(env.Router, "POST", "/api/v1/providers", map[string]interface{}{
		"name": "OpenAI", "type": "openai",
	})
	doRequest(env.Router, "POST", "/api/v1/providers", map[string]interface{}{
		"name": "Anthropic", "type": "anthropic",
	})

	w := doRequest(env.Router, "GET", "/api/v1/providers", nil)
	arr := parseJSONArray(t, w)
	if len(arr) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(arr))
	}
}

func TestProviderHandler_Delete(t *testing.T) {
	env := setupTestEnv(t)
	seedProvider(t, env.DB, "p1", "OpenAI", "openai")

	w := doRequest(env.Router, "DELETE", "/api/v1/providers/p1", nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}

	// Verify deleted
	w = doRequest(env.Router, "GET", "/api/v1/providers", nil)
	arr := parseJSONArray(t, w)
	if len(arr) != 0 {
		t.Fatalf("expected 0 after delete, got %d", len(arr))
	}
}
