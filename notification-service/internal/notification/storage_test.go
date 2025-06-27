package notification

import (
	"context"
	"testing"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/log"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryStorage_StoreAndGetNotifications(t *testing.T) {
	logger, _ := log.NewForTest()
	storage := NewInMemoryStorage(logger)
	ctx := context.Background()

	// Test data
	serviceProviderID := "test-provider-id"
	notification1 := Notification{
		ID:                "notif-1",
		ServiceProviderID: serviceProviderID,
		Message:           "New 5-star rating received",
		RatingID:          "rating-1",
		CreatedAt:         time.Now(),
	}
	notification2 := Notification{
		ID:                "notif-2",
		ServiceProviderID: serviceProviderID,
		Message:           "New 4-star rating received",
		RatingID:          "rating-2",
		CreatedAt:         time.Now().Add(time.Minute),
	}

	// Store notifications
	err := storage.StoreNotification(ctx, notification1)
	assert.NoError(t, err)

	err = storage.StoreNotification(ctx, notification2)
	assert.NoError(t, err)

	// Get all notifications (first call)
	notifications, err := storage.GetNotifications(ctx, serviceProviderID, time.Time{})
	assert.NoError(t, err)
	assert.Len(t, notifications, 2)

	// Get all notifications again (second call) - should return empty since they were already delivered
	notifications, err = storage.GetNotifications(ctx, serviceProviderID, time.Time{})
	assert.NoError(t, err)
	assert.Len(t, notifications, 0, "Second call should return no notifications as they were already delivered")

	// Store a new notification
	notification3 := Notification{
		ID:                "notif-3",
		ServiceProviderID: serviceProviderID,
		Message:           "New 3-star rating received",
		RatingID:          "rating-3",
		CreatedAt:         time.Now().Add(2 * time.Minute),
	}
	err = storage.StoreNotification(ctx, notification3)
	assert.NoError(t, err)

	// Get notifications after storing new one - should return only the new notification
	notifications, err = storage.GetNotifications(ctx, serviceProviderID, time.Time{})
	assert.NoError(t, err)
	assert.Len(t, notifications, 1)
	assert.Equal(t, notification3.ID, notifications[0].ID)

	// Get notifications for non-existent provider
	notifications, err = storage.GetNotifications(ctx, "non-existent", time.Time{})
	assert.NoError(t, err)
	assert.Len(t, notifications, 0)
}

func TestInMemoryStorage_Cleanup(t *testing.T) {
	logger, _ := log.NewForTest()
	storage := NewInMemoryStorage(logger)
	ctx := context.Background()

	// Test data
	serviceProviderID := "test-provider-id"
	oldNotification := Notification{
		ID:                "old-notif",
		ServiceProviderID: serviceProviderID,
		Message:           "Old notification",
		RatingID:          "rating-old",
		CreatedAt:         time.Now().Add(-2 * time.Hour),
	}
	newNotification := Notification{
		ID:                "new-notif",
		ServiceProviderID: serviceProviderID,
		Message:           "New notification",
		RatingID:          "rating-new",
		CreatedAt:         time.Now(),
	}

	// Store notifications
	err := storage.StoreNotification(ctx, oldNotification)
	assert.NoError(t, err)

	err = storage.StoreNotification(ctx, newNotification)
	assert.NoError(t, err)

	// Cleanup old notifications (older than 1 hour)
	err = storage.Cleanup(ctx, time.Hour)
	assert.NoError(t, err)

	// Verify only new notification remains by getting notifications
	notifications, err := storage.GetNotifications(ctx, serviceProviderID, time.Time{})
	assert.NoError(t, err)
	assert.Len(t, notifications, 1, "After cleanup, only new notification should remain")
	assert.Equal(t, newNotification.ID, notifications[0].ID)
}
