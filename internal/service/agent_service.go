package service

import (
	"apihub/internal/model"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
)

var ErrAgentNotFound = errors.New("agent not found")

// AgentRepository defines the interface for agent data access.
type AgentRepository interface {
	List() ([]model.Agent, error)
	GetByID(id string) (*model.Agent, error)
	Create(a model.Agent) error
	Update(id string, a model.Agent) error
	Delete(id string) error
}

// AgentService handles agent business logic.
type AgentService struct {
	repo AgentRepository
}

// NewAgentService creates a new AgentService.
func NewAgentService(repo AgentRepository) *AgentService {
	return &AgentService{repo: repo}
}

// List returns all agents.
func (s *AgentService) List() ([]model.Agent, error) {
	return s.repo.List()
}

// GetByID returns an agent by ID.
func (s *AgentService) GetByID(id string) (*model.Agent, error) {
	a, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAgentNotFound
		}
		return nil, err
	}
	return a, nil
}

// Create creates a new agent with a generated ID.
func (s *AgentService) Create(a model.Agent) (model.Agent, error) {
	a.ID = genAgentID()
	if err := s.repo.Create(a); err != nil {
		return a, err
	}
	return a, nil
}

// Update updates an existing agent.
func (s *AgentService) Update(id string, a model.Agent) error {
	return s.repo.Update(id, a)
}

// Delete removes an agent by ID.
func (s *AgentService) Delete(id string) error {
	return s.repo.Delete(id)
}

func genAgentID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
