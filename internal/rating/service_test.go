package rating

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/berkaykrc/homerun-ratings-system/internal/customer"
	"github.com/berkaykrc/homerun-ratings-system/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/internal/serviceprovider"
	"github.com/berkaykrc/homerun-ratings-system/pkg/log"
	"github.com/stretchr/testify/assert"
)

var errCRUD = errors.New("crud error")

func TestCreateRatingRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateRatingRequest
		wantError bool
	}{
		{"success", CreateRatingRequest{CustomerID: "customer123", ServiceProviderID: "service123", RatingValue: 5, Comment: "good service"}, false},
		{"required", CreateRatingRequest{CustomerID: "", ServiceProviderID: "service123", RatingValue: 5, Comment: "good service"}, true},
		{"too long", CreateRatingRequest{CustomerID: "customer123", ServiceProviderID: "service123", RatingValue: 5, Comment: "a very long comment that exceeds the maximum length of 100 characters.........................................................................................................................."}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.Validate()
			assert.Equal(t, tt.wantError, err != nil)
		})
	}
}

func Test_service_CRUD(t *testing.T) {
	logger, _ := log.NewForTest()

	// Create actual services with mock repositories
	customerService := customer.NewService(&mockCustomerRepository{}, logger)
	serviceProviderService := serviceprovider.NewService(&mockServiceProviderRepository{}, logger)

	s := NewService(&mockRepository{}, customerService, serviceProviderService, logger)

	ctx := context.Background()
	// initial count
	count, _ := s.Count(ctx)
	assert.Equal(t, 0, count)

	// successful creation
	rating, err := s.Create(ctx, CreateRatingRequest{CustomerID: "customer123", ServiceProviderID: "service123", RatingValue: 5, Comment: "good service"})
	assert.Nil(t, err)
	assert.NotEmpty(t, rating.ID)
	id := rating.ID
	assert.Equal(t, "customer123", rating.CustomerID)
	assert.Equal(t, "service123", rating.ServiceProviderID)
	assert.Equal(t, 5, rating.RatingValue)
	assert.Equal(t, "good service", rating.Comment)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// validation error in creation
	_, err = s.Create(ctx, CreateRatingRequest{CustomerID: "", ServiceProviderID: "service123", RatingValue: 5, Comment: "good service"})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// unexpected error in creation
	_, err = s.Create(ctx, CreateRatingRequest{CustomerID: "customer123", ServiceProviderID: "service123", RatingValue: 5, Comment: "error"})
	assert.Equal(t, errCRUD, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	_, _ = s.Create(ctx, CreateRatingRequest{CustomerID: "customer123", ServiceProviderID: "service123", RatingValue: 5, Comment: "good service"})

	// get
	_, err = s.Get(ctx, "none")
	assert.NotNil(t, err)
	rating, err = s.Get(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, "customer123", rating.CustomerID)
	assert.Equal(t, id, rating.ID)

}

type mockRepository struct {
	items []entity.Rating
}

func (m mockRepository) Get(ctx context.Context, id string) (entity.Rating, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return entity.Rating{}, sql.ErrNoRows
}

func (m *mockRepository) Create(ctx context.Context, rating entity.Rating) error {
	if rating.Comment == "error" {
		return errCRUD
	}
	m.items = append(m.items, rating)
	return nil
}

func (m mockRepository) Count(ctx context.Context) (int, error) {
	return len(m.items), nil
}

type mockCustomerRepository struct{}

func (m *mockCustomerRepository) Get(ctx context.Context, id string) (entity.Customer, error) {
	if id == "nonexistent" {
		return entity.Customer{}, sql.ErrNoRows
	}
	return entity.Customer{
		ID:    id,
		Name:  "Test Customer",
		Email: "test@customer.com",
	}, nil
}

func (m *mockCustomerRepository) Create(ctx context.Context, customer entity.Customer) error {
	return nil
}

func (m *mockCustomerRepository) Count(ctx context.Context) (int, error) {
	return 0, nil
}

type mockServiceProviderRepository struct{}

func (m *mockServiceProviderRepository) Get(ctx context.Context, id string) (entity.ServiceProvider, error) {
	if id == "nonexistent" {
		return entity.ServiceProvider{}, sql.ErrNoRows
	}
	return entity.ServiceProvider{
		ID:    id,
		Name:  "Test Service Provider",
		Email: "test@provider.com",
	}, nil
}

func (m *mockServiceProviderRepository) Create(ctx context.Context, serviceProvider entity.ServiceProvider) error {
	return nil
}

func (m *mockServiceProviderRepository) Count(ctx context.Context) (int, error) {
	return 0, nil
}
