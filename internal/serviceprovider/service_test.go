package serviceprovider

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

func TestCreateServiceProviderRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		model     CreateServiceProviderRequest
		wantError bool
	}{
		{"success", CreateServiceProviderRequest{Name: "test", Email: "test@example.com"}, false},
		{"required", CreateServiceProviderRequest{Name: "", Email: "test@example.com"}, true},
		{"too long", CreateServiceProviderRequest{Name: "1234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890", Email: "test@example.com"}, true},
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
	s := NewService(&mockRepository{items: []entity.ServiceProvider{
		{ID: "123", Name: "serviceprovider123", Email: "serviceprovider123@example.com"},
	}}, logger)

	ctx := context.Background()
	// initial count
	count, _ := s.Count(ctx)
	assert.Equal(t, 1, count)

	// get
	serviceprovider, err := s.Get(ctx, "123")
	assert.Nil(t, err)
	assert.Equal(t, "serviceprovider123", serviceprovider.Name)
	_, err = s.Get(ctx, "404")
	assert.Equal(t, sql.ErrNoRows, err)

	// create
	serviceprovider, err = s.Create(ctx, CreateServiceProviderRequest{Name: "test", Email: "test@example.com"})
	assert.Nil(t, err)
	assert.Equal(t, "test", serviceprovider.Name)
	count, _ = s.Count(ctx)
	assert.Equal(t, 2, count)

	// validation error in creation
	_, err = s.Create(ctx, CreateServiceProviderRequest{Name: "", Email: "test@example.com"})
	assert.NotNil(t, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 2, count)

	// unexpected error in creation
	_, err = s.Create(ctx, CreateServiceProviderRequest{Name: "error", Email: "test@example.com"})
	assert.Equal(t, errCRUD, err)
	count, _ = s.Count(ctx)
	assert.Equal(t, 2, count)
}

type mockRepository struct {
	items []entity.ServiceProvider
}

func (m mockRepository) Get(ctx context.Context, id string) (entity.ServiceProvider, error) {
	for _, item := range m.items {
		if item.ID == id {
			return item, nil
		}
	}
	return entity.ServiceProvider{}, sql.ErrNoRows
}

func (m *mockRepository) Create(ctx context.Context, serviceprovider entity.ServiceProvider) error {
	if serviceprovider.Name == "error" {
		return errCRUD
	}
	m.items = append(m.items, serviceprovider)
	return nil
}

func (m mockRepository) Count(ctx context.Context) (int, error) {
	return len(m.items), nil
}
