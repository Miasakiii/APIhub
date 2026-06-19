package jsonl

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"apihub/internal/aggregator"
	"apihub/internal/model"
)

// SyncResult holds the outcome of syncing a single JSONL file.
type SyncResult struct {
	Path      string
	OldOffset int64
	NewOffset int64
	Records   int
	Error     error
}

// SyncIncremental runs incremental sync for all JSONL files.
// It skips files that have already been fully synced.
func SyncIncremental(db *sql.DB, calc *CostCalc, baseDir string) ([]SyncResult, error) {
	files, err := ScanProjects(baseDir)
	if err != nil {
		return nil, fmt.Errorf("scan projects: %w", err)
	}

	var results []SyncResult
	for _, f := range files {
		res := syncSingleFile(db, calc, f)
		results = append(results, res)
	}
	return results, nil
}

func syncSingleFile(db *sql.DB, calc *CostCalc, f FileMeta) SyncResult {
	res := SyncResult{Path: f.Path}

	oldOffset, err := getFileOffset(db, f.Path)
	if err != nil {
		res.Error = fmt.Errorf("get offset: %w", err)
		return res
	}
	res.OldOffset = oldOffset

	// If file shrank (e.g. rotated/truncated), reset to 0
	if oldOffset > f.SizeBytes {
		oldOffset = 0
	}

	// Skip if already fully synced
	if oldOffset >= f.SizeBytes {
		res.NewOffset = oldOffset
		return res
	}

	fh, _, err := OpenFileAt(f.Path, oldOffset)
	if err != nil {
		res.Error = fmt.Errorf("open file: %w", err)
		_ = setFileOffset(db, f.Path, oldOffset, "error", res.Error.Error())
		return res
	}
	defer fh.Close()

	records, finalOffset, err := ParseFileAt(fh, oldOffset)
	if err != nil {
		res.Error = fmt.Errorf("parse file: %w", err)
		_ = setFileOffset(db, f.Path, oldOffset, "error", res.Error.Error())
		return res
	}

	if len(records) > 0 {
		var usageRecords []model.UsageRecord
		for _, r := range records {
			cost := calc.CalcCost(r)
			usageRecords = append(usageRecords, model.UsageRecord{
				ID:           genID(),
				ProviderID:   "claude-code",
				Model:        r.Model,
				AgentID:      "claude-code",
				InputTokens:  r.InputTokens,
				OutputTokens: r.OutputTokens,
				CacheRead:    r.CacheRead,
				CacheCreate:  r.CacheCreate,
				CostUSD:      cost,
				Source:       "jsonl",
				Timestamp:    parseTimestamp(r.Timestamp),
			})
		}

		if err := aggregator.ImportFromCCSwitch(db, usageRecords); err != nil {
			res.Error = fmt.Errorf("import: %w", err)
			_ = setFileOffset(db, f.Path, oldOffset, "error", res.Error.Error())
			return res
		}
	}

	res.NewOffset = finalOffset
	res.Records = len(records)
	_ = setFileOffset(db, f.Path, finalOffset, "ok", "")
	return res
}

// fileStateKey returns a unique sync_state.source value for a JSONL file.
func fileStateKey(path string) string {
	return "jsonl:" + path
}

// getFileOffset queries sync_state for the last synced byte offset of a file.
func getFileOffset(db *sql.DB, path string) (int64, error) {
	var offset int64
	err := db.QueryRow("SELECT offset_val FROM sync_state WHERE source = ?", fileStateKey(path)).Scan(&offset)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return offset, nil
}

// setFileOffset updates (or inserts) the sync_state row for a JSONL file.
func setFileOffset(db *sql.DB, path string, offset int64, status string, errMsg string) error {
	id := fmt.Sprintf("jsonl-%x", time.Now().UnixNano())
	_, err := db.Exec(`
		INSERT INTO sync_state (id, source, last_sync, offset_val, status, error, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(source) DO UPDATE SET
			last_sync = excluded.last_sync,
			offset_val = excluded.offset_val,
			status = excluded.status,
			error = excluded.error,
			updated_at = CURRENT_TIMESTAMP
	`, id, fileStateKey(path), offset, status, errMsg)
	return err
}

func genID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func parseTimestamp(ts string) time.Time {
	t, _ := time.Parse(time.RFC3339, ts)
	if t.IsZero() {
		return time.Now()
	}
	return t
}
