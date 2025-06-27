package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNotification(t *testing.T) {
	req := RatingNotificationRequest{
		ServiceProviderID: "provider-123",
		RatingID:          "rating-456",
		Rating:            5,
		CustomerName:      "John Doe",
		Comment:           "Excellent service!",
	}

	notification := NewNotification(req)

	assert.NotEmpty(t, notification.ID)
	assert.Equal(t, req.ServiceProviderID, notification.ServiceProviderID)
	assert.Equal(t, req.RatingID, notification.RatingID)
	assert.NotEmpty(t, notification.Message)
	assert.Contains(t, notification.Message, "New 5-star rating received")
	assert.Contains(t, notification.Message, "John Doe")
	assert.Contains(t, notification.Message, "Excellent service!")
	assert.False(t, notification.CreatedAt.IsZero())
}

func TestFormatNotificationMessage(t *testing.T) {
	tests := []struct {
		name         string
		rating       int
		customerName string
		comment      string
		expected     string
	}{
		{
			name:         "5-star with customer and comment",
			rating:       5,
			customerName: "John Doe",
			comment:      "Great service!",
			expected:     "New 5-star rating received from John Doe: \"Great service!\"",
		},
		{
			name:         "1-star with customer only",
			rating:       1,
			customerName: "Jane Smith",
			comment:      "",
			expected:     "New 1-star rating received from Jane Smith",
		},
		{
			name:         "3-star with comment only",
			rating:       3,
			customerName: "",
			comment:      "Average service",
			expected:     "New 3-star rating received: \"Average service\"",
		},
		{
			name:         "2-star rating only",
			rating:       2,
			customerName: "",
			comment:      "",
			expected:     "New 2-star rating received",
		},
		{
			name:         "Invalid rating",
			rating:       0,
			customerName: "",
			comment:      "",
			expected:     "New rating received",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatNotificationMessage(tt.rating, tt.customerName, tt.comment)
			assert.Equal(t, tt.expected, result)
		})
	}
}
