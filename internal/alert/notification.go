package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// Notifier sends alert notifications through various channels.
type Notifier struct {
	webhooks []WebhookConfig
}

// WebhookConfig holds a webhook endpoint configuration.
type WebhookConfig struct {
	URL     string
	Headers map[string]string
}

// NewNotifier creates a new notifier with the given webhooks.
func NewNotifier(webhooks []WebhookConfig) *Notifier {
	return &Notifier{webhooks: webhooks}
}

// Send sends a notification to all configured channels.
func (n *Notifier) Send(level, title, message string) error {
	if len(n.webhooks) == 0 {
		return nil
	}

	payload := map[string]interface{}{
		"level":     level,
		"title":     title,
		"message":   message,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	for _, webhook := range n.webhooks {
		go n.sendWebhook(webhook, payloadBytes)
	}

	return nil
}

func (n *Notifier) sendWebhook(config WebhookConfig, payload []byte) {
	req, err := http.NewRequest("POST", config.URL, bytes.NewReader(payload))
	if err != nil {
		log.Printf("webhook request error: %v", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("webhook send error: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("webhook returned status %d", resp.StatusCode)
	}
}
