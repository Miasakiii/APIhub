package api

import (
	"crypto/rand"
	"encoding/hex"
)

// generateID generates a random hex ID.
func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
