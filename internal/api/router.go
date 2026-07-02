package api

import (
	"apihub/internal/crypto"
	"apihub/internal/repository"
	"apihub/internal/service"
	"apihub/internal/syncer"
	"apihub/internal/ws"
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Services holds all service dependencies for the API handlers.
type Services struct {
	Provider     *service.ProviderService
	Key          *service.KeyService
	Usage        *service.UsageService
	Stats        *service.StatsService
	Alert        *service.AlertService
	Subscription *service.SubscriptionService
	Frequency    *service.FrequencyService
	Webhook      *service.WebhookService
	SyncState    *service.SyncStateService
	Session      *service.SessionService
	Agent        *service.AgentService
}

// Register sets up all HTTP routes.
func Register(r *gin.Engine, db *sql.DB, store *crypto.Store, cfg AuthConfig, hub *ws.Hub) {
	// Initialize repositories
	providerRepo := repository.NewProviderRepo(db)
	keyRepo := repository.NewKeyRepo(db)
	keyAuditRepo := repository.NewKeyAuditRepo(db)
	usageRepo := repository.NewUsageRepo(db)
	statsRepo := repository.NewStatsRepo(db)
	alertRepo := repository.NewAlertRepo(db)
	subscriptionRepo := repository.NewSubscriptionRepo(db)
	frequencyRepo := repository.NewFrequencyRepo(db)
	webhookRepo := repository.NewWebhookRepo(db)
	syncStateRepo := repository.NewSyncStateRepo(db)
	sessionRepo := repository.NewSessionRepo(db)
	bucketRepo := repository.NewBucketRepo(db)
	agentRepo := repository.NewAgentRepo(db)

	// Initialize services
	services := &Services{
		Provider:     service.NewProviderService(providerRepo, keyRepo, usageRepo),
		Key:          service.NewKeyService(keyRepo, store).WithAuditLog(keyAuditRepo),
		Usage:        service.NewUsageService(usageRepo, keyRepo),
		Stats:        service.NewStatsService(statsRepo),
		Alert:        service.NewAlertService(alertRepo),
		Subscription: service.NewSubscriptionService(subscriptionRepo),
		Frequency:    service.NewFrequencyService(frequencyRepo),
		Webhook:      service.NewWebhookService(webhookRepo),
		SyncState:    service.NewSyncStateService(syncStateRepo),
		Session:      service.NewSessionService(sessionRepo, bucketRepo),
		Agent:        service.NewAgentService(agentRepo),
	}

	api := r.Group("/api/v1")
	authMW := OptionalAuthMiddleware(cfg)
	sensitiveMW := SensitiveAuthMiddleware(cfg)

	auth := api.Group("/auth")
	RegisterAuth(auth, db, cfg)

	protected := api.Group("")
	protected.Use(authMW)
	registerProviders(protected.Group("/providers"), services.Provider)
	registerKeys(protected.Group("/keys"), services.Key, sensitiveMW, keyAuditRepo)
	registerUsage(protected.Group("/usage"), services.Usage)
	registerStats(protected.Group("/stats"), services.Stats)
	RegisterAlerts(protected.Group("/alerts"), services.Alert)
	RegisterSubscriptions(protected.Group("/subscriptions"), services.Subscription)
	RegisterFrequency(protected.Group("/frequency"), services.Frequency)
	RegisterExport(protected.Group("/export"), services.Usage)
	RegisterProviderDetail(protected.Group("/providers"), services.Provider)
	RegisterPlayground(protected.Group("/playground"), services.Key, store, sensitiveMW)
	RegisterWebhook(protected.Group("/webhooks"), services.Webhook)
	registerSessions(protected.Group("/sessions"), services.Session)
	RegisterScan(protected.Group("/scan"), services.Provider, services.Key)
	RegisterAgents(protected.Group("/agents"), services.Agent)

	// WebSocket endpoint for real-time updates
	r.GET("/ws", ws.Handler(hub))
}

// RegisterSyncRoutes registers sync endpoints on the protected group.
func RegisterSyncRoutes(g *gin.RouterGroup, keySvc *service.KeyService, registry *syncer.Registry, mgr *syncer.Manager, store *crypto.Store, trigger CCSwitchTrigger, syncStateSvc *service.SyncStateService) {
	RegisterSync(g, keySvc, registry, mgr, store, syncStateSvc)
	g.POST("/sync/ccswitch", TriggerCCSwitchSync(trigger))
}

// SyncStatusHandler returns the sync status for all sources.
func SyncStatusHandler(svc *service.SyncStateService) gin.HandlerFunc {
	return func(c *gin.Context) {
		states, err := svc.ListAll()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, states)
	}
}
