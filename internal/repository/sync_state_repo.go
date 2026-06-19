package repository

import (
	"apihub/internal/model"
	"database/sql"
)

// SyncStateRepo handles sync state database operations.
type SyncStateRepo struct {
	db *sql.DB
}

// NewSyncStateRepo creates a new SyncStateRepo.
func NewSyncStateRepo(db *sql.DB) *SyncStateRepo {
	return &SyncStateRepo{db: db}
}

// ListAll returns all sync states ordered by update time.
func (r *SyncStateRepo) ListAll() ([]model.SyncState, error) {
	rows, err := r.db.Query(`
		SELECT id, source, last_sync, offset_val, status, error, updated_at
		FROM sync_state ORDER BY updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var states []model.SyncState
	for rows.Next() {
		var s model.SyncState
		var lastSync, errStr, updatedAt sql.NullString
		if err := rows.Scan(&s.ID, &s.Source, &lastSync, &s.Offset, &s.Status, &errStr, &updatedAt); err != nil {
			return nil, err
		}
		if lastSync.Valid {
			t, _ := parseTime(lastSync.String)
			s.LastSync = &t
		}
		if errStr.Valid {
			s.Error = errStr.String
		}
		if updatedAt.Valid {
			s.UpdatedAt, _ = parseTime(updatedAt.String)
		}
		states = append(states, s)
	}
	if states == nil {
		states = []model.SyncState{}
	}
	return states, nil
}
