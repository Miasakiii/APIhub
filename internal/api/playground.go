package api

import (
	"apihub/internal/crypto"
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterPlayground registers the playground endpoint.
func RegisterPlayground(g *gin.RouterGroup, db *sql.DB, store *crypto.Store, sensitiveMW gin.HandlerFunc) {
	g.POST("/chat", sensitiveMW, func(c *gin.Context) {
		var req struct {
			KeyID    string `json:"key_id" binding:"required"`
			Model    string `json:"model" binding:"required"`
			Prompt   string `json:"prompt" binding:"required"`
			Protocol string `json:"protocol"` // "openai" (default) or "anthropic"
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if req.Protocol == "" {
			req.Protocol = "openai"
		}

		// Get key details
		var encryptedKey []byte
		var providerID, baseURL, providerType string
		err := db.QueryRow(`
			SELECT k.key_encrypted, k.provider_id, COALESCE(p.base_url, ''), COALESCE(p.type, '')
			FROM api_keys k
			JOIN providers p ON k.provider_id = p.id
			WHERE k.id = ?
		`, req.KeyID).Scan(&encryptedKey, &providerID, &baseURL, &providerType)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
			return
		}

		// Decrypt key
		plain, err := store.Decrypt(encryptedKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "decrypt failed"})
			return
		}
		apiKey := string(plain)

		// Auto-detect protocol from provider type if not explicitly set
		if req.Protocol == "openai" && providerType == "anthropic" {
			req.Protocol = "anthropic"
		}

		var url string
		var headers map[string]string
		var body io.Reader
		var payloadBytes []byte

		if req.Protocol == "anthropic" {
			// Anthropic Messages API
			if baseURL == "" {
				baseURL = "https://api.anthropic.com"
			}
			url = baseURL + "/v1/messages"

			payload := map[string]interface{}{
				"model":      req.Model,
				"max_tokens": 256,
				"messages": []map[string]string{
					{"role": "user", "content": req.Prompt},
				},
			}
			payloadBytes, _ = json.Marshal(payload)
			body = bytes.NewReader(payloadBytes)

			headers = map[string]string{
				"Content-Type":      "application/json",
				"x-api-key":         apiKey,
				"anthropic-version": "2023-06-01",
			}
		} else {
			// OpenAI Chat Completions API
			if baseURL == "" {
				baseURL = "https://api.openai.com/v1"
			}
			url = baseURL + "/chat/completions"

			payload := map[string]interface{}{
				"model": req.Model,
				"messages": []map[string]string{
					{"role": "user", "content": req.Prompt},
				},
				"max_tokens": 256,
			}
			payloadBytes, _ = json.Marshal(payload)
			body = bytes.NewReader(payloadBytes)

			headers = map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer " + apiKey,
			}
		}

		// Send request
		httpReq, err := http.NewRequest("POST", url, body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		for k, v := range headers {
			httpReq.Header.Set(k, v)
		}

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != 200 {
			c.JSON(resp.StatusCode, gin.H{
				"error":  "provider returned error",
				"status": resp.StatusCode,
				"body":   string(respBody),
			})
			return
		}

		// Parse response based on protocol
		var result map[string]interface{}
		if err := json.Unmarshal(respBody, &result); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var content string
		var usage interface{}

		if req.Protocol == "anthropic" {
			// Anthropic response: { content: [{ type: "text", text: "..." }], usage: {...} }
			if blocks, ok := result["content"].([]interface{}); ok && len(blocks) > 0 {
				if block, ok := blocks[0].(map[string]interface{}); ok {
					content, _ = block["text"].(string)
				}
			}
			usage = result["usage"]
		} else {
			// OpenAI response: { choices: [{ message: { content: "..." } }], usage: {...} }
			if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
				if choice, ok := choices[0].(map[string]interface{}); ok {
					if msg, ok := choice["message"].(map[string]interface{}); ok {
						content, _ = msg["content"].(string)
					}
				}
			}
			usage = result["usage"]
		}

		c.JSON(http.StatusOK, gin.H{
			"content":    content,
			"model":      req.Model,
			"protocol":   req.Protocol,
			"provider":   providerID,
			"usage":      usage,
			"created_at": result["created"],
		})
	})

	// Validate key endpoint
	g.POST("/validate", sensitiveMW, func(c *gin.Context) {
		var req struct {
			KeyID string `json:"key_id" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var encryptedKey []byte
		var baseURL string
		err := db.QueryRow(`
			SELECT k.key_encrypted, COALESCE(p.base_url, '')
			FROM api_keys k
			JOIN providers p ON k.provider_id = p.id
			WHERE k.id = ?
		`, req.KeyID).Scan(&encryptedKey, &baseURL)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "key not found"})
			return
		}

		plain, err := store.Decrypt(encryptedKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "decrypt failed"})
			return
		}

		if baseURL == "" {
			baseURL = "https://api.openai.com/v1"
		}

		httpReq, err := http.NewRequest("GET", baseURL+"/models", nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		httpReq.Header.Set("Authorization", "Bearer "+string(plain))

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(httpReq)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error(), "valid": false})
			return
		}
		defer resp.Body.Close()

		c.JSON(http.StatusOK, gin.H{
			"valid":   resp.StatusCode == 200,
			"status":  resp.StatusCode,
			"message": "Key validation complete",
		})
	})
}
