package db

import "database/sql"

// modelPricing holds the seed data for the model_pricing table.
// Prices are in USD per million tokens.
type modelPricing struct {
	ModelID           string
	DisplayName       string
	InputPerM         float64
	OutputPerM        float64
	CacheReadPerM     float64
	CacheCreationPerM float64
}

// seedModelPricing inserts default pricing rows. Existing rows are skipped
// (ON CONFLICT DO NOTHING) so user customizations survive upgrades.
func seedModelPricing(tx *sql.Tx) error {
	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO model_pricing
		(model_id, display_name, input_cost_per_million, output_cost_per_million,
		 cache_read_cost_per_million, cache_creation_cost_per_million, is_custom)
		VALUES (?, ?, ?, ?, ?, ?, 0)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range defaultPricing() {
		if _, err := stmt.Exec(p.ModelID, p.DisplayName,
			p.InputPerM, p.OutputPerM, p.CacheReadPerM, p.CacheCreationPerM); err != nil {
			return err
		}
	}
	return nil
}

// defaultPricing returns the built-in model pricing catalog.
// Sources: official provider pricing pages as of 2026-06.
func defaultPricing() []modelPricing {
	return []modelPricing{
		// ── Anthropic Claude ──────────────────────────────────────
		{"claude-opus-4-8", "Claude Opus 4.8", 15.0, 75.0, 1.5, 18.75},
		{"claude-sonnet-4-6", "Claude Sonnet 4.6", 3.0, 15.0, 0.3, 3.75},
		{"claude-haiku-4-5", "Claude Haiku 4.5", 0.8, 4.0, 0.08, 1.0},
		{"claude-fable-5", "Claude Fable 5", 15.0, 75.0, 1.5, 18.75},
		{"claude-3-5-sonnet-20241022", "Claude 3.5 Sonnet", 3.0, 15.0, 0.3, 3.75},
		{"claude-3-5-haiku-20241022", "Claude 3.5 Haiku", 0.8, 4.0, 0.08, 1.0},
		{"claude-3-opus-20240229", "Claude 3 Opus", 15.0, 75.0, 1.5, 18.75},

		// ── OpenAI GPT ───────────────────────────────────────────
		{"gpt-4.1", "GPT-4.1", 2.0, 8.0, 0.5, 2.0},
		{"gpt-4.1-mini", "GPT-4.1 Mini", 0.4, 1.6, 0.1, 0.4},
		{"gpt-4.1-nano", "GPT-4.1 Nano", 0.1, 0.4, 0.025, 0.1},
		{"gpt-4o", "GPT-4o", 2.5, 10.0, 1.25, 2.5},
		{"gpt-4o-mini", "GPT-4o Mini", 0.15, 0.6, 0.075, 0.15},
		{"o3", "o3", 2.0, 8.0, 0.5, 2.0},
		{"o3-mini", "o3-mini", 1.1, 4.4, 0.275, 1.1},
		{"o4-mini", "o4-mini", 1.1, 4.4, 0.275, 1.1},
		{"gpt-5.5", "GPT-5.5", 2.0, 8.0, 0.5, 2.0},

		// ── DeepSeek ─────────────────────────────────────────────
		{"deepseek-chat", "DeepSeek V3", 0.27, 1.1, 0.07, 0.27},
		{"deepseek-reasoner", "DeepSeek R1", 0.55, 2.19, 0.14, 0.55},
		{"deepseek-v4-flash", "DeepSeek V4 Flash", 0.1, 0.4, 0.025, 0.1},
		{"deepseek-v4-pro", "DeepSeek V4 Pro", 0.55, 2.19, 0.14, 0.55},

		// ── Google Gemini ────────────────────────────────────────
		{"gemini-2.5-pro", "Gemini 2.5 Pro", 1.25, 10.0, 0.315, 1.25},
		{"gemini-2.5-flash", "Gemini 2.5 Flash", 0.15, 0.6, 0.0375, 0.15},
		{"gemini-2.0-flash", "Gemini 2.0 Flash", 0.1, 0.4, 0.025, 0.1},

		// ── Mimo (Xiaomi) ───────────────────────────────────────
		{"mimo-v2.5-pro", "Mimo V2.5 Pro", 0.5, 2.0, 0.125, 0.5},
		{"mimo-v2.5", "Mimo V2.5", 0.3, 1.2, 0.075, 0.3},
		{"mimo-v2-flash", "Mimo V2 Flash", 0.1, 0.4, 0.025, 0.1},

		// ── Qwen (Alibaba) ──────────────────────────────────────
		{"qwen3.7-max", "Qwen 3.7 Max", 0.5, 2.0, 0.125, 0.5},
		{"qwen3.6-plus", "Qwen 3.6 Plus", 0.3, 1.2, 0.075, 0.3},
		{"qwen-max", "Qwen Max", 1.6, 6.4, 0.4, 1.6},
		{"qwen-plus", "Qwen Plus", 0.4, 1.2, 0.1, 0.4},
		{"qwen-turbo", "Qwen Turbo", 0.05, 0.2, 0.0125, 0.05},

		// ── GLM (Zhipu) ─────────────────────────────────────────
		{"glm-5.1", "GLM-5.1", 0.5, 2.0, 0.125, 0.5},
		{"glm-4-plus", "GLM-4 Plus", 0.5, 2.0, 0.125, 0.5},

		// ── OpenRouter popular ───────────────────────────────────
		{"meta-llama/llama-4-maverick", "Llama 4 Maverick", 0.2, 0.6, 0.05, 0.2},
		{"meta-llama/llama-4-scout", "Llama 4 Scout", 0.15, 0.42, 0.0375, 0.15},
		{"mistral-large-latest", "Mistral Large", 2.0, 6.0, 0.5, 2.0},

		// ── MiniMax ──────────────────────────────────────────────
		{"MiniMax-M3", "MiniMax M3", 0.3, 1.2, 0.075, 0.3},
	}
}
