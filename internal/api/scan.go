package api

import (
	"apihub/internal/model"
	"apihub/internal/scanner"
	"apihub/internal/service"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterScan registers scan endpoints on the given router group.
func RegisterScan(g *gin.RouterGroup, providerSvc *service.ProviderService, keySvc *service.KeyService) {
	g.POST("", handleScan)
	g.POST("/import", handleScanImport(providerSvc, keySvc))
}

// handleScan triggers a local config scan and returns findings with masked keys.
func handleScan(c *gin.Context) {
	envFindings := scanner.ScanEnv()
	configFindings := scanner.ScanConfigs("")

	all := append(envFindings, configFindings...)

	// Mask keys in response — only show first 8 and last 4 chars
	type maskedFinding struct {
		scanner.Finding
		MaskedKey string `json:"masked_key"`
	}

	results := make([]maskedFinding, len(all))
	for i, f := range all {
		results[i] = maskedFinding{
			Finding:   f,
			MaskedKey: maskKey(f.Key),
		}
		// Clear plaintext key from response
		results[i].Key = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"findings": results,
		"total":    len(results),
	})
}

// scanImportRequest is the request body for scan import.
// The frontend sends the list of finding indices to import.
// The backend re-scans internally to get actual keys.
type scanImportRequest struct {
	Indices []int `json:"indices"` // which findings to import (by index); empty = import all
}

// handleScanImport re-scans locally and imports selected findings.
func handleScanImport(providerSvc *service.ProviderService, keySvc *service.KeyService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req scanImportRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Re-scan to get actual keys
		envFindings := scanner.ScanEnv()
		configFindings := scanner.ScanConfigs("")
		all := append(envFindings, configFindings...)

		if len(all) == 0 {
			c.JSON(http.StatusOK, gin.H{"results": []importResult{}, "total": 0})
			return
		}

		// Filter by indices if provided
		var toImport []scanner.Finding
		if len(req.Indices) > 0 {
			for _, idx := range req.Indices {
				if idx >= 0 && idx < len(all) {
					toImport = append(toImport, all[idx])
				}
			}
		} else {
			toImport = all
		}

		results := make([]importResult, 0, len(toImport))
		for _, f := range toImport {
			results = append(results, importOneFinding(providerSvc, keySvc, f))
		}

		c.JSON(http.StatusOK, gin.H{
			"results": results,
			"total":   len(results),
		})
	}
}

// importResult contains the result of importing a single finding.
type importResult struct {
	Name       string `json:"name"`
	ProviderID string `json:"provider_id,omitempty"`
	KeyID      string `json:"key_id,omitempty"`
	Status     string `json:"status"` // created, skipped, error
	Message    string `json:"message,omitempty"`
}

// importOneFinding creates a provider (if needed) and API key for a single finding.
func importOneFinding(providerSvc *service.ProviderService, keySvc *service.KeyService, f scanner.Finding) importResult {
	if f.Key == "" {
		return importResult{
			Name:    f.Name,
			Status:  "skipped",
			Message: "empty API key",
		}
	}

	// Find or create provider by type
	provider, err := providerSvc.GetByType(f.ProviderType)
	if err != nil {
		p, createErr := providerSvc.Create(model.Provider{
			Name:    f.Name,
			Type:    f.ProviderType,
			BaseURL: f.BaseURL,
			Syncer:  f.ProviderType,
		})
		if createErr != nil {
			log.Printf("scan import: create provider %s: %v", f.Name, createErr)
			return importResult{
				Name:    f.Name,
				Status:  "error",
				Message: "failed to create provider: " + createErr.Error(),
			}
		}
		provider = &p
	}

	// Create API key
	keyResult, err := keySvc.CreateWithSource(service.CreateRequest{
		ProviderID: provider.ID,
		Key:        f.Key,
		Name:       f.Source + " (auto)",
	}, "auto")
	if err != nil {
		if err == service.ErrKeyExists {
			return importResult{
				Name:       f.Name,
				ProviderID: provider.ID,
				Status:     "skipped",
				Message:    "key already exists",
			}
		}
		log.Printf("scan import: create key for %s: %v", f.Name, err)
		return importResult{
			Name:       f.Name,
			ProviderID: provider.ID,
			Status:     "error",
			Message:    "failed to create key: " + err.Error(),
		}
	}

	return importResult{
		Name:       f.Name,
		ProviderID: provider.ID,
		KeyID:      keyResult.ID,
		Status:     "created",
	}
}

// maskKey returns a masked version of the key for display.
func maskKey(key string) string {
	if len(key) <= 12 {
		return "****"
	}
	return key[:8] + "****" + key[len(key)-4:]
}
