package api

import (
	"net/http"
	"testing"
)

func TestAlertHandler_List_Empty(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/alerts", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	arr := parseJSONArray(t, w)
	if len(arr) != 0 {
		t.Fatalf("expected 0 alerts, got %d", len(arr))
	}
}

func TestAlertHandler_Create(t *testing.T) {
	env := setupTestEnv(t)
	seedProvider(t, env.DB, "p1", "OpenAI", "openai")

	w := doRequest(env.Router, "POST", "/api/v1/alerts", map[string]interface{}{
		"name":       "Low Balance",
		"type":       "balance_low",
		"provider_id": "p1",
		"threshold":  10.0,
		"unit":       "usd",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	result := parseJSON(t, w)
	if result["name"] != "Low Balance" {
		t.Fatalf("expected name 'Low Balance', got %q", result["name"])
	}
}

func TestAlertHandler_Create_InvalidJSON(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "POST", "/api/v1/alerts", nil)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAlertHandler_Update(t *testing.T) {
	env := setupTestEnv(t)
	seedProvider(t, env.DB, "p1", "OpenAI", "openai")

	// Create
	doRequest(env.Router, "POST", "/api/v1/alerts", map[string]interface{}{
		"name": "Old Name", "type": "balance_low", "provider_id": "p1", "threshold": 10,
	})

	// Update (need the ID - get from list)
	w := doRequest(env.Router, "GET", "/api/v1/alerts", nil)
	alerts := parseJSONArray(t, w)
	id := alerts[0].(map[string]interface{})["id"].(string)

	w = doRequest(env.Router, "PUT", "/api/v1/alerts/"+id, map[string]interface{}{
		"name": "New Name", "type": "balance_low", "provider_id": "p1", "threshold": 20,
	})
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAlertHandler_Delete(t *testing.T) {
	env := setupTestEnv(t)
	seedProvider(t, env.DB, "p1", "OpenAI", "openai")

	doRequest(env.Router, "POST", "/api/v1/alerts", map[string]interface{}{
		"name": "Test", "type": "balance_low", "provider_id": "p1", "threshold": 10,
	})

	w := doRequest(env.Router, "GET", "/api/v1/alerts", nil)
	alerts := parseJSONArray(t, w)
	id := alerts[0].(map[string]interface{})["id"].(string)

	w = doRequest(env.Router, "DELETE", "/api/v1/alerts/"+id, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestAlertHandler_History_Empty(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/alerts/history", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	arr := parseJSONArray(t, w)
	if len(arr) != 0 {
		t.Fatalf("expected 0 history, got %d", len(arr))
	}
}
