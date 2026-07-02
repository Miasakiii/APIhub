package api

import (
	"apihub/internal/repository"
	"apihub/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// registerKeys registers API key management endpoints.
func registerKeys(g *gin.RouterGroup, svc *service.KeyService, sensitiveMW gin.HandlerFunc, auditRepo *repository.KeyAuditRepo) {
	g.POST("", func(c *gin.Context) {
		var req struct {
			ProviderID string `json:"provider_id" binding:"required"`
			Key        string `json:"key" binding:"required"`
			Name       string `json:"name"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		result, err := svc.Create(service.CreateRequest{
			ProviderID: req.ProviderID,
			Key:        req.Key,
			Name:       req.Name,
		})
		if err != nil {
			if err == service.ErrKeyExists {
				c.JSON(http.StatusConflict, gin.H{"error": "key already exists"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, result)
	})

	g.GET("", func(c *gin.Context) {
		keys, err := svc.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, keys)
	})

	g.GET("/:id/decrypt", sensitiveMW, func(c *gin.Context) {
		key, err := svc.Decrypt(c.Param("id"))
		if err != nil {
			if err == service.ErrKeyNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "decrypt failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"key": key})
	})

	g.POST("/:id/revoke", func(c *gin.Context) {
		if err := svc.Revoke(c.Param("id")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	g.DELETE("/:id", func(c *gin.Context) {
		if err := svc.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	// Key audit log
	if auditRepo != nil {
		g.GET("/:id/audit", func(c *gin.Context) {
			logs, err := auditRepo.ListByKeyID(c.Param("id"))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"audit_logs": logs})
		})
	}
}
