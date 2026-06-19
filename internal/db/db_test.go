package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpen_InMemory(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	defer db.Close()

	// Should be able to query
	var v int
	if err := db.QueryRow("PRAGMA user_version").Scan(&v); err != nil {
		t.Fatalf("query user_version: %v", err)
	}
	if v != 0 {
		t.Fatalf("expected version 0, got %d", v)
	}
}

func TestOpen_File(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("open file db: %v", err)
	}
	defer db.Close()

	// File should exist
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Fatal("db file should exist")
	}
}

func TestMigrate_FromZero(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Check version after migration
	v, err := db.getVersion()
	if err != nil {
		t.Fatalf("get version: %v", err)
	}
	if v != currentVersion {
		t.Fatalf("expected version %d, got %d", currentVersion, v)
	}
}

func TestMigrate_Idempotent(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	// First migration
	if err := db.Migrate(); err != nil {
		t.Fatalf("first migrate: %v", err)
	}

	// Second migration should be a no-op
	if err := db.Migrate(); err != nil {
		t.Fatalf("second migrate: %v", err)
	}

	v, _ := db.getVersion()
	if v != currentVersion {
		t.Fatalf("expected version %d, got %d", currentVersion, v)
	}
}

func TestMigrate_CreatesTables(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Check that key tables exist
	tables := []string{"providers", "api_keys", "usage_records", "daily_stats", "alerts", "subscriptions", "users"}
	for _, table := range tables {
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
		if err != nil {
			t.Fatalf("table %s: %v", table, err)
		}
	}
}

func TestColumnExists(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// providers table should have 'name' column
	if !db.columnExists("providers", "name") {
		t.Fatal("expected 'name' column in providers")
	}

	// providers table should not have 'nonexistent' column
	if db.columnExists("providers", "nonexistent") {
		t.Fatal("expected 'nonexistent' column to not exist")
	}
}
