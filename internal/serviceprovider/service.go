package serviceprovider

import (
	"context"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/pkg/log"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Service encapsulates usecase logic for service providers.
type Service interface {
	Get(ctx context.Context, id string) (ServiceProvider, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, input CreateServiceProviderRequest) (ServiceProvider, error)
}

// ServiceProvider represents the data about a service provider.
type ServiceProvider struct {
	entity.ServiceProvider
}

// CreateServiceProviderRequest represents a service provider creation request.
type CreateServiceProviderRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Validate validates the CreateServiceProviderRequest fields.
func (m CreateServiceProviderRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required, validation.Length(1, 128)),
		validation.Field(&m.Email, validation.Required, is.Email),
	)
}

type service struct {
	repo   Repository
	logger log.Logger
}

// NewService creates a new service provider service.
func NewService(repo Repository, logger log.Logger) Service {
	return service{repo, logger}
}

// Get returns the service provider with the specified the service provider ID.
func (s service) Get(ctx context.Context, id string) (ServiceProvider, error) {
	serviceProvider, err := s.repo.Get(ctx, id)
	if err != nil {
		return ServiceProvider{}, err
	}
	return ServiceProvider{serviceProvider}, nil
}

// Create creates a new service provider.
func (s service) Create(ctx context.Context, req CreateServiceProviderRequest) (ServiceProvider, error) {
	if err := req.Validate(); err != nil {
		return ServiceProvider{}, err
	}
	id := entity.GenerateID()
	now := time.Now()
	err := s.repo.Create(ctx, entity.ServiceProvider{
		ID:        id,
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return ServiceProvider{}, err
	}
	return s.Get(ctx, id)
}

// Count returns the total number of service provider in the storage.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}
