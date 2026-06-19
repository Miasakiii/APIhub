package api

import (
	"net/http"
	"testing"
)

func TestAgentHandler_List_Empty(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/agents", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	arr := parseJSONArray(t, w)
	if len(arr) != 0 {
		t.Fatalf("expected 0, got %d", len(arr))
	}
}

func TestAgentHandler_Create(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "POST", "/api/v1/agents", map[string]interface{}{
		"name": "Claude Code",
		"type": "cli",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	result := parseJSON(t, w)
	if result["name"] != "Claude Code" {
		t.Fatalf("expected name 'Claude Code', got %q", result["name"])
	}
}

func TestAgentHandler_Create_DefaultType(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "POST", "/api/v1/agents", map[string]interface{}{
		"name": "My Agent",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	result := parseJSON(t, w)
	if result["type"] != "cli" {
		t.Fatalf("expected default type 'cli', got %q", result["type"])
	}
}

func TestAgentHandler_Create_MissingName(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "POST", "/api/v1/agents", map[string]interface{}{
		"type": "cli",
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAgentHandler_GetByID(t *testing.T) {
	env := setupTestEnv(t)
	doRequest(env.Router, "POST", "/api/v1/agents", map[string]interface{}{
		"name": "Test Agent", "type": "cli",
	})

	w := doRequest(env.Router, "GET", "/api/v1/agents", nil)
	agents := parseJSONArray(t, w)
	id := agents[0].(map[string]interface{})["id"].(string)

	w = doRequest(env.Router, "GET", "/api/v1/agents/"+id, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAgentHandler_GetByID_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	w := doRequest(env.Router, "GET", "/api/v1/agents/nonexistent", nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestAgentHandler_Update(t *testing.T) {
	env := setupTestEnv(t)
	doRequest(env.Router, "POST", "/api/v1/agents", map[string]interface{}{
		"name": "Old Name", "type": "cli",
	})

	w := doRequest(env.Router, "GET", "/api/v1/agents", nil)
	agents := parseJSONArray(t, w)
	id := agents[0].(map[string]interface{})["id"].(string)

	w = doRequest(env.Router, "PUT", "/api/v1/agents/"+id, map[string]interface{}{
		"name": "New Name", "type": "ide",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAgentHandler_Delete(t *testing.T) {
	env := setupTestEnv(t)
	doRequest(env.Router, "POST", "/api/v1/agents", map[string]interface{}{
		"name": "To Delete", "type": "cli",
	})

	w := doRequest(env.Router, "GET", "/api/v1/agents", nil)
	agents := parseJSONArray(t, w)
	id := agents[0].(map[string]interface{})["id"].(string)

	w = doRequest(env.Router, "DELETE", "/api/v1/agents/"+id, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
