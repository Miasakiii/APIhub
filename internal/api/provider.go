package api

import (
	"apihub/internal/model"
	"apihub/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// registerProviders registers provider CRUD endpoints.
func registerProviders(g *gin.RouterGroup, svc *service.ProviderService) {
	g.GET("", func(c *gin.Context) {
		providers, err := svc.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, providers)
	})

	g.POST("", func(c *gin.Context) {
		var req struct {
			Name       string `json:"name" binding:"required"`
			Type       string `json:"type" binding:"required"`
			BaseURL    string `json:"base_url"`
			ConsoleURL string `json:"console_url"`
			TopUpURL   string `json:"topup_url"`
			DocsURL    string `json:"docs_url"`
			Syncer     string `json:"syncer"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		provider, err := svc.Create(model.Provider{
			Name:       req.Name,
			Type:       req.Type,
			BaseURL:    req.BaseURL,
			ConsoleURL: req.ConsoleURL,
			TopUpURL:   req.TopUpURL,
			DocsURL:    req.DocsURL,
			Syncer:     req.Syncer,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, provider)
	})

	g.DELETE("/:id", func(c *gin.Context) {
		if err := svc.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})
}
