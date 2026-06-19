package api

import (
	"apihub/internal/crypto"
	"apihub/internal/service"
	"apihub/internal/syncer"
	"apihub/internal/syncer/providers"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterSync registers sync-related endpoints on the protected API group.
func RegisterSync(g *gin.RouterGroup, keySvc *service.KeyService, registry *syncer.Registry, mgr *syncer.Manager, store *crypto.Store, syncStateSvc *service.SyncStateService) {
	g.GET("/syncers", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"syncers": registry.Names(),
		})
	})

	g.GET("/sync/status", SyncStatusHandler(syncStateSvc))

	g.POST("/sync/:provider_id", func(c *gin.Context) {
		providerID := c.Param("provider_id")

		fromStr := c.Query("from")
		toStr := c.Query("to")

		from := time.Now().AddDate(0, 0, -7)
		to := time.Now()

		if fromStr != "" {
			if t, err := time.Parse("2006-01-02", fromStr); err == nil {
				from = t
			}
		}
		if toStr != "" {
			if t, err := time.Parse("2006-01-02", toStr); err == nil {
				to = t
			}
		}

		go func() {
			if err := mgr.SyncProvider(context.Background(), providerID, from, to); err != nil {
				fmt.Printf("sync %s: %v\n", providerID, err)
			}
		}()

		c.JSON(http.StatusAccepted, gin.H{
			"status":   "accepted",
			"provider": providerID,
			"from":     from.Format("2006-01-02"),
			"to":       to.Format("2006-01-02"),
		})
	})

	g.POST("/keys/:id/validate", func(c *gin.Context) {
		keyID := c.Param("id")

		detail, err := keySvc.GetByKeyID(keyID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
			return
		}

		plain, err := store.Decrypt(detail.Encrypted)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "decrypt failed"})
			return
		}
		apiKey := string(plain)

		syncerName := detail.Syncer
		if syncerName == "" {
			syncerName = detail.ProviderID
		}

		s, ok := registry.Get(syncerName)
		if !ok {
			c.JSON(http.StatusOK, gin.H{
				"valid":   true,
				"message": "no syncer registered for this provider; key stored successfully",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 15*time.Second)
		defer cancel()

		if err := s.ValidateKey(ctx, apiKey, detail.BaseURL); err != nil {
			c.JSON(http.StatusOK, gin.H{"valid": false, "error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"valid": true, "syncer": s.Name()})
	})
}

// Ensure syncer providers are imported
var (
	_ = &providers.OpenRouterSyncer{}
	_ = providers.NewRelaySyncer("one-api")
)
