package api

import (
	"net/http"
	"testing"
)

func TestSubscriptionHandler_List_Empty(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/subscriptions", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	arr := parseJSONArray(t, w)
	if len(arr) != 0 {
		t.Fatalf("expected 0, got %d", len(arr))
	}
}

func TestSubscriptionHandler_Create(t *testing.T) {
	env := setupTestEnv(t)
	seedProvider(t, env.DB, "p1", "OpenAI", "openai")

	w := doRequest(env.Router, "POST", "/api/v1/subscriptions", map[string]interface{}{
		"provider_id":   "p1",
		"plan_name":     "Pay-as-you-go",
		"price":         0,
		"currency":      "USD",
		"billing_cycle": "pay-as-go",
		"quota_type":    "credits",
		"quota_total":   100,
		"status":        "active",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	result := parseJSON(t, w)
	if result["plan_name"] != "Pay-as-you-go" {
		t.Fatalf("expected plan 'Pay-as-you-go', got %q", result["plan_name"])
	}
}

func TestSubscriptionHandler_GetByID(t *testing.T) {
	env := setupTestEnv(t)
	seedProvider(t, env.DB, "p1", "OpenAI", "openai")

	doRequest(env.Router, "POST", "/api/v1/subscriptions", map[string]interface{}{
		"provider_id": "p1", "plan_name": "Plan A", "status": "active",
	})

	w := doRequest(env.Router, "GET", "/api/v1/subscriptions", nil)
	subs := parseJSONArray(t, w)
	id := subs[0].(map[string]interface{})["id"].(string)

	w = doRequest(env.Router, "GET", "/api/v1/subscriptions/"+id, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestSubscriptionHandler_GetByID_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/subscriptions/nonexistent", nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestSubscriptionHandler_Update(t *testing.T) {
	env := setupTestEnv(t)
	seedProvider(t, env.DB, "p1", "OpenAI", "openai")

	doRequest(env.Router, "POST", "/api/v1/subscriptions", map[string]interface{}{
		"provider_id": "p1", "plan_name": "Old", "status": "active",
	})

	w := doRequest(env.Router, "GET", "/api/v1/subscriptions", nil)
	subs := parseJSONArray(t, w)
	id := subs[0].(map[string]interface{})["id"].(string)

	w = doRequest(env.Router, "PUT", "/api/v1/subscriptions/"+id, map[string]interface{}{
		"provider_id": "p1", "plan_name": "New", "status": "active",
	})
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSubscriptionHandler_Delete(t *testing.T) {
	env := setupTestEnv(t)
	seedProvider(t, env.DB, "p1", "OpenAI", "openai")

	doRequest(env.Router, "POST", "/api/v1/subscriptions", map[string]interface{}{
		"provider_id": "p1", "plan_name": "Test", "status": "active",
	})

	w := doRequest(env.Router, "GET", "/api/v1/subscriptions", nil)
	subs := parseJSONArray(t, w)
	id := subs[0].(map[string]interface{})["id"].(string)

	w = doRequest(env.Router, "DELETE", "/api/v1/subscriptions/"+id, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}
