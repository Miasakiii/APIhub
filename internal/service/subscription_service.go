package service

import (
	"apihub/internal/model"
	"database/sql"
	"errors"
)

var ErrSubscriptionNotFound = errors.New("subscription not found")

// SubscriptionRepository defines the interface for subscription data access.
type SubscriptionRepository interface {
	List() ([]model.Subscription, error)
	GetByID(id string) (*model.Subscription, error)
	Create(s model.Subscription) error
	Update(id string, s model.Subscription) error
	Delete(id string) error
}

// SubscriptionService handles subscription business logic.
type SubscriptionService struct {
	repo SubscriptionRepository
}

// NewSubscriptionService creates a new SubscriptionService.
func NewSubscriptionService(repo SubscriptionRepository) *SubscriptionService {
	return &SubscriptionService{repo: repo}
}

// List returns all subscriptions.
func (s *SubscriptionService) List() ([]model.Subscription, error) {
	return s.repo.List()
}

// GetByID returns a subscription by ID.
func (s *SubscriptionService) GetByID(id string) (*model.Subscription, error) {
	sub, err := s.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSubscriptionNotFound
		}
		return nil, err
	}
	return sub, nil
}

// Create creates a new subscription with a generated ID.
func (s *SubscriptionService) Create(sub model.Subscription) (model.Subscription, error) {
	sub.ID = generateID()
	if err := s.repo.Create(sub); err != nil {
		return sub, err
	}
	return sub, nil
}

// Update updates an existing subscription.
func (s *SubscriptionService) Update(id string, sub model.Subscription) error {
	return s.repo.Update(id, sub)
}

// Delete removes a subscription by ID.
func (s *SubscriptionService) Delete(id string) error {
	return s.repo.Delete(id)
}
