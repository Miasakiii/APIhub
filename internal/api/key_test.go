package api

import (
	"net/http"
	"testing"
)

func TestKeyHandler_List_Empty(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/keys", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	arr := parseJSONArray(t, w)
	if len(arr) != 0 {
		t.Fatalf("expected 0, got %d", len(arr))
	}
}

func TestKeyHandler_Create(t *testing.T) {
	env := setupTestEnv(t)
	seedProvider(t, env.DB, "p1", "OpenAI", "openai")

	w := doRequest(env.Router, "POST", "/api/v1/keys", map[string]interface{}{
		"provider_id": "p1",
		"key":         "sk-test1234567890abcdef",
		"name":        "My Key",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestKeyHandler_Create_Duplicate(t *testing.T) {
	env := setupTestEnv(t)
	seedProvider(t, env.DB, "p1", "OpenAI", "openai")

	doRequest(env.Router, "POST", "/api/v1/keys", map[string]interface{}{
		"provider_id": "p1", "key": "sk-samekey1234567890", "name": "Key 1",
	})
	w := doRequest(env.Router, "POST", "/api/v1/keys", map[string]interface{}{
		"provider_id": "p1", "key": "sk-samekey1234567890", "name": "Key 2",
	})

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestKeyHandler_Create_MissingFields(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "POST", "/api/v1/keys", map[string]interface{}{
		"name": "No Provider",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestKeyHandler_Revoke(t *testing.T) {
	env := setupTestEnv(t)
	seedProvider(t, env.DB, "p1", "OpenAI", "openai")

	doRequest(env.Router, "POST", "/api/v1/keys", map[string]interface{}{
		"provider_id": "p1", "key": "sk-revoke1234567890", "name": "Key",
	})

	w := doRequest(env.Router, "GET", "/api/v1/keys", nil)
	keys := parseJSONArray(t, w)
	id := keys[0].(map[string]interface{})["id"].(string)

	w = doRequest(env.Router, "POST", "/api/v1/keys/"+id+"/revoke", nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestKeyHandler_Delete(t *testing.T) {
	env := setupTestEnv(t)
	seedProvider(t, env.DB, "p1", "OpenAI", "openai")

	doRequest(env.Router, "POST", "/api/v1/keys", map[string]interface{}{
		"provider_id": "p1", "key": "sk-delete1234567890", "name": "Key",
	})

	w := doRequest(env.Router, "GET", "/api/v1/keys", nil)
	keys := parseJSONArray(t, w)
	id := keys[0].(map[string]interface{})["id"].(string)

	w = doRequest(env.Router, "DELETE", "/api/v1/keys/"+id, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}
