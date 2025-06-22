package rating

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/customer"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/serviceprovider"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/test"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	logger, _ := log.NewForTest()
	db := test.DB(t)
	test.ResetTables(t, db, "ratings", "customers", "service_providers")

	// Create repositories
	repo := NewRepository(db, logger)
	customerRepo := customer.NewRepository(db, logger)
	serviceProviderRepo := serviceprovider.NewRepository(db, logger)

	ctx := context.Background()

	// Create required parent records
	err := customerRepo.Create(ctx, entity.Customer{
		ID:        "customer1",
		Name:      "Test Customer",
		Email:     "customer1@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	assert.Nil(t, err)

	err = serviceProviderRepo.Create(ctx, entity.ServiceProvider{
		ID:        "service1",
		Name:      "Test Service Provider",
		Email:     "service1@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	assert.Nil(t, err)

	// count
	count, err := repo.Count(ctx)
	assert.Nil(t, err)
	// create
	err = repo.Create(ctx, entity.Rating{
		ID:                "test1",
		CustomerID:        "customer1",
		ServiceProviderID: "service1",
		RatingValue:       5,
		Comment:           "Great service!",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	})
	assert.Nil(t, err)
	count2, err := repo.Count(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, count2-count)

	// get
	rating, err := repo.Get(ctx, "test1")
	assert.Nil(t, err)
	assert.Equal(t, "customer1", rating.CustomerID)
	assert.Equal(t, "service1", rating.ServiceProviderID)
	assert.Equal(t, 5, rating.RatingValue)
	assert.Equal(t, "Great service!", rating.Comment)
	assert.Nil(t, err)
	_, err = repo.Get(ctx, "test0")
	assert.Equal(t, sql.ErrNoRows, err)

	// Test average rating calculation
	// Add more ratings for the same service provider
	err = repo.Create(ctx, entity.Rating{
		ID:                "test2",
		CustomerID:        "customer1",
		ServiceProviderID: "service1",
		RatingValue:       3,
		Comment:           "Good service",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	})
	assert.Nil(t, err)

	err = repo.Create(ctx, entity.Rating{
		ID:                "test3",
		CustomerID:        "customer1",
		ServiceProviderID: "service1",
		RatingValue:       4,
		Comment:           "Very good service",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	})
	assert.Nil(t, err)

	// Test average rating: (5 + 3 + 4) / 3 = 4.0
	avgRating, totalCount, err := repo.GetAverageRatingByServiceProvider(ctx, "service1")
	assert.Nil(t, err)
	assert.Equal(t, 3, totalCount)
	assert.InDelta(t, 4.0, avgRating, 0.01) // Use InDelta for float comparison

	// Test average rating for non-existent service provider
	avgRating, totalCount, err = repo.GetAverageRatingByServiceProvider(ctx, "nonexistent")
	assert.Nil(t, err)
	assert.Equal(t, 0, totalCount)
	assert.Equal(t, 0.0, avgRating)
}
