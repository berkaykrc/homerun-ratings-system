package notification

import (
	"time"

	"github.com/google/uuid"
)

// Notification represents a notification in the system
type Notification struct {
	ID                string    `json:"id"`
	ServiceProviderID string    `json:"serviceProviderId"`
	Message           string    `json:"message"`
	RatingID          string    `json:"ratingId"`
	CreatedAt         time.Time `json:"createdAt"`
}

// RatingNotificationRequest represents an incoming notification from the rating service
type RatingNotificationRequest struct {
	ServiceProviderID string `json:"serviceProviderId"`
	RatingID          string `json:"ratingId"`
	Rating            int    `json:"rating"`
	CustomerName      string `json:"customerName"`
	Comment           string `json:"comment"`
}

// GetNotificationsResponse represents the response for getting notifications
type GetNotificationsResponse struct {
	Notifications []Notification `json:"notifications"`
	HasMore       bool           `json:"hasMore"`
}

// CreateNotificationResponse represents the response after creating a notification
type CreateNotificationResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

// NewNotification creates a new notification from a rating notification request
func NewNotification(req RatingNotificationRequest) Notification {
	message := formatNotificationMessage(req.Rating, req.CustomerName, req.Comment)

	return Notification{
		ID:                uuid.New().String(),
		ServiceProviderID: req.ServiceProviderID,
		Message:           message,
		RatingID:          req.RatingID,
		CreatedAt:         time.Now(),
	}
}

// formatNotificationMessage formats a notification message based on rating details
func formatNotificationMessage(rating int, customerName, comment string) string {
	baseMessage := ""

	switch rating {
	case 1:
		baseMessage = "New 1-star rating received"
	case 2:
		baseMessage = "New 2-star rating received"
	case 3:
		baseMessage = "New 3-star rating received"
	case 4:
		baseMessage = "New 4-star rating received"
	case 5:
		baseMessage = "New 5-star rating received"
	default:
		baseMessage = "New rating received"
	}

	if customerName != "" {
		baseMessage += " from " + customerName
	}

	if comment != "" {
		baseMessage += ": \"" + comment + "\""
	}

	return baseMessage
}
