package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// DB wraps the APIHub SQLite connection.
type DB struct {
	*sql.DB
}

// Open opens or creates the APIHub database.
func Open(dbPath string) (*DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, err
	}

	// Use a different driver name suffix by adding _busy_timeout
	uri := fmt.Sprintf("file:%s?_busy_timeout=5000", filepath.ToSlash(dbPath))
	d, err := sql.Open("sqlite", uri)
	if err != nil {
		return nil, fmt.Errorf("open apihub db: %w", err)
	}

	// Enable WAL mode
	if _, err := d.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	if _, err := d.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	return &DB{DB: d}, nil
}

// columnExists checks if a column exists in a table.
func (d *DB) columnExists(table, column string) bool {
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name='%s'", table, column)
	d.QueryRow(query).Scan(&count)
	return count > 0
}

// Migrate runs all schema migrations.
func (d *DB) Migrate() error {
	migrations := []string{
		// Providers
		`CREATE TABLE IF NOT EXISTS providers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			base_url TEXT,
			console_url TEXT,
			topup_url TEXT,
			docs_url TEXT,
			api_key TEXT,
			enabled INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// API Keys
		`CREATE TABLE IF NOT EXISTS api_keys (
			id TEXT PRIMARY KEY,
			provider_id TEXT NOT NULL REFERENCES providers(id),
			key_hash TEXT NOT NULL,
			key_encrypted BLOB,
			name TEXT,
			source TEXT NOT NULL DEFAULT 'manual',
			status TEXT NOT NULL DEFAULT 'active',
			balance_usd REAL DEFAULT 0,
			last_checked DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(key_hash)
		)`,
		// Usage Records
		`CREATE TABLE IF NOT EXISTS usage_records (
			id TEXT PRIMARY KEY,
			api_key_id TEXT,
			provider_id TEXT NOT NULL,
			model TEXT NOT NULL,
			input_tokens INTEGER DEFAULT 0,
			output_tokens INTEGER DEFAULT 0,
			cache_read INTEGER DEFAULT 0,
			cache_create INTEGER DEFAULT 0,
			cost_usd REAL DEFAULT 0,
			source TEXT NOT NULL DEFAULT 'syncer',
			timestamp DATETIME NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Daily Stats
		`CREATE TABLE IF NOT EXISTS daily_stats (
			id TEXT PRIMARY KEY,
			provider_id TEXT NOT NULL,
			model TEXT NOT NULL,
			source TEXT NOT NULL,
			date TEXT NOT NULL,
			request_count INTEGER DEFAULT 0,
			input_tokens INTEGER DEFAULT 0,
			output_tokens INTEGER DEFAULT 0,
			cache_read INTEGER DEFAULT 0,
			cache_create INTEGER DEFAULT 0,
			cost_usd REAL DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(provider_id, model, source, date)
		)`,
		// Sync State
		`CREATE TABLE IF NOT EXISTS sync_state (
			id TEXT PRIMARY KEY,
			source TEXT NOT NULL UNIQUE,
			last_sync DATETIME,
			offset_val INTEGER DEFAULT 0,
			status TEXT NOT NULL DEFAULT 'pending',
			error TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Sync Log
		`CREATE TABLE IF NOT EXISTS sync_log (
			id TEXT PRIMARY KEY,
			source TEXT NOT NULL,
			fetched INTEGER DEFAULT 0,
			inserted INTEGER DEFAULT 0,
			error TEXT,
			duration_ms INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Indexes
		// Alerts
		`CREATE TABLE IF NOT EXISTS alerts (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			provider_id TEXT REFERENCES providers(id),
			api_key_id TEXT REFERENCES api_keys(id),
			threshold REAL,
			unit TEXT,
			enabled INTEGER NOT NULL DEFAULT 1,
			last_triggered_at DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Subscriptions
		`CREATE TABLE IF NOT EXISTS subscriptions (
			id TEXT PRIMARY KEY,
			provider_id TEXT NOT NULL REFERENCES providers(id),
			plan_name TEXT NOT NULL,
			price REAL,
			currency TEXT DEFAULT 'USD',
			billing_cycle TEXT DEFAULT 'monthly',
			quota_type TEXT DEFAULT 'tokens',
			quota_total REAL,
			quota_used REAL DEFAULT 0,
			start_date DATE,
			renew_date DATE,
			auto_renew INTEGER DEFAULT 1,
			status TEXT DEFAULT 'active',
			notes TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Users
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Webhook Settings
		`CREATE TABLE IF NOT EXISTS webhook_settings (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			url TEXT NOT NULL,
			headers TEXT DEFAULT '{}',
			enabled INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		// Alert History
		`CREATE TABLE IF NOT EXISTS alert_history (
			id TEXT PRIMARY KEY,
			alert_id TEXT NOT NULL REFERENCES alerts(id),
			message TEXT NOT NULL,
			level TEXT NOT NULL DEFAULT 'warning',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_provider ON usage_records(provider_id)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_model ON usage_records(model)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_timestamp ON usage_records(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_date ON daily_stats(date)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_provider ON daily_stats(provider_id)`,
	}

	for _, ddl := range migrations {
		if _, err := d.Exec(ddl); err != nil {
			return fmt.Errorf("migrate: %s: %w", truncate(ddl, 60), err)
		}
	}

	// Idempotent migrations (ALTER TABLE)
	if !d.columnExists("providers", "syncer") {
		if _, err := d.Exec("ALTER TABLE providers ADD COLUMN syncer TEXT DEFAULT ''"); err != nil {
			return fmt.Errorf("migrate: add syncer column: %w", err)
		}
	}

	return nil
}

func truncate(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) > n {
		return s[:n] + "..."
	}
	return s
}
