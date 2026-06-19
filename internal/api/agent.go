package api

import (
	"apihub/internal/model"
	"apihub/internal/service"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterAgents registers agent endpoints on the given router group.
func RegisterAgents(g *gin.RouterGroup, svc *service.AgentService) {
	g.GET("", listAgents(svc))
	g.POST("", createAgent(svc))
	g.GET("/:id", getAgent(svc))
	g.PUT("/:id", updateAgent(svc))
	g.DELETE("/:id", deleteAgent(svc))
}

func listAgents(svc *service.AgentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		agents, err := svc.List()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, agents)
	}
}

func createAgent(svc *service.AgentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var a model.Agent
		if err := c.ShouldBindJSON(&a); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if a.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
			return
		}
		if a.Type == "" {
			a.Type = "cli"
		}
		created, err := svc.Create(a)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, created)
	}
}

func getAgent(svc *service.AgentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		a, err := svc.GetByID(id)
		if err != nil {
			if errors.Is(err, service.ErrAgentNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "agent not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, a)
	}
}

func updateAgent(svc *service.AgentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var a model.Agent
		if err := c.ShouldBindJSON(&a); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := svc.Update(id, a); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	}
}

func deleteAgent(svc *service.AgentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		if err := svc.Delete(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted"})
	}
}
