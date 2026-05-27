package sync

import (
	"apihub/internal/aggregator"
	"apihub/internal/model"
	"apihub/sources/ccswitch"
	"apihub/sources/jsonl"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"
)

// CCSwitchService handles cc-switch.db synchronization.
type CCSwitchService struct {
	DB     *sql.DB
	CCPath string
}

// ResolvePath returns the cc-switch.db path.
func (s *CCSwitchService) ResolvePath() string {
	if s.CCPath != "" {
		return s.CCPath
	}
	return ccswitch.DefaultPath()
}

// RunInitialOrIncremental performs full import on first run, incremental thereafter.
func (s *CCSwitchService) RunInitialOrIncremental() error {
	var count int
	if err := s.DB.QueryRow("SELECT COUNT(*) FROM sync_state WHERE source = 'ccswitch' AND status = 'ok'").Scan(&count); err == nil && count > 0 {
		return s.RunIncremental()
	}
	return s.RunFull()
}

// RunFull imports all proxy_request_logs from cc-switch.
func (s *CCSwitchService) RunFull() error {
	ccPath := s.ResolvePath()
	if _, err := os.Stat(ccPath); err != nil {
		return fmt.Errorf("cc-switch.db not found: %w", err)
	}

	r, err := ccswitch.Open(ccPath)
	if err != nil {
		return fmt.Errorf("open cc-switch: %w", err)
	}
	defer r.Close()

	if err := r.ValidateSchema(); err != nil {
		return fmt.Errorf("schema check: %w", err)
	}

	start := time.Now()

	logs, err := r.FetchProxyLogs()
	if err != nil {
		return fmt.Errorf("fetch logs: %w", err)
	}

	prices, err := r.FetchModelPrices()
	if err != nil {
		return fmt.Errorf("fetch prices: %w", err)
	}
	priceIdx := ccswitch.NewPriceIndex(prices)

	if err := mirrorProviders(s.DB, r); err != nil {
		return err
	}

	records := logsToRecords(logs, priceIdx)
	log.Printf("importing %d cc-switch usage records...", len(records))
	if err := aggregator.ImportFromCCSwitch(s.DB, records); err != nil {
		return fmt.Errorf("import: %w", err)
	}

	if err := SyncJSONLIncremental(s.DB, priceIdx, ""); err != nil {
		log.Printf("JSONL import: %v", err)
	}

	if err := updateSyncState(s.DB, "ok", ""); err != nil {
		return err
	}

	log.Printf("cc-switch full sync complete: %d records in %s", len(records), time.Since(start))
	return nil
}

// RunIncremental fetches proxy logs since last sync.
func (s *CCSwitchService) RunIncremental() error {
	ccPath := s.ResolvePath()
	if _, err := os.Stat(ccPath); err != nil {
		return fmt.Errorf("cc-switch.db not found: %w", err)
	}

	r, err := ccswitch.Open(ccPath)
	if err != nil {
		return fmt.Errorf("open cc-switch: %w", err)
	}
	defer r.Close()

	if err := r.ValidateSchema(); err != nil {
		return fmt.Errorf("schema check: %w", err)
	}

	since, err := lastSyncTime(s.DB, "ccswitch")
	if err != nil {
		return err
	}

	prices, err := r.FetchModelPrices()
	if err != nil {
		return fmt.Errorf("fetch prices: %w", err)
	}
	priceIdx := ccswitch.NewPriceIndex(prices)

	if err := mirrorProviders(s.DB, r); err != nil {
		return err
	}

	logs, err := r.FetchProxyLogsSince(since)
	if err != nil {
		return fmt.Errorf("fetch incremental logs: %w", err)
	}

	if len(logs) == 0 {
		if err := SyncJSONLIncremental(s.DB, priceIdx, ""); err != nil {
			log.Printf("JSONL sync: %v", err)
		}
		return updateSyncState(s.DB, "ok", "")
	}

	records := logsToRecords(logs, priceIdx)
	if err := aggregator.ImportFromCCSwitch(s.DB, records); err != nil {
		return fmt.Errorf("import: %w", err)
	}

	if err := SyncJSONLIncremental(s.DB, priceIdx, ""); err != nil {
		log.Printf("JSONL sync: %v", err)
	}

	log.Printf("cc-switch incremental sync: +%d records", len(records))
	return updateSyncState(s.DB, "ok", "")
}

