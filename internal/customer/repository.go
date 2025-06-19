package customer

import (
	"context"

	"github.com/berkaykrc/homerun-ratings-system/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/pkg/dbcontext"
	"github.com/berkaykrc/homerun-ratings-system/pkg/log"
)

// Repository encapsulates the logic to access customers from the data source.
type Repository interface {
	// Get returns the customer with the specified customer ID.
	Get(ctx context.Context, id string) (entity.Customer, error)
	// Count returns the total number of customers in the storage.
	Count(ctx context.Context) (int, error)
	// Create saves a new customer in the storage.
	Create(ctx context.Context, customer entity.Customer) error
}

// repository persists customers in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new customer repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the customer with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string) (entity.Customer, error) {
	var customer entity.Customer
	err := r.db.With(ctx).Select().Model(id, &customer)
	return customer, err
}

// Create saves a new customer record in the database.
func (r repository) Create(ctx context.Context, customer entity.Customer) error {
	return r.db.With(ctx).Model(&customer).Insert()
}

// Count returns the number of the customer records in the database.
func (r repository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.With(ctx).Select("COUNT(*)").From("customers").Row(&count)
	return count, err
}
