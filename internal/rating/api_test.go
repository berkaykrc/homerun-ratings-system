package rating

import (
	"net/http"
	"testing"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/internal/customer"
	"github.com/berkaykrc/homerun-ratings-system/internal/entity"
	"github.com/berkaykrc/homerun-ratings-system/internal/serviceprovider"
	"github.com/berkaykrc/homerun-ratings-system/internal/test"
	"github.com/berkaykrc/homerun-ratings-system/pkg/log"
)

func TestAPI(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	customerService := customer.NewService(&mockCustomerRepository{}, logger)
	serviceProviderService := serviceprovider.NewService(&mockServiceProviderRepository{}, logger)
	repo := &mockRepository{items: []entity.Rating{
		{ID: "123", CustomerID: "customer123", ServiceProviderID: "service123", RatingValue: 5, Comment: "Great service!", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}}
	RegisterHandlers(router.Group(""), NewService(repo, customerService, serviceProviderService, logger), logger)

	tests := []test.APITestCase{
		{Name: "get 123", Method: "GET", URL: "/ratings/123", Body: "", WantStatus: http.StatusOK, WantResponse: `*123*`},
		{Name: "get unknown", Method: "GET", URL: "/ratings/1234", Body: "", WantStatus: http.StatusNotFound, WantResponse: ""},
		{Name: "create ok", Method: "POST", URL: "/ratings", Body: `{"customerId":"customer123", "serviceProviderId":"service123", "rating":5, "comment":"Great service!"}`, WantStatus: http.StatusCreated, WantResponse: "**"},
		{Name: "create input error", Method: "POST", URL: "/ratings", Body: `"name":"test"}`, WantStatus: http.StatusBadRequest, WantResponse: ""},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
