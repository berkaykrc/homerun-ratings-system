package rating

import (
	"context"

	"github.com/berkaykrc/homerun-ratings-system/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/pkg/dbcontext"
	"github.com/berkaykrc/homerun-ratings-system/pkg/log"
)

// Repository encapsulates the logic to access ratings from the data source.
type Repository interface {
	// Get returns the rating with the specified rating ID.
	Get(ctx context.Context, id string) (entity.Rating, error)
	// Count returns the total number of ratings in the storage.
	Count(ctx context.Context) (int, error)
	// Create saves a new rating in the storage.
	Create(ctx context.Context, rating entity.Rating) error
}

// repository persists ratings in database
type repository struct {
	db     *dbcontext.DB
	logger log.Logger
}

// NewRepository creates a new rating repository
func NewRepository(db *dbcontext.DB, logger log.Logger) Repository {
	return repository{db, logger}
}

// Get reads the rating with the specified ID from the database.
func (r repository) Get(ctx context.Context, id string) (entity.Rating, error) {
	var rating entity.Rating
	err := r.db.With(ctx).Select().Model(id, &rating)
	return rating, err
}

// Create saves a new rating record in the database.
// It returns the ID of the newly inserted rating record.
func (r repository) Create(ctx context.Context, rating entity.Rating) error {
	return r.db.With(ctx).Model(&rating).Insert()
}

// Count returns the number of the rating records in the database.
func (r repository) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.With(ctx).Select("COUNT(*)").From("ratings").Row(&count)
	return count, err
}
