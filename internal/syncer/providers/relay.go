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

// RelaySyncer implements syncer for one-api / new-api relay platforms.
// These platforms expose OpenAI-compatible endpoints with custom billing paths.
type RelaySyncer struct {
	name string // "one-api" or "new-api"
}

// NewRelaySyncer creates a new relay syncer.
func NewRelaySyncer(name string) *RelaySyncer {
	return &RelaySyncer{name: name}
}

// Name returns the syncer name.
func (s *RelaySyncer) Name() string {
	return s.name
}

// SupportsUsage returns true.
func (s *RelaySyncer) SupportsUsage() bool {
	return false // Most relays don't have a standard usage API
}

// SupportsBalance returns true.
func (s *RelaySyncer) SupportsBalance() bool {
	return true
}

// ValidateKey validates the API key by calling a lightweight endpoint.
func (s *RelaySyncer) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	if baseURL == "" {
		return fmt.Errorf("base_url required for relay syncer")
	}
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/v1/models", nil)
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

// FetchBalance fetches the current balance from the relay platform.
func (s *RelaySyncer) FetchBalance(ctx context.Context, apiKey string, baseURL string) (*syncer.BalanceInfo, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("base_url required for relay syncer")
	}

	// Try the common relay balance endpoint
	url := baseURL + "/api/user/balance"
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
			Balance float64 `json:"balance"`
			Used    float64 `json:"used"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &syncer.BalanceInfo{
		Available: result.Data.Balance - result.Data.Used,
		Total:     result.Data.Balance,
		Currency:  "USD",
		UpdatedAt: time.Now(),
	}, nil
}

// FetchUsage fetches usage records. Most relays don't support this via API.
func (s *RelaySyncer) FetchUsage(ctx context.Context, apiKey string, baseURL string, from, to time.Time) ([]syncer.Record, error) {
	return []syncer.Record{}, nil
}
