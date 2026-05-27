package providers

import (
	"apihub/internal/syncer"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const openrouterBaseURL = "https://openrouter.ai/api/v1"

// OpenRouterSyncer implements syncer for OpenRouter.
type OpenRouterSyncer struct{}

// Name returns the syncer name.
func (s *OpenRouterSyncer) Name() string {
	return "openrouter"
}

// SupportsUsage returns true.
func (s *OpenRouterSyncer) SupportsUsage() bool {
	return true
}

// SupportsBalance returns true.
func (s *OpenRouterSyncer) SupportsBalance() bool {
	return true
}

// ValidateKey validates the API key by calling /auth/key.
func (s *OpenRouterSyncer) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	url := resolveURL(baseURL, openrouterBaseURL) + "/auth/key"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return fmt.Errorf("invalid api key: %d", resp.StatusCode)
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("validate key failed: %d %s", resp.StatusCode, string(body))
	}
	return nil
}

// FetchBalance fetches the current balance.
func (s *OpenRouterSyncer) FetchBalance(ctx context.Context, apiKey string, baseURL string) (*syncer.BalanceInfo, error) {
	url := resolveURL(baseURL, openrouterBaseURL) + "/credits"
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("fetch balance failed: %d %s", resp.StatusCode, string(body))
	}

	var result struct {
		Data struct {
			TotalCredits float64 `json:"total_credits"`
			TotalUsage   float64 `json:"total_usage"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &syncer.BalanceInfo{
		Available: result.Data.TotalCredits - result.Data.TotalUsage,
		Total:     result.Data.TotalCredits,
		Currency:  "USD",
		UpdatedAt: time.Now(),
	}, nil
}

// FetchUsage fetches usage records. OpenRouter does not expose per-request usage
// via a public API, so this returns an empty slice for now.
// In production, users typically ingest usage via cc-switch logs instead.
func (s *OpenRouterSyncer) FetchUsage(ctx context.Context, apiKey string, baseURL string, from, to time.Time) ([]syncer.Record, error) {
	// OpenRouter does not have a public per-request usage endpoint.
	// Return empty - usage is expected to come from cc-switch proxy logs.
	return []syncer.Record{}, nil
}

func resolveURL(base, fallback string) string {
	if base != "" {
		return base
	}
	return fallback
}
