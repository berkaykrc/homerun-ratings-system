package customer

import (
	"net/http"
	"testing"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/internal/test"
	"github.com/berkaykrc/homerun-ratings-system/pkg/log"
)

func TestAPI(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	repo := &mockRepository{items: []entity.Customer{
		{ID: "123", Name: "customer123", Email: "customer123@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}}
	RegisterHandlers(router.Group(""), NewService(repo, logger), logger)

	tests := []test.APITestCase{
		{Name: "get 123", Method: "GET", URL: "/customers/123", Body: "", WantStatus: http.StatusOK, WantResponse: `*customer123*`},
		{Name: "get unknown", Method: "GET", URL: "/customers/1234", Body: "", WantStatus: http.StatusNotFound, WantResponse: ""},
		{Name: "create ok", Method: "POST", URL: "/customers", Body: `{"name":"test", "email":"test@example.com"}`, WantStatus: http.StatusCreated, WantResponse: "*test*"},
		{Name: "create input error", Method: "POST", URL: "/customers", Body: `"name":"test"}`, WantStatus: http.StatusBadRequest, WantResponse: ""},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