// SyncJSONLIncremental runs incremental JSONL sync.
func SyncJSONLIncremental(db *sql.DB, priceIdx ccswitch.PriceIndex, baseDir string) error {
	calc := &jsonl.CostCalc{
		Lookup: func(modelID string) (float64, float64, float64, float64) {
			p := priceIdx.Lookup(modelID)
			return p.InputCost, p.OutputCost, p.CacheRead, p.CacheCreate
		},
	}

	results, err := jsonl.SyncIncremental(db, calc, baseDir)
	if err != nil {
		return err
	}

	var totalRecords int
	for _, res := range results {
		if res.Error != nil {
			log.Printf("JSONL sync error %s: %v", res.Path, res.Error)
			continue
		}
		if res.Records > 0 {
			log.Printf("JSONL synced %s: +%d records (offset %d → %d)",
				res.Path, res.Records, res.OldOffset, res.NewOffset)
		}
		totalRecords += res.Records
	}

	if totalRecords > 0 {
		log.Printf("JSONL incremental sync complete: %d new records", totalRecords)
	}
	return nil
}

func mirrorProviders(db *sql.DB, r *ccswitch.Reader) error {
	ccProviders, err := r.FetchProviders()
	if err != nil {
		return fmt.Errorf("fetch providers: %w", err)
	}

	for _, cp := range ccProviders {
		var existing int
		db.QueryRow("SELECT COUNT(*) FROM providers WHERE id = ?", cp.ID).Scan(&existing)
		if existing == 0 {
			db.Exec(`
				INSERT INTO providers (id, name, type, base_url, enabled)
				VALUES (?, ?, ?, ?, 1)
			`, cp.ID, cp.Name, cp.AppType, cp.BaseURL)
		} else if cp.BaseURL != "" {
			db.Exec("UPDATE providers SET base_url = ? WHERE id = ? AND (base_url IS NULL OR base_url = '')", cp.BaseURL, cp.ID)
		}
	}
	return nil
}

func logsToRecords(logs []ccswitch.ProxyLog, priceIdx ccswitch.PriceIndex) []model.UsageRecord {
	var records []model.UsageRecord
	for _, l := range logs {
		p := priceIdx.Lookup(l.Model)
		cost := ccswitch.CalcCost(p, l.InputTokens, l.OutputTokens, l.CacheRead, l.CacheCreate)
		records = append(records, model.UsageRecord{
			ID:           genID(),
			ProviderID:   l.ProviderID,
			Model:        l.Model,
			InputTokens:  l.InputTokens,
			OutputTokens: l.OutputTokens,
			CacheRead:    l.CacheRead,
			CacheCreate:  l.CacheCreate,
			CostUSD:      cost,
			Source:       "ccswitch",
			Timestamp:    l.CreatedAt,
		})
	}
	return records
}

func lastSyncTime(db *sql.DB, source string) (time.Time, error) {
	var lastSync sql.NullString
	err := db.QueryRow("SELECT last_sync FROM sync_state WHERE source = ?", source).Scan(&lastSync)
	if err == sql.ErrNoRows || !lastSync.Valid {
		return time.Time{}, nil
	}
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
	} {
		if t, err := time.Parse(layout, lastSync.String); err == nil {
			return t, nil
		}
	}
	return time.Time{}, nil
}

func updateSyncState(db *sql.DB, status, errMsg string) error {
	_, err := db.Exec(`
		INSERT OR REPLACE INTO sync_state (id, source, last_sync, offset_val, status, error, updated_at)
		VALUES ('ccswitch-sync', 'ccswitch', ?, 0, ?, ?, CURRENT_TIMESTAMP)
	`, time.Now(), status, errMsg)
	return err
}

func genID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
