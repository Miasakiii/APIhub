package alert

import (
	"apihub/internal/model"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
)

// Engine checks alert rules and triggers notifications.
type Engine struct {
	db       *sql.DB
	notifier *Notifier
}

// NewEngine creates a new alert engine.
func NewEngine(db *sql.DB) *Engine {
	return &Engine{db: db}
}

// SetNotifier sets the notifier for sending alert notifications.
func (e *Engine) SetNotifier(n *Notifier) {
	e.notifier = n
}

// RunOnce runs all enabled alert rules once.
func (e *Engine) RunOnce() error {
	rules, err := e.listEnabledRules()
	if err != nil {
		return fmt.Errorf("list rules: %w", err)
	}

	for _, rule := range rules {
		if err := e.checkRule(rule); err != nil {
			log.Printf("check rule %s: %v", rule.ID, err)
		}
	}
	return nil
}

// listEnabledRules fetches all enabled alert rules.
func (e *Engine) listEnabledRules() ([]model.Alert, error) {
	rows, err := e.db.Query(`
		SELECT id, name, type, provider_id, api_key_id, threshold, unit, enabled
		FROM alerts WHERE enabled = 1
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []model.Alert
	for rows.Next() {
		var a model.Alert
		var enabled int
		if err := rows.Scan(&a.ID, &a.Name, &a.Type, &a.ProviderID, &a.APIKeyID, &a.Threshold, &a.Unit, &enabled); err != nil {
			continue
		}
		a.Enabled = enabled == 1
		rules = append(rules, a)
	}
	return rules, nil
}

// checkRule evaluates a single alert rule.
func (e *Engine) checkRule(rule model.Alert) error {
	switch rule.Type {
	case "balance_low":
		return e.checkBalanceLow(rule)
	case "key_expired":
		return e.checkKeyExpired(rule)
	case "abnormal_frequency":
		return e.checkAbnormalFrequency(rule)
	case "subscription_expiring":
		return e.checkSubscriptionExpiring(rule)
	default:
		return fmt.Errorf("unknown alert type: %s", rule.Type)
	}
}

// checkBalanceLow checks if any key's balance is below threshold.
func (e *Engine) checkBalanceLow(rule model.Alert) error {
	var query string
	var args []any
	if rule.APIKeyID != "" {
		query = "SELECT id, balance_usd FROM api_keys WHERE id = ? AND balance_usd < ?"
		args = []any{rule.APIKeyID, rule.Threshold}
	} else if rule.ProviderID != "" {
		query = "SELECT id, balance_usd FROM api_keys WHERE provider_id = ? AND balance_usd < ?"
		args = []any{rule.ProviderID, rule.Threshold}
	} else {
		query = "SELECT id, balance_usd FROM api_keys WHERE balance_usd < ?"
		args = []any{rule.Threshold}
	}

	rows, err := e.db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var keyID string
		var balance float64
		if err := rows.Scan(&keyID, &balance); err != nil {
			continue
		}
		msg := fmt.Sprintf("Key %s balance $%.2f below threshold $%.2f", keyID, balance, rule.Threshold)
		_ = e.createHistory(rule.ID, msg, "warning")
		e.notify("warning", "余额不足告警", msg)
	}
	return nil
}

// checkKeyExpired checks if any key is marked as expired or invalid.
func (e *Engine) checkKeyExpired(rule model.Alert) error {
	var query string
	var args []any
	if rule.APIKeyID != "" {
		query = "SELECT id, status FROM api_keys WHERE id = ? AND status IN ('expired', 'invalid', 'revoked')"
		args = []any{rule.APIKeyID}
	} else if rule.ProviderID != "" {
		query = "SELECT id, status FROM api_keys WHERE provider_id = ? AND status IN ('expired', 'invalid', 'revoked')"
		args = []any{rule.ProviderID}
	} else {
		query = "SELECT id, status FROM api_keys WHERE status IN ('expired', 'invalid', 'revoked')"
		args = []any{}
	}

	rows, err := e.db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var keyID, status string
		if err := rows.Scan(&keyID, &status); err != nil {
			continue
		}
		msg := fmt.Sprintf("Key %s status is '%s'", keyID, status)
		_ = e.createHistory(rule.ID, msg, "critical")
		e.notify("critical", "Key 状态告警", msg)
	}
	return nil
}

// checkAbnormalFrequency checks if request count exceeds threshold in last hour.
func (e *Engine) checkAbnormalFrequency(rule model.Alert) error {
	threshold := int64(rule.Threshold)
	if threshold == 0 {
		threshold = 100 // default threshold
	}

	var query string
	var args []any
	if rule.APIKeyID != "" {
		query = "SELECT api_key_id, COUNT(*) FROM usage_records WHERE api_key_id = ? AND timestamp > datetime('now', '-1 hour') GROUP BY api_key_id HAVING COUNT(*) > ?"
		args = []any{rule.APIKeyID, threshold}
	} else if rule.ProviderID != "" {
		query = "SELECT provider_id, COUNT(*) FROM usage_records WHERE provider_id = ? AND timestamp > datetime('now', '-1 hour') GROUP BY provider_id HAVING COUNT(*) > ?"
		args = []any{rule.ProviderID, threshold}
	} else {
		query = "SELECT provider_id, COUNT(*) FROM usage_records WHERE timestamp > datetime('now', '-1 hour') GROUP BY provider_id HAVING COUNT(*) > ?"
		args = []any{threshold}
	}

	rows, err := e.db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var pid string
		var count int64
		if err := rows.Scan(&pid, &count); err != nil {
			continue
		}
		msg := fmt.Sprintf("Provider %s had %d requests in last hour (threshold: %d)", pid, count, threshold)
		_ = e.createHistory(rule.ID, msg, "warning")
		e.notify("warning", "异常频率告警", msg)
	}
	return nil
}

// checkSubscriptionExpiring checks if any subscription is about to expire.
func (e *Engine) checkSubscriptionExpiring(rule model.Alert) error {
	// subscription table not yet created, skip for now
	return nil
}

func (e *Engine) createHistory(alertID, message, level string) error {
	id := generateID()
	_, err := e.db.Exec(`
		INSERT INTO alert_history (id, alert_id, message, level, created_at)
		VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, id, alertID, message, level)
	if err != nil {
		return err
	}

	// Update last_triggered_at
	_, _ = e.db.Exec("UPDATE alerts SET last_triggered_at = CURRENT_TIMESTAMP WHERE id = ?", alertID)
	return nil
}

func (e *Engine) notify(level, title, message string) {
	if e.notifier != nil {
		go e.notifier.Send(level, title, message)
	}
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
