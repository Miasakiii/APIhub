package api

import (
	"apihub/internal/crypto"
	"os"
	"time"
)

// AuthConfig holds authentication and CORS settings.
type AuthConfig struct {
	Enabled       bool
	JWTExpiry     time.Duration
	AllowRegister bool
	CORSOrigin    string
	Store         *crypto.Store
}

// LoadAuthConfig reads auth-related settings from environment variables.
func LoadAuthConfig(store *crypto.Store) AuthConfig {
	enabled := os.Getenv("APIHUB_AUTH_ENABLED") == "true"

	expiry := 168 * time.Hour
	if v := os.Getenv("APIHUB_JWT_EXPIRY"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			expiry = d
		}
	}

	allowRegister := os.Getenv("APIHUB_ALLOW_REGISTER") != "false"

	cors := os.Getenv("APIHUB_CORS_ORIGIN")
	if cors == "" {
		cors = "http://localhost:5173"
	}

	return AuthConfig{
		Enabled:       enabled,
		JWTExpiry:     expiry,
		AllowRegister: allowRegister,
		CORSOrigin:    cors,
		Store:         store,
	}
}

// SyncConfig holds background sync intervals.
type SyncConfig struct {
	SyncInterval   time.Duration
	SyncerInterval time.Duration
	CCSwitchPath   string
}

// LoadSyncConfig reads sync-related settings from environment variables.
func LoadSyncConfig() SyncConfig {
	syncInterval := 5 * time.Minute
	if v := os.Getenv("APIHUB_SYNC_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			syncInterval = d
		}
	}

	syncerInterval := 30 * time.Minute
	if v := os.Getenv("APIHUB_SYNCER_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			syncerInterval = d
		}
	}

	return SyncConfig{
		SyncInterval:   syncInterval,
		SyncerInterval: syncerInterval,
		CCSwitchPath:   os.Getenv("APIHUB_CC_SWITCH_PATH"),
	}
}
