package customer

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/internal/test"
	"github.com/berkaykrc/homerun-ratings-system/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestRepository(t *testing.T) {
	logger, _ := log.NewForTest()
	db := test.DB(t)
	test.ResetTables(t, db, "customers")
	repo := NewRepository(db, logger)

	ctx := context.Background()
	// count
	count, err := repo.Count(ctx)
	assert.Nil(t, err)
	// create
	err = repo.Create(ctx, entity.Customer{
		ID:        "test1",
		Name:      "customer1",
		Email:     "customer1@example.com",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	assert.Nil(t, err)
	count2, err := repo.Count(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, count2-count)

	// get
	customer, err := repo.Get(ctx, "test1")
	assert.Nil(t, err)
	assert.Equal(t, "customer1", customer.Name)
	_, err = repo.Get(ctx, "test0")
	assert.Equal(t, sql.ErrNoRows, err)

}
