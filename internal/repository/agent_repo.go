package repository

import (
	"apihub/internal/model"
	"database/sql"
	"time"
)

// AgentRepo handles agent database operations.
type AgentRepo struct {
	db *sql.DB
}

// NewAgentRepo creates a new AgentRepo.
func NewAgentRepo(db *sql.DB) *AgentRepo {
	return &AgentRepo{db: db}
}

// List returns all agents ordered by name.
func (r *AgentRepo) List() ([]model.Agent, error) {
	rows, err := r.db.Query(`
		SELECT id, name, type, icon, created_at, updated_at
		FROM agents ORDER BY name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []model.Agent
	for rows.Next() {
		var a model.Agent
		var createdAt, updatedAt sql.NullString
		if err := rows.Scan(&a.ID, &a.Name, &a.Type, &a.Icon, &createdAt, &updatedAt); err != nil {
			continue
		}
		if createdAt.Valid {
			a.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
		}
		if updatedAt.Valid {
			a.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt.String)
		}
		agents = append(agents, a)
	}
	if agents == nil {
		agents = []model.Agent{}
	}
	return agents, nil
}

// GetByID returns a single agent by ID.
func (r *AgentRepo) GetByID(id string) (*model.Agent, error) {
	var a model.Agent
	var createdAt, updatedAt sql.NullString
	err := r.db.QueryRow(`
		SELECT id, name, type, icon, created_at, updated_at
		FROM agents WHERE id = ?
	`, id).Scan(&a.ID, &a.Name, &a.Type, &a.Icon, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	if createdAt.Valid {
		a.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
	}
	if updatedAt.Valid {
		a.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updatedAt.String)
	}
	return &a, nil
}

// Create inserts a new agent.
func (r *AgentRepo) Create(a model.Agent) error {
	_, err := r.db.Exec(`
		INSERT INTO agents (id, name, type, icon)
		VALUES (?, ?, ?, ?)
	`, a.ID, a.Name, a.Type, a.Icon)
	return err
}

// Update updates an existing agent.
func (r *AgentRepo) Update(id string, a model.Agent) error {
	_, err := r.db.Exec(`
		UPDATE agents SET name=?, type=?, icon=?, updated_at=CURRENT_TIMESTAMP
		WHERE id=?
	`, a.Name, a.Type, a.Icon, id)
	return err
}

// Delete removes an agent by ID.
func (r *AgentRepo) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM agents WHERE id = ?", id)
	return err
}
