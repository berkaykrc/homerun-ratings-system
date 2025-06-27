package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/notification-service/internal/config"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/internal/errors"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/circuitbreaker"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/log"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/retry"
	routing "github.com/go-ozzo/ozzo-routing/v2"
	"github.com/go-ozzo/ozzo-routing/v2/content"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationAPI_CreateNotification(t *testing.T) {
	tests := []struct {
		name           string
		request        interface{}
		requestBody    string
		expectedStatus int
		validateFunc   func(t *testing.T, body []byte)
	}{
		{
			name: "valid notification request",
			request: RatingNotificationRequest{
				ServiceProviderID: "123e4567-e89b-12d3-a456-426614174000",
				RatingID:          "456e7890-e89b-12d3-a456-426614174001",
				Rating:            5,
				CustomerName:      "John Doe",
				Comment:           "Excellent service!",
			},
			expectedStatus: http.StatusCreated,
			validateFunc: func(t *testing.T, body []byte) {
				var response CreateNotificationResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, "Notification created successfully", response.Message)
			},
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, body []byte) {
				// Just validate it returns bad request
			},
		},
		{
			name: "missing required fields",
			request: RatingNotificationRequest{
				Rating: 5,
				// Missing ServiceProviderID and RatingID
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response["message"], "problem with the data")
			},
		},
		{
			name: "invalid rating (too low)",
			request: RatingNotificationRequest{
				ServiceProviderID: "123e4567-e89b-12d3-a456-426614174000",
				RatingID:          "456e7890-e89b-12d3-a456-426614174001",
				Rating:            0,
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response["message"], "problem with the data")
			},
		},
		{
			name: "invalid rating (too high)",
			request: RatingNotificationRequest{
				ServiceProviderID: "123e4567-e89b-12d3-a456-426614174000",
				RatingID:          "456e7890-e89b-12d3-a456-426614174001",
				Rating:            6,
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response["message"], "problem with the data")
			},
		},
		{
			name: "invalid UUID format",
			request: RatingNotificationRequest{
				ServiceProviderID: "invalid-uuid",
				RatingID:          "456e7890-e89b-12d3-a456-426614174001",
				Rating:            5,
			},
			expectedStatus: http.StatusBadRequest,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response["message"], "problem with the data")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger, _ := log.NewForTest()
			storage := NewInMemoryStorage(logger)
			cfg := config.Config{
				ServerPort:     8081,
				Retry:          retry.DefaultRetryConfig(),
				CircuitBreaker: circuitbreaker.DefaultConfig(),
				Cleanup: config.CleanupConfig{
					Interval: 5 * time.Minute,
					MaxAge:   1 * time.Hour,
				},
			}
			service := NewService(storage, logger, cfg)

			// Create router
			router := routing.New()
			router.Use(
				errors.Handler(logger),
				content.TypeNegotiator(content.JSON),
			)
			RegisterHandlers(router, service, logger)

			// Prepare request body
			var requestBody []byte
			if tt.requestBody != "" {
				requestBody = []byte(tt.requestBody)
			} else {
				var err error
				requestBody, err = json.Marshal(tt.request)
				require.NoError(t, err)
			}

			// Create HTTP request
			httpReq := httptest.NewRequest("POST", "/api/internal/notifications", bytes.NewBuffer(requestBody))
			httpReq.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, httpReq)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Run custom validation
			if tt.validateFunc != nil {
				tt.validateFunc(t, w.Body.Bytes())
			}
		})
	}
}

