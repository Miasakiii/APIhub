package api

import (
	"apihub/internal/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterExport registers export-related endpoints.
func RegisterExport(g *gin.RouterGroup, svc *service.UsageService) {
	g.GET("/csv", func(c *gin.Context) {
		data, filename, err := svc.ExportCSV()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Header("Content-Type", "text/csv; charset=utf-8")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		c.String(http.StatusOK, data)
	})
}
