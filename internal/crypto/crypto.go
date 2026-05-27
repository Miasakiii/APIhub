package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"golang.org/x/crypto/hkdf"
)

const keyFile = ".master_key"

// Store holds the derived keys for encryption and signing.
type Store struct {
	dataKey  []byte // AES-256 key for API key encryption
	jwtKey   []byte // JWT signing secret
}

// Init loads or generates the master key and derives sub-keys.
// Returns the store and a boolean indicating if a new master key was generated.
func Init(dataDir string) (*Store, bool, error) {
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, false, fmt.Errorf("create data dir: %w", err)
	}
	master, isNew, err := loadOrGenerateMasterKey(dataDir)
	if err != nil {
		return nil, false, err
	}

	s := &Store{
		dataKey: deriveKey(master, "apihub-data-key"),
		jwtKey:  deriveKey(master, "apihub-jwt-secret"),
	}
	// Zero master after derivation
	for i := range master {
		master[i] = 0
	}
	return s, isNew, nil
}

// Encrypt AES-256-GCM encrypts plaintext, returns (nonce+ciphertext).
func (s *Store) Encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.dataKey)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return aesGCM.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt decrypts a nonce+ciphertext blob.
func (s *Store) Decrypt(blob []byte) ([]byte, error) {
	block, err := aes.NewCipher(s.dataKey)
	if err != nil {
		return nil, err
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := aesGCM.NonceSize()
	if len(blob) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := blob[:nonceSize], blob[nonceSize:]
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}

// JWTKey returns the derived JWT signing secret (caller should copy it).
func (s *Store) JWTKey() []byte {
	return s.jwtKey
}

// KeyHash returns sha256(plaintext)[:16] for deduplication.
func KeyHash(plaintext []byte) string {
	sum := sha256.Sum256(plaintext)
	return hex.EncodeToString(sum[:8])
}

func loadOrGenerateMasterKey(dataDir string) ([]byte, bool, error) {
	path := filepath.Join(dataDir, keyFile)
	if data, err := os.ReadFile(path); err == nil && len(data) == 64 {
		key := make([]byte, 32)
		n, err := hex.Decode(key, data[:64])
		if err == nil && n == 32 {
			return key, false, nil
		}
	}

	// Generate new 256-bit key
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, false, fmt.Errorf("generate master key: %w", err)
	}
	if err := os.WriteFile(path, []byte(hex.EncodeToString(key)), 0600); err != nil {
		return nil, false, fmt.Errorf("write master key: %w", err)
	}
	return key, true, nil
}

func deriveKey(master []byte, info string) []byte {
	r := hkdf.New(sha256.New, master, nil, []byte(info))
	key := make([]byte, 32)
	if _, err := io.ReadFull(r, key); err != nil {
		panic("hkdf derive: " + err.Error())
	}
	return key
}
