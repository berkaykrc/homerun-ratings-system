package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/berkaykrc/homerun-ratings-system/rating-service/pkg/log"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// RetryConfig defines the configuration for retry logic
type RetryConfig struct {
	MaxAttempts   int           `yaml:"max_attempts" json:"maxAttempts"`
	InitialDelay  time.Duration `yaml:"initial_delay" json:"initialDelay"`
	MaxDelay      time.Duration `yaml:"max_delay" json:"maxDelay"`
	BackoffFactor float64       `yaml:"backoff_factor" json:"backoffFactor"`
	Jitter        bool          `yaml:"jitter" json:"jitter"`
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:   3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		Jitter:        true,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func(ctx context.Context) error

// IsRetryableError determines if an error should trigger a retry
type IsRetryableError func(error) bool

// DefaultRetryableError is the default function to determine if an error is retryable
func DefaultRetryableError(err error) bool {
	// For now, all errors as retryable
	return true
}

// Validate validates the retry configuration
func (c RetryConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.MaxAttempts, validation.Required, validation.Min(1), validation.Max(10)),
		validation.Field(&c.InitialDelay, validation.Required, validation.Min(time.Millisecond)),
		validation.Field(&c.MaxDelay, validation.Required, validation.Min(time.Millisecond)),
		validation.Field(&c.BackoffFactor, validation.Required, validation.Min(1.0)),
		validation.Field(&c.Jitter),
	)
}

// WithRetry executes a function with retry logic
func WithRetry(ctx context.Context, config RetryConfig, fn RetryableFunc, isRetryable IsRetryableError, logger log.Logger) error {
	if isRetryable == nil {
		isRetryable = DefaultRetryableError
	}

	var lastErr error

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Try to execute the function
		err := fn(ctx)
		if err == nil {
			if attempt > 1 {
				logger.With(ctx, "attempt", attempt).Info("Function succeeded after retry")
			}
			return nil
		}

		lastErr = err

		if attempt == config.MaxAttempts {
			logger.With(ctx, "attempt", attempt, "error", err).Error("All retry attempts exhausted")
			break
		}

		if !isRetryable(err) {
			logger.With(ctx, "attempt", attempt, "error", err).Info("Error is not retryable, stopping")
			break
		}

		delay := calculateDelay(attempt, config)

		logger.With(ctx, "attempt", attempt, "error", err, "delay_ms", delay.Milliseconds()).
			Infof("Function failed, retrying after delay")

		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("function failed after %d attempts: %w", config.MaxAttempts, lastErr)
}

// calculateDelay calculates the delay for the next retry attempt
func calculateDelay(attempt int, config RetryConfig) time.Duration {
	// Calculate exponential backoff
	delay := min(
		// Apply maximum delay limit
		time.Duration(float64(config.InitialDelay)*math.Pow(config.BackoffFactor, float64(attempt-1))), config.MaxDelay)

	// Add jitter to prevent thundering herd
	if config.Jitter {
		// Add random jitter up to 10% of the delay
		jitter := time.Duration(rand.Float64() * float64(delay) * 0.1)
		delay += jitter
	}

	return delay
}
