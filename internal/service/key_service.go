package service

import (
	"apihub/internal/crypto"
	"apihub/internal/model"
	"apihub/internal/util"
	"errors"
)

var (
	ErrKeyExists   = errors.New("key already exists")
	ErrKeyNotFound = errors.New("key not found")
)

// KeyRepository defines the interface for key data access.
type KeyRepository interface {
	CountByHash(keyHash string) (int, error)
	Create(id, providerID, keyHash string, encrypted []byte, name string) error
	CreateWithSource(id, providerID, keyHash string, encrypted []byte, name, source string) error
	List() ([]model.APIKey, error)
	GetEncryptedKey(id string) ([]byte, error)
	Revoke(id string) error
	Delete(id string) error
	CountActive() (int64, error)
}

// KeyEncryptor defines the interface for key encryption operations.
type KeyEncryptor interface {
	Encrypt(data []byte) ([]byte, error)
	Decrypt(data []byte) ([]byte, error)
}

// KeyService handles API key business logic.
type KeyService struct {
	repo  KeyRepository
	store KeyEncryptor
}

// NewKeyService creates a new KeyService.
func NewKeyService(repo KeyRepository, store KeyEncryptor) *KeyService {
	return &KeyService{repo: repo, store: store}
}

// CreateRequest contains the data needed to create a new API key.
type CreateRequest struct {
	ProviderID string
	Key        string
	Name       string
}

// CreateResult contains the result of creating a new API key.
type CreateResult struct {
	ID      string `json:"id"`
	KeyHash string `json:"key_hash"`
	Name    string `json:"name"`
	Source  string `json:"source"`
	Status  string `json:"status"`
}

// Create adds a new API key with encryption and source 'manual'.
func (s *KeyService) Create(req CreateRequest) (*CreateResult, error) {
	return s.CreateWithSource(req, "manual")
}

// CreateWithSource adds a new API key with encryption and a custom source.
func (s *KeyService) CreateWithSource(req CreateRequest, source string) (*CreateResult, error) {
	kh := crypto.KeyHash([]byte(req.Key))

	// Check for duplicate
	count, err := s.repo.CountByHash(kh)
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, ErrKeyExists
	}

	// Encrypt the key
	encrypted, err := s.store.Encrypt([]byte(req.Key))
	if err != nil {
		return nil, err
	}

	id := util.GenerateID()
	if err := s.repo.CreateWithSource(id, req.ProviderID, kh, encrypted, req.Name, source); err != nil {
		return nil, err
	}

	return &CreateResult{
		ID:      id,
		KeyHash: kh,
		Name:    req.Name,
		Source:  source,
		Status:  "active",
	}, nil
}

// List returns all API keys.
func (s *KeyService) List() ([]model.APIKey, error) {
	return s.repo.List()
}

// Decrypt returns the decrypted key for the given key ID.
func (s *KeyService) Decrypt(id string) (string, error) {
	encrypted, err := s.repo.GetEncryptedKey(id)
	if err != nil {
		return "", ErrKeyNotFound
	}
	plain, err := s.store.Decrypt(encrypted)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// Revoke marks a key as revoked.
func (s *KeyService) Revoke(id string) error {
	return s.repo.Revoke(id)
}

// Delete removes a key.
func (s *KeyService) Delete(id string) error {
	return s.repo.Delete(id)
}

// CountActive returns the number of active keys.
func (s *KeyService) CountActive() (int64, error) {
	return s.repo.CountActive()
}
