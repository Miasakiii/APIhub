package scanner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// Finding represents a discovered API key configuration.
type Finding struct {
	ProviderType string `json:"provider_type"` // anthropic, openai, deepseek, gemini, openrouter, kimi
	Name         string `json:"name"`           // display name
	BaseURL      string `json:"base_url"`       // API base URL
	Key          string `json:"key"`            // plaintext API key (never serialized to DB)
	Source       string `json:"source"`          // env, claude, deepseek, kimi, codex
	ConfigPath   string `json:"config_path"`    // source file path (empty for env)
}

// ScanEnv scans environment variables for API keys.
func ScanEnv() []Finding {
	var findings []Finding

	type envRule struct {
		envKey       string
		providerType string
		name         string
		baseURL      string
		urlEnvKey    string // optional env var for base URL override
	}

	rules := []envRule{
		{"ANTHROPIC_API_KEY", "anthropic", "Anthropic", "https://api.anthropic.com", "ANTHROPIC_BASE_URL"},
		{"OPENAI_API_KEY", "openai", "OpenAI", "https://api.openai.com/v1", "OPENAI_BASE_URL"},
		{"DEEPSEEK_API_KEY", "deepseek", "DeepSeek", "https://api.deepseek.com", "DEEPSEEK_BASE_URL"},
		{"GEMINI_API_KEY", "gemini", "Google Gemini", "", ""},
		{"OPENROUTER_API_KEY", "openrouter", "OpenRouter", "https://openrouter.ai/api/v1", "OPENROUTER_BASE_URL"},
	}

	for _, r := range rules {
		key := strings.TrimSpace(os.Getenv(r.envKey))
		if key == "" {
			continue
		}
		baseURL := r.baseURL
		if r.urlEnvKey != "" {
			if u := strings.TrimSpace(os.Getenv(r.urlEnvKey)); u != "" {
				baseURL = u
			}
		}
		findings = append(findings, Finding{
			ProviderType: r.providerType,
			Name:         r.name,
			BaseURL:      baseURL,
			Key:          key,
			Source:       "env",
		})
	}

	return findings
}

// ScanConfigs scans known AI tool config files for API keys.
func ScanConfigs(homeDir string) []Finding {
	var findings []Finding

	if homeDir == "" {
		h, err := os.UserHomeDir()
		if err != nil {
			return findings
		}
		homeDir = h
	}

	// Claude Code: ~/.claude/settings.json
	if f := parseClaudeSettings(homeDir); f != nil {
		findings = append(findings, *f)
	}

	// DeepSeek: ~/.deepseek/config.toml
	if f := parseDeepseekConfig(homeDir); f != nil {
		findings = append(findings, *f)
	}

	// Kimi Code: ~/.kimi-code/config.toml
	findings = append(findings, parseKimiConfig(homeDir)...)

	// Codex: ~/.codex/auth.json
	if f := parseCodexAuth(homeDir); f != nil {
		findings = append(findings, *f)
	}

	return findings
}

// parseClaudeSettings parses ~/.claude/settings.json for ANTHROPIC_AUTH_TOKEN.
func parseClaudeSettings(homeDir string) *Finding {
	path := filepath.Join(homeDir, ".claude", "settings.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var cfg struct {
		Env map[string]string `json:"env"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil || cfg.Env == nil {
		return nil
	}

	// Prefer ANTHROPIC_AUTH_TOKEN, fall back to ANTHROPIC_API_KEY
	key := cfg.Env["ANTHROPIC_AUTH_TOKEN"]
	if key == "" {
		key = cfg.Env["ANTHROPIC_API_KEY"]
	}
	if key == "" {
		return nil
	}

	baseURL := cfg.Env["ANTHROPIC_BASE_URL"]
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}

	return &Finding{
		ProviderType: "anthropic",
		Name:         "Claude Code",
		BaseURL:      baseURL,
		Key:          key,
		Source:       "claude",
		ConfigPath:   path,
	}
}

// parseDeepseekConfig parses ~/.deepseek/config.toml for api_key.
func parseDeepseekConfig(homeDir string) *Finding {
	path := filepath.Join(homeDir, ".deepseek", "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var cfg struct {
		APIKey string `toml:"api_key"`
	}
	if err := toml.Unmarshal(data, &cfg); err != nil || cfg.APIKey == "" {
		return nil
	}

	return &Finding{
		ProviderType: "deepseek",
		Name:         "DeepSeek",
		BaseURL:      "https://api.deepseek.com",
		Key:          cfg.APIKey,
		Source:       "deepseek",
		ConfigPath:   path,
	}
}

// parseKimiConfig parses ~/.kimi-code/config.toml for provider API keys.
func parseKimiConfig(homeDir string) []Finding {
	path := filepath.Join(homeDir, ".kimi-code", "config.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var cfg struct {
		Providers map[string]struct {
			Type   string `toml:"type"`
			APIKey string `toml:"api_key"`
			BaseURL string `toml:"base_url"`
		} `toml:"providers"`
	}
	if err := toml.Unmarshal(data, &cfg); err != nil || cfg.Providers == nil {
		return nil
	}

	var findings []Finding
	for name, p := range cfg.Providers {
		if p.APIKey == "" {
			continue
		}
		pType := p.Type
		if pType == "" {
			pType = "kimi"
		}
		displayName := "Kimi Code"
		if name != "" {
			displayName = "Kimi (" + name + ")"
		}
		findings = append(findings, Finding{
			ProviderType: pType,
			Name:         displayName,
			BaseURL:      p.BaseURL,
			Key:          p.APIKey,
			Source:       "kimi",
			ConfigPath:   path,
		})
	}
	return findings
}

// parseCodexAuth parses ~/.codex/auth.json for OpenAI access token.
func parseCodexAuth(homeDir string) *Finding {
	path := filepath.Join(homeDir, ".codex", "auth.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var cfg struct {
		OpenAIKey string `json:"OPENAI_API_KEY"`
		Tokens    struct {
			AccessToken string `json:"access_token"`
		} `json:"tokens"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil
	}

	// Prefer explicit OPENAI_API_KEY, fall back to OAuth access_token
	key := cfg.OpenAIKey
	if key == "" {
		key = cfg.Tokens.AccessToken
	}
	if key == "" {
		return nil
	}

	return &Finding{
		ProviderType: "openai",
		Name:         "Codex (OpenAI)",
		BaseURL:      "https://api.openai.com/v1",
		Key:          key,
		Source:       "codex",
		ConfigPath:   path,
	}
}
