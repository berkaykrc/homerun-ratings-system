package serviceprovider

import (
	"context"

	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/pkg/dbcontext"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/pkg/log"
)

// Repository encapsulates the logic to access service providers from the data source.
type Repository interface {
	// Get returns the service provider with the specified service provider ID.
	Get(ctx context.Context, id string) (entity.ServiceProvider, error)
	// Count returns the total number of service providers in the storage.
	Count(ctx context.Context) (int, error)
	// Create saves a new service provider in the storage.
	Create(ctx context.Context, serviceProvider entity.ServiceProvider) error
}

// repository persists service providers in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new service provider repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the service provider with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string) (entity.ServiceProvider, error) {
	var serviceProvider entity.ServiceProvider
	err := r.db.With(ctx).Select().Model(id, &serviceProvider)
	return serviceProvider, err
}

// Create saves a new service provider record in the database.
func (r repository) Create(ctx context.Context, serviceProvider entity.ServiceProvider) error {
	return r.db.With(ctx).Model(&serviceProvider).Insert()
}

// Count returns the number of the service provider records in the database.
func (r repository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.With(ctx).Select("COUNT(*)").From("service_providers").Row(&count)
	return count, err
}
