package customer

import (
	"context"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/pkg/log"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Service encapsulates usecase logic for customers.
type Service interface {
	Get(ctx context.Context, id string) (Customer, error)
	Create(ctx context.Context, input CreateCustomerRequest) (Customer, error)
	Count(ctx context.Context) (int, error)
}

// Customer represents the data about a customer.
type Customer struct {
	entity.Customer
}

// CreateCustomerRequest represents a customer creation request.
type CreateCustomerRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// Validate validates the CreateCustomerRequest fields.
func (m CreateCustomerRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.Name, validation.Required, validation.Length(1, 128)),
		validation.Field(&m.Email, validation.Required, is.Email),
	)
}

type service struct {
	repo   Repository
	logger log.Logger
}

// NewService creates a new customer service.
func NewService(repo Repository, logger log.Logger) Service {
	return service{repo, logger}
}

// Get returns the customer with the specified the customer ID.
func (s service) Get(ctx context.Context, id string) (Customer, error) {
	customer, err := s.repo.Get(ctx, id)
	if err != nil {
		return Customer{}, err
	}
	return Customer{customer}, nil
}

// Create creates a new customer.
func (s service) Create(ctx context.Context, req CreateCustomerRequest) (Customer, error) {
	if err := req.Validate(); err != nil {
		return Customer{}, err
	}
	id := entity.GenerateID()
	now := time.Now()
	err := s.repo.Create(ctx, entity.Customer{
		ID:        id,
		Name:      req.Name,
		Email:     req.Email,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return Customer{}, err
	}
	return s.Get(ctx, id)
}

// Count returns the total number of customers in the storage.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}
