package rating

import (
	"context"
	"database/sql"

	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/pkg/dbcontext"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/pkg/log"
	dbx "github.com/go-ozzo/ozzo-dbx"
)

// Repository encapsulates the logic to access ratings from the data source.
type Repository interface {
	// Get returns the rating with the specified rating ID.
	Get(ctx context.Context, id string) (entity.Rating, error)
	// Count returns the total number of ratings in the storage.
	Count(ctx context.Context) (int, error)
	// Create saves a new rating in the storage.
	Create(ctx context.Context, rating entity.Rating) error
	// GetAverageRatingByServiceProvider returns the average rating and total count for a service provider.
	GetAverageRatingByServiceProvider(ctx context.Context, serviceProviderID string) (float64, int, error)
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

// GetAverageRatingByServiceProvider returns the average rating and total count for a service provider.
func (r repository) GetAverageRatingByServiceProvider(ctx context.Context, serviceProviderID string) (float64, int, error) {
	var avgRating sql.NullFloat64
	var totalCount int

	err := r.db.With(ctx).Select("AVG(rating_value) as avg_rating", "COUNT(*) as total_count").
		From("ratings").
		Where(dbx.HashExp{"service_provider_id": serviceProviderID}).
		Row(&avgRating, &totalCount)
	if err != nil {
		return 0, 0, err
	}

	if !avgRating.Valid || totalCount == 0 {
		return 0, 0, nil
	}

	return avgRating.Float64, totalCount, nil
}
