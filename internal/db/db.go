package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	_ "modernc.org/sqlite"
)

// validIdentifier is a whitelist regex for SQL identifiers (tables, columns).
// Only allows alphanumeric characters and underscores, 1-128 chars.
var validIdentifier = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]{0,127}$`)

const (
	// currentVersion is the target schema version. Bump this when adding new migrations.
	currentVersion = 6
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

	uri := fmt.Sprintf("file:%s?_busy_timeout=5000", filepath.ToSlash(dbPath))
	d, err := sql.Open("sqlite", uri)
	if err != nil {
		return nil, fmt.Errorf("open apihub db: %w", err)
	}

	if _, err := d.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("enable WAL: %w", err)
	}

	if _, err := d.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	return &DB{DB: d}, nil
}

// columnExists checks if a column exists in a table.
// Uses whitelist validation on table and column names to prevent SQL injection.
func (d *DB) columnExists(table, column string) bool {
	if !validIdentifier.MatchString(table) || !validIdentifier.MatchString(column) {
		log.Printf("[db] columnExists: invalid identifier (table=%q, column=%q)", table, column)
		return false
	}
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name='%s'", table, column)
	d.QueryRow(query).Scan(&count)
	return count > 0
}

// getVersion reads the current schema version from PRAGMA user_version.
func (d *DB) getVersion() (int, error) {
	var v int
	err := d.QueryRow("PRAGMA user_version").Scan(&v)
	return v, err
}

// setVersion writes the schema version via PRAGMA user_version.
func (d *DB) setVersion(v int) error {
	// v is always a trusted integer from internal migration logic, but
	// we still enforce a reasonable range as defense-in-depth.
	if v < 0 || v > 9999 {
		return fmt.Errorf("setVersion: invalid version %d", v)
	}
	_, err := d.Exec(fmt.Sprintf("PRAGMA user_version=%d", v))
	return err
}

// Migrate runs incremental schema migrations based on PRAGMA user_version.
// Each migration step is versioned and idempotent. Upgrading from any prior
// version to currentVersion is supported.
func (d *DB) Migrate() error {
	version, err := d.getVersion()
	if err != nil {
		return fmt.Errorf("migrate: read version: %w", err)
	}

	if version >= currentVersion {
		log.Printf("[db] schema version %d (up to date)", version)
		return nil
	}

	log.Printf("[db] schema version %d → %d, running migrations...", version, currentVersion)

	// Define migration steps. Each step migrates from version N to N+1.
	// The index is the target version (step[1] migrates 0→1, step[2] migrates 1→2, etc.)
	migrations := map[int]func(tx *sql.Tx) error{
		1: migrateV1,
		2: migrateV2,
		3: migrateV3,
		4: migrateV4,
		5: migrateV5,
		6: migrateV6,
	}

	for v := version + 1; v <= currentVersion; v++ {
		fn, ok := migrations[v]
		if !ok {
			return fmt.Errorf("migrate: no migration defined for version %d", v)
		}

		tx, err := d.Begin()
		if err != nil {
			return fmt.Errorf("migrate: begin tx for v%d: %w", v, err)
		}

		if err := fn(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migrate: v%d failed: %w", v, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("migrate: commit v%d: %w", v, err)
		}

		// Update version after successful commit
		if err := d.setVersion(v); err != nil {
			return fmt.Errorf("migrate: set version to %d: %w", v, err)
		}

		log.Printf("[db] ✓ migrated to v%d", v)
	}

	return nil
}

// migrateV1 creates the base schema (equivalent to the original flat migrations).
func migrateV1(tx *sql.Tx) error {
	stmts := []string{
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
		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_usage_provider ON usage_records(provider_id)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_model ON usage_records(model)`,
		`CREATE INDEX IF NOT EXISTS idx_usage_timestamp ON usage_records(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_date ON daily_stats(date)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_provider ON daily_stats(provider_id)`,
	}

	for _, ddl := range stmts {
		if _, err := tx.Exec(ddl); err != nil {
			return fmt.Errorf("%s: %w", truncate(ddl, 60), err)
		}
	}

	// Idempotent ALTER TABLE: add syncer column to providers if missing
	if _, err := tx.Exec("ALTER TABLE providers ADD COLUMN syncer TEXT DEFAULT ''"); err != nil {
		// Column already exists is fine (SQLite error code 1: duplicate column)
		if !strings.Contains(err.Error(), "duplicate column") {
			return fmt.Errorf("add syncer column: %w", err)
		}
	}

	return nil
}

