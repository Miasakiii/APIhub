package ccswitch

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const dbPath = ".cc-switch/cc-switch.db"

// Reader is a read-only connection to cc-switch's SQLite database.
type Reader struct {
	db *sql.DB
}

// Open connects to cc-switch.db in read-only mode with busy_timeout for safe
// concurrent access (cc-switch uses journal_mode=delete, not WAL).
func Open(dbFile string) (*Reader, error) {
	abs, err := filepath.Abs(dbFile)
	if err != nil {
		return nil, fmt.Errorf("ccswitch: resolve db path %q: %w", dbFile, err)
	}
	if _, err := os.Stat(abs); err != nil {
		return nil, fmt.Errorf("ccswitch: db not found at %q", abs)
	}

	// ?mode=ro prevents accidental writes, _busy_timeout=5000 gives retry
	// on "database is locked" when cc-switch is actively writing.
	uri := fmt.Sprintf("file:%s?mode=ro&_busy_timeout=5000", filepath.ToSlash(abs))
	db, err := sql.Open("sqlite", uri)
	if err != nil {
		return nil, fmt.Errorf("ccswitch: open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite read-only, serialise to avoid lock contention
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("ccswitch: ping failed: %w", err)
	}

	return &Reader{db: db}, nil
}

// DefaultPath returns the default cc-switch database path for the current user.
func DefaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, dbPath)
}

// Close releases the connection.
func (r *Reader) Close() error {
	if r.db != nil {
		return r.db.Close()
	}
	return nil
}

// DB returns the underlying *sql.DB for direct query when needed.
func (r *Reader) DB() *sql.DB {
	return r.db
}
