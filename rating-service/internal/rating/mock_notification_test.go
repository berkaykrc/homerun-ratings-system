package rating

import (
	"context"

	"github.com/berkaykrc/homerun-ratings-system/rating-service/internal/notification"
)

// MockNotificationClient is a shared mock for testing
type MockNotificationClient struct {
	// Allow tests to control behavior
	ShouldReturnError bool
	LastNotification  *notification.RatingNotification
	CallCount         int
}

func NewMockNotificationClient() *MockNotificationClient {
	return &MockNotificationClient{}
}

func (m *MockNotificationClient) SendRatingNotification(ctx context.Context, notification notification.RatingNotification) error {
	m.CallCount++
	m.LastNotification = &notification

	if m.ShouldReturnError {
		return &MockNotificationError{Message: "mock notification error"}
	}

	return nil
}

// MockNotificationError implements error interface
type MockNotificationError struct {
	Message string
}

func (e *MockNotificationError) Error() string {
	return e.Message
}

// Reset resets the mock state for test isolation
func (m *MockNotificationClient) Reset() {
	m.ShouldReturnError = false
	m.LastNotification = nil
	m.CallCount = 0
}