// migrateV2 adds the model_pricing table with seed data.
func migrateV2(tx *sql.Tx) error {
	// Create model_pricing table
	if _, err := tx.Exec(`CREATE TABLE IF NOT EXISTS model_pricing (
		model_id TEXT PRIMARY KEY,
		display_name TEXT NOT NULL,
		input_cost_per_million REAL NOT NULL DEFAULT 0,
		output_cost_per_million REAL NOT NULL DEFAULT 0,
		cache_read_cost_per_million REAL NOT NULL DEFAULT 0,
		cache_creation_cost_per_million REAL NOT NULL DEFAULT 0,
		is_custom INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`); err != nil {
		return fmt.Errorf("create model_pricing: %w", err)
	}

	// Seed with default pricing data
	if err := seedModelPricing(tx); err != nil {
		return fmt.Errorf("seed model_pricing: %w", err)
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

// migrateV3 adds usage_sessions and usage_activity_buckets tables for three-layer aggregation.
func migrateV3(tx *sql.Tx) error {
	stmts := []string{
		// Sessions: consecutive API calls grouped by (provider, model, source) within a 30-min window.
		`CREATE TABLE IF NOT EXISTS usage_sessions (
			id TEXT PRIMARY KEY,
			provider_id TEXT NOT NULL,
			model TEXT NOT NULL,
			source TEXT NOT NULL,
			started_at DATETIME NOT NULL,
			ended_at DATETIME NOT NULL,
			duration_ms INTEGER NOT NULL DEFAULT 0,
			request_count INTEGER NOT NULL DEFAULT 0,
			input_tokens INTEGER NOT NULL DEFAULT 0,
			output_tokens INTEGER NOT NULL DEFAULT 0,
			cache_read INTEGER NOT NULL DEFAULT 0,
			cache_create INTEGER NOT NULL DEFAULT 0,
			cost_usd REAL NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_started ON usage_sessions(started_at)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_model ON usage_sessions(model)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_provider ON usage_sessions(provider_id)`,

		// Activity buckets: hourly aggregation per provider+model.
		`CREATE TABLE IF NOT EXISTS usage_activity_buckets (
			id TEXT PRIMARY KEY,
			bucket_start DATETIME NOT NULL,
			provider_id TEXT NOT NULL,
			model TEXT NOT NULL,
			request_count INTEGER NOT NULL DEFAULT 0,
			input_tokens INTEGER NOT NULL DEFAULT 0,
			output_tokens INTEGER NOT NULL DEFAULT 0,
			cache_read INTEGER NOT NULL DEFAULT 0,
			cache_create INTEGER NOT NULL DEFAULT 0,
			cost_usd REAL NOT NULL DEFAULT 0,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(bucket_start, provider_id, model)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_buckets_start ON usage_activity_buckets(bucket_start)`,
		`CREATE INDEX IF NOT EXISTS idx_buckets_provider ON usage_activity_buckets(provider_id)`,
	}

	for _, ddl := range stmts {
		if _, err := tx.Exec(ddl); err != nil {
			return fmt.Errorf("%s: %w", truncate(ddl, 60), err)
		}
	}
	return nil
}

func migrateV4(tx *sql.Tx) error {
	// Add source column to subscriptions for auto-detection tracking.
	_, err := tx.Exec(`ALTER TABLE subscriptions ADD COLUMN source TEXT NOT NULL DEFAULT 'manual'`)
	if err != nil {
		// Column may already exist from a previous partial migration
		if !strings.Contains(err.Error(), "duplicate column") {
			return err
		}
	}
	return nil
}

func migrateV5(tx *sql.Tx) error {
	stmts := []string{
		// Agents table
		`CREATE TABLE IF NOT EXISTS agents (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL DEFAULT 'cli',
			icon TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,

		// Add agent_id column to usage tables
		`ALTER TABLE usage_records ADD COLUMN agent_id TEXT DEFAULT ''`,
		`ALTER TABLE daily_stats ADD COLUMN agent_id TEXT DEFAULT ''`,
		`ALTER TABLE usage_sessions ADD COLUMN agent_id TEXT DEFAULT ''`,
		`ALTER TABLE usage_activity_buckets ADD COLUMN agent_id TEXT DEFAULT ''`,

		// Index for agent_id filtering
		`CREATE INDEX IF NOT EXISTS idx_usage_records_agent ON usage_records(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_daily_stats_agent ON daily_stats(agent_id)`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_agent ON usage_sessions(agent_id)`,
	}

	for _, ddl := range stmts {
		if _, err := tx.Exec(ddl); err != nil {
			if strings.Contains(err.Error(), "duplicate column") {
				continue
			}
			return fmt.Errorf("%s: %w", truncate(ddl, 60), err)
		}
	}
	return nil
}

func migrateV6(tx *sql.Tx) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS key_audit_log (
			id TEXT PRIMARY KEY,
			key_id TEXT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
			action TEXT NOT NULL,
			detail TEXT DEFAULT '',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_key_id ON key_audit_log(key_id)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_created ON key_audit_log(created_at)`,
	}

	for _, ddl := range stmts {
		if _, err := tx.Exec(ddl); err != nil {
			return fmt.Errorf("%s: %w", truncate(ddl, 60), err)
		}
	}
	return nil
}
