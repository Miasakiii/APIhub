package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var errAuthNotConfigured = errors.New("auth store not configured")

// RegisterAuth registers auth-related endpoints.
func RegisterAuth(g *gin.RouterGroup, db *sql.DB, cfg AuthConfig) {
	g.GET("/config", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"enabled":        cfg.Enabled,
			"allow_register": cfg.AllowRegister,
		})
	})

	g.POST("/register", func(c *gin.Context) {
		if cfg.Enabled {
			var userCount int
			_ = db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount)
			if userCount > 0 && !cfg.AllowRegister {
				c.JSON(http.StatusForbidden, gin.H{"error": "registration is disabled"})
				return
			}
		}

		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		id := generateID()
		_, err = db.Exec("INSERT INTO users (id, username, password_hash) VALUES (?, ?, ?)",
			id, req.Username, string(hash))
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
			return
		}

		resp := gin.H{"id": id, "username": req.Username}
		if cfg.Enabled {
			token, err := issueToken(cfg, id)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
				return
			}
			resp["token"] = token
		}
		c.JSON(http.StatusCreated, resp)
	})

	g.POST("/login", func(c *gin.Context) {
		if !cfg.Enabled {
			c.JSON(http.StatusBadRequest, gin.H{"error": "authentication is disabled"})
			return
		}

		var req struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var id, hash string
		err := db.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", req.Username).Scan(&id, &hash)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}

		token, err := issueToken(cfg, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"token":    token,
			"username": req.Username,
		})
	})

	g.GET("/me", func(c *gin.Context) {
		if !cfg.Enabled {
			c.JSON(http.StatusOK, gin.H{"auth": "disabled"})
			return
		}
		AuthMiddleware(cfg)(c)
		if c.IsAborted() {
			return
		}
		userID, _ := c.Get("userID")
		c.JSON(http.StatusOK, gin.H{"id": userID})
	})
}

// AuthMiddleware requires a valid JWT when auth is enabled.
func AuthMiddleware(cfg AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}

		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}

		userID, err := validateToken(cfg, parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Next()
	}
}

// OptionalAuthMiddleware applies auth only when enabled.
func OptionalAuthMiddleware(cfg AuthConfig) gin.HandlerFunc {
	return AuthMiddleware(cfg)
}

// SensitiveAuthMiddleware requires auth when enabled (for decrypt/playground).
func SensitiveAuthMiddleware(cfg AuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.Enabled {
			c.Next()
			return
		}
		AuthMiddleware(cfg)(c)
	}
}

type jwtClaims struct {
	jwt.RegisteredClaims
}

func issueToken(cfg AuthConfig, userID string) (string, error) {
	if cfg.Store == nil {
		return "", errAuthNotConfigured
	}
	now := time.Now()
	claims := jwtClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(cfg.JWTExpiry)),
			Issuer:    "apihub",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	key := make([]byte, len(cfg.Store.JWTKey()))
	copy(key, cfg.Store.JWTKey())
	return token.SignedString(key)
}

func validateToken(cfg AuthConfig, tokenStr string) (string, error) {
	if cfg.Store == nil {
		return "", errAuthNotConfigured
	}
	key := make([]byte, len(cfg.Store.JWTKey()))
	copy(key, cfg.Store.JWTKey())

	token, err := jwt.ParseWithClaims(tokenStr, &jwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(*jwtClaims)
	if !ok || !token.Valid || claims.Subject == "" {
		return "", errors.New("invalid token claims")
	}
	return claims.Subject, nil
}

// GenerateTokenForTest exposes token generation for tests.
func GenerateTokenForTest(cfg AuthConfig, userID string) (string, error) {
	return issueToken(cfg, userID)
}

// ValidateTokenForTest exposes token validation for tests.
func ValidateTokenForTest(cfg AuthConfig, token string) (string, error) {
	return validateToken(cfg, token)
}
