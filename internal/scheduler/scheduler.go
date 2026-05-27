package scheduler

import (
	"apihub/internal/alert"
	"apihub/internal/sync"
	"apihub/internal/syncer"
	"context"
	"database/sql"
	"log"
	"time"
)

// Scheduler runs background sync and alert tasks.
type Scheduler struct {
	db           *sql.DB
	syncMgr      *syncer.Manager
	registry     *syncer.Registry
	alertEngine  *alert.Engine
	ccService    *sync.CCSwitchService
	syncInterval time.Duration
	syncerInterval time.Duration
	cancel       context.CancelFunc
}

// New creates a scheduler.
func New(
	db *sql.DB,
	syncMgr *syncer.Manager,
	registry *syncer.Registry,
	alertEngine *alert.Engine,
	ccService *sync.CCSwitchService,
	syncInterval, syncerInterval time.Duration,
) *Scheduler {
	return &Scheduler{
		db:             db,
		syncMgr:        syncMgr,
		registry:       registry,
		alertEngine:    alertEngine,
		ccService:      ccService,
		syncInterval:   syncInterval,
		syncerInterval: syncerInterval,
	}
}

// Start launches background goroutines. Initial cc-switch sync runs once.
func (s *Scheduler) Start(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	s.cancel = cancel

	go func() {
		if err := s.ccService.RunInitialOrIncremental(); err != nil {
			log.Printf("cc-switch initial sync: %v", err)
		}
	}()

	go s.runTicker(ctx, "data-sync", s.syncInterval, func() {
		if err := s.ccService.RunIncremental(); err != nil {
			log.Printf("cc-switch incremental: %v", err)
		}
	})

	go s.runTicker(ctx, "syncer", s.syncerInterval, func() {
		s.runSyncers(ctx)
	})

	go s.runTicker(ctx, "alerts", 5*time.Minute, func() {
		if err := s.alertEngine.RunOnce(); err != nil {
			log.Printf("alert check: %v", err)
		}
	})

	log.Println("scheduler started")
}

// Stop cancels background tasks.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

// TriggerCCSwitch runs cc-switch sync once (incremental if already synced).
func (s *Scheduler) TriggerCCSwitch() error {
	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM sync_state WHERE source = 'ccswitch' AND status = 'ok'").Scan(&count); err == nil && count > 0 {
		return s.ccService.RunIncremental()
	}
	return s.ccService.RunFull()
}

func (s *Scheduler) runTicker(ctx context.Context, name string, interval time.Duration, fn func()) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Printf("scheduler %s stopped", name)
			return
		case <-ticker.C:
			fn()
		}
	}
}

func (s *Scheduler) runSyncers(ctx context.Context) {
	from := time.Now().AddDate(0, 0, -7)
	to := time.Now()

	for _, name := range s.registry.Names() {
		if err := s.syncMgr.SyncProvider(ctx, name, from, to); err != nil {
			log.Printf("syncer %s: %v", name, err)
		}
	}
}

// LoadWebhooks configures alert notifier from database.
func LoadWebhooks(db *sql.DB, engine *alert.Engine) {
	rows, err := db.Query(`SELECT url, headers FROM webhook_settings WHERE enabled = 1`)
	if err != nil {
		return
	}
	defer rows.Close()

	var webhooks []alert.WebhookConfig
	for rows.Next() {
		var url, headersStr string
		if err := rows.Scan(&url, &headersStr); err != nil {
			continue
		}
		webhooks = append(webhooks, alert.WebhookConfig{URL: url})
	}
	if len(webhooks) > 0 {
		engine.SetNotifier(alert.NewNotifier(webhooks))
		log.Printf("loaded %d webhook(s)", len(webhooks))
	}
}
