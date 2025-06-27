package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/rating-service/pkg/circuitbreaker"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/pkg/log"
	"github.com/berkaykrc/homerun-ratings-system/rating-service/pkg/retry"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Client interface for notification service communication
type Client interface {
	SendRatingNotification(ctx context.Context, notification RatingNotification) error
}

// RatingNotification represents the notification payload sent to the notification service
type RatingNotification struct {
	ServiceProviderID string `json:"serviceProviderId"`
	RatingID          string `json:"ratingId"`
	Rating            int    `json:"rating"`
	CustomerName      string `json:"customerName"`
	Comment           string `json:"comment"`
}

// Config represents notification service configuration
type Config struct {
	BaseURL       string                `yaml:"base_url" json:"baseUrl"`
	Timeout       time.Duration         `yaml:"timeout" json:"timeout"`
	RetryConfig   retry.RetryConfig     `yaml:"retry" json:"retry"`
	CircuitConfig circuitbreaker.Config `yaml:"circuit_breaker" json:"circuitBreaker"`
}

// Validate validates the notification service configuration
func (c Config) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.BaseURL, validation.Required, is.URL),
		validation.Field(&c.Timeout, validation.Required),
	)
}

// httpClient implements the Client interface using HTTP
type httpClient struct {
	client         *http.Client
	baseURL        string
	logger         log.Logger
	retryConfig    retry.RetryConfig
	circuitBreaker *circuitbreaker.CircuitBreaker
}

// NewHTTPClient creates a new HTTP-based notification client
func NewHTTPClient(config Config, logger log.Logger) Client {
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	retryConfig := config.RetryConfig
	if retryConfig.MaxAttempts == 0 {
		retryConfig = retry.DefaultRetryConfig()
	}

	circuitConfig := config.CircuitConfig
	if circuitConfig.FailureThreshold == 0 {
		circuitConfig = circuitbreaker.DefaultConfig()
	}

	return &httpClient{
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL:        config.BaseURL,
		logger:         logger,
		retryConfig:    retryConfig,
		circuitBreaker: circuitbreaker.New(circuitConfig, logger),
	}
}

// SendRatingNotification sends a rating notification using HTTP with retry and circuit breaker
func (c *httpClient) SendRatingNotification(ctx context.Context, notification RatingNotification) error {
	// Use circuit breaker to execute the HTTP call with retry
	return c.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		// Use retry mechanism
		return retry.WithRetry(ctx, c.retryConfig, func(ctx context.Context) error {
			return c.sendHTTPNotification(ctx, notification)
		}, c.isRetryableError, c.logger)
	})
}

// sendHTTPNotification performs the actual HTTP request
func (c *httpClient) sendHTTPNotification(ctx context.Context, notification RatingNotification) error {
	// Build the notification service endpoint
	url := fmt.Sprintf("%s/api/internal/notifications", c.baseURL)

	// Convert domain model to JSON
	jsonData, err := json.Marshal(notification)
	if err != nil {
		c.logger.With(ctx, "error", err).Error("Failed to marshal notification")
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Create HTTP request with context for cancellation
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		c.logger.With(ctx, "error", err).Error("Failed to create notification request")
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set HTTP headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "rating-service/1.0")

	// Log the outgoing request
	c.logger.With(ctx, "url", url, "method", "POST").Info("Sending rating notification")

	// Execute HTTP request
	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.With(ctx, "error", err, "url", url).Error("Failed to send notification")
		return fmt.Errorf("failed to send notification: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	// Check HTTP status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.With(ctx, "status_code", resp.StatusCode, "url", url).Error("Notification service returned error")
		return fmt.Errorf("notification service returned status %d", resp.StatusCode)
	}

	c.logger.With(ctx, "status_code", resp.StatusCode).Info("Successfully sent rating notification")
	return nil
}

// isRetryableError determines if an error should trigger a retry
func (c *httpClient) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorStr := err.Error()

	// Retry on network errors
	if strings.Contains(errorStr, "connection refused") ||
		strings.Contains(errorStr, "timeout") ||
		strings.Contains(errorStr, "context deadline exceeded") ||
		strings.Contains(errorStr, "no such host") {
		return true
	}

	// Retry on 5xx HTTP status codes
	if strings.Contains(errorStr, "status 5") {
		return true
	}

	// Don't retry on 4xx status codes (client errors)
	if strings.Contains(errorStr, "status 4") {
		return false
	}

	// By default, retry other errors
	return true
}

// mockClient is a test implementation of Client interface
type mockClient struct {
	shouldFail bool
	logger     log.Logger
}

// NewMockClient creates a mock notification client for testing purposes.
func NewMockClient(shouldFail bool, logger log.Logger) Client {
	return &mockClient{
		shouldFail: shouldFail,
		logger:     logger,
	}
}

func (m *mockClient) SendRatingNotification(ctx context.Context, notification RatingNotification) error {
	m.logger.With(ctx, "notification", notification).Info("Mock: Sending rating notification")

	if m.shouldFail {
		return fmt.Errorf("mock client configured to fail")
	}

	return nil
}
