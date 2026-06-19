package repository

import (
	"apihub/internal/model"
	"database/sql"
	"fmt"
	"time"
)

// ProviderRepo handles provider database operations.
type ProviderRepo struct {
	db *sql.DB
}

// NewProviderRepo creates a new ProviderRepo.
func NewProviderRepo(db *sql.DB) *ProviderRepo {
	return &ProviderRepo{db: db}
}

// List returns all providers ordered by creation date.
func (r *ProviderRepo) List() ([]model.Provider, error) {
	rows, err := r.db.Query(`
		SELECT id, name, type, base_url, console_url, topup_url, docs_url, COALESCE(syncer, ''), enabled, created_at, updated_at
		FROM providers ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []model.Provider
	for rows.Next() {
		p, err := scanProvider(rows)
		if err != nil {
			return nil, err
		}
		providers = append(providers, p)
	}
	if providers == nil {
		providers = []model.Provider{}
	}
	return providers, nil
}

// Create inserts a new provider.
func (r *ProviderRepo) Create(p model.Provider) error {
	_, err := r.db.Exec(`
		INSERT INTO providers (id, name, type, base_url, console_url, topup_url, docs_url, syncer, api_key, enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, '', 1)
	`, p.ID, p.Name, p.Type, p.BaseURL, p.ConsoleURL, p.TopUpURL, p.DocsURL, p.Syncer)
	return err
}

// Delete removes a provider by ID.
func (r *ProviderRepo) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM providers WHERE id = ?", id)
	return err
}

// GetByType returns the first provider matching the given type.
func (r *ProviderRepo) GetByType(pType string) (*model.Provider, error) {
	row := r.db.QueryRow(`
		SELECT id, name, type, base_url, console_url, topup_url, docs_url, COALESCE(syncer, ''), enabled, created_at, updated_at
		FROM providers WHERE type = ? LIMIT 1
	`, pType)
	return scanProviderRow(row)
}

// GetByID returns a single provider by ID.
func (r *ProviderRepo) GetByID(id string) (*model.Provider, error) {
	row := r.db.QueryRow(`
		SELECT id, name, type, base_url, console_url, topup_url, docs_url, COALESCE(syncer, ''), enabled, created_at, updated_at
		FROM providers WHERE id = ?
	`, id)
	p, err := scanProviderRow(row)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func scanProvider(rows *sql.Rows) (model.Provider, error) {
	var p model.Provider
	var enabled int
	var baseURL, consoleURL, topupURL, docsURL, createdAt, updatedAt sql.NullString
	if err := rows.Scan(&p.ID, &p.Name, &p.Type, &baseURL, &consoleURL,
		&topupURL, &docsURL, &p.Syncer, &enabled, &createdAt, &updatedAt); err != nil {
		return p, err
	}
	p.BaseURL = nullStr(baseURL)
	p.ConsoleURL = nullStr(consoleURL)
	p.TopUpURL = nullStr(topupURL)
	p.DocsURL = nullStr(docsURL)
	p.Enabled = enabled == 1
	if createdAt.Valid {
		p.CreatedAt, _ = parseTime(createdAt.String)
	}
	if updatedAt.Valid {
		p.UpdatedAt, _ = parseTime(updatedAt.String)
	}
	return p, nil
}

func scanProviderRow(row *sql.Row) (*model.Provider, error) {
	var p model.Provider
	var enabled int
	var baseURL, consoleURL, topupURL, docsURL, createdAt, updatedAt sql.NullString
	if err := row.Scan(&p.ID, &p.Name, &p.Type, &baseURL, &consoleURL,
		&topupURL, &docsURL, &p.Syncer, &enabled, &createdAt, &updatedAt); err != nil {
		return nil, err
	}
	p.BaseURL = nullStr(baseURL)
	p.ConsoleURL = nullStr(consoleURL)
	p.TopUpURL = nullStr(topupURL)
	p.DocsURL = nullStr(docsURL)
	p.Enabled = enabled == 1
	if createdAt.Valid {
		p.CreatedAt, _ = parseTime(createdAt.String)
	}
	if updatedAt.Valid {
		p.UpdatedAt, _ = parseTime(updatedAt.String)
	}
	return &p, nil
}

func nullStr(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func parseTime(s string) (time.Time, error) {
	for _, layout := range []string{
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.000Z",
		time.RFC3339,
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse time: %s", s)
}
