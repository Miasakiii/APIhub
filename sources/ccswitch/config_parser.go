package ccswitch

import (
	"encoding/json"
	"fmt"
)

// ProviderConfig holds the parsed configuration for a cc-switch provider.
type ProviderConfig struct {
	ID        string
	AppType   string
	Name      string
	Category  string // "official" or "third_party"
	BaseURL   string
	APIKey    string // from settings_config env / top-level field
	RawConfig json.RawMessage
}

// FetchProviders returns all providers from cc-switch.db.
func (r *Reader) FetchProviders() ([]ProviderConfig, error) {
	rows, err := r.db.Query(`
		SELECT id, app_type, name, category, settings_config
		FROM providers
		ORDER BY app_type, name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []ProviderConfig
	for rows.Next() {
		var p ProviderConfig
		var cfgStr string
		var cat sqlNullStr
		if err := rows.Scan(&p.ID, &p.AppType, &p.Name, &cat, &cfgStr); err != nil {
			return nil, err
		}
		p.Category = cat.s
		p.RawConfig = json.RawMessage(cfgStr)
		cfg, err := ParseSettingsConfig(p.AppType, p.RawConfig)
		if err != nil {
			// Log but don't fail — some providers have minimal config
			cfg = &ParsedSettings{}
		}
		p.BaseURL = cfg.BaseURL
		p.APIKey = cfg.APIKey
		providers = append(providers, p)
	}
	return providers, rows.Err()
}

// ParsedSettings holds the extracted endpoint + auth from settings_config JSON.
type ParsedSettings struct {
	BaseURL string
	APIKey  string
}

// ParseSettingsConfig extracts base_url and api_key from the settings_config JSON.
// The field path differs per app_type (verified by spike rounds 2-3):
//   - claude:    env.ANTHROPIC_BASE_URL / env.ANTHROPIC_AUTH_TOKEN
//   - openclaw:  baseUrl / apiKey
//   - hermes:    base_url / api_key
//   - others:    recursive walk for *_URL + *_KEY / *_TOKEN
func ParseSettingsConfig(appType string, raw json.RawMessage) (*ParsedSettings, error) {
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("parse settings_config: %w", err)
	}

	switch appType {
	case "claude":
		return parseClaude(m), nil
	case "openclaw":
		return parseOpenclaw(m), nil
	case "hermes":
		return parseHermes(m), nil
	default:
		return walkGeneric(m), nil
	}
}

func parseClaude(m map[string]any) *ParsedSettings {
	if env, ok := m["env"].(map[string]any); ok {
		return &ParsedSettings{
			BaseURL: str(env, "ANTHROPIC_BASE_URL"),
			APIKey:  str(env, "ANTHROPIC_AUTH_TOKEN"),
		}
	}
	return &ParsedSettings{}
}

func parseOpenclaw(m map[string]any) *ParsedSettings {
	return &ParsedSettings{
		BaseURL: str(m, "baseUrl"),
		APIKey:  str(m, "apiKey"),
	}
}

func parseHermes(m map[string]any) *ParsedSettings {
	return &ParsedSettings{
		BaseURL: str(m, "base_url"),
		APIKey:  str(m, "api_key"),
	}
}

func walkGeneric(m map[string]any) *ParsedSettings {
	var baseURL, apiKey string
	walk(m, func(key string, val string) {
		lk := lower(key)
		if baseURL == "" && (contains(lk, "base") || contains(lk, "url")) && isHTTP(val) {
			baseURL = val
		}
		if apiKey == "" && (contains(lk, "key") || contains(lk, "token") || contains(lk, "auth")) && !isHTTP(val) {
			apiKey = val
		}
	})
	return &ParsedSettings{BaseURL: baseURL, APIKey: apiKey}
}

func walk(v any, fn func(key, val string)) {
	switch x := v.(type) {
	case map[string]any:
		for k, val := range x {
			if s, ok := val.(string); ok {
				fn(k, s)
			} else {
				walk(val, fn)
			}
		}
	case []any:
		for _, item := range x {
			walk(item, fn)
		}
	}
}

func str(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func lower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			b[i] = c + 32
		} else {
			b[i] = c
		}
	}
	return string(b)
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchLower(s, sub)
}

func searchLower(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		match := true
		for j := 0; j < len(sub); j++ {
		cs, csub := s[i+j], sub[j]
		if cs >= 'A' && cs <= 'Z' {
			cs += 32
		}
		if cs != csub {
			match = false
			break
		}
	}
	if match {
		return true
	}
}
	return false
}

func isHTTP(s string) bool {
	return len(s) > 7 && (s[:7] == "http://" || s[:8] == "https://")
}