func TestNotificationAPI_GetNotifications(t *testing.T) {
	tests := []struct {
		name               string
		serviceProviderID  string
		setupNotifications []Notification
		queryParams        string
		expectedStatus     int
		validateFunc       func(t *testing.T, body []byte)
	}{
		{
			name:              "get notifications successfully",
			serviceProviderID: "123e4567-e89b-12d3-a456-426614174000",
			setupNotifications: []Notification{
				{
					ID:                "test-notification-1",
					ServiceProviderID: "123e4567-e89b-12d3-a456-426614174000",
					Message:           "New 5-star rating received",
					RatingID:          "456e7890-e89b-12d3-a456-426614174001",
					CreatedAt:         time.Now(),
				},
				{
					ID:                "test-notification-2",
					ServiceProviderID: "123e4567-e89b-12d3-a456-426614174000",
					Message:           "New 4-star rating received",
					RatingID:          "456e7890-e89b-12d3-a456-426614174002",
					CreatedAt:         time.Now().Add(time.Minute),
				},
			},
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, body []byte) {
				var response GetNotificationsResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Len(t, response.Notifications, 2)
				assert.False(t, response.HasMore)
			},
		},
		{
			name:              "get notifications with lastChecked filter",
			serviceProviderID: "123e4567-e89b-12d3-a456-426614174000",
			setupNotifications: []Notification{
				{
					ID:                "old-notification",
					ServiceProviderID: "123e4567-e89b-12d3-a456-426614174000",
					Message:           "Old notification",
					RatingID:          "456e7890-e89b-12d3-a456-426614174001",
					CreatedAt:         time.Now().Add(-time.Hour),
				},
				{
					ID:                "new-notification",
					ServiceProviderID: "123e4567-e89b-12d3-a456-426614174000",
					Message:           "New notification",
					RatingID:          "456e7890-e89b-12d3-a456-426614174002",
					CreatedAt:         time.Now(),
				},
			},
			queryParams:    "?lastChecked=" + url.QueryEscape(time.Now().Add(-30*time.Minute).Format(time.RFC3339)),
			expectedStatus: http.StatusOK,
			validateFunc: func(t *testing.T, body []byte) {
				var response GetNotificationsResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Len(t, response.Notifications, 1)
				assert.Equal(t, "new-notification", response.Notifications[0].ID)
			},
		},
		{
			name:               "no notifications found",
			serviceProviderID:  "999e4567-e89b-12d3-a456-426614174999",
			setupNotifications: []Notification{},
			expectedStatus:     http.StatusOK,
			validateFunc: func(t *testing.T, body []byte) {
				var response GetNotificationsResponse
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Len(t, response.Notifications, 0)
				assert.False(t, response.HasMore)
			},
		},
		{
			name:              "invalid lastChecked format",
			serviceProviderID: "123e4567-e89b-12d3-a456-426614174000",
			queryParams:       "?lastChecked=invalid-date",
			expectedStatus:    http.StatusBadRequest,
			validateFunc: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response["message"], "Invalid lastChecked format")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			logger, _ := log.NewForTest()
			storage := NewInMemoryStorage(logger)
			cfg := config.Config{
				ServerPort:     8081,
				Retry:          retry.DefaultRetryConfig(),
				CircuitBreaker: circuitbreaker.DefaultConfig(),
				Cleanup: config.CleanupConfig{
					Interval: 5 * time.Minute,
					MaxAge:   1 * time.Hour,
				},
			}
			service := NewService(storage, logger, cfg)

			// Create router
			router := routing.New()
			router.Use(
				errors.Handler(logger),
				content.TypeNegotiator(content.JSON),
			)
			RegisterHandlers(router, service, logger)

			// Setup test notifications
			for _, notification := range tt.setupNotifications {
				err := storage.StoreNotification(context.Background(), notification)
				require.NoError(t, err)
			}

			// Create HTTP request
			url := "/api/notifications/" + tt.serviceProviderID + tt.queryParams
			httpReq := httptest.NewRequest("GET", url, nil)

			// Create response recorder
			w := httptest.NewRecorder()

			// Execute request
			router.ServeHTTP(w, httpReq)

			// Assert response
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Run custom validation
			if tt.validateFunc != nil {
				tt.validateFunc(t, w.Body.Bytes())
			}
		})
	}
}
