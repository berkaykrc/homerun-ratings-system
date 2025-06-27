package notification

import (
	"context"
	"sync"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/log"
)

// Storage represents the notification storage interface
type Storage interface {
	StoreNotification(ctx context.Context, notification Notification) error
	GetNotifications(ctx context.Context, serviceProviderID string, lastChecked time.Time) ([]Notification, error)
	Cleanup(ctx context.Context, maxAge time.Duration) error
}

// inMemoryStorage implements Storage using in-memory data structures
type inMemoryStorage struct {
	mu                     sync.RWMutex
	notifications          map[string][]Notification  // map[serviceProviderID][]Notification
	deliveredNotifications map[string]map[string]bool // map[serviceProviderID]map[notificationID]bool
	logger                 log.Logger
}

// NewInMemoryStorage creates a new in-memory storage instance
func NewInMemoryStorage(logger log.Logger) Storage {
	return &inMemoryStorage{
		notifications:          make(map[string][]Notification),
		deliveredNotifications: make(map[string]map[string]bool),
		logger:                 logger,
	}
}

// StoreNotification stores a notification in memory
func (s *inMemoryStorage) StoreNotification(ctx context.Context, notification Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.With(ctx, "service_provider_id", notification.ServiceProviderID, "notification_id", notification.ID).
		Info("Storing notification")

	s.notifications[notification.ServiceProviderID] = append(
		s.notifications[notification.ServiceProviderID],
		notification,
	)

	return nil
}

// GetNotifications retrieves notifications for a service provider created after the given timestamp
// and marks them as delivered so they won't be returned again
func (s *inMemoryStorage) GetNotifications(ctx context.Context, serviceProviderID string, lastChecked time.Time) ([]Notification, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.With(ctx, "service_provider_id", serviceProviderID, "last_checked", lastChecked).
		Debug("Retrieving notifications")

	allNotifications, exists := s.notifications[serviceProviderID]
	if !exists {
		return []Notification{}, nil
	}

	// Initialize delivered map for this service provider if it doesn't exist
	if s.deliveredNotifications[serviceProviderID] == nil {
		s.deliveredNotifications[serviceProviderID] = make(map[string]bool)
	}

	// Filter notifications created after lastChecked AND not yet delivered
	var newNotifications []Notification
	for _, notification := range allNotifications {
		if notification.CreatedAt.After(lastChecked) && !s.deliveredNotifications[serviceProviderID][notification.ID] {
			newNotifications = append(newNotifications, notification)
			s.deliveredNotifications[serviceProviderID][notification.ID] = true
		}
	}

	s.logger.With(ctx, "service_provider_id", serviceProviderID, "total_count", len(allNotifications), "new_count", len(newNotifications)).
		Debug("Retrieved and marked notifications as delivered")

	return newNotifications, nil
}

// Cleanup removes old notifications to prevent memory leaks
func (s *inMemoryStorage) Cleanup(ctx context.Context, maxAge time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	cutoff := time.Now().Add(-maxAge)
	totalRemoved := 0
	totalDeliveredRemoved := 0

	for serviceProviderID, notifications := range s.notifications {
		var keepNotifications []Notification
		removedCount := 0

		for _, notification := range notifications {
			if notification.CreatedAt.After(cutoff) {
				keepNotifications = append(keepNotifications, notification)
			} else {
				removedCount++
				// Also remove from delivered tracking when we remove the notification
				if s.deliveredNotifications[serviceProviderID] != nil {
					if s.deliveredNotifications[serviceProviderID][notification.ID] {
						delete(s.deliveredNotifications[serviceProviderID], notification.ID)
						totalDeliveredRemoved++
					}
				}
			}
		}

		if removedCount > 0 {
			s.notifications[serviceProviderID] = keepNotifications
			totalRemoved += removedCount
		}

		// Remove empty slices to save memory
		if len(keepNotifications) == 0 {
			delete(s.notifications, serviceProviderID)
		}

		// Remove empty delivered tracking maps to save memory
		if s.deliveredNotifications[serviceProviderID] != nil && len(s.deliveredNotifications[serviceProviderID]) == 0 {
			delete(s.deliveredNotifications, serviceProviderID)
		}
	}

	if totalRemoved > 0 {
		s.logger.With(ctx, "removed_notifications", totalRemoved, "removed_delivered_tracking", totalDeliveredRemoved, "cutoff", cutoff).
			Info("Cleaned up old notifications and delivery tracking")
	}

	return nil
}
