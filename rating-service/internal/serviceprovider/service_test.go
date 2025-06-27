package serviceprovider

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/pkg/log"
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

	type args struct {
		action string
		id     string
		req    CreateServiceProviderRequest
	}
	tests := []struct {
		name      string
		args      args
		wantName  string
		wantErr   error
		wantCount int
	}{
		{
			name:      "initial count",
			args:      args{action: "count"},
			wantCount: 1,
		},
		{
			name:     "get existing",
			args:     args{action: "get", id: "123"},
			wantName: "serviceprovider123",
			wantErr:  nil,
		},
		{
			name:    "get not found",
			args:    args{action: "get", id: "404"},
			wantErr: sql.ErrNoRows,
		},
		{
			name:      "create valid",
			args:      args{action: "create", req: CreateServiceProviderRequest{Name: "test", Email: "test@example.com"}},
			wantName:  "test",
			wantErr:   nil,
			wantCount: 2,
		},
		{
			name:      "create validation error",
			args:      args{action: "create", req: CreateServiceProviderRequest{Name: "", Email: "test@example.com"}},
			wantErr:   errors.New("validation error"), // Will check for not nil
			wantCount: 2,
		},
		{
			name:      "create unexpected error",
			args:      args{action: "create", req: CreateServiceProviderRequest{Name: "error", Email: "test@example.com"}},
			wantErr:   errCRUD,
			wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.args.action {
			case "count":
				count, _ := s.Count(ctx)
				assert.Equal(t, tt.wantCount, count)
			case "get":
				serviceprovider, err := s.Get(ctx, tt.args.id)
				assert.Equal(t, tt.wantErr, err)
				if err == nil {
					assert.Equal(t, tt.wantName, serviceprovider.Name)
				}
			case "create":
				serviceprovider, err := s.Create(ctx, tt.args.req)
				switch tt.wantErr {
				case nil:
					assert.Nil(t, err)
					assert.Equal(t, tt.wantName, serviceprovider.Name)
				case errCRUD:
					assert.Equal(t, errCRUD, err)
				default:
					assert.NotNil(t, err)
				}
				count, _ := s.Count(ctx)
				assert.Equal(t, tt.wantCount, count)
			}
		})
	}
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
