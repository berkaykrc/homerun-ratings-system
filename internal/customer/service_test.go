package customer

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/berkaykrc/homerun-ratings-system/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/pkg/log"
	"github.com/stretchr/testify/assert"
)

var errCRUD = errors.New("crud error")

func TestCreateCustomerRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateCustomerRequest
		wantError bool
	}{
		{"success", CreateCustomerRequest{Name: "test", Email: "test@example.com"}, false},
		{"required", CreateCustomerRequest{Name: "", Email: "test@example.com"}, true},
		{"too long", CreateCustomerRequest{Name: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890", Email: "test@example.com"}, true},
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
	s := NewService(&mockRepository{}, logger)

	ctx := context.Background()
	// initial count
	count, _ := s.Count(ctx)
	assert.Equal(t, 0, count)

	// successful creation
	customer, err := s.Create(ctx, CreateCustomerRequest{Name: "test", Email: "test@example.com"})
	assert.Nil(t, err)
	assert.NotEmpty(t, customer.ID)
	id := customer.ID
	assert.Equal(t, "test", customer.Name)
	assert.NotEmpty(t, customer.CreatedAt)
	assert.NotEmpty(t, customer.UpdatedAt)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// validation error in creation
	_, err = s.Create(ctx, CreateCustomerRequest{Name: "", Email: "test@example.com"})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	// unexpected error in creation
	_, err = s.Create(ctx, CreateCustomerRequest{Name: "error", Email: "error@example.com"})
	assert.Equal(t, errCRUD, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 1, count)

	_, _ = s.Create(ctx, CreateCustomerRequest{Name: "test2", Email: "test2@example.com"})

	// get
	_, err = s.Get(ctx, "none")
	assert.NotNil(t, err)
	customer, err = s.Get(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, "test", customer.Name)
	assert.Equal(t, id, customer.ID)

}

type mockRepository struct {
	items []entity.Customer
}

func (m mockRepository) Get(ctx context.Context, id string) (entity.Customer, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return entity.Customer{}, sql.ErrNoRows
}

func (m *mockRepository) Create(ctx context.Context, customer entity.Customer) error {
	if customer.Name == "error" {
		return errCRUD
	}
	m.items = append(m.items, customer)
	return nil
}

func (m mockRepository) Count(ctx context.Context) (int, error) {
	return len(m.items), nil
}
