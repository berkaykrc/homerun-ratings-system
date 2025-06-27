package notification

import (
	"time"

	"github.com/berkaykrc/homerun-ratings-system/notification-service/internal/errors"
	"github.com/berkaykrc/homerun-ratings-system/notification-service/pkg/log"
	routing "github.com/go-ozzo/ozzo-routing/v2"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// RegisterHandlers sets up the routing of the HTTP handlers.
func RegisterHandlers(rg *routing.Router, service Service, logger log.Logger) {
	res := resource{service, logger}

	// Internal endpoint for receiving notifications from rating service
	rg.Post("/api/internal/notifications", res.createNotification)

	// Public endpoint for service providers to get notifications
	rg.Get("/api/notifications/<serviceProviderId>", res.getNotifications)
}

type resource struct {
	service Service
	logger  log.Logger
}

// createNotification handles POST /api/internal/notifications
func (r resource) createNotification(c *routing.Context) error {
	var req RatingNotificationRequest
	if err := c.Read(&req); err != nil {
		r.logger.With(c.Request.Context(), "error", err).Error("Failed to parse notification request")
		return errors.BadRequest("Invalid request format")
	}

	if err := r.validateCreateNotificationRequest(req); err != nil {
		r.logger.With(c.Request.Context(), "error", err).Error("Invalid notification request")
		return err
	}

	resp, err := r.service.CreateNotification(c.Request.Context(), req)
	if err != nil {
		r.logger.With(c.Request.Context(), "error", err).Error("Failed to create notification")
		return err
	}

	return c.WriteWithStatus(resp, 201)
}

// getNotifications handles GET /api/notifications/{serviceProviderId}
func (r resource) getNotifications(c *routing.Context) error {
	serviceProviderID := c.Param("serviceProviderId")
	if serviceProviderID == "" {
		return errors.BadRequest("Service provider ID is required")
	}

	// Parse lastChecked parameter (optional)
	lastChecked := time.Time{} // Default to epoch if not provided
	if lastCheckedStr := c.Query("lastChecked"); lastCheckedStr != "" {
		parsed, err := time.Parse(time.RFC3339, lastCheckedStr)
		if err != nil {
			r.logger.With(c.Request.Context(), "error", err, "lastChecked", lastCheckedStr).
				Error("Invalid lastChecked format")
			return errors.BadRequest("Invalid lastChecked format. Use RFC3339 format (e.g., 2006-01-02T15:04:05Z)")
		}
		lastChecked = parsed
	}

	resp, err := r.service.GetNotifications(c.Request.Context(), serviceProviderID, lastChecked)
	if err != nil {
		r.logger.With(c.Request.Context(), "error", err, "service_provider_id", serviceProviderID).
			Error("Failed to get notifications")
		return err
	}

	return c.Write(resp)
}

// validateCreateNotificationRequest validates the create notification request
func (r resource) validateCreateNotificationRequest(req RatingNotificationRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ServiceProviderID, validation.Required, is.UUID),
		validation.Field(&req.RatingID, validation.Required, is.UUID),
		validation.Field(&req.Rating, validation.Required, validation.Min(1), validation.Max(5)),
		validation.Field(&req.CustomerName, validation.Length(1, 255)),
		validation.Field(&req.Comment, validation.Length(0, 1000)),
	)
}
