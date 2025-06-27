package notification

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/notification-service/internal/config"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/circuitbreaker"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/log"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/retry"
	"github.com/stretchr/testify/assert"
)

var errStorage = errors.New("storage error")

// mockStorage implements Storage interface for testing
type mockStorage struct {
	notifications []Notification
	storeError    error
	getError      error
	cleanupError  error
}

func (m *mockStorage) StoreNotification(ctx context.Context, notification Notification) error {
	if m.storeError != nil {
		return m.storeError
	}
	m.notifications = append(m.notifications, notification)
	return nil
}

func (m *mockStorage) GetNotifications(ctx context.Context, serviceProviderID string, lastChecked time.Time) ([]Notification, error) {
	if m.getError != nil {
		return nil, m.getError
	}

	var result []Notification
	for _, n := range m.notifications {
		if n.ServiceProviderID == serviceProviderID && n.CreatedAt.After(lastChecked) {
			result = append(result, n)
		}
	}
	return result, nil
}

func (m *mockStorage) Cleanup(ctx context.Context, maxAge time.Duration) error {
	if m.cleanupError != nil {
		return m.cleanupError
	}
	// Simulate cleanup
	return nil
}

func TestService_CreateNotification(t *testing.T) {
	logger, _ := log.NewForTest()
	cfg := config.Config{
		Retry:          retry.DefaultRetryConfig(),
		CircuitBreaker: circuitbreaker.DefaultConfig(),
		Cleanup: config.CleanupConfig{
			Interval: 5 * time.Minute,
			MaxAge:   1 * time.Hour,
		},
	}

	tests := []struct {
		name        string
		storage     *mockStorage
		request     RatingNotificationRequest
		expectError bool
	}{
		{
			name:    "successful creation",
			storage: &mockStorage{},
			request: RatingNotificationRequest{
				ServiceProviderID: "123e4567-e89b-12d3-a456-426614174000",
				RatingID:          "456e7890-e89b-12d3-a456-426614174001",
				Rating:            5,
				CustomerName:      "John Doe",
				Comment:           "Excellent service",
			},
			expectError: false,
		},
		{
			name: "storage error",
			storage: &mockStorage{
				storeError: errStorage,
			},
			request: RatingNotificationRequest{
				ServiceProviderID: "123e4567-e89b-12d3-a456-426614174000",
				RatingID:          "456e7890-e89b-12d3-a456-426614174001",
				Rating:            5,
				CustomerName:      "John Doe",
				Comment:           "Excellent service",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.storage, logger, cfg)
			ctx := context.Background()

			response, err := service.CreateNotification(ctx, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotEmpty(t, response.ID)
				assert.Equal(t, "Notification created successfully", response.Message)
			}
		})
	}
}

func TestService_GetNotifications(t *testing.T) {
	logger, _ := log.NewForTest()
	cfg := config.Config{
		Retry:          retry.DefaultRetryConfig(),
		CircuitBreaker: circuitbreaker.DefaultConfig(),
		Cleanup: config.CleanupConfig{
			Interval: 5 * time.Minute,
			MaxAge:   1 * time.Hour,
		},
	}

	serviceProviderID := "123e4567-e89b-12d3-a456-426614174000"
	pastTime := time.Now().Add(-time.Hour)
	futureTime := time.Now().Add(time.Hour)

	notifications := []Notification{
		{
			ID:                "1",
			ServiceProviderID: serviceProviderID,
			Message:           "New 5-star rating received",
			RatingID:          "rating1",
			CreatedAt:         time.Now(),
		},
		{
			ID:                "2",
			ServiceProviderID: serviceProviderID,
			Message:           "New 4-star rating received",
			RatingID:          "rating2",
			CreatedAt:         time.Now().Add(time.Minute),
		},
	}

	tests := []struct {
		name              string
		storage           *mockStorage
		serviceProviderID string
		lastChecked       time.Time
		expectedCount     int
		expectError       bool
	}{
		{
			name: "successful retrieval",
			storage: &mockStorage{
				notifications: notifications,
			},
			serviceProviderID: serviceProviderID,
			lastChecked:       pastTime,
			expectedCount:     2,
			expectError:       false,
		},
		{
			name: "no notifications after lastChecked",
			storage: &mockStorage{
				notifications: notifications,
			},
			serviceProviderID: serviceProviderID,
			lastChecked:       futureTime,
			expectedCount:     0,
			expectError:       false,
		},
		{
			name: "storage error",
			storage: &mockStorage{
				getError: errStorage,
			},
			serviceProviderID: serviceProviderID,
			lastChecked:       pastTime,
			expectedCount:     0,
			expectError:       true,
		},
		{
			name: "different service provider",
			storage: &mockStorage{
				notifications: notifications,
			},
			serviceProviderID: "different-service-provider",
			lastChecked:       pastTime,
			expectedCount:     0,
			expectError:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(tt.storage, logger, cfg)
			ctx := context.Background()

			response, err := service.GetNotifications(ctx, tt.serviceProviderID, tt.lastChecked)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Len(t, response.Notifications, tt.expectedCount)
				assert.False(t, response.HasMore)
			}
		})
	}
}

func TestService_StartCleanupWorker(t *testing.T) {
	logger, _ := log.NewForTest()
	cfg := config.Config{
		Retry:          retry.DefaultRetryConfig(),
		CircuitBreaker: circuitbreaker.DefaultConfig(),
		Cleanup: config.CleanupConfig{
			Interval: 10 * time.Millisecond, // Short interval for testing
			MaxAge:   1 * time.Hour,
		},
	}

	storage := &mockStorage{}
	service := NewService(storage, logger, cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// This should run and exit gracefully when context is cancelled
	service.StartCleanupWorker(ctx)

	// If we reach here, the cleanup worker exited properly
	assert.True(t, true, "Cleanup worker should exit when context is cancelled")
}

func TestService_isRetryableError(t *testing.T) {
	logger, _ := log.NewForTest()
	cfg := config.Config{
		Retry:          retry.DefaultRetryConfig(),
		CircuitBreaker: circuitbreaker.DefaultConfig(),
		Cleanup: config.CleanupConfig{
			Interval: 5 * time.Minute,
			MaxAge:   1 * time.Hour,
		},
	}

	service := NewService(&mockStorage{}, logger, cfg).(*service)

	// Test that all errors are considered retryable (current implementation)
	assert.True(t, service.isRetryableError(errors.New("some error")))
	assert.True(t, service.isRetryableError(errStorage))
}
