package providers

import (
	"apihub/internal/syncer"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

const anthropicBaseURL = "https://api.anthropic.com/v1"

// AnthropicSyncer implements syncer for Anthropic.
// Anthropic does not currently expose a public usage/billing API,
// so this syncer only supports key validation.
type AnthropicSyncer struct{}

func (s *AnthropicSyncer) Name() string {
	return "anthropic"
}

func (s *AnthropicSyncer) SupportsUsage() bool {
	return false // No public usage API yet
}

func (s *AnthropicSyncer) SupportsBalance() bool {
	return false // No public billing API yet
}

// ValidateKey validates the API key by sending a minimal messages request.
func (s *AnthropicSyncer) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	url := resolveURL(baseURL, anthropicBaseURL) + "/messages"
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 401 = invalid key, 400 = key valid but bad request (expected with nil body)
	if resp.StatusCode == 401 || resp.StatusCode == 403 {
		return fmt.Errorf("invalid api key: %d", resp.StatusCode)
	}
	if resp.StatusCode >= 500 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("validate key failed: %d %s", resp.StatusCode, string(body))
	}
	return nil
}

// FetchBalance is not supported for Anthropic.
func (s *AnthropicSyncer) FetchBalance(ctx context.Context, apiKey string, baseURL string) (*syncer.BalanceInfo, error) {
	return nil, fmt.Errorf("anthropic does not expose a billing API")
}

// FetchUsage is not supported for Anthropic.
func (s *AnthropicSyncer) FetchUsage(ctx context.Context, apiKey string, baseURL string, from, to time.Time) ([]syncer.Record, error) {
	return nil, fmt.Errorf("anthropic does not expose a usage API")
}
