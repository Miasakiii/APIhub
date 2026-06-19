package repository

import (
	apihubDB "apihub/internal/db"
	"testing"
)

// setupTestDB creates an in-memory SQLite database with migrations applied.
func setupTestDB(t *testing.T) *apihubDB.DB {
	t.Helper()
	db, err := apihubDB.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

// seedProvider inserts a test provider into the database.
func seedProvider(t *testing.T, db *apihubDB.DB, id, name, ptype string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO providers (id, name, type, enabled) VALUES (?, ?, ?, 1)`, id, name, ptype)
	if err != nil {
		t.Fatalf("seed provider: %v", err)
	}
}
