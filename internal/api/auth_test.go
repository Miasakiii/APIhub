package api

import (
	"apihub/internal/crypto"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestJWTGenerateAndValidate(t *testing.T) {
	dir := t.TempDir()
	store, _, err := crypto.Init(dir)
	if err != nil {
		t.Fatalf("crypto init: %v", err)
	}

	cfg := AuthConfig{
		Enabled:   true,
		JWTExpiry: time.Hour,
		Store:     store,
	}
	token, err := GenerateTokenForTest(cfg, "user-123")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	userID, err := ValidateTokenForTest(cfg, token)
	if err != nil {
		t.Fatalf("validate token: %v", err)
	}
	if userID != "user-123" {
		t.Fatalf("expected user-123, got %s", userID)
	}
}

func TestJWTExpired(t *testing.T) {
	dir := t.TempDir()
	store, _, err := crypto.Init(dir)
	if err != nil {
		t.Fatalf("crypto init: %v", err)
	}

	cfg := AuthConfig{
		Enabled:   true,
		JWTExpiry: -time.Hour,
		Store:     store,
	}

	token, err := GenerateTokenForTest(cfg, "user-123")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	if _, err := ValidateTokenForTest(cfg, token); err == nil {
		t.Fatal("expected expired token to fail validation")
	}
}

func TestLoadAuthConfigDefaults(t *testing.T) {
	os.Unsetenv("APIHUB_AUTH_ENABLED")
	os.Unsetenv("APIHUB_CORS_ORIGIN")

	dir := t.TempDir()
	store, _, err := crypto.Init(filepath.Join(dir, "data"))
	if err != nil {
		t.Fatalf("crypto init: %v", err)
	}

	cfg := LoadAuthConfig(store)
	if cfg.Enabled {
		t.Fatal("expected auth disabled by default")
	}
	if cfg.CORSOrigin != "http://localhost:5173" {
		t.Fatalf("unexpected cors origin: %s", cfg.CORSOrigin)
	}
}

func TestLoadAuthConfigEnabled(t *testing.T) {
	t.Setenv("APIHUB_AUTH_ENABLED", "true")
	t.Setenv("APIHUB_JWT_EXPIRY", "24h")

	dir := t.TempDir()
	store, _, err := crypto.Init(dir)
	if err != nil {
		t.Fatalf("crypto init: %v", err)
	}

	cfg := LoadAuthConfig(store)
	if !cfg.Enabled {
		t.Fatal("expected auth enabled")
	}
	if cfg.JWTExpiry != 24*time.Hour {
		t.Fatalf("expected 24h expiry, got %v", cfg.JWTExpiry)
	}
}
