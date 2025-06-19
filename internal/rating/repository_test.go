package rating

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/internal/customer"
	"github.com/berkaykrc/homerun-ratings-system/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/internal/serviceprovider"
	"github.com/berkaykrc/homerun-ratings-system/internal/test"
	"github.com/berkaykrc/homerun-ratings-system/pkg/log"
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
}
