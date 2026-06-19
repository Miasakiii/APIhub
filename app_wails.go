//go:build wails

package main

import (
	"apihub/internal/aggregator"
	"apihub/internal/alert"
	"apihub/internal/api"
	"apihub/internal/crypto"
	apihubDB "apihub/internal/db"
	"apihub/internal/repository"
	"apihub/internal/scanner"
	"apihub/internal/scheduler"
	"apihub/internal/service"
	"apihub/internal/sync"
	"apihub/internal/syncer"
	"apihub/internal/syncer/providers"
	"apihub/internal/ws"
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// WailsApp is the main Wails application struct.
type WailsApp struct {
	ctx        context.Context
	port       string
	store      *crypto.Store
	db         *apihubDB.DB
	sched      *scheduler.Scheduler
	hub        *ws.Hub
	minimizeToTray bool
}

// NewWailsApp creates a new WailsApp instance.
func NewWailsApp() *WailsApp {
	return &WailsApp{
		port: "8080",
	}
}

// startup is called when the Wails app starts.
func (a *WailsApp) startup(ctx context.Context) {
	a.ctx = ctx
	a.initBackend()
}

// domReady is called when the DOM is ready.
func (a *WailsApp) domReady(ctx context.Context) {
	// Request notification permission
	runtime.WindowExecJS(ctx, `if('Notification' in window && Notification.permission==='default')Notification.requestPermission()`)
}

// shutdown is called when the Wails app shuts down.
func (a *WailsApp) shutdown(ctx context.Context) {
	if a.sched != nil {
		a.sched.Stop()
	}
	if a.db != nil {
		a.db.Close()
	}
	log.Println("APIHub shutdown complete")
}

// initBackend initializes all backend components and starts the Gin server.
func (a *WailsApp) initBackend() {
	// Data directory
	dataDir := os.Getenv("APIHUB_DATA_DIR")
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".apihub")
	}
	os.MkdirAll(dataDir, 0755)

	// Crypto
	store, isNew, err := crypto.Init(dataDir)
	if err != nil {
		log.Fatalf("crypto init: %v", err)
	}
	a.store = store
	if isNew {
		log.Println("NEW MASTER KEY GENERATED - back up your data directory!")
	}

	// Config
	authCfg := api.LoadAuthConfig(store)
	syncCfg := api.LoadSyncConfig()

	// Database
	dbPath := filepath.Join(dataDir, "apihub.db")
	db, err := apihubDB.Open(dbPath)
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	a.db = db
	if err := db.Migrate(); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Printf("database opened: %s", dbPath)

	// Model pricing
	if err := aggregator.LoadPricing(db.DB); err != nil {
		log.Fatalf("load pricing: %v", err)
	}

	// Scan local configs (report only)
	envFindings := scanner.ScanEnv()
	configFindings := scanner.ScanConfigs("")
	allFindings := append(envFindings, configFindings...)
	if len(allFindings) > 0 {
		log.Printf("local config scan: found %d API key(s)", len(allFindings))
	}

	// WebSocket hub
	hub := ws.NewHub()
	go hub.Run()
	a.hub = hub

	// Aggregator
	agg := aggregator.New(db.DB)
	agg.SetHub(hub)

	// Syncer
	syncRegistry := syncer.NewRegistry()
	syncRegistry.Register(&providers.OpenRouterSyncer{})
	syncRegistry.Register(&providers.OpenAISyncer{})
	syncRegistry.Register(&providers.AnthropicSyncer{})
	syncRegistry.Register(providers.NewRelaySyncer("one-api"))
	syncRegistry.Register(providers.NewRelaySyncer("new-api"))
	syncMgr := syncer.NewManager(db.DB, syncRegistry, store)
	syncMgr.SetHub(hub)

	ccService := &sync.CCSwitchService{
		DB:     db.DB,
		CCPath: syncCfg.CCSwitchPath,
	}

	// Alert engine
	alertEngine := alert.NewEngine(db.DB)
	alertEngine.SetHub(hub)
	scheduler.LoadWebhooks(db.DB, alertEngine)

	// Scheduler
	sched := scheduler.New(
		db.DB,
		syncMgr,
		syncRegistry,
		alertEngine,
		ccService,
		syncCfg.SyncInterval,
		syncCfg.SyncerInterval,
	)
	a.sched = sched
	sched.Start(context.Background())

	// Find available port
	a.port = findAvailablePort("8080")

	// Gin router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(api.CORSMiddleware(authCfg))
	api.Register(r, db.DB, store, authCfg, hub)

	// Register sync routes
	keyRepo := repository.NewKeyRepo(db.DB)
	keySvc := service.NewKeyService(keyRepo, store)
	syncStateSvc := service.NewSyncStateService(repository.NewSyncStateRepo(db.DB))
	protected := r.Group("/api/v1")
	protected.Use(api.OptionalAuthMiddleware(authCfg))
	api.RegisterSyncRoutes(protected, keySvc, syncRegistry, syncMgr, store, sched, syncStateSvc)

	// Health endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "port": a.port})
	})

	// Start Gin server in background
	go func() {
		addr := "127.0.0.1:" + a.port
		log.Printf("API server starting on http://%s", addr)
		if err := r.Run(addr); err != nil {
			log.Printf("gin server error: %v", err)
		}
	}()
}

