package notification

import (
	"context"
	"fmt"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/notification-service/internal/config"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/circuitbreaker"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/log"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/retry"
)

// Service represents the notification service
type Service interface {
	CreateNotification(ctx context.Context, req RatingNotificationRequest) (*CreateNotificationResponse, error)
	GetNotifications(ctx context.Context, serviceProviderID string, lastChecked time.Time) (*GetNotificationsResponse, error)
	StartCleanupWorker(ctx context.Context)
}

// service implements the Service interface
type service struct {
	storage        Storage
	logger         log.Logger
	circuitBreaker *circuitbreaker.CircuitBreaker
	retryConfig    retry.RetryConfig
	cleanupConfig  config.CleanupConfig
}

// NewService creates a new notification service
func NewService(storage Storage, logger log.Logger, cfg config.Config) Service {
	return &service{
		storage:        storage,
		logger:         logger,
		circuitBreaker: circuitbreaker.New(cfg.CircuitBreaker, logger),
		retryConfig:    cfg.Retry,
		cleanupConfig:  cfg.Cleanup,
	}
}

// CreateNotification creates a new notification with circuit breaker and retry logic
func (s *service) CreateNotification(ctx context.Context, req RatingNotificationRequest) (*CreateNotificationResponse, error) {
	notification := NewNotification(req)

	err := s.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		return retry.WithRetry(ctx, s.retryConfig, func(ctx context.Context) error {
			return s.storage.StoreNotification(ctx, notification)
		}, s.isRetryableError, s.logger)
	})

	if err != nil {
		s.logger.With(ctx, "error", err, "service_provider_id", req.ServiceProviderID).
			Error("Failed to create notification")
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	s.logger.With(ctx, "notification_id", notification.ID, "service_provider_id", req.ServiceProviderID).
		Info("Successfully created notification")

	return &CreateNotificationResponse{
		ID:      notification.ID,
		Message: "Notification created successfully",
	}, nil
}

// GetNotifications retrieves notifications with circuit breaker protection
func (s *service) GetNotifications(ctx context.Context, serviceProviderID string, lastChecked time.Time) (*GetNotificationsResponse, error) {
	var notifications []Notification

	err := s.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		return retry.WithRetry(ctx, s.retryConfig, func(ctx context.Context) error {
			var retryErr error
			notifications, retryErr = s.storage.GetNotifications(ctx, serviceProviderID, lastChecked)
			return retryErr
		}, s.isRetryableError, s.logger)
	})

	if err != nil {
		s.logger.With(ctx, "error", err, "service_provider_id", serviceProviderID).
			Error("Failed to get notifications")
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	s.logger.With(ctx, "service_provider_id", serviceProviderID, "count", len(notifications)).
		Debug("Successfully retrieved notifications")
	if notifications == nil {
		notifications = []Notification{}
	}
	return &GetNotificationsResponse{
		Notifications: notifications,
		HasMore:       false,
	}, nil
}

// StartCleanupWorker starts a background worker to clean up old notifications
func (s *service) StartCleanupWorker(ctx context.Context) {
	s.logger.With(ctx, "interval", s.cleanupConfig.Interval, "max_age", s.cleanupConfig.MaxAge).
		Info("Starting notification cleanup worker")

	ticker := time.NewTicker(s.cleanupConfig.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Stopping notification cleanup worker")
			return
		case <-ticker.C:
			if err := s.storage.Cleanup(ctx, s.cleanupConfig.MaxAge); err != nil {
				s.logger.With(ctx, "error", err).Error("Failed to cleanup notifications")
			}
		}
	}
}

// isRetryableError determines if an error should trigger a retry
func (s *service) isRetryableError(err error) bool {
	// For simplicity, considered most errors as retryable
	return true
}
