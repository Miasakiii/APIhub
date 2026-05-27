package crypto

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestInit(t *testing.T) {
	tmpDir := t.TempDir()
	store, isNew, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}
	if store == nil {
		t.Fatal("store is nil")
	}
	if !isNew {
		t.Fatal("expected isNew=true for first init")
	}

	// Verify master key file was created
	masterKeyPath := filepath.Join(tmpDir, ".master_key")
	if _, err := os.Stat(masterKeyPath); os.IsNotExist(err) {
		t.Fatal("master key file not created")
	}

	// Second init should return isNew=false
	_, isNew2, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("second Init failed: %v", err)
	}
	if isNew2 {
		t.Fatal("expected isNew=false for second init")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	tmpDir := t.TempDir()
	store, _, err := Init(tmpDir)
	if err != nil {
		t.Fatalf("Init failed: %v", err)
	}

	plaintext := []byte("hello world this is a secret")
	encrypted, err := store.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}
	if len(encrypted) == 0 {
		t.Fatal("encrypted data is empty")
	}

	decrypted, err := store.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if !bytes.Equal(decrypted, plaintext) {
		t.Fatalf("decrypted data does not match: got %s, want %s", decrypted, plaintext)
	}
}

func TestKeyHash(t *testing.T) {
	key1 := []byte("test-key-123")
	key2 := []byte("test-key-123")
	key3 := []byte("different-key")

	hash1 := KeyHash(key1)
	hash2 := KeyHash(key2)
	hash3 := KeyHash(key3)

	if hash1 != hash2 {
		t.Fatal("same key should produce same hash")
	}
	if hash1 == hash3 {
		t.Fatal("different keys should produce different hashes")
	}
	if len(hash1) == 0 {
		t.Fatal("hash should not be empty")
	}
}
