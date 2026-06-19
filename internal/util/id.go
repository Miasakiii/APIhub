package util

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateID creates a random 32-character hex string (16 bytes).
// Used as unique identifiers for database records.
func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
