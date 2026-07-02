package api

import (
	"apihub/internal/crypto"
	apihubDB "apihub/internal/db"
	"apihub/internal/repository"
	"apihub/internal/service"
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// testEnv holds all test dependencies.
type testEnv struct {
	DB      *apihubDB.DB
	Store   *crypto.Store
	Router  *gin.Engine
	Svc     *Services
	AuthCfg AuthConfig
}

// setupTestEnv creates a full test environment with in-memory DB, services, and router.
func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	db, err := apihubDB.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	dir := t.TempDir()
	store, _, err := crypto.Init(dir)
	if err != nil {
		t.Fatalf("crypto init: %v", err)
	}

	// Repos
	providerRepo := repository.NewProviderRepo(db.DB)
	keyRepo := repository.NewKeyRepo(db.DB)
	usageRepo := repository.NewUsageRepo(db.DB)
	statsRepo := repository.NewStatsRepo(db.DB)
	alertRepo := repository.NewAlertRepo(db.DB)
	subscriptionRepo := repository.NewSubscriptionRepo(db.DB)
	frequencyRepo := repository.NewFrequencyRepo(db.DB)
	webhookRepo := repository.NewWebhookRepo(db.DB)
	syncStateRepo := repository.NewSyncStateRepo(db.DB)
	sessionRepo := repository.NewSessionRepo(db.DB)
	bucketRepo := repository.NewBucketRepo(db.DB)
	agentRepo := repository.NewAgentRepo(db.DB)

	// Services
	svcs := &Services{
		Provider:     service.NewProviderService(providerRepo, keyRepo, usageRepo),
		Key:          service.NewKeyService(keyRepo, store),
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

	authCfg := AuthConfig{Enabled: false, CORSOrigin: "*"}
	r := gin.New()
	r.Use(CORSMiddleware(authCfg))

	api := r.Group("/api/v1")
	RegisterAlerts(api.Group("/alerts"), svcs.Alert)
	RegisterSubscriptions(api.Group("/subscriptions"), svcs.Subscription)
	registerProviders(api.Group("/providers"), svcs.Provider)
	registerKeys(api.Group("/keys"), svcs.Key, func(c *gin.Context) { c.Next() }, nil)
	registerUsage(api.Group("/usage"), svcs.Usage)
	registerStats(api.Group("/stats"), svcs.Stats)
	RegisterFrequency(api.Group("/frequency"), svcs.Frequency)
	RegisterAgents(api.Group("/agents"), svcs.Agent)

	return &testEnv{DB: db, Store: store, Router: r, Svc: svcs, AuthCfg: authCfg}
}

// seedProvider inserts a test provider.
func seedProvider(t *testing.T, db *apihubDB.DB, id, name, ptype string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO providers (id, name, type, enabled) VALUES (?, ?, ?, 1)`, id, name, ptype)
	if err != nil {
		t.Fatalf("seed provider: %v", err)
	}
}

// doRequest performs an HTTP request and returns the recorder.
func doRequest(r *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// parseJSON parses the response body into a map.
func parseJSON(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("parse JSON: %v (body: %s)", err, w.Body.String())
	}
	return result
}

// parseJSONArray parses the response body into a slice.
func parseJSONArray(t *testing.T, w *httptest.ResponseRecorder) []interface{} {
	t.Helper()
	var result []interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("parse JSON array: %v (body: %s)", err, w.Body.String())
	}
	return result
}