// GetAPIPort returns the port the API server is running on.
func (a *WailsApp) GetAPIPort() string {
	return a.port
}

// GetAPIURL returns the full API base URL.
func (a *WailsApp) GetAPIURL() string {
	return fmt.Sprintf("http://127.0.0.1:%s", a.port)
}

// OpenExternalURL opens a URL in the default browser.
func (a *WailsApp) OpenExternalURL(url string) {
	runtime.BrowserOpenURL(a.ctx, url)
}

// GetVersion returns the application version.
func (a *WailsApp) GetVersion() string {
	return "0.14.0"
}

// onBeforeClose is called when the user tries to close the window.
// Returns true to prevent close (minimize to taskbar instead).
func (a *WailsApp) onBeforeClose() bool {
	if a.minimizeToTray {
		runtime.WindowMinimise(a.ctx)
		return true // prevent close, minimize to taskbar
	}
	return false // allow close
}

// MinimizeToTray hides the window to system tray.
func (a *WailsApp) MinimizeToTray() {
	runtime.WindowMinimise(a.ctx)
}

// ShowWindow brings the window back from tray.
func (a *WailsApp) ShowWindow() {
	runtime.WindowShow(a.ctx)
	runtime.WindowUnminimise(a.ctx)
}

// SetMinimizeToTray sets whether to minimize to tray on close.
func (a *WailsApp) SetMinimizeToTray(enable bool) {
	a.minimizeToTray = enable
}

// GetMinimizeToTray returns whether minimize to tray is enabled.
func (a *WailsApp) GetMinimizeToTray() bool {
	return a.minimizeToTray
}

// ShowNotification displays a native OS notification.
func (a *WailsApp) ShowNotification(title, message string) {
	runtime.EventsEmit(a.ctx, "notification", map[string]string{
		"title":   title,
		"message": message,
	})
}

// SetAutoStart enables or disables auto-start on login.
// Uses Windows Startup folder shortcut approach.
func (a *WailsApp) SetAutoStart(enable bool) error {
	startupDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	shortcutPath := filepath.Join(startupDir, "APIHub.url")

	if enable {
		executable, err := os.Executable()
		if err != nil {
			return fmt.Errorf("get executable: %w", err)
		}
		content := fmt.Sprintf("[InternetShortcut]\nURL=file:///%s\n", filepath.ToSlash(executable))
		return os.WriteFile(shortcutPath, []byte(content), 0644)
	}
	// Remove shortcut
	os.Remove(shortcutPath)
	return nil
}

// IsAutoStartEnabled returns whether auto-start is enabled.
func (a *WailsApp) IsAutoStartEnabled() bool {
	startupDir := filepath.Join(os.Getenv("APPDATA"), "Microsoft", "Windows", "Start Menu", "Programs", "Startup")
	shortcutPath := filepath.Join(startupDir, "APIHub.url")
	_, err := os.Stat(shortcutPath)
	return err == nil
}

// GetDataDir returns the data directory path.
func (a *WailsApp) GetDataDir() string {
	dataDir := os.Getenv("APIHUB_DATA_DIR")
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".apihub")
	}
	return dataDir
}

// findAvailablePort tries the preferred port, then falls back to a random port.
func findAvailablePort(preferred string) string {
	ln, err := net.Listen("tcp", "127.0.0.1:"+preferred)
	if err == nil {
		ln.Close()
		return preferred
	}
	// Find a random available port
	ln, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return preferred // fallback
	}
	defer ln.Close()
	return fmt.Sprintf("%d", ln.Addr().(*net.TCPAddr).Port)
}
