package rating

import (
	"context"
	"fmt"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/internal/customer"
	"github.com/berkaykrc/homerun-ratings-system/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/internal/serviceprovider"
	"github.com/berkaykrc/homerun-ratings-system/pkg/log"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Service encapsulates usecase logic for ratings.
type Service interface {
	Get(ctx context.Context, id string) (Rating, error)
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, input CreateRatingRequest) (Rating, error)
	GetAverageRatingByServiceProvider(ctx context.Context, serviceProviderID string) (AverageRating, error)
}

// Rating represents the data about a rating.
type Rating struct {
	entity.Rating
}

// AverageRating represents the average rating data for a service provider.
type AverageRating struct {
	ServiceProviderID string    `json:"serviceProviderId"`
	AverageRating     float64   `json:"averageRating"`
	TotalRatings      int       `json:"totalRatings"`
	LastUpdated       time.Time `json:"lastUpdated"`
}

// CreateRatingRequest represents a rating creation request.
type CreateRatingRequest struct {
	CustomerID        string `json:"customerId"`
	ServiceProviderID string `json:"serviceProviderId"`
	RatingValue       int    `json:"rating"`
	Comment           string `json:"comment"`
}

// Validate validates the CreateRatingRequest fields.
func (m CreateRatingRequest) Validate() error {
	return validation.ValidateStruct(&m,
		validation.Field(&m.CustomerID, validation.Required),
		validation.Field(&m.ServiceProviderID, validation.Required),
		validation.Field(&m.RatingValue, validation.Required, validation.Min(1), validation.Max(5)),
		validation.Field(&m.Comment, validation.Length(0, 100)),
	)
}

type service struct {
	repo                   Repository
	customerService        customer.Service
	serviceProviderService serviceprovider.Service
	logger                 log.Logger
}

// NewService creates a new rating service.
func NewService(repo Repository, customerService customer.Service, serviceProviderService serviceprovider.Service, logger log.Logger) Service {
	return service{repo, customerService, serviceProviderService, logger}
}

// Get returns the rating with the specified the rating ID.
func (s service) Get(ctx context.Context, id string) (Rating, error) {
	rating, err := s.repo.Get(ctx, id)
	if err != nil {
		return Rating{}, err
	}
	return Rating{rating}, nil
}

// Create creates a new rating.
func (s service) Create(ctx context.Context, req CreateRatingRequest) (Rating, error) {
	if err := req.Validate(); err != nil {
		return Rating{}, err
	}

	// Verify customer exists
	_, err := s.customerService.Get(ctx, req.CustomerID)
	if err != nil {
		return Rating{}, fmt.Errorf("customer validation error: %v", err)
	}

	// Verify service provider exists
	_, err = s.serviceProviderService.Get(ctx, req.ServiceProviderID)
	if err != nil {
		return Rating{}, fmt.Errorf("service provider validation error: %v", err)
	}

	id := entity.GenerateID()

	rating := entity.Rating{
		ID:                id,
		CustomerID:        req.CustomerID,
		ServiceProviderID: req.ServiceProviderID,
		RatingValue:       req.RatingValue,
		Comment:           req.Comment,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := s.repo.Create(ctx, rating); err != nil {
		return Rating{}, err
	}

	return Rating{rating}, nil
}

// Count returns the number of albums.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// GetAverageRatingByServiceProvider returns the average rating for a service provider.
func (s service) GetAverageRatingByServiceProvider(ctx context.Context, serviceProviderID string) (AverageRating, error) {
	// First verify that the service provider exists
	_, err := s.serviceProviderService.Get(ctx, serviceProviderID)
	if err != nil {
		return AverageRating{}, err
	}

	avgRating, totalCount, err := s.repo.GetAverageRatingByServiceProvider(ctx, serviceProviderID)
	if err != nil {
		return AverageRating{}, err
	}

	return AverageRating{
		ServiceProviderID: serviceProviderID,
		AverageRating:     avgRating,
		TotalRatings:      totalCount,
		LastUpdated:       time.Now(),
	}, nil
}
