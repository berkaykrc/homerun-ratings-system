package rating

import (
	"context"
	"fmt"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/customer"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/notification"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/serviceprovider"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/pkg/log"
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
	notificationClient     notification.Client
	logger                 log.Logger
}

// NewService creates a new rating service.
func NewService(repo Repository, customerService customer.Service, serviceProviderService serviceprovider.Service, notificationClient notification.Client, logger log.Logger) Service {
	return service{repo, customerService, serviceProviderService, notificationClient, logger}
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

	customer, err := s.customerService.Get(ctx, req.CustomerID)
	if err != nil {
		return Rating{}, fmt.Errorf("customer validation error: %v", err)
	}

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

	notification := notification.RatingNotification{
		ServiceProviderID: req.ServiceProviderID,
		RatingID:          id,
		Rating:            req.RatingValue,
		CustomerName:      customer.Name,
		Comment:           req.Comment,
	}

	go func() {
		// Create a new context for the notification to avoid cancellation
		// when the original request context is done
		notificationCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.notificationClient.SendRatingNotification(notificationCtx, notification); err != nil {
			s.logger.With(notificationCtx, "error", err, "rating_id", id).Error("Failed to send rating notification")
		}
	}()

	return Rating{rating}, nil
}

// Count returns the number of albums.
func (s service) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

// GetAverageRatingByServiceProvider returns the average rating for a service provider.
func (s service) GetAverageRatingByServiceProvider(ctx context.Context, serviceProviderID string) (AverageRating, error) {
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
