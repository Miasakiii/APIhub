package repository

import (
	"apihub/internal/model"
	"database/sql"
)

// KeyRepo handles API key database operations.
type KeyRepo struct {
	db *sql.DB
}

// NewKeyRepo creates a new KeyRepo.
func NewKeyRepo(db *sql.DB) *KeyRepo {
	return &KeyRepo{db: db}
}

// CountByHash returns the number of keys with the given hash.
func (r *KeyRepo) CountByHash(keyHash string) (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM api_keys WHERE key_hash = ?", keyHash).Scan(&count)
	return count, err
}

// Create inserts a new API key with source 'manual'.
func (r *KeyRepo) Create(id, providerID, keyHash string, encrypted []byte, name string) error {
	return r.CreateWithSource(id, providerID, keyHash, encrypted, name, "manual")
}

// CreateWithSource inserts a new API key with a custom source.
func (r *KeyRepo) CreateWithSource(id, providerID, keyHash string, encrypted []byte, name, source string) error {
	_, err := r.db.Exec(`
		INSERT INTO api_keys (id, provider_id, key_hash, key_encrypted, name, source, status)
		VALUES (?, ?, ?, ?, ?, ?, 'active')
	`, id, providerID, keyHash, encrypted, name, source)
	return err
}

// List returns all API keys ordered by creation date.
func (r *KeyRepo) List() ([]model.APIKey, error) {
	rows, err := r.db.Query(`
		SELECT id, provider_id, key_hash, name, source, status, balance_usd, last_checked, created_at
		FROM api_keys ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []model.APIKey
	for rows.Next() {
		var k model.APIKey
		var lastChecked, createdAt sql.NullString
		if err := rows.Scan(&k.ID, &k.ProviderID, &k.KeyHash, &k.Name, &k.Source,
			&k.Status, &k.BalanceUSD, &lastChecked, &createdAt); err != nil {
			return nil, err
		}
		if lastChecked.Valid {
			t, _ := parseTime(lastChecked.String)
			k.LastChecked = &t
		}
		if createdAt.Valid {
			k.CreatedAt, _ = parseTime(createdAt.String)
		}
		keys = append(keys, k)
	}
	if keys == nil {
		keys = []model.APIKey{}
	}
	return keys, nil
}

// GetEncryptedKey returns the encrypted key blob for the given key ID.
func (r *KeyRepo) GetEncryptedKey(id string) ([]byte, error) {
	var encrypted []byte
	err := r.db.QueryRow("SELECT key_encrypted FROM api_keys WHERE id = ?", id).Scan(&encrypted)
	if err != nil {
		return nil, err
	}
	return encrypted, nil
}

// Revoke updates a key's status to 'revoked'.
func (r *KeyRepo) Revoke(id string) error {
	_, err := r.db.Exec("UPDATE api_keys SET status = 'revoked' WHERE id = ?", id)
	return err
}

// Delete removes a key by ID.
func (r *KeyRepo) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM api_keys WHERE id = ?", id)
	return err
}

// CountActive returns the number of active keys.
func (r *KeyRepo) CountActive() (int64, error) {
	var count int64
	err := r.db.QueryRow("SELECT COUNT(*) FROM api_keys WHERE status = 'active'").Scan(&count)
	return count, err
}

// ListByProvider returns all API keys for a specific provider.
func (r *KeyRepo) ListByProvider(providerID string) ([]model.APIKey, error) {
	rows, err := r.db.Query(`
		SELECT id, provider_id, key_hash, name, source, status, balance_usd, last_checked, created_at
		FROM api_keys WHERE provider_id = ? ORDER BY created_at DESC
	`, providerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []model.APIKey
	for rows.Next() {
		var k model.APIKey
		var lastChecked, createdAt sql.NullString
		if err := rows.Scan(&k.ID, &k.ProviderID, &k.KeyHash, &k.Name, &k.Source,
			&k.Status, &k.BalanceUSD, &lastChecked, &createdAt); err != nil {
			return nil, err
		}
		if lastChecked.Valid {
			t, _ := parseTime(lastChecked.String)
			k.LastChecked = &t
		}
		if createdAt.Valid {
			k.CreatedAt, _ = parseTime(createdAt.String)
		}
		keys = append(keys, k)
	}
	if keys == nil {
		keys = []model.APIKey{}
	}
	return keys, nil
}

// GetByKeyID returns key details including encrypted data and provider info.
func (r *KeyRepo) GetByKeyID(id string) (*model.APIKeyDetail, error) {
	var d model.APIKeyDetail
	err := r.db.QueryRow(`
		SELECT k.provider_id, COALESCE(p.syncer, ''), k.key_encrypted, COALESCE(p.base_url, ''), COALESCE(p.type, '')
		FROM api_keys k
		JOIN providers p ON k.provider_id = p.id
		WHERE k.id = ?
	`, id).Scan(&d.ProviderID, &d.Syncer, &d.Encrypted, &d.BaseURL, &d.ProviderType)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// KeyDetail contains key metadata with encrypted data for decryption.
type KeyDetail struct {
	ProviderID   string
	Syncer       string
	Encrypted    []byte
	BaseURL      string
	ProviderType string
}
