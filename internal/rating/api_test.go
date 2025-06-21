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
	repo := &mockRepository{items: []entity.Rating{
		{ID: "123", CustomerID: "customer123", ServiceProviderID: "service123", RatingValue: 5, Comment: "Great service!", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}}
	serviceProviderService := serviceprovider.NewService(&mockServiceProviderRepository{}, logger)
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

func TestAPI_AverageRating(t *testing.T) {
	logger, _ := log.NewForTest()
	router := test.MockRouter(logger)
	customerService := customer.NewService(&mockCustomerRepository{}, logger)
	repo := &mockRepository{items: []entity.Rating{
		{ID: "1", CustomerID: "customer1", ServiceProviderID: "provider1", RatingValue: 5, Comment: "Excellent"},
		{ID: "2", CustomerID: "customer2", ServiceProviderID: "provider1", RatingValue: 4, Comment: "Good"},
		{ID: "3", CustomerID: "customer3", ServiceProviderID: "provider1", RatingValue: 3, Comment: "Average"},
	}}
	serviceProviderService := serviceprovider.NewService(&mockServiceProviderRepository{}, logger)
	RegisterHandlers(router.Group(""), NewService(repo, customerService, serviceProviderService, logger), logger)

	tests := []test.APITestCase{
		{Name: "get average rating for provider1", Method: "GET", URL: "/service-providers/provider1/average-rating", Body: "", WantStatus: http.StatusOK, WantResponse: `*"averageRating":4*`},
		{Name: "get average rating for nonexistent", Method: "GET", URL: "/service-providers/nonexistent/average-rating", Body: "", WantStatus: http.StatusNotFound, WantResponse: ""},
	}
	for _, tc := range tests {
		test.Endpoint(t, router, tc)
	}
}
