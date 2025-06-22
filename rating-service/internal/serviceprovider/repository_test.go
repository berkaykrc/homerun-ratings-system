package serviceprovider

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/test"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	logger, _ := log.NewForTest()
	db := test.DB(t)
	test.ResetTables(t, db, "service_providers")
	repo := NewRepository(db, logger)

	ctx := context.Background()
	// count
	count, err := repo.Count(ctx)
	assert.Nil(t, err)
	// create
	err = repo.Create(ctx, entity.ServiceProvider{
		ID:        "test1",
		Name:      "serviceprovider1",
		Email:     "serviceprovider1@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	assert.Nil(t, err)
	count2, err := repo.Count(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, count2-count)

	// get
	serviceprovider, err := repo.Get(ctx, "test1")
	assert.Nil(t, err)
	assert.Equal(t, "serviceprovider1", serviceprovider.Name)
	_, err = repo.Get(ctx, "test0")
	assert.Equal(t, sql.ErrNoRows, err)

}
