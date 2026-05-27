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

const openaiBaseURL = "https://api.openai.com/v1"

// OpenAISyncer implements syncer for OpenAI.
// Requires an admin API key to access usage/billing endpoints.
type OpenAISyncer struct{}

func (s *OpenAISyncer) Name() string {
	return "openai"
}

func (s *OpenAISyncer) SupportsUsage() bool {
	return true
}

func (s *OpenAISyncer) SupportsBalance() bool {
	return true
}

// ValidateKey validates the API key by listing models.
func (s *OpenAISyncer) ValidateKey(ctx context.Context, apiKey string, baseURL string) error {
	url := resolveURL(baseURL, openaiBaseURL) + "/models"
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

// FetchBalance fetches the current billing balance.
// Uses the dashboard billing API which requires an admin key.
func (s *OpenAISyncer) FetchBalance(ctx context.Context, apiKey string, baseURL string) (*syncer.BalanceInfo, error) {
	url := resolveURL(baseURL, openaiBaseURL) + "/dashboard/billing/subscription"
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
		HardLimitUSD float64 `json:"hard_limit_usd"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	// Also fetch current usage to calculate available balance
	usageURL := resolveURL(baseURL, openaiBaseURL) + "/dashboard/billing/usage"
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	params := fmt.Sprintf("?start_date=%s&end_date=%s", start.Format("2006-01-02"), now.Format("2006-01-02"))

	req2, err := http.NewRequestWithContext(ctx, "GET", usageURL+params, nil)
	if err != nil {
		return nil, err
	}
	req2.Header.Set("Authorization", "Bearer "+apiKey)

	resp2, err := http.DefaultClient.Do(req2)
	if err != nil {
		return nil, err
	}
	defer resp2.Body.Close()

	var totalUsage float64
	if resp2.StatusCode == 200 {
		var usageResult struct {
			TotalUsage float64 `json:"total_usage"` // in cents
		}
		if err := json.NewDecoder(resp2.Body).Decode(&usageResult); err == nil {
			totalUsage = usageResult.TotalUsage / 100.0 // convert to USD
		}
	}

	return &syncer.BalanceInfo{
		Available: result.HardLimitUSD - totalUsage,
		Total:     result.HardLimitUSD,
		Currency:  "USD",
		UpdatedAt: time.Now(),
	}, nil
}

// FetchUsage fetches daily usage for the given time range.
func (s *OpenAISyncer) FetchUsage(ctx context.Context, apiKey string, baseURL string, from, to time.Time) ([]syncer.Record, error) {
	url := resolveURL(baseURL, openaiBaseURL) + "/dashboard/billing/usage"
	params := fmt.Sprintf("?start_date=%s&end_date=%s", from.Format("2006-01-02"), to.Format("2006-01-02"))

	req, err := http.NewRequestWithContext(ctx, "GET", url+params, nil)
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
		return nil, fmt.Errorf("fetch usage failed: %d %s", resp.StatusCode, string(body))
	}

	var result struct {
		DailyCosts []struct {
			Timestamp float64 `json:"timestamp"`
			LineItems []struct {
				Name      string  `json:"name"`
				CostUSD   float64 `json:"cost"`
			} `json:"line_items"`
		} `json:"daily_costs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	var records []syncer.Record
	for _, day := range result.DailyCosts {
		date := time.Unix(int64(day.Timestamp), 0)
		for _, item := range day.LineItems {
			if item.CostUSD <= 0 {
				continue
			}
			records = append(records, syncer.Record{
				Model:        item.Name,
				Date:         date,
				RequestCount: 1,
				CostUSD:      item.CostUSD,
			})
		}
	}
	return records, nil
}
