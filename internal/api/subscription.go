package api

import (
	"apihub/internal/model"
	"apihub/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterSubscriptions registers subscription-related endpoints.
func RegisterSubscriptions(g *gin.RouterGroup, svc *service.SubscriptionService) {
	// List subscriptions
	g.GET("", func(c *gin.Context) {
		subs, err := svc.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, subs)
	})

	// Create subscription
	g.POST("", func(c *gin.Context) {
		var req model.Subscription
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sub, err := svc.Create(req)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, sub)
	})

	// Get subscription detail
	g.GET("/:id", func(c *gin.Context) {
		sub, err := svc.GetByID(c.Param("id"))
		if err != nil {
			if err == service.ErrSubscriptionNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "subscription not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, sub)
	})

	// Update subscription
	g.PUT("/:id", func(c *gin.Context) {
		var req model.Subscription
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if err := svc.Update(c.Param("id"), req); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})

	// Delete subscription
	g.DELETE("/:id", func(c *gin.Context) {
		if err := svc.Delete(c.Param("id")); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Status(http.StatusNoContent)
	})
}
