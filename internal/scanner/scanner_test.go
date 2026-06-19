package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseClaudeSettings(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	// Write test settings.json
	settings := `{
		"env": {
			"ANTHROPIC_AUTH_TOKEN": "sk-test-token-123",
			"ANTHROPIC_BASE_URL": "https://custom.proxy.com/anthropic"
		}
	}`
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(settings), 0644)

	f := parseClaudeSettings(dir)
	if f == nil {
		t.Fatal("expected finding, got nil")
	}
	if f.ProviderType != "anthropic" {
		t.Errorf("type = %q, want anthropic", f.ProviderType)
	}
	if f.Key != "sk-test-token-123" {
		t.Errorf("key = %q, want sk-test-token-123", f.Key)
	}
	if f.BaseURL != "https://custom.proxy.com/anthropic" {
		t.Errorf("base_url = %q, want custom proxy", f.BaseURL)
	}
	if f.Source != "claude" {
		t.Errorf("source = %q, want claude", f.Source)
	}
}

func TestParseClaudeSettings_FallbackToAPIKey(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	settings := `{"env": {"ANTHROPIC_API_KEY": "sk-fallback"}}`
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(settings), 0644)

	f := parseClaudeSettings(dir)
	if f == nil {
		t.Fatal("expected finding, got nil")
	}
	if f.Key != "sk-fallback" {
		t.Errorf("key = %q, want sk-fallback", f.Key)
	}
	if f.BaseURL != "https://api.anthropic.com" {
		t.Errorf("base_url = %q, want default", f.BaseURL)
	}
}

func TestParseClaudeSettings_NoKey(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	settings := `{"env": {"ANTHROPIC_MODEL": "claude-sonnet-4"}}`
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(settings), 0644)

	f := parseClaudeSettings(dir)
	if f != nil {
		t.Errorf("expected nil, got %+v", f)
	}
}

func TestParseDeepseekConfig(t *testing.T) {
	dir := t.TempDir()
	dsDir := filepath.Join(dir, ".deepseek")
	os.MkdirAll(dsDir, 0755)

	toml := `api_key = "sk-deepseek-test"
default_text_model = "deepseek-v4-pro"`
	os.WriteFile(filepath.Join(dsDir, "config.toml"), []byte(toml), 0644)

	f := parseDeepseekConfig(dir)
	if f == nil {
		t.Fatal("expected finding, got nil")
	}
	if f.ProviderType != "deepseek" {
		t.Errorf("type = %q, want deepseek", f.ProviderType)
	}
	if f.Key != "sk-deepseek-test" {
		t.Errorf("key = %q, want sk-deepseek-test", f.Key)
	}
}

func TestParseKimiConfig(t *testing.T) {
	dir := t.TempDir()
	kimiDir := filepath.Join(dir, ".kimi-code")
	os.MkdirAll(kimiDir, 0755)

	toml := `[providers.moonshot-cn]
type = "kimi"
api_key = "sk-kimi-test"
base_url = "https://api.moonshot.cn/v1"

[providers.custom]
type = "openai"
api_key = "sk-custom-key"
base_url = "https://custom.api.com/v1"`
	os.WriteFile(filepath.Join(kimiDir, "config.toml"), []byte(toml), 0644)

	findings := parseKimiConfig(dir)
	if len(findings) != 2 {
		t.Fatalf("expected 2 findings, got %d", len(findings))
	}

	// Find the moonshot-cn entry
	var moonshot, custom *Finding
	for i := range findings {
		switch findings[i].Source {
		case "kimi":
			if findings[i].Key == "sk-kimi-test" {
				moonshot = &findings[i]
			} else if findings[i].Key == "sk-custom-key" {
				custom = &findings[i]
			}
		}
	}

	if moonshot == nil {
		t.Fatal("moonshot finding not found")
	}
	if moonshot.ProviderType != "kimi" {
		t.Errorf("moonshot type = %q, want kimi", moonshot.ProviderType)
	}

	if custom == nil {
		t.Fatal("custom finding not found")
	}
	if custom.ProviderType != "openai" {
		t.Errorf("custom type = %q, want openai", custom.ProviderType)
	}
}

func TestParseCodexAuth(t *testing.T) {
	dir := t.TempDir()
	codexDir := filepath.Join(dir, ".codex")
	os.MkdirAll(codexDir, 0755)

	auth := `{
		"OPENAI_API_KEY": null,
		"tokens": {
			"access_token": "test-access-token"
		}
	}`
	os.WriteFile(filepath.Join(codexDir, "auth.json"), []byte(auth), 0644)

	f := parseCodexAuth(dir)
	if f == nil {
		t.Fatal("expected finding, got nil")
	}
	if f.ProviderType != "openai" {
		t.Errorf("type = %q, want openai", f.ProviderType)
	}
	if f.Key != "test-access-token" {
		t.Errorf("key = %q, want test-access-token", f.Key)
	}
}

func TestParseCodexAuth_WithAPIKey(t *testing.T) {
	dir := t.TempDir()
	codexDir := filepath.Join(dir, ".codex")
	os.MkdirAll(codexDir, 0755)

	auth := `{
		"OPENAI_API_KEY": "sk-explicit-key",
		"tokens": { "access_token": "fallback" }
	}`
	os.WriteFile(filepath.Join(codexDir, "auth.json"), []byte(auth), 0644)

	f := parseCodexAuth(dir)
	if f == nil {
		t.Fatal("expected finding, got nil")
	}
	if f.Key != "sk-explicit-key" {
		t.Errorf("key = %q, want sk-explicit-key", f.Key)
	}
}

func TestScanConfigs(t *testing.T) {
	dir := t.TempDir()

	// Set up Claude config
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0755)
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), []byte(`{"env":{"ANTHROPIC_AUTH_TOKEN":"sk-claude"}}`), 0644)

	// Set up DeepSeek config
	dsDir := filepath.Join(dir, ".deepseek")
	os.MkdirAll(dsDir, 0755)
	os.WriteFile(filepath.Join(dsDir, "config.toml"), []byte(`api_key = "sk-ds"`), 0644)

	findings := ScanConfigs(dir)
	if len(findings) < 2 {
		t.Errorf("expected >= 2 findings, got %d", len(findings))
	}

	sources := make(map[string]bool)
	for _, f := range findings {
		sources[f.Source] = true
	}
	if !sources["claude"] {
		t.Error("missing claude source")
	}
	if !sources["deepseek"] {
		t.Error("missing deepseek source")
	}
}

func TestScanEnv(t *testing.T) {
	// Save and restore env
	oldKey := os.Getenv("OPENAI_API_KEY")
	oldURL := os.Getenv("OPENAI_BASE_URL")
	defer func() {
		os.Setenv("OPENAI_API_KEY", oldKey)
		os.Setenv("OPENAI_BASE_URL", oldURL)
	}()

	os.Setenv("OPENAI_API_KEY", "sk-test-env")
	os.Setenv("OPENAI_BASE_URL", "https://custom.openai.com/v1")

	findings := ScanEnv()
	var found *Finding
	for i := range findings {
		if findings[i].ProviderType == "openai" {
			found = &findings[i]
			break
		}
	}
	if found == nil {
		t.Fatal("openai finding not found in env scan")
	}
	if found.Key != "sk-test-env" {
		t.Errorf("key = %q, want sk-test-env", found.Key)
	}
	if found.BaseURL != "https://custom.openai.com/v1" {
		t.Errorf("base_url = %q, want custom URL", found.BaseURL)
	}
	if found.Source != "env" {
		t.Errorf("source = %q, want env", found.Source)
	}
}
